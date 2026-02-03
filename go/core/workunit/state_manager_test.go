package workunit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// StateManager Creation Tests
// =============================================================================

func TestNewStateManager_CreatesWithRoot(t *testing.T) {
	tmpDir := t.TempDir()

	manager := NewStateManager(tmpDir)

	require.NotNil(t, manager, "NewStateManager should return non-nil manager")
}

func TestNewStateManager_StoresWorkspaceRoot(t *testing.T) {
	tmpDir := t.TempDir()

	manager := NewStateManager(tmpDir)

	// Access the workspaceRoot field (may need to expose via getter or test internal)
	// For now, we verify through Save/Load behavior that root is correctly used
	require.NotNil(t, manager)
}

// =============================================================================
// StateManager.Save() Tests
// =============================================================================

func TestStateManager_Save_CreatesOutputDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	state := &UnitState{
		ID: UnitID{
			Context:   ContextBuild,
			Module:    "test-module",
			Component: "go",
			Tool:      "go",
		},
		SourceHash: "sha256:abc123",
		Passed:     true,
		ExecutedAt: time.Now(),
	}

	// Save should create directory and state file
	err := manager.Save(state)
	require.NoError(t, err, "Save() should succeed")

	// Verify directory was created
	expectedDir := filepath.Join(tmpDir, state.ID.OutDir())
	info, err := os.Stat(expectedDir)
	require.NoError(t, err, "Output directory should exist")
	assert.True(t, info.IsDir(), "Output path should be a directory")
}

