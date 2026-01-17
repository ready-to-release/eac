// Command: get build-deps
// Args: module (required) - Module moniker
// Long:
// Long: Expected Output:
// Long: YAML list of module build dependencies, aggregated from the module and all its transitive
// Long: dependencies. Includes system dependencies resolved from module type capabilities (e.g., go,
// Long: node, docker) and artifact-specific requirements (e.g., upx for compression).
// Flag.format: type=string, usage=Output format (shell, space, yaml, json)
package get

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetBuildDeps)
}

// BuildDepsResult contains the build dependencies for a module.
type BuildDepsResult struct {
	Module    string   `json:"module" yaml:"module"`
	Type      string   `json:"type" yaml:"type"`
	BuildDeps []string `json:"build_deps" yaml:"build_deps"`
}

// buildDepsFlags defines valid flags for the get build-deps command

func GetBuildDeps() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "get", and "build-deps"

	// Parse module moniker and flags from args
	var moniker string
	format := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--format" && i+1 < len(args) {
			format = args[i+1]
			i++
		} else if !strings.HasPrefix(arg, "--") && moniker == "" {
			moniker = arg
		}
	}

	if moniker == "" {
		fmt.Fprintf(os.Stderr, "Usage: get build-deps <module> [--format shell]\n")
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Load module contracts to get module type
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load module contracts: %v\n", err)
		return 1
	}

	// Find the module
	module, exists := moduleReport.Registry.Get(moniker)
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: module not found: %s\n", moniker)
		return 1
	}

	// Get build deps from module types config
	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		fmt.Fprintf(os.Stderr, "Error: module types configuration not loaded\n")
		return 1
	}

	// Aggregate build deps from module and all its dependencies
	buildDeps := aggregateBuildDeps(moniker, moduleReport.Registry, cfg.ModuleTypes, cfg.SystemDependencies)

	// Handle shell format output
	if format == "shell" {
		fmt.Printf("MODULE_TYPE=\"%s\"\n", module.Type)
		fmt.Printf("BUILD_DEPS=\"%s\"\n", strings.Join(buildDeps, ","))
		return 0
	}

	// Use the shared get command helper for output formatting
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return BuildDepsResult{
			Module:    moniker,
			Type:      module.Type,
			BuildDeps: buildDeps,
		}, nil
	})
}

// aggregateBuildDeps collects build dependencies from a module and all its dependencies.
func aggregateBuildDeps(moniker string, registry *modules.Registry, moduleTypes *config.ModuleTypesConfig, sysDeps *config.SystemDependenciesConfig) []string {
	seen := make(map[string]bool)
	depsSet := make(map[string]bool)

	var collect func(m string)
	collect = func(m string) {
		if seen[m] {
			return
		}
		seen[m] = true

		module, exists := registry.Get(m)
		if !exists {
			return
		}

		// Add this module's build deps (resolved from capabilities)
		deps := moduleTypes.GetBuildDepsFromCapabilities(module.Type, sysDeps)
		for _, dep := range deps {
			depsSet[dep] = true
		}

		// Modules with books require docker for mkdocs builds
		if len(module.Books) > 0 {
			depsSet["docker"] = true
		}

		// Check per-module artifacts for compression requirements
		for _, artifact := range module.GetBuildArtifacts() {
			if artifact.Compression == "upx" {
				depsSet["upx"] = true
			}
		}

		// Recurse into dependencies
		for _, depMoniker := range module.DependsOn {
			collect(depMoniker)
		}
	}

	collect(moniker)

	// Convert set to sorted slice for consistent output
	result := make([]string, 0, len(depsSet))
	for dep := range depsSet {
		result = append(result, dep)
	}

	// Sort for consistent output
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// GetBuildDepsPlain returns build deps as comma-separated string (for shell scripts).
func GetBuildDepsPlain(moniker string) (string, error) {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return "", fmt.Errorf("failed to find repository root: %w", err)
	}

	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load module contracts: %w", err)
	}

	_, exists := moduleReport.Registry.Get(moniker)
	if !exists {
		return "", fmt.Errorf("module not found: %s", moniker)
	}

	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		return "", fmt.Errorf("module types configuration not loaded")
	}

	// Aggregate build deps from module and all its dependencies
	buildDeps := aggregateBuildDeps(moniker, moduleReport.Registry, cfg.ModuleTypes, cfg.SystemDependencies)
	if len(buildDeps) == 0 {
		return "", nil
	}

	return strings.Join(buildDeps, ","), nil
}
