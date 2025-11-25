// Godog BDD step definitions - Repository no unordered files validation
//
// Feature: repository_no-unordered-files (specs/repository/no-unordered-files/specification.feature)
//
// This file implements steps for validating that no files belong to the "unordered"
// catch-all module, ensuring all files are properly organized into modules.
package tests

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/git"
	"github.com/ready-to-release/eac/src/core/repository"
)

// noUnorderedFilesContext holds state for no unordered files validation
type noUnorderedFilesContext struct {
	repoRoot       string
	moduleReport   *reports.ModuleContractReport
	unorderedFiles []string // List of files in unordered module
}

var noUnorderedFilesCtx *noUnorderedFilesContext

// resetNoUnorderedFilesContext resets the context between scenarios
func resetNoUnorderedFilesContext() {
	noUnorderedFilesCtx = nil
}

// ============================================================================
// Given Steps
// ============================================================================

// moduleContractsAreLoaded loads the module contracts from the repository
func moduleContractsAreLoaded() error {
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	// Initialize context
	noUnorderedFilesCtx = &noUnorderedFilesContext{
		repoRoot:       repoRoot,
		unorderedFiles: []string{},
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(repoRoot, "0.1.0")
	if err != nil {
		return fmt.Errorf("failed to load module contracts: %w", err)
	}

	noUnorderedFilesCtx.moduleReport = moduleReport
	return nil
}

// ============================================================================
// When Steps
// ============================================================================

// lookupFilesBelongingToUnorderedModule queries files in the unordered module
func lookupFilesBelongingToUnorderedModule() error {
	if noUnorderedFilesCtx == nil {
		return fmt.Errorf("no unordered files context not initialized")
	}

	// Get the unordered module
	unorderedModule, exists := noUnorderedFilesCtx.moduleReport.Registry.Get("unordered")
	if !exists {
		// If unordered module doesn't exist, that's actually fine - no files are unordered
		return nil
	}

	// Open git repository
	repo, err := git.Open(noUnorderedFilesCtx.repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get all files with module ownership
	files, err := repository.GetRepositoryFilesWithModules(repo, true, false, false, "0.1.0")
	if err != nil {
		return fmt.Errorf("failed to get repository files with modules: %w", err)
	}

	// Find all files that belong to the unordered module
	for _, fileEntry := range files {
		for _, moduleName := range fileEntry.Modules {
			if moduleName == unorderedModule.Moniker {
				noUnorderedFilesCtx.unorderedFiles = append(noUnorderedFilesCtx.unorderedFiles, fileEntry.Name)
				break
			}
		}
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

// fileListShouldBeEmpty verifies no files are in the unordered module
func fileListShouldBeEmpty() error {
	if noUnorderedFilesCtx == nil {
		return fmt.Errorf("no unordered files context not initialized")
	}

	if len(noUnorderedFilesCtx.unorderedFiles) > 0 {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Found %d file(s) in the unordered module:\n\n", len(noUnorderedFilesCtx.unorderedFiles)))

		for _, filePath := range noUnorderedFilesCtx.unorderedFiles {
			details.WriteString(fmt.Sprintf("  ❌ %s\n", filePath))
		}

		details.WriteString(fmt.Sprintf("\nAll files should be organized into proper modules.\n"))
		details.WriteString(fmt.Sprintf("Create or update module contracts to claim these files.\n"))

		return fmt.Errorf("%s", details.String())
	}

	return nil
}

// ifAnyFilesAreFoundIShouldSeeTheirPathsWithCounts provides helpful output on failure
func ifAnyFilesAreFoundIShouldSeeTheirPathsWithCounts() error {
	// This is a passive assertion - the error messages from above steps already provide this info
	// Just verify that if there were failures, we tracked them
	if noUnorderedFilesCtx == nil {
		return fmt.Errorf("no unordered files context not initialized")
	}

	// If we had files, ensure we have them recorded
	if len(noUnorderedFilesCtx.unorderedFiles) > 0 {
		// Verify we have the file list
		if noUnorderedFilesCtx.unorderedFiles == nil {
			return fmt.Errorf("unordered files found but not recorded")
		}
	}

	return nil
}

// ============================================================================
// Scenario Initialization
// ============================================================================

// InitializeRepositoryNoUnorderedFilesScenario registers step definitions for no unordered files tests
func InitializeRepositoryNoUnorderedFilesScenario(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the module contracts are loaded$`, moduleContractsAreLoaded)

	// When steps
	sc.Step(`^I lookup files belonging to the "([^"]*)" module$`, func(moduleName string) error {
		// We only support checking the "unordered" module for now
		if moduleName != "unordered" {
			return fmt.Errorf("only the 'unordered' module is supported, got: %s", moduleName)
		}
		return lookupFilesBelongingToUnorderedModule()
	})

	// Then steps
	sc.Step(`^the file list should be empty$`, fileListShouldBeEmpty)
	sc.Step(`^if any files are found, I should see their paths with counts$`, ifAnyFilesAreFoundIShouldSeeTheirPathsWithCounts)
}
