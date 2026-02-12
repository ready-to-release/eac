package itemcache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

// writeCacheFile creates a file in the cache directory with the given content.
func writeCacheFile(t *testing.T, cacheDir, filename, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(cacheDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, filename), []byte(content), 0o644))
}

// writeManifestWithItems persists a manifest containing the given CachedItems
// keyed by item key. The manifest is written to the standard manifest path.
func writeManifestWithItems(t *testing.T, cacheDir string, items map[string]*CachedItem) {
	t.Helper()
	m := &Manifest{
		Version: manifestVersion,
		Items:   items,
	}
	path := filepath.Join(cacheDir, "item-manifest.json")
	require.NoError(t, os.MkdirAll(cacheDir, 0o755))
	data, err := json.MarshalIndent(m, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))
}

// trackingBuildFunc returns a BuildFunc that records each invocation and writes
// dummy content for each item into the cache directory. The caller can inspect
// the recorded calls after Execute returns.
type buildCall struct {
	items    []Item
	cacheDir string
}

func trackingBuildFunc(calls *[]buildCall) BuildFunc {
	return func(items []Item, cacheDir string) (int, error) {
		*calls = append(*calls, buildCall{items: items, cacheDir: cacheDir})
		for _, item := range items {
			path := filepath.Join(cacheDir, item.CacheFilename)
			if err := os.WriteFile(path, []byte("built:"+item.Key), 0o644); err != nil {
				return 0, err
			}
		}
		return len(items), nil
	}
}

// readFile is a shorthand that reads a file and returns its content as string.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

// sampleItems returns a deterministic set of N items with predictable fields.
func sampleItems(n int) []Item {
	items := make([]Item, n)
	for i := 0; i < n; i++ {
		items[i] = Item{
			Key:           fmt.Sprintf("item-%d", i),
			ContentHash:   fmt.Sprintf("hash-%d", i),
			CacheFilename: fmt.Sprintf("item-%d.out", i),
			OutputRelPath: fmt.Sprintf("item-%d.out", i),
		}
	}
	return items
}

// --- tests ---

func TestExecute_AllCacheMisses(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	items := sampleItems(3)
	var calls []buildCall
	bf := trackingBuildFunc(&calls)

	c := New(cacheDir)
	result, err := c.Execute(items, outputDir, bf, false)

	require.NoError(t, err)
	require.NotNil(t, result)

	// All items should be misses.
	assert.Equal(t, 3, result.TotalItems)
	assert.Equal(t, 0, result.CacheHits)
	assert.Equal(t, 3, result.CacheMisses)
	assert.Equal(t, 3, result.Built)
	assert.Equal(t, 0.0, result.HitRate)

	// buildFunc should have been called exactly once with all 3 items.
	require.Len(t, calls, 1)
	assert.Len(t, calls[0].items, 3)
	assert.Equal(t, cacheDir, calls[0].cacheDir)

	// Output files should exist with content written by buildFunc.
	for _, item := range items {
		content := readFile(t, filepath.Join(outputDir, item.OutputRelPath))
		assert.Equal(t, "built:"+item.Key, content)
	}
}

func TestExecute_AllCacheHits(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	items := sampleItems(3)

	// Pre-populate cache: manifest entries AND cache files.
	cachedItems := make(map[string]*CachedItem, len(items))
	for _, item := range items {
		cachedItems[item.Key] = &CachedItem{
			ContentHash:   item.ContentHash,
			CacheFilename: item.CacheFilename,
			CachedAt:      1000,
		}
		writeCacheFile(t, cacheDir, item.CacheFilename, "cached:"+item.Key)
	}
	writeManifestWithItems(t, cacheDir, cachedItems)

	var calls []buildCall
	bf := trackingBuildFunc(&calls)

	c := New(cacheDir)
	result, err := c.Execute(items, outputDir, bf, false)

	require.NoError(t, err)
	require.NotNil(t, result)

	// All items should be hits.
	assert.Equal(t, 3, result.TotalItems)
	assert.Equal(t, 3, result.CacheHits)
	assert.Equal(t, 0, result.CacheMisses)
	assert.Equal(t, 0, result.Built)
	assert.Equal(t, 100.0, result.HitRate)

	// buildFunc should NOT have been called.
	assert.Empty(t, calls)

	// Output files should contain the original cached content.
	for _, item := range items {
		content := readFile(t, filepath.Join(outputDir, item.OutputRelPath))
		assert.Equal(t, "cached:"+item.Key, content)
	}
}

