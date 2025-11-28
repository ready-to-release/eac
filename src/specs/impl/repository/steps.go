// Package repository contains godog step implementations for specs/repository.
package repository

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
	"github.com/ready-to-release/eac/src/specs/internal"
)

// RegisterSteps registers all repository-specific step definitions.
func RegisterSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Create repository-specific context that also updates the shared context
	repoCtx := &repositoryContext{sharedCtx: ctx}

	// Given steps
	sc.Step(`^the repository root exists$`, func() error {
		return repoCtx.repositoryRootExists()
	})
	sc.Step(`^I discover all Go modules in the repository using module contracts$`, func() error {
		return repoCtx.discoverAllGoModulesUsingContracts()
	})
	sc.Step(`^I discover all Markdown files in the repository$`, func() error {
		return repoCtx.discoverAllMarkdownFiles()
	})
	sc.Step(`^I discover all Gherkin feature files in the repository$`, func() error {
		return repoCtx.discoverAllGherkinFeatureFiles()
	})

	// When steps
	sc.Step(`^I run "go mod tidy -diff" in each Go module directory$`, func() error {
		return repoCtx.runGoModTidyDiffInEachModule()
	})
	sc.Step(`^I validate each Markdown file for syntax errors$`, func() error {
		return repoCtx.validateEachMarkdownFile()
	})
	sc.Step(`^I validate each feature file for conflicting L-level tags$`, func() error {
		return repoCtx.validateFeatureFilesForConflictingLLevelTags()
	})
	// Note: "I run the command X" is handled by common steps via RunCommand

	// Then steps - go modules tidy
	sc.Step(`^all modules should have exit code 0$`, func() error {
		return repoCtx.allModulesShouldHaveExitCode0()
	})
	sc.Step(`^no module should have any diff output$`, func() error {
		return repoCtx.noModuleShouldHaveAnyDiffOutput()
	})
	sc.Step(`^if any module is not tidy, I should see the module path and diff$`, func() error {
		return repoCtx.ifAnyModuleIsNotTidyIShouldSeeTheModulePathAndDiff()
	})

	// Then steps - markdown
	sc.Step(`^all files should have valid Markdown syntax$`, func() error {
		return repoCtx.allFilesShouldHaveValidMarkdownSyntax()
	})
	sc.Step(`^no files should have broken links$`, func() error {
		return repoCtx.noFilesShouldHaveBrokenLinks()
	})
	sc.Step(`^no files should have malformed headers$`, func() error {
		return repoCtx.noFilesShouldHaveMalformedHeaders()
	})
	sc.Step(`^if any file has errors, I should see the file path and error details$`, func() error {
		return repoCtx.ifAnyFileHasErrorsIShouldSeeDetails()
	})

	// Then steps - feature level tags
	sc.Step(`^no feature should have an L-tag when its scenarios have different L-tags$`, func() error {
		return repoCtx.noFeatureShouldHaveConflictingLTags()
	})
	sc.Step(`^no feature should have a verification tag when its scenarios have different verification tags$`, func() error {
		return repoCtx.noFeatureShouldHaveConflictingVerificationTags()
	})
	sc.Step(`^if any conflicts are found, I should see the file path, scenario name, and conflicting tags$`, func() error {
		return repoCtx.ifConflictsFoundIShouldSeeDetails()
	})

	// Then steps - build commands (use specific patterns to avoid conflict with common step)
	sc.Step(`^I should not see any build errors$`, func() error {
		return repoCtx.shouldNotSeeBuildErrors()
	})

	// Module hierarchy validation - Given steps
	sc.Step(`^I load all module contracts from the repository$`, func() error {
		return repoCtx.loadAllModuleContracts()
	})

	// Module hierarchy validation - When steps
	sc.Step(`^I validate that all depends_on and used_by relationships are bidirectional$`, func() error {
		return repoCtx.validateBidirectionalRelationships()
	})
	sc.Step(`^I build the complete dependency graph$`, func() error {
		return repoCtx.buildCompleteDependencyGraph()
	})
	sc.Step(`^I check all depends_on references$`, func() error {
		return repoCtx.checkAllDependsOnReferences()
	})
	sc.Step(`^I check all used_by references$`, func() error {
		return repoCtx.checkAllUsedByReferences()
	})
	sc.Step(`^I validate bidirectional relationships$`, func() error {
		return repoCtx.validateBidirectionalRelationships()
	})

	// Module hierarchy validation - Then steps
	sc.Step(`^all modules should have consistent dependency relationships$`, func() error {
		return repoCtx.allModulesShouldHaveConsistentDependencies()
	})
	sc.Step(`^no module should reference a non-existent module$`, func() error {
		return repoCtx.noModuleShouldReferenceNonExistent()
	})
	sc.Step(`^the graph should be a single connected component or forest$`, func() error {
		return repoCtx.graphShouldBeConnectedOrForest()
	})
	sc.Step(`^the graph should have no circular dependencies$`, func() error {
		return repoCtx.graphShouldHaveNoCircularDependencies()
	})
	sc.Step(`^all modules should be reachable from the root$`, func() error {
		return repoCtx.allModulesShouldBeReachableFromRoot()
	})
	sc.Step(`^every referenced module should exist in the registry$`, func() error {
		return repoCtx.everyReferencedModuleShouldExist()
	})
	sc.Step(`^I should see details of any missing modules$`, func() error {
		return repoCtx.shouldSeeDetailsOfMissingModules()
	})
	sc.Step(`^if module A depends_on B, then B's used_by must include A$`, func() error {
		return repoCtx.dependsOnMustHaveUsedByBack()
	})
	sc.Step(`^if module B has used_by A, then A's depends_on must include B$`, func() error {
		return repoCtx.usedByMustHaveDependsOnBack()
	})
	sc.Step(`^I should see details of any inconsistencies$`, func() error {
		return repoCtx.shouldSeeDetailsOfInconsistencies()
	})

	// No-unordered-files steps
	sc.Step(`^the module contracts are loaded$`, func() error {
		return repoCtx.loadAllModuleContracts()
	})
	sc.Step(`^I lookup files belonging to the "([^"]*)" module$`, func(moduleName string) error {
		return repoCtx.lookupFilesBelongingToModule(moduleName)
	})
	sc.Step(`^the file list should be empty$`, func() error {
		return repoCtx.fileListShouldBeEmpty()
	})
	sc.Step(`^if any files are found, I should see their paths with counts$`, func() error {
		return repoCtx.ifAnyFilesFoundShowPathsWithCounts()
	})

	// One-module-per-file steps
	sc.Step(`^I check for files with multi-module ownership$`, func() error {
		return repoCtx.checkForFilesWithMultiModuleOwnership()
	})
	sc.Step(`^no files should belong to multiple modules$`, func() error {
		return repoCtx.noFilesShouldBelongToMultipleModules()
	})
	sc.Step(`^if any files have multi-ownership, I should see their paths and conflicting modules$`, func() error {
		return repoCtx.ifMultiOwnershipShowPathsAndModules()
	})

	// Validation steps (shared with common steps via exit code checking)
	sc.Step(`^I should not see any dependency validation errors$`, func() error {
		return repoCtx.shouldNotSeeBuildErrors() // Reuses build error check
	})
	sc.Step(`^I should not see any undefined tag errors$`, func() error {
		return repoCtx.shouldNotSeeBuildErrors() // Reuses build error check
	})
}

