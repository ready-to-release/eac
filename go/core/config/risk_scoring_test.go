package config

import (
	"testing"

	security "github.com/ready-to-release/eac/contracts/security/0.1.0/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRiskScoringConfig_ImplementsPort(t *testing.T) {
	var _ security.RiskScoringPort = (*RiskScoringConfig)(nil)
}

func TestDefaultRiskScoringConfig(t *testing.T) {
	cfg := DefaultRiskScoringConfig()

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Impact)
	assert.NotNil(t, cfg.Criticality)

	// Check some default values
	assert.Equal(t, 4, cfg.Impact["api"])
	assert.Equal(t, 1, cfg.Impact["docs"])
	assert.Equal(t, 3, cfg.Impact["_default"])

	assert.Equal(t, "high", cfg.Criticality["api"])
	assert.Equal(t, "low", cfg.Criticality["cli"])
	assert.Equal(t, "medium", cfg.Criticality["_default"])

	assert.Equal(t, 4, cfg.SeverityWeights.Critical)
	assert.Equal(t, 3, cfg.SeverityWeights.High)
	assert.Equal(t, 2, cfg.SeverityWeights.Medium)
	assert.Equal(t, 1, cfg.SeverityWeights.Low)
}

func TestRiskScoringConfig_GetImpact(t *testing.T) {
	cfg := DefaultRiskScoringConfig()

	tests := []struct {
		name       string
		moduleType string
		expected   int
	}{
		{"api module", "api", 4},
		{"service module", "service", 4},
		{"gateway module", "gateway", 4},
		{"library module", "library", 3},
		{"core module", "core", 3},
		{"cli module", "cli", 2},
		{"tool module", "tool", 2},
		{"docs module", "docs", 1},
		{"config module", "config", 1},
		{"unknown uses default", "unknown", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetImpact(tt.moduleType)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestRiskScoringConfig_GetImpact_NilConfig(t *testing.T) {
	var cfg *RiskScoringConfig
	assert.Equal(t, 3, cfg.GetImpact("any"))
}

func TestRiskScoringConfig_GetCriticality(t *testing.T) {
	cfg := DefaultRiskScoringConfig()

	tests := []struct {
		name     string
		moniker  string
		expected string
	}{
		{"api module", "api", "high"},
		{"gateway module", "gateway", "high"},
		{"service module", "service", "high"},
		{"core module", "core", "medium"},
		{"library module", "library", "medium"},
		{"cli module", "cli", "low"},
		{"tool module", "tool", "low"},
		{"unknown uses default", "unknown", "medium"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetCriticality(tt.moniker)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestRiskScoringConfig_GetCriticality_NilConfig(t *testing.T) {
	var cfg *RiskScoringConfig
	assert.Equal(t, "medium", cfg.GetCriticality("any"))
}

func TestRiskScoringConfig_GetSeverityWeight(t *testing.T) {
	cfg := DefaultRiskScoringConfig()

	tests := []struct {
		name     string
		severity string
		expected int
	}{
		{"critical", "critical", 4},
		{"high", "high", 3},
		{"medium", "medium", 2},
		{"low", "low", 1},
		{"Critical uppercase", "Critical", 4},
		{"HIGH uppercase", "HIGH", 3},
		{"unknown returns 0", "unknown", 0},
		{"empty returns 0", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetSeverityWeight(tt.severity)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestRiskScoringConfig_GetSeverityWeight_NilConfig(t *testing.T) {
	var cfg *RiskScoringConfig
	assert.Equal(t, 0, cfg.GetSeverityWeight("critical"))
}

func TestRiskScoringConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *RiskScoringConfig
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid default config",
			cfg:       DefaultRiskScoringConfig(),
			wantError: false,
		},
		{
			name: "invalid impact value too low",
			cfg: &RiskScoringConfig{
				Impact:      map[string]int{"api": 0},
				Criticality: map[string]string{},
			},
			wantError: true,
			errorMsg:  "invalid impact value",
		},
		{
			name: "invalid impact value too high",
			cfg: &RiskScoringConfig{
				Impact:      map[string]int{"api": 6},
				Criticality: map[string]string{},
			},
			wantError: true,
			errorMsg:  "invalid impact value",
		},
		{
			name: "invalid criticality value",
			cfg: &RiskScoringConfig{
				Impact:      map[string]int{},
				Criticality: map[string]string{"api": "critical"},
			},
			wantError: true,
			errorMsg:  "invalid criticality value",
		},
		{
			name: "invalid severity weight too high",
			cfg: &RiskScoringConfig{
				Impact:      map[string]int{},
				Criticality: map[string]string{},
				SeverityWeights: SeverityWeightsConfig{
					Critical: 5,
				},
			},
			wantError: true,
			errorMsg:  "invalid critical severity weight",
		},
		{
			name: "invalid severity weight negative",
			cfg: &RiskScoringConfig{
				Impact:      map[string]int{},
				Criticality: map[string]string{},
				SeverityWeights: SeverityWeightsConfig{
					Low: -1,
				},
			},
			wantError: true,
			errorMsg:  "invalid low severity weight",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMergeRiskScoring(t *testing.T) {
	t.Run("nil base returns override", func(t *testing.T) {
		override := &RiskScoringConfig{
			Impact: map[string]int{"api": 5},
		}
		result := mergeRiskScoring(nil, override)
		assert.Equal(t, override, result)
	})

	t.Run("nil override returns base", func(t *testing.T) {
		base := DefaultRiskScoringConfig()
		result := mergeRiskScoring(base, nil)
		assert.Equal(t, base, result)
	})

	t.Run("override replaces base values", func(t *testing.T) {
		base := DefaultRiskScoringConfig()
		override := &RiskScoringConfig{
			Impact: map[string]int{
				"api":  5, // Override existing
				"test": 2, // Add new
			},
			Criticality: map[string]string{
				"api": "low", // Override existing
			},
			SeverityWeights: SeverityWeightsConfig{
				Critical: 3, // Override
			},
		}

		result := mergeRiskScoring(base, override)

		// Overridden values
		assert.Equal(t, 5, result.Impact["api"])
		assert.Equal(t, "low", result.Criticality["api"])
		assert.Equal(t, 3, result.SeverityWeights.Critical)

		// Preserved base values
		assert.Equal(t, 4, result.Impact["service"])
		assert.Equal(t, "high", result.Criticality["gateway"])
		assert.Equal(t, 3, result.SeverityWeights.High)

		// New values
		assert.Equal(t, 2, result.Impact["test"])
	})

	t.Run("zero severity weights not merged", func(t *testing.T) {
		base := DefaultRiskScoringConfig()
		override := &RiskScoringConfig{
			Impact:      map[string]int{},
			Criticality: map[string]string{},
			SeverityWeights: SeverityWeightsConfig{
				Critical: 0, // Zero should not override
				High:     0,
				Medium:   0,
				Low:      0,
			},
		}

		result := mergeRiskScoring(base, override)

		// Base values preserved
		assert.Equal(t, 4, result.SeverityWeights.Critical)
		assert.Equal(t, 3, result.SeverityWeights.High)
		assert.Equal(t, 2, result.SeverityWeights.Medium)
		assert.Equal(t, 1, result.SeverityWeights.Low)
	})
}
