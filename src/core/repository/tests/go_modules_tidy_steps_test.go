// Godog BDD step definitions - Repository Go modules tidy validation
//
// Feature: repository_go-modules-tidy (specs/repository/go-modules-tidy/specification.feature)
//
// This file implements steps for validating that all Go modules in the repository
// have tidy dependencies (go.mod and go.sum are synchronized).
package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

// goModuleTidyContext holds state for go module tidy validation
type goModuleTidyContext struct {
	repoRoot          string              // Repository root path
	discoveredModules []string            // List of discovered Go module paths
	tidyResults       map[string]string   // Module path -> diff output
	failedModules     []string            // Modules that failed tidy check
}

var goModTidyCtx *goModuleTidyContext

// resetGoModuleTidyContext resets the context between scenarios
func resetGoModuleTidyContext() {
	goModTidyCtx = nil
}

// ============================================================================
// Given Steps
// ============================================================================

// repositoryRootExists verifies the repository root is accessible
func repositoryRootExists() error {
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	if _, err := os.Stat(repoRoot); os.IsNotExist(err) {
		return fmt.Errorf("repository root does not exist: %s", repoRoot)
	}

	// Store for later use
	if goModTidyCtx == nil {
		goModTidyCtx = &goModuleTidyContext{
			repoRoot:          repoRoot,
			discoveredModules: []string{},
			tidyResults:       make(map[string]string),
			failedModules:     []string{},
		}
	} else {
		goModTidyCtx.repoRoot = repoRoot
	}

	return nil
}

// discoverAllGoModulesUsingContracts discovers all Go modules via module contracts
func discoverAllGoModulesUsingContracts() error {
	if goModTidyCtx == nil {
		return fmt.Errorf("context not initialized - call repositoryRootExists first")
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(goModTidyCtx.repoRoot, "0.1.0")
	if err != nil {
		return fmt.Errorf("failed to load module contracts: %w", err)
	}

	// Filter for Go modules (go-cli, go-commands, go-library, go-mcp, etc.)
	for _, module := range moduleReport.Registry.All() {
		if isGoModule(module.Type) {
			modulePath := filepath.Join(goModTidyCtx.repoRoot, module.Source.Root)
			goModTidyCtx.discoveredModules = append(goModTidyCtx.discoveredModules, modulePath)
		}
	}

	if len(goModTidyCtx.discoveredModules) == 0 {
		return fmt.Errorf("no Go modules discovered in repository")
	}

	return nil
}

// isGoModule checks if a module type is a Go module
func isGoModule(moduleType string) bool {
	goModuleTypes := []string{
		"go-cli",
		"go-commands",
		"go-library",
		"go-mcp",
		"go-tests",
	}
	for _, t := range goModuleTypes {
		if moduleType == t {
			return true
		}
	}
	return false
}

// ============================================================================
// When Steps
// ============================================================================

// runGoModTidyDiffInEachModule runs go mod tidy -diff in each discovered module
func runGoModTidyDiffInEachModule() error {
	if goModTidyCtx == nil {
		return fmt.Errorf("go module tidy context not initialized")
	}

	for _, modulePath := range goModTidyCtx.discoveredModules {
		// Run go mod tidy -diff
		cmd := exec.Command("go", "mod", "tidy", "-diff")
		cmd.Dir = modulePath

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		output := stdout.String() + stderr.String()

		// Store result
		goModTidyCtx.tidyResults[modulePath] = output

		// If command failed or has output, mark as failed
		if err != nil || strings.TrimSpace(output) != "" {
			goModTidyCtx.failedModules = append(goModTidyCtx.failedModules, modulePath)
		}
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

// allModulesShouldHaveExitCode0 verifies all modules passed the tidy check
func allModulesShouldHaveExitCode0() error {
	if goModTidyCtx == nil {
		return fmt.Errorf("go module tidy context not initialized")
	}

	if len(goModTidyCtx.failedModules) > 0 {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Found %d module(s) with untidy dependencies:\n\n", len(goModTidyCtx.failedModules)))

		for _, modulePath := range goModTidyCtx.failedModules {
			relPath, _ := filepath.Rel(goModTidyCtx.repoRoot, modulePath)
			details.WriteString(fmt.Sprintf("❌ %s\n", relPath))
			if diff := goModTidyCtx.tidyResults[modulePath]; diff != "" {
				details.WriteString(fmt.Sprintf("   Diff:\n%s\n\n", indent(diff, "   ")))
			}
		}

		details.WriteString(fmt.Sprintf("\nTo fix, run: go mod tidy\n"))
		details.WriteString(fmt.Sprintf("Checked %d Go modules total\n", len(goModTidyCtx.discoveredModules)))

		return fmt.Errorf("%s", details.String())
	}

	return nil
}

// noModuleShouldHaveAnyDiffOutput verifies no modules have diff output
func noModuleShouldHaveAnyDiffOutput() error {
	if goModTidyCtx == nil {
		return fmt.Errorf("go module tidy context not initialized")
	}

	var modulesWithDiff []string
	for modulePath, output := range goModTidyCtx.tidyResults {
		if strings.TrimSpace(output) != "" {
			modulesWithDiff = append(modulesWithDiff, modulePath)
		}
	}

	if len(modulesWithDiff) > 0 {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Found %d module(s) with diff output:\n", len(modulesWithDiff)))
		for _, modulePath := range modulesWithDiff {
			relPath, _ := filepath.Rel(goModTidyCtx.repoRoot, modulePath)
			details.WriteString(fmt.Sprintf("  - %s\n", relPath))
		}
		return fmt.Errorf("%s", details.String())
	}

	return nil
}

// ifAnyModuleIsNotTidyIShouldSeeTheModulePathAndDiff provides helpful output on failure
func ifAnyModuleIsNotTidyIShouldSeeTheModulePathAndDiff() error {
	// This is a passive assertion - the error messages from above steps already provide this info
	// Just verify that if there were failures, we tracked them
	if goModTidyCtx == nil {
		return fmt.Errorf("go module tidy context not initialized")
	}

	// If we had failures, ensure we have details
	if len(goModTidyCtx.failedModules) > 0 {
		for _, modulePath := range goModTidyCtx.failedModules {
			if _, exists := goModTidyCtx.tidyResults[modulePath]; !exists {
				return fmt.Errorf("failed module %s has no recorded results", modulePath)
			}
		}
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// indent adds a prefix to each line of text
func indent(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

// ============================================================================
// Scenario Initialization
// ============================================================================

// InitializeRepositoryGoModulesTidyScenario registers step definitions for go modules tidy tests
func InitializeRepositoryGoModulesTidyScenario(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the repository root exists$`, repositoryRootExists)
	sc.Step(`^I discover all Go modules in the repository using module contracts$`, discoverAllGoModulesUsingContracts)

	// When steps
	sc.Step(`^I run "go mod tidy -diff" in each Go module directory$`, runGoModTidyDiffInEachModule)

	// Then steps
	sc.Step(`^all modules should have exit code 0$`, allModulesShouldHaveExitCode0)
	sc.Step(`^no module should have any diff output$`, noModuleShouldHaveAnyDiffOutput)
	sc.Step(`^if any module is not tidy, I should see the module path and diff$`, ifAnyModuleIsNotTidyIShouldSeeTheModulePathAndDiff)
}
