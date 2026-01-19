// Package test provides the test command implementation using cmdframework.
package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/artifacts"
	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/manifests"
	"github.com/ready-to-release/eac/go/eac/commands/impl/show"
	testresults "github.com/ready-to-release/eac/go/eac/commands/impl/test/internal/results"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/locking"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/environments"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	moduledeps "github.com/ready-to-release/eac/go/eac/core/module-deps"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	systemdeps "github.com/ready-to-release/eac/go/eac/core/system-deps"
	"github.com/ready-to-release/eac/go/eac/core/testing"
	"github.com/ready-to-release/eac/go/eac/core/teststate"
)

// TestFrameworkConfig holds test-specific configuration for the framework.
type TestFrameworkConfig struct {
	// Input configuration
	SuiteName   string
	Coverage    bool
	ForceRetest bool
	Parallel    bool
	ListOnly    bool

	// Resolved during execution
	Suite             *testing.TestSuite
	SelectedTests     []testing.TestReference
	TestsByPackage    map[string][]testing.TestReference
	TestsByModulePath map[string][]testing.TestReference
	ModulePathToPkg   map[string]string
	ModuleMapper      *ModuleMapper
	TestRunDir        string
	SuiteTagFilter    string
	OSFilteredCount   int

	// Execution paths
	ParallelPaths   []string
	SequentialPaths []string

	// Stats for summary
	Stats *testSelectionStats

	// Execution state
	ExecCtx       *TestExecutionContext
	TestStartTime time.Time
	Lock          *flock.Flock
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
		defaultSuites := ctx.EACConfig.TestSuites.ListDefault()
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
		for _, s := range ctx.EACConfig.TestSuites.Suites {
			ctx.WriteInit("  - %s", s.Moniker)
		}
		return fmt.Errorf("suite not found: %s", testCfg.SuiteName)
	}
	testCfg.Suite = suite

	// Acquire suite lock
	repoCfg := ctx.EACConfig.Repository
	lockCfg := locking.TestConfig(testCfg.SuiteName, repoCfg.Paths.Out.Test)
	lock, err := locking.Acquire(ctx.WorkspaceRoot, lockCfg)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	testCfg.Lock = lock
	ctx.AddCleanup(func() { locking.Release(lock) })

	// Create test output directory
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
	skipReasons := ctx.EACConfig.TestingTags.GetSkipReasons()
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
			for reason := range skipReasons {
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

	// Convert to module paths and build execution plan
	testsByModulePath, modulePathToPkg := convertToModulePaths(testsByPackage, ctx.WorkspaceRoot, ctx.EACConfig)
	testCfg.TestsByModulePath = testsByModulePath
	testCfg.ModulePathToPkg = modulePathToPkg

	// Separate parallel vs sequential paths
	var parallelPaths, sequentialPaths []string
	moduleTypes := make(map[string]string)

	for modulePath, tests := range testsByModulePath {
		hasSequential := false
		for i := range tests {
			test := &tests[i]
			if test.IsSequential {
				hasSequential = true
				break
			}
		}

		if hasSequential {
			sequentialPaths = append(sequentialPaths, modulePath)
		} else {
			parallelPaths = append(parallelPaths, modulePath)
		}

		if len(tests) > 0 {
			moduleTypes[modulePath] = tests[0].Type
		}
	}

	testCfg.ParallelPaths = parallelPaths
	testCfg.SequentialPaths = sequentialPaths
	testCfg.SuiteTagFilter = buildSuiteTagFilter(suite)

	// Set up execution plan
	ctx.ExecutionPlan = &repository.ExecutionPlan{
		ExecutionOrder: append(parallelPaths, sequentialPaths...),
		Layers:         [][]string{parallelPaths, sequentialPaths},
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
		testsByPackage:  testCfg.TestsByModulePath,
		modulePathToPkg: testCfg.ModulePathToPkg,
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

	return nil
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
		testCfg.TestsByModulePath,
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

// Helper functions

func verifyTestDependencies(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig) error {
	systemDeps := testing.GetSystemDependencies(testCfg.SelectedTests)
	moduleDeps := testing.GetModuleDependencies(testCfg.SelectedTests)

	if len(systemDeps) == 0 && len(moduleDeps) == 0 {
		return nil
	}

	if ctx.Config.SkipDeps {
		return nil
	}

	// Verify only test-phase system dependencies
	sysResults := systemdeps.VerifyAllForPhase(systemDeps, "test")
	var missing []string
	for _, result := range sysResults {
		if !result.Available {
			missing = append(missing, result.Name)
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

	if !artifactValidation.AllPresent {
		ctx.WriteInit("❌ Build artifacts are missing")
		for _, moduleName := range artifactValidation.MissingFrom {
			ctx.WriteInit("  - %s", moduleName)
		}
	}

	if !artifactValidation.AllCurrent {
		ctx.WriteInit("❌ Build artifacts are stale (source changed since build)")
		for _, moduleName := range artifactValidation.StaleModules {
			reason := artifactValidation.StaleReasons[moduleName]
			ctx.WriteInit("  - %s: %s", moduleName, reason)
		}
	}

	ctx.WriteInit("")
	ctx.WriteInit("Resolution: Run 'build <module>' for each module to generate up-to-date artifacts")
	return fmt.Errorf("build artifacts invalid")
}

func filterIncrementalTests(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig, testsByPackage map[string][]testing.TestReference) map[string][]testing.TestReference {
	// Skip incremental if no module registry available
	if ctx.ModuleRegistry == nil {
		log.Debugf("Module registry not available, skipping incremental test detection")
		return testsByPackage
	}

	moduleTestInfo, uniqueModules := buildModuleTestInfo(testsByPackage, ctx.ModuleRegistry, ctx.EACConfig, ctx.WorkspaceRoot)

	changeResult, err := teststate.DetectChanges(ctx.WorkspaceRoot, moduleTestInfo, uniqueModules)
	if err != nil {
		log.Debugf("Failed to detect test changes: %v", err)
		return testsByPackage
	}

	if changeResult.FreshRun {
		return testsByPackage
	}

	changedSet := make(map[string]bool)
	for _, m := range changeResult.ModulesNeedingTest {
		changedSet[m] = true
	}

	filtered := make(map[string][]testing.TestReference)
	for pkgPath, tests := range testsByPackage {
		moduleMoniker := testCfg.ModuleMapper.GetModuleForPackagePath(pkgPath)
		if changedSet[moduleMoniker] {
			filtered[pkgPath] = tests
		}
	}

	if len(filtered) < len(testsByPackage) {
		skipped := len(testsByPackage) - len(filtered)
		ctx.WriteInit("Incremental: skipping %d unchanged package(s)", skipped)
	}

	return filtered
}

func convertToModulePaths(testsByPackage map[string][]testing.TestReference, workspaceRoot string, cfg *config.EACConfig) (map[string][]testing.TestReference, map[string]string) {
	testsByModulePath := make(map[string][]testing.TestReference)
	modulePathToPkg := make(map[string]string)

	for pkgPath, tests := range testsByPackage {
		modulePath := convertPkgPathToModulePath(pkgPath, workspaceRoot, cfg)
		testsByModulePath[modulePath] = tests
		modulePathToPkg[modulePath] = pkgPath
	}

	return testsByModulePath, modulePathToPkg
}

func convertPkgPathToModulePath(pkgPath, workspaceRoot string, cfg *config.EACConfig) string {
	// Handle BDD paths with colons: "featureName:moduleRoot:featurePath"
	colonParts := strings.SplitN(pkgPath, ":", 3)
	if len(colonParts) == 3 {
		// BDD format: featureName:moduleRoot:featurePath
		featureName := colonParts[0]
		moduleRoot := colonParts[1]
		// Find the module moniker for this root - check all package roots and test-impl package
		for i := range cfg.Repository.Modules {
			module := &cfg.Repository.Modules[i]
			// Check all package roots
			for _, entry := range module.Components {
				if entry == nil || entry.Root == "" {
					continue
				}
				pkgRootPath := filepath.ToSlash(entry.Root)
				if moduleRoot == pkgRootPath || strings.HasPrefix(moduleRoot, pkgRootPath+"/") {
					return module.Moniker + "/" + featureName
				}
			}
			// Check test-impl package if it exists
			if testImplEntry, ok := module.Components["test-impl"]; ok && testImplEntry != nil {
				testImplPath := filepath.ToSlash(testImplEntry.Root)
				if testImplPath != "" && (moduleRoot == testImplPath || strings.HasPrefix(moduleRoot, testImplPath+"/")) {
					return module.Moniker + "/" + featureName
				}
			}
		}
		// Module not found - use sanitized path
		return sanitizePathForLog(pkgPath)
	}

	// Try to extract module moniker and subpath from standard paths
	parts := strings.Split(pkgPath, "/")
	if len(parts) == 0 {
		return pkgPath
	}

	// Check if path matches any module's package root
	for i := range cfg.Repository.Modules {
		module := &cfg.Repository.Modules[i]
		for _, entry := range module.Components {
			if entry == nil || entry.Root == "" {
				continue
			}
			moduleRoot := filepath.ToSlash(entry.Root)
			if strings.HasPrefix(pkgPath, moduleRoot+"/") || pkgPath == moduleRoot {
				subPath := strings.TrimPrefix(pkgPath, moduleRoot)
				subPath = strings.TrimPrefix(subPath, "/")
				if subPath == "" {
					return module.Moniker
				}
				return module.Moniker + "/" + subPath
			}
		}
	}

	return pkgPath
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
		SetOutputDir(testCfg.TestRunDir)

	ctx.InitSummary = summary
}

func updateTestState(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig, results []PackageResult) {
	testedModuleResults := make(map[string]bool)

	for modulePath := range testCfg.TestsByModulePath {
		moduleMoniker := strings.Split(modulePath, "/")[0]
		result, exists := testCfg.ExecCtx.results[modulePath]
		if !exists {
			continue
		}

		if existing, ok := testedModuleResults[moduleMoniker]; ok && !existing {
			continue
		}

		passed := !result.PackageFailed && result.TestsFailed == 0
		testedModuleResults[moduleMoniker] = passed
	}

	moduleTestInfo, _ := buildModuleTestInfo(testCfg.TestsByPackage, ctx.ModuleRegistry, ctx.EACConfig, ctx.WorkspaceRoot)

	uniqueModules := make([]string, 0, len(testedModuleResults))
	for m := range testedModuleResults {
		uniqueModules = append(uniqueModules, m)
	}

	if err := teststate.UpdateModuleState(ctx.WorkspaceRoot, testedModuleResults, moduleTestInfo); err != nil {
		log.Warnf("Failed to update test state: %v", err)
	}
}

func showTestTimings(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig) {
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
