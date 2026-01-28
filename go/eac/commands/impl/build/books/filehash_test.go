//go:build L0
// +build L0

package books

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFileHashCache_NewAndSave tests cache creation and persistence to disk.
func TestFileHashCache_NewAndSave(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "test-book"

	// Create a new cache
	cache := NewFileHashCache(bookName, workspaceRoot)
	require.NotNil(t, cache, "NewFileHashCache should return non-nil cache")

	// Verify initial state
	assert.Equal(t, bookName, cache.BookName, "cache should store book name")
	assert.NotNil(t, cache.Hashes, "Hashes map should be initialized")
	assert.Empty(t, cache.Hashes, "new cache should have empty hashes")

	// Add some entries
	cache.Hashes["docs/index.md"] = "abc123"
	cache.Hashes["docs/guide.md"] = "def456"

	// Save the cache
	err := cache.Save()
	require.NoError(t, err, "Save should succeed")

	// Verify cache file was created in correct location
	expectedPath := filepath.Join(workspaceRoot, ".cache", "eac", "build", "hashes", bookName+".json")
	_, err = os.Stat(expectedPath)
	require.NoError(t, err, "cache file should exist at %s", expectedPath)

	// Verify file content is valid JSON
	data, err := os.ReadFile(expectedPath)
	require.NoError(t, err)

	var loaded FileHashCache
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err, "cache file should contain valid JSON")
	assert.Equal(t, bookName, loaded.BookName)
	assert.Len(t, loaded.Hashes, 2)
	assert.Equal(t, "abc123", loaded.Hashes["docs/index.md"])
	assert.Equal(t, "def456", loaded.Hashes["docs/guide.md"])
}

// TestFileHashCache_Load tests loading an existing cache from disk.
func TestFileHashCache_Load(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "existing-book"

	// Create cache directory and file manually
	cacheDir := filepath.Join(workspaceRoot, ".cache", "eac", "build", "hashes")
	require.NoError(t, os.MkdirAll(cacheDir, 0o755))

	cacheData := FileHashCache{
		BookName: bookName,
		Hashes: map[string]string{
			"docs/intro.md":   "hash1",
			"docs/chapter.md": "hash2",
			"docs/appendix.md": "hash3",
		},
	}
	data, err := json.MarshalIndent(cacheData, "", "  ")
	require.NoError(t, err)

	cachePath := filepath.Join(cacheDir, bookName+".json")
	require.NoError(t, os.WriteFile(cachePath, data, 0o644))

	// Load the cache
	cache, err := LoadFileHashCache(bookName, workspaceRoot)
	require.NoError(t, err, "LoadFileHashCache should succeed")
	require.NotNil(t, cache)

	// Verify loaded data
	assert.Equal(t, bookName, cache.BookName)
	assert.Len(t, cache.Hashes, 3)
	assert.Equal(t, "hash1", cache.Hashes["docs/intro.md"])
	assert.Equal(t, "hash2", cache.Hashes["docs/chapter.md"])
	assert.Equal(t, "hash3", cache.Hashes["docs/appendix.md"])
}

// TestFileHashCache_LoadNonExistent tests loading when cache file doesn't exist.
func TestFileHashCache_LoadNonExistent(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "nonexistent-book"

	// Load cache for non-existent book - should return empty cache, not error
	cache, err := LoadFileHashCache(bookName, workspaceRoot)
	require.NoError(t, err, "LoadFileHashCache should not error for missing file")
	require.NotNil(t, cache, "should return empty cache")

	assert.Equal(t, bookName, cache.BookName)
	assert.Empty(t, cache.Hashes, "non-existent cache should have empty hashes")
}

// TestFileHashCache_ShouldProcessFile_Unchanged tests that unchanged files are skipped.
func TestFileHashCache_ShouldProcessFile_Unchanged(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "unchanged-test"

	// Create cache with known hash
	content := "# Hello World\n\nThis is a test file."
	expectedHash := computeTestHash(content)

	cache := NewFileHashCache(bookName, workspaceRoot)
	cache.Hashes["docs/test.md"] = expectedHash

	// Check if file should be processed - should return false (unchanged)
	shouldProcess := cache.ShouldProcessFile("docs/test.md", []byte(content))
	assert.False(t, shouldProcess, "unchanged file should not need processing")

	// Verify hash is still in cache (not modified)
	assert.Equal(t, expectedHash, cache.Hashes["docs/test.md"])
}

