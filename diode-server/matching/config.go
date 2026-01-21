package matching

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration for entity matching
type Config struct {
	// Global settings
	GlobalSettings GlobalMatchingSettings `yaml:"global_settings"`

	// Default entity rules (YAML-first approach - these are the primary source of truth)
	DefaultEntityRules map[string]EntityMatchingRule `yaml:"default_entity_rules"`

	// Entity-specific overrides (for environment-specific adjustments)
	EntityOverrides map[string]EntityMatchingRule `yaml:"entity_overrides"`
}

// GlobalMatchingSettings defines global matching behavior
type GlobalMatchingSettings struct {
	DefaultMinConfidence     float64 `yaml:"default_min_confidence"`
	DefaultRequireAllPrimary bool    `yaml:"default_require_all_primary"`
	EnableFuzzyMatching      bool    `yaml:"enable_fuzzy_matching"`
	DefaultFuzzyThreshold    float64 `yaml:"default_fuzzy_threshold"`
}

// LoadMatchingConfig loads matching configuration from file
func LoadMatchingConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("configuration file path is required")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return &config, nil
}

// GetFinalRules processes YAML configuration and returns final matching rules
func (c *Config) GetFinalRules() map[string]*EntityMatchingRule {
	result := make(map[string]*EntityMatchingRule)

	// Step 1: Start with YAML default rules (primary source of truth)
	for entityType, rule := range c.DefaultEntityRules {
		if entityType == "*" {
			continue // Skip universal rule for now
		}
		newRule := rule // Copy the rule
		result[entityType] = &newRule
	}

	// Step 2: Apply global settings (only when not explicitly set)
	for _, rule := range result {
		if rule.MinConfidence == 0 {
			rule.MinConfidence = MatchConfidence(c.GlobalSettings.DefaultMinConfidence)
		}

		// Only apply global default if RequireAllPrimary was not explicitly set (nil)
		if rule.RequireAllPrimary == nil {
			rule.SetRequireAllPrimary(c.GlobalSettings.DefaultRequireAllPrimary)
		}
	}

	// Step 3: Apply entity-specific overrides (environment-specific adjustments)
	for entityType, override := range c.EntityOverrides {
		if existing, exists := result[entityType]; exists {
			// Merge the override with existing rule
			c.mergeEntityRule(existing, &override)
		} else {
			// Create new rule from override
			newRule := override
			result[entityType] = &newRule
		}
	}

	return result
}

// ApplyWithDefaults is deprecated - use GetFinalRules() for pure YAML configuration
func (c *Config) ApplyWithDefaults(_ map[string]*EntityMatchingRule) map[string]*EntityMatchingRule {
	// Ignore auto-generated rules - YAML is now the only source of truth
	return c.GetFinalRules()
}

// ApplyOverrides is kept for backward compatibility but delegates to GetFinalRules
func (c *Config) ApplyOverrides(_ map[string]*EntityMatchingRule) map[string]*EntityMatchingRule {
	// Ignore passed rules - YAML is now the only source of truth
	return c.GetFinalRules()
}

// mergeEntityRule merges override settings into an existing entity rule
func (c *Config) mergeEntityRule(existing *EntityMatchingRule, override *EntityMatchingRule) {
	// Override top-level settings
	if override.MinConfidence > 0 {
		existing.MinConfidence = override.MinConfidence
	}

	// Override RequireAllPrimary only if explicitly set in override (not nil)
	if override.RequireAllPrimary != nil {
		existing.RequireAllPrimary = override.RequireAllPrimary
	}

	// Override field rules if provided
	if len(override.PrimaryRules) > 0 {
		existing.PrimaryRules = override.PrimaryRules
	}
	if len(override.SecondaryRules) > 0 {
		existing.SecondaryRules = override.SecondaryRules
	}
	if len(override.FallbackRules) > 0 {
		existing.FallbackRules = override.FallbackRules
	}
}
