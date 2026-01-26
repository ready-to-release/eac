// Package docs provides the update docs command and cache pruning functionality.
package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/books"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// PruneResult contains statistics from cache pruning.
type PruneResult struct {
	// Mermaid statistics
	MermaidCacheFiles     int      // Total files in mermaid cache
	MermaidActiveHashes   int      // Hashes computed from current markdown
	MermaidOrphans        []string // Filenames to be pruned (V2 format orphans)
	MermaidBytesRecovered int64    // Space that would be recovered

	// Legacy mermaid statistics (64-char hash naming convention)
	LegacyMermaidFiles []string // Legacy filenames to be pruned
	LegacyMermaidBytes int64    // Space that would be recovered from legacy files

	// Drawio statistics
	DrawioCacheFiles     int      // Total files in drawio cache
	DrawioActiveHashes   int      // Hashes computed from current sources
	DrawioOrphans        []string // Filenames to be pruned
	DrawioBytesRecovered int64    // Space that would be recovered
}

// TotalOrphans returns the total number of orphaned files (including legacy).
func (r *PruneResult) TotalOrphans() int {
	return len(r.MermaidOrphans) + len(r.DrawioOrphans) + len(r.LegacyMermaidFiles)
}

// TotalBytesRecovered returns the total bytes that would be recovered (including legacy).
func (r *PruneResult) TotalBytesRecovered() int64 {
	return r.MermaidBytesRecovered + r.DrawioBytesRecovered + r.LegacyMermaidBytes
}

// PruneCache identifies orphaned cache files that are no longer needed.
// It scans all markdown files for mermaid blocks and all drawio.png sources,
// computes the expected cache filenames, and compares against actual cache files.
//
// This function only identifies orphans - it does not delete them.
// Call DeleteOrphans() to actually remove files.
func PruneCache(repoRoot string, verbose bool) (*PruneResult, error) {
	docsDir := paths.DocsSourcePath(repoRoot)
	cacheDir := paths.DocsCachePath(repoRoot)

	result := &PruneResult{}

	// Compute expected mermaid cache filenames from all markdown files
	mermaidFilenames, err := computeMermaidFilenames(docsDir, cacheDir)
	if err != nil {
		return nil, fmt.Errorf("scanning mermaid: %w", err)
	}
	result.MermaidActiveHashes = len(mermaidFilenames)

	// Compute expected drawio cache filenames from all drawio.png files
	drawioFilenames, err := computeDrawioFilenames(docsDir, cacheDir)
	if err != nil {
		return nil, fmt.Errorf("scanning drawio: %w", err)
	}
	result.DrawioActiveHashes = len(drawioFilenames)

	// Scan mermaid cache directory for orphans and legacy files
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")

	// Find legacy files first (64-char hash naming convention)
	legacyFiles, legacyBytes := FindLegacyMermaidFiles(mermaidCacheDir)
	result.LegacyMermaidFiles = legacyFiles
	result.LegacyMermaidBytes = legacyBytes

	// Find V2 format orphans (excluding legacy files - they're tracked separately)
	mermaidOrphans, mermaidBytes, mermaidTotal := findOrphans(mermaidCacheDir, mermaidFilenames, ".svg", true)
	result.MermaidOrphans = mermaidOrphans
	result.MermaidBytesRecovered = mermaidBytes
	result.MermaidCacheFiles = mermaidTotal

	// Scan drawio cache directory for orphans (no legacy format exists for drawio)
	drawioCacheDir := filepath.Join(cacheDir, "drawio")
	drawioOrphans, drawioBytes, drawioTotal := findOrphans(drawioCacheDir, drawioFilenames, ".png", false)
	result.DrawioOrphans = drawioOrphans
	result.DrawioBytesRecovered = drawioBytes
	result.DrawioCacheFiles = drawioTotal

	return result, nil
}

// DeleteOrphans removes the orphaned cache files identified in a PruneResult.
// Returns the number of files deleted.
func DeleteOrphans(result *PruneResult, cacheDir string) (int, error) {
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")
	drawioCacheDir := filepath.Join(cacheDir, "drawio")

	deleted := 0
	deleted += deleteFiles(mermaidCacheDir, result.MermaidOrphans)
	deleted += deleteFiles(mermaidCacheDir, result.LegacyMermaidFiles)
	deleted += deleteFiles(drawioCacheDir, result.DrawioOrphans)

	return deleted, nil
}

// deleteFiles removes files from a directory, logging warnings for errors.
// Returns the count of files processed (including already-missing files).
func deleteFiles(dir string, filenames []string) int {
	for _, filename := range filenames {
		path := filepath.Join(dir, filename)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Warnf("Failed to delete %s: %v", path, err)
		}
	}
	return len(filenames)
}

