package release

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/tool"
)

type releaseAwaitDepsCommand struct{}

var _ core.SimpleCommandPort = (*releaseAwaitDepsCommand)(nil)

func (c *releaseAwaitDepsCommand) Name() string { return "release await-deps" }

func (c *releaseAwaitDepsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-await-deps",
		Short:         "Wait for dependency CI to pass before release",
		Long:          "Verifies all transitive dependencies have passing CI runs before allowing release.\n\nFor each dependency, finds the commit where it was last changed and verifies\nthat CI passed for that commit. Waits for in-progress CI runs with configurable\ntimeout. This ensures that releasing a module won't proceed if any of its\ndependencies have failing CI.\n\nExpected Output:\n  - Exit code 0 if all dependency CI checks passed\n  - Exit code 1 if any dependency CI failed or timeout occurred\n\nExample:\n  release await-deps eac-ext                    # Check deps for eac-ext\n  release await-deps eac-ext --timeout 600     # Wait up to 10 minutes\n  release await-deps eac-ext --skip-static     # Skip static modules (default)",
		Flags: []core.FlagSpec{
			{Name: "timeout", Type: "int", Usage: "Maximum wait time per dependency in seconds (default: 300)"},
			{Name: "interval", Type: "int", Usage: "Poll interval in seconds (default: 15)"},
			{Name: "skip-static", Type: "bool", Usage: "Skip static modules without CI workflows (default: true)"},
			{Name: "format", Type: "string", Usage: "Output format (text or shell for eval)"},
		},
	}
}

func (c *releaseAwaitDepsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleaseAwaitDeps()
}

// DepCIStatus represents the CI status for a dependency.
type DepCIStatus struct {
	Moniker       string
	LastCommit    string
	LastCommitMsg string
	CIWorkflow    string
	Status        string // "success", "failed", "running", "not_found", "skipped"
	RunID         int64
	RunURL        string
	SkipReason    string
}

