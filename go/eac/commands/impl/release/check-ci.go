// Command: release check-ci
// Short: Check CI status for a commit before releasing
// Long: Waits for a successful CI workflow run on a specific commit.
// Long:
// Long: This command polls GitHub Actions to verify that a CI workflow has
// Long: completed successfully for the given commit SHA. It's used by release
// Long: workflows to ensure code is tested before releasing.
// Long:
// Long: Exit codes:
// Long:   0 - CI workflow succeeded
// Long:   1 - CI workflow failed or timeout
// Long:
// Long: Example:
// Long:   release check-ci --workflow ci-r2r-cli.yaml --commit abc123
// Long:   release check-ci --workflow ci-ext-eac.yaml --commit abc123 --timeout 600
// Flag.workflow: type=string, usage=CI workflow filename (e.g., ci-r2r-cli.yaml)
// Flag.commit: type=string, usage=Commit SHA to check
// Flag.timeout: type=int, usage=Maximum wait time in seconds (default: 300)
// Flag.interval: type=int, usage=Poll interval in seconds (default: 15)
package release

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ReleaseCheckCI)
}

// CIRunStatus represents the status of a GitHub Actions workflow run
type CIRunStatus struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

// CheckCIResult represents the result of the CI check
type CheckCIResult struct {
	Success   bool   `json:"success"`
	Status    string `json:"status"` // "success", "failure", "timeout", "not_found"
	Message   string `json:"message"`
	CommitSHA string `json:"commit_sha"`
	Workflow  string `json:"workflow"`
	Elapsed   string `json:"elapsed"`
}