func TestStateManager_Save_WritesStateJSON(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	state := &UnitState{
		ID: UnitID{
			Context:   ContextTest,
			Module:    "test-module",
			Component: "go",
			Tool:      "gotest",
			Extra:     map[string]string{"testset": "unit"},
		},
		SourceHash: "sha256:source123",
		BuildID:    "build-001",
		Passed:     true,
		ExecutedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	err := manager.Save(state)
	require.NoError(t, err)

	// Verify state.json file exists
	stateFile := filepath.Join(tmpDir, state.ID.StateFile())
	_, err = os.Stat(stateFile)
	require.NoError(t, err, "state.json should exist")

	// Verify contents are valid JSON
	data, err := os.ReadFile(stateFile)
	require.NoError(t, err)

	var loadedState UnitState
	err = json.Unmarshal(data, &loadedState)
	require.NoError(t, err, "state.json should contain valid JSON")
}

func TestStateManager_Save_OverwritesExistingState(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{
		Context:   ContextLint,
		Module:    "test-module",
		Component: "go",
		Tool:      "golangci-lint",
	}

	// Save first state
	state1 := &UnitState{
		ID:         unitID,
		SourceHash: "hash-v1",
		Passed:     false,
		ExecutedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	err := manager.Save(state1)
	require.NoError(t, err)

	// Save second state (should overwrite)
	state2 := &UnitState{
		ID:         unitID,
		SourceHash: "hash-v2",
		Passed:     true,
		ExecutedAt: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
	}
	err = manager.Save(state2)
	require.NoError(t, err)

	// Load and verify it's the second state
	loaded, err := manager.Load(unitID)
	require.NoError(t, err)
	assert.Equal(t, "hash-v2", loaded.SourceHash, "Should have overwritten with new hash")
	assert.True(t, loaded.Passed, "Should have overwritten with new passed value")
}

// =============================================================================
// StateManager.Load() Tests
// =============================================================================

func TestStateManager_Load_ReadsStateFromCorrectPath(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{
		Context:   ContextScan,
		Module:    "test-module",
		Component: "go",
		Tool:      "trivy-vuln",
	}

	// Create state file manually
	stateDir := filepath.Join(tmpDir, unitID.OutDir())
	err := os.MkdirAll(stateDir, 0755)
	require.NoError(t, err)

	expectedState := UnitState{
		ID:         unitID,
		SourceHash: "sha256:manual",
		BuildID:    "build-manual",
		Passed:     true,
		ExecutedAt: time.Date(2024, 6, 20, 15, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(expectedState)
	require.NoError(t, err)

	stateFile := filepath.Join(tmpDir, unitID.StateFile())
	err = os.WriteFile(stateFile, data, 0644)
	require.NoError(t, err)

	// Load should read from correct path
	loaded, err := manager.Load(unitID)
	require.NoError(t, err, "Load() should succeed")
	assert.Equal(t, expectedState.SourceHash, loaded.SourceHash)
	assert.Equal(t, expectedState.BuildID, loaded.BuildID)
	assert.True(t, loaded.Passed)
}

func TestStateManager_Load_ReturnsErrorIfNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{
		Context:   ContextBuild,
		Module:    "nonexistent-module",
		Component: "go",
		Tool:      "go",
	}

	// Load without saving should return error
	_, err := manager.Load(unitID)
	assert.Error(t, err, "Load() should return error when state file doesn't exist")
}

func TestStateManager_Load_ReturnsErrorForInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{
		Context:   ContextTest,
		Module:    "invalid-json-module",
		Component: "go",
		Tool:      "gotest",
		Extra:     map[string]string{"testset": "unit"},
	}

	// Create invalid JSON file
	stateDir := filepath.Join(tmpDir, unitID.OutDir())
	err := os.MkdirAll(stateDir, 0755)
	require.NoError(t, err)

	stateFile := filepath.Join(tmpDir, unitID.StateFile())
	err = os.WriteFile(stateFile, []byte("not valid json {{{"), 0644)
	require.NoError(t, err)

	// Load should fail
	_, err = manager.Load(unitID)
	assert.Error(t, err, "Load() should return error for invalid JSON")
}

// =============================================================================
// StateManager Save/Load Round-Trip Tests
// =============================================================================

func TestStateManager_SaveLoadRoundTrip_PreservesAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	original := &UnitState{
		ID: UnitID{
			Context:   ContextTest,
			Module:    "roundtrip-module",
			Component: "gherkin",
			Tool:      "godog",
			Extra:     map[string]string{"testset": "integration"},
		},
		SourceHash:     "sha256:source-hash-value",
		BuildID:        "build-20240115-001",
		DependencyHash: "sha256:dep-hash-value",
		Passed:         true,
		ExecutedAt:     time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC),
	}

	// Save
	err := manager.Save(original)
	require.NoError(t, err)

	// Load
	loaded, err := manager.Load(original.ID)
	require.NoError(t, err)

	// Verify all fields preserved
	assert.Equal(t, original.ID.Context, loaded.ID.Context)
	assert.Equal(t, original.ID.Module, loaded.ID.Module)
	assert.Equal(t, original.ID.Component, loaded.ID.Component)
	assert.Equal(t, original.ID.Tool, loaded.ID.Tool)
	assert.Equal(t, original.SourceHash, loaded.SourceHash)
	assert.Equal(t, original.BuildID, loaded.BuildID)
	assert.Equal(t, original.DependencyHash, loaded.DependencyHash)
	assert.Equal(t, original.Passed, loaded.Passed)
	assert.True(t, original.ExecutedAt.Equal(loaded.ExecutedAt), "ExecutedAt should be preserved")
}

func TestStateManager_SaveLoadRoundTrip_FailedState(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	original := &UnitState{
		ID: UnitID{
			Context:   ContextLint,
			Module:    "failed-module",
			Component: "go",
			Tool:      "golangci-lint",
		},
		SourceHash: "sha256:failed-source",
		Passed:     false, // Failed state
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
	}

	err := manager.Save(original)
	require.NoError(t, err)

	loaded, err := manager.Load(original.ID)
	require.NoError(t, err)

	assert.False(t, loaded.Passed, "Failed state should be preserved")
}

