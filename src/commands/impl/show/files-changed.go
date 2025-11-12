// Command: show files changed
// Description: Show changed (modified, unstaged) files with their module ownership
// HasSideEffects: false
package show

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/ready-to-release/eac/src/core/repository/reports"
)

func init() {
	registry.Register(ShowFilesChanged)
}

func ShowFilesChanged() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Get list of changed files from git
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	cmd.Dir = workspaceRoot
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting changed files: %v\n", err)
		return 1
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(changedFiles) == 1 && changedFiles[0] == "" {
		return 0
	}

	// Get full report for all tracked files
	report, err := reports.GetFilesModulesReport(true, false, false, workspaceRoot, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Build map of changed files
	changedMap := make(map[string]bool)
	for _, f := range changedFiles {
		changedMap[f] = true
	}

	// Build markdown table
	tb := render.NewTableBuilder().
		WithHeaders("File", "Modules")

	for _, file := range report.AllFiles {
		if changedMap[file.Name] {
			modules := "NONE"
			if len(file.Modules) > 0 {
				modules = strings.Join(file.Modules, ", ")
			}
			tb.AddRow(file.Name, modules)
		}
	}

	result := tb.Build()
	if result != "" {
		fmt.Println(result)
	}

	return 0
}
