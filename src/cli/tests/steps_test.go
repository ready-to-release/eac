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
	commandOutput string
	exitCode      int
	commandError  error
	executablePath string
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
// Scenario Initialization
// ============================================================================

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		initializeContext()
		return ctx, nil
	})

	// All steps
	sc.Step(`^I run "([^"]*)"$`, iRun)
	sc.Step(`^the exit code is (\d+)$`, theExitCodeIs)
	sc.Step(`^I should see "([^"]*)"$`, iShouldSee)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)" or "([^"]*)"$`, iShouldSeeOrOr)
}

func InitializeTestSuite(sc *godog.TestSuiteContext) {
	sc.BeforeSuite(func() {})
	sc.AfterSuite(func() {})
}
