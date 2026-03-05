package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
)

// parsedUpdate holds pre-parsed data for an updated node to avoid repeated JSON unmarshaling.
type parsedUpdate struct {
	key       string
	nodeType  string
	entityMap map[string]any // the inner entity data keyed by lowercase node type
}

// parseUpdatedNodes pre-parses all updated node data once.
func (s *Service) parseUpdatedNodes() []parsedUpdate {
	parsed := make([]parsedUpdate, 0, len(s.updatedNodes))
	for key, node := range s.updatedNodes {
		var data map[string]any
		if json.Unmarshal(node.Data, &data) != nil {
			continue
		}
		entityData, ok := data[strings.ToLower(node.NodeType)]
		if !ok {
			continue
		}
		entityMap, ok := entityData.(map[string]any)
		if !ok {
			continue
		}
		parsed = append(parsed, parsedUpdate{
			key:       key,
			nodeType:  node.NodeType,
			entityMap: entityMap,
		})
	}
	return parsed
}

// propagateNodeUpdates finds nodes that reference updated nodes and refreshes them with the latest data.
func (s *Service) propagateNodeUpdates(ctx context.Context) error {
	if len(s.updatedNodes) == 0 {
		return nil
	}

	// Pre-parse updated node data once instead of on every recursive check.
	updates := s.parseUpdatedNodes()

	// For each node in the cache, check if it contains references to updated nodes
	for cacheKey, cachedNode := range s.nodeCache {
		// Check if this node references any updated nodes by examining its data
		// (Note: we check ALL nodes, even if they were updated themselves, because they
		// might contain references to OTHER updated nodes that need refreshing)
		needsUpdate, err := nodeReferencesUpdated(cachedNode, cacheKey, updates)
		if err != nil {
			s.logger.Warn("failed to check node references", "error", err, "node_type", cachedNode.NodeType)
			continue
		}

		if needsUpdate {
			// Regenerate the node data with updated references
			err = s.refreshNodeWithUpdatedReferences(ctx, cachedNode, updates)
			if err != nil {
				s.logger.Warn("failed to refresh node with updated references",
					"error", err,
					"node_type", cachedNode.NodeType,
					"external_id", cachedNode.ExternalID)
			}
		}
	}

	return nil
}

// nodeReferencesUpdated checks if a node contains references to any updated nodes.
func nodeReferencesUpdated(node *Node, nodeKey string, updates []parsedUpdate) (bool, error) {
	var nodeData map[string]any
	if err := json.Unmarshal(node.Data, &nodeData); err != nil {
		return false, fmt.Errorf("parsing node data: %w", err)
	}
	return checkForUpdatedRefs(nodeData, nodeKey, updates), nil
}

// checkForUpdatedRefs recursively checks if data contains references to updated nodes.
func checkForUpdatedRefs(data any, excludeNodeKey string, updates []parsedUpdate) bool {
	switch v := data.(type) {
	case map[string]any:
		if name, hasName := v["name"]; hasName {
			if nameStr, ok := name.(string); ok {
				for _, u := range updates {
					if u.key == excludeNodeKey {
						continue
					}
					if updatedName, ok := u.entityMap["name"]; ok && updatedName == nameStr {
						return true
					}
				}
			}
		}
		for _, value := range v {
			if checkForUpdatedRefs(value, excludeNodeKey, updates) {
				return true
			}
		}
	case []any:
		for _, item := range v {
			if checkForUpdatedRefs(item, excludeNodeKey, updates) {
				return true
			}
		}
	}
	return false
}

// refreshNodeWithUpdatedReferences updates a node in the database with fresh data that includes updated references.
func (s *Service) refreshNodeWithUpdatedReferences(ctx context.Context, node *Node, updates []parsedUpdate) error {
	existingNode, err := s.repo.FindNode(ctx, FindNodeParams{
		NodeType:   node.NodeType,
		ExternalID: node.ExternalID,
	})
	if err != nil {
		return fmt.Errorf("finding existing node: %w", err)
	}

	var nodeData map[string]any
	if err := json.Unmarshal(existingNode.Data, &nodeData); err != nil {
		return fmt.Errorf("parsing existing node data: %w", err)
	}

	updateRefsInData(nodeData, updates)

	updatedData, err := json.Marshal(nodeData)
	if err != nil {
		return fmt.Errorf("marshaling updated node data: %w", err)
	}

	metadata := existingNode.Metadata
	if metadata == nil {
		metadata = json.RawMessage("{}")
	}

	requestKey := fmt.Sprintf("%s:%s", existingNode.NodeType, existingNode.ExternalID)

	if s.seenInThisRequest[requestKey] {
		_, err = s.repo.UpdateNodeData(ctx, UpdateNodeDataParams{
			NodeType:   existingNode.NodeType,
			ExternalID: existingNode.ExternalID,
			Data:       updatedData,
			Metadata:   metadata,
		})
	} else {
		_, err = s.repo.UpsertNode(ctx, UpsertNodeParams{
			ExternalID: existingNode.ExternalID,
			NodeType:   existingNode.NodeType,
			Data:       updatedData,
			Metadata:   metadata,
		})
		s.seenInThisRequest[requestKey] = true
	}
	if err != nil {
		return fmt.Errorf("updating refreshed node: %w", err)
	}

	return nil
}

// updateRefsInData recursively updates references in node data with latest information from updated nodes.
func updateRefsInData(data map[string]any, updates []parsedUpdate) {
	for _, value := range data {
		switch v := value.(type) {
		case map[string]any:
			matched := false
			if name, hasName := v["name"]; hasName {
				if nameStr, ok := name.(string); ok {
					for _, u := range updates {
						if updatedName, ok := u.entityMap["name"]; ok && updatedName == nameStr {
							maps.Copy(v, u.entityMap)
							matched = true
						}
					}
				}
			}
			// Skip recursion into maps that were just updated via maps.Copy
			// to prevent infinite recursion: the copied entityMap may contain
			// nested maps with matching names that would trigger the same
			// copy indefinitely.
			if !matched {
				updateRefsInData(v, updates)
			}
		case []any:
			for _, item := range v {
				if itemMap, ok := item.(map[string]any); ok {
					updateRefsInData(itemMap, updates)
				}
			}
		}
	}
}
