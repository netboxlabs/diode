-- name: RetrieveDeviations :many
SELECT v_deviations.*
FROM v_deviations
WHERE (v_deviations.state = ANY (sqlc.narg('state')::int[]) OR sqlc.narg('state') IS NULL)
  AND (v_deviations.object_type = ANY (sqlc.narg('object_type')::text[]) OR
       sqlc.narg('object_type') IS NULL)
  AND (v_deviations.change_set ->> 'branch_id' = ANY (sqlc.narg('branch_id')::text[]) OR
       sqlc.narg('branch_id') IS NULL)
  AND (v_deviations.ingestion_ts >= sqlc.narg('ingestion_ts_start') OR
       sqlc.narg('ingestion_ts_start') IS NULL)
  AND (v_deviations.ingestion_ts <= sqlc.narg('ingestion_ts_end') OR
       sqlc.narg('ingestion_ts_end') IS NULL)
ORDER BY v_deviations.id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: RetrieveDeviationByID :one
SELECT v_deviations.*
FROM v_deviations
WHERE v_deviations.external_id = $1;

-- name: ListResultsByJob :many
SELECT v_deviations.*
FROM v_deviations
WHERE (v_deviations.source_metadata->>'job_id' = ANY (sqlc.narg('job_id')::text[]) OR
       sqlc.narg('job_id') IS NULL)
  AND (v_deviations.state = ANY (sqlc.narg('state')::int[]) OR
       sqlc.narg('state') IS NULL)
  AND (v_deviations.object_type = ANY (sqlc.narg('object_type')::text[]) OR
       sqlc.narg('object_type') IS NULL)
  AND (v_deviations.change_set ->> 'branch_id' = ANY (sqlc.narg('branch_id')::text[]) OR
       sqlc.narg('branch_id') IS NULL)
  AND (v_deviations.ingestion_ts >= sqlc.narg('ingestion_ts_start') OR
       sqlc.narg('ingestion_ts_start') IS NULL)
  AND (v_deviations.ingestion_ts <= sqlc.narg('ingestion_ts_end') OR
       sqlc.narg('ingestion_ts_end') IS NULL)
ORDER BY v_deviations.id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
