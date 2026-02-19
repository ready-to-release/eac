//go:build windows

package tool

import (
	"os/exec"
)

// SetProcessGroup is a no-op on Windows.
// Go 1.21+ uses job objects via exec.CommandContext for child process termination,
// making CREATE_NEW_PROCESS_GROUP unnecessary. That flag also breaks grandchild
// processes (e.g., go test spawning PowerShell) by causing exit code 0xFFFFFFFF.
func SetProcessGroup(cmd *exec.Cmd) {
	// no-op: job objects handle process tree cleanup on Windows
}

// KillProcessGroup kills the process on Windows.
// Windows does not support negative PID kill, so we kill the process directly.
// exec.CommandContext already handles child termination via job objects on Go 1.21+.
func KillProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}
