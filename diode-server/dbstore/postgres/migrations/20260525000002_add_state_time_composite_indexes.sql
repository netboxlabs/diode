-- +goose Up

-- Composite indexes to support the common deviation-list COUNT shape:
-- "WHERE state = ANY(...) AND (ingestion_ts | updated_at) BETWEEN ? AND ?".
-- Without these, Postgres has to pick between the single-column state
-- index and the single-column time index and filter by the other axis,
-- which scans many irrelevant rows when the time window is selective.
-- The composite lets it seek to each matching state value and range-scan
-- the time column within.

CREATE INDEX IF NOT EXISTS idx_ingestion_logs_state_ingestion_ts ON ingestion_logs (state, ingestion_ts);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_state_updated_at  ON ingestion_logs (state, updated_at);

-- +goose Down

DROP INDEX IF EXISTS idx_ingestion_logs_state_ingestion_ts;
DROP INDEX IF EXISTS idx_ingestion_logs_state_updated_at;
