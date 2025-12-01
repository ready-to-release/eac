// Package pipelinerunner provides functionality to execute GitHub workflows
// respecting module dependencies
package pipelinerunner

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GitHubCLI defines operations for interacting with GitHub workflows
type GitHubCLI interface {
	TriggerWorkflow(workflowFile string, ref string) (runID string, err error)
	WatchRun(runID string) error
}

// GitHubCLIImpl implements GitHubCLI using the gh CLI tool
type GitHubCLIImpl struct {
	repoPath string
}

// NewGitHubCLI creates a new GitHub CLI wrapper
func NewGitHubCLI(repoPath string) GitHubCLI {
	return &GitHubCLIImpl{
		repoPath: repoPath,
	}
}

// TriggerWorkflow triggers a GitHub workflow and returns the run ID
func (g *GitHubCLIImpl) TriggerWorkflow(workflowFile string, ref string) (string, error) {
	// Trigger the workflow
	cmd := exec.Command("gh", "workflow", "run", workflowFile, "--ref", ref)
	cmd.Dir = g.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to trigger workflow %s: %w\nOutput: %s", workflowFile, err, string(output))
	}

	// Wait a bit for the workflow to be created
	time.Sleep(2 * time.Second)

	// Get the most recent run ID for this workflow
	cmd = exec.Command("gh", "run", "list", "--workflow="+workflowFile, "--limit", "1", "--json", "databaseId", "--jq", ".[0].databaseId")
	cmd.Dir = g.repoPath

	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get run ID for %s: %w", workflowFile, err)
	}

	runID := strings.TrimSpace(string(output))
	if runID == "" || runID == "null" {
		return "", fmt.Errorf("no run ID found for workflow %s", workflowFile)
	}

	return runID, nil
}

// WatchRun watches a workflow run until completion and returns error if it fails
func (g *GitHubCLIImpl) WatchRun(runID string) error {
	cmd := exec.Command("gh", "run", "watch", runID, "--exit-status")
	cmd.Dir = g.repoPath

	// Run and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("workflow run %s failed: %w\nOutput: %s", runID, err, string(output))
	}

	return nil
}
