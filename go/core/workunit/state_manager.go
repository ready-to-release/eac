package workunit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/paths"
)

// StateManager handles persistence of unit states.
type StateManager struct {
	workspaceRoot string
}

// NewStateManager creates a new state manager rooted at the given workspace.
func NewStateManager(root string) *StateManager {
	return &StateManager{workspaceRoot: root}
}

// Load reads the state for a unit from disk.
// Returns an error if the state file doesn't exist or contains invalid JSON.
func (m *StateManager) Load(id UnitID) (*UnitState, error) {
	path := filepath.Join(m.workspaceRoot, id.StateFile())
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var state UnitState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// Save writes the state for a unit to disk.
// Creates the state cache directory if it doesn't exist.
func (m *StateManager) Save(state *UnitState) error {
	dir := filepath.Join(m.workspaceRoot, state.ID.StateCacheDir())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(m.workspaceRoot, state.ID.StateFile())
	return os.WriteFile(path, data, 0644)
}

// NeedsExecution determines if a unit needs to be executed based on its
// cached state and the provided invalidation rule.
// Returns (true, reason) if execution is needed, (false, "") otherwise.
// If cacheConfig is provided and state cache is skipped, always returns true.
func (m *StateManager) NeedsExecution(spec UnitSpec, rule InvalidationRule, currentHash string, cacheConfig *cache.Config) (bool, string) {
	// If state cache is skipped, always execute
	if cacheConfig != nil && cacheConfig.ShouldSkipState() {
		return true, "cache skipped (--skip-cache=state)"
	}

	state, err := m.Load(spec.ID)
	if err != nil {
		return true, "no prior state"
	}

	// Check failure invalidation first (if enabled)
	if rule.OnFailure && !state.Passed {
		return true, "previous failure"
	}

	// Check source change (if enabled)
	if rule.OnSourceChange && state.SourceHash != currentHash {
		return true, "source changed"
	}

	// Check build change (if enabled) - for tests that depend on build output
	if rule.OnBuildChange && spec.Metadata != nil {
		if currentBuildID, ok := spec.Metadata["build_id"].(string); ok {
			if state.BuildID != currentBuildID {
				return true, "build changed"
			}
		}
	}

	// Check dependency change (if enabled) - for integration tests
	if rule.OnDependencyChange && spec.Metadata != nil {
		if currentDepHash, ok := spec.Metadata["dependency_hash"].(string); ok {
			if state.DependencyHash != currentDepHash {
				return true, "dependency changed"
			}
		}
	}

	return false, ""
}

// ChangeResult represents the result of batch change detection.
type ChangeResult struct {
	// Units that need execution (changed or new)
	Changed []UnitSpec
	// Units that are up-to-date (cached)
	UpToDate []UnitSpec
	// Reason for each changed unit (key is Longname)
	ChangeReasons map[string]string
	// Whether this is a fresh run (no prior state for any unit)
	FreshRun bool
	// Time taken for detection
	DetectionTime time.Duration
}

// HashProvider is a function that computes the current source hash for a unit.
type HashProvider func(spec UnitSpec) (string, error)

// DetectChanges performs batch change detection for multiple units.
// Uses the provided HashProvider to compute current hashes.
// Returns which units need execution and which are up-to-date.
// If cacheConfig is provided and state cache is skipped, all units are marked as changed.
func (m *StateManager) DetectChanges(specs []UnitSpec, rule InvalidationRule, hashProvider HashProvider, cacheConfig *cache.Config) (*ChangeResult, error) {
	start := time.Now()
	result := &ChangeResult{
		ChangeReasons: make(map[string]string),
	}

	if len(specs) == 0 {
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// If state cache is skipped, all units need execution
	if cacheConfig != nil && cacheConfig.ShouldSkipState() {
		for _, spec := range specs {
			result.Changed = append(result.Changed, spec)
			result.ChangeReasons[spec.ID.Longname()] = "cache skipped (--skip-cache=state)"
		}
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check if any prior state exists
	hasAnyState := false
	for _, spec := range specs {
		if _, err := m.Load(spec.ID); err == nil {
			hasAnyState = true
			break
		}
	}

	if !hasAnyState {
		// Fresh run - all units need execution
		result.FreshRun = true
		for _, spec := range specs {
			result.Changed = append(result.Changed, spec)
			result.ChangeReasons[spec.ID.Longname()] = "fresh run (no prior state)"
		}
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check each unit
	for _, spec := range specs {
		currentHash := ""
		if hashProvider != nil {
			var err error
			currentHash, err = hashProvider(spec)
			if err != nil {
				result.Changed = append(result.Changed, spec)
				result.ChangeReasons[spec.ID.Longname()] = "hash error: " + err.Error()
				continue
			}
		}

		needsExec, reason := m.NeedsExecution(spec, rule, currentHash, cacheConfig)
		if needsExec {
			result.Changed = append(result.Changed, spec)
			result.ChangeReasons[spec.ID.Longname()] = reason
		} else {
			result.UpToDate = append(result.UpToDate, spec)
		}
	}

	// Sort for deterministic output
	sort.Slice(result.Changed, func(i, j int) bool {
		return result.Changed[i].ID.Longname() < result.Changed[j].ID.Longname()
	})
	sort.Slice(result.UpToDate, func(i, j int) bool {
		return result.UpToDate[i].ID.Longname() < result.UpToDate[j].ID.Longname()
	})

	result.DetectionTime = time.Since(start)
	return result, nil
}

// SaveResult saves the state for a completed unit execution.
// Creates a UnitState from the result and persists it.
func (m *StateManager) SaveResult(spec UnitSpec, exitCode int, sourceHash string) error {
	state := &UnitState{
		ID:         spec.ID,
		SourceHash: sourceHash,
		Passed:     exitCode == 0,
		ExecutedAt: time.Now(),
	}

	// Copy build/dependency hashes from metadata if present
	if spec.Metadata != nil {
		if buildID, ok := spec.Metadata["build_id"].(string); ok {
			state.BuildID = buildID
		}
		if depHash, ok := spec.Metadata["dependency_hash"].(string); ok {
			state.DependencyHash = depHash
		}
	}

	return m.Save(state)
}

// Exists checks if state exists for a unit without loading it.
func (m *StateManager) Exists(id UnitID) bool {
	path := filepath.Join(m.workspaceRoot, id.StateFile())
	_, err := os.Stat(path)
	return err == nil
}

// Delete removes the state file for a unit.
func (m *StateManager) Delete(id UnitID) error {
	path := filepath.Join(m.workspaceRoot, id.StateFile())
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil // Already doesn't exist
	}
	return err
}

// ClearContext removes all state files for a given action type (build, test, lint, scan).
func (m *StateManager) ClearContext(ctx core.ActionType) error {
	dir := filepath.Join(paths.IncrementalCachePath(m.workspaceRoot), string(ctx))
	return os.RemoveAll(dir)
}

// ModuleChangeResult represents change detection results at module granularity.
// This bridges module-level caching (current approach) with unit-level storage.
type ModuleChangeResult struct {
	// Modules that need execution
	ChangedModules []string
	// Modules that are up-to-date
	UpToDateModules []string
	// Reason for each changed module
	ChangeReasons map[string]string
	// Whether this is a fresh run
	FreshRun bool
	// Detection time
	DetectionTime time.Duration
}

// ModuleHashProvider computes the source hash for a module.
type ModuleHashProvider func(module string) (string, error)

// DetectModuleChanges performs change detection at module granularity.
// This is a convenience method that wraps per-unit state checks into module-level results.
// For each module, it creates a representative unit ID and checks its state.
// If cacheConfig is provided and state cache is skipped, all modules are marked as changed.
func (m *StateManager) DetectModuleChanges(ctx core.ActionType, modules []string, rule InvalidationRule, hashProvider ModuleHashProvider, cacheConfig *cache.Config) (*ModuleChangeResult, error) {
	start := time.Now()
	result := &ModuleChangeResult{
		ChangeReasons: make(map[string]string),
	}

	if len(modules) == 0 {
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// If state cache is skipped, all modules need execution
	if cacheConfig != nil && cacheConfig.ShouldSkipState() {
		result.ChangedModules = append(result.ChangedModules, modules...)
		for _, module := range modules {
			result.ChangeReasons[module] = "cache skipped (--skip-cache=state)"
		}
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check if any prior state exists
	hasAnyState := false
	for _, module := range modules {
		// Use a representative unit ID for the module (context:module:_module:_)
		unitID := UnitID{
			Action:    ctx,
			Module:    module,
			Component: "_module", // Representative component for module-level state
			Tool:      "_",
		}
		if m.Exists(unitID) {
			hasAnyState = true
			break
		}
	}

	if !hasAnyState {
		// Fresh run - all modules need execution
		result.FreshRun = true
		result.ChangedModules = append(result.ChangedModules, modules...)
		for _, module := range modules {
			result.ChangeReasons[module] = "fresh run (no prior state)"
		}
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check each module
	for _, module := range modules {
		unitID := UnitID{
			Action:    ctx,
			Module:    module,
			Component: "_module",
			Tool:      "_",
		}
		spec := UnitSpec{ID: unitID}

		currentHash := ""
		if hashProvider != nil {
			var err error
			currentHash, err = hashProvider(module)
			if err != nil {
				result.ChangedModules = append(result.ChangedModules, module)
				result.ChangeReasons[module] = "hash error: " + err.Error()
				continue
			}
		}

		needsExec, reason := m.NeedsExecution(spec, rule, currentHash, cacheConfig)
		if needsExec {
			result.ChangedModules = append(result.ChangedModules, module)
			result.ChangeReasons[module] = reason
		} else {
			result.UpToDateModules = append(result.UpToDateModules, module)
		}
	}

	// Sort for deterministic output
	sort.Strings(result.ChangedModules)
	sort.Strings(result.UpToDateModules)

	result.DetectionTime = time.Since(start)
	return result, nil
}

// SaveModuleResult saves state for a module-level execution.
// Uses a representative unit ID for the module.
func (m *StateManager) SaveModuleResult(ctx core.ActionType, module string, passed bool, sourceHash string) error {
	unitID := UnitID{
		Action:    ctx,
		Module:    module,
		Component: "_module",
		Tool:      "_",
	}

	state := &UnitState{
		ID:         unitID,
		SourceHash: sourceHash,
		Passed:     passed,
		ExecutedAt: time.Now(),
	}

	return m.Save(state)
}

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
		for moniker := range moduleInfo {
			result.ModulesNeedingTest = append(result.ModulesNeedingTest, moniker)
			result.ChangeReasons[moniker] = "cache skipped (--skip-cache=state)"
		}
		sort.Strings(result.ModulesNeedingTest)
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check if any prior state exists
	hasAnyState := false
	for moniker := range moduleInfo {
		unitID := UnitID{
			Action:    core.ActionTest,
			Module:    moniker,
			Component: "_module",
			Tool:      "_",
			Extra:     map[string]string{"testset": string(testSet)},
		}
		if m.Exists(unitID) {
			hasAnyState = true
			break
		}
	}

	if !hasAnyState {
		// Fresh run - all modules need testing
		result.FreshRun = true
		for moniker := range moduleInfo {
			result.ModulesNeedingTest = append(result.ModulesNeedingTest, moniker)
			result.ChangeReasons[moniker] = "fresh run (no prior state)"
		}
		sort.Strings(result.ModulesNeedingTest)
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Phase 1: Detect directly changed modules
	directlyChanged := make(map[string]bool)
	for moniker, info := range moduleInfo {
		unitID := UnitID{
			Action:    core.ActionTest,
			Module:    moniker,
			Component: "_module",
			Tool:      "_",
			Extra:     map[string]string{"testset": string(testSet)},
		}

		state, err := m.Load(unitID)
		if err != nil {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "new module (not previously tested)"
			continue
		}

		// Check if previous run failed
		if rule.OnFailure && !state.Passed {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "previous test failed"
			continue
		}

		// Check build change
		if rule.OnBuildChange && info.BuildID != "" && state.BuildID != "" && info.BuildID != state.BuildID {
			directlyChanged[moniker] = true
			result.ChangeReasons[moniker] = "build changed (rebuild detected)"
			continue
		}

		// Check source files changed
		if hashProvider != nil {
			currentSourceHash, err := hashProvider(moniker)
			if err != nil {
				directlyChanged[moniker] = true
				result.ChangeReasons[moniker] = "hash error: " + err.Error()
				continue
			}

			if rule.OnSourceChange && currentSourceHash != state.SourceHash {
				directlyChanged[moniker] = true
				result.ChangeReasons[moniker] = "source files changed"
				continue
			}
		}
	}

	// Phase 2: Apply invalidation rules
	if !rule.OnDependencyChange {
		// Unit tests: only directly changed modules
		for moniker := range moduleInfo {
			if directlyChanged[moniker] {
				result.ModulesNeedingTest = append(result.ModulesNeedingTest, moniker)
			} else {
				result.UpToDateModules = append(result.UpToDateModules, moniker)
			}
		}
	} else {
		// Integration tests: propagate through dependency chain
		needsTest := make(map[string]bool)
		checked := make(map[string]bool)

		var checkNeedsTest func(moniker string) bool
		checkNeedsTest = func(moniker string) bool {
			if result, ok := needsTest[moniker]; ok {
				return result
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

			// Check dependencies
			for _, dep := range info.Dependencies {
				// Check if dependency's BuildID changed
				if loadDepBuildID != nil {
					currentBuildID := loadDepBuildID(dep)
					// Load dep's state to compare BuildID
					depUnitID := UnitID{
						Action:    core.ActionTest,
						Module:    dep,
						Component: "_module",
						Tool:      "_",
						Extra:     map[string]string{"testset": string(testSet)},
					}
					depState, _ := m.Load(depUnitID)
					if depState != nil && currentBuildID != "" && depState.BuildID != "" && currentBuildID != depState.BuildID {
						needsTest[moniker] = true
						if result.ChangeReasons[moniker] == "" {
							result.ChangeReasons[moniker] = "dependency rebuilt: " + dep
						}
						return true
					}
				}

				// Recursively check dependency
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

	sort.Strings(result.ModulesNeedingTest)
	sort.Strings(result.UpToDateModules)
	result.DetectionTime = time.Since(start)

	return result, nil
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
		Action:    core.ActionTest,
		Module:    module,
		Component: "_module",
		Tool:      "_",
		Extra:     map[string]string{"testset": string(testSet)},
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
