// Package scriptsinstaller contains godog step implementations for specs/scripts-cli-installer.
//
// Features:
// - specs/scripts-cli-installer/cli-installation/
//
// These tests invoke the installer scripts and verify they work correctly.
// Platform-specific scenarios use runtime detection to skip on non-matching platforms.
package scriptscliinstaller

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/specs/internal"
)

// installerContext holds state between steps for installer tests
type installerContext struct {
	sharedCtx      *internal.TestContext
	scriptsRoot    string
	tempInstallDir string
}

var instCtx *installerContext

// RegisterSteps registers all scripts-cli-installer specific step definitions.
func RegisterSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	instCtx = &installerContext{sharedCtx: ctx}

	// Initialize scripts root and temp install dir
	initializeInstallerContext()

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

func initializeInstallerContext() {
	// Find scripts directory relative to test location
	workspaceRoot := filepath.Join("..", "..", "..", "..")
	instCtx.scriptsRoot = filepath.Join(workspaceRoot, "scripts")

	if _, err := os.Stat(instCtx.scriptsRoot); os.IsNotExist(err) {
		absRoot, _ := filepath.Abs(instCtx.scriptsRoot)
		instCtx.scriptsRoot = absRoot
	}

	// Create isolated temp directory for this test scenario
	tempDir, err := os.MkdirTemp("", "r2r-installer-test-*")
	if err == nil {
		instCtx.tempInstallDir = tempDir
	}
}

func cleanupInstallerContext() {
	if instCtx != nil && instCtx.tempInstallDir != "" {
		os.RemoveAll(instCtx.tempInstallDir)
	}
}

// ============================================================================
// Platform Detection Steps
// ============================================================================

func iAmOnWindowsWithPowerShell() error {
	if runtime.GOOS != "windows" {
		return godog.ErrPending
	}
	return nil
}

func iAmOnLinuxWithBashAndCurl() error {
	if runtime.GOOS != "linux" {
		return godog.ErrPending
	}
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
	return nil
}

// ============================================================================
// Action Steps
// ============================================================================

func iRunThePowerShellInstaller() error {
	scriptPath := filepath.Join(instCtx.scriptsRoot, "pwsh", "cli", "install.ps1")

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath,
		"-InstallDir", instCtx.tempInstallDir)
	output, err := cmd.CombinedOutput()
	instCtx.sharedCtx.CommandOutput = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			instCtx.sharedCtx.ExitCode = exitErr.ExitCode()
		} else {
			instCtx.sharedCtx.ExitCode = 1
		}
	} else {
		instCtx.sharedCtx.ExitCode = 0
	}

	return nil
}

func iRunThePowerShellInstallerWithArgs(args string) error {
	scriptPath := filepath.Join(instCtx.scriptsRoot, "pwsh", "cli", "install.ps1")

	cmdArgs := []string{"-ExecutionPolicy", "Bypass", "-File", scriptPath, "-InstallDir", instCtx.tempInstallDir}
	cmdArgs = append(cmdArgs, strings.Fields(args)...)

	cmd := exec.Command("powershell", cmdArgs...)
	output, err := cmd.CombinedOutput()
	instCtx.sharedCtx.CommandOutput = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			instCtx.sharedCtx.ExitCode = exitErr.ExitCode()
		} else {
			instCtx.sharedCtx.ExitCode = 1
		}
	} else {
		instCtx.sharedCtx.ExitCode = 0
	}

	return nil
}

func iRunTheBashInstaller() error {
	scriptPath := filepath.Join(instCtx.scriptsRoot, "sh", "cli", "install.sh")

	cmd := exec.Command("bash", scriptPath, "--install-dir", instCtx.tempInstallDir)
	output, err := cmd.CombinedOutput()
	instCtx.sharedCtx.CommandOutput = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			instCtx.sharedCtx.ExitCode = exitErr.ExitCode()
		} else {
			instCtx.sharedCtx.ExitCode = 1
		}
	} else {
		instCtx.sharedCtx.ExitCode = 0
	}

	return nil
}

func iRunTheBashInstallerWithArgs(args string) error {
	scriptPath := filepath.Join(instCtx.scriptsRoot, "sh", "cli", "install.sh")

	cmdArgs := []string{scriptPath, "--install-dir", instCtx.tempInstallDir}
	cmdArgs = append(cmdArgs, strings.Fields(args)...)

	cmd := exec.Command("bash", cmdArgs...)
	output, err := cmd.CombinedOutput()
	instCtx.sharedCtx.CommandOutput = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			instCtx.sharedCtx.ExitCode = exitErr.ExitCode()
		} else {
			instCtx.sharedCtx.ExitCode = 1
		}
	} else {
		instCtx.sharedCtx.ExitCode = 0
	}

	return nil
}

// ============================================================================
// Assertion Steps
// ============================================================================

func theLatestVersionIsFetchedFromGitHubAPI() error {
	output := strings.ToLower(instCtx.sharedCtx.CommandOutput)
	if !strings.Contains(output, "version") && !strings.Contains(output, "fetching") && !strings.Contains(output, "latest") {
		return fmt.Errorf("expected output to mention version fetching, got:\n%s", instCtx.sharedCtx.CommandOutput)
	}
	return nil
}

func theBinaryIsInstalledTo(expectedPath string) error {
	var binaryName string
	if runtime.GOOS == "windows" {
		binaryName = "r2r.exe"
	} else {
		binaryName = "r2r"
	}

	actualPath := filepath.Join(instCtx.tempInstallDir, binaryName)

	if _, err := os.Stat(actualPath); os.IsNotExist(err) {
		return fmt.Errorf("binary not found at %s (isolated test dir)\n\nInstaller output:\n%s", actualPath, instCtx.sharedCtx.CommandOutput)
	}
	return nil
}

func theInstallationIsVerifiedByRunningVersion() error {
	output := strings.ToLower(instCtx.sharedCtx.CommandOutput)
	if !strings.Contains(output, "verified") && !strings.Contains(output, "version") && !strings.Contains(output, "successfully") {
		return fmt.Errorf("expected installation verification in output, got:\n%s", instCtx.sharedCtx.CommandOutput)
	}
	return nil
}

func theInstallerDisplaysAnErrorAboutTheFailedDownload() error {
	output := strings.ToLower(instCtx.sharedCtx.CommandOutput)
	if !strings.Contains(output, "failed") && !strings.Contains(output, "error") && !strings.Contains(output, "not found") {
		return fmt.Errorf("expected error message about failed download, got:\n%s", instCtx.sharedCtx.CommandOutput)
	}
	return nil
}

func exitsWithANonZeroExitCode() error {
	if instCtx.sharedCtx.ExitCode == 0 {
		return fmt.Errorf("expected non-zero exit code, got 0")
	}
	return nil
}
