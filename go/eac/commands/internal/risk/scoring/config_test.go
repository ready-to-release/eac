package scoring

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultRiskScoringConfig(t *testing.T) {
	config := DefaultRiskScoringConfig()

	// Verify impact mappings match original hardcoded values from GetDefaultImpact
	impactTests := []struct {
		moduleType string
		want       int
	}{
		{"api", 4},
		{"service", 4},
		{"gateway", 4},
		{"library", 3},
		{"core", 3},
		{"cli", 2},
		{"tool", 2},
		{"docs", 1},
		{"config", 1},
		{"_default", 3},
	}

	for _, tt := range impactTests {
		if got := config.Impact[tt.moduleType]; got != tt.want {
			t.Errorf("Impact[%q] = %d, want %d", tt.moduleType, got, tt.want)
		}
	}

	// Verify criticality mappings match original hardcoded values from determineCriticality
	criticalityTests := []struct {
		moduleType string
		want       string
	}{
		{"api", "high"},
		{"gateway", "high"},
		{"service", "high"},
		{"core", "medium"},
		{"library", "medium"},
		{"cli", "low"},
		{"tool", "low"},
		{"_default", "medium"},
	}

	for _, tt := range criticalityTests {
		if got := config.Criticality[tt.moduleType]; got != tt.want {
			t.Errorf("Criticality[%q] = %q, want %q", tt.moduleType, got, tt.want)
		}
	}

	// Verify severity weights
	if config.SeverityWeights.Critical != 4 {
		t.Errorf("SeverityWeights.Critical = %d, want 4", config.SeverityWeights.Critical)
	}
	if config.SeverityWeights.High != 3 {
		t.Errorf("SeverityWeights.High = %d, want 3", config.SeverityWeights.High)
	}
	if config.SeverityWeights.Medium != 2 {
		t.Errorf("SeverityWeights.Medium = %d, want 2", config.SeverityWeights.Medium)
	}
	if config.SeverityWeights.Low != 1 {
		t.Errorf("SeverityWeights.Low = %d, want 1", config.SeverityWeights.Low)
	}
}

func TestRiskScoringConfig_GetImpact(t *testing.T) {
	config := DefaultRiskScoringConfig()

	tests := []struct {
		moduleType string
		want       int
	}{
		{"api", 4},
		{"service", 4},
		{"gateway", 4},
		{"library", 3},
		{"core", 3},
		{"cli", 2},
		{"tool", 2},
		{"docs", 1},
		{"config", 1},
		{"unknown", 3},        // Falls back to _default
		{"", 3},               // Falls back to _default
		{"microservice", 3},   // Unknown type, falls back to _default
	}

	for _, tt := range tests {
		t.Run(tt.moduleType, func(t *testing.T) {
			got := config.GetImpact(tt.moduleType)
			if got != tt.want {
				t.Errorf("GetImpact(%q) = %d, want %d", tt.moduleType, got, tt.want)
			}
		})
	}
}

func TestRiskScoringConfig_GetCriticality(t *testing.T) {
	config := DefaultRiskScoringConfig()

	tests := []struct {
		moduleType string
		want       string
	}{
		{"api", "high"},
		{"gateway", "high"},
		{"service", "high"},
		{"core", "medium"},
		{"library", "medium"},
		{"cli", "low"},
		{"tool", "low"},
		{"unknown", "medium"},      // Falls back to _default
		{"", "medium"},             // Falls back to _default
		{"microservice", "medium"}, // Unknown type, falls back to _default
	}

	for _, tt := range tests {
		t.Run(tt.moduleType, func(t *testing.T) {
			got := config.GetCriticality(tt.moduleType)
			if got != tt.want {
				t.Errorf("GetCriticality(%q) = %q, want %q", tt.moduleType, got, tt.want)
			}
		})
	}
}

func TestRiskScoringConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *RiskScoringConfig
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultRiskScoringConfig(),
			wantErr: false,
		},
		{
			name: "impact too low",
			config: &RiskScoringConfig{
				Impact:      map[string]int{"api": 0},
				Criticality: map[string]string{},
				SeverityWeights: SeverityWeightsConfig{
					Critical: 4,
					High:     3,
					Medium:   2,
					Low:      1,
				},
			},
			wantErr: true,
		},
		{
			name: "impact too high",
			config: &RiskScoringConfig{
				Impact:      map[string]int{"api": 6},
				Criticality: map[string]string{},
				SeverityWeights: SeverityWeightsConfig{
					Critical: 4,
					High:     3,
					Medium:   2,
					Low:      1,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid criticality value",
			config: &RiskScoringConfig{
				Impact:      map[string]int{},
				Criticality: map[string]string{"api": "extreme"},
				SeverityWeights: SeverityWeightsConfig{
					Critical: 4,
					High:     3,
					Medium:   2,
					Low:      1,
				},
			},
			wantErr: true,
		},
		{
			name: "severity weight too high",
			config: &RiskScoringConfig{
				Impact:      map[string]int{},
				Criticality: map[string]string{},
				SeverityWeights: SeverityWeightsConfig{
					Critical: 5, // Invalid: must be 0-4
					High:     3,
					Medium:   2,
					Low:      1,
				},
			},
			wantErr: true,
		},
		{
			name: "severity weight negative",
			config: &RiskScoringConfig{
				Impact:      map[string]int{},
				Criticality: map[string]string{},
				SeverityWeights: SeverityWeightsConfig{
					Critical: 4,
					High:     -1, // Invalid: must be >= 0
					Medium:   2,
					Low:      1,
				},
			},
			wantErr: true,
		},
		{
			name: "valid custom config",
			config: &RiskScoringConfig{
				Impact:      map[string]int{"microservice": 5, "batch-job": 2},
				Criticality: map[string]string{"microservice": "high", "batch-job": "low"},
				SeverityWeights: SeverityWeightsConfig{
					Critical: 4,
					High:     3,
					Medium:   2,
					Low:      1,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadRiskScoringConfig_NoFiles(t *testing.T) {
	// Load from a directory with no config files
	// Should return default config
	tmpDir := t.TempDir()

	config, err := LoadRiskScoringConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadRiskScoringConfig() error = %v", err)
	}

	// Should match defaults
	if config.GetImpact("api") != 4 {
		t.Errorf("GetImpact(api) = %d, want 4", config.GetImpact("api"))
	}
	if config.GetCriticality("api") != "high" {
		t.Errorf("GetCriticality(api) = %q, want high", config.GetCriticality("api"))
	}
}

func TestLoadRiskScoringConfig_WithOverrides(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .r2r/eac directory
	r2rDir := filepath.Join(tmpDir, ".r2r", "eac")
	if err := os.MkdirAll(r2rDir, 0755); err != nil {
		t.Fatalf("Failed to create .r2r/eac directory: %v", err)
	}

	// Write override config
	overrideYAML := `
impact:
  microservice: 5
  api: 5  # Override default
criticality:
  microservice: high
  cli: medium  # Override default
`
	overridePath := filepath.Join(r2rDir, "risk-scoring.yml")
	if err := os.WriteFile(overridePath, []byte(overrideYAML), 0644); err != nil {
		t.Fatalf("Failed to write override config: %v", err)
	}

	config, err := LoadRiskScoringConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadRiskScoringConfig() error = %v", err)
	}

	// Check overridden values
	if config.GetImpact("microservice") != 5 {
		t.Errorf("GetImpact(microservice) = %d, want 5", config.GetImpact("microservice"))
	}
	if config.GetImpact("api") != 5 {
		t.Errorf("GetImpact(api) = %d, want 5 (overridden)", config.GetImpact("api"))
	}
	if config.GetCriticality("microservice") != "high" {
		t.Errorf("GetCriticality(microservice) = %q, want high", config.GetCriticality("microservice"))
	}
	if config.GetCriticality("cli") != "medium" {
		t.Errorf("GetCriticality(cli) = %q, want medium (overridden)", config.GetCriticality("cli"))
	}

	// Check non-overridden values are preserved
	if config.GetImpact("service") != 4 {
		t.Errorf("GetImpact(service) = %d, want 4 (default)", config.GetImpact("service"))
	}
	if config.GetCriticality("gateway") != "high" {
		t.Errorf("GetCriticality(gateway) = %q, want high (default)", config.GetCriticality("gateway"))
	}
}

func TestLoadRiskScoringConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .r2r/eac directory
	r2rDir := filepath.Join(tmpDir, ".r2r", "eac")
	if err := os.MkdirAll(r2rDir, 0755); err != nil {
		t.Fatalf("Failed to create .r2r/eac directory: %v", err)
	}

	// Write invalid YAML
	invalidYAML := `
impact:
  api: "not a number"  # Should be int
`
	overridePath := filepath.Join(r2rDir, "risk-scoring.yml")
	if err := os.WriteFile(overridePath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write override config: %v", err)
	}

	_, err := LoadRiskScoringConfig(tmpDir)
	if err == nil {
		t.Error("LoadRiskScoringConfig() should return error for invalid YAML")
	}
}

func TestLoadRiskScoringConfig_InvalidValues(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .r2r/eac directory
	r2rDir := filepath.Join(tmpDir, ".r2r", "eac")
	if err := os.MkdirAll(r2rDir, 0755); err != nil {
		t.Fatalf("Failed to create .r2r/eac directory: %v", err)
	}

	// Write config with invalid impact value
	invalidConfig := `
impact:
  api: 10  # Invalid: must be 1-5
`
	overridePath := filepath.Join(r2rDir, "risk-scoring.yml")
	if err := os.WriteFile(overridePath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to write override config: %v", err)
	}

	_, err := LoadRiskScoringConfig(tmpDir)
	if err == nil {
		t.Error("LoadRiskScoringConfig() should return error for invalid impact value")
	}
}

func TestGetRiskScoringConfig_DefaultWhenNotInitialized(t *testing.T) {
	// Reset global config
	riskConfig = nil

	config := GetRiskScoringConfig()
	if config == nil {
		t.Fatal("GetRiskScoringConfig() returned nil")
	}

	// Should return defaults
	if config.GetImpact("api") != 4 {
		t.Errorf("GetImpact(api) = %d, want 4", config.GetImpact("api"))
	}
}

func TestInitRiskScoringConfig(t *testing.T) {
	// Reset global config
	riskConfig = nil

	tmpDir := t.TempDir()

	err := InitRiskScoringConfig(tmpDir)
	if err != nil {
		t.Fatalf("InitRiskScoringConfig() error = %v", err)
	}

	// Now GetRiskScoringConfig should return the initialized config
	config := GetRiskScoringConfig()
	if config == nil {
		t.Fatal("GetRiskScoringConfig() returned nil after init")
	}

	// Reset for other tests
	riskConfig = nil
}

// TestBackwardCompatibility ensures the config system doesn't change behavior
// for existing code that calls GetDefaultImpact and determineCriticality
func TestBackwardCompatibility_GetDefaultImpact(t *testing.T) {
	// Reset global config to ensure we use defaults
	riskConfig = nil

	// These are the exact expectations from the original TestGetDefaultImpact test
	tests := []struct {
		moduleType string
		want       int
	}{
		{"service", 4},
		{"library", 3},
		{"unknown", 3},
		{"", 3},
	}

	for _, tt := range tests {
		t.Run(tt.moduleType, func(t *testing.T) {
			// This tests GetDefaultImpact which will use the config system
			got := GetDefaultImpact(tt.moduleType)
			if got != tt.want {
				t.Errorf("GetDefaultImpact(%q) = %d, want %d", tt.moduleType, got, tt.want)
			}
		})
	}
}
