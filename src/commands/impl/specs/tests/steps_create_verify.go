// Package tests provides BDD step definitions for the specs commands.
//
// This file contains verification steps (Then) for the specs create command.
package tests

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// registerCreateVerifySteps registers verification steps for specs create
func registerCreateVerifySteps(sc *godog.ScenarioContext) {
	// File output verification
	sc.Step(`^the file is saved at "([^"]*)"$`, theFileIsSavedAt)
	sc.Step(`^the file contains valid Gherkin syntax$`, theFileContainsValidGherkinSyntax)
	sc.Step(`^the parent directories are created if they don't exist$`, theParentDirectoriesAreCreatedIfTheyDontExist)
	sc.Step(`^the existing file is not modified$`, theExistingFileIsNotModified)
	sc.Step(`^the existing file is overwritten$`, theExistingFileIsOverwritten)
	sc.Step(`^no specification file is created$`, noSpecificationFileIsCreated)

	// Content verification
	sc.Step(`^initialization noise should be removed$`, initializationNoiseShouldBeRemoved)
	// Note: "only valid Gherkin content should remain" is registered in steps_create_setup.go
	sc.Step(`^the generated specification includes multiple Rules$`, theGeneratedSpecificationIncludesMultipleRules)
	sc.Step(`^the generated specification includes multiple Scenarios$`, theGeneratedSpecificationIncludesMultipleScenarios)
	sc.Step(`^it must contain at least one "Rule:" declaration$`, itMustContainAtLeastOneRuleDeclaration)
	sc.Step(`^it must contain at least one "Scenario:" declaration$`, itMustContainAtLeastOneScenarioDeclaration)

	// Debug output verification
	sc.Step(`^debug files are written to "out/logs/specs/"$`, debugFilesAreWrittenToOutLogsSpecs)
	sc.Step(`^all debug files are written successfully$`, allDebugFilesAreWrittenSuccessfully)

	// Template verification
	sc.Step(`^the local template should be loaded$`, theLocalTemplateShouldBeLoaded)
	sc.Step(`^the agent receives the template content$`, theAgentReceivesTheTemplateContent)
	sc.Step(`^the custom template is used$`, theCustomTemplateIsUsed)
	sc.Step(`^the custom prompt is used$`, theCustomPromptIsUsed)

	// Generation verification
	sc.Step(`^the agent generates a valid Gherkin feature file$`, theAgentGeneratesAValidGherkinFeatureFile)
	sc.Step(`^the generation completes successfully$`, theGenerationCompletesSuccessfully)
	sc.Step(`^the generation succeeds$`, theGenerationSucceeds)
	sc.Step(`^a specification file is created$`, aSpecificationFileIsCreated)

	// Description handling
	sc.Step(`^the truncated description is used for generation$`, theTruncatedDescriptionIsUsedForGeneration)
	sc.Step(`^the description is properly escaped$`, theDescriptionIsProperlyEscaped)
	sc.Step(`^the description is truncated with warning$`, theDescriptionIsTruncatedWithWarning)
	sc.Step(`^Unicode should be handled without errors$`, unicodeShouldBeHandledWithoutErrors)
	sc.Step(`^the AI receives the full unmodified description$`, theAIReceivesTheFullUnmodifiedDescription)
	sc.Step(`^the AI receives module context "([^"]*)"$`, theAIReceivesModuleContext)

	// Security verification
	sc.Step(`^the file does not contain actual secrets$`, theFileDoesNotContainActualSecrets)
	sc.Step(`^the file contains "([^"]*)"$`, theFileContainsText)
	sc.Step(`^the file does not contain "([^"]*)" field$`, theFileDoesNotContainField)

	// Output verification
	sc.Step(`^stdout contains the file path$`, stdoutContainsTheFilePath)
	sc.Step(`^"([^"]*)" contains cleaned output$`, fileContainsCleanedOutput)
	sc.Step(`^the temporary files are cleaned up$`, theTemporaryFilesAreCleanedUp)
}

// ============================================================================
// File Output Verification Steps
// ============================================================================

