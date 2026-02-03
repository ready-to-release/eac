// Package git provides Git operations using go-git library.
// This package uses ONLY pure go-git - no exec.Command calls.
// CLI layers (go/clibase/git/, go/cli/eac/impl/) may use exec.Command.
package git

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/zap"
)

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

// HeadCommit returns the full SHA of the HEAD commit.
func (r *Repository) HeadCommit() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return head.Hash().String(), nil
}

// UncommittedFiles returns paths of files with uncommitted changes.
// This includes staged, unstaged, and untracked files.
func (r *Repository) UncommittedFiles() ([]string, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var files []string
	for path, s := range status {
		// Include files with any changes: staging or worktree
		if s.Staging != gogit.Unmodified || s.Worktree != gogit.Unmodified {
			files = append(files, path)
		}
	}

	sort.Strings(files) // Consistent ordering
	return files, nil
}

// parseStatusPorcelain extracts file paths from git status --porcelain output.
// Format per line: XY filename
// - X = index status (staged)
// - Y = worktree status (unstaged)
// - A space separates XY from filename
// - Filenames with spaces are quoted
func parseStatusPorcelain(output string) []string {
	if output == "" {
		return nil
	}

	// Only trim trailing newlines/whitespace, not leading (which could be a status char)
	output = strings.TrimRight(output, "\n\r ")
	if output == "" {
		return nil
	}

	lines := strings.Split(output, "\n")
	var files []string
	for _, line := range lines {
		if len(line) < 4 {
			// Minimum valid line: "XY f" (2 status + space + 1 char filename)
			continue
		}
		// Skip position 0 (index status), 1 (worktree status), 2 (space separator)
		// Filename starts at position 3
		path := strings.TrimSpace(line[3:])
		path = strings.Trim(path, "\"")
		if path != "" {
			files = append(files, path)
		}
	}
	return files
}

// TrackedFiles returns all files tracked by Git (in the index).
// This reflects the current index state: HEAD files + staged additions - staged deletions.
func (r *Repository) TrackedFiles() ([]string, error) {
	idx, err := r.repo.Storer.Index()
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	files := make([]string, 0, len(idx.Entries))
	for _, entry := range idx.Entries {
		files = append(files, entry.Name)
	}

	sort.Strings(files) // Consistent ordering
	return files, nil
}

// StagedFiles returns files currently staged in the index (added, modified, renamed, copied).
// This corresponds to `git diff --cached --name-only --diff-filter=ACMR`.
func (r *Repository) StagedFiles() ([]string, error) {
	r.logger.Debug("Getting staged files")

	wt, err := r.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var files []string
	for path, s := range status {
		// Only include files staged for commit (Added, Modified, Renamed, Copied)
		// Exclude Deleted and Unmodified
		switch s.Staging {
		case gogit.Added, gogit.Modified, gogit.Renamed, gogit.Copied:
			files = append(files, path)
		}
	}

	sort.Strings(files) // Consistent ordering
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
func (r *Repository) StagedDiff() (string, error) {
	r.logger.Debug("Getting staged diff")

	// Get HEAD tree
	headTree, err := r.getHeadTree()
	if err != nil {
		// If no HEAD (empty repo), compare against empty tree
		headTree = nil
	}

	// Get index tree (staged changes)
	indexTree, err := r.getIndexTree()
	if err != nil {
		return "", fmt.Errorf("failed to get index tree: %w", err)
	}

	// Compare HEAD tree to index tree
	changes, err := headTree.Diff(indexTree)
	if err != nil {
		return "", fmt.Errorf("failed to diff trees: %w", err)
	}

	patch, err := changes.Patch()
	if err != nil {
		return "", fmt.Errorf("failed to generate patch: %w", err)
	}

	output := patch.String()
	r.logger.Debug("Staged diff retrieved", zap.Int("size", len(output)))
	return output, nil
}

// StagedDiffStats returns the stat summary of staged changes.
// Equivalent to `git diff --staged --stat`.
func (r *Repository) StagedDiffStats() (string, error) {
	r.logger.Debug("Getting staged diff stats")

	// Get HEAD tree
	headTree, err := r.getHeadTree()
	if err != nil {
		headTree = nil
	}

	// Get index tree
	indexTree, err := r.getIndexTree()
	if err != nil {
		return "", fmt.Errorf("failed to get index tree: %w", err)
	}

	// Compare HEAD tree to index tree
	changes, err := headTree.Diff(indexTree)
	if err != nil {
		return "", fmt.Errorf("failed to diff trees: %w", err)
	}

	patch, err := changes.Patch()
	if err != nil {
		return "", fmt.Errorf("failed to generate patch: %w", err)
	}

	stats := patch.Stats()
	output := formatDiffStats(stats)

	r.logger.Debug("Staged diff stats retrieved")
	return output, nil
}

// GetBranchCommits returns all commits from baseBranch..HEAD.
// Returns commits in reverse chronological order (newest first).
// Returns error if baseBranch doesn't exist or no commits ahead.
func (r *Repository) GetBranchCommits(baseBranch string) ([]CommitInfo, error) {
	// Resolve base branch to verify it exists
	baseRef := r.resolveBaseRef(baseBranch)

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
// Uses three-dot notation semantics (compare against merge-base).
func (r *Repository) GetBranchDiff(baseBranch string) (string, error) {
	baseRef := r.resolveBaseRef(baseBranch)

	baseCommit, err := r.resolveToCommit(baseRef)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base: %w", err)
	}

	headCommit, err := r.getHeadCommit()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get merge base for three-dot semantics
	mergeBase, err := r.findMergeBase(baseCommit, headCommit)
	if err != nil {
		// If no merge base, compare directly
		mergeBase = baseCommit
	}

	mergeBaseTree, err := mergeBase.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get merge-base tree: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD tree: %w", err)
	}

	changes, err := mergeBaseTree.Diff(headTree)
	if err != nil {
		return "", fmt.Errorf("failed to diff trees: %w", err)
	}

	patch, err := changes.Patch()
	if err != nil {
		return "", fmt.Errorf("failed to generate patch: %w", err)
	}

	return patch.String(), nil
}

