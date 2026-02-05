// Package test provides the test command implementation using cmdframework.
package test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/artifacts"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/manifests"
	"github.com/ready-to-release/eac/go/cli/eac/impl/show"
	testresults "github.com/ready-to-release/eac/go/cli/eac/impl/test/internal/results"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/execution"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/logging"
	moduledeps "github.com/ready-to-release/eac/go/core/module-deps"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func init() {
	// Register test component-level execution support
	cmdframework.RegisterUnitProvider(cmdframework.CommandTypeTest, ResolveTestUnitSpecs)
	cmdframework.RegisterUnitWorker(cmdframework.CommandTypeTest, testUnitWorker)
}

// TestFrameworkConfig holds test-specific configuration for the framework.
type TestFrameworkConfig struct {
	// Input configuration
	SuiteName   string
	Coverage    bool
	ForceRetest bool
	Parallel    bool
	ListOnly    bool

	// Resolved during execution
	Suite           *testing.TestSuite
	SelectedTests   []testing.TestReference
	TestsByPackage  map[string][]testing.TestReference // Package path -> tests
	ModuleMapper    *ModuleMapper
	TestRunDir      string
	SuiteTagFilter  string
	OSFilteredCount int

	// Execution paths (using package paths directly)
	ParallelPaths   []string
	SequentialPaths []string

	// Stats for summary
	Stats *testSelectionStats

	// UoW-level incremental state
	CachedUoWs    map[string]bool      // UoW longname -> cached
	UoWCacheTimes map[string]time.Time // UoW longname -> cache time
	Tracker       *coreoutput.InMemoryTracker

	// Module-level incremental state (aggregated from UoWs for TUI compatibility)
	CachedModules map[string]bool      // Modules that are up-to-date (cache hits)
	CacheTimes    map[string]time.Time // Module-level cache times for TUI

	// Component mapping for clean directory names
	// Maps "cleanComponent" (e.g., "docs-drawio-cache/godog") to full pkgPath
	ComponentToPkgPath map[string]string

	// Execution state
	ExecCtx       *TestExecutionContext
	TestStartTime time.Time
	// NOTE: Lock acquisition moved to testUnitWorker (after cache check)
}

// testSelectionStats tracks test selection statistics.
type testSelectionStats struct {
	TotalDiscovered  int
	Skipped          int
	NotMatchingSuite int
	Selected         int
	OSFiltered       int
	ModulesRequested []string
	ModulesInScope   []string
	ModulesNoTests   []string
}

// RunTestWithFramework executes tests using the cmdframework.
func RunTestWithFramework(cmdCfg *cmdframework.CommandConfig, testCfg *TestFrameworkConfig) int {
	// Store test config in Extra for access in hooks/worker
	if cmdCfg.Extra == nil {
		cmdCfg.Extra = make(map[string]interface{})
	}
	cmdCfg.Extra["testConfig"] = testCfg
	testCfg.Stats = &testSelectionStats{ModulesRequested: cmdCfg.Monikers}

	// Set up hooks
	hooks := &cmdframework.Hooks{
		AfterInit:     testAfterInit,
		AfterResolve:  testAfterResolve,
		BeforeExecute: testBeforeExecute,
		AfterExecute:  testAfterExecute,
	}

	// Don't register validators - test handles them inline with custom logic
	// cmdframework.SetArtifactValidator(nil)
	// cmdframework.SetDepsVerifier(nil)

	return cmdframework.Run(cmdCfg, testWorker, hooks)
}

