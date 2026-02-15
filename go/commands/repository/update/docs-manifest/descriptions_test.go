package docsmanifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDescriptions_FileNotFound(t *testing.T) {
	descriptions, err := LoadDescriptions(filepath.Join(t.TempDir(), "nonexistent.yml"))

	require.NoError(t, err, "missing file should not be an error")
	assert.NotNil(t, descriptions)
	assert.Empty(t, descriptions)
}

func TestLoadDescriptions_EmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "descriptions.yml")
	require.NoError(t, os.WriteFile(path, []byte(""), 0o644))

	descriptions, err := LoadDescriptions(path)

	require.NoError(t, err)
	assert.NotNil(t, descriptions)
	assert.Empty(t, descriptions)
}

func TestLoadDescriptions_NilAssets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "descriptions.yml")
	content := "assets:\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	descriptions, err := LoadDescriptions(path)

	require.NoError(t, err)
	assert.NotNil(t, descriptions)
	assert.Empty(t, descriptions)
}

func TestLoadDescriptions_ValidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "descriptions.yml")
	content := `assets:
  architecture/overview.drawio.png:
    active: true
    description: "System overview diagram"
  icons/logo.png:
    active: false
    description: "Deprecated logo"
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	descriptions, err := LoadDescriptions(path)

	require.NoError(t, err)
	require.Len(t, descriptions, 2)

	overview := descriptions["architecture/overview.drawio.png"]
	assert.True(t, overview.Active)
	assert.Equal(t, "System overview diagram", overview.Description)

	logo := descriptions["icons/logo.png"]
	assert.False(t, logo.Active)
	assert.Equal(t, "Deprecated logo", logo.Description)
}

func TestLoadDescriptions_InvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "descriptions.yml")
	content := "this is not: valid: yaml: [["
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	descriptions, err := LoadDescriptions(path)

	assert.Error(t, err)
	assert.Nil(t, descriptions)
}

func TestSaveDescriptions_CreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "descriptions.yml")
	descriptions := map[string]Description{
		"test/image.png": {
			Active:      true,
			Description: "A test image",
		},
	}

	err := SaveDescriptions(path, descriptions)

	require.NoError(t, err)
	assert.FileExists(t, path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "Human-maintained asset descriptions")
	assert.Contains(t, content, "test/image.png")
	assert.Contains(t, content, "A test image")
	assert.Contains(t, content, "active: true")
}

func TestSaveDescriptions_EmptyMap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "descriptions.yml")
	descriptions := map[string]Description{}

	err := SaveDescriptions(path, descriptions)

	require.NoError(t, err)
	assert.FileExists(t, path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "Human-maintained asset descriptions")
}

func TestSaveDescriptions_HeaderComment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "descriptions.yml")
	descriptions := map[string]Description{
		"img.png": {Active: true, Description: "test"},
	}

	err := SaveDescriptions(path, descriptions)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)

	lines := strings.Split(content, "\n")
	assert.True(t, strings.HasPrefix(lines[0], "#"), "first line should be a comment")
	assert.Contains(t, content, "git-tracked")
	assert.Contains(t, content, ".manifest-cache.json")
}

func TestSaveAndLoadDescriptions_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "descriptions.yml")

	original := map[string]Description{
		"alpha/first.png": {
			Active:      true,
			Description: "First image",
		},
		"beta/second.drawio.png": {
			Active:      false,
			Description: "Deprecated diagram",
		},
		"gamma/third.png": {
			Active:      true,
			Description: "",
		},
	}

	err := SaveDescriptions(path, original)
	require.NoError(t, err)

	loaded, err := LoadDescriptions(path)
	require.NoError(t, err)

	require.Len(t, loaded, len(original))
	for key, orig := range original {
		got, exists := loaded[key]
		assert.True(t, exists, "key %q should exist after round-trip", key)
		assert.Equal(t, orig.Active, got.Active, "Active mismatch for %q", key)
		assert.Equal(t, orig.Description, got.Description, "Description mismatch for %q", key)
	}
}

func TestSaveDescriptions_MultipleAssets_SortedOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "descriptions.yml")

	descriptions := map[string]Description{
		"zebra/z.png": {Active: true, Description: "last"},
		"alpha/a.png": {Active: true, Description: "first"},
		"mid/m.png":   {Active: true, Description: "middle"},
	}

	err := SaveDescriptions(path, descriptions)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)

	// All keys should appear in the output
	assert.Contains(t, content, "alpha/a.png")
	assert.Contains(t, content, "mid/m.png")
	assert.Contains(t, content, "zebra/z.png")
}

func TestLoadDescriptions_ActiveDefaultsToFalse(t *testing.T) {
	// YAML bool default is false when omitted
	path := filepath.Join(t.TempDir(), "descriptions.yml")
	content := `assets:
  test/image.png:
    description: "no active field"
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	descriptions, err := LoadDescriptions(path)
	require.NoError(t, err)

	desc := descriptions["test/image.png"]
	assert.False(t, desc.Active, "active should default to false when omitted in YAML")
	assert.Equal(t, "no active field", desc.Description)
}

func TestLoadDescriptions_EmptyDescription(t *testing.T) {
	path := filepath.Join(t.TempDir(), "descriptions.yml")
	content := `assets:
  new/asset.png:
    active: true
    description: ""
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	descriptions, err := LoadDescriptions(path)
	require.NoError(t, err)

	desc := descriptions["new/asset.png"]
	assert.True(t, desc.Active)
	assert.Empty(t, desc.Description)
}
