package show

import (
	"context"
	"fmt"
	"os"
	"runtime"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/paths"
)

type showModulesCommand struct{}

var _ core.SimpleCommandPort = (*showModulesCommand)(nil)

func (c *showModulesCommand) Name() string { return "show modules" }

func (c *showModulesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-modules",
		Short:         "Display all module contracts in a human-readable table",
		Long:          "The show modules command displays all module contracts defined in the repository,\nincluding each module's moniker (identifier), group, and component types.\nModules are sorted by dependency depth, then by group, then by declaration order.\nThis information helps understand the modular structure and build order of the repository.\nUse --with-artifacts to include artifact statistics (count, missing, overrides) in the output.\nUse --markdown to output a pipe-separated Markdown table instead of the default console table.\n\nExpected Output:\n- Console table with columns: Moniker, Group, Components\n- If --markdown flag is used: pipe-separated Markdown table\n- If --with-artifacts flag is used: additional columns for Artifacts (count), Missing (count), Overrides (count)\n- Each row represents one module in the repository",
		Flags: []core.FlagSpec{
			{Name: "with-artifacts", Type: "bool", DefaultValue: "false", Usage: "Include artifact statistics (count, missing, overrides)"},
			{Name: "markdown", Type: "bool", DefaultValue: "false", Usage: "Output a pipe-separated Markdown table instead of the console table"},
		},
	}
}

func (c *showModulesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowModules()
}

func ShowModules() int {
	return ExecuteShowCommand(showModulesImpl)
}

func showModulesImpl() int {
	// Parse flags
	args := os.Args[1:]
	withArtifacts := flags.HasFlag(args, "--with-artifacts", "")
	markdown := flags.HasFlag(args, "--markdown", "")

	// Load config (always needed for DisplayOrder; also used for artifacts)
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		return 1
	}

	workspaceRoot := cfg.RepoRoot

	// Generate module contracts report for component type display
	report, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Use precomputed display order
	monikers := cfg.Repository.DisplayOrder.Modules

	// Build markdown table
	tb := render.NewTableBuilder()
	if withArtifacts {
		tb.WithHeaders("Moniker", "Group", "Components", "Artifacts", "Missing", "Overrides")
	} else {
		tb.WithHeaders("Moniker", "Group", "Components")
	}

	for _, moniker := range monikers {
		mod, ok := report.Registry.Get(moniker)
		if !ok {
			continue
		}

		pkgDisplay := mod.GetComponentTypesDisplay()
		group := mod.Group

		if withArtifacts {
			artifactStats := getArtifactStats(mod, cfg, workspaceRoot)
			tb.AddRow(
				mod.Moniker,
				group,
				pkgDisplay,
				fmt.Sprintf("%d", artifactStats.Total),
				fmt.Sprintf("%d", artifactStats.Missing),
				fmt.Sprintf("%d", artifactStats.Overrides),
			)
		} else {
			tb.AddRow(mod.Moniker, group, pkgDisplay)
		}
	}

	if markdown {
		fmt.Println(tb.BuildMarkdown())
	} else {
		fmt.Println(tb.Build())
	}
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
	buildDir := paths.BuildOutputPath(workspaceRoot, mod.Moniker)

	_, summary, err := config.ResolveArtifactsForModuleWithConfig(
		module, buildDir, runtime.GOOS, runtime.GOARCH, cfg,
	)
	if err != nil {
		return artifactStats{}
	}

	return artifactStats{
		Total:     summary.Total,
		Missing:   summary.Missing,
		Overrides: summary.Overrides,
	}
}
