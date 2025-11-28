// Godog step definitions for src-commands features
//
// Features:
// - specs/src-commands/ai-commit-generation/
// - specs/src-commands/build-module/
package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	coretesting "github.com/ready-to-release/eac/src/core/testing"

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
	_ "github.com/ready-to-release/eac/src/commands/impl/specs"
	_ "github.com/ready-to-release/eac/src/commands/impl/specs/create"
	_ "github.com/ready-to-release/eac/src/commands/impl/specs/validate"
	_ "github.com/ready-to-release/eac/src/commands/impl/templates"
	_ "github.com/ready-to-release/eac/src/commands/impl/templates/list"
	_ "github.com/ready-to-release/eac/src/commands/impl/templates/apply/docs"
	_ "github.com/ready-to-release/eac/src/commands/impl/templates/install/reports"
	_ "github.com/ready-to-release/eac/src/commands/impl/test"
	_ "github.com/ready-to-release/eac/src/commands/impl/work"

	// Import templates test packages
	listtests "github.com/ready-to-release/eac/src/commands/impl/templates/list/tests"
)

// sharedCtx is the shared test context used by child test packages.
// Child packages (commit/tests, specs/tests, etc.) receive this pointer directly,
// so changes are immediately visible without manual synchronization.
// This replaces the old BeforeStep/AfterStep sync pattern.
var sharedCtx *coretesting.SharedTestContext

// testContext holds state between steps for the main test runner.
// This is kept for backward compatibility with existing step definitions.
// On each step, we sync between ctx and sharedCtx to keep them in sync.
type testContext struct {
	commandOutput     string
	exitCode          int
	commandError      error
	testCommitMessage string   // For commit validation testing
	affectedModules   []string // Modules affected by current changes
	validationErrors  []string // Validation error codes for assertions
}

var ctx *testContext

// originalRepoRoot stores the actual eac repository root
// This is set once at test initialization and used for running go commands
// even when tests change the current working directory for git isolation
var originalRepoRoot string

// syncCtxToShared copies state from ctx to sharedCtx.
// Called before steps that child packages might use.
func syncCtxToShared() {
	if sharedCtx == nil || ctx == nil {
		return
	}
	sharedCtx.CommandOutput = ctx.commandOutput
	sharedCtx.ExitCode = ctx.exitCode
	sharedCtx.CommandError = ctx.commandError
	sharedCtx.TestCommitMessage = ctx.testCommitMessage
	sharedCtx.AffectedModules = ctx.affectedModules
	sharedCtx.ValidationErrors = ctx.validationErrors
}

// syncSharedToCtx copies state from sharedCtx to ctx.
// Called after steps that child packages might have modified.
func syncSharedToCtx() {
	if sharedCtx == nil || ctx == nil {
		return
	}
	ctx.commandOutput = sharedCtx.CommandOutput
	ctx.exitCode = sharedCtx.ExitCode
	ctx.commandError = sharedCtx.CommandError
	ctx.testCommitMessage = sharedCtx.TestCommitMessage
	ctx.affectedModules = sharedCtx.AffectedModules
	ctx.validationErrors = sharedCtx.ValidationErrors
}

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

	// Only print test execution details if GODOG_VERBOSE is set
	if os.Getenv("GODOG_VERBOSE") != "" {
		fmt.Printf("[TEST] Executing read-only command: %s\n", execCtx.commandName)
	}

	return executeCommand(execCtx)
}

