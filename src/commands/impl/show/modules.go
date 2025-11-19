// Command: show modules
// Description: Show all module contracts in the repository
// Short: Display all module contracts in a human-readable table
// Long: The show modules command displays all module contracts defined in the repository,
// Long: including each module's moniker (identifier), type, and root path.
// Long: This information helps understand the modular structure of the repository and identify available modules.
// Long: The output is formatted as a Markdown table for easy reading.
// Long: Use this command before making changes to understand which modules exist and their locations.
// HasSideEffects: false
package show

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(ShowModules)
}

func ShowModules() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Generate module contracts report
	report, err := reports.GetModuleContracts(workspaceRoot, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Build markdown table
	tb := render.NewTableBuilder().
		WithHeaders("Moniker", "Type", "Root Path")

	for _, mod := range report.Modules {
		tb.AddRow(mod.Moniker, mod.Type, mod.Source.Root)
	}

	fmt.Println(tb.Build())
	return 0
}
