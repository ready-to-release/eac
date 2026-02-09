// Package docs provides cache pruning tests for the update docs command.
// These tests verify that the cache pruning logic correctly identifies orphaned
// cache files by computing hashes that exactly match the caching hash functions.
package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/diagrams"
	"github.com/ready-to-release/eac/go/core/paths"
)

// =============================================================================
// Hash Computation Tests - Verify exact match with caching hash functions
// =============================================================================

// TestComputeMermaidCacheHash_ProducesValidHash verifies that the prune module's
// mermaid hash computation produces a valid SHA256 hash.
// The content hash is used as part of the traceable filename.
func TestComputeMermaidCacheHash_ProducesValidHash(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "simple flowchart",
			code: "flowchart TD\n    A --> B",
		},
		{
			name: "complex sequence diagram",
			code: `sequenceDiagram
    participant A as Alice
    participant B as Bob
    A->>B: Hello Bob
    B-->>A: Hi Alice`,
		},
		{
			name: "empty diagram",
			code: "",
		},
		{
			name: "diagram with special characters",
			code: "flowchart TD\n    A[\"Label with 'quotes'\"] --> B[\"Unicode: \u00e9\u00e8\u00e0\"]",
		},
		{
			name: "diagram with newlines",
			code: "flowchart TD\n\n    A --> B\n\n    B --> C\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := computeMermaidCacheHash(tt.code)

			// Verify hash is 64 hex characters (SHA256)
			if len(hash) != 64 {
				t.Errorf("Expected 64-char hash, got %d chars: %s", len(hash), hash)
			}

			// Verify hash is valid hex
			for _, c := range hash {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("Invalid hex character in hash: %c", c)
				}
			}
		})
	}
}

// TestComputeMermaidCacheHash_WithSizeDirective verifies that size directives
// are properly stripped before hashing, matching the hash function behavior.
func TestComputeMermaidCacheHash_WithSizeDirective(t *testing.T) {
	// Content WITH size directive
	contentWithDirective := `%%{size:medium}%%
flowchart TD
    A --> B`

	// Same content WITHOUT size directive (should produce same hash)
	contentWithout := `flowchart TD
    A --> B`

	// Strip the directive (as the prune logic should do)
	strippedContent := diagrams.StripSizeDirective(contentWithDirective)

	// Hash the stripped content
	hashFromDirective := computeMermaidCacheHash(strippedContent)
	hashFromPlain := computeMermaidCacheHash(contentWithout)

	if hashFromDirective != hashFromPlain {
		t.Errorf("Hash mismatch after stripping size directive:\n  with directive (stripped): %s\n  without directive: %s",
			hashFromDirective, hashFromPlain)
	}
}

// TestComputeDrawioCacheHash_ProducesValidHash verifies that the prune module's
// drawio hash computation produces a valid SHA256 hash.
// The content hash is used as part of the traceable filename.
func TestComputeDrawioCacheHash_ProducesValidHash(t *testing.T) {
	tests := []struct {
		name       string
		sourceHash string
		maxWidth   int
	}{
		{
			name:       "standard hash and width",
			sourceHash: "abc123def456789",
			maxWidth:   diagrams.MaxImageWidthPDF,
		},
		{
			name:       "full 64-char hash",
			sourceHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			maxWidth:   diagrams.MaxImageWidthPDF,
		},
		{
			name:       "empty source hash",
			sourceHash: "",
			maxWidth:   diagrams.MaxImageWidthPDF,
		},
		{
			name:       "different max width",
			sourceHash: "abc123",
			maxWidth:   800,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := computeDrawioCacheHash(tt.sourceHash, tt.maxWidth)

			// Verify hash is 64 hex characters (SHA256)
			if len(hash) != 64 {
				t.Errorf("Expected 64-char hash, got %d chars: %s", len(hash), hash)
			}

			// Verify hash is valid hex
			for _, c := range hash {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("Invalid hex character in hash: %c", c)
				}
			}
		})
	}
}

// =============================================================================
// PruneCache Tests - Verify orphan identification
// =============================================================================

