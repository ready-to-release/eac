package oscal

import (
	"testing"
	"time"
)

func TestNewProfileDocument(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		catalogURL string
		controlIDs []string
		wantErr    bool
	}{
		{
			name:       "valid profile",
			title:      "Test Profile",
			catalogURL: "https://example.com/catalog.json",
			controlIDs: []string{"ac-2", "ia-2"},
			wantErr:    false,
		},
		{
			name:       "empty controls",
			title:      "Empty Profile",
			catalogURL: "https://example.com/catalog.json",
			controlIDs: []string{},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oscalDoc, err := NewProfileDocument(tt.title, tt.catalogURL, tt.controlIDs)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewProfileDocument returned error: %v", err)
			}

			if oscalDoc == nil || oscalDoc.Profile == nil {
				t.Fatal("NewProfileDocument returned nil profile")
			}

			profile := oscalDoc.Profile

			if profile.UUID == "" {
				t.Error("UUID is empty")
			}

			if profile.Metadata.Title != tt.title {
				t.Errorf("Title = %s, want %s", profile.Metadata.Title, tt.title)
			}

			if profile.Metadata.OscalVersion == "" {
				t.Error("OscalVersion is empty")
			}

			if len(profile.Imports) != 1 {
				t.Errorf("Imports length = %d, want 1", len(profile.Imports))
			}

			if profile.Imports[0].Href != tt.catalogURL {
				t.Errorf("Catalog URL = %s, want %s", profile.Imports[0].Href, tt.catalogURL)
			}

			// Verify last-modified is recent
			if profile.Metadata.LastModified.IsZero() {
				t.Error("LastModified is zero time")
			}
			if time.Since(profile.Metadata.LastModified) > time.Minute {
				t.Errorf("LastModified is not recent: %s", profile.Metadata.LastModified)
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

			if ar.UUID != tt.uuid {
				t.Errorf("UUID = %s, want %s", ar.UUID, tt.uuid)
			}

			if ar.Metadata.Title != tt.title {
				t.Errorf("Title = %s, want %s", ar.Metadata.Title, tt.title)
			}

			if ar.Metadata.OscalVersion == "" {
				t.Error("OscalVersion is empty")
			}

			if ar.ImportAp.Href != tt.profileRef {
				t.Errorf("ProfileRef = %s, want %s", ar.ImportAp.Href, tt.profileRef)
			}

			// Results array is empty by default - results are added via AddResult
			if ar.Results == nil {
				t.Error("Results should not be nil")
			}
		})
	}
}

func TestConstants(t *testing.T) {
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
