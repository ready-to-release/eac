// Command: show artifacts
// Description: Show resolved artifacts for a module in human-readable format
// Short: Display artifacts for a module with status
// Long: The show artifacts command displays all build artifacts for a module in a formatted table.
// Long: Shows resolved names with metadata overrides applied, existence status, and file paths.
// Long: By default shows artifacts for the current platform. Use --all-platforms to see all platforms.
// Long:
// Long: Expected Output:
// Long: - Formatted table with artifact names, existence status (checkmark or X), and resolved file paths
// Long: - Summary header showing module name, type, build directory, and platform
// Long: - Breakdown of build modes (default vs --all)
// Long: - Metadata overrides section if any metadata is defined
package show

import (
	"fmt"
	"os"
	"runtime"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	showinternal "github.com/ready-to-release/eac/go/eac/commands/impl/show/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowArtifacts)
}

// ShowArtifacts displays resolved artifacts for a module
func ShowArtifacts() int {
	args := os.Args[3:] // Skip program name, "show", and "artifacts"

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: show artifacts <module> [--all-platforms] [--missing-only]")
		return 1
	}

	moduleName := args[0]
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH
	allPlatforms := false
	missingOnly := false

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--all-platforms":
			allPlatforms = true
		case "--missing-only":
			missingOnly = true
		case "--os":
			if i+1 < len(args) {
				targetOS = args[i+1]
				i++
			}
		case "--arch":
			if i+1 < len(args) {
				targetArch = args[i+1]
				i++
			}
		}
	}

	return showArtifactsForModule(moduleName, targetOS, targetArch, allPlatforms, missingOnly)
}

func showArtifactsForModule(moduleName, targetOS, targetArch string, allPlatforms, missingOnly bool) int {
	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Load configuration
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		return 1
	}

	// Get module
	module, ok := cfg.Repository.GetModule(moduleName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: module not found: %s\n", moduleName)
		return 1
	}

	// Get module type
	moduleType := cfg.ModuleTypes.Get(module.Type)
	if moduleType == nil {
		fmt.Fprintf(os.Stderr, "Error: module type not found: %s\n", module.Type)
		return 1
	}

	if moduleType.Build == nil || len(moduleType.Build.Artifacts) == 0 {
		log.Debugf("No artifacts defined for module type %s", module.Type)
		return 0
	}

	// Build directory
	buildDir := cfg.Repository.BuildOutputPathAbs(workspaceRoot, moduleName)

	// Resolve artifacts
	var results []implinternal.ResolvedArtifact
	var summary *implinternal.ArtifactResolutionSummary

	if allPlatforms {
		// Get all platform combinations (same logic as GET artifacts)
		platforms := getPlatformsForModule(moduleType)
		summary = &implinternal.ArtifactResolutionSummary{}

		// Use a map to deduplicate by resolved path
		seenPaths := make(map[string]bool)

		for _, plat := range platforms {
			platformResults, _, err := implinternal.ResolveArtifactsForModuleWithConfig(
				module, moduleType, buildDir, plat.OS, plat.Arch, cfg,
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to resolve artifacts for %s/%s: %v\n", plat.OS, plat.Arch, err)
				continue
			}

			// Deduplicate by resolved path
			for _, art := range platformResults {
				if !seenPaths[art.ResolvedPath] {
					seenPaths[art.ResolvedPath] = true
					results = append(results, art)
					summary.Total++
					if art.Exists {
						summary.Exists++
					} else {
						summary.Missing++
					}
					if art.MetadataOverride != "" {
						summary.Overrides++
					}
				}
			}
		}
	} else {
		// Single platform
		var err error
		results, summary, err = implinternal.ResolveArtifactsForModuleWithConfig(
			module, moduleType, buildDir, targetOS, targetArch, cfg,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to resolve artifacts: %v\n", err)
			return 1
		}
	}

	// Filter if missing-only
	if missingOnly {
		var filtered []implinternal.ResolvedArtifact
		for _, art := range results {
			if !art.Exists {
				filtered = append(filtered, art)
			}
		}
		results = filtered
		if len(results) == 0 {
			fmt.Println("All artifacts exist")
			return 0
		}
	}

	// Format output
	var output string

	// Header
	platform := fmt.Sprintf("%s/%s", targetOS, targetArch)
	if allPlatforms {
		platform = "all platforms"
	}
	metadataCount := countArtifactMetadata(module.Metadata)
	output += showinternal.FormatArtifactSummaryHeader(moduleName, module.Type, buildDir, platform, metadataCount)

	// Show default vs all artifact breakdown
	if !missingOnly {
		output += "\n"
		output += formatArtifactModeBreakdown(module, moduleType, cfg)
	}

	// Table
	output += showinternal.FormatArtifactTable(results, summary, allPlatforms)

	// Metadata overrides
	if metadataCount > 0 {
		output += "\n"
		output += showinternal.FormatMetadataOverrides(module.Metadata, results)
	}

	// Help text
	if !allPlatforms {
		output += "\n\nUse --all-platforms to see all platforms\n"
	}
	if !missingOnly && summary.Missing > 0 {
		output += "Use --missing-only to show only missing artifacts\n"
	}

	fmt.Print(output)
	return 0
}