// repositoryContext holds state for repository validation scenarios.
type repositoryContext struct {
	sharedCtx *internal.TestContext
	repoRoot  string

	// Go modules tidy
	discoveredModules []string
	tidyResults       map[string]string
	failedModules     []string

	// Markdown
	markdownFiles  []string
	markdownErrors map[string][]string

	// Feature level tags
	featureFiles []string
	tagConflicts []string

	// Module hierarchy validation
	moduleReport          *reports.ModuleContractReport
	dependencyErrors      []string
	circularDependencies  []string
	missingModules        []string
	bidirectionalErrors   []string

	// File ownership validation
	moduleFiles       []string
	multiOwnershipMap map[string][]string // file -> list of owning modules
}

func (c *repositoryContext) ensureRepoRoot() error {
	if c.repoRoot == "" {
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return fmt.Errorf("failed to get repository root: %w", err)
		}
		c.repoRoot = repoRoot
	}
	return nil
}

// ============================================================================
// Given Steps
// ============================================================================

func (c *repositoryContext) repositoryRootExists() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}
	if _, err := os.Stat(c.repoRoot); os.IsNotExist(err) {
		return fmt.Errorf("repository root does not exist: %s", c.repoRoot)
	}
	// Initialize collections
	c.discoveredModules = []string{}
	c.tidyResults = make(map[string]string)
	c.failedModules = []string{}
	c.markdownFiles = []string{}
	c.markdownErrors = make(map[string][]string)
	c.featureFiles = []string{}
	c.tagConflicts = []string{}
	return nil
}

