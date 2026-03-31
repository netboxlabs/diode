package graph

//go:generate go run ../cmd/protograph -proto=../../diode-proto/diode/v1/ingester.proto -output=../gen/protograph/entity_mappings.go

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/entityhash"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/matching"
)

const (
	// DefaultSnapshotRetention is the default number of snapshots to keep
	DefaultSnapshotRetention = 5
	// ContentHashMatchConfidence is the confidence score for content hash matches
	ContentHashMatchConfidence = 0.9

	// DefaultListLimit is used when the caller does not specify a limit.
	DefaultListLimit = 100
	// MaxListLimit is the maximum number of nodes that can be returned in a single ListEntities call.
	MaxListLimit = 1000
)

// EntityStore persists and queries entities in the graph.
type EntityStore interface {
	UpsertEntity(ctx context.Context, entity *diodepb.Entity, requestMetadata ...map[string]any) (*Node, error)
	ListEntities(ctx context.Context, params ListNodesParams) ([]NodeWithLatestSnapshot, error)
}

// Compile-time check that Service implements EntityStore.
var _ EntityStore = (*Service)(nil)

// Service handles extraction and persistence of entity relationships.
// A Service is not safe for concurrent use; each goroutine should use its own
// instance, or calls to UpsertEntity must be serialized.
type Service struct {
	repo              Repository
	logger            *slog.Logger
	nodeCache         map[string]*Node       // Cache for deduplication: "nodeType:externalID" -> *Node
	entityMatcher     matching.EntityMatcher // Confidence-based entity matcher
	updatedNodes      map[string]*Node       // Track nodes that were updated during processing
	seenInThisRequest map[string]bool        // Track which nodes have been seen in this ingestion request
	matchingConfig    *matching.Config       // Entity matching configuration for extracting attributes
	snapshotRetention int                    // Number of snapshots to retain per node
	requestMetadata   map[string]any         // Request-level metadata (e.g. run_id) merged into entity metadata
}

// Option configures a Service.
type Option func(*Service)

// WithEntityMatcher sets the entity matcher for confidence-based matching.
func WithEntityMatcher(matcher matching.EntityMatcher) Option {
	return func(s *Service) { s.entityMatcher = matcher }
}

// WithMatchingConfig sets the entity matching configuration.
func WithMatchingConfig(config *matching.Config) Option {
	return func(s *Service) { s.matchingConfig = config }
}

// WithSnapshotRetention sets the number of snapshots to retain per node.
// Values <= 0 are ignored (the default is used).
func WithSnapshotRetention(count int) Option {
	return func(s *Service) {
		if count > 0 {
			s.snapshotRetention = count
		}
	}
}

