package oscal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProfile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a valid profile file
	validProfile := `{
		"profile": {
			"uuid": "test-uuid",
			"metadata": {
				"title": "Test Profile",
				"last-modified": "2024-01-01T00:00:00Z",
				"oscal-version": "1.1.2"
			},
			"imports": [{
				"href": "https://example.com/catalog.json",
				"include-controls": [{"with-id": "ac-2"}]
			}]
		}
	}`
	validPath := filepath.Join(tmpDir, "valid.profile.json")
	if err := os.WriteFile(validPath, []byte(validProfile), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create an invalid JSON file
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("not json"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid profile",
			path:    validPath,
			wantErr: false,
		},
		{
			name:    "file not found",
			path:    filepath.Join(tmpDir, "nonexistent.json"),
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			path:    invalidPath,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := LoadProfile(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if profile == nil {
				t.Error("Profile is nil")
				return
			}

			if profile.Profile.UUID != "test-uuid" {
				t.Errorf("UUID = %s, want test-uuid", profile.Profile.UUID)
			}
		})
	}
}

func TestLoadAssessmentResults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid assessment-results file
	validAR := `{
		"assessment-results": {
			"uuid": "ar-uuid",
			"metadata": {
				"title": "Test Assessment",
				"last-modified": "2024-01-01T00:00:00Z",
				"oscal-version": "1.1.2"
			},
			"import-ap": {"href": "profile.json"},
			"results": [{
				"uuid": "result-uuid",
				"title": "Test Result",
				"start": "2024-01-01T00:00:00Z",
				"findings": []
			}]
		}
	}`
	validPath := filepath.Join(tmpDir, "assessment-results.json")
	if err := os.WriteFile(validPath, []byte(validAR), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid assessment-results",
			path:    validPath,
			wantErr: false,
		},
		{
			name:    "file not found",
			path:    filepath.Join(tmpDir, "nonexistent.json"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar, err := LoadAssessmentResults(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if ar == nil {
				t.Error("AssessmentResults is nil")
				return
			}

			if ar.AssessmentResults.UUID != "ar-uuid" {
				t.Errorf("UUID = %s, want ar-uuid", ar.AssessmentResults.UUID)
			}
		})
	}
}

func TestDetectOSCALDocumentType(t *testing.T) {
	tmpDir := t.TempDir()

	// Create profile file
	profileContent := `{"profile": {"uuid": "test"}}`
	profilePath := filepath.Join(tmpDir, "profile.json")
	os.WriteFile(profilePath, []byte(profileContent), 0644)

	// Create assessment-results file
	arContent := `{"assessment-results": {"uuid": "test"}}`
	arPath := filepath.Join(tmpDir, "assessment.json")
	os.WriteFile(arPath, []byte(arContent), 0644)

	// Create unknown file
	unknownContent := `{"other": {"data": "test"}}`
	unknownPath := filepath.Join(tmpDir, "unknown.json")
	os.WriteFile(unknownPath, []byte(unknownContent), 0644)

	tests := []struct {
		name     string
		path     string
		wantType string
		wantErr  bool
	}{
		{
			name:     "detect profile",
			path:     profilePath,
			wantType: "profile",
			wantErr:  false,
		},
		{
			name:     "detect assessment-results",
			path:     arPath,
			wantType: "assessment-results",
			wantErr:  false,
		},
		{
			name:     "unknown type",
			path:     unknownPath,
			wantType: "",
			wantErr:  false,
		},
		{
			name:     "file not found",
			path:     filepath.Join(tmpDir, "nonexistent.json"),
			wantType: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docType, err := DetectOSCALDocumentType(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if docType != tt.wantType {
				t.Errorf("DocumentType = %s, want %s", docType, tt.wantType)
			}
		})
	}
}

func TestDiscoverProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create specs/risk-controls directory
	riskDir := filepath.Join(tmpDir, "specs", "risk-controls")
	os.MkdirAll(riskDir, 0755)

	// Create profile files
	profile1 := `{"profile": {"uuid": "p1", "metadata": {"title": "Profile 1"}}}`
	profile2 := `{"profile": {"uuid": "p2", "metadata": {"title": "Profile 2"}}}`
	os.WriteFile(filepath.Join(riskDir, "billing.profile.json"), []byte(profile1), 0644)
	os.WriteFile(filepath.Join(riskDir, "auth.profile.json"), []byte(profile2), 0644)

	profiles, err := DiscoverProfiles(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverProfiles failed: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("Found %d profiles, want 2", len(profiles))
	}

	if _, ok := profiles["billing"]; !ok {
		t.Error("billing profile not found")
	}

	if _, ok := profiles["auth"]; !ok {
		t.Error("auth profile not found")
	}
}

func TestDiscoverAssessmentResults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create out/risk directories
	billingDir := filepath.Join(tmpDir, "out", "risk", "billing")
	authDir := filepath.Join(tmpDir, "out", "risk", "auth")
	os.MkdirAll(billingDir, 0755)
	os.MkdirAll(authDir, 0755)

	// Create assessment-results files
	ar1 := `{"assessment-results": {"uuid": "ar1"}}`
	ar2 := `{"assessment-results": {"uuid": "ar2"}}`
	os.WriteFile(filepath.Join(billingDir, "assessment-results.json"), []byte(ar1), 0644)
	os.WriteFile(filepath.Join(authDir, "assessment-results.json"), []byte(ar2), 0644)

	results, err := DiscoverAssessmentResults(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverAssessmentResults failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Found %d results, want 2", len(results))
	}

	if _, ok := results["billing"]; !ok {
		t.Error("billing results not found")
	}

	if _, ok := results["auth"]; !ok {
		t.Error("auth results not found")
	}
}