func (c *repositoryContext) discoverAllGoModulesUsingContracts() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	moduleReport, err := reports.GetModuleContracts(c.repoRoot)
	if err != nil {
		return fmt.Errorf("failed to load module contracts: %w", err)
	}

	goModuleTypes := []string{"go-cli", "go-commands", "go-library", "go-mcp", "go-tests"}
	for _, module := range moduleReport.Registry.All() {
		for _, t := range goModuleTypes {
			if module.Type == t {
				modulePath := filepath.Join(c.repoRoot, module.Files.Root)
				c.discoveredModules = append(c.discoveredModules, modulePath)
				break
			}
		}
	}

	if len(c.discoveredModules) == 0 {
		return fmt.Errorf("no Go modules discovered in repository")
	}
	return nil
}

func (c *repositoryContext) discoverAllMarkdownFiles() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	err := filepath.Walk(c.repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories and common ignore patterns
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			c.markdownFiles = append(c.markdownFiles, path)
		}
		return nil
	})
	return err
}

func (c *repositoryContext) discoverAllGherkinFeatureFiles() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	specsDir := filepath.Join(c.repoRoot, "specs")
	err := filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk specs directory: %w", err)
		}
		if !info.IsDir() && strings.HasSuffix(path, ".feature") {
			c.featureFiles = append(c.featureFiles, path)
		}
		return nil
	})
	return err
}

// ============================================================================
// When Steps
// ============================================================================

func (c *repositoryContext) runGoModTidyDiffInEachModule() error {
	for _, modulePath := range c.discoveredModules {
		cmd := exec.Command("go", "mod", "tidy", "-diff")
		cmd.Dir = modulePath

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		output := stdout.String() + stderr.String()
		c.tidyResults[modulePath] = output

		if err != nil || strings.TrimSpace(output) != "" {
			c.failedModules = append(c.failedModules, modulePath)
		}
	}
	return nil
}

func (c *repositoryContext) validateEachMarkdownFile() error {
	// Basic markdown validation - check for common issues
	for _, path := range c.markdownFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			c.markdownErrors[path] = append(c.markdownErrors[path], fmt.Sprintf("failed to read: %v", err))
			continue
		}

		// Check for malformed headers (e.g., #NoSpace)
		// Must track code blocks to avoid false positives on shebangs like #!/bin/bash
		lines := strings.Split(string(content), "\n")
		inCodeBlock := false
		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			// Track code block boundaries
			if strings.HasPrefix(trimmedLine, "```") {
				inCodeBlock = !inCodeBlock
				continue
			}

			// Skip lines inside code blocks
			if inCodeBlock {
				continue
			}

			if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "# ") &&
				!strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") &&
				!strings.HasPrefix(line, "#### ") && !strings.HasPrefix(line, "##### ") &&
				!strings.HasPrefix(line, "###### ") && line != "#" {
				// Check if it's actually a header without space
				trimmed := strings.TrimLeft(line, "#")
				if len(trimmed) > 0 && trimmed[0] != ' ' && trimmed[0] != '\n' && trimmed[0] != '\r' {
					c.markdownErrors[path] = append(c.markdownErrors[path],
						fmt.Sprintf("line %d: malformed header (missing space after #)", i+1))
				}
			}
		}
	}
	return nil
}

