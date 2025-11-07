// Godog BDD step definitions for src-cli features
//
// Features:
// - specs/src-cli/cli-invocation/
// - specs/src-cli/verify-configuration/
//
// Prerequisites:
// - Requires pre-built executable from "build module src-cli"
// - Executable location: out/src-cli/r2r-cli (or r2r-cli.exe on Windows)
package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// Test context holds state between steps
type testContext struct {
	commandOutput   string
	exitCode        int
	commandError    error
	executablePath  string
	testFolder      string
	originalDir     string
}

var ctx *testContext

// ============================================================================
// Common Steps
// ============================================================================

func iRun(cmdLine string) error {
	// Parse the command line
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Check if this is a command that should use the built executable
	if parts[0] == "simply" || parts[0] == "r2r" {
		// Verify executable exists
		if ctx.executablePath == "" {
			return fmt.Errorf("executable not found - please run 'build module src-cli' first")
		}

		// Replace command with executable path
		parts[0] = ctx.executablePath
	}

	return runCommandWithArgs(parts...)
}

func theExitCodeIs(expectedCode int) error {
	if ctx.exitCode != expectedCode {
		return fmt.Errorf("expected exit code %d, got %d. Output:\n%s",
			expectedCode, ctx.exitCode, ctx.commandOutput)
	}
	return nil
}

func iShouldSee(text string) error {
	if !strings.Contains(ctx.commandOutput, text) {
		return fmt.Errorf("expected output to contain '%s', got:\n%s", text, ctx.commandOutput)
	}
	return nil
}

func iShouldSeeOrOr(text1, text2, text3 string) error {
	if strings.Contains(ctx.commandOutput, text1) ||
		strings.Contains(ctx.commandOutput, text2) ||
		strings.Contains(ctx.commandOutput, text3) {
		return nil
	}
	return fmt.Errorf("expected output to contain one of '%s', '%s', or '%s', got:\n%s",
		text1, text2, text3, ctx.commandOutput)
}

func iShouldSeeOr(text1, text2 string) error {
	if strings.Contains(ctx.commandOutput, text1) ||
		strings.Contains(ctx.commandOutput, text2) {
		return nil
	}
	return fmt.Errorf("expected output to contain one of '%s' or '%s', got:\n%s",
		text1, text2, ctx.commandOutput)
}

func iShouldSeeVersionNumber() error {
	// Check for common version indicators
	if strings.Contains(ctx.commandOutput, "Version:") ||
		strings.Contains(ctx.commandOutput, "version") ||
		strings.Contains(ctx.commandOutput, "BuildTime:") {
		return nil
	}
	return fmt.Errorf("expected output to contain version information, got:\n%s", ctx.commandOutput)
}

// ============================================================================
// Test Folder Management Steps
// ============================================================================

func iCreateATestFolder(folderName string) error {
	// Create test folder in temp directory
	tempDir := os.TempDir()
	ctx.testFolder = filepath.Join(tempDir, folderName)

	// Remove if exists, then create fresh
	os.RemoveAll(ctx.testFolder)
	if err := os.MkdirAll(ctx.testFolder, 0755); err != nil {
		return fmt.Errorf("failed to create test folder: %w", err)
	}

	return nil
}

func iCreateAFolderInTheTestFolder(folderName string) error {
	if ctx.testFolder == "" {
		return fmt.Errorf("test folder not created yet")
	}

	folderPath := filepath.Join(ctx.testFolder, folderName)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return fmt.Errorf("failed to create folder '%s': %w", folderName, err)
	}

	return nil
}

func iChangeDirectoryToTheTestFolder() error {
	if ctx.testFolder == "" {
		return fmt.Errorf("test folder not created yet")
	}

	// Save original directory
	if ctx.originalDir == "" {
		var err error
		ctx.originalDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Change to test folder
	if err := os.Chdir(ctx.testFolder); err != nil {
		return fmt.Errorf("failed to change to test folder: %w", err)
	}

	return nil
}

// ============================================================================
// Build Steps
// ============================================================================

func iBuildTheCLIWith(buildCommand string) error {
	// The build is already done - this is just verifying the executable exists
	// We don't actually build during tests
	return nil
}

func theBuildSucceeds() error {
	// Verify the executable exists
	if ctx.executablePath == "" {
		return fmt.Errorf("executable not found - please run 'build module src-cli' first")
	}

	if _, err := os.Stat(ctx.executablePath); err != nil {
		return fmt.Errorf("executable not found at %s - please run 'build module src-cli' first", ctx.executablePath)
	}

	return nil
}

// ============================================================================
// Config File Steps
// ============================================================================

func noConfigFileExistsInTheTestFolder() error {
	// Ensure no config file exists
	configPath := filepath.Join(ctx.testFolder, "r2r-cli.yml")
	os.Remove(configPath)
	return nil
}

func iCreateATestConfigFileWithValidSettings(filename string) error {
	if ctx.testFolder == "" {
		return fmt.Errorf("test folder not created yet")
	}

	if ctx.executablePath == "" {
		return fmt.Errorf("executable not found - please run 'build module src-cli' first")
	}

	// Use the real CLI to generate a valid config file
	cmd := exec.Command(ctx.executablePath, "init")
	cmd.Dir = ctx.testFolder

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run 'r2r init': %w\nOutput: %s", err, string(output))
	}

	// Verify the config file was created
	configPath := filepath.Join(ctx.testFolder, "r2r-cli.yml")
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("config file not created by 'r2r init': %w", err)
	}

	return nil
}