func TestStateManager_SaveLoadRoundTrip_EmptyOptionalFields(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	original := &UnitState{
		ID: UnitID{
			Context:   ContextBuild,
			Module:    "minimal-module",
			Component: "go",
			Tool:      "go",
		},
		SourceHash:     "sha256:minimal",
		BuildID:        "", // Empty
		DependencyHash: "", // Empty
		Passed:         true,
		ExecutedAt:     time.Now().UTC().Truncate(time.Second),
	}

	err := manager.Save(original)
	require.NoError(t, err)

	loaded, err := manager.Load(original.ID)
	require.NoError(t, err)

	assert.Empty(t, loaded.BuildID, "Empty BuildID should be preserved as empty")
	assert.Empty(t, loaded.DependencyHash, "Empty DependencyHash should be preserved as empty")
}

// =============================================================================
// StateManager.NeedsExecution() Tests
// =============================================================================

func TestStateManager_NeedsExecution_NoStateExists(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	spec := UnitSpec{
		ID: UnitID{
			Context:   ContextBuild,
			Module:    "new-module",
			Component: "go",
			Tool:      "go",
		},
	}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	needs, reason := manager.NeedsExecution(spec, rule, "current-hash", nil)

	assert.True(t, needs, "Should need execution when no prior state exists")
	assert.Equal(t, "no prior state", reason)
}

func TestStateManager_NeedsExecution_PreviousFailure(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{
		Context:   ContextTest,
		Module:    "failed-module",
		Component: "go",
		Tool:      "gotest",
		Extra:     map[string]string{"testset": "unit"},
	}

	// Save a failed state
	failedState := &UnitState{
		ID:         unitID,
		SourceHash: "sha256:same-hash",
		Passed:     false, // Previous failure
		ExecutedAt: time.Now(),
	}
	err := manager.Save(failedState)
	require.NoError(t, err)

	spec := UnitSpec{ID: unitID}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	needs, reason := manager.NeedsExecution(spec, rule, "sha256:same-hash", nil)

	assert.True(t, needs, "Should need execution when previous run failed")
	assert.Equal(t, "previous failure", reason)
}

func TestStateManager_NeedsExecution_PreviousFailure_RuleDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{
		Context:   ContextTest,
		Module:    "failed-module",
		Component: "go",
		Tool:      "gotest",
		Extra:     map[string]string{"testset": "unit"},
	}

	// Save a failed state
	failedState := &UnitState{
		ID:         unitID,
		SourceHash: "sha256:same-hash",
		Passed:     false,
		ExecutedAt: time.Now(),
	}
	err := manager.Save(failedState)
	require.NoError(t, err)

	spec := UnitSpec{ID: unitID}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: false} // OnFailure disabled

	needs, reason := manager.NeedsExecution(spec, rule, "sha256:same-hash", nil)

	assert.False(t, needs, "Should NOT need execution when OnFailure is false")
	assert.Empty(t, reason)
}

func TestStateManager_NeedsExecution_SourceChanged(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{
		Context:   ContextBuild,
		Module:    "changed-module",
		Component: "go",
		Tool:      "go",
	}

	// Save a successful state with old hash
	oldState := &UnitState{
		ID:         unitID,
		SourceHash: "sha256:old-hash",
		Passed:     true,
		ExecutedAt: time.Now(),
	}
	err := manager.Save(oldState)
	require.NoError(t, err)

	spec := UnitSpec{ID: unitID}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	needs, reason := manager.NeedsExecution(spec, rule, "sha256:new-hash", nil) // Different hash

	assert.True(t, needs, "Should need execution when source hash changed")
	assert.Equal(t, "source changed", reason)
}

