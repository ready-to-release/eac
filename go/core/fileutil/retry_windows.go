//go:build windows

package fileutil

import (
	"errors"
	"os"
	"syscall"
	"time"
)

// Windows error codes for file locking conflicts.
const (
	errorSharingViolation = syscall.Errno(0x20) // ERROR_SHARING_VIOLATION
	errorLockViolation    = syscall.Errno(0x21) // ERROR_LOCK_VIOLATION
	errorAccessDenied     = syscall.Errno(0x05) // ERROR_ACCESS_DENIED
)

// RemoveAllWithRetry removes path, retrying on Windows file locking errors.
// Up to 5 retries with exponential backoff (100ms, 200ms, 400ms, 800ms, 1.6s; ~3.1s total).
func RemoveAllWithRetry(path string) error {
	var err error
	delay := 100 * time.Millisecond
	for attempt := 0; attempt < 5; attempt++ {
		err = os.RemoveAll(path)
		if err == nil || os.IsNotExist(err) {
			return nil
		}
		if !isRetryableError(err) {
			return err
		}
		time.Sleep(delay)
		delay *= 2
	}
	return err
}

// isRetryableError returns true for Windows file locking errors.
func isRetryableError(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case syscall.EACCES, errorSharingViolation, errorLockViolation, errorAccessDenied:
			return true
		}
	}
	return false
}
