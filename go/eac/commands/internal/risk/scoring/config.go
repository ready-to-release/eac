// Package scoring provides risk scoring configuration management.
package scoring

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Global config instance (set during initialization)
var riskConfig *RiskScoringConfig

// RiskScoringConfig holds configurable risk scoring mappings.
type RiskScoringConfig struct {
	// Impact maps module type to default impact rating (1-5)
	Impact map[string]int `yaml:"impact"`

	// Criticality maps module type to criticality level (high/medium/low)
	Criticality map[string]string `yaml:"criticality"`

	// SeverityWeights maps severity level to likelihood increment
	SeverityWeights SeverityWeightsConfig `yaml:"severity_weights"`
}

// SeverityWeightsConfig holds severity-to-likelihood weights.
type SeverityWeightsConfig struct {
	Critical int `yaml:"critical"`
	High     int `yaml:"high"`
	Medium   int `yaml:"medium"`
	Low      int `yaml:"low"`
}

// DefaultRiskScoringConfig returns the hardcoded default configuration.
// These values match the original hardcoded behavior in GetDefaultImpact and determineCriticality.
func DefaultRiskScoringConfig() *RiskScoringConfig {
	return &RiskScoringConfig{
		Impact: map[string]int{
			"api":      4,
			"service":  4,
			"gateway":  4,
			"library":  3,
			"core":     3,
			"cli":      2,
			"tool":     2,
			"docs":     1,
			"config":   1,
			"_default": 3,
		},
		Criticality: map[string]string{
			"api":      "high",
			"gateway":  "high",
			"service":  "high",
			"core":     "medium",
			"library":  "medium",
			"cli":      "low",
			"tool":     "low",
			"_default": "medium",
		},
		SeverityWeights: SeverityWeightsConfig{
			Critical: 4,
			High:     3,
			Medium:   2,
			Low:      1,
		},
	}
}

// InitRiskScoringConfig initializes the global risk scoring configuration.
// Should be called once during application startup.
func InitRiskScoringConfig(repoRoot string) error {
	config, err := LoadRiskScoringConfig(repoRoot)
	if err != nil {
		return err
	}
	riskConfig = config
	return nil
}

// GetRiskScoringConfig returns the current risk scoring configuration.
// Returns default config if not initialized.
func GetRiskScoringConfig() *RiskScoringConfig {
	if riskConfig == nil {
		return DefaultRiskScoringConfig()
	}
	return riskConfig
}

// LoadRiskScoringConfig loads risk scoring configuration with defaults.
// Looks for risk-scoring.yml in contracts/eac-core/defaults/ and .r2r/eac/.
// Falls back to hardcoded defaults if no config file found.
func LoadRiskScoringConfig(repoRoot string) (*RiskScoringConfig, error) {
	// Start with defaults
	config := DefaultRiskScoringConfig()

	// Try to load from contracts defaults
	contractsPath := filepath.Join(repoRoot, "contracts", "eac-core", "0.1.0", "defaults", "risk-scoring.yml")
	if err := loadAndMerge(contractsPath, config); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("loading contracts risk-scoring.yml: %w", err)
	}

	// Try to load project-level overrides
	projectPath := filepath.Join(repoRoot, ".r2r", "eac", "risk-scoring.yml")
	if err := loadAndMerge(projectPath, config); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("loading project risk-scoring.yml: %w", err)
	}

	// Validate the merged config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validating risk scoring config: %w", err)
	}

	return config, nil
}

// loadAndMerge loads a YAML file and merges it into the existing config.
func loadAndMerge(path string, config *RiskScoringConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var override RiskScoringConfig
	if err := yaml.Unmarshal(data, &override); err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	// Merge impact mappings
	for k, v := range override.Impact {
		config.Impact[k] = v
	}

	// Merge criticality mappings
	for k, v := range override.Criticality {
		config.Criticality[k] = v
	}

	// Override severity weights if any are set (non-zero)
	if override.SeverityWeights.Critical != 0 {
		config.SeverityWeights.Critical = override.SeverityWeights.Critical
	}
	if override.SeverityWeights.High != 0 {
		config.SeverityWeights.High = override.SeverityWeights.High
	}
	if override.SeverityWeights.Medium != 0 {
		config.SeverityWeights.Medium = override.SeverityWeights.Medium
	}
	if override.SeverityWeights.Low != 0 {
		config.SeverityWeights.Low = override.SeverityWeights.Low
	}

	return nil
}

// Validate checks that all config values are within valid ranges.
func (c *RiskScoringConfig) Validate() error {
	// Validate impact values (1-5)
	for k, v := range c.Impact {
		if v < 1 || v > 5 {
			return fmt.Errorf("invalid impact value for %q: %d (must be 1-5)", k, v)
		}
	}

	// Validate criticality values
	validCriticality := map[string]bool{"high": true, "medium": true, "low": true}
	for k, v := range c.Criticality {
		if !validCriticality[v] {
			return fmt.Errorf("invalid criticality value for %q: %q (must be high/medium/low)", k, v)
		}
	}

	// Validate severity weights (0-4, since base is 1 and max is 5)
	if c.SeverityWeights.Critical < 0 || c.SeverityWeights.Critical > 4 {
		return fmt.Errorf("invalid critical severity weight: %d (must be 0-4)", c.SeverityWeights.Critical)
	}
	if c.SeverityWeights.High < 0 || c.SeverityWeights.High > 4 {
		return fmt.Errorf("invalid high severity weight: %d (must be 0-4)", c.SeverityWeights.High)
	}
	if c.SeverityWeights.Medium < 0 || c.SeverityWeights.Medium > 4 {
		return fmt.Errorf("invalid medium severity weight: %d (must be 0-4)", c.SeverityWeights.Medium)
	}
	if c.SeverityWeights.Low < 0 || c.SeverityWeights.Low > 4 {
		return fmt.Errorf("invalid low severity weight: %d (must be 0-4)", c.SeverityWeights.Low)
	}

	return nil
}

// GetImpact returns the impact rating for a module type.
// Falls back to _default if the type is not found.
func (c *RiskScoringConfig) GetImpact(moduleType string) int {
	if v, ok := c.Impact[moduleType]; ok {
		return v
	}
	if v, ok := c.Impact["_default"]; ok {
		return v
	}
	return 3 // Ultimate fallback: Medium
}

// GetCriticality returns the criticality level for a module type.
// Falls back to _default if the type is not found.
func (c *RiskScoringConfig) GetCriticality(moduleType string) string {
	if v, ok := c.Criticality[moduleType]; ok {
		return v
	}
	if v, ok := c.Criticality["_default"]; ok {
		return v
	}
	return "medium" // Ultimate fallback
}