func TestStateManager_NeedsExecution_SourceChanged_RuleDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{
		Context:   ContextBuild,
		Module:    "changed-module",
		Component: "go",
		Tool:      "go",
	}

	// Save a successful state with old hash
	oldState := &UnitState{
		ID:         unitID,
		SourceHash: "sha256:old-hash",
		Passed:     true,
		ExecutedAt: time.Now(),
	}
	err := manager.Save(oldState)
	require.NoError(t, err)

	spec := UnitSpec{ID: unitID}
	rule := InvalidationRule{OnSourceChange: false, OnFailure: true} // OnSourceChange disabled

	needs, reason := manager.NeedsExecution(spec, rule, "sha256:new-hash", nil)

	assert.False(t, needs, "Should NOT need execution when OnSourceChange is false")
	assert.Empty(t, reason)
}

func TestStateManager_NeedsExecution_ValidCacheHit(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{
		Context:   ContextLint,
		Module:    "cached-module",
		Component: "go",
		Tool:      "golangci-lint",
	}

	// Save a successful state with same hash
	validState := &UnitState{
		ID:         unitID,
		SourceHash: "sha256:current-hash",
		Passed:     true,
		ExecutedAt: time.Now(),
	}
	err := manager.Save(validState)
	require.NoError(t, err)

	spec := UnitSpec{ID: unitID}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	needs, reason := manager.NeedsExecution(spec, rule, "sha256:current-hash", nil) // Same hash

	assert.False(t, needs, "Should NOT need execution when cache is valid")
	assert.Empty(t, reason, "Reason should be empty when no execution needed")
}

func TestStateManager_NeedsExecution_AllRulesDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{
		Context:   ContextScan,
		Module:    "always-cached",
		Component: "go",
		Tool:      "trivy-vuln",
	}

	// Save a failed state with different hash
	state := &UnitState{
		ID:         unitID,
		SourceHash: "sha256:old-hash",
		Passed:     false,
		ExecutedAt: time.Now(),
	}
	err := manager.Save(state)
	require.NoError(t, err)

	spec := UnitSpec{ID: unitID}
	rule := InvalidationRule{
		OnSourceChange: false,
		OnBuildChange:  false,
		OnFailure:      false,
	}

	needs, reason := manager.NeedsExecution(spec, rule, "sha256:new-hash", nil)

	assert.False(t, needs, "Should NOT need execution when all rules disabled")
	assert.Empty(t, reason)
}

func TestStateManager_NeedsExecution_CacheSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{
		Context:   ContextBuild,
		Module:    "cached-module",
		Component: "go",
		Tool:      "go",
	}

	// Save a successful state with same hash
	validState := &UnitState{
		ID:         unitID,
		SourceHash: "sha256:current-hash",
		Passed:     true,
		ExecutedAt: time.Now(),
	}
	err := manager.Save(validState)
	require.NoError(t, err)

	spec := UnitSpec{ID: unitID}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	// Create cache config that skips state
	cacheConfig := cache.NewConfig()
	cacheConfig.Skip(cache.Spec{Level: cache.LevelLocal, Type: cache.TypeState})

	needs, reason := manager.NeedsExecution(spec, rule, "sha256:current-hash", cacheConfig)

	assert.True(t, needs, "Should need execution when cache is skipped")
	assert.Equal(t, "cache skipped (--skip-cache=state)", reason)
}

func TestStateManager_DetectChanges_CacheSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	// Save state for all specs
	for _, module := range []string{"mod-a", "mod-b"} {
		state := &UnitState{
			ID:         UnitID{Context: ContextBuild, Module: module, Component: "go", Tool: "go"},
			SourceHash: "hash-" + module,
			Passed:     true,
			ExecutedAt: time.Now(),
		}
		err := manager.Save(state)
		require.NoError(t, err)
	}

	specs := []UnitSpec{
		{ID: UnitID{Context: ContextBuild, Module: "mod-a", Component: "go", Tool: "go"}},
		{ID: UnitID{Context: ContextBuild, Module: "mod-b", Component: "go", Tool: "go"}},
	}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	// Hash provider that returns matching hashes (would normally be cache hit)
	hashProvider := func(spec UnitSpec) (string, error) {
		return "hash-" + spec.ID.Module, nil
	}

	// Create cache config that skips state
	cacheConfig := cache.NewConfig()
	cacheConfig.Skip(cache.Spec{Level: cache.LevelAll, Type: cache.TypeState})

	result, err := manager.DetectChanges(specs, rule, hashProvider, cacheConfig)
	require.NoError(t, err)

	assert.Len(t, result.Changed, 2, "All specs should need execution when cache is skipped")
	assert.Empty(t, result.UpToDate, "No specs should be up-to-date when cache is skipped")
	assert.Equal(t, "cache skipped (--skip-cache=state)", result.ChangeReasons[specs[0].ID.Longname()])
}

