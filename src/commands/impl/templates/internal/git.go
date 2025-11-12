package templates

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GitCloner handles cloning Git repositories
type GitCloner struct {
	repoURL    string
	targetDir  string
	branch     string
}

// NewGitCloner creates a new Git cloner
func NewGitCloner(repoURL string) *GitCloner {
	return &GitCloner{
		repoURL: repoURL,
		branch:  "main", // Always use main branch
	}
}

// CloneToTemp clones the repository to a temporary directory
// Returns the path to the cloned directory
func (g *GitCloner) CloneToTemp() (string, error) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "templates-clone-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	g.targetDir = tmpDir

	// Clone repository with depth=1 for speed (only latest commit)
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", g.branch, g.repoURL, tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tmpDir) // Clean up on error
		return "", fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, string(output))
	}

	// Remove .git directory to save space and avoid confusion
	gitDir := filepath.Join(tmpDir, ".git")
	os.RemoveAll(gitDir)

	return tmpDir, nil
}

// Cleanup removes the temporary clone directory
func (g *GitCloner) Cleanup() error {
	if g.targetDir != "" {
		return os.RemoveAll(g.targetDir)
	}
	return nil
}

// IsGitRepository checks if a path is a Git repository URL
func IsGitRepository(path string) bool {
	// Check for common Git URL patterns
	// https://github.com/user/repo
	// git@github.com:user/repo.git
	// https://gitlab.com/user/repo.git
	if len(path) < 4 {
		return false
	}

	// HTTP/HTTPS URLs
	if len(path) > 8 && (path[:8] == "https://" || path[:7] == "http://") {
		return true
	}

	// SSH URLs (git@...)
	if len(path) > 4 && path[:4] == "git@" {
		return true
	}

	// Git protocol
	if len(path) > 6 && path[:6] == "git://" {
		return true
	}

	return false
}
