-- name: CreateIngestionLog :one
-- Inserts the log and its stats row in one statement. ingestion_logs is
-- insert-only: the columns that change after insert live in
-- ingestion_logs_stats. Readers use v_ingestion_logs, which joins the two
-- into the pre-split row shape; state and bookkeeping writes touch only
-- ingestion_logs_stats.
WITH il AS (
    INSERT INTO ingestion_logs (external_id, object_type, request_id, ingestion_ts, source_ts, producer_app_name,
                                producer_app_version, sdk_name, sdk_version, entity, source_metadata, entity_hash)
    VALUES (sqlc.arg('external_id'), sqlc.narg('object_type'), sqlc.narg('request_id'), sqlc.narg('ingestion_ts'),
            sqlc.narg('source_ts'), sqlc.narg('producer_app_name'), sqlc.narg('producer_app_version'),
            sqlc.narg('sdk_name'), sqlc.narg('sdk_version'), sqlc.arg('entity'), sqlc.arg('source_metadata'),
            sqlc.narg('entity_hash'))
    RETURNING id, object_type, ingestion_ts
), stats AS (
    INSERT INTO ingestion_logs_stats (ingestion_log_id, state, object_type, ingestion_ts)
    SELECT id, sqlc.narg('state')::int4, object_type, ingestion_ts FROM il
)
SELECT id FROM il;

-- name: UpdateIngestionLogStateWithError :exec
UPDATE ingestion_logs_stats
SET state = sqlc.narg('state'),
    error = sqlc.arg('error'),
    requeued_from_state = NULL
WHERE ingestion_log_id = sqlc.arg('id');

-- name: CountIngestionLogsPerState :many
SELECT state, COUNT(*) AS count
FROM ingestion_logs_stats
GROUP BY state;

-- name: RetrieveIngestionLogs :many
SELECT *
FROM v_ingestion_logs
WHERE (state = sqlc.narg('state') OR sqlc.narg('state') IS NULL)
  AND (object_type = sqlc.narg('object_type') OR sqlc.narg('object_type') IS NULL)
  AND (ingestion_ts >= sqlc.narg('ingestion_ts_start') OR sqlc.narg('ingestion_ts_start') IS NULL)
  AND (ingestion_ts <= sqlc.narg('ingestion_ts_end') OR sqlc.narg('ingestion_ts_end') IS NULL)
ORDER BY id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: RetrieveIngestionLogByExternalID :one
SELECT *
FROM v_ingestion_logs
WHERE external_id = $1;

-- name: RetrieveIngestionLogsWithChangeSets :many
SELECT v_deviations.*
FROM v_deviations
WHERE (v_deviations.state = sqlc.narg('state') OR sqlc.narg('state') IS NULL)
  AND (v_deviations.object_type = sqlc.narg('object_type') OR sqlc.narg('object_type') IS NULL)
  AND (v_deviations.ingestion_ts >= sqlc.narg('ingestion_ts_start') OR
       sqlc.narg('ingestion_ts_start') IS NULL)
  AND (v_deviations.ingestion_ts <= sqlc.narg('ingestion_ts_end') OR
       sqlc.narg('ingestion_ts_end') IS NULL)
ORDER BY v_deviations.id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: FindPriorIngestionLogByEntityHash :one
SELECT il.*
FROM v_ingestion_logs il
LEFT JOIN LATERAL (
    SELECT branch_id
    FROM change_sets cs
    WHERE cs.ingestion_log_id = il.id
    ORDER BY cs.id DESC
    LIMIT 1
) lcs ON true
WHERE il.entity_hash = sqlc.arg('entity_hash')
  AND lcs.branch_id IS NOT DISTINCT FROM sqlc.narg('branch_id')::text
ORDER BY il.created_at DESC, il.id DESC
LIMIT 1;

-- name: FindPriorIngestionLogsByEntityHashes :many
SELECT il.*
FROM unnest(@entity_hashes::text[]) AS h(entity_hash)
CROSS JOIN LATERAL (
    SELECT il2.*
    FROM v_ingestion_logs il2
    WHERE il2.entity_hash = h.entity_hash
      AND (
          SELECT cs.branch_id
          FROM change_sets cs
          WHERE cs.ingestion_log_id = il2.id
          ORDER BY cs.id DESC
          LIMIT 1
      ) IS NOT DISTINCT FROM sqlc.narg('branch_id')::text
    ORDER BY il2.created_at DESC, il2.id DESC
    LIMIT 1
) il;