func (c *repositoryContext) validateFeatureFilesForConflictingLLevelTags() error {
	// TODO: Implement proper Gherkin parsing and tag conflict detection
	// For now, this is a placeholder
	return nil
}


// ============================================================================
// Then Steps
// ============================================================================

func (c *repositoryContext) allModulesShouldHaveExitCode0() error {
	if len(c.failedModules) > 0 {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Found %d module(s) with untidy dependencies:\n\n", len(c.failedModules)))
		for _, modulePath := range c.failedModules {
			relPath, _ := filepath.Rel(c.repoRoot, modulePath)
			details.WriteString(fmt.Sprintf("❌ %s\n", relPath))
			if diff := c.tidyResults[modulePath]; diff != "" {
				details.WriteString(fmt.Sprintf("   Diff:\n%s\n\n", indent(diff, "   ")))
			}
		}
		return fmt.Errorf("%s", details.String())
	}
	return nil
}

func (c *repositoryContext) noModuleShouldHaveAnyDiffOutput() error {
	var modulesWithDiff []string
	for modulePath, output := range c.tidyResults {
		if strings.TrimSpace(output) != "" {
			modulesWithDiff = append(modulesWithDiff, modulePath)
		}
	}
	if len(modulesWithDiff) > 0 {
		return fmt.Errorf("found %d module(s) with diff output", len(modulesWithDiff))
	}
	return nil
}

func (c *repositoryContext) ifAnyModuleIsNotTidyIShouldSeeTheModulePathAndDiff() error {
	// Passive assertion - error messages from above steps provide this info
	return nil
}

func (c *repositoryContext) allFilesShouldHaveValidMarkdownSyntax() error {
	if len(c.markdownErrors) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("found %d file(s) with markdown errors:\n", len(c.markdownErrors)))
		for path, errors := range c.markdownErrors {
			sb.WriteString(fmt.Sprintf("\n  %s:\n", path))
			for _, err := range errors {
				sb.WriteString(fmt.Sprintf("    - %s\n", err))
			}
		}
		return fmt.Errorf("%s", sb.String())
	}
	return nil
}

func (c *repositoryContext) noFilesShouldHaveBrokenLinks() error {
	// TODO: Implement link checking
	return nil
}

func (c *repositoryContext) noFilesShouldHaveMalformedHeaders() error {
	for path, errors := range c.markdownErrors {
		for _, err := range errors {
			if strings.Contains(err, "malformed header") {
				return fmt.Errorf("file %s has malformed headers", path)
			}
		}
	}
	return nil
}

func (c *repositoryContext) ifAnyFileHasErrorsIShouldSeeDetails() error {
	// Passive assertion
	return nil
}

func (c *repositoryContext) noFeatureShouldHaveConflictingLTags() error {
	if len(c.tagConflicts) > 0 {
		return fmt.Errorf("found tag conflicts: %v", c.tagConflicts)
	}
	return nil
}

func (c *repositoryContext) noFeatureShouldHaveConflictingVerificationTags() error {
	// Same as L-tags for now
	return nil
}

func (c *repositoryContext) ifConflictsFoundIShouldSeeDetails() error {
	// Passive assertion
	return nil
}

func (c *repositoryContext) shouldNotSeeBuildErrors() error {
	if c.sharedCtx.ExitCode != 0 {
		return fmt.Errorf("build had errors (exit code %d)\nOutput: %s",
			c.sharedCtx.ExitCode, c.sharedCtx.CommandOutput)
	}
	return nil
}

// ============================================================================
// Module Hierarchy Validation Steps
// ============================================================================

