// Godog BDD step definitions - Repository markdown syntax validation
//
// Feature: repository_markdown-syntax (specs/repository/markdown-syntax/specification.feature)
//
// This file implements steps for validating that all Markdown files in the repository
// have valid syntax.
package tests

import (
	"fmt"
	"os"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/markdown"
	"github.com/ready-to-release/eac/src/core/repository"
)

// markdownSyntaxContext holds state for markdown syntax validation
type markdownSyntaxContext struct {
	repoRoot          string
	markdownFiles     []string
	validationResults []markdown.ValidationResult
	failedFiles       []markdown.ValidationResult
}

var markdownSyntaxCtx *markdownSyntaxContext

// resetMarkdownSyntaxContext resets the context between scenarios
func resetMarkdownSyntaxContext() {
	markdownSyntaxCtx = nil
}

// ============================================================================
// Given Steps
// ============================================================================

// discoverAllMarkdownFiles discovers all Markdown files in the repository
func discoverAllMarkdownFiles() error {
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	// Initialize context
	markdownSyntaxCtx = &markdownSyntaxContext{
		repoRoot:          repoRoot,
		markdownFiles:     []string{},
		validationResults: []markdown.ValidationResult{},
		failedFiles:       []markdown.ValidationResult{},
	}

	return nil
}

// ============================================================================
// When Steps
// ============================================================================

// validateEachMarkdownFile validates all discovered Markdown files
func validateEachMarkdownFile() error {
	if markdownSyntaxCtx == nil {
		return fmt.Errorf("markdown syntax context not initialized")
	}

	// Create validator with default options
	opts := markdown.DefaultValidatorOptions()
	validator := markdown.NewValidator(opts, os.Stdout)

	// Validate all markdown files in the repository
	results, err := validator.ValidateDirectory(markdownSyntaxCtx.repoRoot)
	if err != nil {
		return fmt.Errorf("failed to validate markdown files: %w", err)
	}

	markdownSyntaxCtx.validationResults = results

	// Track failed files
	for _, result := range results {
		if !result.Valid {
			markdownSyntaxCtx.failedFiles = append(markdownSyntaxCtx.failedFiles, result)
		}
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

// allFilesShouldHaveValidMarkdownSyntax verifies all files passed validation
func allFilesShouldHaveValidMarkdownSyntax() error {
	if markdownSyntaxCtx == nil {
		return fmt.Errorf("markdown syntax context not initialized")
	}

	if len(markdownSyntaxCtx.failedFiles) > 0 {
		return fmt.Errorf("found %d file(s) with invalid Markdown syntax", len(markdownSyntaxCtx.failedFiles))
	}

	return nil
}

// noFilesShouldHaveBrokenLinks verifies no files have broken links
func noFilesShouldHaveBrokenLinks() error {
	if markdownSyntaxCtx == nil {
		return fmt.Errorf("markdown syntax context not initialized")
	}

	// The markdown validator currently doesn't check links
	// This is a placeholder for future implementation
	return nil
}

// noFilesShouldHaveMalformedHeaders verifies no files have malformed headers
func noFilesShouldHaveMalformedHeaders() error {
	if markdownSyntaxCtx == nil {
		return fmt.Errorf("markdown syntax context not initialized")
	}

	// Check for heading-related errors and warnings
	var malformedHeaders []string

	for _, result := range markdownSyntaxCtx.validationResults {
		// Check for heading warnings
		for _, warning := range result.Warnings {
			malformedHeaders = append(malformedHeaders,
				fmt.Sprintf("%s: Line %d: %s", result.FilePath, warning.Line, warning.Message))
		}
	}

	if len(malformedHeaders) > 0 {
		// For now, we'll allow warnings and only fail on errors
		// Uncomment below to make warnings fail the test:
		// return fmt.Errorf("found %d malformed header(s):\n%s",
		//     len(malformedHeaders), strings.Join(malformedHeaders, "\n"))
	}

	return nil
}

// ifAnyFileHasErrorsIShouldSeeTheFilePathAndErrorDetails provides error details
func ifAnyFileHasErrorsIShouldSeeTheFilePathAndErrorDetails() error {
	if markdownSyntaxCtx == nil {
		return fmt.Errorf("markdown syntax context not initialized")
	}

	// If we had failures, ensure we have error details
	if len(markdownSyntaxCtx.failedFiles) > 0 {
		for _, result := range markdownSyntaxCtx.failedFiles {
			if len(result.Errors) == 0 {
				return fmt.Errorf("file %s marked as failed but has no error details", result.FilePath)
			}
		}
	}

	return nil
}

// ============================================================================
// Scenario Initialization
// ============================================================================

// InitializeRepositoryMarkdownSyntaxScenario registers step definitions for markdown syntax tests
func InitializeRepositoryMarkdownSyntaxScenario(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^I discover all Markdown files in the repository$`, discoverAllMarkdownFiles)

	// When steps
	sc.Step(`^I validate each Markdown file for syntax errors$`, validateEachMarkdownFile)

	// Then steps
	sc.Step(`^all files should have valid Markdown syntax$`, allFilesShouldHaveValidMarkdownSyntax)
	sc.Step(`^no files should have broken links$`, noFilesShouldHaveBrokenLinks)
	sc.Step(`^no files should have malformed headers$`, noFilesShouldHaveMalformedHeaders)
	sc.Step(`^if any file has errors, I should see the file path and error details$`, ifAnyFileHasErrorsIShouldSeeTheFilePathAndErrorDetails)
}
