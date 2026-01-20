package fileutil

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gofrs/flock"
)

const DefaultLockTimeout = 30 * time.Second

// AtomicWriteJSONWithLock combines atomic write with file locking.
// Pattern extracted from oscal/writer.go WriteAssessmentResults().
//
// Use this when multiple processes might write to the same file concurrently.
// Provides both atomicity (no partial writes) and mutual exclusion (no race conditions).
func AtomicWriteJSONWithLock(path string, v interface{}, perm os.FileMode) error {
	// Acquire exclusive lock
	lockPath := path + ".lock"
	lock := flock.New(lockPath)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultLockTimeout)
	defer cancel()

	locked, err := lock.TryLockContext(ctx, 100*time.Millisecond)
	if err != nil {
		return fmt.Errorf("acquiring lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("file locked by another process (timeout after %s)", DefaultLockTimeout)
	}
	defer func() {
		//nolint:errcheck // best-effort cleanup
		lock.Unlock()
		os.Remove(lockPath) // best-effort cleanup
	}()

	// Atomic write (protected by lock)
	return AtomicWriteJSON(path, v, perm)
}
