package release

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/repository"
)

type releaseCheckCICommand struct{}

var _ core.SimpleCommandPort = (*releaseCheckCICommand)(nil)

func (c *releaseCheckCICommand) Name() string { return "release check-ci" }

func (c *releaseCheckCICommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-check-ci",
		Short:         "Check CI status for a commit before releasing",
		Long: "Waits for a successful CI workflow run on a specific commit.\n\nThis command polls GitHub Actions to verify that a CI workflow has\ncompleted successfully for the given commit SHA. It's used by release\nworkflows to ensure code is tested before releasing.",
		Notes: "Expected Output:\n  - Exit code 0 if CI workflow succeeded\n  - Exit code 1 if CI workflow failed or timeout occurred\n\nExit codes:\n  0 - CI workflow succeeded\n  1 - CI workflow failed or timeout",
		Examples: []string{
			"eac release check-ci --workflow ci-clie.yaml --commit abc123",
			"eac release check-ci --workflow ci-eac-ext.yaml --commit abc123 --timeout 600",
			"eac release check-ci --workflow ci-clie.yaml --commit abc123 --strict",
		},
		Flags: []core.FlagSpec{
			{Name: "workflow", Type: "string", Usage: "CI workflow filename (e.g., ci-clie.yaml)"},
			{Name: "commit", Type: "string", Usage: "Commit SHA to check"},
			{Name: "timeout", Type: "int", Usage: "Maximum wait time in seconds (default: 300)"},
			{Name: "interval", Type: "int", Usage: "Poll interval in seconds (default: 15)"},
			{Name: "strict", Type: "bool", Usage: "Require exact commit match (no ancestor check)"},
		},
	}
}

func (c *releaseCheckCICommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleaseCheckCI()
}

// CIRunStatus represents the status of a GitHub Actions workflow run.
type CIRunStatus struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	RunID      int64  `json:"databaseId"`
	HeadSHA    string `json:"headSha"`
}