func ReleaseAwaitDeps() int {
	s, exitCode := newReleaseScaffold(withModules())
	if s == nil {
		return exitCode
	}

	// Parse arguments
	module := ""
	timeout := 300
	interval := 15
	skipStatic := true
	format := "text"

	args := os.Args[3:] // Skip "commands release await-deps"
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--timeout":
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					timeout = v
				}
				i++
			}
		case "--interval":
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					interval = v
				}
				i++
			}
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--skip-static":
			skipStatic = true
		case "--no-skip-static":
			skipStatic = false
		case "--help", "-h":
			printAwaitDepsUsage()
			return 0
		default:
			if strings.HasPrefix(arg, "--format=") {
				format = strings.TrimPrefix(arg, "--format=")
			} else if !strings.HasPrefix(arg, "--") && module == "" {
				module = arg
			}
		}
	}

	if module == "" {
		log.Errorf("Error: module name is required")
		log.Errorf("Usage: release await-deps <module> [--timeout N] [--interval N]")
		return 1
	}

	workspaceRoot := s.WorkspaceRoot
	reg := s.ModuleRegistry

	// Verify module exists
	targetModule, exists := reg.Get(module)
	if !exists {
		log.Errorf("Error: module '%s' not found", module)
		return 1
	}

	// Get transitive dependencies
	deps := getTransitiveDeps(module, reg)
	if len(deps) == 0 {
		if format == "shell" {
			fmt.Printf("HAS_FAILED=%q\n", "false")
			fmt.Printf("PASSED_COUNT=%q\n", "0")
			fmt.Printf("SKIPPED_COUNT=%q\n", "0")
			fmt.Printf("DEPS_LIST=%q\n", "")
		} else {
			log.Infof("Module %s has no dependencies", module)
		}
		return 0
	}

	log.Infof("Awaiting CI for %s dependencies...", module)
	log.Infof("")
	log.Infof("Dependencies: %s", strings.Join(deps, ", "))
	log.Infof("")

	// Check each dependency
	results := make([]DepCIStatus, 0, len(deps))
	anyFailed := false

	for _, dep := range deps {
		depModule, _ := reg.Get(dep)

		// Check if we should skip this dep
		if skipStatic && shouldSkipModule(depModule, workspaceRoot) {
			status := DepCIStatus{
				Moniker:    dep,
				Status:     "skipped",
				SkipReason: getSkipReason(depModule, workspaceRoot),
			}
			results = append(results, status)
			log.Infof("  ○ %s (skipped: %s)", dep, status.SkipReason)
			continue
		}

		// Find last commit that changed this dependency
		lastCommit, commitMsg, err := findLastChangedCommit(depModule, workspaceRoot)
		if err != nil {
			log.Errorf("  ✗ %s: failed to find last changed commit: %v", dep, err)
			anyFailed = true
			results = append(results, DepCIStatus{
				Moniker: dep,
				Status:  "failed",
			})
			continue
		}

		workflow := fmt.Sprintf("ci-%s.yaml", dep)
		log.Infof("  Checking %s", dep)
		log.Infof("    Last changed: %s (%s)", lastCommit[:7], truncateString(commitMsg, 50))

		// Wait for CI on this commit
		status := waitForDepCI(dep, workflow, lastCommit, timeout, interval, targetModule.Moniker, workspaceRoot)
		status.Moniker = dep
		status.LastCommit = lastCommit
		status.LastCommitMsg = commitMsg
		status.CIWorkflow = workflow
		results = append(results, status)

		switch status.Status {
		case "success":
			log.Infof("    CI run: #%d ✓ passed", status.RunID)
		case "failed":
			log.Infof("    CI run: #%d ✗ failed", status.RunID)
			anyFailed = true
		case "timeout":
			log.Infof("    CI run: #%d ◐ still running (timeout)", status.RunID)
			anyFailed = true
		case "not_found":
			log.Infof("    CI: ✗ not found for commit %s", lastCommit[:7])
			anyFailed = true
		}
		log.Infof("")
	}

	// Count results
	passedCount := 0
	skippedCount := 0
	for _, r := range results {
		switch r.Status {
		case "success":
			passedCount++
		case "skipped":
			skippedCount++
		}
	}

	// Shell format output (for eval in bash)
	if format == "shell" {
		fmt.Printf("HAS_FAILED=%q\n", fmt.Sprintf("%t", anyFailed))
		fmt.Printf("PASSED_COUNT=%q\n", fmt.Sprintf("%d", passedCount))
		fmt.Printf("SKIPPED_COUNT=%q\n", fmt.Sprintf("%d", skippedCount))
		depsList := make([]string, 0, len(results))
		for _, r := range results {
			depsList = append(depsList, r.Moniker)
		}
		fmt.Printf("DEPS_LIST=%q\n", strings.Join(depsList, ","))
		if anyFailed {
			return 1
		}
		return 0
	}

	// Summary (text format)
	if anyFailed {
		log.Infof("✗ Dependency CI check failed")
		log.Infof("")
		for _, r := range results {
			switch r.Status {
			case "failed":
				log.Infof("  %s: CI failed", r.Moniker)
				if r.RunURL != "" {
					log.Infof("  Run: %s", r.RunURL)
				}
			case "not_found":
				log.Infof("  %s: No CI found for commit %s", r.Moniker, r.LastCommit[:7])
				log.Infof("  Trigger: gh workflow run %s", r.CIWorkflow)
			}
		}
		return 1
	}

	if skippedCount > 0 {
		log.Infof("✓ All %d dependency CI checks passed (%d skipped)", passedCount, skippedCount)
	} else {
		log.Infof("✓ All %d dependency CI checks passed", passedCount)
	}
	return 0
}

// getTransitiveDeps returns all transitive dependencies of a module (sorted).
func getTransitiveDeps(module string, reg *modules.Registry) []string {
	visited := make(map[string]bool)
	var collect func(m string)
	collect = func(m string) {
		mod, exists := reg.Get(m)
		if !exists {
			return
		}
		for _, dep := range mod.DependsOn {
			if !visited[dep] {
				visited[dep] = true
				collect(dep)
			}
		}
	}

	collect(module)

	deps := make([]string, 0, len(visited))
	for dep := range visited {
		deps = append(deps, dep)
	}
	sort.Strings(deps)
	return deps
}

// shouldSkipModule returns true if this module should be skipped for CI checks.
func shouldSkipModule(mod *modules.ModuleContract, workspaceRoot string) bool {
	// Skip if no CI workflow exists
	workflowPath := filepath.Join(workspaceRoot, ".github", "workflows", fmt.Sprintf("ci-%s.yaml", mod.Moniker))
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		return true
	}
	return false
}

