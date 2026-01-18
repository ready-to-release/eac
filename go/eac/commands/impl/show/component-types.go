// Command: show component-types
// Short: Display all component types grouped by count
// Long: The show component-types command displays all component types used in the repository,
// Long: grouped by count to show which types are most commonly used.
// Long: This helps understand the technology mix and module distribution in the repository.
// Long: The output is formatted as a Markdown table with a footer showing total types.
package show

import (
	"fmt"
	"os"
	"sort"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowComponentTypes)
}

func ShowComponentTypes() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

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

	// Count by component type across all modules
	// Use GetComponentType to resolve actual types, not just names
	typeCounts := make(map[string]int)
	for _, mod := range report.Modules {
		for _, compName := range mod.GetEnabledComponents() {
			// Resolve actual type (may differ from name for named components like book types)
			compType := mod.Components.GetComponentType(compName)
			typeCounts[compType]++
		}
	}

	// Sort types alphabetically
	types := make([]string, 0, len(typeCounts))
	for t := range typeCounts {
		types = append(types, t)
	}
	sort.Strings(types)

	// Build markdown table
	tb := render.NewTableBuilder()
	tb.WithHeaders("Component Type", "Count")

	for _, t := range types {
		tb.AddRow(t, fmt.Sprintf("%d", typeCounts[t]))
	}

	// Add footer with total
	tb.WithFooter("Total Types", fmt.Sprintf("%d", len(types)))

	fmt.Println(tb.Build())
	return 0
}