// testAfterInit handles suite resolution and lock acquisition.
func testAfterInit(ctx *cmdframework.ExecutionContext) error {
	testCfg, ok := ctx.Config.Extra["testConfig"].(*TestFrameworkConfig)
	if !ok {
		return fmt.Errorf("testConfig not found or wrong type")
	}

	// Resolve suite from config if not specified
	if testCfg.SuiteName == "" {
		defaultSuites := ctx.EACConfig.Testing.GetDefaultSuites()
		if len(defaultSuites) == 0 {
			return fmt.Errorf("no default suites configured")
		}
		testCfg.SuiteName = strings.Join(defaultSuites, "+")
		log.Debugf("Using default suites: %s", testCfg.SuiteName)
	}

	// Load suite configuration
	suite, err := testing.GetSuite(testCfg.SuiteName)
	if err != nil {
		ctx.WriteInit("Suite not found: %s", testCfg.SuiteName)
		ctx.WriteInit("Available suites:")
		for _, s := range ctx.EACConfig.Testing.ListSuites() {
			ctx.WriteInit("  - %s", s)
		}
		return fmt.Errorf("suite not found: %s", testCfg.SuiteName)
	}
	testCfg.Suite = suite

	// NOTE: Lock acquisition moved to testUnitWorker (after cache check)
	// This ensures locks are only held during actual test execution, not during TUI init

	// Create test output directory
	repoCfg := ctx.EACConfig.Repository
	testCfg.TestRunDir = filepath.Join(ctx.WorkspaceRoot, repoCfg.TestOutputDir())
	if err := os.MkdirAll(testCfg.TestRunDir, 0o755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	// Load module contracts and registry (test command uses SkipResolve, so we load it here)
	moduleReport, err := reports.GetModuleContracts(ctx.WorkspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to load module contracts: %w", err)
	}
	ctx.ModuleReport = moduleReport
	ctx.ModuleRegistry = moduleReport.Registry

	return nil
}

// testAfterResolve handles test discovery, filtering, and execution plan setup.
func testAfterResolve(ctx *cmdframework.ExecutionContext) error {
	testCfg, ok := ctx.Config.Extra["testConfig"].(*TestFrameworkConfig)
	if !ok {
		return fmt.Errorf("testConfig not found or wrong type")
	}
	suite := testCfg.Suite
	stats := testCfg.Stats

	// Build module mapper early - used for test filtering and module ownership
	testCfg.ModuleMapper = NewModuleMapper(ctx.ModuleRegistry, ctx.WorkspaceRoot)

	// Test Discovery with suite-specific inferences
	allTests, err := testing.DiscoverAndEnrich(ctx.WorkspaceRoot, testing.DiscoveryOptions{
		Inferences: suite.Inferences,
	})
	if err != nil {
		return fmt.Errorf("failed to discover tests: %w", err)
	}
	stats.TotalDiscovered = len(allTests)
	log.Debugf("Discovered %d tests with %d inference rules applied", len(allTests), len(suite.Inferences))

	// Filter tests by suite and modules
	validSkipReasons := ctx.EACConfig.Testing.GetValidSkipReasons()
	skipReasonSet := make(map[string]bool)
	for _, reason := range validSkipReasons {
		skipReasonSet[reason] = true
	}
	requestedMonikers := make(map[string]bool)
	for _, m := range ctx.Config.Monikers {
		requestedMonikers[m] = true
	}

	var selectedTests []testing.TestReference
	var unmappedTests []string
	for i := range allTests {
		test := &allTests[i]
		// Check for @skip tags
		isSkipped := false
		for _, tag := range test.Tags {
			for reason := range skipReasonSet {
				if tag == "@skip:"+reason {
					isSkipped = true
					break
				}
			}
			if isSkipped {
				break
			}
		}
		if isSkipped {
			stats.Skipped++
			continue
		}

		// Check if test matches suite
		if !suite.Matches(*test) {
			stats.NotMatchingSuite++
			continue
		}

		// Get module ownership from registry (not heuristic)
		testModule := testCfg.ModuleMapper.GetModuleForFile(test.FilePath)
		if testModule == "" {
			// Track unmapped tests for fail-fast reporting
			unmappedTests = append(unmappedTests, test.FilePath)
			continue
		}

		// Check module filter if specified
		if len(requestedMonikers) > 0 && !requestedMonikers[testModule] {
			continue
		}

		selectedTests = append(selectedTests, *test)
	}

	// Fail fast if tests couldn't be mapped to modules
	if len(unmappedTests) > 0 {
		log.Warnf("Found %d tests that could not be mapped to any module:", len(unmappedTests))
		for _, path := range unmappedTests {
			log.Warnf("  - %s", path)
		}
		return fmt.Errorf("test discovery failed: %d tests have no module ownership - check repository.yml module patterns", len(unmappedTests))
	}

	stats.Selected = len(selectedTests)

	// Filter by OS compatibility
	osCompatibleTests := filterByOSCompatibility(selectedTests, os.Stdout)
	stats.OSFiltered = len(selectedTests) - len(osCompatibleTests)
	selectedTests = osCompatibleTests
	testCfg.OSFilteredCount = stats.OSFiltered

	// Calculate module stats
	stats.ModulesInScope = getUniqueModulesFromTests(selectedTests, testCfg.ModuleMapper)

	// Find modules with no tests
	inScopeSet := make(map[string]bool)
	for _, m := range stats.ModulesInScope {
		inScopeSet[m] = true
	}
	var modulesToCheck []string
	if len(ctx.Config.Monikers) > 0 {
		modulesToCheck = ctx.Config.Monikers
	} else {
		modulesToCheck = ctx.EACConfig.Repository.AllMonikers()
	}
	for _, m := range modulesToCheck {
		if !inScopeSet[m] {
			stats.ModulesNoTests = append(stats.ModulesNoTests, m)
		}
	}

	testCfg.SelectedTests = selectedTests

	// Handle list-only mode
	if testCfg.ListOnly {
		ctx.WriteInit("=== Selected Tests ===")
		for i := range selectedTests {
			test := &selectedTests[i]
			ctx.WriteInit("%d. %s (%s)", i+1, test.TestName, test.Type)
			ctx.WriteInit("   File: %s", test.FilePath)
			ctx.WriteInit("   Tags: %s", strings.Join(test.Tags, ", "))
			ctx.WriteInit("")
		}
		// Return early - execution plan will be empty
		ctx.ExecutionPlan = &repository.ExecutionPlan{}
		return nil
	}

	// Dependency Verification
	if err := verifyTestDependencies(ctx, testCfg); err != nil {
		return err
	}

	// Build Artifact Validation
	if err := validateTestArtifacts(ctx, testCfg); err != nil {
		return err
	}

	// Group tests by package
	testsByPackage := groupTestsByPackage(selectedTests, ctx.WorkspaceRoot, ctx.EACConfig)
	testCfg.TestsByPackage = testsByPackage

	if len(testsByPackage) == 0 {
		ctx.WriteInit("No test packages to execute")
		ctx.ExecutionPlan = &repository.ExecutionPlan{}
		return nil
	}

	// UoW-based incremental test detection (devbox only)
	if !environments.IsCI() && !testCfg.ForceRetest && !ctx.Config.DryRun {
		testsByPackage = detectUoWIncrementalTestChanges(ctx, testCfg, testsByPackage)
		testCfg.TestsByPackage = testsByPackage
	}

	// Separate parallel vs sequential paths using package paths directly
	// (No conversion to module paths - tests use path-based keys)
	var parallelPaths, sequentialPaths []string
	moduleTypes := make(map[string]string)
	moduleTypeSet := make(map[string]map[string]bool) // Track unique test types per moniker

	for pkgPath, tests := range testsByPackage {
		hasSequential := false
		for i := range tests {
			test := &tests[i]
			if test.IsSequential {
				hasSequential = true
				break
			}
		}

		if hasSequential {
			sequentialPaths = append(sequentialPaths, pkgPath)
		} else {
			parallelPaths = append(parallelPaths, pkgPath)
		}

		// Aggregate test types by moniker using module mapper
		if len(tests) > 0 {
			moniker := extractMonikerFromPath(pkgPath, testCfg.ModuleMapper)
			if moduleTypeSet[moniker] == nil {
				moduleTypeSet[moniker] = make(map[string]bool)
			}
			// Capture ALL test types from this package, not just the first
			for i := range tests {
				moduleTypeSet[moniker][tests[i].Type] = true
			}
		}
	}

	// Build moduleTypes map with aggregated types (e.g., "gotest, godog")
	for moniker, types := range moduleTypeSet {
		var typeList []string
		for t := range types {
			typeList = append(typeList, t)
		}
		sort.Strings(typeList)
		moduleTypes[moniker] = strings.Join(typeList, ", ")
	}

	testCfg.ParallelPaths = parallelPaths
	testCfg.SequentialPaths = sequentialPaths
	testCfg.SuiteTagFilter = buildSuiteTagFilter(suite)

	// Convert package paths to module monikers for TUI tree building.
	// ExecutionPlan.ExecutionOrder must contain monikers (not paths) so buildTreeFromUnitSpecs
	// can match them with UnitSpec.ID.Module for proper tab pre-filling.
	parallelModules := extractUniqueModulesFromPaths(parallelPaths, testCfg.ModuleMapper)
	sequentialModules := extractUniqueModulesFromPaths(sequentialPaths, testCfg.ModuleMapper)

	// Remove modules that appear in both lists (keep in parallel only)
	sequentialModules = removeExistingModules(sequentialModules, parallelModules)

	// All modules for execution order (parallel first, then sequential)
	var allModules []string
	allModules = append(allModules, parallelModules...)
	allModules = append(allModules, sequentialModules...)

	// Set up execution plan with module monikers
	ctx.ExecutionPlan = &repository.ExecutionPlan{
		ExecutionOrder: allModules,
	}
	ctx.ComponentTypesDisplay = moduleTypes
	ctx.Orchestrator.SetComponentTypesDisplay(moduleTypes)

	// Build init summary
	buildTestInitSummary(ctx, testCfg)

	return nil
}

// testBeforeExecute initializes the test execution context.
func testBeforeExecute(ctx *cmdframework.ExecutionContext) error {
	testCfg, ok := ctx.Config.Extra["testConfig"].(*TestFrameworkConfig)
	if !ok {
		return fmt.Errorf("testConfig not found or wrong type")
	}

	if len(ctx.ExecutionPlan.ExecutionOrder) == 0 {
		return nil // Nothing to execute
	}

	repoCfg := ctx.EACConfig.Repository
	testParallelism := repoCfg.EffectiveParallelism(environments.IsCI())
	if !testCfg.Parallel {
		testParallelism = 1
	}

	testCfg.ExecCtx = &TestExecutionContext{
		testsByPackage:  testCfg.TestsByPackage,
		modulePathToPkg: nil, // No longer needed - using package paths directly
		testParallelism: testParallelism,
		testRunDir:      testCfg.TestRunDir,
		coverage:        testCfg.Coverage,
		suiteTagFilter:  testCfg.SuiteTagFilter,
		workspaceRoot:   ctx.WorkspaceRoot,
		moduleMapper:    testCfg.ModuleMapper,
		suiteMoniker:    testCfg.Suite.Moniker,
		repoCfg:         repoCfg,
		results:         make(map[string]PackageResult),
	}
	testCfg.TestStartTime = time.Now()

	// Enable early cache detection for fast TUI feedback
	// Tabs will progressively "light up" blue as cache hits are detected
	if (len(testCfg.CachedUoWs) > 0 || len(testCfg.CachedModules) > 0) && ctx.Orchestrator != nil {
		verifier := &TestCacheVerifier{
			cachedUoWs:    testCfg.CachedUoWs,
			uowCacheTimes: testCfg.UoWCacheTimes,
			cachedModules: testCfg.CachedModules,
		}
		ctx.Orchestrator.SetCacheDetection(verifier, testCfg.CachedModules)

		// Pass cache times if available
		if len(testCfg.CacheTimes) > 0 {
			ctx.Orchestrator.SetCacheTimes(testCfg.CacheTimes)
		}
	}

	return nil
}

// TestCacheVerifier implements execution.CacheVerifier for test commands.
// Uses UoW-level cache for fine-grained test caching.
type TestCacheVerifier struct {
	cachedUoWs    map[string]bool      // UoW longname -> cached
	uowCacheTimes map[string]time.Time // UoW longname -> cache time
	cachedModules map[string]bool      // Module-level cache (aggregated from UoWs for TUI)
}

// Verify implements execution.CacheVerifier.
func (v *TestCacheVerifier) Verify(ctx context.Context, unit workunit.UnitSpec) (execution.CacheResult, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return execution.CacheResult{}, ctx.Err()
	default:
	}

	longname := unit.ID.Longname()

	// Check UoW-level cache first
	if v.cachedUoWs != nil && v.cachedUoWs[longname] {
		return execution.CacheResult{
			Cached:    true,
			CacheTime: v.uowCacheTimes[longname],
		}, nil
	}

	// Fall back to module-level check for backwards compatibility
	if v.cachedModules != nil && v.cachedModules[unit.ID.Module] {
		return execution.CacheResult{Cached: true}, nil
	}

	return execution.CacheResult{}, nil
}

