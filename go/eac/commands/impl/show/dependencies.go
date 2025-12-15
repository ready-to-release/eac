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
// Long: Expected Output:
// Long: - Markdown table showing module dependencies with columns: Module, Depends On, Used By
// Long: - Statistics table with metrics like total modules, total dependencies, root/leaf modules
// Long: - Execution order table showing layers of modules that can run in parallel
// Long:
// Long: Example:
// Long:   show dependencies
package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowDependencies)
}

func ShowDependencies() int {
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

	// Calculate and show execution order
	plan, err := repository.CalculateExecutionOrder(nil, workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not calculate execution order: %v\n", err)
	} else {
		fmt.Println("## Execution Order")
		fmt.Println("")
		fmt.Printf("Total layers: %d\n", plan.LayerCount)
		fmt.Println("")

		layerTable := render.NewTableBuilder().
			WithHeaders("Layer", "Modules (can run in parallel)", "Count")

		for i, layer := range plan.Layers {
			layerTable.AddRow(
				fmt.Sprintf("Layer %d", i),
				strings.Join(layer, ", "),
				len(layer),
			)
		}

		fmt.Println(layerTable.Build())
		fmt.Println("")
	}

	return 0
}
