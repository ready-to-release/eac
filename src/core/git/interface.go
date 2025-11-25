package git

import gogit "github.com/go-git/go-git/v5"

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

	// ConfigSet sets a Git configuration value.
	ConfigSet(section, key, value string) error

	// GoGitRepo returns the underlying go-git repository for advanced operations.
	// Returns nil for mock implementations.
	GoGitRepo() *gogit.Repository
}

// Ensure Repository implements GitRepository
var _ GitRepository = (*Repository)(nil)
