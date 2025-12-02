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

	for {
		elapsed := time.Since(startTime)
		if elapsed.Seconds() > float64(timeout) {
			log.Infof("")
			log.Infof("✗ Timeout after %v", elapsed.Round(time.Second))
			log.Infof("  Trigger CI: gh workflow run %s", workflow)
			return 1
		}

		// Query GitHub for workflow runs on this commit
		runs, err := getWorkflowRuns(workflow, commitSHA)
		if err != nil {
			log.Errorf("Warning: failed to query runs: %v", err)
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		// Count statuses
		var successCount, failedCount, runningCount int
		for _, run := range runs {
			if run.Status == "completed" {
				if run.Conclusion == "success" {
					successCount++
				} else {
					failedCount++
				}
			} else if run.Status == "in_progress" || run.Status == "queued" {
				runningCount++
			}
		}

		// Determine current state and print status
		var currentStatus string
		elapsedStr := formatElapsed(elapsed)

		if successCount > 0 {
			log.Infof("\r⏱ %s  ✓ CI passed", elapsedStr)
			return 0
		} else if failedCount > 0 {
			log.Infof("\r⏱ %s  ✗ CI failed", elapsedStr)
			return 1
		} else if runningCount > 0 {
			currentStatus = fmt.Sprintf("⏱ %s  ◐ CI running", elapsedStr)
		} else {
			currentStatus = fmt.Sprintf("⏱ %s  ○ Waiting for CI", elapsedStr)
		}

		// Update status line (only if changed to reduce flicker)
		if currentStatus != lastStatus {
			log.Infof("\r%s", currentStatus)
			lastStatus = currentStatus
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// getWorkflowRuns queries GitHub for workflow runs on a specific commit
func getWorkflowRuns(workflow, commitSHA string) ([]CIRunStatus, error) {
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

// formatElapsed formats elapsed time as M:SS
func formatElapsed(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
