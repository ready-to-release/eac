package docsmanifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadCache_FileNotFound(t *testing.T) {
	cache, err := LoadCache(filepath.Join(t.TempDir(), "nonexistent.json"))

	require.NoError(t, err, "missing file should not be an error")
	assert.Nil(t, cache, "should return nil for non-existent file")
}

func TestLoadCache_ValidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".manifest-cache.json")
	content := `{
  "generated": "2025-01-15T10:00:00Z",
  "assets": {
    "arch/overview.drawio.png": {
      "usedIn": ["docs/index.md", "docs/arch.md"],
      "type": "drawio",
      "category": "arch",
      "lastModified": "2025-01-10T08:00:00Z",
      "size": 4096,
      "hash": "abc123"
    },
    "icons/logo.png": {
      "usedIn": [],
      "type": "image",
      "category": "icons"
    }
  },
  "stats": {
    "total": 2,
    "used": 1,
    "unused": 1,
    "byCategory": {
      "arch": {"total": 1, "used": 1, "unused": 0},
      "icons": {"total": 1, "used": 0, "unused": 1}
    }
  }
}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	cache, err := LoadCache(path)

	require.NoError(t, err)
	require.NotNil(t, cache)
	assert.Equal(t, "2025-01-15T10:00:00Z", cache.Generated)
	assert.Len(t, cache.Assets, 2)

	overview := cache.Assets["arch/overview.drawio.png"]
	assert.Equal(t, []string{"docs/index.md", "docs/arch.md"}, overview.UsedIn)
	assert.Equal(t, "drawio", overview.Type)
	assert.Equal(t, "arch", overview.Category)
	assert.Equal(t, "2025-01-10T08:00:00Z", overview.LastModified)
	assert.Equal(t, int64(4096), overview.Size)
	assert.Equal(t, "abc123", overview.Hash)

	logo := cache.Assets["icons/logo.png"]
	assert.Empty(t, logo.UsedIn)
	assert.Equal(t, "image", logo.Type)

	require.NotNil(t, cache.Stats)
	assert.Equal(t, 2, cache.Stats.Total)
	assert.Equal(t, 1, cache.Stats.Used)
	assert.Equal(t, 1, cache.Stats.Unused)
	assert.Len(t, cache.Stats.ByCategory, 2)
}

func TestLoadCache_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".manifest-cache.json")
	require.NoError(t, os.WriteFile(path, []byte("{not valid json"), 0o644))

	cache, err := LoadCache(path)

	assert.Error(t, err)
	assert.Nil(t, cache)
}

func TestLoadCache_EmptyAssetsMap(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".manifest-cache.json")
	content := `{"generated": "2025-01-15T10:00:00Z", "assets": {}}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	cache, err := LoadCache(path)

	require.NoError(t, err)
	require.NotNil(t, cache)
	assert.Empty(t, cache.Assets)
}

func TestSaveCache_CreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".manifest-cache.json")
	cache := &CacheManifest{
		Generated: "2025-01-15T10:00:00Z",
		Assets: map[string]CacheAsset{
			"test/image.png": {
				UsedIn:   []string{"docs/readme.md"},
				Type:     "image",
				Category: "test",
				Size:     1024,
				Hash:     "deadbeef",
			},
		},
		Stats: &ManifestStats{
			Total:  1,
			Used:   1,
			Unused: 0,
		},
	}

	err := SaveCache(cache, path)

	require.NoError(t, err)
	assert.FileExists(t, path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Contains(t, parsed, "generated")
	assert.Contains(t, parsed, "assets")
}

func TestSaveCache_IndentedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".manifest-cache.json")
	cache := &CacheManifest{
		Generated: "2025-01-15T10:00:00Z",
		Assets: map[string]CacheAsset{
			"img.png": {UsedIn: []string{}},
		},
	}

	err := SaveCache(cache, path)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)

	// MarshalIndent with 2-space indentation
	assert.Contains(t, content, "  \"generated\"")
}

