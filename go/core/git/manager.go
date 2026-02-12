package git

import (
	"fmt"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"go.uber.org/zap"
)

// RepositoryManager manages git repository instances with shared logging configuration.
// It implements constructor-based dependency injection where the logger is provided once
// during manager creation and shared across all repository instances.
//
// Usage:
//
//	// Create manager once with logger
//	gitMgr := git.NewManager(logging.C().Zap())
//
//	// Open repositories as needed - all share the manager's logger
//	repo, err := gitMgr.Open(workspaceRoot)
//	files, _ := repo.StagedFiles()
//	diff, _ := repo.StagedDiff()
//
//	// Can open multiple repos with same logger
//	repo2, err := gitMgr.Open(otherPath)
type RepositoryManager struct {
	logger *zap.Logger
	clock  Clock // Optional clock override; nil means time.Now
}

// NewManager creates a new RepositoryManager with the given logger.
// If logger is nil, a no-op logger is used (no logging output).
//
// This follows constructor-based dependency injection:
// - Logger is provided once during manager creation
// - All repositories created by this manager share the same logger
// - No nil checks needed in repository methods
//
// Example:
//
//	// With logging (commands layer)
//	mgr := git.NewManager(logging.C().Zap())
//
//	// Without logging (core-to-core calls)
//	mgr := git.NewManager(nil)
func NewManager(logger *zap.Logger) *RepositoryManager {
	if logger == nil {
		logger = zap.NewNop() // Default to no-op logger
	}
	return &RepositoryManager{
		logger: logger,
	}
}

// WithClock returns a copy of the manager that injects the given clock
// into every repository it creates. This is intended for tests that need
// deterministic timestamps without mutable package-level state.
//
// Example:
//
//	fixed := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
//	mgr := git.NewManager(nil).WithClock(func() time.Time { return fixed })
//	repo, _ := mgr.Init(tmpDir)
//	// repo.Commit(...) will use the fixed time
func (m *RepositoryManager) WithClock(c Clock) *RepositoryManager {
	return &RepositoryManager{
		logger: m.logger,
		clock:  c,
	}
}

// Open opens an existing Git repository at the given path.
// If path is empty, uses the current working directory.
// It searches upward through parent directories to find the repository root.
//
// The returned repository shares this manager's logger configuration.
//
// Example:
//
//	mgr := git.NewManager(logger)
//	repo, err := mgr.Open("/path/to/workspace")
//	if err != nil {
//	    return fmt.Errorf("failed to open repository: %w", err)
//	}
func (m *RepositoryManager) Open(path string) (*Repository, error) {
	m.logger.Debug("Opening git repository", zap.String("path", path))

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

	m.logger.Debug("Git repository opened successfully", zap.String("rootPath", wt.Filesystem.Root()))

	return &Repository{
		repo:     repo,
		rootPath: wt.Filesystem.Root(),
		logger:   m.logger, // Share manager's logger
		clock:    m.clock,  // Share manager's clock (nil means time.Now)
	}, nil
}

// Init initializes a new Git repository at the given path.
// If path is empty, uses the current working directory.
//
// The returned repository shares this manager's logger configuration.
//
// Example:
//
//	mgr := git.NewManager(logger)
//	repo, err := mgr.Init("/path/to/new/repo")
//	if err != nil {
//	    return fmt.Errorf("failed to init repository: %w", err)
//	}
func (m *RepositoryManager) Init(path string) (*Repository, error) {
	m.logger.Debug("Initializing git repository", zap.String("path", path))

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

	// Initialize repository
	repo, err := gogit.PlainInit(absPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	m.logger.Debug("Git repository initialized successfully", zap.String("path", absPath))

	return &Repository{
		repo:     repo,
		rootPath: absPath,
		logger:   m.logger, // Share manager's logger
		clock:    m.clock,  // Share manager's clock (nil means time.Now)
	}, nil
}
