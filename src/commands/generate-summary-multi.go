// Command: generate summary multi
// Description: Generate test summary markdown from multiple module test results
// Usage: generate summary multi <test-run-id>
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/internal/reports/cucumber"
)

func init() {
	Register("generate summary multi", GenerateSummaryMulti)
}

// GenerateSummaryMulti generates a multi-module summary.md from test-run-id directory
func GenerateSummaryMulti() int {
	// Parse arguments
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "Error: missing test-run-id\n")
		fmt.Fprintf(os.Stderr, "Usage: generate summary multi <test-run-id>\n")
		return 1
	}

	testRunID := os.Args[4]

	// Get repository root (two levels up from src/commands)
	workspaceRoot, err := filepath.Abs("../..")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to determine workspace root: %v\n", err)
		return 1
	}

	// Construct paths
	testRunDir := filepath.Join(workspaceRoot, "out", "test-results", testRunID)
	summaryPath := filepath.Join(testRunDir, "summary.md")

	// Check if test-run-id directory exists
	if _, err := os.Stat(testRunDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: test run directory not found: %s\n", testRunDir)
		return 1
	}

	// Find all module directories (subdirectories with cucumber.json)
	modules, err := findModulesWithResults(testRunDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find module results: %v\n", err)
		return 1
	}

	if len(modules) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no module results found in %s\n", testRunDir)
		return 1
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
		fmt.Fprintf(os.Stderr, "Error: failed to write summary.md: %v\n", err)
		return 1
	}

	fmt.Printf("✅ Generated: %s\n", summaryPath)
	return 0
}