// TestPruneCache_IdentifiesMermaidOrphans verifies that PruneCache correctly
// identifies mermaid cache files that are no longer referenced by markdown files.
func TestPruneCache_IdentifiesMermaidOrphans(t *testing.T) {
	// Setup temp directory structure
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	cacheDir := filepath.Join(tmpDir, ".cache", "eac")
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")

	// Create directories
	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create mermaid cache dir: %v", err)
	}

	// Create markdown with one mermaid block
	mdPath := filepath.Join(docsDir, "test.md")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}
	mdContent := `# Test Document

Here is a diagram:

` + "```mermaid\nflowchart TD\n    A --> B\n```" + `

End of document.
`

	if err := os.WriteFile(mdPath, []byte(mdContent), 0o644); err != nil {
		t.Fatalf("Failed to write markdown: %v", err)
	}

	// Compute the valid filename for the mermaid block (traceable format)
	validHash := computeMermaidCacheHash("flowchart TD\n    A --> B")
	validFilename := filepath.Base(paths.MermaidCachePath(cacheDir, mdPath, 0, validHash))

	// Create cache files: one valid, one orphan
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, validFilename), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write valid cache file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, "orphan123abc.svg"), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write orphan cache file: %v", err)
	}

	// Run prune
	result, err := PruneCache(tmpDir, false)
	if err != nil {
		t.Fatalf("PruneCache failed: %v", err)
	}

	// Verify results
	if result.MermaidActiveHashes != 1 {
		t.Errorf("Expected 1 active mermaid hash, got %d", result.MermaidActiveHashes)
	}

	if len(result.MermaidOrphans) != 1 {
		t.Errorf("Expected 1 mermaid orphan, got %d: %v", len(result.MermaidOrphans), result.MermaidOrphans)
	} else if result.MermaidOrphans[0] != "orphan123abc.svg" {
		t.Errorf("Expected orphan123abc.svg, got %s", result.MermaidOrphans[0])
	}

	// Valid file should not be an orphan
	for _, orphan := range result.MermaidOrphans {
		if orphan == validFilename {
			t.Errorf("Valid cache file %s incorrectly identified as orphan", validFilename)
		}
	}
}

// TestPruneCache_IdentifiesDrawioOrphans verifies that PruneCache correctly
// identifies drawio cache files that are no longer referenced.
func TestPruneCache_IdentifiesDrawioOrphans(t *testing.T) {
	// Setup temp directory structure
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	assetsDir := filepath.Join(docsDir, "assets")
	cacheDir := filepath.Join(tmpDir, ".cache", "eac")
	drawioCacheDir := filepath.Join(cacheDir, "drawio")

	// Create directories
	if err := os.MkdirAll(drawioCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create drawio cache dir: %v", err)
	}
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("Failed to create assets dir: %v", err)
	}

	// Create a drawio.png source file with known content
	drawioContent := []byte("PNG\x89\x50\x4e\x47test-drawio-content")
	drawioPath := filepath.Join(assetsDir, "diagram.drawio.png")
	if err := os.WriteFile(drawioPath, drawioContent, 0o644); err != nil {
		t.Fatalf("Failed to write drawio source: %v", err)
	}

	// Compute the source file hash
	sourceHash, err := diagrams.HashFileContent(drawioPath)
	if err != nil {
		t.Fatalf("Failed to hash drawio source: %v", err)
	}

	// Compute the valid cache filename (traceable format)
	cacheHash := computeDrawioCacheHash(sourceHash, diagrams.MaxImageWidthPDF)
	validFilename := filepath.Base(paths.DrawioCachePath(cacheDir, drawioPath, cacheHash))

	// Create cache files: one valid, one orphan
	if err := os.WriteFile(filepath.Join(drawioCacheDir, validFilename), []byte("optimized"), 0o644); err != nil {
		t.Fatalf("Failed to write valid cache file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(drawioCacheDir, "orphan456def.png"), []byte("orphan"), 0o644); err != nil {
		t.Fatalf("Failed to write orphan cache file: %v", err)
	}

	// Run prune
	result, err := PruneCache(tmpDir, false)
	if err != nil {
		t.Fatalf("PruneCache failed: %v", err)
	}

	// Verify results
	if result.DrawioActiveHashes != 1 {
		t.Errorf("Expected 1 active drawio hash, got %d", result.DrawioActiveHashes)
	}

	if len(result.DrawioOrphans) != 1 {
		t.Errorf("Expected 1 drawio orphan, got %d: %v", len(result.DrawioOrphans), result.DrawioOrphans)
	} else if result.DrawioOrphans[0] != "orphan456def.png" {
		t.Errorf("Expected orphan456def.png, got %s", result.DrawioOrphans[0])
	}

	// Valid file should not be an orphan
	for _, orphan := range result.DrawioOrphans {
		if orphan == validFilename {
			t.Errorf("Valid cache file %s incorrectly identified as orphan", validFilename)
		}
	}
}

