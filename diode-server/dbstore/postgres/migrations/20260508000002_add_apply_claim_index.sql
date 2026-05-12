-- +goose Up

-- Partial index for APPLYING state (8) to support ResetApplyingIngestionLogs on
-- startup. The claim query for AutoApplyProcessor (QUEUED -> APPLYING) reads
-- from state=1 and uses the existing idx_ingestion_logs_state_id composite index.
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_applying ON ingestion_logs (id) WHERE state = 8;

-- +goose Down

DROP INDEX IF EXISTS idx_ingestion_logs_applying;