func TestExecute_PartialHits(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	items := sampleItems(3) // item-0, item-1, item-2

	// Only item-0 is cached.
	cachedItems := map[string]*CachedItem{
		items[0].Key: {
			ContentHash:   items[0].ContentHash,
			CacheFilename: items[0].CacheFilename,
			CachedAt:      1000,
		},
	}
	writeManifestWithItems(t, cacheDir, cachedItems)
	writeCacheFile(t, cacheDir, items[0].CacheFilename, "cached:"+items[0].Key)

	var calls []buildCall
	bf := trackingBuildFunc(&calls)

	c := New(cacheDir)
	result, err := c.Execute(items, outputDir, bf, false)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 3, result.TotalItems)
	assert.Equal(t, 1, result.CacheHits)
	assert.Equal(t, 2, result.CacheMisses)
	assert.Equal(t, 2, result.Built)

	// HitRate = 1/3 * 100 ~= 33.33...
	assert.InDelta(t, 33.33, result.HitRate, 0.01)

	// buildFunc called once with the 2 misses.
	require.Len(t, calls, 1)
	assert.Len(t, calls[0].items, 2)

	missKeys := make(map[string]bool)
	for _, m := range calls[0].items {
		missKeys[m.Key] = true
	}
	assert.True(t, missKeys["item-1"])
	assert.True(t, missKeys["item-2"])
	assert.False(t, missKeys["item-0"])

	// All output files exist.
	for _, item := range items {
		_, err := os.Stat(filepath.Join(outputDir, item.OutputRelPath))
		assert.NoError(t, err, "output file missing for %s", item.Key)
	}
}

func TestExecute_ForceRebuild(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	items := sampleItems(2)

	// Pre-populate full cache.
	cachedItems := make(map[string]*CachedItem, len(items))
	for _, item := range items {
		cachedItems[item.Key] = &CachedItem{
			ContentHash:   item.ContentHash,
			CacheFilename: item.CacheFilename,
			CachedAt:      1000,
		}
		writeCacheFile(t, cacheDir, item.CacheFilename, "old:"+item.Key)
	}
	writeManifestWithItems(t, cacheDir, cachedItems)

	var calls []buildCall
	bf := trackingBuildFunc(&calls)

	c := New(cacheDir)
	result, err := c.Execute(items, outputDir, bf, true) // forceRebuild = true

	require.NoError(t, err)
	require.NotNil(t, result)

	// All should be misses despite being cached.
	assert.Equal(t, 2, result.TotalItems)
	assert.Equal(t, 0, result.CacheHits)
	assert.Equal(t, 2, result.CacheMisses)
	assert.Equal(t, 2, result.Built)
	assert.Equal(t, 0.0, result.HitRate)

	// buildFunc called with all items.
	require.Len(t, calls, 1)
	assert.Len(t, calls[0].items, 2)

	// Output files contain rebuilt content, not old.
	for _, item := range items {
		content := readFile(t, filepath.Join(outputDir, item.OutputRelPath))
		assert.Equal(t, "built:"+item.Key, content)
	}
}

