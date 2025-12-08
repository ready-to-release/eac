// docker.go - Shared Docker utilities for build system
package buildutil

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// IsDockerInDocker detects if we're running inside a Docker container.
// Uses multiple signals for robust detection:
// 1. R2R_DOCKER_MODE env var (set by r2r CLI when launching containers)
// 2. R2R_HOST_REPOROOT with Windows path while running on Linux (indicates container)
// 3. /.dockerenv file exists (standard Docker container indicator)
func IsDockerInDocker() bool {
	// Primary check: explicit env var
	if os.Getenv("R2R_DOCKER_MODE") == "true" {
		return true
	}

	// Fallback: R2R_HOST_REPOROOT is set with Windows path but we're on Linux
	// This catches old r2r CLI binaries that don't set R2R_DOCKER_MODE
	hostRoot := os.Getenv("R2R_HOST_REPOROOT")
	if hostRoot != "" && runtime.GOOS == "linux" {
		// Windows paths have backslashes or drive letters
		if strings.Contains(hostRoot, "\\") || (len(hostRoot) >= 2 && hostRoot[1] == ':') {
			return true
		}
	}

	// Final fallback: check for Docker container indicator file
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	return false
}

// IsDockerAvailable checks if Docker daemon is accessible by attempting to run 'docker info'.
// Returns true if Docker daemon is accessible, false otherwise.
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// FormatDockerVolumePath formats a path for use as a Docker volume mount source
// On Windows, converts C:\path to /c/path for Docker compatibility
func FormatDockerVolumePath(path string) string {
	if len(path) >= 2 && path[1] == ':' {
		driveLetter := strings.ToLower(string(path[0]))
		rest := strings.ReplaceAll(path[2:], "\\", "/")
		return "/" + driveLetter + rest
	}
	return strings.ReplaceAll(path, "\\", "/")
}
