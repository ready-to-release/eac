package workunit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

// LockInfo represents information about an acquired lock.
type LockInfo struct {
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
	UnitID    string    `json:"unit_id"`
}

// ErrLocked is returned when a lock cannot be acquired because it's held.
type ErrLocked struct {
	UnitID   string
	HeldBy   *LockInfo
	LockPath string
}

func (e *ErrLocked) Error() string {
	if e.HeldBy != nil {
		return fmt.Sprintf("unit %s is locked by PID %d since %s",
			e.UnitID, e.HeldBy.PID, e.HeldBy.StartedAt.Format(time.RFC3339))
	}
	return fmt.Sprintf("unit %s is locked", e.UnitID)
}

// Lock acquires exclusive access to this unit.
// Creates the lock file with information about the current process.
// Returns ErrLocked if the lock is already held.
func (u UnitID) Lock() error {
	return u.LockWithRoot("")
}

// LockWithRoot acquires exclusive access using a workspace root prefix.
// If root is empty, uses relative paths.
func (u UnitID) LockWithRoot(root string) error {
	outDir := u.OutDir()
	lockFile := u.LockFile()
	if root != "" {
		outDir = root + "/" + outDir
		lockFile = root + "/" + lockFile
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			// Try to read existing lock info
			info := u.ReadLockInfoWithRoot(root)
			return &ErrLocked{
				UnitID:   u.Longname(),
				HeldBy:   info,
				LockPath: lockFile,
			}
		}
		return fmt.Errorf("failed to create lock file: %w", err)
	}
	defer f.Close()

	info := LockInfo{
		PID:       os.Getpid(),
		StartedAt: time.Now(),
		UnitID:    u.Longname(),
	}

	return json.NewEncoder(f).Encode(info)
}

// Unlock releases the lock.
// This operation is idempotent - it succeeds even if the lock doesn't exist.
func (u UnitID) Unlock() error {
	return u.UnlockWithRoot("")
}

// UnlockWithRoot releases the lock using a workspace root prefix.
func (u UnitID) UnlockWithRoot(root string) error {
	lockFile := u.LockFile()
	if root != "" {
		lockFile = root + "/" + lockFile
	}

	err := os.Remove(lockFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}
	return nil
}

// IsLocked returns true if this unit is currently locked.
func (u UnitID) IsLocked() bool {
	return u.IsLockedWithRoot("")
}

// IsLockedWithRoot returns true if this unit is currently locked.
func (u UnitID) IsLockedWithRoot(root string) bool {
	lockFile := u.LockFile()
	if root != "" {
		lockFile = root + "/" + lockFile
	}
	_, err := os.Stat(lockFile)
	return err == nil
}

// ReadLockInfo reads the lock information if the unit is locked.
// Returns nil if not locked or if the lock file can't be read.
func (u UnitID) ReadLockInfo() *LockInfo {
	return u.ReadLockInfoWithRoot("")
}

// ReadLockInfoWithRoot reads the lock information with a workspace root prefix.
func (u UnitID) ReadLockInfoWithRoot(root string) *LockInfo {
	lockFile := u.LockFile()
	if root != "" {
		lockFile = root + "/" + lockFile
	}

	data, err := os.ReadFile(lockFile)
	if err != nil {
		return nil
	}

	var info LockInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil
	}

	return &info
}

// LockWaitConfig configures lock waiting behavior.
type LockWaitConfig struct {
	Timeout      time.Duration // Maximum time to wait (default: 5 minutes)
	PollInterval time.Duration // How often to retry (default: 200ms)
}

// DefaultLockWaitConfig returns sensible defaults for lock waiting.
func DefaultLockWaitConfig() LockWaitConfig {
	return LockWaitConfig{
		Timeout:      5 * time.Minute,
		PollInterval: 200 * time.Millisecond,
	}
}

// LockWithWait attempts to acquire the lock, waiting if blocked.
// Returns when the lock is acquired or the context is cancelled/timeout reached.
func (u UnitID) LockWithWait(ctx context.Context, root string, cfg LockWaitConfig) error {
	// Use defaults if not specified
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Minute
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 200 * time.Millisecond
	}

	// Try immediately first
	err := u.LockWithRoot(root)
	if err == nil {
		return nil
	}

	// If not a lock error, return immediately
	if _, ok := err.(*ErrLocked); !ok {
		return err
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			if ctx.Err() != nil {
				return fmt.Errorf("lock acquisition cancelled for %s", u.Longname())
			}
			// Return the last lock error with holder info
			info := u.ReadLockInfoWithRoot(root)
			return &ErrLocked{
				UnitID: u.Longname(),
				HeldBy: info,
			}

		case <-ticker.C:
			err := u.LockWithRoot(root)
			if err == nil {
				return nil
			}
			if _, ok := err.(*ErrLocked); !ok {
				return err
			}
			// Still locked, continue waiting
		}
	}
}

// TryBreakStaleLock attempts to break a lock if the holding process is dead.
// Returns true if the lock was broken, false if it's still held by a live process.
// This is a best-effort operation and may have race conditions.
func (u UnitID) TryBreakStaleLock(root string) bool {
	info := u.ReadLockInfoWithRoot(root)
	if info == nil {
		return false // No lock or can't read it
	}

	// Check if the process is still alive
	// On Unix, sending signal 0 checks process existence without killing it.
	// We use syscall.Signal(0) because process.Signal(nil) is not reliable
	// or supported on all platforms/Go versions for this purpose.
	err := syscall.Kill(info.PID, 0)
	if err != nil {
		// If err is ESRCH, the process definitely doesn't exist.
		// For other errors (like EPERM), the process might exist but we can't signal it.
		// We only break the lock if we are sure the process is gone.
		if errors.Is(err, syscall.ESRCH) {
			_ = u.UnlockWithRoot(root)
			return true
		}
	}

	return false // Process is still alive (or we can't tell)
}
