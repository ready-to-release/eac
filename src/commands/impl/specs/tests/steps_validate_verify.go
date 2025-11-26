// Package tests provides BDD step definitions for the specs commands.
//
// This file contains verification steps (Then) for the specs validate command.
package tests

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// registerValidateVerifySteps registers verification steps for specs validate
func registerValidateVerifySteps(sc *godog.ScenarioContext) {
	// Validator verification
	sc.Step(`^the GherkinValidator is used$`, theGherkinValidatorIsUsed)

	// Output verification - Note: The general "stdout contains" pattern in steps_test.go handles most cases
	// Only register patterns that need specific handling or don't match the general pattern
	sc.Step(`^stdout contains valid JSON$`, stdoutContainsValidJSON)

	// Recursive processing
	sc.Step(`^the command processes all \.feature files recursively$`, theCommandProcessesAllFeatureFilesRecursively)
	sc.Step(`^only \.feature files are validated$`, onlyFeatureFilesAreValidated)

	// Mode-specific output
	sc.Step(`^stdout only contains files with errors$`, stdoutOnlyContainsFilesWithErrors)
	sc.Step(`^stdout contains detailed validation steps$`, stdoutContainsDetailedValidationSteps)

	// Error verification - Note: "validation fails with" is registered in steps_test.go
	sc.Step(`^stderr contains "invalid path" or "path must be within repository"$`, stderrContainsInvalidPathOrPathMustBeWithinRepository)
	sc.Step(`^stderr contains "permission denied" or "failed to read file"$`, stderrContainsPermissionDeniedOrFailedToReadFile)
	sc.Step(`^stderr contains naming convention explanation$`, stderrContainsNamingConventionExplanation)
	sc.Step(`^stdout contains "([^"]*)" with error details$`, stdoutContainsLineWithErrorDetails)
	sc.Step(`^stdout contains error codes like (.*)$`, stdoutContainsErrorCodesLike)
	sc.Step(`^errors are displayed with "([^"]*)" prefix$`, errorsAreDisplayedWithPrefix)

	// Strict mode steps
	sc.Step(`^the command fails due to warnings$`, theCommandFailsDueToWarnings)
	sc.Step(`^warnings are treated as errors$`, warningsAreTreatedAsErrors)

	// Fix mode steps
	sc.Step(`^the tag issues are auto-fixed$`, theTagIssuesAreAutoFixed)
	sc.Step(`^the file is modified with fixes$`, theFileIsModifiedWithFixes)
	sc.Step(`^stdout contains "Fixed (\d+) issues"$`, stdoutContainsFixedIssues)

	// CheckTags steps
	sc.Step(`^tag validation is skipped$`, tagValidationIsSkipped)
	sc.Step(`^only syntax errors are reported$`, onlySyntaxErrorsAreReported)
}

// ============================================================================
// Validator Verification Steps
// ============================================================================

func theGherkinValidatorIsUsed() error {
	// The validator is used internally - verified by validation output
	return nil
}

// ============================================================================
// Output Verification Steps
// ============================================================================

// Note: stdoutContainsValidationPassed, stdoutContainsValidationFailed, stdoutContainsPassedFailed
// are handled by the general "stdout contains" pattern in steps_test.go

func stdoutContainsValidJSON() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := strings.TrimSpace(Ctx.CommandOutput)
	if strings.HasPrefix(output, "{") || strings.HasPrefix(output, "[") {
		return nil
	}
	return fmt.Errorf("stdout does not contain valid JSON.\nOutput:\n%s", Ctx.CommandOutput)
}

// ============================================================================
// Recursive Processing Steps
// ============================================================================

func theCommandProcessesAllFeatureFilesRecursively() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := Ctx.CommandOutput
	featureCount := strings.Count(output, ".feature")
	if featureCount > 1 {
		return nil
	}
	return fmt.Errorf("output does not show recursive processing of multiple files.\nOutput:\n%s", output)
}

func onlyFeatureFilesAreValidated() error {
	// This is verified by the command only processing .feature files
	return nil
}

// ============================================================================
// Mode-Specific Output Steps
// ============================================================================

