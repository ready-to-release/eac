package config

import (
	"fmt"
	"strings"

	security "github.com/ready-to-release/eac/contracts/security/0.1.0/interfaces"
)

// RiskScoringConfig holds configurable risk scoring mappings.
// It implements security.RiskScoringPort.
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

// Verify RiskScoringConfig implements RiskScoringPort.
var _ security.RiskScoringPort = (*RiskScoringConfig)(nil)

// DefaultRiskScoringConfig returns the hardcoded default configuration.
// These values are the canonical defaults for impact and criticality.
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

// GetImpact returns the impact rating for a module type.
// Falls back to _default if the type is not found.
func (c *RiskScoringConfig) GetImpact(moduleType string) int {
	if c == nil || c.Impact == nil {
		return 3 // Ultimate fallback: Medium
	}
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
	if c == nil || c.Criticality == nil {
		return "medium" // Ultimate fallback
	}
	if v, ok := c.Criticality[moduleType]; ok {
		return v
	}
	if v, ok := c.Criticality["_default"]; ok {
		return v
	}
	return "medium" // Ultimate fallback
}

// GetSeverityWeight returns the likelihood increment for a severity level.
// Valid severity levels: critical, high, medium, low.
func (c *RiskScoringConfig) GetSeverityWeight(severity string) int {
	if c == nil {
		return 0
	}
	switch strings.ToLower(severity) {
	case "critical":
		return c.SeverityWeights.Critical
	case "high":
		return c.SeverityWeights.High
	case "medium":
		return c.SeverityWeights.Medium
	case "low":
		return c.SeverityWeights.Low
	default:
		return 0
	}
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

// mergeRiskScoring merges override values into base config.
// Non-zero values in override replace values in base.
func mergeRiskScoring(base, override *RiskScoringConfig) *RiskScoringConfig {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	// Deep copy base
	result := &RiskScoringConfig{
		Impact:          make(map[string]int),
		Criticality:     make(map[string]string),
		SeverityWeights: base.SeverityWeights,
	}

	// Copy base values
	for k, v := range base.Impact {
		result.Impact[k] = v
	}
	for k, v := range base.Criticality {
		result.Criticality[k] = v
	}

	// Merge override values
	for k, v := range override.Impact {
		result.Impact[k] = v
	}
	for k, v := range override.Criticality {
		result.Criticality[k] = v
	}

	// Override severity weights if any are set (non-zero)
	if override.SeverityWeights.Critical != 0 {
		result.SeverityWeights.Critical = override.SeverityWeights.Critical
	}
	if override.SeverityWeights.High != 0 {
		result.SeverityWeights.High = override.SeverityWeights.High
	}
	if override.SeverityWeights.Medium != 0 {
		result.SeverityWeights.Medium = override.SeverityWeights.Medium
	}
	if override.SeverityWeights.Low != 0 {
		result.SeverityWeights.Low = override.SeverityWeights.Low
	}

	return result
}
