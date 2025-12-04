package riskprofile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidControlID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{"valid NIST format", "ac-2", true},
		{"valid with uppercase", "AC-2", true},
		{"valid multi-digit", "si-10", true},
		{"invalid no dash", "ac2", false},
		{"invalid single char family", "a-2", false},
		{"invalid three char family", "abc-2", false},
		{"invalid non-numeric", "ac-abc", false},
		{"empty string", "", false},
		{"just dash", "-", false},
		{"missing number", "ac-", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidControlID(tt.id)
			if result != tt.expected {
				t.Errorf("isValidControlID(%q) = %v, want %v", tt.id, result, tt.expected)
			}
		})
	}
}

func TestValidateProfile_ValidProfile(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "valid-profile.json")

	validProfile := `{
		"profile": {
			"uuid": "12345678-1234-4234-8234-123456789abc",
			"metadata": {
				"title": "Test Profile",
				"last-modified": "2025-01-01T00:00:00Z",
				"version": "1.0.0",
				"oscal-version": "1.1.2"
			},
			"imports": [{
				"href": "https://example.com/catalog.json",
				"include-controls": [{
					"with-ids": ["ac-2", "ia-5"]
				}]
			}]
		}
	}`

	err := os.WriteFile(profilePath, []byte(validProfile), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	result := validateProfile(config)

	if !result.Valid {
		t.Errorf("Expected valid profile, got invalid. Errors: %+v", result.Errors)
	}

	if len(result.Errors) > 0 {
		t.Errorf("Expected no errors, got %d: %+v", len(result.Errors), result.Errors)
	}
}

func TestValidateProfile_MissingUUID(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "no-uuid-profile.json")

	profileWithoutUUID := `{
		"profile": {
			"metadata": {
				"title": "Test Profile",
				"last-modified": "2025-01-01T00:00:00Z",
				"version": "1.0.0"
			},
			"imports": [{
				"href": "https://example.com/catalog.json",
				"include-controls": [{
					"with-ids": ["ac-2"]
				}]
			}]
		}
	}`

	err := os.WriteFile(profilePath, []byte(profileWithoutUUID), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	result := validateProfile(config)

	if result.Valid {
		t.Error("Expected invalid profile (missing UUID)")
	}

	foundUUIDError := false
	for _, err := range result.Errors {
		if err.Field == "profile.uuid" {
			foundUUIDError = true
			break
		}
	}

	if !foundUUIDError {
		t.Errorf("Expected UUID error. Errors: %+v", result.Errors)
	}
}

func TestValidateProfile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "invalid.json")

	invalidJSON := `{ "profile": { invalid`

	err := os.WriteFile(profilePath, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	result := validateProfile(config)

	if result.Valid {
		t.Error("Expected invalid result for malformed JSON")
	}

	foundJSONError := false
	for _, err := range result.Errors {
		if err.Field == "json" {
			foundJSONError = true
			break
		}
	}

	if !foundJSONError {
		t.Errorf("Expected JSON error. Errors: %+v", result.Errors)
	}
}

func TestValidateProfile_NoImports(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "no-imports-profile.json")

	profileWithoutImports := `{
		"profile": {
			"uuid": "12345678-1234-4234-8234-123456789abc",
			"metadata": {
				"title": "Test Profile",
				"last-modified": "2025-01-01T00:00:00Z",
				"version": "1.0.0"
			},
			"imports": []
		}
	}`

	err := os.WriteFile(profilePath, []byte(profileWithoutImports), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	result := validateProfile(config)

	if result.Valid {
		t.Error("Expected invalid profile (no imports)")
	}

	foundImportsError := false
	for _, err := range result.Errors {
		if err.Field == "profile.imports" {
			foundImportsError = true
			break
		}
	}

	if !foundImportsError {
		t.Errorf("Expected imports error. Errors: %+v", result.Errors)
	}
}

func TestValidateProfile_InvalidControlIDWarning(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "invalid-control-profile.json")

	profileWithInvalidControl := `{
		"profile": {
			"uuid": "12345678-1234-4234-8234-123456789abc",
			"metadata": {
				"title": "Test Profile",
				"last-modified": "2025-01-01T00:00:00Z",
				"version": "1.0.0",
				"oscal-version": "1.1.2"
			},
			"imports": [{
				"href": "https://example.com/catalog.json",
				"include-controls": [{
					"with-ids": ["ac-2", "invalid-control-id"]
				}]
			}]
		}
	}`

	err := os.WriteFile(profilePath, []byte(profileWithInvalidControl), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	result := validateProfile(config)

	// Should be valid but with warnings
	if !result.Valid {
		t.Errorf("Expected valid profile (invalid control should be warning). Errors: %+v", result.Errors)
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for invalid control ID format")
	}

	foundControlWarning := false
	for _, warn := range result.Warnings {
		if warn.Message == "control ID 'invalid-control-id' may not be valid NIST 800-53 format" {
			foundControlWarning = true
			break
		}
	}

	if !foundControlWarning {
		t.Errorf("Expected warning about invalid control ID. Warnings: %+v", result.Warnings)
	}
}

func TestValidateProfile_OscalVersionMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "version-mismatch-profile.json")

	profileWithWrongVersion := `{
		"profile": {
			"uuid": "12345678-1234-4234-8234-123456789abc",
			"metadata": {
				"title": "Test Profile",
				"last-modified": "2025-01-01T00:00:00Z",
				"version": "1.0.0",
				"oscal-version": "1.0.0"
			},
			"imports": [{
				"href": "https://example.com/catalog.json",
				"include-controls": [{
					"with-ids": ["ac-2"]
				}]
			}]
		}
	}`

	err := os.WriteFile(profilePath, []byte(profileWithWrongVersion), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	result := validateProfile(config)

	// Should be valid but with version warning
	if !result.Valid {
		t.Errorf("Expected valid (version mismatch should be warning). Errors: %+v", result.Errors)
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warning for OSCAL version mismatch")
	}
}