func TestStateManager_DetectModuleChanges_CacheSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	// Save state for all modules
	for _, module := range []string{"mod-a", "mod-b"} {
		err := manager.SaveModuleResult(ContextLint, module, true, "hash-"+module)
		require.NoError(t, err)
	}

	modules := []string{"mod-a", "mod-b"}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	// Hash provider that returns matching hashes (would normally be cache hit)
	hashProvider := func(module string) (string, error) {
		return "hash-" + module, nil
	}

	// Create cache config that skips state
	cacheConfig := cache.NewConfig()
	cacheConfig.SkipAll()

	result, err := manager.DetectModuleChanges(ContextLint, modules, rule, hashProvider, cacheConfig)
	require.NoError(t, err)

	assert.Len(t, result.ChangedModules, 2, "All modules should need execution when cache is skipped")
	assert.Empty(t, result.UpToDateModules, "No modules should be up-to-date when cache is skipped")
	assert.Equal(t, "cache skipped (--skip-cache=state)", result.ChangeReasons["mod-a"])
}

// =============================================================================
// StateManager Edge Cases
// =============================================================================

func TestStateManager_WorksWithTestsets(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitState := &UnitState{
		ID: UnitID{
			Context:   ContextTest,
			Module:    "testset-module",
			Component: "go",
			Tool:      "gotest",
			Extra:     map[string]string{"testset": "integration"},
		},
		SourceHash: "sha256:testset-hash",
		Passed:     true,
		ExecutedAt: time.Now(),
	}

	// Save and load should work with testset in path
	err := manager.Save(unitState)
	require.NoError(t, err)

	loaded, err := manager.Load(unitState.ID)
	require.NoError(t, err)
	assert.Equal(t, unitState.SourceHash, loaded.SourceHash)
}

func TestStateManager_IsolatesStatesBetweenUnits(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unit1 := UnitID{
		Context:   ContextBuild,
		Module:    "module-a",
		Component: "go",
		Tool:      "go",
	}

	unit2 := UnitID{
		Context:   ContextBuild,
		Module:    "module-b",
		Component: "go",
		Tool:      "go",
	}

	state1 := &UnitState{
		ID:         unit1,
		SourceHash: "hash-for-module-a",
		Passed:     true,
		ExecutedAt: time.Now(),
	}

	state2 := &UnitState{
		ID:         unit2,
		SourceHash: "hash-for-module-b",
		Passed:     false,
		ExecutedAt: time.Now(),
	}

	// Save both
	err := manager.Save(state1)
	require.NoError(t, err)
	err = manager.Save(state2)
	require.NoError(t, err)

	// Load and verify isolation
	loaded1, err := manager.Load(unit1)
	require.NoError(t, err)
	assert.Equal(t, "hash-for-module-a", loaded1.SourceHash)
	assert.True(t, loaded1.Passed)

	loaded2, err := manager.Load(unit2)
	require.NoError(t, err)
	assert.Equal(t, "hash-for-module-b", loaded2.SourceHash)
	assert.False(t, loaded2.Passed)
}

