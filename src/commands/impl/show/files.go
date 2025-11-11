// Command: show files
// Description: Show repository files with their module ownership
package show

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/commands/render"
	"github.com/ready-to-release/eac/src/internal/repository"
	"github.com/ready-to-release/eac/src/internal/repository/reports"
)

func init() {
	registry.Register("show files", ShowFiles)
}

func ShowFiles() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Generate report for all tracked files (tracked only, no ignored, not staged only)
	report, err := reports.GetFilesModulesReport(true, false, false, workspaceRoot, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Sort by last module in the list (if multiple modules)
	sort.Slice(report.AllFiles, func(i, j int) bool {
		// Get last module for each file (or empty string if no modules)
		lastModuleI := ""
		if len(report.AllFiles[i].Modules) > 0 {
			lastModuleI = report.AllFiles[i].Modules[len(report.AllFiles[i].Modules)-1]
		}

		lastModuleJ := ""
		if len(report.AllFiles[j].Modules) > 0 {
			lastModuleJ = report.AllFiles[j].Modules[len(report.AllFiles[j].Modules)-1]
		}

		// Sort by last module, then by file name if modules are equal
		if lastModuleI != lastModuleJ {
			return lastModuleI < lastModuleJ
		}
		return report.AllFiles[i].Name < report.AllFiles[j].Name
	})

	// Build markdown table with File first, then Modules
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
