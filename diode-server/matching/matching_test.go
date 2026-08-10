package matching

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

// Helper function to create string pointer
func ptrString(s string) *string {
	return &s
}

func TestExtractEntityMetadata_NilEntity(t *testing.T) {
	result := ExtractEntityMetadata(nil)
	assert.Nil(t, result)
}

func TestExtractEntityMetadata_NilEntityContent(t *testing.T) {
	entity := &diodepb.Entity{}
	result := ExtractEntityMetadata(entity)
	assert.Nil(t, result)
}

func TestExtractEntityMetadata_NoMetadata(t *testing.T) {
	device := &diodepb.Device{
		Name: ptrString("test-device"),
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	result := ExtractEntityMetadata(entity)
	assert.Nil(t, result)
}

func TestExtractEntityMetadata_WithMetadata(t *testing.T) {
	metadata, err := structpb.NewStruct(map[string]any{
		"source_match": map[string]any{
			"diode_id": "device-123",
		},
		"agent_id": "agent-1",
	})
	require.NoError(t, err)

	device := &diodepb.Device{
		Name:     ptrString("test-device"),
		Metadata: metadata,
	}
	entity := &diodepb.Entity{
		Entity: &diodepb.Entity_Device{Device: device},
	}

	result := ExtractEntityMetadata(entity)

	require.NotNil(t, result)
	assert.Equal(t, "agent-1", result["agent_id"])

	sourceMatch, ok := result["source_match"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "device-123", sourceMatch["diode_id"])
}

func TestExtractEntityMetadata_DifferentEntityTypes(t *testing.T) {
	tests := []struct {
		name     string
		entity   *diodepb.Entity
		hasValue bool
	}{
		{
			name: "Site with metadata",
			entity: func() *diodepb.Entity {
				meta, _ := structpb.NewStruct(map[string]any{"key": "value"})
				return &diodepb.Entity{
					Entity: &diodepb.Entity_Site{Site: &diodepb.Site{
						Name:     "test-site",
						Metadata: meta,
					}},
				}
			}(),
			hasValue: true,
		},
		{
			name: "Interface with metadata",
			entity: func() *diodepb.Entity {
				meta, _ := structpb.NewStruct(map[string]any{"port_id": "eth0"})
				return &diodepb.Entity{
					Entity: &diodepb.Entity_Interface{Interface: &diodepb.Interface{
						Name:     "eth0",
						Metadata: meta,
					}},
				}
			}(),
			hasValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEntityMetadata(tt.entity)
			if tt.hasValue {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestFlattenMetadata_EmptyMap(t *testing.T) {
	result := FlattenMetadata(map[string]any{}, "")
	assert.Empty(t, result)
}

func TestFlattenMetadata_FlatMap(t *testing.T) {
	data := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}

	result := FlattenMetadata(data, "")

	assert.Len(t, result, 2)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "value2", result["key2"])
}

func TestFlattenMetadata_NestedMap(t *testing.T) {
	data := map[string]any{
		"source": map[string]any{
			"id": "123",
		},
	}

	result := FlattenMetadata(data, "")

	assert.Len(t, result, 1)
	assert.Equal(t, "123", result["source.id"])
}

func TestFlattenMetadata_DeeplyNested(t *testing.T) {
	data := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": "deep_value",
			},
		},
	}

	result := FlattenMetadata(data, "")

	assert.Len(t, result, 1)
	assert.Equal(t, "deep_value", result["level1.level2.level3"])
}

func TestFlattenMetadata_MixedNesting(t *testing.T) {
	data := map[string]any{
		"flat_key": "flat_value",
		"nested": map[string]any{
			"inner_key": "inner_value",
		},
	}

	result := FlattenMetadata(data, "")

	assert.Len(t, result, 2)
	assert.Equal(t, "flat_value", result["flat_key"])
	assert.Equal(t, "inner_value", result["nested.inner_key"])
}

func TestFlattenMetadata_WithPrefix(t *testing.T) {
	data := map[string]any{
		"key": "value",
	}

	result := FlattenMetadata(data, "prefix")

	assert.Len(t, result, 1)
	assert.Equal(t, "value", result["prefix.key"])
}

func TestBuildNestedFilter_SingleKey(t *testing.T) {
	result := BuildNestedFilter("key", "value")

	assert.Len(t, result, 1)
	assert.Equal(t, "value", result["key"])
}

