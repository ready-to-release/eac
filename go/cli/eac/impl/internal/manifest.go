// Package internal provides build manifest functionality
package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ModuleManifest represents the build manifest for a single module.
// This is used as a display/view format aggregated from UoW manifests
// via GetModuleManifestFromUoWs() in manifest_adapter.go.
type ModuleManifest struct {
	BuildID             string         `json:"build_id"`                        // Unique identifier for this build (UUID locally, GitHub run ID in CI)
	BuildAgent          string         `json:"build_agent"`                     // Build agent type: "ci" or "devbox"
	Moniker             string         `json:"moniker"`                         // Module identifier
	Type                string         `json:"type"`                            // Module type
	BuildTime           time.Time      `json:"build_time"`                      // When this module was built
	DurationSeconds     float64        `json:"duration_seconds,omitempty"`      // Build duration in seconds
	GitCommit           string         `json:"git_commit,omitempty"`            // Git commit SHA at build time
	InputHash           string         `json:"input_hash,omitempty"`            // SHA-256 hash of source files at build time
	RequestedArtifacts  []string       `json:"requested_artifacts,omitempty"`   // Artifact IDs requested to build
	Artifacts           []ArtifactInfo `json:"artifacts"`                       // Artifacts actually produced
	Files               []string       `json:"files,omitempty"`                 // All files in build output (relative paths)
	Platforms           []PlatformInfo `json:"platforms"`                       // Platforms this was built for
	VerifiedUnchangedAt string         `json:"verified_unchanged_at,omitempty"` // Git SHA when builder verified module unchanged
	Version             string         `json:"version"`                         // Manifest format version
}

// ArtifactInfo describes a single built artifact.
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

// PlatformInfo describes a platform that was built.
type PlatformInfo struct {
	OS   string `json:"os"`   // Operating system (windows, linux, darwin)
	Arch string `json:"arch"` // Architecture (amd64, arm64)
}

const (
	manifestVersion  = "2.0"
	manifestFileName = "build.manifest.json"
)

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

// BuildAgentCI is the build agent value for CI builds (GitHub Actions).
const BuildAgentCI = "ci"

// BuildAgentDevbox is the build agent value for local developer builds.
const BuildAgentDevbox = "devbox"

// GetRequestedArtifacts returns the list of artifact IDs that were requested to be built.
func (m *ModuleManifest) GetRequestedArtifacts() []string {
	return m.RequestedArtifacts
}

// GetPlatforms returns the platforms this module was built for.
func (m *ModuleManifest) GetPlatforms() []PlatformInfo {
	return m.Platforms
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
