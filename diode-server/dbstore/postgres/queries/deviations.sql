-- name: RetrieveDeviations :many
SELECT v_ingestion_logs_with_change_set.*
FROM v_ingestion_logs_with_change_set
WHERE (v_ingestion_logs_with_change_set.state = ANY (sqlc.narg('state')::text[]) OR sqlc.narg('state') IS NULL)
  AND (v_ingestion_logs_with_change_set.object_type = ANY (sqlc.narg('object_type')::text[]) OR
       sqlc.narg('object_type') IS NULL)
  AND (v_ingestion_logs_with_change_set.branch_id = ANY (sqlc.narg('branch_id')::text[]) OR
       sqlc.narg('branch_id') IS NULL)
  AND (v_ingestion_logs_with_change_set.ingestion_ts >= sqlc.narg('ingestion_ts_start') OR
       sqlc.narg('ingestion_ts_start') IS NULL)
  AND (v_ingestion_logs_with_change_set.ingestion_ts <= sqlc.narg('ingestion_ts_end') OR
       sqlc.narg('ingestion_ts_end') IS NULL)
ORDER BY v_ingestion_logs_with_change_set.id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