func iCreateATestConfigFileWithInvalidSettings(filename string) error {
	if ctx.testFolder == "" {
		return fmt.Errorf("test folder not created yet")
	}

	// The CLI looks for r2r-cli.yml
	configPath := filepath.Join(ctx.testFolder, "r2r-cli.yml")
	invalidConfig := `this is not valid YAML: [[[broken`

	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	return nil
}

// ============================================================================
// CLI Execution Steps
// ============================================================================

func iRunTheBuiltCLIWith(args string) error {
	if ctx.executablePath == "" {
		return fmt.Errorf("executable not found - please run 'build module src-cli' first")
	}

	// Parse arguments
	argParts := strings.Fields(args)
	allArgs := append([]string{ctx.executablePath}, argParts...)

	return runCommandWithArgs(allArgs...)
}

// ============================================================================
// Setup/Initialization
// ============================================================================

func initializeContext() error {
	ctx = &testContext{}

	// Find the pre-built executable
	// Expected location: out/src-cli/r2r-cli (or r2r-cli.exe on Windows)
	workspaceRoot := filepath.Join("..", "..", "..")

	// Try both with and without .exe extension (works on all platforms)
	possiblePaths := []string{
		filepath.Join(workspaceRoot, "out", "src-cli", "r2r-cli.exe"), // Windows
		filepath.Join(workspaceRoot, "out", "src-cli", "r2r-cli"),     // Linux/Mac
	}

	for _, execPath := range possiblePaths {
		if _, err := os.Stat(execPath); err == nil {
			absPath, _ := filepath.Abs(execPath)
			ctx.executablePath = absPath
			break
		}
	}

	// If executable not found, tests will fail with helpful error message
	return nil
}

func cleanupContext() {
	// Restore original directory if changed
	if ctx.originalDir != "" {
		os.Chdir(ctx.originalDir)
	}

	// Clean up test folder
	if ctx.testFolder != "" {
		os.RemoveAll(ctx.testFolder)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func runCommandWithArgs(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	// Don't set cmd.Dir - we want to run from current directory

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
// Scenario Initialization
// ============================================================================

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		initializeContext()
		return ctx, nil
	})

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		cleanupContext()
		return ctx, nil
	})

	// Command execution steps
	sc.Step(`^I run "([^"]*)"$`, iRun)
	sc.Step(`^I run the built CLI with "([^"]*)"$`, iRunTheBuiltCLIWith)

	// Assertion steps
	sc.Step(`^the exit code is (\d+)$`, theExitCodeIs)
	sc.Step(`^I should see "([^"]*)"$`, iShouldSee)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)"$`, iShouldSeeOr)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)" or "([^"]*)"$`, iShouldSeeOrOr)
	sc.Step(`^I should see version number$`, iShouldSeeVersionNumber)

	// Test folder management steps
	sc.Step(`^I create a test folder "([^"]*)"$`, iCreateATestFolder)
	sc.Step(`^I create a "([^"]*)" folder in the test folder$`, iCreateAFolderInTheTestFolder)
	sc.Step(`^I change directory to the test folder$`, iChangeDirectoryToTheTestFolder)

	// Build steps
	sc.Step(`^I build the CLI with "([^"]*)"$`, iBuildTheCLIWith)
	sc.Step(`^the build succeeds$`, theBuildSucceeds)

	// Config file steps
	sc.Step(`^no config file exists in the test folder$`, noConfigFileExistsInTheTestFolder)
	sc.Step(`^I create a test config file "([^"]*)" with valid settings$`, iCreateATestConfigFileWithValidSettings)
	sc.Step(`^I create a test config file "([^"]*)" with invalid settings$`, iCreateATestConfigFileWithInvalidSettings)
}

func InitializeTestSuite(sc *godog.TestSuiteContext) {
	sc.BeforeSuite(func() {})
	sc.AfterSuite(func() {})
}
