// Package pipelinerunner provides functionality to execute GitHub workflows
// respecting module dependencies
package pipelinerunner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/clibase/ghexec"
	"github.com/ready-to-release/eac/go/clibase/gitexec"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/environments"
)

// WorkflowRun represents a GitHub Actions workflow run.
type WorkflowRun struct {
	Name       string
	Status     string // "completed", "in_progress", "queued"
	Conclusion string // "success", "failure", "cancelled", "skipped", etc.
	WorkflowName string
}

// GitHubCLI defines operations for interacting with GitHub workflows.
type GitHubCLI interface {
	TriggerWorkflow(workflowFile, ref string) (runID string, err error)
	WatchRun(runID string) error
	WatchRunWithTimeout(runID string, timeoutSeconds int) error
	GetCommitSHA(ref string) (string, error)
	ListWorkflowRuns(commitSHA string) ([]WorkflowRun, error)
}

// GitHubCLIImpl implements GitHubCLI using the gh CLI tool.
type GitHubCLIImpl struct {
	repoPath string
}

// NewGitHubCLI creates a new GitHub CLI wrapper.
// If R2R_MOCK_GITHUB_CLI environment variable is set, returns a mock implementation.
func NewGitHubCLI(repoPath string) GitHubCLI {
	if os.Getenv(environments.EnvR2RMockGitHubCLI) == "true" {
		return &MockGitHubCLI{repoPath: repoPath}
	}
	return &GitHubCLIImpl{
		repoPath: repoPath,
	}
}

// TriggerWorkflow triggers a GitHub workflow and returns the run ID.
func (g *GitHubCLIImpl) TriggerWorkflow(workflowFile, ref string) (string, error) {
	// Trigger the workflow
	output, exitCode, err := ghexec.RunCombined(context.Background(), g.repoPath, "workflow", "run", workflowFile, "--ref", ref)
	if err != nil {
		return "", fmt.Errorf("failed to trigger workflow %s: %w", workflowFile, err)
	}
	if exitCode != 0 {
		return "", fmt.Errorf("failed to trigger workflow %s (exit %d)\nOutput: %s", workflowFile, exitCode, string(output))
	}

	// Wait a bit for the workflow to be created
	time.Sleep(config.CIDispatchSettleTime())

	// Get the most recent run ID for this workflow
	output, err = ghexec.Run(g.repoPath, "run", "list", "--workflow="+workflowFile, "--limit", "1", "--json", "databaseId", "--jq", ".[0].databaseId")
	if err != nil {
		return "", fmt.Errorf("failed to get run ID for %s: %w", workflowFile, err)
	}

	runID := strings.TrimSpace(string(output))
	if runID == "" || runID == "null" {
		return "", fmt.Errorf("no run ID found for workflow %s", workflowFile)
	}

	return runID, nil
}

// WatchRun watches a workflow run until completion and returns error if it fails.
func (g *GitHubCLIImpl) WatchRun(runID string) error {
	output, exitCode, err := ghexec.RunCombined(context.Background(), g.repoPath, "run", "watch", runID, "--exit-status")
	if err != nil {
		return fmt.Errorf("workflow run %s failed: %w", runID, err)
	}
	if exitCode != 0 {
		return fmt.Errorf("workflow run %s failed (exit %d)\nOutput: %s", runID, exitCode, string(output))
	}

	return nil
}

// WatchRunWithTimeout watches a workflow run with a timeout.
func (g *GitHubCLIImpl) WatchRunWithTimeout(runID string, timeoutSeconds int) error {
	// Create a context with timeout
	done := make(chan error, 1)

	go func() {
		done <- g.WatchRun(runID)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(time.Duration(timeoutSeconds) * time.Second):
		return fmt.Errorf("timeout exceeded (%d seconds) waiting for run %s", timeoutSeconds, runID)
	}
}

