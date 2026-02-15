// Package docs provides the update docs command and cache pruning functionality.
package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/commands/build/docprep/caching"
	"github.com/ready-to-release/eac/go/commands/build/docprep/diagrams"
	"github.com/ready-to-release/eac/go/core/paths"
)

// PruneResult contains statistics from cache pruning.
type PruneResult struct {
	// Mermaid statistics
	MermaidCacheFiles     int      // Total files in mermaid cache
	MermaidActiveHashes   int      // Hashes computed from current markdown
	MermaidOrphans        []string // Filenames to be pruned
	MermaidBytesRecovered int64    // Space that would be recovered

	// Drawio statistics
	DrawioCacheFiles     int      // Total files in drawio cache
	DrawioActiveHashes   int      // Hashes computed from current sources
	DrawioOrphans        []string // Filenames to be pruned
	DrawioBytesRecovered int64    // Space that would be recovered
}

// TotalOrphans returns the total number of orphaned files.
func (r *PruneResult) TotalOrphans() int {
	return len(r.MermaidOrphans) + len(r.DrawioOrphans)
}

// TotalBytesRecovered returns the total bytes that would be recovered.
func (r *PruneResult) TotalBytesRecovered() int64 {
	return r.MermaidBytesRecovered + r.DrawioBytesRecovered
}

// PruneCache identifies orphaned cache files that are no longer needed.
// It scans all markdown files for mermaid blocks and all drawio.png sources,
// computes the expected cache filenames, and compares against actual cache files.
//
// This function only identifies orphans - it does not delete them.
// Call DeleteOrphans() to actually remove files.
func PruneCache(repoRoot string, verbose bool) (*PruneResult, error) {
	docsDir := paths.DocsSourcePath(repoRoot)
	cacheDir := paths.CacheRootPath(repoRoot)

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

	// Scan mermaid cache directory for orphans
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")

	mermaidOrphans, mermaidBytes, mermaidTotal := findOrphans(mermaidCacheDir, mermaidFilenames, ".svg")
	result.MermaidOrphans = mermaidOrphans
	result.MermaidBytesRecovered = mermaidBytes
	result.MermaidCacheFiles = mermaidTotal

	// Scan drawio cache directory for orphans
	drawioCacheDir := filepath.Join(cacheDir, "drawio")
	drawioOrphans, drawioBytes, drawioTotal := findOrphans(drawioCacheDir, drawioFilenames, ".png")
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
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		blocks := diagrams.ExtractMermaidBlocks(string(content), path, docsDir)
		for _, block := range blocks {
			// Compute the full cache path using the same hash algorithm
			cleanContent := diagrams.StripSizeDirective(block.Content)
			hash := computeMermaidCacheHash(cleanContent)
			cachePath := paths.MermaidCachePath(cacheRoot, block.SourceFile, block.BlockIndex, hash)
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

	images, err := diagrams.FindDrawioImages(docsDir)
	if err != nil {
		return nil, fmt.Errorf("finding drawio images: %w", err)
	}

	for _, img := range images {
		// Compute the full cache path using the same hash algorithm
		hash := computeDrawioCacheHash(img.Hash, diagrams.MaxImageWidthPDF)
		cachePath := paths.DrawioCachePath(cacheRoot, img.SourceFile, hash)
		filename := filepath.Base(cachePath)
		filenames[filename] = true
	}

	return filenames, nil
}

// computeMermaidCacheHash computes the cache key hash for mermaid content.
// Uses the exported caching.HashMermaidKey to ensure consistency.
func computeMermaidCacheHash(code string) string {
	return caching.HashMermaidKey(caching.MermaidCacheKey{Code: code})
}

// computeDrawioCacheHash computes the cache key hash for drawio images.
// Uses the exported caching.HashDrawioKey to ensure consistency.
func computeDrawioCacheHash(sourceHash string, maxWidth int) string {
	return caching.HashDrawioKey(caching.DrawioCacheKey{SourceHash: sourceHash, MaxWidth: maxWidth})
}

// findOrphans scans a cache directory and identifies files that are not in the
// active filenames set.
// Returns: (orphan filenames, total orphan bytes, total cache files)
func findOrphans(cacheDir string, activeFilenames map[string]bool, ext string) ([]string, int64, int) {
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

