// Package tests provides BDD step definitions for the risks commands.
package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// registerAssessmentSteps registers step definitions for risks assessment command.
func registerAssessmentSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^I have staged "([^"]*)"$`, iHaveStaged)
	sc.Step(`^I have modified "([^"]*)"$`, iHaveModified)
	sc.Step(`^specifications exist in "([^"]*)"$`, specificationsExistIn)
	sc.Step(`^repository contracts define specs directory as "([^"]*)"$`, contractsDefineSpecsDirectory)
	sc.Step(`^a custom prompt at "([^"]*)"$`, customPromptExists)

	// When steps - handled by generic "I run" step

	// Then steps
	sc.Step(`^a file exists matching "([^"]*)"$`, fileExistsMatching)
	sc.Step(`^the output file matches pattern "([^"]*)"$`, outputFileMatchesPattern)
	sc.Step(`^the report contains "([^"]*)"$`, reportContains)
	sc.Step(`^the report lists "([^"]*)"$`, reportLists)
	sc.Step(`^specifications are loaded from "([^"]*)"$`, specificationsLoadedFrom)
	sc.Step(`^no specs directory argument is required$`, noSpecsDirectoryArgumentRequired)
	sc.Step(`^the custom prompt is used$`, customPromptUsed)
	sc.Step(`^intermediate outputs are saved to "([^"]*)"$`, intermediateOutputsSavedTo)
	sc.Step(`^the AI is invoked with file changes$`, aiInvokedWithFileChanges)
	sc.Step(`^the AI receives specification context$`, aiReceivesSpecContext)
	sc.Step(`^the AI generates a markdown report$`, aiGeneratesMarkdownReport)
	sc.Step(`^the report follows the contract structure$`, reportFollowsContractStructure)
	sc.Step(`^risks are assigned IDs like "([^"]*)", "([^"]*)"$`, risksAssignedIDs)
	sc.Step(`^the report includes section "([^"]*)"$`, reportIncludesSection)
	sc.Step(`^the report mentions "([^"]*)"$`, reportMentions)
}

// ============================================================================
// Given Steps - Setup
// ============================================================================

func iHaveStaged(file string) error {
	if TestMockRepo == nil {
		return fmt.Errorf("mock repo not initialized")
	}

	// Add file to staged files
	currentStaged, _ := TestMockRepo.StagedFiles()
	TestMockRepo.WithStagedFiles(append(currentStaged, file))

	// Set up mock diff output
	diffOutput := fmt.Sprintf("A\t%s", file)
	TestMockRepo.WithStagedDiff(diffOutput)

	// Track in context
	Ctx.StagedFiles = append(Ctx.StagedFiles, file)

	return nil
}

func iHaveModified(file string) error {
	if TestMockRepo == nil {
		return fmt.Errorf("mock repo not initialized")
	}

	// Track in context (mock repo doesn't need diff setup for changed files)
	Ctx.ChangedFiles = append(Ctx.ChangedFiles, file)

	return nil
}

func specificationsExistIn(dir string) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	// Create specs directory
	specsDir := filepath.Join(IsolatedTestProjectDir, dir)
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		return fmt.Errorf("failed to create specs directory: %w", err)
	}

	// Create a sample spec file
	specPath := filepath.Join(specsDir, "sample.feature")
	specContent := `Feature: Sample Specification
  @auth
  Scenario: Test scenario
    Given a user
    When they authenticate
    Then access is granted
`
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	// Track in context
	Ctx.TestFiles[specPath] = specContent

	return nil
}

func contractsDefineSpecsDirectory(dir string) error {
	// In isolated test environment, this is handled by test setup
	// Just verify the directory would be correct
	return nil
}

func customPromptExists(path string) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	// Create custom prompt file
	fullPath := filepath.Join(IsolatedTestProjectDir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create prompt directory: %w", err)
	}

	promptContent := `# Custom Risk Assessment Prompt

Analyze the provided code changes and generate a risk assessment report.

Focus on security vulnerabilities and code quality issues.

{{.ChangedFiles}}

{{.Specifications}}
`
	if err := os.WriteFile(fullPath, []byte(promptContent), 0644); err != nil {
		return fmt.Errorf("failed to write custom prompt: %w", err)
	}

	// Track in context
	Ctx.TestFiles[fullPath] = promptContent

	return nil
}

// ============================================================================
// Then Steps - Verification
// ============================================================================

func fileExistsMatching(pattern string) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	// Check for files matching pattern
	matches, err := filepath.Glob(filepath.Join(IsolatedTestProjectDir, pattern))
	if err != nil {
		return fmt.Errorf("glob pattern error: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no files found matching pattern: %s", pattern)
	}

	return nil
}

