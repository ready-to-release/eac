package caching

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/staging"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/paths"
)

// AssetCache provides content-addressable caching for expensive operations
// like mermaid SVG rendering. Uses SHA256 hashing for cache keys.
type AssetCache struct {
	cacheRoot   string
	cacheConfig *cache.Config
	stats       CacheStats
	debugf      func(string, ...any)
}

// CacheStats tracks cache hit/miss statistics for reporting.
type CacheStats struct {
	MermaidHits   int
	MermaidMisses int
	DrawioHits    int
	DrawioMisses  int
}

// MermaidCacheKey contains all inputs that affect mermaid rendering.
type MermaidCacheKey struct {
	SourceFile string
	BlockIndex int
	Code       string
	Width      int
	Height     int
	Theme      string
}

// DrawioCacheKey contains all inputs that affect drawio image optimization.
type DrawioCacheKey struct {
	SourcePath string
	SourceHash string
	MaxWidth   int
}

// NewAssetCache creates a new asset cache rooted at docs/assets/cache.
func NewAssetCache(workspaceRoot string, cacheConfig *cache.Config, debugf func(string, ...any)) *AssetCache {
	if debugf == nil {
		debugf = func(string, ...any) {}
	}
	return &AssetCache{
		cacheRoot:   paths.DocsCachePath(workspaceRoot),
		cacheConfig: cacheConfig,
		debugf:      debugf,
	}
}

// GetMermaid checks if a mermaid SVG is already cached.
// Returns: (cachePath, cacheHit).
func (c *AssetCache) GetMermaid(key MermaidCacheKey) (string, bool) {
	hash := HashMermaidKey(key)
	cachePath := paths.MermaidCachePath(c.cacheRoot, key.SourceFile, key.BlockIndex, hash)

	if c.cacheConfig != nil && c.cacheConfig.ShouldSkipAsset() {
		c.stats.MermaidMisses++
		c.debugf("cache: mermaid SKIP (--skip-cache=asset) block=%d hash=%s", key.BlockIndex, hash[:8])
		return cachePath, false
	}

	if _, err := os.Stat(cachePath); err == nil {
		c.stats.MermaidHits++
		c.debugf("cache: mermaid HIT block=%d hash=%s path=%s", key.BlockIndex, hash[:8], cachePath)
		return cachePath, true
	}

	c.stats.MermaidMisses++
	c.debugf("cache: mermaid MISS block=%d hash=%s path=%s", key.BlockIndex, hash[:8], cachePath)
	return cachePath, false
}

// PutMermaid stores a rendered SVG in the cache for future reuse.
func (c *AssetCache) PutMermaid(svgPath string, key MermaidCacheKey) error {
	hash := HashMermaidKey(key)
	cachePath := paths.MermaidCachePath(c.cacheRoot, key.SourceFile, key.BlockIndex, hash)

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}

	tempPath := cachePath + ".tmp"
	if err := staging.CopyFile(svgPath, tempPath); err != nil {
		return err
	}

	return os.Rename(tempPath, cachePath)
}

// HashMermaidKey computes a deterministic hash for a mermaid cache key.
func HashMermaidKey(key MermaidCacheKey) string {
	h := sha256.New()
	fmt.Fprintf(h, "code:%s\n", key.Code)
	fmt.Fprintf(h, "width:%d\n", key.Width)
	fmt.Fprintf(h, "height:%d\n", key.Height)
	fmt.Fprintf(h, "theme:%s\n", key.Theme)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// GetDrawio checks if an optimized drawio PNG is already cached.
// Returns: (cachePath, cacheHit).
func (c *AssetCache) GetDrawio(key DrawioCacheKey) (string, bool) {
	hash := HashDrawioKey(key)
	cachePath := paths.DrawioCachePath(c.cacheRoot, key.SourcePath, hash)

	if c.cacheConfig != nil && c.cacheConfig.ShouldSkipAsset() {
		c.stats.DrawioMisses++
		c.debugf("cache: drawio SKIP (--skip-cache=asset) hash=%s", hash[:8])
		return cachePath, false
	}

	if _, err := os.Stat(cachePath); err == nil {
		c.stats.DrawioHits++
		c.debugf("cache: drawio HIT hash=%s path=%s", hash[:8], cachePath)
		return cachePath, true
	}

	c.stats.DrawioMisses++
	c.debugf("cache: drawio MISS hash=%s path=%s", hash[:8], cachePath)
	return cachePath, false
}

// PutDrawio stores an optimized PNG in the cache for future reuse.
func (c *AssetCache) PutDrawio(pngPath string, key DrawioCacheKey) error {
	hash := HashDrawioKey(key)
	cachePath := paths.DrawioCachePath(c.cacheRoot, key.SourcePath, hash)

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}

	tempPath := cachePath + ".tmp"
	if err := staging.CopyFile(pngPath, tempPath); err != nil {
		return err
	}

	return os.Rename(tempPath, cachePath)
}

// HashDrawioKey computes a deterministic hash for a drawio cache key.
func HashDrawioKey(key DrawioCacheKey) string {
	h := sha256.New()
	fmt.Fprintf(h, "source:%s\n", key.SourceHash)
	fmt.Fprintf(h, "maxWidth:%d\n", key.MaxWidth)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Stats returns current cache statistics.
func (c *AssetCache) Stats() CacheStats {
	return c.stats
}

// Clear removes all cached assets.
func (c *AssetCache) Clear() error {
	return os.RemoveAll(c.cacheRoot)
}