// getSkipReason returns the reason why a module is skipped.
func getSkipReason(mod *modules.ModuleContract, workspaceRoot string) string {
	workflowPath := filepath.Join(workspaceRoot, ".github", "workflows", fmt.Sprintf("ci-%s.yaml", mod.Moniker))
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		return "no CI workflow"
	}
	return "unknown"
}

// findLastChangedCommit finds the most recent commit that changed files owned by this module.
func findLastChangedCommit(mod *modules.ModuleContract, workspaceRoot string) (string, string, error) {
	// Get file patterns for this module
	patterns := mod.GetGlobPatterns()
	if len(patterns) == 0 {
		// Fallback to package roots
		for _, root := range mod.GetComponentRoots() {
			if root != "" && root != "/" {
				patterns = append(patterns, root+"/**")
			}
		}
	}

	// Convert patterns to paths for git log
	// Git log works better with directory paths than globs
	var paths []string
	for _, p := range patterns {
		// Strip /**/* suffixes for git
		p = strings.TrimSuffix(p, "/**/*")
		p = strings.TrimSuffix(p, "/**")
		p = strings.TrimSuffix(p, "/*")
		if p != "" {
			paths = append(paths, p)
		}
	}

	if len(paths) == 0 {
		return "", "", fmt.Errorf("no file paths for module %s", mod.Moniker)
	}

	// Use git log to find last commit touching these paths
	args := []string{"log", "-1", "--format=%H%n%s", "--"}
	args = append(args, paths...)

	ts := tool.GlobalToolSystem()
	if ts == nil {
		return "", "", fmt.Errorf("tool system not initialized")
	}
	output, err := ts.RunTool(context.Background(), "git", workspaceRoot, args...)
	if err != nil {
		return "", "", fmt.Errorf("git log failed: %w", err)
	}

	lines := strings.SplitN(strings.TrimSpace(string(output)), "\n", 2)
	if len(lines) < 1 || lines[0] == "" {
		return "", "", fmt.Errorf("no commits found for module files")
	}

	commitSHA := lines[0]
	commitMsg := ""
	if len(lines) > 1 {
		commitMsg = lines[1]
	}

	return commitSHA, commitMsg, nil
}

