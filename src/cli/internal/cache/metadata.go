package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

// MetadataCache holds cached extension metadata with image digest for invalidation
type MetadataCache struct {
	ExtensionName string         `json:"extension_name"`
	ImageDigest   string         `json:"image_digest"` // Docker image digest for cache invalidation
	Metadata      *ExtensionMeta `json:"metadata"`
	CachedAt      time.Time      `json:"cached_at"`
}

// ExtensionMeta represents the parsed extension metadata from extension-meta command
type ExtensionMeta struct {
	Name               string            `json:"name" yaml:"name"`
	Version            string            `json:"version" yaml:"version"`
	Description        string            `json:"description" yaml:"description"`
	SchemaVersion      string            `json:"schema_version" yaml:"schema-version"`
	Capabilities       []string          `json:"capabilities" yaml:"capabilities"`
	Volumes            []VolumeRequest   `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Requirements       *MetaRequirements `json:"requirements,omitempty" yaml:"requirements,omitempty"`
	ExpectedHostImages []string          `json:"expected_host_images,omitempty" yaml:"expected-host-images,omitempty"`
}

// VolumeRequest represents a volume mount requested by an extension
type VolumeRequest struct {
	Name   string `json:"name" yaml:"name"`     // Logical name (used for Docker volume naming)
	Target string `json:"target" yaml:"target"` // Container path to mount
	Type   string `json:"type" yaml:"type"`     // "cache" for named volumes, "bind" for bind mounts
}

// MetaRequirements represents extension requirements from metadata
type MetaRequirements struct {
	CLIVersion       string `json:"cli_version,omitempty" yaml:"r2r-version,omitempty"`
	ContainerRuntime string `json:"container_runtime,omitempty" yaml:"container-runtime,omitempty"`
	MinimumMemory    string `json:"minimum_memory,omitempty" yaml:"minimum-memory,omitempty"`
	MinimumCPU       string `json:"minimum_cpu,omitempty" yaml:"minimum-cpu,omitempty"`
}

// GetMetadataCacheDir returns the path to the metadata cache directory (.r2r/metadata/)
func GetMetadataCacheDir(repoRoot string) string {
	return filepath.Join(repoRoot, ".r2r", "metadata")
}

// GetMetadataCachePath returns the path to a specific extension's metadata cache file
func GetMetadataCachePath(repoRoot, extensionName string) string {
	cacheDir := GetMetadataCacheDir(repoRoot)
	return filepath.Join(cacheDir, fmt.Sprintf("%s.json", extensionName))
}

// LoadMetadataCache loads cached metadata for an extension from disk
func LoadMetadataCache(repoRoot, extensionName string) (*MetadataCache, error) {
	cachePath := GetMetadataCachePath(repoRoot, extensionName)
	log.Debug().
		Str("extension", extensionName).
		Str("path", cachePath).
		Msg("Loading metadata cache")

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debug().
				Str("extension", extensionName).
				Msg("Metadata cache not found")
			return nil, nil // Cache miss, not an error
		}
		return nil, fmt.Errorf("failed to read metadata cache: %w", err)
	}

	var cache MetadataCache
	if err := json.Unmarshal(data, &cache); err != nil {
		log.Warn().
			Err(err).
			Str("extension", extensionName).
			Msg("Failed to parse metadata cache, will refresh")
		return nil, nil // Treat parse errors as cache miss
	}

	log.Debug().
		Str("extension", extensionName).
		Str("digest", cache.ImageDigest).
		Time("cached_at", cache.CachedAt).
		Msg("Loaded metadata cache")

	return &cache, nil
}

// SaveMetadataCache saves extension metadata to the cache
func SaveMetadataCache(repoRoot string, cache *MetadataCache) error {
	cacheDir := GetMetadataCacheDir(repoRoot)

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata cache directory: %w", err)
	}

	cachePath := GetMetadataCachePath(repoRoot, cache.ExtensionName)

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata cache: %w", err)
	}

	// Write atomically by writing to temp file then renaming
	tmpPath := cachePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata cache: %w", err)
	}

	if err := os.Rename(tmpPath, cachePath); err != nil {
		os.Remove(tmpPath) // Clean up temp file on failure
		return fmt.Errorf("failed to finalize metadata cache: %w", err)
	}

	log.Debug().
		Str("extension", cache.ExtensionName).
		Str("path", cachePath).
		Str("digest", cache.ImageDigest).
		Msg("Saved metadata cache")

	return nil
}

// IsMetadataCacheValid checks if the cached metadata is still valid for the given image digest
func IsMetadataCacheValid(cache *MetadataCache, currentDigest string) bool {
	if cache == nil || cache.Metadata == nil {
		return false
	}

	// Cache is valid if the image digest matches
	if cache.ImageDigest != currentDigest {
		log.Debug().
			Str("extension", cache.ExtensionName).
			Str("cached_digest", cache.ImageDigest).
			Str("current_digest", currentDigest).
			Msg("Metadata cache invalidated: image digest changed")
		return false
	}

	return true
}

// NewMetadataCache creates a new MetadataCache instance
func NewMetadataCache(extensionName, imageDigest string, metadata *ExtensionMeta) *MetadataCache {
	return &MetadataCache{
		ExtensionName: extensionName,
		ImageDigest:   imageDigest,
		Metadata:      metadata,
		CachedAt:      time.Now(),
	}
}
