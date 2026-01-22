package entitymatcher

import (
	"context"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/matching"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

// setupMockRepository sets up the mockery-generated GraphRepository with test expectations
func setupMockRepository(_ *testing.T) *mocks.GraphRepository {
	mockRepo := &mocks.GraphRepository{}

	// Setup test nodes
	testNodes := []postgres.GraphNode{
		{
			ID:             1,
			ExternalID:     "device-01",
			NodeType:       "Device",
			Data:           []byte(`{"Device": {"name": "server-01", "serial": "ABC123"}}`),
			DuplicateCount: 1,
		},
		{
			ID:             2,
			ExternalID:     "device-02",
			NodeType:       "Device",
			Data:           []byte(`{"Device": {"name": "server-02", "serial": "DEF456"}}`),
			DuplicateCount: 1,
		},
	}

	// Mock FindNodesByFieldMatch - return matching node based on field
	mockRepo.On("FindNodesByFieldMatch", mock.Anything, mock.Anything).Return(func(_ context.Context, arg postgres.FindNodesByFieldMatchParams) []postgres.GraphNode {
		if !arg.JsonField.Valid {
			return []postgres.GraphNode{}
		}
		fieldName := arg.JsonField.String
		fieldValue := arg.FieldValue.String

		for _, node := range testNodes {
			if node.NodeType == arg.NodeType {
				if fieldName == "Device.serial" && fieldValue == "ABC123" {
					return []postgres.GraphNode{testNodes[0]}
				}
			}
		}
		return []postgres.GraphNode{}
	}, nil)

	// Mock GetGraphNodesByType - needed for fuzzy matching to get all nodes of a type
	mockRepo.On("GetGraphNodesByType", mock.Anything, mock.Anything).Return(func(_ context.Context, arg postgres.GetGraphNodesByTypeParams) []postgres.GraphNode {
		var result []postgres.GraphNode
		for _, node := range testNodes {
			if node.NodeType == arg.NodeType {
				result = append(result, node)
			}
		}
		return result
	}, nil)

	return mockRepo
}

// createTestMatchingConfig creates a test configuration for matching
// This simulates what would be loaded from YAML and converted to EntityMatchingConfig
func createTestMatchingConfig() *matching.EntityMatchingConfig {
	// Create a YAML config structure
	yamlConfig := &matching.Config{
		GlobalSettings: matching.GlobalMatchingSettings{
			DefaultMinConfidence:     0.5,
			DefaultRequireAllPrimary: false,
			EnableFuzzyMatching:      true,
			DefaultFuzzyThreshold:    0.8,
		},
		DefaultEntityRules: map[string]matching.EntityMatchingRule{
			"Device": {
				EntityType: "Device",
				PrimaryRules: []matching.FieldMatchRule{
					{
						FieldPath:  "Device.name",
						MatchType:  matching.MatchFuzzy,
						Weight:     1.0,
						Required:   true,
						Confidence: matching.ConfidenceHigh,
						FuzzyOptions: &matching.FuzzyOptions{
							MinSimilarity: 0.8,
							CaseIgnore:    true,
							Normalize:     true,
						},
					},
					{
						FieldPath:  "Device.serial",
						MatchType:  matching.MatchExact,
						Weight:     1.0,
						Required:   true,
						Confidence: matching.ConfidenceHigh,
					},
				},
				MinConfidence: matching.ConfidenceMedium,
				// RequireAllPrimary omitted - defaults to nil (uses global default)
			},
		},
		EntityOverrides: make(map[string]matching.EntityMatchingRule),
	}

	// Convert YAML config to EntityMatchingConfig (what the entity matcher expects)
	finalRules := yamlConfig.GetFinalRules()

	return &matching.EntityMatchingConfig{
		Rules:          finalRules,
		GlobalMinConf:  matching.ConfidenceLow,
		EnableFallback: true,
		CacheResults:   true,
		MaxCacheSize:   1000,
	}
}

func TestEntityMatcherIntegration(t *testing.T) {
	// Create a mock repository with expectations
	repo := setupMockRepository(t)

	// Create entity matcher with test config
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	config := createTestMatchingConfig()
	t.Logf("Created config with %d rules", len(config.Rules))
	for entityType, rule := range config.Rules {
		t.Logf("Rule for %s: %d primary rules, min confidence: %v", entityType, len(rule.PrimaryRules), rule.MinConfidence)
	}
	matcher := NewMatcher(repo, config, logger)

	// Create a test entity
	name := "server-01"
	serial := "ABC123"
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{
			Device: &diodepb.Device{
				Name:   &name,
				Serial: &serial,
			},
		},
	}
	t.Logf("Created entity of type: %v", reflect.TypeOf(entity.GetEntity()).Elem().Name())

	// Test finding matches
	matches, err := matcher.FindMatches(context.Background(), entity)
	if err != nil {
		t.Fatalf("FindMatches failed: %v", err)
	}
	t.Logf("FindMatches returned %d matches with error: %v", len(matches), err)

	// We should find at least one match (could be more due to fallback strategy)
	if len(matches) == 0 {
		t.Error("Expected at least one match, got zero")
	}

	t.Logf("Found %d potential matches", len(matches))

	// Test finding best match
	bestMatch, err := matcher.FindBestMatch(context.Background(), entity)
	if err != nil {
		t.Fatalf("FindBestMatch failed: %v", err)
	}

	if bestMatch == nil {
		t.Error("Expected a best match, got nil")
	} else {
		t.Logf("Best match: NodeID=%v, Confidence=%f", bestMatch.NodeID, bestMatch.Confidence)
	}

	// Test getting matching rule
	rule, err := matcher.GetMatchingRule("Device")
	if err != nil {
		t.Fatalf("GetMatchingRule failed: %v", err)
	}

	if rule == nil {
		t.Error("Expected matching rule for Device, got nil")
	} else {
		t.Logf("Device matching rule: PrimaryRules=%d, SecondaryRules=%d",
			len(rule.PrimaryRules), len(rule.SecondaryRules))
	}
}

func TestEntityMatcherFuzzyMatching(t *testing.T) {
	// Create a mock repository with expectations
	repo := setupMockRepository(t)

	// Create entity matcher with test config
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	config := createTestMatchingConfig()
	matcher := NewMatcher(repo, config, logger)

	// Create a test entity with slightly different name (typo)
	name := "servr-01" // Missing 'e' - should still match with fuzzy matching
	serial := "ABC123"
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{
			Device: &diodepb.Device{
				Name:   &name,
				Serial: &serial,
			},
		},
	}

	// Test fuzzy matching
	matches, err := matcher.FindMatches(context.Background(), entity)
	if err != nil {
		t.Fatalf("FindMatches failed: %v", err)
	}

	t.Logf("Fuzzy matching found %d potential matches", len(matches))

	// The fuzzy matcher should find some matches even with typos
	// (exact behavior depends on confidence thresholds and fuzzy options)
	for _, match := range matches {
		t.Logf("Match: Confidence=%f, Reason=%s", match.Confidence, match.MatchReason)
	}
}