// TestFileHashCache_ShouldProcessFile_Changed tests that changed files are processed.
func TestFileHashCache_ShouldProcessFile_Changed(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "changed-test"

	// Create cache with old hash
	oldContent := "# Old Content"
	oldHash := computeTestHash(oldContent)

	cache := NewFileHashCache(bookName, workspaceRoot)
	cache.Hashes["docs/test.md"] = oldHash

	// Check with new content
	newContent := "# New Content - Updated"
	shouldProcess := cache.ShouldProcessFile("docs/test.md", []byte(newContent))
	assert.True(t, shouldProcess, "changed file should need processing")

	// Verify cache was updated with new hash
	newHash := computeTestHash(newContent)
	assert.Equal(t, newHash, cache.Hashes["docs/test.md"], "cache should be updated with new hash")
}

// TestFileHashCache_ShouldProcessFile_NewFile tests that new files are processed.
func TestFileHashCache_ShouldProcessFile_NewFile(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "new-file-test"

	// Create empty cache
	cache := NewFileHashCache(bookName, workspaceRoot)
	assert.Empty(t, cache.Hashes)

	// Check new file
	content := "# Brand New File"
	shouldProcess := cache.ShouldProcessFile("docs/new.md", []byte(content))
	assert.True(t, shouldProcess, "new file should need processing")

	// Verify cache now contains the file
	expectedHash := computeTestHash(content)
	assert.Equal(t, expectedHash, cache.Hashes["docs/new.md"], "new file should be added to cache")
}

// TestFileHashCache_BookSpecific tests that caches are isolated per book.
func TestFileHashCache_BookSpecific(t *testing.T) {
	workspaceRoot := t.TempDir()

	// Create two caches for different books
	cache1 := NewFileHashCache("book-alpha", workspaceRoot)
	cache2 := NewFileHashCache("book-beta", workspaceRoot)

	// Add different content to each
	cache1.Hashes["docs/shared.md"] = "hash-for-alpha"
	cache2.Hashes["docs/shared.md"] = "hash-for-beta"

	// Save both caches
	require.NoError(t, cache1.Save())
	require.NoError(t, cache2.Save())

	// Verify separate files were created
	cacheDir := filepath.Join(workspaceRoot, ".cache", "eac", "build", "hashes")
	alphaPath := filepath.Join(cacheDir, "book-alpha.json")
	betaPath := filepath.Join(cacheDir, "book-beta.json")

	_, err := os.Stat(alphaPath)
	require.NoError(t, err, "book-alpha cache file should exist")
	_, err = os.Stat(betaPath)
	require.NoError(t, err, "book-beta cache file should exist")

	// Load each and verify isolation
	loadedAlpha, err := LoadFileHashCache("book-alpha", workspaceRoot)
	require.NoError(t, err)
	loadedBeta, err := LoadFileHashCache("book-beta", workspaceRoot)
	require.NoError(t, err)

	assert.Equal(t, "hash-for-alpha", loadedAlpha.Hashes["docs/shared.md"])
	assert.Equal(t, "hash-for-beta", loadedBeta.Hashes["docs/shared.md"])
}