// runMutatingCommand executes commands with side-effects
func runMutatingCommand(execCtx *commandExecutionContext) error {
	// Allow mutating commands to run in tests
	// Test isolation is handled through:
	// - Temporary git repositories (testRepoRoot)
	// - AI test doubles (testAIResponse)
	// - Each scenario runs in a clean environment

	// Only print test execution details if GODOG_VERBOSE is set
	if os.Getenv("GODOG_VERBOSE") != "" {
		fmt.Printf("[TEST] Executing mutating command: %s\n", execCtx.commandName)
	}

	return executeCommand(execCtx)
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

	// Determine working directory
	// For isolated tests, use the isolated directory; otherwise use real repo
	var workDir string
	if isolatedTestProjectDir != "" {
		// Isolated test - run from the isolated temp directory
		workDir = isolatedTestProjectDir
	} else {
		// Non-isolated test - use the original repository root
		workDir = filepath.Join(originalRepoRoot, "src", "commands")
	}

	// Set working directory - must be done BEFORE building the command
	// For isolated tests, this needs to be the source directory
	if isolatedTestProjectDir != "" {
		// For isolated tests, we need to build/run from the source but pretend we're in the isolated dir
		// So use the real source directory
		cmd.Dir = filepath.Join(originalRepoRoot, "src", "commands")
	} else {
		cmd.Dir = workDir
	}

	// Build environment with test overrides
	env := os.Environ()

	// For isolated tests, set environment variables to use the isolated directory
	if isolatedTestProjectDir != "" {
		// R2R_PWD: Where user invoked command (for relative path resolution)
		env = append(env, fmt.Sprintf("R2R_PWD=%s", isolatedTestProjectDir))
		// R2R_REPO_ROOT: Repository root override
		env = append(env, fmt.Sprintf("R2R_REPO_ROOT=%s", isolatedTestProjectDir))
	}

	cmd.Env = env

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

	// Check if this is a "go run ." command
	if len(parts) >= 3 && parts[0] == "go" && parts[1] == "run" {
		args := parts[3:] // Skip "go run ."
		return runCommandWithArgs(args...)
	}

	// Otherwise, treat it as a direct command (e.g., "specs validate file.feature")
	return runCommandWithArgs(parts...)
}

