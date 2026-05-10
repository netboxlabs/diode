-- +goose Up

-- Partial index for ClaimOpenIngestionLogs: only indexes rows in state=2 (OPEN),
-- the transient state between planning and applying.
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_claim_apply ON ingestion_logs (id) WHERE state = 2;

-- Partial index for APPLYING state (8) to support ResetApplyingIngestionLogs on startup.
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_applying ON ingestion_logs (id) WHERE state = 8;

-- +goose Down

DROP INDEX IF EXISTS idx_ingestion_logs_applying;
DROP INDEX IF EXISTS idx_ingestion_logs_claim_apply;
