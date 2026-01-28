// do not delete log.Debugf calls: we always want to be able to track cache usage
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
	DrawioHits    int
	DrawioMisses  int
}

// MermaidCacheKey contains all inputs that affect mermaid rendering
// Any change to these values will produce a different cache key
type MermaidCacheKey struct {
	SourceFile string // Path to .md file containing the block (for traceable naming)
	BlockIndex int    // Index of mermaid block in file (0, 1, 2...)
	Code       string
	Width      int
	Height     int
	Theme      string
}

// DrawioCacheKey contains all inputs that affect drawio image optimization
// Any change to these values will produce a different cache key
type DrawioCacheKey struct {
	SourcePath string // Path to source .drawio.png (for traceable naming)
	SourceHash string // SHA256 of original image file content
	MaxWidth   int    // Target max width for optimization
}

// NewAssetCache creates a new asset cache rooted at docs/assets/cache
// This cache is git-tracked for CI optimization (pre-rendered mermaid diagrams)
func NewAssetCache(workspaceRoot string) *AssetCache {
	return &AssetCache{
		cacheRoot: paths.DocsCachePath(workspaceRoot),
	}
}

// GetMermaid checks if a mermaid SVG is already cached
// Returns: (cachePath, cacheHit)
func (c *AssetCache) GetMermaid(key MermaidCacheKey) (string, bool) {
	hash := c.hashMermaid(key)
	cachePath := paths.MermaidCachePath(c.cacheRoot, key.SourceFile, key.BlockIndex, hash)

	if _, err := os.Stat(cachePath); err == nil {
		c.stats.MermaidHits++
		log.Debugf("cache: mermaid HIT block=%d hash=%s path=%s", key.BlockIndex, hash[:8], cachePath)
		return cachePath, true
	}

	c.stats.MermaidMisses++
	log.Debugf("cache: mermaid MISS block=%d hash=%s path=%s", key.BlockIndex, hash[:8], cachePath)
	return cachePath, false
}

// PutMermaid stores a rendered SVG in the cache for future reuse
func (c *AssetCache) PutMermaid(svgPath string, key MermaidCacheKey) error {
	hash := c.hashMermaid(key)
	cachePath := paths.MermaidCachePath(c.cacheRoot, key.SourceFile, key.BlockIndex, hash)

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
	return HashMermaidKey(key)
}

// HashMermaidKey computes a deterministic hash for a mermaid cache key.
// Exported for use by cache pruning logic.
func HashMermaidKey(key MermaidCacheKey) string {
	h := sha256.New()
	fmt.Fprintf(h, "code:%s\n", key.Code)
	fmt.Fprintf(h, "width:%d\n", key.Width)
	fmt.Fprintf(h, "height:%d\n", key.Height)
	fmt.Fprintf(h, "theme:%s\n", key.Theme)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// GetDrawio checks if an optimized drawio PNG is already cached
// Returns: (cachePath, cacheHit)
func (c *AssetCache) GetDrawio(key DrawioCacheKey) (string, bool) {
	hash := c.hashDrawio(key)
	cachePath := paths.DrawioCachePath(c.cacheRoot, key.SourcePath, hash)

	if _, err := os.Stat(cachePath); err == nil {
		c.stats.DrawioHits++
		log.Debugf("cache: drawio HIT hash=%s path=%s", hash[:8], cachePath)
		return cachePath, true
	}

	c.stats.DrawioMisses++
	log.Debugf("cache: drawio MISS hash=%s path=%s", hash[:8], cachePath)
	return cachePath, false
}

// PutDrawio stores an optimized PNG in the cache for future reuse
func (c *AssetCache) PutDrawio(pngPath string, key DrawioCacheKey) error {
	hash := c.hashDrawio(key)
	cachePath := paths.DrawioCachePath(c.cacheRoot, key.SourcePath, hash)

	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	// Atomic write: write to temp file, then rename
	// This prevents corruption if multiple builds run concurrently
	tempPath := cachePath + ".tmp"
	if err := copyFile(pngPath, tempPath); err != nil {
		return err
	}

	return os.Rename(tempPath, cachePath)
}

// hashDrawio creates a deterministic hash of all drawio optimization inputs
func (c *AssetCache) hashDrawio(key DrawioCacheKey) string {
	return HashDrawioKey(key)
}

// HashDrawioKey computes a deterministic hash for a drawio cache key.
// Exported for use by cache pruning logic.
func HashDrawioKey(key DrawioCacheKey) string {
	h := sha256.New()
	fmt.Fprintf(h, "source:%s\n", key.SourceHash)
	fmt.Fprintf(h, "maxWidth:%d\n", key.MaxWidth)
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
