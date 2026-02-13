package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/netboxlabs/diode/diode-server/entityhash"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/protograph"
	"github.com/netboxlabs/diode/diode-server/matching"
)

const (
	// sourceMatchKey is the metadata key containing correlation IDs for entity matching
	sourceMatchKey = "source_match"
	// diodeIDKey is the internal Diode entity ID - highest priority for matching
	diodeIDKey = "diode_id"
)

// findEntityMatch attempts to find an existing entity match.
// Priority order:
// 1. Metadata match (correlation IDs) - works without entity matcher config
// 2. Field-based matching via entityMatcher (requires config)
// 3. Content hash fallback - last resort when entity matcher config is missing
// Returns nil if no match found.
func (s *Service) findEntityMatch(ctx context.Context, entity *diodepb.Entity, nodeType string, contentHash string) *matching.MatchResult {
	// Priority 1: Check metadata for correlation IDs (works without entity matcher)
	if metadataMatch := s.findMatchByMetadata(ctx, entity, nodeType); metadataMatch != nil {
		return metadataMatch
	}

	// Priority 2: Fall back to field-based matching if entity matcher is configured
	if s.entityMatcher != nil {
		bestMatch, err := s.entityMatcher.FindBestMatch(ctx, entity)
		if err != nil {
			s.logger.Warn("failed to find entity matches", "error", err, "entity_type", nodeType)
		} else if bestMatch != nil && bestMatch.NodeID != nil {
			s.logger.Debug("found confident match for entity",
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
		if hashMatch := s.findNodeByContentHash(ctx, nodeType, contentHash); hashMatch != nil {
			return hashMatch
		}
	}

	return nil
}

// findMatchByMetadata attempts to find an existing node by checking the source_match metadata.
// Priority order:
// 1. source_match.diode_id - if found, return immediately (internal Diode entity)
// 2. Other source_match.* keys - iterate and return on first match found
// Returns a match result with confidence 1.0 if found, nil if no match.
func (s *Service) findMatchByMetadata(ctx context.Context, entity *diodepb.Entity, nodeType string) *matching.MatchResult {
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
			if result := s.findNodeByExternalID(ctx, nodeType, diodeIDStr); result != nil {
				s.logger.Debug("found diode_id match for entity",
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

		if result := s.findNodeBySourceMatchKey(ctx, nodeType, key, value); result != nil {
			s.logger.Debug("found metadata match for entity",
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
func (s *Service) findNodeBySourceMatchKey(ctx context.Context, nodeType, key string, value any) *matching.MatchResult {
	// Build filter: {"source_match": {key: value}}
	metadataFilter := map[string]any{
		sourceMatchKey: map[string]any{
			key: value,
		},
	}
	filterJSON, err := json.Marshal(metadataFilter)
	if err != nil {
		s.logger.Warn("failed to marshal metadata filter", "error", err, "key", key)
		return nil
	}

	// Search for existing node with this metadata
	node, err := s.repo.FindNodeByMetadata(ctx, FindNodeByMetadataParams{
		NodeType:       nodeType,
		MetadataFilter: filterJSON,
	})
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			s.logger.Warn("metadata lookup failed", "error", err, "key", key, "value", value)
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
func (s *Service) findNodeByExternalID(ctx context.Context, nodeType, externalID string) *matching.MatchResult {
	node, err := s.repo.FindNode(ctx, FindNodeParams{
		NodeType:   nodeType,
		ExternalID: externalID,
	})
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			s.logger.Warn("externalID lookup failed", "error", err, "external_id", externalID)
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
func (s *Service) findNodeByContentHash(ctx context.Context, nodeType, contentHash string) *matching.MatchResult {
	node, err := s.repo.FindNodeByContentHash(ctx, FindNodeByContentHashParams{
		NodeType:    nodeType,
		ContentHash: contentHash,
	})
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			s.logger.Warn("content hash lookup failed", "error", err, "content_hash", contentHash)
		}
		return nil
	}

	s.logger.Debug("found content hash match for entity",
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

// findTargetNodeFromField attempts to find or create a target node from a field value
func (s *Service) findTargetNodeFromField(ctx context.Context, fieldValue any, fieldName string) (*Node, error) {
	// Use generated function to create entity from field value
	entity := protograph.CreateEntityFromInterface(fieldValue)

	if entity != nil {
		// Use comprehensive functions if we successfully created an entity
		nodeType := getEntityTypeName(entity)
		if nodeType == "" {
			return nil, fmt.Errorf("unknown nested entity type")
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
		return s.findNodeByTypeAndID(ctx, nodeType, externalID)
	}

	// Fallback to the old approach if entity type not recognized
	if hasName, ok := fieldValue.(interface{ GetName() string }); ok {
		name := hasName.GetName()
		if name == "" {
			return nil, nil
		}

		// Determine node type from field name
		nodeType := getNodeTypeFromFieldName(fieldName)

		// Try to find existing node
		return s.findNodeByTypeAndID(ctx, nodeType, name)
	}

	return nil, nil
}

// findNodeByTypeAndID looks up an existing node by type and external ID
func (s *Service) findNodeByTypeAndID(ctx context.Context, nodeType, externalID string) (*Node, error) {
	result, err := s.repo.FindNode(ctx, FindNodeParams{
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
