package repository

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/src/core/git"
)

// GitContext contains Git repository context information
type GitContext struct {
	RepositoryURL string // GitHub repository URL (e.g., "https://github.com/owner/repo")
	BaseCommit    string // Closest server-known commit SHA
	CurrentBranch string // Current branch name
}

// GetGitContext retrieves Git context for generating stable GitHub links.
// It finds:
// - The GitHub repository URL from remote 'origin'
// - The closest server-known commit (merge-base with origin/main or main)
// - The current branch name
//
// Parameters:
//   - repo: GitRepository interface for git operations
func GetGitContext(repo git.GitRepository) (*GitContext, error) {
	ctx := &GitContext{}
	rootPath := repo.RootPath()

	// Get remote URL
	remoteURL, err := repo.RemoteURL("origin")
	if err != nil {
		return nil, NewRepositoryError("remote get-url", rootPath, err, "failed to get remote URL")
	}
	ctx.RepositoryURL = normalizeGitHubURL(strings.TrimSpace(remoteURL))

	// Get current branch
	branch, err := repo.CurrentBranch()
	if err != nil {
		return nil, NewRepositoryError("branch", rootPath, err, "failed to get current branch")
	}
	ctx.CurrentBranch = branch

	// Get base commit (merge-base with main)
	// For pre-commit/pre-push documentation, use "main" branch name instead of SHA
	// This creates stable links that work once changes are committed and pushed
	ctx.BaseCommit = "main"

	return ctx, nil
}

// normalizeGitHubURL converts various Git URL formats to HTTPS GitHub URL
// Examples:
//
//	git@github.com:owner/repo.git -> https://github.com/owner/repo
//	https://github.com/owner/repo.git -> https://github.com/owner/repo
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
