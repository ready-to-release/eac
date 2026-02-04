//go:build !windows

package workunit

import (
	"errors"
	"syscall"
)

// isProcessAlive checks if a process with the given PID is still running.
// On Unix, sending signal 0 checks process existence without killing it.
func isProcessAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true // Process exists and we can signal it
	}
	// ESRCH means the process doesn't exist
	// EPERM means it exists but we can't signal it (still alive)
	if errors.Is(err, syscall.ESRCH) {
		return false
	}
	return true // Assume alive if we can't tell
}
