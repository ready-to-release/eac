// Godog glue for src-cli-installers features
//
// Features:
// - specs/src-cli-installers/cli-installation/
//
// These tests invoke the installer scripts and verify they work correctly.
// Platform-specific scenarios use build tags to run only on the appropriate OS.
package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cucumber/godog"
)

// testContext holds state between steps
type testContext struct {
	commandOutput  string
	exitCode       int
	commandError   error
	scriptsRoot    string // Path to scripts directory
	tempInstallDir string // Isolated temp directory for installation
}

var ctx *testContext

// ============================================================================
// Platform Detection Steps
// ============================================================================

func iAmOnWindowsWithPowerShell() error {
	if runtime.GOOS != "windows" {
		return godog.ErrPending // Skip on non-Windows
	}
	return nil
}

func iAmOnLinuxWithBashAndCurl() error {
	if runtime.GOOS != "linux" {
		return godog.ErrPending // Skip on non-Linux
	}
	// Verify bash and curl are available
	if _, err := exec.LookPath("bash"); err != nil {
		return fmt.Errorf("bash not found: %w", err)
	}
	if _, err := exec.LookPath("curl"); err != nil {
		return fmt.Errorf("curl not found: %w", err)
	}
	return nil
}

func iAmOnWindows() error {
	if runtime.GOOS != "windows" {
		return godog.ErrPending
	}
	return nil
}

func iAmOnLinux() error {
	if runtime.GOOS != "linux" {
		return godog.ErrPending
	}
	return nil
}

// ============================================================================
// Background Steps
// ============================================================================

func theGitHubRepositoryHasReleasesAvailable(repo string) error {
	// This is a precondition - we assume GitHub is accessible
	// In a real test, we might verify connectivity
	return nil
}

// ============================================================================
// Action Steps
// ============================================================================

func iRunThePowerShellInstaller() error {
	scriptPath := filepath.Join(ctx.scriptsRoot, "pwsh", "cli", "install.ps1")

	// Run PowerShell with the script, using isolated temp install directory
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath,
		"-InstallDir", ctx.tempInstallDir)
	output, err := cmd.CombinedOutput()
	ctx.commandOutput = string(output)
	ctx.commandError = err

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			ctx.exitCode = exitErr.ExitCode()
		} else {
			ctx.exitCode = 1
		}
	} else {
		ctx.exitCode = 0
	}

	return nil
}

func iRunThePowerShellInstallerWithArgs(args string) error {
	scriptPath := filepath.Join(ctx.scriptsRoot, "pwsh", "cli", "install.ps1")

	// Parse args and build command, always include isolated install dir
	cmdArgs := []string{"-ExecutionPolicy", "Bypass", "-File", scriptPath, "-InstallDir", ctx.tempInstallDir}
	cmdArgs = append(cmdArgs, strings.Fields(args)...)

	cmd := exec.Command("powershell", cmdArgs...)
	output, err := cmd.CombinedOutput()
	ctx.commandOutput = string(output)
	ctx.commandError = err

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			ctx.exitCode = exitErr.ExitCode()
		} else {
			ctx.exitCode = 1
		}
	} else {
		ctx.exitCode = 0
	}

	return nil
}

func iRunTheBashInstaller() error {
	scriptPath := filepath.Join(ctx.scriptsRoot, "sh", "cli", "install.sh")

	// Run bash with isolated install directory
	cmd := exec.Command("bash", scriptPath, "--install-dir", ctx.tempInstallDir)
	output, err := cmd.CombinedOutput()
	ctx.commandOutput = string(output)
	ctx.commandError = err

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			ctx.exitCode = exitErr.ExitCode()
		} else {
			ctx.exitCode = 1
		}
	} else {
		ctx.exitCode = 0
	}

	return nil
}

func iRunTheBashInstallerWithArgs(args string) error {
	scriptPath := filepath.Join(ctx.scriptsRoot, "sh", "cli", "install.sh")

	// Build command with args, always include isolated install dir
	cmdArgs := []string{scriptPath, "--install-dir", ctx.tempInstallDir}
	cmdArgs = append(cmdArgs, strings.Fields(args)...)

	cmd := exec.Command("bash", cmdArgs...)
	output, err := cmd.CombinedOutput()
	ctx.commandOutput = string(output)
	ctx.commandError = err

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			ctx.exitCode = exitErr.ExitCode()
		} else {
			ctx.exitCode = 1
		}
	} else {
		ctx.exitCode = 0
	}

	return nil
}

// ============================================================================
// Assertion Steps
// ============================================================================

func theLatestVersionIsFetchedFromGitHubAPI() error {
	// Check output mentions fetching or version
	output := strings.ToLower(ctx.commandOutput)
	if !strings.Contains(output, "version") && !strings.Contains(output, "fetching") && !strings.Contains(output, "latest") {
		return fmt.Errorf("expected output to mention version fetching, got:\n%s", ctx.commandOutput)
	}
	return nil
}

