package show

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/repository/reports"
)

type showFilesCommand struct{}

var _ core.SimpleCommandPort = (*showFilesCommand)(nil)

func (c *showFilesCommand) Name() string { return "show files" }

func (c *showFilesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-files",
		Short:         "Display all tracked files and their owning modules",
		Long:          "The show files command displays all tracked files in the repository with their module ownership.\nShows which files belong to which modules, helping understand repository structure.\n\nExpected Output:\n- Table with file paths and owning modules (comma-separated if multiple)\n- Files with no module ownership shown as \"NONE\"\n- Sorted by module for easy grouping",
	}
}

func (c *showFilesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowFiles()
}

func ShowFiles() int {
	return ExecuteShowCommand(showFilesImpl)
}

func showFilesImpl() int {
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