func TestBuildNestedFilter_TwoLevels(t *testing.T) {
	result := BuildNestedFilter("source.id", "123")

	assert.Len(t, result, 1)
	source, ok := result["source"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "123", source["id"])
}

func TestBuildNestedFilter_ThreeLevels(t *testing.T) {
	result := BuildNestedFilter("a.b.c", "value")

	assert.Len(t, result, 1)
	a, ok := result["a"].(map[string]any)
	require.True(t, ok)
	b, ok := a["b"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", b["c"])
}

func TestBuildNestedFilter_DifferentValueTypes(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value any
	}{
		{"string value", "key", "string"},
		{"int value", "key", 123},
		{"float value", "key", 1.23},
		{"bool value", "key", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNestedFilter(tt.key, tt.value)
			assert.Equal(t, tt.value, result[tt.key])
		})
	}
}

// Tests for types.go

func TestMatchConfidence_IsHighConfidence(t *testing.T) {
	tests := []struct {
		confidence MatchConfidence
		expected   bool
	}{
		{1.0, true},
		{0.95, true},
		{0.9, true},
		{0.89, false},
		{0.5, false},
		{0.0, false},
	}

	for _, tt := range tests {
		result := tt.confidence.IsHighConfidence()
		assert.Equal(t, tt.expected, result, "confidence %v", tt.confidence)
	}
}

func TestMatchConfidence_IsMediumConfidence(t *testing.T) {
	tests := []struct {
		confidence MatchConfidence
		expected   bool
	}{
		{0.89, true},
		{0.8, true},
		{0.7, true},
		{0.9, false},  // High, not medium
		{0.69, false}, // Low, not medium
		{0.5, false},
	}

	for _, tt := range tests {
		result := tt.confidence.IsMediumConfidence()
		assert.Equal(t, tt.expected, result, "confidence %v", tt.confidence)
	}
}

func TestMatchConfidence_IsLowConfidence(t *testing.T) {
	tests := []struct {
		confidence MatchConfidence
		expected   bool
	}{
		{0.69, true},
		{0.6, true},
		{0.5, true},
		{0.7, false},  // Medium, not low
		{0.49, false}, // None
		{0.0, false},
	}

	for _, tt := range tests {
		result := tt.confidence.IsLowConfidence()
		assert.Equal(t, tt.expected, result, "confidence %v", tt.confidence)
	}
}

func TestMatchConfidence_String(t *testing.T) {
	tests := []struct {
		confidence MatchConfidence
		expected   string
	}{
		{1.0, "High"},
		{0.9, "High"},
		{0.89, "Medium"},
		{0.7, "Medium"},
		{0.69, "Low"},
		{0.5, "Low"},
		{0.49, "None"},
		{0.0, "None"},
	}

	for _, tt := range tests {
		result := tt.confidence.String()
		assert.Equal(t, tt.expected, result, "confidence %v", tt.confidence)
	}
}

func TestEntityMatchingRule_GetRequireAllPrimary(t *testing.T) {
	t.Run("nil value returns default true", func(t *testing.T) {
		rule := &EntityMatchingRule{}
		assert.True(t, rule.GetRequireAllPrimary(true))
	})

	t.Run("nil value returns default false", func(t *testing.T) {
		rule := &EntityMatchingRule{}
		assert.False(t, rule.GetRequireAllPrimary(false))
	})

	t.Run("set to true returns true", func(t *testing.T) {
		rule := &EntityMatchingRule{}
		rule.SetRequireAllPrimary(true)
		assert.True(t, rule.GetRequireAllPrimary(false))
	})

	t.Run("set to false returns false", func(t *testing.T) {
		rule := &EntityMatchingRule{}
		rule.SetRequireAllPrimary(false)
		assert.False(t, rule.GetRequireAllPrimary(true))
	})
}

func TestEntityMatchingRule_SetRequireAllPrimary(t *testing.T) {
	rule := &EntityMatchingRule{}
	assert.Nil(t, rule.RequireAllPrimary)

	rule.SetRequireAllPrimary(true)
	require.NotNil(t, rule.RequireAllPrimary)
	assert.True(t, *rule.RequireAllPrimary)

	rule.SetRequireAllPrimary(false)
	require.NotNil(t, rule.RequireAllPrimary)
	assert.False(t, *rule.RequireAllPrimary)
}

