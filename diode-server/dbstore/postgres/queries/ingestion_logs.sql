-- name: CreateIngestionLog :one
INSERT INTO ingestion_logs (external_id, object_type, state, request_id, ingestion_ts, source_ts, producer_app_name,
                            producer_app_version, sdk_name, sdk_version, entity, source_metadata, entity_hash)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: UpdateIngestionLogStateWithError :exec
UPDATE ingestion_logs
SET state = $2,
    error = $3,
    requeued_from_state = NULL
WHERE id = $1
RETURNING *;

-- name: CountIngestionLogsPerState :many
SELECT state, COUNT(*) AS count
FROM ingestion_logs
GROUP BY state;

-- name: RetrieveIngestionLogs :many
SELECT *
FROM ingestion_logs
WHERE (state = sqlc.narg('state') OR sqlc.narg('state') IS NULL)
  AND (object_type = sqlc.narg('object_type') OR sqlc.narg('object_type') IS NULL)
  AND (ingestion_ts >= sqlc.narg('ingestion_ts_start') OR sqlc.narg('ingestion_ts_start') IS NULL)
  AND (ingestion_ts <= sqlc.narg('ingestion_ts_end') OR sqlc.narg('ingestion_ts_end') IS NULL)
ORDER BY id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: RetrieveIngestionLogByExternalID :one
SELECT *
FROM ingestion_logs
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
FROM ingestion_logs il
LEFT JOIN LATERAL (
    SELECT branch_id
    FROM change_sets cs
    WHERE cs.ingestion_log_id = il.id
    ORDER BY cs.id DESC
    LIMIT 1
) lcs ON true
WHERE il.entity_hash = sqlc.arg('entity_hash')
  AND lcs.branch_id IS NOT DISTINCT FROM sqlc.narg('branch_id')::text
ORDER BY il.created_at DESC
LIMIT 1;

-- name: FindPriorIngestionLogsByEntityHashes :many
SELECT il.*
FROM unnest(@entity_hashes::text[]) AS h(entity_hash)
CROSS JOIN LATERAL (
    SELECT il2.*
    FROM ingestion_logs il2
    WHERE il2.entity_hash = h.entity_hash
      AND (
          SELECT cs.branch_id
          FROM change_sets cs
          WHERE cs.ingestion_log_id = il2.id
          ORDER BY cs.id DESC
          LIMIT 1
      ) IS NOT DISTINCT FROM sqlc.narg('branch_id')::text
    ORDER BY il2.created_at DESC
    LIMIT 1
) il;

-- name: BulkCreateIngestionLogs :copyfrom
INSERT INTO ingestion_logs (id, external_id, object_type, state, request_id, ingestion_ts, source_ts,
                            producer_app_name, producer_app_version, sdk_name, sdk_version,
                            entity, source_metadata, entity_hash)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

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
-- The inner SELECT ... ORDER BY id FOR UPDATE acquires row locks in the same
-- global order as ClaimQueuedIngestionLogs/ClaimQueuedForAutoApply to avoid
-- deadlocks with concurrent claim/persist updates. The state guard is
-- evaluated on the locked row version, so a concurrent claim (1->2 / 1->8)
-- committed before we acquire the lock is correctly skipped.
UPDATE ingestion_logs il
SET duplicate_count = il.duplicate_count + 1,
    last_seen = CURRENT_TIMESTAMP,
    state = CASE WHEN locked.prev_state IN (3, 4, 5) THEN 1 ELSE il.state END,
    error = CASE WHEN locked.prev_state IN (3, 4, 5) THEN NULL ELSE il.error END,
    requeued_from_state = CASE WHEN locked.prev_state IN (3, 4, 5) THEN locked.prev_state ELSE il.requeued_from_state END
FROM (
    SELECT id, state AS prev_state
    FROM ingestion_logs
    WHERE id = ANY(@ids::int4[])
    ORDER BY id
    FOR UPDATE
) locked
WHERE il.id = locked.id
RETURNING il.id, (locked.prev_state IN (3, 4, 5))::bool AS requeued;

-- name: CloneIngestionLogForDrift :one
-- Clone a prior ingestion log into a new row representing a freshly detected
-- deviation (NetBox state drifted since the prior was applied). Entity data,
-- hash and source fields are copied; the new row gets its own external ID,
-- terminal state and ingestion timestamp. Dedup lookups pick the newest row
-- per entity hash, so subsequent duplicates track the new deviation.
INSERT INTO ingestion_logs (external_id, object_type, state, request_id, ingestion_ts, source_ts,
                            producer_app_name, producer_app_version, sdk_name, sdk_version,
                            entity, source_metadata, entity_hash)
SELECT @new_external_id, object_type, @new_state::int4, request_id, @ingestion_ts::bigint, source_ts,
       producer_app_name, producer_app_version, sdk_name, sdk_version,
       entity, source_metadata, entity_hash
FROM ingestion_logs AS prior
WHERE prior.id = @prior_id
RETURNING ingestion_logs.id;

-- name: BulkUpdateIngestionLogStates :exec
UPDATE ingestion_logs il
SET state = bulk.new_state,
    error = NULL,
    requeued_from_state = NULL
FROM (
    SELECT unnest(@ids::int4[]) AS id,
           unnest(@states::int4[]) AS new_state
) bulk
WHERE il.id = bulk.id;

-- name: ClaimQueuedIngestionLogs :many
UPDATE ingestion_logs
SET state = 2
WHERE id IN (
    SELECT id FROM ingestion_logs
    WHERE state = 1
    ORDER BY id
    LIMIT sqlc.arg('batch_size')
    FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: ClaimQueuedForAutoApply :many
-- Claim a batch of QUEUED ingestion logs for the AutoApplyProcessor (combined
-- plan+apply via /bulk-plan-apply). Transitions QUEUED (1) -> APPLYING (8).
-- A row stays in APPLYING for the duration of the NetBox round-trip and is
-- reset back to QUEUED on reconciler startup via ResetApplyingIngestionLogs.
UPDATE ingestion_logs
SET state = 8
WHERE id IN (
    SELECT id FROM ingestion_logs
    WHERE state = 1
    ORDER BY id
    LIMIT sqlc.arg('batch_size')
    FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: ResetApplyingIngestionLogs :exec
-- Reset rows stuck in APPLYING (worker died mid-batch) back to QUEUED so the
-- AutoApplyProcessor reclaims them. Idempotent — safe to run on every startup.
UPDATE ingestion_logs
SET state = 1
WHERE state = 8;
