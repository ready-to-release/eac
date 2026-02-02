package scoring

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultRiskScoringConfig(t *testing.T) {
	config := DefaultRiskScoringConfig()

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
		{"unknown", 3},
		{"", 3},
		{"microservice", 3},
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
		{"unknown", "medium"},
		{"", "medium"},
		{"microservice", "medium"},
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

func TestLoadRiskScoringConfig_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	config, err := LoadRiskScoringConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadRiskScoringConfig() error = %v", err)
	}

	if config.GetImpact("api") != 4 {
		t.Errorf("GetImpact(api) = %d, want 4", config.GetImpact("api"))
	}
	if config.GetCriticality("api") != "high" {
		t.Errorf("GetCriticality(api) = %q, want high", config.GetCriticality("api"))
	}
}

func TestLoadRiskScoringConfig_WithOverrides(t *testing.T) {
	tmpDir := t.TempDir()

	// Create contracts/security/0.1.0/defaults directory
	contractsDir := filepath.Join(tmpDir, "contracts", "security", "0.1.0", "defaults")
	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("Failed to create contracts directory: %v", err)
	}

	// Write risk-config.yml with overrides
	configYAML := `
scoring:
  impact:
    microservice: 5
    api: 5
  criticality:
    microservice: high
    cli: medium
`
	configPath := filepath.Join(contractsDir, "risk-config.yml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	config, err := LoadRiskScoringConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadRiskScoringConfig() error = %v", err)
	}

	if config.GetImpact("microservice") != 5 {
		t.Errorf("GetImpact(microservice) = %d, want 5", config.GetImpact("microservice"))
	}
	if config.GetImpact("api") != 5 {
		t.Errorf("GetImpact(api) = %d, want 5", config.GetImpact("api"))
	}
	if config.GetCriticality("microservice") != "high" {
		t.Errorf("GetCriticality(microservice) = %q, want high", config.GetCriticality("microservice"))
	}
	if config.GetCriticality("cli") != "medium" {
		t.Errorf("GetCriticality(cli) = %q, want medium", config.GetCriticality("cli"))
	}

	// Non-overridden values preserved
	if config.GetImpact("service") != 4 {
		t.Errorf("GetImpact(service) = %d, want 4", config.GetImpact("service"))
	}
	if config.GetCriticality("gateway") != "high" {
		t.Errorf("GetCriticality(gateway) = %q, want high", config.GetCriticality("gateway"))
	}
}

func TestGetRiskScoringConfig_DefaultWhenNotInitialized(t *testing.T) {
	riskConfig = nil

	config := GetRiskScoringConfig()
	if config == nil {
		t.Fatal("GetRiskScoringConfig() returned nil")
	}

	if config.GetImpact("api") != 4 {
		t.Errorf("GetImpact(api) = %d, want 4", config.GetImpact("api"))
	}
}

func TestInitRiskScoringConfig(t *testing.T) {
	riskConfig = nil
	tmpDir := t.TempDir()

	err := InitRiskScoringConfig(tmpDir)
	if err != nil {
		t.Fatalf("InitRiskScoringConfig() error = %v", err)
	}

	config := GetRiskScoringConfig()
	if config == nil {
		t.Fatal("GetRiskScoringConfig() returned nil after init")
	}

	riskConfig = nil
}

func TestBackwardCompatibility_GetDefaultImpact(t *testing.T) {
	riskConfig = nil

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
			got := GetDefaultImpact(tt.moduleType)
			if got != tt.want {
				t.Errorf("GetDefaultImpact(%q) = %d, want %d", tt.moduleType, got, tt.want)
			}
		})
	}
}
