// Package srccli contains godog step implementations for specs/r2r-cli.
//
// Features:
// - specs/r2r-cli/cli-invocation/
// - specs/r2r-cli/verify-configuration/
//
// Prerequisites:
// - Requires pre-built executable from "build module r2r-cli"
// - Executable location: out/build/r2r-cli/windows-r2r-cli.exe (or linux-r2r-cli, darwin-r2r-cli)
package srccli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// cliContext holds state between steps for CLI tests.
type cliContext struct {
	sharedCtx      *internal.TestContext
	executablePath string
	testFolderPath string
	currentDir     string
}

var cliCtx *cliContext

// RegisterSteps registers all r2r-cli specific step definitions.
func RegisterSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	cliCtx = &cliContext{sharedCtx: ctx}

	// Initialize executable path
	initializeExecutablePath()

	// Common CLI steps
	sc.Step(`^I run "([^"]*)"$`, iRun)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)" or "([^"]*)"$`, iShouldSeeOrOr)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)"$`, iShouldSeeOr)
	sc.Step(`^I should see version number$`, iShouldSeeVersionNumber)

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
}

func initializeExecutablePath() {
	// Find the pre-built executable
	// Expected location: out/build/r2r-cli/<platform>-r2r-cli (or <platform>-r2r-cli.exe on Windows)
	workspaceRoot := filepath.Join("..", "..", "..", "..")

	possiblePaths := []string{
		filepath.Join(workspaceRoot, "out", "build", "r2r-cli", "windows-r2r-cli.exe"),
		filepath.Join(workspaceRoot, "out", "build", "r2r-cli", "linux-r2r-cli"),
		filepath.Join(workspaceRoot, "out", "build", "r2r-cli", "darwin-r2r-cli"),
		filepath.Join(workspaceRoot, "out", "build", "r2r-cli", "r2r-cli.exe"),
		filepath.Join(workspaceRoot, "out", "build", "r2r-cli", "r2r-cli"),
	}

	for _, execPath := range possiblePaths {
		if _, err := os.Stat(execPath); err == nil {
			absPath, err := filepath.Abs(execPath)
			if err == nil {
				cliCtx.executablePath = absPath
			}
			break
		}
	}
}

// ============================================================================
// Common Steps
// ============================================================================

func iRun(cmdLine string) error {
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	if parts[0] == "r2r" {
		if cliCtx.executablePath == "" {
			return fmt.Errorf("executable not found - please run 'build module r2r-cli' first")
		}
		parts[0] = cliCtx.executablePath
	}

	return runCommandWithArgs(parts...)
}

func iShouldSeeOrOr(text1, text2, text3 string) error {
	if strings.Contains(cliCtx.sharedCtx.CommandOutput, text1) ||
		strings.Contains(cliCtx.sharedCtx.CommandOutput, text2) ||
		strings.Contains(cliCtx.sharedCtx.CommandOutput, text3) {
		return nil
	}
	return fmt.Errorf("expected output to contain one of '%s', '%s', or '%s', got:\n%s",
		text1, text2, text3, cliCtx.sharedCtx.CommandOutput)
}

func iShouldSeeOr(text1, text2 string) error {
	if strings.Contains(cliCtx.sharedCtx.CommandOutput, text1) ||
		strings.Contains(cliCtx.sharedCtx.CommandOutput, text2) {
		return nil
	}
	return fmt.Errorf("expected output to contain '%s' or '%s', got:\n%s",
		text1, text2, cliCtx.sharedCtx.CommandOutput)
}

func iShouldSeeVersionNumber() error {
	output := strings.ToLower(cliCtx.sharedCtx.CommandOutput)

	hasVersion := strings.Contains(output, "version") ||
		strings.Contains(output, "v0.") ||
		strings.Contains(output, "v1.") ||
		strings.Contains(output, "v2.") ||
		strings.Contains(output, "0.0.") ||
		strings.Contains(output, "1.0.") ||
		strings.Contains(output, "2.0.")

	if !hasVersion {
		return fmt.Errorf("expected version number in output, got:\n%s", cliCtx.sharedCtx.CommandOutput)
	}

	return nil
}