// formatArtifactModeBreakdown shows which artifacts are built in default vs --all mode
func formatArtifactModeBreakdown(module *config.Module, moduleType *config.ModuleTypeDef, cfg *config.EACConfig) string {
	// Get default and all artifact IDs
	defaultArtifacts := implinternal.DetermineRequestedArtifacts(module, moduleType, false, cfg)
	allArtifacts := implinternal.DetermineRequestedArtifacts(module, moduleType, true, cfg)

	var output string
	output += "Build Modes:\n"
	output += fmt.Sprintf("  Default (current platform): %v\n", defaultArtifacts)
	output += fmt.Sprintf("  --all (all platforms):      %v\n", allArtifacts)

	return output
}

// countArtifactMetadata counts how many metadata entries are artifact-related
func countArtifactMetadata(metadata map[string]string) int {
	count := 0
	artifactPrefixes := []string{"executable-", "file-", "directory-", "image-", "marker-"}

	for key := range metadata {
		for _, prefix := range artifactPrefixes {
			if len(key) > len(prefix) && key[:len(prefix)] == prefix {
				count++
				break
			}
		}
	}

	return count
}

type platformInfo struct {
	OS   string
	Arch string
}

// getPlatformsForModule returns all platform combinations for a module type
func getPlatformsForModule(moduleType *config.ModuleTypeDef) []platformInfo {
	platformSet := make(map[string]bool)
	var platforms []platformInfo

	if moduleType.Build == nil {
		return []platformInfo{{OS: runtime.GOOS, Arch: runtime.GOARCH}}
	}

	// Extract unique platform combinations from artifacts
	for _, artifact := range moduleType.Build.Artifacts {
		if artifact.Type == config.ArtifactTypeExecutable && len(artifact.Platforms) > 0 {
			// Executables specify platforms
			for _, os := range artifact.Platforms {
				// Add both amd64 and arm64 for each OS
				if os == "windows" {
					key := fmt.Sprintf("%s-amd64", os)
					if !platformSet[key] {
						platformSet[key] = true
						platforms = append(platforms, platformInfo{OS: os, Arch: "amd64"})
					}
				} else {
					// linux and darwin support both amd64 and arm64
					for _, arch := range []string{"amd64", "arm64"} {
						key := fmt.Sprintf("%s-%s", os, arch)
						if !platformSet[key] {
							platformSet[key] = true
							platforms = append(platforms, platformInfo{OS: os, Arch: arch})
						}
					}
				}
			}
		}
	}

	// If no platforms found, default to current
	if len(platforms) == 0 {
		platforms = append(platforms, platformInfo{OS: runtime.GOOS, Arch: runtime.GOARCH})
	}

	return platforms
}
