package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/graph"
)

// Compile-time check that GraphRepository implements graph.Repository.
var _ graph.Repository = (*GraphRepository)(nil)

// GraphRepository is the postgres adapter for graph.Repository.
// It translates between domain types and SQLC-generated types.
type GraphRepository struct {
	queries *postgres.Queries
}

// NewGraphRepository creates a new GraphRepository.
func NewGraphRepository(pool *pgxpool.Pool) *GraphRepository {
	return &GraphRepository{
		queries: postgres.New(pool),
	}
}

func toNode(n postgres.GraphNode) graph.Node {
	var lastSeen *time.Time
	if n.LastSeenTs.Valid {
		lastSeen = &n.LastSeenTs.Time
	}
	var contentHash *string
	if n.ContentHash.Valid {
		contentHash = &n.ContentHash.String
	}
	return graph.Node{
		ID:             n.ID,
		ExternalID:     n.ExternalID,
		NodeType:       n.NodeType,
		Data:           json.RawMessage(n.Data),
		DuplicateCount: n.DuplicateCount,
		CreatedAt:      n.CreatedAt.Time,
		UpdatedAt:      n.UpdatedAt.Time,
		LastSeenTs:     lastSeen,
		Metadata:       json.RawMessage(n.Metadata),
		ContentHash:    contentHash,
	}
}

func toNodes(nodes []postgres.GraphNode) []graph.Node {
	result := make([]graph.Node, len(nodes))
	for i, n := range nodes {
		result[i] = toNode(n)
	}
	return result
}

func toSnapshot(s postgres.GraphNodeSnapshot) graph.Snapshot {
	return graph.Snapshot{
		ID:             s.ID,
		NodeID:         s.NodeID,
		SnapshotData:   json.RawMessage(s.SnapshotData),
		SequenceNumber: s.SequenceNumber,
		CreatedAt:      s.CreatedAt.Time,
	}
}

func toSnapshots(snapshots []postgres.GraphNodeSnapshot) []graph.Snapshot {
	result := make([]graph.Snapshot, len(snapshots))
	for i, s := range snapshots {
		result[i] = toSnapshot(s)
	}
	return result
}

func toOptionalPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func wrapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return graph.ErrNotFound
	}
	return err
}

// UpsertNode implements graph.Repository.
func (r *GraphRepository) UpsertNode(ctx context.Context, arg graph.UpsertNodeParams) (graph.Node, error) {
	result, err := r.queries.UpsertGraphNode(ctx, postgres.UpsertGraphNodeParams{
		ExternalID:  arg.ExternalID,
		NodeType:    arg.NodeType,
		Data:        []byte(arg.Data),
		Metadata:    []byte(arg.Metadata),
		ContentHash: toOptionalPgText(arg.ContentHash),
	})
	if err != nil {
		return graph.Node{}, err
	}
	return toNode(result), nil
}

// UpdateNodeData implements graph.Repository.
func (r *GraphRepository) UpdateNodeData(ctx context.Context, arg graph.UpdateNodeDataParams) (graph.Node, error) {
	result, err := r.queries.UpdateGraphNodeData(ctx, postgres.UpdateGraphNodeDataParams{
		NodeType:    arg.NodeType,
		ExternalID:  arg.ExternalID,
		Data:        []byte(arg.Data),
		Metadata:    []byte(arg.Metadata),
		ContentHash: toOptionalPgText(arg.ContentHash),
	})
	if err != nil {
		return graph.Node{}, wrapNotFound(err)
	}
	return toNode(result), nil
}

// FindNode implements graph.Repository.
func (r *GraphRepository) FindNode(ctx context.Context, arg graph.FindNodeParams) (graph.Node, error) {
	result, err := r.queries.FindGraphNode(ctx, postgres.FindGraphNodeParams{
		NodeType:   arg.NodeType,
		ExternalID: arg.ExternalID,
	})
	if err != nil {
		return graph.Node{}, wrapNotFound(err)
	}
	return toNode(result), nil
}

// UpsertEdge implements graph.Repository.
func (r *GraphRepository) UpsertEdge(ctx context.Context, arg graph.UpsertEdgeParams) error {
	return r.queries.UpsertGraphEdge(ctx, postgres.UpsertGraphEdgeParams{
		SourceNodeID: arg.SourceNodeID,
		TargetNodeID: arg.TargetNodeID,
		EdgeType:     arg.EdgeType,
		Properties:   []byte(arg.Properties),
	})
}

