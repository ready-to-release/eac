// Package internal provides build manifest functionality
package internal

import (
	"os"
	"time"

	"github.com/google/uuid"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// GetModuleManifestFromUoWs aggregates UoW manifests into legacy ModuleManifest format.
// This provides backward compatibility during migration from per-module manifests
// to per-UoW (Unit of Work) manifests.
//
// workspaceRoot: Repository root path
// ctx: Operation context (build, test, lint, scan)
// module: Module moniker
// moduleType: Module type (e.g., "cli-app", "library")
//
// Returns nil and error if no UoW manifests exist for the module.
func GetModuleManifestFromUoWs(workspaceRoot string, ctx workunit.Context, module, moduleType string) (*ModuleManifest, error) {
	reader := coreoutput.NewReader(workspaceRoot)
	moduleView, err := reader.GetModule(ctx, module)
	if err != nil {
		return nil, err
	}

	// If no components, return nil (no manifest)
	if len(moduleView.Components) == 0 {
		return nil, nil
	}

	return ConvertModuleViewToManifest(moduleView, moduleType), nil
}

// ConvertModuleViewToManifest converts a ModuleView (from UoW aggregation) to legacy ModuleManifest.
// This enables callers that expect the old format to work with the new UoW-based system.
func ConvertModuleViewToManifest(view *coreoutput.ModuleView, moduleType string) *ModuleManifest {
	if view == nil {
		return nil
	}

	// Determine build agent
	buildAgent := BuildAgentDevbox
	if runID := os.Getenv("GITHUB_RUN_ID"); runID != "" {
		buildAgent = BuildAgentCI
	}

	// Compute total duration and find earliest execution time
	var totalDuration time.Duration
	var earliestTime time.Time
	var latestTime time.Time

	for _, comp := range view.Components {
		for _, uow := range comp.UoWs {
			totalDuration += uow.Duration
			if earliestTime.IsZero() || uow.ExecutedAt.Before(earliestTime) {
				earliestTime = uow.ExecutedAt
			}
			endTime := uow.ExecutedAt.Add(uow.Duration)
			if latestTime.IsZero() || endTime.After(latestTime) {
				latestTime = endTime
			}
		}
	}

	// Generate build ID
	buildID := uuid.New().String()
	if runID := os.Getenv("GITHUB_RUN_ID"); runID != "" {
		buildID = runID
	}

	// Build artifacts list from all UoWs
	var artifacts []ArtifactInfo
	for _, comp := range view.Components {
		for _, uow := range comp.UoWs {
			for _, artifact := range uow.Artifacts {
				// Convert UoW artifact to legacy ArtifactInfo
				info := ArtifactInfo{
					ID:     artifact.ID,
					Path:   artifact.Path,
					Type:   artifact.Type,
					Size:   artifact.Size,
					SHA256: artifact.SHA256,
				}
				// Try to extract platform from artifact ID or path
				platform := inferPlatformFromPath(artifact.Path)
				if platform != "" {
					info.Platform = platform
				}
				artifacts = append(artifacts, info)
			}
		}
	}

	// If no artifacts, still create a manifest but with empty artifacts
	if artifacts == nil {
		artifacts = []ArtifactInfo{}
	}

	// Build platforms list from artifacts
	platforms := extractPlatformsFromArtifacts(artifacts)

	// Get input hash from first UoW (they should all have the same source hash)
	var inputHash string
	for _, comp := range view.Components {
		for _, uow := range comp.UoWs {
			if uow.InputHash != "" {
				inputHash = uow.InputHash
				break
			}
		}
		if inputHash != "" {
			break
		}
	}

	return &ModuleManifest{
		BuildID:         buildID,
		BuildAgent:      buildAgent,
		Moniker:         view.Module,
		Type:            moduleType,
		BuildTime:       earliestTime,
		DurationSeconds: totalDuration.Seconds(),
		InputHash:       inputHash,
		Artifacts:       artifacts,
		Platforms:       platforms,
		Version:         manifestVersion,
	}
}

// inferPlatformFromPath tries to extract platform info from artifact path.
// Common patterns: "eac-linux-amd64", "app-darwin-arm64.exe", "linux/amd64"
func inferPlatformFromPath(path string) string {
	// Check for common OS-arch patterns
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

	for _, p := range patterns {
		if containsIgnoreCase(path, p.pattern) {
			return p.platform
		}
	}

	return ""
}

// containsIgnoreCase checks if s contains substr (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (contains(toLower(s), toLower(substr))))
}

// toLower is a simple lowercase helper.
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// contains is a simple substring check.
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// extractPlatformsFromArtifacts extracts unique platforms from artifact list.
func extractPlatformsFromArtifacts(artifacts []ArtifactInfo) []PlatformInfo {
	seen := make(map[string]bool)
	var platforms []PlatformInfo

	for _, art := range artifacts {
		if art.Platform == "" {
			continue
		}
		if seen[art.Platform] {
			continue
		}
		seen[art.Platform] = true

		// Parse platform string (e.g., "linux-amd64" -> {OS: "linux", Arch: "amd64"})
		os, arch := parsePlatformString(art.Platform)
		if os != "" && arch != "" {
			platforms = append(platforms, PlatformInfo{OS: os, Arch: arch})
		}
	}

	return platforms
}

// parsePlatformString parses "os-arch" or "os/arch" format into separate OS and Arch.
func parsePlatformString(platform string) (os, arch string) {
	// Try "-" separator first
	for i := len(platform) - 1; i >= 0; i-- {
		if platform[i] == '-' || platform[i] == '/' {
			return platform[:i], platform[i+1:]
		}
	}
	return "", ""
}
