package hash

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestLoadCache_NonExistent(t *testing.T) {
	c := LoadCache(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if c == nil {
		t.Fatal("expected non-nil cache")
	}
	if c.Version != cacheVersion {
		t.Errorf("expected version %d, got %d", cacheVersion, c.Version)
	}
	if len(c.Modules) != 0 {
		t.Errorf("expected empty modules, got %d", len(c.Modules))
	}
}

func TestLoadCache_CorruptJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")
	if err := os.WriteFile(cachePath, []byte("{corrupt"), 0644); err != nil {
		t.Fatal(err)
	}

	c := LoadCache(cachePath)
	if c == nil {
		t.Fatal("expected non-nil cache")
	}
	if len(c.Modules) != 0 {
		t.Errorf("expected empty modules for corrupt cache, got %d", len(c.Modules))
	}
}

func TestLoadCache_WrongVersion(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")
	data, _ := json.Marshal(map[string]interface{}{
		"version": 999,
		"modules": map[string]interface{}{
			"core": map[string]interface{}{
				"hash": "abc123",
			},
		},
	})
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		t.Fatal(err)
	}

	c := LoadCache(cachePath)
	if len(c.Modules) != 0 {
		t.Errorf("expected empty modules for wrong version, got %d", len(c.Modules))
	}
}

func TestLoadCache_ValidCache(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")
	data, _ := json.Marshal(&HashCache{
		Version: cacheVersion,
		Modules: map[string]*ModuleCache{
			"core": {
				Patterns: []string{"**/*.go"},
				Hash:     "abc123",
				Files:    map[string]int64{"main.go": 12345},
			},
		},
	})
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		t.Fatal(err)
	}

	c := LoadCache(cachePath)
	if len(c.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(c.Modules))
	}
	if c.Modules["core"].Hash != "abc123" {
		t.Errorf("expected hash abc123, got %s", c.Modules["core"].Hash)
	}
}

func TestSave_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")

	c := &HashCache{
		Version: cacheVersion,
		Modules: map[string]*ModuleCache{
			"mymod": {
				Patterns: []string{"src/**/*.go"},
				Hash:     "deadbeef",
				Files:    map[string]int64{"src/main.go": 99999},
			},
		},
		path:  cachePath,
		dirty: true,
	}

	if err := c.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded := LoadCache(cachePath)
	if loaded.Modules["mymod"].Hash != "deadbeef" {
		t.Errorf("expected hash deadbeef, got %s", loaded.Modules["mymod"].Hash)
	}
	if !slices.Equal(loaded.Modules["mymod"].Patterns, []string{"src/**/*.go"}) {
		t.Errorf("patterns mismatch")
	}
}

func TestSave_CreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "deep", "nested", "cache.json")

	c := &HashCache{
		Version: cacheVersion,
		Modules: map[string]*ModuleCache{},
		path:    cachePath,
		dirty:   true,
	}

	if err := c.Save(); err != nil {
		t.Fatalf("save should create directories: %v", err)
	}

	if _, err := os.Stat(cachePath); err != nil {
		t.Errorf("cache file not created: %v", err)
	}
}

func TestSave_SkipsWhenNotDirty(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")

	c := &HashCache{
		Version: cacheVersion,
		Modules: map[string]*ModuleCache{},
		path:    cachePath,
		dirty:   false,
	}

	if err := c.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// File should NOT exist since not dirty
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Error("expected cache file to not be written when not dirty")
	}
}

