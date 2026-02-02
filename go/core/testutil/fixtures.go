// Package testutil provides test fixtures and helpers for unit testing.
package testutil

import (
	"time"

	"github.com/ready-to-release/eac/go/core/github"
)

// --- GitHub Mock Helpers ---

// NewMockGitHub creates a pre-configured MockAPI for testing.
func NewMockGitHub() *github.MockAPI {
	return github.NewMockAPI()
}

// WithSuccessfulRun adds a successful workflow run to the mock.
func WithSuccessfulRun(mock *github.MockAPI, workflow, sha string) *github.MockAPI {
	return mock.AddRun(workflow, github.WorkflowRun{
		ID:         len(mock.WorkflowRuns[workflow]) + 1,
		HeadSHA:    sha,
		Status:     "completed",
		Conclusion: "success",
		CreatedAt:  time.Now(),
	})
}

// WithFailedRun adds a failed workflow run to the mock.
func WithFailedRun(mock *github.MockAPI, workflow, sha string) *github.MockAPI {
	return mock.AddRun(workflow, github.WorkflowRun{
		ID:         len(mock.WorkflowRuns[workflow]) + 1,
		HeadSHA:    sha,
		Status:     "completed",
		Conclusion: "failure",
		CreatedAt:  time.Now(),
	})
}

// WithInProgressRun adds an in-progress workflow run to the mock.
func WithInProgressRun(mock *github.MockAPI, workflow, sha string) *github.MockAPI {
	return mock.AddRun(workflow, github.WorkflowRun{
		ID:         len(mock.WorkflowRuns[workflow]) + 1,
		HeadSHA:    sha,
		Status:     "in_progress",
		Conclusion: "",
		CreatedAt:  time.Now(),
	})
}

// WithOldSuccessfulRun adds a successful run from the past.
func WithOldSuccessfulRun(mock *github.MockAPI, workflow, sha string, age time.Duration) *github.MockAPI {
	return mock.AddRun(workflow, github.WorkflowRun{
		ID:         len(mock.WorkflowRuns[workflow]) + 1,
		HeadSHA:    sha,
		Status:     "completed",
		Conclusion: "success",
		CreatedAt:  time.Now().Add(-age),
	})
}

// WithRelease adds a release to the mock.
func WithRelease(mock *github.MockAPI, tag, name string) *github.MockAPI {
	return mock.AddRelease(tag, name)
}

// WithTreeFiles adds tree files for a SHA.
func WithTreeFiles(mock *github.MockAPI, sha string, files []string) *github.MockAPI {
	return mock.AddTreeFiles(sha, files)
}

// --- Assertion Helpers ---

// AssertCalled panics if the method was not called on the mock.
func AssertCalled(mock *github.MockAPI, method string) {
	if mock.CallCount(method) == 0 {
		panic("expected " + method + " to be called")
	}
}

// AssertNotCalled panics if the method was called on the mock.
func AssertNotCalled(mock *github.MockAPI, method string) {
	if mock.CallCount(method) > 0 {
		panic("expected " + method + " to not be called")
	}
}

// AssertCallCount panics if the method was not called exactly n times.
func AssertCallCount(mock *github.MockAPI, method string, expected int) {
	actual := mock.CallCount(method)
	if actual != expected {
		panic("expected " + method + " to be called " + string(rune(expected+'0')) + " times, got " + string(rune(actual+'0')))
	}
}