func TestStateManager_SameModuleDifferentContexts(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	buildUnit := UnitID{
		Context:   ContextBuild,
		Module:    "shared-module",
		Component: "go",
		Tool:      "go",
	}

	testUnit := UnitID{
		Context:   ContextTest,
		Module:    "shared-module",
		Component: "go",
		Tool:      "gotest",
		Extra:     map[string]string{"testset": "unit"},
	}

	buildState := &UnitState{
		ID:         buildUnit,
		SourceHash: "build-hash",
		Passed:     true,
		ExecutedAt: time.Now(),
	}

	testState := &UnitState{
		ID:         testUnit,
		SourceHash: "test-hash",
		BuildID:    "build-001",
		Passed:     false,
		ExecutedAt: time.Now(),
	}

	// Save both
	err := manager.Save(buildState)
	require.NoError(t, err)
	err = manager.Save(testState)
	require.NoError(t, err)

	// Load and verify isolation by context
	loadedBuild, err := manager.Load(buildUnit)
	require.NoError(t, err)
	assert.Equal(t, "build-hash", loadedBuild.SourceHash)

	loadedTest, err := manager.Load(testUnit)
	require.NoError(t, err)
	assert.Equal(t, "test-hash", loadedTest.SourceHash)
	assert.Equal(t, "build-001", loadedTest.BuildID)
}

// =============================================================================
// StateManager.DetectChanges() Tests
// =============================================================================

func TestStateManager_DetectChanges_FreshRun(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	specs := []UnitSpec{
		{ID: UnitID{Context: ContextBuild, Module: "mod-a", Component: "go", Tool: "go"}},
		{ID: UnitID{Context: ContextBuild, Module: "mod-b", Component: "go", Tool: "go"}},
	}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	result, err := manager.DetectChanges(specs, rule, nil, nil)
	require.NoError(t, err)

	assert.True(t, result.FreshRun, "Should be a fresh run")
	assert.Len(t, result.Changed, 2, "All specs should need execution")
	assert.Empty(t, result.UpToDate, "No specs should be up-to-date")
}

func TestStateManager_DetectChanges_MixedState(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	// Save state for mod-a only
	state := &UnitState{
		ID:         UnitID{Context: ContextLint, Module: "mod-a", Component: "go", Tool: "golangci-lint"},
		SourceHash: "hash-a",
		Passed:     true,
		ExecutedAt: time.Now(),
	}
	err := manager.Save(state)
	require.NoError(t, err)

	specs := []UnitSpec{
		{ID: UnitID{Context: ContextLint, Module: "mod-a", Component: "go", Tool: "golangci-lint"}},
		{ID: UnitID{Context: ContextLint, Module: "mod-b", Component: "go", Tool: "golangci-lint"}},
	}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	// Provide hash provider that returns matching hash for mod-a
	hashProvider := func(spec UnitSpec) (string, error) {
		if spec.ID.Module == "mod-a" {
			return "hash-a", nil // Same hash = cache hit
		}
		return "hash-b", nil
	}

	result, err := manager.DetectChanges(specs, rule, hashProvider, nil)
	require.NoError(t, err)

	assert.False(t, result.FreshRun)
	assert.Len(t, result.Changed, 1, "Only mod-b should need execution")
	assert.Len(t, result.UpToDate, 1, "mod-a should be up-to-date")
	assert.Equal(t, "mod-a", result.UpToDate[0].ID.Module)
	assert.Equal(t, "mod-b", result.Changed[0].ID.Module)
}

func TestStateManager_DetectChanges_EmptySpecs(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	result, err := manager.DetectChanges(nil, InvalidationRule{}, nil, nil)
	require.NoError(t, err)

	assert.False(t, result.FreshRun)
	assert.Empty(t, result.Changed)
	assert.Empty(t, result.UpToDate)
}

// =============================================================================
// StateManager.SaveResult() Tests
// =============================================================================

