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
	commandOutput  string
	exitCode       int
	commandError   error
	executablePath string
	testFolderPath string // For integration tests with temp folders
	currentDir     string // Track current working directory
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
// CLI Integration Test Steps (for verify-configuration feature)
// ============================================================================

func iCreateATestFolder(folderName string) error {
	// Create temp folder in OS temp directory
	tempDir := os.TempDir()
	testPath := filepath.Join(tempDir, folderName)

	// Clean up if it already exists
	os.RemoveAll(testPath)

	// Create fresh folder
	if err := os.MkdirAll(testPath, 0755); err != nil {
		return fmt.Errorf("failed to create test folder: %w", err)
	}

	ctx.testFolderPath = testPath
	return nil
}

func iCreateAFolderInTheTestFolder(folderName string) error {
	if ctx.testFolderPath == "" {
		return fmt.Errorf("test folder not created yet")
	}

	folderPath := filepath.Join(ctx.testFolderPath, folderName)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	return nil
}

func iBuildTheCLIWith(buildCommand string) error {
	// The CLI is already built via @dep:internal-src-cli
	// This step just verifies it exists
	if ctx.executablePath == "" {
		return fmt.Errorf("CLI executable not found - @dep:internal-src-cli should have verified this")
	}
	return nil
}

func theBuildSucceeds() error {
	// Since we're using pre-built executable, just verify it's accessible
	if ctx.executablePath == "" {
		return fmt.Errorf("CLI executable not available")
	}

	// Verify executable is actually executable
	info, err := os.Stat(ctx.executablePath)
	if err != nil {
		return fmt.Errorf("cannot access executable: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("executable path is a directory")
	}

	return nil
}

func iChangeDirectoryToTheTestFolder() error {
	if ctx.testFolderPath == "" {
		return fmt.Errorf("test folder not created yet")
	}

	// Save current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	ctx.currentDir = cwd

	// Change to test folder
	if err := os.Chdir(ctx.testFolderPath); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	return nil
}

func noConfigFileExistsInTheTestFolder() error {
	if ctx.testFolderPath == "" {
		return fmt.Errorf("test folder not created yet")
	}

	// Ensure no config file exists (the CLI looks for "r2r-cli.yml")
	configPath := filepath.Join(ctx.testFolderPath, "r2r-cli.yml")
	os.Remove(configPath) // Ignore error - file might not exist

	return nil
}

func iCreateATestConfigFileWithValidSettings(filename string) error {
	if ctx.testFolderPath == "" {
		return fmt.Errorf("test folder not created yet")
	}

	// Create a minimal valid config
	// The CLI looks for "r2r-cli.yml", so use that name regardless of parameter
	configContent := `# Valid R2R CLI configuration
version: "1.0"
project:
  name: "test-project"
extensions: []
`

	configPath := filepath.Join(ctx.testFolderPath, "r2r-cli.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func iCreateATestConfigFileWithInvalidSettings(filename string) error {
	if ctx.testFolderPath == "" {
		return fmt.Errorf("test folder not created yet")
	}

	// Create an invalid config (malformed YAML)
	// The CLI looks for "r2r-cli.yml", so use that name regardless of parameter
	configContent := `# Invalid R2R CLI configuration
version: "1.0"
project:
  name: [this is invalid yaml syntax
`

	configPath := filepath.Join(ctx.testFolderPath, "r2r-cli.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func iRunTheBuiltCLIWith(args string) error {
	if ctx.executablePath == "" {
		return fmt.Errorf("executable not found")
	}

	// Parse arguments
	argsList := strings.Fields(args)

	// Prepend executable path
	cmdArgs := append([]string{ctx.executablePath}, argsList...)

	return runCommandWithArgs(cmdArgs...)
}

func iShouldSeeVersionNumber() error {
	// Look for version pattern like "1.0.0" or "v1.0.0" or "version 1.0.0"
	output := strings.ToLower(ctx.commandOutput)

	// Check for common version indicators
	hasVersion := strings.Contains(output, "version") ||
		strings.Contains(output, "v0.") ||
		strings.Contains(output, "v1.") ||
		strings.Contains(output, "v2.") ||
		strings.Contains(output, "0.0.") ||
		strings.Contains(output, "1.0.") ||
		strings.Contains(output, "2.0.")

	if !hasVersion {
		return fmt.Errorf("expected version number in output, got:\n%s", ctx.commandOutput)
	}

	return nil
}

func iShouldSeeOr(text1, text2 string) error {
	if strings.Contains(ctx.commandOutput, text1) ||
		strings.Contains(ctx.commandOutput, text2) {
		return nil
	}
	return fmt.Errorf("expected output to contain '%s' or '%s', got:\n%s",
		text1, text2, ctx.commandOutput)
}

// ============================================================================
// Scenario Initialization
// ============================================================================

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		initializeContext()
		return ctx, nil
	})

	sc.After(func(context.Context, *godog.Scenario, error) (context.Context, error) {
		// Cleanup: restore original directory if changed
		if ctx != nil && ctx.currentDir != "" {
			os.Chdir(ctx.currentDir)
		}

		// Cleanup: remove test folder if created
		if ctx != nil && ctx.testFolderPath != "" {
			os.RemoveAll(ctx.testFolderPath)
		}

		return nil, nil
	})

	// Common steps
	sc.Step(`^I run "([^"]*)"$`, iRun)
	sc.Step(`^the exit code is (\d+)$`, theExitCodeIs)
	sc.Step(`^I should see "([^"]*)"$`, iShouldSee)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)" or "([^"]*)"$`, iShouldSeeOrOr)

	// CLI integration test steps
	sc.Step(`^I create a test folder "([^"]*)"$`, iCreateATestFolder)
	sc.Step(`^I create a "([^"]*)" folder in the test folder$`, iCreateAFolderInTheTestFolder)
	sc.Step(`^I build the CLI with "([^"]*)"$`, iBuildTheCLIWith)
	sc.Step(`^the build succeeds$`, theBuildSucceeds)
	sc.Step(`^I change directory to the test folder$`, iChangeDirectoryToTheTestFolder)
	sc.Step(`^no config file exists in the test folder$`, noConfigFileExistsInTheTestFolder)
	sc.Step(`^I create a test config file "([^"]*)" with valid settings$`, iCreateATestConfigFileWithValidSettings)
	sc.Step(`^I create a test config file "([^"]*)" with invalid settings$`, iCreateATestConfigFileWithInvalidSettings)
	sc.Step(`^I run the built CLI with "([^"]*)"$`, iRunTheBuiltCLIWith)
	sc.Step(`^I should see version number$`, iShouldSeeVersionNumber)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)"$`, iShouldSeeOr)
}

func InitializeTestSuite(sc *godog.TestSuiteContext) {
	sc.BeforeSuite(func() {})
	sc.AfterSuite(func() {})
}
