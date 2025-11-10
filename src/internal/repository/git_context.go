package repository

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitContext contains Git repository context information
type GitContext struct {
	RepositoryURL string // GitHub repository URL (e.g., "https://github.com/owner/repo")
	BaseCommit    string // Closest server-known commit SHA
	CurrentBranch string // Current branch name
}

// GetGitContext retrieves Git context for generating stable GitHub links
// It finds:
// - The GitHub repository URL from remote 'origin'
// - The closest server-known commit (merge-base with origin/main or main)
// - The current branch name
func GetGitContext(rootPath string) (*GitContext, error) {
	if rootPath == "" {
		var err error
		rootPath, err = GetRepositoryRoot("")
		if err != nil {
			return nil, err
		}
	}

	ctx := &GitContext{}

	// Get remote URL
	remoteURL, err := getRemoteURL(rootPath)
	if err != nil {
		return nil, err
	}
	ctx.RepositoryURL = normalizeGitHubURL(remoteURL)

	// Get current branch
	branch, err := getCurrentBranch(rootPath)
	if err != nil {
		return nil, err
	}
	ctx.CurrentBranch = branch

	// Get base commit (merge-base with main)
	baseCommit, err := getBaseCommit(rootPath)
	if err != nil {
		return nil, err
	}
	ctx.BaseCommit = baseCommit

	return ctx, nil
}

// getRemoteURL gets the remote URL for 'origin'
func getRemoteURL(rootPath string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = rootPath
	output, err := cmd.Output()
	if err != nil {
		return "", NewRepositoryError("remote get-url", rootPath, err, "failed to get remote URL")
	}

	return strings.TrimSpace(string(output)), nil
}

// getCurrentBranch gets the current branch name
func getCurrentBranch(rootPath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = rootPath
	output, err := cmd.Output()
	if err != nil {
		return "", NewRepositoryError("branch --show-current", rootPath, err, "failed to get current branch")
	}

	branch := strings.TrimSpace(string(output))
	if branch == "" {
		// Detached HEAD state - try to get commit SHA
		cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
		cmd.Dir = rootPath
		output, err = cmd.Output()
		if err != nil {
			return "", NewRepositoryError("rev-parse HEAD", rootPath, err, "failed to get HEAD commit")
		}
		return "detached-" + strings.TrimSpace(string(output)), nil
	}

	return branch, nil
}

// getBaseCommit finds the closest server-known commit
// Strategy:
// For pre-commit/pre-push documentation, use "main" branch name instead of SHA
// This creates stable links that work once changes are committed and pushed
//
// Note: Using branch name "main" instead of commit SHA ensures links work
// regardless of local commit state, as long as changes eventually reach main branch
func getBaseCommit(rootPath string) (string, error) {
	// Return "main" as the ref - GitHub will resolve this to current main HEAD
	// This handles uncommitted, committed-but-not-pushed, and pushed states
	return "main", nil
}

// normalizeGitHubURL converts various Git URL formats to HTTPS GitHub URL
// Examples:
//   git@github.com:owner/repo.git -> https://github.com/owner/repo
//   https://github.com/owner/repo.git -> https://github.com/owner/repo
func normalizeGitHubURL(remoteURL string) string {
	// Remove .git suffix
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// Convert SSH format to HTTPS
	if strings.HasPrefix(remoteURL, "git@github.com:") {
		remoteURL = strings.Replace(remoteURL, "git@github.com:", "https://github.com/", 1)
	}

	return remoteURL
}

// BuildGitHubFileURL builds a GitHub URL for a file at the base commit
// Example: https://github.com/owner/repo/blob/abc123/path/to/file.feature
func (ctx *GitContext) BuildGitHubFileURL(filePath string) string {
	// Normalize path separators to forward slashes
	filePath = strings.ReplaceAll(filePath, "\\", "/")

	return fmt.Sprintf("%s/blob/%s/%s", ctx.RepositoryURL, ctx.BaseCommit, filePath)
}

// BuildGitHubBlobURL is an alias for BuildGitHubFileURL
func (ctx *GitContext) BuildGitHubBlobURL(filePath string) string {
	return ctx.BuildGitHubFileURL(filePath)
}
