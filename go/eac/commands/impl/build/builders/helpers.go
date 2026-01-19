// helpers.go - Shared utility functions for builders
package builders

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ready-to-release/eac/go/eac/commands/internal/dockerutil"
	"github.com/ready-to-release/eac/go/eac/core/platform"
)

// Logln writes a formatted string with platform-specific line ending to the writer.
func Logln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+platform.LineEnding, args...)
}

// RunCommandWithLog executes a command in the specified directory
// Output is written to the provided writer
// Returns exit code (0 = success, non-zero = failure).
func RunCommandWithLog(dir string, logWriter io.Writer, name string, args ...string) int {
	// Use platform-aware command wrapper (handles .cmd files on Windows)
	wrappedName, wrappedArgs := platform.WrapCommand(name, args...)
	cmd := exec.Command(wrappedName, wrappedArgs...)
	cmd.Dir = dir
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	if err := cmd.Run(); err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		Logln(logWriter, "\nError: failed to execute command: %v", err)
		return 1
	}

	return 0
}

// RunCommandWithEnv executes a command with custom environment variables.
func RunCommandWithEnv(dir string, logWriter io.Writer, env []string, name string, args ...string) int {
	// Use platform-aware command wrapper (handles .cmd files on Windows)
	wrappedName, wrappedArgs := platform.WrapCommand(name, args...)
	cmd := exec.Command(wrappedName, wrappedArgs...)
	cmd.Dir = dir
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	cmd.Env = append(os.Environ(), env...)

	if err := cmd.Run(); err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		Logln(logWriter, "\nError: failed to execute command: %v", err)
		return 1
	}

	return 0
}

// CopyFile copies a file from src to dst, preserving permissions.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

// FormatDockerVolumePath formats a path for use as a Docker volume mount source
// On Windows, converts C:\path to /c/path for Docker compatibility.
func FormatDockerVolumePath(path string) string {
	return dockerutil.FormatDockerVolume(path)
}

// IsDockerInDocker detects if we're running inside a Docker container.
func IsDockerInDocker() bool {
	return dockerutil.IsDinD()
}

// IsDockerAvailable checks if Docker daemon is accessible.
func IsDockerAvailable() bool {
	return dockerutil.IsDockerAvailable()
}

// ExecutePostBuildSteps runs any post-build steps defined for the module.
// Returns non-zero exit code if any step fails.
func ExecutePostBuildSteps(moniker, workspaceRoot, outputDir string, logWriter io.Writer) int {
	// Post-build steps are expected to be defined at the module level
	// This function is a placeholder for future implementation
	return 0
}

// substituteVars replaces variable placeholders in a string.
func substituteVars(s string, vars map[string]string) string {
	result := s
	for k, v := range vars {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}

// CopyBuildOutput copies files from build output to target directory.
// If include patterns are specified, only matching files are copied.
// Exclude patterns filter out files from the copy.
func CopyBuildOutput(srcDir, dstDir string, include, exclude []string, logWriter io.Writer) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Walk source directory and copy files
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Skip log files and manifest
		if strings.HasSuffix(relPath, ".log") || relPath == "build.manifest.json" {
			return nil
		}

		// Convert to forward slashes for glob matching
		matchPath := filepath.ToSlash(relPath)

		// Check include patterns (if specified, file must match at least one)
		if len(include) > 0 {
			matched := false
			for _, pattern := range include {
				if m, matchErr := doublestar.Match(pattern, matchPath); matchErr == nil && m {
					matched = true
					break
				}
			}
			if !matched {
				if info.IsDir() {
					return nil // Continue into directory but don't copy
				}
				return nil
			}
		}

		// Check exclude patterns
		for _, pattern := range exclude {
			if m, matchErr := doublestar.Match(pattern, matchPath); matchErr == nil && m {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Calculate destination path
		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return CopyFile(path, dstPath)
	})
}
