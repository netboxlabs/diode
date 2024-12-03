-- name: CreateIngestionLog :one
INSERT INTO ingestion_logs (ingestion_log_ksuid, data_type, state, request_id, ingestion_ts, producer_app_name,
                            producer_app_version, sdk_name, sdk_version, entity, source_metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING *;