// TestPruneCache_MultipleMermaidBlocks verifies correct handling of files with
// multiple mermaid blocks.
func TestPruneCache_MultipleMermaidBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	cacheDir := filepath.Join(tmpDir, ".cache", "eac")
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")

	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	// Create markdown with multiple mermaid blocks
	mdPath := filepath.Join(docsDir, "multi.md")
	mdContent := `# Multiple Diagrams

` + "```mermaid\nflowchart TD\n    A --> B\n```" + `

Some text.

` + "```mermaid\nsequenceDiagram\n    A->>B: Hello\n```" + `

More text.

` + "```mermaid\npie\n    title Pets\n    \"Dogs\": 45\n    \"Cats\": 30\n```" + `
`

	if err := os.WriteFile(mdPath, []byte(mdContent), 0o644); err != nil {
		t.Fatalf("Failed to write markdown: %v", err)
	}

	// Compute valid filenames for all three blocks (traceable format)
	hash1 := computeMermaidCacheHash("flowchart TD\n    A --> B")
	hash2 := computeMermaidCacheHash("sequenceDiagram\n    A->>B: Hello")
	hash3 := computeMermaidCacheHash("pie\n    title Pets\n    \"Dogs\": 45\n    \"Cats\": 30")

	filename1 := filepath.Base(paths.MermaidCachePath(cacheDir, mdPath, 0, hash1))
	filename2 := filepath.Base(paths.MermaidCachePath(cacheDir, mdPath, 1, hash2))
	filename3 := filepath.Base(paths.MermaidCachePath(cacheDir, mdPath, 2, hash3))

	// Create all valid cache files plus one orphan
	for _, fn := range []string{filename1, filename2, filename3} {
		if err := os.WriteFile(filepath.Join(mermaidCacheDir, fn), []byte("<svg></svg>"), 0o644); err != nil {
			t.Fatalf("Failed to write cache file: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, "orphanxyz.svg"), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write orphan: %v", err)
	}

	// Run prune
	result, err := PruneCache(tmpDir, false)
	if err != nil {
		t.Fatalf("PruneCache failed: %v", err)
	}

	if result.MermaidActiveHashes != 3 {
		t.Errorf("Expected 3 active mermaid hashes, got %d", result.MermaidActiveHashes)
	}

	if len(result.MermaidOrphans) != 1 {
		t.Errorf("Expected 1 orphan, got %d: %v", len(result.MermaidOrphans), result.MermaidOrphans)
	}
}

// TestPruneCache_EmptyCache verifies behavior when cache directories are empty.
func TestPruneCache_EmptyCache(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	mermaidCacheDir := filepath.Join(tmpDir, ".cache", "eac", "mermaid")
	drawioCacheDir := filepath.Join(tmpDir, ".cache", "eac", "drawio")

	// Create directories but leave them empty
	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create mermaid cache dir: %v", err)
	}
	if err := os.MkdirAll(drawioCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create drawio cache dir: %v", err)
	}
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	// Create a markdown file with mermaid
	mdContent := "# Test\n\n```mermaid\nflowchart TD\n    A\n```\n"
	if err := os.WriteFile(filepath.Join(docsDir, "test.md"), []byte(mdContent), 0o644); err != nil {
		t.Fatalf("Failed to write markdown: %v", err)
	}

	// Run prune
	result, err := PruneCache(tmpDir, false)
	if err != nil {
		t.Fatalf("PruneCache failed: %v", err)
	}

	// Empty cache = no orphans, even though we have content
	if len(result.MermaidOrphans) != 0 {
		t.Errorf("Expected 0 mermaid orphans with empty cache, got %d", len(result.MermaidOrphans))
	}
	if len(result.DrawioOrphans) != 0 {
		t.Errorf("Expected 0 drawio orphans with empty cache, got %d", len(result.DrawioOrphans))
	}
}

