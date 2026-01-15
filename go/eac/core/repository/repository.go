package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"go.uber.org/zap"
)

// Repository represents a Git repository
type Repository struct {
	root    string
	gitRepo git.GitRepository
	logger  *zap.Logger // Optional logger for observability (nil = no logging)
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
	// Check for explicit workspace root override first
	// This takes precedence over R2R_DOCKER_MODE to allow test isolation
	// Used by CLI wrapper and tests to specify the repository root
	if repoRoot := GetWorkspaceRoot(); repoRoot != "" {
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

	// Search for .git directory by walking up the tree
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
			break
		}
		currentPath = parentPath
	}

	// If in Docker mode and no git root found, fall back to /var/task
	// This handles the case where we're in a subdirectory of the mounted repo
	if os.Getenv("R2R_DOCKER_MODE") == "true" {
		return "/var/task", nil
	}

	// Not in Docker mode and no git root found - error
	return "", NewRepositoryError("find", absPath, nil, "not a git repository (or any parent up to mount point)")
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
// Pass logger for observability (nil = no logging).
func New(path string, logger *zap.Logger) (*Repository, error) {
	if logger != nil {
		logger.Debug("Creating repository instance", zap.String("path", path))
	}

	root, err := GetRepositoryRoot(path)
	if err != nil {
		return nil, err
	}

	gitRepo, err := git.Open(root, logger)
	if err != nil {
		return nil, NewRepositoryError("open", root, err, "failed to open git repository")
	}

	if logger != nil {
		logger.Debug("Repository instance created", zap.String("root", root))
	}

	return &Repository{
		root:    root,
		gitRepo: gitRepo,
		logger:  logger, // ✅ Store logger
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

// GetContainerRoot returns the container's internal root directory if running
// inside a container (R2R_CONTAINER_ROOT is set), otherwise returns empty string.
//
// This is used to locate container-internal files (pre-built binaries, test assets)
// which may differ from the mounted repository root in container environments.
func GetContainerRoot() string {
	return os.Getenv("R2R_CONTAINER_ROOT")
}

// GetDistRoot returns the distribution root where tool assets live (contracts,
// schemas, defaults). In container mode, this is R2R_CONTAINER_ROOT. Otherwise
// falls back to the provided fallback (typically the repo root).
//
// Use this for loading:
//   - Contract schemas (contracts/eac-core/*/schemas/)
//   - Default configs (contracts/eac-core/*/defaults/)
//   - Tool definitions and templates
//
// These are part of the TOOL distribution, not the user's workspace.
func GetDistRoot(fallback string) string {
	if containerRoot := GetContainerRoot(); containerRoot != "" {
		return containerRoot
	}
	return fallback
}

// GetWorkspaceRoot returns the workspace root override (R2R_REPO_ROOT) if set,
// otherwise returns empty string. The caller should fall back to git-based
// discovery if this returns empty.
//
// Use this for loading:
//   - User configs (.r2r/eac/*.yml)
//   - Project source files
//   - Build outputs
//
// These are part of the USER's workspace, not the tool distribution.
func GetWorkspaceRoot() string {
	return os.Getenv("R2R_REPO_ROOT")
}

// GetRepoEACConfigRoot returns the path to the EAC repository configuration directory.
// This is where repository-specific EAC contracts are stored (modules, tags, environments).
// The relative path is defined by EACConfigRelPath constant in paths.go.
//
// This is distinct from tool contracts (AI prompts, CLI schema) which remain in contracts/.
func GetRepoEACConfigRoot(startPath string) (string, error) {
	repoRoot, err := GetRepositoryRoot(startPath)
	if err != nil {
		return "", err
	}
	return paths.EACConfigPath(repoRoot), nil
}
