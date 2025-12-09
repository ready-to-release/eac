// Command: show modules
// Description: Show all module contracts in the repository
// Short: Display all module contracts in a human-readable table
// Long: The show modules command displays all module contracts defined in the repository,
// Long: including each module's moniker (identifier), type, and root path.
// Long: This information helps understand the modular structure of the repository and identify available modules.
// Long: The output is formatted as a Markdown table for easy reading.
// Long: Use this command before making changes to understand which modules exist and their locations.
// Long: Use --with-artifacts to include artifact statistics (count, missing, overrides) in the output.
package show

import (
	"fmt"
	"os"
	"runtime"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowModules)
}

func ShowModules() int {
	// Parse flags
	args := os.Args[1:]
	withArtifacts := false
	for _, arg := range args {
		if arg == "--with-artifacts" {
			withArtifacts = true
		}
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Generate module contracts report
	report, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Load configuration if artifacts are requested
	var cfg *config.EACConfig
	if withArtifacts {
		cfg, err = config.Load(config.DefaultLoadOptions())
		if err != nil {
			log.Errorf("failed to load config: %v", err)
			return 1
		}
	}

	// Build markdown table
	tb := render.NewTableBuilder()
	if withArtifacts {
		tb.WithHeaders("Moniker", "Type", "Root Path", "Artifacts", "Missing", "Overrides")
	} else {
		tb.WithHeaders("Moniker", "Type", "Root Path")
	}

	for _, mod := range report.Modules {
		if withArtifacts {
			// Get artifact statistics for this module
			artifactStats := getArtifactStats(mod, cfg, workspaceRoot)
			tb.AddRow(
				mod.Moniker,
				mod.Type,
				mod.Files.Root,
				fmt.Sprintf("%d", artifactStats.Total),
				fmt.Sprintf("%d", artifactStats.Missing),
				fmt.Sprintf("%d", artifactStats.Overrides),
			)
		} else {
			tb.AddRow(mod.Moniker, mod.Type, mod.Files.Root)
		}
	}

	log.Info(tb.Build())
	return 0
}

// artifactStats holds summary statistics for a module's artifacts
type artifactStats struct {
	Total     int
	Missing   int
	Overrides int
}

// getArtifactStats calculates artifact statistics for a module
func getArtifactStats(mod *modules.ModuleContract, cfg *config.EACConfig, workspaceRoot string) artifactStats {
	stats := artifactStats{}

	if cfg == nil {
		return stats
	}

	// Get module from config
	module, ok := cfg.Modules.GetModule(mod.Moniker)
	if !ok {
		return stats
	}

	// Get module type
	moduleType := cfg.ModuleTypes.Get(mod.Type)
	if moduleType == nil || moduleType.Build == nil {
		return stats
	}

	// Build directory
	buildDir := cfg.Repository.BuildOutputPathAbs(workspaceRoot, mod.Moniker)

	// Resolve artifacts for current platform
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH

	_, summary, err := implinternal.ResolveArtifactsForModuleWithConfig(
		module, moduleType, buildDir, targetOS, targetArch, cfg,
	)
	if err != nil {
		return stats
	}

	stats.Total = summary.Total
	stats.Missing = summary.Missing
	stats.Overrides = summary.Overrides

	return stats
}
