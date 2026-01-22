package entitymatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/protograph"
	"github.com/netboxlabs/diode/diode-server/matching"
	"github.com/netboxlabs/diode/diode-server/reconciler"
)

// Matcher implements confidence-based entity matching
type Matcher struct {
	repo         reconciler.GraphRepository
	config       *matching.EntityMatchingConfig
	logger       *slog.Logger
	fuzzyMatcher *matching.FuzzyMatcher

	// Cache for matching results
	cache      map[string][]matching.MatchResult
	cacheMutex sync.RWMutex
}

// NewMatcher creates a new EntityMatcher with the given configuration
func NewMatcher(repo reconciler.GraphRepository, config *matching.EntityMatchingConfig, logger *slog.Logger) *Matcher {
	if config == nil {
		config = DefaultEntityMatchingConfig()
	}

	return &Matcher{
		repo:         repo,
		config:       config,
		logger:       logger,
		fuzzyMatcher: matching.NewFuzzyMatcher(),
		cache:        make(map[string][]matching.MatchResult),
	}
}

// DefaultEntityMatchingConfig returns the default matching configuration
// Note: This is deprecated - use YAML configuration instead
func DefaultEntityMatchingConfig() *matching.EntityMatchingConfig {
	return &matching.EntityMatchingConfig{
		Rules:          make(map[string]*matching.EntityMatchingRule), // Empty - use YAML config instead
		GlobalMinConf:  matching.ConfidenceLow,
		EnableFallback: true,
		CacheResults:   true,
		MaxCacheSize:   1000,
	}
}

// getEntityTypeName extracts the type name from a diodepb.Entity wrapper
func getEntityTypeName(entity *diodepb.Entity) string {
	if entity == nil || entity.Entity == nil {
		return ""
	}

	// Extract the inner value from the Entity oneof
	innerValue := reflect.ValueOf(entity.Entity)
	if innerValue.Kind() == reflect.Ptr || innerValue.Kind() == reflect.Interface {
		innerValue = innerValue.Elem()
	}

	// Get the actual entity value from the oneof struct
	if innerValue.Kind() == reflect.Struct && innerValue.NumField() > 0 {
		field := innerValue.Field(0)
		if field.CanInterface() {
			return protograph.GetEntityTypeName(field.Interface())
		}
	}

	return ""
}

// extractFieldValue extracts a field value from entity data using a JSON path
func extractFieldValue(data interface{}, fieldPath string) (interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("data is nil")
	}

	parts := strings.Split(fieldPath, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			// Try exact match first, then case-insensitive
			if val, exists := v[part]; exists {
				current = val
			} else {
				// Try case-insensitive match
				found := false
				for key, val := range v {
					if strings.EqualFold(key, part) {
						current = val
						found = true
						break
					}
				}
				if !found {
					current = nil
				}
			}
		case *map[string]interface{}:
			if v != nil {
				// Try exact match first, then case-insensitive
				if val, exists := (*v)[part]; exists {
					current = val
				} else {
					// Try case-insensitive match
					found := false
					for key, val := range *v {
						if strings.EqualFold(key, part) {
							current = val
							found = true
							break
						}
					}
					if !found {
						current = nil
					}
				}
			} else {
				current = nil
			}
		default:
			// Use reflection for struct field access
			val := reflect.ValueOf(current)
			if val.Kind() == reflect.Ptr {
				if val.IsNil() {
					return nil, nil
				}
				val = val.Elem()
			}

			if val.Kind() != reflect.Struct {
				return nil, fmt.Errorf("cannot navigate field path %s: not a struct at %s", fieldPath, part)
			}

			field := val.FieldByName(strings.ToUpper(part[:1]) + part[1:])
			if !field.IsValid() {
				// Try with the original case
				field = val.FieldByName(part)
				if !field.IsValid() {
					return nil, nil
				}
			}

			current = field.Interface()
		}

		if current == nil {
			return nil, nil
		}
	}

	return current, nil
}

