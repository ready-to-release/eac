// Package internal provides build manifest functionality
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// BuildManifest represents the complete build manifest
type BuildManifest struct {
	BuildID    string                `json:"build_id"`              // Unique identifier for this build event
	Timestamp  time.Time             `json:"timestamp"`             // When this build was created
	GitCommit  string                `json:"git_commit,omitempty"`  // Git commit SHA at build time
	Modules    map[string]ModuleBuild `json:"modules"`               // Modules built in this event
	Version    string                `json:"version"`               // Manifest format version
}

// ModuleBuild represents a single module's build information
type ModuleBuild struct {
	Moniker            string              `json:"moniker"`                       // Module identifier
	Type               string              `json:"type"`                          // Module type
	BuildTime          time.Time           `json:"build_time"`                    // When this module was built
	RequestedArtifacts []string            `json:"requested_artifacts,omitempty"` // Artifact IDs requested to build
	Artifacts          []ArtifactInfo      `json:"artifacts"`                     // Artifacts actually produced
	Platforms          []PlatformInfo      `json:"platforms"`                     // Platforms this was built for
}

// ArtifactInfo describes a single built artifact
type ArtifactInfo struct {
	Type         string `json:"type"`                   // Artifact type (executable, file, etc.)
	ID           string `json:"id"`                     // Artifact identifier
	Name         string `json:"name"`                   // Resolved artifact name
	Path         string `json:"path"`                   // Relative path from build root
	Platform     string `json:"platform,omitempty"`     // Platform (e.g., "windows-amd64") if applicable
}

// PlatformInfo describes a platform that was built
type PlatformInfo struct {
	OS   string `json:"os"`   // Operating system (windows, linux, darwin)
	Arch string `json:"arch"` // Architecture (amd64, arm64)
}

const manifestVersion = "1.0"
const manifestFileName = ".manifest.json"

// NewBuildManifest creates a new build manifest
func NewBuildManifest(gitCommit string) *BuildManifest {
	return &BuildManifest{
		BuildID:   uuid.New().String(),
		Timestamp: time.Now(),
		GitCommit: gitCommit,
		Modules:   make(map[string]ModuleBuild),
		Version:   manifestVersion,
	}
}

// AddModule adds a module build to the manifest
func (m *BuildManifest) AddModule(module ModuleBuild) {
	m.Modules[module.Moniker] = module
}

// Save writes the manifest to the build output directory
func (m *BuildManifest) Save(buildOutputDir string) error {
	manifestPath := filepath.Join(buildOutputDir, manifestFileName)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(buildOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write to file
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// LoadManifest loads the build manifest from the build output directory
func LoadManifest(buildOutputDir string) (*BuildManifest, error) {
	manifestPath := filepath.Join(buildOutputDir, manifestFileName)

	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest not found at %s", manifestPath)
	}

	// Read file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	// Unmarshal
	var manifest BuildManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// GetPlatformsForModule returns the platforms a module was built for
func (m *BuildManifest) GetPlatformsForModule(moniker string) []PlatformInfo {
	if module, ok := m.Modules[moniker]; ok {
		return module.Platforms
	}
	return nil
}

// HasModule checks if a module was built in this manifest
func (m *BuildManifest) HasModule(moniker string) bool {
	_, ok := m.Modules[moniker]
	return ok
}

// GetRequestedArtifacts returns the list of artifact IDs that were requested to be built
// Returns empty slice if module not found or no artifacts requested (backward compatibility)
func (m *BuildManifest) GetRequestedArtifacts(moniker string) []string {
	if module, ok := m.Modules[moniker]; ok {
		return module.RequestedArtifacts
	}
	return []string{}
}
