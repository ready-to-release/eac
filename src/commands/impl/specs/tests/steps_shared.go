// Package tests provides BDD step definitions for the specs commands.
//
// This file contains shared steps used by both create and validate commands.
package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// registerSharedSteps registers steps used by both create and validate commands
func registerSharedSteps(sc *godog.ScenarioContext) {
	// File setup steps
	sc.Step(`^a valid specification file at "([^"]*)"$`, aValidSpecificationFileAt)
	sc.Step(`^a specification file at "([^"]*)"$`, aValidSpecificationFileAt)
	sc.Step(`^a specification file with missing Rule at "([^"]*)"$`, aSpecificationFileWithMissingRuleAt)
	sc.Step(`^a specification file with multiple errors at "([^"]*)"$`, aSpecificationFileWithMultipleErrorsAt)
	sc.Step(`^an empty file at "([^"]*)"$`, anEmptyFileAt)
	sc.Step(`^a file with invalid UTF-8 encoding at "([^"]*)"$`, aFileWithInvalidUTF8EncodingAt)

	// Directory setup steps
	sc.Step(`^multiple specification files in "([^"]*)"$`, multipleSpecificationFilesIn)
	sc.Step(`^(\d+) valid specifications and (\d+) invalid specifications in "([^"]*)"$`, validSpecificationsAndInvalidSpecificationsIn)
	sc.Step(`^an empty directory at "([^"]*)"$`, anEmptyDirectoryAt)
	sc.Step(`^a directory with only \.md files at "([^"]*)"$`, aDirectoryWithOnlyMdFilesAt)

	// Contract setup steps
	sc.Step(`^contract files are missing from "([^"]*)"$`, contractFilesAreMissingFrom)
	// Note: "specification contracts are available" is registered in src/commands/tests/steps_test.go

	// Common verification steps
	// Note: Most verification steps are already registered in src/commands/tests/steps_test.go
	// Only register specs-specific verification steps here
	sc.Step(`^debug messages are shown on stderr$`, debugMessagesAreShownOnStderr)
}

// ============================================================================
// File Setup Steps
// ============================================================================

func aValidSpecificationFileAt(filePath string) error {
	return createTestFile(filePath, validSpecContent())
}

func aSpecificationFileWithMissingRuleAt(filePath string) error {
	return createTestFile(filePath, invalidSpecContent())
}

func aSpecificationFileWithMultipleErrorsAt(filePath string) error {
	return createTestFile(filePath, multipleErrorsSpecContent())
}

func anEmptyFileAt(filePath string) error {
	return createTestFile(filePath, "")
}

func aFileWithInvalidUTF8EncodingAt(filePath string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	fullPath := filepath.Join(Ctx.TestDir, filePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write invalid UTF-8 bytes
	invalidBytes := []byte{0xff, 0xfe, 0xfd}
	if err := os.WriteFile(fullPath, invalidBytes, 0644); err != nil {
		return err
	}

	Ctx.CreatedFiles = append(Ctx.CreatedFiles, fullPath)
	return nil
}

// ============================================================================
// Directory Setup Steps
// ============================================================================

func multipleSpecificationFilesIn(dirPath string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	dir := filepath.Join(Ctx.TestDir, dirPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create 3 valid files with @test-framework tag
	for i := 1; i <= 3; i++ {
		filePath := filepath.Join(dirPath, fmt.Sprintf("test%d.feature", i))
		if err := createTestFile(filePath, validSpecContent()); err != nil {
			return err
		}
	}
	return nil
}

func validSpecificationsAndInvalidSpecificationsIn(validCount, invalidCount int, dirPath string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	dir := filepath.Join(Ctx.TestDir, dirPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create valid files
	for i := 1; i <= validCount; i++ {
		filePath := filepath.Join(dirPath, fmt.Sprintf("valid%d.feature", i))
		if err := createTestFile(filePath, validSpecContent()); err != nil {
			return err
		}
	}

	// Create invalid files
	for i := 1; i <= invalidCount; i++ {
		filePath := filepath.Join(dirPath, fmt.Sprintf("invalid%d.feature", i))
		if err := createTestFile(filePath, invalidSpecContent()); err != nil {
			return err
		}
	}

	return nil
}

func anEmptyDirectoryAt(dirPath string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	fullPath := filepath.Join(Ctx.TestDir, dirPath)
	return os.MkdirAll(fullPath, 0755)
}

func aDirectoryWithOnlyMdFilesAt(dirPath string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	fullPath := filepath.Join(Ctx.TestDir, dirPath)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return err
	}

	// Create some .md files
	mdContent := "# Documentation\n\nThis is a markdown file."
	for i := 1; i <= 3; i++ {
		mdFile := filepath.Join(fullPath, fmt.Sprintf("doc%d.md", i))
		if err := os.WriteFile(mdFile, []byte(mdContent), 0644); err != nil {
			return err
		}
		Ctx.CreatedFiles = append(Ctx.CreatedFiles, mdFile)
	}

	return nil
}

// ============================================================================
// Contract Setup Steps
// ============================================================================

func contractFilesAreMissingFrom(contractPath string) error {
	// For this test, we rely on the command failing to load contracts
	// The actual contracts should exist in the real directory
	if Ctx != nil {
		Ctx.RemovedContracts = true
	}
	return nil
}

func specificationContractsAreAvailable() error {
	// Contracts are available by default in the repository
	return nil
}

// ============================================================================
// Common Verification Steps
// ============================================================================

func theCommandExitsWithCode(expectedCode int) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	if Ctx.ExitCode != expectedCode {
		return fmt.Errorf("expected exit code %d, got %d. Output:\n%s",
			expectedCode, Ctx.ExitCode, Ctx.CommandOutput)
	}
	return nil
}

func stderrContains(text string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	if !strings.Contains(Ctx.CommandOutput, text) {
		return fmt.Errorf("stderr does not contain '%s'.\nOutput:\n%s", text, Ctx.CommandOutput)
	}
	return nil
}

func stdoutContains(text string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	if !strings.Contains(Ctx.CommandOutput, text) {
		return fmt.Errorf("stdout does not contain '%s'.\nOutput:\n%s", text, Ctx.CommandOutput)
	}
	return nil
}

func debugMessagesAreShownOnStderr() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}
	output := Ctx.CommandOutput
	if strings.Contains(output, "DEBUG") || strings.Contains(output, "debug") {
		return nil
	}
	return fmt.Errorf("no debug messages on stderr.\nOutput:\n%s", output)
}
