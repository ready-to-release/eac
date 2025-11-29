// dispatch.go - Test function dispatch and type definitions
package test

import (
	"github.com/ready-to-release/eac/src/commands/impl/test/testers"
	"github.com/ready-to-release/eac/src/core/config"
)

// TestFunc is an alias to the testers package type
type TestFunc = testers.TestFunc

// testFunctions maps module types to their test functions.
// Types not in this map fall back to build system handlers.
var testFunctions = map[string]TestFunc{
	// Go family
	"go-cli":           testers.TestGoCLI,
	"go-commands":      testers.TestGoCommands,
	"go-mcp":           testers.TestGoMCP,
	"go-library":       testers.TestGoLibrary,
	"go-tests":         testers.TestGoTests,
	"go-r2r-extension": testers.TestStaticModule, // Extensions tested via Docker

	// Documentation
	"mkdocs-site": testers.TestMkDocsSite,

	// Repository
	"repository-root": testers.TestRepositoryRoot,

	// Scripts
	"scripts-package": testers.TestScriptsPackage,

	// Static/config modules - validation done at build time
	"catch-all":     testers.TestStaticModule,
	"claude-config": testers.TestStaticModule,
	"configuration": testers.TestStaticModule,
	"templates":     testers.TestStaticModule,
	"vscode-config": testers.TestStaticModule,
	"vscode-ext":    testers.TestStaticModule,
}

// testSystemHandlers maps build systems to default test functions.
// Used when no type-specific handler exists in testFunctions.
var testSystemHandlers = map[string]TestFunc{
	"go":     testers.TestGoDefault,
	"mkdocs": testers.TestStaticModule,
	"docker": testers.TestStaticModule,
	"vscode": testers.TestStaticModule,
	"none":   testers.TestStaticModule,
}

// GetTestFunc returns the appropriate test function for a module type.
// It first checks for a type-specific handler, then falls back to build system handlers.
func GetTestFunc(moduleType string) TestFunc {
	// First, check for type-specific handler
	if fn, exists := testFunctions[moduleType]; exists {
		return fn
	}

	// Fall back to build system handler from type registry
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		buildSystem := cfg.ModuleTypes.GetBuildSystem(moduleType)
		if fn, exists := testSystemHandlers[buildSystem]; exists {
			return fn
		}
	}

	// Ultimate fallback: static module test (no-op)
	return testers.TestStaticModule
}

// initTestSuiteRunner initializes the test suite runner function in the testers package.
// This avoids circular imports between test and testers packages.
func initTestSuiteRunner() {
	testers.RunTestSuiteForModule = runTestSuiteForModule
}

// runTestSuiteForModule runs the test suite command with a module filter.
// This ensures proper test discovery, inference, and suite-based filtering.
func runTestSuiteForModule(moniker string, suiteName string) int {
	return RunTestSuiteForModuleImpl(moniker, suiteName)
}

// RunTestSuiteForModuleImpl is the actual implementation that will be set by suite.go
// to avoid circular dependency issues within the package.
var RunTestSuiteForModuleImpl func(moniker string, suiteName string) int

func init() {
	initTestSuiteRunner()
}
