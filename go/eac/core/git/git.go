// Package git provides Git operations using go-git library.
// It replaces direct exec.Command("git", ...) calls with pure Go implementations.
package git

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/zap"
)

// runGitCommand executes a git command in the repository and returns the output.
// This is used for performance-critical operations where native git is faster than go-git.
func (r *Repository) runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.rootPath

	output, err := cmd.Output()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("git %s failed: %s", args[0], string(exitErr.Stderr))
		}
		return "", err
	}

	return string(output), nil
}

// Repository wraps a go-git repository with convenience methods.
// Should be created via RepositoryManager for proper logger injection.
type Repository struct {
	repo     *gogit.Repository
	rootPath string
	logger   *zap.Logger // Logger for observability (guaranteed non-nil when created via manager)
}

// RootPath returns the repository root directory path.
func (r *Repository) RootPath() string {
	return r.rootPath
}

// RemoteURL returns the URL for the named remote (typically "origin").
func (r *Repository) RemoteURL(remoteName string) (string, error) {
	remote, err := r.repo.Remote(remoteName)
	if err != nil {
		return "", fmt.Errorf("failed to get remote %q: %w", remoteName, err)
	}

	cfg := remote.Config()
	if len(cfg.URLs) == 0 {
		return "", fmt.Errorf("remote %q has no URLs configured", remoteName)
	}

	return cfg.URLs[0], nil
}

// CurrentBranch returns the current branch name.
// If in detached HEAD state, returns "detached-<short-sha>".
func (r *Repository) CurrentBranch() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}

	// Detached HEAD state - return short SHA
	shortSHA := head.Hash().String()[:7]
	return "detached-" + shortSHA, nil
}

// HeadShortSHA returns the short SHA of the HEAD commit.
func (r *Repository) HeadShortSHA() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return head.Hash().String()[:7], nil
}

// TrackedFiles returns all files tracked by Git (in the index).
// This reflects the current index state: HEAD files + staged additions - staged deletions.
// Uses native git command for performance (go-git's wt.Status() is slow).
func (r *Repository) TrackedFiles() ([]string, error) {
	// Use native git ls-files which is fast locally, slow in CI
	// This lists all files in the index (tracked files)
	output, err := r.runGitCommand("ls-files")
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked files: %w", err)
	}

	if output == "" {
		return []string{}, nil
	}

	files := strings.Split(strings.TrimSpace(output), "\n")
	return files, nil
}

// StagedFiles returns files currently staged in the index (added, modified, renamed).
// This corresponds to `git diff --cached --name-only --diff-filter=ACMR`.
// Uses native git command for performance (avoids slow working tree scan).
func (r *Repository) StagedFiles() ([]string, error) {
	r.logger.Debug("Getting staged files")

	output, err := r.runGitCommand("diff", "--cached", "--name-only", "--diff-filter=ACMR")
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	if output == "" {
		return []string{}, nil
	}

	files := strings.Split(strings.TrimSpace(output), "\n")

	r.logger.Debug("Staged files retrieved", zap.Int("count", len(files)))
	return files, nil
}

// IsFileTracked checks if a specific file is tracked by Git (in the index).
// This reflects the current index state, accounting for staged additions and deletions.
func (r *Repository) IsFileTracked(relPath string) bool {
	// Normalize path separators
	relPath = filepath.ToSlash(relPath)

	// Check if in HEAD
	inHead := false
	head, err := r.repo.Head()
	if err == nil {
		commit, err := r.repo.CommitObject(head.Hash())
		if err == nil {
			tree, err := commit.Tree()
			if err == nil {
				_, err = tree.File(relPath)
				inHead = (err == nil)
			}
		}
	}

	// Check staging status
	wt, err := r.repo.Worktree()
	if err != nil {
		return inHead
	}

	status, err := wt.Status()
	if err != nil {
		return inHead
	}

	if fileStatus, exists := status[relPath]; exists {
		switch fileStatus.Staging {
		case gogit.Added:
			return true
		case gogit.Deleted:
			return false
		}
	}

	return inHead
}

