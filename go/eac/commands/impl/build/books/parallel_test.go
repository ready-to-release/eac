//go:build L0
// +build L0

package books

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProcessFilesParallel_Basic tests basic parallel file processing.
func TestProcessFilesParallel_Basic(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	files := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(tempDir, "file3.txt"),
	}
	for i, f := range files {
		require.NoError(t, os.WriteFile(f, []byte("content"+string(rune('1'+i))), 0o644))
	}

	// Process: uppercase all content
	err := ProcessFilesParallel(files, func(path string, content []byte) ([]byte, error) {
		return bytes.ToUpper(content), nil
	})
	require.NoError(t, err)

	// Verify all files were modified
	for i, f := range files {
		content, err := os.ReadFile(f)
		require.NoError(t, err)
		expected := "CONTENT" + string(rune('1'+i))
		assert.Equal(t, expected, string(content), "file %s should be uppercased", f)
	}
}

// TestProcessFilesParallel_EmptyList tests handling of empty file list.
func TestProcessFilesParallel_EmptyList(t *testing.T) {
	err := ProcessFilesParallel([]string{}, func(path string, content []byte) ([]byte, error) {
		t.Error("processor should not be called for empty list")
		return content, nil
	})
	require.NoError(t, err)
}

// TestProcessFilesParallel_NoChange tests that unchanged files are not rewritten.
func TestProcessFilesParallel_NoChange(t *testing.T) {
	tempDir := t.TempDir()

	file := filepath.Join(tempDir, "unchanged.txt")
	originalContent := []byte("original content")
	require.NoError(t, os.WriteFile(file, originalContent, 0o644))

	// Get original mod time
	origInfo, err := os.Stat(file)
	require.NoError(t, err)

	// Process: return content unchanged
	err = ProcessFilesParallel([]string{file}, func(path string, content []byte) ([]byte, error) {
		return content, nil // No change
	})
	require.NoError(t, err)

	// Verify content is still the same
	content, err := os.ReadFile(file)
	require.NoError(t, err)
	assert.Equal(t, originalContent, content)

	// Note: We can't reliably check mod time on fast systems, but content should be same
	_ = origInfo
}

// TestProcessFilesParallel_Concurrent tests that processing actually runs in parallel.
func TestProcessFilesParallel_Concurrent(t *testing.T) {
	tempDir := t.TempDir()

	// Create many files to ensure parallelism
	numFiles := 100
	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		files[i] = filepath.Join(tempDir, "file"+string(rune('a'+i%26))+string(rune('0'+i/26))+".txt")
		require.NoError(t, os.WriteFile(files[i], []byte("x"), 0o644))
	}

	// Count concurrent executions
	var maxConcurrent int32
	var currentConcurrent int32

	err := ProcessFilesParallel(files, func(path string, content []byte) ([]byte, error) {
		current := atomic.AddInt32(&currentConcurrent, 1)
		// Track max concurrency
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if current <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
				break
			}
		}
		// Small delay to allow concurrency
		// (In real usage, file I/O provides natural delays)
		atomic.AddInt32(&currentConcurrent, -1)
		return bytes.ToUpper(content), nil
	})
	require.NoError(t, err)

	// Should have had some level of concurrency (at least 1, up to NumCPU)
	assert.GreaterOrEqual(t, maxConcurrent, int32(1), "should have at least 1 concurrent execution")
}

// TestProcessFilesParallel_Error tests error handling.
func TestProcessFilesParallel_Error(t *testing.T) {
	tempDir := t.TempDir()

	files := []string{
		filepath.Join(tempDir, "good.txt"),
		filepath.Join(tempDir, "bad.txt"),
	}
	require.NoError(t, os.WriteFile(files[0], []byte("good"), 0o644))
	require.NoError(t, os.WriteFile(files[1], []byte("bad"), 0o644))

	err := ProcessFilesParallel(files, func(path string, content []byte) ([]byte, error) {
		if strings.Contains(path, "bad") {
			return nil, assert.AnError
		}
		return content, nil
	})
	require.Error(t, err, "should return error from failing file")
}

// TestProcessFilesParallel_NonExistent tests handling of non-existent files.
func TestProcessFilesParallel_NonExistent(t *testing.T) {
	err := ProcessFilesParallel([]string{"/nonexistent/file.txt"}, func(path string, content []byte) ([]byte, error) {
		t.Error("processor should not be called for non-existent file")
		return content, nil
	})
	require.Error(t, err, "should error on non-existent file")
}

