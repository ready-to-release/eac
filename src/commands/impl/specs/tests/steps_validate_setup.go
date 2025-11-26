// Package tests provides BDD step definitions for the specs commands.
//
// This file contains setup steps (Given/When) for the specs validate command.
package tests

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// registerValidateSetupSteps registers setup steps for specs validate
func registerValidateSetupSteps(sc *godog.ScenarioContext) {
	// Command execution - Note: The general "I run" step in steps_test.go handles command execution
	// Only register specs-validate-specific steps not covered by the general step
	sc.Step(`^I run "specs validate" without arguments$`, iRunSpecsValidateWithoutArguments)

	// Validation-specific file setup
	sc.Step(`^a specification file with naming warnings at "([^"]*)"$`, aSpecificationFileWithNamingWarningsAt)
	sc.Step(`^a specification with feature name "([^"]*)"$`, aSpecificationWithFeatureName)
	sc.Step(`^a specification with scenarios missing verification tags$`, aSpecificationWithScenariosMissingVerificationTags)
	sc.Step(`^a specification with Rule before Feature$`, aSpecificationWithRuleBeforeFeature)
	sc.Step(`^a specification without Rule blocks$`, aSpecificationWithoutRuleBlocks)
	sc.Step(`^a specification without Scenario blocks$`, aSpecificationWithoutScenarioBlocks)
	sc.Step(`^a specification with (\d+) errors and (\d+) warnings$`, aSpecificationWithErrorsAndWarnings)
	sc.Step(`^a specification with errors on lines (\d+), (\d+), and (\d+)$`, aSpecificationWithErrorsOnLines)
	sc.Step(`^a specification with validation errors$`, aSpecificationWithValidationErrors)
	sc.Step(`^a specification with errors and warnings$`, aSpecificationWithErrorsAndWarningsSimple)
	sc.Step(`^specification files in nested directories under "([^"]*)"$`, specificationFilesInNestedDirectoriesUnder)
	sc.Step(`^a directory with \.feature, \.md, and \.txt files$`, aDirectoryWithFeatureMdAndTxtFiles)
	sc.Step(`^a specification file with no read permissions$`, aSpecificationFileWithNoReadPermissions)

	// Fixture setup for new flags
	sc.Step(`^a specification file with errors at "([^"]*)"$`, aSpecificationFileWithErrorsAt)
	sc.Step(`^a specification file with warnings at "([^"]*)"$`, aSpecificationFileWithWarningsAt)
	sc.Step(`^a specification file with fixable tag issues at "([^"]*)"$`, aSpecificationFileWithFixableTagIssuesAt)
	sc.Step(`^a specification file with tag errors at "([^"]*)"$`, aSpecificationFileWithTagErrorsAt)
}

// ============================================================================
// Command Execution Steps
// ============================================================================

func iRunSpecsValidateWithoutArguments() error {
	if RunCommand != nil {
		return RunCommand("specs validate")
	}
	return nil
}

// ============================================================================
// Validation-Specific File Setup Steps
// ============================================================================

func aSpecificationFileWithNamingWarningsAt(filePath string) error {
	return createTestFile(filePath, specWithWarningsContent())
}

func aSpecificationWithFeatureName(featureName string) error {
	content := fmt.Sprintf(`@test-framework @deps:go @ov
Feature: %s

  As a developer
  I want to test
  So that it works

  Background:
    Given setup

  Rule: Test rule

    @L2 @ov
    Scenario: Test
      When action
      Then result
`, featureName)
	return createTestFile("specs/src-commands/test-fixtures/custom-name.feature", content)
}

func aSpecificationWithScenariosMissingVerificationTags() error {
	return createTestFile("specs/test/spec.feature", specMissingVerificationTagsContent())
}

func aSpecificationWithRuleBeforeFeature() error {
	return createTestFile("specs/src-commands/test-fixtures/wrong-order.feature", specWrongOrderContent())
}

func aSpecificationWithoutRuleBlocks() error {
	return createTestFile("specs/test/spec.feature", invalidSpecContent())
}

func aSpecificationWithoutScenarioBlocks() error {
	return createTestFile("specs/src-commands/test-fixtures/no-scenarios.feature", specNoScenariosContent())
}

func aSpecificationWithErrorsAndWarnings(errorCount, warningCount int) error {
	_ = errorCount
	_ = warningCount
	return createTestFile("specs/src-commands/test-fixtures/errors-warnings.feature", multipleErrorsSpecContent())
}

func aSpecificationWithErrorsAndWarningsSimple() error {
	return createTestFile("specs/src-commands/test-fixtures/errors-warnings.feature", multipleErrorsSpecContent())
}

func aSpecificationWithValidationErrors() error {
	return createTestFile("specs/test/spec.feature", multipleErrorsSpecContent())
}

func aSpecificationWithErrorsOnLines(line1, line2, line3 int) error {
	_ = line1
	_ = line2
	_ = line3
	return createTestFile("specs/src-commands/test-fixtures/multi-line-errors.feature", multipleErrorsSpecContent())
}

func specificationFilesInNestedDirectoriesUnder(rootPath string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	// Create nested directory structure with .feature files
	dir1 := filepath.Join(rootPath, "level1")
	dir2 := filepath.Join(rootPath, "level1/level2")

	os.MkdirAll(filepath.Join(Ctx.TestDir, dir1), 0755)
	os.MkdirAll(filepath.Join(Ctx.TestDir, dir2), 0755)

	createTestFile(filepath.Join(dir1, "test1.feature"), validSpecContent())
	createTestFile(filepath.Join(dir2, "test2.feature"), validSpecContent())
	return nil
}

func aDirectoryWithFeatureMdAndTxtFiles() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	dir := "specs/test/mixed"
	os.MkdirAll(filepath.Join(Ctx.TestDir, dir), 0755)

	createTestFile(filepath.Join(dir, "test.feature"), validSpecContent())
	createTestFile(filepath.Join(dir, "readme.md"), "# README")
	createTestFile(filepath.Join(dir, "notes.txt"), "notes")
	return nil
}

func aSpecificationFileWithNoReadPermissions() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	filePath := "specs/src-commands/test-fixtures/no-read.feature"
	if err := createTestFile(filePath, validSpecContent()); err != nil {
		return err
	}

	// Remove read permissions (chmod 000)
	// Note: On Windows this might not work the same way as Unix
	fullPath := filepath.Join(Ctx.TestDir, filePath)
	if err := os.Chmod(fullPath, 0000); err != nil {
		return fmt.Errorf("failed to remove read permissions: %w", err)
	}

	return nil
}

// ============================================================================
// New Flag Fixture Setup Steps
// ============================================================================

func aSpecificationFileWithErrorsAt(filePath string) error {
	return createTestFile(filePath, multipleErrorsSpecContent())
}

func aSpecificationFileWithWarningsAt(filePath string) error {
	return createTestFile(filePath, specWithWarningsContent())
}

func aSpecificationFileWithFixableTagIssuesAt(filePath string) error {
	return createTestFile(filePath, specFixableTagsContent())
}

func aSpecificationFileWithTagErrorsAt(filePath string) error {
	return createTestFile(filePath, specBadTagsContent())
}
