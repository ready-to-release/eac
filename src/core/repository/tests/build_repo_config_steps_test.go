// Godog BDD step definitions - Repository repo-config build validation
//
// Feature: repository_build-repo-config (specs/repository/build-repo-config/specification.feature)
//
// This file implements steps for validating that the repo-config module builds successfully.
package tests

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/repository"
)

// buildRepoConfigContext holds state for repo-config build validation
type buildRepoConfigContext struct {
	repoRoot string // Repository root path
	exitCode int    // Exit code from build command
	stdout   string // Standard output
	stderr   string // Standard error
	cmdErr   error  // Command execution error
}

var buildRepoConfigCtx *buildRepoConfigContext

// resetBuildRepoConfigContext resets the context between scenarios
func resetBuildRepoConfigContext() {
	buildRepoConfigCtx = nil
}

// ============================================================================
// When Steps
// ============================================================================

// runBuildRepoConfigCommand runs the "build repo-config" command
func runBuildRepoConfigCommand() error {
	// Initialize context if needed
	if buildRepoConfigCtx == nil {
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return fmt.Errorf("failed to get repository root: %w", err)
		}
		buildRepoConfigCtx = &buildRepoConfigContext{
			repoRoot: repoRoot,
		}
	}

	// Run the build repo-config command from repository root
	// The command is pwd-independent (works from any directory inside the repo)
	cmd := exec.Command("go", "run", "./src/commands", "build", "repo-config")
	cmd.Dir = buildRepoConfigCtx.repoRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	buildRepoConfigCtx.stdout = stdout.String()
	buildRepoConfigCtx.stderr = stderr.String()
	buildRepoConfigCtx.cmdErr = err

	// Extract exit code
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			buildRepoConfigCtx.exitCode = exitError.ExitCode()
		} else {
			buildRepoConfigCtx.exitCode = 1
		}
	} else {
		buildRepoConfigCtx.exitCode = 0
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

// checkExitCodeIs verifies the exit code matches expected value
// This function works with buildRepoConfigCtx, buildContractsCtx, validateDepsCtx, and testTagsCtx
func checkExitCodeIs(expectedCode int) error {
	var exitCode int
	var stdout, stderr string
	var cmdErr error
	var contextName string

	// Check which context is initialized
	if buildRepoConfigCtx != nil {
		exitCode = buildRepoConfigCtx.exitCode
		stdout = buildRepoConfigCtx.stdout
		stderr = buildRepoConfigCtx.stderr
		cmdErr = buildRepoConfigCtx.cmdErr
		contextName = "repo-config"
	} else if buildContractsCtx != nil {
		exitCode = buildContractsCtx.exitCode
		stdout = buildContractsCtx.stdout
		stderr = buildContractsCtx.stderr
		cmdErr = buildContractsCtx.cmdErr
		contextName = "contracts"
	} else if validateDepsCtx != nil {
		exitCode = validateDepsCtx.exitCode
		stdout = validateDepsCtx.stdout
		stderr = validateDepsCtx.stderr
		cmdErr = validateDepsCtx.cmdErr
		contextName = "validate-dependencies"
	} else if testTagsCtx != nil {
		exitCode = testTagsCtx.exitCode
		stdout = testTagsCtx.stdout
		stderr = testTagsCtx.stderr
		cmdErr = testTagsCtx.cmdErr
		contextName = "test-tags"
	} else {
		return fmt.Errorf("no command context initialized")
	}

	if exitCode != expectedCode {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Expected exit code %d, got %d (context: %s)\n", expectedCode, exitCode, contextName))
		if stdout != "" {
			details.WriteString(fmt.Sprintf("\nStdout:\n%s\n", stdout))
		}
		if stderr != "" {
			details.WriteString(fmt.Sprintf("\nStderr:\n%s\n", stderr))
		}
		if cmdErr != nil {
			details.WriteString(fmt.Sprintf("\nError: %v\n", cmdErr))
		}
		return fmt.Errorf("%s", details.String())
	}

	return nil
}

// shouldNotSeeBuildErrors verifies no build errors in output
// This function works with both buildRepoConfigCtx and buildContractsCtx
func shouldNotSeeBuildErrors() error {
	var stdout, stderr string
	var contextName string

	// Check which context is initialized
	if buildRepoConfigCtx != nil {
		stdout = buildRepoConfigCtx.stdout
		stderr = buildRepoConfigCtx.stderr
		contextName = "repo-config"
	} else if buildContractsCtx != nil {
		stdout = buildContractsCtx.stdout
		stderr = buildContractsCtx.stderr
		contextName = "contracts"
	} else {
		return fmt.Errorf("no build context initialized")
	}

	// Check for common error indicators in output
	combinedOutput := stdout + stderr
	errorIndicators := []string{
		"error",
		"failed",
		"fatal",
		"panic",
		"cannot",
		"undefined",
	}

	var foundErrors []string
	for _, indicator := range errorIndicators {
		if strings.Contains(strings.ToLower(combinedOutput), indicator) {
			foundErrors = append(foundErrors, indicator)
		}
	}

	if len(foundErrors) > 0 {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Found build errors in %s:\n", contextName))
		details.WriteString(fmt.Sprintf("  Error indicators found: %v\n\n", foundErrors))
		if stdout != "" {
			details.WriteString(fmt.Sprintf("Stdout:\n%s\n", stdout))
		}
		if stderr != "" {
			details.WriteString(fmt.Sprintf("Stderr:\n%s\n", stderr))
		}
		return fmt.Errorf("%s", details.String())
	}

	return nil
}

// ============================================================================
// Scenario Initialization
// ============================================================================

// InitializeBuildRepoConfigScenario registers step definitions for repo-config build validation tests
func InitializeBuildRepoConfigScenario(sc *godog.ScenarioContext) {
	// Given steps - repositoryRootExists is shared and registered by other step files

	// When steps
	sc.Step(`^I run the command "build repo-config"$`, runBuildRepoConfigCommand)

	// Then steps (shared across repository validation specs)
	sc.Step(`^the exit code is (\d+)$`, checkExitCodeIs)
	sc.Step(`^I should not see any build errors$`, shouldNotSeeBuildErrors)
}