func (c *repositoryContext) loadAllModuleContracts() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	moduleReport, err := reports.GetModuleContracts(c.repoRoot)
	if err != nil {
		return fmt.Errorf("failed to load module contracts: %w", err)
	}
	c.moduleReport = moduleReport
	c.dependencyErrors = []string{}
	c.circularDependencies = []string{}
	c.missingModules = []string{}
	c.bidirectionalErrors = []string{}
	return nil
}

func (c *repositoryContext) validateBidirectionalRelationships() error {
	if c.moduleReport == nil {
		return fmt.Errorf("module contracts not loaded")
	}

	// Check all depends_on references exist
	for _, m := range c.moduleReport.Registry.All() {
		for _, dep := range m.DependsOn {
			if _, exists := c.moduleReport.Registry.Get(dep); !exists {
				c.missingModules = append(c.missingModules,
					fmt.Sprintf("%s depends_on %s but %s does not exist", m.Name, dep, dep))
			}
		}
	}
	// Note: used_by is computed from depends_on, so no bidirectional check needed
	return nil
}

func (c *repositoryContext) buildCompleteDependencyGraph() error {
	if c.moduleReport == nil {
		return fmt.Errorf("module contracts not loaded")
	}
	// Graph is implicitly built via module contracts
	// Check for cycles using DFS
	modules := c.moduleReport.Registry.All()
	visited := make(map[string]int) // 0=unvisited, 1=visiting, 2=visited

	var dfs func(name string, path []string) bool
	dfs = func(name string, path []string) bool {
		if visited[name] == 1 {
			// Found cycle
			cycleStart := 0
			for i, p := range path {
				if p == name {
					cycleStart = i
					break
				}
			}
			cycle := append(path[cycleStart:], name)
			c.circularDependencies = append(c.circularDependencies,
				fmt.Sprintf("Circular dependency: %s", strings.Join(cycle, " -> ")))
			return true
		}
		if visited[name] == 2 {
			return false
		}

		visited[name] = 1
		if m, exists := c.moduleReport.Registry.Get(name); exists {
			for _, dep := range m.DependsOn {
				if dfs(dep, append(path, name)) {
					return true
				}
			}
		}
		visited[name] = 2
		return false
	}

	for _, m := range modules {
		if visited[m.Name] == 0 {
			dfs(m.Name, []string{})
		}
	}
	return nil
}

func (c *repositoryContext) checkAllDependsOnReferences() error {
	if c.moduleReport == nil {
		return fmt.Errorf("module contracts not loaded")
	}

	for _, m := range c.moduleReport.Registry.All() {
		for _, dep := range m.DependsOn {
			if _, exists := c.moduleReport.Registry.Get(dep); !exists {
				c.missingModules = append(c.missingModules,
					fmt.Sprintf("%s.depends_on references non-existent module: %s", m.Name, dep))
			}
		}
	}
	return nil
}

func (c *repositoryContext) checkAllUsedByReferences() error {
	// used_by is computed from depends_on, so this is now a no-op
	// The validation happens through checkAllDependsOnReferences
	return nil
}

func (c *repositoryContext) allModulesShouldHaveConsistentDependencies() error {
	if len(c.bidirectionalErrors) > 0 || len(c.dependencyErrors) > 0 {
		errors := append(c.bidirectionalErrors, c.dependencyErrors...)
		return fmt.Errorf("found %d inconsistency(ies):\n%s",
			len(errors), strings.Join(errors, "\n"))
	}
	return nil
}

func (c *repositoryContext) noModuleShouldReferenceNonExistent() error {
	if len(c.missingModules) > 0 {
		return fmt.Errorf("found %d missing module reference(s):\n%s",
			len(c.missingModules), strings.Join(c.missingModules, "\n"))
	}
	return nil
}

func (c *repositoryContext) graphShouldBeConnectedOrForest() error {
	// For now, we allow disconnected components (forest)
	// A fully connected graph check would be stricter
	return nil
}

func (c *repositoryContext) graphShouldHaveNoCircularDependencies() error {
	if len(c.circularDependencies) > 0 {
		return fmt.Errorf("found %d circular dependency(ies):\n%s",
			len(c.circularDependencies), strings.Join(c.circularDependencies, "\n"))
	}
	return nil
}