func TestExecute_CacheFileDeletedButManifestExists(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	items := sampleItems(1)

	// Manifest says item is cached, but cache file does NOT exist on disk.
	cachedItems := map[string]*CachedItem{
		items[0].Key: {
			ContentHash:   items[0].ContentHash,
			CacheFilename: items[0].CacheFilename,
			CachedAt:      1000,
		},
	}
	writeManifestWithItems(t, cacheDir, cachedItems)
	// Deliberately NOT creating the cache file.

	var calls []buildCall
	bf := trackingBuildFunc(&calls)

	c := New(cacheDir)
	result, err := c.Execute(items, outputDir, bf, false)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Item should be a miss because the cache file is missing.
	assert.Equal(t, 1, result.TotalItems)
	assert.Equal(t, 0, result.CacheHits)
	assert.Equal(t, 1, result.CacheMisses)
	assert.Equal(t, 1, result.Built)

	// buildFunc called.
	require.Len(t, calls, 1)
	assert.Len(t, calls[0].items, 1)
	assert.Equal(t, items[0].Key, calls[0].items[0].Key)
}

func TestExecute_ContentHashChanged(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	items := []Item{
		{
			Key:           "doc",
			ContentHash:   "newhash",
			CacheFilename: "doc.out",
			OutputRelPath: "doc.out",
		},
	}

	// Manifest has OLD hash.
	cachedItems := map[string]*CachedItem{
		"doc": {
			ContentHash:   "oldhash",
			CacheFilename: "doc.out",
			CachedAt:      1000,
		},
	}
	writeManifestWithItems(t, cacheDir, cachedItems)
	writeCacheFile(t, cacheDir, "doc.out", "stale content")

	var calls []buildCall
	bf := trackingBuildFunc(&calls)

	c := New(cacheDir)
	result, err := c.Execute(items, outputDir, bf, false)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 0, result.CacheHits)
	assert.Equal(t, 1, result.CacheMisses)

	// buildFunc received the item with new hash.
	require.Len(t, calls, 1)
	assert.Equal(t, "newhash", calls[0].items[0].ContentHash)

	// Output has rebuilt content.
	content := readFile(t, filepath.Join(outputDir, "doc.out"))
	assert.Equal(t, "built:doc", content)
}

func TestExecute_EmptyItemList(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	var calls []buildCall
	bf := trackingBuildFunc(&calls)

	c := New(cacheDir)
	result, err := c.Execute([]Item{}, outputDir, bf, false)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 0, result.TotalItems)
	assert.Equal(t, 0, result.CacheHits)
	assert.Equal(t, 0, result.CacheMisses)
	assert.Equal(t, 0, result.Built)
	assert.Equal(t, 0.0, result.HitRate)

	// buildFunc should NOT have been called.
	assert.Empty(t, calls)
}

func TestExecute_BuildFuncFailure(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	items := sampleItems(2)

	failBuild := func(items []Item, cacheDir string) (int, error) {
		return 0, fmt.Errorf("build exploded")
	}

	c := New(cacheDir)
	result, err := c.Execute(items, outputDir, failBuild, false)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "build exploded")
	assert.Contains(t, err.Error(), "building items")
}

func TestManifest_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "item-manifest.json")

	original := &Manifest{
		Version: manifestVersion,
		Items: map[string]*CachedItem{
			"alpha": {
				ContentHash:   "aaa111",
				CacheFilename: "alpha.out",
				CachedAt:      1700000000,
			},
			"beta": {
				ContentHash:   "bbb222",
				CacheFilename: "beta.out",
				CachedAt:      1700000001,
			},
		},
	}

	err := saveManifest(original, path)
	require.NoError(t, err)

	loaded := loadManifest(path, logging.C())
	require.NotNil(t, loaded)
	assert.Equal(t, manifestVersion, loaded.Version)
	require.Len(t, loaded.Items, 2)

	assert.Equal(t, "aaa111", loaded.Items["alpha"].ContentHash)
	assert.Equal(t, "alpha.out", loaded.Items["alpha"].CacheFilename)
	assert.Equal(t, int64(1700000000), loaded.Items["alpha"].CachedAt)

	assert.Equal(t, "bbb222", loaded.Items["beta"].ContentHash)
	assert.Equal(t, "beta.out", loaded.Items["beta"].CacheFilename)
	assert.Equal(t, int64(1700000001), loaded.Items["beta"].CachedAt)
}