func TestGetOrCompute_FirstRun(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	cachePath := filepath.Join(tmpDir, ".cache", "hashes.json")
	c := LoadCache(cachePath)

	h, files, err := c.GetOrCompute("mymod", []string{"*.go"}, tmpDir)
	if err != nil {
		t.Fatalf("GetOrCompute failed: %v", err)
	}
	if h == "" {
		t.Error("expected non-empty hash")
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
	if !c.dirty {
		t.Error("expected cache to be dirty after first run")
	}
}

func TestGetOrCompute_CacheHit(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	cachePath := filepath.Join(tmpDir, ".cache", "hashes.json")
	c := LoadCache(cachePath)

	// First call: cache miss
	h1, files1, err := c.GetOrCompute("mymod", []string{"*.go"}, tmpDir)
	if err != nil {
		t.Fatalf("first GetOrCompute failed: %v", err)
	}

	// Reset dirty flag to verify cache hit doesn't set it again
	c.dirty = false

	// Second call: cache hit (same patterns, same mtimes)
	h2, files2, err := c.GetOrCompute("mymod", []string{"*.go"}, tmpDir)
	if err != nil {
		t.Fatalf("second GetOrCompute failed: %v", err)
	}

	if h1 != h2 {
		t.Errorf("expected same hash on cache hit: %s != %s", h1, h2)
	}
	if len(files1) != len(files2) {
		t.Errorf("expected same files on cache hit: %d != %d", len(files1), len(files2))
	}
	if c.dirty {
		t.Error("expected cache to not be dirty on cache hit")
	}
}

func TestGetOrCompute_MtimeChanged(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(filePath, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	cachePath := filepath.Join(tmpDir, ".cache", "hashes.json")
	c := LoadCache(cachePath)

	// First call
	h1, _, err := c.GetOrCompute("mymod", []string{"*.go"}, tmpDir)
	if err != nil {
		t.Fatalf("first GetOrCompute failed: %v", err)
	}

	// Change mtime (without changing content)
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(filePath, future, future); err != nil {
		t.Fatal(err)
	}

	c.dirty = false

	// Should recompute (mtime changed)
	h2, _, err := c.GetOrCompute("mymod", []string{"*.go"}, tmpDir)
	if err != nil {
		t.Fatalf("second GetOrCompute failed: %v", err)
	}

	// Same content -> same hash, but dirty should be true since mtimes differ
	if h1 != h2 {
		t.Errorf("expected same hash for same content: %s != %s", h1, h2)
	}
	if !c.dirty {
		t.Error("expected cache to be dirty after mtime change")
	}
}

func TestGetOrCompute_FileDeleted(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(filePath, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	cachePath := filepath.Join(tmpDir, ".cache", "hashes.json")
	c := LoadCache(cachePath)

	// First call
	_, _, err := c.GetOrCompute("mymod", []string{"*.go"}, tmpDir)
	if err != nil {
		t.Fatalf("first GetOrCompute failed: %v", err)
	}

	// Delete the file
	os.Remove(filePath)
	c.dirty = false

	// Should recompute (file missing -> mtime check fails)
	h2, files2, err := c.GetOrCompute("mymod", []string{"*.go"}, tmpDir)
	if err != nil {
		t.Fatalf("second GetOrCompute failed: %v", err)
	}

	// No files match, so hash should be empty
	if h2 != "" {
		t.Errorf("expected empty hash when no files, got %s", h2)
	}
	if len(files2) != 0 {
		t.Errorf("expected 0 files, got %d", len(files2))
	}
	if !c.dirty {
		t.Error("expected cache to be dirty after file deletion")
	}
}

func TestGetOrCompute_PatternChange(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# readme"), 0644); err != nil {
		t.Fatal(err)
	}

	cachePath := filepath.Join(tmpDir, ".cache", "hashes.json")
	c := LoadCache(cachePath)

	// First call with *.go
	_, _, err := c.GetOrCompute("mymod", []string{"*.go"}, tmpDir)
	if err != nil {
		t.Fatalf("first GetOrCompute failed: %v", err)
	}

	c.dirty = false

	// Second call with different patterns
	_, files2, err := c.GetOrCompute("mymod", []string{"*.go", "*.md"}, tmpDir)
	if err != nil {
		t.Fatalf("second GetOrCompute failed: %v", err)
	}

	if len(files2) != 2 {
		t.Errorf("expected 2 files with new patterns, got %d", len(files2))
	}
	if !c.dirty {
		t.Error("expected cache to be dirty after pattern change")
	}
}

func TestGetOrCompute_DeterministicHash(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.go"), []byte("package a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.go"), []byte("package b"), 0644); err != nil {
		t.Fatal(err)
	}

	cachePath1 := filepath.Join(tmpDir, ".cache", "h1.json")
	cachePath2 := filepath.Join(tmpDir, ".cache", "h2.json")
	c1 := LoadCache(cachePath1)
	c2 := LoadCache(cachePath2)

	h1, _, err := c1.GetOrCompute("mymod", []string{"*.go"}, tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	h2, _, err := c2.GetOrCompute("mymod", []string{"*.go"}, tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if h1 != h2 {
		t.Errorf("expected deterministic hash: %s != %s", h1, h2)
	}
}

func TestGetOrCompute_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"a.go", "b.go", "c.go"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("package main"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cachePath := filepath.Join(tmpDir, ".cache", "hashes.json")
	c := LoadCache(cachePath)

	var wg sync.WaitGroup
	errs := make(chan error, 10)

	// Run concurrent GetOrCompute calls for different modules
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			module := "mod" + string(rune('a'+idx))
			_, _, err := c.GetOrCompute(module, []string{"*.go"}, tmpDir)
			if err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent GetOrCompute failed: %v", err)
	}

	// All modules should be cached
	if len(c.Modules) != 10 {
		t.Errorf("expected 10 cached modules, got %d", len(c.Modules))
	}
}
