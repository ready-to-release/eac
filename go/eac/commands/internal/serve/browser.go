package serve

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/ready-to-release/eac/go/eac/commands/internal/dockerutil"
	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
)

// shouldOpenBrowser checks if browser should be opened based on environment.
// Returns true only in local interactive console (not in tests, CI, containers, or DinD).
func shouldOpenBrowser() bool {
	env := environment.Detect()

	// Don't open browsers in test contexts, CI, or containers
	if env.IsTestContext || env.IsCI || env.IsContainer {
		return false
	}

	// Also check DinD mode for backward compatibility
	// (R2R_HOST_REPOROOT indicates Docker-in-Docker)
	if dockerutil.IsDinD() {
		return false
	}

	return true
}

// OpenBrowser opens the default web browser to the given URL.
// In DinD mode or when R2R_NO_BROWSER=true, this is a no-op.
// Returns an error only if browser opening fails in normal mode.
func OpenBrowser(url string) error {
	if !shouldOpenBrowser() {
		return nil
	}

	return openBrowserNative(url)
}

// OpenBrowserWithFallback opens the browser and returns whether it was skipped.
// Returns (false, nil) if browser opening is disabled (DinD mode or R2R_NO_BROWSER=true).
// This allows callers to show appropriate messages.
func OpenBrowserWithFallback(url string) (opened bool, err error) {
	if !shouldOpenBrowser() {
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