// FindMatches finds potential matches for an entity with confidence scores
func (m *Matcher) FindMatches(ctx context.Context, entity *diodepb.Entity) ([]matching.MatchResult, error) {
	if entity == nil {
		return nil, fmt.Errorf("entity is nil")
	}

	// Get entity type - use proto names for consistency with YAML config
	entityType := getEntityTypeName(entity)
	if entityType == "" {
		return nil, fmt.Errorf("unable to determine entity type")
	}

	// Check cache first
	if m.config.CacheResults {
		if cached := m.getCachedResults(entity, entityType); cached != nil {
			return cached, nil
		}
	}

	// Get matching rule for this entity type
	rule, exists := m.config.Rules[entityType]
	if !exists {
		m.logger.Debug("no matching rule found for entity type", "entity_type", entityType)
		return []matching.MatchResult{}, nil
	}

	// Search for potential matches in database
	candidates, err := m.findCandidateNodes(ctx, entity, entityType, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to find candidate nodes: %w", err)
	}

	// Score each candidate
	var results []matching.MatchResult
	for _, candidate := range candidates {
		result, err := m.scoreMatch(entity, candidate, rule)
		if err != nil {
			m.logger.Warn("failed to score match candidate", "error", err, "candidate_id", candidate.ID)
			continue
		}

		if result.Confidence >= rule.MinConfidence {
			results = append(results, *result)
		}
	}

	// Cache results if enabled
	if m.config.CacheResults {
		m.setCachedResults(entity, entityType, results)
	}

	return results, nil
}

// FindBestMatch finds the best match for an entity above the confidence threshold
func (m *Matcher) FindBestMatch(ctx context.Context, entity *diodepb.Entity) (*matching.MatchResult, error) {
	matches, err := m.FindMatches(ctx, entity)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, nil
	}

	// Find the highest confidence match
	best := &matches[0]
	for i := 1; i < len(matches); i++ {
		if matches[i].Confidence > best.Confidence {
			best = &matches[i]
		}
	}

	// Only return if above global minimum confidence
	if best.Confidence >= m.config.GlobalMinConf {
		return best, nil
	}

	return nil, nil
}

// GetMatchingRule returns the matching rule for a specific entity type
func (m *Matcher) GetMatchingRule(entityType string) (*matching.EntityMatchingRule, error) {
	rule, exists := m.config.Rules[entityType]
	if !exists {
		return nil, fmt.Errorf("no matching rule found for entity type: %s", entityType)
	}
	return rule, nil
}

// UpdateMatchingRule updates or adds a matching rule for an entity type
func (m *Matcher) UpdateMatchingRule(entityType string, rule *matching.EntityMatchingRule) error {
	if rule == nil {
		return fmt.Errorf("rule cannot be nil")
	}

	m.config.Rules[entityType] = rule

	// Clear cache for this entity type
	if m.config.CacheResults {
		m.clearCacheForEntityType(entityType)
	}

	return nil
}

// findCandidateNodes searches the database for potential matching nodes
func (m *Matcher) findCandidateNodes(ctx context.Context, entity *diodepb.Entity, entityType string, rule *matching.EntityMatchingRule) ([]*reconciler.GraphNode, error) {
	// Extract entity data for comparison
	entityData, err := m.entityToMap(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to convert entity to map: %w", err)
	}

	m.logger.Debug("searching for candidate nodes",
		"entity_type", entityType,
		"primary_rules", len(rule.PrimaryRules),
		"secondary_rules", len(rule.SecondaryRules),
		"fallback_rules", len(rule.FallbackRules))

	var allCandidates []*reconciler.GraphNode

	// Strategy 1: Try exact matches on primary fields first
	primaryCandidates, err := m.findCandidatesByPrimaryFields(ctx, entityType, entityData, rule.PrimaryRules)
	if err != nil {
		m.logger.Error("failed to search by primary fields", "error", err)
	} else {
		allCandidates = append(allCandidates, primaryCandidates...)
		if len(primaryCandidates) > 0 {
			// If we found good primary matches, we can be more selective
			m.logger.Debug("found primary field matches", "count", len(primaryCandidates))
			return allCandidates, nil
		}
	}

	// Strategy 2: Try secondary field matches if no primary matches
	secondaryCandidates, err := m.findCandidatesBySecondaryFields(ctx, entityType, entityData, rule.SecondaryRules)
	if err != nil {
		m.logger.Error("failed to search by secondary fields", "error", err)
	} else {
		allCandidates = append(allCandidates, secondaryCandidates...)
	}

	// Strategy 3: Fallback to broader search if still no good matches
	if len(allCandidates) < 5 && len(rule.FallbackRules) > 0 {
		fallbackCandidates, err := m.findCandidatesByFallbackStrategy(ctx, entityType, entityData, rule.FallbackRules)
		if err != nil {
			m.logger.Error("failed to search by fallback strategy", "error", err)
		} else {
			allCandidates = append(allCandidates, fallbackCandidates...)
		}
	}

	// Remove duplicates and limit results
	candidates := m.deduplicateCandidates(allCandidates)
	if len(candidates) > 50 {
		// Limit to top 50 candidates to avoid performance issues
		candidates = candidates[:50]
	}

	m.logger.Debug("found candidate nodes", "total", len(candidates))
	return candidates, nil
}