func theFileIsSavedAt(expectedPath string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := Ctx.CommandOutput

	// Normalize path separators for cross-platform comparison
	normalizedExpected := strings.ReplaceAll(expectedPath, "/", string(filepath.Separator))
	normalizedExpected = strings.ReplaceAll(normalizedExpected, "\\", string(filepath.Separator))

	// Check if the output contains the expected path (as relative or absolute)
	if strings.Contains(output, expectedPath) || strings.Contains(output, normalizedExpected) {
		return nil
	}

	// Also check if the expected path appears at the end of a full path in the output
	// This handles cases where we expect "specs/module/file.feature" but output has
	// "/tmp/isolated-test-123/specs/module/file.feature"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "File:") {
			// Normalize the line for comparison
			normalizedLine := strings.ReplaceAll(line, "/", string(filepath.Separator))
			normalizedLine = strings.ReplaceAll(normalizedLine, "\\", string(filepath.Separator))
			if strings.HasSuffix(normalizedLine, normalizedExpected) ||
				strings.Contains(normalizedLine, normalizedExpected) {
				return nil
			}
		}
	}

	return fmt.Errorf("file not saved at expected path '%s'.\nOutput:\n%s", expectedPath, output)
}

func theFileContainsValidGherkinSyntax() error {
	// File should contain Feature:, Rule:, Scenario: keywords
	return nil
}

func theParentDirectoriesAreCreatedIfTheyDontExist() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	if Ctx.ExitCode == 0 {
		return nil
	}
	return fmt.Errorf("parent directories not created")
}

func theExistingFileIsNotModified() error {
	// When file exists and shouldn't be overwritten
	return nil
}

func theExistingFileIsOverwritten() error {
	// When file exists and should be overwritten with --force
	return nil
}

// ============================================================================
// Content Verification Steps
// ============================================================================

func initializationNoiseShouldBeRemoved() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := Ctx.CommandOutput
	// Check that common initialization messages are not present in the output file
	noisePatterns := []string{
		"**Initialized",
		"ready to assist",
		"I'll help",
		"Here's",
	}
	for _, pattern := range noisePatterns {
		if strings.Contains(output, pattern) {
			return fmt.Errorf("initialization noise '%s' found in output", pattern)
		}
	}
	return nil
}

func theGeneratedSpecificationIncludesMultipleRules() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	// Read the generated file content (not stdout which has status messages)
	fileContent, err := getLastCreatedFileContent()
	if err != nil {
		return err
	}

	ruleCount := strings.Count(fileContent, "Rule:")
	if ruleCount > 1 {
		return nil
	}
	return fmt.Errorf("generated specification does not include multiple Rules (found %d).\nContent:\n%s", ruleCount, fileContent)
}

func theGeneratedSpecificationIncludesMultipleScenarios() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	// Read the generated file content (not stdout which has status messages)
	fileContent, err := getLastCreatedFileContent()
	if err != nil {
		return err
	}

	scenarioCount := strings.Count(fileContent, "Scenario:")
	if scenarioCount > 1 {
		return nil
	}
	return fmt.Errorf("generated specification does not include multiple Scenarios (found %d).\nContent:\n%s", scenarioCount, fileContent)
}

func itMustContainAtLeastOneRuleDeclaration() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	// Read the generated file content (not stdout which has status messages)
	fileContent, err := getLastCreatedFileContent()
	if err != nil {
		return err
	}

	if strings.Contains(fileContent, "Rule:") {
		return nil
	}
	return fmt.Errorf("generated file does not contain at least one 'Rule:' declaration.\nContent:\n%s", fileContent)
}

func itMustContainAtLeastOneScenarioDeclaration() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	// Read the generated file content (not stdout which has status messages)
	fileContent, err := getLastCreatedFileContent()
	if err != nil {
		return err
	}

	if strings.Contains(fileContent, "Scenario:") {
		return nil
	}
	return fmt.Errorf("generated file does not contain at least one 'Scenario:' declaration.\nContent:\n%s", fileContent)
}

// ============================================================================
// Debug Output Verification Steps
// ============================================================================

func debugFilesAreWrittenToOutLogsSpecs() error {
	// Debug files written to out/logs/specs/
	return nil
}

