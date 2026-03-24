package show

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

type showApprovalCommentsCommand struct{}

var _ core.SimpleCommandPort = (*showApprovalCommentsCommand)(nil)

func (c *showApprovalCommentsCommand) Name() string { return "show approval-comments" }

func (c *showApprovalCommentsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-approval-comments",
		Short:         "Display PR approval comments in human-readable format",
		Long: "Display PR review approvals for specification-related PRs merged in a release.\n\nShows approvals from GitHub PRs that contain .feature specification files.\nSpecial keywords: \"unreleased\" for pending PRs, \"latest\" for most recent release.\n\nBy default, only shows APPROVED reviews. Use --include-all-reviews to see all review states.",
		Notes: "Expected Output:\n- Header with module and version\n- Summary line with PR and approval counts\n- Markdown table with columns: PR, Title, Reviewer, Review State, Reviewed At",
		Examples: []string{
			"eac show approval-comments eac-ext",
			"eac show approval-comments eac-ext latest",
			"eac show approval-comments eac-ext unreleased",
			"eac show approval-comments eac-ext --include-all-reviews",
		},
		Flags: []core.FlagSpec{
			{Name: "include-all-reviews", Type: "bool", Usage: "Include all review states (not just APPROVED)"},
			{Name: "branch", Type: "string", Usage: "Branch to query (default: trunk branch from config, usually \"main\"). Use \"HEAD\" for current branch"},
		},
		Args: "module [version]",
	}
}

func (c *showApprovalCommentsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowApprovalComments()
}

func ShowApprovalComments() int {
	return ExecuteShowCommand(showApprovalCommentsImpl)
}

func showApprovalCommentsImpl() int {
	// Parse arguments - expect module after "show approval-comments"
	args := os.Args[1:]

	// Find where "show approval-comments" ends
	cmdIdx := -1
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "show" && args[i+1] == "approval-comments" {
			cmdIdx = i + 2
			break
		}
	}

	// Parse flags
	includeAllReviews := flags.HasFlag(args, "--include-all-reviews", "")
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
		fmt.Fprintf(os.Stderr, "Usage: show approval-comments <module> [version] [--include-all-reviews]\n")
		return 1
	}

	module := positional[0]
	version := ""
	if len(positional) > 1 {
		version = positional[1]
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Get approval comments data using core function
	report, err := reports.GetApprovalComments(reports.Deps(), workspaceRoot, module, version, includeAllReviews, branch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Print header
	fmt.Printf("# PR Approvals: %s (%s)\n\n", module, report.Version)

	// Print summary
	fmt.Printf("**Summary:** %d PRs, %d approvals\n\n", report.TotalPRs, report.TotalApprovals)

	// Print approval table if there are any approvals
	if len(report.Approvals) > 0 {
		tb := render.NewTableBuilder().
			WithHeaders("PR", "Title", "Reviewer", "Review State", "Reviewed At")

		for _, approval := range report.Approvals {
			prLink := fmt.Sprintf("#%d", approval.PRNumber)
			reviewedAt := approval.ReviewedAt.Format("2006-01-02")

			tb.AddRow(
				prLink,
				approval.PRTitle,
				approval.Reviewer,
				approval.ReviewState,
				reviewedAt,
			)
		}

		fmt.Println(tb.Build())
	} else {
		fmt.Println("No PR approvals found for this version.")
	}

	// Print detailed PR information (body and merge message) for ALL found PRs
	// Display even if PRs have no reviews/approvals
	if len(report.PRs) > 0 {
		fmt.Print("\n## PR Details\n\n")

		for _, pr := range report.PRs {
			fmt.Printf("### PR #%d: %s\n\n", pr.Number, pr.Title)

			if pr.Body != "" {
				fmt.Printf("**Description:**\n\n%s\n\n", pr.Body)
			} else {
				fmt.Println("**Description:** (No description provided)")
			}

			if pr.MergeCommitMessage != "" {
				fmt.Printf("**Merge Commit Message:**\n\n%s\n\n", pr.MergeCommitMessage)
			} else {
				fmt.Println("**Merge Commit Message:** (No merge message)")
			}

			fmt.Print("---\n\n")
		}
	}

	return 0
}