// findCandidatesByPrimaryFields searches using primary matching fields (supports both exact and fuzzy matches)
func (m *Matcher) findCandidatesByPrimaryFields(ctx context.Context, entityType string, entityData map[string]interface{}, primaryRules []matching.FieldMatchRule) ([]*reconciler.GraphNode, error) {
	var candidates []*reconciler.GraphNode

	for _, rule := range primaryRules {
		// Extract field value from entity data
		fieldValue, exists := m.extractFieldFromMap(entityData, rule.FieldPath)
		if !exists || fieldValue == nil {
			if rule.Required {
				m.logger.Debug("required primary field missing", "field", rule.FieldPath)
				continue
			}
			continue
		}

		valueStr := fmt.Sprintf("%v", fieldValue)
		if valueStr == "" {
			continue
		}

		var dbNodes []postgres.GraphNode
		var err error

		if rule.MatchType == matching.MatchExact {
			// Use exact field match query for exact matches
			params := postgres.FindNodesByFieldMatchParams{
				NodeType:   entityType,
				JsonField:  pgtype.Text{String: rule.FieldPath, Valid: true},
				FieldValue: pgtype.Text{String: valueStr, Valid: true},
				Limit:      10,
				Offset:     0,
			}

			dbNodes, err = m.repo.FindNodesByFieldMatch(ctx, params)
			if err != nil {
				m.logger.Error("exact field match query failed", "field", rule.FieldPath, "error", err)
				continue
			}

			m.logger.Debug("found exact primary field matches",
				"field", rule.FieldPath, "value", valueStr, "count", len(dbNodes))

		} else if rule.MatchType == matching.MatchFuzzy {
			// For fuzzy matches, get all nodes of this type and perform fuzzy matching
			fuzzyNodes, err := m.findFuzzyCandidatesForField(ctx, entityType, rule.FieldPath, valueStr, rule.FuzzyOptions)
			if err != nil {
				m.logger.Error("fuzzy field match query failed", "field", rule.FieldPath, "error", err)
				continue
			}

			dbNodes = fuzzyNodes
			m.logger.Debug("found fuzzy primary field matches",
				"field", rule.FieldPath, "value", valueStr, "count", len(dbNodes))
		}

		// Convert to GraphNode format
		for _, dbNode := range dbNodes {
			candidates = append(candidates, &reconciler.GraphNode{
				ID:             dbNode.ID,
				ExternalID:     dbNode.ExternalID,
				NodeType:       dbNode.NodeType,
				Data:           dbNode.Data,
				DuplicateCount: dbNode.DuplicateCount,
			})
		}

		// If we found matches on a required primary field, prioritize these
		if len(dbNodes) > 0 && rule.Required {
			break // Use first good required primary field match
		}
	}

	return candidates, nil
}

