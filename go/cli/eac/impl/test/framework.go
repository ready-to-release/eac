// Package test provides the test command implementation using cmdframework.
package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	implinternal "github.com/ready-to-release/eac/go/cli/eac/impl/internal"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/artifacts"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/manifests"
	"github.com/ready-to-release/eac/go/cli/eac/impl/show"
	testresults "github.com/ready-to-release/eac/go/cli/eac/impl/test/internal/results"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
	moduledeps "github.com/ready-to-release/eac/go/core/module-deps"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func init() {
	// Register test component-level execution support
	cmdframework.RegisterUnitProvider(cmdframework.CommandTypeTest, FlattenModulesToTestUnits)
	cmdframework.RegisterUnitWorker(cmdframework.CommandTypeTest, testUnitWorker)
	cmdframework.RegisterUnitLayersProvider(cmdframework.CommandTypeTest, getTestUnitLayers)
}

// getTestUnitLayers returns the component execution layers as string slices.
// For spec tests (godog, tscucumber), uses "module:spec:specname" format for TUI tree matching.
// This format matches the Longname() output from UnitID when Spec is set.
func getTestUnitLayers(ctx *cmdframework.ExecutionContext) [][]string {
	layers := FlattenModulesToTestUnits(ctx)
	result := make([][]string, len(layers))
	for i, layer := range layers {
		result[i] = make([]string, len(layer))
		for j, work := range layer {
			if work.ID.Spec != "" {
				// Spec test: use Longname() which returns "module:spec:specname"
				result[i][j] = work.ID.Longname()
			} else if work.ID.Tool != "" {
				// Regular test with tool: "module:component:tool"
				result[i][j] = fmt.Sprintf("%s:%s:%s", work.ID.Module, work.ID.Component, work.ID.Tool)
			} else {
				// Regular test without tool: "module:component"
				// (testType is already included in component as "path:testType")
				result[i][j] = fmt.Sprintf("%s:%s", work.ID.Module, work.ID.Component)
			}
		}
	}
	return result
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

	// Incremental state
	CachedModules map[string]bool // Modules that are up-to-date (cache hits)

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

	// Incremental test detection (devbox only)
	if !environments.IsCI() && !testCfg.ForceRetest && !ctx.Config.DryRun {
		testsByPackage = filterIncrementalTests(ctx, testCfg, testsByPackage)
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
	// ExecutionPlan.Layers must contain monikers (not paths) so buildTreeFromUnitSpecs
	// can match them with UnitSpec.ID.Module for proper tab pre-filling.
	parallelModules := extractUniqueModulesFromPaths(parallelPaths, testCfg.ModuleMapper)
	sequentialModules := extractUniqueModulesFromPaths(sequentialPaths, testCfg.ModuleMapper)

	// Remove modules that appear in both layers (keep in parallel only)
	sequentialModules = removeExistingModules(sequentialModules, parallelModules)

	// Build layers (only include non-empty layers)
	var layers [][]string
	if len(parallelModules) > 0 {
		layers = append(layers, parallelModules)
	}
	if len(sequentialModules) > 0 {
		layers = append(layers, sequentialModules)
	}

	// All modules for execution order
	var allModules []string
	allModules = append(allModules, parallelModules...)
	allModules = append(allModules, sequentialModules...)

	// Set up execution plan with module monikers
	ctx.ExecutionPlan = &repository.ExecutionPlan{
		ExecutionOrder: allModules,
		Layers:         layers,
	}
	ctx.ModuleTypes = moduleTypes
	ctx.Orchestrator.SetModuleTypes(moduleTypes)

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
	if len(testCfg.CachedModules) > 0 && ctx.Orchestrator != nil {
		ctx.Orchestrator.SetCacheDetection(testCacheVerifierFunc(testCfg.CachedModules), testCfg.CachedModules)
	}

	return nil
}

// testCacheVerifierFunc returns a CacheVerifier for test incremental caching.
// Test caching is simpler than build - just checks if module is in cached set.
// No artifact integrity verification needed since test state is tracked separately.
func testCacheVerifierFunc(cachedModules map[string]bool) orchestrator.CacheVerifier {
	return func(workspaceRoot string, spec workunit.UnitSpec, _ map[string]bool, cacheTimes map[string]time.Time) (bool, time.Time) {
		module := spec.ID.Module
		if cachedModules[module] {
			return true, cacheTimes[module]
		}
		return false, time.Time{}
	}
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

	// Update incremental test state (devbox only)
	if !environments.IsCI() && ctx.ModuleRegistry != nil {
		updateTestState(ctx, testCfg, results)
	}

	// Show timing analysis if requested
	if ctx.Config.ShowTimings {
		showTestTimings(ctx, testCfg)
	}

	return nil
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
// The component parameter is in "componentName:toolName" format (e.g., "config:gotest" or "docs-drawio-cache:godog").
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

	// Parse component:tool format (e.g., "config:gotest")
	componentName := component
	testType := ""
	if idx := strings.LastIndex(component, ":"); idx > 0 {
		componentName = component[:idx]
		testType = component[idx+1:]
	}

	// Look up pkgPath from component mapping
	pkgPath, ok := testCfg.ComponentToPkgPath[componentName]
	if !ok {
		fmt.Fprintf(logWriter, "Error: no pkgPath mapping for component %s\n", componentName)
		return 1
	}

	// Check incremental cache - if module is cached, skip immediately (blue in TUI)
	log.Debugf("[TUI-CACHE] Test worker for %s: CachedModules=%v, isCached=%v",
		module, testCfg.CachedModules != nil, testCfg.CachedModules != nil && testCfg.CachedModules[module])
	if testCfg.CachedModules != nil && testCfg.CachedModules[module] {
		fmt.Fprintf(logWriter, "⏭️  Cached (unchanged)\n")
		return -1 // -1 = skipped/cached = blue in TUI
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

	// Run tests - UoW manages log file, we just write to logWriter
	// Use the componentName as result key for aggregation
	resultKey := componentName
	result := testCfg.ExecCtx.runPackageTestsDirect(pkgPath, tests, logWriter)

	testCfg.ExecCtx.mu.Lock()
	testCfg.ExecCtx.results[resultKey] = result
	testCfg.ExecCtx.mu.Unlock()

	// Pass test counts to orchestrator for summary display
	ctx.Orchestrator.SetUnitExtras(module, componentName, orchestrator.UnitExtras{
		TestsTotal:   result.TestsTotal,
		TestsPassed:  result.TestsPassed,
		TestsFailed:  result.TestsFailed,
		TestsSkipped: result.TestsSkipped,
	})

	// Atomically update test state for this module immediately after completion
	// This ensures interrupted runs still preserve cache for completed modules
	passed := !result.PackageFailed && result.TestsFailed == 0
	if err := updateModuleTestStateAtomic(ctx, testCfg, module, passed); err != nil {
		log.Debugf("Failed to update test state for %s: %v", module, err)
	}

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

	artifactValidation := artifacts.ValidateBuildArtifacts(
		stats.ModulesInScope,
		ctx.EACConfig,
		ctx.WorkspaceRoot,
		ctx.ModuleRegistry,
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

func filterIncrementalTests(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig, testsByPackage map[string][]testing.TestReference) map[string][]testing.TestReference {
	startTime := time.Now()
	defer func() {
		ctx.SetChangeDetectionTiming(time.Since(startTime))
	}()

	// Skip incremental if no module registry available
	if ctx.ModuleRegistry == nil {
		log.Debugf("Module registry not available, skipping incremental test detection")
		return testsByPackage
	}

	moduleTestInfo, depBuildIDLoader := buildModuleTestInfo(testsByPackage, ctx.ModuleRegistry, ctx.EACConfig, ctx.WorkspaceRoot)

	// Determine test set from suite L-tags
	testSet := workunit.TestSetUnit // Default to unit
	var ltags []string
	if ctx.EACConfig != nil && ctx.EACConfig.Testing != nil {
		ltags = ctx.EACConfig.Testing.GetSuiteLTags(testCfg.SuiteName)
		testSet = workunit.ClassifyTestByTags(ltags)
	}
	log.Debugf("[TUI-CACHE] Incremental detection: suite=%s ltags=%v testSet=%s moduleCount=%d",
		testCfg.SuiteName, ltags, testSet, len(moduleTestInfo))

	// Create hash provider for source files
	hashProvider := func(module string) (string, error) {
		info, ok := moduleTestInfo[module]
		if !ok {
			return "", fmt.Errorf("no info for module %s", module)
		}
		return hash.Files(ctx.WorkspaceRoot, info.SourceFiles)
	}

	stateMgr := workunit.NewStateManager(ctx.WorkspaceRoot)
	changeResult, err := stateMgr.DetectTestModuleChanges(moduleTestInfo, testSet, hashProvider, depBuildIDLoader)
	if err != nil {
		log.Debugf("Failed to detect test changes: %v", err)
		return testsByPackage
	}

	if changeResult.FreshRun {
		log.Debugf("[TUI-CACHE] Test incremental: FreshRun=true, no cache available")
		return testsByPackage
	}

	// Build set of modules that need testing
	changedSet := make(map[string]bool)
	for _, m := range changeResult.ModulesNeedingTest {
		changedSet[m] = true
		if reason, ok := changeResult.ChangeReasons[m]; ok {
			log.Debugf("[TUI-CACHE] Module needs test: %s reason=%s", m, reason)
		}
	}
	log.Debugf("[TUI-CACHE] Detection result: needsTest=%d upToDate=%d",
		len(changeResult.ModulesNeedingTest), len(changeResult.UpToDateModules))

	// Build set of cached modules (modules that are up-to-date)
	// These will be skipped at the component worker level, not filtered from the plan.
	// This keeps all packages visible in the TUI.
	testCfg.CachedModules = make(map[string]bool)
	var changedCount, cachedCount int

	for pkgPath := range testsByPackage {
		moduleMoniker := testCfg.ModuleMapper.GetModuleForPackagePath(pkgPath)
		if changedSet[moduleMoniker] {
			changedCount++
			log.Debugf("[TUI-CACHE] Test incremental: pkgPath=%s -> module=%s (CHANGED)", pkgPath, moduleMoniker)
		} else {
			testCfg.CachedModules[moduleMoniker] = true
			cachedCount++
			log.Debugf("[TUI-CACHE] Test incremental: pkgPath=%s -> module=%s (CACHED)", pkgPath, moduleMoniker)
		}
	}

	if cachedCount > 0 {
		log.Debugf("Incremental test: %d packages to test, %d cached (will show blue in TUI)",
			changedCount, cachedCount)
	}

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

// updateModuleTestStateAtomic updates test state for a single module immediately after completion.
// This ensures interrupted runs preserve cache for completed modules (atomic caching).
func updateModuleTestStateAtomic(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig, module string, passed bool) error {
	if ctx.ModuleRegistry == nil {
		log.Debugf("[ATOMIC-STATE] ModuleRegistry is nil for module=%s", module)
		return nil
	}

	contract, exists := ctx.ModuleRegistry.Get(module)
	if !exists {
		log.Debugf("[ATOMIC-STATE] Module not found in registry: module=%s", module)
		return nil
	}
	log.Debugf("[ATOMIC-STATE] Updating test state for module=%s passed=%v", module, passed)

	// Build module info for just this module
	var sourceFiles []string
	var buildID string

	// Load build manifest to get BuildID
	moduleBuildDir := ctx.EACConfig.Repository.BuildOutputPathAbs(ctx.WorkspaceRoot, module)
	if manifest, err := implinternal.LoadModuleManifest(moduleBuildDir); err == nil {
		buildID = manifest.BuildID
	}

	// Get source files from module definition
	sourcePatterns := contract.GetGlobPatterns()
	files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, sourcePatterns)
	if err == nil {
		for _, f := range files {
			if !isTestFile(f) {
				sourceFiles = append(sourceFiles, f)
			}
		}
	}

	// Determine test set from suite L-tags
	testSet := workunit.TestSetUnit
	if ctx.EACConfig != nil && ctx.EACConfig.Testing != nil {
		ltags := ctx.EACConfig.Testing.GetSuiteLTags(testCfg.SuiteName)
		testSet = workunit.ClassifyTestByTags(ltags)
	}

	// Compute source hash
	sourceHash, _ := hash.Files(ctx.WorkspaceRoot, sourceFiles)

	log.Debugf("[ATOMIC-STATE] Saving test state: module=%s testSet=%s sourceFiles=%d", module, testSet, len(sourceFiles))
	stateMgr := workunit.NewStateManager(ctx.WorkspaceRoot)
	err = stateMgr.SaveTestModuleResult(module, testSet, passed, sourceHash, buildID, "")
	if err != nil {
		log.Debugf("[ATOMIC-STATE] SaveTestModuleResult failed: module=%s err=%v", module, err)
	}
	return err
}

func updateTestState(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig, _ []PackageResult) {
	testedModuleResults := make(map[string]bool)

	// Iterate over actual results (which use path:testType format like "go/eac/core/config:gotest")
	for resultKey, result := range testCfg.ExecCtx.results {
		// Extract pkgPath from resultKey (strip test type suffix if present)
		pkgPath := resultKey
		if colonIdx := strings.LastIndex(resultKey, ":"); colonIdx > 0 {
			// Check if suffix is a valid test type
			suffix := resultKey[colonIdx+1:]
			if testing.IsValidTestType(suffix) {
				pkgPath = resultKey[:colonIdx]
			}
		}

		// Get module moniker from package path using module mapper
		moduleMoniker := extractMonikerFromPath(pkgPath, testCfg.ModuleMapper)

		if existing, ok := testedModuleResults[moduleMoniker]; ok && !existing {
			continue
		}

		passed := !result.PackageFailed && result.TestsFailed == 0
		testedModuleResults[moduleMoniker] = passed
	}

	moduleTestInfo, _ := buildModuleTestInfo(testCfg.TestsByPackage, ctx.ModuleRegistry, ctx.EACConfig, ctx.WorkspaceRoot)

	// Determine test set from suite L-tags
	testSet := workunit.TestSetUnit
	if ctx.EACConfig != nil && ctx.EACConfig.Testing != nil {
		ltags := ctx.EACConfig.Testing.GetSuiteLTags(testCfg.SuiteName)
		testSet = workunit.ClassifyTestByTags(ltags)
	}

	// Save test state for each module
	stateMgr := workunit.NewStateManager(ctx.WorkspaceRoot)
	for module, passed := range testedModuleResults {
		info, ok := moduleTestInfo[module]
		if !ok {
			continue
		}

		sourceHash, _ := hash.Files(ctx.WorkspaceRoot, info.SourceFiles)
		if err := stateMgr.SaveTestModuleResult(module, testSet, passed, sourceHash, info.BuildID, ""); err != nil {
			log.Warnf("Failed to update test state for %s: %v", module, err)
		}
	}
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
// Used to convert test package paths to module monikers for ExecutionPlan.Layers.
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
