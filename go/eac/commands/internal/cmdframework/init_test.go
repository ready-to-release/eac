package cmdframework

import (
	"testing"
)

// TestCalculateMaxConcurrency tests the max concurrency calculation logic.
// Note: turbo no longer affects MaxConcurrency - it's handled by TurboMultiplier.
// MaxConcurrency returns 0 for dynamic mode (orchestrator calculates from CPU×RAM).
func TestCalculateMaxConcurrency(t *testing.T) {
	tests := []struct {
		name              string
		configConcurrency int  // MaxConcurrency from CommandConfig
		repoConcurrency   int  // EffectiveParallelism from repo config
		turbo             bool // Turbo flag (no longer affects MaxConcurrency)
		sequential        bool // Sequential flag
		expected          int
	}{
		{
			name:              "default returns 0 for dynamic mode",
			configConcurrency: 0,
			repoConcurrency:   0,
			turbo:             false,
			sequential:        false,
			expected:          0, // Dynamic - orchestrator calculates
		},
		{
			name:              "explicit config is used as ceiling",
			configConcurrency: 8,
			repoConcurrency:   0,
			turbo:             false,
			sequential:        false,
			expected:          8,
		},
		{
			name:              "turbo does not affect MaxConcurrency",
			configConcurrency: 0,
			repoConcurrency:   0,
			turbo:             true,
			sequential:        false,
			expected:          0, // Still dynamic - turbo handled separately
		},
		{
			name:              "sequential overrides everything",
			configConcurrency: 0,
			repoConcurrency:   0,
			turbo:             true,
			sequential:        true,
			expected:          1,
		},
		{
			name:              "sequential forces 1 even with explicit config",
			configConcurrency: 8,
			repoConcurrency:   0,
			turbo:             false,
			sequential:        true,
			expected:          1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateMaxConcurrency(tt.configConcurrency, tt.repoConcurrency, tt.turbo, tt.sequential)
			if result != tt.expected {
				t.Errorf("CalculateMaxConcurrency(%d, %d, %v, %v) = %d, want %d",
					tt.configConcurrency, tt.repoConcurrency, tt.turbo, tt.sequential, result, tt.expected)
			}
		})
	}
}

// TestDefaultTurboMultiplier verifies the turbo multiplier value is consistent.
func TestDefaultTurboMultiplier(t *testing.T) {
	if DefaultTurboMultiplier != 4 {
		t.Errorf("DefaultTurboMultiplier = %d, want 4", DefaultTurboMultiplier)
	}
}

// TestCalculateTurboMultiplier verifies turbo multiplier calculation.
func TestCalculateTurboMultiplier(t *testing.T) {
	if CalculateTurboMultiplier(false) != 1 {
		t.Errorf("CalculateTurboMultiplier(false) = %d, want 1", CalculateTurboMultiplier(false))
	}
	if CalculateTurboMultiplier(true) != DefaultTurboMultiplier {
		t.Errorf("CalculateTurboMultiplier(true) = %d, want %d", CalculateTurboMultiplier(true), DefaultTurboMultiplier)
	}
}

// TestCommandConfigTurboMutualExclusivity documents the expected behavior
// when both Turbo and Sequential are set.
func TestCommandConfigTurboMutualExclusivity(t *testing.T) {
	cfg := &CommandConfig{
		Turbo:      true,
		Sequential: true,
	}

	defaultConcurrency := 4

	// Sequential should take precedence
	result := CalculateMaxConcurrency(cfg.MaxConcurrency, defaultConcurrency, cfg.Turbo, cfg.Sequential)
	if result != 1 {
		t.Errorf("Sequential should override Turbo: got %d, want 1", result)
	}
}
