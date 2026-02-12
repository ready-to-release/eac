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
type runStatusInfo struct {
	HeadSHA    string `json:"headSha"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
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
