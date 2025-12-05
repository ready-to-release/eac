// Package internal provides path conversion utilities for Docker on Windows.
package internal

import (
	"path/filepath"
	"runtime"
	"strings"
)

// ToDockerPath converts a Windows path to Docker-compatible format for bind mounts.
// On Windows, Docker Desktop requires paths in Unix-style format.
//
// Examples:
//   - C:\Users\name\project → /c/Users/name/project
//   - D:\code\myapp → /d/code/myapp
//   - /already/unix/path → /already/unix/path (unchanged)
//
// On non-Windows platforms, returns the path unchanged.
func ToDockerPath(path string) string {
	// On non-Windows, return as-is
	if runtime.GOOS != "windows" {
		return filepath.ToSlash(path)
	}

	// Convert to absolute path if relative
	absPath, err := filepath.Abs(path)
	if err != nil {
		// If we can't get absolute path, try to convert what we have
		absPath = path
	}

	// Check if it's a Windows path (has drive letter)
	if len(absPath) >= 2 && absPath[1] == ':' {
		// Extract drive letter and path
		drive := strings.ToLower(string(absPath[0]))
		pathPart := absPath[2:] // Skip "C:"

		// Convert backslashes to forward slashes
		pathPart = filepath.ToSlash(pathPart)

		// Return in Docker format: /c/path/to/dir
		return "/" + drive + pathPart
	}

	// If already in Unix format or WSL path, return with forward slashes
	return filepath.ToSlash(absPath)
}
