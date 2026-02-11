package show

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/repository/reports"
)

type showFilesStagedCommand struct{}

var _ core.SimpleCommandPort = (*showFilesStagedCommand)(nil)

func (c *showFilesStagedCommand) Name() string { return "show files-staged" }

func (c *showFilesStagedCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-files-staged",
		Short:         "Show staged files with their module ownership",
		Long:          "The show files-staged command displays staged files (git diff --cached) with their module ownership.\nUseful for identifying which modules are affected by staged changes before committing.\n\nExpected Output:\n- Table with staged file paths and owning modules (comma-separated if multiple)\n- Files with no module ownership shown as \"NONE\"\n- Empty output if no files are staged",
	}
}

func (c *showFilesStagedCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowFilesStaged()
}

func ShowFilesStaged() int {
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
