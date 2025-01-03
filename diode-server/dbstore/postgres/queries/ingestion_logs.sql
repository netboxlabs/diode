-- name: CreateIngestionLog :one
INSERT INTO ingestion_logs (external_id, object_type, state, request_id, ingestion_ts, producer_app_name,
                            producer_app_version, sdk_name, sdk_version, entity, source_metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
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
SELECT v_ingestion_logs_with_change_set.*
FROM v_ingestion_logs_with_change_set
WHERE (v_ingestion_logs_with_change_set.state = sqlc.narg('state') OR sqlc.narg('state') IS NULL)
  AND (v_ingestion_logs_with_change_set.object_type = sqlc.narg('object_type') OR sqlc.narg('object_type') IS NULL)
  AND (v_ingestion_logs_with_change_set.ingestion_ts >= sqlc.narg('ingestion_ts_start') OR
       sqlc.narg('ingestion_ts_start') IS NULL)
  AND (v_ingestion_logs_with_change_set.ingestion_ts <= sqlc.narg('ingestion_ts_end') OR
       sqlc.narg('ingestion_ts_end') IS NULL)
ORDER BY v_ingestion_logs_with_change_set.id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
