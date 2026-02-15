package get

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getBuildDepsCommand struct{}

var _ core.SimpleCommandPort = (*getBuildDepsCommand)(nil)

func (c *getBuildDepsCommand) Name() string { return "get build-deps" }

func (c *getBuildDepsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-build-deps",
		Short:         "Get aggregated build dependencies for a module",
		Long: "Expected Output:\n" +
			"YAML list of module build dependencies, aggregated from the module and all its transitive\n" +
			"dependencies. Includes system dependencies resolved from module type capabilities (e.g., go,\n" +
			"node, docker) and artifact-specific requirements (e.g., upx for compression).",
		Args: "module (required) - Module moniker",
		Flags: []core.FlagSpec{
			{Name: "format", Type: "string", Usage: "Output format (shell, space, yaml, json)"},
		},
	}
}

func (c *getBuildDepsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetBuildDeps()
}

// BuildDepsResult contains the build dependencies for a module.
type BuildDepsResult struct {
	Module    string   `json:"module" yaml:"module"`
	Type      string   `json:"type" yaml:"type"`
	BuildDeps []string `json:"build_deps" yaml:"build_deps"`
}

func GetBuildDeps() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Check for help flag
	for _, arg := range os.Args[2:] {
		if arg == "--help" || arg == "-h" {
			printBuildDepsUsage()
			return 0
		}
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

	// Get build deps from package types config
	cfg := config.Global()
	if cfg == nil || cfg.ComponentKinds == nil {
		fmt.Fprintf(os.Stderr, "Error: package types configuration not loaded\n")
		return 1
	}

	// Aggregate build deps from module and all its dependencies
	buildDeps := aggregateBuildDeps(moniker, moduleReport.Registry)

	// Handle shell format output
	if format == "shell" {
		fmt.Printf("MODULE_PACKAGES=\"%s\"\n", module.GetComponentTypesDisplay())
		fmt.Printf("BUILD_DEPS=\"%s\"\n", strings.Join(buildDeps, ","))
		return 0
	}

	// Use the shared get command helper for output formatting
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return BuildDepsResult{
			Module:    moniker,
			Type:      module.GetComponentTypesDisplay(),
			BuildDeps: buildDeps,
		}, nil
	})
}

// aggregateBuildDeps collects build dependencies from a module and all its dependencies.
// Docker is the only true system dependency - required for any container-based tool execution.
// UPX is included when artifacts require upx compression.
func aggregateBuildDeps(moniker string, registry *modules.Registry) []string {
	seen := make(map[string]bool)
	depsSet := make(map[string]bool)

	// Docker is always a dependency for builds that use container tools.
	// The tool system handles the binding (system vs container) at execution time.
	// Since we can't know at this point which tools will be used (that's determined
	// by bindings), we conservatively include docker as a build dep.
	depsSet["docker"] = true

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
	if cfg == nil || cfg.ComponentKinds == nil {
		return "", fmt.Errorf("package types configuration not loaded")
	}

	// Aggregate build deps from module and all its dependencies
	buildDeps := aggregateBuildDeps(moniker, moduleReport.Registry)
	if len(buildDeps) == 0 {
		return "", nil
	}

	return strings.Join(buildDeps, ","), nil
}

func printBuildDepsUsage() {
	fmt.Println("Get build dependencies for a module")
	fmt.Println()
	fmt.Println("Usage: get build-deps <module> [flags]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  module    Module moniker to get build dependencies for")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --format <type>    Output format (shell, space, yaml, json)")
	fmt.Println("  -h, --help         Show this help message")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  YAML list of module build dependencies, aggregated from the module")
	fmt.Println("  and all its transitive dependencies.")
}
