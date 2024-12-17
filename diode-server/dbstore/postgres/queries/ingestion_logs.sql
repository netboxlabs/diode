-- name: CreateIngestionLog :one
INSERT INTO ingestion_logs (ingestion_log_uuid, data_type, state, request_id, ingestion_ts, producer_app_name,
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
  AND (data_type = sqlc.narg('data_type') OR sqlc.narg('data_type') IS NULL)
  AND (ingestion_ts >= sqlc.narg('ingestion_ts_start') OR sqlc.narg('ingestion_ts_start') IS NULL)
  AND (ingestion_ts <= sqlc.narg('ingestion_ts_end') OR sqlc.narg('ingestion_ts_end') IS NULL)
ORDER BY id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: RetrieveIngestionLogsWithChangeSets :many
SELECT sqlc.embed(ingestion_logs), sqlc.embed(change_sets), sqlc.embed(changes_view)
FROM ingestion_logs
         LEFT JOIN change_sets on ingestion_logs.id = change_sets.ingestion_log_id
         LEFT JOIN changes_view on change_sets.id = changes_view.change_set_id
WHERE (ingestion_logs.state = sqlc.narg('state') OR sqlc.narg('state') IS NULL)
  AND (ingestion_logs.data_type = sqlc.narg('data_type') OR sqlc.narg('data_type') IS NULL)
  AND (ingestion_logs.ingestion_ts >= sqlc.narg('ingestion_ts_start') OR sqlc.narg('ingestion_ts_start') IS NULL)
  AND (ingestion_logs.ingestion_ts <= sqlc.narg('ingestion_ts_end') OR sqlc.narg('ingestion_ts_end') IS NULL)
ORDER BY ingestion_logs.id DESC, changes_view.sequence_number ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
