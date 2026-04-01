-- +goose Up

-- Add data_hash column to graph_node_snapshots for deduplication.
-- When entity data hasn't changed between runs, we reuse the existing snapshot
-- and only insert a new metadata row.
ALTER TABLE graph_node_snapshots ADD COLUMN IF NOT EXISTS data_hash TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_graph_node_snapshots_data_hash ON graph_node_snapshots (node_id, data_hash);

-- Create graph_node_snapshot_metadata table for tracking per-ingestion metadata.
-- Each row records the metadata (e.g., run_id) from a single ingestion event,
-- pointing to the deduplicated snapshot that represents the entity state at that time.
CREATE TABLE IF NOT EXISTS graph_node_snapshot_metadata (
    id BIGSERIAL PRIMARY KEY,
    snapshot_id BIGINT NOT NULL REFERENCES graph_node_snapshots(id) ON DELETE CASCADE,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- GIN index for efficient @> containment queries on snapshot metadata
CREATE INDEX IF NOT EXISTS idx_graph_node_snapshot_metadata_gin ON graph_node_snapshot_metadata USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_graph_node_snapshot_metadata_snapshot_id ON graph_node_snapshot_metadata (snapshot_id);

-- +goose Down

DROP INDEX IF EXISTS idx_graph_node_snapshot_metadata_snapshot_id;
DROP INDEX IF EXISTS idx_graph_node_snapshot_metadata_gin;
DROP TABLE IF EXISTS graph_node_snapshot_metadata;
DROP INDEX IF EXISTS idx_graph_node_snapshots_data_hash;
ALTER TABLE graph_node_snapshots DROP COLUMN IF EXISTS data_hash;