// ============================================================================
// CLI Integration Test Steps
// ============================================================================

func iCreateATestFolder(folderName string) error {
	tempDir := os.TempDir()
	testPath := filepath.Join(tempDir, folderName)

	os.RemoveAll(testPath)

	if err := os.MkdirAll(testPath, 0o750); err != nil {
		return fmt.Errorf("failed to create test folder: %w", err)
	}

	cliCtx.testFolderPath = testPath
	return nil
}

func iCreateAFolderInTheTestFolder(folderName string) error {
	if cliCtx.testFolderPath == "" {
		return fmt.Errorf("test folder not created yet")
	}

	folderPath := filepath.Join(cliCtx.testFolderPath, folderName)
	if err := os.MkdirAll(folderPath, 0o750); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	return nil
}

func iBuildTheCLIWith(buildCommand string) error {
	if cliCtx.executablePath == "" {
		return fmt.Errorf("CLI executable not found - @depm:r2r-cli should have verified this")
	}
	return nil
}

func theBuildSucceeds() error {
	if cliCtx.executablePath == "" {
		return fmt.Errorf("CLI executable not available")
	}

	info, err := os.Stat(cliCtx.executablePath)
	if err != nil {
		return fmt.Errorf("cannot access executable: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("executable path is a directory")
	}

	return nil
}

func iChangeDirectoryToTheTestFolder() error {
	if cliCtx.testFolderPath == "" {
		return fmt.Errorf("test folder not created yet")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	cliCtx.currentDir = cwd

	if err := os.Chdir(cliCtx.testFolderPath); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	return nil
}

func noConfigFileExistsInTheTestFolder() error {
	if cliCtx.testFolderPath == "" {
		return fmt.Errorf("test folder not created yet")
	}

	configPath := filepath.Join(cliCtx.testFolderPath, ".r2r", "r2r-cli.yml")
	os.Remove(configPath)

	return nil
}

func iCreateATestConfigFileWithValidSettings(filename string) error {
	if cliCtx.testFolderPath == "" {
		return fmt.Errorf("test folder not created yet")
	}

	configContent := `# Valid R2R CLI configuration
version: "1.0"
project:
  name: "test-project"
extensions: []
`

	// Ensure .r2r directory exists
	r2rDir := filepath.Join(cliCtx.testFolderPath, ".r2r")
	if err := os.MkdirAll(r2rDir, 0o750); err != nil {
		return fmt.Errorf("failed to create .r2r directory: %w", err)
	}

	configPath := filepath.Join(r2rDir, "r2r-cli.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func iCreateATestConfigFileWithInvalidSettings(filename string) error {
	if cliCtx.testFolderPath == "" {
		return fmt.Errorf("test folder not created yet")
	}

	configContent := `# Invalid R2R CLI configuration
version: "1.0"
project:
  name: [this is invalid yaml syntax
`

	// Ensure .r2r directory exists
	r2rDir := filepath.Join(cliCtx.testFolderPath, ".r2r")
	if err := os.MkdirAll(r2rDir, 0o750); err != nil {
		return fmt.Errorf("failed to create .r2r directory: %w", err)
	}

	configPath := filepath.Join(r2rDir, "r2r-cli.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func iRunTheBuiltCLIWith(args string) error {
	if cliCtx.executablePath == "" {
		return fmt.Errorf("executable not found")
	}

	argsList := strings.Fields(args)
	cmdArgs := append([]string{cliCtx.executablePath}, argsList...)

	return runCommandWithArgs(cmdArgs...)
}

// ============================================================================
// Helper Functions
// ============================================================================

func runCommandWithArgs(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec // Args are controlled test input

	output, err := cmd.CombinedOutput()
	cliCtx.sharedCtx.CommandOutput = string(output)

	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			cliCtx.sharedCtx.ExitCode = exitErr.ExitCode()
		} else {
			cliCtx.sharedCtx.ExitCode = 1
		}
	} else {
		cliCtx.sharedCtx.ExitCode = 0
	}

	return nil
}