// FindNodesByFieldMatch implements graph.Repository.
func (r *GraphRepository) FindNodesByFieldMatch(ctx context.Context, arg graph.FindNodesByFieldMatchParams) ([]graph.Node, error) {
	result, err := r.queries.FindNodesByFieldMatch(ctx, postgres.FindNodesByFieldMatchParams{
		NodeType:    arg.NodeType,
		JsonField:   pgtype.Text{String: arg.JSONField, Valid: arg.JSONField != ""},
		FieldValue:  pgtype.Text{String: arg.FieldValue, Valid: arg.FieldValue != ""},
		NestedPath:  arg.NestedPath,
		NestedValue: pgtype.Text{String: arg.NestedValue, Valid: arg.NestedValue != ""},
		Offset:      arg.Offset,
		Limit:       arg.Limit,
	})
	if err != nil {
		return nil, err
	}
	return toNodes(result), nil
}

// GetNodesByType implements graph.Repository.
func (r *GraphRepository) GetNodesByType(ctx context.Context, arg graph.GetNodesByTypeParams) ([]graph.Node, error) {
	result, err := r.queries.GetGraphNodesByType(ctx, postgres.GetGraphNodesByTypeParams{
		NodeType: arg.NodeType,
		Offset:   arg.Offset,
		Limit:    arg.Limit,
	})
	if err != nil {
		return nil, err
	}
	return toNodes(result), nil
}

// FindNodeByMetadata implements graph.Repository.
func (r *GraphRepository) FindNodeByMetadata(ctx context.Context, arg graph.FindNodeByMetadataParams) (graph.Node, error) {
	result, err := r.queries.FindNodeByMetadata(ctx, postgres.FindNodeByMetadataParams{
		NodeType:       arg.NodeType,
		MetadataFilter: []byte(arg.MetadataFilter),
	})
	if err != nil {
		return graph.Node{}, wrapNotFound(err)
	}
	return toNode(result), nil
}

// FindNodeByContentHash implements graph.Repository.
func (r *GraphRepository) FindNodeByContentHash(ctx context.Context, arg graph.FindNodeByContentHashParams) (graph.Node, error) {
	result, err := r.queries.FindNodeByContentHash(ctx, postgres.FindNodeByContentHashParams{
		NodeType:    arg.NodeType,
		ContentHash: pgtype.Text{String: arg.ContentHash, Valid: arg.ContentHash != ""},
	})
	if err != nil {
		return graph.Node{}, wrapNotFound(err)
	}
	return toNode(result), nil
}

// InsertSnapshot implements graph.Repository.
func (r *GraphRepository) InsertSnapshot(ctx context.Context, arg graph.InsertSnapshotParams) (graph.Snapshot, error) {
	result, err := r.queries.InsertSnapshot(ctx, postgres.InsertSnapshotParams{
		NodeID:       arg.NodeID,
		SnapshotData: []byte(arg.SnapshotData),
	})
	if err != nil {
		return graph.Snapshot{}, err
	}
	return toSnapshot(result), nil
}

// GetLatestSnapshot implements graph.Repository.
func (r *GraphRepository) GetLatestSnapshot(ctx context.Context, nodeID int64) (graph.Snapshot, error) {
	result, err := r.queries.GetLatestSnapshot(ctx, nodeID)
	if err != nil {
		return graph.Snapshot{}, wrapNotFound(err)
	}
	return toSnapshot(result), nil
}

// CleanupOldSnapshots implements graph.Repository.
func (r *GraphRepository) CleanupOldSnapshots(ctx context.Context, arg graph.CleanupOldSnapshotsParams) error {
	return r.queries.CleanupOldSnapshots(ctx, postgres.CleanupOldSnapshotsParams{
		NodeID: arg.NodeID,
		Limit:  arg.Limit,
	})
}

// nodeWithSnapshotFields is a common shape shared by GetNodeWithLatestSnapshotRow and ListNodesRow.
type nodeWithSnapshotFields struct {
	ID                int64
	ExternalID        string
	NodeType          string
	MatchingData      []byte
	DuplicateCount    int32
	LastSeenTs        pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	Metadata          []byte
	SnapshotData      []byte
	SequenceNumber    int32
	SnapshotCreatedAt pgtype.Timestamptz
}

