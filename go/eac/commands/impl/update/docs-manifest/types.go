// Package docsmanifest implements the update docs-manifest command.
package docsmanifest

import "time"

// DescriptionsFile represents the YAML file containing human-maintained asset descriptions.
// This file is git-tracked and contains static metadata that doesn't change automatically.
type DescriptionsFile struct {
	Assets map[string]Description `yaml:"assets"`
}

// Description represents human-maintained metadata for an asset.
type Description struct {
	Active      bool   `yaml:"active"`
	Description string `yaml:"description"`
}

// CacheManifest represents the auto-generated cache file with dynamic data.
// This file is git-ignored and regenerated on each update.
type CacheManifest struct {
	Generated string                `json:"generated"`
	Assets    map[string]CacheAsset `json:"assets"`
	Stats     *ManifestStats        `json:"stats,omitempty"`
}

// CacheAsset represents auto-generated metadata for a single asset in the cache.
type CacheAsset struct {
	UsedIn       []string `json:"usedIn"`
	Type         string   `json:"type,omitempty"`
	Category     string   `json:"category,omitempty"`
	LastModified string   `json:"lastModified,omitempty"`
	Size         int64    `json:"size,omitempty"`
	Hash         string   `json:"hash,omitempty"`
}

// ManifestStats contains summary statistics about the assets.
type ManifestStats struct {
	Total      int                      `json:"total"`
	Used       int                      `json:"used"`
	Unused     int                      `json:"unused"`
	ByCategory map[string]CategoryStats `json:"byCategory,omitempty"`
}

// CategoryStats contains statistics for a single category.
type CategoryStats struct {
	Total  int `json:"total"`
	Used   int `json:"used"`
	Unused int `json:"unused"`
}

// DiscoveredAsset represents an asset found during scanning.
type DiscoveredAsset struct {
	RelPath      string    // Path relative to assets directory (forward slashes)
	AbsPath      string    // Absolute path on filesystem
	Category     string    // Subdirectory category (e.g., "architecture")
	Type         string    // Asset type: "drawio", "drawio-source", "image"
	Size         int64     // File size in bytes
	LastModified time.Time // Last modification time
	Hash         string    // SHA-256 content hash
}

// UpdateResult captures the changes made during a manifest update.
type UpdateResult struct {
	Added                []string // Asset paths that were added
	Removed              []string // Asset paths that were removed
	UpdatedUsedIn        []string // Asset paths with changed usedIn
	TotalAssets          int
	UsedAssets           int
	UnusedAssets         int
	NeedsDescriptionsWrite bool   // Whether descriptions.yml needs to be written
	NeedsCacheWrite      bool     // Whether cache needs to be written
	DescriptionsPath     string   // Path to descriptions.yml file
	CachePath            string   // Path to cache file
}

// Legacy types for migration support

// Manifest represents the legacy manifest format (for migration).
type Manifest struct {
	Schema    string           `json:"$schema"`
	Generated string           `json:"generated"`
	Generator string           `json:"generator"`
	Stats     *ManifestStats   `json:"stats,omitempty"`
	Assets    map[string]Asset `json:"assets"`
}

// Asset represents the legacy asset format in manifest.json (for migration).
type Asset struct {
	Description  string   `json:"description"`
	UsedIn       []string `json:"usedIn"`
	Type         string   `json:"type,omitempty"`
	Category     string   `json:"category,omitempty"`
	LastModified string   `json:"lastModified,omitempty"`
	Size         int64    `json:"size,omitempty"`
	Hash         string   `json:"hash,omitempty"`
}
