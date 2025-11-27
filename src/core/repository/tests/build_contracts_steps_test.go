// Godog BDD step definitions - Repository contracts build validation
//
// Feature: repository_build-contracts (specs/repository/build-contracts/specification.feature)
//
// This file implements steps for validating that the contracts module builds successfully.
package tests

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/repository"
)

// buildContractsContext holds state for contracts build validation
type buildContractsContext struct {
	repoRoot string // Repository root path
	exitCode int    // Exit code from build command
	stdout   string // Standard output
	stderr   string // Standard error
	cmdErr   error  // Command execution error
}

var buildContractsCtx *buildContractsContext

// resetBuildContractsContext resets the context between scenarios
func resetBuildContractsContext() {
	buildContractsCtx = nil
}

// ============================================================================
// When Steps
// ============================================================================

// runBuildContractsCommand runs the "build contracts" command
func runBuildContractsCommand() error {
	// Initialize context if needed
	if buildContractsCtx == nil {
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return fmt.Errorf("failed to get repository root: %w", err)
		}
		buildContractsCtx = &buildContractsContext{
			repoRoot: repoRoot,
		}
	}

	// Run the build contracts command from repository root
	// The command is pwd-independent (works from any directory inside the repo)
	cmd := exec.Command("go", "run", "./src/commands", "build", "contracts")
	cmd.Dir = buildContractsCtx.repoRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	buildContractsCtx.stdout = stdout.String()
	buildContractsCtx.stderr = stderr.String()
	buildContractsCtx.cmdErr = err

	// Extract exit code
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			buildContractsCtx.exitCode = exitError.ExitCode()
		} else {
			buildContractsCtx.exitCode = 1
		}
	} else {
		buildContractsCtx.exitCode = 0
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

// buildContractsExitCodeIs verifies the exit code matches expected value
func buildContractsExitCodeIs(expectedCode int) error {
	if buildContractsCtx == nil {
		return fmt.Errorf("build contracts context not initialized")
	}

	if buildContractsCtx.exitCode != expectedCode {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Expected exit code %d, got %d\n", expectedCode, buildContractsCtx.exitCode))
		if buildContractsCtx.stdout != "" {
			details.WriteString(fmt.Sprintf("\nStdout:\n%s\n", buildContractsCtx.stdout))
		}
		if buildContractsCtx.stderr != "" {
			details.WriteString(fmt.Sprintf("\nStderr:\n%s\n", buildContractsCtx.stderr))
		}
		if buildContractsCtx.cmdErr != nil {
			details.WriteString(fmt.Sprintf("\nError: %v\n", buildContractsCtx.cmdErr))
		}
		return fmt.Errorf("%s", details.String())
	}

	return nil
}

// ============================================================================
// Scenario Initialization
// ============================================================================

// InitializeBuildContractsScenario registers step definitions for contracts build validation tests
func InitializeBuildContractsScenario(sc *godog.ScenarioContext) {
	// Given steps - repositoryRootExists is shared and registered by other step files

	// When steps
	sc.Step(`^I run the command "build contracts"$`, runBuildContractsCommand)

	// Then steps - shared steps registered by build_repo_config_steps_test.go:
	// - "the exit code is (\d+)" -> checkExitCodeIs
	// - "I should not see any build errors" -> shouldNotSeeBuildErrors
}
