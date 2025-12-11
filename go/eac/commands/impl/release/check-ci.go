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
// Long:   release check-ci --workflow ci-r2r-cli.yaml --commit abc123 --strict
// Flag.workflow: type=string, usage=CI workflow filename (e.g., ci-r2r-cli.yaml)
// Flag.commit: type=string, usage=Commit SHA to check
// Flag.timeout: type=int, usage=Maximum wait time in seconds (default: 300)
// Flag.interval: type=int, usage=Poll interval in seconds (default: 15)
// Flag.strict: type=bool, usage=Require exact commit match (no ancestor check)
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
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ReleaseCheckCI)
}

// CIRunStatus represents the status of a GitHub Actions workflow run
type CIRunStatus struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

func ReleaseCheckCI() int {
	// Parse flags
	workflow := ""
	commitSHA := ""
	timeout := 300
	interval := 15
	strict := false

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
		case "--strict":
			strict = true
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

	if strict {
		log.Infof("Checking CI for %s @ %s (strict mode)", moduleName, commitSHA[:7])
	} else {
		log.Infof("Checking CI for %s @ %s", moduleName, commitSHA[:7])
	}

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
		moduleStatus, err := getModuleCIStatus(workflow, commitSHA, strict)
		if err != nil {
			log.Errorf("Warning: failed to query module CI: %v", err)
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		// Check module CI result
		if moduleStatus.success {
			log.Infof("⏱ %s  ✓ CI passed", elapsedStr)
			return 0
		}
		if moduleStatus.failed {
			log.Infof("⏱ %s  ✗ CI failed", elapsedStr)
			return 1
		}
		if moduleStatus.running {
			// Module CI is running - just wait
			currentStatus := fmt.Sprintf("⏱ %s  ◐ %s running", elapsedStr, moduleName)
			if currentStatus != lastStatus {
				log.Infof("%s", currentStatus)
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
				// Check if we can inherit CI from previous successful run
				// (safe when only changelog changed for this module)
				workspaceRoot, rootErr := repository.GetRepositoryRoot("")
				if rootErr == nil {
					canInherit, inheritedSHA, inheritMsg := canInheritCIFromPrevious(moduleName, workflow, commitSHA, workspaceRoot)
					if canInherit {
						log.Infof("")
						log.Infof("✓ %s", inheritMsg)
						// Export INHERITED_CI_SHA to GitHub Actions environment
						if ghEnv := os.Getenv("GITHUB_ENV"); ghEnv != "" {
							if f, err := os.OpenFile(ghEnv, os.O_APPEND|os.O_WRONLY, 0644); err == nil {
								fmt.Fprintf(f, "INHERITED_CI_SHA=%s\n", inheritedSHA)
								f.Close()
							}
						}
						return 0
					}
					// If we have a specific reason why we can't inherit, show it
					if inheritMsg != "" {
						log.Infof("")
						log.Infof("✗ CI chain completed but %s was not tested", moduleName)
						log.Infof("")
						log.Infof("  %s", inheritMsg)
						log.Infof("  Manually trigger: gh workflow run %s", workflow)
						return 1
					}
				}
				// Default failure message
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
			log.Infof("%s", currentStatus)
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
// If strict is true, only exact commit matches are accepted (no ancestor check)
func getModuleCIStatus(workflow, commitSHA string, strict bool) (ModuleCIStatus, error) {
	runs, err := getWorkflowRuns(workflow, commitSHA, strict)
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
}

// getWorkflowRuns queries GitHub for workflow runs on a specific commit
// If strict is true, only exact commit matches are accepted
// If strict is false, falls back to checking if commit is ancestor of recent runs
func getWorkflowRuns(workflow, commitSHA string, strict bool) ([]CIRunStatus, error) {
	// First try exact commit match (fast path)
	runs, err := queryRunsByCommit(workflow, commitSHA)
	if err != nil {
		return nil, err
	}
	if len(runs) > 0 {
		return runs, nil
	}

	// In strict mode, only accept exact matches - no ancestor fallback
	if strict {
		return nil, nil
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

// canInheritCIFromPrevious checks if we can safely inherit CI status from a previous
// successful run. This is safe when the only module-owned files that changed since
// the last successful CI are changelog files.
//
// Returns (canInherit bool, inheritedSHA string, message string)
// - canInherit=true, inheritedSHA=sha, message=success message: Safe to inherit
// - canInherit=false, inheritedSHA="", message="": No previous CI or error
// - canInherit=false, inheritedSHA="", message=reason: Specific reason we can't inherit
func canInheritCIFromPrevious(moduleName, workflow, releaseCommit, workspaceRoot string) (bool, string, string) {
	// 1. Get the last successful CI SHA for this module's workflow
	lastCISHA, err := getLastSuccessfulModuleCISHA(workflow, "main", workspaceRoot)
	if err != nil || lastCISHA == "" {
		return false, "", "" // No previous successful CI
	}

	// 2. Get changed files between last CI and release commit
	changedFiles, err := getChangedFilesBetweenCommits(lastCISHA, releaseCommit, workspaceRoot)
	if err != nil {
		return false, "", "" // Can't determine changes
	}

	if len(changedFiles) == 0 {
		// No files changed - this shouldn't happen but is safe
		return true, lastCISHA, fmt.Sprintf("CI inherited from %s (no files changed)", lastCISHA[:7])
	}

	// 3. Load module registry to determine file ownership and changelog path
	registry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return false, "", "" // Can't load module registry
	}

	// Find the module contract
	module, found := registry.Get(moduleName)
	if !found {
		return false, "", "" // Module not found
	}

	// 4. Get the changelog path for this module (normalized)
	changelogPath := normalizeSlashes(module.GetChangelogPath())

	// 5. Filter changed files to only those owned by this module
	moduleChangedFiles := []string{}
	for _, file := range changedFiles {
		normalizedFile := normalizeSlashes(file)
		if module.MatchesFile(normalizedFile) {
			moduleChangedFiles = append(moduleChangedFiles, normalizedFile)
		}
	}

	// 6. Check if all module changes are just the changelog
	if len(moduleChangedFiles) == 0 {
		// No module files changed - safe to inherit
		return true, lastCISHA, fmt.Sprintf("CI inherited from %s (no module files changed)", lastCISHA[:7])
	}

	// Check each changed file - must be the changelog
	nonChangelogFiles := []string{}
	for _, file := range moduleChangedFiles {
		if file != changelogPath {
			nonChangelogFiles = append(nonChangelogFiles, file)
		}
	}

	if len(nonChangelogFiles) > 0 {
		// Module source files changed - cannot inherit
		if len(nonChangelogFiles) == 1 {
			return false, "", fmt.Sprintf("Module file changed: %s", nonChangelogFiles[0])
		}
		return false, "", fmt.Sprintf("Module files changed: %s (and %d more)", nonChangelogFiles[0], len(nonChangelogFiles)-1)
	}

	// Only changelog changed - safe to inherit!
	return true, lastCISHA, fmt.Sprintf("CI inherited from %s (only changelog changed)", lastCISHA[:7])
}

// getLastSuccessfulModuleCISHA queries gh CLI for the last successful workflow run SHA
func getLastSuccessfulModuleCISHA(workflow, branch, workspaceRoot string) (string, error) {
	cmd := exec.Command("gh", "run", "list",
		"-b", branch,
		"-s", "success",
		"-w", workflow,
		"-L", "1",
		"--json", "headSha",
		"-q", ".[0].headSha",
	)
	cmd.Dir = workspaceRoot

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("gh command failed: %w", err)
	}

	sha := strings.TrimSpace(string(output))
	return sha, nil
}

// getChangedFilesBetweenCommits gets the list of files changed between two commits
func getChangedFilesBetweenCommits(baseSHA, headSHA, workspaceRoot string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", baseSHA+".."+headSHA)
	cmd.Dir = workspaceRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

// normalizeSlashes converts backslashes to forward slashes for consistent path comparison
func normalizeSlashes(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

