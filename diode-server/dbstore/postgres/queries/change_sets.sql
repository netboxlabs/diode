-- name: CreateChangeSet :one

INSERT INTO change_sets (external_id, ingestion_log_id, branch_id, deviation_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateChange :one

INSERT INTO changes (external_id, change_set_id, change_type, object_type, object_id, object_version, data,
                     sequence_number)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