func theBinaryIsInstalledTo(expectedPath string) error {
	// The spec uses placeholder paths like %LOCALAPPDATA%\r2r\r2r.exe or $HOME/.local/bin/r2r
	// In isolated tests, we always install to ctx.tempInstallDir, so check there instead
	var binaryName string
	if runtime.GOOS == "windows" {
		binaryName = "r2r.exe"
	} else {
		binaryName = "r2r"
	}

	actualPath := filepath.Join(ctx.tempInstallDir, binaryName)

	if _, err := os.Stat(actualPath); os.IsNotExist(err) {
		return fmt.Errorf("binary not found at %s (isolated test dir)\n\nInstaller output:\n%s", actualPath, ctx.commandOutput)
	}
	return nil
}


func theInstallationIsVerifiedByRunningVersion() error {
	// The installer should have run --version as part of verification
	output := strings.ToLower(ctx.commandOutput)
	if !strings.Contains(output, "verified") && !strings.Contains(output, "version") && !strings.Contains(output, "successfully") {
		return fmt.Errorf("expected installation verification in output, got:\n%s", ctx.commandOutput)
	}
	return nil
}

func theInstallerDisplaysAnErrorAboutTheFailedDownload() error {
	output := strings.ToLower(ctx.commandOutput)
	if !strings.Contains(output, "failed") && !strings.Contains(output, "error") && !strings.Contains(output, "not found") {
		return fmt.Errorf("expected error message about failed download, got:\n%s", ctx.commandOutput)
	}
	return nil
}

func exitsWithANonZeroExitCode() error {
	if ctx.exitCode == 0 {
		return fmt.Errorf("expected non-zero exit code, got 0")
	}
	return nil
}

// ============================================================================
// Setup/Initialization
// ============================================================================

func initializeContext() error {
	ctx = &testContext{}

	// Find scripts directory relative to test location
	// Tests run from src/cli-installers/tests, scripts are at ../../../scripts
	workspaceRoot := filepath.Join("..", "..", "..")
	ctx.scriptsRoot = filepath.Join(workspaceRoot, "scripts")

	// Verify scripts directory exists
	if _, err := os.Stat(ctx.scriptsRoot); os.IsNotExist(err) {
		// Try absolute path resolution
		absRoot, _ := filepath.Abs(ctx.scriptsRoot)
		ctx.scriptsRoot = absRoot
	}

	// Create isolated temp directory for this test scenario
	tempDir, err := os.MkdirTemp("", "r2r-installer-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp install directory: %w", err)
	}
	ctx.tempInstallDir = tempDir

	return nil
}

func cleanupContext() {
	if ctx != nil && ctx.tempInstallDir != "" {
		os.RemoveAll(ctx.tempInstallDir)
	}
}

// ============================================================================
// Scenario Initialization
// ============================================================================

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		if err := initializeContext(); err != nil {
			return ctx, err
		}
		return ctx, nil
	})

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		cleanupContext()
		return ctx, nil
	})

	// Background
	sc.Step(`^the GitHub repository "([^"]*)" has releases available$`, theGitHubRepositoryHasReleasesAvailable)

	// Platform detection
	sc.Step(`^I am on Windows with PowerShell 5\.1 or later$`, iAmOnWindowsWithPowerShell)
	sc.Step(`^I am on Linux with bash and curl available$`, iAmOnLinuxWithBashAndCurl)
	sc.Step(`^I am on Windows$`, iAmOnWindows)
	sc.Step(`^I am on Linux$`, iAmOnLinux)

	// Actions
	sc.Step(`^I run the PowerShell installer$`, iRunThePowerShellInstaller)
	sc.Step(`^I run the PowerShell installer with "([^"]*)"$`, iRunThePowerShellInstallerWithArgs)
	sc.Step(`^I run the bash installer$`, iRunTheBashInstaller)
	sc.Step(`^I run the bash installer with "([^"]*)"$`, iRunTheBashInstallerWithArgs)

	// Assertions
	sc.Step(`^the latest version is fetched from GitHub API$`, theLatestVersionIsFetchedFromGitHubAPI)
	sc.Step(`^the binary is installed to "([^"]*)"$`, theBinaryIsInstalledTo)
	sc.Step(`^the installation is verified by running "([^"]*)"$`, theInstallationIsVerifiedByRunningVersion)
	sc.Step(`^the installer displays an error about the failed download$`, theInstallerDisplaysAnErrorAboutTheFailedDownload)
	sc.Step(`^exits with a non-zero exit code$`, exitsWithANonZeroExitCode)
}
