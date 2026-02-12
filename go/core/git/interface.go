package git

import (
	"time"

	gogit "github.com/go-git/go-git/v5"
)

// GitReader provides read-only repository state queries.
// Use this sub-interface when a consumer only needs to inspect repository state
// without modifying it (e.g., change detection, file listing, status checks).
type GitReader interface {
	// RootPath returns the repository root directory path.
	RootPath() string

	// RemoteURL returns the URL for the named remote (typically "origin").
	RemoteURL(remoteName string) (string, error)

	// CurrentBranch returns the current branch name.
	// If in detached HEAD state, returns "detached-<short-sha>".
	CurrentBranch() (string, error)

	// HeadShortSHA returns the short SHA of the HEAD commit.
	HeadShortSHA() (string, error)

	// HeadCommit returns the full SHA of the HEAD commit.
	HeadCommit() (string, error)

	// UncommittedFiles returns paths of files with uncommitted changes.
	// Uses git status --porcelain format parsing.
	UncommittedFiles() ([]string, error)

	// TrackedFiles returns all files tracked by Git (in the index).
	TrackedFiles() ([]string, error)

	// StagedFiles returns files currently staged in the index.
	StagedFiles() ([]string, error)

	// StagedDiff returns the unified diff of all staged changes.
	// Equivalent to `git diff --staged`.
	StagedDiff() (string, error)

	// StagedDiffStats returns the stat summary of staged changes.
	// Equivalent to `git diff --staged --stat`.
	StagedDiffStats() (string, error)

	// IsFileTracked checks if a specific file is tracked by Git.
	IsFileTracked(relPath string) bool

	// IsFileIgnored checks if a file matches .gitignore patterns.
	IsFileIgnored(relPath string) bool

	// GoGitRepo returns the underlying go-git repository for advanced operations.
	// Returns nil for mock implementations.
	GoGitRepo() *gogit.Repository
}

// GitWriter provides repository mutation operations.
// Use this sub-interface when a consumer needs to stage files, create commits,
// or modify repository configuration.
type GitWriter interface {
	// Add stages a file for commit.
	Add(path string) error

	// Commit creates a new commit with the staged changes.
	// Returns the commit hash.
	Commit(message, authorName, authorEmail string) (string, error)

	// AddRemote adds a new remote to the repository.
	AddRemote(name, url string) error

	// ConfigSet sets a Git configuration value.
	ConfigSet(section, key, value string) error
}

// GitReleaseQuerier provides changelog and tag operations.
// Use this sub-interface for release tooling that needs to query commit history
// and tag information (e.g., changelog generation, version resolution).
type GitReleaseQuerier interface {
	// CommitsBetween returns commits between two references (tag/SHA/branch).
	// If fromRef is empty, returns all commits up to toRef.
	// Returns commits in reverse chronological order (newest first).
	CommitsBetween(fromRef, toRef string) ([]CommitInfo, error)

	// CommitsSince returns all commits since a given tag or reference.
	CommitsSince(ref string) ([]CommitInfo, error)

	// TagsMatching returns tags matching a glob pattern (e.g., "clie/*").
	TagsMatching(pattern string) ([]string, error)

	// LatestTag returns the most recent tag matching the pattern.
	LatestTag(pattern string) (string, error)

	// TagCommit returns the commit SHA that a tag points to.
	TagCommit(tagName string) (string, error)

	// TagDate returns the date of a tag.
	TagDate(tagName string) (time.Time, error)

	// TagExists checks if a tag with the given name exists.
	TagExists(tagName string) (bool, error)
}

// GitBranchComparer provides branch comparison operations.
// Use this sub-interface for tools that compare branches, such as
// squash commit message generators and PR diff viewers.
type GitBranchComparer interface {
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
}

// GitRepository is the full repository interface combining all capabilities.
// This allows mocking in tests and supports alternative implementations.
//
// Consumers should prefer accepting the narrowest sub-interface that satisfies
// their needs (GitReader, GitWriter, GitReleaseQuerier, or GitBranchComparer)
// rather than requiring the full GitRepository. This follows the Interface
// Segregation Principle and makes dependencies explicit.
type GitRepository interface {
	GitReader
	GitWriter
	GitReleaseQuerier
	GitBranchComparer
}

// Compile-time interface satisfaction checks.
var (
	_ GitRepository = (*Repository)(nil)
	_ GitRepository = (*MockRepository)(nil)

	// Sub-interface satisfaction for Repository.
	_ GitReader          = (*Repository)(nil)
	_ GitWriter          = (*Repository)(nil)
	_ GitReleaseQuerier  = (*Repository)(nil)
	_ GitBranchComparer  = (*Repository)(nil)

	// Sub-interface satisfaction for MockRepository.
	_ GitReader          = (*MockRepository)(nil)
	_ GitWriter          = (*MockRepository)(nil)
	_ GitReleaseQuerier  = (*MockRepository)(nil)
	_ GitBranchComparer  = (*MockRepository)(nil)
)
