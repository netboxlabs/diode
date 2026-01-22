-- +goose Up

-- Create graph_nodes table for storing entities as nodes
CREATE TABLE IF NOT EXISTS graph_nodes (
    id BIGSERIAL PRIMARY KEY,
    external_id TEXT NOT NULL,
    node_type TEXT NOT NULL,
    data JSONB NOT NULL,
    duplicate_count INTEGER DEFAULT 1 NOT NULL,
    matching_schema_version INTEGER DEFAULT 1 NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    CONSTRAINT unique_entity UNIQUE (node_type, external_id)
);

-- Create graph_node_snapshots table for storing historical entity data
CREATE TABLE IF NOT EXISTS graph_node_snapshots (
    id BIGSERIAL PRIMARY KEY,
    node_id BIGINT NOT NULL REFERENCES graph_nodes(id) ON DELETE CASCADE,
    snapshot_data JSONB NOT NULL,
    sequence_number INTEGER NOT NULL CHECK (sequence_number > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    CONSTRAINT unique_node_sequence UNIQUE (node_id, sequence_number)
);

-- Create graph_edges table for storing relationships between nodes
CREATE TABLE IF NOT EXISTS graph_edges (
    id BIGSERIAL PRIMARY KEY,
    source_node_id BIGINT NOT NULL REFERENCES graph_nodes(id) ON DELETE CASCADE,
    target_node_id BIGINT NOT NULL REFERENCES graph_nodes(id) ON DELETE CASCADE,
    edge_type TEXT NOT NULL,
    edge_subtype TEXT,
    properties JSONB DEFAULT '{}',
    confidence_score REAL DEFAULT 1.0 NOT NULL CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    match_reason TEXT,
    matching_fields JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    CONSTRAINT unique_edge UNIQUE (source_node_id, target_node_id, edge_type)
);

-- Essential indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_graph_nodes_type ON graph_nodes (node_type);
CREATE INDEX IF NOT EXISTS idx_graph_nodes_duplicate_count ON graph_nodes (duplicate_count);
CREATE INDEX IF NOT EXISTS idx_graph_nodes_data_gin ON graph_nodes USING GIN (data);
CREATE INDEX IF NOT EXISTS idx_graph_nodes_updated_at ON graph_nodes (updated_at);
CREATE INDEX IF NOT EXISTS idx_graph_nodes_schema_version ON graph_nodes (matching_schema_version);

-- Standard indexes for commonly searched fields (fuzzy matching now handled in application layer)
CREATE INDEX IF NOT EXISTS idx_graph_nodes_data_name ON graph_nodes ((data->>'name'));
CREATE INDEX IF NOT EXISTS idx_graph_nodes_data_serial ON graph_nodes ((data->>'serial'));

-- Indexes for graph_node_snapshots
CREATE INDEX IF NOT EXISTS idx_graph_node_snapshots_node_id ON graph_node_snapshots (node_id);
CREATE INDEX IF NOT EXISTS idx_graph_node_snapshots_sequence ON graph_node_snapshots (node_id, sequence_number DESC);
CREATE INDEX IF NOT EXISTS idx_graph_node_snapshots_data_gin ON graph_node_snapshots USING GIN (snapshot_data);

CREATE INDEX IF NOT EXISTS idx_graph_edges_source ON graph_edges (source_node_id);
CREATE INDEX IF NOT EXISTS idx_graph_edges_target ON graph_edges (target_node_id);
CREATE INDEX IF NOT EXISTS idx_graph_edges_type ON graph_edges (edge_type);
CREATE INDEX IF NOT EXISTS idx_graph_edges_subtype ON graph_edges (edge_subtype);
CREATE INDEX IF NOT EXISTS idx_graph_edges_confidence ON graph_edges (confidence_score);
CREATE INDEX IF NOT EXISTS idx_graph_edges_properties_gin ON graph_edges USING GIN (properties);
CREATE INDEX IF NOT EXISTS idx_graph_edges_matching_fields_gin ON graph_edges USING GIN (matching_fields);

-- Trigger to automatically update updated_at on graph_nodes (idempotent)
DROP TRIGGER IF EXISTS update_graph_nodes_updated_at ON graph_nodes;
CREATE TRIGGER update_graph_nodes_updated_at
    BEFORE UPDATE ON graph_nodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON COLUMN graph_nodes.matching_schema_version IS 'Version of the matching configuration used to extract matching attributes';
COMMENT ON COLUMN graph_node_snapshots.snapshot_data IS 'Complete entity protojson snapshot for historical tracking';
COMMENT ON COLUMN graph_node_snapshots.sequence_number IS 'Sequential number for ordering snapshots, higher is newer';
COMMENT ON COLUMN graph_edges.confidence_score IS 'Confidence score (0.0-1.0) for the relationship match';
COMMENT ON COLUMN graph_edges.match_reason IS 'Human-readable explanation of why entities were matched';
COMMENT ON COLUMN graph_edges.matching_fields IS 'JSON array of field names that contributed to the match';
COMMENT ON COLUMN graph_edges.edge_subtype IS 'Subtype of edge based on confidence: high_confidence, medium_confidence, low_confidence, possible_duplicate';

-- +goose Down

-- Remove comments
COMMENT ON COLUMN graph_nodes.matching_schema_version IS NULL;
COMMENT ON COLUMN graph_node_snapshots.snapshot_data IS NULL;
COMMENT ON COLUMN graph_node_snapshots.sequence_number IS NULL;
COMMENT ON COLUMN graph_edges.confidence_score IS NULL;
COMMENT ON COLUMN graph_edges.match_reason IS NULL;
COMMENT ON COLUMN graph_edges.matching_fields IS NULL;
COMMENT ON COLUMN graph_edges.edge_subtype IS NULL;

-- Drop trigger only (function is shared across migrations)
DROP TRIGGER IF EXISTS update_graph_nodes_updated_at ON graph_nodes;

-- Drop indexes
DROP INDEX IF EXISTS idx_graph_edges_matching_fields_gin;
DROP INDEX IF EXISTS idx_graph_edges_properties_gin;
DROP INDEX IF EXISTS idx_graph_edges_confidence;
DROP INDEX IF EXISTS idx_graph_edges_subtype;
DROP INDEX IF EXISTS idx_graph_edges_type;
DROP INDEX IF EXISTS idx_graph_edges_target;
DROP INDEX IF EXISTS idx_graph_edges_source;

DROP INDEX IF EXISTS idx_graph_node_snapshots_data_gin;
DROP INDEX IF EXISTS idx_graph_node_snapshots_sequence;
DROP INDEX IF EXISTS idx_graph_node_snapshots_node_id;

DROP INDEX IF EXISTS idx_graph_nodes_data_serial;
DROP INDEX IF EXISTS idx_graph_nodes_data_name;
DROP INDEX IF EXISTS idx_graph_nodes_schema_version;
DROP INDEX IF EXISTS idx_graph_nodes_updated_at;
DROP INDEX IF EXISTS idx_graph_nodes_data_gin;
DROP INDEX IF EXISTS idx_graph_nodes_duplicate_count;
DROP INDEX IF EXISTS idx_graph_nodes_type;

-- Drop tables
DROP TABLE IF EXISTS graph_edges;
DROP TABLE IF EXISTS graph_node_snapshots;
DROP TABLE IF EXISTS graph_nodes;