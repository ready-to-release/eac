package repository

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository represents a Git repository
type Repository struct {
	root string
}

// RepositoryError represents a repository-related error
type RepositoryError struct {
	Op      string // Operation that failed
	Path    string // Path related to the error
	Err     error  // Underlying error
	Message string // Additional context
}

func (e *RepositoryError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("repository %s failed for %s: %s", e.Op, e.Path, e.Message)
	}
	return fmt.Sprintf("repository %s failed: %s", e.Op, e.Message)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewRepositoryError creates a new RepositoryError
func NewRepositoryError(op, path string, err error, message string) *RepositoryError {
	return &RepositoryError{
		Op:      op,
		Path:    path,
		Err:     err,
		Message: message,
	}
}

// GetRepositoryRoot finds and returns the root directory of the Git repository
// starting from the given path (or current directory if empty).
//
// It searches upward through parent directories until it finds a .git directory,
// or returns an error if no repository is found.
//
// Example:
//   root, err := repository.GetRepositoryRoot("")
//   root, err := repository.GetRepositoryRoot("/path/to/subdir")
func GetRepositoryRoot(startPath string) (string, error) {
	// Use current directory if no path provided
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return "", NewRepositoryError("getwd", "", err, "failed to get current directory")
		}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", NewRepositoryError("abs", startPath, err, "failed to get absolute path")
	}

	// Try using git to find the root (most reliable)
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = absPath
	output, err := cmd.Output()
	if err == nil {
		root := strings.TrimSpace(string(output))
		// Normalize path separators for Windows
		root = filepath.Clean(root)
		return root, nil
	}

	// Fallback: manually search for .git directory
	currentPath := absPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			// Found .git - check if it's a directory or file (submodule/worktree)
			if info.IsDir() || info.Mode().IsRegular() {
				return currentPath, nil
			}
		}

		// Move to parent directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached filesystem root without finding .git
			return "", NewRepositoryError("find", absPath, nil, "not a git repository (or any parent up to mount point)")
		}
		currentPath = parentPath
	}
}

// FileInfo represents information about a repository file
type FileInfo struct {
	Path         string // Relative path from repository root
	AbsolutePath string // Absolute filesystem path
	IsTracked    bool   // Whether the file is tracked by git
	IsIgnored    bool   // Whether the file is ignored by .gitignore
}

// RepositoryFileWithModule represents a file with its owning module(s)
type RepositoryFileWithModule struct {
	Name    string   // File path relative to repo root with forward slashes
	Modules []string // Module monikers that own this file (can be multiple)
}

// GetRepositoryFiles returns a list of all files in the repository.
//
// Options:
//   - trackedOnly: if true, only return files tracked by Git
//   - includeIgnored: if true, include files ignored by .gitignore
//   - includeGitInternalFiles: if true, include .gitignore and .gitkeep files (default: false)
//   - stagedOnly: if true, only return files currently staged in Git index (added, removed, or changed)
//   - rootPath: repository root (if empty, will be detected automatically)
//
// Example:
//   files, err := repository.GetRepositoryFiles(true, false, false, false, "")
func GetRepositoryFiles(trackedOnly bool, includeIgnored bool, includeGitInternalFiles bool, stagedOnly bool, rootPath string) ([]FileInfo, error) {
	// Get repository root if not provided
	if rootPath == "" {
		var err error
		rootPath, err = GetRepositoryRoot("")
		if err != nil {
			return nil, err
		}
	}

	var files []FileInfo

	// If stagedOnly is true, get only staged files
	if stagedOnly {
		cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACMR")
		cmd.Dir = rootPath
		output, err := cmd.Output()
		if err != nil {
			return nil, NewRepositoryError("diff --cached", rootPath, err, "failed to list staged files")
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Filter out Git internal files unless explicitly included
			if !includeGitInternalFiles && isGitInternalFile(line) {
				continue
			}

			absPath := filepath.Join(rootPath, line)
			files = append(files, FileInfo{
				Path:         line,
				AbsolutePath: absPath,
				IsTracked:    true,
				IsIgnored:    false,
			})
		}
		return files, nil
	}

	if trackedOnly {
		// Get tracked files from git
		cmd := exec.Command("git", "ls-files")
		cmd.Dir = rootPath
		output, err := cmd.Output()
		if err != nil {
			return nil, NewRepositoryError("ls-files", rootPath, err, "failed to list tracked files")
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Filter out Git internal files unless explicitly included
			if !includeGitInternalFiles && isGitInternalFile(line) {
				continue
			}

			absPath := filepath.Join(rootPath, line)
			files = append(files, FileInfo{
				Path:         line,
				AbsolutePath: absPath,
				IsTracked:    true,
				IsIgnored:    false,
			})
		}
	} else {
		// Walk all files in repository
		err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				// Skip .git directory
				if info.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}

			// Get relative path
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}

			// Filter out Git internal files unless explicitly included
			if !includeGitInternalFiles && isGitInternalFile(relPath) {
				return nil
			}

			// Check if file is tracked
			isTracked := isFileTracked(rootPath, relPath)

			// Check if file is ignored
			isIgnored := false
			if !isTracked {
				isIgnored = isFileIgnored(rootPath, relPath)
			}

			// Skip ignored files if not requested
			if isIgnored && !includeIgnored {
				return nil
			}

			files = append(files, FileInfo{
				Path:         relPath,
				AbsolutePath: path,
				IsTracked:    isTracked,
				IsIgnored:    isIgnored,
			})

			return nil
		})

		if err != nil {
			return nil, NewRepositoryError("walk", rootPath, err, "failed to walk repository files")
		}
	}

	return files, nil
}

// isFileTracked checks if a file is tracked by git
func isFileTracked(repoRoot, relPath string) bool {
	cmd := exec.Command("git", "ls-files", "--", relPath)
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

// isFileIgnored checks if a file is ignored by .gitignore
func isFileIgnored(repoRoot, relPath string) bool {
	cmd := exec.Command("git", "check-ignore", "-q", relPath)
	cmd.Dir = repoRoot
	err := cmd.Run()
	// Exit code 0 means file is ignored
	return err == nil
}

// isGitInternalFile checks if a file is a Git internal file (.gitignore, .gitkeep)
// that should be filtered out from repository operations
func isGitInternalFile(relPath string) bool {
	basename := filepath.Base(relPath)
	return basename == ".gitignore" || basename == ".gitkeep"
}

// New creates a Repository instance from a given path
// If path is empty, uses current directory
func New(path string) (*Repository, error) {
	root, err := GetRepositoryRoot(path)
	if err != nil {
		return nil, err
	}

	return &Repository{
		root: root,
	}, nil
}

// Root returns the repository root path
func (r *Repository) Root() string {
	return r.root
}

// Files returns all files in the repository with the given options
func (r *Repository) Files(trackedOnly bool, includeIgnored bool) ([]FileInfo, error) {
	return GetRepositoryFiles(trackedOnly, includeIgnored, false, false, r.root)
}

// GetRepositoryFilesDefault is a convenience function that calls GetRepositoryFiles
// with includeGitInternalFiles and stagedOnly defaulted to false
func GetRepositoryFilesDefault(trackedOnly bool, includeIgnored bool, rootPath string) ([]FileInfo, error) {
	return GetRepositoryFiles(trackedOnly, includeIgnored, false, false, rootPath)
}

// IsGitRepository checks if the given path is within a git repository
func IsGitRepository(path string) bool {
	_, err := GetRepositoryRoot(path)
	return err == nil
}
