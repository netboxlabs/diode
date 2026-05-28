-- +goose Up

-- Retry accounting for automatic retry of failed applies.
--   retry_count   - failed apply attempts for this row.
--   next_retry_at - earliest re-claim time; NULL means eligible now.
ALTER TABLE ingestion_logs
    ADD COLUMN IF NOT EXISTS retry_count   INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMP WITH TIME ZONE;

-- Partial index for the due-retry claim predicate (PENDING_RETRY by due time).
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_pending_retry
    ON ingestion_logs (next_retry_at) WHERE state = 9;

-- +goose Down

DROP INDEX IF EXISTS idx_ingestion_logs_pending_retry;
ALTER TABLE ingestion_logs
    DROP COLUMN IF EXISTS next_retry_at,
    DROP COLUMN IF EXISTS retry_count;
