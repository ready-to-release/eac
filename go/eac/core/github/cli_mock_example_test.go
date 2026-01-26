package github_test

import (
	"fmt"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/github"
)

// Example_cliMock demonstrates basic usage of the GitHub CLI mock.
func Example_cliMock() {
	// Create a new mock
	mock := github.NewCLIMock()

	// Configure workflow runs
	mock.WithWorkflowRuns("ci.yml", []github.WorkflowRun{
		{
			ID:         123456789,
			HeadSHA:    "abc123",
			Status:     "completed",
			Conclusion: "success",
			CreatedAt:  time.Now(),
			Name:       "CI",
		},
	})

	// Execute command
	output, err := mock.Exec("run", "list", "-w", "ci.yml", "--json", "databaseId,headSha,status,conclusion,createdAt,name")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Output available: %v\n", len(output) > 0)
	// Output: Output available: true
}

// Example_cliMock_fluentAPI demonstrates the fluent builder pattern.
func Example_cliMock_fluentAPI() {
	// Build mock with fluent API
	mock := github.NewCLIMock().
		WithWorkflowRuns("ci.yml", []github.WorkflowRun{
			{ID: 123, HeadSHA: "abc", Status: "completed", Conclusion: "success"},
		}).
		WithRelease("v1.0.0", "Release v1.0.0").
		WithWorkflow("ci.yml", "CI Workflow")

	// All data is configured
	fmt.Printf("Mock configured: %v\n", mock != nil)
	// Output: Mock configured: true
}

// Example_cliMock_errorInjection demonstrates error testing.
func Example_cliMock_errorInjection() {
	// Configure error for specific command
	mock := github.NewCLIMock().
		WithError("run list -w missing.yml", fmt.Errorf("workflow not found"))

	// Execute command that triggers error
	_, err := mock.Exec("run", "list", "-w", "missing.yml", "--json", "databaseId,headSha,status,conclusion,createdAt,name")

	fmt.Printf("Error occurred: %v\n", err != nil)
	// Output: Error occurred: true
}

// Example_cliMock_callTracking demonstrates call tracking for assertions.
func Example_cliMock_callTracking() {
	mock := github.NewCLIMock().
		WithWorkflowRuns("ci.yml", []github.WorkflowRun{})

	// Execute multiple commands
	mock.Exec("run", "list", "-w", "ci.yml", "--json", "databaseId,headSha,status,conclusion,createdAt,name")
	mock.Exec("release", "list", "--limit", "100", "--json", "tagName,name,isDraft")

	// Check calls
	calls := mock.GetCalls()
	fmt.Printf("Total calls: %d\n", len(calls))
	fmt.Printf("First command: %s\n", calls[0].Command)
	// Output: Total calls: 2
	// First command: run list
}

// Example_cliMock_fixtures demonstrates loading data from fixture files.
func Example_cliMock_fixtures() {
	mock := github.NewCLIMock()

	// Load fixtures from JSON files (when running from correct directory)
	// err := mock.LoadWorkflowRunsFixture("ci.yml", "workflow_runs.json")
	// err = mock.LoadReleasesFixture("releases.json")

	fmt.Printf("Mock created: %v\n", mock != nil)
	// Output: Mock created: true
}

// Example_cliMock_realWorldScenario demonstrates a realistic test scenario.
func Example_cliMock_realWorldScenario() {
	// Setup: CI pipeline with multiple states
	mock := github.NewCLIMock().
		WithWorkflowRuns("ci.yml", []github.WorkflowRun{
			{ID: 1, HeadSHA: "sha1", Status: "completed", Conclusion: "success"},
			{ID: 2, HeadSHA: "sha2", Status: "completed", Conclusion: "failure"},
			{ID: 3, HeadSHA: "sha3", Status: "in_progress", Conclusion: ""},
		}).
		WithRelease("v1.0.0", "Release v1.0.0").
		WithRelease("v1.1.0", "Release v1.1.0")

	// Test 1: Check workflow runs
	runsOutput, _ := mock.Exec("run", "list", "-w", "ci.yml", "--json", "databaseId,headSha,status,conclusion,createdAt,name")
	fmt.Printf("Runs output available: %v\n", len(runsOutput) > 0)

	// Test 2: Check specific run
	runOutput, _ := mock.Exec("run", "view", "1", "--json", "databaseId,headSha,status,conclusion,name")
	fmt.Printf("Run view available: %v\n", len(runOutput) > 0)

	// Test 3: Check releases
	releasesOutput, _ := mock.Exec("release", "list", "--limit", "100", "--json", "tagName,name,isDraft")
	fmt.Printf("Releases available: %v\n", len(releasesOutput) > 0)

	// Output: Runs output available: true
	// Run view available: true
	// Releases available: true
}
