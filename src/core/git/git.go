// Package git provides Git operations using go-git library.
// It replaces direct exec.Command("git", ...) calls with pure Go implementations.
package git

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Repository wraps a go-git repository with convenience methods.
type Repository struct {
	repo     *gogit.Repository
	rootPath string
}

// Open opens an existing Git repository at the given path.
// If path is empty, uses the current working directory.
// It searches upward through parent directories to find the repository root.
func Open(path string) (*Repository, error) {
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
	repo, err := gogit.PlainOpenWithOptions(absPath, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the worktree to find the root path
	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

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
func (r *Repository) TrackedFiles() ([]string, error) {
	head, err := r.repo.Head()
	if err != nil {
		// Empty repository - no tracked files
		if err == plumbing.ErrReferenceNotFound {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	var files []string
	err = tree.Files().ForEach(func(f *object.File) error {
		files = append(files, f.Name)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate files: %w", err)
	}

	return files, nil
}

// StagedFiles returns files currently staged in the index (added, modified, renamed).
// This corresponds to `git diff --cached --name-only --diff-filter=ACMR`.
func (r *Repository) StagedFiles() ([]string, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var files []string
	for path, fileStatus := range status {
		// Check staging status (index status)
		// A = Added, M = Modified, R = Renamed, C = Copied
		switch fileStatus.Staging {
		case gogit.Added, gogit.Modified, gogit.Renamed, gogit.Copied:
			files = append(files, path)
		}
	}

	return files, nil
}

// IsFileTracked checks if a specific file is tracked by Git.
func (r *Repository) IsFileTracked(relPath string) bool {
	// Normalize path separators
	relPath = filepath.ToSlash(relPath)

	head, err := r.repo.Head()
	if err != nil {
		return false
	}

	commit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return false
	}

	tree, err := commit.Tree()
	if err != nil {
		return false
	}

	_, err = tree.File(relPath)
	return err == nil
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
func (r *Repository) StagedDiff() (string, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	// Get HEAD commit tree (for comparison)
	var headTree *object.Tree
	head, err := r.repo.Head()
	if err == nil {
		commit, err := r.repo.CommitObject(head.Hash())
		if err == nil {
			headTree, _ = commit.Tree()
		}
	}
	// If no HEAD (empty repo), headTree will be nil - that's fine

	var diffBuilder strings.Builder

	for path, fileStatus := range status {
		// Only process staged changes
		switch fileStatus.Staging {
		case gogit.Added, gogit.Modified, gogit.Renamed, gogit.Copied, gogit.Deleted:
			// Get the staged content from the index
			idx, err := r.repo.Storer.Index()
			if err != nil {
				continue
			}

			// Find the entry in the index
			var stagedContent string
			var stagedBlob plumbing.Hash
			for _, entry := range idx.Entries {
				if entry.Name == path {
					stagedBlob = entry.Hash
					break
				}
			}

			if stagedBlob != plumbing.ZeroHash {
				blob, err := r.repo.BlobObject(stagedBlob)
				if err == nil {
					reader, err := blob.Reader()
					if err == nil {
						content, _ := io.ReadAll(reader)
						stagedContent = string(content)
						reader.Close()
					}
				}
			}

			// Get the original content from HEAD (if exists)
			var originalContent string
			if headTree != nil && fileStatus.Staging != gogit.Added {
				file, err := headTree.File(path)
				if err == nil {
					originalContent, _ = file.Contents()
				}
			}

			// Generate unified diff
			diffOutput := generateUnifiedDiff(path, originalContent, stagedContent, fileStatus.Staging)
			diffBuilder.WriteString(diffOutput)
		}
	}

	return diffBuilder.String(), nil
}

// StagedDiffStats returns the stat summary of staged changes.
// Equivalent to `git diff --staged --stat`.
func (r *Repository) StagedDiffStats() (string, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	// Get HEAD commit tree (for comparison)
	var headTree *object.Tree
	head, err := r.repo.Head()
	if err == nil {
		commit, err := r.repo.CommitObject(head.Hash())
		if err == nil {
			headTree, _ = commit.Tree()
		}
	}

	var stats []string
	var totalInsertions, totalDeletions int
	filesChanged := 0

	for path, fileStatus := range status {
		// Only process staged changes
		switch fileStatus.Staging {
		case gogit.Added, gogit.Modified, gogit.Renamed, gogit.Copied, gogit.Deleted:
			filesChanged++

			// Get the staged content from the index
			idx, err := r.repo.Storer.Index()
			if err != nil {
				continue
			}

			var stagedContent string
			var stagedBlob plumbing.Hash
			for _, entry := range idx.Entries {
				if entry.Name == path {
					stagedBlob = entry.Hash
					break
				}
			}

			if stagedBlob != plumbing.ZeroHash {
				blob, err := r.repo.BlobObject(stagedBlob)
				if err == nil {
					reader, err := blob.Reader()
					if err == nil {
						content, _ := io.ReadAll(reader)
						stagedContent = string(content)
						reader.Close()
					}
				}
			}

			// Get the original content from HEAD (if exists)
			var originalContent string
			if headTree != nil && fileStatus.Staging != gogit.Added {
				file, err := headTree.File(path)
				if err == nil {
					originalContent, _ = file.Contents()
				}
			}

			// Calculate insertions/deletions
			originalLines := strings.Split(originalContent, "\n")
			stagedLines := strings.Split(stagedContent, "\n")

			if originalContent == "" {
				originalLines = []string{}
			}
			if stagedContent == "" {
				stagedLines = []string{}
			}

			insertions := 0
			deletions := 0

			if fileStatus.Staging == gogit.Added {
				insertions = len(stagedLines)
			} else if fileStatus.Staging == gogit.Deleted {
				deletions = len(originalLines)
			} else {
				// Simple line count diff (not a true diff algorithm, but good for stats)
				insertions = max(0, len(stagedLines)-len(originalLines))
				deletions = max(0, len(originalLines)-len(stagedLines))
				// For modified files, count actual changes
				if insertions == 0 && deletions == 0 && originalContent != stagedContent {
					// Files differ but same line count - count as modifications
					insertions = 1
					deletions = 1
				}
			}

			totalInsertions += insertions
			totalDeletions += deletions

			// Format: " path | changes +++ ---"
			changeIndicator := ""
			if insertions > 0 {
				changeIndicator += strings.Repeat("+", min(insertions, 50))
			}
			if deletions > 0 {
				changeIndicator += strings.Repeat("-", min(deletions, 50))
			}

			stats = append(stats, fmt.Sprintf(" %s | %d %s", path, insertions+deletions, changeIndicator))
		}
	}

	if len(stats) == 0 {
		return "", nil
	}

	// Build the final stats output
	var result strings.Builder
	for _, stat := range stats {
		result.WriteString(stat)
		result.WriteString("\n")
	}
	result.WriteString(fmt.Sprintf(" %d file(s) changed, %d insertion(s)(+), %d deletion(s)(-)\n",
		filesChanged, totalInsertions, totalDeletions))

	return result.String(), nil
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

// GoGitRepo returns the underlying go-git repository for advanced operations.
// Use sparingly - prefer the wrapper methods when possible.
func (r *Repository) GoGitRepo() *gogit.Repository {
	return r.repo
}