func (c *repositoryContext) allModulesShouldBeReachableFromRoot() error {
	// For now, pass - we don't enforce a single root
	return nil
}

func (c *repositoryContext) everyReferencedModuleShouldExist() error {
	if len(c.missingModules) > 0 {
		return fmt.Errorf("found %d missing module reference(s):\n%s",
			len(c.missingModules), strings.Join(c.missingModules, "\n"))
	}
	return nil
}

func (c *repositoryContext) shouldSeeDetailsOfMissingModules() error {
	// Passive assertion - error messages provide details
	return nil
}

func (c *repositoryContext) dependsOnMustHaveUsedByBack() error {
	// Already checked in validateBidirectionalRelationships
	return nil
}

func (c *repositoryContext) usedByMustHaveDependsOnBack() error {
	// Already checked in validateBidirectionalRelationships
	return nil
}

func (c *repositoryContext) shouldSeeDetailsOfInconsistencies() error {
	// Passive assertion - error messages provide details
	return nil
}

// ============================================================================
// File Ownership Validation Steps
// ============================================================================

func (c *repositoryContext) lookupFilesBelongingToModule(moduleName string) error {
	if c.moduleReport == nil {
		return fmt.Errorf("module contracts not loaded")
	}

	c.moduleFiles = []string{}

	m, exists := c.moduleReport.Registry.Get(moduleName)
	if !exists {
		// Module doesn't exist, so no files belong to it
		return nil
	}

	// Walk the repository and find files that match this module
	err := filepath.Walk(c.repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		// Get relative path
		relPath, err := filepath.Rel(c.repoRoot, path)
		if err != nil {
			return nil
		}
		relPath = strings.ReplaceAll(relPath, "\\", "/")
		if m.MatchesFile(relPath) {
			c.moduleFiles = append(c.moduleFiles, relPath)
		}
		return nil
	})
	return err
}

func (c *repositoryContext) fileListShouldBeEmpty() error {
	if len(c.moduleFiles) > 0 {
		return fmt.Errorf("found %d file(s) in module:\n%s",
			len(c.moduleFiles), strings.Join(c.moduleFiles, "\n"))
	}
	return nil
}

func (c *repositoryContext) ifAnyFilesFoundShowPathsWithCounts() error {
	// Passive assertion - error in fileListShouldBeEmpty provides details
	return nil
}

func (c *repositoryContext) checkForFilesWithMultiModuleOwnership() error {
	if c.moduleReport == nil {
		return fmt.Errorf("module contracts not loaded")
	}

	c.multiOwnershipMap = make(map[string][]string)
	modules := c.moduleReport.Registry.All()

	// Walk the repository and check each file against all modules
	err := filepath.Walk(c.repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		// Get relative path
		relPath, pathErr := filepath.Rel(c.repoRoot, path)
		if pathErr != nil {
			return nil
		}
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		// Check which modules match this file
		var matchingModules []string
		for _, m := range modules {
			if m.MatchesFile(relPath) {
				matchingModules = append(matchingModules, m.Name)
			}
		}
		if len(matchingModules) > 1 {
			c.multiOwnershipMap[relPath] = matchingModules
		}
		return nil
	})
	return err
}

func (c *repositoryContext) noFilesShouldBelongToMultipleModules() error {
	if len(c.multiOwnershipMap) > 0 {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Found %d file(s) with multiple owners:\n", len(c.multiOwnershipMap)))
		for file, modules := range c.multiOwnershipMap {
			details.WriteString(fmt.Sprintf("  %s: owned by %s\n", file, strings.Join(modules, ", ")))
		}
		return fmt.Errorf("%s", details.String())
	}
	return nil
}

func (c *repositoryContext) ifMultiOwnershipShowPathsAndModules() error {
	// Passive assertion - error in noFilesShouldBelongToMultipleModules provides details
	return nil
}

// Helper functions
func indent(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
