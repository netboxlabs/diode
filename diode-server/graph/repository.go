package graph

import (
	"context"
)

// Repository defines the interface for graph database operations.
// All types are domain types defined in types.go.
// The adapter in dbstore/postgres/graph_repository.go translates between
// these domain types and SQLC-generated types.
type Repository interface {
	// Node operations
	UpsertNode(ctx context.Context, arg UpsertNodeParams) (Node, error)
	UpdateNodeData(ctx context.Context, arg UpdateNodeDataParams) (Node, error)
	FindNode(ctx context.Context, arg FindNodeParams) (Node, error)

	// Edge operations
	UpsertEdge(ctx context.Context, arg UpsertEdgeParams) error

	// Entity matching queries
	FindNodesByFieldMatch(ctx context.Context, arg FindNodesByFieldMatchParams) ([]Node, error)
	GetNodesByType(ctx context.Context, arg GetNodesByTypeParams) ([]Node, error)
	FindNodeByMetadata(ctx context.Context, arg FindNodeByMetadataParams) (Node, error)
	FindNodeByContentHash(ctx context.Context, arg FindNodeByContentHashParams) (Node, error)

	// Snapshot management
	InsertSnapshot(ctx context.Context, arg InsertSnapshotParams) (Snapshot, error)
	GetLatestSnapshot(ctx context.Context, nodeID int64) (Snapshot, error)
	CleanupOldSnapshots(ctx context.Context, arg CleanupOldSnapshotsParams) error
	GetNodeWithLatestSnapshot(ctx context.Context, arg GetNodeWithLatestSnapshotParams) (NodeWithLatestSnapshot, error)
	GetSnapshotsByNode(ctx context.Context, arg GetSnapshotsByNodeParams) ([]Snapshot, error)

	// Schema update queries
	FindNodesNeedingSchemaUpdate(ctx context.Context, arg FindNodesNeedingSchemaUpdateParams) ([]Node, error)
}
