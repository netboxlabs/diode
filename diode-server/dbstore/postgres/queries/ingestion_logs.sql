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
WHERE (duplicate_of_id IS NULL OR sqlc.narg('include_duplicates')::boolean = true)
GROUP BY state;

-- name: RetrieveIngestionLogs :many
SELECT *
FROM ingestion_logs
WHERE (state = sqlc.narg('state') OR sqlc.narg('state') IS NULL)
  AND (object_type = sqlc.narg('object_type') OR sqlc.narg('object_type') IS NULL)
  AND (ingestion_ts >= sqlc.narg('ingestion_ts_start') OR sqlc.narg('ingestion_ts_start') IS NULL)
  AND (ingestion_ts <= sqlc.narg('ingestion_ts_end') OR sqlc.narg('ingestion_ts_end') IS NULL)
  AND (duplicate_of_id IS NULL OR sqlc.narg('include_duplicates')::boolean = true)
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
  AND (v_deviations.duplicate_of_id IS NULL OR sqlc.narg('include_duplicates')::boolean = true)
ORDER BY v_deviations.id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: FindPriorIngestionLogByEntityHash :one
WITH latest_change_sets AS (
    SELECT DISTINCT ON (ingestion_log_id) 
        ingestion_log_id, 
        branch_id
    FROM change_sets
    ORDER BY ingestion_log_id, id DESC
)
SELECT il.*
FROM ingestion_logs il
LEFT JOIN latest_change_sets lcs ON il.id = lcs.ingestion_log_id
WHERE il.entity_hash = sqlc.arg('entity_hash')
  AND il.duplicate_of_id IS NULL
  AND (
    (sqlc.narg('branch_id')::text IS NOT NULL AND lcs.branch_id = sqlc.narg('branch_id')::text)
    OR
    (sqlc.narg('branch_id')::text IS NULL AND lcs.branch_id IS NULL)
  )
ORDER BY il.created_at DESC
LIMIT 1;

-- name: SetIngestionLogDuplicateOfID :exec
UPDATE ingestion_logs
SET duplicate_of_id = $2
WHERE id = $1;

-- name: RetrieveIngestionLogDuplicateOfID :one
SELECT duplicate_of_id
FROM ingestion_logs
WHERE id = $1;

