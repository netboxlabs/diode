-- name: CreateIngestionLog :one
INSERT INTO ingestion_logs (external_id, object_type, state, request_id, ingestion_ts, source_ts, producer_app_name,
                            producer_app_version, sdk_name, sdk_version, entity, source_metadata, entity_hash)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: UpdateIngestionLogStateWithError :exec
UPDATE ingestion_logs
SET state = $2,
    error = $3
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
  -- Exclude terminal ERRORED (7) so a re-ingest after the system gives up
  -- re-queues instead of deduping against the dead row.
  AND il.state IS DISTINCT FROM 7
ORDER BY il.created_at DESC
LIMIT 1;

-- name: IncrementDuplicateCount :exec
UPDATE ingestion_logs
SET duplicate_count = duplicate_count + 1,
    last_seen = CURRENT_TIMESTAMP
WHERE id = $1;

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
      -- Exclude terminal ERRORED (7): see FindPriorIngestionLogByEntityHash.
      -- Re-ingest after the retrier gives up must re-queue, not dedupe.
      AND il2.state IS DISTINCT FROM 7
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

-- name: BulkIncrementDuplicateCounts :exec
UPDATE ingestion_logs
SET duplicate_count = duplicate_count + 1,
    last_seen = CURRENT_TIMESTAMP
WHERE id = ANY(@ids::int4[]);

-- name: BulkUpdateIngestionLogStates :exec
UPDATE ingestion_logs il
SET state = bulk.new_state,
    error = NULL
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
-- Claim a batch for the AutoApplyProcessor: fresh QUEUED (1) plus retry-eligible
-- PENDING_RETRY (9) rows whose backoff has elapsed, transitioned to APPLYING (8).
-- Ordered by id (not fresh-first) so a due retry is processed in line rather than
-- starved behind fresh work when the queue never empties.
UPDATE ingestion_logs
SET state = 8
WHERE id IN (
    SELECT id FROM ingestion_logs
    WHERE state = 1
       OR (state = 9 AND (next_retry_at IS NULL OR next_retry_at <= NOW()))
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

-- name: MarkIngestionLogRetry :exec
-- Record a failed apply: increment retry_count and either re-arm as PENDING_RETRY
-- (9) with a jittered exponential backoff (base*2^n capped at max, ×random[0.5,1)
-- to spread retry herds), or retire to terminal ERRORED (7) once the budget is
-- spent. SET expressions read the pre-update retry_count.
UPDATE ingestion_logs
SET retry_count = retry_count + 1,
    state = CASE
        WHEN retry_count + 1 >= sqlc.arg('max_retries')::int THEN 7
        ELSE 9
    END,
    next_retry_at = CASE
        WHEN retry_count + 1 >= sqlc.arg('max_retries')::int THEN NULL
        ELSE NOW() + make_interval(secs => GREATEST(1, (
            LEAST(
                sqlc.arg('base_backoff_secs')::bigint * (1::bigint << LEAST(retry_count, 30)),
                sqlc.arg('max_backoff_secs')::bigint
            )::double precision * (0.5 + random() * 0.5)
        )::int))
    END,
    error = sqlc.arg('error')
WHERE id = sqlc.arg('id');
