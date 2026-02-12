// Package locking provides file-based locking for build and test commands.
// It prevents concurrent execution of the same module build or test suite.
package locking

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/google/uuid"
	"github.com/ready-to-release/eac/go/clibase/locktracker"
	"github.com/ready-to-release/eac/go/core/logging"
)

// Action represents the type of operation being locked.
type Action string

const (
	// ActionBuild is used for module or component build locks.
	ActionBuild Action = "build"
	// ActionTest is used for test suite or component test locks.
	ActionTest Action = "test"
	// ActionScan is used for module or component scan locks.
	ActionScan Action = "scan"
	// ActionLint is used for module or component lint locks.
	ActionLint Action = "lint"
	// ActionWrite is used for file-write locks (e.g. manual test results).
	ActionWrite Action = "write"
)

// actionMeta maps each Action to its resource-type defaults and messaging.
type actionMeta struct {
	moduleResourceType    string // resource type for module-level locks
	componentResourceType string // resource type for unit/component-level locks
	actionVerb            string // used in "already being ..." error messages
	waitLabel             string // user-friendly label for "Waiting for <label> lock..."
}

var actionTable = map[Action]actionMeta{
	ActionBuild: {moduleResourceType: "module", componentResourceType: "component", actionVerb: "already being built", waitLabel: "building"},
	ActionTest:  {moduleResourceType: "test suite", componentResourceType: "test component", actionVerb: "already running", waitLabel: "testing"},
	ActionScan:  {moduleResourceType: "module", componentResourceType: "component", actionVerb: "already being scanned", waitLabel: "scanning"},
	ActionLint:  {moduleResourceType: "module", componentResourceType: "component", actionVerb: "already being linted", waitLabel: "linting"},
	ActionWrite: {moduleResourceType: "manual test file", componentResourceType: "manual test file", actionVerb: "already being written", waitLabel: "writing"},
}

// NewConfig returns a Config for a module-level lock of the given action.
// identifier is typically the module moniker or suite name; baseDir is the
// relative output directory (e.g. "out/build").
func NewConfig(action Action, identifier, baseDir string) Config {
	meta := actionTable[action]
	return Config{
		BaseDir:      baseDir,
		Identifier:   identifier,
		ResourceType: meta.moduleResourceType,
		ActionVerb:   meta.actionVerb,
	}
}

// NewUnitConfig returns a Config for a component-level lock within a module.
// module and component are joined to form a unique identifier.
func NewUnitConfig(action Action, module, component, baseDir string) Config {
	meta := actionTable[action]
	sep := "-"
	if action == ActionBuild {
		sep = ":"
	}
	return Config{
		BaseDir:      baseDir,
		Identifier:   module + sep + component,
		ResourceType: meta.componentResourceType,
		ActionVerb:   meta.actionVerb,
	}
}

// NewFileConfig returns a Config for locking a specific file path.
// The file's base name is used as the lock identifier.
func NewFileConfig(action Action, filePath, baseDir string) Config {
	meta := actionTable[action]
	return Config{
		BaseDir:      baseDir,
		Identifier:   filepath.Base(filePath),
		ResourceType: meta.moduleResourceType,
		ActionVerb:   meta.actionVerb,
	}
}

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
	Timeout      time.Duration          // Maximum time to wait for lock (default: 5 minutes)
	PollInterval time.Duration          // How often to retry acquiring lock (default: 200ms)
	Writer       io.Writer              // Destination for "Waiting for lock" messages (nil = discard)
	Logger       *logging.ComponentLogger // Structured logger for wait messages (nil = use Writer)
}

// DefaultWaitConfig returns sensible defaults for lock waiting.
func DefaultWaitConfig() WaitConfig {
	return WaitConfig{
		Timeout:      5 * time.Minute,
		PollInterval: 200 * time.Millisecond,
		Logger:       logging.C(),
	}
}

// AcquireWithWait attempts to acquire a lock, waiting with visual feedback if blocked.
// Unlike AcquireTracked, this will wait (up to timeout) instead of failing immediately.
// The wait state is tracked in the registry for TUI visualization.
// ctx can be used to cancel the wait early.
func AcquireWithWait(ctx context.Context, workspaceRoot string, cfg Config, registry *locktracker.Registry, waitCfg WaitConfig) (*TrackedLock, error) {
	lock, lockName, err := prepareLock(workspaceRoot, cfg)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()

	// First try without waiting
	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock at %s: %w", lock.Path(), err)
	}
	if locked {
		return registerHeldLock(lock, id, lockName, registry), nil
	}

	// Lock is held by another process -- enter polling loop
	return pollForLock(ctx, lock, id, lockName, cfg, registry, waitCfg)
}

// prepareLock creates the lock directory and returns a flock handle.
func prepareLock(workspaceRoot string, cfg Config) (*flock.Flock, string, error) {
	lockDir := filepath.Join(workspaceRoot, cfg.BaseDir)
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		return nil, "", fmt.Errorf("failed to create lock directory %s: %w", lockDir, err)
	}
	lockPath := filepath.Join(lockDir, fmt.Sprintf(".lock-%s", cfg.Identifier))
	lockName := fmt.Sprintf("%s:%s", cfg.ResourceType, cfg.Identifier)
	return flock.New(lockPath), lockName, nil
}

