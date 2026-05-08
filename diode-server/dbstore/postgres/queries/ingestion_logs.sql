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
