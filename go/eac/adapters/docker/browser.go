package docker

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/ready-to-release/eac/go/eac/adapters/docker/util"
)

// shouldOpenBrowser checks if browser should be opened based on environment.
// Returns true only in local interactive console (not in tests, CI, containers, or DinD).
func shouldOpenBrowser() bool {
	// Don't open browsers in CI
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("GITLAB_CI") != "" {
		return false
	}

	// Don't open browsers in test contexts
	if os.Getenv("R2R_TEST_RUN_ID") != "" || os.Getenv("GODOG_FORMAT") != "" {
		return false
	}

	// Check DinD mode for backward compatibility
	// (R2R_HOST_REPOROOT indicates Docker-in-Docker)
	if util.IsDinD() {
		return false
	}

	// Check if running in container mode
	if os.Getenv("R2R_CONTEXT") == "r2r-cli" {
		return false
	}

	// Check explicit no-browser flag
	if os.Getenv("R2R_NO_BROWSER") == "true" {
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