// registerHeldLock registers an acquired lock with the tracker and returns a TrackedLock.
func registerHeldLock(lock *flock.Flock, id, lockName string, registry *locktracker.Registry) *TrackedLock {
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
	}
}

// pollForLock retries lock acquisition until success, timeout, or cancellation.
func pollForLock(ctx context.Context, lock *flock.Flock, id, lockName string, cfg Config, registry *locktracker.Registry, waitCfg WaitConfig) (*TrackedLock, error) {
	// Apply defaults
	if waitCfg.Timeout == 0 {
		waitCfg.Timeout = 5 * time.Minute
	}
	if waitCfg.PollInterval == 0 {
		waitCfg.PollInterval = 200 * time.Millisecond
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, waitCfg.Timeout)
	defer cancel()

	ticker := time.NewTicker(waitCfg.PollInterval)
	defer ticker.Stop()

	const waitDisplayDelay = 500 * time.Millisecond
	waitStart := time.Now()
	waitID := uuid.New().String()
	waitRegistered := false

	defer func() {
		if waitRegistered {
			registry.Unregister(waitID)
		}
	}()

	for {
		select {
		case <-timeoutCtx.Done():
			if ctx.Err() != nil {
				return nil, fmt.Errorf("lock acquisition cancelled for %s '%s'", cfg.ResourceType, cfg.Identifier)
			}
			return nil, fmt.Errorf("timeout waiting for %s '%s' (held by another process)", cfg.ResourceType, cfg.Identifier)

		case <-ticker.C:
			// Register as "waiting" after delay (avoids flashing for quick waits)
			if !waitRegistered && time.Since(waitStart) >= waitDisplayDelay {
				registry.Register(&locktracker.LockInfo{
					ID:      waitID,
					Type:    locktracker.LockTypeFileLock,
					Name:    lockName,
					Waiting: 1,
				})
				waitRegistered = true

				action := extractAction(cfg.ActionVerb)
				if waitCfg.Logger != nil {
					waitCfg.Logger.Infof("Waiting for %s lock...", action)
				} else if waitCfg.Writer != nil {
					fmt.Fprintf(waitCfg.Writer, "Waiting for %s lock...\n", action)
				}
			}

			locked, err := lock.TryLock()
			if err != nil {
				return nil, fmt.Errorf("failed to acquire lock at %s: %w", lock.Path(), err)
			}
			if locked {
				// Clear waiting registration before registering as held
				if waitRegistered {
					registry.Unregister(waitID)
					waitRegistered = false
				}
				return registerHeldLock(lock, id, lockName, registry), nil
			}
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

// Deprecated convenience functions -- kept for backward compatibility.
// Prefer NewConfig, NewUnitConfig, or NewFileConfig.

// BuildConfig returns a Config for module build locking.
// Deprecated: use NewConfig(ActionBuild, moniker, baseDir).
func BuildConfig(moniker, baseDir string) Config { return NewConfig(ActionBuild, moniker, baseDir) }

// UnitBuildConfig returns a Config for component-level build locking.
// Deprecated: use NewUnitConfig(ActionBuild, module, component, baseDir).
func UnitBuildConfig(module, component, baseDir string) Config {
	return NewUnitConfig(ActionBuild, module, component, baseDir)
}

// TestConfig returns a Config for test suite locking.
// Deprecated: use NewConfig(ActionTest, suiteName, baseDir).
func TestConfig(suiteName, baseDir string) Config { return NewConfig(ActionTest, suiteName, baseDir) }

// UnitTestConfig returns a Config for component-level test locking.
// Deprecated: use NewUnitConfig(ActionTest, module, component, baseDir).
func UnitTestConfig(module, component, baseDir string) Config {
	return NewUnitConfig(ActionTest, module, component, baseDir)
}

// ScanConfig returns a Config for module scan locking.
// Deprecated: use NewConfig(ActionScan, moniker, baseDir).
func ScanConfig(moniker, baseDir string) Config { return NewConfig(ActionScan, moniker, baseDir) }

// UnitScanConfig returns a Config for component-level scan locking.
// Deprecated: use NewUnitConfig(ActionScan, module, component, baseDir).
func UnitScanConfig(module, component, baseDir string) Config {
	return NewUnitConfig(ActionScan, module, component, baseDir)
}

// LintConfig returns a Config for module lint locking.
// Deprecated: use NewConfig(ActionLint, moniker, baseDir).
func LintConfig(moniker, baseDir string) Config { return NewConfig(ActionLint, moniker, baseDir) }

// UnitLintConfig returns a Config for component-level lint locking.
// Deprecated: use NewUnitConfig(ActionLint, module, component, baseDir).
func UnitLintConfig(module, component, baseDir string) Config {
	return NewUnitConfig(ActionLint, module, component, baseDir)
}

// ManualTestFileConfig returns a Config for locking manual test result files.
// Deprecated: use NewFileConfig(ActionWrite, filePath, baseDir).
func ManualTestFileConfig(filePath, baseDir string) Config {
	return NewFileConfig(ActionWrite, filePath, baseDir)
}
