// Package test provides the test command implementation using cmdframework.
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
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/artifacts"
	"github.com/ready-to-release/eac/go/cli/eac/impl/show"
	testresults "github.com/ready-to-release/eac/go/cli/eac/impl/test/internal/results"
	"github.com/ready-to-release/eac/go/clibase/caching"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
	moduledeps "github.com/ready-to-release/eac/go/core/module-deps"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func init() {
	// Register test component-level execution support
	cmdframework.RegisterUnitProvider(core.ActionTest, ResolveTestUnitSpecs)
	cmdframework.RegisterUnitWorker(core.ActionTest, testUnitWorker)
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
	SuiteTagFilter  core.TagFilter
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

	// UoWTags stores pre-computed tag summaries keyed by UoW longname.
	// Populated by ResolveTestUnitSpecs, consumed by writeUoWTestManifest.
	UoWTags map[string]workunit.TagSummary

	// Pre-computed module input hashes for cache consistency.
	// Computed once before test execution, used by both detection and workers.
	// Prevents hash divergence when go generate modifies source files.
	ModuleInputHashes map[string]string

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
	// Store test config in typed field for access in hooks/worker
	cmdCfg.TestCmdConfig = testCfg
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
		log.Infof("Aggregated cucumber report: %s", path)
	}
	if path := testresults.AggregateCTRFReports(testCfg.TestRunDir); path != "" {
		log.Infof("Aggregated CTRF report: %s", path)
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

// extractUniqueModulesFromPaths extracts unique module monikers from package paths.
// Used to convert test package paths to module monikers for ScopeMonikers.
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

// preComputeModuleHashes computes input hashes for all modules in scope.
// Called once before test execution to ensure detection and workers use identical hashes.
func preComputeModuleHashes(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig) map[string]string {
	if ctx.ModuleRegistry == nil || testCfg.TestsByPackage == nil {
		return nil
	}

	moduleHashes := make(map[string]string)
	seen := make(map[string]bool)

	for pkgPath := range testCfg.TestsByPackage {
		moniker := testCfg.ModuleMapper.GetModuleForPackagePath(pkgPath)
		if moniker == "" || seen[moniker] {
			continue
		}
		seen[moniker] = true

		contract, ok := ctx.ModuleRegistry.Get(moniker)
		if !ok {
			continue
		}

		h, err := computeTestInputHash(ctx, contract)
		if err != nil {
			log.Debugf("Failed to pre-compute hash for %s: %v", moniker, err)
			continue
		}
		moduleHashes[moniker] = h
	}

	return moduleHashes
}

// getSuitesIncluded parses composite suite syntax.
// Handles "unit+integration" -> ["unit", "integration"]. For single suites, returns nil.
func getSuitesIncluded(suiteMoniker string) []string {
	if strings.Contains(suiteMoniker, "+") {
		return strings.Split(suiteMoniker, "+")
	}
	return nil
}