// findFuzzyCandidatesForField performs fuzzy matching for a specific field by fetching all nodes and comparing
func (m *Matcher) findFuzzyCandidatesForField(ctx context.Context, entityType, fieldPath, searchValue string, fuzzyOptions *matching.FuzzyOptions) ([]postgres.GraphNode, error) {
	if fuzzyOptions == nil {
		return nil, fmt.Errorf("fuzzy options required for fuzzy matching")
	}

	// Get all nodes of this type to search against
	params := postgres.GetGraphNodesByTypeParams{
		NodeType: entityType,
		Limit:    100, // Reasonable limit to avoid performance issues
		Offset:   0,
	}

	allNodes, err := m.repo.GetGraphNodesByType(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nodes for fuzzy matching: %w", err)
	}

	var candidates []postgres.GraphNode

	// Check each node for fuzzy match
	for _, node := range allNodes {
		// Parse node data to extract field value
		var nodeData map[string]interface{}
		if err := json.Unmarshal(node.Data, &nodeData); err != nil {
			m.logger.Warn("failed to parse node data for fuzzy matching", "node_id", node.ID, "error", err)
			continue
		}

		m.logger.Debug("checking node for fuzzy match",
			"node_id", node.ID,
			"field_path", fieldPath,
			"search_value", searchValue,
			"node_data_keys", getMapKeys(nodeData))

		// Extract the field value
		fieldValue, exists := m.extractFieldFromMap(nodeData, fieldPath)
		if !exists || fieldValue == nil {
			m.logger.Debug("field not found in node",
				"node_id", node.ID,
				"field_path", fieldPath,
				"exists", exists,
				"field_value_nil", fieldValue == nil)
			continue
		}

		fieldStr := fmt.Sprintf("%v", fieldValue)
		if fieldStr == "" {
			continue
		}

		// Perform fuzzy comparison
		similarity := m.fuzzyMatcher.CalculateSimilarity(searchValue, fieldStr, fuzzyOptions)
		m.logger.Debug("calculated similarity",
			"search_value", searchValue,
			"field_value", fieldStr,
			"similarity", similarity,
			"min_required", fuzzyOptions.MinSimilarity,
			"field_path", fieldPath)

		if similarity >= fuzzyOptions.MinSimilarity {
			candidates = append(candidates, node)
			m.logger.Debug("fuzzy match found",
				"search_value", searchValue,
				"field_value", fieldStr,
				"similarity", similarity,
				"field_path", fieldPath,
				"node_id", node.ID)
		}
	}

	return candidates, nil
}

// findCandidatesBySecondaryFields searches using secondary matching fields (supports both exact and fuzzy matches)
func (m *Matcher) findCandidatesBySecondaryFields(ctx context.Context, entityType string, entityData map[string]interface{}, secondaryRules []matching.FieldMatchRule) ([]*reconciler.GraphNode, error) {
	var candidates []*reconciler.GraphNode

	// For secondary rules, try both exact and fuzzy matching approaches
	for _, rule := range secondaryRules {
		fieldValue, exists := m.extractFieldFromMap(entityData, rule.FieldPath)
		if !exists || fieldValue == nil {
			continue
		}

		valueStr := fmt.Sprintf("%v", fieldValue)
		if valueStr == "" {
			continue
		}

		var dbNodes []postgres.GraphNode
		var err error

		if rule.MatchType == matching.MatchExact {
			// Use single field exact match for secondary rules
			params := postgres.FindNodesByFieldMatchParams{
				NodeType:   entityType,
				JsonField:  pgtype.Text{String: rule.FieldPath, Valid: true},
				FieldValue: pgtype.Text{String: valueStr, Valid: true},
				Limit:      15, // Moderate limit for secondary matches
				Offset:     0,
			}

			dbNodes, err = m.repo.FindNodesByFieldMatch(ctx, params)
			if err != nil {
				m.logger.Error("exact secondary field match query failed", "field", rule.FieldPath, "error", err)
				continue
			}

			m.logger.Debug("found exact secondary field matches",
				"field", rule.FieldPath, "value", valueStr, "count", len(dbNodes))

		} else if rule.MatchType == matching.MatchFuzzy {
			// For fuzzy secondary matches
			fuzzyNodes, err := m.findFuzzyCandidatesForField(ctx, entityType, rule.FieldPath, valueStr, rule.FuzzyOptions)
			if err != nil {
				m.logger.Error("fuzzy secondary field match query failed", "field", rule.FieldPath, "error", err)
				continue
			}

			dbNodes = fuzzyNodes
			m.logger.Debug("found fuzzy secondary field matches",
				"field", rule.FieldPath, "value", valueStr, "count", len(dbNodes))
		}

		// Convert to GraphNode format and add to candidates
		for _, dbNode := range dbNodes {
			candidates = append(candidates, &reconciler.GraphNode{
				ID:             dbNode.ID,
				ExternalID:     dbNode.ExternalID,
				NodeType:       dbNode.NodeType,
				Data:           dbNode.Data,
				DuplicateCount: dbNode.DuplicateCount,
			})
		}

		// Don't break early for secondary rules - collect from multiple fields
		if len(dbNodes) > 0 {
			m.logger.Debug("accumulated secondary candidates", "field", rule.FieldPath, "total_candidates", len(candidates))
		}
	}

	return candidates, nil
}

