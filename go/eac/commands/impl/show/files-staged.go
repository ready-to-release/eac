// Command: show files-staged
// Short: Show staged files with their module ownership
package show

import (
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/repository/reports"
)

func init() {
	registry.Register(ShowFilesStaged)
}

func ShowFilesStaged() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Generate report for staged files only
	report, err := reports.GetFilesModulesReport(true, false, true, workspaceRoot)
	if err != nil {
		log.Errorf("%v", err)
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

	log.Info(tb.Build())
	return 0
}
