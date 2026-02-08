// Command: pipeline await-ci
// Short: Wait for CI workflows to complete for a specific commit
// Long: Wait for CI workflows to complete. Can wait by pattern+SHA or by run ID.
// Long:
// Long: Mode 1 - Pattern+SHA (default):
// Long:   Polls GitHub Actions for in_progress or queued CI workflow runs
// Long:   that match the specified SHA and waits until all complete.
// Long:
// Long: Mode 2 - Run ID:
// Long:   Wait for a specific workflow run to complete by its run ID.
// Long:
// Long: SHA Detection (in order of precedence):
// Long:   1. --sha flag (explicit override)
// Long:   2. GITHUB_SHA environment variable (GitHub Actions)
// Long:   3. git rev-parse HEAD (local development)
// Long:
// Long: Expected Output:
// Long:   - Live progress display showing active workflow count
// Long:   - Exit code 0 when all workflows complete successfully
// Long:   - Exit code 1 on timeout or failure
// Long:
// Long: Example:
// Long:   pipeline await-ci                              # Auto-detect SHA, all ci-*.yaml
// Long:   pipeline await-ci --sha abc123                 # Explicit SHA
// Long:   pipeline await-ci --pattern ci-clie-cli.yaml    # Specific workflow
// Long:   pipeline await-ci --run-id 12345               # Wait for specific run
// Long:   pipeline await-ci --timeout 600                # 10 minute timeout
// Long:   pipeline await-ci --exclude ci-foo             # Exclude workflow
// Flag.sha: type=string, usage=Commit SHA to filter runs (auto-detected if not provided)
// Flag.run-id: type=string, usage=Specific workflow run ID to wait for (alternative to pattern+sha)
// Flag.timeout: type=int, default=1800, usage=Maximum wait time in seconds (default: 1800)
// Flag.interval: type=int, default=30, usage=Poll interval in seconds (default: 30)
// Flag.pattern: type=string, default=ci-*.yaml, usage=Workflow file pattern to match
// Flag.exclude: type=string, usage=Workflow name substring to exclude (e.g., clie-eac-bundle)
package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get"
	"github.com/ready-to-release/eac/go/clibase/ghexec"
	"github.com/ready-to-release/eac/go/core/repository"
)

func PipelineAwaitCI() int {
	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Parse flags
	sha := ""
	runID := ""
	timeout := 1800 // 30 minutes default
	interval := 30  // 30 seconds default
	pattern := "ci-*.yaml"
	exclude := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--sha" && i+1 < len(os.Args):
			sha = os.Args[i+1]
			i++
		case arg == "--run-id" && i+1 < len(os.Args):
			runID = os.Args[i+1]
			i++
		case arg == "--timeout" && i+1 < len(os.Args):
			if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
				timeout = v
			}
			i++
		case arg == "--interval" && i+1 < len(os.Args):
			if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
				interval = v
			}
			i++
		case arg == "--pattern" && i+1 < len(os.Args):
			pattern = os.Args[i+1]
			i++
		case arg == "--exclude" && i+1 < len(os.Args):
			exclude = os.Args[i+1]
			i++
		}
	}

	// Mode 2: Wait for specific run ID
	if runID != "" {
		return awaitRunByID(runID, timeout, interval)
	}

	// Mode 1: Wait by pattern+SHA
	// Auto-detect SHA using shared detection logic
	result, err := get.DetectCurrentSHA(workspaceRoot, sha)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	sha = result.SHA

	return awaitWorkflows(workspaceRoot, pattern, exclude, sha, timeout, interval, "CI")
}

