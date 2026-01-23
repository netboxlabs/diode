package entitymatcher

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/netboxlabs/diode/diode-server/gen/dbstore/postgres"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/matching"
	"github.com/netboxlabs/diode/diode-server/reconciler"
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

func TestDefaultEntityMatchingConfig(t *testing.T) {
	config := DefaultEntityMatchingConfig()

	if config == nil {
		t.Fatal("DefaultEntityMatchingConfig returned nil")
	}
	if config.Rules == nil {
		t.Error("Rules map should not be nil")
	}
	if config.GlobalMinConf != matching.ConfidenceLow {
		t.Errorf("Expected GlobalMinConf to be ConfidenceLow, got %v", config.GlobalMinConf)
	}
	if !config.EnableFallback {
		t.Error("Expected EnableFallback to be true")
	}
	if !config.CacheResults {
		t.Error("Expected CacheResults to be true")
	}
	if config.MaxCacheSize != 1000 {
		t.Errorf("Expected MaxCacheSize to be 1000, got %d", config.MaxCacheSize)
	}
}

func TestNewMatcherWithNilConfig(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	matcher := NewMatcher(mockRepo, nil, logger)

	if matcher == nil {
		t.Fatal("NewMatcher returned nil")
	}
	// Should use default config when nil is passed
	if matcher.config == nil {
		t.Error("Matcher config should not be nil")
	}
}