// IsFileIgnored checks if a file matches .gitignore patterns.
func (r *Repository) IsFileIgnored(relPath string) bool {
	wt, err := r.repo.Worktree()
	if err != nil {
		return false
	}

	// Load gitignore patterns
	patterns, err := gitignore.ReadPatterns(wt.Filesystem, nil)
	if err != nil {
		return false
	}

	matcher := gitignore.NewMatcher(patterns)

	// Normalize path and split into components
	relPath = filepath.ToSlash(relPath)
	pathParts := strings.Split(relPath, "/")

	return matcher.Match(pathParts, false)
}

// ConfigSet sets a Git configuration value.
func (r *Repository) ConfigSet(section, key, value string) error {
	cfg, err := r.repo.Config()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Handle user.name and user.email specially
	if section == "user" {
		if cfg.User.Name == "" && key == "name" {
			cfg.User.Name = value
		} else if cfg.User.Email == "" && key == "email" {
			cfg.User.Email = value
		}

		switch key {
		case "name":
			cfg.User.Name = value
		case "email":
			cfg.User.Email = value
		}
	} else {
		// For other config, use raw sections
		if cfg.Raw.Section(section) == nil {
			cfg.Raw.AddOption(section, "", key, value)
		} else {
			cfg.Raw.Section(section).SetOption(key, value)
		}
	}

	return r.repo.SetConfig(cfg)
}

// Add stages a file for commit.
func (r *Repository) Add(path string) error {
	wt, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = wt.Add(path)
	if err != nil {
		return fmt.Errorf("failed to add file %q: %w", path, err)
	}

	return nil
}

// Commit creates a new commit with the staged changes.
func (r *Repository) Commit(message, authorName, authorEmail string) (string, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	sig := &object.Signature{
		Name:  authorName,
		Email: authorEmail,
		When:  timeNow(),
	}

	hash, err := wt.Commit(message, &gogit.CommitOptions{
		Author: sig,
	})
	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	return hash.String(), nil
}

// AddRemote adds a new remote to the repository.
func (r *Repository) AddRemote(name, url string) error {
	_, err := r.repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
	if err != nil {
		return fmt.Errorf("failed to add remote %q: %w", name, err)
	}
	return nil
}

// StagedDiff returns the unified diff of all staged changes.
// Equivalent to `git diff --staged`.
// Uses native git command for performance (avoids slow working tree scan).
func (r *Repository) StagedDiff() (string, error) {
	r.logger.Debug("Getting staged diff")

	output, err := r.runGitCommand("diff", "--cached")
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	r.logger.Debug("Staged diff retrieved", zap.Int("size", len(output)))
	return output, nil
}

// StagedDiffStats returns the stat summary of staged changes.
// Equivalent to `git diff --staged --stat`.
// Uses native git command for performance (avoids slow working tree scan).
func (r *Repository) StagedDiffStats() (string, error) {
	r.logger.Debug("Getting staged diff stats")

	output, err := r.runGitCommand("diff", "--cached", "--stat")
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff stats: %w", err)
	}

	r.logger.Debug("Staged diff stats retrieved")
	return output, nil
}

// GetBranchCommits returns all commits from baseBranch..HEAD.
// Returns commits in reverse chronological order (newest first).
// Returns error if baseBranch doesn't exist or no commits ahead.
func (r *Repository) GetBranchCommits(baseBranch string) ([]CommitInfo, error) {
	// Resolve base branch to verify it exists
	baseRef := baseBranch
	if !strings.HasPrefix(baseBranch, "origin/") && !strings.HasPrefix(baseBranch, "refs/") {
		// Try origin/baseBranch first (for remote tracking)
		_, err := r.runGitCommand("rev-parse", "--verify", "origin/"+baseBranch)
		if err == nil {
			baseRef = "origin/" + baseBranch
		}
	}

	// Use CommitsBetween to get commits from baseBranch..HEAD
	// This uses the two-dot notation which shows commits reachable from HEAD but not from baseBranch
	commits, err := r.CommitsBetween(baseRef, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get branch commits: %w", err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits ahead of %s", baseBranch)
	}

	return commits, nil
}

// GetBranchDiff returns the cumulative diff from baseBranch...HEAD.
// Uses three-dot notation to compare against merge-base.
func (r *Repository) GetBranchDiff(baseBranch string) (string, error) {
	// Resolve base branch reference
	baseRef := baseBranch
	if !strings.HasPrefix(baseBranch, "origin/") && !strings.HasPrefix(baseBranch, "refs/") {
		// Try origin/baseBranch first
		_, err := r.runGitCommand("rev-parse", "--verify", "origin/"+baseBranch)
		if err == nil {
			baseRef = "origin/" + baseBranch
		}
	}

	// Use three-dot notation to compare against merge-base
	output, err := r.runGitCommand("diff", baseRef+"...HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get branch diff: %w", err)
	}

	return output, nil
}

