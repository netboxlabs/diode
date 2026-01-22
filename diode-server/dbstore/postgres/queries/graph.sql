-- name: UpsertGraphNode :one
INSERT INTO graph_nodes (external_id, node_type, data, duplicate_count, matching_schema_version)
VALUES ($1, $2, $3, 1, $4)
ON CONFLICT (node_type, external_id)
DO UPDATE SET
    data = EXCLUDED.data,
    duplicate_count = graph_nodes.duplicate_count + 1,
    matching_schema_version = EXCLUDED.matching_schema_version,
    updated_at = NOW()
RETURNING id, external_id, node_type, data, duplicate_count, matching_schema_version;

-- name: UpdateGraphNodeData :one
UPDATE graph_nodes
SET data = $3, matching_schema_version = $4, updated_at = NOW()
WHERE node_type = $1 AND external_id = $2
RETURNING id, external_id, node_type, data, duplicate_count, matching_schema_version;

-- name: FindGraphNode :one
SELECT id, external_id, node_type, data, duplicate_count, matching_schema_version
FROM graph_nodes
WHERE node_type = $1 AND external_id = $2;

-- name: UpsertGraphEdge :exec
INSERT INTO graph_edges (source_node_id, target_node_id, edge_type, properties)
VALUES ($1, $2, $3, $4)
ON CONFLICT (source_node_id, target_node_id, edge_type)
DO UPDATE SET properties = EXCLUDED.properties;

-- name: UpsertGraphEdgeWithConfidence :exec
INSERT INTO graph_edges (source_node_id, target_node_id, edge_type, properties, confidence_score, match_reason, matching_fields, edge_subtype)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (source_node_id, target_node_id, edge_type)
DO UPDATE SET 
    properties = EXCLUDED.properties,
    confidence_score = EXCLUDED.confidence_score,
    match_reason = EXCLUDED.match_reason,
    matching_fields = EXCLUDED.matching_fields,
    edge_subtype = EXCLUDED.edge_subtype;

-- name: GetGraphNodesByType :many
SELECT id, external_id, node_type, data, duplicate_count, matching_schema_version, created_at, updated_at
FROM graph_nodes
WHERE node_type = $1
ORDER BY updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetConnectedNodes :many
SELECT
    target.id,
    target.external_id,
    target.node_type,
    target.data,
    target.duplicate_count,
    target.matching_schema_version,
    edges.edge_type,
    edges.properties as edge_properties
FROM graph_edges edges
JOIN graph_nodes target ON edges.target_node_id = target.id
WHERE edges.source_node_id = $1;

-- name: GetNodesByFrequency :many
SELECT id, external_id, node_type, data, duplicate_count, matching_schema_version, created_at, updated_at
FROM graph_nodes
WHERE duplicate_count >= $1
ORDER BY duplicate_count DESC, updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: SearchGraphNodes :many
SELECT id, external_id, node_type, data, duplicate_count, matching_schema_version, created_at, updated_at
FROM graph_nodes
WHERE (node_type = sqlc.narg('node_type') OR sqlc.narg('node_type') IS NULL)
  AND (
    external_id ILIKE '%' || sqlc.narg('search_term') || '%'
    OR data::text ILIKE '%' || sqlc.narg('search_term') || '%'
    OR sqlc.narg('search_term') IS NULL
  )
ORDER BY duplicate_count DESC, updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetGraphStats :one
SELECT 
    COUNT(*) as total_nodes,
    COUNT(DISTINCT node_type) as unique_node_types,
    SUM(duplicate_count) as total_observations,
    AVG(duplicate_count) as avg_observations_per_node
FROM graph_nodes;

-- name: GetGraphStatsByType :many
SELECT
    node_type,
    COUNT(*) as node_count,
    SUM(duplicate_count) as total_observations,
    AVG(duplicate_count) as avg_observations_per_node,
    MAX(duplicate_count) as max_observations,
    MIN(created_at) as first_seen,
    MAX(updated_at) as last_seen
FROM graph_nodes
GROUP BY node_type
ORDER BY node_count DESC;

-- Entity matching queries for confidence-based matching

