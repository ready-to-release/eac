//go:build !windows

package tool

import (
	"os/exec"
	"syscall"
)

// SetProcessGroup puts the command in its own process group.
// This allows killing all child processes when the parent is cancelled.
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// KillProcessGroup sends SIGKILL to the entire process group.
// Falls back to killing just the process if the group ID can't be obtained.
func KillProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return cmd.Process.Kill()
	}
	return syscall.Kill(-pgid, syscall.SIGKILL)
}