func toNodeWithLatestSnapshot(r nodeWithSnapshotFields) graph.NodeWithLatestSnapshot {
	var lastSeen *time.Time
	if r.LastSeenTs.Valid {
		lastSeen = &r.LastSeenTs.Time
	}
	var snapshotCreatedAt *time.Time
	if r.SnapshotCreatedAt.Valid {
		snapshotCreatedAt = &r.SnapshotCreatedAt.Time
	}
	return graph.NodeWithLatestSnapshot{
		ID:                r.ID,
		ExternalID:        r.ExternalID,
		NodeType:          r.NodeType,
		MatchingData:      json.RawMessage(r.MatchingData),
		DuplicateCount:    r.DuplicateCount,
		LastSeenTs:        lastSeen,
		CreatedAt:         r.CreatedAt.Time,
		UpdatedAt:         r.UpdatedAt.Time,
		Metadata:          json.RawMessage(r.Metadata),
		SnapshotData:      json.RawMessage(r.SnapshotData),
		SequenceNumber:    r.SequenceNumber,
		SnapshotCreatedAt: snapshotCreatedAt,
	}
}

// GetNodeWithLatestSnapshot implements graph.Repository.
func (r *GraphRepository) GetNodeWithLatestSnapshot(ctx context.Context, arg graph.GetNodeWithLatestSnapshotParams) (graph.NodeWithLatestSnapshot, error) {
	result, err := r.queries.GetNodeWithLatestSnapshot(ctx, postgres.GetNodeWithLatestSnapshotParams{
		NodeType:   arg.NodeType,
		ExternalID: arg.ExternalID,
	})
	if err != nil {
		return graph.NodeWithLatestSnapshot{}, wrapNotFound(err)
	}
	return toNodeWithLatestSnapshot(nodeWithSnapshotFields{
		ID:                result.ID,
		ExternalID:        result.ExternalID,
		NodeType:          result.NodeType,
		MatchingData:      result.MatchingData,
		DuplicateCount:    result.DuplicateCount,
		LastSeenTs:        result.LastSeenTs,
		CreatedAt:         result.CreatedAt,
		UpdatedAt:         result.UpdatedAt,
		Metadata:          result.Metadata,
		SnapshotData:      result.SnapshotData,
		SequenceNumber:    result.SequenceNumber,
		SnapshotCreatedAt: result.SnapshotCreatedAt,
	}), nil
}

// ListNodes implements graph.Repository.
func (r *GraphRepository) ListNodes(ctx context.Context, arg graph.ListNodesParams) ([]graph.NodeWithLatestSnapshot, error) {
	rows, err := r.queries.ListNodes(ctx, postgres.ListNodesParams{
		NodeTypes:      arg.NodeTypes,
		MetadataFilter: []byte(arg.MetadataFilter),
		Limit:          arg.Limit,
		Offset:         arg.Offset,
	})
	if err != nil {
		return nil, err
	}
	result := make([]graph.NodeWithLatestSnapshot, len(rows))
	for i, row := range rows {
		result[i] = toNodeWithLatestSnapshot(nodeWithSnapshotFields{
			ID:                row.ID,
			ExternalID:        row.ExternalID,
			NodeType:          row.NodeType,
			MatchingData:      row.MatchingData,
			DuplicateCount:    row.DuplicateCount,
			LastSeenTs:        row.LastSeenTs,
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
			Metadata:          row.Metadata,
			SnapshotData:      row.SnapshotData,
			SequenceNumber:    row.SequenceNumber,
			SnapshotCreatedAt: row.SnapshotCreatedAt,
		})
	}
	return result, nil
}

// GetSnapshotsByNode implements graph.Repository.
func (r *GraphRepository) GetSnapshotsByNode(ctx context.Context, arg graph.GetSnapshotsByNodeParams) ([]graph.Snapshot, error) {
	result, err := r.queries.GetSnapshotsByNode(ctx, postgres.GetSnapshotsByNodeParams{
		NodeID: arg.NodeID,
		Offset: arg.Offset,
		Limit:  arg.Limit,
	})
	if err != nil {
		return nil, err
	}
	return toSnapshots(result), nil
}
