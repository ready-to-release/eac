package show

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/repository"
)

type showDependenciesCommand struct{}

var _ core.SimpleCommandPort = (*showDependenciesCommand)(nil)

func (c *showDependenciesCommand) Name() string { return "show dependencies" }

func (c *showDependenciesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-dependencies",
		Short:         "Show module dependency graph in a human-readable table format",
		Long:          "Show module dependency graph in a human-readable table format.\n\nThis command displays the dependency relationships between modules in the repository.\nThe output shows which modules depend on which other modules, helping you understand\nthe module architecture and plan changes that respect dependencies.\n\nThe table format makes it easy to see the full dependency graph at a glance.\n\nExpected Output:\n- Markdown table showing module dependencies with columns: Module, Depends On, Used By\n- Statistics table with metrics like total modules, total dependencies, root/leaf modules\n- Display order table showing modules ordered by depth, group, and declaration order\n\nExample:\n  show dependencies",
	}
}

func (c *showDependenciesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowDependencies()
}

func ShowDependencies() int {
	return ExecuteShowCommand(showDependenciesImpl)
}

func showDependenciesImpl() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Get dependency graph
	graph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Print header
	fmt.Println("# Module Dependency Graph")
	fmt.Println("")

	// Print statistics
	fmt.Println("## Statistics")
	fmt.Println("")
	stats := render.NewTableBuilder().
		WithHeaders("Metric", "Value")

	stats.AddRow("Total Modules", graph.Stats.TotalModules)
	stats.AddRow("Total Dependencies", graph.Stats.TotalDependencies)
	stats.AddRow("Root Modules (no dependencies)", graph.Stats.RootModules)
	stats.AddRow("Leaf Modules (no dependents)", graph.Stats.LeafModules)
	stats.AddRow("Max Dependencies", graph.Stats.MaxDependencies)
	stats.AddRow("Max Dependents", graph.Stats.MaxDependents)

	fmt.Println(stats.Build())
	fmt.Println("")

	// Print module dependencies table
	fmt.Println("## Module Dependencies")
	fmt.Println("")

	tb := render.NewTableBuilder().
		WithHeaders("Module", "Depends On", "Used By")

	for _, moniker := range graph.Modules {
		deps := graph.Dependencies[moniker]
		depts := graph.Dependents[moniker]

		depsStr := "-"
		if len(deps) > 0 {
			depsStr = strings.Join(deps, ", ")
		}

		deptsStr := "-"
		if len(depts) > 0 {
			deptsStr = strings.Join(depts, ", ")
		}

		tb.AddRow(moniker, depsStr, deptsStr)
	}

	fmt.Println(tb.Build())
	fmt.Println("")

	// Show display order from precomputed DisplayOrder
	cfg, cfgErr := config.Load(config.DefaultLoadOptions())
	if cfgErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load config for display order: %v\n", cfgErr)
	} else if cfg.Repository.DisplayOrder != nil {
		displayOrder := cfg.Repository.DisplayOrder
		fmt.Println("## Display Order")
		fmt.Println("")
		fmt.Printf("Total modules: %d\n", len(displayOrder.Modules))
		fmt.Println("")

		orderTable := render.NewTableBuilder().
			WithHeaders("Order", "Module", "Depth")

		for i, module := range displayOrder.Modules {
			depth := fmt.Sprintf("%d", displayOrder.Depth[module])
			if displayOrder.IsBaseline[module] {
				depth += " (baseline)"
			}
			orderTable.AddRow(fmt.Sprintf("%d", i+1), module, depth)
		}

		fmt.Println(orderTable.Build())
		fmt.Println("")
	}

	return 0
}
