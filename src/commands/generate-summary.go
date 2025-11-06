// Command: generate summary
// Description: Generate test summary markdown from cucumber.json results
// Usage: generate summary <moniker>
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/reports/cucumber"
)

func init() {
	Register("generate summary", GenerateSummary)
}

// GenerateSummary generates a summary.md file from cucumber.json test results
func GenerateSummary() int {
	// Parse arguments
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: missing module moniker\n")
		fmt.Fprintf(os.Stderr, "Usage: generate summary <moniker>\n")
		return 1
	}

	moniker := os.Args[3]

	// Get repository root (two levels up from src/commands)
	workspaceRoot, err := filepath.Abs("../..")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to determine workspace root: %v\n", err)
		return 1
	}

	// Construct paths
	outputDir := filepath.Join(workspaceRoot, "out", moniker)
	cucumberPath := filepath.Join(outputDir, "cucumber.json")
	summaryPath := filepath.Join(outputDir, "summary.md")

	// Check if cucumber.json exists
	if _, err := os.Stat(cucumberPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: cucumber.json not found: %s\n", cucumberPath)
		fmt.Fprintf(os.Stderr, "Hint: Run 'test module %s --as-cucumber' first\n", moniker)
		return 1
	}

	// Parse cucumber.json
	fmt.Printf("Parsing test results: %s\n", cucumberPath)
	report, err := cucumber.ParseFile(cucumberPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse cucumber.json: %v\n", err)
		return 1
	}

	fmt.Printf("Found %d features\n", len(report))

	// Generate summary markdown with Appendix A
	var summary string
	summary += "# Test Summary\n\n"
	summary += cucumber.RenderAllFeatures(report, nil)
	summary += "\n---\n\n"
	summary += cucumber.RenderAppendixA(report, workspaceRoot)

	// Write summary.md
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write summary.md: %v\n", err)
		return 1
	}

	fmt.Printf("✅ Generated: %s\n", summaryPath)
	return 0
}
