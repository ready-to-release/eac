// Godog BDD step definitions for src-commands features
//
// Features:
// - specs/src-commands/ai-commit-generation/
// - specs/src-commands/build-module/
package tests

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cucumber/godog"
)

// Test context holds state between steps
type testContext struct {
	commandOutput string
	exitCode      int
	commandError  error
}

var ctx *testContext

// ============================================================================
// Common Steps
// ============================================================================

func iRun(cmdLine string) error {
	parts := strings.Fields(cmdLine)
	if len(parts) < 3 || parts[0] != "go" || parts[1] != "run" {
		return fmt.Errorf("invalid command format: %s", cmdLine)
	}

	args := parts[3:] // Skip "go run ."
	return runCommandWithArgs(args...)
}

func theExitCodeIs(expectedCode int) error {
	if ctx.exitCode != expectedCode {
		return fmt.Errorf("expected exit code %d, got %d. Output:\n%s",
			expectedCode, ctx.exitCode, ctx.commandOutput)
	}
	return nil
}

func theExitCodeIsOr(code1, code2 int) error {
	if ctx.exitCode == code1 || ctx.exitCode == code2 {
		return nil
	}
	return fmt.Errorf("expected exit code %d or %d, got %d. Output:\n%s",
		code1, code2, ctx.exitCode, ctx.commandOutput)
}

func iShouldSee(text string) error {
	if !strings.Contains(ctx.commandOutput, text) {
		return fmt.Errorf("expected output to contain '%s', got:\n%s", text, ctx.commandOutput)
	}
	return nil
}

func iShouldSeeOr(text1, text2 string) error {
	if strings.Contains(ctx.commandOutput, text1) ||
		strings.Contains(ctx.commandOutput, text2) {
		return nil
	}
	return fmt.Errorf("expected output to contain one of '%s' or '%s', got:\n%s",
		text1, text2, ctx.commandOutput)
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

func iShouldSeeOrOrOr(text1, text2, text3, text4 string) error {
	if strings.Contains(ctx.commandOutput, text1) ||
		strings.Contains(ctx.commandOutput, text2) ||
		strings.Contains(ctx.commandOutput, text3) ||
		strings.Contains(ctx.commandOutput, text4) {
		return nil
	}
	return fmt.Errorf("expected output to contain one of '%s', '%s', '%s', or '%s', got:\n%s",
		text1, text2, text3, text4, ctx.commandOutput)
}

func iShouldSeeOnStderr(text string) error {
	if !strings.Contains(ctx.commandOutput, text) {
		return fmt.Errorf("expected '%s' in stderr/output", text)
	}
	return nil
}

func iRunOr(cmd1, cmd2, cmd3 string) error {
	return iRun("go run . " + cmd1)
}

// ============================================================================
// Setup/Initialization
// ============================================================================

func initializeContext() error {
	ctx = &testContext{}
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func runCommandWithArgs(args ...string) error {
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = ".." // src/commands directory

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
	sc.Step(`^I run "([^"]*)" or "([^"]*)" or "([^"]*)"$`, iRunOr)
	sc.Step(`^I run "([^"]*)", "([^"]*)", or "([^"]*)"$`, iRunOr)
	sc.Step(`^the exit code is (\d+)$`, theExitCodeIs)
	sc.Step(`^the exit code is (\d+) or the exit code is (\d+)$`, theExitCodeIsOr)
	sc.Step(`^I should see "([^"]*)"$`, iShouldSee)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)"$`, iShouldSeeOr)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)" or "([^"]*)"$`, iShouldSeeOrOr)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)" or "([^"]*)" or "([^"]*)"$`, iShouldSeeOrOrOr)
	sc.Step(`^I should see "([^"]*)" on stderr$`, iShouldSeeOnStderr)

	// Design command steps
	InitializeDesignScenario(sc)

	// Templates command steps
	InitializeTemplatesScenario(sc)
}

func InitializeTestSuite(sc *godog.TestSuiteContext) {
	sc.BeforeSuite(func() {})
	sc.AfterSuite(func() {})
}