// computeMermaidFilenames scans all markdown files and returns a set of
// expected mermaid cache filenames (traceable format: {identifier}_{blockIndex}_{hash8}.svg).
func computeMermaidFilenames(docsDir, cacheRoot string) (map[string]bool, error) {
	filenames := make(map[string]bool)

	err := filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip the cache directory itself
			if strings.Contains(path, "assets"+string(os.PathSeparator)+"cache") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		blocks := books.ExtractMermaidBlocks(string(content), path, docsDir)
		for _, block := range blocks {
			// Use same algorithm as AssetCache - compute the full cache path
			cleanContent := books.StripSizeDirective(block.Content)
			hash := computeMermaidCacheHash(cleanContent)
			cachePath := paths.MermaidCachePathV2(cacheRoot, block.SourceFile, block.BlockIndex, hash)
			filename := filepath.Base(cachePath)
			filenames[filename] = true
		}
		return nil
	})

	return filenames, err
}

// computeDrawioFilenames scans all drawio.png files and returns a set of
// expected cache filenames (traceable format: {identifier}_{hash8}.png).
func computeDrawioFilenames(docsDir, cacheRoot string) (map[string]bool, error) {
	filenames := make(map[string]bool)

	images, err := books.FindDrawioImages(docsDir)
	if err != nil {
		return nil, fmt.Errorf("finding drawio images: %w", err)
	}

	for _, img := range images {
		// Use same algorithm as AssetCache - compute the full cache path
		hash := computeDrawioCacheHash(img.Hash, books.MaxImageWidthPDF)
		cachePath := paths.DrawioCachePathV2(cacheRoot, img.SourceFile, hash)
		filename := filepath.Base(cachePath)
		filenames[filename] = true
	}

	return filenames, nil
}

// computeMermaidCacheHash computes the cache key hash for mermaid content.
// Uses the exported books.HashMermaidKey to ensure consistency.
func computeMermaidCacheHash(code string) string {
	return books.HashMermaidKey(books.MermaidCacheKey{Code: code})
}

// computeDrawioCacheHash computes the cache key hash for drawio images.
// Uses the exported books.HashDrawioKey to ensure consistency.
func computeDrawioCacheHash(sourceHash string, maxWidth int) string {
	return books.HashDrawioKey(books.DrawioCacheKey{SourceHash: sourceHash, MaxWidth: maxWidth})
}

// findOrphans scans a cache directory and identifies files that are not in the
// active filenames set. If skipLegacy is true, files matching the legacy mermaid
// naming convention (64-char SHA256 hash) are excluded from orphan detection.
// Returns: (orphan filenames, total orphan bytes, total cache files)
func findOrphans(cacheDir string, activeFilenames map[string]bool, ext string, skipLegacy bool) ([]string, int64, int) {
	var orphans []string
	var bytes int64
	total := 0

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil, 0, 0
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ext) {
			continue
		}

		total++

		// Skip legacy files if requested (they're tracked separately)
		if skipLegacy && IsLegacyMermaidCacheFile(entry.Name()) {
			continue
		}

		if !activeFilenames[entry.Name()] {
			orphans = append(orphans, entry.Name())
			if info, err := entry.Info(); err == nil {
				bytes += info.Size()
			}
		}
	}

	return orphans, bytes, total
}

// formatBytes formats bytes as human-readable string (B, KB, MB).
func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}

// IsLegacyMermaidCacheFile returns true if the filename matches the legacy
// 64-character SHA256 hash naming convention (e.g., "abc123...def.svg").
// Legacy files are pure hash filenames without the traceable identifier prefix.
func IsLegacyMermaidCacheFile(filename string) bool {
	basename, found := strings.CutSuffix(filename, ".svg")
	if !found || len(basename) != 64 {
		return false
	}

	for _, c := range basename {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// FindLegacyMermaidFiles scans the mermaid cache directory and returns all
// legacy cache files (64-char hash naming convention) along with their total size.
// Returns: (legacy filenames, total bytes)
func FindLegacyMermaidFiles(mermaidCacheDir string) ([]string, int64) {
	var legacyFiles []string
	var totalBytes int64

	entries, err := os.ReadDir(mermaidCacheDir)
	if err != nil {
		// Directory doesn't exist or other error - return empty
		return nil, 0
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if IsLegacyMermaidCacheFile(entry.Name()) {
			legacyFiles = append(legacyFiles, entry.Name())
			if info, err := entry.Info(); err == nil {
				totalBytes += info.Size()
			}
		}
	}

	return legacyFiles, totalBytes
}