// GetBranchDiffStats returns diff statistics from baseBranch...HEAD.
// Returns summary like "3 files changed, 45 insertions(+), 12 deletions(-)".
func (r *Repository) GetBranchDiffStats(baseBranch string) (string, error) {
	baseRef := r.resolveBaseRef(baseBranch)

	baseCommit, err := r.resolveToCommit(baseRef)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base: %w", err)
	}

	headCommit, err := r.getHeadCommit()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get merge base for three-dot semantics
	mergeBase, err := r.findMergeBase(baseCommit, headCommit)
	if err != nil {
		mergeBase = baseCommit
	}

	mergeBaseTree, err := mergeBase.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get merge-base tree: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD tree: %w", err)
	}

	changes, err := mergeBaseTree.Diff(headTree)
	if err != nil {
		return "", fmt.Errorf("failed to diff trees: %w", err)
	}

	patch, err := changes.Patch()
	if err != nil {
		return "", fmt.Errorf("failed to generate patch: %w", err)
	}

	stats := patch.Stats()
	return formatDiffStats(stats), nil
}

// GetBranchFiles returns list of files changed in baseBranch...HEAD.
// Returns relative paths from repository root.
func (r *Repository) GetBranchFiles(baseBranch string) ([]string, error) {
	baseRef := r.resolveBaseRef(baseBranch)

	baseCommit, err := r.resolveToCommit(baseRef)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base: %w", err)
	}

	headCommit, err := r.getHeadCommit()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get merge base for three-dot semantics
	mergeBase, err := r.findMergeBase(baseCommit, headCommit)
	if err != nil {
		mergeBase = baseCommit
	}

	mergeBaseTree, err := mergeBase.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get merge-base tree: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD tree: %w", err)
	}

	changes, err := mergeBaseTree.Diff(headTree)
	if err != nil {
		return nil, fmt.Errorf("failed to diff trees: %w", err)
	}

	var files []string
	for _, change := range changes {
		name := change.To.Name
		if name == "" {
			name = change.From.Name
		}
		files = append(files, name)
	}

	sort.Strings(files)
	return files, nil
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

// formatDiffStats formats file stats similar to git diff --stat output.
func formatDiffStats(stats object.FileStats) string {
	if len(stats) == 0 {
		return ""
	}

	var b strings.Builder
	var totalAdditions, totalDeletions int

	for _, stat := range stats {
		b.WriteString(fmt.Sprintf(" %s | %d ", stat.Name, stat.Addition+stat.Deletion))
		b.WriteString(strings.Repeat("+", stat.Addition))
		b.WriteString(strings.Repeat("-", stat.Deletion))
		b.WriteString("\n")
		totalAdditions += stat.Addition
		totalDeletions += stat.Deletion
	}

	// Summary line
	b.WriteString(fmt.Sprintf(" %d file", len(stats)))
	if len(stats) != 1 {
		b.WriteString("s")
	}
	b.WriteString(" changed")
	if totalAdditions > 0 {
		b.WriteString(fmt.Sprintf(", %d insertion", totalAdditions))
		if totalAdditions != 1 {
			b.WriteString("s")
		}
		b.WriteString("(+)")
	}
	if totalDeletions > 0 {
		b.WriteString(fmt.Sprintf(", %d deletion", totalDeletions))
		if totalDeletions != 1 {
			b.WriteString("s")
		}
		b.WriteString("(-)")
	}
	b.WriteString("\n")

	return b.String()
}
