package reconciler

import (
	"context"
	"encoding/json"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
)

// GraphNode is an alias to the SQLC-generated GraphNode type
type GraphNode = postgres.GraphNode

// GraphEdge represents an edge between two nodes
type GraphEdge struct {
	ID           int64           `json:"id"`
	SourceNodeID int64           `json:"source_node_id"`
	TargetNodeID int64           `json:"target_node_id"`
	EdgeType     string          `json:"edge_type"`
	Properties   json.RawMessage `json:"properties"`
}

// GraphRepository defines the interface for graph database operations.
// This interface wraps the SQLC-generated Queries to allow for mocking in tests.
// Used by EntityMatcher and GraphBuilder.
type GraphRepository interface {
	// Node operations
	UpsertGraphNode(ctx context.Context, arg postgres.UpsertGraphNodeParams) (postgres.GraphNode, error)
	UpdateGraphNodeData(ctx context.Context, arg postgres.UpdateGraphNodeDataParams) (postgres.GraphNode, error)
	FindGraphNode(ctx context.Context, arg postgres.FindGraphNodeParams) (postgres.GraphNode, error)

	// Edge operations
	UpsertGraphEdge(ctx context.Context, arg postgres.UpsertGraphEdgeParams) error

	// Entity matching queries
	FindNodesByFieldMatch(ctx context.Context, arg postgres.FindNodesByFieldMatchParams) ([]postgres.GraphNode, error)
	FindNodesByMultiFieldMatch(ctx context.Context, arg postgres.FindNodesByMultiFieldMatchParams) ([]postgres.GraphNode, error)
	GetGraphNodesByType(ctx context.Context, arg postgres.GetGraphNodesByTypeParams) ([]postgres.GraphNode, error)

	// Snapshot management
	InsertSnapshot(ctx context.Context, arg postgres.InsertSnapshotParams) (postgres.GraphNodeSnapshot, error)
	GetLatestSnapshot(ctx context.Context, nodeID int64) (postgres.GraphNodeSnapshot, error)
	CleanupOldSnapshots(ctx context.Context, arg postgres.CleanupOldSnapshotsParams) error
	GetNodeWithLatestSnapshot(ctx context.Context, arg postgres.GetNodeWithLatestSnapshotParams) (postgres.GetNodeWithLatestSnapshotRow, error)
	GetSnapshotsByNode(ctx context.Context, arg postgres.GetSnapshotsByNodeParams) ([]postgres.GraphNodeSnapshot, error)

	// Schema update queries
	FindNodesNeedingSchemaUpdate(ctx context.Context, arg postgres.FindNodesNeedingSchemaUpdateParams) ([]postgres.GraphNode, error)
}