-- name: BulkCreateIngestionLogs :copyfrom
INSERT INTO ingestion_logs (id, external_id, object_type, request_id, ingestion_ts, source_ts,
                            producer_app_name, producer_app_version, sdk_name, sdk_version,
                            entity, source_metadata, entity_hash)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: BulkCreateIngestionLogsStats :copyfrom
-- Stats rows for BulkCreateIngestionLogs, copied in the same transaction.
INSERT INTO ingestion_logs_stats (ingestion_log_id, state, object_type, ingestion_ts)
VALUES ($1, $2, $3, $4);

-- name: FindIngestionLogIDsByExternalIDs :many
SELECT id, external_id FROM ingestion_logs WHERE external_id = ANY(@external_ids::text[]);

-- name: BulkMarkDuplicates :many
-- Increment duplicate_count/last_seen for prior ingestion logs of
-- deduplicated entities and atomically requeue rows whose NetBox state may
-- have drifted: APPLIED (3), FAILED (4), NO_CHANGES (5) -> QUEUED (1) so the
-- pollers re-plan (and re-apply under auto-apply). QUEUED (1) is already
-- pending, OPEN (2) may be claimed or awaiting review, IGNORED (6) is a user
-- opt-out, APPLYING (8) is in flight -- those keep their state and only get
-- the duplicate bookkeeping.
--
-- @ids and @increments are parallel arrays (ids unique): a batch carrying n
-- copies of the same entity increments its row's duplicate_count by n in one
-- statement instead of n round-trips.
--
-- The inner SELECT ... ORDER BY id FOR UPDATE acquires row locks in the same
-- global order as ClaimQueuedIngestionLogs/ClaimQueuedForAutoApply to avoid
-- deadlocks with concurrent claim/persist updates. The state guard is
-- evaluated on the locked row version, so a concurrent claim (1->2 / 1->8)
-- committed before we acquire the lock is correctly skipped.
UPDATE ingestion_logs_stats s
SET duplicate_count = s.duplicate_count + bulk.increment,
    last_seen = CURRENT_TIMESTAMP,
    state = CASE WHEN locked.prev_state IN (3, 4, 5) THEN 1 ELSE s.state END,
    error = CASE WHEN locked.prev_state IN (3, 4, 5) THEN NULL ELSE s.error END,
    requeued_from_state = CASE WHEN locked.prev_state IN (3, 4, 5) THEN locked.prev_state ELSE s.requeued_from_state END
FROM (
    SELECT ingestion_log_id, state AS prev_state
    FROM ingestion_logs_stats
    WHERE ingestion_log_id = ANY(@ids::int4[])
    ORDER BY ingestion_log_id
    FOR UPDATE
) locked
JOIN (
    SELECT unnest(@ids::int4[]) AS id,
           unnest(@increments::int4[]) AS increment
) bulk ON bulk.id = locked.ingestion_log_id
WHERE s.ingestion_log_id = locked.ingestion_log_id
RETURNING s.ingestion_log_id AS id, (locked.prev_state IN (3, 4, 5))::bool AS requeued;

-- name: AcquireEntityHashLocks :exec
-- Take transaction-scoped advisory locks on each entity hash, in sorted order
-- to avoid lock-order deadlocks between concurrent dedup transactions. Held
-- until commit/rollback, these serialize the find-prior -> insert window per
-- entity hash so two concurrent batches of the same entity cannot both miss
-- the prior lookup and insert their own rows. Locks on distinct hashes do not
-- contend (modulo 64-bit hashtextextended collisions, which only cost a
-- harmless extra serialization).
SELECT pg_advisory_xact_lock(hashtextextended(h, 0))
FROM (SELECT unnest(@entity_hashes::text[]) AS h ORDER BY 1) AS hashes;