// TestPruneCache_NoMarkdownFiles verifies that all cache files become orphans
// when there are no markdown files to reference them.
func TestPruneCache_NoMarkdownFiles(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	mermaidCacheDir := filepath.Join(tmpDir, ".cache", "eac", "mermaid")

	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	// Create cache files but no markdown
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, "cache1.svg"), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write cache1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, "cache2.svg"), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write cache2: %v", err)
	}

	// Run prune
	result, err := PruneCache(tmpDir, false)
	if err != nil {
		t.Fatalf("PruneCache failed: %v", err)
	}

	// All cache files should be orphans
	if len(result.MermaidOrphans) != 2 {
		t.Errorf("Expected 2 mermaid orphans (all files), got %d", len(result.MermaidOrphans))
	}
	if result.MermaidActiveHashes != 0 {
		t.Errorf("Expected 0 active hashes with no markdown, got %d", result.MermaidActiveHashes)
	}
}

// TestPruneCache_NoCacheDirectories verifies graceful handling when cache
// directories don't exist.
func TestPruneCache_NoCacheDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")

	// Only create docs dir, no cache subdirectories
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	// Create a markdown file
	mdContent := "# Test\n\n```mermaid\nflowchart TD\n    A\n```\n"
	if err := os.WriteFile(filepath.Join(docsDir, "test.md"), []byte(mdContent), 0o644); err != nil {
		t.Fatalf("Failed to write markdown: %v", err)
	}

	// Run prune - should not error
	result, err := PruneCache(tmpDir, false)
	if err != nil {
		t.Fatalf("PruneCache failed unexpectedly: %v", err)
	}

	// Should have active hashes but no orphans (no cache to have orphans in)
	if result.MermaidActiveHashes != 1 {
		t.Errorf("Expected 1 active hash, got %d", result.MermaidActiveHashes)
	}
	if len(result.MermaidOrphans) != 0 {
		t.Errorf("Expected 0 orphans with no cache dir, got %d", len(result.MermaidOrphans))
	}
}

// TestPruneCache_IgnoresNonCacheFiles verifies that files with wrong extensions
// or names are not counted as cache files.
func TestPruneCache_IgnoresNonCacheFiles(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	mermaidCacheDir := filepath.Join(tmpDir, ".cache", "eac", "mermaid")

	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	// Create various non-SVG files that should be ignored
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, "readme.txt"), []byte("info"), 0o644); err != nil {
		t.Fatalf("Failed to write txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, "diagram.png"), []byte("png"), 0o644); err != nil {
		t.Fatalf("Failed to write png: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, ".gitkeep"), []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to write gitkeep: %v", err)
	}
	// One actual SVG that should be counted
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, "actualcache.svg"), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write svg: %v", err)
	}

	// Run prune (no markdown, so all SVGs are orphans)
	result, err := PruneCache(tmpDir, false)
	if err != nil {
		t.Fatalf("PruneCache failed: %v", err)
	}

	// Only the .svg file should be counted as an orphan
	if len(result.MermaidOrphans) != 1 {
		t.Errorf("Expected 1 mermaid orphan (only .svg), got %d: %v",
			len(result.MermaidOrphans), result.MermaidOrphans)
	}
	if len(result.MermaidOrphans) == 1 && result.MermaidOrphans[0] != "actualcache.svg" {
		t.Errorf("Expected actualcache.svg, got %s", result.MermaidOrphans[0])
	}
}