// findCandidatesByFallbackStrategy searches using broader criteria
func (m *Matcher) findCandidatesByFallbackStrategy(ctx context.Context, entityType string, _ map[string]interface{}, _ []matching.FieldMatchRule) ([]*reconciler.GraphNode, error) {
	// For fallback, get all nodes of this type and let the scoring algorithm handle fuzzy matching
	params := postgres.GetGraphNodesByTypeParams{
		NodeType: entityType,
		Limit:    30, // Reasonable limit for fallback strategy
		Offset:   0,
	}

	dbNodes, err := m.repo.GetGraphNodesByType(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fallback node retrieval failed: %w", err)
	}

	var candidates []*reconciler.GraphNode
	for _, dbNode := range dbNodes {
		candidates = append(candidates, &reconciler.GraphNode{
			ID:             dbNode.ID,
			ExternalID:     dbNode.ExternalID,
			NodeType:       dbNode.NodeType,
			Data:           dbNode.Data,
			DuplicateCount: dbNode.DuplicateCount,
		})
	}

	m.logger.Debug("found fallback candidates", "count", len(candidates))
	return candidates, nil
}

// deduplicateCandidates removes duplicate nodes from the candidate list
func (m *Matcher) deduplicateCandidates(candidates []*reconciler.GraphNode) []*reconciler.GraphNode {
	seen := make(map[int64]bool)
	var unique []*reconciler.GraphNode

	for _, candidate := range candidates {
		if !seen[candidate.ID] {
			seen[candidate.ID] = true
			unique = append(unique, candidate)
		}
	}

	return unique
}

// extractFieldFromMap extracts a field value from a map using dot notation
func (m *Matcher) extractFieldFromMap(data map[string]interface{}, fieldPath string) (interface{}, bool) {
	value, err := extractFieldValue(data, fieldPath)
	return value, err == nil && value != nil
}

// scoreMatch calculates the confidence score for a potential match
func (m *Matcher) scoreMatch(entity *diodepb.Entity, candidate *reconciler.GraphNode, rule *matching.EntityMatchingRule) (*matching.MatchResult, error) {
	// Convert entity to comparable format
	entityData, err := m.entityToMap(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to convert entity to map: %w", err)
	}

	// Parse candidate data
	var candidateData map[string]interface{}
	if err := json.Unmarshal(candidate.Data, &candidateData); err != nil {
		return nil, fmt.Errorf("failed to parse candidate data: %w", err)
	}

	result := &matching.MatchResult{
		NodeID:         &candidate.ID,
		Confidence:     matching.ConfidenceNone,
		MatchingFields: []string{},
		MatchReason:    "",
		ExistingData:   candidate.Data,
		ExternalID:     &candidate.ExternalID, // Include external ID for reuse
	}

	// Score primary rules
	primaryScore, primaryMatches, primaryReason := m.scoreFieldRules(entityData, candidateData, rule.PrimaryRules)

	// Score secondary rules
	secondaryScore, secondaryMatches, secondaryReason := m.scoreFieldRules(entityData, candidateData, rule.SecondaryRules)

	// Score fallback rules
	fallbackScore, fallbackMatches, fallbackReason := m.scoreFieldRules(entityData, candidateData, rule.FallbackRules)

	// Determine which ruleset to use based on scores and configuration
	var finalScore matching.MatchConfidence
	var matchingFields []string
	var reason string

	if primaryScore >= rule.MinConfidence && (!rule.GetRequireAllPrimary(true) || len(primaryMatches) == len(rule.PrimaryRules)) {
		finalScore = primaryScore
		matchingFields = primaryMatches
		reason = fmt.Sprintf("Primary rules: %s", primaryReason)
	} else if secondaryScore >= rule.MinConfidence {
		finalScore = secondaryScore
		matchingFields = secondaryMatches
		reason = fmt.Sprintf("Secondary rules: %s", secondaryReason)
	} else if m.config.EnableFallback && fallbackScore >= rule.MinConfidence {
		finalScore = fallbackScore
		matchingFields = fallbackMatches
		reason = fmt.Sprintf("Fallback rules: %s", fallbackReason)
	} else {
		finalScore = matching.ConfidenceNone
		reason = "No rules met minimum confidence threshold"
	}

	result.Confidence = finalScore
	result.MatchingFields = matchingFields
	result.MatchReason = reason

	return result, nil
}

