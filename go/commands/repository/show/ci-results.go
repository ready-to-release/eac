package show

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/gh"
	"github.com/ready-to-release/eac/go/commands/repository/get"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

type showCIResultsCommand struct{}

var _ core.SimpleCommandPort = (*showCIResultsCommand)(nil)

func (c *showCIResultsCommand) Name() string { return "show ci-results" }

func (c *showCIResultsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-ci-results",
		Short:         "Show CI workflow run results with job details and download links",
		Long:          "Shows a pretty-formatted summary of CI workflow run results with\nper-module job status tables, artifact listings, and copy-pasteable\ndiagnostic commands for investigating failures.\n\nInput Detection:\n  - 40-char hex or 7+ hex prefix: treated as commit SHA\n  - Numeric value: treated as a specific run ID\n  - Omitted: auto-detects SHA (GITHUB_SHA -> origin/main -> git HEAD)\n\nExample:\n  show ci-results                           # Current HEAD\n  show ci-results abc1234                    # Specific commit\n  show ci-results abc1234 core clibase       # Specific modules\n  show ci-results 12345678                    # Specific run ID",
		Args:          "[sha-or-run-id] [module...]",
	}
}

func (c *showCIResultsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowCIResults()
}

func ShowCIResults() int {
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse positional args (skip flags)
	args := os.Args[3:] // Skip program, "show", "ci-results"
	var positional []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			continue
		}
		positional = append(positional, arg)
	}

	// Classify first arg
	iType, shaOrRunID := classifyShowInput(positional)
	modules := positional
	if len(positional) > 0 && iType != showInputAuto {
		modules = positional[1:]
	}

	// Resolve SHA
	var headSHA string
	var specificRunID int

	switch iType {
	case showInputRunID:
		id, _ := strconv.Atoi(shaOrRunID)
		specificRunID = id
	case showInputSHA:
		headSHA = shaOrRunID
	case showInputAuto:
		result, err := get.DetectCurrentSHA(workspaceRoot, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to detect SHA: %v\n", err)
			return 1
		}
		headSHA = result.SHA
	}

	// Build GitHub API client
	api := github.NewGHClient(gh.New(tool.GlobalToolSystem(), workspaceRoot), workspaceRoot)

	// Get repo name for links
	repo := getCIResultsRepoName(workspaceRoot)

	var summary *get.CIResultsSummary

	if specificRunID > 0 {
		summary, err = get.GetCIResultsForRunID(specificRunID, api, repo, workspaceRoot)
	} else {
		summary, err = get.GetCIResultsForSHA(headSHA, modules, api, repo, workspaceRoot)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Render pretty output
	renderCIResults(summary)
	return 0
}

// tableRow holds pre-computed values for one table row.
type tableRow struct {
	name   string
	status string
	link   string
}

func renderCIResults(summary *get.CIResultsSummary) {
	sha := summary.HeadSHA
	if len(sha) > 7 {
		sha = sha[:7]
	}

	if summary.TotalRuns == 0 && summary.Orchestrator == nil {
		fmt.Printf("CI Results: %s — no runs found\n", sha)
		return
	}

	fmt.Printf("CI Results: %s — %d passed, %d failed\n\n",
		sha, summary.Passed, summary.Failed)

	// Build rows: orchestrator first, then modules
	var rows []tableRow

	if summary.Orchestrator != nil {
		rows = append(rows, tableRow{
			name:   summary.Orchestrator.Workflow,
			status: runStatus(summary.Orchestrator),
			link:   runLink(summary.Orchestrator),
		})
	}

	for i := range summary.Runs {
		run := &summary.Runs[i]
		rows = append(rows, tableRow{
			name:   run.Workflow,
			status: runStatus(run),
			link:   runLink(run),
		})
	}

	// Compute column widths
	nameW := len("Workflow")
	statusW := len("Status")
	for _, r := range rows {
		if len(r.name) > nameW {
			nameW = len(r.name)
		}
		if len(r.status) > statusW {
			statusW = len(r.status)
		}
	}

	// Header
	fmt.Printf("  %-*s  %-*s  %s\n", nameW, "Workflow", statusW, "Status", "Link")
	fmt.Printf("  %-*s  %-*s  %s\n", nameW, dashes(nameW), statusW, dashes(statusW), dashes(4))

	// Rows
	for _, r := range rows {
		fmt.Printf("  %-*s  %-*s  %s\n", nameW, r.name, statusW, r.status, r.link)
	}
}

// runStatus returns a compact status string.
// In-progress runs show their status; completed failures include the failed job name.
func runStatus(run *get.CIRunResult) string {
	if run.Conclusion == "" {
		// Still running — use status (in_progress, queued, etc.)
		return run.Status
	}
	if run.Conclusion != "failure" {
		return run.Conclusion
	}
	// Find first failed job
	for _, job := range run.Jobs {
		if job.Conclusion == "failure" {
			return "FAILED: " + stripJobPrefix(job.Name)
		}
	}
	return "FAILED"
}

// runLink returns the most useful link for the run.
// For failures: the failed job URL. Otherwise: the run web URL.
func runLink(run *get.CIRunResult) string {
	if run.Conclusion == "failure" {
		for _, job := range run.Jobs {
			if job.Conclusion == "failure" && job.HTMLURL != "" {
				return job.HTMLURL
			}
		}
	}
	return run.Links.WebURL
}

func dashes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '-'
	}
	return string(b)
}

// stripJobPrefix removes the "ci / " prefix GitHub adds to reusable workflow job names.
func stripJobPrefix(name string) string {
	if idx := strings.Index(name, " / "); idx != -1 {
		return name[idx+3:]
	}
	return name
}

// showInputType classifies the first positional argument (show-local copy to avoid import cycle).
type showInputType int

const (
	showInputSHA   showInputType = iota
	showInputRunID showInputType = iota
	showInputAuto  showInputType = iota
)

func classifyShowInput(positional []string) (showInputType, string) {
	if len(positional) == 0 {
		return showInputAuto, ""
	}

	first := positional[0]

	// Pure numeric → run ID
	if _, err := strconv.Atoi(first); err == nil {
		return showInputRunID, first
	}

	// Hex string 7-40 chars → SHA
	if isHexString(first) && len(first) >= 7 && len(first) <= 40 {
		return showInputSHA, first
	}

	// Otherwise it's a module name → auto-detect SHA
	return showInputAuto, ""
}

func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return len(s) > 0
}

func getCIResultsRepoName(workspaceRoot string) string {
	if repo := os.Getenv("GITHUB_REPOSITORY"); repo != "" {
		return repo
	}
	if repo := os.Getenv("GH_REPO"); repo != "" {
		return repo
	}
	output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", workspaceRoot, "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	if err == nil {
		name := strings.TrimSpace(string(output))
		if name != "" {
			return name
		}
	}
	return "{owner}/{repo}"
}
