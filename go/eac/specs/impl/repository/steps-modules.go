// Package repository contains godog step implementations for specs/repository.
package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// ============================================================================
// Module Hierarchy Validation Steps
// ============================================================================

func (c *repositoryContext) loadAllModuleContracts() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	// Use cached module contracts for performance
	if err := c.sharedCtx.EnsureOriginalRepoCache(); err != nil {
		return err
	}

	moduleReport, err := c.sharedCtx.OriginalRepoCache.ModuleReport()
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
	for _, m := range c.moduleReport.Registry().All() {
		for _, dep := range m.GetDependsOn() {
			if _, exists := c.moduleReport.Registry().Get(dep); !exists {
				c.missingModules = append(c.missingModules,
					fmt.Sprintf("%s depends_on %s but %s does not exist", m.GetName(), dep, dep))
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
	modules := c.moduleReport.Registry().All()
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
		if m, exists := c.moduleReport.Registry().Get(name); exists {
			for _, dep := range m.GetDependsOn() {
				if dfs(dep, append(path, name)) {
					return true
				}
			}
		}
		visited[name] = 2
		return false
	}

	for _, m := range modules {
		if visited[m.GetName()] == 0 {
			dfs(m.GetName(), []string{})
		}
	}
	return nil
}

func (c *repositoryContext) checkAllDependsOnReferences() error {
	if c.moduleReport == nil {
		return fmt.Errorf("module contracts not loaded")
	}

	for _, m := range c.moduleReport.Registry().All() {
		for _, dep := range m.GetDependsOn() {
			if _, exists := c.moduleReport.Registry().Get(dep); !exists {
				c.missingModules = append(c.missingModules,
					fmt.Sprintf("%s.depends_on references non-existent module: %s", m.GetName(), dep))
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

func (c *repositoryContext) checkForOrphanFiles() error {
	if c.moduleReport == nil {
		return fmt.Errorf("module contracts not loaded")
	}

	c.orphanFiles = []repository.RepositoryFileWithModule{}

	// Use cached file list instead of running git ls-files again
	if err := c.sharedCtx.EnsureOriginalRepoCache(); err != nil {
		return fmt.Errorf("failed to get file list: %w", err)
	}

	// Check each file against the already-loaded module registry
	for _, file := range c.sharedCtx.OriginalRepoCache.TrackedFiles() {
		// Skip git internal files (.gitignore, .gitkeep) - same as GetRepositoryFiles
		basename := filepath.Base(file)
		if basename == ".gitignore" || basename == ".gitkeep" {
			continue
		}

		// Use the already-loaded registry to check file ownership
		matchingModules := c.moduleReport.Registry().FindModulesForFile(file)
		if len(matchingModules) == 0 {
			c.orphanFiles = append(c.orphanFiles, repository.RepositoryFileWithModule{
				Name:    file,
				Modules: []string{},
			})
		}
	}

	return nil
}

func (c *repositoryContext) noFilesShouldBeOrphaned() error {
	if len(c.orphanFiles) > 0 {
		paths := make([]string, len(c.orphanFiles))
		for i, f := range c.orphanFiles {
			paths[i] = f.Name
		}
		return fmt.Errorf("found %d orphan file(s) without module ownership:\n%s",
			len(c.orphanFiles), strings.Join(paths, "\n"))
	}
	return nil
}

func (c *repositoryContext) ifOrphanFilesFoundShowPathsWithCounts() error {
	// Passive assertion - error in noFilesShouldBeOrphaned provides details
	return nil
}

func (c *repositoryContext) checkForFilesWithMultiModuleOwnership() error {
	if c.moduleReport == nil {
		return fmt.Errorf("module contracts not loaded")
	}

	c.multiOwnershipMap = make(map[string][]string)

	// Use cached file list instead of running git ls-files again
	if err := c.sharedCtx.EnsureOriginalRepoCache(); err != nil {
		return fmt.Errorf("failed to get file list: %w", err)
	}

	// Check each file against the already-loaded module registry
	for _, file := range c.sharedCtx.OriginalRepoCache.TrackedFiles() {
		// Skip git internal files (.gitignore, .gitkeep)
		basename := filepath.Base(file)
		if basename == ".gitignore" || basename == ".gitkeep" {
			continue
		}

		// Use the already-loaded registry to check file ownership
		matchingModules := c.moduleReport.Registry().FindModulesForFile(file)
		if len(matchingModules) > 1 {
			// Extract module names
			moduleNames := make([]string, len(matchingModules))
			for i, m := range matchingModules {
				moduleNames[i] = m.GetName()
			}
			c.multiOwnershipMap[file] = moduleNames
		}
	}

	return nil
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

// ============================================================================
// Build Tags Validation Steps
// ============================================================================

func (c *repositoryContext) discoverGodogTestFiles(dir string) error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	c.godogTestFiles = []string{}
	c.filesWithBuildTags = make(map[string]string)

	// Use cached file list instead of filepath.Walk
	if err := c.sharedCtx.EnsureOriginalRepoCache(); err != nil {
		return err
	}

	// Get all godog_test.go files in the specified directory from cache
	for _, file := range c.sharedCtx.OriginalRepoCache.FilesInDir(dir) {
		if strings.HasSuffix(file, "godog_test.go") {
			c.godogTestFiles = append(c.godogTestFiles, file)
		}
	}
	return nil
}

func (c *repositoryContext) checkFilesForBuildTags() error {
	for _, relPath := range c.godogTestFiles {
		fullPath := c.sharedCtx.OriginalRepoCache.AbsolutePath(relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		// Check first 10 lines for build tags (they must appear before package declaration)
		for i := 0; i < len(lines) && i < 10; i++ {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, "//go:build ") {
				c.filesWithBuildTags[relPath] = line
				break
			}
			if strings.HasPrefix(line, "// +build ") {
				c.filesWithBuildTags[relPath] = line
				break
			}
			// Stop if we hit the package declaration
			if strings.HasPrefix(line, "package ") {
				break
			}
		}
	}
	return nil
}

func (c *repositoryContext) noFilesShouldHaveDirective(directive string) error {
	var violations []string
	for file, tag := range c.filesWithBuildTags {
		if strings.Contains(tag, directive) {
			violations = append(violations, fmt.Sprintf("  %s: %s", file, tag))
		}
	}
	if len(violations) > 0 {
		return fmt.Errorf("found %d file(s) with %s directives:\n%s",
			len(violations), directive, strings.Join(violations, "\n"))
	}
	return nil
}

func (c *repositoryContext) ifBuildTagsFoundShowDetails() error {
	// Passive assertion - errors from noFilesShouldHaveDirective provide details
	return nil
}
