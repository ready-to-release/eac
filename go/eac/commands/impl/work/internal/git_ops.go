// Package internal provides shared utilities for work commands
package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

// WorkGitOperations defines git operations needed by work commands.
// This interface allows for mocking in tests.
type WorkGitOperations interface {
	// Worktree operations
	CreateWorktree(path, branch, base string) error
	RemoveWorktree(path string) error
	ListWorktrees() ([]Worktree, error)
	WorktreeExists(branch string) (bool, error)

	// Branch operations
	BranchExists(branch string) (bool, error)
	GetCurrentBranch(path string) (string, error)
	DeleteBranch(branch string, force bool) error

	// Status operations
	IsWorktreeClean(path string) (bool, error)

	// Remote operations
	FetchBranch(branch string) error
	PushBranch(branch string, force bool) error

	// Merge/Rebase operations
	Rebase(target string) error
	Merge(branch string, squash bool) error
	MergeAbort() error
	RebaseAbort() error

	// Stash operations
	Stash(message string) error
	StashPop() error

	// Commit operations
	GetCommitCount(base, head string) (int, error)
	GetConflictingFiles() ([]string, error)
}

// defaultGitOps implements WorkGitOperations using real git commands.
type defaultGitOps struct {
	repoRoot string
}

// gitOps holds the current git operations implementation.
// In production, this is nil and defaults to real git commands.
// In tests, this can be set to a mock implementation.
var gitOps WorkGitOperations

// GetGitOps returns the git operations interface.
// If a mock has been set via SetGitOps, returns that.
// Otherwise, returns a new real implementation.
func GetGitOps(repoRoot string) WorkGitOperations {
	if gitOps != nil {
		return gitOps
	}
	return &defaultGitOps{repoRoot: repoRoot}
}

// SetGitOps allows tests to inject a mock implementation.
func SetGitOps(ops WorkGitOperations) {
	gitOps = ops
}

// ResetGitOps clears the mock implementation (for test cleanup).
func ResetGitOps() {
	gitOps = nil
}

// CreateWorktree creates a new git worktree.
func (g *defaultGitOps) CreateWorktree(path, branch, base string) error {
	cmd := exec.Command("git", "worktree", "add", path, "-b", branch, base)
	cmd.Dir = g.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// RemoveWorktree removes a git worktree.
func (g *defaultGitOps) RemoveWorktree(path string) error {
	cmd := exec.Command("git", "worktree", "remove", path)
	cmd.Dir = g.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// ListWorktrees returns all worktrees in the repository.
func (g *defaultGitOps) ListWorktrees() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = g.repoRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	return parseWorktreeList(string(output))
}

// WorktreeExists checks if a worktree exists for a given branch.
func (g *defaultGitOps) WorktreeExists(branch string) (bool, error) {
	worktrees, err := g.ListWorktrees()
	if err != nil {
		return false, err
	}

	for _, wt := range worktrees {
		if wt.Branch == branch {
			return true, nil
		}
	}
	return false, nil
}

// BranchExists checks if a branch exists in the repository.
func (g *defaultGitOps) BranchExists(branch string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", fmt.Sprintf("refs/heads/%s", branch))
	cmd.Dir = g.repoRoot
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 128 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check branch: %w", err)
	}
	return true, nil
}

// GetCurrentBranch returns the current branch name.
func (g *defaultGitOps) GetCurrentBranch(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// DeleteBranch deletes a branch.
func (g *defaultGitOps) DeleteBranch(branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	cmd := exec.Command("git", "branch", flag, branch)
	cmd.Dir = g.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %w\nOutput: %s", branch, err, string(output))
	}
	return nil
}

// IsWorktreeClean checks if a worktree has uncommitted changes.
func (g *defaultGitOps) IsWorktreeClean(path string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check worktree status: %w", err)
	}
	return len(strings.TrimSpace(string(output))) == 0, nil
}

// FetchBranch fetches a branch from origin.
func (g *defaultGitOps) FetchBranch(branch string) error {
	cmd := exec.Command("git", "fetch", "origin", branch)
	cmd.Dir = g.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to fetch origin/%s: %w\nOutput: %s", branch, err, string(output))
	}
	return nil
}

// PushBranch pushes a branch to origin.
func (g *defaultGitOps) PushBranch(branch string, force bool) error {
	args := []string{"push", "origin", branch}
	if force {
		args = append(args, "--force")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push %s: %w\nOutput: %s", branch, err, string(output))
	}
	return nil
}

// Rebase rebases current branch onto target.
func (g *defaultGitOps) Rebase(target string) error {
	cmd := exec.Command("git", "rebase", fmt.Sprintf("origin/%s", target))
	cmd.Dir = g.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rebase failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// Merge merges a branch.
func (g *defaultGitOps) Merge(branch string, squash bool) error {
	args := []string{"merge"}
	if squash {
		args = append(args, "--squash")
	}
	args = append(args, branch)

	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("merge failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// MergeAbort aborts an in-progress merge.
func (g *defaultGitOps) MergeAbort() error {
	cmd := exec.Command("git", "merge", "--abort")
	cmd.Dir = g.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("merge abort failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// RebaseAbort aborts an in-progress rebase.
func (g *defaultGitOps) RebaseAbort() error {
	cmd := exec.Command("git", "rebase", "--abort")
	cmd.Dir = g.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rebase abort failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// Stash stashes uncommitted changes.
func (g *defaultGitOps) Stash(message string) error {
	cmd := exec.Command("git", "stash", "push", "-m", message)
	cmd.Dir = g.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stash changes: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// StashPop reapplies stashed changes.
func (g *defaultGitOps) StashPop() error {
	cmd := exec.Command("git", "stash", "pop")
	cmd.Dir = g.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reapply stashed changes: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// GetCommitCount returns the number of commits between base and head.
func (g *defaultGitOps) GetCommitCount(base, head string) (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("%s..%s", base, head))
	cmd.Dir = g.repoRoot
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}

	var count int
	_, err = fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count)
	if err != nil {
		return 0, fmt.Errorf("failed to parse commit count: %w", err)
	}
	return count, nil
}

// GetConflictingFiles returns a list of files with merge conflicts.
func (g *defaultGitOps) GetConflictingFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	cmd.Dir = g.repoRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get conflicting files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var conflicts []string
	for _, line := range lines {
		if line != "" {
			conflicts = append(conflicts, line)
		}
	}
	return conflicts, nil
}
