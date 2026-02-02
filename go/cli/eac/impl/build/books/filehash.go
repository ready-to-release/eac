// Package books provides book preprocessing for MkDocs sites.
package books

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/ready-to-release/eac/go/core/paths"
)

// FileHashCache tracks content hashes for files to enable incremental processing.
// It stores SHA256 hashes keyed by file path, allowing the preprocessor to skip
// files that haven't changed since the last build.
type FileHashCache struct {
	BookName string            `json:"book_name"`
	Hashes   map[string]string `json:"hashes"`

	workspaceRoot string
	mu            sync.RWMutex
	stats         CacheHitStats
}

// CacheHitStats tracks cache hit/miss statistics for monitoring performance.
type CacheHitStats struct {
	Hits   int
	Misses int
}

// NewFileHashCache creates a new empty cache for a book.
func NewFileHashCache(bookName, workspaceRoot string) *FileHashCache {
	return &FileHashCache{
		BookName:      bookName,
		Hashes:        make(map[string]string),
		workspaceRoot: workspaceRoot,
	}
}

// LoadFileHashCache loads an existing cache from disk.
// Returns an empty cache if the file doesn't exist or is corrupt (graceful degradation).
func LoadFileHashCache(bookName, workspaceRoot string) (*FileHashCache, error) {
	cache := NewFileHashCache(bookName, workspaceRoot)

	data, err := os.ReadFile(cache.CachePath())
	if err != nil || len(data) == 0 {
		return cache, nil
	}

	// Try to unmarshal - if it fails, return empty cache
	var loaded FileHashCache
	if err := json.Unmarshal(data, &loaded); err != nil {
		return cache, nil
	}

	// Copy loaded hashes into cache (preserve bookName and workspaceRoot from parameters)
	if loaded.Hashes != nil {
		cache.Hashes = loaded.Hashes
	}

	return cache, nil
}

// Save persists the cache to disk.
// Creates parent directories if they don't exist.
func (c *FileHashCache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create cache directory if needed
	cacheDir := filepath.Dir(c.CachePath())
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.CachePath(), data, 0o644)
}

// CachePath returns the path where this cache is stored.
// Path: .cache/eac/build/hashes/{bookName}.json
func (c *FileHashCache) CachePath() string {
	return paths.FileHashCachePath(c.workspaceRoot, c.BookName)
}

// ShouldProcessFile checks if a file needs processing by comparing its content hash.
// Returns true if the file is new or changed, false if unchanged.
// Always updates the cache with the current hash.
func (c *FileHashCache) ShouldProcessFile(path string, content []byte) bool {
	newHash := computeHash(content)

	c.mu.Lock()
	defer c.mu.Unlock()

	existingHash, exists := c.Hashes[path]
	if exists && existingHash == newHash {
		c.stats.Hits++
		log.Debugf("cache: filehash HIT path=%s hash=%s", path, newHash[:8])
		return false // unchanged
	}

	// New or changed file - update cache
	c.Hashes[path] = newHash
	c.stats.Misses++
	if exists {
		log.Debugf("cache: filehash MISS (changed) path=%s prev=%s curr=%s", path, existingHash[:8], newHash[:8])
	} else {
		log.Debugf("cache: filehash MISS (new) path=%s hash=%s", path, newHash[:8])
	}
	return true
}

// RemoveFile removes a file entry from the cache.
// This is used when a file is deleted from the source.
func (c *FileHashCache) RemoveFile(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.Hashes, path)
}

// Clear removes all entries from the cache.
func (c *FileHashCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Hashes = make(map[string]string)
}

// Stats returns cache hit/miss statistics.
func (c *FileHashCache) Stats() CacheHitStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.stats
}

// GetTrackedPaths returns all file paths currently in the cache.
// Used to detect orphaned files that exist in staging but were deleted from source.
func (c *FileHashCache) GetTrackedPaths() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	paths := make([]string, 0, len(c.Hashes))
	for path := range c.Hashes {
		paths = append(paths, path)
	}
	return paths
}

// computeHash calculates the SHA256 hash of content and returns it as a hex string.
func computeHash(content []byte) string {
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:])
}
