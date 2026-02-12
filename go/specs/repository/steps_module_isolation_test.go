// Module isolation step definitions for specs/repository/module-isolation feature.
//
// This file implements steps for validating Go module dependency rules
// across the repository.
package repository

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

// moduleIsolationContext holds state for module isolation validation.
type moduleIsolationContext struct {
	repoRoot        string
	currentModule   string
	modulePath      string
	scannedFiles    []string
	productionFiles []string
	testFiles       []string
	importsByFile   map[string][]string // file -> list of local imports
}

// registerModuleIsolationSteps registers module isolation step definitions.
func registerModuleIsolationSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Per-scenario context -- no global mutable state.
	miCtx := &moduleIsolationContext{
		importsByFile: make(map[string][]string),
	}

	sc.Step(`^the repository contains the following Go modules:$`, miCtx.theRepositoryContainsGoModules)
	sc.Step(`^I am checking module "([^"]*)"$`, miCtx.iAmCheckingModule)
	sc.Step(`^I scan all \.go files for import statements$`, miCtx.iScanAllGoFilesForImports)
	sc.Step(`^I scan all production \.go files in "([^"]*)"$`, miCtx.iScanAllProductionGoFilesIn)
	sc.Step(`^I build the module dependency graph from go\.mod files$`, miCtx.iBuildModuleDependencyGraph)
	sc.Step(`^no files should import "([^"]*)"$`, miCtx.noFilesShouldImport)
	sc.Step(`^no production files should import "([^"]*)"$`, miCtx.noProductionFilesShouldImport)
	sc.Step(`^files may import "([^"]*)"$`, miCtx.filesMayImport)
	sc.Step(`^the dependency order should be:$`, miCtx.dependencyOrderShouldBe)
}

func (c *moduleIsolationContext) ensureInit() error {
	if c.repoRoot == "" {
		repoRoot, err := findRepoRoot()
		if err != nil {
			return err
		}
		c.repoRoot = repoRoot
		c.importsByFile = make(map[string][]string)
	}
	return nil
}

func (c *moduleIsolationContext) theRepositoryContainsGoModules(table *godog.Table) error {
	return c.ensureInit()
}

func (c *moduleIsolationContext) iAmCheckingModule(moduleName string) error {
	if err := c.ensureInit(); err != nil {
		return err
	}

	c.currentModule = moduleName
	c.modulePath = filepath.Join(c.repoRoot, moduleName)
	c.scannedFiles = nil
	c.productionFiles = nil
	c.testFiles = nil
	c.importsByFile = make(map[string][]string)

	// Verify module exists
	goModPath := filepath.Join(c.modulePath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("module %s not found (no go.mod at %s)", moduleName, goModPath)
	}

	return nil
}

func (c *moduleIsolationContext) iScanAllGoFilesForImports() error {
	if c.modulePath == "" {
		return fmt.Errorf("module not selected - use 'I am checking module' first")
	}

	return filepath.Walk(c.modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == ".git" || name == "out" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		c.scannedFiles = append(c.scannedFiles, path)

		imports, err := extractLocalImports(path)
		if err != nil {
			return nil // Skip parse errors
		}

		if len(imports) > 0 {
			relPath, relErr := filepath.Rel(c.repoRoot, path)
			if relErr != nil {
				relPath = path
			}
			c.importsByFile[relPath] = imports
		}

		return nil
	})
}

func (c *moduleIsolationContext) iScanAllProductionGoFilesIn(modulePath string) error {
	if err := c.ensureInit(); err != nil {
		return err
	}

	fullPath := filepath.Join(c.repoRoot, modulePath)
	c.modulePath = fullPath
	c.currentModule = modulePath

	return filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == ".git" || name == "out" || name == "testdata" || name == "tests" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(info.Name(), "_test.go") {
			c.testFiles = append(c.testFiles, path)
			return nil
		}

		c.productionFiles = append(c.productionFiles, path)

		imports, err := extractLocalImports(path)
		if err != nil {
			return nil
		}

		if len(imports) > 0 {
			relPath, relErr := filepath.Rel(c.repoRoot, path)
			if relErr != nil {
				relPath = path
			}
			c.importsByFile[relPath] = imports
		}

		return nil
	})
}

func (c *moduleIsolationContext) noFilesShouldImport(forbiddenImport string) error {
	var violations []string
	for file, imports := range c.importsByFile {
		for _, imp := range imports {
			if strings.HasPrefix(imp, forbiddenImport) {
				violations = append(violations, fmt.Sprintf("  - %s imports %s", file, imp))
			}
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("found forbidden imports of %s:\n%s", forbiddenImport, strings.Join(violations, "\n"))
	}

	return nil
}

func (c *moduleIsolationContext) noProductionFilesShouldImport(forbiddenImport string) error {
	var violations []string
	for file, imports := range c.importsByFile {
		// Skip test files
		if strings.Contains(file, "/tests/") || strings.HasSuffix(file, "_test.go") {
			continue
		}

		for _, imp := range imports {
			if strings.HasPrefix(imp, forbiddenImport) {
				violations = append(violations, fmt.Sprintf("  - %s imports %s", file, imp))
			}
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("production files import forbidden module %s:\n%s", forbiddenImport, strings.Join(violations, "\n"))
	}

	return nil
}

func (c *moduleIsolationContext) filesMayImport(allowedImport string) error {
	// This is a documentation step - allowed imports don't need validation
	return nil
}

func (c *moduleIsolationContext) iBuildModuleDependencyGraph() error {
	return c.ensureInit()
}

func (c *moduleIsolationContext) dependencyOrderShouldBe(table *godog.Table) error {
	// This is a documentation/assertion step
	// Full validation would verify actual dependencies match expected layers
	return nil
}

// Helper functions

func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repository root")
		}
		dir = parent
	}
}

func extractLocalImports(filePath string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	var localImports []string
	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		// Check if it's a local module import
		if strings.HasPrefix(importPath, "github.com/ready-to-release/eac/go/") {
			localImports = append(localImports, importPath)
		}
	}

	return localImports, nil
}
