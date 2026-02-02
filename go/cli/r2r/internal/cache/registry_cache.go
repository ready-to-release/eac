package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/go/cli/r2r/internal/logging"
)

// RegistryCache manages cached GitHub Container Registry data
type RegistryCache struct {
	Version    string                     `json:"version"`
	Extensions map[string]*ExtensionCache `json:"extensions"`
	UpdatedAt  time.Time                  `json:"updated_at"`
}

// ExtensionCache holds cached data for a single extension
type ExtensionCache struct {
	Name         string            `json:"name"`
	LatestSHA    string            `json:"latest_sha"`               // e.g., "sha-84f1a65"
	Tags         []string          `json:"tags"`                     // All available tags
	ImageDigests map[string]string `json:"image_digests,omitempty"` // tag -> digest mapping for fast lookup
	UpdatedAt    time.Time         `json:"updated_at"`
}

const (
	cacheVersion = "1.0"
)

// GetRegistryCachePath returns the path to the registry cache file
func GetRegistryCachePath(repoRoot string) string {
	return filepath.Join(GetCacheDir(repoRoot), "registry.json")
}

// LoadRegistryCache reads the registry cache from disk
func LoadRegistryCache(repoRoot string) (*RegistryCache, error) {
	cachePath := GetRegistryCachePath(repoRoot)
	logging.Debugf("Loading registry cache from disk: path=%s", cachePath)

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
		logging.Warnf("Failed to parse cache file, creating new cache: %v", err)
		// Return empty cache if parsing fails
		return &RegistryCache{
			Version:    cacheVersion,
			Extensions: make(map[string]*ExtensionCache),
			UpdatedAt:  time.Time{},
		}, nil
	}

	// Check version compatibility
	if cache.Version != cacheVersion {
		logging.Debugf("Cache version mismatch, creating new cache: cache_version=%s expected_version=%s", cache.Version, cacheVersion)
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
	logging.Debugf("Saving registry cache: path=%s", cachePath)

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

	logging.Debugf("Saved registry cache: path=%s extensions=%d", cachePath, len(c.Extensions))

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

// GetImageDigest returns the cached digest for a specific image tag
func (c *RegistryCache) GetImageDigest(extensionName, imageTag string) (string, bool) {
	if ext, ok := c.Extensions[extensionName]; ok && ext.ImageDigests != nil {
		digest, found := ext.ImageDigests[imageTag]
		return digest, found
	}
	return "", false
}

// SetImageDigest updates the cached digest for a specific image tag
func (c *RegistryCache) SetImageDigest(extensionName, imageTag, digest string) {
	if c.Extensions == nil {
		c.Extensions = make(map[string]*ExtensionCache)
	}

	ext, ok := c.Extensions[extensionName]
	if !ok {
		ext = &ExtensionCache{
			Name:         extensionName,
			ImageDigests: make(map[string]string),
			UpdatedAt:    time.Now(),
		}
		c.Extensions[extensionName] = ext
	}

	if ext.ImageDigests == nil {
		ext.ImageDigests = make(map[string]string)
	}

	ext.ImageDigests[imageTag] = digest
	ext.UpdatedAt = time.Now()
	c.UpdatedAt = time.Now()
}

// DigestChanged checks if the cached digest differs from the current digest
func (c *RegistryCache) DigestChanged(extensionName, imageTag, currentDigest string) bool {
	cachedDigest, found := c.GetImageDigest(extensionName, imageTag)
	if !found {
		// No cached digest, consider it changed (needs pull)
		return true
	}

	// Digests are different if they don't match
	return cachedDigest != currentDigest
}
