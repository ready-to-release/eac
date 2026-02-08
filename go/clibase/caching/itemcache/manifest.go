package itemcache

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const manifestVersion = 1

// Manifest tracks cached items and their content hashes.
type Manifest struct {
	Version int                    `json:"version"`
	Items   map[string]*CachedItem `json:"items"`
}

// CachedItem tracks the cached state of a single item.
type CachedItem struct {
	ContentHash   string `json:"content_hash"`
	CacheFilename string `json:"cache_filename"`
	CachedAt      int64  `json:"cached_at"`
}

// loadManifest reads the manifest from disk. Returns empty manifest if missing/corrupt.
func loadManifest(path string) *Manifest {
	m := &Manifest{
		Version: manifestVersion,
		Items:   make(map[string]*CachedItem),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return m
	}

	var loaded Manifest
	if err := json.Unmarshal(data, &loaded); err != nil {
		return m
	}
	if loaded.Version != manifestVersion {
		return m
	}
	if loaded.Items == nil {
		loaded.Items = make(map[string]*CachedItem)
	}

	m.Items = loaded.Items
	return m
}

// saveManifest writes the manifest atomically (temp + rename).
func saveManifest(m *Manifest, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
