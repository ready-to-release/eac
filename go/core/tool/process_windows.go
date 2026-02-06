//go:build windows

package tool

import (
	"os/exec"
	"syscall"
)

// SetProcessGroup puts the command in a new process group on Windows.
// CREATE_NEW_PROCESS_GROUP allows terminating the group independently.
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
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
