package session

import (
	"fmt"
	"os"
	"strings"
)

// GetIdentifier returns a unique identifier for the current shell session
// This works across different platforms and shells
func GetIdentifier() string {
	// Try to get parent process ID (works on all platforms)
	ppid := os.Getppid()

	// Also check for common shell session environment variables
	// These provide better session identification when available

	// PowerShell session ID (Windows)
	if psSession := os.Getenv("POWERSHELL_DISTRIBUTION_CHANNEL"); psSession != "" {
		// Combine with PID for uniqueness
		if pid := os.Getenv("PID"); pid != "" {
			return fmt.Sprintf("pwsh-%s-%s", pid, psSession)
		}
	}

	// Terminal session identifiers (various shells)
	if termSession := os.Getenv("TERM_SESSION_ID"); termSession != "" {
		return fmt.Sprintf("term-%s", termSession)
	}

	// SSH session
	if sshClient := os.Getenv("SSH_CLIENT"); sshClient != "" {
		// Use first part of SSH_CLIENT (client IP/port)
		parts := strings.Fields(sshClient)
		if len(parts) > 0 {
			return fmt.Sprintf("ssh-%s-%d", parts[0], ppid)
		}
	}

	// WSL session (Windows Subsystem for Linux)
	if wslDistro := os.Getenv("WSL_DISTRO_NAME"); wslDistro != "" {
		return fmt.Sprintf("wsl-%s-%d", wslDistro, ppid)
	}

	// tmux/screen session
	if tmux := os.Getenv("TMUX"); tmux != "" {
		// Extract session ID from tmux environment
		parts := strings.Split(tmux, ",")
		if len(parts) >= 2 {
			return fmt.Sprintf("tmux-%s", parts[1])
		}
	}

	if screenSession := os.Getenv("STY"); screenSession != "" {
		return fmt.Sprintf("screen-%s", screenSession)
	}

	// Console/TTY based identification
	if tty := os.Getenv("TTY"); tty != "" {
		// Replace slashes to make valid filename
		ttyClean := strings.ReplaceAll(tty, "/", "_")
		return fmt.Sprintf("tty-%s-%d", ttyClean, ppid)
	}

	// Windows Console Host Process ID
	if winPid := os.Getenv("WT_SESSION"); winPid != "" {
		// Windows Terminal session
		return fmt.Sprintf("wt-%s", winPid)
	}

	// Default: use parent process ID
	// This works on all platforms but may not be perfect for all cases
	return fmt.Sprintf("pid-%d", ppid)
}
