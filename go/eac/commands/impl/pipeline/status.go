// Command: pipeline status
// Short: Show CI status for the head of trunk
// Long: Show the CI pipeline status for the head of the main branch.
// Long:
// Long: This command queries GitHub Actions to display the status of all CI
// Long: workflows for the latest commit on the main branch. It shows a compact
// Long: summary of each workflow's status.
// Long:
// Long: Expected Output:
// Long:   - Compact summary of all workflow statuses
// Long:   - Status icons (✓, ✗, ◐, ○) and workflow names
// Long:   - Exit code 0 if all succeed, 1 if any fail
// Long:
// Long: Example:
// Long:   pipeline status              # Show status for main branch HEAD
// Long:   pipeline status --ref dev    # Show status for dev branch HEAD
// Long:   pipeline status --commit abc # Show status for specific commit
// Flag.ref: type=string, usage=Git ref to check (default: main)
// Flag.commit: type=string, usage=Specific commit SHA to check
package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)


func init() {
	registry.Register(PipelineStatus)
}

// StatusWorkflowRunInfo represents workflow run information from GitHub
type StatusWorkflowRunInfo struct {
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

// StatusCommitInfo represents commit information
type StatusCommitInfo struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
	Author  string `json:"author"`
}

func PipelineStatus() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

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
		sha, err := getStatusHeadCommit(ref)
		if err != nil {
			log.Errorf("failed to get HEAD commit for %s: %v", ref, err)
			return 1
		}
		commitSHA = sha
	}

	// Get commit info
	commitInfo, err := getStatusCommitInfo(commitSHA)
	if err != nil {
		log.Errorf("Warning: failed to get commit info: %v", err)
		commitInfo = &StatusCommitInfo{SHA: commitSHA, Message: "(unknown)", Author: "(unknown)"}
	}

	// Print header
	log.Infof("Pipeline Status: %s", ref)
	log.Infof("Commit: %s", commitSHA[:7])
	// Truncate message at first newline
	message := commitInfo.Message
	if idx := strings.Index(message, "\n"); idx > 0 {
		message = message[:idx]
	}
	if len(message) > 60 {
		message = message[:57] + "..."
	}
	log.Infof("Message: %s", message)
	log.Info("")

	// Get all workflow runs for this commit
	runs, err := getStatusWorkflowRunsForCommit(commitSHA)
	if err != nil {
		log.Errorf("failed to get workflow runs: %v", err)
		return 1
	}

	if len(runs) == 0 {
		log.Info("No workflow runs found for this commit.")
		log.Info("")
		log.Info("This could mean:")
		log.Info("  - The commit was just pushed and workflows haven't started yet")
		log.Info("  - No workflows are configured to run on this commit")
		log.Info("  - The commit was skipped by workflow filters")
		return 0
	}

	// Group runs by workflow name (keep only most recent per workflow)
	latestRuns := make(map[string]*StatusWorkflowRunInfo)
	for i := range runs {
		run := &runs[i]
		existing, exists := latestRuns[run.Name]
		if !exists || run.CreatedAt > existing.CreatedAt {
			latestRuns[run.Name] = run
		}
	}

	// Print status table
	log.Infof("%-40s  %-12s  %s", "Workflow", "Status", "Conclusion")
	log.Infof("%-40s  %-12s  %s", strings.Repeat("-", 40), strings.Repeat("-", 12), strings.Repeat("-", 12))

	allSuccess := true
	anyFailed := false
	anyRunning := false

	for _, run := range latestRuns {
		icon := getStatusWorkflowIcon(run.Status, run.Conclusion)
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

		log.Infof("%s %-38s  %-12s  %s", icon, truncateStatus(run.Name, 38), run.Status, conclusion)
	}

	log.Info("")

	// Print summary
	if allSuccess {
		log.Info("✅ All workflows passed")
		return 0
	} else if anyFailed {
		log.Info("❌ One or more workflows failed")
		return 1
	} else if anyRunning {
		log.Info("◐ Workflows in progress")
		return 0
	}

	return 0
}

// getStatusHeadCommit gets the HEAD commit SHA for a ref
func getStatusHeadCommit(ref string) (string, error) {
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

// getStatusCommitInfo gets commit information
func getStatusCommitInfo(sha string) (*StatusCommitInfo, error) {
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/{owner}/{repo}/commits/%s", sha),
		"--jq", `{sha: .sha, message: .commit.message, author: .commit.author.name}`)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %w", err)
	}

	var info StatusCommitInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse commit info: %w", err)
	}

	return &info, nil
}

// getStatusWorkflowRunsForCommit gets all workflow runs for a specific commit
func getStatusWorkflowRunsForCommit(sha string) ([]StatusWorkflowRunInfo, error) {
	cmd := exec.Command("gh", "run", "list",
		"--commit", sha,
		"--json", "databaseId,name,displayTitle,status,conclusion,headSha,url,createdAt,updatedAt,event",
		"--limit", "50",
	)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list runs: %w", err)
	}

	var runs []StatusWorkflowRunInfo
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse runs: %w", err)
	}

	return runs, nil
}

// getStatusWorkflowIcon returns an icon for the workflow status
func getStatusWorkflowIcon(status, conclusion string) string {
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

// truncateStatus truncates a string to maxLen characters
func truncateStatus(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