// waitForDepCI waits for CI to pass for a dependency at a specific commit.
func waitForDepCI(dep, workflow, commitSHA string, timeout, interval int, targetModule, workspaceRoot string) DepCIStatus {
	startTime := time.Now()
	var lastStatus DepCIStatus

	for {
		elapsed := time.Since(startTime)
		if elapsed.Seconds() > float64(timeout) {
			// Timeout - return last known status or not_found
			if lastStatus.Status == "running" {
				return DepCIStatus{Status: "timeout", RunID: lastStatus.RunID, RunURL: lastStatus.RunURL}
			}
			return DepCIStatus{Status: "not_found"}
		}

		// Check CI status
		status, err := checkDepCIStatus(workflow, commitSHA, workspaceRoot)
		if err != nil {
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		lastStatus = status

		if status.Status == "success" || status.Status == "failed" {
			return status
		}

		if status.Status == "running" {
			elapsedStr := formatElapsed(elapsed)
			log.Infof("    ⏱ %s CI run #%d ◐ in_progress", elapsedStr, status.RunID)
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		// Not found - check if CI might be inherited
		canInherit, inheritedCI, _ := canInheritCIForDep(dep, workflow, commitSHA, workspaceRoot)
		if canInherit {
			return DepCIStatus{
				Status: "success",
				RunID:  inheritedCI.RunID,
			}
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// checkDepCIStatus checks the CI status for a workflow at a specific commit
// It checks:
// 1. Exact commit match - CI ran on this exact commit
// 2. Descendant match - CI ran on a newer commit that includes this commit's changes.
func checkDepCIStatus(workflow, commitSHA, workspaceRoot string) (DepCIStatus, error) {
	// First try exact commit match
	output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", workspaceRoot, "run", "list",
		"--commit", commitSHA,
		"--workflow", workflow,
		"--json", "status,conclusion,databaseId,url",
		"--limit", "5",
	)
	if err != nil {
		return DepCIStatus{}, fmt.Errorf("gh run list failed: %w", err)
	}

	var runs []struct {
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		DatabaseID int64  `json:"databaseId"`
		URL        string `json:"url"`
	}
	if err := json.Unmarshal(output, &runs); err != nil {
		return DepCIStatus{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check exact match results - prioritize success over running over failed
	var runningStatus *DepCIStatus
	var failedStatus *DepCIStatus

	for _, run := range runs {
		status := DepCIStatus{
			RunID:  run.DatabaseID,
			RunURL: run.URL,
		}

		switch run.Status {
		case "completed":
			if run.Conclusion == "success" {
				status.Status = "success"
				return status, nil // Success takes priority - return immediately
			}
			// Remember failed, but keep looking for success
			if failedStatus == nil {
				status.Status = "failed"
				failedStatus = &status
			}
		case "in_progress", "queued":
			// Remember running, but keep looking for success
			if runningStatus == nil {
				status.Status = "running"
				runningStatus = &status
			}
		}
	}

	// Return in priority order: running > failed (since running might succeed)
	if runningStatus != nil {
		return *runningStatus, nil
	}
	if failedStatus != nil {
		return *failedStatus, nil
	}

	// No exact match - check if any recent successful run includes this commit
	// (i.e., the CI ran on a descendant of commitSHA)
	recentRuns, err := getRecentSuccessfulRuns(workflow, workspaceRoot)
	if err != nil {
		return DepCIStatus{Status: "not_found"}, nil
	}

	for _, run := range recentRuns {
		// Check if run.HeadSHA is commitSHA or a descendant of commitSHA
		if run.HeadSHA == commitSHA || isAncestor(commitSHA, run.HeadSHA) {
			return DepCIStatus{
				Status: "success",
				RunID:  run.DatabaseID,
				RunURL: run.URL,
			}, nil
		}
	}

	return DepCIStatus{Status: "not_found"}, nil
}

// getRecentSuccessfulRuns gets recent successful CI runs for a workflow.
func getRecentSuccessfulRuns(workflow, workspaceRoot string) ([]struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	DatabaseID int64  `json:"databaseId"`
	URL        string `json:"url"`
	HeadSHA    string `json:"headSha"`
}, error,
) {
	output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", workspaceRoot, "run", "list",
		"--workflow", workflow,
		"--branch", "main",
		"--status", "success",
		"--json", "status,conclusion,databaseId,url,headSha",
		"--limit", "10",
	)
	if err != nil {
		return nil, fmt.Errorf("gh run list failed: %w", err)
	}

	var runs []struct {
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		DatabaseID int64  `json:"databaseId"`
		URL        string `json:"url"`
		HeadSHA    string `json:"headSha"`
	}
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return runs, nil
}

// canInheritCIForDep checks if we can inherit CI from a previous successful run
// Similar to canInheritCIFromPrevious but for dependencies.
func canInheritCIForDep(dep, workflow, releaseCommit, workspaceRoot string) (bool, CIRunInfo, string) {
	// Get last successful CI for this dep
	lastCI, err := getLastSuccessfulModuleCIInfo(workflow, "main", workspaceRoot)
	if err != nil || lastCI.SHA == "" {
		return false, CIRunInfo{}, ""
	}

	// Check if releaseCommit is descendant of lastCI.SHA
	// If so, CI covers our commit
	if isAncestor(lastCI.SHA, releaseCommit) {
		return true, lastCI, fmt.Sprintf("CI inherited from %s", lastCI.SHA[:7])
	}

	return false, CIRunInfo{}, ""
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func printAwaitDepsUsage() {
	log.Info("Wait for dependency CI to pass before release")
	log.Info("")
	log.Info("Usage: release await-deps <module> [options]")
	log.Info("")
	log.Info("Options:")
	log.Info("  --timeout N       Maximum wait time per dependency in seconds (default: 300)")
	log.Info("  --interval N      Poll interval in seconds (default: 15)")
	log.Info("  --skip-static     Skip modules without CI workflows (default)")
	log.Info("  --no-skip-static  Require CI for all dependencies")
	log.Info("  --format FORMAT   Output format: text (default) or shell (for eval)")
	log.Info("")
	log.Info("Shell format output (for bash eval):")
	log.Info("  HAS_FAILED=\"true|false\"")
	log.Info("  PASSED_COUNT=\"N\"")
	log.Info("  SKIPPED_COUNT=\"N\"")
	log.Info("  DEPS_LIST=\"dep1,dep2,...\"")
	log.Info("")
	log.Info("Examples:")
	log.Info("  release await-deps eac-ext")
	log.Info("  release await-deps eac-ext --timeout 600")
	log.Info("  eval $(release await-deps eac-ext --format shell)")
}
