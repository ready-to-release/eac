package git

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"go.uber.org/zap"
)

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
