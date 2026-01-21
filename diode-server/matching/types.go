package matching

import (
	"context"
	"encoding/json"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

// MatchConfidence represents the confidence level of an entity match
type MatchConfidence float64

const (
	// ConfidenceHigh is the threshold for high confidence matches (>= 0.9)
	ConfidenceHigh MatchConfidence = 0.9
	// ConfidenceMedium is the threshold for medium confidence matches (0.7-0.89)
	ConfidenceMedium MatchConfidence = 0.7
	// ConfidenceLow is the threshold for low confidence matches (0.5-0.69)
	ConfidenceLow MatchConfidence = 0.5
	// ConfidenceNone indicates no match (< 0.5)
	ConfidenceNone MatchConfidence = 0.0
)

// MatchResult represents the result of an entity matching operation
type MatchResult struct {
	NodeID         *int64          `json:"node_id,omitempty"`
	ExternalID     *string         `json:"external_id,omitempty"`
	Confidence     MatchConfidence `json:"confidence"`
	MatchingFields []string        `json:"matching_fields"`
	MatchReason    string          `json:"match_reason"`
	ExistingData   json.RawMessage `json:"existing_data,omitempty"`
}

// FieldMatchRule defines how a specific field should be matched
type FieldMatchRule struct {
	FieldPath    string          `json:"field_path" yaml:"field_path"`                 // JSON path to the field (e.g., "name", "device_type.name")
	MatchType    FieldMatchType  `json:"match_type" yaml:"match_type"`                 // How to match this field
	Weight       float64         `json:"weight" yaml:"weight"`                         // Importance weight (0.0-1.0)
	Required     bool            `json:"required" yaml:"required"`                     // If true, match fails if this field doesn't match
	Confidence   MatchConfidence `json:"confidence" yaml:"confidence"`                 // Confidence contribution when this field matches
	FuzzyOptions *FuzzyOptions   `json:"fuzzy_options,omitempty" yaml:"fuzzy_options"` // Options for fuzzy matching
}

// FieldMatchType defines different ways to match field values
type FieldMatchType string

const (
	// MatchExact performs exact string/value matching
	MatchExact FieldMatchType = "exact"
	// MatchFuzzy performs fuzzy string matching
	MatchFuzzy FieldMatchType = "fuzzy"
	// MatchContains checks if one string contains another
	MatchContains FieldMatchType = "contains"
	// MatchNumeric performs numeric comparison with tolerance
	MatchNumeric FieldMatchType = "numeric"
	// MatchExists checks if field exists (not null/empty)
	MatchExists FieldMatchType = "exists"
	// MatchRegex performs regular expression matching
	MatchRegex FieldMatchType = "regex"
)

// FuzzyOptions contains options for fuzzy matching
type FuzzyOptions struct {
	MinSimilarity float64 `json:"min_similarity" yaml:"min_similarity"` // Minimum similarity ratio (0.0-1.0)
	CaseIgnore    bool    `json:"case_ignore" yaml:"case_ignore"`       // Ignore case differences
	Normalize     bool    `json:"normalize" yaml:"normalize"`           // Normalize whitespace and punctuation
}

// EntityMatchingRule defines how to match entities of a specific type
type EntityMatchingRule struct {
	EntityType         string           `json:"entity_type" yaml:"entity_type"`                                       // Type of entity (e.g., "Device", "Interface")
	PrimaryRules       []FieldMatchRule `json:"primary_rules" yaml:"primary_rules"`                                   // Primary matching criteria
	SecondaryRules     []FieldMatchRule `json:"secondary_rules" yaml:"secondary_rules"`                               // Secondary matching criteria (lower confidence)
	FallbackRules      []FieldMatchRule `json:"fallback_rules" yaml:"fallback_rules"`                                 // Fallback matching criteria (lowest confidence)
	MinConfidence      MatchConfidence  `json:"min_confidence" yaml:"min_confidence"`                                 // Minimum confidence threshold for this entity type
	RequireAllPrimary  *bool            `json:"require_all_primary,omitempty" yaml:"require_all_primary,omitempty"`   // If true, all primary rules must match; nil means use global default
	EdgePropertyFields []string         `json:"edge_property_fields,omitempty" yaml:"edge_property_fields,omitempty"` // Fields to store as edge properties instead of node properties
}

// GetRequireAllPrimary returns the RequireAllPrimary value, with a fallback default
func (r *EntityMatchingRule) GetRequireAllPrimary(defaultValue bool) bool {
	if r.RequireAllPrimary == nil {
		return defaultValue
	}
	return *r.RequireAllPrimary
}

// SetRequireAllPrimary sets the RequireAllPrimary value
func (r *EntityMatchingRule) SetRequireAllPrimary(value bool) {
	r.RequireAllPrimary = &value
}

// EntityMatchingConfig contains all matching rules and global configuration
type EntityMatchingConfig struct {
	Rules          map[string]*EntityMatchingRule `json:"rules"`           // Rules by entity type
	GlobalMinConf  MatchConfidence                `json:"global_min_conf"` // Global minimum confidence
	EnableFallback bool                           `json:"enable_fallback"` // Enable fallback matching
	CacheResults   bool                           `json:"cache_results"`   // Cache matching results
	MaxCacheSize   int                            `json:"max_cache_size"`  // Maximum cache entries
}

// EntityMatcher interface defines methods for entity matching with confidence scoring
type EntityMatcher interface {
	// FindMatches finds potential matches for an entity with confidence scores
	FindMatches(ctx context.Context, entity *diodepb.Entity) ([]MatchResult, error)

	// FindBestMatch finds the best match for an entity above the confidence threshold
	FindBestMatch(ctx context.Context, entity *diodepb.Entity) (*MatchResult, error)

	// GetMatchingRule returns the matching rule for a specific entity type
	GetMatchingRule(entityType string) (*EntityMatchingRule, error)

	// UpdateMatchingRule updates or adds a matching rule for an entity type
	UpdateMatchingRule(entityType string, rule *EntityMatchingRule) error
}

// IsHighConfidence returns true if the confidence is high enough to treat as same entity
func (c MatchConfidence) IsHighConfidence() bool {
	return c >= ConfidenceHigh
}

// IsMediumConfidence returns true if the confidence requires manual review
func (c MatchConfidence) IsMediumConfidence() bool {
	return c >= ConfidenceMedium && c < ConfidenceHigh
}

// IsLowConfidence returns true if the confidence indicates a potential match
func (c MatchConfidence) IsLowConfidence() bool {
	return c >= ConfidenceLow && c < ConfidenceMedium
}

// String returns a human-readable description of the confidence level
func (c MatchConfidence) String() string {
	switch {
	case c >= ConfidenceHigh:
		return "High"
	case c >= ConfidenceMedium:
		return "Medium"
	case c >= ConfidenceLow:
		return "Low"
	default:
		return "None"
	}
}
