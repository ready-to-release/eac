// Godog BDD step definitions - src-cli module isolation
//
// Feature: src-cli_module-isolation (specs/src-cli/module-isolation/specification.feature)
//
// This file implements steps for validating that src-cli production code
// remains isolated from other local modules, while allowing test code to
// import them.
package tests

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// moduleIsolationContext holds state for module isolation validation
type moduleIsolationContext struct {
	modulePath           string
	productionFiles      []string
	testFiles            []string
	violatingFiles       map[string][]string // file -> imported local modules
	allowedTestImports   map[string]bool     // test files that import local modules (allowed)
}

var moduleIsolationCtx *moduleIsolationContext

// resetModuleIsolationContext resets the context between scenarios
func resetModuleIsolationContext() {
	moduleIsolationCtx = nil
}

// ============================================================================
// Given Steps
// ============================================================================

func iAmCheckingModule(moduleName string) error {
	// Get repository root
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	modulePath := filepath.Join(repoRoot, "src", "cli")
	if moduleName != "src/cli" {
		return fmt.Errorf("this test only supports src/cli module, got: %s", moduleName)
	}

	moduleIsolationCtx = &moduleIsolationContext{
		modulePath:         modulePath,
		productionFiles:    []string{},
		testFiles:          []string{},
		violatingFiles:     make(map[string][]string),
		allowedTestImports: make(map[string]bool),
	}

	return nil
}