// Tests for config.go

func TestLoadMatchingConfig_EmptyPath(t *testing.T) {
	_, err := LoadMatchingConfig("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file path is required")
}

func TestLoadMatchingConfig_NonExistentFile(t *testing.T) {
	_, err := LoadMatchingConfig("/nonexistent/path/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file not found")
}

func TestLoadMatchingConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0o644)
	require.NoError(t, err)

	_, err = LoadMatchingConfig(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestLoadMatchingConfig_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
global_settings:
  default_min_confidence: 0.8
  default_require_all_primary: true
  enable_fuzzy_matching: true
  default_fuzzy_threshold: 0.85

default_entity_rules:
  Device:
    entity_type: Device
    primary_rules:
      - field_path: name
        match_type: exact
        weight: 1.0
        required: true
        confidence: 0.9
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	config, err := LoadMatchingConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, 0.8, config.GlobalSettings.DefaultMinConfidence)
	assert.True(t, config.GlobalSettings.DefaultRequireAllPrimary)
	assert.True(t, config.GlobalSettings.EnableFuzzyMatching)
	assert.Equal(t, 0.85, config.GlobalSettings.DefaultFuzzyThreshold)

	deviceRule, exists := config.DefaultEntityRules["Device"]
	assert.True(t, exists)
	assert.Len(t, deviceRule.PrimaryRules, 1)
	assert.Equal(t, "name", deviceRule.PrimaryRules[0].FieldPath)
}

func TestConfig_GetFinalRules_Basic(t *testing.T) {
	config := &Config{
		GlobalSettings: GlobalMatchingSettings{
			DefaultMinConfidence:     0.7,
			DefaultRequireAllPrimary: true,
		},
		DefaultEntityRules: map[string]EntityMatchingRule{
			"Device": {
				EntityType: "Device",
				PrimaryRules: []FieldMatchRule{
					{FieldPath: "name", MatchType: MatchExact},
				},
			},
		},
	}

	rules := config.GetFinalRules()

	require.Contains(t, rules, "Device")
	assert.Equal(t, MatchConfidence(0.7), rules["Device"].MinConfidence)
	assert.True(t, rules["Device"].GetRequireAllPrimary(false))
}

func TestConfig_GetFinalRules_SkipsWildcard(t *testing.T) {
	config := &Config{
		DefaultEntityRules: map[string]EntityMatchingRule{
			"*": {
				EntityType: "*",
				PrimaryRules: []FieldMatchRule{
					{FieldPath: "name", MatchType: MatchExact},
				},
			},
			"Device": {
				EntityType: "Device",
			},
		},
	}

	rules := config.GetFinalRules()

	assert.NotContains(t, rules, "*")
	assert.Contains(t, rules, "Device")
}

func TestConfig_GetFinalRules_WithOverrides(t *testing.T) {
	boolTrue := true
	config := &Config{
		GlobalSettings: GlobalMatchingSettings{
			DefaultMinConfidence:     0.7,
			DefaultRequireAllPrimary: false,
		},
		DefaultEntityRules: map[string]EntityMatchingRule{
			"Device": {
				EntityType:    "Device",
				MinConfidence: 0.6,
				PrimaryRules: []FieldMatchRule{
					{FieldPath: "name", MatchType: MatchExact},
				},
			},
		},
		EntityOverrides: map[string]EntityMatchingRule{
			"Device": {
				EntityType:        "Device",
				MinConfidence:     0.9,
				RequireAllPrimary: &boolTrue,
			},
		},
	}

	rules := config.GetFinalRules()

	require.Contains(t, rules, "Device")
	assert.Equal(t, MatchConfidence(0.9), rules["Device"].MinConfidence)
	assert.True(t, rules["Device"].GetRequireAllPrimary(false))
}

func TestConfig_GetFinalRules_NewEntityFromOverrides(t *testing.T) {
	config := &Config{
		DefaultEntityRules: map[string]EntityMatchingRule{},
		EntityOverrides: map[string]EntityMatchingRule{
			"NewEntity": {
				EntityType:    "NewEntity",
				MinConfidence: 0.8,
			},
		},
	}

	rules := config.GetFinalRules()

	require.Contains(t, rules, "NewEntity")
	assert.Equal(t, MatchConfidence(0.8), rules["NewEntity"].MinConfidence)
}

