package graph

import (
	"encoding/json"
	"errors"
	"time"
)

// ErrNotFound is a domain sentinel error returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// Node represents a node in the entity graph.
type Node struct {
	ID             int64           `json:"id"`
	ExternalID     string          `json:"external_id"`
	NodeType       string          `json:"node_type"`
	Data           json.RawMessage `json:"data"`
	DuplicateCount int32           `json:"duplicate_count"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	LastSeenTs     *time.Time      `json:"last_seen_ts,omitempty"`
	Metadata       json.RawMessage `json:"metadata"`
	ContentHash    *string         `json:"content_hash,omitempty"`
}

// Snapshot represents a point-in-time snapshot of a graph node's data.
type Snapshot struct {
	ID             int64           `json:"id"`
	NodeID         int64           `json:"node_id"`
	SnapshotData   json.RawMessage `json:"snapshot_data"`
	SequenceNumber int32           `json:"sequence_number"`
	CreatedAt      time.Time       `json:"created_at"`
	DataHash       string          `json:"data_hash"`
}

// SnapshotMetadata represents ingestion metadata associated with a snapshot.
type SnapshotMetadata struct {
	ID         int64           `json:"id"`
	SnapshotID int64           `json:"snapshot_id"`
	Metadata   json.RawMessage `json:"metadata"`
	CreatedAt  time.Time       `json:"created_at"`
}

// Edge represents an edge between two nodes.
type Edge struct {
	ID           int64           `json:"id"`
	SourceNodeID int64           `json:"source_node_id"`
	TargetNodeID int64           `json:"target_node_id"`
	EdgeType     string          `json:"edge_type"`
	Properties   json.RawMessage `json:"properties"`
}

// NodeWithLatestSnapshot combines a graph node with its snapshot data.
// When filtered by metadata, returns the matching snapshot rather than the latest.
type NodeWithLatestSnapshot struct {
	ID                int64           `json:"id"`
	ExternalID        string          `json:"external_id"`
	NodeType          string          `json:"node_type"`
	MatchingData      json.RawMessage `json:"matching_data"`
	DuplicateCount    int32           `json:"duplicate_count"`
	LastSeenTs        *time.Time      `json:"last_seen_ts,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	Metadata          json.RawMessage `json:"metadata"`
	SnapshotData      json.RawMessage `json:"snapshot_data"`
	SequenceNumber    int32           `json:"sequence_number"`
	SnapshotCreatedAt *time.Time      `json:"snapshot_created_at,omitempty"`
	SnapshotMetadata  json.RawMessage `json:"snapshot_metadata"`
}

// CompleteNodeData represents a node with its complete entity data.
type CompleteNodeData struct {
	ID                     int64           `json:"id"`
	ExternalID             string          `json:"external_id"`
	NodeType               string          `json:"node_type"`
	MatchingData           json.RawMessage `json:"matching_data"`
	CompleteData           json.RawMessage `json:"complete_data"`
	DuplicateCount         int32           `json:"duplicate_count"`
	SnapshotSequenceNumber *int32          `json:"snapshot_sequence_number,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
	SnapshotCreatedAt      *time.Time      `json:"snapshot_created_at,omitempty"`
}

// ListNodesParams contains parameters for listing nodes with pagination and filtering.
type ListNodesParams struct {
	NodeTypes      []string
	MetadataFilter json.RawMessage
	Limit          int32
	Offset         int32
}

// ListNodesBySnapshotMetadataParams contains parameters for listing nodes filtered by snapshot metadata.
type ListNodesBySnapshotMetadataParams struct {
	MetadataFilter json.RawMessage
	NodeTypes      []string
	Limit          int32
	Offset         int32
}

// UpsertNodeParams contains parameters for upserting a graph node.
type UpsertNodeParams struct {
	ExternalID  string
	NodeType    string
	Data        json.RawMessage
	Metadata    json.RawMessage
	ContentHash *string
}

// UpdateNodeDataParams contains parameters for updating a graph node's data.
type UpdateNodeDataParams struct {
	NodeType    string
	ExternalID  string
	Data        json.RawMessage
	Metadata    json.RawMessage
	ContentHash *string
}

// FindNodeParams contains parameters for finding a graph node.
type FindNodeParams struct {
	NodeType   string
	ExternalID string
}

// UpsertEdgeParams contains parameters for upserting a graph edge.
type UpsertEdgeParams struct {
	SourceNodeID int64
	TargetNodeID int64
	EdgeType     string
	Properties   json.RawMessage
}

// FindNodesByFieldMatchParams contains parameters for finding nodes by field match.
type FindNodesByFieldMatchParams struct {
	NodeType    string
	JSONField   string
	FieldValue  string
	NestedPath  []string
	NestedValue string
	Offset      int32
	Limit       int32
}

// GetNodesByTypeParams contains parameters for getting graph nodes by type.
type GetNodesByTypeParams struct {
	NodeType string
	Offset   int32
	Limit    int32
}

// FindNodeByMetadataParams contains parameters for finding a node by metadata.
type FindNodeByMetadataParams struct {
	NodeType       string
	MetadataFilter json.RawMessage
}

// FindNodeByContentHashParams contains parameters for finding a node by content hash.
type FindNodeByContentHashParams struct {
	NodeType    string
	ContentHash string
}

// InsertSnapshotParams contains parameters for inserting a snapshot.
type InsertSnapshotParams struct {
	NodeID       int64
	SnapshotData json.RawMessage
	DataHash     string
}

// FindLatestSnapshotByHashParams contains parameters for finding a snapshot by data hash.
type FindLatestSnapshotByHashParams struct {
	NodeID   int64
	DataHash string
}

// InsertSnapshotMetadataParams contains parameters for inserting snapshot metadata.
type InsertSnapshotMetadataParams struct {
	SnapshotID int64
	Metadata   json.RawMessage
}

// CleanupOldSnapshotsParams contains parameters for cleaning up old snapshots by count.
type CleanupOldSnapshotsParams struct {
	NodeID int64
	Limit  int32
}

// CleanupExpiredSnapshotsParams contains parameters for cleaning up snapshots by age.
type CleanupExpiredSnapshotsParams struct {
	NodeID        int64
	RetentionDays int32
}

// GetNodeWithLatestSnapshotParams contains parameters for getting a node with its latest snapshot.
type GetNodeWithLatestSnapshotParams struct {
	NodeType   string
	ExternalID string
}

// GetSnapshotsByNodeParams contains parameters for getting snapshots by node.
type GetSnapshotsByNodeParams struct {
	NodeID int64
	Offset int32
	Limit  int32
}

// ptrIfNonEmpty returns a pointer to s if s is non-empty, otherwise nil.
func ptrIfNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
