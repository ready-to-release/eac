package git

import (
	"time"

	gogit "github.com/go-git/go-git/v5"
)

// GitRepository defines the interface for git operations.
// This allows mocking in tests and supports alternative implementations.
type GitRepository interface {
	// RootPath returns the repository root directory path.
	RootPath() string

	// RemoteURL returns the URL for the named remote (typically "origin").
	RemoteURL(remoteName string) (string, error)

	// AddRemote adds a new remote to the repository.
	AddRemote(name, url string) error

	// CurrentBranch returns the current branch name.
	// If in detached HEAD state, returns "detached-<short-sha>".
	CurrentBranch() (string, error)

	// HeadShortSHA returns the short SHA of the HEAD commit.
	HeadShortSHA() (string, error)

	// TrackedFiles returns all files tracked by Git (in the index).
	TrackedFiles() ([]string, error)

	// StagedFiles returns files currently staged in the index.
	StagedFiles() ([]string, error)

	// IsFileTracked checks if a specific file is tracked by Git.
	IsFileTracked(relPath string) bool

	// IsFileIgnored checks if a file matches .gitignore patterns.
	IsFileIgnored(relPath string) bool

	// Add stages a file for commit.
	Add(path string) error

	// Commit creates a new commit with the staged changes.
	// Returns the commit hash.
	Commit(message, authorName, authorEmail string) (string, error)

	// StagedDiff returns the unified diff of all staged changes.
	// Equivalent to `git diff --staged`.
	StagedDiff() (string, error)

	// StagedDiffStats returns the stat summary of staged changes.
	// Equivalent to `git diff --staged --stat`.
	StagedDiffStats() (string, error)

	// ConfigSet sets a Git configuration value.
	ConfigSet(section, key, value string) error

	// GoGitRepo returns the underlying go-git repository for advanced operations.
	// Returns nil for mock implementations.
	GoGitRepo() *gogit.Repository

	// --- Changelog/Release related operations ---

	// CommitsBetween returns commits between two references (tag/SHA/branch).
	// If fromRef is empty, returns all commits up to toRef.
	// Returns commits in reverse chronological order (newest first).
	CommitsBetween(fromRef, toRef string) ([]CommitInfo, error)

	// CommitsSince returns all commits since a given tag or reference.
	CommitsSince(ref string) ([]CommitInfo, error)

	// TagsMatching returns tags matching a glob pattern (e.g., "r2r-cli/*").
	TagsMatching(pattern string) ([]string, error)

	// LatestTag returns the most recent tag matching the pattern.
	LatestTag(pattern string) (string, error)

	// TagCommit returns the commit SHA that a tag points to.
	TagCommit(tagName string) (string, error)

	// TagDate returns the date of a tag.
	TagDate(tagName string) (time.Time, error)

	// TagExists checks if a tag with the given name exists.
	TagExists(tagName string) (bool, error)

	// --- Branch comparison operations (for squash commit messages) ---

	// GetBranchCommits returns all commits from baseBranch..HEAD.
	// Returns commits in reverse chronological order (newest first).
	// Returns error if baseBranch doesn't exist or no commits ahead.
	GetBranchCommits(baseBranch string) ([]CommitInfo, error)

	// GetBranchDiff returns the cumulative diff from baseBranch...HEAD.
	// Uses three-dot notation to compare against merge-base.
	GetBranchDiff(baseBranch string) (string, error)

	// GetBranchDiffStats returns diff statistics from baseBranch...HEAD.
	// Returns summary like "3 files changed, 45 insertions(+), 12 deletions(-)".
	GetBranchDiffStats(baseBranch string) (string, error)

	// GetBranchFiles returns list of files changed in baseBranch...HEAD.
	// Returns relative paths from repository root.
	GetBranchFiles(baseBranch string) ([]string, error)

	// --- Worktree operations (for workspace management) ---

	// WorktreeList returns information about all worktrees.
	WorktreeList() ([]WorktreeEntry, error)

	// WorktreeIsDirty checks if a worktree has uncommitted changes.
	WorktreeIsDirty(worktreePath string) (bool, error)
}

// WorktreeEntry represents a git worktree.
type WorktreeEntry struct {
	Path   string // Absolute path to worktree
	Branch string // Branch checked out in worktree
	SHA    string // Short SHA of HEAD commit
}

// Ensure Repository implements GitRepository.
var _ GitRepository = (*Repository)(nil)