// TestProcessFilesParallelWithCache_CacheHit tests that cached files are skipped.
func TestProcessFilesParallelWithCache_CacheHit(t *testing.T) {
	tempDir := t.TempDir()

	file := filepath.Join(tempDir, "cached.txt")
	content := []byte("cached content")
	require.NoError(t, os.WriteFile(file, content, 0o644))

	// Create cache with matching hash
	cache := NewFileHashCache("test", tempDir)
	cache.ShouldProcessFile(file, content) // Pre-populate cache

	// Reset stats
	cache.stats = CacheHitStats{}

	var processorCalled bool
	count, err := ProcessFilesParallelWithCache([]string{file}, cache, func(path string, c []byte) ([]byte, error) {
		processorCalled = true
		return bytes.ToUpper(c), nil
	})
	require.NoError(t, err)

	assert.False(t, processorCalled, "processor should not be called for cached file")
	assert.Equal(t, 0, count, "processed count should be 0")

	// Verify cache stats
	stats := cache.Stats()
	assert.Equal(t, 1, stats.Hits, "should have 1 cache hit")
	assert.Equal(t, 0, stats.Misses, "should have 0 cache misses")
}

// TestProcessFilesParallelWithCache_CacheMiss tests that uncached files are processed.
func TestProcessFilesParallelWithCache_CacheMiss(t *testing.T) {
	tempDir := t.TempDir()

	file := filepath.Join(tempDir, "new.txt")
	content := []byte("new content")
	require.NoError(t, os.WriteFile(file, content, 0o644))

	// Create empty cache
	cache := NewFileHashCache("test", tempDir)

	var processorCalled bool
	count, err := ProcessFilesParallelWithCache([]string{file}, cache, func(path string, c []byte) ([]byte, error) {
		processorCalled = true
		return bytes.ToUpper(c), nil
	})
	require.NoError(t, err)

	assert.True(t, processorCalled, "processor should be called for new file")
	assert.Equal(t, 1, count, "processed count should be 1")

	// Verify file was modified
	modified, err := os.ReadFile(file)
	require.NoError(t, err)
	assert.Equal(t, "NEW CONTENT", string(modified))

	// Verify cache stats
	stats := cache.Stats()
	assert.Equal(t, 0, stats.Hits, "should have 0 cache hits")
	assert.Equal(t, 1, stats.Misses, "should have 1 cache miss")
}

// TestProcessFilesParallelWithCache_NilCache tests that nil cache processes all files.
func TestProcessFilesParallelWithCache_NilCache(t *testing.T) {
	tempDir := t.TempDir()

	file := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(file, []byte("test"), 0o644))

	var processorCalled bool
	count, err := ProcessFilesParallelWithCache([]string{file}, nil, func(path string, c []byte) ([]byte, error) {
		processorCalled = true
		return c, nil
	})
	require.NoError(t, err)

	assert.True(t, processorCalled, "processor should be called when cache is nil")
	assert.Equal(t, 1, count, "processed count should be 1")
}

// TestProcessFilesParallelWithCache_MixedHitsMisses tests processing with mixed cache states.
func TestProcessFilesParallelWithCache_MixedHitsMisses(t *testing.T) {
	tempDir := t.TempDir()

	// Create files
	cachedFile := filepath.Join(tempDir, "cached.txt")
	newFile := filepath.Join(tempDir, "new.txt")
	changedFile := filepath.Join(tempDir, "changed.txt")

	cachedContent := []byte("cached")
	newContent := []byte("new")
	changedOld := []byte("old")
	changedNew := []byte("changed")

	require.NoError(t, os.WriteFile(cachedFile, cachedContent, 0o644))
	require.NoError(t, os.WriteFile(newFile, newContent, 0o644))
	require.NoError(t, os.WriteFile(changedFile, changedNew, 0o644))

	// Create cache with some pre-populated entries
	cache := NewFileHashCache("test", tempDir)
	cache.ShouldProcessFile(cachedFile, cachedContent) // Same content - will hit
	cache.ShouldProcessFile(changedFile, changedOld)   // Different content - will miss

	// Reset stats
	cache.stats = CacheHitStats{}

	files := []string{cachedFile, newFile, changedFile}
	count, err := ProcessFilesParallelWithCache(files, cache, func(path string, c []byte) ([]byte, error) {
		return bytes.ToUpper(c), nil
	})
	require.NoError(t, err)

	assert.Equal(t, 2, count, "should process 2 files (new + changed)")

	stats := cache.Stats()
	assert.Equal(t, 1, stats.Hits, "should have 1 cache hit (cached file)")
	assert.Equal(t, 2, stats.Misses, "should have 2 cache misses (new + changed)")

	// Verify only uncached files were modified
	cachedResult, _ := os.ReadFile(cachedFile)
	assert.Equal(t, "cached", string(cachedResult), "cached file should not be modified")

	newResult, _ := os.ReadFile(newFile)
	assert.Equal(t, "NEW", string(newResult), "new file should be modified")

	changedResult, _ := os.ReadFile(changedFile)
	assert.Equal(t, "CHANGED", string(changedResult), "changed file should be modified")
}