func TestStateManager_SaveResult_Success(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	spec := UnitSpec{
		ID: UnitID{Context: ContextBuild, Module: "mod-a", Component: "go", Tool: "go"},
	}

	err := manager.SaveResult(spec, 0, "source-hash-123")
	require.NoError(t, err)

	// Verify state was saved
	loaded, err := manager.Load(spec.ID)
	require.NoError(t, err)
	assert.True(t, loaded.Passed)
	assert.Equal(t, "source-hash-123", loaded.SourceHash)
}

func TestStateManager_SaveResult_Failure(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	spec := UnitSpec{
		ID: UnitID{Context: ContextTest, Module: "mod-a", Component: "go", Tool: "gotest"},
	}

	err := manager.SaveResult(spec, 1, "source-hash-456")
	require.NoError(t, err)

	loaded, err := manager.Load(spec.ID)
	require.NoError(t, err)
	assert.False(t, loaded.Passed, "Failed execution should have Passed=false")
}

func TestStateManager_SaveResult_WithMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	spec := UnitSpec{
		ID: UnitID{Context: ContextTest, Module: "mod-a", Component: "go", Tool: "gotest"},
		Metadata: map[string]any{
			"build_id":        "build-001",
			"dependency_hash": "dep-hash-xyz",
		},
	}

	err := manager.SaveResult(spec, 0, "source-hash")
	require.NoError(t, err)

	loaded, err := manager.Load(spec.ID)
	require.NoError(t, err)
	assert.Equal(t, "build-001", loaded.BuildID)
	assert.Equal(t, "dep-hash-xyz", loaded.DependencyHash)
}

// =============================================================================
// StateManager.Exists() Tests
// =============================================================================

func TestStateManager_Exists_True(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{Context: ContextBuild, Module: "mod-a", Component: "go", Tool: "go"}
	state := &UnitState{ID: unitID, SourceHash: "hash", Passed: true, ExecutedAt: time.Now()}
	err := manager.Save(state)
	require.NoError(t, err)

	assert.True(t, manager.Exists(unitID))
}

func TestStateManager_Exists_False(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{Context: ContextBuild, Module: "nonexistent", Component: "go", Tool: "go"}
	assert.False(t, manager.Exists(unitID))
}

// =============================================================================
// StateManager.Delete() Tests
// =============================================================================

func TestStateManager_Delete_ExistingState(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{Context: ContextLint, Module: "mod-a", Component: "go", Tool: "golangci-lint"}
	state := &UnitState{ID: unitID, SourceHash: "hash", Passed: true, ExecutedAt: time.Now()}
	err := manager.Save(state)
	require.NoError(t, err)

	assert.True(t, manager.Exists(unitID))

	err = manager.Delete(unitID)
	require.NoError(t, err)

	assert.False(t, manager.Exists(unitID))
}

func TestStateManager_Delete_NonexistentState(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	unitID := UnitID{Context: ContextScan, Module: "nonexistent", Component: "go", Tool: "trivy"}

	// Should not error when deleting non-existent state
	err := manager.Delete(unitID)
	assert.NoError(t, err)
}

// =============================================================================
// StateManager.ClearContext() Tests
// =============================================================================

func TestStateManager_ClearContext_RemovesAllStatesForContext(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	// Save states for multiple contexts
	buildState := &UnitState{
		ID:         UnitID{Context: ContextBuild, Module: "mod-a", Component: "go", Tool: "go"},
		SourceHash: "build-hash",
		Passed:     true,
		ExecutedAt: time.Now(),
	}
	testState := &UnitState{
		ID:         UnitID{Context: ContextTest, Module: "mod-a", Component: "go", Tool: "gotest"},
		SourceHash: "test-hash",
		Passed:     true,
		ExecutedAt: time.Now(),
	}

	err := manager.Save(buildState)
	require.NoError(t, err)
	err = manager.Save(testState)
	require.NoError(t, err)

	// Clear build context
	err = manager.ClearContext(ContextBuild)
	require.NoError(t, err)

	// Build state should be gone
	assert.False(t, manager.Exists(buildState.ID))
	// Test state should still exist
	assert.True(t, manager.Exists(testState.ID))
}

