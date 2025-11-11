// Command: show files staged
// Description: Show staged files with their module ownership
package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/commands/render"
	"github.com/ready-to-release/eac/src/internal/repository"
	"github.com/ready-to-release/eac/src/internal/repository/reports"
)

func init() {
	registry.Register("show files staged", ShowFilesStaged)
}

func ShowFilesStaged() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Generate report for staged files only
	report, err := reports.GetFilesModulesReport(true, false, true, workspaceRoot, "0.1.0")
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
