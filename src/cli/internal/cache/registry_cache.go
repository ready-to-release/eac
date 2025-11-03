package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/src/cli/internal/session"
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
	cacheVersion  = "1.0"
)

// GetCachePath returns the path to the session-specific cache file
func GetCachePath() string {
	// Get session identifier for session-specific cache
	sessionID := session.GetIdentifier()
	
	// Use temp directory for cache file
	cacheDir := filepath.Join(os.TempDir(), "r2r-cli-cache")
	// Create directory if it doesn't exist
	os.MkdirAll(cacheDir, 0755)
	
	// Session-specific cache file name
	cacheFileName := fmt.Sprintf("r2r-cli-cache-%s.json", sessionID)
	return filepath.Join(cacheDir, cacheFileName)
}

// Load reads the cache from disk
func Load() (*RegistryCache, error) {
	cachePath := GetCachePath()
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

// Save writes the cache to disk
func (c *RegistryCache) Save() error {
	cachePath := GetCachePath()
	log.Debug().Str("path", cachePath).Msg("Saving registry cache")
	
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}
	
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
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