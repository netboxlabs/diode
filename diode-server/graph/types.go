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
	ID                    int64           `json:"id"`
	ExternalID            string          `json:"external_id"`
	NodeType              string          `json:"node_type"`
	Data                  json.RawMessage `json:"data"`
	DuplicateCount        int32           `json:"duplicate_count"`
	MatchingSchemaVersion int32           `json:"matching_schema_version"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	LastSeenTs            *time.Time      `json:"last_seen_ts,omitempty"`
	Metadata              json.RawMessage `json:"metadata"`
	ContentHash           *string         `json:"content_hash,omitempty"`
}

// Snapshot represents a point-in-time snapshot of a graph node's data.
type Snapshot struct {
	ID             int64           `json:"id"`
	NodeID         int64           `json:"node_id"`
	SnapshotData   json.RawMessage `json:"snapshot_data"`
	SequenceNumber int32           `json:"sequence_number"`
	CreatedAt      time.Time       `json:"created_at"`
}

// Edge represents an edge between two nodes.
type Edge struct {
	ID           int64           `json:"id"`
	SourceNodeID int64           `json:"source_node_id"`
	TargetNodeID int64           `json:"target_node_id"`
	EdgeType     string          `json:"edge_type"`
	Properties   json.RawMessage `json:"properties"`
}

// NodeWithLatestSnapshot combines a graph node with its latest snapshot data.
type NodeWithLatestSnapshot struct {
	ID                    int64           `json:"id"`
	ExternalID            string          `json:"external_id"`
	NodeType              string          `json:"node_type"`
	MatchingData          json.RawMessage `json:"matching_data"`
	DuplicateCount        int32           `json:"duplicate_count"`
	MatchingSchemaVersion int32           `json:"matching_schema_version"`
	LastSeenTs            *time.Time      `json:"last_seen_ts,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	Metadata              json.RawMessage `json:"metadata"`
	SnapshotData          json.RawMessage `json:"snapshot_data"`
	SequenceNumber        int32           `json:"sequence_number"`
	SnapshotCreatedAt     *time.Time      `json:"snapshot_created_at,omitempty"`
}

// UpsertNodeParams contains parameters for upserting a graph node.
type UpsertNodeParams struct {
	ExternalID            string
	NodeType              string
	Data                  json.RawMessage
	MatchingSchemaVersion int32
	Metadata              json.RawMessage
	ContentHash           *string
}

// UpdateNodeDataParams contains parameters for updating a graph node's data.
type UpdateNodeDataParams struct {
	NodeType              string
	ExternalID            string
	Data                  json.RawMessage
	MatchingSchemaVersion int32
	Metadata              json.RawMessage
	ContentHash           *string
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
}

// CleanupOldSnapshotsParams contains parameters for cleaning up old snapshots.
type CleanupOldSnapshotsParams struct {
	NodeID int64
	Limit  int32
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

// FindNodesNeedingSchemaUpdateParams contains parameters for finding nodes that need schema updates.
type FindNodesNeedingSchemaUpdateParams struct {
	MatchingSchemaVersion int32
	Offset                int32
	Limit                 int32
}

// ptrIfNonEmpty returns a pointer to s if s is non-empty, otherwise nil.
func ptrIfNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
