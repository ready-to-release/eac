// Command: show pipeline-status
// Short: Show CI status for the head of trunk
// Long: Show the CI pipeline status for the head of the main branch.
// Long:
// Long: This command queries GitHub Actions to display the status of all CI
// Long: workflows for the latest commit on the main branch. It shows a compact
// Long: summary of each workflow's status.
// Long:
// Long: Example:
// Long:   show pipeline-status              # Show status for main branch HEAD
// Long:   show pipeline-status --ref dev    # Show status for dev branch HEAD
// Long:   show pipeline-status --commit abc # Show status for specific commit
// Flag.ref: type=string, usage=Git ref to check (default: main)
// Flag.commit: type=string, usage=Specific commit SHA to check
package show

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/src/commands/registry"
)

func init() {
	registry.Register(ShowPipelineStatus)
}

// PipelineWorkflowRunInfo represents workflow run information from GitHub
type PipelineWorkflowRunInfo struct {
	DatabaseID  int    `json:"databaseId"`
	Name        string `json:"name"`
	DisplayName string `json:"displayTitle"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion"`
	HeadSha     string `json:"headSha"`
	URL         string `json:"url"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Event       string `json:"event"`
}

// PipelineCommitInfo represents commit information
type PipelineCommitInfo struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
	Author  string `json:"author"`
}

func ShowPipelineStatus() int {
	// Parse flags
	ref := "main"
	commitSHA := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--ref" && i+1 < len(os.Args) {
			ref = os.Args[i+1]
			i++
		} else if arg == "--commit" && i+1 < len(os.Args) {
			commitSHA = os.Args[i+1]
			i++
		}
	}

	// Get commit SHA if not specified
	if commitSHA == "" {
		sha, err := getHeadCommit(ref)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get HEAD commit for %s: %v\n", ref, err)
			return 1
		}
		commitSHA = sha
	}

	// Get commit info
	commitInfo, err := getPipelineCommitInfo(commitSHA)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to get commit info: %v\n", err)
		commitInfo = &PipelineCommitInfo{SHA: commitSHA, Message: "(unknown)", Author: "(unknown)"}
	}

	// Print header
	fmt.Printf("Pipeline Status: %s\n", ref)
	fmt.Printf("Commit: %s\n", commitSHA[:7])
	// Truncate message at first newline
	message := commitInfo.Message
	if idx := strings.Index(message, "\n"); idx > 0 {
		message = message[:idx]
	}
	if len(message) > 60 {
		message = message[:57] + "..."
	}
	fmt.Printf("Message: %s\n", message)
	fmt.Println()

	// Get all workflow runs for this commit
	runs, err := getWorkflowRunsForCommit(commitSHA)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get workflow runs: %v\n", err)
		return 1
	}

	if len(runs) == 0 {
		fmt.Println("No workflow runs found for this commit.")
		fmt.Println()
		fmt.Println("This could mean:")
		fmt.Println("  - The commit was just pushed and workflows haven't started yet")
		fmt.Println("  - No workflows are configured to run on this commit")
		fmt.Println("  - The commit was skipped by workflow filters")
		return 0
	}

	// Group runs by workflow name (keep only most recent per workflow)
	latestRuns := make(map[string]*PipelineWorkflowRunInfo)
	for i := range runs {
		run := &runs[i]
		existing, exists := latestRuns[run.Name]
		if !exists || run.CreatedAt > existing.CreatedAt {
			latestRuns[run.Name] = run
		}
	}

	// Print status table
	fmt.Printf("%-40s  %-12s  %s\n", "Workflow", "Status", "Conclusion")
	fmt.Printf("%-40s  %-12s  %s\n", strings.Repeat("-", 40), strings.Repeat("-", 12), strings.Repeat("-", 12))

	allSuccess := true
	anyFailed := false
	anyRunning := false

	for _, run := range latestRuns {
		icon := getStatusIcon(run.Status, run.Conclusion)
		conclusion := run.Conclusion
		if conclusion == "" {
			conclusion = "-"
		}

		// Track overall status
		if run.Status != "completed" {
			anyRunning = true
			allSuccess = false
		} else if run.Conclusion != "success" {
			anyFailed = true
			allSuccess = false
		}

		fmt.Printf("%s %-38s  %-12s  %s\n", icon, truncate(run.Name, 38), run.Status, conclusion)
	}

	fmt.Println()

	// Print summary
	if allSuccess {
		fmt.Println("✅ All workflows passed")
		return 0
	} else if anyFailed {
		fmt.Println("❌ One or more workflows failed")
		return 1
	} else if anyRunning {
		fmt.Println("◐ Workflows in progress")
		return 0
	}

	return 0
}

// getHeadCommit gets the HEAD commit SHA for a ref
func getHeadCommit(ref string) (string, error) {
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/{owner}/{repo}/commits/%s", ref), "--jq", ".sha")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to git
		cmd = exec.Command("git", "rev-parse", ref)
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get commit: %w", err)
		}
	}
	return strings.TrimSpace(string(output)), nil
}

// getPipelineCommitInfo gets commit information
func getPipelineCommitInfo(sha string) (*PipelineCommitInfo, error) {
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/{owner}/{repo}/commits/%s", sha),
		"--jq", `{sha: .sha, message: .commit.message, author: .commit.author.name}`)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %w", err)
	}

	var info PipelineCommitInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse commit info: %w", err)
	}

	return &info, nil
}

// getWorkflowRunsForCommit gets all workflow runs for a specific commit
func getWorkflowRunsForCommit(sha string) ([]PipelineWorkflowRunInfo, error) {
	cmd := exec.Command("gh", "run", "list",
		"--commit", sha,
		"--json", "databaseId,name,displayTitle,status,conclusion,headSha,url,createdAt,updatedAt,event",
		"--limit", "50",
	)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list runs: %w", err)
	}

	var runs []PipelineWorkflowRunInfo
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse runs: %w", err)
	}

	return runs, nil
}

// getStatusIcon returns an icon for the workflow status
func getStatusIcon(status, conclusion string) string {
	if status == "completed" {
		switch conclusion {
		case "success":
			return "✓"
		case "failure":
			return "✗"
		case "cancelled":
			return "⊘"
		case "skipped":
			return "⊘"
		default:
			return "?"
		}
	}

	switch status {
	case "in_progress":
		return "◐"
	case "queued":
		return "○"
	case "waiting":
		return "○"
	default:
		return "?"
	}
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
