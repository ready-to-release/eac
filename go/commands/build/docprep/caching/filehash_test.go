package caching

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileHashCache(t *testing.T) {
	cache := NewFileHashCache("test-book", "/tmp/workspace", nil)

	if cache.BookName != "test-book" {
		t.Errorf("Expected BookName 'test-book', got %q", cache.BookName)
	}
	if len(cache.Hashes) != 0 {
		t.Errorf("Expected empty hashes, got %d", len(cache.Hashes))
	}
}

func TestFileHashCache_ShouldProcessFile(t *testing.T) {
	cache := NewFileHashCache("test-book", "/tmp/workspace", nil)

	content := []byte("hello world")

	// First time: should process (new file)
	if !cache.ShouldProcessFile("test.md", content) {
		t.Errorf("Expected ShouldProcessFile=true for new file")
	}

	// Same content: should skip (cached)
	if cache.ShouldProcessFile("test.md", content) {
		t.Errorf("Expected ShouldProcessFile=false for unchanged file")
	}

	// Changed content: should process
	changed := []byte("hello world updated")
	if !cache.ShouldProcessFile("test.md", changed) {
		t.Errorf("Expected ShouldProcessFile=true for changed file")
	}
}

func TestFileHashCache_Stats(t *testing.T) {
	cache := NewFileHashCache("test-book", "/tmp/workspace", nil)

	content := []byte("hello")
	cache.ShouldProcessFile("a.md", content) // miss (new)
	cache.ShouldProcessFile("a.md", content) // hit
	cache.ShouldProcessFile("b.md", content) // miss (new)
	cache.ShouldProcessFile("b.md", content) // hit

	stats := cache.Stats()
	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 2 {
		t.Errorf("Expected 2 misses, got %d", stats.Misses)
	}
}

func TestFileHashCache_RemoveFile(t *testing.T) {
	cache := NewFileHashCache("test-book", "/tmp/workspace", nil)

	content := []byte("hello")
	cache.ShouldProcessFile("a.md", content)
	cache.RemoveFile("a.md")

	// After removal, file should be treated as new again
	if !cache.ShouldProcessFile("a.md", content) {
		t.Errorf("Expected ShouldProcessFile=true after RemoveFile")
	}
}

func TestFileHashCache_Clear(t *testing.T) {
	cache := NewFileHashCache("test-book", "/tmp/workspace", nil)

	cache.ShouldProcessFile("a.md", []byte("a"))
	cache.ShouldProcessFile("b.md", []byte("b"))
	cache.Clear()

	paths := cache.GetTrackedPaths()
	if len(paths) != 0 {
		t.Errorf("Expected 0 tracked paths after Clear, got %d", len(paths))
	}
}

func TestFileHashCache_GetTrackedPaths(t *testing.T) {
	cache := NewFileHashCache("test-book", "/tmp/workspace", nil)

	cache.ShouldProcessFile("a.md", []byte("a"))
	cache.ShouldProcessFile("b.md", []byte("b"))

	paths := cache.GetTrackedPaths()
	if len(paths) != 2 {
		t.Errorf("Expected 2 tracked paths, got %d", len(paths))
	}
}

func TestFileHashCache_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create and populate cache
	cache := NewFileHashCache("test-book", tmpDir, nil)
	cache.ShouldProcessFile("a.md", []byte("content-a"))
	cache.ShouldProcessFile("b.md", []byte("content-b"))

	// Save
	if err := cache.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(cache.CachePath()); os.IsNotExist(err) {
		t.Fatalf("Cache file not created")
	}

	// Load
	loaded, err := LoadFileHashCache("test-book", tmpDir, nil)
	if err != nil {
		t.Fatalf("LoadFileHashCache failed: %v", err)
	}

	if len(loaded.Hashes) != 2 {
		t.Errorf("Expected 2 hashes, got %d", len(loaded.Hashes))
	}

	// Unchanged file should be cache hit
	if loaded.ShouldProcessFile("a.md", []byte("content-a")) {
		t.Errorf("Expected cache hit for unchanged file after load")
	}
}

func TestLoadFileHashCache_CorruptFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the cache directory and write corrupt data
	cache := NewFileHashCache("test-book", tmpDir, nil)
	cacheDir := filepath.Dir(cache.CachePath())
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}
	if err := os.WriteFile(cache.CachePath(), []byte("not json"), 0o644); err != nil {
		t.Fatalf("Failed to write corrupt file: %v", err)
	}

	// Load should return empty cache (graceful degradation)
	loaded, err := LoadFileHashCache("test-book", tmpDir, nil)
	if err != nil {
		t.Fatalf("LoadFileHashCache should not error on corrupt data: %v", err)
	}

	if len(loaded.Hashes) != 0 {
		t.Errorf("Expected empty hashes from corrupt file, got %d", len(loaded.Hashes))
	}
}

func TestLoadFileHashCache_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Load non-existent cache
	loaded, err := LoadFileHashCache("nonexistent-book", tmpDir, nil)
	if err != nil {
		t.Fatalf("LoadFileHashCache should not error on missing file: %v", err)
	}

	if len(loaded.Hashes) != 0 {
		t.Errorf("Expected empty hashes from missing file, got %d", len(loaded.Hashes))
	}
}

func TestComputeHash(t *testing.T) {
	hash1 := ComputeHash([]byte("hello"))
	hash2 := ComputeHash([]byte("hello"))
	hash3 := ComputeHash([]byte("world"))

	if hash1 != hash2 {
		t.Errorf("Same content should produce same hash")
	}
	if hash1 == hash3 {
		t.Errorf("Different content should produce different hash")
	}
	if len(hash1) != 64 {
		t.Errorf("SHA256 hex should be 64 chars, got %d", len(hash1))
	}
}