// testAfterExecute handles manifest generation and state updates.
func testAfterExecute(ctx *cmdframework.ExecutionContext) error {
	testCfg, ok := ctx.Config.Extra["testConfig"].(*TestFrameworkConfig)
	if !ok {
		return fmt.Errorf("testConfig not found or wrong type")
	}

	if testCfg.ExecCtx == nil || ctx.Config.DryRun {
		return nil
	}

	results := testCfg.ExecCtx.collectResults()
	testDuration := time.Since(testCfg.TestStartTime)
	repoCfg := ctx.EACConfig.Repository

	// Parse cucumber results by module
	cucumberResultsByModule := make(map[string][]manifests.CucumberTestResult)
	for _, result := range results {
		moniker := result.ModuleMoniker
		if moniker == "" {
			parts := strings.Split(result.PackageName, "/")
			if len(parts) > 0 {
				moniker = parts[0]
			}
		}
		if moniker != "" {
			if _, exists := cucumberResultsByModule[moniker]; !exists {
				moduleTestDir := repoCfg.TestModuleDirAbs(ctx.WorkspaceRoot, moniker)
				for _, r := range testresults.ParseCucumberResults(moduleTestDir) {
					cucumberResultsByModule[moniker] = append(cucumberResultsByModule[moniker], manifests.CucumberTestResult{
						ScenarioName: r.ScenarioName,
						FeaturePath:  r.FeaturePath,
						Status:       r.Status,
						DurationMs:   r.DurationMs,
						Tags:         r.Tags,
					})
				}
			}
		}
	}

	// Generate test manifests
	manifests.GenerateTestManifests(
		results,
		testCfg.TestsByPackage,
		cucumberResultsByModule,
		testCfg.SuiteName,
		testDuration,
		ctx.WorkspaceRoot,
		repoCfg,
		ctx.EACConfig,
	)

	// Aggregate reports
	if path := testresults.AggregateCucumberReports(testCfg.TestRunDir); path != "" {
		log.Infof("Aggregated cucumber report: %s", path)
	}
	if path := testresults.AggregateCTRFReports(testCfg.TestRunDir); path != "" {
		log.Infof("Aggregated CTRF report: %s", path)
	}

	// UoW manifests are written immediately in writeUoWTestManifest via RecordComplete
	// No explicit state update needed here

	// Show timing analysis if requested
	if ctx.Config.ShowTimings {
		showTestTimings(ctx, testCfg)
	}

	// Assert all UoWs have valid manifests
	if err := assertTestManifestsExist(ctx); err != nil {
		return err
	}

	return nil
}

