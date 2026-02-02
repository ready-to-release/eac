package riskprofile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/validation/formats/oscal"
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
			result := oscal.IsValidControlID(tt.id)
			if result != tt.expected {
				t.Errorf("IsValidControlID(%q) = %v, want %v", tt.id, result, tt.expected)
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

	err := os.WriteFile(profilePath, []byte(validProfile), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateProfile(config)

	if len(errors) > 0 {
		t.Errorf("Expected valid profile, got invalid. Errors: %+v", errors)
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

	err := os.WriteFile(profilePath, []byte(profileWithoutUUID), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateProfile(config)

	if len(errors) == 0 {
		t.Error("Expected invalid profile (missing UUID)")
	}

	foundUUIDError := false
	for _, err := range errors {
		if err.GetCode() == "OSCAL_MISSING_UUID" {
			foundUUIDError = true
			break
		}
	}

	if !foundUUIDError {
		t.Errorf("Expected UUID error. Errors: %+v", errors)
	}
}

func TestValidateProfile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "invalid.json")

	invalidJSON := `{ "profile": { invalid`

	err := os.WriteFile(profilePath, []byte(invalidJSON), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateProfile(config)

	if len(errors) == 0 {
		t.Error("Expected invalid result for malformed JSON")
	}

	foundJSONError := false
	for _, err := range errors {
		if err.GetCode() == "INVALID_JSON" {
			foundJSONError = true
			break
		}
	}

	if !foundJSONError {
		t.Errorf("Expected JSON error. Errors: %+v", errors)
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

	err := os.WriteFile(profilePath, []byte(profileWithoutImports), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateProfile(config)

	if len(errors) == 0 {
		t.Error("Expected invalid profile (no imports)")
	}

	foundImportsError := false
	for _, err := range errors {
		if err.GetCode() == "OSCAL_MISSING_IMPORTS" {
			foundImportsError = true
			break
		}
	}

	if !foundImportsError {
		t.Errorf("Expected imports error. Errors: %+v", errors)
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

	err := os.WriteFile(profilePath, []byte(profileWithInvalidControl), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateProfile(config)

	// Should have warnings but no errors
	var warnings []string
	hasErrors := false
	for _, err := range errors {
		if err.IsWarning() {
			warnings = append(warnings, err.GetCode())
		} else {
			hasErrors = true
		}
	}

	if hasErrors {
		t.Errorf("Expected no errors, only warnings. Errors: %+v", errors)
	}

	if len(warnings) == 0 {
		t.Error("Expected warnings for invalid control ID format")
	}

	foundControlWarning := false
	for _, err := range errors {
		if err.GetCode() == "OSCAL_INVALID_CONTROL_ID" && err.IsWarning() {
			foundControlWarning = true
			break
		}
	}

	if !foundControlWarning {
		t.Errorf("Expected warning about invalid control ID. Errors: %+v", errors)
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

	err := os.WriteFile(profilePath, []byte(profileWithWrongVersion), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      profilePath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateProfile(config)

	// Should have warnings but no errors
	hasErrors := false
	hasWarnings := false
	for _, err := range errors {
		if err.IsWarning() {
			hasWarnings = true
		} else {
			hasErrors = true
		}
	}

	if hasErrors {
		t.Errorf("Expected no errors, only warnings (version mismatch should be warning). Errors: %+v", errors)
	}

	if !hasWarnings {
		t.Error("Expected warning for OSCAL version mismatch")
	}
}
