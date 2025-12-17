// Command: show moduletypes
// Description: Show all module types grouped by count
// Short: Display module types and their usage counts
// Long: The show moduletypes command displays all module types in the repository with their usage counts.
// Long: Shows how many modules of each type exist, sorted alphabetically by type name.
// Long:
// Long: Expected Output:
// Long: - Table with columns: Module Type, Count
// Long: - Footer row showing total number of unique module types
// Long: - Types sorted alphabetically
package show

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowModuleTypes)
}

func ShowModuleTypes() int {
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

	// Group modules by type
	typeCount := make(map[string]int)
	for _, mod := range report.Modules {
		typeCount[mod.Type]++
	}

	// Sort types alphabetically
	var types []string
	for t := range typeCount {
		types = append(types, t)
	}
	for i := 0; i < len(types); i++ {
		for j := i + 1; j < len(types); j++ {
			if types[i] > types[j] {
				types[i], types[j] = types[j], types[i]
			}
		}
	}

	// Build markdown table
	tb := render.NewTableBuilder().
		WithHeaders("Module Type", "Count")

	for _, modType := range types {
		tb.AddRow(modType, typeCount[modType])
	}

	// Add footer with total
	tb.WithFooter("Total Types", len(types))

	fmt.Println(tb.Build())
	return 0
}
