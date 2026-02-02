package domain

import (
	"testing"
)

// TestGoMatchesSchema is the critical test that ensures Go constants
// stay in sync with shared-definitions.schema.json.
// If this test fails, update the Go constants in shared_definitions.go.
func TestGoMatchesSchema(t *testing.T) {
	if err := ValidateGoMatchesSchema(); err != nil {
		t.Fatalf("Go/Schema mismatch:\n%v", err)
	}
}

func TestValidScannerCategories(t *testing.T) {
	categories := ValidScannerCategories()
	if len(categories) == 0 {
		t.Fatal("ValidScannerCategories() returned empty map")
	}

	// Test known valid categories
	validCats := []string{"sbom", "vuln", "secrets", "iac", "compliance", "sast", "zap"}
	for _, cat := range validCats {
		if !categories[cat] {
			t.Errorf("ValidScannerCategories() missing %q", cat)
		}
	}
}

func TestIsValidScannerCategory(t *testing.T) {
	tests := []struct {
		category string
		want     bool
	}{
		{"sbom", true},
		{"vuln", true},
		{"secrets", true},
		{"iac", true},
		{"compliance", true},
		{"sast", true},
		{"zap", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			if got := IsValidScannerCategory(tt.category); got != tt.want {
				t.Errorf("IsValidScannerCategory(%q) = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}

func TestValidServerTypes(t *testing.T) {
	serverTypes := ValidServerTypes()
	if len(serverTypes) == 0 {
		t.Fatal("ValidServerTypes() returned empty map")
	}

	validTypes := []string{"static-site", "mkdocs-live", "structurizr-lite"}
	for _, st := range validTypes {
		if !serverTypes[st] {
			t.Errorf("ValidServerTypes() missing %q", st)
		}
	}
}

func TestIsValidServerType(t *testing.T) {
	tests := []struct {
		serverType string
		want       bool
	}{
		{"static-site", true},
		{"mkdocs-live", true},
		{"structurizr-lite", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.serverType, func(t *testing.T) {
			if got := IsValidServerType(tt.serverType); got != tt.want {
				t.Errorf("IsValidServerType(%q) = %v, want %v", tt.serverType, got, tt.want)
			}
		})
	}
}

func TestValidPlatforms(t *testing.T) {
	platforms := ValidPlatforms()
	if len(platforms) == 0 {
		t.Fatal("ValidPlatforms() returned empty map")
	}

	validPlatforms := []string{"linux", "darwin", "windows"}
	for _, p := range validPlatforms {
		if !platforms[p] {
			t.Errorf("ValidPlatforms() missing %q", p)
		}
	}
}

func TestIsValidPlatform(t *testing.T) {
	tests := []struct {
		platform string
		want     bool
	}{
		{"linux", true},
		{"darwin", true},
		{"windows", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			if got := IsValidPlatform(tt.platform); got != tt.want {
				t.Errorf("IsValidPlatform(%q) = %v, want %v", tt.platform, got, tt.want)
			}
		})
	}
}

func TestValidSeverities(t *testing.T) {
	severities := ValidSeverities()
	if len(severities) == 0 {
		t.Fatal("ValidSeverities() returned empty map")
	}

	validSeverities := []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}
	for _, s := range validSeverities {
		if !severities[s] {
			t.Errorf("ValidSeverities() missing %q", s)
		}
	}
}

func TestIsValidSeverity(t *testing.T) {
	tests := []struct {
		severity string
		want     bool
	}{
		{"LOW", true},
		{"MEDIUM", true},
		{"HIGH", true},
		{"CRITICAL", true},
		{"low", false}, // Case sensitive
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			if got := IsValidSeverity(tt.severity); got != tt.want {
				t.Errorf("IsValidSeverity(%q) = %v, want %v", tt.severity, got, tt.want)
			}
		})
	}
}
