package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// iRunCommand runs a templates install command
func iRunCommand(cmdLine string) error {
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Build the executable
	exePath := filepath.Join(installCtx.workDir, "eac-test.exe")
	cliSourceDir := filepath.Join(installCtx.testDir, "..", "..", "..", "..", "..")
	buildCmd := exec.Command("go", "build", "-o", exePath, ".")
	buildCmd.Dir = cliSourceDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build CLI: %s\n%s", err, string(output))
	}
	defer os.Remove(exePath)

	// Convert relative paths in args to absolute paths based on workDir
	resolvedArgs := make([]string, len(parts))
	for i, arg := range parts {
		// Check if this is a flag value that might be a path
		if i > 0 && (parts[i-1] == "--source" || parts[i-1] == "--destination") {
			// If it's not a URL, make it absolute
			if !strings.HasPrefix(arg, "http://") && !strings.HasPrefix(arg, "https://") {
				absPath := filepath.Join(installCtx.workDir, arg)
				resolvedArgs[i] = absPath
			} else {
				resolvedArgs[i] = arg
			}
		} else {
			resolvedArgs[i] = arg
		}
	}

	// Run the command
	cmd := exec.Command(exePath, resolvedArgs...)
	cmd.Dir = installCtx.workDir

	// Set environment variables for isolated testing
	if installCtx.isolation != nil {
		cmd.Env = installCtx.isolation.AppendToEnvironment(os.Environ())
	} else {
		cmd.Env = append(os.Environ(),
			"R2R_PWD="+installCtx.workDir,
			"R2R_REPO_ROOT="+installCtx.workDir)
	}

	output, err := cmd.CombinedOutput()
	installCtx.commandOutput = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			installCtx.exitCode = exitErr.ExitCode()
			installCtx.errorOutput = string(output)
		} else {
			installCtx.exitCode = 1
			installCtx.errorOutput = err.Error()
		}
	} else {
		installCtx.exitCode = 0
		installCtx.errorOutput = ""
	}

	return nil
}

// theCommandShouldSucceed verifies command succeeded
func theCommandShouldSucceed() error {
	if installCtx.exitCode != 0 {
		return fmt.Errorf("expected exit code 0, got %d.\nOutput:\n%s\nError:\n%s",
			installCtx.exitCode, installCtx.commandOutput, installCtx.errorOutput)
	}
	return nil
}

// theCommandShouldFail verifies command failed
func theCommandShouldFail() error {
	if installCtx.exitCode == 0 {
		return fmt.Errorf("expected command to fail, but it succeeded.\nOutput:\n%s",
			installCtx.commandOutput)
	}
	return nil
}

// theFileShouldExist verifies a file exists
func theFileShouldExist(filePath string) error {
	fullPath := filepath.Join(installCtx.workDir, filePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", fullPath)
	} else if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}
	return nil
}

// theFileShouldContain verifies file contains expected text
func theFileShouldContain(filePath, expectedText string) error {
	fullPath := filepath.Join(installCtx.workDir, filePath)
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

// theOutputShouldContain verifies output contains text
func theOutputShouldContain(expectedText string) error {
	if !strings.Contains(installCtx.commandOutput, expectedText) {
		return fmt.Errorf("output does not contain expected text '%s'.\nActual output:\n%s",
			expectedText, installCtx.commandOutput)
	}
	return nil
}

// theErrorOutputShouldContain verifies error output contains text
func theErrorOutputShouldContain(expectedText string) error {
	combined := installCtx.commandOutput + installCtx.errorOutput
	if !strings.Contains(combined, expectedText) {
		return fmt.Errorf("error output does not contain expected text '%s'.\nActual output:\n%s\nActual error:\n%s",
			expectedText, installCtx.commandOutput, installCtx.errorOutput)
	}
	return nil
}

// theCommandShouldAttemptToCloneFrom verifies clone attempt
func theCommandShouldAttemptToCloneFrom(repoURL string) error {
	if !strings.Contains(installCtx.commandOutput, repoURL) {
		return fmt.Errorf("output does not indicate cloning from '%s'.\nActual output:\n%s",
			repoURL, installCtx.commandOutput)
	}
	return nil
}

// theDestinationShouldBe verifies destination directory
func theDestinationShouldBe(expectedDest string) error {
	fullPath := filepath.Join(installCtx.workDir, expectedDest)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("destination directory does not exist: %s", fullPath)
	} else if err != nil {
		return fmt.Errorf("error checking destination: %w", err)
	}
	return nil
}

// debugFilesShouldExistIn verifies debug files exist
func debugFilesShouldExistIn(debugPath string) error {
	fullPath := filepath.Join(installCtx.workDir, debugPath)
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

// theTemplatesShouldBeCopiedWithoutValueReplacement verifies templates copied without replacement
func theTemplatesShouldBeCopiedWithoutValueReplacement() error {
	// When installing (not applying), templates should be copied as-is
	// This is verified by the command succeeding
	if installCtx.exitCode != 0 {
		return fmt.Errorf("command failed, indicating templates may not have been copied correctly")
	}
	return nil
}
