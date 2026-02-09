package graph

//go:generate go run ../cmd/protograph -proto=../../diode-proto/diode/v1/ingester.proto -output=../gen/protograph/entity_mappings.go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/netboxlabs/diode/diode-server/entityhash"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/protograph"
	"github.com/netboxlabs/diode/diode-server/matching"
	"github.com/netboxlabs/diode/diode-server/strcase"
)

const (
	// CurrentSchemaVersion represents the current matching schema version
	CurrentSchemaVersion = 1
	// DefaultSnapshotRetention is the default number of snapshots to keep
	DefaultSnapshotRetention = 5
	// ContentHashMatchConfidence is the confidence score for content hash matches
	ContentHashMatchConfidence = 0.9
)

// Builder handles extraction and persistence of entity relationships
type Builder struct {
	repo              Repository
	logger            *slog.Logger
	nodeCache         map[string]*Node       // Cache for deduplication: "nodeType:externalID" -> *Node
	entityMatcher     matching.EntityMatcher // Confidence-based entity matcher
	updatedNodes      map[string]*Node       // Track nodes that were updated during processing
	seenInThisRequest map[string]bool        // Track which nodes have been seen in this ingestion request
	matchingConfig    *matching.Config       // Entity matching configuration for extracting attributes
	snapshotRetention int                    // Number of snapshots to retain per node
}

// BuilderOption configures a Builder.
type BuilderOption func(*Builder)

// WithEntityMatcher sets the entity matcher for confidence-based matching.
func WithEntityMatcher(matcher matching.EntityMatcher) BuilderOption {
	return func(b *Builder) { b.entityMatcher = matcher }
}

// WithMatchingConfig sets the entity matching configuration.
func WithMatchingConfig(config *matching.Config) BuilderOption {
	return func(b *Builder) { b.matchingConfig = config }
}

// WithSnapshotRetention sets the number of snapshots to retain per node.
// Values <= 0 are ignored (the default is used).
func WithSnapshotRetention(count int) BuilderOption {
	return func(b *Builder) {
		if count > 0 {
			b.snapshotRetention = count
		}
	}
}

// NewBuilder creates a new Builder instance.
func NewBuilder(repo Repository, logger *slog.Logger, opts ...BuilderOption) *Builder {
	b := &Builder{
		repo:              repo,
		logger:            logger,
		nodeCache:         make(map[string]*Node),
		updatedNodes:      make(map[string]*Node),
		seenInThisRequest: make(map[string]bool),
		snapshotRetention: DefaultSnapshotRetention,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// extractMetadata extracts metadata from an entity.
// Returns a JSON-encoded map of metadata key-value pairs.
//
// TODO: Add support for merging request-level metadata (IngestRequest.metadata) with
// entity-level metadata. This would require:
// 1. Pass IngestRequest.metadata to Builder (via SetSourceMetadata or ExtractGraph parameter)
// 2. Merge source metadata with entity metadata (entity takes precedence for duplicate keys)
// 3. Wire up the call in IngestionProcessor.CreateIngestionLogs before ExtractGraph
func (gb *Builder) extractMetadata(entity *diodepb.Entity) (json.RawMessage, error) {
	if entity == nil || entity.GetEntity() == nil {
		return json.RawMessage("{}"), nil
	}

	result := gb.getEntityMetadata(entity)
	if len(result) == 0 {
		return json.RawMessage("{}"), nil
	}

	return json.Marshal(result)
}

// getEntityMetadata extracts the metadata field from the inner entity.
// Each entity type (Device, Site, etc.) has a GetMetadata() method returning *structpb.Struct.
func (gb *Builder) getEntityMetadata(entity *diodepb.Entity) map[string]any {
	return matching.ExtractEntityMetadata(entity)
}

// ensureDiodeID sets source_match.diode_id to the externalID.
// This allows future lookups by diode_id to find the node directly by externalID.
// Always overwrites any client-provided diode_id to ensure consistency.
func (gb *Builder) ensureDiodeID(metadata json.RawMessage, externalID string) json.RawMessage {
	var metaMap map[string]any
	if err := json.Unmarshal(metadata, &metaMap); err != nil {
		gb.logger.Debug("failed to unmarshal metadata, creating new map", "error", err)
		metaMap = make(map[string]any)
	}

	// Get or create source_match
	sourceMatch, ok := metaMap[sourceMatchKey].(map[string]any)
	if !ok {
		sourceMatch = make(map[string]any)
	}

	// Always set diode_id to externalID (system controls identity)
	sourceMatch[diodeIDKey] = externalID
	metaMap[sourceMatchKey] = sourceMatch

	if updated, err := json.Marshal(metaMap); err == nil {
		return updated
	}

	return metadata
}

// extractMatchingAttributes extracts only the attributes needed for matching from the full entity data
func (gb *Builder) extractMatchingAttributes(entityType string, fullEntityData json.RawMessage) (json.RawMessage, error) {
	// If no matching config is set, return the full entity data as fallback
	if gb.matchingConfig == nil {
		return fullEntityData, nil
	}

	// Get matching rules for this entity type
	var entityRule matching.EntityMatchingRule
	var found bool

	// Try to get entity-specific rule first
	if entityRule, found = gb.matchingConfig.DefaultEntityRules[entityType]; !found {
		// Fall back to wildcard rule
		if entityRule, found = gb.matchingConfig.DefaultEntityRules["*"]; !found {
			// No matching rules found, return full data
			return fullEntityData, nil
		}
	}

	// Parse the full entity data
	var fullData map[string]any
	if err := json.Unmarshal(fullEntityData, &fullData); err != nil {
		return nil, fmt.Errorf("failed to parse entity data: %w", err)
	}

	// Extract only the fields used in matching rules
	matchingData := make(map[string]any)

	// Process primary rules
	for _, rule := range entityRule.PrimaryRules {
		if value := gb.extractFieldByPath(fullData, rule.FieldPath); value != nil {
			gb.setFieldByPath(matchingData, rule.FieldPath, value)
		}
	}

	// Process secondary rules
	for _, rule := range entityRule.SecondaryRules {
		if value := gb.extractFieldByPath(fullData, rule.FieldPath); value != nil {
			gb.setFieldByPath(matchingData, rule.FieldPath, value)
		}
	}

	// If no matching data was extracted (empty object), fall back to full entity data
	if len(matchingData) == 0 {
		return fullEntityData, nil
	}

	// Marshal the extracted matching data back to JSON
	matchingJSON, err := json.Marshal(matchingData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal matching data: %w", err)
	}

	return matchingJSON, nil
}

// extractFieldByPath extracts a value from nested data using a field path like "Device.name" or "Device.site.name"
func (gb *Builder) extractFieldByPath(data map[string]any, fieldPath string) any {
	parts := strings.Split(fieldPath, ".")
	current := data

	for _, part := range parts {
		if current == nil {
			return nil
		}

		if nextValue, exists := current[part]; exists {
			if nextMap, ok := nextValue.(map[string]any); ok {
				current = nextMap
			} else {
				// This is the final value
				return nextValue
			}
		} else {
			return nil
		}
	}

	return current
}

// setFieldByPath sets a value in nested data using a field path like "Device.name" or "Device.site.name"
func (gb *Builder) setFieldByPath(data map[string]any, fieldPath string, value any) {
	parts := strings.Split(fieldPath, ".")
	current := data

	// Navigate to the parent of the target field
	for _, part := range parts[:len(parts)-1] {
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]any)
		}
		if nextMap, ok := current[part].(map[string]any); ok {
			current = nextMap
		} else {
			// Path conflicts with existing non-map value, create a new path
			current[part] = make(map[string]any)
			current = current[part].(map[string]any)
		}
	}

	// Set the final value
	finalKey := parts[len(parts)-1]
	current[finalKey] = value
}

