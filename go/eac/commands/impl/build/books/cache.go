package books

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// AssetCache provides content-addressable caching for expensive operations
// like mermaid SVG rendering. Uses SHA256 hashing for cache keys.
type AssetCache struct {
	cacheRoot string
	stats     CacheStats
}

// CacheStats tracks cache hit/miss statistics for reporting
type CacheStats struct {
	MermaidHits   int
	MermaidMisses int
}

// MermaidCacheKey contains all inputs that affect mermaid rendering
// Any change to these values will produce a different cache key
type MermaidCacheKey struct {
	Code   string
	Width  int
	Height int
	Theme  string
}

// NewAssetCache creates a new asset cache rooted at workspace_root/out/cache
func NewAssetCache(workspaceRoot string) *AssetCache {
	return &AssetCache{
		cacheRoot: paths.CachePath(workspaceRoot),
	}
}

// GetMermaid checks if a mermaid SVG is already cached
// Returns: (cachePath, cacheHit)
func (c *AssetCache) GetMermaid(key MermaidCacheKey) (string, bool) {
	hash := c.hashMermaid(key)
	cachePath := paths.MermaidCachePath(c.cacheRoot, hash)

	if _, err := os.Stat(cachePath); err == nil {
		c.stats.MermaidHits++
		return cachePath, true
	}

	c.stats.MermaidMisses++
	return cachePath, false
}

// PutMermaid stores a rendered SVG in the cache for future reuse
func (c *AssetCache) PutMermaid(svgPath string, key MermaidCacheKey) error {
	hash := c.hashMermaid(key)
	cachePath := paths.MermaidCachePath(c.cacheRoot, hash)

	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	// Atomic write: write to temp file, then rename
	// This prevents corruption if multiple builds run concurrently
	tempPath := cachePath + ".tmp"
	if err := copyFile(svgPath, tempPath); err != nil {
		return err
	}

	return os.Rename(tempPath, cachePath)
}

// hashMermaid creates a deterministic hash of all mermaid rendering inputs
func (c *AssetCache) hashMermaid(key MermaidCacheKey) string {
	h := sha256.New()
	fmt.Fprintf(h, "code:%s\n", key.Code)
	fmt.Fprintf(h, "width:%d\n", key.Width)
	fmt.Fprintf(h, "height:%d\n", key.Height)
	fmt.Fprintf(h, "theme:%s\n", key.Theme)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Stats returns current cache statistics
func (c *AssetCache) Stats() CacheStats {
	return c.stats
}

// Clear removes all cached assets (useful for troubleshooting)
func (c *AssetCache) Clear() error {
	return os.RemoveAll(c.cacheRoot)
}
