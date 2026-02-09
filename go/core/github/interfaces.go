// Package github provides abstractions for GitHub API interactions.
// All external GitHub calls go through these interfaces, enabling
// unit testing with mocks.
package github

import (
	"context"
	"time"
)

// WorkflowRun represents a GitHub Actions workflow run.
type WorkflowRun struct {
	ID         int       `json:"databaseId"`
	HeadSHA    string    `json:"headSha"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	CreatedAt  time.Time `json:"createdAt"`
	Name       string    `json:"name"`
}

// Release represents a GitHub release.
type Release struct {
	TagName string `json:"tagName"`
	Name    string `json:"name"`
	Draft   bool   `json:"isDraft"`
	Body    string `json:"body"` // Release notes body (for bundle module extraction)
}

// ListRunsOpts contains options for listing workflow runs.
type ListRunsOpts struct {
	Status string // Filter by status: success, failure, in_progress, etc.
	Branch string // Filter by branch
	Limit  int    // Maximum number of runs to return
}

// API defines the interface for GitHub operations.
// Implementations include GHClient (real) and MockAPI (testing).
type API interface {
	// Tree operations
	GetTreeFiles(sha string) ([]string, error)

	// Workflow run operations
	ListRuns(workflow string, opts ListRunsOpts) ([]WorkflowRun, error)
	FindRunBySHA(workflow, sha string, limit int) (*WorkflowRun, error)
	HasRecentSuccess(workflow, sha string, since time.Duration) (bool, error)

	// Release operations
	ListReleases(limit int) ([]Release, error)
	ReleaseExists(tag string) (bool, error)
}


// CLIExecutor defines the interface for executing GitHub CLI commands.
// Implementations include GHClient (real) and CLIMock (testing).
type CLIExecutor interface {
	// Exec executes a GitHub CLI command and returns output.
	Exec(args ...string) ([]byte, error)

	// ExecContext executes a GitHub CLI command with context support.
	ExecContext(ctx context.Context, args ...string) ([]byte, error)
}
