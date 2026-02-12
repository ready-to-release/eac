package git

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/zap"
)

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
