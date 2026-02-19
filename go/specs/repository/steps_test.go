// Package repository contains godog step implementations for specs/repository.
package repository

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

// RegisterSteps registers all repository-specific step definitions.
func RegisterSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Create repository-specific context that also updates the shared context
	repoCtx := &repositoryContext{sharedCtx: ctx}

	// Register module isolation steps
	registerModuleIsolationSteps(sc, ctx)

	// Register test sanity steps
	registerTestSanitySteps(sc, ctx)

	// Register command docs coverage steps
	registerCommandDocsSteps(sc, ctx)

	// Initialize version consistency contexts
	// Note: contexts will get repoRoot lazily when needed
	goVersionCtx := &goVersionContext{
		actionVersions: make(map[string]string),
	}
	nodeVersionCtx := &nodeVersionContext{
		actionVersions: make(map[string]string),
	}

	// Register Go version consistency steps (also initializes shared Node.js steps)
	registerGoVersionSteps(sc, goVersionCtx, nodeVersionCtx, repoCtx)

	// Register Node.js version consistency steps
	registerNodeVersionSteps(sc, nodeVersionCtx, repoCtx)

	// Given steps
	// Note: "the repository root exists" is registered in common steps (internal/steps.go)
	// Repository context is initialized lazily via ensureRepoRoot() when needed
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

	// No-unordered-files steps (orphan file detection)
	sc.Step(`^the module contracts are loaded$`, func() error {
		return repoCtx.loadAllModuleContracts()
	})
	sc.Step(`^I check for orphan files$`, func() error {
		return repoCtx.checkForOrphanFiles()
	})
	sc.Step(`^no files should be orphaned$`, func() error {
		return repoCtx.noFilesShouldBeOrphaned()
	})
	sc.Step(`^if any orphan files are found, I should see their paths with counts$`, func() error {
		return repoCtx.ifOrphanFilesFoundShowPathsWithCounts()
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

	// Build tags in step files validation
	sc.Step(`^I discover all godog_test\.go files in "([^"]*)"$`, func(dir string) error {
		return repoCtx.discoverGodogTestFiles(dir)
	})
	sc.Step(`^I check each file for Go build tags$`, func() error {
		return repoCtx.checkFilesForBuildTags()
	})
	sc.Step(`^no files should have "([^"]*)" directives$`, func(directive string) error {
		return repoCtx.noFilesShouldHaveDirective(directive)
	})
	sc.Step(`^if any files have build tags, I should see the file path and the tag$`, func() error {
		return repoCtx.ifBuildTagsFoundShowDetails()
	})

	// Script location validation steps
	sc.Step(`^the following script extensions are tracked:$`, func(table *godog.Table) error {
		return repoCtx.theFollowingScriptExtensionsAreTracked(table)
	})
	sc.Step(`^I scan the repository for script files$`, func() error {
		return repoCtx.scanRepositoryForScriptFiles()
	})
	sc.Step(`^all scripts should be in one of these locations:$`, func(table *godog.Table) error {
		return repoCtx.allScriptsShouldBeInApprovedLocations(table)
	})
	sc.Step(`^no scripts should exist in:$`, func(table *godog.Table) error {
		return repoCtx.noScriptsShouldExistIn(table)
	})
	sc.Step(`^these locations are excluded from validation:$`, func(table *godog.Table) error {
		return repoCtx.theseLocationsAreExcludedFromValidation(table)
	})
	sc.Step(`^I scan the scripts directory structure$`, func() error {
		return repoCtx.scanScriptsDirectoryStructure()
	})
	sc.Step(`^each script type directory should contain only package subdirectories$`, func() error {
		return repoCtx.eachScriptTypeDirectoryShouldContainOnlyPackageSubdirectories()
	})
	sc.Step(`^package names should be lowercase with hyphens$`, func() error {
		return repoCtx.packageNamesShouldBeLowercaseWithHyphens()
	})
	sc.Step(`^each package should contain at least one script file$`, func() error {
		return repoCtx.eachPackageShouldContainAtLeastOneScriptFile()
	})
	sc.Step(`^I check the scripts directory structure$`, func() error {
		return repoCtx.checkScriptsDirectoryStructure()
	})
	sc.Step(`^scripts/pwsh/ should contain only directories$`, func() error {
		return repoCtx.scriptsTypeDirShouldContainOnlyDirectories("scripts/pwsh")
	})
	sc.Step(`^scripts/sh/ should contain only directories$`, func() error {
		return repoCtx.scriptsTypeDirShouldContainOnlyDirectories("scripts/sh")
	})
	sc.Step(`^scripts/cmd/ should contain only directories if it exists$`, func() error {
		return repoCtx.scriptsCmdDirShouldContainOnlyDirectoriesIfExists()
	})

	// Release folder validation steps
	sc.Step(`^I scan the release directory for subdirectories$`, func() error {
		return repoCtx.scanReleaseDirectoryForSubdirectories()
	})
	sc.Step(`^every release subdirectory should correspond to a registered module$`, func() error {
		return repoCtx.everyReleaseSubdirShouldCorrespondToModule()
	})
	sc.Step(`^if any orphan directories are found, I should see their names$`, func() error {
		return repoCtx.ifOrphanReleaseDirsFoundShowNames()
	})
	sc.Step(`^I scan the release directory for changelog files$`, func() error {
		return repoCtx.scanReleaseDirectoryForChangelogFiles()
	})
	sc.Step(`^every changelog should have valid format$`, func() error {
		return repoCtx.everyChangelogShouldHaveValidFormat()
	})
	sc.Step(`^no unexpected files should exist in release subdirectories$`, func() error {
		return repoCtx.noUnexpectedFilesInReleaseSubdirs()
	})
}

// repositoryContext holds state for repository validation scenarios.
type repositoryContext struct {
	sharedCtx *eacgodog.TestContext
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
	moduleReport         core.ModuleReportPort
	dependencyErrors     []string
	circularDependencies []string
	missingModules       []string
	bidirectionalErrors  []string

	// File ownership validation
	orphanFiles       []repository.RepositoryFileWithModule // files with no owner
	multiOwnershipMap map[string][]string                   // file -> list of owning modules

	// Build tags validation
	godogTestFiles     []string
	filesWithBuildTags map[string]string // file -> build tag found

	// Script location validation
	scriptExtensions   []string
	discoveredScripts  []string
	disallowedScripts  []string
	looseScriptsInType []string

	// Release folder validation
	releaseSubdirs         []string
	orphanReleaseDirs      []string
	releaseChangelogErrs   map[string][]string
	unexpectedReleaseFiles []string
}


func (c *repositoryContext) ensureRepoRoot() error {
	if c.repoRoot == "" {
		repoRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return fmt.Errorf("failed to get repository root: %w", err)
		}
		c.repoRoot = repoRoot
		// Initialize collections
		c.discoveredModules = []string{}
		c.tidyResults = make(map[string]string)
		c.failedModules = []string{}
		c.markdownFiles = []string{}
		c.markdownErrors = make(map[string][]string)
		c.featureFiles = []string{}
		c.tagConflicts = []string{}
	}
	return nil
}

