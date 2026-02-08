//go:build windows

package workunit

import (
	"errors"
	"syscall"
)

// Windows error codes for file locking conflicts.
const (
	errSharingViolation = syscall.Errno(0x20) // ERROR_SHARING_VIOLATION
	errAccessDenied     = syscall.Errno(0x05) // ERROR_ACCESS_DENIED
)

// isWindowsAccessDenied returns true if the error is a transient Windows
// file locking error (access denied or sharing violation). These occur
// when a file handle has not been fully released after os.Remove.
func isWindowsAccessDenied(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == errAccessDenied || errno == errSharingViolation
	}
	return false
}