// =============================================================================
// Module-Level Helper Tests
// =============================================================================

func TestStateManager_DetectModuleChanges_FreshRun(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	modules := []string{"mod-a", "mod-b", "mod-c"}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	result, err := manager.DetectModuleChanges(ContextLint, modules, rule, nil, nil)
	require.NoError(t, err)

	assert.True(t, result.FreshRun)
	assert.Len(t, result.ChangedModules, 3)
	assert.Empty(t, result.UpToDateModules)
}

func TestStateManager_DetectModuleChanges_MixedState(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	// Save state for mod-a only
	err := manager.SaveModuleResult(ContextLint, "mod-a", true, "hash-a")
	require.NoError(t, err)

	modules := []string{"mod-a", "mod-b"}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	hashProvider := func(module string) (string, error) {
		if module == "mod-a" {
			return "hash-a", nil // Same hash = cache hit
		}
		return "hash-b", nil
	}

	result, err := manager.DetectModuleChanges(ContextLint, modules, rule, hashProvider, nil)
	require.NoError(t, err)

	assert.False(t, result.FreshRun)
	assert.Len(t, result.ChangedModules, 1)
	assert.Equal(t, "mod-b", result.ChangedModules[0])
	assert.Len(t, result.UpToDateModules, 1)
	assert.Equal(t, "mod-a", result.UpToDateModules[0])
}

func TestStateManager_DetectModuleChanges_SourceChanged(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	// Save state with old hash
	err := manager.SaveModuleResult(ContextLint, "mod-a", true, "old-hash")
	require.NoError(t, err)

	modules := []string{"mod-a"}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	hashProvider := func(module string) (string, error) {
		return "new-hash", nil // Different hash = needs execution
	}

	result, err := manager.DetectModuleChanges(ContextLint, modules, rule, hashProvider, nil)
	require.NoError(t, err)

	assert.Len(t, result.ChangedModules, 1)
	assert.Equal(t, "source changed", result.ChangeReasons["mod-a"])
}

func TestStateManager_DetectModuleChanges_PreviousFailure(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	// Save failed state
	err := manager.SaveModuleResult(ContextLint, "mod-a", false, "hash-a")
	require.NoError(t, err)

	modules := []string{"mod-a"}
	rule := InvalidationRule{OnSourceChange: true, OnFailure: true}

	hashProvider := func(module string) (string, error) {
		return "hash-a", nil // Same hash but previous failure
	}

	result, err := manager.DetectModuleChanges(ContextLint, modules, rule, hashProvider, nil)
	require.NoError(t, err)

	assert.Len(t, result.ChangedModules, 1)
	assert.Equal(t, "previous failure", result.ChangeReasons["mod-a"])
}

func TestStateManager_SaveModuleResult_Success(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	err := manager.SaveModuleResult(ContextLint, "mod-a", true, "source-hash")
	require.NoError(t, err)

	// Verify state was saved with representative unit ID
	unitID := UnitID{Context: ContextLint, Module: "mod-a", Component: "_module", Tool: "_"}
	loaded, err := manager.Load(unitID)
	require.NoError(t, err)
	assert.True(t, loaded.Passed)
	assert.Equal(t, "source-hash", loaded.SourceHash)
}

func TestStateManager_SaveModuleResult_Failure(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewStateManager(tmpDir)

	err := manager.SaveModuleResult(ContextLint, "mod-a", false, "source-hash")
	require.NoError(t, err)

	unitID := UnitID{Context: ContextLint, Module: "mod-a", Component: "_module", Tool: "_"}
	loaded, err := manager.Load(unitID)
	require.NoError(t, err)
	assert.False(t, loaded.Passed)
}