func TestConfig_ApplyWithDefaults(t *testing.T) {
	config := &Config{
		GlobalSettings: GlobalMatchingSettings{
			DefaultMinConfidence: 0.7,
		},
		DefaultEntityRules: map[string]EntityMatchingRule{
			"Device": {EntityType: "Device"},
		},
	}

	// ApplyWithDefaults should ignore the passed map and use YAML config
	rules := config.ApplyWithDefaults(map[string]*EntityMatchingRule{
		"Ignored": {EntityType: "Ignored"},
	})

	assert.Contains(t, rules, "Device")
	assert.NotContains(t, rules, "Ignored")
}

func TestConfig_ApplyOverrides(t *testing.T) {
	config := &Config{
		GlobalSettings: GlobalMatchingSettings{
			DefaultMinConfidence: 0.7,
		},
		DefaultEntityRules: map[string]EntityMatchingRule{
			"Device": {EntityType: "Device"},
		},
	}

	// ApplyOverrides should ignore the passed map and use YAML config
	rules := config.ApplyOverrides(map[string]*EntityMatchingRule{
		"Ignored": {EntityType: "Ignored"},
	})

	assert.Contains(t, rules, "Device")
	assert.NotContains(t, rules, "Ignored")
}

func TestConfig_MergeEntityRule(t *testing.T) {
	config := &Config{}

	t.Run("merges MinConfidence when override is positive", func(t *testing.T) {
		existing := &EntityMatchingRule{MinConfidence: 0.5}
		override := &EntityMatchingRule{MinConfidence: 0.9}

		config.mergeEntityRule(existing, override)

		assert.Equal(t, MatchConfidence(0.9), existing.MinConfidence)
	})

	t.Run("keeps existing MinConfidence when override is zero", func(t *testing.T) {
		existing := &EntityMatchingRule{MinConfidence: 0.5}
		override := &EntityMatchingRule{MinConfidence: 0}

		config.mergeEntityRule(existing, override)

		assert.Equal(t, MatchConfidence(0.5), existing.MinConfidence)
	})

	t.Run("merges RequireAllPrimary when override is not nil", func(t *testing.T) {
		boolFalse := false
		existing := &EntityMatchingRule{}
		existing.SetRequireAllPrimary(true)
		override := &EntityMatchingRule{RequireAllPrimary: &boolFalse}

		config.mergeEntityRule(existing, override)

		assert.False(t, existing.GetRequireAllPrimary(true))
	})

	t.Run("keeps existing RequireAllPrimary when override is nil", func(t *testing.T) {
		existing := &EntityMatchingRule{}
		existing.SetRequireAllPrimary(true)
		override := &EntityMatchingRule{RequireAllPrimary: nil}

		config.mergeEntityRule(existing, override)

		assert.True(t, existing.GetRequireAllPrimary(false))
	})

	t.Run("merges PrimaryRules when override has them", func(t *testing.T) {
		existing := &EntityMatchingRule{
			PrimaryRules: []FieldMatchRule{{FieldPath: "old"}},
		}
		override := &EntityMatchingRule{
			PrimaryRules: []FieldMatchRule{{FieldPath: "new"}},
		}

		config.mergeEntityRule(existing, override)

		assert.Len(t, existing.PrimaryRules, 1)
		assert.Equal(t, "new", existing.PrimaryRules[0].FieldPath)
	})

	t.Run("merges SecondaryRules when override has them", func(t *testing.T) {
		existing := &EntityMatchingRule{
			SecondaryRules: []FieldMatchRule{{FieldPath: "old"}},
		}
		override := &EntityMatchingRule{
			SecondaryRules: []FieldMatchRule{{FieldPath: "new"}},
		}

		config.mergeEntityRule(existing, override)

		assert.Len(t, existing.SecondaryRules, 1)
		assert.Equal(t, "new", existing.SecondaryRules[0].FieldPath)
	})

	t.Run("merges FallbackRules when override has them", func(t *testing.T) {
		existing := &EntityMatchingRule{
			FallbackRules: []FieldMatchRule{{FieldPath: "old"}},
		}
		override := &EntityMatchingRule{
			FallbackRules: []FieldMatchRule{{FieldPath: "new"}},
		}

		config.mergeEntityRule(existing, override)

		assert.Len(t, existing.FallbackRules, 1)
		assert.Equal(t, "new", existing.FallbackRules[0].FieldPath)
	})
}
