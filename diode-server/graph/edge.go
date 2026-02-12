package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/protograph"
	"github.com/netboxlabs/diode/diode-server/strcase"
)

// createEdgesForNode extracts and creates all edges for a node.
func (s *Service) createEdgesForNode(ctx context.Context, entity *diodepb.Entity, node *Node, nodeType, externalID string) {
	edges, err := s.extractEdgesRecursively(ctx, entity, node)
	if err != nil {
		s.logger.Warn("failed to extract edges recursively",
			"node_type", nodeType,
			"external_id", externalID,
			"error", err)
		return
	}

	for _, edge := range edges {
		if err := s.upsertEdge(ctx, edge); err != nil {
			s.logger.Warn("failed to create edge",
				"source_id", edge.SourceNodeID,
				"target_id", edge.TargetNodeID,
				"edge_type", edge.EdgeType,
				"error", err)
		}
	}
}

// extractEdgesRecursively finds all relationships within the entity and creates edges recursively
func (s *Service) extractEdgesRecursively(ctx context.Context, entity *diodepb.Entity, sourceNode *Node) ([]*Edge, error) {
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
			fieldEdges, err := s.processFieldRecursively(ctx, sourceNode, field, fieldName)
			if err != nil {
				s.logger.Warn("failed to process field recursively", "field", fieldName, "error", err)
				continue
			}
			if len(fieldEdges) > 0 {
				edges = append(edges, fieldEdges...)
			}
		}

		// Process slice fields (repeated references)
		if field.Kind() == reflect.Slice {
			sliceEdges, err := s.processSliceFieldRecursively(ctx, sourceNode, field, fieldName)
			if err != nil {
				s.logger.Warn("failed to process slice field recursively", "field", fieldName, "error", err)
				continue
			}
			edges = append(edges, sliceEdges...)
		}

		// Process oneof fields (interface types in generated Go code)
		if field.Kind() == reflect.Interface && !field.IsNil() {
			// Unwrap the interface to get the concrete value
			concreteValue := reflect.ValueOf(field.Interface())
			if concreteValue.IsValid() && concreteValue.Kind() == reflect.Ptr && !concreteValue.IsNil() {
				fieldEdges, err := s.processFieldRecursively(ctx, sourceNode, concreteValue, fieldName)
				if err != nil {
					s.logger.Warn("failed to process oneof field recursively", "field", fieldName, "error", err)
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
func (s *Service) extractEdgeProperties(fieldValue any) json.RawMessage {
	if fieldValue == nil {
		return json.RawMessage("{}")
	}

	// If no matching config, return empty properties
	if s.matchingConfig == nil {
		return json.RawMessage("{}")
	}

	// Get the entity type name from the field value
	entityTypeName := getEntityTypeNameFromValue(fieldValue)
	if entityTypeName == "" {
		return json.RawMessage("{}")
	}

	// Look up the matching rule for this entity type to get edge property fields
	entityRule, found := s.matchingConfig.DefaultEntityRules[entityTypeName]
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
func getEntityTypeNameFromValue(fieldValue any) string {
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
func (s *Service) processFieldRecursively(ctx context.Context, sourceNode *Node, field reflect.Value, fieldName string) ([]*Edge, error) {
	if !field.IsValid() || field.IsNil() {
		return nil, nil
	}

	fieldInterface := field.Interface()

	// Try to recursively process this field as a nested entity first
	targetNode, err := s.processNestedEntityRecursively(ctx, fieldInterface)
	if err != nil {
		return nil, err
	}

	// If recursive processing didn't work, try the fallback approach
	if targetNode == nil {
		targetNode, err = s.findTargetNodeFromField(ctx, fieldInterface, fieldName)
		if err != nil {
			return nil, err
		}
	}

	if targetNode == nil {
		return nil, nil
	}

	// Create bidirectional edges
	edgeTypes := getEdgeTypesForField(fieldName, sourceNode.NodeType)

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
func (s *Service) processSliceFieldRecursively(ctx context.Context, sourceNode *Node, field reflect.Value, fieldName string) ([]*Edge, error) {
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
		edgeProperties := s.extractEdgeProperties(item.Interface())

		targetNode, err := s.processNestedEntityRecursively(ctx, item.Interface())
		if err != nil {
			s.logger.Warn("failed to process slice item", "field", fieldName, "index", i, "error", err)
			continue
		}

		if targetNode == nil {
			continue
		}

		// Create bidirectional edges
		edgeTypes := getEdgeTypesForField(fieldName, sourceNode.NodeType)

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
func (s *Service) processNestedEntityRecursively(ctx context.Context, fieldValue any) (*Node, error) {
	// Use generated function to create entity from field value
	entity := protograph.CreateEntityFromInterface(fieldValue)
	if entity == nil {
		return nil, nil
	}

	// If we successfully created an entity, recursively process it
	return s.processEntityRecursively(ctx, entity)
}

// upsertEdge creates or updates an edge
func (s *Service) upsertEdge(ctx context.Context, edge *Edge) error {
	return s.repo.UpsertEdge(ctx, UpsertEdgeParams{
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
func getEdgeTypesForField(fieldName, sourceEntityType string) protograph.EdgeTypePair {
	return protograph.GetEdgeTypesForField(fieldName, sourceEntityType)
}

// getNodeTypeFromFieldName returns the node type (entity type name) for a given field name using generated mappings.
func getNodeTypeFromFieldName(fieldName string) string {
	return protograph.GetNodeTypeForField(fieldName)
}
