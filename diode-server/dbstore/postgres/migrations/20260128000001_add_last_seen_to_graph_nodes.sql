-- +goose Up

-- Add last_seen_ts column to graph_nodes for tracking when entity was last observed
ALTER TABLE graph_nodes ADD COLUMN IF NOT EXISTS last_seen_ts TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL;

-- Create index for efficient time-based queries
CREATE INDEX IF NOT EXISTS idx_graph_nodes_last_seen_ts ON graph_nodes (last_seen_ts);

-- Add comment for documentation
COMMENT ON COLUMN graph_nodes.last_seen_ts IS 'Timestamp when this entity was last seen/ingested';

-- +goose Down

-- Remove comment
COMMENT ON COLUMN graph_nodes.last_seen_ts IS NULL;

-- Drop index
DROP INDEX IF EXISTS idx_graph_nodes_last_seen_ts;

-- Remove column
ALTER TABLE graph_nodes DROP COLUMN IF EXISTS last_seen_ts;