// scoreFieldRules scores a set of field rules and returns confidence, matching fields, and reason
func (m *Matcher) scoreFieldRules(entityData, candidateData map[string]interface{}, rules []matching.FieldMatchRule) (matching.MatchConfidence, []string, string) {
	if len(rules) == 0 {
		return matching.ConfidenceNone, nil, "no rules defined"
	}

	var totalConfidence matching.MatchConfidence
	var matchingFields []string
	var reasons []string

	for _, rule := range rules {
		entityValue, err1 := extractFieldValue(entityData, rule.FieldPath)
		candidateValue, err2 := extractFieldValue(candidateData, rule.FieldPath)

		if err1 != nil || err2 != nil {
			if rule.Required {
				return matching.ConfidenceNone, nil, fmt.Sprintf("required field %s not accessible", rule.FieldPath)
			}
			continue
		}

		matches, confidence := m.compareFieldValues(entityValue, candidateValue, rule)
		if matches {
			totalConfidence += confidence * matching.MatchConfidence(rule.Weight)
			matchingFields = append(matchingFields, rule.FieldPath)
			reasons = append(reasons, fmt.Sprintf("%s(%.2f)", rule.FieldPath, float64(confidence)))
		} else if rule.Required {
			return matching.ConfidenceNone, nil, fmt.Sprintf("required field %s did not match", rule.FieldPath)
		}
	}

	// Cap the confidence at 1.0
	if totalConfidence > 1.0 {
		totalConfidence = 1.0
	}

	return totalConfidence, matchingFields, strings.Join(reasons, ", ")
}

// compareFieldValues compares two field values based on the match rule
func (m *Matcher) compareFieldValues(entityValue, candidateValue interface{}, rule matching.FieldMatchRule) (bool, matching.MatchConfidence) {
	if entityValue == nil || candidateValue == nil {
		// MatchExists requires both values to be present; if either is nil, no match
		return false, matching.ConfidenceNone
	}

	switch rule.MatchType {
	case matching.MatchExact:
		return m.compareExact(entityValue, candidateValue), rule.Confidence

	case matching.MatchFuzzy:
		if rule.FuzzyOptions == nil {
			return false, matching.ConfidenceNone
		}
		return m.compareFuzzy(entityValue, candidateValue, rule.FuzzyOptions, rule.Confidence)

	case matching.MatchContains:
		return m.compareContains(entityValue, candidateValue), rule.Confidence

	case matching.MatchNumeric:
		return m.compareNumeric(entityValue, candidateValue), rule.Confidence

	case matching.MatchExists:
		return true, rule.Confidence // Both values exist

	case matching.MatchRegex:
		return m.compareRegex(entityValue, candidateValue), rule.Confidence

	default:
		m.logger.Warn("unknown match type", "match_type", rule.MatchType)
		return false, matching.ConfidenceNone
	}
}

// compareExact performs exact value comparison
func (m *Matcher) compareExact(entityValue, candidateValue interface{}) bool {
	// Convert both to strings for comparison
	entityStr := fmt.Sprintf("%v", entityValue)
	candidateStr := fmt.Sprintf("%v", candidateValue)
	return entityStr == candidateStr
}

