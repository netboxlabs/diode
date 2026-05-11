-- +goose Up

-- Replace single-column state index with composite (state, id) to support
-- efficient ORDER BY id LIMIT N scans in ClaimQueuedIngestionLogs.
DROP INDEX IF EXISTS idx_ingestion_logs_state;
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_state_id ON ingestion_logs (state, id);

-- +goose Down

DROP INDEX IF EXISTS idx_ingestion_logs_state_id;
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_state ON ingestion_logs (state);