// assertTestManifestsExist verifies that all executed UoWs have valid manifests.
func assertTestManifestsExist(ctx *cmdframework.ExecutionContext) error {
	return cmdframework.AssertManifestsExist(ctx, "test", ResolveTestUnitSpecs(ctx))
}

// testWorker runs tests for a module path.
func testWorker(ctx *cmdframework.ExecutionContext, modulePath string, logWriter io.Writer) int {
	testCfg, ok := ctx.Config.Extra["testConfig"].(*TestFrameworkConfig)
	if !ok || testCfg == nil {
		fmt.Fprintf(logWriter, "Error: testConfig not found or wrong type\n")
		return 1
	}

	if testCfg.ExecCtx == nil {
		return 1
	}

	tests := testCfg.ExecCtx.testsByPackage[modulePath]
	result := testCfg.ExecCtx.runPackageTests(modulePath, tests, logWriter)

	testCfg.ExecCtx.mu.Lock()
	testCfg.ExecCtx.results[modulePath] = result
	testCfg.ExecCtx.mu.Unlock()

	if result.PackageFailed || result.TestsFailed > 0 {
		return 1
	}
	return 0
}

// testUnitWorker runs tests for a package path using component-level execution.
// This is called by the UnitScheduler for parallel test component execution.
// The component parameter is in "componentType:toolName:testname" format (e.g., "go:gotest:impl-build").
// The orchestrator (UoW) creates the log file and output directory - worker just writes to logWriter.
func testUnitWorker(ctx *cmdframework.ExecutionContext, module, component string, logWriter io.Writer) int {
	testCfg, ok := ctx.Config.Extra["testConfig"].(*TestFrameworkConfig)
	if !ok || testCfg == nil {
		fmt.Fprintf(logWriter, "Error: testConfig not found or wrong type\n")
		return 1
	}

	if testCfg.ExecCtx == nil {
		fmt.Fprintf(logWriter, "Error: test execution context not initialized\n")
		return 1
	}

	// Parse component:tool:testname format (e.g., "go:gotest:impl-build")
	// Parts: [0]=componentType, [1]=tool, [2]=testname
	parts := strings.Split(component, ":")
	componentType := parts[0]
	testType := ""
	testname := ""
	if len(parts) >= 2 {
		testType = parts[1]
	}
	if len(parts) >= 3 {
		testname = parts[2]
	}

	// Build UnitID for UoW-level cache lookup
	toolName := testType
	if toolName == "" {
		toolName = "none"
	}
	unitID := workunit.UnitID{
		Context:   workunit.ContextTest,
		Module:    module,
		Component: componentType,
		Tool:      toolName,
		Extra:     map[string]string{"testname": testname},
	}

	// Check UoW-level cache first
	isCached := testCfg.CachedUoWs != nil && testCfg.CachedUoWs[unitID.Longname()]
	log.Debugf("[TEST-UOW-CACHE] Test worker for %s: unitID=%s, isCached=%v", component, unitID.Longname(), isCached)

	if isCached {
		fmt.Fprintf(logWriter, "⏭️  Cached (unchanged)\n")
		return -1 // -1 = skipped/cached = blue in TUI
	}

	// Look up pkgPath from component mapping using testname
	// testname is the unique identifier within module:componentType (e.g., "impl-build")
	pkgPath, ok := testCfg.ComponentToPkgPath[testname]
	if !ok {
		fmt.Fprintf(logWriter, "Error: no pkgPath mapping for testname %s\n", testname)
		return 1
	}

	tests := testCfg.ExecCtx.testsByPackage[pkgPath]
	if len(tests) == 0 {
		fmt.Fprintf(logWriter, "No tests: Success\n")
		return 0
	}

	// Filter tests by type if specified
	if testType != "" {
		tests = filterTestsByType(tests, testType)
		if len(tests) == 0 {
			fmt.Fprintf(logWriter, "No %s tests: Success\n", testType)
			return 0
		}
	}

	// Compute input hash before running tests
	var inputHash string
	if contract, ok := ctx.ModuleRegistry.Get(module); ok {
		inputHash, _ = computeTestInputHash(ctx, contract)
	}

	startTime := time.Now()

	// Run tests - UoW manages log file, we just write to logWriter
	// Use the testname as result key for aggregation (unique within module:componentType)
	resultKey := testname
	result := testCfg.ExecCtx.runPackageTestsDirect(pkgPath, tests, logWriter)

	testCfg.ExecCtx.mu.Lock()
	testCfg.ExecCtx.results[resultKey] = result
	testCfg.ExecCtx.mu.Unlock()

	// Pass test counts to orchestrator for summary display
	// Use componentType as the component name for orchestrator tracking
	ctx.Orchestrator.SetUnitExtras(module, componentType, orchestrator.UnitExtras{
		TestsTotal:   result.TestsTotal,
		TestsPassed:  result.TestsPassed,
		TestsFailed:  result.TestsFailed,
		TestsSkipped: result.TestsSkipped,
	})

	// Write UoW manifest for incremental cache
	passed := !result.PackageFailed && result.TestsFailed == 0
	exitCode := 0
	if !passed {
		exitCode = 1
	}
	writeUoWTestManifest(ctx, testCfg, unitID, inputHash, startTime, exitCode)

	if result.PackageFailed || result.TestsFailed > 0 {
		return 1
	}
	return 0
}

