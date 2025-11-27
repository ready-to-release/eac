// Godog BDD step definitions - Repository test tag validation
//
// Feature: repository_test-tags-contracted (specs/repository/test-tags-contracted/specification.feature)
//
// This file implements steps for validating that all test tags are defined in the tag contract.
package tests

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/repository"
)

// testTagsContractedContext holds state for test tag validation
type testTagsContractedContext struct {
	repoRoot string // Repository root path
	exitCode int    // Exit code from validate test-tags command
	stdout   string // Standard output
	stderr   string // Standard error
	cmdErr   error  // Command execution error
}

var testTagsCtx *testTagsContractedContext

// resetTestTagsContractedContext resets the context between scenarios
func resetTestTagsContractedContext() {
	testTagsCtx = nil
}

// ============================================================================
// Given Steps
// ============================================================================

// (Uses shared repositoryRootExists from go_modules_tidy_steps_test.go)

// ============================================================================
// When Steps
// ============================================================================

// runValidateTestTagsCommand runs the "validate test-tags" command
func runValidateTestTagsCommand() error {
	// Initialize context if needed
	if testTagsCtx == nil {
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return fmt.Errorf("failed to get repository root: %w", err)
		}
		testTagsCtx = &testTagsContractedContext{
			repoRoot: repoRoot,
		}
	}

	// Run the validate test-tags command from repository root
	// The command is pwd-independent (works from any directory inside the repo)
	cmd := exec.Command("go", "run", "./src/commands", "validate", "test-tags")
	cmd.Dir = testTagsCtx.repoRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	testTagsCtx.stdout = stdout.String()
	testTagsCtx.stderr = stderr.String()
	testTagsCtx.cmdErr = err

	// Extract exit code
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			testTagsCtx.exitCode = exitError.ExitCode()
		} else {
			testTagsCtx.exitCode = 1
		}
	} else {
		testTagsCtx.exitCode = 0
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

// shouldNotSeeUndefinedTagErrors verifies no undefined tag errors in output
func shouldNotSeeUndefinedTagErrors() error {
	if testTagsCtx == nil {
		return fmt.Errorf("test tags context not initialized")
	}

	// Check for common error indicators in output
	combinedOutput := testTagsCtx.stdout + testTagsCtx.stderr
	errorIndicators := []string{
		"undefined tag",
		"invalid tag",
		"unknown tag",
		"not found in contract",
		"not defined in tag contract",
	}

	var foundErrors []string
	for _, indicator := range errorIndicators {
		if strings.Contains(strings.ToLower(combinedOutput), indicator) {
			foundErrors = append(foundErrors, indicator)
		}
	}

	if len(foundErrors) > 0 {
		var details strings.Builder
		details.WriteString("Found undefined tag errors:\n")
		details.WriteString(fmt.Sprintf("  Error indicators found: %v\n\n", foundErrors))
		if testTagsCtx.stdout != "" {
			details.WriteString(fmt.Sprintf("Stdout:\n%s\n", testTagsCtx.stdout))
		}
		if testTagsCtx.stderr != "" {
			details.WriteString(fmt.Sprintf("Stderr:\n%s\n", testTagsCtx.stderr))
		}
		return fmt.Errorf("%s", details.String())
	}

	return nil
}

// ============================================================================
// Scenario Initialization
// ============================================================================

// InitializeTestTagsContractedScenario registers step definitions for test tag validation
func InitializeTestTagsContractedScenario(sc *godog.ScenarioContext) {
	// Given steps - repositoryRootExists is shared and registered by other step files

	// When steps
	sc.Step(`^I run the command "validate test-tags"$`, runValidateTestTagsCommand)

	// Then steps
	// Note: "the exit code is (\d+)" step is shared and registered by build_repo_config_steps_test.go -> checkExitCodeIs
	sc.Step(`^I should not see any undefined tag errors$`, shouldNotSeeUndefinedTagErrors)
}
