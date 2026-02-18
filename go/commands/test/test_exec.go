package test

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/ready-to-release/eac/go/clibase/testrunners"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/testing"
)

// TestExecutionContext holds shared state for parallel test execution.
type TestExecutionContext struct {
	testsByPackage  map[string][]testing.TestReference // Keyed by module path (e.g., core/config)
	modulePathToPkg map[string]string                  // Maps module path -> original package path (e.g., go/eac/core/config)
	testParallelism int
	testRunDir      string
	coverage        bool
	suiteTagFilter  core.TagFilter
	workspaceRoot   string
	moduleMapper    *ModuleMapper // Maps package paths to module monikers

	// Suite routing
	suiteMoniker string                   // Suite moniker (single or composite like "unit+integration")
	repoCfg      *config.RepositoryConfig // For computing suite output paths

	// Thread-safe result storage
	mu      sync.Mutex
	results map[string]PackageResult
}

// PackageResult holds per-package test execution results.
type PackageResult struct {
	ModuleMoniker string // Module this package belongs to (for aggregation)
	PackageName   string
	LogFilePath   string
	TestsPassed   int
	TestsFailed   int
	TestsSkipped  int
	TestsTotal    int
	PackageFailed bool
	Duration      time.Duration
}

// createWorker returns an orchestrator worker function for test execution.
func (ctx *TestExecutionContext) createWorker() orchestrator.WorkerFunc {
	return func(goCtx context.Context, pkgPath string, tuiWriter io.Writer) int {
		tests := ctx.testsByPackage[pkgPath]
		result := ctx.runPackageTests(goCtx, pkgPath, tests, tuiWriter)

		ctx.mu.Lock()
		ctx.results[pkgPath] = result
		ctx.mu.Unlock()

		if result.PackageFailed || result.TestsFailed > 0 {
			return 1
		}
		return 0
	}
}

// getEffectiveTestRunDir returns the test run directory for a package.
// With the module-based output structure, all tests go to the same root directory
// (out/test) and the module-specific paths are constructed by the module mapper.
func (ctx *TestExecutionContext) getEffectiveTestRunDir(tests []testing.TestReference) string {
	// Module-based structure: always use the test output root
	// Individual module paths are handled by BuildModuleOutputPath
	return ctx.testRunDir
}

// runPackageTests executes tests for a single package with streaming output
// modulePath is the module-based path (e.g., core/config) used for output organization.
func (ctx *TestExecutionContext) runPackageTests(goCtx context.Context, modulePath string, tests []testing.TestReference, tuiWriter io.Writer) PackageResult {
	// Look up original package path for test execution
	// modulePath is what the orchestrator uses for output directories
	// originalPkgPath is what the runner uses to find test files
	originalPkgPath := ctx.modulePathToPkg[modulePath]
	if originalPkgPath == "" {
		originalPkgPath = modulePath // Fallback if not found
	}

	// Determine test type and get appropriate runner
	testType := getPackageTestType(tests)
	testRunner := testrunners.Get(testType)

	// Extract module moniker using the module mapper which correctly handles
	// godog BDD paths (format: "featureName:testRoot:featurePath")
	moduleMoniker := ""
	if ctx.moduleMapper != nil {
		moduleMoniker = ctx.moduleMapper.GetModuleForPackagePath(originalPkgPath)
	}
	if moduleMoniker == "" {
		// Fallback: simple extraction from modulePath
		moduleMoniker = modulePath
		if idx := strings.Index(modulePath, "/"); idx > 0 {
			moduleMoniker = modulePath[:idx]
		}
	}

	// Get effective test run dir (routes to correct suite folder for composite suites)
	effectiveTestRunDir := ctx.getEffectiveTestRunDir(tests)

	// Use the module path directly as the output path (orchestrator already uses this)
	cfg := testrunners.RunConfig{
		Ctx:              goCtx,
		WorkspaceRoot:    ctx.workspaceRoot,
		TestRunDir:       effectiveTestRunDir,
		Coverage:         ctx.coverage,
		SuiteTagFilter:   ctx.suiteTagFilter,
		Parallelism:      ctx.testParallelism,
		ModuleMoniker:    moduleMoniker, // For result aggregation
		ModuleOutputPath: modulePath,    // Orchestrator creates this directory
	}
	// Pass original package path to runner for test execution
	runResult := testRunner.Execute(originalPkgPath, tests, tuiWriter, cfg)
	return PackageResult{
		ModuleMoniker: runResult.ModuleMoniker,
		PackageName:   runResult.PackageName,
		LogFilePath:   runResult.LogFilePath,
		TestsPassed:   runResult.TestsPassed,
		TestsFailed:   runResult.TestsFailed,
		TestsSkipped:  runResult.TestsSkipped,
		TestsTotal:    runResult.TestsTotal,
		PackageFailed: runResult.PackageFailed,
		Duration:      runResult.Duration,
	}
}

// runPackageTestsDirect executes tests for a single package.
// The orchestrator (UoW) manages log files and output directories.
// logWriter is provided by the orchestrator for all output.
// outputDir is the pre-created UoW output directory (e.g., out/test/<module>/<dirname>).
func (ctx *TestExecutionContext) runPackageTestsDirect(goCtx context.Context, pkgPath string, tests []testing.TestReference, logWriter io.Writer, outputDir string) PackageResult {
	// Determine test type and get appropriate runner
	testType := getPackageTestType(tests)
	testRunner := testrunners.Get(testType)

	// Get effective test run dir (routes to correct suite folder for composite suites)
	effectiveTestRunDir := ctx.getEffectiveTestRunDir(tests)

	cfg := testrunners.RunConfig{
		Ctx:            goCtx,
		WorkspaceRoot:  ctx.workspaceRoot,
		TestRunDir:     effectiveTestRunDir,
		Coverage:       ctx.coverage,
		SuiteTagFilter: ctx.suiteTagFilter,
		Parallelism:    ctx.testParallelism,
		OutputDir:      outputDir,
	}

	runResult := testRunner.Execute(pkgPath, tests, logWriter, cfg)
	return PackageResult{
		ModuleMoniker: runResult.ModuleMoniker,
		PackageName:   runResult.PackageName,
		TestsPassed:   runResult.TestsPassed,
		TestsFailed:   runResult.TestsFailed,
		TestsSkipped:  runResult.TestsSkipped,
		TestsTotal:    runResult.TestsTotal,
		PackageFailed: runResult.PackageFailed,
		Duration:      runResult.Duration,
	}
}

// collectResults returns all collected test results.
func (ctx *TestExecutionContext) collectResults() []PackageResult {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	results := make([]PackageResult, 0, len(ctx.results))
	for _, r := range ctx.results {
		results = append(results, r)
	}
	return results
}
