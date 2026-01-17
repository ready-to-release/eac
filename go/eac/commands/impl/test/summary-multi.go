// Internal helper: Generate test summary markdown from multiple module test results
package test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/commands/impl/test/internal/cucumber"
	"github.com/ready-to-release/eac/go/eac/commands/impl/test/testers"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// generateSummaryMulti generates a multi-module summary.md from test-run-id directory.
func generateSummaryMulti(testRunID string) error {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	// Load config for path resolution
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Construct paths
	testRunDir := filepath.Join(cfg.Repository.TestOutputDirAbs(workspaceRoot), testRunID)
	summaryPath := filepath.Join(testRunDir, "summary.md")

	// Check if test-run-id directory exists
	if _, err := os.Stat(testRunDir); os.IsNotExist(err) {
		return fmt.Errorf("test run directory not found: %s", testRunDir)
	}

	// Find all module directories (subdirectories with cucumber.json)
	modules, err := testers.FindModulesWithResults(testRunDir)
	if err != nil {
		return fmt.Errorf("failed to find module results: %w", err)
	}

	if len(modules) == 0 {
		return fmt.Errorf("no module results found in %s", testRunDir)
	}

	log.Infof("Found %d module(s) with test results", len(modules))

	// Generate multi-module summary
	var summary string
	summary += "# Test Summary\n\n"
	summary += fmt.Sprintf("**Test Run ID**: %s\n\n", testRunID)

	// Render each module as a section
	for i, moduleName := range modules {
		log.Infof("Processing module: %s", moduleName)

		moduleDir := filepath.Join(testRunDir, moduleName)
		cucumberPath := filepath.Join(moduleDir, "cucumber.json")

		// Parse cucumber.json
		report, err := cucumber.ParseFile(cucumberPath)
		if err != nil {
			log.Errorf("Warning: failed to parse %s: %v", cucumberPath, err)
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
	if err := os.WriteFile(summaryPath, []byte(summary), 0o644); err != nil {
		return fmt.Errorf("failed to write summary.md: %w", err)
	}

	log.Infof("Generated: %s", summaryPath)
	return nil
}
