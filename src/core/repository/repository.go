package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/core/git"
)

// Repository represents a Git repository
type Repository struct {
	root    string
	gitRepo git.GitRepository
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
//
//	root, err := repository.GetRepositoryRoot("")
//	root, err := repository.GetRepositoryRoot("/path/to/subdir")
func GetRepositoryRoot(startPath string) (string, error) {
	// Check for Docker R2R mode - repository is mounted at /var/task
	if os.Getenv("DOCKER_R2R_MODE") == "true" {
		// In R2R CLI Docker mode, the repository is always at /var/task
		return "/var/task", nil
	}

	// Check for repository root override environment variable
	// Used by CLI wrapper and tests to specify the repository root
	if repoRoot := os.Getenv("R2R_REPO_ROOT"); repoRoot != "" {
		return filepath.Clean(repoRoot), nil
	}

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

	// Try using go-git to find the repository root
	repo, err := git.Open(absPath)
	if err == nil {
		root := repo.RootPath()
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
// Parameters:
//   - repo: GitRepository interface for git operations
//   - trackedOnly: if true, only return files tracked by Git
//   - includeIgnored: if true, include files ignored by .gitignore
//   - includeGitInternalFiles: if true, include .gitignore and .gitkeep files (default: false)
//   - stagedOnly: if true, only return files currently staged in Git index
//
// Example:
//
//	repo, _ := git.Open("/path/to/repo")
//	files, err := repository.GetRepositoryFiles(repo, true, false, false, false)
func GetRepositoryFiles(repo git.GitRepository, trackedOnly, includeIgnored, includeGitInternalFiles, stagedOnly bool) ([]FileInfo, error) {
	rootPath := repo.RootPath()
	var files []FileInfo

	// If stagedOnly is true, get only staged files
	if stagedOnly {
		stagedFiles, err := repo.StagedFiles()
		if err != nil {
			return nil, NewRepositoryError("staged files", rootPath, err, "failed to list staged files")
		}

		for _, line := range stagedFiles {
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
		// Get tracked files using the provided repo
		trackedFiles, err := repo.TrackedFiles()
		if err != nil {
			return nil, NewRepositoryError("tracked files", rootPath, err, "failed to list tracked files")
		}

		for _, line := range trackedFiles {
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

			// Check if file is tracked (using the provided repo)
			isTracked := repo.IsFileTracked(relPath)

			// Check if file is ignored
			isIgnored := false
			if !isTracked {
				isIgnored = repo.IsFileIgnored(relPath)
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

	gitRepo, err := git.Open(root)
	if err != nil {
		return nil, NewRepositoryError("open", root, err, "failed to open git repository")
	}

	return &Repository{
		root:    root,
		gitRepo: gitRepo,
	}, nil
}

// Root returns the repository root path
func (r *Repository) Root() string {
	return r.root
}

// GitRepo returns the underlying GitRepository interface
func (r *Repository) GitRepo() git.GitRepository {
	return r.gitRepo
}

// Files returns all files in the repository with the given options
func (r *Repository) Files(trackedOnly bool, includeIgnored bool) ([]FileInfo, error) {
	return GetRepositoryFiles(r.gitRepo, trackedOnly, includeIgnored, false, false)
}

// IsGitRepository checks if the given path is within a git repository
func IsGitRepository(path string) bool {
	_, err := GetRepositoryRoot(path)
	return err == nil
}
