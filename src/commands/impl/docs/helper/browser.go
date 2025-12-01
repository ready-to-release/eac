package docs

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/ready-to-release/eac/src/core/logging"
	"go.uber.org/zap"
)

// openBrowser opens the default web browser to the given URL
func openBrowser(url string, logger *logging.Logger) error {
	logger.Debug("Attempting to open browser", zap.String("url", url), zap.String("platform", runtime.GOOS))

	command := detectBrowser()
	if command == "" {
		logger.Error("Unable to detect browser command", zap.String("platform", runtime.GOOS))
		return fmt.Errorf("unable to detect browser command for platform: %s", runtime.GOOS)
	}

	logger.Debug("Browser command detected", zap.String("command", command))

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
		logger.Debug("Using Windows browser command", zap.Strings("args", []string{"cmd", "/c", "start", url}))
	case "darwin":
		cmd = exec.Command(command, url)
		logger.Debug("Using macOS browser command", zap.Strings("args", []string{command, url}))
	default: // linux, freebsd, etc.
		cmd = exec.Command(command, url)
		logger.Debug("Using Linux/Unix browser command", zap.Strings("args", []string{command, url}))
	}

	err := cmd.Start()
	if err != nil {
		logger.Error("Failed to start browser process", zap.Error(err))
		return fmt.Errorf("failed to open browser: %w", err)
	}

	logger.Debug("Browser process started successfully")
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
