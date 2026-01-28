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
	"github.com/ready-to-release/eac/go/eac/core/config"
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

	err := cmd.Run()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			exitCode := exitErr.ExitCode()
			Logln(logWriter, "[debug] command exited with code %d", exitCode)
			return exitCode
		}
		Logln(logWriter, "\nError: failed to execute command: %v", err)
		return 1
	}

	Logln(logWriter, "[debug] command completed successfully (exit code 0)")
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

// ExecutePostBuildSteps runs any post-build steps defined for the component.
// Parameters:
//   - moniker: Module moniker
//   - component: Component name (e.g., "typescript", "go")
//   - workspaceRoot: Absolute path to repository root
//   - outputDir: Absolute path to component build output (out/build/<module>/<component>/)
//   - logWriter: Writer for log output
//
// Returns 0 on success, non-zero on failure.
func ExecutePostBuildSteps(moniker, component, workspaceRoot, outputDir string, logWriter io.Writer) int {
	cfg := config.Global()
	if cfg == nil || cfg.Repository == nil {
		return 0
	}

	// Get module configuration
	module := cfg.Repository.GetByMoniker(moniker)
	if module == nil {
		return 0
	}

	// Get component entry
	compEntry, ok := module.Components[component]
	if !ok || compEntry == nil || compEntry.Build == nil || compEntry.Build.PostBuild == nil {
		return 0
	}

	postBuild := compEntry.Build.PostBuild

	// Execute copy_to if configured
	if postBuild.CopyTo != "" {
		targetPath := filepath.Join(workspaceRoot, postBuild.CopyTo)

		// Validate target is within workspace (security check)
		relPath, err := filepath.Rel(workspaceRoot, targetPath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			Logln(logWriter, "Error: post-build copy_to path must be within workspace")
			return 1
		}

		if err := executeCopyTo(outputDir, targetPath, logWriter); err != nil {
			Logln(logWriter, "Error: post-build copy failed: %v", err)
			return 1
		}
		Logln(logWriter, "Post-build: copied output to %s", postBuild.CopyTo)
	}

	return 0
}

// executeCopyTo copies build output to the target directory.
// The target directory is cleaned first to avoid stale files.
func executeCopyTo(srcDir, dstDir string, logWriter io.Writer) error {
	// Clean the destination first to avoid stale files
	if err := os.RemoveAll(dstDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean target directory: %w", err)
	}

	// Copy using existing CopyBuildOutput helper
	// Skip log files and manifest, copy everything else
	return CopyBuildOutput(srcDir, dstDir, nil, nil, logWriter)
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
