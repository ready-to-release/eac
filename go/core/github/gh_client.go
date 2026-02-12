package github

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Compile-time interface checks
var (
	_ API         = (*GHClient)(nil)
	_ CLIExecutor = (*GHClient)(nil)
)

// ErrRunNotFound is returned when a workflow run is not found for the given SHA.
var ErrRunNotFound = errors.New("workflow run not found")

// GHClient implements API using an injected CLIExecutor.
// The executor is provided by the outer layer (CLI bootstrap) and routes
// through the tool registry for binary resolution.
type GHClient struct {
	executor CLIExecutor
	workDir  string
}

// NewGHClient creates a new GitHub client.
// executor: a CLIExecutor that handles actual gh binary invocation.
// workDir: working directory for git context (empty = current).
func NewGHClient(executor CLIExecutor, workDir string) *GHClient {
	return &GHClient{executor: executor, workDir: workDir}
}

// GetTreeFiles returns all file paths in the repository at the given SHA.
// Uses the GitHub Trees API for fast file listing.
func (c *GHClient) GetTreeFiles(sha string) ([]string, error) {
	output, err := c.executor.Exec("api",
		fmt.Sprintf("repos/{owner}/{repo}/git/trees/%s", sha),
		"-q", ".tree[] | select(.type==\"blob\") | .path",
		"--paginate",
	)
	if err != nil {
		return nil, fmt.Errorf("gh api failed: %w", err)
	}

	return filterNonEmptyLines(string(output)), nil
}

// ListRuns returns workflow runs matching the given options.
func (c *GHClient) ListRuns(workflow string, opts ListRunsOpts) ([]WorkflowRun, error) {
	args := []string{"run", "list", "-w", workflow, "--json", "databaseId,headSha,status,conclusion,createdAt,name"}

	if opts.Status != "" {
		args = append(args, "-s", opts.Status)
	}
	if opts.Branch != "" {
		args = append(args, "-b", opts.Branch)
	}
	if opts.Limit > 0 {
		args = append(args, "-L", fmt.Sprintf("%d", opts.Limit))
	}

	output, err := c.executor.Exec(args...)
	if err != nil {
		return nil, fmt.Errorf("gh run list failed: %w", err)
	}

	return parseWorkflowRuns(output)
}

// FindRunBySHA finds a workflow run with the given HEAD SHA.
// Returns nil if no matching run is found.
func (c *GHClient) FindRunBySHA(workflow, sha string, limit int) (*WorkflowRun, error) {
	if limit <= 0 {
		limit = 10
	}

	runs, err := c.ListRuns(workflow, ListRunsOpts{Limit: limit})
	if err != nil {
		return nil, err
	}

	for _, run := range runs {
		if run.HeadSHA == sha {
			return &run, nil
		}
	}

	return nil, ErrRunNotFound
}

// HasRecentSuccess checks if a successful run exists for the given SHA
// within the specified time window.
func (c *GHClient) HasRecentSuccess(workflow, sha string, since time.Duration) (bool, error) {
	runs, err := c.ListRuns(workflow, ListRunsOpts{
		Status: "success",
		Limit:  20,
	})
	if err != nil {
		return false, err
	}

	cutoff := time.Now().Add(-since)
	for _, run := range runs {
		if run.HeadSHA == sha && run.CreatedAt.After(cutoff) {
			return true, nil
		}
	}

	return false, nil
}

// ListReleases returns the most recent releases.
func (c *GHClient) ListReleases(limit int) ([]Release, error) {
	if limit <= 0 {
		limit = 100
	}

	output, err := c.executor.Exec("release", "list",
		"--limit", fmt.Sprintf("%d", limit),
		"--json", "tagName,name,isDraft",
	)
	if err != nil {
		return nil, fmt.Errorf("gh release list failed: %w", err)
	}

	return parseReleases(output)
}

// ReleaseExists checks if a release with the given tag exists.
func (c *GHClient) ReleaseExists(tag string) (bool, error) {
	_, err := c.executor.Exec("release", "view", tag, "--json", "tagName")
	if err != nil {
		// gh returns error if release doesn't exist
		return false, nil
	}
	return true, nil
}

// Exec delegates to the injected executor.
func (c *GHClient) Exec(args ...string) ([]byte, error) {
	return c.executor.Exec(args...)
}

// ExecContext delegates to the injected executor.
func (c *GHClient) ExecContext(ctx context.Context, args ...string) ([]byte, error) {
	return c.executor.ExecContext(ctx, args...)
}
