// Command: show files
// Description: Show repository files with their module ownership
// Short: Display all tracked files and their owning modules
// Long: The show files command displays all tracked files in the repository with their module ownership.
// Long: Shows which files belong to which modules, helping understand repository structure.
// Long:
// Long: Expected Output:
// Long: - Table with file paths and owning modules (comma-separated if multiple)
// Long: - Files with no module ownership shown as "NONE"
// Long: - Sorted by module for easy grouping
package show

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/repository/reports"
)

func init() {
	registry.Register(ShowFiles)
}

func ShowFiles() int {
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

	// Generate report for all tracked files (tracked only, no ignored, not staged only)
	report, err := reports.GetFilesModulesReport(true, false, false, workspaceRoot)
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
