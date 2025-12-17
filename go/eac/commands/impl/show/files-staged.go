// Command: show files-staged
// Short: Show staged files with their module ownership
// Long: The show files-staged command displays staged files (git diff --cached) with their module ownership.
// Long: Useful for identifying which modules are affected by staged changes before committing.
// Long:
// Long: Expected Output:
// Long: - Table with staged file paths and owning modules (comma-separated if multiple)
// Long: - Files with no module ownership shown as "NONE"
// Long: - Empty output if no files are staged
package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/repository/reports"
)

func init() {
	registry.Register(ShowFilesStaged)
}

func ShowFilesStaged() int {
	args := os.Args[3:] // Skip program name, "show", and "files-staged"

	// Validate flags (no flags expected for this command)
	commandFlags := []flags.FlagDefinition{}
	if err := flags.ValidateFlags(args, commandFlags); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Generate report for staged files only
	report, err := reports.GetFilesModulesReport(true, false, true, workspaceRoot)
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
