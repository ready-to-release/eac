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
	"github.com/ready-to-release/eac/src/commands/internal/registry"

	// Import all command packages to trigger their init() and Register() calls
	_ "github.com/ready-to-release/eac/src/commands/impl/build"
	_ "github.com/ready-to-release/eac/src/commands/impl/commit"
	_ "github.com/ready-to-release/eac/src/commands/impl/describe"
	_ "github.com/ready-to-release/eac/src/commands/impl/design"
	_ "github.com/ready-to-release/eac/src/commands/impl/docs"
	_ "github.com/ready-to-release/eac/src/commands/impl/get"
	_ "github.com/ready-to-release/eac/src/commands/impl/list"
	_ "github.com/ready-to-release/eac/src/commands/impl/pipeline"
	_ "github.com/ready-to-release/eac/src/commands/impl/show"
	_ "github.com/ready-to-release/eac/src/commands/impl/templates"
	_ "github.com/ready-to-release/eac/src/commands/impl/test"
)

// Test context holds state between steps
type testContext struct {
	commandOutput string
	exitCode      int
	commandError  error
}

var ctx *testContext

// ============================================================================
// Command Execution Context
// ============================================================================

// commandExecutionContext holds metadata about the command being executed
type commandExecutionContext struct {
	commandName    string
	args           []string
	hasSideEffects bool
	registration   *registry.CommandRegistration
}

// ============================================================================
// Command Execution Abstraction
// ============================================================================

// parseAndLookupCommand extracts command info and looks up metadata from registry
func parseAndLookupCommand(cmdLine string) (*commandExecutionContext, error) {
	parts := strings.Fields(cmdLine)
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid command format: %s", cmdLine)
	}

	// Try to find longest matching command in registry
	// E.g., "design new" should match before "design"
	var foundCanonical string
	var remainingArgs []string

	for i := len(parts); i > 0; i-- {
		candidate := strings.Join(parts[:i], " ")
		canonical := registry.GetCanonicalName(candidate)

		if reg := registry.GetCommandByCanonical(canonical); reg != nil {
			foundCanonical = canonical
			remainingArgs = parts[i:]
			break
		}
	}

	if foundCanonical == "" {
		return nil, fmt.Errorf("command not found in registry: %s", parts[0])
	}

	// Get the registration for metadata
	reg := registry.GetCommandByCanonical(foundCanonical)

	return &commandExecutionContext{
		commandName:    foundCanonical,
		args:           remainingArgs,
		hasSideEffects: reg.HasSideEffects,
		registration:   reg,
	}, nil
}

// runReadOnlyCommand executes commands without side-effects
func runReadOnlyCommand(execCtx *commandExecutionContext) error {
	// Extension point: Add read-only specific logic here
	// - No confirmation needed
	// - Can run in parallel
	// - No audit logging
	// - Safe to retry

	fmt.Printf("[TEST] Executing read-only command: %s\n", execCtx.commandName)

	return executeCommand(execCtx)
}

// runMutatingCommand executes commands with side-effects
func runMutatingCommand(execCtx *commandExecutionContext) error {
	// Extension point: Add mutating command specific logic here
	// For now: BLOCK execution of side-effect commands in tests

	return fmt.Errorf("BLOCKED: Command '%s' has side-effects and cannot be executed in tests. Side-effect commands must be tested differently", execCtx.commandName)
}

// executeCommand is the common execution logic
func executeCommand(execCtx *commandExecutionContext) error {
	// Use ActualCommand from registration (has spaces, e.g., "list commands")
	// This is the actual command users type, not the internal canonical moniker
	actualCommand := execCtx.registration.ActualCommand

	// Build full command line
	// Split actualCommand into parts (e.g., "build module" → ["build", "module"])
	commandParts := strings.Fields(actualCommand)
	allArgs := append(commandParts, execCtx.args...)

	cmdArgs := append([]string{"run", "."}, allArgs...)
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

func iRunTheCommand(cmdLine string) error {
	// New step for "When I run the command <command>"
	// Takes just the command without "go run ."

	// Parse command and lookup metadata from registry
	execCtx, err := parseAndLookupCommand(cmdLine)
	if err != nil {
		// Fallback: use legacy path if registry lookup fails
		fmt.Printf("[TEST] Warning: Could not lookup command in registry, using legacy path: %v\n", err)
		parts := strings.Fields(cmdLine)
		return runCommandWithArgs(parts...)
	}

	// Route based on side-effects
	if execCtx.hasSideEffects {
		return runMutatingCommand(execCtx)
	} else {
		return runReadOnlyCommand(execCtx)
	}
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
	sc.Step(`^I run the command "([^"]*)"$`, iRunTheCommand)
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
