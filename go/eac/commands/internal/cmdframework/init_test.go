package cmdframework

import (
	"testing"
)

// TestCalculateMaxConcurrency tests the turbo boost calculation logic.
func TestCalculateMaxConcurrency(t *testing.T) {
	tests := []struct {
		name              string
		configConcurrency int  // MaxConcurrency from CommandConfig
		repoConcurrency   int  // EffectiveParallelism from repo config
		turbo             bool // Turbo flag
		sequential        bool // Sequential flag
		expected          int
	}{
		{
			name:              "default uses repo parallelism",
			configConcurrency: 0,
			repoConcurrency:   4,
			turbo:             false,
			sequential:        false,
			expected:          4,
		},
		{
			name:              "explicit config overrides repo default",
			configConcurrency: 8,
			repoConcurrency:   4,
			turbo:             false,
			sequential:        false,
			expected:          8,
		},
		{
			name:              "turbo adds 2 to default parallelism",
			configConcurrency: 0,
			repoConcurrency:   4,
			turbo:             true,
			sequential:        false,
			expected:          6, // 4 + 2
		},
		{
			name:              "turbo adds 2 to explicit config",
			configConcurrency: 8,
			repoConcurrency:   4,
			turbo:             true,
			sequential:        false,
			expected:          10, // 8 + 2
		},
		{
			name:              "sequential overrides turbo",
			configConcurrency: 0,
			repoConcurrency:   4,
			turbo:             true,
			sequential:        true,
			expected:          1,
		},
		{
			name:              "sequential forces 1",
			configConcurrency: 8,
			repoConcurrency:   4,
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

// TestTurboBoostConstant verifies the turbo boost value is consistent.
func TestTurboBoostConstant(t *testing.T) {
	if TurboBoost != 2 {
		t.Errorf("TurboBoost = %d, want 2", TurboBoost)
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