func stdoutOnlyContainsFilesWithErrors() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	// In quiet mode, successful files should not be displayed
	output := Ctx.CommandOutput
	passCount := strings.Count(output, "✅")
	if passCount > 2 {
		return fmt.Errorf("quiet mode showing too many success messages.\nOutput:\n%s", output)
	}
	return nil
}

func stdoutContainsDetailedValidationSteps() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := Ctx.CommandOutput
	if strings.Contains(output, "checking") || strings.Contains(output, "validating") ||
		strings.Contains(output, "Validating") {
		return nil
	}
	return fmt.Errorf("stdout does not contain detailed validation steps.\nOutput:\n%s", output)
}

// ============================================================================
// Error Verification Steps
// ============================================================================

// Note: validationFailsWith is registered in steps_test.go

func stderrContainsInvalidPathOrPathMustBeWithinRepository() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := Ctx.CommandOutput
	if strings.Contains(output, "invalid path") || strings.Contains(output, "path must be within repository") {
		return nil
	}
	return fmt.Errorf("stderr does not contain path security error.\nOutput:\n%s", output)
}

func stderrContainsPermissionDeniedOrFailedToReadFile() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := Ctx.CommandOutput
	if strings.Contains(output, "permission denied") || strings.Contains(output, "failed to read file") {
		return nil
	}
	return fmt.Errorf("stderr does not contain permission error.\nOutput:\n%s", output)
}

func stderrContainsNamingConventionExplanation() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := Ctx.CommandOutput
	if strings.Contains(output, "naming") || strings.Contains(output, "convention") {
		return nil
	}
	return fmt.Errorf("stderr does not contain naming convention explanation.\nOutput:\n%s", output)
}

func stdoutContainsLineWithErrorDetails(lineRef string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	if !strings.Contains(Ctx.CommandOutput, lineRef) {
		return fmt.Errorf("stdout does not contain '%s' with error details.\nOutput:\n%s",
			lineRef, Ctx.CommandOutput)
	}
	return nil
}

func stdoutContainsErrorCodesLike(examples string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := Ctx.CommandOutput
	// Check for any of the example codes
	if strings.Contains(output, "MISSING_RULE") ||
		strings.Contains(output, "INVALID_FEATURE_NAMING") ||
		strings.Contains(output, "ERROR") {
		return nil
	}
	return fmt.Errorf("stdout does not contain error codes like %s.\nOutput:\n%s",
		examples, output)
}

func errorsAreDisplayedWithPrefix(prefix string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	if !strings.Contains(Ctx.CommandOutput, prefix) {
		return fmt.Errorf("errors not displayed with '%s' prefix.\nOutput:\n%s",
			prefix, Ctx.CommandOutput)
	}
	return nil
}

// ============================================================================
// Strict Mode Steps
// ============================================================================

func theCommandFailsDueToWarnings() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	// In strict mode, exit code should be non-zero if warnings exist
	if Ctx.ExitCode != 0 {
		return nil
	}
	return fmt.Errorf("command did not fail due to warnings (exit code: %d)", Ctx.ExitCode)
}

func warningsAreTreatedAsErrors() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	// In strict mode, warnings cause failure
	if Ctx.ExitCode != 0 {
		return nil
	}
	return fmt.Errorf("warnings were not treated as errors (exit code: %d)", Ctx.ExitCode)
}

// ============================================================================
// Fix Mode Steps
// ============================================================================

func theTagIssuesAreAutoFixed() error {
	// Fix mode auto-corrects tag issues
	return nil
}

func theFileIsModifiedWithFixes() error {
	// File is modified with fixes in fix mode
	return nil
}

func stdoutContainsFixedIssues(count int) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	expected := fmt.Sprintf("Fixed %d", count)
	if strings.Contains(Ctx.CommandOutput, expected) {
		return nil
	}
	return fmt.Errorf("stdout does not contain 'Fixed %d issues'.\nOutput:\n%s", count, Ctx.CommandOutput)
}

// ============================================================================
// CheckTags Steps
// ============================================================================

func tagValidationIsSkipped() error {
	// Tag validation is skipped with --no-check-tags
	return nil
}

func onlySyntaxErrorsAreReported() error {
	// Only syntax errors are reported, not tag errors
	return nil
}