func iRunTheCommand(cmdLine string) error {
	// New step for "When I run the command <command>"
	// Takes just the command without "go run ."

	// Parse command and lookup metadata from registry
	execCtx, err := parseAndLookupCommand(cmdLine)
	if err != nil {
		// Fallback: use legacy path if registry lookup fails
		fmt.Printf("[TEST] Note: Could not lookup command in registry, using legacy path: %v\n", err)
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
	// Initialize shared context for child packages
	sharedCtx = coretesting.NewSharedTestContext()
	sharedCtx.OriginalRepoRoot = originalRepoRoot
	sharedCtx.RunCommand = iRunTheCommand
	return nil
}

// isolatedTestProjectDir stores the temp directory for isolated test projects
// Created when @env:isolated-test-project tag is present
var isolatedTestProjectDir string

// testIsolation holds the current test isolation instance for cleanup
var testIsolation *coretesting.TestIsolation

// initializeIsolatedTestProject creates an isolated test environment.
// This prevents tests from polluting the real repository with .r2r, custom, .docs folders.
//
// Note: No .git directory is created - the R2R_REPO_ROOT environment variable
// is used instead to tell GetRepositoryRoot() where the repo root is.
func initializeIsolatedTestProject() error {
	testIsolation = coretesting.NewTestIsolation().
		WithOriginalRepoRoot(originalRepoRoot).
		WithCopyContracts(true)

	if err := testIsolation.Setup(); err != nil {
		return fmt.Errorf("failed to setup test isolation: %w", err)
	}

	isolatedTestProjectDir = testIsolation.IsolatedDir()

	// Update shared context with isolation info
	if sharedCtx != nil {
		sharedCtx.SetIsolation(originalRepoRoot, isolatedTestProjectDir)
	}

	return nil
}

// cleanupScenario cleans up resources created during scenario execution
func cleanupScenario() error {
	// Clean up test isolation (handles temp directory removal)
	if testIsolation != nil {
		testIsolation.Cleanup()
		testIsolation = nil
		isolatedTestProjectDir = ""
	}

	// Clear isolation from shared context
	if sharedCtx != nil {
		sharedCtx.ClearIsolation()
	}

	// Reset commit mocks (git repo and AI response)
	cleanupCommitMocks()

	// Reset specs mocks (git repo and AI response)
	cleanupSpecsMocks()

	// Reset init mocks (git repo)
	cleanupInitMocks()

	// Templates context cleanup is handled by the individual test packages
	// (list/tests, apply/tests, install/tests) via their own Before/After hooks

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func runCommandWithArgs(args ...string) error {
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)

	// Determine working directory based on isolation
	if isolatedTestProjectDir != "" {
		// For isolated tests, build/run from source but use isolated directory
		cmd.Dir = filepath.Join(originalRepoRoot, "src", "commands")
	} else {
		cmd.Dir = ".." // src/commands directory
	}

	// Build environment with test overrides (same as executeCommand)
	env := os.Environ()

	// For isolated tests, set environment variables to use the isolated directory
	if isolatedTestProjectDir != "" {
		// R2R_PWD: Where user invoked command (for relative path resolution)
		env = append(env, fmt.Sprintf("R2R_PWD=%s", isolatedTestProjectDir))
		// R2R_REPO_ROOT: Repository root override
		env = append(env, fmt.Sprintf("R2R_REPO_ROOT=%s", isolatedTestProjectDir))
	}

	cmd.Env = env

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
	// Initialize shared context FIRST, before any child initializers run.
	// This ensures sharedCtx is available when child packages set up their state.
	sharedCtx = coretesting.NewSharedTestContext()
	sharedCtx.OriginalRepoRoot = originalRepoRoot
	sharedCtx.RunCommand = iRunTheCommand

	// Templates command steps (must be FIRST to override generic "I run the command")
	// All templates scenarios (list, apply, install) are handled by InitializeListScenario
	listtests.InitializeListScenario(sc)

	// Specs command steps
	InitializeSpecsScenario(sc)

	// Init command steps
	InitializeInitScenario(sc)

	// Design command steps
	InitializeDesignScenario(sc)

	// Docs command steps
	InitializeDocsScenario(sc)

	// Commit command steps
	InitializeCommitScenario(sc)

	// Work command steps
	InitializeWorkScenario(sc)

	sc.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
		initializeContext()
		// Re-sync shared context after initializeContext() resets ctx
		syncCtxToShared()

		// Check for @env:isolated-test-project tag
		for _, tag := range scenario.Tags {
			if tag.Name == "@env:isolated-test-project" {
				if err := initializeIsolatedTestProject(); err != nil {
					return ctx, fmt.Errorf("failed to initialize isolated test project: %w", err)
				}
				// Set up commit mocks automatically for isolated tests
				if err := setupCommitMocks(); err != nil {
					return ctx, fmt.Errorf("failed to setup commit mocks: %w", err)
				}
				// Set up specs mocks automatically for isolated tests
				if err := setupSpecsMocks(); err != nil {
					return ctx, fmt.Errorf("failed to setup specs mocks: %w", err)
				}
				// Set up docs mocks automatically for isolated tests
				if err := setupDocsMocks(); err != nil {
					return ctx, fmt.Errorf("failed to setup docs mocks: %w", err)
				}
				// Set up init mocks automatically for isolated tests
				if err := setupInitMocks(); err != nil {
					return ctx, fmt.Errorf("failed to setup init mocks: %w", err)
				}
				break
			}
		}

		return ctx, nil
	})

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		cleanupScenario()
		return ctx, nil
	})

	// All steps
	sc.Step(`^I run "([^"]*)"$`, iRun)
	sc.Step(`^I run the command "([^"]*)"$`, iRunTheCommand)
	sc.Step(`^I run the command "([^"]*)" without arguments$`, func(cmd string) error {
		return iRunTheCommand(cmd)
	})
	sc.Step(`^I run "([^"]*)" or "([^"]*)" or "([^"]*)"$`, iRunOr)
	sc.Step(`^I run "([^"]*)", "([^"]*)", or "([^"]*)"$`, iRunOr)
	sc.Step(`^the exit code is (\d+)$`, theExitCodeIs)
	sc.Step(`^the exit code is (\d+) or (\d+)$`, theExitCodeIsOr)
	sc.Step(`^the exit code is (\d+) or the exit code is (\d+)$`, theExitCodeIsOr)
	sc.Step(`^I should see "([^"]*)"$`, iShouldSee)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)"$`, iShouldSeeOr)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)" or "([^"]*)"$`, iShouldSeeOrOr)
	sc.Step(`^I should see "([^"]*)" or "([^"]*)" or "([^"]*)" or "([^"]*)"$`, iShouldSeeOrOrOr)
	sc.Step(`^I should see "([^"]*)" on stderr$`, iShouldSeeOnStderr)

	// Common steps for stdout/stderr verification
	sc.Step(`^stdout contains "([^"]*)"$`, stdoutContains)
	sc.Step(`^stderr contains "([^"]*)"$`, stderrContains)
	sc.Step(`^stdout contains "([^"]*)" or "([^"]*)"$`, stdoutContainsOr)
	sc.Step(`^stdout contains validation summary$`, stdoutContainsValidationSummary)
	sc.Step(`^stdout lists all invalid files with their errors$`, stdoutListsAllInvalidFilesWithTheirErrors)
	sc.Step(`^stdout shows which contract rules are checked$`, stdoutShowsWhichContractRulesAreChecked)
	sc.Step(`^stdout contains error count "([^"]*)"$`, stdoutContainsErrorCount)
	sc.Step(`^stdout contains "([^"]*)" and file count$`, stdoutContainsAndFileCount)
	sc.Step(`^stdout contains validation results for all files$`, stdoutContainsValidationResultsForAllFiles)

	// Exit code steps
	sc.Step(`^the command exits with code (\d+)$`, theCommandExitsWithCode)
	sc.Step(`^the exit code should be (\d+)$`, theExitCodeShouldBe)

	// Validation and contract steps
	sc.Step(`^contract rules from "([^"]*)" are applied$`, contractRulesAreApplied)
	sc.Step(`^validation proceeds normally$`, validationProceedsNormally)
	sc.Step(`^validation fails with "([^"]*)"$`, validationFailsWith)
	sc.Step(`^the path is accepted$`, thePathIsAccepted)

	// Template and file operation steps
	sc.Step(`^template placeholders should remain unchanged$`, templatePlaceholdersShouldRemainUnchanged)
	sc.Step(`^the templates should be copied from local path "([^"]*)"$`, theTemplatesShouldBeCopiedFromLocalPath)
	sc.Step(`^the templates should be read from subdirectory "([^"]*)"$`, theTemplatesShouldBeReadFromSubdirectory)
	sc.Step(`^no files should be written outside the output directory$`, noFilesShouldBeWrittenOutsideTheOutputDirectory)
	sc.Step(`^no partial files should be created$`, noPartialFilesShouldBeCreated)

	// Output format steps
	sc.Step(`^successful validations are not displayed$`, successfulValidationsAreNotDisplayed)
	sc.Step(`^warnings are displayed with "([^"]*)" prefix$`, warningsAreDisplayedWithPrefix)
	sc.Step(`^JSON includes file path, errors array, and summary$`, jsonIncludesFilePathErrorsArrayAndSummary)

	// Edge case steps
	sc.Step(`^non-feature files are ignored silently$`, nonFeatureFilesAreIgnoredSilently)
	sc.Step(`^no validation errors should occur$`, noValidationErrorsShouldOccur)

	// Background/setup steps
	sc.Step(`^I am in a git repository$`, iAmInAGitRepository)
	sc.Step(`^specification contracts are available$`, specificationContractsAreAvailable)

	// Module-specific steps
	sc.Step(`^each section should have module-specific context$`, eachSectionShouldHaveModuleSpecificContext)
	sc.Step(`^metrics include module count of (\d+)$`, metricsIncludeModuleCount)
	sc.Step(`^missing module stub is added for "([^"]*)"$`, missingModuleStubIsAdded)

	// Additional common steps
	// Note: "only that file's diff should be included", "other files should be excluded",
	// and "no module sections should be created" are registered in commit_steps_test.go
	// Note: "only valid Gherkin content should remain" is registered in specs_bridge_test.go
	sc.Step(`^all sections are in the correct order$`, allSectionsAreInTheCorrectOrder)

	// BDD common steps (alternative phrasings from spec features)
	sc.Step(`^the command is executed$`, theCommandIsExecuted)
	sc.Step(`^the file is readable$`, theFileIsReadable)
	sc.Step(`^the file "([^"]*)" contains "([^"]*)"$`, theFileContainsExpectedContent)
	sc.Step(`^the directory "([^"]*)" contains a file "([^"]*)"$`, theDirectoryContainsAFile)
	sc.Step(`^the directory is empty$`, theDirectoryIsEmpty)
	sc.Step(`^the standard output contains "([^"]*)"$`, theStandardOutputContains)
	sc.Step(`^the standard error contains "([^"]*)"$`, theStandardErrorContains)
	sc.Step(`^the standard output is empty$`, theStandardOutputIsEmpty)
	sc.Step(`^the standard error is not empty$`, theStandardErrorIsNotEmpty)
	sc.Step(`^the standard output does not contain "([^"]*)"$`, theStandardOutputDoesNotContain)
	sc.Step(`^the standard output matches the pattern "([^"]*)"$`, theStandardOutputMatchesThePattern)
	sc.Step(`^the exit code indicates a timeout condition$`, theExitCodeIndicatesATimeoutCondition)
	sc.Step(`^the exit code is not (\d+)$`, theExitCodeIsNot)
	sc.Step(`^the exit code should be (\d+) for both$`, theExitCodeShouldBeForBoth)
	sc.Step(`^the command processes the arguments without truncation$`, theCommandProcessesTheArgumentsWithoutTruncation)
	sc.Step(`^a parsing error should be returned$`, aParsingErrorShouldBeReturned)
	sc.Step(`^an empty string should be returned$`, anEmptyStringShouldBeReturned)
	sc.Step(`^an error should be returned$`, anErrorShouldBeReturned)

	// BDD common Given steps - command setup
	sc.Step(`^a command "([^"]*)"$`, aCommand)
	sc.Step(`^a command that does not exist$`, aCommandThatDoesNotExist)
	sc.Step(`^a command that fails$`, aCommandThatFails)
	sc.Step(`^a command that produces output$`, aCommandThatProducesOutput)
	sc.Step(`^a command that produces formatted output$`, aCommandThatProducesFormattedOutput)
	sc.Step(`^a command with filtered output$`, aCommandWithFilteredOutput)
	sc.Step(`^a command with no output$`, aCommandWithNoOutput)
	sc.Step(`^a command that writes to a file$`, aCommandThatWritesToAFile)
	sc.Step(`^a command with invalid arguments$`, aCommandWithInvalidArguments)
	sc.Step(`^a command with a very long argument list$`, aCommandWithAVeryLongArgumentList)
	sc.Step(`^a valid command$`, aValidCommand)
	sc.Step(`^the environment variable "([^"]*)" is set to "([^"]*)"$`, theEnvironmentVariableIsSetTo)
	sc.Step(`^no authentication credentials are provided$`, noAuthenticationCredentialsAreProvided)
	sc.Step(`^the command is executed with insufficient permissions$`, theCommandIsExecutedWithInsufficientPermissions)

	// BDD common Given steps - file/directory setup
	sc.Step(`^a file exists at "([^"]*)"$`, aFileExistsAt)
	sc.Step(`^a file is created at "([^"]*)"$`, aFileIsCreatedAt)
	sc.Step(`^the file "([^"]*)" exists$`, theFileExists)
	sc.Step(`^the file no longer exists$`, theFileNoLongerExists)
	sc.Step(`^the output file is created$`, theOutputFileIsCreated)

	// BDD common Then steps - additional verifications
	sc.Step(`^both commands complete successfully$`, bothCommandsCompleteSuccessfully)
	sc.Step(`^the command is terminated$`, theCommandIsTerminated)
	sc.Step(`^the command history does not contain the secret value$`, theCommandHistoryDoesNotContainTheSecretValue)
	// Note: "execution context is built" is registered in commit_steps_test.go
	sc.Step(`^no data race warnings are reported$`, noDataRaceWarningsAreReported)
	sc.Step(`^no file corruption occurs$`, noFileCorruptionOccurs)
	sc.Step(`^metrics include total time$`, metricsIncludeTotalTime)
	sc.Step(`^the speedup is at least (\d+) percent$`, theSpeedupIsAtLeastPercent)
	sc.Step(`^the total generation time is approximately the sum of individual times$`, theTotalGenerationTimeIsApproximatelyTheSumOfIndividualTimes)

	// BDD common - additional setup/background steps
	sc.Step(`^the working directory is clean$`, theWorkingDirectoryIsClean)
	sc.Step(`^a test directory exists$`, aTestDirectoryExists)
	sc.Step(`^a command that creates a file$`, aCommandThatCreatesAFile)
	sc.Step(`^a command that creates a directory$`, aCommandThatCreatesADirectory)
	sc.Step(`^the directory "([^"]*)" exists$`, theDirectoryExists)
	sc.Step(`^the directory contains a file "([^"]*)"$`, theDirectoryContainsAFile)
	sc.Step(`^the command is run$`, theCommandIsRun)
	sc.Step(`^the commands are executed concurrently$`, theCommandsAreExecutedConcurrently)
	sc.Step(`^the command is executed with a timeout of (\d+) seconds$`, theCommandIsExecutedWithATimeoutOfSeconds)
	sc.Step(`^the cleanup command is executed$`, theCleanupCommandIsExecuted)

	// BDD common - security & authentication steps
	sc.Step(`^a command that handles secrets$`, aCommandThatHandlesSecrets)
	sc.Step(`^a command that processes sensitive data$`, aCommandThatProcessesSensitiveData)
	sc.Step(`^a command that requires authentication$`, aCommandThatRequiresAuthentication)
	sc.Step(`^the standard output contains a usage message$`, theStandardOutputContainsAUsageMessage)

	// Specs steps are registered via InitializeSpecsScenario() in specs_bridge_test.go
	// which delegates to src/commands/impl/specs/tests/

	// Templates steps are registered via InitializeListScenario(), InitializeApplyScenario(),
	// and InitializeInstallScenario() which delegate to src/commands/impl/templates/*/tests/
}

func InitializeTestSuite(sc *godog.TestSuiteContext) {
	sc.BeforeSuite(func() {})
	sc.AfterSuite(func() {})
}
