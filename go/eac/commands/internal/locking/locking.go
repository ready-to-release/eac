// Package locking provides file-based locking for build and test commands.
// It prevents concurrent execution of the same module build or test suite.
package locking

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/google/uuid"
	"github.com/ready-to-release/eac/go/eac/commands/internal/locktracker"
)

// Config configures lock acquisition.
type Config struct {
	BaseDir      string // Directory for lock files (e.g., "out/build", "out/test")
	Identifier   string // Lock identifier (module moniker or suite name)
	ResourceType string // For error messages: "module" or "test suite"
	ActionVerb   string // For error messages: "already being built" or "already running"
}

// Acquire creates and acquires an exclusive lock.
// Returns the lock handle on success, or an error if the lock cannot be acquired.
func Acquire(workspaceRoot string, cfg Config) (*flock.Flock, error) {
	lockDir := filepath.Join(workspaceRoot, cfg.BaseDir)

	// Ensure directory exists with proper permissions
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory %s: %w", lockDir, err)
	}

	lockPath := filepath.Join(lockDir, fmt.Sprintf(".lock-%s", cfg.Identifier))
	lock := flock.New(lockPath)

	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock at %s: %w", lockPath, err)
	}
	if !locked {
		return nil, fmt.Errorf("%s '%s' is %s", cfg.ResourceType, cfg.Identifier, cfg.ActionVerb)
	}

	return lock, nil
}

// Release releases the lock and removes the lock file.
func Release(lock *flock.Flock) {
	if lock == nil {
		return
	}
	lockPath := lock.Path()
	//nolint:errcheck // best-effort cleanup
	lock.Unlock()
	os.Remove(lockPath) // best-effort cleanup
}

// TrackedLock wraps a flock.Flock with lock tracking.
type TrackedLock struct {
	*flock.Flock
	id       string
	registry *locktracker.Registry
}

// AcquireTracked creates and acquires an exclusive lock with tracking.
// The lock is registered with the provided registry for visualization.
// Returns the tracked lock handle on success, or an error if the lock cannot be acquired.
func AcquireTracked(workspaceRoot string, cfg Config, registry *locktracker.Registry) (*TrackedLock, error) {
	lockDir := filepath.Join(workspaceRoot, cfg.BaseDir)

	// Ensure directory exists with proper permissions
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory %s: %w", lockDir, err)
	}

	lockPath := filepath.Join(lockDir, fmt.Sprintf(".lock-%s", cfg.Identifier))
	lock := flock.New(lockPath)

	id := uuid.New().String()
	lockName := fmt.Sprintf("%s:%s", cfg.ResourceType, cfg.Identifier)

	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock at %s: %w", lockPath, err)
	}
	if !locked {
		return nil, fmt.Errorf("%s '%s' is %s", cfg.ResourceType, cfg.Identifier, cfg.ActionVerb)
	}

	// Register with the lock tracker
	registry.Register(locktracker.LockInfo{
		ID:         id,
		Type:       locktracker.LockTypeFileLock,
		Name:       lockName,
		AcquiredAt: time.Now(),
		Used:       1, // File lock is binary: held (1) or not (0)
	})

	return &TrackedLock{
		Flock:    lock,
		id:       id,
		registry: registry,
	}, nil
}

// ReleaseTracked releases the tracked lock and removes it from tracking.
func ReleaseTracked(lock *TrackedLock) {
	if lock == nil {
		return
	}

	// Unregister from tracking
	lock.registry.Unregister(lock.id)

	// Release the underlying lock
	lockPath := lock.Path()
	//nolint:errcheck // best-effort cleanup
	lock.Unlock()
	os.Remove(lockPath) // best-effort cleanup
}

// BuildConfig returns a Config for module build locking.
func BuildConfig(moniker, baseDir string) Config {
	return Config{
		BaseDir:      baseDir,
		Identifier:   moniker,
		ResourceType: "module",
		ActionVerb:   "already being built",
	}
}

// ComponentBuildConfig returns a Config for component-level build locking.
// Use this when building components within a module in parallel.
func ComponentBuildConfig(module, component, baseDir string) Config {
	return Config{
		BaseDir:      baseDir,
		Identifier:   module + ":" + component,
		ResourceType: "component",
		ActionVerb:   "already being built",
	}
}

// TestConfig returns a Config for test suite locking.
func TestConfig(suiteName, baseDir string) Config {
	return Config{
		BaseDir:      baseDir,
		Identifier:   suiteName,
		ResourceType: "test suite",
		ActionVerb:   "already running",
	}
}

// ScanConfig returns a Config for module scan locking.
func ScanConfig(moniker, baseDir string) Config {
	return Config{
		BaseDir:      baseDir,
		Identifier:   moniker,
		ResourceType: "module",
		ActionVerb:   "already being scanned",
	}
}

// ComponentScanConfig returns a Config for component-level scan locking.
func ComponentScanConfig(module, component, baseDir string) Config {
	return Config{
		BaseDir:      baseDir,
		Identifier:   module + "-" + component, // Use dash to avoid path issues
		ResourceType: "component",
		ActionVerb:   "already being scanned",
	}
}

// LintConfig returns a Config for module lint locking.
func LintConfig(moniker, baseDir string) Config {
	return Config{
		BaseDir:      baseDir,
		Identifier:   moniker,
		ResourceType: "module",
		ActionVerb:   "already being linted",
	}
}

// ComponentLintConfig returns a Config for component-level lint locking.
// Use this when linting components within a module in parallel.
func ComponentLintConfig(module, component, baseDir string) Config {
	return Config{
		BaseDir:      baseDir,
		Identifier:   module + "-" + component, // Use dash to avoid path issues
		ResourceType: "component",
		ActionVerb:   "already being linted",
	}
}

// ManualTestFileConfig returns a Config for locking manual test result files.
// Use this when importing/merging manual test results to prevent concurrent writes.
func ManualTestFileConfig(filePath, baseDir string) Config {
	// Use file base name as identifier
	identifier := filepath.Base(filePath)
	return Config{
		BaseDir:      baseDir,
		Identifier:   identifier,
		ResourceType: "manual test file",
		ActionVerb:   "already being written",
	}
}
