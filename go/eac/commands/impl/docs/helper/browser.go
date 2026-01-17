package docs

import (
	"fmt"
	"os/exec"
	"runtime"
)

// openBrowser opens the default web browser to the given URL.
func openBrowser(url string) error {
	log.Debugf("Attempting to open browser: url=%s, platform=%s", url, runtime.GOOS)

	command := detectBrowser()
	if command == "" {
		log.Errorf("Unable to detect browser command: platform=%s", runtime.GOOS)
		return fmt.Errorf("unable to detect browser command for platform: %s", runtime.GOOS)
	}

	log.Debugf("Browser command detected: command=%s", command)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
		log.Debugf("Using Windows browser command: args=[cmd /c start %s]", url)
	case "darwin":
		cmd = exec.Command(command, url)
		log.Debugf("Using macOS browser command: args=[%s %s]", command, url)
	default: // linux, freebsd, etc.
		cmd = exec.Command(command, url)
		log.Debugf("Using Linux/Unix browser command: args=[%s %s]", command, url)
	}

	err := cmd.Start()
	if err != nil {
		log.Errorf("Failed to start browser process: error=%v", err)
		return fmt.Errorf("failed to open browser: %w", err)
	}

	log.Debugf("Browser process started successfully")
	return nil
}

// detectBrowser returns the command to open a URL in the default browser.
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

// DetectBrowser is exported for testing.
func DetectBrowser() string {
	return detectBrowser()
}
