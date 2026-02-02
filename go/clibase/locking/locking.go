// Package locking provides file-based locking for build and test commands.
// It prevents concurrent execution of the same module build or test suite.
package locking

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/google/uuid"
	"github.com/ready-to-release/eac/go/clibase/locktracker"
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
	registry.Register(&locktracker.LockInfo{
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

// WaitConfig configures lock waiting behavior.
type WaitConfig struct {
	Timeout      time.Duration // Maximum time to wait for lock (default: 5 minutes)
	PollInterval time.Duration // How often to retry acquiring lock (default: 200ms)
}

// DefaultWaitConfig returns sensible defaults for lock waiting.
func DefaultWaitConfig() WaitConfig {
	return WaitConfig{
		Timeout:      5 * time.Minute,
		PollInterval: 200 * time.Millisecond,
	}
}

// AcquireWithWait attempts to acquire a lock, waiting with visual feedback if blocked.
// Unlike AcquireTracked, this will wait (up to timeout) instead of failing immediately.
// The wait state is tracked in the registry for TUI visualization.
// ctx can be used to cancel the wait early.
func AcquireWithWait(ctx context.Context, workspaceRoot string, cfg Config, registry *locktracker.Registry, waitCfg WaitConfig) (*TrackedLock, error) {
	lockDir := filepath.Join(workspaceRoot, cfg.BaseDir)

	// Ensure directory exists with proper permissions
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory %s: %w", lockDir, err)
	}

	lockPath := filepath.Join(lockDir, fmt.Sprintf(".lock-%s", cfg.Identifier))
	lock := flock.New(lockPath)

	id := uuid.New().String()
	lockName := fmt.Sprintf("%s:%s", cfg.ResourceType, cfg.Identifier)

	// First try without waiting
	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock at %s: %w", lockPath, err)
	}
	if locked {
		// Got it immediately - register and return
		registry.Register(&locktracker.LockInfo{
			ID:         id,
			Type:       locktracker.LockTypeFileLock,
			Name:       lockName,
			AcquiredAt: time.Now(),
			Used:       1,
		})
		return &TrackedLock{
			Flock:    lock,
			id:       id,
			registry: registry,
		}, nil
	}

	// Lock is held by another process - poll and only show "waiting" after 500ms
	// Use defaults if not specified
	if waitCfg.Timeout == 0 {
		waitCfg.Timeout = 5 * time.Minute
	}
	if waitCfg.PollInterval == 0 {
		waitCfg.PollInterval = 200 * time.Millisecond
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, waitCfg.Timeout)
	defer cancel()

	ticker := time.NewTicker(waitCfg.PollInterval)
	defer ticker.Stop()

	// Only register as "waiting" after 500ms delay (avoid flashing for quick locks)
	const waitDisplayDelay = 500 * time.Millisecond
	waitStart := time.Now()
	waitID := uuid.New().String()
	waitRegistered := false

	for {
		select {
		case <-timeoutCtx.Done():
			// Timeout or cancelled - remove waiting registration if we registered
			if waitRegistered {
				registry.Unregister(waitID)
			}
			if ctx.Err() != nil {
				return nil, fmt.Errorf("lock acquisition cancelled for %s '%s'", cfg.ResourceType, cfg.Identifier)
			}
			return nil, fmt.Errorf("timeout waiting for %s '%s' (held by another process)", cfg.ResourceType, cfg.Identifier)

		case <-ticker.C:
			// Register as "waiting" only after 500ms delay (avoid flashing for quick waits)
			if !waitRegistered && time.Since(waitStart) >= waitDisplayDelay {
				registry.Register(&locktracker.LockInfo{
					ID:      waitID,
					Type:    locktracker.LockTypeFileLock,
					Name:    lockName,
					Waiting: 1,
				})
				waitRegistered = true

				// Print user-friendly message to stdout
				action := extractAction(cfg.ActionVerb)
				fmt.Printf("Waiting for %s lock...\n", action)
			}

			// Try to acquire again
			locked, err := lock.TryLock()
			if err != nil {
				if waitRegistered {
					registry.Unregister(waitID)
				}
				return nil, fmt.Errorf("failed to acquire lock at %s: %w", lockPath, err)
			}
			if locked {
				// Got it! Remove waiting registration (if any) and register as held
				if waitRegistered {
					registry.Unregister(waitID)
				}
				registry.Register(&locktracker.LockInfo{
					ID:         id,
					Type:       locktracker.LockTypeFileLock,
					Name:       lockName,
					AcquiredAt: time.Now(),
					Used:       1,
				})
				return &TrackedLock{
					Flock:    lock,
					id:       id,
					registry: registry,
				}, nil
			}
			// Still locked, continue waiting
		}
	}
}

// extractAction converts ActionVerb to a user-friendly action name for waiting messages.
// "already being built" → "building", "already running" → "testing", etc.
func extractAction(actionVerb string) string {
	switch actionVerb {
	case "already being built":
		return "building"
	case "already running":
		return "testing"
	case "already being scanned":
		return "scanning"
	case "already being linted":
		return "linting"
	default:
		// Fallback: extract verb from "already being X" or "already X"
		if len(actionVerb) > 14 && actionVerb[:14] == "already being " {
			return actionVerb[14:]
		}
		return "lock"
	}
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

// UnitBuildConfig returns a Config for component-level build locking.
// Use this when building components within a module in parallel.
func UnitBuildConfig(module, component, baseDir string) Config {
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

// UnitTestConfig returns a Config for component-level test locking.
// Use this when testing components within a module in parallel.
func UnitTestConfig(module, component, baseDir string) Config {
	return Config{
		BaseDir:      baseDir,
		Identifier:   module + "-" + component, // Use dash to avoid path issues
		ResourceType: "test component",
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

// UnitScanConfig returns a Config for component-level scan locking.
func UnitScanConfig(module, component, baseDir string) Config {
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

// UnitLintConfig returns a Config for component-level lint locking.
// Use this when linting components within a module in parallel.
func UnitLintConfig(module, component, baseDir string) Config {
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