// TestPruneCache_BytesRecovered verifies that byte counts are calculated correctly.
func TestPruneCache_BytesRecovered(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	mermaidCacheDir := filepath.Join(tmpDir, ".cache", "eac", "mermaid")

	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	// Create orphan files with known sizes
	content1 := strings.Repeat("a", 1000) // 1000 bytes
	content2 := strings.Repeat("b", 2500) // 2500 bytes
	expectedBytes := int64(3500)

	if err := os.WriteFile(filepath.Join(mermaidCacheDir, "orphan1.svg"), []byte(content1), 0o644); err != nil {
		t.Fatalf("Failed to write orphan1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, "orphan2.svg"), []byte(content2), 0o644); err != nil {
		t.Fatalf("Failed to write orphan2: %v", err)
	}

	// Run prune
	result, err := PruneCache(tmpDir, false)
	if err != nil {
		t.Fatalf("PruneCache failed: %v", err)
	}

	if result.MermaidBytesRecovered != expectedBytes {
		t.Errorf("Expected %d bytes recovered, got %d", expectedBytes, result.MermaidBytesRecovered)
	}
}

// =============================================================================
// DeleteOrphans Tests
// =============================================================================

// TestDeleteOrphans_RemovesFiles verifies that DeleteOrphans actually deletes
// the orphaned files.
func TestDeleteOrphans_RemovesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, ".cache", "eac")
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")
	drawioCacheDir := filepath.Join(cacheDir, "drawio")

	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create mermaid dir: %v", err)
	}
	if err := os.MkdirAll(drawioCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create drawio dir: %v", err)
	}

	// Create orphan files
	mermaidOrphan := filepath.Join(mermaidCacheDir, "orphan1.svg")
	drawioOrphan := filepath.Join(drawioCacheDir, "orphan2.png")

	if err := os.WriteFile(mermaidOrphan, []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write mermaid orphan: %v", err)
	}
	if err := os.WriteFile(drawioOrphan, []byte("png"), 0o644); err != nil {
		t.Fatalf("Failed to write drawio orphan: %v", err)
	}

	// Create result with known orphans
	result := &PruneResult{
		MermaidOrphans: []string{"orphan1.svg"},
		DrawioOrphans:  []string{"orphan2.png"},
	}

	// Delete orphans
	deleted, err := DeleteOrphans(result, cacheDir)
	if err != nil {
		t.Fatalf("DeleteOrphans failed: %v", err)
	}

	if deleted != 2 {
		t.Errorf("Expected 2 files deleted, got %d", deleted)
	}

	// Verify files are gone
	if _, err := os.Stat(mermaidOrphan); !os.IsNotExist(err) {
		t.Errorf("Mermaid orphan should have been deleted")
	}
	if _, err := os.Stat(drawioOrphan); !os.IsNotExist(err) {
		t.Errorf("Drawio orphan should have been deleted")
	}
}

// TestDeleteOrphans_EmptyResult verifies behavior with no orphans.
func TestDeleteOrphans_EmptyResult(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	result := &PruneResult{
		MermaidOrphans: []string{},
		DrawioOrphans:  []string{},
	}

	deleted, err := DeleteOrphans(result, cacheDir)
	if err != nil {
		t.Fatalf("DeleteOrphans failed: %v", err)
	}

	if deleted != 0 {
		t.Errorf("Expected 0 files deleted with empty result, got %d", deleted)
	}
}

// TestDeleteOrphans_ContinuesOnMissingFiles verifies that DeleteOrphans continues
// even if some files are already gone (e.g., deleted manually).
func TestDeleteOrphans_ContinuesOnMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, ".cache", "eac")
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")

	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create mermaid dir: %v", err)
	}

	// Only create one of the two orphans
	existingOrphan := filepath.Join(mermaidCacheDir, "exists.svg")
	if err := os.WriteFile(existingOrphan, []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write orphan: %v", err)
	}

	// Result includes a file that doesn't exist
	result := &PruneResult{
		MermaidOrphans: []string{"exists.svg", "missing.svg"},
	}

	// Should not error, should delete what it can
	deleted, err := DeleteOrphans(result, cacheDir)
	if err != nil {
		t.Fatalf("DeleteOrphans failed: %v", err)
	}

	// Should have deleted at least the existing file
	if deleted < 1 {
		t.Errorf("Expected at least 1 file deleted, got %d", deleted)
	}

	// Existing orphan should be gone
	if _, err := os.Stat(existingOrphan); !os.IsNotExist(err) {
		t.Errorf("Existing orphan should have been deleted")
	}
}

// =============================================================================
// PruneResult Helper Tests
// =============================================================================