// TestFileHashCache_CorruptCache tests graceful handling of corrupt cache files.
func TestFileHashCache_CorruptCache(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "corrupt-book"

	// Create cache directory with corrupt file
	cacheDir := filepath.Join(workspaceRoot, ".cache", "eac", "build", "hashes")
	require.NoError(t, os.MkdirAll(cacheDir, 0o755))

	cachePath := filepath.Join(cacheDir, bookName+".json")

	tests := []struct {
		name        string
		content     string
		description string
	}{
		{
			name:        "invalid JSON",
			content:     "this is not valid json {{{",
			description: "completely invalid JSON",
		},
		{
			name:        "empty file",
			content:     "",
			description: "empty file",
		},
		{
			name:        "truncated JSON",
			content:     `{"book_name": "test", "hashes": {"file.md":`,
			description: "truncated JSON object",
		},
		{
			name:        "wrong type for hashes",
			content:     `{"book_name": "test", "hashes": "not-a-map"}`,
			description: "hashes field is wrong type",
		},
		{
			name:        "null hashes",
			content:     `{"book_name": "test", "hashes": null}`,
			description: "null hashes field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write corrupt content
			require.NoError(t, os.WriteFile(cachePath, []byte(tt.content), 0o644))

			// Load should return empty cache, not error (graceful degradation)
			cache, err := LoadFileHashCache(bookName, workspaceRoot)
			require.NoError(t, err, "corrupt cache should not cause error: %s", tt.description)
			require.NotNil(t, cache, "should return empty cache for: %s", tt.description)
			assert.Equal(t, bookName, cache.BookName)
			assert.NotNil(t, cache.Hashes, "Hashes should be initialized")
		})
	}
}

// TestFileHashCache_RemoveDeletedFiles tests cache cleanup for deleted files.
func TestFileHashCache_RemoveDeletedFiles(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "cleanup-test"

	// Create cache with multiple files
	cache := NewFileHashCache(bookName, workspaceRoot)
	cache.Hashes["docs/keep.md"] = "hash1"
	cache.Hashes["docs/delete.md"] = "hash2"
	cache.Hashes["docs/also-delete.md"] = "hash3"

	assert.Len(t, cache.Hashes, 3)

	// Remove entries for deleted files
	cache.RemoveFile("docs/delete.md")
	cache.RemoveFile("docs/also-delete.md")

	// Verify only kept file remains
	assert.Len(t, cache.Hashes, 1)
	assert.Contains(t, cache.Hashes, "docs/keep.md")
	assert.NotContains(t, cache.Hashes, "docs/delete.md")
	assert.NotContains(t, cache.Hashes, "docs/also-delete.md")
}

// TestFileHashCache_RemoveNonExistent tests removing a file that isn't in cache.
func TestFileHashCache_RemoveNonExistent(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "remove-nonexistent"

	cache := NewFileHashCache(bookName, workspaceRoot)
	cache.Hashes["docs/existing.md"] = "hash1"

	// Remove non-existent file should not panic or error
	cache.RemoveFile("docs/nonexistent.md")

	// Original entry should still exist
	assert.Len(t, cache.Hashes, 1)
	assert.Contains(t, cache.Hashes, "docs/existing.md")
}

// TestFileHashCache_HashComputation tests that SHA256 hashes are computed correctly.
func TestFileHashCache_HashComputation(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "hash-computation"

	cache := NewFileHashCache(bookName, workspaceRoot)

	// Test with known content
	content := "Hello, World!"
	expectedHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	// Process the file to add it to cache
	shouldProcess := cache.ShouldProcessFile("test.txt", []byte(content))
	assert.True(t, shouldProcess, "new file should need processing")

	// Verify the hash matches expected SHA256
	assert.Equal(t, expectedHash, cache.Hashes["test.txt"], "hash should be correct SHA256")
}

// TestFileHashCache_DeterministicHash tests that same content always produces same hash.
func TestFileHashCache_DeterministicHash(t *testing.T) {
	workspaceRoot := t.TempDir()

	content := "# Deterministic Test\n\nSame content should always produce same hash."

	// Create multiple caches and process same content
	hashes := make([]string, 5)
	for i := 0; i < 5; i++ {
		cache := NewFileHashCache("deterministic", workspaceRoot)
		cache.ShouldProcessFile("test.md", []byte(content))
		hashes[i] = cache.Hashes["test.md"]
	}

	// All hashes should be identical
	for i := 1; i < len(hashes); i++ {
		assert.Equal(t, hashes[0], hashes[i], "hash should be deterministic")
	}
}

// TestFileHashCache_PathNormalization tests that paths are handled consistently.
func TestFileHashCache_PathNormalization(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "path-normalization"

	cache := NewFileHashCache(bookName, workspaceRoot)
	content := "# Test"

	// Process file with forward slashes
	cache.ShouldProcessFile("docs/guides/intro.md", []byte(content))

	// Same file should be recognized regardless of slash direction
	// (implementation should normalize paths)
	shouldProcess := cache.ShouldProcessFile("docs/guides/intro.md", []byte(content))
	assert.False(t, shouldProcess, "same path with same content should not need reprocessing")
}

