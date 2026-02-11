// Package internal provides shared infrastructure for GET and SHOW commands
package internal

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
)

// FormatArtifactStatus returns a human-readable status string for an artifact.
func FormatArtifactStatus(artifact ResolvedArtifact) string {
	if artifact.Exists {
		if artifact.MetadataOverride != "" {
			return fmt.Sprintf("✓ (Override: %s)", artifact.MetadataOverride)
		}
		return "✓"
	}

	if artifact.Error != "" {
		return fmt.Sprintf("✗ (%s)", artifact.Error)
	}

	return "✗"
}

// FormatArtifactSize returns a human-readable size string for an artifact.
func FormatArtifactSize(sizeBytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case sizeBytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(sizeBytes)/float64(GB))
	case sizeBytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(sizeBytes)/float64(MB))
	case sizeBytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(sizeBytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", sizeBytes)
	}
}

// isBookModule checks if a module type is a book module (uses mkdocs handler).
func isBookModule(moduleType string) bool {
	// Check by type name - container modules with mkdocs handler
	return moduleType == "container" || strings.Contains(moduleType, "mkdocs")
}

// isContainerModule checks if a module produces container images that are pushed to registry.
// For such modules, we don't download build artifacts - we pull from registry instead.
func isContainerModule(module *config.Module) bool {
	// Check if module has docker_build with push enabled
	dockerConfig := module.GetDockerBuildConfig()
	return dockerConfig != nil && dockerConfig.ShouldPush()
}

// expandBookArtifacts expands wildcard PDF patterns to specific book PDFs
// The first book in the module's books list is the default; others require --all flag.
func expandBookArtifacts(module *config.Module, artifacts []config.Artifact, cfg *config.EACConfig, buildAll bool) []config.Artifact {
	var expanded []config.Artifact

	for _, artifact := range artifacts {
		// Check if this is a PDF wildcard that needs expansion
		if artifact.Type == config.ArtifactTypeFile && artifact.Pattern == "*.pdf" {
			// Get books for this module
			books := cfg.GetBooksByModule(module.Moniker)

			// Expand to specific book PDFs
			for i, book := range books {
				// Skip non-default books (not first) unless --all flag is used
				isDefault := i == 0
				if !buildAll && !isDefault {
					continue
				}

				// Get the output mode (theme)
				output := book.GetOutput()

				// Parse theme from output (e.g., "pdf-dark" -> "dark")
				theme := "dark" // default
				if strings.HasPrefix(output, "pdf-") {
					theme = strings.TrimPrefix(output, "pdf-")
					// Handle "pdf-all" by creating artifacts for both themes
					if theme == "all" {
						for _, t := range []string{"dark", "light"} {
							pdfName := fmt.Sprintf("%s-%s.pdf", book.Name, t)
							expanded = append(expanded, config.Artifact{
								Type:    artifact.Type,
								Pattern: pdfName,
								ID:      fmt.Sprintf("%s-%s", book.Name, t),
								Verify:  artifact.Verify,
							})
						}
						continue
					}
				}

				// Create artifact for this book's PDF
				pdfName := fmt.Sprintf("%s-%s.pdf", book.Name, theme)
				expanded = append(expanded, config.Artifact{
					Type:    artifact.Type,
					Pattern: pdfName,
					ID:      book.Name,
					Verify:  artifact.Verify,
				})
			}
		} else {
			// Keep non-PDF artifacts as-is
			expanded = append(expanded, artifact)
		}
	}

	return expanded
}

// extractRequestedArtifactsFromModuleView extracts artifact IDs from a ModuleView.
// Note: UoW manifests track actual artifacts produced, not "requested" artifacts.
// This returns the IDs of all artifacts that were actually produced.
// Callers should treat empty result as "validate all artifacts".
func extractRequestedArtifactsFromModuleView(view *coreoutput.ModuleView) []string {
	if view == nil {
		return nil
	}

	var ids []string
	for _, comp := range view.Components {
		for _, uow := range comp.UoWs {
			for _, art := range uow.Artifacts {
				if art.ID != "" {
					ids = append(ids, art.ID)
				}
			}
		}
	}
	return ids
}

// PlatformInfo describes a platform that was built.
type PlatformInfo struct {
	OS   string `json:"os"`   // Operating system (windows, linux, darwin)
	Arch string `json:"arch"` // Architecture (amd64, arm64)
}

// ArtifactInfo describes a single built artifact.
type ArtifactInfo struct {
	Type     string   `json:"type"`               // Artifact type (executable, file, directory, image)
	ID       string   `json:"id"`                 // Artifact identifier
	Name     string   `json:"name"`               // Resolved artifact name or image reference
	Path     string   `json:"path"`               // Relative path from build root, or image reference for type=image
	Platform string   `json:"platform,omitempty"` // Platform (e.g., "windows-amd64", "linux/amd64") if applicable
	Size     int64    `json:"size,omitempty"`      // File size in bytes (for file-based artifacts)
	SHA256   string   `json:"sha256,omitempty"`    // SHA-256 hash of artifact content (for file-based artifacts)
	Digest   string   `json:"digest,omitempty"`    // Image digest (for type=image, e.g., "sha256:abc123...")
	Tags     []string `json:"tags,omitempty"`      // Image tags (for type=image)
	Registry string   `json:"registry,omitempty"` // Container registry (for type=image, e.g., "ghcr.io")
}

// extractPlatformsFromModuleView extracts platform info from artifact paths in a ModuleView.
func extractPlatformsFromModuleView(view *coreoutput.ModuleView) []PlatformInfo {
	if view == nil {
		return nil
	}

	seen := make(map[string]bool)
	var platforms []PlatformInfo

	for _, comp := range view.Components {
		for _, uow := range comp.UoWs {
			for _, art := range uow.Artifacts {
				platform := inferPlatformFromPath(art.Path)
				if platform == "" || seen[platform] {
					continue
				}
				seen[platform] = true

				os, arch := parsePlatformString(platform)
				if os != "" && arch != "" {
					platforms = append(platforms, PlatformInfo{OS: os, Arch: arch})
				}
			}
		}
	}

	return platforms
}

// inferPlatformFromPath tries to extract platform info from artifact path.
// Common patterns: "eac-linux-amd64", "app-darwin-arm64.exe", "linux/amd64"
func inferPlatformFromPath(path string) string {
	patterns := []struct {
		pattern  string
		platform string
	}{
		{"linux-amd64", "linux-amd64"},
		{"linux-arm64", "linux-arm64"},
		{"darwin-amd64", "darwin-amd64"},
		{"darwin-arm64", "darwin-arm64"},
		{"windows-amd64", "windows-amd64"},
		{"windows-arm64", "windows-arm64"},
		{"linux/amd64", "linux-amd64"},
		{"linux/arm64", "linux-arm64"},
	}

	lowerPath := strings.ToLower(path)
	for _, p := range patterns {
		if strings.Contains(lowerPath, p.pattern) {
			return p.platform
		}
	}

	return ""
}

// parsePlatformString parses "os-arch" or "os/arch" format into separate OS and Arch.
func parsePlatformString(platform string) (os, arch string) {
	for i := len(platform) - 1; i >= 0; i-- {
		if platform[i] == '-' || platform[i] == '/' {
			return platform[:i], platform[i+1:]
		}
	}
	return "", ""
}
