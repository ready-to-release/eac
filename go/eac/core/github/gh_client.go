package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ErrRunNotFound is returned when a workflow run is not found for the given SHA.
var ErrRunNotFound = errors.New("workflow run not found")

// GHClient implements API using the gh CLI tool.
type GHClient struct {
	workDir string
}

// NewGHClient creates a new GitHub client.
// workDir is the working directory for git operations (empty = current).
func NewGHClient(workDir string) *GHClient {
	return &GHClient{workDir: workDir}
}

// GetTreeFiles returns all file paths in the repository at the given SHA.
// Uses the GitHub Trees API for fast file listing.
func (c *GHClient) GetTreeFiles(sha string) ([]string, error) {
	// gh api repos/{owner}/{repo}/git/trees/{sha}?recursive=1
	cmd := exec.Command("gh", "api", //nolint:gosec // G204: sha is a git commit hash from controlled source
		fmt.Sprintf("repos/{owner}/{repo}/git/trees/%s", sha),
		"-q", ".tree[] | select(.type==\"blob\") | .path",
		"--paginate",
	)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh api failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	// Filter empty lines
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
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

	cmd := exec.Command("gh", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh run list failed: %w", err)
	}

	var runs []WorkflowRun
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse runs: %w", err)
	}

	return runs, nil
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

	cmd := exec.Command("gh", "release", "list", //nolint:gosec // G204: limit is an int, safe from injection
		"--limit", fmt.Sprintf("%d", limit),
		"--json", "tagName,name,isDraft",
	)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh release list failed: %w", err)
	}

	var releases []Release
	if err := json.Unmarshal(output, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	return releases, nil
}

// ReleaseExists checks if a release with the given tag exists.
func (c *GHClient) ReleaseExists(tag string) (bool, error) {
	cmd := exec.Command("gh", "release", "view", tag, "--json", "tagName")
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	err := cmd.Run()
	if err != nil {
		// gh returns error if release doesn't exist
		return false, nil
	}
	return true, nil
}
