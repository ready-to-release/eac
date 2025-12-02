package oscal

import (
	"testing"
	"time"
)

func TestNewProfile(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		title      string
		catalogURL string
		controlIDs []string
		wantErr    bool
	}{
		{
			name:       "valid profile",
			uuid:       "test-uuid-123",
			title:      "Test Profile",
			catalogURL: "https://example.com/catalog.json",
			controlIDs: []string{"ac-2", "ia-2"},
			wantErr:    false,
		},
		{
			name:       "empty controls",
			uuid:       "test-uuid-456",
			title:      "Empty Profile",
			catalogURL: "https://example.com/catalog.json",
			controlIDs: []string{},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := NewProfile(tt.uuid, tt.title, tt.catalogURL, tt.controlIDs)

			if profile == nil {
				t.Fatal("NewProfile returned nil")
			}

			if profile.Profile.UUID != tt.uuid {
				t.Errorf("UUID = %s, want %s", profile.Profile.UUID, tt.uuid)
			}

			if profile.Profile.Metadata.Title != tt.title {
				t.Errorf("Title = %s, want %s", profile.Profile.Metadata.Title, tt.title)
			}

			if profile.Profile.Metadata.OSCALVersion != OSCALVersion {
				t.Errorf("OSCALVersion = %s, want %s", profile.Profile.Metadata.OSCALVersion, OSCALVersion)
			}

			if len(profile.Profile.Imports) != 1 {
				t.Errorf("Imports length = %d, want 1", len(profile.Profile.Imports))
			}

			if profile.Profile.Imports[0].Href != tt.catalogURL {
				t.Errorf("Catalog URL = %s, want %s", profile.Profile.Imports[0].Href, tt.catalogURL)
			}

			if len(profile.Profile.Imports[0].IncludeControls) != len(tt.controlIDs) {
				t.Errorf("IncludeControls length = %d, want %d",
					len(profile.Profile.Imports[0].IncludeControls), len(tt.controlIDs))
			}

			// Verify last-modified is recent
			lastMod, err := time.Parse(time.RFC3339, profile.Profile.Metadata.LastModified)
			if err != nil {
				t.Errorf("Failed to parse LastModified: %v", err)
			}
			if time.Since(lastMod) > time.Minute {
				t.Errorf("LastModified is not recent: %s", profile.Profile.Metadata.LastModified)
			}
		})
	}
}

func TestNewAssessmentResults(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		title      string
		profileRef string
		wantErr    bool
	}{
		{
			name:       "valid assessment-results",
			uuid:       "ar-uuid-123",
			title:      "Assessment for Billing",
			profileRef: "specs/risk-controls/billing.profile.json",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := NewAssessmentResults(tt.uuid, tt.title, tt.profileRef)

			if ar == nil {
				t.Fatal("NewAssessmentResults returned nil")
			}

			if ar.AssessmentResults.UUID != tt.uuid {
				t.Errorf("UUID = %s, want %s", ar.AssessmentResults.UUID, tt.uuid)
			}

			if ar.AssessmentResults.Metadata.Title != tt.title {
				t.Errorf("Title = %s, want %s", ar.AssessmentResults.Metadata.Title, tt.title)
			}

			if ar.AssessmentResults.Metadata.OSCALVersion != OSCALVersion {
				t.Errorf("OSCALVersion = %s, want %s", ar.AssessmentResults.Metadata.OSCALVersion, OSCALVersion)
			}

			if ar.AssessmentResults.ImportAP.Href != tt.profileRef {
				t.Errorf("ProfileRef = %s, want %s", ar.AssessmentResults.ImportAP.Href, tt.profileRef)
			}

			// Results array is empty by default - results are added via AddResult
			if ar.AssessmentResults.Results == nil {
				t.Error("Results should not be nil")
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if OSCALVersion != "1.1.2" {
		t.Errorf("OSCALVersion = %s, want 1.1.2", OSCALVersion)
	}

	if StateSatisfied != "satisfied" {
		t.Errorf("StateSatisfied = %s, want satisfied", StateSatisfied)
	}

	if StateNotSatisfied != "not-satisfied" {
		t.Errorf("StateNotSatisfied = %s, want not-satisfied", StateNotSatisfied)
	}

	if NIST80053Rev5CatalogURL == "" {
		t.Error("NIST80053Rev5CatalogURL should not be empty")
	}
}
