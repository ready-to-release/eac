// Package internal provides shared utilities for work commands
package internal

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/core/logging"
)

var log = logging.C()

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
	log.Debugf("Git operation: createWorktree | path=%s branch=%s base=%s repoRoot=%s", path, branch, base, g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "worktree", "add", path, "-b", branch, base)
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.CombinedOutput()

	log.Debugf("Git operation completed: createWorktree | duration=%v output=%q err=%v", time.Since(start), string(output), err)

	if err != nil {
		return fmt.Errorf("failed to create worktree: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// RemoveWorktree removes a git worktree.
func (g *defaultGitOps) RemoveWorktree(path string) error {
	log.Debugf("Git operation: removeWorktree | path=%s repoRoot=%s", path, g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "worktree", "remove", path)
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.CombinedOutput()

	log.Debugf("Git operation completed: removeWorktree | duration=%v output=%q err=%v", time.Since(start), string(output), err)

	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// ListWorktrees returns all worktrees in the repository.
func (g *defaultGitOps) ListWorktrees() ([]Worktree, error) {
	log.Debugf("Git operation: listWorktrees | repoRoot=%s", g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.Output()

	log.Debugf("Git operation completed: listWorktrees | duration=%v output=%q err=%v", time.Since(start), string(output), err)

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
	log.Debugf("Git operation: branchExists | branch=%s repoRoot=%s", branch, g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "rev-parse", "--verify", fmt.Sprintf("refs/heads/%s", branch))
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	err := cmd.Run()

	log.Debugf("Git operation completed: branchExists | duration=%v exists=%v err=%v", time.Since(start), err == nil, err)

	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check branch: %w", err)
	}
	return true, nil
}

// GetCurrentBranch returns the current branch name.
func (g *defaultGitOps) GetCurrentBranch(path string) (string, error) {
	log.Debugf("Git operation: getCurrentBranch | path=%s repoRoot=%s", path, g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.Output()
	branch := strings.TrimSpace(string(output))

	log.Debugf("Git operation completed: getCurrentBranch | duration=%v branch=%s output=%q err=%v", time.Since(start), branch, string(output), err)

	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return branch, nil
}

// DeleteBranch deletes a branch.
func (g *defaultGitOps) DeleteBranch(branch string, force bool) error {
	log.Debugf("Git operation: deleteBranch | branch=%s force=%v repoRoot=%s", branch, force, g.repoRoot)

	start := time.Now()

	flag := "-d"
	if force {
		flag = "-D"
	}
	cmd := exec.Command("git", "branch", flag, branch)
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.CombinedOutput()

	log.Debugf("Git operation completed: deleteBranch | duration=%v output=%q err=%v", time.Since(start), string(output), err)

	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %w\nOutput: %s", branch, err, string(output))
	}
	return nil
}

// IsWorktreeClean checks if a worktree has uncommitted changes.
func (g *defaultGitOps) IsWorktreeClean(path string) (bool, error) {
	log.Debugf("Git operation: isWorktreeClean | path=%s repoRoot=%s", path, g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.Output()
	isClean := strings.TrimSpace(string(output)) == ""

	log.Debugf("Git operation completed: isWorktreeClean | duration=%v isClean=%v output=%q err=%v", time.Since(start), isClean, string(output), err)

	if err != nil {
		return false, fmt.Errorf("failed to check worktree status: %w", err)
	}
	return isClean, nil
}

// FetchBranch fetches a branch from origin.
func (g *defaultGitOps) FetchBranch(branch string) error {
	log.Debugf("Git operation: fetchBranch | branch=%s repoRoot=%s", branch, g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "fetch", "origin", branch)
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.CombinedOutput()

	log.Debugf("Git operation completed: fetchBranch | duration=%v output=%q err=%v", time.Since(start), string(output), err)

	if err != nil {
		return fmt.Errorf("failed to fetch origin/%s: %w\nOutput: %s", branch, err, string(output))
	}
	return nil
}

// PushBranch pushes a branch to origin.
func (g *defaultGitOps) PushBranch(branch string, force bool) error {
	log.Debugf("Git operation: pushBranch | branch=%s force=%v repoRoot=%s", branch, force, g.repoRoot)

	start := time.Now()

	args := []string{"push", "origin", branch}
	if force {
		args = append(args, "--force")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.CombinedOutput()

	log.Debugf("Git operation completed: pushBranch | duration=%v output=%q err=%v", time.Since(start), string(output), err)

	if err != nil {
		return fmt.Errorf("failed to push %s: %w\nOutput: %s", branch, err, string(output))
	}
	return nil
}

// Rebase rebases current branch onto target.
func (g *defaultGitOps) Rebase(target string) error {
	log.Debugf("Git operation: rebase | target=%s repoRoot=%s", target, g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "rebase", fmt.Sprintf("origin/%s", target))
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.CombinedOutput()

	log.Debugf("Git operation completed: rebase | duration=%v output=%q err=%v", time.Since(start), string(output), err)

	if err != nil {
		return fmt.Errorf("rebase failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// Merge merges a branch.
func (g *defaultGitOps) Merge(branch string, squash bool) error {
	log.Debugf("Git operation: merge | branch=%s squash=%v repoRoot=%s", branch, squash, g.repoRoot)

	start := time.Now()

	args := []string{"merge"}
	if squash {
		args = append(args, "--squash")
	}
	args = append(args, branch)

	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.CombinedOutput()

	log.Debugf("Git operation completed: merge | duration=%v output=%q err=%v", time.Since(start), string(output), err)

	if err != nil {
		return fmt.Errorf("merge failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// MergeAbort aborts an in-progress merge.
func (g *defaultGitOps) MergeAbort() error {
	log.Debugf("Git operation: mergeAbort | repoRoot=%s", g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "merge", "--abort")
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.CombinedOutput()

	log.Debugf("Git operation completed: mergeAbort | duration=%v output=%q err=%v", time.Since(start), string(output), err)

	if err != nil {
		return fmt.Errorf("merge abort failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// RebaseAbort aborts an in-progress rebase.
func (g *defaultGitOps) RebaseAbort() error {
	log.Debugf("Git operation: rebaseAbort | repoRoot=%s", g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "rebase", "--abort")
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.CombinedOutput()

	log.Debugf("Git operation completed: rebaseAbort | duration=%v output=%q err=%v", time.Since(start), string(output), err)

	if err != nil {
		return fmt.Errorf("rebase abort failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// Stash stashes uncommitted changes.
func (g *defaultGitOps) Stash(message string) error {
	log.Debugf("Git operation: stash | message=%s repoRoot=%s", message, g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "stash", "push", "-m", message)
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.CombinedOutput()

	log.Debugf("Git operation completed: stash | duration=%v output=%q err=%v", time.Since(start), string(output), err)

	if err != nil {
		return fmt.Errorf("failed to stash changes: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// StashPop reapplies stashed changes.
func (g *defaultGitOps) StashPop() error {
	log.Debugf("Git operation: stashPop | repoRoot=%s", g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "stash", "pop")
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.CombinedOutput()

	log.Debugf("Git operation completed: stashPop | duration=%v output=%q err=%v", time.Since(start), string(output), err)

	if err != nil {
		return fmt.Errorf("failed to reapply stashed changes: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// GetCommitCount returns the number of commits between base and head.
func (g *defaultGitOps) GetCommitCount(base, head string) (int, error) {
	log.Debugf("Git operation: getCommitCount | base=%s head=%s repoRoot=%s", base, head, g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("%s..%s", base, head))
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.Output()

	var count int
	if err == nil {
		_, err = fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count)
		if err != nil {
			err = fmt.Errorf("failed to parse commit count: %w", err)
		}
	}

	log.Debugf("Git operation completed: getCommitCount | duration=%v count=%d output=%q err=%v", time.Since(start), count, string(output), err)

	if err != nil {
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}
	return count, nil
}

// GetConflictingFiles returns a list of files with merge conflicts.
func (g *defaultGitOps) GetConflictingFiles() ([]string, error) {
	log.Debugf("Git operation: getConflictingFiles | repoRoot=%s", g.repoRoot)

	start := time.Now()

	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	cmd.Dir = g.repoRoot

	log.Debugf("Executing git command: %s (dir=%s)", cmd.String(), cmd.Dir)

	output, err := cmd.Output()

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var conflicts []string
	for _, line := range lines {
		if line != "" {
			conflicts = append(conflicts, line)
		}
	}

	log.Debugf("Git operation completed: getConflictingFiles | duration=%v conflictCount=%d conflicts=%v output=%q err=%v", time.Since(start), len(conflicts), conflicts, string(output), err)

	if err != nil {
		return nil, fmt.Errorf("failed to get conflicting files: %w", err)
	}
	return conflicts, nil
}
