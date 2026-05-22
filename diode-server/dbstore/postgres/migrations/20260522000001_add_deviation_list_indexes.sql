-- +goose Up

-- Two B-tree indexes on ingestion_logs to support fast ORDER BY on the
-- ingestion_ts and updated_at columns. Without these, queries listing
-- ingestion_logs ordered by either column fall back to a parallel
-- sequential scan + top-N heapsort over the full table; on a tenant
-- with several million rows that runs ~700ms-1.2s hot and times out
-- the 5s client budget under concurrent traffic. Plain (non-partial)
-- B-tree because list-style consumers may filter by any subset of
-- state/object_type/branch in addition to ordering by these columns.
--
-- B-tree indexes are bidirectional, so a single index per column covers
-- both ASC and DESC orderings without a second index.

CREATE INDEX IF NOT EXISTS idx_ingestion_logs_ingestion_ts ON ingestion_logs (ingestion_ts);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_updated_at  ON ingestion_logs (updated_at);

-- +goose Down

DROP INDEX IF EXISTS idx_ingestion_logs_ingestion_ts;
DROP INDEX IF EXISTS idx_ingestion_logs_updated_at;