func ReleaseCheckCI() int {
	s, exitCode := newReleaseScaffold()
	if s == nil {
		return exitCode
	}

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
	inheritanceChecked := false       // Only check inheritance once
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
			// Export CI run info to GitHub Actions environment
			exportCIRunInfo(moduleStatus.runID, moduleStatus.ciSHA)
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
		// Question 2: Can we inherit from previous CI?
		// Try inheritance FIRST before waiting for CI chain (only check once)
		// ===========================================
		if !inheritanceChecked {
			inheritanceChecked = true
			workspaceRoot, rootErr := repository.GetRepositoryRoot("")
			if rootErr == nil {
				canInherit, inheritedCI, inheritMsg := canInheritCIFromPrevious(moduleName, workflow, commitSHA, workspaceRoot)
				if canInherit {
					log.Infof("")
					log.Infof("✓ %s", inheritMsg)
					// Export CI run info to GitHub Actions environment
					if ghEnv := os.Getenv("GITHUB_ENV"); ghEnv != "" {
						if f, err := os.OpenFile(ghEnv, os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
							fmt.Fprintf(f, "INHERITED_CI_SHA=%s\n", inheritedCI.SHA)
							f.Close()
						}
					}
					exportCIRunInfo(inheritedCI.RunID, inheritedCI.SHA)
					return 0
				}
				// If we have a specific reason why we can't inherit, fail fast
				if inheritMsg != "" {
					log.Infof("")
					log.Infof("✗ Cannot inherit CI: %s", inheritMsg)
					log.Infof("  Manually trigger: gh workflow run %s", workflow)
					return 1
				}
			}
		}

		// ===========================================
		// Question 3: Should we keep waiting?
		// Module CI not found and can't inherit - check if CI chain is in progress
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
			// CI chain completed but our module wasn't triggered and we couldn't inherit
			chainCompletedCount++
			if chainCompletedCount >= chainCompletedThreshold {
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

// ModuleCIStatus represents the status of the target module's CI.
type ModuleCIStatus struct {
	success bool
	failed  bool
	running bool
	runID   int64  // Run ID of the successful CI (if success=true)
	ciSHA   string // Commit SHA of the CI run (if success=true)
}

// getModuleCIStatus checks if the specific module CI has run for the commit
// If strict is true, only exact commit matches are accepted (no ancestor check).
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
				status.runID = run.RunID
				status.ciSHA = run.HeadSHA
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

// CIChainStatus represents the status of the overall CI chain.
type CIChainStatus struct {
	running   bool
	completed bool
}

// getCIChainStatus checks if any CI workflow is running/completed that covers our commit.
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

// getWorkflowRuns queries GitHub for workflow runs on a specific commit
// If strict is true, only exact commit matches are accepted
// If strict is false, falls back to checking if commit is ancestor of recent runs.
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

// queryRunsByCommit queries runs filtered by exact commit SHA.
func queryRunsByCommit(workflow, commitSHA string) ([]CIRunStatus, error) {
	output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", ".", "run", "list",
		"--commit", commitSHA,
		"--workflow", workflow,
		"--json", "status,conclusion,databaseId,headSha",
		"--limit", "5",
	)
	if err != nil {
		return nil, fmt.Errorf("gh run list failed: %w", err)
	}

	var runs []CIRunStatus
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return runs, nil
}

// queryRecentRunsForCommit checks if targetCommit is an ancestor of any recent CI run.
func queryRecentRunsForCommit(workflow, targetCommit string) ([]CIRunStatus, error) {
	// Get recent runs without commit filter
	output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", ".", "run", "list",
		"--workflow", workflow,
		"--branch", "main",
		"--json", "status,conclusion,headSha,databaseId",
		"--limit", "10",
	)
	if err != nil {
		return nil, fmt.Errorf("gh run list failed: %w", err)
	}

	var runs []CIRunStatus
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check each run to see if targetCommit is ancestor of its headSha
	var matchingRuns []CIRunStatus
	for _, run := range runs {
		if run.HeadSHA == targetCommit {
			// Exact match
			matchingRuns = append(matchingRuns, run)
		} else if run.Status == "completed" && run.Conclusion == "success" {
			// Check if targetCommit is ancestor of this successful run
			if isAncestor(targetCommit, run.HeadSHA) {
				matchingRuns = append(matchingRuns, run)
			}
		}
	}

	return matchingRuns, nil
}

// isAncestor checks if potentialAncestor is an ancestor of commit.
func isAncestor(potentialAncestor, commit string) bool {
	ts := tool.GlobalToolSystem()
	if ts == nil {
		return false
	}
	_, exitCode, err := ts.RunToolCombined(context.Background(), "git", ".", "merge-base", "--is-ancestor", potentialAncestor, commit)
	return err == nil && exitCode == 0
}

// ciRunCoversCommit returns true if a CI run with the given headSHA covers targetCommit
// This is true when:
// - headSHA == targetCommit (exact match)
// - targetCommit is ancestor of headSHA (CI is testing newer code that includes targetCommit).
func ciRunCoversCommit(headSHA, targetCommit string) bool {
	if headSHA == targetCommit {
		return true
	}
	// Check if targetCommit is ancestor of headSHA
	// (meaning the CI run includes changes from targetCommit)
	return isAncestor(targetCommit, headSHA)
}

// CIRunWithWorkflow includes workflow name for chain detection.
type CIRunWithWorkflow struct {
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	HeadSHA      string `json:"headSha"`
	WorkflowName string `json:"workflowName"`
}

// queryAllRecentRuns queries recent runs across all workflows on a branch.
func queryAllRecentRuns(branch string, limit int) ([]CIRunWithWorkflow, error) {
	output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", ".", "run", "list",
		"--branch", branch,
		"--json", "status,conclusion,headSha,workflowName",
		"--limit", fmt.Sprintf("%d", limit),
	)
	if err != nil {
		return nil, fmt.Errorf("gh run list failed: %w", err)
	}

	var runs []CIRunWithWorkflow
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return runs, nil
}