// TestPruneResult_TotalOrphans verifies the TotalOrphans helper method.
func TestPruneResult_TotalOrphans(t *testing.T) {
	result := &PruneResult{
		MermaidOrphans: []string{"a.svg", "b.svg"},
		DrawioOrphans:  []string{"c.png"},
	}

	if result.TotalOrphans() != 3 {
		t.Errorf("Expected 3 total orphans, got %d", result.TotalOrphans())
	}
}

// TestPruneResult_TotalBytesRecovered verifies the TotalBytesRecovered helper.
func TestPruneResult_TotalBytesRecovered(t *testing.T) {
	result := &PruneResult{
		MermaidBytesRecovered: 1000,
		DrawioBytesRecovered:  2500,
	}

	if result.TotalBytesRecovered() != 3500 {
		t.Errorf("Expected 3500 total bytes, got %d", result.TotalBytesRecovered())
	}
}

// =============================================================================
// formatBytes Tests
// =============================================================================

// TestFormatBytes verifies the formatBytes helper for various sizes.
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{10240, "10.0 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{10485760, "10.0 MB"},
		{104857600, "100.0 MB"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d_bytes", tt.bytes), func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Edge Cases and Integration Tests
// =============================================================================

// TestPruneCache_SameContentDifferentFiles verifies that the same mermaid block
// appearing in multiple files produces different cache entries (traceable naming).
func TestPruneCache_SameContentDifferentFiles(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	cacheDir := filepath.Join(tmpDir, ".cache", "eac")
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")

	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	// Create two markdown files with the exact same mermaid block
	mdContent := "# Doc\n\n```mermaid\nflowchart TD\n    SHARED\n```\n"

	doc1Path := filepath.Join(docsDir, "doc1.md")
	doc2Path := filepath.Join(docsDir, "doc2.md")
	if err := os.WriteFile(doc1Path, []byte(mdContent), 0o644); err != nil {
		t.Fatalf("Failed to write doc1: %v", err)
	}
	if err := os.WriteFile(doc2Path, []byte(mdContent), 0o644); err != nil {
		t.Fatalf("Failed to write doc2: %v", err)
	}

	// With traceable naming, each file gets its own cache entry
	contentHash := computeMermaidCacheHash("flowchart TD\n    SHARED")

	cache1Path := paths.MermaidCachePath(cacheDir, doc1Path, 0, contentHash)
	cache2Path := paths.MermaidCachePath(cacheDir, doc2Path, 0, contentHash)
	cache1Filename := filepath.Base(cache1Path)
	cache2Filename := filepath.Base(cache2Path)

	// Create both valid cache files
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, cache1Filename), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write cache1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, cache2Filename), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write cache2: %v", err)
	}

	// Run prune
	result, err := PruneCache(tmpDir, false)
	if err != nil {
		t.Fatalf("PruneCache failed: %v", err)
	}

	// With traceable naming, same content from different files = 2 active hashes
	if result.MermaidActiveHashes != 2 {
		t.Errorf("Expected 2 active hashes (one per file), got %d", result.MermaidActiveHashes)
	}

	// No orphans since we created both valid files
	if len(result.MermaidOrphans) != 0 {
		t.Errorf("Expected 0 orphans, got %d", len(result.MermaidOrphans))
	}
}

// TestPruneCache_NestedDocsDirectories verifies correct handling of nested
// directory structures within docs/.
func TestPruneCache_NestedDocsDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	cacheDir := filepath.Join(tmpDir, ".cache", "eac")
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")

	// Create nested structure
	nestedDir := filepath.Join(docsDir, "guide", "advanced", "diagrams")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}
	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Create markdown in nested directory
	mdPath := filepath.Join(nestedDir, "deep.md")
	mdContent := "# Nested\n\n```mermaid\nflowchart TD\n    NESTED\n```\n"
	if err := os.WriteFile(mdPath, []byte(mdContent), 0o644); err != nil {
		t.Fatalf("Failed to write nested md: %v", err)
	}

	// Compute hash and create cache using traceable naming
	nestedHash := computeMermaidCacheHash("flowchart TD\n    NESTED")
	cachePath := paths.MermaidCachePath(cacheDir, mdPath, 0, nestedHash)
	cacheFilename := filepath.Base(cachePath)
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, cacheFilename), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write cache: %v", err)
	}

	// Also create an orphan
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, "orphan.svg"), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write orphan: %v", err)
	}

	// Run prune
	result, err := PruneCache(tmpDir, false)
	if err != nil {
		t.Fatalf("PruneCache failed: %v", err)
	}

	// Should find the nested diagram
	if result.MermaidActiveHashes != 1 {
		t.Errorf("Expected 1 active hash from nested dir, got %d", result.MermaidActiveHashes)
	}

	// Should identify the orphan
	if len(result.MermaidOrphans) != 1 {
		t.Errorf("Expected 1 orphan, got %d", len(result.MermaidOrphans))
	}
}

