//go:build windows

package workunit

import (
	"os"
)

// isProcessAlive checks if a process with the given PID is still running.
// On Windows, we use os.FindProcess and attempt to check process state.
func isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false // Can't find process
	}

	// On Windows, FindProcess always succeeds, so we need another check.
	// We attempt to wait with WNOHANG equivalent, but Go doesn't expose this well.
	// As a fallback, we try to signal the process - on Windows this will fail
	// for processes we don't own, so we assume alive if we can't tell.
	err = process.Signal(os.Signal(nil))
	if err != nil {
		// "os: process already finished" indicates the process is dead
		if err.Error() == "os: process already finished" {
			return false
		}
		// Other errors (like access denied) mean process likely exists
	}
	return true // Assume alive if we can't tell
}