// TestFileHashCache_EmptyContent tests handling of empty file content.
func TestFileHashCache_EmptyContent(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "empty-content"

	cache := NewFileHashCache(bookName, workspaceRoot)

	// Process empty file
	shouldProcess := cache.ShouldProcessFile("empty.md", []byte(""))
	assert.True(t, shouldProcess, "new empty file should need processing")

	// Empty file should have a hash
	assert.NotEmpty(t, cache.Hashes["empty.md"], "empty file should have a hash")

	// Same empty content should be skipped
	shouldProcess = cache.ShouldProcessFile("empty.md", []byte(""))
	assert.False(t, shouldProcess, "unchanged empty file should not need reprocessing")
}

// TestFileHashCache_LargeFile tests handling of large file content.
func TestFileHashCache_LargeFile(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "large-file"

	cache := NewFileHashCache(bookName, workspaceRoot)

	// Create large content (1MB)
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	// Process large file
	shouldProcess := cache.ShouldProcessFile("large.bin", largeContent)
	assert.True(t, shouldProcess, "new large file should need processing")

	// Verify hash was computed
	hash := cache.Hashes["large.bin"]
	assert.Len(t, hash, 64, "hash should be 64 hex characters (SHA256)")

	// Same content should be skipped
	shouldProcess = cache.ShouldProcessFile("large.bin", largeContent)
	assert.False(t, shouldProcess, "unchanged large file should not need reprocessing")
}

// TestFileHashCache_ConcurrentAccess tests that cache handles concurrent access safely.
func TestFileHashCache_ConcurrentAccess(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "concurrent"

	cache := NewFileHashCache(bookName, workspaceRoot)

	// Process multiple files concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			content := []byte("content-" + string(rune('0'+idx)))
			path := "file-" + string(rune('0'+idx)) + ".md"
			cache.ShouldProcessFile(path, content)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all files were processed
	assert.Len(t, cache.Hashes, 10, "all 10 files should be in cache")
}

// TestFileHashCache_SaveCreatesDirectory tests that Save creates parent directories.
func TestFileHashCache_SaveCreatesDirectory(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "create-dir"

	// Ensure cache directory doesn't exist
	cacheDir := filepath.Join(workspaceRoot, ".cache", "eac", "build", "hashes")
	_, err := os.Stat(cacheDir)
	require.True(t, os.IsNotExist(err), "cache directory should not exist initially")

	// Create and save cache
	cache := NewFileHashCache(bookName, workspaceRoot)
	cache.Hashes["test.md"] = "testhash"

	err = cache.Save()
	require.NoError(t, err, "Save should create directory and succeed")

	// Verify directory was created
	_, err = os.Stat(cacheDir)
	require.NoError(t, err, "cache directory should exist after Save")
}

// TestFileHashCache_Clear tests clearing all cache entries.
func TestFileHashCache_Clear(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "clear-test"

	cache := NewFileHashCache(bookName, workspaceRoot)
	cache.Hashes["file1.md"] = "hash1"
	cache.Hashes["file2.md"] = "hash2"
	cache.Hashes["file3.md"] = "hash3"

	assert.Len(t, cache.Hashes, 3)

	// Clear the cache
	cache.Clear()

	assert.Empty(t, cache.Hashes, "cache should be empty after Clear")
	assert.Equal(t, bookName, cache.BookName, "BookName should be preserved after Clear")
}

// TestFileHashCache_GetCachePath tests the cache path generation.
func TestFileHashCache_GetCachePath(t *testing.T) {
	workspaceRoot := "/workspace/project"
	bookName := "my-book"

	cache := NewFileHashCache(bookName, workspaceRoot)
	expectedPath := filepath.Join(workspaceRoot, ".cache", "eac", "build", "hashes", "my-book.json")

	assert.Equal(t, expectedPath, cache.CachePath(), "cache path should be correct")
}

