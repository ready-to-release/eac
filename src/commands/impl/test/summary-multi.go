// Internal helper: Generate test summary markdown from multiple module test results
package test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/ready-to-release/eac/src/commands/impl/test/internal/cucumber"
)

// generateSummaryMulti generates a multi-module summary.md from test-run-id directory
func generateSummaryMulti(testRunID string) error {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	// Construct paths
	testRunDir := filepath.Join(workspaceRoot, "out", "test-results", testRunID)
	summaryPath := filepath.Join(testRunDir, "summary.md")

	// Check if test-run-id directory exists
	if _, err := os.Stat(testRunDir); os.IsNotExist(err) {
		return fmt.Errorf("test run directory not found: %s", testRunDir)
	}

	// Find all module directories (subdirectories with cucumber.json)
	modules, err := FindModulesWithResults(testRunDir)
	if err != nil {
		return fmt.Errorf("failed to find module results: %w", err)
	}

	if len(modules) == 0 {
		return fmt.Errorf("no module results found in %s", testRunDir)
	}

	fmt.Printf("Found %d module(s) with test results\n", len(modules))

	// Generate multi-module summary
	var summary string
	summary += "# Test Summary\n\n"
	summary += fmt.Sprintf("**Test Run ID**: %s\n\n", testRunID)

	// Render each module as a section
	for i, moduleName := range modules {
		fmt.Printf("Processing module: %s\n", moduleName)

		moduleDir := filepath.Join(testRunDir, moduleName)
		cucumberPath := filepath.Join(moduleDir, "cucumber.json")

		// Parse cucumber.json
		report, err := cucumber.ParseFile(cucumberPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", cucumberPath, err)
			continue
		}

		// Add module section header
		summary += fmt.Sprintf("### Module: %s\n\n", moduleName)

		// Render features for this module
		summary += cucumber.RenderAllFeatures(report, nil)

		// Add separator between modules (but not after the last one)
		if i < len(modules)-1 {
			summary += "\n---\n\n"
		}
	}

	// Add Appendix A with all specifications
	summary += "\n---\n\n"
	summary += "## Appendix A: Specifications and Test Results\n\n"

	for _, moduleName := range modules {
		moduleDir := filepath.Join(testRunDir, moduleName)
		cucumberPath := filepath.Join(moduleDir, "cucumber.json")

		report, err := cucumber.ParseFile(cucumberPath)
		if err != nil {
			continue
		}

		// Render appendix for this module
		summary += cucumber.RenderAppendixA(report, workspaceRoot)
	}

	// Write summary.md
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		return fmt.Errorf("failed to write summary.md: %w", err)
	}

	fmt.Printf("✅ Generated: %s\n", summaryPath)
	return nil
}
