// Godog BDD step definitions - Repository one module per file validation
//
// Feature: repository_one-module-per-file (specs/repository/one-module-per-file/specification.feature)
//
// This file implements steps for validating that each file belongs to exactly one module,
// ensuring clear ownership and no module overlap.
package tests

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/git"
	"github.com/ready-to-release/eac/src/core/repository"
)

// oneModulePerFileContext holds state for one module per file validation
type oneModulePerFileContext struct {
	repoRoot               string
	moduleReport           *reports.ModuleContractReport
	multiOwnershipFiles    []repository.RepositoryFileWithModule // Files with multiple module owners
}

var oneModulePerFileCtx *oneModulePerFileContext

// resetOneModulePerFileContext resets the context between scenarios
func resetOneModulePerFileContext() {
	oneModulePerFileCtx = nil
}

// ============================================================================
// Given Steps
// ============================================================================

// moduleContractsAreLoadedForOneModule loads the module contracts (shared with no-unordered-files)
// This reuses the same step name, so we need to ensure it works for both contexts

// ============================================================================
// When Steps
// ============================================================================

// checkForFilesWithMultiModuleOwnership finds files owned by multiple modules
func checkForFilesWithMultiModuleOwnership() error {
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	// Initialize context
	oneModulePerFileCtx = &oneModulePerFileContext{
		repoRoot:            repoRoot,
		multiOwnershipFiles: []repository.RepositoryFileWithModule{},
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to load module contracts: %w", err)
	}
	oneModulePerFileCtx.moduleReport = moduleReport

	// Open git repository
	repo, err := git.Open(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get all files with module ownership
	files, err := repository.GetRepositoryFilesWithModules(repo, true, false, false)
	if err != nil {
		return fmt.Errorf("failed to get repository files with modules: %w", err)
	}

	// Find files with multiple module ownership
	for _, file := range files {
		if len(file.Modules) > 1 {
			oneModulePerFileCtx.multiOwnershipFiles = append(oneModulePerFileCtx.multiOwnershipFiles, file)
		}
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

// noFilesShouldBelongToMultipleModules verifies no files have multi-ownership
func noFilesShouldBelongToMultipleModules() error {
	if oneModulePerFileCtx == nil {
		return fmt.Errorf("one module per file context not initialized")
	}

	if len(oneModulePerFileCtx.multiOwnershipFiles) > 0 {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Found %d file(s) with multi-module ownership:\n\n", len(oneModulePerFileCtx.multiOwnershipFiles)))

		for _, file := range oneModulePerFileCtx.multiOwnershipFiles {
			details.WriteString(fmt.Sprintf("  ❌ %s\n", file.Name))
			details.WriteString(fmt.Sprintf("     Owned by modules: %s\n", strings.Join(file.Modules, ", ")))
		}

		details.WriteString(fmt.Sprintf("\nEach file should belong to exactly one module.\n"))
		details.WriteString(fmt.Sprintf("Review module contracts and adjust glob patterns to prevent overlap.\n"))

		return fmt.Errorf("%s", details.String())
	}

	return nil
}

// ifAnyFilesHaveMultiOwnershipIShouldSeeTheirPathsAndConflictingModules provides helpful output
func ifAnyFilesHaveMultiOwnershipIShouldSeeTheirPathsAndConflictingModules() error {
	// This is a passive assertion - the error messages from above steps already provide this info
	// Just verify that if there were failures, we tracked them
	if oneModulePerFileCtx == nil {
		return fmt.Errorf("one module per file context not initialized")
	}

	// If we had multi-ownership files, ensure we have them recorded
	if len(oneModulePerFileCtx.multiOwnershipFiles) > 0 {
		// Verify we have the file list with module details
		if oneModulePerFileCtx.multiOwnershipFiles == nil {
			return fmt.Errorf("multi-ownership files found but not recorded")
		}

		// Verify each file has multiple modules listed
		for _, file := range oneModulePerFileCtx.multiOwnershipFiles {
			if len(file.Modules) <= 1 {
				return fmt.Errorf("file %s marked as multi-ownership but has %d modules", file.Name, len(file.Modules))
			}
		}
	}

	return nil
}

// ============================================================================
// Scenario Initialization
// ============================================================================

// InitializeRepositoryOneModulePerFileScenario registers step definitions for one module per file tests
func InitializeRepositoryOneModulePerFileScenario(sc *godog.ScenarioContext) {
	// Given steps
	// Reuses: sc.Step(`^the module contracts are loaded$`, moduleContractsAreLoaded)
	// This is already registered by no_unordered_files_steps_test.go

	// When steps
	sc.Step(`^I check for files with multi-module ownership$`, checkForFilesWithMultiModuleOwnership)

	// Then steps
	sc.Step(`^no files should belong to multiple modules$`, noFilesShouldBelongToMultipleModules)
	sc.Step(`^if any files have multi-ownership, I should see their paths and conflicting modules$`, ifAnyFilesHaveMultiOwnershipIShouldSeeTheirPathsAndConflictingModules)
}
