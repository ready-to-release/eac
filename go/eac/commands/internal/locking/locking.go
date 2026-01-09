// Package locking provides file-based locking for build and test commands.
// It prevents concurrent execution of the same module build or test suite.
package locking

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
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
	if err := os.MkdirAll(lockDir, 0755); err != nil {
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
	lock.Unlock()
	os.Remove(lockPath)
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