// awaitWorkflows waits for all matching workflows to complete for a specific SHA.
func awaitWorkflows(workspaceRoot, pattern, exclude, sha string, timeout, interval int, workflowType string) int {
	log.Infof("Waiting for %s workflows to complete...", workflowType)

	// Parse comma-separated exclude list
	var excludeList []string
	if exclude != "" {
		for _, e := range strings.Split(exclude, ",") {
			if trimmed := strings.TrimSpace(e); trimmed != "" {
				excludeList = append(excludeList, trimmed)
			}
		}
		log.Infof("  Excluding: %s", strings.Join(excludeList, ", "))
	}
	log.Info("")

	startTime := time.Now()
	workflowsDir := filepath.Join(workspaceRoot, ".github", "workflows")

	for {
		elapsed := time.Since(startTime)
		if elapsed.Seconds() >= float64(timeout) {
			log.Warnf("Timeout waiting for %s workflows after %v", workflowType, elapsed.Round(time.Second))
			return 1
		}

		// Find all matching workflow files
		matches, err := filepath.Glob(filepath.Join(workflowsDir, pattern))
		if err != nil {
			log.Errorf("Error matching workflow pattern: %v", err)
			return 1
		}

		// Check each workflow for active runs matching our SHA
		activeCount := 0
		failedCount := 0
		var activeWorkflows []string

		for _, wfPath := range matches {
			wfName := filepath.Base(wfPath)

			// Skip excluded workflows (exact match on filename)
			excluded := false
			for _, e := range excludeList {
				if wfName == e {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}

			status, count := getRunStatusForSHA(wfName, sha)
			switch status {
			case "active":
				activeCount += count
				activeWorkflows = append(activeWorkflows, wfName)
			case "failed":
				failedCount++
			}
		}

		if activeCount == 0 {
			if failedCount > 0 {
				log.Warnf("❌ %d %s workflow(s) failed", failedCount, workflowType)
				return 1
			}
			log.Infof("✅ All %s workflows completed", workflowType)
			return 0
		}

		log.Infof("  Waiting for %d active %s workflow(s)... (%v)", activeCount, workflowType, elapsed.Round(time.Second))
		for _, wf := range activeWorkflows {
			log.Infof("    - %s", wf)
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// runInfo holds workflow run status from GitHub.
type runInfo struct {
	HeadSHA    string `json:"headSha"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

// getRunStatusForSHA checks workflow runs for a specific SHA
// Returns: status ("active", "success", "failed", "none"), count of active runs
//
// Important: Only the MOST RECENT run for this SHA determines success/failure.
// If a workflow was re-run after a failure, only the latest run matters.
// GitHub returns runs in order (most recent first), so we track which runs
// we've already seen for this SHA and only count the first completed one.
func getRunStatusForSHA(workflowName, sha string) (string, int) {
	// Get recent runs for this workflow
	output, err := ghexec.Run(".", "run", "list", "-w", workflowName, "--limit", "20",
		"--json", "headSha,status,conclusion")
	if err != nil {
		return "none", 0
	}

	var runs []runInfo
	if err := json.Unmarshal(output, &runs); err != nil {
		return "none", 0
	}

	// Find runs matching our SHA
	// Count all active runs (we want to wait for all of them)
	// But for completed runs, only consider the most recent one
	// (GitHub returns runs in reverse chronological order)
	activeCount := 0
	foundCompleted := false
	mostRecentFailed := false

	for _, run := range runs {
		if run.HeadSHA != sha {
			continue
		}

		switch run.Status {
		case "in_progress", "queued", "waiting", "pending":
			activeCount++
		case "completed":
			// Only the first (most recent) completed run determines success/failure
			if !foundCompleted {
				foundCompleted = true
				if run.Conclusion != "success" && run.Conclusion != "skipped" {
					mostRecentFailed = true
				}
			}
			// Ignore older completed runs - they may be failed retries
		}
	}

	if activeCount > 0 {
		return "active", activeCount
	}
	if mostRecentFailed {
		return "failed", 0
	}
	if foundCompleted {
		return "success", 0
	}
	return "none", 0
}

// runByIDInfo holds workflow run info for a specific run ID.
type runByIDInfo struct {
	DatabaseID   int64  `json:"databaseId"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	WorkflowName string `json:"workflowName"`
}

// awaitRunByID waits for a specific workflow run to complete.
func awaitRunByID(runID string, timeout, interval int) int {
	log.Infof("Waiting for workflow run %s to complete...", runID)
	log.Info("")

	startTime := time.Now()

	for {
		elapsed := time.Since(startTime)
		if elapsed.Seconds() >= float64(timeout) {
			log.Warnf("Timeout waiting for run %s after %v", runID, elapsed.Round(time.Second))
			return 1
		}

		// Get run status
		output, err := ghexec.Run(".", "run", "view", runID, "--json", "databaseId,status,conclusion,workflowName")
		if err != nil {
			log.Errorf("Error getting run status: %v", err)
			return 1
		}

		var run runByIDInfo
		if err := json.Unmarshal(output, &run); err != nil {
			log.Errorf("Error parsing run status: %v", err)
			return 1
		}

		switch run.Status {
		case "completed":
			if run.Conclusion == "success" || run.Conclusion == "skipped" {
				log.Infof("✅ Workflow %s completed successfully", run.WorkflowName)
				return 0
			}
			log.Warnf("❌ Workflow %s failed with conclusion: %s", run.WorkflowName, run.Conclusion)
			return 1
		case "in_progress", "queued", "waiting", "pending":
			log.Infof("  Run %s (%s) is %s... (%v)", runID, run.WorkflowName, run.Status, elapsed.Round(time.Second))
		default:
			log.Warnf("  Run %s has unexpected status: %s", runID, run.Status)
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
