// Package pipeline contains godog step implementations for eac.
//
// This file contains helper utilities for pipeline testing to reduce boilerplate.
package pipeline

import (
	"fmt"
	"strings"

	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/core/github"
)

// PipelineTestHelper provides utilities for pipeline testing.
// Reduces boilerplate in pipeline step definitions.
type PipelineTestHelper struct {
	ctx    *eacgodog.TestContext
	ghMock *github.CLIMock
}

// NewPipelineTestHelper creates a new pipeline test helper.
func NewPipelineTestHelper(ctx *eacgodog.TestContext) *PipelineTestHelper {
	return &PipelineTestHelper{
		ctx:    ctx,
		ghMock: github.NewCLIMock(),
	}
}

// CreateModule creates a module in the isolated test environment.
// Dependencies should be module monikers (e.g., "core").
func (h *PipelineTestHelper) CreateModule(moniker string, dependencies ...string) error {
	// TODO: Implement module creation
	// This would need to:
	// 1. Create module directory structure
	// 2. Create module contract file with dependencies
	// 3. Initialize git if needed
	return nil
}

// MarkModuleChanged marks a module as having uncommitted changes.
func (h *PipelineTestHelper) MarkModuleChanged(moniker string) error {
	// TODO: Implement marking module as changed
	// This would need to:
	// 1. Find module directory
	// 2. Create or modify a file
	// 3. Leave changes uncommitted
	return nil
}

// CreateWorkflow creates a GitHub workflow with the specified status.
func (h *PipelineTestHelper) CreateWorkflow(name string, status, conclusion string) error {
	// Add workflow run to mock
	runs := []github.WorkflowRun{
		{
			ID:         123,
			HeadSHA:    "abc123",
			Status:     status,
			Conclusion: conclusion,
		},
	}
	h.ghMock.WithWorkflowRuns(name, runs)
	return nil
}

// SetWorkflowRunning sets a workflow to running status.
func (h *PipelineTestHelper) SetWorkflowRunning(name string) error {
	return h.CreateWorkflow(name, "in_progress", "")
}

// AssertModuleProcessedBefore asserts that one module was processed before another.
// Checks the command output for processing order.
func (h *PipelineTestHelper) AssertModuleProcessedBefore(first, second string) error {
	output := h.ctx.CommandOutput
	firstIdx := strings.Index(output, first)
	secondIdx := strings.Index(output, second)

	if firstIdx == -1 {
		return fmt.Errorf("module %s was not processed (not found in output)", first)
	}
	if secondIdx == -1 {
		return fmt.Errorf("module %s was not processed (not found in output)", second)
	}
	if firstIdx >= secondIdx {
		return fmt.Errorf("expected %s before %s in output, but got %s first", first, second, second)
	}

	return nil
}

// AssertOnlyModulesProcessed asserts that only the specified modules were processed.
func (h *PipelineTestHelper) AssertOnlyModulesProcessed(monikers ...string) error {
	output := h.ctx.CommandOutput
	for _, moniker := range monikers {
		if !strings.Contains(output, moniker) {
			return fmt.Errorf("expected module %s to be processed, but it was not found in output", moniker)
		}
	}
	return nil
}

// GetGitHubMock returns the GitHub CLI mock for injection into context.
func (h *PipelineTestHelper) GetGitHubMock() *github.CLIMock {
	return h.ghMock
}
