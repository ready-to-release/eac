package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

// RegistryCache manages cached GitHub Container Registry data
type RegistryCache struct {
	Version    string                       `json:"version"`
	Extensions map[string]*ExtensionCache   `json:"extensions"`
	UpdatedAt  time.Time                    `json:"updated_at"`
}

// ExtensionCache holds cached data for a single extension
type ExtensionCache struct {
	Name      string    `json:"name"`
	LatestSHA string    `json:"latest_sha"`  // e.g., "sha-84f1a65"
	Tags      []string  `json:"tags"`        // All available tags
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	cacheVersion = "1.0"
)

// GetCacheDir returns the path to the .r2r/cache directory
func GetCacheDir(repoRoot string) string {
	return filepath.Join(repoRoot, ".r2r", "cache")
}

// GetRegistryCachePath returns the path to the registry cache file
func GetRegistryCachePath(repoRoot string) string {
	return filepath.Join(GetCacheDir(repoRoot), "registry.json")
}

// LoadRegistryCache reads the registry cache from disk
func LoadRegistryCache(repoRoot string) (*RegistryCache, error) {
	cachePath := GetRegistryCachePath(repoRoot)
	log.Debug().Str("path", cachePath).Msg("Loading registry cache from disk")

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty cache if file doesn't exist
			return &RegistryCache{
				Version:    cacheVersion,
				Extensions: make(map[string]*ExtensionCache),
				UpdatedAt:  time.Time{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	var cache RegistryCache
	if err := json.Unmarshal(data, &cache); err != nil {
		log.Warn().Err(err).Msg("Failed to parse cache file, creating new cache")
		// Return empty cache if parsing fails
		return &RegistryCache{
			Version:    cacheVersion,
			Extensions: make(map[string]*ExtensionCache),
			UpdatedAt:  time.Time{},
		}, nil
	}

	// Check version compatibility
	if cache.Version != cacheVersion {
		log.Debug().
			Str("cache_version", cache.Version).
			Str("expected_version", cacheVersion).
			Msg("Cache version mismatch, creating new cache")
		return &RegistryCache{
			Version:    cacheVersion,
			Extensions: make(map[string]*ExtensionCache),
			UpdatedAt:  time.Time{},
		}, nil
	}

	// Initialize map if nil
	if cache.Extensions == nil {
		cache.Extensions = make(map[string]*ExtensionCache)
	}

	return &cache, nil
}

// SaveRegistryCache writes the cache to disk
func (c *RegistryCache) SaveRegistryCache(repoRoot string) error {
	cacheDir := GetCacheDir(repoRoot)

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cachePath := GetRegistryCachePath(repoRoot)
	log.Debug().Str("path", cachePath).Msg("Saving registry cache")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Write atomically by writing to temp file then renaming
	tmpPath := cachePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	if err := os.Rename(tmpPath, cachePath); err != nil {
		os.Remove(tmpPath) // Clean up temp file on failure
		return fmt.Errorf("failed to finalize cache: %w", err)
	}

	log.Debug().
		Str("path", cachePath).
		Int("extensions", len(c.Extensions)).
		Msg("Saved registry cache")

	return nil
}

// IsExpired checks if the cache needs refresh based on the configured TTL
func (c *RegistryCache) IsExpired(ttlSeconds int) bool {
	if c.UpdatedAt.IsZero() {
		return true // Never updated
	}
	
	ttl := time.Duration(ttlSeconds) * time.Second
	return time.Since(c.UpdatedAt) > ttl
}

// GetExtension returns cached data for an extension
func (c *RegistryCache) GetExtension(name string) (*ExtensionCache, bool) {
	ext, ok := c.Extensions[name]
	return ext, ok
}

// SetExtension updates or adds extension cache data
func (c *RegistryCache) SetExtension(name string, latestSHA string, tags []string) {
	if c.Extensions == nil {
		c.Extensions = make(map[string]*ExtensionCache)
	}
	
	c.Extensions[name] = &ExtensionCache{
		Name:      name,
		LatestSHA: latestSHA,
		Tags:      tags,
		UpdatedAt: time.Now(),
	}
	c.UpdatedAt = time.Now()
}

// GetLatestSHA returns the cached latest SHA for an extension
func (c *RegistryCache) GetLatestSHA(extensionName string) (string, bool) {
	if ext, ok := c.Extensions[extensionName]; ok && ext.LatestSHA != "" {
		return ext.LatestSHA, true
	}
	return "", false
}

// Clear removes all cached data
func (c *RegistryCache) Clear() {
	c.Extensions = make(map[string]*ExtensionCache)
	c.UpdatedAt = time.Time{}
}