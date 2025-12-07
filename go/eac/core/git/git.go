// Package git provides Git operations using go-git library.
// It replaces direct exec.Command("git", ...) calls with pure Go implementations.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// log is the package-level logger for git operations
var log = logging.C()

// runGitCommand executes a git command in the repository and returns the output.
// This is used for performance-critical operations where native git is faster than go-git.
func (r *Repository) runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.rootPath

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s failed: %s", args[0], string(exitErr.Stderr))
		}
		return "", err
	}

	return string(output), nil
}

// Repository wraps a go-git repository with convenience methods.
type Repository struct {
	repo     *gogit.Repository
	rootPath string
}

// Open opens an existing Git repository at the given path.
// If path is empty, uses the current working directory.
// It searches upward through parent directories to find the repository root.
func Open(path string) (*Repository, error) {
	log.Debug("Open: start")
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Open repository, detecting .git directory by walking up
	log.Debug("Open: calling PlainOpenWithOptions")
	repo, err := gogit.PlainOpenWithOptions(absPath, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	log.Debug("Open: PlainOpenWithOptions complete")

	// Get the worktree to find the root path
	log.Debug("Open: getting worktree")
	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}
	log.Debug("Open: complete")

	return &Repository{
		repo:     repo,
		rootPath: wt.Filesystem.Root(),
	}, nil
}

// Init initializes a new Git repository at the given path.
func Init(path string) (*Repository, error) {
	repo, err := gogit.PlainInit(path, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	return &Repository{
		repo:     repo,
		rootPath: path,
	}, nil
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
	log.Debug("StagedFiles: start")
	log.Debug("StagedFiles: calling git diff --cached --name-only")

	output, err := r.runGitCommand("diff", "--cached", "--name-only", "--diff-filter=ACMR")
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}
	log.Debug("StagedFiles: git command complete")

	if output == "" {
		return []string{}, nil
	}

	files := strings.Split(strings.TrimSpace(output), "\n")
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
func (r *Repository) Commit(message string, authorName, authorEmail string) (string, error) {
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
	log.Debug("StagedDiff: start")
	log.Debug("StagedDiff: calling git diff --cached")

	output, err := r.runGitCommand("diff", "--cached")
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}
	log.Debug("StagedDiff: git command complete")

	return output, nil
}

// StagedDiffStats returns the stat summary of staged changes.
// Equivalent to `git diff --staged --stat`.
// Uses native git command for performance (avoids slow working tree scan).
func (r *Repository) StagedDiffStats() (string, error) {
	log.Debug("StagedDiffStats: start")
	log.Debug("StagedDiffStats: calling git diff --cached --stat")

	output, err := r.runGitCommand("diff", "--cached", "--stat")
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff stats: %w", err)
	}
	log.Debug("StagedDiffStats: git command complete")

	return output, nil
}

// generateUnifiedDiff creates a unified diff format for a single file
func generateUnifiedDiff(path, original, staged string, status gogit.StatusCode) string {
	var diff strings.Builder

	diff.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", path, path))

	switch status {
	case gogit.Added:
		diff.WriteString("new file mode 100644\n")
		diff.WriteString("--- /dev/null\n")
		diff.WriteString(fmt.Sprintf("+++ b/%s\n", path))

		lines := strings.Split(staged, "\n")
		if len(lines) > 0 {
			diff.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", len(lines)))
			for _, line := range lines {
				diff.WriteString("+" + line + "\n")
			}
		}

	case gogit.Deleted:
		diff.WriteString("deleted file mode 100644\n")
		diff.WriteString(fmt.Sprintf("--- a/%s\n", path))
		diff.WriteString("+++ /dev/null\n")

		lines := strings.Split(original, "\n")
		if len(lines) > 0 {
			diff.WriteString(fmt.Sprintf("@@ -1,%d +0,0 @@\n", len(lines)))
			for _, line := range lines {
				diff.WriteString("-" + line + "\n")
			}
		}

	default: // Modified, Renamed, Copied
		diff.WriteString(fmt.Sprintf("--- a/%s\n", path))
		diff.WriteString(fmt.Sprintf("+++ b/%s\n", path))

		// Simple diff: show all lines changed
		origLines := strings.Split(original, "\n")
		newLines := strings.Split(staged, "\n")

		diff.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", len(origLines), len(newLines)))
		for _, line := range origLines {
			diff.WriteString("-" + line + "\n")
		}
		for _, line := range newLines {
			diff.WriteString("+" + line + "\n")
		}
	}

	return diff.String()
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
