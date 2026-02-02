// Package mkdocs provides MkDocs documentation build handlers.
//
// This package implements the base-site, site-render, and pdf-render component types
// as part of the MkDocs documentation architecture. The key design principle is
// decoupling preprocessing (expensive) from rendering (Docker-based).
//
// Component Types:
//   - base-site: Preprocessing only - produces staged markdown ready for MkDocs
//   - site-render: HTML generation - consumes base-site output, produces HTML site
//   - pdf-render: PDF generation - consumes base-site output, produces PDF files
package mkdocs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ComponentType identifies the type of MkDocs component.
type ComponentType string

const (
	// ComponentTypeBaseSite is preprocessing-only (no Docker required).
	ComponentTypeBaseSite ComponentType = "base-site"

	// ComponentTypeSiteRender is HTML generation (requires Docker).
	ComponentTypeSiteRender ComponentType = "site-render"

	// ComponentTypePDFRender is PDF generation (requires Docker).
	ComponentTypePDFRender ComponentType = "pdf-render"
)

// ManifestVersion is the current schema version for component manifests.
const ManifestVersion = "1.0"

// ComponentManifest tracks the state of a built component.
// This enables manifest-based caching where render components
// depend on base-site output hash for cache invalidation.
type ComponentManifest struct {
	// Version is the manifest schema version.
	Version string `json:"version"`

	// Type identifies the component type (base-site, site-render, pdf-render).
	Type ComponentType `json:"type"`

	// Module is the module moniker (e.g., "docs", "books").
	Module string `json:"module"`

	// Component is the component name within the module.
	Component string `json:"component"`

	// InputHash is the SHA256 hash of all inputs (sources, config, dependencies).
	InputHash string `json:"input_hash"`

	// OutputHash is the SHA256 hash of all outputs (staged content, artifacts).
	OutputHash string `json:"output_hash"`

	// Timestamp is when the build completed.
	Timestamp time.Time `json:"timestamp"`

	// Duration is how long the build took.
	Duration time.Duration `json:"duration"`

	// Dependencies maps dependency component names to their output hashes.
	// For site-render/pdf-render, this includes the base-site output hash.
	Dependencies map[string]string `json:"dependencies,omitempty"`

	// Artifacts lists the output file/directory paths relative to the output directory.
	Artifacts []string `json:"artifacts"`

	// Metadata holds additional type-specific data (e.g., theme for PDF).
	Metadata map[string]string `json:"metadata,omitempty"`
}

// BuildResult represents the outcome of a component build.
type BuildResult struct {
	// ExitCode is the process exit code (0 = success, non-zero = failure).
	ExitCode int

	// Cached indicates if the build was skipped due to cache hit.
	Cached bool

	// Manifest is the manifest written after a successful build.
	Manifest *ComponentManifest

	// Artifacts lists the produced artifact paths.
	Artifacts []string

	// Duration is how long the build took.
	Duration time.Duration
}

// ExitCodeSkipped is returned when a build is skipped due to cache hit.
// This is distinct from success (0) to signal the TUI for different display.
const ExitCodeSkipped = -1

// ManifestStore manages component manifests in the cache directory.
type ManifestStore struct {
	cacheDir string
}

// NewManifestStore creates a new manifest store.
// The cache directory is .cache/eac/components/ under the workspace root.
func NewManifestStore(workspaceRoot string) *ManifestStore {
	return &ManifestStore{
		cacheDir: filepath.Join(workspaceRoot, ".cache", "eac", "components"),
	}
}

// ManifestPath returns the path to the manifest file for a component.
func (s *ManifestStore) ManifestPath(module, component string) string {
	return filepath.Join(s.cacheDir, module, component, ".manifest.json")
}

// Load loads a manifest for a component.
// Returns nil, nil if the manifest doesn't exist.
func (s *ManifestStore) Load(module, component string) (*ComponentManifest, error) {
	path := s.ManifestPath(module, component)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var m ComponentManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Save saves a manifest for a component.
func (s *ManifestStore) Save(m *ComponentManifest) error {
	dir := filepath.Join(s.cacheDir, m.Module, m.Component)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	path := filepath.Join(dir, ".manifest.json")
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// CacheStats tracks cache hit/miss statistics for reporting.
type CacheStats struct {
	ManifestHits   int
	ManifestMisses int
	MermaidHits    int
	MermaidMisses  int
	DrawioHits     int
	DrawioMisses   int
}
