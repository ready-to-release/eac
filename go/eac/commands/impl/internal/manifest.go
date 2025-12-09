// Package internal provides build manifest functionality
package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ModuleManifest represents the build manifest for a single module.
// This is stored per-module at out/build/<module>/build.manifest.json and is immutable after creation.
// The VerifiedUnchangedAt field can be updated by the builder when it verifies the module is up-to-date.
type ModuleManifest struct {
	BuildID             string         `json:"build_id"`                        // Unique identifier for this build (UUID locally, GitHub run ID in CI)
	BuildAgent          string         `json:"build_agent"`                     // Build agent type: "ci" or "devbox"
	Moniker             string         `json:"moniker"`                         // Module identifier
	Type                string         `json:"type"`                            // Module type
	BuildTime           time.Time      `json:"build_time"`                      // When this module was built
	GitCommit           string         `json:"git_commit,omitempty"`            // Git commit SHA at build time
	InputHash           string         `json:"input_hash,omitempty"`            // SHA-256 hash of source files at build time
	RequestedArtifacts  []string       `json:"requested_artifacts,omitempty"`   // Artifact IDs requested to build
	Artifacts           []ArtifactInfo `json:"artifacts"`                       // Artifacts actually produced
	Files               []string       `json:"files,omitempty"`                 // All files in build output (relative paths)
	Platforms           []PlatformInfo `json:"platforms"`                       // Platforms this was built for
	VerifiedUnchangedAt string         `json:"verified_unchanged_at,omitempty"` // Git SHA when builder verified module unchanged
	Version             string         `json:"version"`                         // Manifest format version
}

// ArtifactInfo describes a single built artifact
type ArtifactInfo struct {
	Type     string   `json:"type"`               // Artifact type (executable, file, directory, image)
	ID       string   `json:"id"`                 // Artifact identifier
	Name     string   `json:"name"`               // Resolved artifact name or image reference
	Path     string   `json:"path"`               // Relative path from build root, or image reference for type=image
	Platform string   `json:"platform,omitempty"` // Platform (e.g., "windows-amd64", "linux/amd64") if applicable
	Size     int64    `json:"size,omitempty"`     // File size in bytes (for file-based artifacts)
	SHA256   string   `json:"sha256,omitempty"`   // SHA-256 hash of artifact content (for file-based artifacts)
	Digest   string   `json:"digest,omitempty"`   // Image digest (for type=image, e.g., "sha256:abc123...")
	Tags     []string `json:"tags,omitempty"`     // Image tags (for type=image)
	Registry string   `json:"registry,omitempty"` // Container registry (for type=image, e.g., "ghcr.io")
}

// PlatformInfo describes a platform that was built
type PlatformInfo struct {
	OS   string `json:"os"`   // Operating system (windows, linux, darwin)
	Arch string `json:"arch"` // Architecture (amd64, arm64)
}

const manifestVersion = "2.0"
const manifestFileName = "build.manifest.json"

// CollectBuildFiles walks the build directory and returns all file paths relative to the directory.
// It excludes the manifest file itself, build logs, and intermediate build artifacts like staging directories.
func CollectBuildFiles(buildDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path first (needed for directory skip logic)
		relPath, err := filepath.Rel(buildDir, path)
		if err != nil {
			return err
		}

		// Skip hidden directories (like .staging) and their contents entirely
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip manifest file and build log
		if relPath == manifestFileName || relPath == "build.log" {
			return nil
		}

		// Convert to forward slashes for consistency
		relPath = filepath.ToSlash(relPath)
		files = append(files, relPath)
		return nil
	})

	return files, err
}

// BuildAgentCI is the build agent value for CI builds (GitHub Actions)
const BuildAgentCI = "ci"

// BuildAgentDevbox is the build agent value for local developer builds
const BuildAgentDevbox = "devbox"

// NewModuleManifest creates a new module manifest.
// In CI (GITHUB_RUN_ID set), uses the run ID as BuildID and sets BuildAgent to "ci".
// Locally, generates a UUID for BuildID and sets BuildAgent to "devbox".
func NewModuleManifest(moniker, moduleType, gitCommit string) *ModuleManifest {
	buildID := uuid.New().String()
	buildAgent := BuildAgentDevbox

	// In CI, use GitHub run ID as build identifier for traceability
	if runID := os.Getenv("GITHUB_RUN_ID"); runID != "" {
		buildID = runID
		buildAgent = BuildAgentCI
	}

	return &ModuleManifest{
		BuildID:    buildID,
		BuildAgent: buildAgent,
		Moniker:    moniker,
		Type:       moduleType,
		BuildTime:  time.Now(),
		GitCommit:  gitCommit,
		Version:    manifestVersion,
	}
}

