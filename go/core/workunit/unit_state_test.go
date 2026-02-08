package workunit

import (
	"encoding/json"
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// UnitState Type Tests
// =============================================================================

func TestUnitState_FieldsExist(t *testing.T) {
	now := time.Now()
	state := UnitState{
		ID:             UnitID{Action: ActionBuild, Module: "mod", Component: "comp", Tool: "tool"},
		SourceHash:     "abc123",
		BuildID:        "build-456",
		DependencyHash: "dep789",
		Passed:         true,
		ExecutedAt:     now,
	}

	assert.Equal(t, "mod", state.ID.Module)
	assert.Equal(t, "abc123", state.SourceHash)
	assert.Equal(t, "build-456", state.BuildID)
	assert.Equal(t, "dep789", state.DependencyHash)
	assert.True(t, state.Passed)
	assert.Equal(t, now, state.ExecutedAt)
}

func TestUnitState_FieldsAreAccessible(t *testing.T) {
	tests := []struct {
		name  string
		state UnitState
	}{
		{
			name: "build state with all fields",
			state: UnitState{
				ID:         UnitID{Action: ActionBuild, Module: "core", Component: "go", Tool: "go"},
				SourceHash: "sha256:abc123def456",
				Passed:     true,
				ExecutedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "test state with build dependency",
			state: UnitState{
				ID:         UnitID{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest", Extra: map[string]string{"testset": "unit"}},
				SourceHash: "sha256:source123",
				BuildID:    "build-20240115-001",
				Passed:     true,
				ExecutedAt: time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC),
			},
		},
		{
			name: "integration test state with dependency hash",
			state: UnitState{
				ID:             UnitID{Action: ActionTest, Module: "eac-cli", Component: "gherkin", Tool: "godog", Extra: map[string]string{"testset": "integration"}},
				SourceHash:     "sha256:source456",
				BuildID:        "build-20240115-002",
				DependencyHash: "sha256:deps789",
				Passed:         false,
				ExecutedAt:     time.Date(2024, 1, 15, 10, 40, 0, 0, time.UTC),
			},
		},
		{
			name: "failed scan state",
			state: UnitState{
				ID:         UnitID{Action: ActionScan, Module: "web-app", Component: "docker", Tool: "trivy-vuln"},
				SourceHash: "sha256:scan123",
				BuildID:    "build-20240115-003",
				Passed:     false,
				ExecutedAt: time.Date(2024, 1, 15, 10, 45, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simply verify fields are accessible without panicking
			_ = tt.state.ID
			_ = tt.state.SourceHash
			_ = tt.state.BuildID
			_ = tt.state.DependencyHash
			_ = tt.state.Passed
			_ = tt.state.ExecutedAt
		})
	}
}

// =============================================================================
// UnitState JSON Marshaling Tests
// =============================================================================

func TestUnitState_MarshalJSON(t *testing.T) {
	state := UnitState{
		ID:         UnitID{Action: ActionBuild, Module: "core", Component: "go", Tool: "go"},
		SourceHash: "sha256:abc123",
		Passed:     true,
		ExecutedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(state)
	require.NoError(t, err)

	// Verify JSON contains expected fields
	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"source_hash"`)
	assert.Contains(t, jsonStr, `"passed"`)
	assert.Contains(t, jsonStr, `"executed_at"`)
}

func TestUnitState_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"id": {
			"Context": "build",
			"Module": "core",
			"Component": "go",
			"Tool": "go"
		},
		"source_hash": "sha256:abc123",
		"passed": true,
		"executed_at": "2024-01-15T10:30:00Z"
	}`

	var state UnitState
	err := json.Unmarshal([]byte(jsonData), &state)
	require.NoError(t, err)

	assert.Equal(t, "sha256:abc123", state.SourceHash)
	assert.True(t, state.Passed)
}

func TestUnitState_MarshalUnmarshalRoundTrip(t *testing.T) {
	original := UnitState{
		ID:             UnitID{Action: ActionTest, Module: "core", Component: "go", Tool: "gotest", Extra: map[string]string{"testset": "unit"}},
		SourceHash:     "sha256:source123",
		BuildID:        "build-001",
		DependencyHash: "sha256:deps456",
		Passed:         true,
		ExecutedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var restored UnitState
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Compare fields
	assert.Equal(t, original.SourceHash, restored.SourceHash)
	assert.Equal(t, original.BuildID, restored.BuildID)
	assert.Equal(t, original.DependencyHash, restored.DependencyHash)
	assert.Equal(t, original.Passed, restored.Passed)
	assert.True(t, original.ExecutedAt.Equal(restored.ExecutedAt))
}

func TestUnitState_JSONOmitsEmptyBuildID(t *testing.T) {
	state := UnitState{
		ID:         UnitID{Action: ActionBuild, Module: "core", Component: "go", Tool: "go"},
		SourceHash: "sha256:abc123",
		BuildID:    "", // Empty - should be omitted
		Passed:     true,
		ExecutedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(state)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.NotContains(t, jsonStr, `"build_id"`, "Empty BuildID should be omitted")
}

func TestUnitState_JSONOmitsEmptyDependencyHash(t *testing.T) {
	state := UnitState{
		ID:             UnitID{Action: ActionBuild, Module: "core", Component: "go", Tool: "go"},
		SourceHash:     "sha256:abc123",
		DependencyHash: "", // Empty - should be omitted
		Passed:         true,
		ExecutedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(state)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.NotContains(t, jsonStr, `"dependency_hash"`, "Empty DependencyHash should be omitted")
}

// =============================================================================
// InvalidationRule Type Tests
// =============================================================================

func TestInvalidationRule_FieldsExist(t *testing.T) {
	rule := InvalidationRule{
		OnSourceChange:     true,
		OnBuildChange:      true,
		OnDependencyChange: true,
		OnFailure:          true,
	}

	assert.True(t, rule.OnSourceChange)
	assert.True(t, rule.OnBuildChange)
	assert.True(t, rule.OnDependencyChange)
	assert.True(t, rule.OnFailure)
}

func TestInvalidationRule_DefaultIsAllFalse(t *testing.T) {
	var rule InvalidationRule

	assert.False(t, rule.OnSourceChange)
	assert.False(t, rule.OnBuildChange)
	assert.False(t, rule.OnDependencyChange)
	assert.False(t, rule.OnFailure)
}

// =============================================================================
// DefaultRules Tests
// =============================================================================

func TestDefaultRules_Exists(t *testing.T) {
	// DefaultRules should be a map from Context to InvalidationRule
	require.NotNil(t, DefaultRules)
}

func TestDefaultRules_HasAllContexts(t *testing.T) {
	contexts := []core.ActionType{core.ActionBuild, core.ActionTest, core.ActionLint, core.ActionScan}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			_, exists := DefaultRules[ctx]
			assert.True(t, exists, "DefaultRules should have entry for %s", ctx)
		})
	}
}

func TestDefaultRules_BuildContext(t *testing.T) {
	rule, exists := DefaultRules[core.ActionBuild]
	require.True(t, exists)

	assert.True(t, rule.OnSourceChange, "Build should invalidate on source change")
	assert.False(t, rule.OnBuildChange, "Build should not depend on build change")
	assert.False(t, rule.OnDependencyChange, "Build should not depend on dependency change by default")
	assert.True(t, rule.OnFailure, "Build should invalidate on previous failure")
}

func TestDefaultRules_TestContext(t *testing.T) {
	rule, exists := DefaultRules[core.ActionTest]
	require.True(t, exists)

	assert.True(t, rule.OnSourceChange, "Test should invalidate on source change")
	assert.True(t, rule.OnBuildChange, "Test should invalidate on build change")
	assert.False(t, rule.OnDependencyChange, "Test should not invalidate on dependency change by default (use IntegrationTestRule)")
	assert.True(t, rule.OnFailure, "Test should invalidate on previous failure")
}

func TestDefaultRules_LintContext(t *testing.T) {
	rule, exists := DefaultRules[core.ActionLint]
	require.True(t, exists)

	assert.True(t, rule.OnSourceChange, "Lint should invalidate on source change")
	assert.False(t, rule.OnBuildChange, "Lint should not depend on build change")
	assert.False(t, rule.OnDependencyChange, "Lint should not depend on dependency change")
	assert.True(t, rule.OnFailure, "Lint should invalidate on previous failure")
}

func TestDefaultRules_ScanContext(t *testing.T) {
	rule, exists := DefaultRules[core.ActionScan]
	require.True(t, exists)

	assert.True(t, rule.OnSourceChange, "Scan should invalidate on source change")
	assert.True(t, rule.OnBuildChange, "Scan should invalidate on build change")
	assert.False(t, rule.OnDependencyChange, "Scan should not depend on dependency change by default")
	assert.True(t, rule.OnFailure, "Scan should invalidate on previous failure")
}

// =============================================================================
// IntegrationTestRule Tests
// =============================================================================

func TestIntegrationTestRule_Exists(t *testing.T) {
	// IntegrationTestRule should be defined
	_ = IntegrationTestRule
}

func TestIntegrationTestRule_HasOnDependencyChangeTrue(t *testing.T) {
	assert.True(t, IntegrationTestRule.OnDependencyChange,
		"IntegrationTestRule must have OnDependencyChange=true")
}

func TestIntegrationTestRule_HasAllFlags(t *testing.T) {
	assert.True(t, IntegrationTestRule.OnSourceChange, "Should invalidate on source change")
	assert.True(t, IntegrationTestRule.OnBuildChange, "Should invalidate on build change")
	assert.True(t, IntegrationTestRule.OnDependencyChange, "Should invalidate on dependency change")
	assert.True(t, IntegrationTestRule.OnFailure, "Should invalidate on previous failure")
}

func TestIntegrationTestRule_DiffersFromDefaultTestRule(t *testing.T) {
	defaultRule := DefaultRules[core.ActionTest]

	// The key difference is OnDependencyChange
	assert.False(t, defaultRule.OnDependencyChange,
		"Default test rule should NOT have OnDependencyChange")
	assert.True(t, IntegrationTestRule.OnDependencyChange,
		"Integration test rule SHOULD have OnDependencyChange")
}

// =============================================================================
// InvalidationRule Usage Scenarios
// =============================================================================

func TestInvalidationRule_UsageScenarios(t *testing.T) {
	tests := []struct {
		name             string
		rule             InvalidationRule
		sourceChanged    bool
		buildChanged     bool
		depChanged       bool
		prevFailed       bool
		shouldInvalidate bool
	}{
		{
			name:             "build with source change",
			rule:             DefaultRules[core.ActionBuild],
			sourceChanged:    true,
			shouldInvalidate: true,
		},
		{
			name:             "build with no changes and previous success",
			rule:             DefaultRules[core.ActionBuild],
			sourceChanged:    false,
			prevFailed:       false,
			shouldInvalidate: false,
		},
		{
			name:             "test with build change",
			rule:             DefaultRules[core.ActionTest],
			buildChanged:     true,
			shouldInvalidate: true,
		},
		{
			name:             "integration test with dependency change",
			rule:             IntegrationTestRule,
			depChanged:       true,
			shouldInvalidate: true,
		},
		{
			name:             "default test ignores dependency change",
			rule:             DefaultRules[core.ActionTest],
			depChanged:       true,
			shouldInvalidate: false,
		},
		{
			name:             "any rule invalidates on previous failure",
			rule:             DefaultRules[core.ActionLint],
			prevFailed:       true,
			shouldInvalidate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate expected invalidation based on rule
			shouldInvalidate := false
			if tt.rule.OnSourceChange && tt.sourceChanged {
				shouldInvalidate = true
			}
			if tt.rule.OnBuildChange && tt.buildChanged {
				shouldInvalidate = true
			}
			if tt.rule.OnDependencyChange && tt.depChanged {
				shouldInvalidate = true
			}
			if tt.rule.OnFailure && tt.prevFailed {
				shouldInvalidate = true
			}

			assert.Equal(t, tt.shouldInvalidate, shouldInvalidate)
		})
	}
}

// =============================================================================
// UnitState Zero Value Tests
// =============================================================================

func TestUnitState_ZeroValueIsValid(t *testing.T) {
	var state UnitState

	// Zero value should be usable
	assert.Empty(t, state.SourceHash)
	assert.Empty(t, state.BuildID)
	assert.Empty(t, state.DependencyHash)
	assert.False(t, state.Passed)
	assert.True(t, state.ExecutedAt.IsZero())
}

// =============================================================================
// UnitState Time Handling Tests
// =============================================================================

func TestUnitState_ExecutedAtPreservesTimezone(t *testing.T) {
	utcTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	state := UnitState{
		ID:         UnitID{Action: ActionBuild, Module: "mod", Component: "comp", Tool: "tool"},
		SourceHash: "hash",
		Passed:     true,
		ExecutedAt: utcTime,
	}

	// Marshal and unmarshal
	data, err := json.Marshal(state)
	require.NoError(t, err)

	var restored UnitState
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Times should be equal (may be in different zones, but same instant)
	assert.True(t, state.ExecutedAt.Equal(restored.ExecutedAt))
}

// =============================================================================
// TestSet Type Tests
// =============================================================================

func TestTestSet_Constants(t *testing.T) {
	// Verify constants exist and have expected values
	assert.Equal(t, TestSet("unit"), TestSetUnit)
	assert.Equal(t, TestSet("integration"), TestSetIntegration)
}

func TestTestSet_CanBeUsedAsMapKey(t *testing.T) {
	m := map[TestSet]string{
		TestSetUnit:        "unit tests",
		TestSetIntegration: "integration tests",
	}

	assert.Equal(t, "unit tests", m[TestSetUnit])
	assert.Equal(t, "integration tests", m[TestSetIntegration])
}

// =============================================================================
// GetTestSetForLTag Tests
// =============================================================================

func TestGetTestSetForLTag_L0ReturnsUnit(t *testing.T) {
	result := GetTestSetForLTag("@L0")
	assert.Equal(t, TestSetUnit, result)
}

func TestGetTestSetForLTag_L1ReturnsUnit(t *testing.T) {
	result := GetTestSetForLTag("@L1")
	assert.Equal(t, TestSetUnit, result)
}

func TestGetTestSetForLTag_L2ReturnsIntegration(t *testing.T) {
	result := GetTestSetForLTag("@L2")
	assert.Equal(t, TestSetIntegration, result)
}

func TestGetTestSetForLTag_L3ReturnsIntegration(t *testing.T) {
	result := GetTestSetForLTag("@L3")
	assert.Equal(t, TestSetIntegration, result)
}

func TestGetTestSetForLTag_L4ReturnsIntegration(t *testing.T) {
	result := GetTestSetForLTag("@L4")
	assert.Equal(t, TestSetIntegration, result)
}

func TestGetTestSetForLTag_UnknownDefaultsToUnit(t *testing.T) {
	tests := []string{"@L5", "@L99", "L0", "invalid", "", "@other"}
	for _, tag := range tests {
		t.Run(tag, func(t *testing.T) {
			result := GetTestSetForLTag(tag)
			assert.Equal(t, TestSetUnit, result, "Unknown tag %q should default to unit", tag)
		})
	}
}

// =============================================================================
// ClassifyTestByTags Tests
// =============================================================================

func TestClassifyTestByTags_EmptyTagsReturnsUnit(t *testing.T) {
	result := ClassifyTestByTags([]string{})
	assert.Equal(t, TestSetUnit, result)
}

func TestClassifyTestByTags_NilTagsReturnsUnit(t *testing.T) {
	result := ClassifyTestByTags(nil)
	assert.Equal(t, TestSetUnit, result)
}

func TestClassifyTestByTags_OnlyUnitTagsReturnsUnit(t *testing.T) {
	result := ClassifyTestByTags([]string{"@L0", "@L1", "@fast"})
	assert.Equal(t, TestSetUnit, result)
}

func TestClassifyTestByTags_IntegrationTagReturnsIntegration(t *testing.T) {
	result := ClassifyTestByTags([]string{"@L2", "@slow"})
	assert.Equal(t, TestSetIntegration, result)
}

func TestClassifyTestByTags_MixedTagsReturnsIntegration(t *testing.T) {
	// If ANY tag is integration, the whole test is classified as integration
	result := ClassifyTestByTags([]string{"@L0", "@L2", "@other"})
	assert.Equal(t, TestSetIntegration, result)
}

func TestClassifyTestByTags_NoLTagsDefaultsToUnit(t *testing.T) {
	result := ClassifyTestByTags([]string{"@fast", "@unit", "@smoke"})
	assert.Equal(t, TestSetUnit, result)
}

// =============================================================================
// GetRuleForTestSet Tests
// =============================================================================

func TestGetRuleForTestSet_UnitReturnsDefaultTestRule(t *testing.T) {
	rule := GetRuleForTestSet(TestSetUnit)
	expected := DefaultRules[core.ActionTest]

	assert.Equal(t, expected.OnSourceChange, rule.OnSourceChange)
	assert.Equal(t, expected.OnBuildChange, rule.OnBuildChange)
	assert.Equal(t, expected.OnDependencyChange, rule.OnDependencyChange)
	assert.Equal(t, expected.OnFailure, rule.OnFailure)
}

func TestGetRuleForTestSet_IntegrationReturnsIntegrationRule(t *testing.T) {
	rule := GetRuleForTestSet(TestSetIntegration)

	assert.Equal(t, IntegrationTestRule.OnSourceChange, rule.OnSourceChange)
	assert.Equal(t, IntegrationTestRule.OnBuildChange, rule.OnBuildChange)
	assert.Equal(t, IntegrationTestRule.OnDependencyChange, rule.OnDependencyChange)
	assert.Equal(t, IntegrationTestRule.OnFailure, rule.OnFailure)
}

func TestGetRuleForTestSet_IntegrationHasOnDependencyChange(t *testing.T) {
	rule := GetRuleForTestSet(TestSetIntegration)
	assert.True(t, rule.OnDependencyChange, "Integration test rule must propagate dependency changes")
}

func TestGetRuleForTestSet_UnitDoesNotHaveOnDependencyChange(t *testing.T) {
	rule := GetRuleForTestSet(TestSetUnit)
	assert.False(t, rule.OnDependencyChange, "Unit test rule must NOT propagate dependency changes")
}

// =============================================================================
// ComputeDependencyBuildHash Tests
// =============================================================================

func TestComputeDependencyBuildHash_EmptyMapReturnsEmpty(t *testing.T) {
	result := ComputeDependencyBuildHash(map[string]string{})
	assert.Empty(t, result)
}

func TestComputeDependencyBuildHash_NilMapReturnsEmpty(t *testing.T) {
	result := ComputeDependencyBuildHash(nil)
	assert.Empty(t, result)
}

func TestComputeDependencyBuildHash_SingleDependency(t *testing.T) {
	deps := map[string]string{
		"core": "build-123",
	}
	result := ComputeDependencyBuildHash(deps)

	assert.NotEmpty(t, result)
	assert.Len(t, result, 16) // 8 bytes = 16 hex chars
}

func TestComputeDependencyBuildHash_MultipleDependencies(t *testing.T) {
	deps := map[string]string{
		"core": "build-123",
		"eac-cli":  "build-456",
		"eac-web":  "build-789",
	}
	result := ComputeDependencyBuildHash(deps)

	assert.NotEmpty(t, result)
	assert.Len(t, result, 16)
}

func TestComputeDependencyBuildHash_IsDeterministic(t *testing.T) {
	deps := map[string]string{
		"core": "build-123",
		"eac-cli":  "build-456",
	}

	result1 := ComputeDependencyBuildHash(deps)
	result2 := ComputeDependencyBuildHash(deps)

	assert.Equal(t, result1, result2, "Same input should produce same hash")
}

func TestComputeDependencyBuildHash_OrderIndependent(t *testing.T) {
	// Create maps with same content, different insertion order
	deps1 := map[string]string{
		"aaa": "build-1",
		"bbb": "build-2",
		"ccc": "build-3",
	}
	deps2 := map[string]string{
		"ccc": "build-3",
		"aaa": "build-1",
		"bbb": "build-2",
	}

	result1 := ComputeDependencyBuildHash(deps1)
	result2 := ComputeDependencyBuildHash(deps2)

	assert.Equal(t, result1, result2, "Hash should be order-independent")
}

func TestComputeDependencyBuildHash_DifferentValuesProduceDifferentHashes(t *testing.T) {
	deps1 := map[string]string{"core": "build-123"}
	deps2 := map[string]string{"core": "build-456"}

	result1 := ComputeDependencyBuildHash(deps1)
	result2 := ComputeDependencyBuildHash(deps2)

	assert.NotEqual(t, result1, result2, "Different values should produce different hashes")
}

func TestComputeDependencyBuildHash_DifferentKeysProduceDifferentHashes(t *testing.T) {
	deps1 := map[string]string{"core": "build-123"}
	deps2 := map[string]string{"eac-cli": "build-123"}

	result1 := ComputeDependencyBuildHash(deps1)
	result2 := ComputeDependencyBuildHash(deps2)

	assert.NotEqual(t, result1, result2, "Different keys should produce different hashes")
}
