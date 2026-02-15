package get

import (
	"context"
	"fmt"
	"os"
	"runtime"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/commands/base"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getArtifactsCommand struct{}

var _ core.SimpleCommandPort = (*getArtifactsCommand)(nil)

func (c *getArtifactsCommand) Name() string { return "get artifacts" }

func (c *getArtifactsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-artifacts",
		Short:         "Get resolved artifacts for a module",
		Long:          "The get artifacts command returns all build artifacts for a module with metadata overrides applied.\nOutput includes resolved names, paths, existence status, and which overrides were used.\nBy default, shows artifacts for the current platform. Use --all-platforms to see all platforms.\n\nExpected Output:\nYAML list of build artifacts with metadata, including:\n  - Resolved artifact names and paths\n  - Artifact existence status (exists/missing)\n  - Metadata override information showing which overrides were applied\n  - Build modes (default and all)\n  - Summary statistics (total, exists, missing, overrides)",
	}
}

func (c *getArtifactsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetArtifacts()
}

// artifactsFlags defines valid flags for the get artifacts command

// GetArtifacts returns resolved artifacts for a module.
func GetArtifacts() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "get", and "artifacts"

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: get artifacts <module> [--os <os>] [--arch <arch>] [--all-platforms]")
		return 1
	}

	moduleName := args[0]
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH
	allPlatforms := false

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--all-platforms":
			allPlatforms = true
		case arg == "--os" && i+1 < len(args):
			targetOS = args[i+1]
			i++
		case arg == "--arch" && i+1 < len(args):
			targetArch = args[i+1]
			i++
		}
	}

	return getArtifactsForModule(moduleName, targetOS, targetArch, allPlatforms)
}

func getArtifactsForModule(moduleName, targetOS, targetArch string, allPlatforms bool) int {
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

	// Build directory
	buildDir := paths.BuildOutputPath(workspaceRoot, moduleName)

	// Resolve artifacts
	var allResults []config.ResolvedArtifact
	var totalSummary *config.ArtifactResolutionSummary

	if allPlatforms {
		// Get all platform combinations
		platforms := getAllPlatformsFromModule(module)
		totalSummary = &config.ArtifactResolutionSummary{}

		// Use a map to deduplicate by resolved path
		seenPaths := make(map[string]bool)

		for _, plat := range platforms {
			results, _, err := config.ResolveArtifactsForModuleWithConfig(
				module, buildDir, plat.OS, plat.Arch, cfg,
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to resolve artifacts for %s/%s: %v\n", plat.OS, plat.Arch, err)
				continue
			}

			// Deduplicate by resolved path
			for i := range results {
				art := &results[i]
				if !seenPaths[art.ResolvedPath] {
					seenPaths[art.ResolvedPath] = true
					allResults = append(allResults, *art)
					totalSummary.Total++
					if art.Exists {
						totalSummary.Exists++
					} else {
						totalSummary.Missing++
					}
					if art.MetadataOverride != "" {
						totalSummary.Overrides++
					}
				}
			}
		}
	} else {
		// Single platform
		results, summary, err := config.ResolveArtifactsForModuleWithConfig(
			module, buildDir, targetOS, targetArch, cfg,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to resolve artifacts: %v\n", err)
			return 1
		}
		allResults = results
		totalSummary = summary
	}

	// Get build mode breakdown
	defaultArtifacts := base.DetermineRequestedArtifacts(module, false, cfg)
	allArtifactsList := base.DetermineRequestedArtifacts(module, true, cfg)

	// Output
	output := struct {
		Module     string                                  `json:"module" yaml:"module"`
		Type       string                                  `json:"type" yaml:"type"`
		BuildDir   string                                  `json:"build_dir" yaml:"build_dir"`
		OS         string                                  `json:"os,omitempty" yaml:"os,omitempty"`
		Arch       string                                  `json:"arch,omitempty" yaml:"arch,omitempty"`
		BuildModes *ArtifactBuildModes                     `json:"build_modes" yaml:"build_modes"`
		Metadata   map[string]string                       `json:"metadata,omitempty" yaml:"metadata,omitempty"`
		Artifacts  []config.ResolvedArtifact         `json:"artifacts" yaml:"artifacts"`
		Summary    *config.ArtifactResolutionSummary `json:"summary" yaml:"summary"`
	}{
		Module:   moduleName,
		Type:     module.GetComponentTypesDisplay(),
		BuildDir: buildDir,
		BuildModes: &ArtifactBuildModes{
			Default: defaultArtifacts,
			All:     allArtifactsList,
		},
		Metadata:  module.Metadata,
		Artifacts: allResults,
		Summary:   totalSummary,
	}

	if !allPlatforms {
		output.OS = targetOS
		output.Arch = targetArch
	}

	// Use shared helper for consistent YAML/JSON/TOML output
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return output, nil
	})
}

// ArtifactBuildModes shows which artifacts are built in each mode.
type ArtifactBuildModes struct {
	Default []string `json:"default" yaml:"default"` // Artifacts built in default mode (current platform)
	All     []string `json:"all" yaml:"all"`         // Artifacts built in --all mode (all platforms)
}

type platform struct {
	OS   string
	Arch string
}

func getAllPlatformsFromModule(module *config.Module) []platform {
	platformSet := make(map[string]bool)
	var platforms []platform

	// Check if module has executable artifacts (infer platforms from artifact patterns)
	hasExecutables := false
	for _, pkg := range module.Components {
		if pkg != nil && pkg.Build != nil {
			for _, artifact := range pkg.Build.Artifacts {
				if artifact.Type == config.ArtifactTypeExecutable {
					hasExecutables = true
					break
				}
			}
		}
		if hasExecutables {
			break
		}
	}

	if hasExecutables {
		// For executables, use common cross-platform targets
		// Windows only supports amd64, linux and darwin support both amd64 and arm64
		commonPlatforms := []platform{
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
			{OS: "darwin", Arch: "amd64"},
			{OS: "darwin", Arch: "arm64"},
			{OS: "windows", Arch: "amd64"},
		}
		for _, plat := range commonPlatforms {
			key := fmt.Sprintf("%s-%s", plat.OS, plat.Arch)
			if !platformSet[key] {
				platformSet[key] = true
				platforms = append(platforms, plat)
			}
		}
	}

	// If no platforms found, default to current
	if len(platforms) == 0 {
		platforms = append(platforms, platform{OS: runtime.GOOS, Arch: runtime.GOARCH})
	}

	return platforms
}
