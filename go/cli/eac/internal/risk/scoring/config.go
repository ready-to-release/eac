// Package scoring provides risk scoring configuration management.
//
// DEPRECATED: This package delegates to github.com/ready-to-release/eac/go/core/config.RiskConfig.
// New code should use core/config.LoadRiskConfig() directly.
package scoring

import (
	coreConfig "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
)

// Global config instance (set during initialization)
var riskConfig *RiskScoringConfig

// RiskScoringConfig holds configurable risk scoring mappings.
type RiskScoringConfig struct {
	Impact          map[string]int           `yaml:"impact"`
	Criticality     map[string]string        `yaml:"criticality"`
	SeverityWeights SeverityWeightsConfig    `yaml:"severity_weights"`
}

// SeverityWeightsConfig holds severity-to-likelihood weights.
type SeverityWeightsConfig struct {
	Critical int `yaml:"critical"`
	High     int `yaml:"high"`
	Medium   int `yaml:"medium"`
	Low      int `yaml:"low"`
}

// DefaultRiskScoringConfig returns the default configuration.
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
func InitRiskScoringConfig(repoRoot string) error {
	config, err := LoadRiskScoringConfig(repoRoot)
	if err != nil {
		return err
	}
	riskConfig = config
	return nil
}

// GetRiskScoringConfig returns the current risk scoring configuration.
func GetRiskScoringConfig() *RiskScoringConfig {
	if riskConfig == nil {
		return DefaultRiskScoringConfig()
	}
	return riskConfig
}

// LoadRiskScoringConfig loads risk scoring configuration.
// Delegates to core/config.LoadRiskConfig().
func LoadRiskScoringConfig(repoRoot string) (*RiskScoringConfig, error) {
	configRoot := paths.EACConfigPath(repoRoot)
	riskCfg, err := coreConfig.LoadRiskConfig(repoRoot, configRoot)
	if err != nil {
		return DefaultRiskScoringConfig(), nil
	}

	scoring := riskCfg.GetScoring()
	if scoring == nil {
		return DefaultRiskScoringConfig(), nil
	}

	return wrapRiskScoringPort(scoring), nil
}

// wrapRiskScoringPort wraps RiskScoringPort in the old struct type.
func wrapRiskScoringPort(port interface{}) *RiskScoringConfig {
	// Type assert to concrete type to access maps directly
	if coreCfg, ok := port.(*coreConfig.RiskScoringConfig); ok {
		config := &RiskScoringConfig{
			Impact:      make(map[string]int),
			Criticality: make(map[string]string),
			SeverityWeights: SeverityWeightsConfig{
				Critical: coreCfg.GetSeverityWeight("critical"),
				High:     coreCfg.GetSeverityWeight("high"),
				Medium:   coreCfg.GetSeverityWeight("medium"),
				Low:      coreCfg.GetSeverityWeight("low"),
			},
		}
		// Copy all map entries directly
		for k, v := range coreCfg.Impact {
			config.Impact[k] = v
		}
		for k, v := range coreCfg.Criticality {
			config.Criticality[k] = v
		}
		return config
	}

	return DefaultRiskScoringConfig()
}

// GetImpact returns the impact rating for a module type.
func (c *RiskScoringConfig) GetImpact(moduleType string) int {
	if v, ok := c.Impact[moduleType]; ok {
		return v
	}
	if v, ok := c.Impact["_default"]; ok {
		return v
	}
	return 3
}

// GetCriticality returns the criticality level for a module type.
func (c *RiskScoringConfig) GetCriticality(moduleType string) string {
	if v, ok := c.Criticality[moduleType]; ok {
		return v
	}
	if v, ok := c.Criticality["_default"]; ok {
		return v
	}
	return "medium"
}
