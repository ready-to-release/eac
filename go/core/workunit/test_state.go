package workunit

import (
	"sort"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/cache"
)

// TestModuleInfo contains file and dependency information for a module's tests.
type TestModuleInfo struct {
	// Source files (the code being tested)
	SourceFiles []string
	// Test files (the tests themselves)
	TestFiles []string
	// Direct dependencies (module monikers this module depends on)
	Dependencies []string
	// BuildID from the module's build manifest (empty if no manifest)
	BuildID string
}

// DependencyBuildIDLoader loads the current BuildID for a dependency module.
type DependencyBuildIDLoader func(moniker string) string

// TestChangeResult represents the result of test change detection.
type TestChangeResult struct {
	// Test set this result applies to
	TestSet TestSet
	// Modules that need retesting
	ModulesNeedingTest []string
	// Modules that are up-to-date
	UpToDateModules []string
	// Reason for each module needing test
	ChangeReasons map[string]string
	// Whether this is a fresh run (no prior state)
	FreshRun bool
	// Detection time
	DetectionTime time.Duration
}

// DetectTestModuleChanges performs test-set-aware change detection for modules.
// For unit tests (L0/L1): only direct module changes trigger retest.
// For integration tests (L2+): direct AND transitive dependency changes trigger retest.
// If cacheConfig is provided and state cache is skipped, all modules are marked as needing test.
func (m *StateManager) DetectTestModuleChanges(
	moduleInfo map[string]TestModuleInfo,
	testSet TestSet,
	hashProvider ModuleHashProvider,
	loadDepBuildID DependencyBuildIDLoader,
	cacheConfig *cache.Config,
) (*TestChangeResult, error) {
	start := time.Now()
	rule := GetRuleForTestSet(testSet)

	result := &TestChangeResult{
		TestSet:       testSet,
		ChangeReasons: make(map[string]string),
	}

	if len(moduleInfo) == 0 {
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// If state cache is skipped, all modules need testing
	if cacheConfig != nil && cacheConfig.ShouldSkipState() {
		markAllModulesNeedTest(result, moduleInfo, "cache skipped (--skip-cache=state)")
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check if any prior state exists; if not, this is a fresh run
	if !m.hasAnyTestState(moduleInfo, testSet) {
		result.FreshRun = true
		markAllModulesNeedTest(result, moduleInfo, "fresh run (no prior state)")
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Phase 1: Detect directly changed modules
	directlyChanged := m.detectDirectlyChangedModules(moduleInfo, testSet, rule, hashProvider, result)

	// Phase 2: Apply invalidation rules and partition modules
	m.applyTestInvalidationRules(moduleInfo, testSet, rule, directlyChanged, loadDepBuildID, result)

	sort.Strings(result.ModulesNeedingTest)
	sort.Strings(result.UpToDateModules)
	result.DetectionTime = time.Since(start)

	return result, nil
}

// markAllModulesNeedTest adds every module in moduleInfo to the result as needing test
// with the given reason string.
func markAllModulesNeedTest(result *TestChangeResult, moduleInfo map[string]TestModuleInfo, reason string) {
	for moniker := range moduleInfo {
		result.ModulesNeedingTest = append(result.ModulesNeedingTest, moniker)
		result.ChangeReasons[moniker] = reason
	}
	sort.Strings(result.ModulesNeedingTest)
}

// hasAnyTestState checks whether the state manager contains prior state for any
// module in the given map under the specified test set.
func (m *StateManager) hasAnyTestState(moduleInfo map[string]TestModuleInfo, testSet TestSet) bool {
	for moniker := range moduleInfo {
		if m.Exists(testModuleUnitID(moniker, testSet)) {
			return true
		}
	}
	return false
}

// testModuleUnitID constructs the canonical UnitID used to store per-module test state
// for a given test set.
func testModuleUnitID(moniker string, testSet TestSet) UnitID {
	return UnitID{
		Action:        core.ActionTest,
		Module:        moniker,
		ComponentType: "_module",
		ComponentName: "_module",
		Tool:          "_",
		Extra:         map[string]string{"testset": string(testSet)},
	}
}

// detectDirectlyChangedModules loads prior state for every module and checks whether
// it has changed due to source edits, build changes, or previous failures.
// It populates result.ChangeReasons for each changed module and returns the set of
// directly changed monikers.
func (m *StateManager) detectDirectlyChangedModules(
	moduleInfo map[string]TestModuleInfo,
	testSet TestSet,
	rule InvalidationRule,
	hashProvider ModuleHashProvider,
	result *TestChangeResult,
) map[string]bool {
	directlyChanged := make(map[string]bool)

	for moniker, info := range moduleInfo {
		if reason := m.classifyModuleChange(moniker, info, testSet, rule, hashProvider); reason != "" {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = reason
		}
	}

	return directlyChanged
}

// classifyModuleChange checks a single module against its prior state and returns
// a non-empty reason string if the module needs retesting, or "" if it is up-to-date.
func (m *StateManager) classifyModuleChange(
	moniker string,
	info TestModuleInfo,
	testSet TestSet,
	rule InvalidationRule,
	hashProvider ModuleHashProvider,
) string {
	state, err := m.Load(testModuleUnitID(moniker, testSet))
	if err != nil {
		return "new module (not previously tested)"
	}

	if rule.OnFailure && !state.Passed {
		return "previous test failed"
	}

	if rule.OnBuildChange && info.BuildID != "" && state.BuildID != "" && info.BuildID != state.BuildID {
		return "build changed (rebuild detected)"
	}

	if hashProvider != nil {
		currentSourceHash, err := hashProvider(moniker)
		if err != nil {
			return "hash error: " + err.Error()
		}
		if rule.OnSourceChange && currentSourceHash != state.SourceHash {
			return "source files changed"
		}
	}

	return ""
}

// applyTestInvalidationRules partitions modules into ModulesNeedingTest and
// UpToDateModules based on the invalidation rule. For unit tests only directly
// changed modules are included. For integration tests the dependency chain is
// walked transitively.
func (m *StateManager) applyTestInvalidationRules(
	moduleInfo map[string]TestModuleInfo,
	testSet TestSet,
	rule InvalidationRule,
	directlyChanged map[string]bool,
	loadDepBuildID DependencyBuildIDLoader,
	result *TestChangeResult,
) {
	if !rule.OnDependencyChange {
		// Unit tests: only directly changed modules
		partitionModulesByDirectChange(moduleInfo, directlyChanged, result)
		return
	}

	// Integration tests: propagate through dependency chain
	m.propagateDependencyChanges(moduleInfo, testSet, directlyChanged, loadDepBuildID, result)
}

// partitionModulesByDirectChange splits modules into needing-test vs up-to-date
// based solely on the directlyChanged set (no dependency propagation).
func partitionModulesByDirectChange(
	moduleInfo map[string]TestModuleInfo,
	directlyChanged map[string]bool,
	result *TestChangeResult,
) {
	for moniker := range moduleInfo {
		if directlyChanged[moniker] {
			result.ModulesNeedingTest = append(result.ModulesNeedingTest, moniker)
		} else {
			result.UpToDateModules = append(result.UpToDateModules, moniker)
		}
	}
}

// propagateDependencyChanges walks the dependency graph transitively, marking a
// module as needing test if any of its direct or transitive dependencies have
// changed or been rebuilt.
func (m *StateManager) propagateDependencyChanges(
	moduleInfo map[string]TestModuleInfo,
	testSet TestSet,
	directlyChanged map[string]bool,
	loadDepBuildID DependencyBuildIDLoader,
	result *TestChangeResult,
) {
	needsTest := make(map[string]bool)
	checked := make(map[string]bool)

	var checkNeedsTest func(moniker string) bool
	checkNeedsTest = func(moniker string) bool {
		if cached, ok := needsTest[moniker]; ok {
			return cached
		}
		if checked[moniker] {
			return false
		}
		checked[moniker] = true

		if directlyChanged[moniker] {
			needsTest[moniker] = true
			return true
		}

		info, exists := moduleInfo[moniker]
		if !exists {
			needsTest[moniker] = false
			return false
		}

		if m.hasRebuiltDependency(moniker, info, testSet, loadDepBuildID, result) {
			needsTest[moniker] = true
			return true
		}

		for _, dep := range info.Dependencies {
			if checkNeedsTest(dep) {
				needsTest[moniker] = true
				if result.ChangeReasons[moniker] == "" {
					result.ChangeReasons[moniker] = "dependency changed: " + dep
				}
				return true
			}
		}

		needsTest[moniker] = false
		return false
	}

	for moniker := range moduleInfo {
		if checkNeedsTest(moniker) {
			result.ModulesNeedingTest = append(result.ModulesNeedingTest, moniker)
		} else {
			result.UpToDateModules = append(result.UpToDateModules, moniker)
		}
	}
}

// hasRebuiltDependency checks whether any direct dependency of the given module
// has a BuildID that differs from the stored state, indicating it was rebuilt.
// If a rebuilt dependency is found, it records the reason in result.ChangeReasons.
func (m *StateManager) hasRebuiltDependency(
	moniker string,
	info TestModuleInfo,
	testSet TestSet,
	loadDepBuildID DependencyBuildIDLoader,
	result *TestChangeResult,
) bool {
	if loadDepBuildID == nil {
		return false
	}

	for _, dep := range info.Dependencies {
		currentBuildID := loadDepBuildID(dep)
		depState, _ := m.Load(testModuleUnitID(dep, testSet))
		if depState != nil && currentBuildID != "" && depState.BuildID != "" && currentBuildID != depState.BuildID {
			if result.ChangeReasons[moniker] == "" {
				result.ChangeReasons[moniker] = "dependency rebuilt: " + dep
			}
			return true
		}
	}

	return false
}

// SaveTestModuleResult saves test state for a module with test-set-specific storage.
func (m *StateManager) SaveTestModuleResult(
	module string,
	testSet TestSet,
	passed bool,
	sourceHash string,
	buildID string,
	dependencyHash string,
) error {
	unitID := UnitID{
		Action:        core.ActionTest,
		Module:        module,
		ComponentType: "_module",
		ComponentName: "_module",
		Tool:          "_",
		Extra:         map[string]string{"testset": string(testSet)},
	}

	state := &UnitState{
		ID:             unitID,
		SourceHash:     sourceHash,
		BuildID:        buildID,
		DependencyHash: dependencyHash,
		Passed:         passed,
		ExecutedAt:     time.Now(),
	}

	return m.Save(state)
}