// createSnapshot creates a new snapshot for the given node with the full entity data
func (gb *Builder) createSnapshot(ctx context.Context, nodeID int64, fullEntityData json.RawMessage) error {
	// Insert the new snapshot (sequence number is auto-computed by the SQL query)
	_, err := gb.repo.InsertSnapshot(ctx, InsertSnapshotParams{
		NodeID:       nodeID,
		SnapshotData: fullEntityData,
	})
	if err != nil {
		return fmt.Errorf("failed to insert snapshot: %w", err)
	}

	// Clean up old snapshots to maintain retention limit
	err = gb.repo.CleanupOldSnapshots(ctx, CleanupOldSnapshotsParams{
		NodeID: nodeID,
		Limit:  int32(gb.snapshotRetention),
	})
	if err != nil {
		gb.logger.Warn("failed to cleanup old snapshots", "error", err, "node_id", nodeID)
		// Don't fail the entire operation for cleanup issues
	}

	return nil
}

// needsSchemaUpdate checks if a node needs its matching data updated due to schema version mismatch
func (gb *Builder) needsSchemaUpdate(node *Node) bool {
	return node.MatchingSchemaVersion < CurrentSchemaVersion
}

// updateNodeMatchingData updates a node's matching data to the current schema version
func (gb *Builder) updateNodeMatchingData(ctx context.Context, node *Node, fullEntityData json.RawMessage) error {
	// Extract matching attributes based on current schema
	matchingData, err := gb.extractMatchingAttributes(node.NodeType, fullEntityData)
	if err != nil {
		return fmt.Errorf("failed to extract matching attributes: %w", err)
	}

	// Keep existing metadata (schema update doesn't change metadata)
	metadata := node.Metadata
	if metadata == nil {
		metadata = json.RawMessage("{}")
	}

	// Update the node with new matching data and schema version
	_, err = gb.repo.UpdateNodeData(ctx, UpdateNodeDataParams{
		NodeType:              node.NodeType,
		ExternalID:            node.ExternalID,
		Data:                  matchingData,
		MatchingSchemaVersion: CurrentSchemaVersion,
		Metadata:              metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to update node matching data: %w", err)
	}

	return nil
}

// GetCompleteNodeData retrieves a node with its complete entity data by combining matching data and latest snapshot
func (gb *Builder) GetCompleteNodeData(ctx context.Context, nodeType, externalID string) (*CompleteNodeData, error) {
	// Get the node with its latest snapshot in a single query
	nodeWithSnapshot, err := gb.repo.GetNodeWithLatestSnapshot(ctx, GetNodeWithLatestSnapshotParams{
		NodeType:   nodeType,
		ExternalID: externalID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get node with latest snapshot: %w", err)
	}

	result := &CompleteNodeData{
		ID:                    nodeWithSnapshot.ID,
		ExternalID:            nodeWithSnapshot.ExternalID,
		NodeType:              nodeWithSnapshot.NodeType,
		MatchingData:          nodeWithSnapshot.MatchingData,
		DuplicateCount:        nodeWithSnapshot.DuplicateCount,
		MatchingSchemaVersion: nodeWithSnapshot.MatchingSchemaVersion,
		CreatedAt:             nodeWithSnapshot.CreatedAt,
		UpdatedAt:             nodeWithSnapshot.UpdatedAt,
	}

	// Check if we have snapshot data
	if len(nodeWithSnapshot.SnapshotData) > 0 {
		result.CompleteData = nodeWithSnapshot.SnapshotData
		result.SnapshotSequenceNumber = &nodeWithSnapshot.SequenceNumber
		result.SnapshotCreatedAt = nodeWithSnapshot.SnapshotCreatedAt
	} else {
		// No snapshot available, use matching data as fallback
		result.CompleteData = nodeWithSnapshot.MatchingData
		gb.logger.Warn("no snapshot available for node, using matching data as complete data",
			"node_type", nodeType,
			"external_id", externalID)
	}

	return result, nil
}

// GetSnapshots retrieves all snapshots for a given node
func (gb *Builder) GetSnapshots(ctx context.Context, nodeID int64, limit, offset int) ([]Snapshot, error) {
	return gb.repo.GetSnapshotsByNode(ctx, GetSnapshotsByNodeParams{
		NodeID: nodeID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
}

// UpsertNode creates or updates a graph node.
func (gb *Builder) UpsertNode(ctx context.Context, params UpsertNodeParams) (Node, error) {
	return gb.repo.UpsertNode(ctx, params)
}

// FindNode looks up a graph node by type and external ID.
func (gb *Builder) FindNode(ctx context.Context, params FindNodeParams) (Node, error) {
	return gb.repo.FindNode(ctx, params)
}

// UpsertEdge creates or updates an edge between two nodes.
func (gb *Builder) UpsertEdge(ctx context.Context, params UpsertEdgeParams) error {
	return gb.repo.UpsertEdge(ctx, params)
}

// CompleteNodeData represents a node with its complete entity data
type CompleteNodeData struct {
	ID                     int64           `json:"id"`
	ExternalID             string          `json:"external_id"`
	NodeType               string          `json:"node_type"`
	MatchingData           json.RawMessage `json:"matching_data"` // Data used for matching
	CompleteData           json.RawMessage `json:"complete_data"` // Complete entity data from latest snapshot
	DuplicateCount         int32           `json:"duplicate_count"`
	MatchingSchemaVersion  int32           `json:"matching_schema_version"`
	SnapshotSequenceNumber *int32          `json:"snapshot_sequence_number,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
	SnapshotCreatedAt      *time.Time      `json:"snapshot_created_at,omitempty"`
}

// MigrateNodesToCurrentSchema performs bulk migration of nodes to the current schema version
func (gb *Builder) MigrateNodesToCurrentSchema(ctx context.Context, batchSize int) (int, error) {
	if gb.matchingConfig == nil {
		return 0, fmt.Errorf("matching configuration not set, cannot perform schema migration")
	}

	migratedCount := 0
	offset := 0

	for {
		// Find nodes needing schema update
		nodes, err := gb.repo.FindNodesNeedingSchemaUpdate(ctx, FindNodesNeedingSchemaUpdateParams{
			MatchingSchemaVersion: int32(CurrentSchemaVersion),
			Limit:                 int32(batchSize),
			Offset:                int32(offset),
		})
		if err != nil {
			return migratedCount, fmt.Errorf("failed to find nodes needing schema update: %w", err)
		}

		if len(nodes) == 0 {
			break // No more nodes to migrate
		}

		// Process each node
		for _, node := range nodes {
			// Get the latest snapshot for this node to reconstruct full entity data
			latestSnapshot, err := gb.repo.GetLatestSnapshot(ctx, node.ID)
			if err != nil {
				gb.logger.Warn("failed to get latest snapshot for schema migration, skipping node",
					"error", err,
					"node_id", node.ID,
					"node_type", node.NodeType,
					"external_id", node.ExternalID)
				continue
			}

			// Update the node's matching data based on current schema
			err = gb.updateNodeMatchingData(ctx, &node, latestSnapshot.SnapshotData)
			if err != nil {
				gb.logger.Warn("failed to update node matching data during schema migration",
					"error", err,
					"node_id", node.ID,
					"node_type", node.NodeType,
					"external_id", node.ExternalID)
				continue
			}

			migratedCount++
			gb.logger.Debug("migrated node to current schema",
				"node_id", node.ID,
				"node_type", node.NodeType,
				"external_id", node.ExternalID,
				"old_version", node.MatchingSchemaVersion,
				"new_version", CurrentSchemaVersion)
		}

		offset += len(nodes)
	}

	gb.logger.Info("completed schema migration",
		"migrated_count", migratedCount,
		"target_schema_version", CurrentSchemaVersion)

	return migratedCount, nil
}

// ExtractGraph processes an entity and creates/updates its graph representation recursively
func (gb *Builder) ExtractGraph(ctx context.Context, entity *diodepb.Entity) error {
	if entity == nil || entity.GetEntity() == nil {
		return fmt.Errorf("entity or entity content is nil")
	}

	// Clear caches for each top-level extraction
	gb.nodeCache = make(map[string]*Node)
	gb.updatedNodes = make(map[string]*Node)
	gb.seenInThisRequest = make(map[string]bool)

	// Recursively process the entire entity tree
	rootNode, err := gb.processEntityRecursively(ctx, entity)
	if err != nil {
		return fmt.Errorf("failed to process entity tree: %w", err)
	}

	// Propagate updates to dependent nodes
	if len(gb.updatedNodes) > 0 {
		err = gb.propagateNodeUpdates(ctx)
		if err != nil {
			gb.logger.Warn("failed to propagate node updates", "error", err)
		}
	}

	gb.logger.Debug("completed recursive graph extraction",
		"root_node_type", rootNode.NodeType,
		"root_external_id", rootNode.ExternalID,
		"total_nodes_processed", len(gb.nodeCache),
		"nodes_updated", len(gb.updatedNodes))

	return nil
}

// processEntityRecursively processes an entity and all its nested entities recursively
func (gb *Builder) processEntityRecursively(ctx context.Context, entity *diodepb.Entity) (*Node, error) {
	if entity == nil || entity.GetEntity() == nil {
		return nil, nil
	}

	// Extract node type using proto names for consistency
	nodeType := getEntityTypeName(entity)
	if nodeType == "" {
		return nil, fmt.Errorf("failed to get object type from entity")
	}

	// Generate content hash for same-request deduplication only
	fingerprinter := entityhash.NewEntityFingerprinter()
	contentHash, err := fingerprinter.GenerateEntityHash(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entity hash: %w", err)
	}

	// Check cache using content hash (same-request deduplication)
	cacheKey := fmt.Sprintf("%s:%s", nodeType, contentHash)
	if cachedNode, exists := gb.nodeCache[cacheKey]; exists {
		return cachedNode, nil
	}

	// Try to find existing entity match (diode_id checked first, then other metadata, then field-based, then content hash)
	bestMatch := gb.findEntityMatch(ctx, entity, nodeType, contentHash)

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
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	// Process node: either update matched node or create new one
	var node *Node
	if bestMatch != nil && bestMatch.NodeID != nil {
		node, err = gb.updateMatchedNode(ctx, bestMatch, nodeType, entityData, entity, contentHash)
		if err != nil {
			return nil, err
		}
	} else {
		node, err = gb.upsertNode(ctx, externalID, nodeType, entityData, entity, contentHash)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert node: %w", err)
		}
	}

	// Cache the node for deduplication
	gb.nodeCache[cacheKey] = node

	// Recursively process all nested entities and create edges
	gb.createEdgesForNode(ctx, entity, node, nodeType, externalID)

	return node, nil
}

// findEntityMatch attempts to find an existing entity match.
// Priority order:
// 1. Metadata match (correlation IDs) - works without entity matcher config
// 2. Field-based matching via entityMatcher (requires config)
// 3. Content hash fallback - last resort when entity matcher config is missing
// Returns nil if no match found.
func (gb *Builder) findEntityMatch(ctx context.Context, entity *diodepb.Entity, nodeType string, contentHash string) *matching.MatchResult {
	// Priority 1: Check metadata for correlation IDs (works without entity matcher)
	if metadataMatch := gb.findMatchByMetadata(ctx, entity, nodeType); metadataMatch != nil {
		return metadataMatch
	}

	// Priority 2: Fall back to field-based matching if entity matcher is configured
	if gb.entityMatcher != nil {
		bestMatch, err := gb.entityMatcher.FindBestMatch(ctx, entity)
		if err != nil {
			gb.logger.Warn("failed to find entity matches", "error", err, "entity_type", nodeType)
		} else if bestMatch != nil && bestMatch.NodeID != nil {
			gb.logger.Debug("found confident match for entity",
				"entity_type", nodeType,
				"matched_node_id", *bestMatch.NodeID,
				"confidence", float64(bestMatch.Confidence),
				"reason", bestMatch.MatchReason,
				"matching_fields", bestMatch.MatchingFields)
			return bestMatch
		}
	}

	// Priority 3: Content hash fallback - last resort when entity matcher config is missing
	if contentHash != "" {
		if hashMatch := gb.findNodeByContentHash(ctx, nodeType, contentHash); hashMatch != nil {
			return hashMatch
		}
	}

	return nil
}

const (
	// sourceMatchKey is the metadata key containing correlation IDs for entity matching
	sourceMatchKey = "source_match"
	// diodeIDKey is the internal Diode entity ID - highest priority for matching
	diodeIDKey = "diode_id"
)

// findMatchByMetadata attempts to find an existing node by checking the source_match metadata.
// Priority order:
// 1. source_match.diode_id - if found, return immediately (internal Diode entity)
// 2. Other source_match.* keys - iterate and return on first match found
// Returns a match result with confidence 1.0 if found, nil if no match.
func (gb *Builder) findMatchByMetadata(ctx context.Context, entity *diodepb.Entity, nodeType string) *matching.MatchResult {
	if entity == nil {
		return nil
	}

	// Extract metadata from the entity
	entityMetadata := matching.ExtractEntityMetadata(entity)
	if len(entityMetadata) == 0 {
		return nil
	}

	// Get source_match metadata - only this is used for matching
	sourceMatchRaw, exists := entityMetadata[sourceMatchKey]
	if !exists {
		return nil
	}
	sourceMatch, ok := sourceMatchRaw.(map[string]any)
	if !ok || len(sourceMatch) == 0 {
		return nil
	}

	// Priority 1: Check diode_id first (internal Diode entity)
	// diode_id equals externalID, so we can do a direct lookup
	if diodeID, exists := sourceMatch[diodeIDKey]; exists && diodeID != nil {
		if diodeIDStr, ok := diodeID.(string); ok && diodeIDStr != "" {
			if result := gb.findNodeByExternalID(ctx, nodeType, diodeIDStr); result != nil {
				gb.logger.Debug("found diode_id match for entity",
					"entity_type", nodeType,
					"matched_node_id", *result.NodeID,
					"diode_id", diodeID)
				return result
			}
		}
	}

	// Priority 2: Check other source_match keys (any match)
	for key, value := range sourceMatch {
		if key == diodeIDKey || value == nil {
			continue
		}

		if result := gb.findNodeBySourceMatchKey(ctx, nodeType, key, value); result != nil {
			gb.logger.Debug("found metadata match for entity",
				"entity_type", nodeType,
				"matched_node_id", *result.NodeID,
				"matching_key", key,
				"matching_value", value)
			return result
		}
	}

	return nil
}

// findNodeBySourceMatchKey searches for a node by a single source_match key-value pair.
func (gb *Builder) findNodeBySourceMatchKey(ctx context.Context, nodeType, key string, value any) *matching.MatchResult {
	// Build filter: {"source_match": {key: value}}
	metadataFilter := map[string]any{
		sourceMatchKey: map[string]any{
			key: value,
		},
	}
	filterJSON, err := json.Marshal(metadataFilter)
	if err != nil {
		gb.logger.Warn("failed to marshal metadata filter", "error", err, "key", key)
		return nil
	}

	// Search for existing node with this metadata
	node, err := gb.repo.FindNodeByMetadata(ctx, FindNodeByMetadataParams{
		NodeType:       nodeType,
		MetadataFilter: filterJSON,
	})
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			gb.logger.Warn("metadata lookup failed", "error", err, "key", key, "value", value)
		}
		return nil
	}

	return &matching.MatchResult{
		NodeID:         &node.ID,
		ExternalID:     &node.ExternalID,
		Confidence:     1.0,
		MatchingFields: []string{fmt.Sprintf("%s.%s", sourceMatchKey, key)},
		MatchReason:    fmt.Sprintf("Metadata match: %s.%s=%v", sourceMatchKey, key, value),
		ExistingData:   node.Data,
	}
}

// findNodeByExternalID searches for a node directly by externalID (used for diode_id lookup).
func (gb *Builder) findNodeByExternalID(ctx context.Context, nodeType, externalID string) *matching.MatchResult {
	node, err := gb.repo.FindNode(ctx, FindNodeParams{
		NodeType:   nodeType,
		ExternalID: externalID,
	})
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			gb.logger.Warn("externalID lookup failed", "error", err, "external_id", externalID)
		}
		return nil
	}

	return &matching.MatchResult{
		NodeID:         &node.ID,
		ExternalID:     &node.ExternalID,
		Confidence:     1.0,
		MatchingFields: []string{fmt.Sprintf("%s.%s", sourceMatchKey, diodeIDKey)},
		MatchReason:    fmt.Sprintf("Direct lookup: %s.%s=%s", sourceMatchKey, diodeIDKey, externalID),
		ExistingData:   node.Data,
	}
}

// findNodeByContentHash searches for a node by content hash (last resort fallback).
// Used when entity matcher config is missing and no metadata match found.
func (gb *Builder) findNodeByContentHash(ctx context.Context, nodeType, contentHash string) *matching.MatchResult {
	node, err := gb.repo.FindNodeByContentHash(ctx, FindNodeByContentHashParams{
		NodeType:    nodeType,
		ContentHash: contentHash,
	})
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			gb.logger.Warn("content hash lookup failed", "error", err, "content_hash", contentHash)
		}
		return nil
	}

	gb.logger.Debug("found content hash match for entity",
		"entity_type", nodeType,
		"matched_node_id", node.ID,
		"content_hash", contentHash)

	return &matching.MatchResult{
		NodeID:         &node.ID,
		ExternalID:     &node.ExternalID,
		Confidence:     ContentHashMatchConfidence,
		MatchingFields: []string{"content_hash"},
		MatchReason:    fmt.Sprintf("Content hash match: %s", contentHash),
		ExistingData:   node.Data,
	}
}

// updateMatchedNode updates an existing matched node with new entity data.
func (gb *Builder) updateMatchedNode(ctx context.Context, bestMatch *matching.MatchResult, nodeType string, entityData json.RawMessage, entity *diodepb.Entity, contentHash string) (*Node, error) {
	if bestMatch.ExternalID == nil {
		return nil, fmt.Errorf("matched node missing external ID")
	}

	existingNode, err := gb.repo.FindNode(ctx, FindNodeParams{
		NodeType:   nodeType,
		ExternalID: *bestMatch.ExternalID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find existing matched node: %w", err)
	}

	// Extract matching attributes from the new entity data
	matchingData, err := gb.extractMatchingAttributes(nodeType, entityData)
	if err != nil {
		gb.logger.Warn("failed to extract matching attributes, using full data", "error", err, "node_type", nodeType)
		matchingData = entityData
	}

	// Extract metadata from entity and source
	metadata, err := gb.extractMetadata(entity)
	if err != nil {
		gb.logger.Warn("failed to extract metadata", "error", err, "node_type", nodeType)
		metadata = json.RawMessage("{}")
	}

	// Ensure source_match.diode_id is set to externalID for future lookups
	metadata = gb.ensureDiodeID(metadata, existingNode.ExternalID)

	// Check if existing node needs schema update (lazy migration)
	gb.maybeUpdateNodeSchema(ctx, &existingNode, nodeType, entityData)

	// Upsert the node with appropriate duplicate count handling
	result, err := gb.upsertMatchedNodeData(ctx, nodeType, existingNode.ExternalID, matchingData, metadata, contentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to update matched node: %w", err)
	}

	// Create snapshot with full entity data
	if err := gb.createSnapshot(ctx, result.ID, entityData); err != nil {
		gb.logger.Warn("failed to create entity snapshot", "error", err, "node_id", result.ID)
	}

	node := &Node{
		ID:                    result.ID,
		ExternalID:            result.ExternalID,
		NodeType:              result.NodeType,
		Data:                  result.Data,
		DuplicateCount:        result.DuplicateCount,
		MatchingSchemaVersion: result.MatchingSchemaVersion,
		Metadata:              result.Metadata,
	}

	// Track this node as updated for propagation
	nodeKey := fmt.Sprintf("%s:%s", nodeType, result.ExternalID)
	gb.updatedNodes[nodeKey] = node

	gb.logger.Debug("reused existing matched node",
		"matched_node_id", *bestMatch.NodeID,
		"confidence", float64(bestMatch.Confidence),
		"reason", bestMatch.MatchReason,
		"duplicate_count", node.DuplicateCount,
		"external_id", result.ExternalID)

	return node, nil
}

// maybeUpdateNodeSchema checks if a node needs schema migration and updates it if necessary.
func (gb *Builder) maybeUpdateNodeSchema(ctx context.Context, existingNode *Node, nodeType string, entityData json.RawMessage) {
	if gb.needsSchemaUpdate(existingNode) {
		gb.logger.Info("updating node matching data due to schema change",
			"node_type", nodeType,
			"external_id", existingNode.ExternalID,
			"old_version", existingNode.MatchingSchemaVersion,
			"new_version", CurrentSchemaVersion)

		if err := gb.updateNodeMatchingData(ctx, existingNode, entityData); err != nil {
			gb.logger.Warn("failed to update node schema, continuing", "error", err)
		}
	}
}

// upsertMatchedNodeData handles the upsert logic for a matched node, respecting duplicate count rules.
func (gb *Builder) upsertMatchedNodeData(ctx context.Context, nodeType, externalID string, matchingData, metadata json.RawMessage, contentHash string) (Node, error) {
	requestKey := fmt.Sprintf("%s:%s", nodeType, externalID)

	contentHashPtr := ptrIfNonEmpty(contentHash)

	if gb.seenInThisRequest[requestKey] {
		// Already seen in this request - update data but don't increment duplicate count
		updateResult, err := gb.repo.UpdateNodeData(ctx, UpdateNodeDataParams{
			NodeType:              nodeType,
			ExternalID:            externalID,
			Data:                  matchingData,
			MatchingSchemaVersion: CurrentSchemaVersion,
			Metadata:              metadata,
			ContentHash:           contentHashPtr,
		})
		if err != nil {
			return Node{}, err
		}
		return updateResult, nil
	}

	// First time seeing this node in this request - normal upsert with duplicate count increment
	params := UpsertNodeParams{
		ExternalID:            externalID,
		NodeType:              nodeType,
		Data:                  matchingData,
		MatchingSchemaVersion: CurrentSchemaVersion,
		Metadata:              metadata,
		ContentHash:           contentHashPtr,
	}
	result, err := gb.repo.UpsertNode(ctx, params)
	if err != nil {
		return Node{}, err
	}

	gb.seenInThisRequest[requestKey] = true
	return result, nil
}

// createEdgesForNode extracts and creates all edges for a node.
func (gb *Builder) createEdgesForNode(ctx context.Context, entity *diodepb.Entity, node *Node, nodeType, externalID string) {
	edges, err := gb.extractEdgesRecursively(ctx, entity, node)
	if err != nil {
		gb.logger.Warn("failed to extract edges recursively",
			"node_type", nodeType,
			"external_id", externalID,
			"error", err)
		return
	}

	for _, edge := range edges {
		if err := gb.upsertEdge(ctx, edge); err != nil {
			gb.logger.Warn("failed to create edge",
				"source_id", edge.SourceNodeID,
				"target_id", edge.TargetNodeID,
				"edge_type", edge.EdgeType,
				"error", err)
		}
	}
}

// extractEdgesRecursively finds all relationships within the entity and creates edges recursively
func (gb *Builder) extractEdgesRecursively(ctx context.Context, entity *diodepb.Entity, sourceNode *Node) ([]*Edge, error) {
	var edges []*Edge

	// Get the actual entity from the wrapper (e.g., Device from Entity_Device)
	wrapperValue := reflect.ValueOf(entity.GetEntity())
	if wrapperValue.Kind() == reflect.Ptr {
		wrapperValue = wrapperValue.Elem()
	}

	// The wrapper has one field containing the actual entity
	if wrapperValue.NumField() != 1 {
		return edges, fmt.Errorf("expected wrapper to have exactly 1 field, got %d", wrapperValue.NumField())
	}

	// Get the actual entity (e.g., the Device struct inside Entity_Device)
	actualEntityField := wrapperValue.Field(0)
	if !actualEntityField.IsValid() || actualEntityField.IsNil() {
		return edges, nil
	}

	entityValue := actualEntityField
	if entityValue.Kind() == reflect.Ptr {
		entityValue = entityValue.Elem()
	}

	entityType := entityValue.Type()

	for i := 0; i < entityValue.NumField(); i++ {
		field := entityValue.Field(i)
		fieldType := entityType.Field(i)
		fieldName := fieldType.Name

		// Skip non-exported fields
		if !field.CanInterface() {
			continue
		}

		// Process single entity field
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			fieldEdges, err := gb.processFieldRecursively(ctx, sourceNode, field, fieldName)
			if err != nil {
				gb.logger.Warn("failed to process field recursively", "field", fieldName, "error", err)
				continue
			}
			if len(fieldEdges) > 0 {
				edges = append(edges, fieldEdges...)
			}
		}

		// Process slice fields (repeated references)
		if field.Kind() == reflect.Slice {
			sliceEdges, err := gb.processSliceFieldRecursively(ctx, sourceNode, field, fieldName)
			if err != nil {
				gb.logger.Warn("failed to process slice field recursively", "field", fieldName, "error", err)
				continue
			}
			edges = append(edges, sliceEdges...)
		}

		// Process oneof fields (interface types in generated Go code)
		if field.Kind() == reflect.Interface && !field.IsNil() {
			// Unwrap the interface to get the concrete value
			concreteValue := reflect.ValueOf(field.Interface())
			if concreteValue.IsValid() && concreteValue.Kind() == reflect.Ptr && !concreteValue.IsNil() {
				fieldEdges, err := gb.processFieldRecursively(ctx, sourceNode, concreteValue, fieldName)
				if err != nil {
					gb.logger.Warn("failed to process oneof field recursively", "field", fieldName, "error", err)
					continue
				}
				if len(fieldEdges) > 0 {
					edges = append(edges, fieldEdges...)
				}
			}
		}
	}

	return edges, nil
}

// extractEdgeProperties extracts properties to be stored on an edge based on reflection.
// It automatically identifies fields that should be edge properties (not node properties)
// based on the edge_property_fields configuration in the matching config.
func (gb *Builder) extractEdgeProperties(fieldValue any) json.RawMessage {
	if fieldValue == nil {
		return json.RawMessage("{}")
	}

	// If no matching config, return empty properties
	if gb.matchingConfig == nil {
		return json.RawMessage("{}")
	}

	// Get the entity type name from the field value
	entityTypeName := gb.getEntityTypeNameFromValue(fieldValue)
	if entityTypeName == "" {
		return json.RawMessage("{}")
	}

	// Look up the matching rule for this entity type to get edge property fields
	entityRule, found := gb.matchingConfig.DefaultEntityRules[entityTypeName]
	if !found || len(entityRule.EdgePropertyFields) == 0 {
		return json.RawMessage("{}")
	}

	propertyFieldNames := entityRule.EdgePropertyFields

	// Use reflection to extract the configured fields
	properties := make(map[string]any)
	val := reflect.ValueOf(fieldValue)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Extract each configured edge property field
	for _, fieldName := range propertyFieldNames {
		if field := val.FieldByName(fieldName); field.IsValid() && field.CanInterface() {
			fv := field.Interface()

			// Only include non-zero values
			if !reflect.DeepEqual(fv, reflect.Zero(field.Type()).Interface()) {
				// Convert field name to snake_case for JSON
				jsonFieldName := strcase.ToSnakeCase(fieldName)
				properties[jsonFieldName] = fv
			}
		}
	}

	// If we extracted any properties, return them as JSON
	if len(properties) > 0 {
		if propsJSON, err := json.Marshal(properties); err == nil {
			return propsJSON
		}
	}

	// Default: return empty JSON object for edges without special properties
	return json.RawMessage("{}")
}

// getEntityTypeNameFromValue extracts the entity type name from a reflect.Value
func (gb *Builder) getEntityTypeNameFromValue(fieldValue any) string {
	typ := reflect.TypeOf(fieldValue)
	if typ == nil {
		return ""
	}

	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// Get the type name (e.g., "ServicePort")
	return typ.Name()
}

// processFieldRecursively processes a single field and creates bidirectional edges if it references another entity.
// Returns both forward (BELONGS_TO) and reverse (HAS) edges.
func (gb *Builder) processFieldRecursively(ctx context.Context, sourceNode *Node, field reflect.Value, fieldName string) ([]*Edge, error) {
	if !field.IsValid() || field.IsNil() {
		return nil, nil
	}

	fieldInterface := field.Interface()

	// Try to recursively process this field as a nested entity first
	targetNode, err := gb.processNestedEntityRecursively(ctx, fieldInterface)
	if err != nil {
		return nil, err
	}

	// If recursive processing didn't work, try the fallback approach
	if targetNode == nil {
		targetNode, err = gb.findTargetNodeFromField(ctx, fieldInterface, fieldName)
		if err != nil {
			return nil, err
		}
	}

	if targetNode == nil {
		return nil, nil
	}

	// Create bidirectional edges
	edgeTypes := gb.getEdgeTypesForField(fieldName, sourceNode.NodeType)

	return []*Edge{
		{
			// Forward: Source BELONGS_TO Target (e.g., Device BELONGS_TO_SITE Site)
			SourceNodeID: sourceNode.ID,
			TargetNodeID: targetNode.ID,
			EdgeType:     edgeTypes.Forward,
			Properties:   json.RawMessage("{}"),
		},
		{
			// Reverse: Target HAS Source (e.g., Site HAS_DEVICE Device)
			SourceNodeID: targetNode.ID,
			TargetNodeID: sourceNode.ID,
			EdgeType:     edgeTypes.Reverse,
			Properties:   json.RawMessage("{}"),
		},
	}, nil
}

// processSliceFieldRecursively processes slice fields that may contain multiple references.
// Creates bidirectional edges (forward and reverse) for each nested entity.
func (gb *Builder) processSliceFieldRecursively(ctx context.Context, sourceNode *Node, field reflect.Value, fieldName string) ([]*Edge, error) {
	var edges []*Edge

	for i := 0; i < field.Len(); i++ {
		item := field.Index(i)
		if !item.IsValid() {
			continue
		}

		// Only check IsNil for types that support it (ptr, map, slice, chan, func, interface)
		// Scalar types like string, int, etc. will panic on IsNil
		if canBeNil(item) && item.IsNil() {
			continue
		}

		// Skip scalar types (string, int, etc.) - they don't represent entity references
		if !canBeNil(item) && item.Kind() != reflect.Struct {
			continue
		}

		// Extract edge properties before processing (for ServicePort, extract port_state)
		edgeProperties := gb.extractEdgeProperties(item.Interface())

		targetNode, err := gb.processNestedEntityRecursively(ctx, item.Interface())
		if err != nil {
			gb.logger.Warn("failed to process slice item", "field", fieldName, "index", i, "error", err)
			continue
		}

		if targetNode == nil {
			continue
		}

		// Create bidirectional edges
		edgeTypes := gb.getEdgeTypesForField(fieldName, sourceNode.NodeType)

		// Forward: Source BELONGS_TO Target
		edges = append(edges, &Edge{
			SourceNodeID: sourceNode.ID,
			TargetNodeID: targetNode.ID,
			EdgeType:     edgeTypes.Forward,
			Properties:   edgeProperties,
		})

		// Reverse: Target HAS Source
		edges = append(edges, &Edge{
			SourceNodeID: targetNode.ID,
			TargetNodeID: sourceNode.ID,
			EdgeType:     edgeTypes.Reverse,
			Properties:   edgeProperties,
		})
	}

	return edges, nil
}

// processNestedEntityRecursively attempts to process a nested entity recursively.
func (gb *Builder) processNestedEntityRecursively(ctx context.Context, fieldValue any) (*Node, error) {
	// Use generated function to create entity from field value
	entity := protograph.CreateEntityFromInterface(fieldValue)
	if entity == nil {
		return nil, nil
	}

	// If we successfully created an entity, recursively process it
	return gb.processEntityRecursively(ctx, entity)
}

// upsertNode creates or updates a node, incrementing duplicate count only once per ingestion request
func (gb *Builder) upsertNode(ctx context.Context, externalID, nodeType string, fullEntityData json.RawMessage, entity *diodepb.Entity, contentHash string) (*Node, error) {
	// Extract matching attributes from full entity data
	matchingData, err := gb.extractMatchingAttributes(nodeType, fullEntityData)
	if err != nil {
		gb.logger.Warn("failed to extract matching attributes, using full data", "error", err, "node_type", nodeType)
		matchingData = fullEntityData
	}

	// Extract metadata from entity and source
	metadata, err := gb.extractMetadata(entity)
	if err != nil {
		gb.logger.Warn("failed to extract metadata", "error", err, "node_type", nodeType)
		metadata = json.RawMessage("{}")
	}

	// Ensure source_match.diode_id is set to externalID for future lookups
	metadata = gb.ensureDiodeID(metadata, externalID)

	// Create a unique key for this node in this request
	requestKey := fmt.Sprintf("%s:%s", nodeType, externalID)

	// Check if we've already processed this node in this ingestion request
	alreadySeenInRequest := gb.seenInThisRequest[requestKey]

	contentHashPtr := ptrIfNonEmpty(contentHash)

	var result Node

	if alreadySeenInRequest {
		// This node was already seen in this request - update data but don't increment duplicate count
		result, err = gb.repo.UpdateNodeData(ctx, UpdateNodeDataParams{
			NodeType:              nodeType,
			ExternalID:            externalID,
			Data:                  matchingData,
			MatchingSchemaVersion: CurrentSchemaVersion,
			Metadata:              metadata,
			ContentHash:           contentHashPtr,
		})
		if err != nil {
			return nil, fmt.Errorf("updating node %s/%s: %w", nodeType, externalID, err)
		}
	} else {
		// First time seeing this node in this request - normal upsert with duplicate count increment
		result, err = gb.repo.UpsertNode(ctx, UpsertNodeParams{
			ExternalID:            externalID,
			NodeType:              nodeType,
			Data:                  matchingData,
			MatchingSchemaVersion: CurrentSchemaVersion,
			Metadata:              metadata,
			ContentHash:           contentHashPtr,
		})
		if err != nil {
			return nil, fmt.Errorf("upserting node %s/%s: %w", nodeType, externalID, err)
		}
		// Mark this node as seen in this request
		gb.seenInThisRequest[requestKey] = true
	}

	// Create snapshot with full entity data
	err = gb.createSnapshot(ctx, result.ID, fullEntityData)
	if err != nil {
		gb.logger.Warn("failed to create entity snapshot", "error", err, "node_id", result.ID)
		// Don't fail the entire operation for snapshot issues
	}

	return &Node{
		ID:                    result.ID,
		ExternalID:            result.ExternalID,
		NodeType:              result.NodeType,
		Data:                  result.Data,
		DuplicateCount:        result.DuplicateCount,
		MatchingSchemaVersion: result.MatchingSchemaVersion,
		Metadata:              result.Metadata,
	}, nil
}

// findTargetNodeFromField attempts to find or create a target node from a field value
func (gb *Builder) findTargetNodeFromField(ctx context.Context, fieldValue any, fieldName string) (*Node, error) {
	// Use generated function to create entity from field value
	entity := protograph.CreateEntityFromInterface(fieldValue)

	if entity != nil {
		// Use comprehensive functions if we successfully created an entity
		nodeType := getEntityTypeName(entity)
		if nodeType == "" {
			return nil, fmt.Errorf("failed to get object type for nested entity")
		}

		fingerprinter := entityhash.NewEntityFingerprinter()
		externalID, err := fingerprinter.GenerateEntityHash(entity)
		if err != nil {
			return nil, err
		}

		if externalID == "" {
			return nil, nil
		}

		// Try to find existing node
		return gb.findNodeByTypeAndID(ctx, nodeType, externalID)
	}

	// Fallback to the old approach if entity type not recognized
	if hasName, ok := fieldValue.(interface{ GetName() string }); ok {
		name := hasName.GetName()
		if name == "" {
			return nil, nil
		}

		// Determine node type from field name
		nodeType := gb.getNodeTypeFromFieldName(fieldName)

		// Try to find existing node
		return gb.findNodeByTypeAndID(ctx, nodeType, name)
	}

	return nil, nil
}

// findNodeByTypeAndID looks up an existing node by type and external ID
func (gb *Builder) findNodeByTypeAndID(ctx context.Context, nodeType, externalID string) (*Node, error) {
	result, err := gb.repo.FindNode(ctx, FindNodeParams{
		NodeType:   nodeType,
		ExternalID: externalID,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("finding node %s/%s: %w", nodeType, externalID, err)
	}

	return &result, nil
}

// upsertEdge creates or updates an edge
func (gb *Builder) upsertEdge(ctx context.Context, edge *Edge) error {
	return gb.repo.UpsertEdge(ctx, UpsertEdgeParams{
		SourceNodeID: edge.SourceNodeID,
		TargetNodeID: edge.TargetNodeID,
		EdgeType:     edge.EdgeType,
		Properties:   edge.Properties,
	})
}

// canBeNil returns true if the reflect.Value's kind supports nil checks.
// Calling IsNil on other types (string, int, struct, etc.) will panic.
func canBeNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return true
	}
	return false
}

// getEntityTypeName extracts the entity type name from an entity using proto names
func getEntityTypeName(entity *diodepb.Entity) string {
	if entity == nil || entity.GetEntity() == nil {
		return ""
	}

	// Simple approach using reflection to get the type name
	entityWrapper := entity.GetEntity()
	entityType := reflect.TypeOf(entityWrapper)
	if entityType == nil {
		return ""
	}

	// Extract the type name from the wrapper (e.g., "Entity_Device" -> "Device")
	typeName := entityType.Elem().Name()
	if name, found := strings.CutPrefix(typeName, "Entity_"); found {
		return name
	}

	return typeName
}

// getEdgeTypesForField returns both forward and reverse edge types for bidirectional relationships.
// Forward: BELONGS_TO_X (source references target)
// Reverse: HAS_X (target contains source)
func (gb *Builder) getEdgeTypesForField(fieldName, sourceEntityType string) protograph.EdgeTypePair {
	return protograph.GetEdgeTypesForField(fieldName, sourceEntityType)
}

// getNodeTypeFromFieldName returns the node type (entity type name) for a given field name using generated mappings.
func (gb *Builder) getNodeTypeFromFieldName(fieldName string) string {
	return protograph.GetNodeTypeForField(fieldName)
}

// propagateNodeUpdates finds nodes that reference updated nodes and refreshes them with the latest data
func (gb *Builder) propagateNodeUpdates(ctx context.Context) error {
	if len(gb.updatedNodes) == 0 {
		return nil
	}

	// For each node in the cache, check if it contains references to updated nodes
	for cacheKey, cachedNode := range gb.nodeCache {
		// Check if this node references any updated nodes by examining its data
		// (Note: we check ALL nodes, even if they were updated themselves, because they
		// might contain references to OTHER updated nodes that need refreshing)
		needsUpdate, err := gb.nodeReferencesUpdatedNodes(cachedNode, cacheKey)
		if err != nil {
			gb.logger.Warn("failed to check node references", "error", err, "node_type", cachedNode.NodeType)
			continue
		}

		if needsUpdate {
			// Regenerate the node data with updated references
			err = gb.refreshNodeWithUpdatedReferences(ctx, cachedNode)
			if err != nil {
				gb.logger.Warn("failed to refresh node with updated references",
					"error", err,
					"node_type", cachedNode.NodeType,
					"external_id", cachedNode.ExternalID)
			}
		}
	}

	return nil
}

// nodeReferencesUpdatedNodes checks if a node contains references to any updated nodes
func (gb *Builder) nodeReferencesUpdatedNodes(node *Node, nodeKey string) (bool, error) {
	// Parse the node's JSON data to look for references
	var nodeData map[string]any
	if err := json.Unmarshal(node.Data, &nodeData); err != nil {
		return false, fmt.Errorf("failed to parse node data: %w", err)
	}

	// Recursively check for references to updated node types (excluding this node itself)
	return gb.checkDataForUpdatedReferences(nodeData, nodeKey), nil
}

// checkDataForUpdatedReferences recursively checks if data contains references to updated nodes
func (gb *Builder) checkDataForUpdatedReferences(data any, excludeNodeKey string) bool {
	switch v := data.(type) {
	case map[string]any:
		// Check if this map represents an entity that might be updated
		if name, hasName := v["name"]; hasName {
			if nameStr, ok := name.(string); ok {
				// Check if any updated node has this name (excluding the node being checked itself)
				for updatedKey, updatedNode := range gb.updatedNodes {
					// Skip checking the node against itself
					if updatedKey == excludeNodeKey {
						continue
					}
					var updatedData map[string]any
					if json.Unmarshal(updatedNode.Data, &updatedData) == nil {
						// Look for the name in the nested entity data structure
						if entityData, hasEntity := updatedData[strings.ToLower(updatedNode.NodeType)]; hasEntity {
							if entityMap, ok := entityData.(map[string]any); ok {
								if updatedName, ok := entityMap["name"]; ok && updatedName == nameStr {
									// Check if the node types could be related (e.g., Manufacturer)
									if gb.couldBeRelatedNodeType(updatedNode.NodeType) {
										return true
									}
								}
							}
						}
					}
				}
			}
		}

		// Recursively check nested objects
		for _, value := range v {
			if gb.checkDataForUpdatedReferences(value, excludeNodeKey) {
				return true
			}
		}
	case []any:
		// Check array elements
		for _, item := range v {
			if gb.checkDataForUpdatedReferences(item, excludeNodeKey) {
				return true
			}
		}
	}
	return false
}

// couldBeRelatedNodeType checks if an updated node type could be referenced by the given data.
func (gb *Builder) couldBeRelatedNodeType(updatedNodeType string) bool {
	// The logic here is: if we updated a node of type X, and this referenceData
	// represents a reference to type X (by having the same name), then it's related

	// For now, we'll use a simple heuristic: if the reference data looks like
	// it could represent the same type of entity (has similar structure)
	switch updatedNodeType {
	case "Manufacturer":
		// Check if this reference data represents a manufacturer reference
		// (it should have manufacturer-like fields)
		return true // For manufacturer, any named reference could potentially be a manufacturer
	case "DeviceType":
		// Check if this could be a device type reference
		return true
	case "Platform":
		// Check if this could be a platform reference
		return true
	case "Site":
		// Check if this could be a site reference
		return true
	}

	// Default to true for now - we can make this more sophisticated later
	return true
}

// refreshNodeWithUpdatedReferences updates a node in the database with fresh data that includes updated references
func (gb *Builder) refreshNodeWithUpdatedReferences(ctx context.Context, node *Node) error {
	// Find the existing node in the database to get its latest data
	existingNode, err := gb.repo.FindNode(ctx, FindNodeParams{
		NodeType:   node.NodeType,
		ExternalID: node.ExternalID,
	})
	if err != nil {
		return fmt.Errorf("failed to find existing node: %w", err)
	}

	// Parse the existing data to reconstruct the entity
	var nodeData map[string]any
	if err := json.Unmarshal(existingNode.Data, &nodeData); err != nil {
		return fmt.Errorf("failed to parse existing node data: %w", err)
	}

	// Update references in the node data with the latest information from updated nodes
	gb.updateReferencesInNodeData(nodeData)

	// Re-serialize the updated data
	updatedData, err := json.Marshal(nodeData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated node data: %w", err)
	}

	// Keep existing metadata (reference updates don't change metadata)
	metadata := existingNode.Metadata
	if metadata == nil {
		metadata = json.RawMessage("{}")
	}

	// Update the node in the database using request-level duplicate counting
	requestKey := fmt.Sprintf("%s:%s", existingNode.NodeType, existingNode.ExternalID)

	if gb.seenInThisRequest[requestKey] {
		// Already seen in this request - update data but don't increment duplicate count
		_, err = gb.repo.UpdateNodeData(ctx, UpdateNodeDataParams{
			NodeType:              existingNode.NodeType,
			ExternalID:            existingNode.ExternalID,
			Data:                  updatedData,
			MatchingSchemaVersion: existingNode.MatchingSchemaVersion, // Keep existing schema version for reference updates
			Metadata:              metadata,
		})
	} else {
		// First time seeing this node in this request - increment duplicate count
		_, err = gb.repo.UpsertNode(ctx, UpsertNodeParams{
			ExternalID:            existingNode.ExternalID,
			NodeType:              existingNode.NodeType,
			Data:                  updatedData,
			MatchingSchemaVersion: existingNode.MatchingSchemaVersion, // Keep existing schema version for reference updates
			Metadata:              metadata,
		})
		gb.seenInThisRequest[requestKey] = true
	}
	if err != nil {
		return fmt.Errorf("failed to update refreshed node: %w", err)
	}

	return nil
}

// updateReferencesInNodeData recursively updates references in node data with latest information from updated nodes
func (gb *Builder) updateReferencesInNodeData(data map[string]any) {
	for _, value := range data {
		switch v := value.(type) {
		case map[string]any:
			// Check if this nested object represents a reference that should be updated
			if name, hasName := v["name"]; hasName {
				if nameStr, ok := name.(string); ok {
					// Find matching updated node and update the reference
					for _, updatedNode := range gb.updatedNodes {
						var updatedData map[string]any
						if json.Unmarshal(updatedNode.Data, &updatedData) == nil {
							// Look for the name in the nested entity data structure
							if entityData, hasEntity := updatedData[strings.ToLower(updatedNode.NodeType)]; hasEntity {
								if entityMap, ok := entityData.(map[string]any); ok {
									if updatedName, ok := entityMap["name"]; ok && updatedName == nameStr {
										// Update this reference with the latest entity data
										maps.Copy(v, entityMap)
									}
								}
							}
						}
					}
				}
			}
			// Recursively update nested references
			gb.updateReferencesInNodeData(v)
		case []any:
			// Handle array of references
			for _, item := range v {
				if itemMap, ok := item.(map[string]any); ok {
					gb.updateReferencesInNodeData(itemMap)
				}
			}
		}
	}
}