// ============================================================================
// Given Steps
// ============================================================================

func (c *repositoryContext) discoverAllGoModulesUsingContracts() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	// Load module contracts directly (bypasses port abstraction to access
	// component type information needed for Go module discovery).
	moduleReport, err := reports.GetModuleContracts(c.repoRoot)
	if err != nil {
		return fmt.Errorf("failed to load module contracts: %w", err)
	}

	for _, module := range moduleReport.Registry.All() {
		// Find Go components by type (not name), since list-format component
		// names are derived from root paths (e.g., "eac"), not the type "go".
		goComps := module.Components.GetComponentsByType("go")
		for _, comp := range goComps {
			if comp != nil && comp.Root != "" {
				modulePath := filepath.Join(c.repoRoot, comp.Root)
				c.discoveredModules = append(c.discoveredModules, modulePath)
			}
		}
	}

	if len(c.discoveredModules) == 0 {
		return fmt.Errorf("no Go modules discovered in repository")
	}
	return nil
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
		// Only use stdout for the diff — stderr contains "go: downloading" noise
		// that doesn't indicate an untidy module.
		diff := stdout.String()
		c.tidyResults[modulePath] = diff

		if err != nil || strings.TrimSpace(diff) != "" {
			c.failedModules = append(c.failedModules, modulePath)
		}
	}
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
			relPath, err := filepath.Rel(c.repoRoot, modulePath)
			if err != nil {
				relPath = modulePath
			}
			details.WriteString(fmt.Sprintf("X %s\n", relPath))
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

func (c *repositoryContext) shouldNotSeeBuildErrors() error {
	if c.sharedCtx.ExitCode != 0 {
		return fmt.Errorf("build had errors (exit code %d)\nOutput: %s",
			c.sharedCtx.ExitCode, c.sharedCtx.CommandOutput)
	}
	return nil
}
