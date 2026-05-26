-- +goose Up

-- Composite descending index to support the deviation-list LIST shape:
-- "ORDER BY updated_at DESC, id DESC LIMIT N". Without this index Postgres
-- falls back to Parallel Seq Scan + top-N heapsort over the full
-- ingestion_logs table because updated_at carries many duplicate values
-- from bulk operations, so the existing single-column
-- idx_ingestion_logs_updated_at cannot satisfy the multi-column sort
-- via Incremental Sort. With the composite, the planner does an
-- Index Scan over (updated_at DESC, id DESC) and stops at LIMIT N,
-- bringing top-N list queries from seconds to sub-millisecond on
-- multi-million-row tables.
--
-- Complements idx_ingestion_logs_state_updated_at from migration
-- 20260525000002, which helps state-filtered COUNTs but does not satisfy
-- the DESC sort shape: state = ANY(...) requires merging multiple
-- ascending index streams, which the planner estimates as costlier than
-- a seq scan + sort. The composite below lets the planner satisfy the
-- DESC sort directly and apply state as a cheap post-scan filter.

CREATE INDEX IF NOT EXISTS idx_ingestion_logs_updated_at_id_desc
    ON ingestion_logs (updated_at DESC, id DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_ingestion_logs_updated_at_id_desc;
