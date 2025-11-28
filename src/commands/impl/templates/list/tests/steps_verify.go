package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	coretesting "github.com/ready-to-release/eac/src/core/testing"
)

// iRunCommand runs a templates command
func iRunCommand(cmdLine string) error {
	ctx, err := getActiveContext()
	if err != nil {
		return err
	}

	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Build the executable
	exePath := filepath.Join(ctx.WorkDir(), "eac-test.exe")
	// Go up to repo root, then build src/commands
	repoRoot := filepath.Join(ctx.TestDir(), "..", "..", "..")
	buildCmd := exec.Command("go", "build", "-o", exePath, "./src/commands")
	buildCmd.Dir = repoRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build CLI: %s\n%s", err, string(output))
	}
	defer os.Remove(exePath)

	// Run the command
	cmd := exec.Command(exePath, parts...)
	cmd.Dir = ctx.WorkDir()

	// Set environment variables for isolated testing
	if isolation := ctx.Isolation(); isolation != nil {
		if testIsolation, ok := isolation.(*coretesting.TestIsolation); ok {
			cmd.Env = testIsolation.AppendToEnvironment(os.Environ())
		}
	} else {
		cmd.Env = append(os.Environ(),
			"R2R_PWD="+ctx.WorkDir(),
			"R2R_REPO_ROOT="+ctx.WorkDir())
	}

	output, err := cmd.CombinedOutput()
	ctx.SetCommandOutput(string(output))

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			ctx.SetExitCode(exitErr.ExitCode())
			ctx.SetErrorOutput(string(output))
		} else {
			ctx.SetExitCode(1)
			ctx.SetErrorOutput(err.Error())
		}
	} else {
		ctx.SetExitCode(0)
		ctx.SetErrorOutput("")
	}

	return nil
}

// theCommandShouldSucceed verifies command succeeded
func theCommandShouldSucceed() error {
	ctx, err := getActiveContext()
	if err != nil {
		return err
	}

	if ctx.ExitCode() != 0 {
		return fmt.Errorf("expected exit code 0, got %d.\nOutput:\n%s\nError:\n%s",
			ctx.ExitCode(), ctx.CommandOutput(), ctx.ErrorOutput())
	}
	return nil
}

// theCommandShouldFail verifies command failed
func theCommandShouldFail() error {
	ctx, err := getActiveContext()
	if err != nil {
		return err
	}

	if ctx.ExitCode() == 0 {
		return fmt.Errorf("expected command to fail, but it succeeded.\nOutput:\n%s",
			ctx.CommandOutput())
	}
	return nil
}

// theOutputShouldContain verifies output contains text
func theOutputShouldContain(expectedText string) error {
	ctx, err := getActiveContext()
	if err != nil {
		return err
	}

	if !strings.Contains(ctx.CommandOutput(), expectedText) {
		return fmt.Errorf("output does not contain expected text '%s'.\nActual output:\n%s",
			expectedText, ctx.CommandOutput())
	}
	return nil
}

// theCommandShouldAttemptToCloneFrom verifies clone attempt
func theCommandShouldAttemptToCloneFrom(repoURL string) error {
	ctx, err := getActiveContext()
	if err != nil {
		return err
	}

	if !strings.Contains(ctx.CommandOutput(), repoURL) {
		return fmt.Errorf("output does not indicate cloning from '%s'.\nActual output:\n%s",
			repoURL, ctx.CommandOutput())
	}
	return nil
}

// debugFilesShouldExistIn verifies debug files exist
func debugFilesShouldExistIn(debugPath string) error {
	ctx, err := getActiveContext()
	if err != nil {
		return err
	}

	fullPath := filepath.Join(ctx.WorkDir(), debugPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("debug directory does not exist: %s", fullPath)
	}

	// Check for at least one debug file
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read debug directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("debug directory is empty: %s", fullPath)
	}

	return nil
}

// theFileShouldExist verifies a file exists
func theFileShouldExist(filePath string) error {
	ctx, err := getActiveContext()
	if err != nil {
		return err
	}

	fullPath := filepath.Join(ctx.WorkDir(), filePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", fullPath)
	} else if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}
	return nil
}

// theFileShouldContain verifies file contains expected text
func theFileShouldContain(filePath, expectedText string) error {
	ctx, err := getActiveContext()
	if err != nil {
		return err
	}

	fullPath := filepath.Join(ctx.WorkDir(), filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}

	if !strings.Contains(string(content), expectedText) {
		return fmt.Errorf("file %s does not contain expected text '%s'.\nActual content:\n%s",
			filePath, expectedText, string(content))
	}
	return nil
}