// NewService creates a new Service instance.
func NewService(repo Repository, logger *slog.Logger, opts ...Option) *Service {
	s := &Service{
		repo:              repo,
		logger:            logger,
		nodeCache:         make(map[string]*Node),
		updatedNodes:      make(map[string]*Node),
		seenInThisRequest: make(map[string]bool),
		snapshotRetention: DefaultSnapshotRetention,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// UpsertEntity processes an entity and creates/updates its graph representation recursively.
func (s *Service) UpsertEntity(ctx context.Context, entity *diodepb.Entity, requestMetadata ...map[string]any) (*Node, error) {
	if entity == nil || entity.GetEntity() == nil {
		return nil, fmt.Errorf("entity or entity content is nil")
	}

	// Clear caches for each top-level extraction
	s.nodeCache = make(map[string]*Node)
	s.updatedNodes = make(map[string]*Node)
	s.seenInThisRequest = make(map[string]bool)

	// Store request-level metadata for merging during entity processing
	s.requestMetadata = nil
	if len(requestMetadata) > 0 && requestMetadata[0] != nil {
		s.requestMetadata = requestMetadata[0]
	}

	// Recursively process the entire entity tree
	rootNode, err := s.processEntityRecursively(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("processing entity tree: %w", err)
	}

	// Propagate updates to dependent nodes
	if len(s.updatedNodes) > 0 {
		err = s.propagateNodeUpdates(ctx)
		if err != nil {
			s.logger.Warn("failed to propagate node updates", "error", err)
		}
	}

	s.logger.Debug("entity upserted",
		"root_node_type", rootNode.NodeType,
		"root_external_id", rootNode.ExternalID,
		"total_nodes_processed", len(s.nodeCache),
		"nodes_updated", len(s.updatedNodes))

	return rootNode, nil
}

// ListEntities returns a paginated list of entities with optional filtering.
func (s *Service) ListEntities(ctx context.Context, params ListNodesParams) ([]NodeWithLatestSnapshot, error) {
	if params.Limit <= 0 {
		params.Limit = DefaultListLimit
	} else if params.Limit > MaxListLimit {
		params.Limit = MaxListLimit
	}
	return s.repo.ListNodes(ctx, params)
}

// GetCompleteNodeData retrieves a node with its complete entity data by combining matching data and latest snapshot.
func (s *Service) GetCompleteNodeData(ctx context.Context, nodeType, externalID string) (*CompleteNodeData, error) {
	// Get the node with its latest snapshot in a single query
	nodeWithSnapshot, err := s.repo.GetNodeWithLatestSnapshot(ctx, GetNodeWithLatestSnapshotParams{
		NodeType:   nodeType,
		ExternalID: externalID,
	})
	if err != nil {
		return nil, fmt.Errorf("getting node with latest snapshot: %w", err)
	}

	result := &CompleteNodeData{
		ID:             nodeWithSnapshot.ID,
		ExternalID:     nodeWithSnapshot.ExternalID,
		NodeType:       nodeWithSnapshot.NodeType,
		MatchingData:   nodeWithSnapshot.MatchingData,
		DuplicateCount: nodeWithSnapshot.DuplicateCount,
		CreatedAt:      nodeWithSnapshot.CreatedAt,
		UpdatedAt:      nodeWithSnapshot.UpdatedAt,
	}

	// Check if we have snapshot data
	if len(nodeWithSnapshot.SnapshotData) > 0 {
		result.CompleteData = nodeWithSnapshot.SnapshotData
		result.SnapshotSequenceNumber = &nodeWithSnapshot.SequenceNumber
		result.SnapshotCreatedAt = nodeWithSnapshot.SnapshotCreatedAt
	} else {
		// No snapshot available, use matching data as fallback
		result.CompleteData = nodeWithSnapshot.MatchingData
		s.logger.Warn("no snapshot available for node, using matching data as complete data",
			"node_type", nodeType,
			"external_id", externalID)
	}

	return result, nil
}

// GetSnapshots retrieves all snapshots for a given node.
func (s *Service) GetSnapshots(ctx context.Context, nodeID int64, limit, offset int) ([]Snapshot, error) {
	return s.repo.GetSnapshotsByNode(ctx, GetSnapshotsByNodeParams{
		NodeID: nodeID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
}

// processEntityRecursively processes an entity and all its nested entities recursively.
func (s *Service) processEntityRecursively(ctx context.Context, entity *diodepb.Entity) (*Node, error) {
	if entity == nil || entity.GetEntity() == nil {
		return nil, nil
	}

	// Extract node type using proto names for consistency
	nodeType := getEntityTypeName(entity)
	if nodeType == "" {
		return nil, fmt.Errorf("unknown entity type")
	}

	// Generate content hash for same-request deduplication only
	fingerprinter := entityhash.NewEntityFingerprinter()
	contentHash, err := fingerprinter.GenerateEntityHash(entity)
	if err != nil {
		return nil, fmt.Errorf("generating entity hash: %w", err)
	}

	// Check cache using content hash (same-request deduplication)
	cacheKey := fmt.Sprintf("%s:%s", nodeType, contentHash)
	if cachedNode, exists := s.nodeCache[cacheKey]; exists {
		return cachedNode, nil
	}

	// Try to find existing entity match (diode_id checked first, then other metadata, then field-based, then content hash)
	bestMatch := s.findEntityMatch(ctx, entity, nodeType, contentHash)

	// Determine externalID: use existing node's UUID or generate new one
	var externalID string
	if bestMatch != nil && bestMatch.ExternalID != nil {
		externalID = *bestMatch.ExternalID
	} else {
		externalID = uuid.New().String()
	}

	// Marshal entity data
	entityData, err := protojson.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("marshaling entity: %w", err)
	}

	// Process node: either update matched node or create new one
	var node *Node
	if bestMatch != nil && bestMatch.NodeID != nil {
		node, err = s.updateMatchedNode(ctx, bestMatch, nodeType, entityData, entity, contentHash)
		if err != nil {
			return nil, err
		}
	} else {
		node, err = s.upsertNode(ctx, externalID, nodeType, entityData, entity, contentHash)
		if err != nil {
			return nil, fmt.Errorf("upserting node: %w", err)
		}
	}

	// Cache the node for deduplication
	s.nodeCache[cacheKey] = node

	// Recursively process all nested entities and create edges
	s.createEdgesForNode(ctx, entity, node, nodeType, externalID)

	return node, nil
}

// updateMatchedNode updates an existing matched node with new entity data.
func (s *Service) updateMatchedNode(ctx context.Context, bestMatch *matching.MatchResult, nodeType string, entityData json.RawMessage, entity *diodepb.Entity, contentHash string) (*Node, error) {
	if bestMatch.ExternalID == nil {
		return nil, fmt.Errorf("matched node missing external ID")
	}

	existingNode, err := s.repo.FindNode(ctx, FindNodeParams{
		NodeType:   nodeType,
		ExternalID: *bestMatch.ExternalID,
	})
	if err != nil {
		return nil, fmt.Errorf("finding existing matched node: %w", err)
	}

	// Extract matching attributes from the new entity data
	matchingData, err := s.extractMatchingAttributes(nodeType, entityData)
	if err != nil {
		s.logger.Warn("failed to extract matching attributes, using full data", "error", err, "node_type", nodeType)
		matchingData = entityData
	}

	// Extract metadata from entity and source
	metadata, err := s.extractMetadata(entity)
	if err != nil {
		s.logger.Warn("failed to extract metadata", "error", err, "node_type", nodeType)
		metadata = json.RawMessage("{}")
	}

	// Ensure source_match.diode_id is set to externalID for future lookups
	metadata = s.ensureDiodeID(metadata, existingNode.ExternalID)

	// Upsert the node with appropriate duplicate count handling
	result, err := s.upsertMatchedNodeData(ctx, nodeType, existingNode.ExternalID, matchingData, metadata, contentHash)
	if err != nil {
		return nil, fmt.Errorf("updating matched node: %w", err)
	}

	// Create snapshot with full entity data and current metadata
	if err := s.createSnapshot(ctx, result.ID, entityData, metadata); err != nil {
		s.logger.Warn("failed to create entity snapshot", "error", err, "node_id", result.ID)
	}

	node := &Node{
		ID:             result.ID,
		ExternalID:     result.ExternalID,
		NodeType:       result.NodeType,
		Data:           result.Data,
		DuplicateCount: result.DuplicateCount,
		Metadata:       result.Metadata,
	}

	// Track this node as updated for propagation
	nodeKey := fmt.Sprintf("%s:%s", nodeType, result.ExternalID)
	s.updatedNodes[nodeKey] = node

	s.logger.Debug("reused existing matched node",
		"matched_node_id", *bestMatch.NodeID,
		"confidence", float64(bestMatch.Confidence),
		"reason", bestMatch.MatchReason,
		"duplicate_count", node.DuplicateCount,
		"external_id", result.ExternalID)

	return node, nil
}

// upsertMatchedNodeData handles the upsert logic for a matched node, respecting duplicate count rules.
func (s *Service) upsertMatchedNodeData(ctx context.Context, nodeType, externalID string, matchingData, metadata json.RawMessage, contentHash string) (Node, error) {
	requestKey := fmt.Sprintf("%s:%s", nodeType, externalID)

	contentHashPtr := ptrIfNonEmpty(contentHash)

	if s.seenInThisRequest[requestKey] {
		// Already seen in this request - update data but don't increment duplicate count
		updateResult, err := s.repo.UpdateNodeData(ctx, UpdateNodeDataParams{
			NodeType:    nodeType,
			ExternalID:  externalID,
			Data:        matchingData,
			Metadata:    metadata,
			ContentHash: contentHashPtr,
		})
		if err != nil {
			return Node{}, err
		}
		return updateResult, nil
	}

	// First time seeing this node in this request - normal upsert with duplicate count increment
	params := UpsertNodeParams{
		ExternalID:  externalID,
		NodeType:    nodeType,
		Data:        matchingData,
		Metadata:    metadata,
		ContentHash: contentHashPtr,
	}
	result, err := s.repo.UpsertNode(ctx, params)
	if err != nil {
		return Node{}, err
	}

	s.seenInThisRequest[requestKey] = true
	return result, nil
}

// upsertNode creates or updates a node, incrementing duplicate count only once per ingestion request.
func (s *Service) upsertNode(ctx context.Context, externalID, nodeType string, fullEntityData json.RawMessage, entity *diodepb.Entity, contentHash string) (*Node, error) {
	// Extract matching attributes from full entity data
	matchingData, err := s.extractMatchingAttributes(nodeType, fullEntityData)
	if err != nil {
		s.logger.Warn("failed to extract matching attributes, using full data", "error", err, "node_type", nodeType)
		matchingData = fullEntityData
	}

	// Extract metadata from entity and source
	metadata, err := s.extractMetadata(entity)
	if err != nil {
		s.logger.Warn("failed to extract metadata", "error", err, "node_type", nodeType)
		metadata = json.RawMessage("{}")
	}

	// Ensure source_match.diode_id is set to externalID for future lookups
	metadata = s.ensureDiodeID(metadata, externalID)

	// Create a unique key for this node in this request
	requestKey := fmt.Sprintf("%s:%s", nodeType, externalID)

	// Check if we've already processed this node in this ingestion request
	alreadySeenInRequest := s.seenInThisRequest[requestKey]

	contentHashPtr := ptrIfNonEmpty(contentHash)

	var result Node

	if alreadySeenInRequest {
		// This node was already seen in this request - update data but don't increment duplicate count
		result, err = s.repo.UpdateNodeData(ctx, UpdateNodeDataParams{
			NodeType:    nodeType,
			ExternalID:  externalID,
			Data:        matchingData,
			Metadata:    metadata,
			ContentHash: contentHashPtr,
		})
		if err != nil {
			return nil, fmt.Errorf("updating node %s/%s: %w", nodeType, externalID, err)
		}
	} else {
		// First time seeing this node in this request - normal upsert with duplicate count increment
		result, err = s.repo.UpsertNode(ctx, UpsertNodeParams{
			ExternalID:  externalID,
			NodeType:    nodeType,
			Data:        matchingData,
			Metadata:    metadata,
			ContentHash: contentHashPtr,
		})
		if err != nil {
			return nil, fmt.Errorf("upserting node %s/%s: %w", nodeType, externalID, err)
		}
		// Mark this node as seen in this request
		s.seenInThisRequest[requestKey] = true
	}

	// Create snapshot with full entity data and current metadata
	err = s.createSnapshot(ctx, result.ID, fullEntityData, metadata)
	if err != nil {
		s.logger.Warn("failed to create entity snapshot", "error", err, "node_id", result.ID)
		// Don't fail the entire operation for snapshot issues
	}

	return &Node{
		ID:             result.ID,
		ExternalID:     result.ExternalID,
		NodeType:       result.NodeType,
		Data:           result.Data,
		DuplicateCount: result.DuplicateCount,
		Metadata:       result.Metadata,
	}, nil
}

// createSnapshot creates a new snapshot for the given node with the full entity data
// and the metadata that was current at the time of ingestion (e.g. run_id).
func (s *Service) createSnapshot(ctx context.Context, nodeID int64, fullEntityData, metadata json.RawMessage) error {
	// Insert the new snapshot (sequence number is auto-computed by the SQL query)
	_, err := s.repo.InsertSnapshot(ctx, InsertSnapshotParams{
		NodeID:       nodeID,
		SnapshotData: fullEntityData,
		Metadata:     metadata,
	})
	if err != nil {
		return fmt.Errorf("inserting snapshot: %w", err)
	}

	// Clean up old snapshots to maintain retention limit
	err = s.repo.CleanupOldSnapshots(ctx, CleanupOldSnapshotsParams{
		NodeID: nodeID,
		Limit:  int32(s.snapshotRetention),
	})
	if err != nil {
		s.logger.Warn("failed to cleanup old snapshots", "error", err, "node_id", nodeID)
		// Don't fail the entire operation for cleanup issues
	}

	return nil
}
