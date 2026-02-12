package graph

import (
	"context"
)

// NodeWriter persists node mutations.
type NodeWriter interface {
	UpsertNode(ctx context.Context, arg UpsertNodeParams) (Node, error)
	UpdateNodeData(ctx context.Context, arg UpdateNodeDataParams) (Node, error)
}

// NodeReader queries nodes.
type NodeReader interface {
	FindNode(ctx context.Context, arg FindNodeParams) (Node, error)
	FindNodesByFieldMatch(ctx context.Context, arg FindNodesByFieldMatchParams) ([]Node, error)
	GetNodesByType(ctx context.Context, arg GetNodesByTypeParams) ([]Node, error)
	FindNodeByMetadata(ctx context.Context, arg FindNodeByMetadataParams) (Node, error)
	FindNodeByContentHash(ctx context.Context, arg FindNodeByContentHashParams) (Node, error)
	ListNodes(ctx context.Context, arg ListNodesParams) ([]NodeWithLatestSnapshot, error)
}

// EdgeWriter persists edge mutations.
type EdgeWriter interface {
	UpsertEdge(ctx context.Context, arg UpsertEdgeParams) error
}

// SnapshotWriter persists snapshot mutations.
type SnapshotWriter interface {
	InsertSnapshot(ctx context.Context, arg InsertSnapshotParams) (Snapshot, error)
	CleanupOldSnapshots(ctx context.Context, arg CleanupOldSnapshotsParams) error
}

// SnapshotReader queries snapshots.
type SnapshotReader interface {
	GetLatestSnapshot(ctx context.Context, nodeID int64) (Snapshot, error)
	GetNodeWithLatestSnapshot(ctx context.Context, arg GetNodeWithLatestSnapshotParams) (NodeWithLatestSnapshot, error)
	GetSnapshotsByNode(ctx context.Context, arg GetSnapshotsByNodeParams) ([]Snapshot, error)
}

// Repository composes all graph database operations.
// The adapter in dbstore/postgres/graph_repository.go translates between
// these domain types and SQLC-generated types.
type Repository interface {
	NodeWriter
	NodeReader
	EdgeWriter
	SnapshotWriter
	SnapshotReader
}