// formatElapsed formats elapsed time as M:SS.
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
// Returns (canInherit bool, inheritedInfo CIRunInfo, message string)
// - canInherit=true, inheritedInfo with SHA and RunID, message=success message: Safe to inherit
// - canInherit=false, empty inheritedInfo, message="": No previous CI or error
// - canInherit=false, empty inheritedInfo, message=reason: Specific reason we can't inherit.
func canInheritCIFromPrevious(moduleName, workflow, releaseCommit, workspaceRoot string) (bool, CIRunInfo, string) {
	// 1. Get the last successful CI info for this module's workflow
	lastCI, err := getLastSuccessfulModuleCIInfo(workflow, "main", workspaceRoot)
	if err != nil || lastCI.SHA == "" {
		return false, CIRunInfo{}, "" // No previous successful CI
	}

	// 2. Get changed files between last CI and release commit
	changedFiles, err := getChangedFilesBetweenCommits(lastCI.SHA, releaseCommit, workspaceRoot)
	if err != nil {
		return false, CIRunInfo{}, "" // Can't determine changes
	}

	if len(changedFiles) == 0 {
		// No files changed - this shouldn't happen but is safe
		return true, lastCI, fmt.Sprintf("CI inherited from %s (no files changed)", lastCI.SHA[:7])
	}

	// 3. Load module registry to determine file ownership and changelog path
	registry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return false, CIRunInfo{}, "" // Can't load module registry
	}

	// Find the module contract
	module, found := registry.Get(moduleName)
	if !found {
		return false, CIRunInfo{}, "" // Module not found
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
		return true, lastCI, fmt.Sprintf("CI inherited from %s (no module files changed)", lastCI.SHA[:7])
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
			return false, CIRunInfo{}, fmt.Sprintf("Module file changed: %s", nonChangelogFiles[0])
		}
		return false, CIRunInfo{}, fmt.Sprintf("Module files changed: %s (and %d more)", nonChangelogFiles[0], len(nonChangelogFiles)-1)
	}

	// Only changelog changed - safe to inherit!
	return true, lastCI, fmt.Sprintf("CI inherited from %s (only changelog changed)", lastCI.SHA[:7])
}

// CIRunInfo contains information about a CI run.
type CIRunInfo struct {
	SHA   string
	RunID int64
}

// getLastSuccessfulModuleCIInfo queries gh CLI for the last successful workflow run info.
func getLastSuccessfulModuleCIInfo(workflow, branch, workspaceRoot string) (CIRunInfo, error) {
	output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", workspaceRoot, "run", "list",
		"-b", branch,
		"-s", "success",
		"-w", workflow,
		"-L", "1",
		"--json", "headSha,databaseId",
	)
	if err != nil {
		return CIRunInfo{}, fmt.Errorf("gh command failed: %w", err)
	}

	var runs []struct {
		HeadSHA    string `json:"headSha"`
		DatabaseID int64  `json:"databaseId"`
	}
	if err := json.Unmarshal(output, &runs); err != nil {
		return CIRunInfo{}, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(runs) == 0 {
		return CIRunInfo{}, nil
	}

	return CIRunInfo{
		SHA:   runs[0].HeadSHA,
		RunID: runs[0].DatabaseID,
	}, nil
}

// getChangedFilesBetweenCommits gets the list of files changed between two commits.
func getChangedFilesBetweenCommits(baseSHA, headSHA, workspaceRoot string) ([]string, error) {
	ts := tool.GlobalToolSystem()
	if ts == nil {
		return nil, fmt.Errorf("tool system not initialized")
	}
	output, err := ts.RunTool(context.Background(), "git", workspaceRoot, "diff", "--name-only", baseSHA+".."+headSHA)
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

// normalizeSlashes converts backslashes to forward slashes for consistent path comparison.
func normalizeSlashes(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// exportCIRunInfo exports CI run information to GitHub Actions environment
// This allows workflows to access the run ID for artifact download.
func exportCIRunInfo(runID int64, ciSHA string) {
	ghEnv := os.Getenv("GITHUB_ENV")
	if ghEnv == "" {
		return // Not running in GitHub Actions
	}

	f, err := os.OpenFile(ghEnv, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	if runID > 0 {
		fmt.Fprintf(f, "CI_RUN_ID=%d\n", runID)
	}
	if ciSHA != "" {
		fmt.Fprintf(f, "CI_SHA=%s\n", ciSHA)
	}
}
