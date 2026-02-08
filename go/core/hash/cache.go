// Package hash provides deterministic file content hashing for change detection.
package hash

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync"
)

const cacheVersion = 1

// HashCache caches computed input hashes with file mtime metadata.
// On subsequent runs, if all file mtimes match cached values, the cached hash
// is returned without reading file contents.
type HashCache struct {
	Version int                     `json:"version"`
	Modules map[string]*ModuleCache `json:"modules"`
	mu      sync.RWMutex
	path    string
	dirty   bool
}

// ModuleCache stores the cached hash and file mtimes for a single module.
type ModuleCache struct {
	Patterns []string         `json:"patterns"`
	Hash     string           `json:"hash"`
	Files    map[string]int64 `json:"files"` // relative path -> mtime UnixNano
}

// LoadCache loads from disk. Returns empty cache if missing or corrupt.
func LoadCache(filePath string) *HashCache {
	c := &HashCache{
		Version: cacheVersion,
		Modules: make(map[string]*ModuleCache),
		path:    filePath,
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return c
	}

	var loaded HashCache
	if err := json.Unmarshal(data, &loaded); err != nil {
		return c
	}

	if loaded.Version != cacheVersion {
		return c
	}

	if loaded.Modules == nil {
		loaded.Modules = make(map[string]*ModuleCache)
	}

	c.Modules = loaded.Modules
	return c
}

// Save persists the cache atomically (temp + rename). Only writes if dirty.
func (c *HashCache) Save() error {
	c.mu.RLock()
	if !c.dirty {
		c.mu.RUnlock()
		return nil
	}
	data, err := json.MarshalIndent(c, "", "  ")
	c.mu.RUnlock()
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Atomic write: temp file + rename
	tmp := c.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, c.path)
}

// GetOrCompute returns (hash, files, error).
// Fast path: patterns match AND all mtimes unchanged -> return cached hash.
// Slow path: expand globs, hash files, update cache.
func (c *HashCache) GetOrCompute(module string, patterns []string, workspaceRoot string) (string, []string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache hit
	if cached, ok := c.Modules[module]; ok {
		if slices.Equal(cached.Patterns, patterns) && mtimesUnchanged(workspaceRoot, cached.Files) {
			// Reconstruct sorted file list from cache keys
			files := make([]string, 0, len(cached.Files))
			for f := range cached.Files {
				files = append(files, f)
			}
			sort.Strings(files)
			return cached.Hash, files, nil
		}
	}

	// Cache miss: full computation
	files, err := ExpandGlobPatterns(workspaceRoot, patterns)
	if err != nil {
		return "", nil, err
	}

	h, err := Files(workspaceRoot, files)
	if err != nil {
		return "", nil, err
	}

	mtimes := collectMtimes(workspaceRoot, files)

	c.Modules[module] = &ModuleCache{
		Patterns: patterns,
		Hash:     h,
		Files:    mtimes,
	}
	c.dirty = true

	return h, files, nil
}
