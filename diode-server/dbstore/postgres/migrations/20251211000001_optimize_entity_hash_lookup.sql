-- +goose Up

-- Composite index for filtering by entity_hash and ordering by created_at
CREATE INDEX idx_ingestion_logs_entity_hash_created_at 
ON ingestion_logs(entity_hash, created_at DESC);

-- Composite index to speed up latest change_set lookup per ingestion_log
CREATE INDEX idx_change_sets_ingestion_log_id_id 
ON change_sets(ingestion_log_id, id DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_change_sets_ingestion_log_id_id;
DROP INDEX IF EXISTS idx_ingestion_logs_entity_hash_created_at;