// Save writes the manifest to the module's build output directory.
// The manifest is stored at <moduleBuildDir>/build.manifest.json
func (m *ModuleManifest) Save(moduleBuildDir string) error {
	manifestPath := filepath.Join(moduleBuildDir, manifestFileName)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(moduleBuildDir, 0755); err != nil {
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

// LoadModuleManifest loads a module's manifest from its build output directory
func LoadModuleManifest(moduleBuildDir string) (*ModuleManifest, error) {
	manifestPath := filepath.Join(moduleBuildDir, manifestFileName)

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
	var manifest ModuleManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// GetRequestedArtifacts returns the list of artifact IDs that were requested to be built
func (m *ModuleManifest) GetRequestedArtifacts() []string {
	return m.RequestedArtifacts
}

// GetPlatforms returns the platforms this module was built for
func (m *ModuleManifest) GetPlatforms() []PlatformInfo {
	return m.Platforms
}

// UpdateVerifiedUnchangedAt updates the verification timestamp and saves the manifest.
// This is the only field that can be updated after initial creation - it records when
// the builder verified the module was unchanged and didn't need rebuilding.
func (m *ModuleManifest) UpdateVerifiedUnchangedAt(moduleBuildDir, gitCommit string) error {
	m.VerifiedUnchangedAt = gitCommit
	return m.Save(moduleBuildDir)
}

// HashArtifactFile computes the SHA-256 hash and size of an artifact file.
// Returns (size, sha256hex, error). For directories or non-existent files,
// returns (0, "", nil) - the caller should handle these cases appropriately.
func HashArtifactFile(path string) (int64, string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, "", nil // File doesn't exist, not an error
		}
		return 0, "", err
	}

	// Skip directories - they don't have content hashes
	if info.IsDir() {
		return 0, "", nil
	}

	f, err := os.Open(path)
	if err != nil {
		return 0, "", err
	}
	defer f.Close()

	h := sha256.New()
	size, err := io.Copy(h, f)
	if err != nil {
		return 0, "", err
	}

	return size, hex.EncodeToString(h.Sum(nil)), nil
}

// =============================================================================
// Legacy support - BuildManifest (deprecated, for backward compatibility)
// =============================================================================

// BuildManifest represents the old global build manifest (deprecated).
// Kept for backward compatibility during migration.
// New code should use ModuleManifest instead.
type BuildManifest struct {
	BuildID   string                 `json:"build_id"`
	Timestamp time.Time              `json:"timestamp"`
	GitCommit string                 `json:"git_commit,omitempty"`
	Modules   map[string]ModuleBuild `json:"modules"`
	Version   string                 `json:"version"`
}

// ModuleBuild represents a single module's build information (legacy)
type ModuleBuild struct {
	Moniker            string         `json:"moniker"`
	Type               string         `json:"type"`
	BuildTime          time.Time      `json:"build_time"`
	RequestedArtifacts []string       `json:"requested_artifacts,omitempty"`
	Artifacts          []ArtifactInfo `json:"artifacts"`
	Platforms          []PlatformInfo `json:"platforms"`
}

// NewBuildManifest creates a new build manifest (legacy)
func NewBuildManifest(gitCommit string) *BuildManifest {
	return &BuildManifest{
		BuildID:   uuid.New().String(),
		Timestamp: time.Now(),
		GitCommit: gitCommit,
		Modules:   make(map[string]ModuleBuild),
		Version:   manifestVersion,
	}
}

// AddModule adds a module build to the manifest (legacy)
func (m *BuildManifest) AddModule(module ModuleBuild) {
	m.Modules[module.Moniker] = module
}

// Save writes the manifest to the build output directory (legacy)
func (m *BuildManifest) Save(buildOutputDir string) error {
	manifestPath := filepath.Join(buildOutputDir, manifestFileName)

	if err := os.MkdirAll(buildOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// LoadManifest loads the build manifest from the build output directory (legacy)
func LoadManifest(buildOutputDir string) (*BuildManifest, error) {
	manifestPath := filepath.Join(buildOutputDir, manifestFileName)

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest not found at %s", manifestPath)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest BuildManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// GetPlatformsForModule returns the platforms a module was built for (legacy)
func (m *BuildManifest) GetPlatformsForModule(moniker string) []PlatformInfo {
	if module, ok := m.Modules[moniker]; ok {
		return module.Platforms
	}
	return nil
}

// HasModule checks if a module was built in this manifest (legacy)
func (m *BuildManifest) HasModule(moniker string) bool {
	_, ok := m.Modules[moniker]
	return ok
}

// GetRequestedArtifacts returns the list of artifact IDs that were requested to be built (legacy)
func (m *BuildManifest) GetRequestedArtifacts(moniker string) []string {
	if module, ok := m.Modules[moniker]; ok {
		return module.RequestedArtifacts
	}
	return []string{}
}
