package show

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

type showSpecsCommand struct{}

var _ core.SimpleCommandPort = (*showSpecsCommand)(nil)

func (c *showSpecsCommand) Name() string { return "show specs" }

func (c *showSpecsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-specs",
		Short:         "Display specifications for a release",
		Long: "Shows .feature specification files that were added or modified for a release.",
		Notes: "Expected Output:\n- Version header with module and version number\n- Summary line with counts (added, modified, deleted, total scenarios)\n- Markdown table with columns: File, Status, Scenarios, Feature\n- Status icons: ✨ Added, 📝 Modified, 🗑️ Deleted",
		Examples: []string{
			"eac show specs eac-ext",
			"eac show specs eac-ext latest",
			"eac show specs eac-ext unreleased",
			"eac show specs eac-ext 0.0.7",
		},
		Args: "module [version]",
		Flags: []core.FlagSpec{
			{Name: "branch", Type: "string", Usage: "Branch to query (default: trunk branch from config, usually \"main\"). Use \"HEAD\" for current branch"},
		},
	}
}

func (c *showSpecsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowSpecs()
}

// ShowSpecs displays specifications in human-readable markdown format.
func ShowSpecs() int {
	return ExecuteShowCommand(showSpecsImpl)
}

func showSpecsImpl() int {
	// Parse arguments - expect module after "show specs"
	args := os.Args[1:]

	// Find where "show specs" ends
	cmdIdx := -1
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "show" && args[i+1] == "specs" {
			cmdIdx = i + 2
			break
		}
	}

	// Parse flags
	branch := flags.GetFlagValue(args, "--branch")

	// Collect positional arguments (non-flag arguments after command)
	var positional []string
	if cmdIdx != -1 && cmdIdx < len(args) {
		for i := cmdIdx; i < len(args); i++ {
			if args[i] != "" && args[i][0] != '-' {
				positional = append(positional, args[i])
			}
		}
	}

	if len(positional) < 1 {
		fmt.Fprintf(os.Stderr, "Error: module argument required\n")
		fmt.Fprintf(os.Stderr, "Usage: show specs <module> [version]\n")
		return 1
	}

	module := positional[0]
	var version string
	if len(positional) > 1 {
		version = positional[1]
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Get specs data using core function
	report, err := reports.GetSpecs(reports.Deps(), workspaceRoot, module, version, branch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Print header
	fmt.Printf("# Specifications: %s (%s)\n\n", module, report.Version)

	// Print summary
	fmt.Printf("**Summary:** %d added, %d modified, %d deleted (%d total scenarios)\n\n",
		report.AddedCount, report.ModifiedCount, report.DeletedCount, report.TotalScenarios)

	if len(report.SpecFiles) == 0 {
		fmt.Printf("No specification changes for this version.\n")
		return 0
	}

	// Print table header
	fmt.Println("| File | Status | Scenarios | Feature |")
	fmt.Println("|------|--------|-----------|---------|")

	// Print each spec file
	for _, spec := range report.SpecFiles {
		statusIcon := ""
		switch spec.Status {
		case "Added":
			statusIcon = "✨ Added"
		case "Modified":
			statusIcon = "📝 Modified"
		case "Deleted":
			statusIcon = "🗑️ Deleted"
		default:
			statusIcon = spec.Status
		}

		scenarios := "-"
		if spec.Status != "Deleted" {
			scenarios = fmt.Sprintf("%d", spec.ScenarioCount)
		}

		fmt.Printf("| %s | %s | %s | %s |\n",
			spec.RelativePath,
			statusIcon,
			scenarios,
			spec.FeatureName)
	}

	return 0
}
