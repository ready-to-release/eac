package config

import (
	"os"
	"path/filepath"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	scanner "github.com/ready-to-release/eac/contracts/scanner/0.1.0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileWrapper_ImplementsPort(t *testing.T) {
	var _ scanner.ProfilePort = (*ProfileWrapper)(nil)
}

func TestNewProfileWrapper(t *testing.T) {
	profile := &oscalTypes.Profile{
		Metadata: oscalTypes.Metadata{
			Title:   "Test Profile",
			Version: "1.0.0",
		},
	}

	wrapper := NewProfileWrapper(profile)
	assert.NotNil(t, wrapper)
	assert.Equal(t, "Test Profile", wrapper.Title())
	assert.Equal(t, "1.0.0", wrapper.Version())
}

func TestProfileWrapper_NilProfile(t *testing.T) {
	wrapper := &ProfileWrapper{}

	assert.Nil(t, wrapper.ControlIDs())
	assert.False(t, wrapper.HasControl("ac-1"))
	assert.Empty(t, wrapper.Title())
	assert.Empty(t, wrapper.Version())
	assert.Empty(t, wrapper.CatalogHref())
}

func TestProfileWrapper_ControlIDs(t *testing.T) {
	controlIDs := []string{"ac-1", "ac-2", "ia-1"}
	profile := createTestProfile("Test", "1.0", "https://example.com/catalog.json", controlIDs)

	wrapper := NewProfileWrapper(profile)
	ids := wrapper.ControlIDs()

	assert.Len(t, ids, 3)
	assert.Contains(t, ids, "ac-1")
	assert.Contains(t, ids, "ac-2")
	assert.Contains(t, ids, "ia-1")
}

func TestProfileWrapper_ControlIDs_NoDuplicates(t *testing.T) {
	// Create a profile with duplicate control IDs
	controlIDs := []string{"ac-1", "ac-2", "ac-1"} // ac-1 appears twice
	profile := createTestProfile("Test", "1.0", "https://example.com/catalog.json", controlIDs)

	wrapper := NewProfileWrapper(profile)
	ids := wrapper.ControlIDs()

	// Should deduplicate
	assert.Len(t, ids, 2)
}

func TestProfileWrapper_HasControl(t *testing.T) {
	controlIDs := []string{"ac-1", "ac-2", "ia-1"}
	profile := createTestProfile("Test", "1.0", "https://example.com/catalog.json", controlIDs)

	wrapper := NewProfileWrapper(profile)

	assert.True(t, wrapper.HasControl("ac-1"))
	assert.True(t, wrapper.HasControl("AC-1")) // Case-insensitive
	assert.True(t, wrapper.HasControl("ia-1"))
	assert.False(t, wrapper.HasControl("ac-3"))
	assert.False(t, wrapper.HasControl(""))
}

func TestProfileWrapper_CatalogHref(t *testing.T) {
	catalogURL := "https://raw.githubusercontent.com/usnistgov/oscal-content/main/nist.gov/SP800-53/rev5/json/NIST_SP-800-53_rev5_catalog.json"
	profile := createTestProfile("Test", "1.0", catalogURL, []string{"ac-1"})

	wrapper := NewProfileWrapper(profile)
	assert.Equal(t, catalogURL, wrapper.CatalogHref())
}

func TestLoadProfileWrapper(t *testing.T) {
	// Create a temporary profile file
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "test-profile.json")

	profileJSON := `{
  "profile": {
    "uuid": "test-uuid",
    "metadata": {
      "title": "Test Profile",
      "version": "1.0.0",
      "oscal-version": "1.1.3",
      "last-modified": "2024-01-01T00:00:00Z"
    },
    "imports": [
      {
        "href": "https://example.com/catalog.json",
        "include-controls": [
          {
            "with-ids": ["ac-1", "ac-2", "ia-1"]
          }
        ]
      }
    ]
  }
}`

	err := os.WriteFile(profilePath, []byte(profileJSON), 0o644)
	require.NoError(t, err)

	wrapper, err := LoadProfileWrapper(profilePath)
	require.NoError(t, err)

	assert.Equal(t, "Test Profile", wrapper.Title())
	assert.Equal(t, "1.0.0", wrapper.Version())
	assert.Equal(t, "https://example.com/catalog.json", wrapper.CatalogHref())
	assert.Len(t, wrapper.ControlIDs(), 3)
	assert.True(t, wrapper.HasControl("ac-1"))
}

func TestLoadProfileWrapper_FileNotFound(t *testing.T) {
	_, err := LoadProfileWrapper("/nonexistent/path/profile.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading profile file")
}

func TestLoadProfileWrapper_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(profilePath, []byte("not valid json"), 0o644)
	require.NoError(t, err)

	_, err = LoadProfileWrapper(profilePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing profile JSON")
}

func TestLoadProfileWrapper_NotAProfile(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "catalog.json")

	// Valid JSON but not a profile
	catalogJSON := `{"catalog": {"uuid": "test"}}`
	err := os.WriteFile(profilePath, []byte(catalogJSON), 0o644)
	require.NoError(t, err)

	_, err = LoadProfileWrapper(profilePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain a profile")
}

// Helper function to create a test profile
func createTestProfile(title, version, catalogHref string, controlIDs []string) *oscalTypes.Profile {
	withIDs := controlIDs
	includeControls := []oscalTypes.SelectControlById{
		{WithIds: &withIDs},
	}

	return &oscalTypes.Profile{
		UUID: "test-uuid",
		Metadata: oscalTypes.Metadata{
			Title:   title,
			Version: version,
		},
		Imports: []oscalTypes.Import{
			{
				Href:            catalogHref,
				IncludeControls: &includeControls,
			},
		},
	}
}