func TestSaveAndLoadCache_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".manifest-cache.json")

	original := &CacheManifest{
		Generated: "2025-01-15T10:00:00Z",
		Assets: map[string]CacheAsset{
			"arch/diagram.drawio.png": {
				UsedIn:       []string{"docs/arch.md", "docs/index.md"},
				Type:         "drawio",
				Category:     "arch",
				LastModified: "2025-01-10T08:00:00Z",
				Size:         4096,
				Hash:         "abc123def456",
			},
			"icons/logo.png": {
				UsedIn:   []string{},
				Type:     "image",
				Category: "icons",
				Size:     512,
			},
		},
		Stats: &ManifestStats{
			Total:  2,
			Used:   1,
			Unused: 1,
			ByCategory: map[string]CategoryStats{
				"arch":  {Total: 1, Used: 1, Unused: 0},
				"icons": {Total: 1, Used: 0, Unused: 1},
			},
		},
	}

	err := SaveCache(original, path)
	require.NoError(t, err)

	loaded, err := LoadCache(path)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, original.Generated, loaded.Generated)
	assert.Len(t, loaded.Assets, len(original.Assets))

	for key, origAsset := range original.Assets {
		gotAsset, exists := loaded.Assets[key]
		require.True(t, exists, "asset %q should exist after round-trip", key)
		assert.Equal(t, origAsset.UsedIn, gotAsset.UsedIn, "UsedIn mismatch for %q", key)
		assert.Equal(t, origAsset.Type, gotAsset.Type, "Type mismatch for %q", key)
		assert.Equal(t, origAsset.Category, gotAsset.Category, "Category mismatch for %q", key)
		assert.Equal(t, origAsset.LastModified, gotAsset.LastModified, "LastModified mismatch for %q", key)
		assert.Equal(t, origAsset.Size, gotAsset.Size, "Size mismatch for %q", key)
		assert.Equal(t, origAsset.Hash, gotAsset.Hash, "Hash mismatch for %q", key)
	}

	require.NotNil(t, loaded.Stats)
	assert.Equal(t, original.Stats.Total, loaded.Stats.Total)
	assert.Equal(t, original.Stats.Used, loaded.Stats.Used)
	assert.Equal(t, original.Stats.Unused, loaded.Stats.Unused)
	assert.Len(t, loaded.Stats.ByCategory, len(original.Stats.ByCategory))
}

func TestSaveCache_NilStats(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".manifest-cache.json")
	cache := &CacheManifest{
		Generated: "2025-01-15T10:00:00Z",
		Assets:    map[string]CacheAsset{},
		Stats:     nil,
	}

	err := SaveCache(cache, path)
	require.NoError(t, err)

	loaded, err := LoadCache(path)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Nil(t, loaded.Stats, "nil stats should remain nil after round-trip due to omitempty")
}

func TestSaveCache_EmptyUsedIn(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".manifest-cache.json")
	cache := &CacheManifest{
		Generated: "2025-01-15T10:00:00Z",
		Assets: map[string]CacheAsset{
			"unused.png": {
				UsedIn:   []string{},
				Type:     "image",
				Category: "misc",
			},
		},
	}

	err := SaveCache(cache, path)
	require.NoError(t, err)

	loaded, err := LoadCache(path)
	require.NoError(t, err)

	asset := loaded.Assets["unused.png"]
	assert.Empty(t, asset.UsedIn)
}

func TestSaveCache_OmitsEmptyOptionalFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".manifest-cache.json")
	cache := &CacheManifest{
		Generated: "2025-01-15T10:00:00Z",
		Assets: map[string]CacheAsset{
			"minimal.png": {
				UsedIn: []string{"docs/a.md"},
				// All optional fields left at zero values
			},
		},
	}

	err := SaveCache(cache, path)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)

	// Fields with omitempty should not appear for zero values
	assert.NotContains(t, content, "\"type\"")
	assert.NotContains(t, content, "\"category\"")
	assert.NotContains(t, content, "\"lastModified\"")
	assert.NotContains(t, content, "\"hash\"")
}

func TestCacheManifest_JSONFieldNames(t *testing.T) {
	cache := &CacheManifest{
		Generated: "2025-01-15T10:00:00Z",
		Assets: map[string]CacheAsset{
			"test.png": {
				UsedIn:       []string{"a.md"},
				Type:         "image",
				Category:     "test",
				LastModified: "2025-01-10T08:00:00Z",
				Size:         100,
				Hash:         "abc",
			},
		},
		Stats: &ManifestStats{
			Total:  1,
			Used:   1,
			Unused: 0,
			ByCategory: map[string]CategoryStats{
				"test": {Total: 1, Used: 1, Unused: 0},
			},
		},
	}

	data, err := json.Marshal(cache)
	require.NoError(t, err)
	content := string(data)

	// Verify JSON field naming conventions
	assert.Contains(t, content, "\"generated\"")
	assert.Contains(t, content, "\"assets\"")
	assert.Contains(t, content, "\"usedIn\"")
	assert.Contains(t, content, "\"lastModified\"")
	assert.Contains(t, content, "\"byCategory\"")
}
