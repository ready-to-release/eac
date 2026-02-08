package workunit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
func Lock(u UnitID) error {
	return LockWithRoot(u, "")
}

// LockWithRoot acquires exclusive access using a workspace root prefix.
// If root is empty, uses relative paths.
func LockWithRoot(u UnitID, root string) error {
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
			info := ReadLockInfoWithRoot(u, root)
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
func Unlock(u UnitID) error {
	return UnlockWithRoot(u, "")
}

// UnlockWithRoot releases the lock using a workspace root prefix.
func UnlockWithRoot(u UnitID, root string) error {
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
func IsLocked(u UnitID) bool {
	return IsLockedWithRoot(u, "")
}

// IsLockedWithRoot returns true if this unit is currently locked.
func IsLockedWithRoot(u UnitID, root string) bool {
	lockFile := u.LockFile()
	if root != "" {
		lockFile = root + "/" + lockFile
	}
	_, err := os.Stat(lockFile)
	return err == nil
}

// ReadLockInfo reads the lock information if the unit is locked.
// Returns nil if not locked or if the lock file can't be read.
func ReadLockInfo(u UnitID) *LockInfo {
	return ReadLockInfoWithRoot(u, "")
}

// ReadLockInfoWithRoot reads the lock information with a workspace root prefix.
func ReadLockInfoWithRoot(u UnitID, root string) *LockInfo {
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
func LockWithWait(u UnitID, ctx context.Context, root string, cfg LockWaitConfig) error {
	// Use defaults if not specified
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Minute
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 200 * time.Millisecond
	}

	// Try immediately first
	err := LockWithRoot(u, root)
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
			info := ReadLockInfoWithRoot(u, root)
			return &ErrLocked{
				UnitID: u.Longname(),
				HeldBy: info,
			}

		case <-ticker.C:
			err := LockWithRoot(u, root)
			if err == nil {
				return nil
			}
			if _, ok := err.(*ErrLocked); !ok {
				// On Windows, access-denied during lock creation is transient
				// (file handle not fully released after os.Remove)
				if !isWindowsAccessDenied(err) {
					return err
				}
			}
			// Still locked or transient Windows error, continue waiting
		}
	}
}

// TryBreakStaleLock attempts to break a lock if the holding process is dead.
// Returns true if the lock was broken, false if it's still held by a live process.
// This is a best-effort operation and may have race conditions.
func TryBreakStaleLock(u UnitID, root string) bool {
	info := ReadLockInfoWithRoot(u, root)
	if info == nil {
		return false // No lock or can't read it
	}

	// Check if the process is still alive using cross-platform approach
	if isProcessAlive(info.PID) {
		return false // Process is still alive
	}

	// Process is dead, break the lock
	_ = UnlockWithRoot(u, root)
	return true
}
