// Command: show modules
// Description: Show all module contracts in the repository
package main

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/render"
	"github.com/ready-to-release/eac/src/contracts/reports"
)

func init() {
	Register("show modules", ShowModules)
}

func ShowModules() int {
	// Generate module contracts report
	report, err := reports.GetModuleContracts("../..", "0.1.0")
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
