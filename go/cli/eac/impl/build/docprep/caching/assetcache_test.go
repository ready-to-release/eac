package caching

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAssetCache(t *testing.T) {
	tmpDir := t.TempDir()
	ac := NewAssetCache(tmpDir, nil, nil)

	if ac.cacheRoot == "" {
		t.Errorf("Expected non-empty cacheRoot")
	}

	stats := ac.Stats()
	if stats.MermaidHits != 0 || stats.MermaidMisses != 0 {
		t.Errorf("Expected zero stats on new cache")
	}
}

func TestAssetCache_GetMermaid_Miss(t *testing.T) {
	tmpDir := t.TempDir()
	ac := NewAssetCache(tmpDir, nil, nil)

	key := MermaidCacheKey{
		SourceFile: "test.md",
		BlockIndex: 0,
		Code:       "graph TD\n    A --> B",
		Width:      800,
		Height:     600,
		Theme:      "default",
	}

	cachePath, hit := ac.GetMermaid(key)

	if hit {
		t.Errorf("Expected cache miss for new key")
	}
	if cachePath == "" {
		t.Errorf("Expected non-empty cachePath even on miss")
	}

	stats := ac.Stats()
	if stats.MermaidMisses != 1 {
		t.Errorf("Expected 1 mermaid miss, got %d", stats.MermaidMisses)
	}
}

func TestAssetCache_PutAndGetMermaid(t *testing.T) {
	tmpDir := t.TempDir()
	ac := NewAssetCache(tmpDir, nil, nil)

	key := MermaidCacheKey{
		SourceFile: "test.md",
		BlockIndex: 0,
		Code:       "graph TD\n    A --> B",
		Width:      800,
		Height:     600,
		Theme:      "default",
	}

	// Create a temp SVG file to cache
	svgFile := filepath.Join(tmpDir, "temp.svg")
	if err := os.WriteFile(svgFile, []byte("<svg>test</svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write SVG: %v", err)
	}

	// Put into cache
	if err := ac.PutMermaid(svgFile, key); err != nil {
		t.Fatalf("PutMermaid failed: %v", err)
	}

	// Get from cache - should be hit
	cachePath, hit := ac.GetMermaid(key)
	if !hit {
		t.Errorf("Expected cache hit after PutMermaid")
	}

	// Verify cached file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Errorf("Cached file doesn't exist at %s", cachePath)
	}
}

func TestAssetCache_GetDrawio_Miss(t *testing.T) {
	tmpDir := t.TempDir()
	ac := NewAssetCache(tmpDir, nil, nil)

	key := DrawioCacheKey{
		SourcePath: "diagram.drawio.png",
		SourceHash: "abc123",
		MaxWidth:   1200,
	}

	cachePath, hit := ac.GetDrawio(key)

	if hit {
		t.Errorf("Expected cache miss for new key")
	}
	if cachePath == "" {
		t.Errorf("Expected non-empty cachePath even on miss")
	}

	stats := ac.Stats()
	if stats.DrawioMisses != 1 {
		t.Errorf("Expected 1 drawio miss, got %d", stats.DrawioMisses)
	}
}

func TestAssetCache_PutAndGetDrawio(t *testing.T) {
	tmpDir := t.TempDir()
	ac := NewAssetCache(tmpDir, nil, nil)

	key := DrawioCacheKey{
		SourcePath: "diagram.drawio.png",
		SourceHash: "abc123def456",
		MaxWidth:   1200,
	}

	// Create a temp PNG file to cache
	pngFile := filepath.Join(tmpDir, "temp.png")
	if err := os.WriteFile(pngFile, []byte("PNG-data"), 0o644); err != nil {
		t.Fatalf("Failed to write PNG: %v", err)
	}

	// Put into cache
	if err := ac.PutDrawio(pngFile, key); err != nil {
		t.Fatalf("PutDrawio failed: %v", err)
	}

	// Get from cache - should be hit
	cachePath, hit := ac.GetDrawio(key)
	if !hit {
		t.Errorf("Expected cache hit after PutDrawio")
	}

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Errorf("Cached file doesn't exist at %s", cachePath)
	}
}

func TestHashMermaidKey_Deterministic(t *testing.T) {
	key := MermaidCacheKey{
		SourceFile: "test.md",
		BlockIndex: 0,
		Code:       "graph TD\n    A --> B",
		Width:      800,
		Height:     600,
		Theme:      "default",
	}

	hash1 := HashMermaidKey(key)
	hash2 := HashMermaidKey(key)

	if hash1 != hash2 {
		t.Errorf("Hash should be deterministic: %s != %s", hash1, hash2)
	}
}

func TestHashMermaidKey_DifferentInputs(t *testing.T) {
	key1 := MermaidCacheKey{Code: "graph TD\n    A --> B", Theme: "default"}
	key2 := MermaidCacheKey{Code: "graph TD\n    A --> C", Theme: "default"}
	key3 := MermaidCacheKey{Code: "graph TD\n    A --> B", Theme: "dark"}

	hash1 := HashMermaidKey(key1)
	hash2 := HashMermaidKey(key2)
	hash3 := HashMermaidKey(key3)

	if hash1 == hash2 {
		t.Errorf("Different code should produce different hash")
	}
	if hash1 == hash3 {
		t.Errorf("Different theme should produce different hash")
	}
}

func TestHashDrawioKey_Deterministic(t *testing.T) {
	key := DrawioCacheKey{SourceHash: "abc123", MaxWidth: 1200}

	hash1 := HashDrawioKey(key)
	hash2 := HashDrawioKey(key)

	if hash1 != hash2 {
		t.Errorf("Hash should be deterministic: %s != %s", hash1, hash2)
	}
}

func TestHashDrawioKey_DifferentInputs(t *testing.T) {
	key1 := DrawioCacheKey{SourceHash: "abc123", MaxWidth: 1200}
	key2 := DrawioCacheKey{SourceHash: "def456", MaxWidth: 1200}
	key3 := DrawioCacheKey{SourceHash: "abc123", MaxWidth: 800}

	hash1 := HashDrawioKey(key1)
	hash2 := HashDrawioKey(key2)
	hash3 := HashDrawioKey(key3)

	if hash1 == hash2 {
		t.Errorf("Different source hash should produce different key")
	}
	if hash1 == hash3 {
		t.Errorf("Different max width should produce different key")
	}
}

func TestAssetCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	ac := NewAssetCache(tmpDir, nil, nil)

	// Create a file in the cache root
	cacheDir := ac.cacheRoot
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "test.svg"), []byte("<svg/>"), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Clear
	if err := ac.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify directory is removed
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Errorf("Cache directory should be removed after Clear")
	}
}
