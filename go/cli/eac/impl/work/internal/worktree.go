// Package internal provides shared utilities for work commands
package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/gitexec"
)

// Worktree represents a git worktree.
type Worktree struct {
	Path   string
	Branch string
	SHA    string
	Clean  bool
}

// GetWorktrees returns all worktrees in the repository.
func GetWorktrees(repoRoot string) ([]Worktree, error) {
	return GetGitOps(repoRoot).ListWorktrees()
}

// parseWorktreeList parses the output of `git worktree list --porcelain`.
func parseWorktreeList(output string) ([]Worktree, error) {
	var worktrees []Worktree
	lines := strings.Split(strings.TrimSpace(output), "\n")

	var current *Worktree
	for _, line := range lines {
		if line == "" {
			if current != nil {
				worktrees = append(worktrees, *current)
				current = nil
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current = &Worktree{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		} else if strings.HasPrefix(line, "HEAD ") && current != nil {
			current.SHA = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") && current != nil {
			branch := strings.TrimPrefix(line, "branch ")
			// Remove refs/heads/ prefix
			branch = strings.TrimPrefix(branch, "refs/heads/")
			current.Branch = branch
		}
	}

	// Add last worktree if exists
	if current != nil {
		worktrees = append(worktrees, *current)
	}

	// Check clean status for each worktree
	for i := range worktrees {
		clean, err := IsWorktreeClean(worktrees[i].Path)
		if err != nil {
			// If we can't determine, assume dirty for safety
			worktrees[i].Clean = false
		} else {
			worktrees[i].Clean = clean
		}
	}

	return worktrees, nil
}

// FindWorktreeByBranch finds a worktree by branch name.
func FindWorktreeByBranch(branch, repoRoot string) (*Worktree, error) {
	worktrees, err := GetWorktrees(repoRoot)
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Branch == branch {
			return &wt, nil
		}
	}

	return nil, fmt.Errorf("worktree not found for branch: %s", branch)
}

// GetMainWorktree returns the main worktree (usually on main branch).
func GetMainWorktree(repoRoot string) (*Worktree, error) {
	worktrees, err := GetWorktrees(repoRoot)
	if err != nil {
		return nil, err
	}

	// Look for main or master branch
	for _, wt := range worktrees {
		if wt.Branch == "main" || wt.Branch == "master" {
			return &wt, nil
		}
	}

	// If not found, return first worktree
	if len(worktrees) > 0 {
		return &worktrees[0], nil
	}

	return nil, fmt.Errorf("no worktrees found")
}

// GenerateWorktreePath creates a standard worktree path from repo name and branch.
func GenerateWorktreePath(repoName, branchName string) string {
	// Sanitize branch name for path (replace / with -)
	safeBranch := strings.ReplaceAll(branchName, "/", "-")
	return filepath.Join("..", fmt.Sprintf("%s-%s", repoName, safeBranch))
}

// IsWorktreeClean checks if a worktree has uncommitted changes.
func IsWorktreeClean(path string) (bool, error) {
	// Note: GetGitOps requires repoRoot, but IsWorktreeClean in GitOps
	// uses the path parameter as the directory, so we pass path as repoRoot
	return GetGitOps(path).IsWorktreeClean(path)
}

// GetRepoName extracts the repository name from the repo root path.
func GetRepoName(repoRoot string) string {
	return filepath.Base(repoRoot)
}

// WorktreeExists checks if a worktree already exists for a branch.
func WorktreeExists(branch, repoRoot string) (bool, error) {
	_, err := FindWorktreeByBranch(branch, repoRoot)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// BranchExists checks if a branch exists in the repository.
func BranchExists(branch, repoRoot string) (bool, error) {
	return GetGitOps(repoRoot).BranchExists(branch)
}

// GetCurrentBranch returns the current branch name.
func GetCurrentBranch(path string) (string, error) {
	return GetGitOps(path).GetCurrentBranch(path)
}

// EnsureInGitRepo checks if we're in a git repository.
func EnsureInGitRepo() error {
	_, err := gitexec.Run(".", "rev-parse", "--git-dir")
	if err != nil {
		return fmt.Errorf("not in a git repository")
	}
	return nil
}

// GetAbsolutePath converts a relative path to absolute.
func GetAbsolutePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	absPath, err := filepath.Abs(filepath.Join(cwd, path))
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}
