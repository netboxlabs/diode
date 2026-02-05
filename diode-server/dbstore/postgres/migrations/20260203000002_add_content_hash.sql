-- +goose Up

-- Add content_hash column to graph_nodes for last-resort entity matching fallback
-- Content hash is computed from entity attributes and used when entity matching config is missing
ALTER TABLE graph_nodes ADD COLUMN IF NOT EXISTS content_hash TEXT;

-- Index on (node_type, content_hash) for efficient fallback lookups
CREATE INDEX IF NOT EXISTS idx_graph_nodes_content_hash ON graph_nodes (node_type, content_hash) WHERE content_hash IS NOT NULL;

-- Comment for documentation
COMMENT ON COLUMN graph_nodes.content_hash IS 'Entity content hash for fallback matching when entity matcher config is missing';

-- +goose Down

-- Remove comment
COMMENT ON COLUMN graph_nodes.content_hash IS NULL;

-- Drop index
DROP INDEX IF EXISTS idx_graph_nodes_content_hash;

-- Remove column
ALTER TABLE graph_nodes DROP COLUMN IF EXISTS content_hash;
