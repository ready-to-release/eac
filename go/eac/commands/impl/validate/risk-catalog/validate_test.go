package riskcatalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidControlID(t *testing.T) {
	// Note: This function doesn't exist yet in catalog validation,
	// but this test shows the pattern for when it's added
	t.Skip("isValidControlID not yet implemented in risk-catalog")
}

func TestValidateCatalog_ValidMinimalCatalog(t *testing.T) {
	// Create temp file with valid minimal catalog
	tmpDir := t.TempDir()
	catalogPath := filepath.Join(tmpDir, "test-catalog.json")

	validCatalog := `{
		"catalog": {
			"uuid": "12345678-1234-4234-8234-123456789abc",
			"metadata": {
				"title": "Test Catalog",
				"last-modified": "2025-01-01T00:00:00Z",
				"version": "1.0.0",
				"oscal-version": "1.1.3"
			},
			"groups": [{
				"id": "ac",
				"title": "Access Control",
				"controls": [{
					"id": "ac-1",
					"title": "Policy and Procedures"
				}]
			}]
		}
	}`

	err := os.WriteFile(catalogPath, []byte(validCatalog), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      catalogPath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateCatalog(config)

	if len(errors) > 0 {
		t.Errorf("Expected valid catalog, but got invalid. Errors: %+v", errors)
	}
}

func TestValidateCatalog_MissingUUID(t *testing.T) {
	tmpDir := t.TempDir()
	catalogPath := filepath.Join(tmpDir, "catalog-no-uuid.json")

	catalogWithoutUUID := `{
		"catalog": {
			"metadata": {
				"title": "Test Catalog",
				"last-modified": "2025-01-01T00:00:00Z",
				"version": "1.0.0"
			},
			"groups": [{
				"id": "ac",
				"title": "Access Control",
				"controls": [{
					"id": "ac-1",
					"title": "Policy"
				}]
			}]
		}
	}`

	err := os.WriteFile(catalogPath, []byte(catalogWithoutUUID), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      catalogPath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateCatalog(config)

	if len(errors) == 0 {
		t.Error("Expected invalid catalog (missing UUID), but got valid")
	}

	foundUUIDError := false
	for _, err := range errors {
		if err.GetCode() == "OSCAL_MISSING_UUID" {
			foundUUIDError = true
			break
		}
	}

	if !foundUUIDError {
		t.Errorf("Expected UUID error, but didn't find it. Errors: %+v", errors)
	}
}

func TestValidateCatalog_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	catalogPath := filepath.Join(tmpDir, "invalid.json")

	invalidJSON := `{ invalid json`

	err := os.WriteFile(catalogPath, []byte(invalidJSON), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      catalogPath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateCatalog(config)

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
		t.Errorf("Expected JSON parsing error. Errors: %+v", errors)
	}
}

func TestValidateCatalog_NotACatalogDocument(t *testing.T) {
	tmpDir := t.TempDir()
	catalogPath := filepath.Join(tmpDir, "not-catalog.json")

	notCatalog := `{
		"profile": {
			"uuid": "12345678-1234-4234-8234-123456789abc",
			"metadata": {
				"title": "Test Profile"
			}
		}
	}`

	err := os.WriteFile(catalogPath, []byte(notCatalog), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      catalogPath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateCatalog(config)

	if len(errors) == 0 {
		t.Error("Expected invalid result for non-catalog document")
	}

	foundDocError := false
	for _, err := range errors {
		if err.GetCode() == "OSCAL_INVALID_DOCUMENT" {
			foundDocError = true
			break
		}
	}

	if !foundDocError {
		t.Errorf("Expected document type error. Errors: %+v", errors)
	}
}

func TestValidateCatalog_MissingTitleAndLastModified(t *testing.T) {
	tmpDir := t.TempDir()
	catalogPath := filepath.Join(tmpDir, "incomplete-catalog.json")

	incompleteCatalog := `{
		"catalog": {
			"uuid": "12345678-1234-4234-8234-123456789abc",
			"metadata": {
				"version": "1.0.0"
			},
			"groups": [{
				"id": "ac",
				"controls": [{
					"id": "ac-1",
					"title": "Policy"
				}]
			}]
		}
	}`

	err := os.WriteFile(catalogPath, []byte(incompleteCatalog), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      catalogPath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateCatalog(config)

	if len(errors) == 0 {
		t.Error("Expected invalid catalog (missing required metadata fields)")
	}

	// Should have errors for both missing title and last-modified
	if len(errors) < 2 {
		t.Errorf("Expected at least 2 errors (title and last-modified), got %d: %+v",
			len(errors), errors)
	}
}

func TestValidateCatalog_NoControlsOrGroups(t *testing.T) {
	tmpDir := t.TempDir()
	catalogPath := filepath.Join(tmpDir, "empty-catalog.json")

	emptyCatalog := `{
		"catalog": {
			"uuid": "12345678-1234-4234-8234-123456789abc",
			"metadata": {
				"title": "Empty Catalog",
				"last-modified": "2025-01-01T00:00:00Z",
				"version": "1.0.0"
			}
		}
	}`

	err := os.WriteFile(catalogPath, []byte(emptyCatalog), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &Config{
		FilePath:      catalogPath,
		WorkspaceRoot: tmpDir,
	}

	errors := validateCatalog(config)

	if len(errors) == 0 {
		t.Error("Expected invalid catalog (no controls or groups)")
	}

	foundControlError := false
	for _, err := range errors {
		if err.GetCode() == "OSCAL_EMPTY_CATALOG" {
			foundControlError = true
			break
		}
	}

	if !foundControlError {
		t.Errorf("Expected error about missing controls/groups. Errors: %+v", errors)
	}
}
