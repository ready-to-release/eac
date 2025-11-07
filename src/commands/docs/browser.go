package docs

import (
	"fmt"
	"os/exec"
	"runtime"
)

// openBrowser opens the default web browser to the given URL
func openBrowser(url string) error {
	command := detectBrowser()
	if command == "" {
		return fmt.Errorf("unable to detect browser command for platform: %s", runtime.GOOS)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command(command, url)
	default: // linux, freebsd, etc.
		cmd = exec.Command(command, url)
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	return nil
}

// detectBrowser returns the command to open a URL in the default browser
func detectBrowser() string {
	switch runtime.GOOS {
	case "windows":
		return "cmd /c start"
	case "darwin":
		return "open"
	default: // linux, freebsd, openbsd, netbsd
		return "xdg-open"
	}
}

// DetectBrowser is exported for testing
func DetectBrowser() string {
	return detectBrowser()
}
