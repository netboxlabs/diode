-- +goose Up
-- +goose StatementBegin
-- Rows bulk-inserted in one transaction share an identical created_at
-- (CURRENT_TIMESTAMP is transaction start time), so "newest prior by
-- created_at" was nondeterministic for same-hash rows and duplicate-count
-- increments could scatter across several rows. FindPriorIngestionLog* now
-- break ties by id DESC; extend the lookup index to match that sort order.
DROP INDEX IF EXISTS idx_ingestion_logs_entity_hash_created_at;
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_entity_hash_created_at
ON ingestion_logs(entity_hash, created_at DESC, id DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_ingestion_logs_entity_hash_created_at;
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_entity_hash_created_at
ON ingestion_logs(entity_hash, created_at DESC);
-- +goose StatementEnd