func ReleaseCheckCI() int {
	// Parse flags
	workflow := ""
	commitSHA := ""
	timeout := 300
	interval := 15

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "--workflow":
			if i+1 < len(os.Args) {
				workflow = os.Args[i+1]
				i++
			}
		case "--commit":
			if i+1 < len(os.Args) {
				commitSHA = os.Args[i+1]
				i++
			}
		case "--timeout":
			if i+1 < len(os.Args) {
				if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
					timeout = v
				}
				i++
			}
		case "--interval":
			if i+1 < len(os.Args) {
				if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
					interval = v
				}
				i++
			}
		}
	}

	if workflow == "" {
		log.Errorf("Error: --workflow is required")
		return 1
	}
	if commitSHA == "" {
		log.Errorf("Error: --commit is required")
		return 1
	}

	// Extract module name from workflow for display
	moduleName := strings.TrimPrefix(workflow, "ci-")
	moduleName = strings.TrimSuffix(moduleName, ".yaml")
	moduleName = strings.TrimSuffix(moduleName, ".yml")

	log.Infof("Checking CI for %s @ %s", moduleName, commitSHA[:7])

	startTime := time.Now()
	lastStatus := ""
	noCIFoundCount := 0
	chainCompletedCount := 0
	const noCIFoundThreshold = 4      // After 4 checks (~60s), fail fast if no CI exists
	const chainCompletedThreshold = 2 // After 2 checks (~30s), confirm chain really completed (runner startup buffer)

	for {
		elapsed := time.Since(startTime)
		if elapsed.Seconds() > float64(timeout) {
			log.Infof("")
			log.Infof("✗ Timeout after %v", elapsed.Round(time.Second))
			log.Infof("  Trigger CI: gh workflow run %s", workflow)
			return 1
		}

		elapsedStr := formatElapsed(elapsed)

		// ===========================================
		// Question 1: Has our MODULE CI completed?
		// ===========================================
		moduleStatus, err := getModuleCIStatus(workflow, commitSHA)
		if err != nil {
			log.Errorf("Warning: failed to query module CI: %v", err)
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		// Check module CI result
		if moduleStatus.success {
			log.Infof("\r⏱ %s  ✓ CI passed", elapsedStr)
			return 0
		}
		if moduleStatus.failed {
			log.Infof("\r⏱ %s  ✗ CI failed", elapsedStr)
			return 1
		}
		if moduleStatus.running {
			// Module CI is running - just wait
			currentStatus := fmt.Sprintf("⏱ %s  ◐ %s running", elapsedStr, moduleName)
			if currentStatus != lastStatus {
				log.Infof("\r%s", currentStatus)
				lastStatus = currentStatus
			}
			noCIFoundCount = 0
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		// ===========================================
		// Question 2: Should we keep waiting?
		// Module CI not found - check if CI chain is in progress
		// ===========================================
		chainStatus, chainErr := getCIChainStatus(commitSHA)
		if chainErr != nil {
			log.Errorf("Warning: failed to check CI chain: %v", chainErr)
		}

		var currentStatus string
		if chainStatus.running {
			// CI chain is running, our module hasn't started yet
			currentStatus = fmt.Sprintf("⏱ %s  ○ Waiting (CI chain in progress)", elapsedStr)
			noCIFoundCount = 0      // Reset - CI is happening
			chainCompletedCount = 0 // Reset - chain is still running
		} else if chainStatus.completed {
			// CI chain appears completed but our module CI wasn't triggered
			// Wait a bit longer to account for GitHub runner startup delays
			chainCompletedCount++
			if chainCompletedCount >= chainCompletedThreshold {
				// Confirmed: chain completed but module wasn't in changed set
				log.Infof("")
				log.Infof("✗ CI chain completed but %s was not tested", moduleName)
				log.Infof("")
				log.Infof("  The module may not have been in the changed set.")
				log.Infof("  Manually trigger: gh workflow run %s", workflow)
				return 1
			}
			currentStatus = fmt.Sprintf("⏱ %s  ○ Waiting (runner startup)", elapsedStr)
			noCIFoundCount = 0 // Reset - we saw CI activity
		} else {
			// No CI chain found at all
			noCIFoundCount++
			if noCIFoundCount >= noCIFoundThreshold {
				log.Infof("")
				log.Infof("✗ No CI found for commit %s", commitSHA[:7])
				log.Infof("")
				log.Infof("  CI must run before release. Options:")
				log.Infof("  1. Push changes to trigger CI via change-trigger workflow")
				log.Infof("  2. Manually trigger: gh workflow run %s", workflow)
				return 1
			}
			currentStatus = fmt.Sprintf("⏱ %s  ○ Waiting for CI", elapsedStr)
		}

		if currentStatus != lastStatus {
			log.Infof("\r%s", currentStatus)
			lastStatus = currentStatus
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// ModuleCIStatus represents the status of the target module's CI
type ModuleCIStatus struct {
	success bool
	failed  bool
	running bool
}

// getModuleCIStatus checks if the specific module CI has run for the commit
func getModuleCIStatus(workflow, commitSHA string) (ModuleCIStatus, error) {
	runs, err := getWorkflowRuns(workflow, commitSHA)
	if err != nil {
		return ModuleCIStatus{}, err
	}

	var status ModuleCIStatus
	for _, run := range runs {
		switch run.Status {
		case "completed":
			if run.Conclusion == "success" {
				status.success = true
				return status, nil // Success - done
			} else {
				status.failed = true
			}
		case "in_progress", "queued":
			status.running = true
		}
	}

	return status, nil
}

// CIChainStatus represents the status of the overall CI chain
type CIChainStatus struct {
	running   bool
	completed bool
}

// getCIChainStatus checks if any CI workflow is running/completed that covers our commit
func getCIChainStatus(commitSHA string) (CIChainStatus, error) {
	var status CIChainStatus

	// Check all recent CI runs on main
	allRuns, err := queryAllRecentRuns("main", 30)
	if err != nil {
		return status, err
	}

	for _, run := range allRuns {
		// Only consider CI-related workflows
		if !strings.HasPrefix(run.WorkflowName, "ci-") && !strings.Contains(run.WorkflowName, "CI") {
			continue
		}

		// Check if this run covers our commit
		if !ciRunCoversCommit(run.HeadSHA, commitSHA) {
			continue
		}

		// This run is relevant to our commit
		if run.Status == "in_progress" || run.Status == "queued" {
			status.running = true
			return status, nil // If anything is running, we should wait
		}
		if run.Status == "completed" {
			status.completed = true
			// Don't return yet - keep checking for running ones
		}
	}

	return status, nil
}

// CIRunWithSHA includes headSha for commit matching
type CIRunWithSHA struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"headSha"`
	HeadBranch string `json:"headBranch"`
}

// getWorkflowRuns queries GitHub for workflow runs on a specific commit
// It first tries exact commit match, then falls back to checking recent runs
// to handle cases where CI headSha differs from the target commit
func getWorkflowRuns(workflow, commitSHA string) ([]CIRunStatus, error) {
	// First try exact commit match (fast path)
	runs, err := queryRunsByCommit(workflow, commitSHA)
	if err != nil {
		return nil, err
	}
	if len(runs) > 0 {
		return runs, nil
	}

	// If no exact match, check if target commit is ancestor of any recent successful run
	// This handles the case where CI ran after more commits were pushed
	return queryRecentRunsForCommit(workflow, commitSHA)
}

// queryRunsByCommit queries runs filtered by exact commit SHA
func queryRunsByCommit(workflow, commitSHA string) ([]CIRunStatus, error) {
	cmd := exec.Command("gh", "run", "list",
		"--commit", commitSHA,
		"--workflow", workflow,
		"--json", "status,conclusion",
		"--limit", "5",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh run list failed: %w", err)
	}

	var runs []CIRunStatus
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return runs, nil
}

// queryRecentRunsForCommit checks if targetCommit is an ancestor of any recent CI run
func queryRecentRunsForCommit(workflow, targetCommit string) ([]CIRunStatus, error) {
	// Get recent runs without commit filter
	cmd := exec.Command("gh", "run", "list",
		"--workflow", workflow,
		"--branch", "main",
		"--json", "status,conclusion,headSha",
		"--limit", "10",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh run list failed: %w", err)
	}

	var runs []CIRunWithSHA
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check each run to see if targetCommit is ancestor of its headSha
	var matchingRuns []CIRunStatus
	for _, run := range runs {
		if run.HeadSHA == targetCommit {
			// Exact match
			matchingRuns = append(matchingRuns, CIRunStatus{
				Status:     run.Status,
				Conclusion: run.Conclusion,
			})
		} else if run.Status == "completed" && run.Conclusion == "success" {
			// Check if targetCommit is ancestor of this successful run
			if isAncestor(targetCommit, run.HeadSHA) {
				matchingRuns = append(matchingRuns, CIRunStatus{
					Status:     run.Status,
					Conclusion: run.Conclusion,
				})
			}
		}
	}

	return matchingRuns, nil
}

// isAncestor checks if potentialAncestor is an ancestor of commit
func isAncestor(potentialAncestor, commit string) bool {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", potentialAncestor, commit)
	err := cmd.Run()
	return err == nil
}

// ciRunCoversCommit returns true if a CI run with the given headSHA covers targetCommit
// This is true when:
// - headSHA == targetCommit (exact match)
// - targetCommit is ancestor of headSHA (CI is testing newer code that includes targetCommit)
func ciRunCoversCommit(headSHA, targetCommit string) bool {
	if headSHA == targetCommit {
		return true
	}
	// Check if targetCommit is ancestor of headSHA
	// (meaning the CI run includes changes from targetCommit)
	return isAncestor(targetCommit, headSHA)
}

// CIRunWithWorkflow includes workflow name for chain detection
type CIRunWithWorkflow struct {
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	HeadSHA      string `json:"headSha"`
	WorkflowName string `json:"workflowName"`
}

// queryAllRecentRuns queries recent runs across all workflows on a branch
func queryAllRecentRuns(branch string, limit int) ([]CIRunWithWorkflow, error) {
	cmd := exec.Command("gh", "run", "list",
		"--branch", branch,
		"--json", "status,conclusion,headSha,workflowName",
		"--limit", fmt.Sprintf("%d", limit),
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh run list failed: %w", err)
	}

	var runs []CIRunWithWorkflow
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return runs, nil
}

// formatElapsed formats elapsed time as M:SS
func formatElapsed(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
