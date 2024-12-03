-- name: CreateChangeSet :one

INSERT INTO change_sets (change_set_ksuid, ingestion_log_id, branch_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateChange :one

INSERT INTO changes (change_ksuid, change_set_id, change_type, object_type, object_id, object_version, data,
                     sequence_number)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