-- name: CloneIngestionLogForDrift :one
-- Clone a prior ingestion log into a new row representing a freshly detected
-- deviation (NetBox state drifted since the prior was applied). Entity data,
-- hash and source fields are copied; the new row gets its own external ID,
-- terminal state and ingestion timestamp. Dedup lookups pick the newest row
-- per entity hash, so subsequent duplicates track the new deviation.
WITH il AS (
    INSERT INTO ingestion_logs (external_id, object_type, request_id, ingestion_ts, source_ts,
                                producer_app_name, producer_app_version, sdk_name, sdk_version,
                                entity, source_metadata, entity_hash)
    SELECT @new_external_id, object_type, request_id, @ingestion_ts::bigint, source_ts,
           producer_app_name, producer_app_version, sdk_name, sdk_version,
           entity, source_metadata, entity_hash
    FROM ingestion_logs AS prior
    WHERE prior.id = @prior_id
    RETURNING id, object_type, ingestion_ts
), stats AS (
    INSERT INTO ingestion_logs_stats (ingestion_log_id, state, object_type, ingestion_ts)
    SELECT id, @new_state::int4, object_type, ingestion_ts FROM il
)
SELECT id FROM il;

-- name: BulkUpdateIngestionLogStates :exec
UPDATE ingestion_logs_stats s
SET state = bulk.new_state,
    error = NULL,
    requeued_from_state = NULL
FROM (
    SELECT unnest(@ids::int4[]) AS id,
           unnest(@states::int4[]) AS new_state
) bulk
WHERE s.ingestion_log_id = bulk.id;

-- name: ClaimQueuedIngestionLogs :many
-- The claimed rows come back in the v_ingestion_logs shape. A statement
-- cannot read its own CTE's update through the view, so the changed columns
-- come from RETURNING and the rest from ingestion_logs.
WITH claimed AS (
    UPDATE ingestion_logs_stats
    SET state = 2
    WHERE ingestion_log_id IN (
        SELECT ingestion_log_id FROM ingestion_logs_stats
        WHERE state = 1
        ORDER BY ingestion_log_id
        LIMIT sqlc.arg('batch_size')
        FOR UPDATE SKIP LOCKED
    )
    RETURNING *
)
SELECT il.id, il.external_id, il.object_type, c.state, il.request_id, il.ingestion_ts, il.source_ts,
       il.producer_app_name, il.producer_app_version, il.sdk_name, il.sdk_version, il.entity, c.error,
       il.source_metadata, il.created_at, c.updated_at, il.entity_hash, c.last_seen, c.duplicate_count,
       c.requeued_from_state
FROM claimed c
JOIN ingestion_logs il ON il.id = c.ingestion_log_id;

-- name: ClaimQueuedForAutoApply :many
-- Claim a batch of QUEUED ingestion logs for the AutoApplyProcessor (combined
-- plan+apply via /bulk-plan-apply). Transitions QUEUED (1) -> APPLYING (8).
-- A row stays in APPLYING for the duration of the NetBox round-trip and is
-- reset back to QUEUED on reconciler startup via ResetApplyingIngestionLogs.
WITH claimed AS (
    UPDATE ingestion_logs_stats
    SET state = 8
    WHERE ingestion_log_id IN (
        SELECT ingestion_log_id FROM ingestion_logs_stats
        WHERE state = 1
        ORDER BY ingestion_log_id
        LIMIT sqlc.arg('batch_size')
        FOR UPDATE SKIP LOCKED
    )
    RETURNING *
)
SELECT il.id, il.external_id, il.object_type, c.state, il.request_id, il.ingestion_ts, il.source_ts,
       il.producer_app_name, il.producer_app_version, il.sdk_name, il.sdk_version, il.entity, c.error,
       il.source_metadata, il.created_at, c.updated_at, il.entity_hash, c.last_seen, c.duplicate_count,
       c.requeued_from_state
FROM claimed c
JOIN ingestion_logs il ON il.id = c.ingestion_log_id;

-- name: ResetApplyingIngestionLogs :exec
-- Reset rows stuck in APPLYING (worker died mid-batch) back to QUEUED so the
-- AutoApplyProcessor reclaims them. Idempotent — safe to run on every startup.
UPDATE ingestion_logs_stats
SET state = 1
WHERE state = 8;
