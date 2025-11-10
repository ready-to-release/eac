// Command: show files staged
// Description: Show staged files with their module ownership
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/render"
	"github.com/ready-to-release/eac/src/internal/repository/reports"
)

func init() {
	Register("show files staged", ShowFilesStaged)
}

func ShowFilesStaged() int {
	// Generate report for staged files only
	report, err := reports.GetFilesModulesReport(true, false, true, "../..", "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Build markdown table
	tb := render.NewTableBuilder().
		WithHeaders("File", "Modules")

	for _, file := range report.AllFiles {
		modules := "NONE"
		if len(file.Modules) > 0 {
			modules = strings.Join(file.Modules, ", ")
		}
		tb.AddRow(file.Name, modules)
	}

	fmt.Println(tb.Build())
	return 0
}
