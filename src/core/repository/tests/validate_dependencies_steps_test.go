// Godog BDD step definitions - Repository dependency validation
//
// Feature: repository_validate-dependencies (specs/repository/validate-dependencies/specification.feature)
//
// This file implements steps for validating that all module dependencies are valid
// and consistent according to the dependency contracts.
package tests

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/repository"
)

// validateDependenciesContext holds state for dependency validation
type validateDependenciesContext struct {
	repoRoot   string // Repository root path
	exitCode   int    // Exit code from validate dependencies command
	stdout     string // Standard output
	stderr     string // Standard error
	cmdErr     error  // Command execution error
}

var validateDepsCtx *validateDependenciesContext

// resetValidateDependenciesContext resets the context between scenarios
func resetValidateDependenciesContext() {
	validateDepsCtx = nil
}

// ============================================================================
// Given Steps
// ============================================================================

// (Uses shared repositoryRootExists from go_modules_tidy_steps_test.go)

// ============================================================================
// When Steps
// ============================================================================

// runValidateDependenciesCommand runs the "validate dependencies" command
func runValidateDependenciesCommand() error {
	// Initialize context if needed
	if validateDepsCtx == nil {
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return fmt.Errorf("failed to get repository root: %w", err)
		}
		validateDepsCtx = &validateDependenciesContext{
			repoRoot: repoRoot,
		}
	}

	// Run the validate dependencies command from repository root
	// The command is pwd-independent (works from any directory inside the repo)
	cmd := exec.Command("go", "run", "./src/commands", "validate", "dependencies")
	cmd.Dir = validateDepsCtx.repoRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	validateDepsCtx.stdout = stdout.String()
	validateDepsCtx.stderr = stderr.String()
	validateDepsCtx.cmdErr = err

	// Extract exit code
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			validateDepsCtx.exitCode = exitError.ExitCode()
		} else {
			validateDepsCtx.exitCode = 1
		}
	} else {
		validateDepsCtx.exitCode = 0
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

// validateDependenciesExitCodeIs verifies the exit code matches expected value
func validateDependenciesExitCodeIs(expectedCode int) error {
	if validateDepsCtx == nil {
		return fmt.Errorf("validate dependencies context not initialized")
	}

	if validateDepsCtx.exitCode != expectedCode {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Expected exit code %d, got %d\n", expectedCode, validateDepsCtx.exitCode))
		if validateDepsCtx.stdout != "" {
			details.WriteString(fmt.Sprintf("\nStdout:\n%s\n", validateDepsCtx.stdout))
		}
		if validateDepsCtx.stderr != "" {
			details.WriteString(fmt.Sprintf("\nStderr:\n%s\n", validateDepsCtx.stderr))
		}
		if validateDepsCtx.cmdErr != nil {
			details.WriteString(fmt.Sprintf("\nError: %v\n", validateDepsCtx.cmdErr))
		}
		return fmt.Errorf("%s", details.String())
	}

	return nil
}

// shouldNotSeeDependencyValidationErrors verifies no validation errors in output
func shouldNotSeeDependencyValidationErrors() error {
	if validateDepsCtx == nil {
		return fmt.Errorf("validate dependencies context not initialized")
	}

	// Check for common error indicators in output
	combinedOutput := validateDepsCtx.stdout + validateDepsCtx.stderr
	errorIndicators := []string{
		"error",
		"failed",
		"invalid",
		"violation",
		"dependency cycle",
		"not found",
	}

	var foundErrors []string
	for _, indicator := range errorIndicators {
		if strings.Contains(strings.ToLower(combinedOutput), indicator) {
			foundErrors = append(foundErrors, indicator)
		}
	}

	if len(foundErrors) > 0 {
		var details strings.Builder
		details.WriteString("Found dependency validation errors:\n")
		details.WriteString(fmt.Sprintf("  Error indicators found: %v\n\n", foundErrors))
		if validateDepsCtx.stdout != "" {
			details.WriteString(fmt.Sprintf("Stdout:\n%s\n", validateDepsCtx.stdout))
		}
		if validateDepsCtx.stderr != "" {
			details.WriteString(fmt.Sprintf("Stderr:\n%s\n", validateDepsCtx.stderr))
		}
		return fmt.Errorf("%s", details.String())
	}

	return nil
}

// ============================================================================
// Scenario Initialization
// ============================================================================

// InitializeValidateDependenciesScenario registers step definitions for dependency validation tests
func InitializeValidateDependenciesScenario(sc *godog.ScenarioContext) {
	// Given steps - repositoryRootExists is shared and registered by other step files

	// When steps
	sc.Step(`^I run the command "validate dependencies"$`, runValidateDependenciesCommand)

	// Then steps
	// Note: "the exit code is (\d+)" step is shared and registered by build_repo_config_steps_test.go -> checkExitCodeIs
	sc.Step(`^I should not see any dependency validation errors$`, shouldNotSeeDependencyValidationErrors)
}
