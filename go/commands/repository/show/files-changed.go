package show

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/repository/reports"
)

type showFilesChangedCommand struct{}

var _ core.SimpleCommandPort = (*showFilesChangedCommand)(nil)

func (c *showFilesChangedCommand) Name() string { return "show files-changed" }

func (c *showFilesChangedCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-files-changed",
		Short:         "Show changed (modified, unstaged) files with their module ownership",
		Long:          "The show files-changed command displays modified files (git diff HEAD) with their module ownership.\nUseful for identifying which modules are affected by uncommitted changes.\n\nExpected Output:\n- Table with changed file paths and owning modules (comma-separated if multiple)\n- Files with no module ownership shown as \"NONE\"\n- Empty output if no files are changed",
	}
}

func (c *showFilesChangedCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowFilesChanged()
}

func ShowFilesChanged() int {
	return ExecuteShowCommand(showFilesChangedImpl)
}

func showFilesChangedImpl() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Get list of changed files from git
	ts := tool.GlobalToolSystem()
	if ts == nil {
		fmt.Fprintf(os.Stderr, "Error: tool system not initialized\n")
		return 1
	}
	output, err := ts.RunTool(context.Background(), "git", workspaceRoot, "diff", "--name-only", "HEAD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: getting changed files: %v\n", err)
		return 1
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(changedFiles) == 1 && changedFiles[0] == "" {
		return 0
	}

	// Get full report for all tracked files
	report, err := reports.GetFilesModulesReport(true, false, false, workspaceRoot)
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