// TestPruneCache_MixedContentTypes verifies correct handling when both mermaid
// and drawio content exist together.
func TestPruneCache_MixedContentTypes(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	assetsDir := filepath.Join(docsDir, "assets")
	cacheDir := filepath.Join(tmpDir, ".cache", "eac")
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")
	drawioCacheDir := filepath.Join(cacheDir, "drawio")

	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create mermaid cache: %v", err)
	}
	if err := os.MkdirAll(drawioCacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create drawio cache: %v", err)
	}
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("Failed to create assets dir: %v", err)
	}

	// Create markdown with mermaid
	mdPath := filepath.Join(docsDir, "mixed.md")
	mdContent := "# Mixed\n\n```mermaid\nflowchart TD\n    MIXED\n```\n"
	if err := os.WriteFile(mdPath, []byte(mdContent), 0o644); err != nil {
		t.Fatalf("Failed to write md: %v", err)
	}

	// Create drawio source
	drawioContent := []byte("DRAWIO-SOURCE-CONTENT")
	drawioPath := filepath.Join(assetsDir, "arch.drawio.png")
	if err := os.WriteFile(drawioPath, drawioContent, 0o644); err != nil {
		t.Fatalf("Failed to write drawio: %v", err)
	}

	// Compute valid hashes and create cache files using traceable naming
	mermaidHash := computeMermaidCacheHash("flowchart TD\n    MIXED")
	mermaidCachePath := paths.MermaidCachePath(cacheDir, mdPath, 0, mermaidHash)
	mermaidCacheFilename := filepath.Base(mermaidCachePath)

	drawioSourceHash, _ := diagrams.HashFileContent(drawioPath)
	drawioHash := computeDrawioCacheHash(drawioSourceHash, diagrams.MaxImageWidthPDF)
	drawioCachePath := paths.DrawioCachePath(cacheDir, drawioPath, drawioHash)
	drawioCacheFilename := filepath.Base(drawioCachePath)

	// Create valid cache files with traceable names
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, mermaidCacheFilename), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write mermaid cache: %v", err)
	}
	if err := os.WriteFile(filepath.Join(drawioCacheDir, drawioCacheFilename), []byte("opt"), 0o644); err != nil {
		t.Fatalf("Failed to write drawio cache: %v", err)
	}

	// Create orphans of each type
	if err := os.WriteFile(filepath.Join(mermaidCacheDir, "mermaid_orphan.svg"), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to write mermaid orphan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(drawioCacheDir, "drawio_orphan.png"), []byte("opt"), 0o644); err != nil {
		t.Fatalf("Failed to write drawio orphan: %v", err)
	}

	// Run prune
	result, err := PruneCache(tmpDir, false)
	if err != nil {
		t.Fatalf("PruneCache failed: %v", err)
	}

	// Verify both types tracked correctly
	if result.MermaidActiveHashes != 1 {
		t.Errorf("Expected 1 active mermaid hash, got %d", result.MermaidActiveHashes)
	}
	if result.DrawioActiveHashes != 1 {
		t.Errorf("Expected 1 active drawio hash, got %d", result.DrawioActiveHashes)
	}

	// Verify orphans identified
	if len(result.MermaidOrphans) != 1 {
		t.Errorf("Expected 1 mermaid orphan, got %d", len(result.MermaidOrphans))
	}
	if len(result.DrawioOrphans) != 1 {
		t.Errorf("Expected 1 drawio orphan, got %d", len(result.DrawioOrphans))
	}

	// Total should be 2 orphans
	if result.TotalOrphans() != 2 {
		t.Errorf("Expected 2 total orphans, got %d", result.TotalOrphans())
	}
}
