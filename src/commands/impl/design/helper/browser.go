package design

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenBrowser opens the given URL in the default browser
// Returns an error if the browser could not be opened
func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows: use 'cmd /c start'
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		// macOS: use 'open'
		cmd = exec.Command("open", url)
	case "linux":
		// Linux: use 'xdg-open'
		cmd = exec.Command("xdg-open", url)
	default:
		// FreeBSD, OpenBSD, etc: try xdg-open
		cmd = exec.Command("xdg-open", url)
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	return nil
}

// DetectBrowser returns the command used to open browsers on this platform
func DetectBrowser() string {
	switch runtime.GOOS {
	case "windows":
		return "cmd /c start"
	case "darwin":
		return "open"
	case "linux":
		return "xdg-open"
	default:
		return "xdg-open"
	}
}