// filterTestsByType returns only tests matching the specified type.
func filterTestsByType(tests []testing.TestReference, testType string) []testing.TestReference {
	var filtered []testing.TestReference
	for i := range tests {
		if tests[i].Type == testType {
			filtered = append(filtered, tests[i])
		}
	}
	return filtered
}

// Helper functions

// extractMonikerFromPath extracts the module moniker from a package path.
// Returns empty string if mapper is nil or path doesn't map to any module.
func extractMonikerFromPath(pkgPath string, mapper *ModuleMapper) string {
	if mapper == nil {
		return ""
	}
	return mapper.GetModuleForPackagePath(pkgPath)
}

func verifyTestDependencies(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig) error {
	systemDeps := testing.GetSystemDependencies(testCfg.SelectedTests)
	moduleDeps := testing.GetModuleDependencies(testCfg.SelectedTests)

	if len(systemDeps) == 0 && len(moduleDeps) == 0 {
		return nil
	}

	if ctx.Config.SkipDeps {
		return nil
	}

	// Strip @deps: prefix from system dependencies to get tool IDs
	// @deps: tags use tool IDs (e.g., @deps:go, @deps:docker, @deps:go-lint)
	toolIDs := make([]string, 0, len(systemDeps))
	for _, dep := range systemDeps {
		toolID := strings.TrimPrefix(dep, "@deps:")
		toolIDs = append(toolIDs, toolID)
	}

	// Filter out platform-incompatible tools before verification
	toolIDs = tool.FilterPlatformSupported(toolIDs)

	// Verify system dependencies using tool registry
	registry := tool.GlobalRegistry()
	sysResults := registry.VerifyAll(toolIDs)
	var missing []string
	for _, result := range sysResults {
		if !result.Available {
			missing = append(missing, result.ToolID)
		}
	}

	if len(missing) > 0 {
		ctx.WriteInit("❌ Required system dependencies are missing: %s", strings.Join(missing, ", "))
		ctx.WriteInit("   Use --skip-deps to run tests anyway")
		return fmt.Errorf("missing system dependencies: %s", strings.Join(missing, ", "))
	}

	// Verify module dependencies
	modResults := moduledeps.VerifyAll(moduleDeps)
	for _, result := range modResults {
		if !result.Available {
			log.Warnf("Module dependency not available: %s", result.Dependency)
		}
	}

	return nil
}

