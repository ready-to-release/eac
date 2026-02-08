// Package itemcache provides a per-item content-addressable cache for builders.
// Each builder type gets its own isolated cache directory and manifest.
// Items are cached individually by content hash -- only changed items are rebuilt.
package itemcache

// Item represents a single cacheable unit of work within a builder.
type Item struct {
	// Key uniquely identifies this item within the builder scope.
	Key string

	// ContentHash is the content hash (typically 8 chars of SHA256).
	// Changes to this value trigger a rebuild of this item.
	ContentHash string

	// CacheFilename is the filename in the cache directory.
	CacheFilename string

	// OutputRelPath is the relative path in the output directory.
	OutputRelPath string
}

// BuildFunc renders cache-missed items.
// Receives items that need building and the cache directory to write results to.
// Each item's CacheFilename specifies where to write the result within cacheDir.
// Returns (built count, error).
type BuildFunc func(items []Item, cacheDir string) (int, error)

// Result summarizes the outcome of a cache-accelerated build.
type Result struct {
	TotalItems  int
	CacheHits   int
	CacheMisses int
	Built       int
	HitRate     float64
}
