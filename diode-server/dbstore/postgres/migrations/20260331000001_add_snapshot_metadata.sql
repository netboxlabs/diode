-- +goose Up
-- Add metadata column to graph_node_snapshots so each snapshot preserves
-- the metadata (including run_id) that was current at the time of ingestion.
-- This prevents run_id loss when the same entity is upserted across multiple
-- discovery runs, since the node-level metadata only stores the latest run_id.

ALTER TABLE graph_node_snapshots
    ADD COLUMN metadata jsonb NOT NULL DEFAULT '{}';

-- GIN index allows efficient @> containment queries (e.g. filter by run_id)
CREATE INDEX IF NOT EXISTS idx_graph_node_snapshots_metadata_gin
    ON graph_node_snapshots USING GIN (metadata);

-- +goose Down
DROP INDEX IF EXISTS idx_graph_node_snapshots_metadata_gin;

ALTER TABLE graph_node_snapshots
    DROP COLUMN IF EXISTS metadata;
