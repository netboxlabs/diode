-- +goose Up

-- Add metadata column to graph_nodes for storing correlation IDs and source tracking
ALTER TABLE graph_nodes ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}' NOT NULL;

-- GIN index for efficient @> containment queries on metadata
CREATE INDEX IF NOT EXISTS idx_graph_nodes_metadata_gin ON graph_nodes USING GIN (metadata);

-- Add comment for documentation
COMMENT ON COLUMN graph_nodes.metadata IS 'Entity metadata including correlation IDs';

-- +goose Down

-- Remove comment
COMMENT ON COLUMN graph_nodes.metadata IS NULL;

-- Drop index
DROP INDEX IF EXISTS idx_graph_nodes_metadata_gin;

-- Remove column
ALTER TABLE graph_nodes DROP COLUMN IF EXISTS metadata;
