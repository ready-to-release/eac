package serve

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/ready-to-release/eac/go/eac/commands/internal/dockerutil"
)

// OpenBrowser opens the default web browser to the given URL.
// In DinD mode, this is a no-op since there's no browser available inside containers.
// Returns an error only if not in DinD mode and browser opening fails.
func OpenBrowser(url string) error {
	if dockerutil.IsDinD() {
		// In DinD mode, we can't open a browser - this is expected
		// Return nil to indicate "success" (no error, just nothing to do)
		return nil
	}

	return openBrowserNative(url)
}

// OpenBrowserWithFallback opens the browser and returns whether it was skipped due to DinD.
// This allows callers to show appropriate messages.
func OpenBrowserWithFallback(url string) (opened bool, err error) {
	if dockerutil.IsDinD() {
		return false, nil
	}

	err = openBrowserNative(url)
	if err != nil {
		return false, err
	}
	return true, nil
}

// openBrowserNative opens the browser using platform-specific commands.
func openBrowserNative(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	return nil
}