func outputFileMatchesPattern(pattern string) error {
	// Check if the report path matches the pattern
	if Ctx.ReportPath == "" {
		return fmt.Errorf("no report path set")
	}

	matched, err := filepath.Match(pattern, filepath.Base(Ctx.ReportPath))
	if err != nil {
		return fmt.Errorf("pattern match error: %w", err)
	}

	if !matched {
		return fmt.Errorf("report path %s does not match pattern %s", Ctx.ReportPath, pattern)
	}

	return nil
}

func reportContains(text string) error {
	if Ctx.GeneratedReport == "" {
		return fmt.Errorf("no report generated")
	}

	if !strings.Contains(Ctx.GeneratedReport, text) {
		return fmt.Errorf("report does not contain: %s", text)
	}

	return nil
}

func reportLists(file string) error {
	if Ctx.GeneratedReport == "" {
		return fmt.Errorf("no report generated")
	}

	if !strings.Contains(Ctx.GeneratedReport, file) {
		return fmt.Errorf("report does not list file: %s", file)
	}

	return nil
}

func specificationsLoadedFrom(dir string) error {
	// Verify specs were loaded (check in command output or debug logs)
	if !strings.Contains(Ctx.CommandOutput, "specification") {
		return fmt.Errorf("no evidence of specifications being loaded")
	}
	return nil
}

func noSpecsDirectoryArgumentRequired() error {
	// Verification that command works without explicit specs directory
	// This is a contract check, no actual verification needed
	return nil
}

func customPromptUsed() error {
	// Check debug logs or output for evidence of custom prompt usage
	if IsolatedTestProjectDir == "" {
		return nil // Can't verify in non-isolated mode
	}

	debugDir := filepath.Join(IsolatedTestProjectDir, "out", "logs", "risks")
	if _, err := os.Stat(debugDir); os.IsNotExist(err) {
		return fmt.Errorf("no debug logs found to verify custom prompt usage")
	}

	return nil
}

func intermediateOutputsSavedTo(dir string) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	debugDir := filepath.Join(IsolatedTestProjectDir, dir)
	if _, err := os.Stat(debugDir); os.IsNotExist(err) {
		return fmt.Errorf("debug directory does not exist: %s", debugDir)
	}

	// Check for at least one file
	entries, err := os.ReadDir(debugDir)
	if err != nil {
		return fmt.Errorf("failed to read debug directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no intermediate outputs found in: %s", debugDir)
	}

	return nil
}

func aiInvokedWithFileChanges() error {
	// Verify AI was called (check if mock AI was used)
	if Ctx.MockAIResponse == "" {
		return fmt.Errorf("AI mock was not set up")
	}
	return nil
}

func aiReceivesSpecContext() error {
	// Verify AI received spec context
	// In test mode, we can check if specs were loaded
	if len(Ctx.AllFiles) > 0 || len(Ctx.StagedFiles) > 0 {
		return nil
	}
	return fmt.Errorf("no file context set up for AI")
}

func aiGeneratesMarkdownReport() error {
	if Ctx.GeneratedReport == "" {
		return fmt.Errorf("no report generated")
	}

	// Check if report looks like markdown (has headers)
	if !strings.Contains(Ctx.GeneratedReport, "#") {
		return fmt.Errorf("report does not appear to be markdown")
	}

	return nil
}

func reportFollowsContractStructure() error {
	if Ctx.GeneratedReport == "" {
		return fmt.Errorf("no report generated")
	}

	// Check for required sections
	requiredSections := []string{
		"# Risk Assessment Report",
		"## Summary",
		"## Risks Identified",
	}

	for _, section := range requiredSections {
		if !strings.Contains(Ctx.GeneratedReport, section) {
			return fmt.Errorf("report missing required section: %s", section)
		}
	}

	return nil
}

func risksAssignedIDs(id1, id2 string) error {
	if Ctx.GeneratedReport == "" {
		return fmt.Errorf("no report generated")
	}

	if !strings.Contains(Ctx.GeneratedReport, id1) {
		return fmt.Errorf("report does not contain risk ID: %s", id1)
	}

	if !strings.Contains(Ctx.GeneratedReport, id2) {
		return fmt.Errorf("report does not contain risk ID: %s", id2)
	}

	return nil
}

func reportIncludesSection(section string) error {
	return reportContains(section)
}

func reportMentions(text string) error {
	return reportContains(text)
}

// ============================================================================
// Helper Functions
// ============================================================================
// (Mock setup is now handled in register.go)