// compareFuzzy performs fuzzy string matching using advanced algorithms
func (m *Matcher) compareFuzzy(entityValue, candidateValue interface{}, options *matching.FuzzyOptions, baseConfidence matching.MatchConfidence) (bool, matching.MatchConfidence) {
	entityStr := fmt.Sprintf("%v", entityValue)
	candidateStr := fmt.Sprintf("%v", candidateValue)

	// Use the database-agnostic fuzzy matcher with Jaro-Winkler and Levenshtein algorithms
	similarity := m.fuzzyMatcher.CalculateSimilarity(entityStr, candidateStr, options)

	if similarity >= options.MinSimilarity {
		// Scale confidence based on similarity
		adjustedConfidence := baseConfidence * matching.MatchConfidence(similarity)
		return true, adjustedConfidence
	}

	return false, matching.ConfidenceNone
}

// compareContains checks if one string contains the other
func (m *Matcher) compareContains(entityValue, candidateValue interface{}) bool {
	entityStr := strings.ToLower(fmt.Sprintf("%v", entityValue))
	candidateStr := strings.ToLower(fmt.Sprintf("%v", candidateValue))

	return strings.Contains(entityStr, candidateStr) || strings.Contains(candidateStr, entityStr)
}

// compareNumeric performs numeric comparison with tolerance
func (m *Matcher) compareNumeric(entityValue, candidateValue interface{}) bool {
	entityFloat, err1 := parseFloat(entityValue)
	candidateFloat, err2 := parseFloat(candidateValue)

	if err1 != nil || err2 != nil {
		return false
	}

	// Use small tolerance for floating point comparison
	tolerance := 0.001
	return abs(entityFloat-candidateFloat) <= tolerance
}

// compareRegex performs regular expression matching
func (m *Matcher) compareRegex(entityValue, candidateValue interface{}) bool {
	pattern := fmt.Sprintf("%v", entityValue)
	text := fmt.Sprintf("%v", candidateValue)

	matched, err := regexp.MatchString(pattern, text)
	if err != nil {
		return false
	}

	return matched
}

// entityToMap converts a protobuf entity to a map for field extraction
func (m *Matcher) entityToMap(entity *diodepb.Entity) (map[string]interface{}, error) {
	// Get the actual entity from the wrapper
	actualEntity := entity.GetEntity()
	if actualEntity == nil {
		return nil, fmt.Errorf("entity wrapper contains no actual entity")
	}

	// Convert to JSON and back to map for easy field access
	data, err := json.Marshal(actualEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity to map: %w", err)
	}

	return result, nil
}

// Cache management methods
func (m *Matcher) getCachedResults(entity *diodepb.Entity, entityType string) []matching.MatchResult {
	if !m.config.CacheResults {
		return nil
	}

	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	key := m.getCacheKey(entity, entityType)
	return m.cache[key]
}

func (m *Matcher) setCachedResults(entity *diodepb.Entity, entityType string, results []matching.MatchResult) {
	if !m.config.CacheResults {
		return
	}

	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// Implement cache size limit
	if len(m.cache) >= m.config.MaxCacheSize {
		// Simple eviction: clear half the cache
		for k := range m.cache {
			delete(m.cache, k)
			if len(m.cache) < m.config.MaxCacheSize/2 {
				break
			}
		}
	}

	key := m.getCacheKey(entity, entityType)
	m.cache[key] = results
}

func (m *Matcher) clearCacheForEntityType(entityType string) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// Clear cache entries that contain the entity type in their key
	for key := range m.cache {
		if strings.Contains(key, entityType) {
			delete(m.cache, key)
		}
	}
}

func (m *Matcher) getCacheKey(entity *diodepb.Entity, entityType string) string {
	// Create a simple cache key based on entity type and primary identifying fields
	// This is a simplified approach - in production you might want a more sophisticated key
	return fmt.Sprintf("%s:%p", entityType, entity)
}

// Utility functions

func parseFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// getMapKeys returns the keys of a map as a slice for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