func validateTestArtifacts(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig) error {
	stats := testCfg.Stats

	if len(stats.ModulesInScope) == 0 || ctx.Config.SkipDepm {
		return nil
	}

	// Get expected build UoWs to filter validation (ignores orphaned manifests)
	expectedBuildUoWs := getExpectedBuildUoWs(ctx)

	artifactValidation := artifacts.ValidateBuildArtifactsWithExpected(
		stats.ModulesInScope,
		ctx.EACConfig,
		ctx.WorkspaceRoot,
		ctx.ModuleRegistry,
		expectedBuildUoWs,
	)

	if artifactValidation.AllValid() {
		return nil
	}

	// Collect affected modules for the resolution command
	var affectedModules []string

	ctx.WriteInit("")
	ctx.WriteInit("┌─────────────────────────────────────────────────────────────────┐")
	ctx.WriteInit("│  Build Required                                                 │")
	ctx.WriteInit("└─────────────────────────────────────────────────────────────────┘")
	ctx.WriteInit("")

	if !artifactValidation.AllPresent {
		ctx.WriteInit("Missing build artifacts:")
		for _, moduleName := range artifactValidation.MissingFrom {
			ctx.WriteInit("  • %s", moduleName)
			affectedModules = append(affectedModules, moduleName)
		}
		ctx.WriteInit("")
	}

	if !artifactValidation.AllCurrent {
		ctx.WriteInit("Stale build artifacts (source changed since last build):")
		for _, moduleName := range artifactValidation.StaleModules {
			reason := artifactValidation.StaleReasons[moduleName]
			ctx.WriteInit("  • %s: %s", moduleName, reason)
			// Only add if not already in the list
			found := false
			for _, m := range affectedModules {
				if m == moduleName {
					found = true
					break
				}
			}
			if !found {
				affectedModules = append(affectedModules, moduleName)
			}
		}
		ctx.WriteInit("")
	}

	// Show resolution command
	ctx.WriteInit("To fix, run:")
	if len(affectedModules) == 1 {
		ctx.WriteInit("  eac build %s", affectedModules[0])
	} else if len(affectedModules) <= 3 {
		ctx.WriteInit("  eac build %s", strings.Join(affectedModules, " "))
	} else {
		ctx.WriteInit("  eac build")
	}
	ctx.WriteInit("")

	// Return informational exit - no additional error message needed
	return cmdframework.NewInformationalExit("build required")
}

