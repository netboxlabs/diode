-- name: CreateChangeSet :one

INSERT INTO change_sets (external_id, ingestion_log_id, branch_id, deviation_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateChange :one

INSERT INTO changes (external_id, change_set_id, change_type, object_type, object_primary_value, object_id,
                     ref_id, object_version, before, after, new_refs,
                     sequence_number)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: BulkCreateChanges :copyfrom
INSERT INTO changes (external_id, change_set_id, change_type, object_type, object_primary_value, object_id,
                     ref_id, object_version, before, after, new_refs,
                     sequence_number)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: TruncateChangeSets :exec
DELETE FROM change_sets cs1
WHERE cs1.ingestion_log_id = $1
  AND cs1.id NOT IN (
    SELECT cs2.id
    FROM change_sets cs2
    WHERE cs2.ingestion_log_id = $1
    ORDER BY cs2.id DESC
    LIMIT $2
  );

-- name: BulkCreateChangeSets :copyfrom
INSERT INTO change_sets (id, external_id, ingestion_log_id, branch_id, deviation_name)
VALUES ($1, $2, $3, $4, $5);

-- name: BulkTruncateChangeSets :exec
DELETE FROM change_sets
WHERE change_sets.id IN (
    SELECT change_sets.id
    FROM change_sets
    WHERE change_sets.ingestion_log_id = ANY(@ingestion_log_ids::int4[])
      AND change_sets.id NOT IN (
          SELECT cs2.id
          FROM change_sets cs2
          WHERE cs2.ingestion_log_id = change_sets.ingestion_log_id
          ORDER BY cs2.id DESC
          LIMIT @keep_count
      )
);