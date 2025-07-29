-- +goose Up

ALTER TABLE ingestion_logs
    ADD COLUMN entity_hash VARCHAR(64),
    ADD COLUMN last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN duplicate_count INTEGER DEFAULT 0 NOT NULL;

CREATE INDEX idx_ingestion_logs_entity_hash ON ingestion_logs(entity_hash);

-- +goose Down


DROP INDEX IF EXISTS idx_ingestion_logs_entity_hash;

ALTER TABLE ingestion_logs
    DROP COLUMN IF EXISTS entity_hash,
    DROP COLUMN IF EXISTS last_seen,
    DROP COLUMN IF EXISTS duplicate_count;