// GetBranchDiffStats returns diff statistics from baseBranch...HEAD.
// Returns summary like "3 files changed, 45 insertions(+), 12 deletions(-)".
func (r *Repository) GetBranchDiffStats(baseBranch string) (string, error) {
	// Resolve base branch reference
	baseRef := baseBranch
	if !strings.HasPrefix(baseBranch, "origin/") && !strings.HasPrefix(baseBranch, "refs/") {
		// Try origin/baseBranch first
		_, err := r.runGitCommand("rev-parse", "--verify", "origin/"+baseBranch)
		if err == nil {
			baseRef = "origin/" + baseBranch
		}
	}

	// Use three-dot notation with --stat
	output, err := r.runGitCommand("diff", baseRef+"...HEAD", "--stat")
	if err != nil {
		return "", fmt.Errorf("failed to get branch diff stats: %w", err)
	}

	return output, nil
}

// GetBranchFiles returns list of files changed in baseBranch...HEAD.
// Returns relative paths from repository root.
func (r *Repository) GetBranchFiles(baseBranch string) ([]string, error) {
	// Resolve base branch reference
	baseRef := baseBranch
	if !strings.HasPrefix(baseBranch, "origin/") && !strings.HasPrefix(baseBranch, "refs/") {
		// Try origin/baseBranch first
		_, err := r.runGitCommand("rev-parse", "--verify", "origin/"+baseBranch)
		if err == nil {
			baseRef = "origin/" + baseBranch
		}
	}

	// Use three-dot notation with --name-only
	output, err := r.runGitCommand("diff", baseRef+"...HEAD", "--name-only")
	if err != nil {
		return nil, fmt.Errorf("failed to get branch files: %w", err)
	}

	// Split output into lines and filter empty lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// GoGitRepo returns the underlying go-git repository for advanced operations.
// Use sparingly - prefer the wrapper methods when possible.
func (r *Repository) GoGitRepo() *gogit.Repository {
	return r.repo
}

// WorktreeList returns information about all worktrees.
// Uses native git command for reliable worktree detection.
func (r *Repository) WorktreeList() ([]WorktreeEntry, error) {
	r.logger.Debug("Listing worktrees")

	// Use git worktree list --porcelain for structured output
	output, err := r.runGitCommand("worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Parse porcelain output
	// Format:
	// worktree <path>
	// HEAD <sha>
	// branch refs/heads/<branch>
	// <blank line>
	// (repeat for each worktree)

	var worktrees []WorktreeEntry
	var current WorktreeEntry

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			// End of current worktree entry
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = WorktreeEntry{}
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "worktree":
			current.Path = parts[1]
		case "HEAD":
			sha := parts[1]
			if len(sha) > 7 {
				current.SHA = sha[:7]
			} else {
				current.SHA = sha
			}
		case "branch":
			// Extract branch name from refs/heads/<branch>
			branch := strings.TrimPrefix(parts[1], "refs/heads/")
			current.Branch = branch
		case "detached":
			// Detached HEAD state
			current.Branch = "detached-" + current.SHA
		}
	}

	// Don't forget the last entry if output doesn't end with blank line
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	r.logger.Debug("Worktrees listed", zap.Int("count", len(worktrees)))
	return worktrees, nil
}

// WorktreeIsDirty checks if a worktree has uncommitted changes.
// Returns true if there are modified, added, or deleted files.
func (r *Repository) WorktreeIsDirty(worktreePath string) (bool, error) {
	r.logger.Debug("Checking worktree status", zap.String("path", worktreePath))

	// Use git status --porcelain in the worktree directory
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = worktreePath

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check worktree status: %w", err)
	}

	// If output is empty, worktree is clean
	isDirty := strings.TrimSpace(string(output)) != ""

	r.logger.Debug("Worktree status checked",
		zap.String("path", worktreePath),
		zap.Bool("isDirty", isDirty))

	return isDirty, nil
}
