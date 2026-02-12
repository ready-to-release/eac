// Package git provides Git operations using go-git library.
// This package uses ONLY pure go-git - no exec.Command calls.
// CLI layers (go/clibase/git/, go/cli/eac/impl/) may use exec.Command.
package git

import (
	"fmt"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/zap"
)

// Repository wraps a go-git repository with convenience methods.
// Should be created via RepositoryManager for proper logger injection.
type Repository struct {
	repo     *gogit.Repository
	rootPath string
	logger   *zap.Logger // Logger for observability (guaranteed non-nil when created via manager)
	clock    Clock       // Time source for commits (defaults to time.Now)
}

// now returns the current time using the repository's clock.
// Falls back to time.Now if no clock was injected.
func (r *Repository) now() time.Time {
	if r.clock != nil {
		return r.clock()
	}
	return time.Now()
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

// HeadCommit returns the full SHA of the HEAD commit.
func (r *Repository) HeadCommit() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return head.Hash().String(), nil
}

// GoGitRepo returns the underlying go-git repository for advanced operations.
// Use sparingly - prefer the wrapper methods when possible.
func (r *Repository) GoGitRepo() *gogit.Repository {
	return r.repo
}

// --- Helper methods for pure go-git operations ---

// resolveBaseRef resolves a base branch reference, trying origin/ prefix if needed.
func (r *Repository) resolveBaseRef(baseBranch string) string {
	if strings.HasPrefix(baseBranch, "origin/") || strings.HasPrefix(baseBranch, "refs/") {
		return baseBranch
	}

	// Try origin/baseBranch first (for remote tracking)
	originRef := "origin/" + baseBranch
	if _, err := r.resolveToCommit(originRef); err == nil {
		return originRef
	}

	return baseBranch
}

// resolveToCommit resolves a ref (branch/tag/sha) to a commit object.
func (r *Repository) resolveToCommit(ref string) (*object.Commit, error) {
	hash, err := r.repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", ref, err)
	}
	return r.repo.CommitObject(*hash)
}

// getHeadCommit returns the HEAD commit object.
func (r *Repository) getHeadCommit() (*object.Commit, error) {
	head, err := r.repo.Head()
	if err != nil {
		return nil, err
	}
	return r.repo.CommitObject(head.Hash())
}

// getHeadTree returns the tree for the HEAD commit.
func (r *Repository) getHeadTree() (*object.Tree, error) {
	commit, err := r.getHeadCommit()
	if err != nil {
		return nil, err
	}
	return commit.Tree()
}

// getIndexTree creates a pseudo-tree from the index for diff comparison.
// This is used for staged diff operations.
func (r *Repository) getIndexTree() (*object.Tree, error) {
	// For staged diff, we need to compare HEAD to the index
	// go-git doesn't have a direct "index as tree" API, so we use worktree status
	// to identify staged files and create the diff based on that

	// Get the HEAD tree as a baseline
	headTree, err := r.getHeadTree()
	if err != nil {
		// Empty repository - return nil tree
		return nil, nil
	}

	// For staged diff, we actually compare HEAD tree to itself
	// but filter based on staging status
	// This is a simplification - return HEAD tree and let the status-based
	// methods handle the actual staging detection
	return headTree, nil
}

// findMergeBase finds the merge base between two commits.
func (r *Repository) findMergeBase(c1, c2 *object.Commit) (*object.Commit, error) {
	// Simple implementation: walk c1's ancestors looking for c2's ancestors
	// This is a naive O(n*m) algorithm but works for most cases
	c2Ancestors := make(map[plumbing.Hash]bool)

	// Collect c2's ancestors
	iter := object.NewCommitIterCTime(c2, nil, nil)
	err := iter.ForEach(func(c *object.Commit) error {
		c2Ancestors[c.Hash] = true
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Find first c1 ancestor that's also c2 ancestor
	iter = object.NewCommitIterCTime(c1, nil, nil)
	var mergeBase *object.Commit
	err = iter.ForEach(func(c *object.Commit) error {
		if c2Ancestors[c.Hash] {
			mergeBase = c
			return fmt.Errorf("found") // Stop iteration
		}
		return nil
	})
	if mergeBase != nil {
		return mergeBase, nil
	}
	if err != nil && err.Error() != "found" {
		return nil, err
	}

	return nil, fmt.Errorf("no merge base found")
}
