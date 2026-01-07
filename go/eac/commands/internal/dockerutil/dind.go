// Package dockerutil provides shared Docker utilities for Docker-in-Docker (DinD) scenarios
// and general Docker operations used across multiple commands.
package dockerutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/repository"
)

const (
	// EnvHostRepoRoot is the environment variable that contains the host repository root
	// when running inside a Docker container (DinD mode)
	EnvHostRepoRoot = "R2R_HOST_REPOROOT"

	// EnvContainerRepoRoot is the environment variable that contains the container's
	// view of the repository root (typically /var/task)
	EnvContainerRepoRoot = "R2R_CONTAINER_REPOROOT"

	// EnvDockerMode is explicitly set by r2r CLI when launching containers
	EnvDockerMode = "R2R_DOCKER_MODE"

	// DefaultContainerRepoRoot is the default path where the repository is mounted
	// inside the r2r CLI container
	DefaultContainerRepoRoot = "/var/task"
)

// IsDinD returns true if running inside a Docker container (DinD mode).
// Uses multiple signals for robust detection:
// 1. R2R_HOST_REPOROOT environment variable (primary indicator)
// 2. R2R_DOCKER_MODE explicit flag
// 3. Windows host path while running on Linux
// 4. /.dockerenv file exists
func IsDinD() bool {
	// Primary check: R2R_HOST_REPOROOT is set
	if os.Getenv(EnvHostRepoRoot) != "" {
		return true
	}

	// Secondary check: explicit env var
	if os.Getenv(EnvDockerMode) == "true" {
		return true
	}

	// Fallback: R2R_HOST_REPOROOT is set with Windows path but we're on Linux
	// This catches old r2r CLI binaries that don't set R2R_DOCKER_MODE
	hostRoot := os.Getenv(EnvHostRepoRoot)
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

// GetHostRepoRoot returns the host's repository root path.
// In DinD mode, this comes from the R2R_HOST_REPOROOT environment variable.
// In direct host mode, this returns the actual repository root.
func GetHostRepoRoot() (string, error) {
	if hostRoot := os.Getenv(EnvHostRepoRoot); hostRoot != "" {
		return hostRoot, nil
	}
	return repository.GetRepositoryRoot("")
}

// GetContainerRepoRoot returns the container's view of the repository root.
// In DinD mode, this is typically /var/task.
// In direct host mode, this returns the actual repository root.
func GetContainerRepoRoot() (string, error) {
	if containerRoot := os.Getenv(EnvContainerRepoRoot); containerRoot != "" {
		return containerRoot, nil
	}
	if IsDinD() {
		return DefaultContainerRepoRoot, nil
	}
	return repository.GetRepositoryRoot("")
}

// TranslatePathForMount translates a local path to the appropriate path for
// Docker volume mounting in DinD environments.
//
// In DinD mode:
//   - Input path is relative to container repo root (/var/task)
//   - Output path is translated to host repo root (R2R_HOST_REPOROOT)
//
// In direct host mode:
//   - Path is returned unchanged (already a host path)
//
// Example (DinD mode):
//
//	Input: /var/task/docs
//	R2R_HOST_REPOROOT: C:\projects\eac
//	Output: C:\projects\eac\docs
func TranslatePathForMount(localPath string) (string, error) {
	if !IsDinD() {
		// Direct host mode: path is already a host path
		return localPath, nil
	}

	hostRoot := os.Getenv(EnvHostRepoRoot)
	containerRoot, err := GetContainerRepoRoot()
	if err != nil {
		return "", err
	}

	// Make the path relative to container root, then join with host root
	relPath, err := makeRelative(localPath, containerRoot)
	if err != nil {
		// Path is not under container root, return as-is
		// This handles edge cases where absolute paths are used
		return localPath, nil
	}

	// Join with host root
	// When running on Linux but the host is Windows (DinD on Windows),
	// filepath.Join uses '/' but we need '\' to match the Windows host path.
	// Detect Windows paths by checking for drive letter (e.g., "C:\")
	if len(hostRoot) >= 2 && hostRoot[1] == ':' {
		// Windows host path: use backslash separator
		if relPath == "" {
			return hostRoot, nil
		}
		// Convert forward slashes in relPath to backslashes
		relPath = strings.ReplaceAll(relPath, "/", "\\")
		return hostRoot + "\\" + relPath, nil
	}

	// Unix host path: use filepath.Join which handles separators
	return filepath.Join(hostRoot, relPath), nil
}

// makeRelative returns the relative path from base to target.
// Returns an error if target is not under base.
func makeRelative(target, base string) (string, error) {
	// Normalize paths for comparison
	target = filepath.Clean(target)
	base = filepath.Clean(base)

	// Handle the case where paths use different separators
	// This can happen when comparing Windows paths with Unix paths
	targetNorm := strings.ReplaceAll(target, "\\", "/")
	baseNorm := strings.ReplaceAll(base, "\\", "/")

	if !strings.HasPrefix(targetNorm, baseNorm) {
		return "", os.ErrNotExist
	}

	rel := strings.TrimPrefix(targetNorm, baseNorm)
	rel = strings.TrimPrefix(rel, "/")
	return rel, nil
}

// FormatDockerVolume formats a path for use as a Docker volume mount source.
// On Windows, this converts paths like C:\path to /c/path for Docker compatibility.
func FormatDockerVolume(path string) string {
	// Check if this is a Windows absolute path (e.g., C:\...)
	if len(path) >= 2 && path[1] == ':' {
		// Convert C:\path to /c/path
		driveLetter := strings.ToLower(string(path[0]))
		rest := strings.ReplaceAll(path[2:], "\\", "/")
		return "/" + driveLetter + rest
	}

	// Already Unix-style or relative path
	return strings.ReplaceAll(path, "\\", "/")
}

// IsDockerAvailable checks if Docker daemon is accessible by attempting to run 'docker info'.
// Returns true if Docker daemon is accessible, false otherwise.
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}
