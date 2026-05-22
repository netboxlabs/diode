package graph

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/netboxlabs/diode/diode-server/matching"
)

// extractMatchingAttributes extracts only the attributes needed for matching from the full entity data.
func (s *Service) extractMatchingAttributes(entityType string, fullEntityData json.RawMessage) (json.RawMessage, error) {
	// If no matching config is set, return the full entity data as fallback
	if s.matchingConfig == nil {
		return fullEntityData, nil
	}

	// Get matching rules for this entity type
	var entityRule matching.EntityMatchingRule
	var found bool

	// Try to get entity-specific rule first
	if entityRule, found = s.matchingConfig.DefaultEntityRules[entityType]; !found {
		// Fall back to wildcard rule
		if entityRule, found = s.matchingConfig.DefaultEntityRules["*"]; !found {
			// No matching rules found, return full data
			return fullEntityData, nil
		}
	}

	// Parse the full entity data
	var fullData map[string]any
	if err := json.Unmarshal(fullEntityData, &fullData); err != nil {
		return nil, fmt.Errorf("parsing entity data: %w", err)
	}

	// Extract only the fields used in matching rules
	matchingData := make(map[string]any)

	// Process primary rules
	for _, rule := range entityRule.PrimaryRules {
		if value := s.extractFieldByPath(fullData, rule.FieldPath); value != nil {
			s.setFieldByPath(matchingData, rule.FieldPath, value)
		}
	}

	// Process secondary rules
	for _, rule := range entityRule.SecondaryRules {
		if value := s.extractFieldByPath(fullData, rule.FieldPath); value != nil {
			s.setFieldByPath(matchingData, rule.FieldPath, value)
		}
	}

	// If no matching data was extracted (empty object), fall back to full entity data
	if len(matchingData) == 0 {
		return fullEntityData, nil
	}

	// Marshal the extracted matching data back to JSON
	matchingJSON, err := json.Marshal(matchingData)
	if err != nil {
		return nil, fmt.Errorf("marshaling matching data: %w", err)
	}

	return matchingJSON, nil
}

// extractFieldByPath extracts a value from nested data using a field path like "Device.name" or "Device.site.name".
func (s *Service) extractFieldByPath(data map[string]any, fieldPath string) any {
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

// setFieldByPath sets a value in nested data using a field path like "Device.name" or "Device.site.name".
func (s *Service) setFieldByPath(data map[string]any, fieldPath string, value any) {
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