func theGoModFileLists(dependency string) error {
	if moduleIsolationCtx == nil {
		return fmt.Errorf("module isolation context not initialized")
	}

	// Read go.mod file
	goModPath := filepath.Join(moduleIsolationCtx.modulePath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Check if dependency is listed
	if !strings.Contains(string(content), dependency) {
		return fmt.Errorf("dependency %s not found in go.mod", dependency)
	}

	return nil
}

// ============================================================================
// When Steps
// ============================================================================

func iScanAllProductionGoFilesIn(moduleName string) error {
	if moduleIsolationCtx == nil {
		return fmt.Errorf("module isolation context not initialized")
	}

	// Walk through the module directory
	err := filepath.Walk(moduleIsolationCtx.modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip vendor, .git, and out directories
			if info.Name() == "vendor" || info.Name() == ".git" || info.Name() == "out" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// Determine if this is a production file or test file
		relPath, _ := filepath.Rel(moduleIsolationCtx.modulePath, path)

		// Test files are:
		// 1. Files in tests/ directory
		// 2. Files ending with _test.go
		isTestFile := strings.Contains(relPath, "tests"+string(filepath.Separator)) || strings.HasSuffix(info.Name(), "_test.go")

		if isTestFile {
			moduleIsolationCtx.testFiles = append(moduleIsolationCtx.testFiles, path)
		} else {
			moduleIsolationCtx.productionFiles = append(moduleIsolationCtx.productionFiles, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	return nil
}

func iScanTestGoFilesIn(testDir string) error {
	if moduleIsolationCtx == nil {
		return fmt.Errorf("module isolation context not initialized")
	}

	testPath := filepath.Join(moduleIsolationCtx.modulePath, "tests")

	// Walk through the tests directory
	err := filepath.Walk(testPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		moduleIsolationCtx.testFiles = append(moduleIsolationCtx.testFiles, path)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan test files: %w", err)
	}

	return nil
}

func iVerifyTheDependencyIsOnlyUsedInTestFiles() error {
	if moduleIsolationCtx == nil {
		return fmt.Errorf("module isolation context not initialized")
	}

	// Check all production files for local module imports
	for _, file := range moduleIsolationCtx.productionFiles {
		imports, err := getLocalModuleImports(file)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", file, err)
		}

		if len(imports) > 0 {
			relPath, _ := filepath.Rel(moduleIsolationCtx.modulePath, file)
			moduleIsolationCtx.violatingFiles[relPath] = imports
		}
	}

	// Check test files (for informational purposes)
	for _, file := range moduleIsolationCtx.testFiles {
		imports, err := getLocalModuleImports(file)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", file, err)
		}

		if len(imports) > 0 {
			relPath, _ := filepath.Rel(moduleIsolationCtx.modulePath, file)
			moduleIsolationCtx.allowedTestImports[relPath] = true
		}
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

func noProductionFilesShouldImportLocalModulesFrom(localModulePrefix string) error {
	if moduleIsolationCtx == nil {
		return fmt.Errorf("module isolation context not initialized")
	}

	// If we haven't scanned imports yet, do it now
	if len(moduleIsolationCtx.violatingFiles) == 0 && len(moduleIsolationCtx.productionFiles) > 0 {
		for _, file := range moduleIsolationCtx.productionFiles {
			imports, err := getLocalModuleImports(file)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", file, err)
			}

			if len(imports) > 0 {
				relPath, _ := filepath.Rel(moduleIsolationCtx.modulePath, file)
				moduleIsolationCtx.violatingFiles[relPath] = imports
			}
		}
	}

	// Check if any production files import this local module
	var violatingFiles []string
	for file, imports := range moduleIsolationCtx.violatingFiles {
		for _, imp := range imports {
			if strings.HasPrefix(imp, localModulePrefix) {
				violatingFiles = append(violatingFiles, fmt.Sprintf("  - %s imports %s", file, imp))
			}
		}
	}

	if len(violatingFiles) > 0 {
		return fmt.Errorf("production files import local module %s:\n%s", localModulePrefix, strings.Join(violatingFiles, "\n"))
	}

	return nil
}

func noProductionFilesShouldImportAnyOtherLocalModulesOutside(allowedModule string) error {
	if moduleIsolationCtx == nil {
		return fmt.Errorf("module isolation context not initialized")
	}

	if len(moduleIsolationCtx.violatingFiles) > 0 {
		var details []string
		for file, imports := range moduleIsolationCtx.violatingFiles {
			details = append(details, fmt.Sprintf("  - %s imports: %s", file, strings.Join(imports, ", ")))
		}
		return fmt.Errorf("production files import local modules:\n%s", strings.Join(details, "\n"))
	}

	return nil
}

func testFilesMayImportLocalModulesLike(localModulePrefix string) error {
	if moduleIsolationCtx == nil {
		return fmt.Errorf("module isolation context not initialized")
	}

	// Scan test files for imports if not done yet
	if len(moduleIsolationCtx.allowedTestImports) == 0 && len(moduleIsolationCtx.testFiles) > 0 {
		for _, file := range moduleIsolationCtx.testFiles {
			imports, err := getLocalModuleImports(file)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", file, err)
			}

			if len(imports) > 0 {
				relPath, _ := filepath.Rel(moduleIsolationCtx.modulePath, file)
				moduleIsolationCtx.allowedTestImports[relPath] = true
			}
		}
	}

	// This is informational - test files are allowed to import local modules
	// Just verify we found some test files doing this
	if len(moduleIsolationCtx.allowedTestImports) == 0 {
		return fmt.Errorf("expected to find test files importing local modules, but found none")
	}

	return nil
}

func thisIsAllowedForTestInfrastructurePurposes() error {
	// This is just a documentation step - always passes
	return nil
}

func theDependencyShouldOnlyBeImportedByFilesMatching(table *godog.Table) error {
	if moduleIsolationCtx == nil {
		return fmt.Errorf("module isolation context not initialized")
	}

	// Verify no production files violate the rule
	if len(moduleIsolationCtx.violatingFiles) > 0 {
		var details []string
		for file, imports := range moduleIsolationCtx.violatingFiles {
			details = append(details, fmt.Sprintf("  - %s imports: %s", file, strings.Join(imports, ", ")))
		}
		return fmt.Errorf("production files violate isolation rule:\n%s", strings.Join(details, "\n"))
	}

	return nil
}

func noDependencyShouldImportThisDependency() error {
	// Placeholder - this step would check production files don't import the dependency
	return noProductionFilesShouldImportAnyOtherLocalModulesOutside("src/cli")
}

// ============================================================================
// Helper Functions
// ============================================================================

// getLocalModuleImports parses a Go file and returns local module imports
func getLocalModuleImports(filePath string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	var localImports []string
	for _, imp := range f.Imports {
		// Remove quotes from import path
		importPath := strings.Trim(imp.Path.Value, `"`)

		// Check if it's a local module import (starts with github.com/ready-to-release/eac/src/)
		if strings.HasPrefix(importPath, "github.com/ready-to-release/eac/src/") {
			// Exclude imports from src/cli itself
			if !strings.HasPrefix(importPath, "github.com/ready-to-release/eac/src/cli") {
				localImports = append(localImports, importPath)
			}
		}
	}

	return localImports, nil
}

// ============================================================================
// Scenario Initialization (called from main godog_test.go)
// ============================================================================

func initializeModuleIsolationScenario(sc *godog.ScenarioContext) {
	// Reset context before each scenario
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		resetModuleIsolationContext()
		return ctx, nil
	})

	// Given steps
	sc.Step(`^I am checking module "([^"]*)"$`, iAmCheckingModule)
	sc.Step(`^the go\.mod file lists "([^"]*)" as a dependency$`, theGoModFileLists)

	// When steps
	sc.Step(`^I scan all production \.go files in "([^"]*)"$`, iScanAllProductionGoFilesIn)
	sc.Step(`^I scan test \.go files in "([^"]*)"$`, iScanTestGoFilesIn)
	sc.Step(`^I verify the dependency is only used in test files$`, iVerifyTheDependencyIsOnlyUsedInTestFiles)

	// Then steps
	sc.Step(`^no production files should import local modules from "([^"]*)"$`, noProductionFilesShouldImportLocalModulesFrom)
	sc.Step(`^no production files should import any other local modules outside "([^"]*)"$`, noProductionFilesShouldImportAnyOtherLocalModulesOutside)
	sc.Step(`^test files MAY import local modules like "([^"]*)"$`, testFilesMayImportLocalModulesLike)
	sc.Step(`^this is allowed for test infrastructure purposes$`, thisIsAllowedForTestInfrastructurePurposes)
	sc.Step(`^the dependency should only be imported by files matching:$`, theDependencyShouldOnlyBeImportedByFilesMatching)
	sc.Step(`^no production files should import this dependency$`, noDependencyShouldImportThisDependency)
}
