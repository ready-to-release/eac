package itemcache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/go/core/logging"
)

var log = logging.C()

// Cache manages per-item caching for a single builder type.
type Cache struct {
	cacheDir     string
	manifestPath string
	manifest     *Manifest
}

// New creates a new per-item cache for a builder type.
func New(cacheDir string) *Cache {
	return &Cache{
		cacheDir:     cacheDir,
		manifestPath: filepath.Join(cacheDir, "item-manifest.json"),
	}
}

// Execute runs a cache-accelerated build.
//
// Algorithm:
//  1. Load manifest
//  2. Classify items as cache hits or misses
//  3. Call buildFunc for misses only
//  4. Copy ALL items to outputDir
//  5. Update and save manifest
//  6. Prune stale entries
func (c *Cache) Execute(
	items []Item,
	outputDir string,
	buildFunc BuildFunc,
	forceRebuild bool,
) (*Result, error) {
	// 1. Load manifest
	c.manifest = loadManifest(c.manifestPath)

	// 2. Ensure directories exist
	if err := os.MkdirAll(c.cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating output dir: %w", err)
	}

	// 3. Classify items
	var hits, misses []Item
	for _, item := range items {
		if forceRebuild {
			misses = append(misses, item)
			continue
		}

		cached, ok := c.manifest.Items[item.Key]
		if !ok || cached.ContentHash != item.ContentHash {
			misses = append(misses, item)
			continue
		}

		// Verify cache file actually exists on disk
		cachePath := filepath.Join(c.cacheDir, item.CacheFilename)
		if _, err := os.Stat(cachePath); err != nil {
			misses = append(misses, item)
			continue
		}

		hits = append(hits, item)
	}

	log.Debugf("itemcache: %d items, %d hits, %d misses (force=%v)",
		len(items), len(hits), len(misses), forceRebuild)

	// 4. Build cache misses
	built := 0
	if len(misses) > 0 {
		var err error
		built, err = buildFunc(misses, c.cacheDir)
		if err != nil {
			return nil, fmt.Errorf("building items: %w", err)
		}
	}

	// 5. Copy ALL items from cache to output
	now := time.Now().Unix()
	for _, item := range items {
		cachePath := filepath.Join(c.cacheDir, item.CacheFilename)
		outputPath := filepath.Join(outputDir, item.OutputRelPath)

		// Verify cache file exists (buildFunc should have produced it)
		if _, err := os.Stat(cachePath); err != nil {
			return nil, fmt.Errorf("expected item not found in cache: %s", cachePath)
		}

		// Ensure output subdirectory exists
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return nil, fmt.Errorf("creating output subdir: %w", err)
		}

		// Copy from cache to output
		if err := copyFile(cachePath, outputPath); err != nil {
			return nil, fmt.Errorf("copying %s to output: %w", item.Key, err)
		}

		// Update manifest
		c.manifest.Items[item.Key] = &CachedItem{
			ContentHash:   item.ContentHash,
			CacheFilename: item.CacheFilename,
			CachedAt:      now,
		}
	}

	// 6. Prune stale entries
	currentKeys := make(map[string]bool, len(items))
	for _, item := range items {
		currentKeys[item.Key] = true
	}
	c.prune(currentKeys)

	// 7. Save manifest
	if err := saveManifest(c.manifest, c.manifestPath); err != nil {
		log.Debugf("itemcache: failed to save manifest: %v", err)
		// Non-fatal: build succeeded even if manifest save fails
	}

	// 8. Compute result
	hitRate := 0.0
	if len(items) > 0 {
		hitRate = float64(len(hits)) / float64(len(items)) * 100
	}

	return &Result{
		TotalItems:  len(items),
		CacheHits:   len(hits),
		CacheMisses: len(misses),
		Built:       built,
		HitRate:     hitRate,
	}, nil
}

// prune removes manifest entries and cache files not in currentKeys.
func (c *Cache) prune(currentKeys map[string]bool) {
	for key, cached := range c.manifest.Items {
		if currentKeys[key] {
			continue
		}

		// Remove cache file
		cachePath := filepath.Join(c.cacheDir, cached.CacheFilename)
		if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
			log.Debugf("itemcache: failed to prune %s: %v", cachePath, err)
		}

		// Remove manifest entry
		delete(c.manifest.Items, key)
		log.Debugf("itemcache: pruned stale entry: %s", key)
	}
}

// copyFile copies src to dst, preserving permissions.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}
