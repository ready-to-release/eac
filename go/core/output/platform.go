package output

import (
	"strings"
)

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

// ExtractRequestedArtifactsFromModuleView extracts artifact IDs from a ModuleView.
// Note: UoW manifests track actual artifacts produced, not "requested" artifacts.
// This returns the IDs of all artifacts that were actually produced.
// Callers should treat empty result as "validate all artifacts".
func ExtractRequestedArtifactsFromModuleView(view *ModuleView) []string {
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

// ExtractPlatformsFromModuleView extracts platform info from artifact paths in a ModuleView.
func ExtractPlatformsFromModuleView(view *ModuleView) []PlatformInfo {
	if view == nil {
		return nil
	}

	seen := make(map[string]bool)
	var platforms []PlatformInfo

	for _, comp := range view.Components {
		for _, uow := range comp.UoWs {
			for _, art := range uow.Artifacts {
				platform := InferPlatformFromPath(art.Path)
				if platform == "" || seen[platform] {
					continue
				}
				seen[platform] = true

				os, arch := ParsePlatformString(platform)
				if os != "" && arch != "" {
					platforms = append(platforms, PlatformInfo{OS: os, Arch: arch})
				}
			}
		}
	}

	return platforms
}

// knownOS and knownArch define recognized platform components for inference.
var (
	knownOS   = []string{"linux", "darwin", "windows", "freebsd"}
	knownArch = []string{"amd64", "arm64", "arm", "386", "s390x", "ppc64le", "riscv64"}
)

// InferPlatformFromPath tries to extract platform info from artifact path.
// Supports patterns like "eac-linux-amd64", "app-darwin-arm64.exe", "linux/amd64".
// Uses a combinatorial approach from known OS/architecture pairs rather than a
// hardcoded pattern list, so new platforms are automatically supported.
func InferPlatformFromPath(path string) string {
	lowerPath := strings.ToLower(path)

	for _, os := range knownOS {
		for _, arch := range knownArch {
			// Check "os-arch" (filename convention) and "os/arch" (Docker convention)
			if strings.Contains(lowerPath, os+"-"+arch) || strings.Contains(lowerPath, os+"/"+arch) {
				return os + "-" + arch
			}
		}
	}

	return ""
}

// ParsePlatformString parses "os-arch" or "os/arch" format into separate OS and Arch.
func ParsePlatformString(platform string) (os, arch string) {
	for i := len(platform) - 1; i >= 0; i-- {
		if platform[i] == '-' || platform[i] == '/' {
			return platform[:i], platform[i+1:]
		}
	}
	return "", ""
}