func TestManifest_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "item-manifest.json")

	require.NoError(t, os.WriteFile(path, []byte("{not valid json!!!"), 0o644))

	loaded := loadManifest(path, logging.C())
	require.NotNil(t, loaded)
	assert.Equal(t, manifestVersion, loaded.Version)
	assert.Empty(t, loaded.Items)
}

func TestManifest_VersionMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "item-manifest.json")

	wrongVersion := &Manifest{
		Version: 999,
		Items: map[string]*CachedItem{
			"old": {ContentHash: "x", CacheFilename: "old.out", CachedAt: 1},
		},
	}
	data, err := json.Marshal(wrongVersion)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))

	loaded := loadManifest(path, logging.C())
	require.NotNil(t, loaded)
	assert.Equal(t, manifestVersion, loaded.Version)
	assert.Empty(t, loaded.Items, "items from wrong-version manifest should be discarded")
}

func TestPrune_RemovesStaleEntries(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	// 5 items total in the manifest, but only 3 in the current item set.
	allItems := sampleItems(5)
	currentItems := allItems[:3] // item-0, item-1, item-2

	// Pre-populate cache for all 5 items.
	cachedMap := make(map[string]*CachedItem)
	for _, item := range allItems {
		cachedMap[item.Key] = &CachedItem{
			ContentHash:   item.ContentHash,
			CacheFilename: item.CacheFilename,
			CachedAt:      1000,
		}
		writeCacheFile(t, cacheDir, item.CacheFilename, "cached:"+item.Key)
	}
	writeManifestWithItems(t, cacheDir, cachedMap)

	var calls []buildCall
	bf := trackingBuildFunc(&calls)

	c := New(cacheDir)
	result, err := c.Execute(currentItems, outputDir, bf, false)

	require.NoError(t, err)
	assert.Equal(t, 3, result.TotalItems)
	assert.Equal(t, 3, result.CacheHits)

	// buildFunc not called since all 3 current items are hits.
	assert.Empty(t, calls)

	// Verify manifest was persisted with only 3 entries.
	persistedManifest := loadManifest(filepath.Join(cacheDir, "item-manifest.json"), logging.C())
	assert.Len(t, persistedManifest.Items, 3)
	for _, item := range currentItems {
		assert.Contains(t, persistedManifest.Items, item.Key)
	}
	// Stale items should be gone from the manifest.
	assert.NotContains(t, persistedManifest.Items, "item-3")
	assert.NotContains(t, persistedManifest.Items, "item-4")

	// Stale cache files should be deleted from disk.
	for _, stale := range allItems[3:] {
		_, err := os.Stat(filepath.Join(cacheDir, stale.CacheFilename))
		assert.True(t, os.IsNotExist(err), "cache file for %s should have been pruned", stale.Key)
	}

	// Current cache files should still exist.
	for _, item := range currentItems {
		_, err := os.Stat(filepath.Join(cacheDir, item.CacheFilename))
		assert.NoError(t, err, "cache file for %s should still exist", item.Key)
	}
}

