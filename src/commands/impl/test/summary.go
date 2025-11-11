// Internal helper: Generate test summary markdown from cucumber.json results
package test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/ready-to-release/eac/src/commands/impl/test/internal/cucumber"
)

// generateSummary generates a summary.md file from cucumber.json test results for a single module
// Uses legacy path: out/<moniker>/cucumber.json (for --generate-only flag)
func generateSummary(moniker string) error {

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	// Construct paths (legacy format)
	outputDir := filepath.Join(workspaceRoot, "out", moniker)
	return generateSummaryForOutputDir(outputDir)
}

// generateSummaryForOutputDir generates a summary.md file from cucumber.json in the given output directory
func generateSummaryForOutputDir(outputDir string) error {
	// Get repository root for Appendix A file path rendering
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	// Construct paths
	cucumberPath := filepath.Join(outputDir, "cucumber.json")
	summaryPath := filepath.Join(outputDir, "summary.md")

	// Check if cucumber.json exists
	if _, err := os.Stat(cucumberPath); os.IsNotExist(err) {
		return fmt.Errorf("cucumber.json not found: %s", cucumberPath)
	}

	// Parse cucumber.json
	fmt.Printf("📊 Parsing test results: %s\n", cucumberPath)
	report, err := cucumber.ParseFile(cucumberPath)
	if err != nil {
		return fmt.Errorf("failed to parse cucumber.json: %w", err)
	}

	fmt.Printf("📝 Found %d features\n", len(report))

	// Generate summary markdown with Appendix A
	var summary string
	summary += "# Test Summary\n\n"
	summary += cucumber.RenderAllFeatures(report, nil)
	summary += "\n---\n\n"
	summary += cucumber.RenderAppendixA(report, workspaceRoot)

	// Write summary.md
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		return fmt.Errorf("failed to write summary.md: %w", err)
	}

	fmt.Printf("✅ Generated summary: %s\n", summaryPath)
	return nil
}
