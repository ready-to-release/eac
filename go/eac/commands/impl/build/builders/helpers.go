// helpers.go - Shared utility functions for builders
package builders

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ready-to-release/eac/go/eac/commands/impl/build/buildutil"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/platform"
)

// Logln writes a formatted string with platform-specific line ending to the writer
func Logln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+platform.LineEnding, args...)
}

// RunCommandWithLog executes a command in the specified directory
// Output is written to the provided writer
// Returns exit code (0 = success, non-zero = failure)
func RunCommandWithLog(dir string, logWriter io.Writer, name string, args ...string) int {
	// Use platform-aware command wrapper (handles .cmd files on Windows)
	wrappedName, wrappedArgs := platform.WrapCommand(name, args...)
	cmd := exec.Command(wrappedName, wrappedArgs...)
	cmd.Dir = dir
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		Logln(logWriter, "\nError: failed to execute command: %v", err)
		return 1
	}

	return 0
}

// RunCommandWithEnv executes a command with custom environment variables
func RunCommandWithEnv(dir string, logWriter io.Writer, env []string, name string, args ...string) int {
	// Use platform-aware command wrapper (handles .cmd files on Windows)
	wrappedName, wrappedArgs := platform.WrapCommand(name, args...)
	cmd := exec.Command(wrappedName, wrappedArgs...)
	cmd.Dir = dir
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	cmd.Env = append(os.Environ(), env...)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		Logln(logWriter, "\nError: failed to execute command: %v", err)
		return 1
	}

	return 0
}

// CopyFile copies a file from src to dst, preserving permissions
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
// On Windows, converts C:\path to /c/path for Docker compatibility
func FormatDockerVolumePath(path string) string {
	return buildutil.FormatDockerVolumePath(path)
}

// IsDockerInDocker detects if we're running inside a Docker container
func IsDockerInDocker() bool {
	return buildutil.IsDockerInDocker()
}

// IsDockerAvailable checks if Docker daemon is accessible
func IsDockerAvailable() bool {
	return buildutil.IsDockerAvailable()
}

// ExecutePostBuildSteps runs any post-build steps defined for the module type.
// Returns non-zero exit code if any step fails.
func ExecutePostBuildSteps(moduleType, moniker, workspaceRoot, outputDir string, logWriter io.Writer) int {
	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		return 0
	}

	steps := cfg.ModuleTypes.GetPostBuildSteps(moduleType)
	if len(steps) == 0 {
		return 0
	}

	Logln(logWriter, "")
	Logln(logWriter, "📋 Running post-build steps...")

	for i, step := range steps {
		// Substitute variables in target/script
		vars := map[string]string{
			"{moniker}":    moniker,
			"{root}":       workspaceRoot,
			"{output_dir}": outputDir,
		}

		switch step.Action {
		case config.PostBuildActionCopy:
			target := substituteVars(step.Target, vars)
			targetPath := filepath.Join(workspaceRoot, target)

			Logln(logWriter, "   [%d] Copy to %s", i+1, target)

			if err := CopyBuildOutput(outputDir, targetPath, step.Include, step.Exclude, logWriter); err != nil {
				Logln(logWriter, "   ❌ Copy failed: %v", err)
				return 1
			}

		case config.PostBuildActionScript:
			script := substituteVars(step.Script, vars)
			Logln(logWriter, "   [%d] Running: %s", i+1, script)

			// Run script in workspace root
			exitCode := RunCommandWithLog(workspaceRoot, logWriter, "sh", "-c", script)
			if exitCode != 0 {
				Logln(logWriter, "   ❌ Script failed with exit code %d", exitCode)
				return exitCode
			}
		}
	}

	Logln(logWriter, "   ✅ Post-build steps completed")
	return 0
}

// substituteVars replaces variable placeholders in a string
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
	if err := os.MkdirAll(dstDir, 0755); err != nil {
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
		if strings.HasSuffix(relPath, ".log") || relPath == ".manifest.json" {
			return nil
		}

		// Convert to forward slashes for glob matching
		matchPath := filepath.ToSlash(relPath)

		// Check include patterns (if specified, file must match at least one)
		if len(include) > 0 {
			matched := false
			for _, pattern := range include {
				if m, _ := doublestar.Match(pattern, matchPath); m {
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
			if m, _ := doublestar.Match(pattern, matchPath); m {
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