func TestPrune_HandlesDeletedFiles(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	currentItems := sampleItems(1) // item-0

	// Manifest has item-0 (current) plus a stale entry whose file is already gone.
	cachedMap := map[string]*CachedItem{
		currentItems[0].Key: {
			ContentHash:   currentItems[0].ContentHash,
			CacheFilename: currentItems[0].CacheFilename,
			CachedAt:      1000,
		},
		"stale-gone": {
			ContentHash:   "gone",
			CacheFilename: "stale-gone.out",
			CachedAt:      500,
		},
	}
	writeManifestWithItems(t, cacheDir, cachedMap)
	writeCacheFile(t, cacheDir, currentItems[0].CacheFilename, "cached:"+currentItems[0].Key)
	// Deliberately NOT creating stale-gone.out -- it's already deleted on disk.

	var calls []buildCall
	bf := trackingBuildFunc(&calls)

	c := New(cacheDir)
	result, err := c.Execute(currentItems, outputDir, bf, false)

	// Should succeed without error even though the stale cache file doesn't exist.
	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalItems)
	assert.Equal(t, 1, result.CacheHits)

	// Stale entry removed from manifest.
	persistedManifest := loadManifest(filepath.Join(cacheDir, "item-manifest.json"), logging.C())
	assert.NotContains(t, persistedManifest.Items, "stale-gone")
	assert.Contains(t, persistedManifest.Items, currentItems[0].Key)
}

func TestExecute_OutputSubdirectories(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	items := []Item{
		{
			Key:           "nested-a",
			ContentHash:   "ha",
			CacheFilename: "nested-a.out",
			OutputRelPath: filepath.Join("sub", "dir", "a.out"),
		},
		{
			Key:           "nested-b",
			ContentHash:   "hb",
			CacheFilename: "nested-b.out",
			OutputRelPath: filepath.Join("another", "b.out"),
		},
		{
			Key:           "top-level",
			ContentHash:   "hc",
			CacheFilename: "top-level.out",
			OutputRelPath: "c.out",
		},
	}

	var calls []buildCall
	bf := trackingBuildFunc(&calls)

	c := New(cacheDir)
	result, err := c.Execute(items, outputDir, bf, false)

	require.NoError(t, err)
	assert.Equal(t, 3, result.TotalItems)
	assert.Equal(t, 3, result.Built)

	// All output files should exist in their nested locations.
	for _, item := range items {
		outPath := filepath.Join(outputDir, item.OutputRelPath)
		content := readFile(t, outPath)
		assert.Equal(t, "built:"+item.Key, content, "wrong content for %s", item.Key)
	}

	// Verify subdirectories were actually created.
	info, err := os.Stat(filepath.Join(outputDir, "sub", "dir"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	info, err = os.Stat(filepath.Join(outputDir, "another"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestExecute_ManifestPersistedAfterBuild(t *testing.T) {
	cacheDir := t.TempDir()
	outputDir := t.TempDir()

	items := []Item{
		{
			Key:           "alpha",
			ContentHash:   "h-alpha",
			CacheFilename: "alpha.out",
			OutputRelPath: "alpha.out",
		},
		{
			Key:           "beta",
			ContentHash:   "h-beta",
			CacheFilename: "beta.out",
			OutputRelPath: "beta.out",
		},
	}

	var calls []buildCall
	bf := trackingBuildFunc(&calls)

	c := New(cacheDir)
	_, err := c.Execute(items, outputDir, bf, false)
	require.NoError(t, err)

	// Manifest JSON file should exist on disk.
	manifestPath := filepath.Join(cacheDir, "item-manifest.json")
	_, statErr := os.Stat(manifestPath)
	require.NoError(t, statErr, "manifest file should exist after Execute")

	// Load and verify contents.
	persisted := loadManifest(manifestPath, logging.C())
	require.Len(t, persisted.Items, 2)

	assert.Equal(t, "h-alpha", persisted.Items["alpha"].ContentHash)
	assert.Equal(t, "alpha.out", persisted.Items["alpha"].CacheFilename)
	assert.NotZero(t, persisted.Items["alpha"].CachedAt)

	assert.Equal(t, "h-beta", persisted.Items["beta"].ContentHash)
	assert.Equal(t, "beta.out", persisted.Items["beta"].CacheFilename)
	assert.NotZero(t, persisted.Items["beta"].CachedAt)

	// Verify the raw JSON is valid and parseable.
	raw, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	var rawManifest map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &rawManifest))
	assert.Equal(t, float64(manifestVersion), rawManifest["version"])
}