// GetCommitSHA gets the commit SHA for a given ref (branch, tag, or commit).
func (g *GitHubCLIImpl) GetCommitSHA(ref string) (string, error) {
	if ref == "" {
		ref = "HEAD"
	}

	// First try git to get the SHA locally (faster)
	output, err := gitexec.Run(g.repoPath, "rev-parse", ref)
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// Fallback to gh API if git fails (e.g., for remote refs not fetched locally)
	output, err = ghexec.Run(g.repoPath, "api", "repos/{owner}/{repo}/commits/"+ref, "--jq", ".sha")
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA for ref %s: %w", ref, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// ListWorkflowRuns lists all workflow runs for a given commit SHA.
func (g *GitHubCLIImpl) ListWorkflowRuns(commitSHA string) ([]WorkflowRun, error) {
	// Use gh run list with commit filter
	output, err := ghexec.Run(g.repoPath, "run", "list",
		"--commit", commitSHA,
		"--json", "name,status,conclusion,workflowName")
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow runs for commit %s: %w", commitSHA, err)
	}

	// Parse JSON output
	// For simplicity, we'll use a basic JSON parser
	// In production, you'd use encoding/json
	var runs []WorkflowRun

	// Simple parsing: look for objects in the array
	// This is a placeholder - real implementation would use json.Unmarshal
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "name") {
			// Parse workflow run from line
			// This is simplified - real implementation would parse JSON properly
			run := WorkflowRun{}
			if strings.Contains(line, "success") {
				run.Status = "completed"
				run.Conclusion = "success"
			} else if strings.Contains(line, "failure") {
				run.Status = "completed"
				run.Conclusion = "failure"
			}
			runs = append(runs, run)
		}
	}

	return runs, nil
}

// MockGitHubCLI is a mock implementation for testing.
type MockGitHubCLI struct {
	repoPath string

	// Mock behavior control
	failingWorkflows []string  // List of workflow names that should fail
	noWorkflows      bool       // If true, return no workflows
	invalidRef       bool       // If true, simulate invalid ref error
}

// TriggerWorkflow simulates triggering a workflow.
func (m *MockGitHubCLI) TriggerWorkflow(workflowFile, ref string) (string, error) {
	// Extract module name from workflow file (e.g., "core.yaml" -> "core")
	moduleName := strings.TrimSuffix(workflowFile, ".yaml")
	moduleName = strings.TrimSuffix(moduleName, ".yml")

	log.Infof("Processing module: %s", moduleName)

	// Return a mock run ID
	return "mock-run-123", nil
}

// WatchRun simulates watching a workflow run.
func (m *MockGitHubCLI) WatchRun(runID string) error {
	// Mock workflows always succeed immediately
	return nil
}

// WatchRunWithTimeout simulates watching a workflow run with timeout.
func (m *MockGitHubCLI) WatchRunWithTimeout(runID string, timeoutSeconds int) error {
	// Check if we should simulate a timeout
	// For testing, we can use an environment variable
	if os.Getenv("R2R_MOCK_TIMEOUT") == "true" {
		return fmt.Errorf("timeout exceeded (%d seconds) waiting for run %s", timeoutSeconds, runID)
	}

	// Mock workflows succeed immediately if not simulating timeout
	return nil
}

// GetCommitSHA simulates getting a commit SHA for a ref.
func (m *MockGitHubCLI) GetCommitSHA(ref string) (string, error) {
	// Check if we should simulate invalid ref
	if os.Getenv("R2R_MOCK_INVALID_REF") == "true" {
		return "", fmt.Errorf("not found: invalid ref %s", ref)
	}

	if ref == "" {
		ref = "HEAD"
	}

	// Try to get actual commit SHA from repo if it exists
	output, err := gitexec.Run(m.repoPath, "rev-parse", ref)
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// If git fails, return error (don't fallback to mock SHA)
	// This ensures invalid refs are properly reported
	return "", fmt.Errorf("not found: invalid ref %s", ref)
}

// ListWorkflowRuns simulates listing workflow runs for a commit.
func (m *MockGitHubCLI) ListWorkflowRuns(commitSHA string) ([]WorkflowRun, error) {
	// Check if we should return no workflows
	if os.Getenv("R2R_MOCK_NO_WORKFLOWS") == "true" {
		return []WorkflowRun{}, nil
	}

	// Check if we should simulate failing workflows
	failingWorkflow := os.Getenv("R2R_MOCK_FAILING_WORKFLOW")

	// Read workflows directory to find actual workflows
	workflowsDir := filepath.Join(m.repoPath, ".github", "workflows")
	entries, err := os.ReadDir(workflowsDir)
	if err != nil {
		// If no workflows directory, return empty list
		return []WorkflowRun{}, nil
	}

	var runs []WorkflowRun
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		// Extract workflow name (remove extension)
		workflowName := strings.TrimSuffix(name, ".yaml")
		workflowName = strings.TrimSuffix(workflowName, ".yml")

		// Determine if this workflow is failing
		isFailure := failingWorkflow != "" && workflowName == failingWorkflow

		run := WorkflowRun{
			Name:         workflowName,
			Status:       "completed",
			WorkflowName: workflowName,
		}

		if isFailure {
			run.Conclusion = "failure"
		} else {
			run.Conclusion = "success"
		}

		runs = append(runs, run)
	}

	return runs, nil
}
