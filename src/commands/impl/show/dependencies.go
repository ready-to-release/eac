// Command: show dependencies
// Short: Show module dependency graph in a human-readable table format
// Long: Show module dependency graph in a human-readable table format.
// Long:
// Long: This command displays the dependency relationships between modules in the repository.
// Long: The output shows which modules depend on which other modules, helping you understand
// Long: the module architecture and plan changes that respect dependencies.
// Long:
// Long: The table format makes it easy to see the full dependency graph at a glance.
// Long:
// Long: Example:
// Long:   show dependencies
package show

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(ShowDependencies)
}

func ShowDependencies() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Get dependency graph
	graph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Print header
	log.Info("# Module Dependency Graph")
	log.Info("")

	// Print statistics
	log.Info("## Statistics")
	log.Info("")
	stats := render.NewTableBuilder().
		WithHeaders("Metric", "Value")

	stats.AddRow("Total Modules", graph.Stats.TotalModules)
	stats.AddRow("Total Dependencies", graph.Stats.TotalDependencies)
	stats.AddRow("Root Modules (no dependencies)", graph.Stats.RootModules)
	stats.AddRow("Leaf Modules (no dependents)", graph.Stats.LeafModules)
	stats.AddRow("Max Dependencies", graph.Stats.MaxDependencies)
	stats.AddRow("Max Dependents", graph.Stats.MaxDependents)

	log.Info(stats.Build())
	log.Info("")

	// Print module dependencies table
	log.Info("## Module Dependencies")
	log.Info("")

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

	log.Info(tb.Build())
	log.Info("")

	// Calculate and show execution order
	plan, err := repository.CalculateExecutionOrder(nil, workspaceRoot)
	if err != nil {
		log.Errorf("Warning: Could not calculate execution order: %v", err)
	} else {
		log.Info("## Execution Order")
		log.Info("")
		log.Infof("Total layers: %d", plan.LayerCount)
		log.Info("")

		layerTable := render.NewTableBuilder().
			WithHeaders("Layer", "Modules (can run in parallel)", "Count")

		for i, layer := range plan.Layers {
			layerTable.AddRow(
				fmt.Sprintf("Layer %d", i),
				strings.Join(layer, ", "),
				len(layer),
			)
		}

		log.Info(layerTable.Build())
		log.Info("")
	}

	return 0
}
