package test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build"
	"github.com/ready-to-release/eac/go/cli/eac/impl/show"
	testresults "github.com/ready-to-release/eac/go/cli/eac/impl/test/internal/results"
	"github.com/ready-to-release/eac/go/clibase/caching"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// testAfterInit handles suite resolution and lock acquisition.
func testAfterInit(ctx *cmdframework.ExecutionContext) error {
	testCfg, ok := ctx.Config.TestCmdConfig.(*TestFrameworkConfig)
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
	testCfg, ok := ctx.Config.TestCmdConfig.(*TestFrameworkConfig)
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
		// Return early - scope will be empty
		ctx.ScopeMonikers = nil
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
		ctx.ScopeMonikers = nil
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
	// ScopeMonikers must contain monikers (not paths) so buildTreeFromUnitSpecs
	// can match them with UnitSpec.ID.Module for proper tab pre-filling.
	parallelModules := extractUniqueModulesFromPaths(parallelPaths, testCfg.ModuleMapper)
	sequentialModules := extractUniqueModulesFromPaths(sequentialPaths, testCfg.ModuleMapper)

	// Remove modules that appear in both lists (keep in parallel only)
	sequentialModules = removeExistingModules(sequentialModules, parallelModules)

	// All modules for execution order (parallel first, then sequential)
	var allModules []string
	allModules = append(allModules, parallelModules...)
	allModules = append(allModules, sequentialModules...)

	// Set scope monikers for execution
	ctx.ScopeMonikers = allModules
	ctx.ComponentTypesDisplay = moduleTypes
	ctx.Orchestrator.SetComponentTypesDisplay(moduleTypes)

	// Build init summary
	buildTestInitSummary(ctx, testCfg)

	return nil
}

// testBeforeExecute initializes the test execution context.
func testBeforeExecute(ctx *cmdframework.ExecutionContext) error {
	testCfg, ok := ctx.Config.TestCmdConfig.(*TestFrameworkConfig)
	if !ok {
		return fmt.Errorf("testConfig not found or wrong type")
	}

	if len(ctx.ScopeMonikers) == 0 {
		return nil // Nothing to execute
	}

	repoCfg := ctx.EACConfig.Repository
	testParallelism := repoCfg.EffectiveParallelism(environments.IsCI())
	if !testCfg.Parallel {
		testParallelism = 1
	}

	// Pre-compute module input hashes if detection was skipped (CI, force-retest, dry-run)
	if testCfg.ModuleInputHashes == nil {
		testCfg.ModuleInputHashes = preComputeModuleHashes(ctx, testCfg)
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

// testAfterExecute handles manifest generation and state updates.
func testAfterExecute(ctx *cmdframework.ExecutionContext) error {
	testCfg, ok := ctx.Config.TestCmdConfig.(*TestFrameworkConfig)
	if !ok {
		return fmt.Errorf("testConfig not found or wrong type")
	}

	if testCfg.ExecCtx == nil || ctx.Config.DryRun {
		return nil
	}

	// UoW manifests are written per-worker in writeUoWTestManifest via RecordComplete.

	// Aggregate reports
	if path := testresults.AggregateCucumberReports(testCfg.TestRunDir); path != "" {
		log.Debugf("Aggregated cucumber report: %s", path)
	}
	if path := testresults.AggregateCTRFReports(testCfg.TestRunDir); path != "" {
		log.Debugf("Aggregated CTRF report: %s", path)
	}

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
	// Skip incremental if no module registry available
	if ctx.ModuleRegistry == nil {
		log.Debugf("Module registry not available, skipping incremental test detection")
		return testsByPackage
	}

	specs := ResolveTestUnitSpecs(ctx)
	result := caching.DetectIncrementalChanges(ctx, core.ActionTest, specs, "TEST")
	if result == nil {
		return testsByPackage
	}

	// Always store pre-computed hashes for worker reuse
	testCfg.ModuleInputHashes = result.ModuleInputHashes

	if result.FreshRun {
		return testsByPackage
	}

	// Copy aggregated results to test config
	testCfg.CachedUoWs = result.CachedUoWs
	testCfg.UoWCacheTimes = result.UoWCacheTimes
	testCfg.CachedModules = result.CachedModules
	testCfg.CacheTimes = result.ModuleCacheTimes

	// Return ALL packages - filtering happens at component worker level
	return testsByPackage
}

func buildSuiteTagFilter(suite *testing.TestSuite) core.TagFilter {
	return suite.ToTagFilter()
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
			SuitesIncluded:        getSuitesIncluded(suite.Moniker),
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
