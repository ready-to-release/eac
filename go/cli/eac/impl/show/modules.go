// Command: show modules
// Short: Display all module contracts in a human-readable table
// Long: The show modules command displays all module contracts defined in the repository,
// Long: including each module's dependency layer, moniker (identifier), and component types.
// Long: Modules are sorted by layer first, then alphabetically within each layer.
// Long: Layer 0 contains modules with no dependencies; higher layers depend on lower layers.
// Long: This information helps understand the modular structure and build order of the repository.
// Long: Use --with-artifacts to include artifact statistics (count, missing, overrides) in the output.
// Long:
// Long: Expected Output:
// Long: - Markdown table with columns: Layer, Moniker, Components
// Long: - If --with-artifacts flag is used: additional columns for Artifacts (count), Missing (count), Overrides (count)
// Long: - Each row represents one module in the repository
package show

import (
	"fmt"
	"os"
	"runtime"

	implinternal "github.com/ready-to-release/eac/go/cli/eac/impl/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(ShowModules)
}

func ShowModules() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Parse flags
	args := os.Args[1:]
	withArtifacts := flags.HasFlag(args, "--with-artifacts", "")

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Generate module contracts report
	report, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Calculate execution order to get layer assignments
	execPlan, err := repository.CalculateExecutionOrder(nil, workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to calculate execution order: %v\n", err)
		return 1
	}

	// Load configuration if artifacts are requested
	var cfg *config.EACConfig
	if withArtifacts {
		cfg, err = config.Load(config.DefaultLoadOptions())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
			return 1
		}
	}

	// Build markdown table
	tb := render.NewTableBuilder()
	if withArtifacts {
		tb.WithHeaders("Moniker", "Components", "Artifacts", "Missing", "Overrides")
	} else {
		tb.WithHeaders("Moniker", "Components")
	}

	// Iterate in execution order
	for _, moniker := range execPlan.ExecutionOrder {
		// Find the module contract
		mod, ok := report.Registry.Get(moniker)
		if !ok {
			continue // Module not in registry, skip
		}

		// Format components list
		pkgDisplay := mod.GetComponentTypesDisplay()

		if withArtifacts {
			// Get artifact statistics for this module
			artifactStats := getArtifactStats(mod, cfg, workspaceRoot)
			tb.AddRow(
				mod.Moniker,
				pkgDisplay,
				fmt.Sprintf("%d", artifactStats.Total),
				fmt.Sprintf("%d", artifactStats.Missing),
				fmt.Sprintf("%d", artifactStats.Overrides),
			)
		} else {
			tb.AddRow(mod.Moniker, pkgDisplay)
		}
	}

	fmt.Println(tb.Build())
	return 0
}

// artifactStats holds summary statistics for a module's artifacts.
type artifactStats struct {
	Total     int
	Missing   int
	Overrides int
}

// getArtifactStats calculates artifact statistics for a module.
func getArtifactStats(mod *modules.ModuleContract, cfg *config.EACConfig, workspaceRoot string) artifactStats {
	stats := artifactStats{}

	if cfg == nil {
		return stats
	}

	// Get module from config
	module, ok := cfg.Repository.GetModule(mod.Moniker)
	if !ok {
		return stats
	}

	// Check if module has build artifacts defined in any package
	hasArtifacts := false
	for _, pkg := range module.Components {
		if pkg != nil && pkg.Build != nil && len(pkg.Build.Artifacts) > 0 {
			hasArtifacts = true
			break
		}
	}
	if !hasArtifacts {
		return stats
	}

	// Build directory
	buildDir := cfg.Repository.BuildOutputPathAbs(workspaceRoot, mod.Moniker)

	// Resolve artifacts for current platform
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH

	_, summary, err := implinternal.ResolveArtifactsForModuleWithConfig(
		module, buildDir, targetOS, targetArch, cfg,
	)
	if err != nil {
		return stats
	}

	stats.Total = summary.Total
	stats.Missing = summary.Missing
	stats.Overrides = summary.Overrides

	return stats
}