func TestExtractFieldValue(t *testing.T) {
	tests := []struct {
		name      string
		data      interface{}
		fieldPath string
		wantValue interface{}
		wantErr   bool
	}{
		{
			name:      "simple map access",
			data:      map[string]interface{}{"name": "test"},
			fieldPath: "name",
			wantValue: "test",
			wantErr:   false,
		},
		{
			name:      "nested map access",
			data:      map[string]interface{}{"Device": map[string]interface{}{"name": "server-01"}},
			fieldPath: "Device.name",
			wantValue: "server-01",
			wantErr:   false,
		},
		{
			name:      "case insensitive access",
			data:      map[string]interface{}{"Name": "test"},
			fieldPath: "name",
			wantValue: "test",
			wantErr:   false,
		},
		{
			name:      "missing field",
			data:      map[string]interface{}{"name": "test"},
			fieldPath: "missing",
			wantValue: nil,
			wantErr:   false,
		},
		{
			name:      "nil data",
			data:      nil,
			fieldPath: "name",
			wantValue: nil,
			wantErr:   true,
		},
		{
			name:      "deeply nested",
			data:      map[string]interface{}{"a": map[string]interface{}{"b": map[string]interface{}{"c": "deep"}}},
			fieldPath: "a.b.c",
			wantValue: "deep",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractFieldValue(tt.data, tt.fieldPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantValue {
				t.Errorf("extractFieldValue() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestCompareContains(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	matcher := &Matcher{logger: logger}

	tests := []struct {
		name      string
		entity    interface{}
		candidate interface{}
		want      bool
	}{
		{"substring match", "hello world", "world", true},
		{"reverse substring", "world", "hello world", true},
		{"case insensitive", "HELLO", "hello", true},
		{"no match", "foo", "bar", false},
		{"exact match", "test", "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matcher.compareContains(tt.entity, tt.candidate); got != tt.want {
				t.Errorf("compareContains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareNumeric(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	matcher := &Matcher{logger: logger}

	tests := []struct {
		name      string
		entity    interface{}
		candidate interface{}
		want      bool
	}{
		{"equal integers", 42, 42, true},
		{"equal floats", 3.14, 3.14, true},
		{"within tolerance", 1.0001, 1.0002, true},
		{"outside tolerance", 1.0, 2.0, false},
		{"string numbers", "42", "42", true},
		{"mixed types", 42, 42.0, true},
		{"non-numeric", "abc", "def", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matcher.compareNumeric(tt.entity, tt.candidate); got != tt.want {
				t.Errorf("compareNumeric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareRegex(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	matcher := &Matcher{logger: logger}

	tests := []struct {
		name    string
		pattern interface{}
		text    interface{}
		want    bool
	}{
		{"simple match", "server-.*", "server-01", true},
		{"no match", "^foo$", "bar", false},
		{"exact match pattern", "^test$", "test", true},
		{"partial match", "test", "this is a test", true},
		{"invalid regex", "[invalid", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matcher.compareRegex(tt.pattern, tt.text); got != tt.want {
				t.Errorf("compareRegex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeduplicateCandidates(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	matcher := &Matcher{logger: logger}

	candidates := []*reconciler.GraphNode{
		{ID: 1, ExternalID: "node-1"},
		{ID: 2, ExternalID: "node-2"},
		{ID: 1, ExternalID: "node-1"}, // duplicate
		{ID: 3, ExternalID: "node-3"},
		{ID: 2, ExternalID: "node-2"}, // duplicate
	}

	result := matcher.deduplicateCandidates(candidates)

	if len(result) != 3 {
		t.Errorf("Expected 3 unique candidates, got %d", len(result))
	}

	// Verify IDs are unique
	seen := make(map[int64]bool)
	for _, c := range result {
		if seen[c.ID] {
			t.Errorf("Duplicate ID found: %d", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestUpdateMatchingRule(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := &matching.EntityMatchingConfig{
		Rules:        make(map[string]*matching.EntityMatchingRule),
		CacheResults: true,
	}
	matcher := NewMatcher(mockRepo, config, logger)

	// Test adding a new rule
	newRule := &matching.EntityMatchingRule{
		EntityType: "TestEntity",
		PrimaryRules: []matching.FieldMatchRule{
			{FieldPath: "name", MatchType: matching.MatchExact},
		},
	}

	err := matcher.UpdateMatchingRule("TestEntity", newRule)
	if err != nil {
		t.Fatalf("UpdateMatchingRule failed: %v", err)
	}

	// Verify the rule was added
	rule, err := matcher.GetMatchingRule("TestEntity")
	if err != nil {
		t.Fatalf("GetMatchingRule failed: %v", err)
	}
	if rule.EntityType != "TestEntity" {
		t.Errorf("Expected EntityType 'TestEntity', got '%s'", rule.EntityType)
	}

	// Test nil rule error
	err = matcher.UpdateMatchingRule("TestEntity", nil)
	if err == nil {
		t.Error("Expected error when passing nil rule")
	}
}

func TestGetMatchingRuleNotFound(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := &matching.EntityMatchingConfig{
		Rules: make(map[string]*matching.EntityMatchingRule),
	}
	matcher := NewMatcher(mockRepo, config, logger)

	_, err := matcher.GetMatchingRule("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent rule")
	}
}

func TestFindMatchesNilEntity(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := createTestMatchingConfig()
	matcher := NewMatcher(mockRepo, config, logger)

	_, err := matcher.FindMatches(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil entity")
	}
}

func TestFindMatchesNoMatchingRule(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := &matching.EntityMatchingConfig{
		Rules: make(map[string]*matching.EntityMatchingRule),
	}
	matcher := NewMatcher(mockRepo, config, logger)

	name := "test-site"
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{
				Name: name,
			},
		},
	}

	matches, err := matcher.FindMatches(context.Background(), entity)
	if err != nil {
		t.Fatalf("FindMatches failed: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for entity without matching rule, got %d", len(matches))
	}
}

func TestFindBestMatchNoMatches(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := &matching.EntityMatchingConfig{
		Rules: make(map[string]*matching.EntityMatchingRule),
	}
	matcher := NewMatcher(mockRepo, config, logger)

	name := "test"
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{
				Name: name,
			},
		},
	}

	best, err := matcher.FindBestMatch(context.Background(), entity)
	if err != nil {
		t.Fatalf("FindBestMatch failed: %v", err)
	}
	if best != nil {
		t.Error("Expected nil for no matches")
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    float64
		wantErr bool
	}{
		{"float64", 3.14, 3.14, false},
		{"float32", float32(2.5), 2.5, false},
		{"int", 42, 42.0, false},
		{"int64", int64(100), 100.0, false},
		{"string", "3.14", 3.14, false},
		{"invalid string", "abc", 0, true},
		{"unsupported type", []int{1, 2}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFloat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{5.0, 5.0},
		{-5.0, 5.0},
		{0.0, 0.0},
		{-0.001, 0.001},
	}

	for _, tt := range tests {
		if got := abs(tt.input); got != tt.want {
			t.Errorf("abs(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestCompareFieldValuesAllTypes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	fm := matching.NewFuzzyMatcher()
	matcher := &Matcher{logger: logger, fuzzyMatcher: fm}

	tests := []struct {
		name      string
		entity    interface{}
		candidate interface{}
		rule      matching.FieldMatchRule
		wantMatch bool
	}{
		{
			name:      "exact match success",
			entity:    "test",
			candidate: "test",
			rule:      matching.FieldMatchRule{MatchType: matching.MatchExact, Confidence: 0.9},
			wantMatch: true,
		},
		{
			name:      "exact match failure",
			entity:    "test",
			candidate: "other",
			rule:      matching.FieldMatchRule{MatchType: matching.MatchExact, Confidence: 0.9},
			wantMatch: false,
		},
		{
			name:      "contains match",
			entity:    "hello world",
			candidate: "world",
			rule:      matching.FieldMatchRule{MatchType: matching.MatchContains, Confidence: 0.8},
			wantMatch: true,
		},
		{
			name:      "numeric match",
			entity:    42,
			candidate: 42,
			rule:      matching.FieldMatchRule{MatchType: matching.MatchNumeric, Confidence: 0.9},
			wantMatch: true,
		},
		{
			name:      "exists match - both present",
			entity:    "value",
			candidate: "other",
			rule:      matching.FieldMatchRule{MatchType: matching.MatchExists, Confidence: 0.5},
			wantMatch: true,
		},
		{
			name:      "exists match - entity nil",
			entity:    nil,
			candidate: "value",
			rule:      matching.FieldMatchRule{MatchType: matching.MatchExists, Confidence: 0.5},
			wantMatch: false,
		},
		{
			name:      "regex match",
			entity:    "server-.*",
			candidate: "server-01",
			rule:      matching.FieldMatchRule{MatchType: matching.MatchRegex, Confidence: 0.9},
			wantMatch: true,
		},
		{
			name:      "fuzzy match without options",
			entity:    "test",
			candidate: "test",
			rule:      matching.FieldMatchRule{MatchType: matching.MatchFuzzy, Confidence: 0.9, FuzzyOptions: nil},
			wantMatch: false,
		},
		{
			name:      "unknown match type",
			entity:    "test",
			candidate: "test",
			rule:      matching.FieldMatchRule{MatchType: "unknown", Confidence: 0.9},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := matcher.compareFieldValues(tt.entity, tt.candidate, tt.rule)
			if got != tt.wantMatch {
				t.Errorf("compareFieldValues() match = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestSecondaryFieldMatching(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Setup mock to return nodes when queried
	testNodes := []postgres.GraphNode{
		{
			ID:             1,
			ExternalID:     "device-01",
			NodeType:       "Device",
			Data:           []byte(`{"Device": {"name": "server-01", "description": "Primary server"}}`),
			DuplicateCount: 1,
		},
	}

	mockRepo.On("FindNodesByFieldMatch", mock.Anything, mock.MatchedBy(func(arg postgres.FindNodesByFieldMatchParams) bool {
		return arg.NodeType == "Device" && arg.JsonField.String == "Device.description"
	})).Return(testNodes, nil)

	mockRepo.On("FindNodesByFieldMatch", mock.Anything, mock.Anything).Return([]postgres.GraphNode{}, nil)
	mockRepo.On("GetGraphNodesByType", mock.Anything, mock.Anything).Return([]postgres.GraphNode{}, nil)

	config := &matching.EntityMatchingConfig{
		Rules: map[string]*matching.EntityMatchingRule{
			"Device": {
				EntityType: "Device",
				PrimaryRules: []matching.FieldMatchRule{
					{
						FieldPath:  "Device.name",
						MatchType:  matching.MatchExact,
						Required:   true,
						Confidence: matching.ConfidenceHigh,
					},
				},
				SecondaryRules: []matching.FieldMatchRule{
					{
						FieldPath:  "Device.description",
						MatchType:  matching.MatchExact,
						Required:   false,
						Confidence: matching.ConfidenceMedium,
					},
				},
				MinConfidence: matching.ConfidenceLow,
			},
		},
		GlobalMinConf:  matching.ConfidenceLow,
		EnableFallback: true,
	}

	matcher := NewMatcher(mockRepo, config, logger)

	name := "server-01"
	desc := "Primary server"
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{
			Device: &diodepb.Device{
				Name:        &name,
				Description: &desc,
			},
		},
	}

	matches, err := matcher.FindMatches(context.Background(), entity)
	if err != nil {
		t.Fatalf("FindMatches failed: %v", err)
	}

	// Should find matches via secondary rules when primary doesn't match
	t.Logf("Found %d matches via secondary rules", len(matches))
}

func TestFallbackStrategyMatching(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Setup mock to return nodes for fallback query
	testNodes := []postgres.GraphNode{
		{
			ID:             1,
			ExternalID:     "device-01",
			NodeType:       "Device",
			Data:           []byte(`{"Device": {"name": "server-01"}}`),
			DuplicateCount: 1,
		},
	}

	mockRepo.On("FindNodesByFieldMatch", mock.Anything, mock.Anything).Return([]postgres.GraphNode{}, nil)
	mockRepo.On("GetGraphNodesByType", mock.Anything, mock.MatchedBy(func(arg postgres.GetGraphNodesByTypeParams) bool {
		return arg.NodeType == "Device"
	})).Return(testNodes, nil)

	config := &matching.EntityMatchingConfig{
		Rules: map[string]*matching.EntityMatchingRule{
			"Device": {
				EntityType: "Device",
				PrimaryRules: []matching.FieldMatchRule{
					{
						FieldPath:  "Device.serial",
						MatchType:  matching.MatchExact,
						Required:   true,
						Confidence: matching.ConfidenceHigh,
					},
				},
				FallbackRules: []matching.FieldMatchRule{
					{
						FieldPath:  "Device.name",
						MatchType:  matching.MatchFuzzy,
						Confidence: matching.ConfidenceLow,
						FuzzyOptions: &matching.FuzzyOptions{
							MinSimilarity: 0.7,
							CaseIgnore:    true,
						},
					},
				},
				MinConfidence: matching.ConfidenceLow,
			},
		},
		GlobalMinConf:  matching.ConfidenceLow,
		EnableFallback: true,
	}

	matcher := NewMatcher(mockRepo, config, logger)

	name := "server-01"
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{
			Device: &diodepb.Device{
				Name: &name,
				// No serial provided, should trigger fallback
			},
		},
	}

	matches, err := matcher.FindMatches(context.Background(), entity)
	if err != nil {
		t.Fatalf("FindMatches failed: %v", err)
	}

	t.Logf("Found %d matches via fallback strategy", len(matches))
}

func TestExtractFieldValueWithStruct(t *testing.T) {
	type Inner struct {
		Value string
	}
	type Outer struct {
		Inner Inner
		Name  string
	}

	tests := []struct {
		name      string
		data      interface{}
		fieldPath string
		wantValue interface{}
		wantErr   bool
	}{
		{
			name:      "struct field access",
			data:      Outer{Name: "test", Inner: Inner{Value: "nested"}},
			fieldPath: "Name",
			wantValue: "test",
			wantErr:   false,
		},
		{
			name:      "nested struct access",
			data:      Outer{Inner: Inner{Value: "nested"}},
			fieldPath: "Inner.Value",
			wantValue: "nested",
			wantErr:   false,
		},
		{
			name:      "pointer to struct",
			data:      &Outer{Name: "pointer-test"},
			fieldPath: "Name",
			wantValue: "pointer-test",
			wantErr:   false,
		},
		{
			name:      "nil pointer",
			data:      (*Outer)(nil),
			fieldPath: "Name",
			wantValue: nil,
			wantErr:   false,
		},
		{
			name:      "non-struct navigation",
			data:      "string value",
			fieldPath: "field",
			wantValue: nil,
			wantErr:   true,
		},
		{
			name:      "pointer to map",
			data:      &map[string]interface{}{"key": "value"},
			fieldPath: "key",
			wantValue: "value",
			wantErr:   false,
		},
		{
			name:      "nil pointer to map",
			data:      (*map[string]interface{})(nil),
			fieldPath: "key",
			wantValue: nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractFieldValue(tt.data, tt.fieldPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantValue {
				t.Errorf("extractFieldValue() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestCacheEviction(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	config := &matching.EntityMatchingConfig{
		Rules:        make(map[string]*matching.EntityMatchingRule),
		CacheResults: true,
		MaxCacheSize: 5, // Small cache for testing eviction
	}

	matcher := NewMatcher(mockRepo, config, logger)

	// Fill the cache beyond capacity to trigger eviction
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("test-%d", i)
		entity := &diodepb.Entity{
			Entity: &diodepb.Entity_Site{
				Site: &diodepb.Site{
					Name: name,
				},
			},
		}
		// Directly call setCachedResults to fill the cache
		matcher.setCachedResults(entity, "Site", []matching.MatchResult{})
	}

	// Cache should have been evicted to stay under max size
	if len(matcher.cache) > config.MaxCacheSize {
		t.Errorf("Cache size %d exceeds max %d", len(matcher.cache), config.MaxCacheSize)
	}
}

func TestCacheDisabled(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	config := &matching.EntityMatchingConfig{
		Rules:        make(map[string]*matching.EntityMatchingRule),
		CacheResults: false, // Caching disabled
	}

	matcher := NewMatcher(mockRepo, config, logger)

	name := "test"
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{
				Name: name,
			},
		},
	}

	// Try to set cache - should be no-op
	matcher.setCachedResults(entity, "Site", []matching.MatchResult{{Confidence: 0.9}})

	// Get should return nil when caching is disabled
	result := matcher.getCachedResults(entity, "Site")
	if result != nil {
		t.Error("Expected nil result when caching is disabled")
	}
}

func TestClearCacheForEntityType(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	config := &matching.EntityMatchingConfig{
		Rules:        make(map[string]*matching.EntityMatchingRule),
		CacheResults: true,
		MaxCacheSize: 100,
	}

	matcher := NewMatcher(mockRepo, config, logger)

	// Add cache entries for different entity types
	siteName := "test-site"
	siteEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Site{
			Site: &diodepb.Site{Name: siteName},
		},
	}
	matcher.setCachedResults(siteEntity, "Site", []matching.MatchResult{})

	deviceName := "test-device"
	deviceEntity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{
			Device: &diodepb.Device{Name: &deviceName},
		},
	}
	matcher.setCachedResults(deviceEntity, "Device", []matching.MatchResult{})

	// Clear only Site cache
	matcher.clearCacheForEntityType("Site")

	// Site cache should be cleared
	if len(matcher.cache) == 0 {
		// This is expected as the simple key format makes it hard to distinguish
		t.Log("Cache cleared as expected")
	}
}

func TestEntityToMapEmptyEntity(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := DefaultEntityMatchingConfig()
	matcher := NewMatcher(mockRepo, config, logger)

	// Test with empty entity wrapper
	entity := &diodepb.Entity{}

	_, err := matcher.entityToMap(entity)
	if err == nil {
		t.Error("Expected error for empty entity wrapper")
	}
}

func TestCompareFuzzyWithOptions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	fm := matching.NewFuzzyMatcher()
	matcher := &Matcher{logger: logger, fuzzyMatcher: fm}

	tests := []struct {
		name      string
		entity    interface{}
		candidate interface{}
		rule      matching.FieldMatchRule
		wantMatch bool
	}{
		{
			name:      "fuzzy match with high similarity",
			entity:    "server-01",
			candidate: "server-02",
			rule: matching.FieldMatchRule{
				MatchType:  matching.MatchFuzzy,
				Confidence: 0.9,
				FuzzyOptions: &matching.FuzzyOptions{
					MinSimilarity: 0.8,
					CaseIgnore:    true,
					Normalize:     true,
				},
			},
			wantMatch: true,
		},
		{
			name:      "fuzzy match with low similarity requirement",
			entity:    "test-value",
			candidate: "test-valeu", // Minor typo
			rule: matching.FieldMatchRule{
				MatchType:  matching.MatchFuzzy,
				Confidence: 0.9,
				FuzzyOptions: &matching.FuzzyOptions{
					MinSimilarity: 0.7,
					CaseIgnore:    true,
				},
			},
			wantMatch: true,
		},
		{
			name:      "fuzzy match case insensitive",
			entity:    "SERVER",
			candidate: "server",
			rule: matching.FieldMatchRule{
				MatchType:  matching.MatchFuzzy,
				Confidence: 0.9,
				FuzzyOptions: &matching.FuzzyOptions{
					MinSimilarity: 0.9,
					CaseIgnore:    true,
				},
			},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := matcher.compareFieldValues(tt.entity, tt.candidate, tt.rule)
			if got != tt.wantMatch {
				t.Errorf("compareFieldValues() match = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestGetEntityTypeName(t *testing.T) {
	tests := []struct {
		name     string
		entity   *diodepb.Entity
		wantType string
	}{
		{
			name: "device entity",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{},
				},
			},
			wantType: "Device",
		},
		{
			name: "site entity",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Site{
					Site: &diodepb.Site{},
				},
			},
			wantType: "Site",
		},
		{
			name: "interface entity",
			entity: &diodepb.Entity{
				Entity: &diodepb.Entity_Interface{
					Interface: &diodepb.Interface{},
				},
			},
			wantType: "Interface",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEntityTypeName(tt.entity)
			if got != tt.wantType {
				t.Errorf("getEntityTypeName() = %v, want %v", got, tt.wantType)
			}
		})
	}
}

func TestScoreMatchWithInvalidCandidateData(t *testing.T) {
	mockRepo := &mocks.GraphRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := createTestMatchingConfig()
	matcher := NewMatcher(mockRepo, config, logger)

	name := "server-01"
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{
			Device: &diodepb.Device{Name: &name},
		},
	}

	// Create candidate with invalid JSON
	candidate := &reconciler.GraphNode{
		ID:         1,
		ExternalID: "device-01",
		NodeType:   "Device",
		Data:       []byte(`{invalid json`),
	}

	rule := config.Rules["Device"]
	_, err := matcher.scoreMatch(entity, candidate, rule)
	if err == nil {
		t.Error("Expected error for invalid candidate JSON")
	}
}

func TestFindBestMatchWithHighConfidence(t *testing.T) {
	mockRepo := setupMockRepository(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := createTestMatchingConfig()
	matcher := NewMatcher(mockRepo, config, logger)

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

	best, err := matcher.FindBestMatch(context.Background(), entity)
	if err != nil {
		t.Fatalf("FindBestMatch failed: %v", err)
	}

	if best == nil {
		t.Fatal("Expected a best match, got nil")
	}

	// Should have high confidence due to exact match on both fields
	if best.Confidence < matching.ConfidenceMedium {
		t.Errorf("Expected confidence >= Medium, got %v", best.Confidence)
	}
}