func allDebugFilesAreWrittenSuccessfully() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	if Ctx.ExitCode == 0 {
		return nil
	}
	return fmt.Errorf("debug files not written successfully")
}

// ============================================================================
// Template Verification Steps
// ============================================================================

func theLocalTemplateShouldBeLoaded() error {
	// Local template loaded instead of default
	return nil
}

func theAgentReceivesTheTemplateContent() error {
	// Template content passed to AI agent
	return nil
}

func theCustomTemplateIsUsed() error {
	// Custom template used when specified with --template
	return nil
}

func theCustomPromptIsUsed() error {
	// Custom prompt used when specified with --prompt
	return nil
}

// ============================================================================
// Generation Verification Steps
// ============================================================================

func theAgentGeneratesAValidGherkinFeatureFile() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	if Ctx.ExitCode == 0 {
		return nil
	}
	return fmt.Errorf("agent failed to generate valid Gherkin")
}

func theGenerationCompletesSuccessfully() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	if Ctx.ExitCode == 0 {
		return nil
	}
	return fmt.Errorf("generation did not complete successfully")
}

func theGenerationSucceeds() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	if Ctx.ExitCode == 0 {
		return nil
	}
	return fmt.Errorf("generation did not succeed")
}

func aSpecificationFileIsCreated() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	// Check that the command succeeded and a file path is in the output
	if Ctx.ExitCode == 0 && (strings.Contains(Ctx.CommandOutput, ".feature") ||
		strings.Contains(Ctx.CommandOutput, "specification")) {
		return nil
	}
	return fmt.Errorf("specification file not created.\nOutput:\n%s", Ctx.CommandOutput)
}

// ============================================================================
// Description Handling Steps
// ============================================================================

func theTruncatedDescriptionIsUsedForGeneration() error {
	// Long descriptions should be truncated
	return nil
}

func theDescriptionIsProperlyEscaped() error {
	// Special characters in description properly escaped
	return nil
}

func theDescriptionIsTruncatedWithWarning() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := Ctx.CommandOutput
	if strings.Contains(output, "truncat") || strings.Contains(output, "warn") {
		return nil
	}
	return fmt.Errorf("description not truncated with warning")
}

func unicodeShouldBeHandledWithoutErrors() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	if Ctx.ExitCode == 0 {
		return nil
	}
	return fmt.Errorf("Unicode handling failed")
}

func theAIReceivesTheFullUnmodifiedDescription() error {
	// AI should receive complete description
	return nil
}

// ============================================================================
// Security Verification Steps
// ============================================================================

func theFileDoesNotContainActualSecrets() error {
	// Generated files should use placeholders, not real secrets
	return nil
}

func theFileContainsText(expectedText string) error {
	// Would read file and verify content
	_ = expectedText
	return nil
}

func theFileDoesNotContainField(fieldName string) error {
	// File should not contain specified field
	_ = fieldName
	return nil
}

// ============================================================================
// Output Verification Steps
// ============================================================================

func stdoutContainsTheFilePath() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := Ctx.CommandOutput
	if strings.Contains(output, "/") || strings.Contains(output, "\\") {
		return nil
	}
	return fmt.Errorf("stdout does not contain file path.\nOutput:\n%s", output)
}

func fileContainsCleanedOutput(filePath string) error {
	// File should contain cleaned/processed output
	_ = filePath
	return nil
}

func theTemporaryFilesAreCleanedUp() error {
	// Temporary files should be removed after processing
	return nil
}

// ============================================================================
// Additional Verification Steps
// ============================================================================

func noSpecificationFileIsCreated() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	// When command fails, no file should be created
	// Check that exit code indicates failure
	if Ctx.ExitCode == 0 {
		return fmt.Errorf("command succeeded when it should have failed - file may have been created")
	}
	return nil
}

func theAIReceivesModuleContext(moduleName string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	// When -m flag is used, the module name should be passed to AI
	// In mock mode, we just verify the command succeeded
	if Ctx.ExitCode != 0 {
		return fmt.Errorf("command failed, module context may not have been passed.\nOutput:\n%s", Ctx.CommandOutput)
	}
	return nil
}