-- name: FindNodesByFieldMatch :many
SELECT id, external_id, node_type, data, duplicate_count, matching_schema_version, created_at, updated_at
FROM graph_nodes
WHERE node_type = $1
  AND (
    -- Support matching by single JSON field (only if json_field is provided)
    (sqlc.narg('json_field')::text IS NOT NULL AND data->>sqlc.narg('json_field')::text = sqlc.narg('field_value')::text)
    -- Support matching by nested JSON path (only if nested_path is provided)
    OR (sqlc.narg('nested_path')::text[] IS NOT NULL AND data#>>sqlc.narg('nested_path')::text[] = sqlc.narg('nested_value')::text)
  )
ORDER BY duplicate_count DESC, updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: FindNodesByMultiFieldMatch :many
SELECT id, external_id, node_type, data, duplicate_count, matching_schema_version, created_at, updated_at
FROM graph_nodes
WHERE node_type = $1
  AND (
    -- Primary field match (required)
    data->>sqlc.arg('primary_field')::text = sqlc.arg('primary_value')::text
    -- Secondary field match (optional)
    AND (sqlc.narg('secondary_field')::text IS NULL
         OR data->>sqlc.narg('secondary_field')::text = sqlc.narg('secondary_value')::text)
    -- Tertiary field match (optional)
    AND (sqlc.narg('tertiary_field')::text IS NULL
         OR data->>sqlc.narg('tertiary_field')::text = sqlc.narg('tertiary_value')::text)
  )
ORDER BY duplicate_count DESC, updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: FindNodesByFieldPattern :many
-- Database-agnostic fuzzy matching using LIKE patterns and JSONB operations
SELECT id, external_id, node_type, data, duplicate_count, matching_schema_version, created_at, updated_at
FROM graph_nodes
WHERE node_type = $1
  AND (
    -- Exact match (highest priority)
    data->>sqlc.arg('field_name')::text = sqlc.arg('search_value')::text
    -- Prefix match (second priority)
    OR data->>sqlc.arg('field_name')::text ILIKE sqlc.arg('search_value')::text || '%'
    -- Contains match (third priority)
    OR data->>sqlc.arg('field_name')::text ILIKE '%' || sqlc.arg('search_value')::text || '%'
    -- Suffix match (fourth priority)
    OR data->>sqlc.arg('field_name')::text ILIKE '%' || sqlc.arg('search_value')::text
  )
ORDER BY
  CASE
    WHEN data->>sqlc.arg('field_name')::text = sqlc.arg('search_value')::text THEN 1
    WHEN data->>sqlc.arg('field_name')::text ILIKE sqlc.arg('search_value')::text || '%' THEN 2
    WHEN data->>sqlc.arg('field_name')::text ILIKE '%' || sqlc.arg('search_value')::text || '%' THEN 3
    ELSE 4
  END,
  duplicate_count DESC,
  updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: FindNodesByComplexMatch :many
SELECT id, external_id, node_type, data, duplicate_count, matching_schema_version, created_at, updated_at
FROM graph_nodes
WHERE node_type = $1
  AND (
    -- Support JSONB containment queries
    (sqlc.narg('contains_filter')::jsonb IS NULL OR data @> sqlc.narg('contains_filter')::jsonb)
    -- Support JSONB path existence
    AND (sqlc.narg('path_exists')::text IS NULL OR data ? sqlc.narg('path_exists')::text)
    -- Support regex matching on text fields
    AND (sqlc.narg('regex_field')::text IS NULL OR sqlc.narg('regex_pattern')::text IS NULL
         OR data->>sqlc.narg('regex_field')::text ~ sqlc.narg('regex_pattern')::text)
  )
ORDER BY duplicate_count DESC, updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- Snapshot management queries

-- name: InsertSnapshot :one
INSERT INTO graph_node_snapshots (node_id, snapshot_data, sequence_number)
VALUES ($1, $2, $3)
RETURNING id, node_id, snapshot_data, sequence_number, created_at;

-- name: GetLatestSnapshot :one
SELECT id, node_id, snapshot_data, sequence_number, created_at
FROM graph_node_snapshots
WHERE node_id = $1
ORDER BY sequence_number DESC
LIMIT 1;

-- name: GetSnapshotsByNode :many
SELECT id, node_id, snapshot_data, sequence_number, created_at
FROM graph_node_snapshots
WHERE node_id = $1
ORDER BY sequence_number DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetNextSequenceNumber :one
SELECT COALESCE(MAX(sequence_number), 0) + 1 as next_sequence
FROM graph_node_snapshots
WHERE node_id = $1;

-- name: CleanupOldSnapshots :exec
DELETE FROM graph_node_snapshots outer_snap
WHERE outer_snap.node_id = $1
  AND outer_snap.id NOT IN (
    SELECT s.id
    FROM graph_node_snapshots s
    WHERE s.node_id = $1
    ORDER BY s.sequence_number DESC
    LIMIT $2
  );

-- name: GetNodeWithLatestSnapshot :one
-- Note: snapshot fields use COALESCE defaults for NULL safety when no snapshot exists
-- sequence_number = 0 and empty snapshot_data indicate no snapshot
SELECT
    n.id, n.external_id, n.node_type, n.data as matching_data, n.duplicate_count, n.matching_schema_version, n.created_at, n.updated_at,
    COALESCE(s.snapshot_data, '{}'::jsonb) as snapshot_data,
    COALESCE(s.sequence_number, 0) as sequence_number,
    s.created_at as snapshot_created_at
FROM graph_nodes n
LEFT JOIN LATERAL (
    SELECT snapshot_data, sequence_number, created_at
    FROM graph_node_snapshots
    WHERE node_id = n.id
    ORDER BY sequence_number DESC
    LIMIT 1
) s ON true
WHERE n.node_type = $1 AND n.external_id = $2;

-- name: FindNodesNeedingSchemaUpdate :many
SELECT id, external_id, node_type, data, duplicate_count, matching_schema_version, created_at, updated_at
FROM graph_nodes
WHERE matching_schema_version < $1
ORDER BY updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');