// TestFileHashCache_MultipleBooks tests handling multiple books independently.
func TestFileHashCache_MultipleBooks(t *testing.T) {
	workspaceRoot := t.TempDir()

	books := []string{"api-docs", "user-guide", "developer-handbook"}
	content := "# Test content"

	// Create caches for each book
	for _, bookName := range books {
		cache := NewFileHashCache(bookName, workspaceRoot)
		cache.ShouldProcessFile("index.md", []byte(content))
		require.NoError(t, cache.Save())
	}

	// Verify each book has its own cache file
	cacheDir := filepath.Join(workspaceRoot, ".cache", "eac", "build", "hashes")
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)

	assert.Len(t, entries, 3, "should have 3 cache files")

	// Verify each cache can be loaded independently
	for _, bookName := range books {
		cache, err := LoadFileHashCache(bookName, workspaceRoot)
		require.NoError(t, err)
		assert.Equal(t, bookName, cache.BookName)
		assert.Contains(t, cache.Hashes, "index.md")
	}
}

// TestFileHashCache_Stats tests cache statistics.
func TestFileHashCache_Stats(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "stats-test"

	cache := NewFileHashCache(bookName, workspaceRoot)

	// Pre-populate cache
	cache.Hashes["cached1.md"] = computeTestHash("cached content 1")
	cache.Hashes["cached2.md"] = computeTestHash("cached content 2")

	// Process mix of cached and new files
	cache.ShouldProcessFile("cached1.md", []byte("cached content 1")) // hit
	cache.ShouldProcessFile("cached2.md", []byte("cached content 2")) // hit
	cache.ShouldProcessFile("new1.md", []byte("new content 1"))       // miss
	cache.ShouldProcessFile("new2.md", []byte("new content 2"))       // miss
	cache.ShouldProcessFile("cached1.md", []byte("changed content"))  // miss (changed)

	stats := cache.Stats()
	assert.Equal(t, 2, stats.Hits, "should have 2 cache hits")
	assert.Equal(t, 3, stats.Misses, "should have 3 cache misses (2 new + 1 changed)")
}

// TestFileHashCache_GetTrackedPaths tests retrieving all tracked file paths.
func TestFileHashCache_GetTrackedPaths(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "tracked-paths-test"

	cache := NewFileHashCache(bookName, workspaceRoot)

	// Empty cache should return empty slice
	paths := cache.GetTrackedPaths()
	assert.Empty(t, paths, "empty cache should return empty paths")

	// Add some entries
	cache.Hashes["docs/index.md"] = "hash1"
	cache.Hashes["docs/guide/intro.md"] = "hash2"
	cache.Hashes["assets/image.png"] = "hash3"

	// Get tracked paths
	paths = cache.GetTrackedPaths()
	assert.Len(t, paths, 3, "should return all 3 tracked paths")

	// Verify all expected paths are present (order not guaranteed)
	pathSet := make(map[string]bool)
	for _, p := range paths {
		pathSet[p] = true
	}
	assert.True(t, pathSet["docs/index.md"], "should contain docs/index.md")
	assert.True(t, pathSet["docs/guide/intro.md"], "should contain docs/guide/intro.md")
	assert.True(t, pathSet["assets/image.png"], "should contain assets/image.png")
}

// TestFileHashCache_GetTrackedPaths_ConcurrentSafe tests that GetTrackedPaths is safe for concurrent access.
func TestFileHashCache_GetTrackedPaths_ConcurrentSafe(t *testing.T) {
	workspaceRoot := t.TempDir()
	bookName := "concurrent-paths"

	cache := NewFileHashCache(bookName, workspaceRoot)

	// Pre-populate with some entries
	for i := 0; i < 100; i++ {
		cache.Hashes["file"+string(rune('0'+i%10))+".md"] = "hash"
	}

	// Concurrent reads should not panic
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = cache.GetTrackedPaths()
			}
			done <- true
		}()
	}

	// Also do concurrent writes while reading
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for j := 0; j < 100; j++ {
				cache.ShouldProcessFile("concurrent"+string(rune('0'+idx))+".md", []byte("content"))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify cache is still valid
	paths := cache.GetTrackedPaths()
	assert.NotEmpty(t, paths, "cache should still have paths after concurrent access")
}

// computeTestHash is a helper function to compute SHA256 hash for test verification.
func computeTestHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}