// getExpectedBuildUoWs returns the expected build UoWs for each module.
// This is used to filter artifact validation to only check manifests for currently-configured
// components, preventing false positives from orphaned manifests left by removed component types.
func getExpectedBuildUoWs(ctx *cmdframework.ExecutionContext) map[string][]workunit.UnitID {
	// Use build command's resolver to get expected build UoWs
	buildSpecs := build.ResolveUnitSpecs(ctx)
	if len(buildSpecs) == 0 {
		return nil
	}

	result := make(map[string][]workunit.UnitID)
	for _, spec := range buildSpecs {
		result[spec.ID.Module] = append(result[spec.ID.Module], spec.ID)
	}
	return result
}

// detectUoWIncrementalTestChanges performs UoW-level incremental test detection.
// Instead of checking at module granularity, it checks each component:tool UoW.
// This enables partial caching - some test components can be cached while others retest.
func detectUoWIncrementalTestChanges(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig, testsByPackage map[string][]testing.TestReference) map[string][]testing.TestReference {
	startTime := time.Now()
	defer func() {
		ctx.SetChangeDetectionTiming(time.Since(startTime))
	}()

	// Skip incremental if no module registry available
	if ctx.ModuleRegistry == nil {
		log.Debugf("Module registry not available, skipping incremental test detection")
		return testsByPackage
	}

	// Get all expected UoWs from resolved unit specs
	specs := ResolveTestUnitSpecs(ctx)
	if len(specs) == 0 {
		return testsByPackage
	}

	// Build list of expected UoWs
	var expectedUoWs []workunit.UnitID
	for _, spec := range specs {
		expectedUoWs = append(expectedUoWs, spec.ID)
	}

	if len(expectedUoWs) == 0 {
		return testsByPackage
	}

	// Collect module files for hash computation
	moduleFiles := make(map[string][]string)
	for _, id := range expectedUoWs {
		if _, ok := moduleFiles[id.Module]; ok {
			continue // Already collected
		}
		if contract, ok := ctx.ModuleRegistry.Get(id.Module); ok {
			patterns := contract.GetGlobPatterns()
			files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
			if err != nil {
				log.Debugf("Failed to expand patterns for %s: %v", id.Module, err)
				continue
			}
			moduleFiles[id.Module] = files
		}
	}

	// Use shared helpers for change detection and aggregation
	reader := coreoutput.NewReader(ctx.WorkspaceRoot)
	getInputHash := coreoutput.InputHashProvider(hash.NewModuleInputHashProvider(ctx.WorkspaceRoot, moduleFiles))

	aggResult, err := coreoutput.AggregateUoWChanges(reader, workunit.ContextTest, expectedUoWs, getInputHash)
	if err != nil {
		log.Debugf("Failed to detect UoW changes: %v", err)
		return testsByPackage
	}

	// Log change detection results
	log.Debugf("[TEST-UOW-CACHE] DetectUoWChanges result: FreshRun=%v Changed=%d UpToDate=%d",
		aggResult.UoWResult.FreshRun, len(aggResult.UoWResult.Changed), len(aggResult.UoWResult.UpToDate))
	for longname, reason := range aggResult.UoWResult.ChangeReasons {
		log.Debugf("[TEST-UOW-CACHE] Changed: %s -> %s", longname, reason)
	}

	detectionTime := time.Since(startTime)

	if aggResult.UoWResult.FreshRun {
		log.Debugf("Fresh test detected (UoW mode), all tests will run")
		if ctx.InitSummary != nil {
			ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
				Enabled:       true,
				DetectionTime: detectionTime,
				FreshBuild:    true,
			})
		}
		return testsByPackage
	}

	// Copy aggregated results to test config
	testCfg.CachedUoWs = aggResult.CachedUoWs
	testCfg.UoWCacheTimes = aggResult.UoWCacheTimes
	testCfg.CachedModules = aggResult.CachedModules
	testCfg.CacheTimes = aggResult.ModuleCacheTimes

	// Log module-level aggregation
	agg := workunit.NewUoWAggregator(expectedUoWs)
	for _, id := range aggResult.UoWResult.UpToDate {
		agg.MarkCached(id)
	}
	for module := range aggResult.CachedModules {
		total, cached := agg.Stats(module)
		log.Debugf("[TEST-UOW-CACHE] Module %s: %d/%d UoWs cached -> module cached=%v",
			module, cached, total, true)
	}
	for _, module := range aggResult.ChangedModules {
		total, cached := agg.Stats(module)
		log.Debugf("[TEST-UOW-CACHE] Module %s: %d/%d UoWs cached -> module cached=%v",
			module, cached, total, false)
	}

	// Report incremental detection in init summary
	if ctx.InitSummary != nil {
		ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
			Enabled:       true,
			DetectionTime: detectionTime,
			Changed:       aggResult.ChangedModules,
			UpToDate:      aggResult.UpToDateModules,
			FreshBuild:    false,
		})
	}

	log.Debugf("Incremental (UoW mode): %d modules to test, %d cached, %d UoWs cached",
		len(aggResult.ChangedModules), len(aggResult.UpToDateModules), len(testCfg.CachedUoWs))

	// Return ALL packages - filtering happens at component worker level
	return testsByPackage
}

