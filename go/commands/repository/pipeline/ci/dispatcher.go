package ci

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/tool"
)

// ModuleRunStatus holds the status of a module from a batch query.
type ModuleRunStatus struct {
	Status     string
	Conclusion string
}

// CIWorkflowDispatcher dispatches and monitors CI workflows.
// Port interface for testability — mock in tests, real GitHub in production.
type CIWorkflowDispatcher interface {
	// Dispatch triggers a CI workflow for a module and returns immediately.
	// The workflow name is derived as "ci-{module}.yaml".
	Dispatch(ctx context.Context, module, ref, sha, triggerRunID string) error

	// GetStatus returns the current status of a module's CI workflow for a given SHA.
	// Returns (status, conclusion, error) where:
	//   status: "queued", "in_progress", "completed", "none"
	//   conclusion: "success", "failure", "cancelled" (only meaningful when status == "completed")
	GetStatus(ctx context.Context, module, sha string) (status, conclusion string, err error)

	// BatchGetStatus returns the status of multiple modules' CI workflows in a single API call.
	// Uses `gh run list --commit <sha>` to fetch all runs for the commit, then filters
	// by workflow name prefix "CI: " to extract module monikers.
	BatchGetStatus(ctx context.Context, modules []string, sha string) (map[string]ModuleRunStatus, error)
}

// ghWorkflowDispatcher implements CIWorkflowDispatcher using the GitHub CLI.
type ghWorkflowDispatcher struct {
	workspaceRoot string
}

// NewGHWorkflowDispatcher creates a CIWorkflowDispatcher backed by the GitHub CLI.
func NewGHWorkflowDispatcher(workspaceRoot string) CIWorkflowDispatcher {
	return &ghWorkflowDispatcher{workspaceRoot: workspaceRoot}
}

// Dispatch triggers a CI workflow for the given module.
func (d *ghWorkflowDispatcher) Dispatch(ctx context.Context, module, ref, sha, triggerRunID string) error {
	workflow := fmt.Sprintf("ci-%s.yaml", module)

	args := []string{
		"workflow", "run", workflow,
		"--ref", ref,
		"-f", fmt.Sprintf("ref=refs/heads/%s", ref),
		"-f", fmt.Sprintf("sha=%s", sha),
	}
	if triggerRunID != "" {
		args = append(args, "-f", fmt.Sprintf("trigger_run_id=%s", triggerRunID))
	}

	_, err := tool.GlobalToolSystem().RunTool(ctx, "gh", d.workspaceRoot, args...)
	if err != nil {
		return fmt.Errorf("dispatch %s: %w", workflow, err)
	}

	// Brief pause to avoid GitHub API rate limits when dispatching multiple workflows.
	time.Sleep(config.CIDispatchSettleTime())
	return nil
}

// runStatusInfo holds workflow run status from GitHub JSON output.
// WorkflowName is populated by BatchGetStatus (omitted by GetStatus queries).
type runStatusInfo struct {
	HeadSHA      string `json:"headSha"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	WorkflowName string `json:"workflowName,omitempty"`
}

// GetStatus queries the CI workflow status for a module at a specific SHA.
// It follows the same logic as getRunStatusForSHA in await_ci.go:
// counts active runs, and for completed runs only considers the most recent.
func (d *ghWorkflowDispatcher) GetStatus(ctx context.Context, module, sha string) (string, string, error) {
	workflow := fmt.Sprintf("ci-%s.yaml", module)

	output, err := tool.GlobalToolSystem().RunTool(ctx, "gh", d.workspaceRoot,
		"run", "list", "-w", workflow, "--limit", "20",
		"--json", "headSha,status,conclusion")
	if err != nil {
		return "none", "", nil
	}

	var runs []runStatusInfo
	if err := json.Unmarshal(output, &runs); err != nil {
		return "none", "", nil
	}

	// Find runs matching our SHA.
	// Count all active runs. For completed runs, only the most recent matters.
	activeCount := 0
	foundCompleted := false
	var mostRecentConclusion string

	for _, run := range runs {
		if run.HeadSHA != sha {
			continue
		}

		status := strings.ToLower(run.Status)
		switch status {
		case "in_progress", "queued", "waiting", "pending":
			activeCount++
		case "completed":
			if !foundCompleted {
				foundCompleted = true
				mostRecentConclusion = run.Conclusion
			}
		}
	}

	if activeCount > 0 {
		return "in_progress", "", nil
	}
	if foundCompleted {
		return "completed", mostRecentConclusion, nil
	}
	return "none", "", nil
}

// BatchGetStatus queries CI workflow status for multiple modules in a single API call.
// Instead of one `gh run list -w <workflow>` per module, it fetches all runs for the
// commit SHA and filters by workflow name prefix "CI: ".
func (d *ghWorkflowDispatcher) BatchGetStatus(ctx context.Context, modules []string, sha string) (map[string]ModuleRunStatus, error) {
	output, err := tool.GlobalToolSystem().RunTool(ctx, "gh", d.workspaceRoot,
		"run", "list", "--commit", sha, "--limit", "100",
		"--json", "headSha,status,conclusion,workflowName")
	if err != nil {
		return nil, fmt.Errorf("batch status query: %w", err)
	}

	var runs []runStatusInfo
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("parse batch status: %w", err)
	}

	// Build module set for filtering.
	moduleSet := make(map[string]bool, len(modules))
	for _, m := range modules {
		moduleSet[m] = true
	}

	// Process runs — same active-trumps-completed logic as GetStatus, but across all modules.
	type moduleRunState struct {
		activeCount      int
		foundCompleted   bool
		latestConclusion string
	}
	states := make(map[string]*moduleRunState)

	for _, run := range runs {
		module := extractModuleFromWorkflowName(run.WorkflowName)
		if module == "" || !moduleSet[module] {
			continue
		}

		state, ok := states[module]
		if !ok {
			state = &moduleRunState{}
			states[module] = state
		}

		status := strings.ToLower(run.Status)
		switch status {
		case "in_progress", "queued", "waiting", "pending":
			state.activeCount++
		case "completed":
			if !state.foundCompleted {
				state.foundCompleted = true
				state.latestConclusion = run.Conclusion
			}
		}
	}

	// Build result map.
	result := make(map[string]ModuleRunStatus, len(states))
	for module, state := range states {
		if state.activeCount > 0 {
			result[module] = ModuleRunStatus{Status: "in_progress"}
		} else if state.foundCompleted {
			result[module] = ModuleRunStatus{Status: "completed", Conclusion: state.latestConclusion}
		}
	}

	return result, nil
}

// extractModuleFromWorkflowName extracts the module moniker from a GitHub Actions
// workflow name. CI workflows are named "CI: <module>".
func extractModuleFromWorkflowName(name string) string {
	after, found := strings.CutPrefix(name, "CI: ")
	if !found {
		return ""
	}
	return after
}
