// Command: show moduletypes
// Description: Show all module types grouped by count
package main

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/render"
	"github.com/ready-to-release/eac/src/internal/contracts/reports"
)

func init() {
	Register("show moduletypes", ShowModuleTypes)
}

func ShowModuleTypes() int {
	// Generate module contracts report
	report, err := reports.GetModuleContracts("../..", "0.1.0")
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