func buildSuiteTagFilter(suite *testing.TestSuite) string {
	return suite.BuildGodogTagFilter()
}

func buildTestInitSummary(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig) {
	stats := testCfg.Stats
	suite := testCfg.Suite

	summary := initsummary.New("test").
		SetRequest(ctx.Config.Monikers, ctx.Config.Monikers).
		SetExecutionContext(string(logging.GetExecutionContext())).
		SetFlags(initsummary.Flags{
			ListOnly:    testCfg.ListOnly,
			ShowTimings: ctx.Config.ShowTimings,
			DebugMode:   ctx.Config.DebugMode,
			UseTUI:      ctx.Config.UseTUI,
		}).
		SetTestInfo(&initsummary.TestInfo{
			SuiteName:             suite.Name,
			SuiteDescription:      suite.Description,
			SuitesIncluded:        manifests.GetSuitesIncluded(suite.Moniker),
			TotalDiscovered:       stats.TotalDiscovered,
			Skipped:               stats.Skipped,
			NotMatchingSuite:      stats.NotMatchingSuite,
			OSFiltered:            stats.OSFiltered,
			Selected:              stats.Selected,
			ModulesRequested:      stats.ModulesRequested,
			ModulesInScope:        stats.ModulesInScope,
			ModulesNoTests:        stats.ModulesNoTests,
			InferenceRulesApplied: len(suite.Inferences),
		}).
		SetUoWCount(len(testCfg.TestsByPackage)).
		SetOutputDir(testCfg.TestRunDir)

	ctx.InitSummary = summary
}

func showTestTimings(_ *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig) {
	testedModules := make(map[string]bool)
	for pkgPath := range testCfg.TestsByPackage {
		moduleMoniker := testCfg.ModuleMapper.GetModuleForPackagePath(pkgPath)
		if moduleMoniker != "" {
			testedModules[moduleMoniker] = true
		}
	}

	moduleList := make([]string, 0, len(testedModules))
	for module := range testedModules {
		moduleList = append(moduleList, module)
	}

	wallClockSeconds := time.Since(testCfg.TestStartTime).Seconds()
	show.ShowTestTimingsForModules(moduleList, 5, testCfg.TestRunDir, wallClockSeconds)
}

// extractUniqueModulesFromPaths extracts unique module monikers from package paths.
// Used to convert test package paths to module monikers for ExecutionPlan.ExecutionOrder.
func extractUniqueModulesFromPaths(paths []string, mapper *ModuleMapper) []string {
	seen := make(map[string]bool)
	var modules []string
	for _, path := range paths {
		moniker := mapper.GetModuleForPackagePath(path)
		if moniker != "" && !seen[moniker] {
			seen[moniker] = true
			modules = append(modules, moniker)
		}
	}
	sort.Strings(modules)
	return modules
}

// removeExistingModules returns items from 'from' that are not in 'existing'.
// Used to deduplicate modules between parallel and sequential layers.
func removeExistingModules(from, existing []string) []string {
	set := make(map[string]bool)
	for _, s := range existing {
		set[s] = true
	}
	var result []string
	for _, s := range from {
		if !set[s] {
			result = append(result, s)
		}
	}
	return result
}

// computeTestInputHash computes the input hash for a module's tests.
func computeTestInputHash(ctx *cmdframework.ExecutionContext, contract interface{ GetGlobPatterns() []string }) (string, error) {
	patterns := contract.GetGlobPatterns()
	files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
	if err != nil {
		return "", err
	}
	return hash.Files(ctx.WorkspaceRoot, files)
}

// writeUoWTestManifest writes a UoW manifest for a completed test.
func writeUoWTestManifest(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig, unitID workunit.UnitID, inputHash string, startTime time.Time, exitCode int) {
	// Initialize tracker if needed (thread-safe)
	testCfg.ExecCtx.mu.Lock()
	if testCfg.Tracker == nil {
		testCfg.Tracker = coreoutput.NewTracker(ctx.WorkspaceRoot, workunit.ContextTest)
	}
	tracker := testCfg.Tracker
	testCfg.ExecCtx.mu.Unlock()

	// Create and record the manifest
	// Include Extra field for testname which ensures unique directory names
	manifest := &coreoutput.UoWManifest{
		Context:    workunit.ContextTest,
		Module:     unitID.Module,
		Component:  unitID.Component,
		Tool:       unitID.Tool,
		Extra:      unitID.Extra, // Include testname for unique directory path
		InputHash:  inputHash,
		ExecutedAt: startTime,
		ExitCode:   exitCode,
		Duration:   time.Since(startTime),
	}

	if err := tracker.RecordComplete(unitID, manifest); err != nil {
		log.Debugf("[TEST-UOW-CACHE] Failed to write UoW manifest for %s: %v", unitID.Longname(), err)
	} else {
		log.Debugf("[TEST-UOW-CACHE] Wrote UoW manifest for %s (exitCode=%d)", unitID.Longname(), exitCode)
	}
}
