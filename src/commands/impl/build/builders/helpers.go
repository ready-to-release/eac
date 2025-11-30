// helpers.go - Shared utility functions for builders
package builders

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/src/core/platform"
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
	if len(path) >= 2 && path[1] == ':' {
		driveLetter := strings.ToLower(string(path[0]))
		rest := strings.ReplaceAll(path[2:], "\\", "/")
		return "/" + driveLetter + rest
	}
	return strings.ReplaceAll(path, "\\", "/")
}
