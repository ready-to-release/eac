package teststate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTestSetForLTag(t *testing.T) {
	tests := []struct {
		name     string
		ltag     string
		expected TestSet
	}{
		{"L0 is unit", "@L0", TestSetUnit},
		{"L1 is unit", "@L1", TestSetUnit},
		{"L2 is integration", "@L2", TestSetIntegration},
		{"L3 is integration", "@L3", TestSetIntegration},
		{"L4 is integration", "@L4", TestSetIntegration},
		{"empty defaults to unit", "", TestSetUnit},
		{"unrecognized defaults to unit", "@L5", TestSetUnit},
		{"non-level tag defaults to unit", "@smoke", TestSetUnit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTestSetForLTag(tt.ltag)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestClassifyTestByTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected TestSet
	}{
		{"no tags defaults to unit", []string{}, TestSetUnit},
		{"L0 only is unit", []string{"@L0"}, TestSetUnit},
		{"L1 only is unit", []string{"@L1"}, TestSetUnit},
		{"L2 only is integration", []string{"@L2"}, TestSetIntegration},
		{"L3 only is integration", []string{"@L3"}, TestSetIntegration},
		{"L4 only is integration", []string{"@L4"}, TestSetIntegration},
		{"L0 and L2 is integration (L2 wins)", []string{"@L0", "@L2"}, TestSetIntegration},
		{"other tags with L1 is unit", []string{"@smoke", "@L1", "@fast"}, TestSetUnit},
		{"other tags with L2 is integration", []string{"@smoke", "@L2", "@slow"}, TestSetIntegration},
		{"non-level tags only defaults to unit", []string{"@smoke", "@regression"}, TestSetUnit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyTestByTags(tt.tags)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDefaultInvalidationRules(t *testing.T) {
	rules := DefaultInvalidationRules()

	t.Run("unit rules", func(t *testing.T) {
		unitRule, ok := rules[TestSetUnit]
		require.True(t, ok, "unit rule should exist")

		assert.Equal(t, TestSetUnit, unitRule.TestSet)
		assert.True(t, unitRule.OnModuleChange, "unit tests should invalidate on module change")
		assert.False(t, unitRule.OnTransitiveChange, "unit tests should NOT invalidate on transitive change")
		assert.True(t, unitRule.OnBuildChange, "unit tests should invalidate on build change")
		assert.False(t, unitRule.OnDependencyBuildChange, "unit tests should NOT invalidate on dependency build change")
	})

	t.Run("integration rules", func(t *testing.T) {
		integrationRule, ok := rules[TestSetIntegration]
		require.True(t, ok, "integration rule should exist")

		assert.Equal(t, TestSetIntegration, integrationRule.TestSet)
		assert.True(t, integrationRule.OnModuleChange, "integration tests should invalidate on module change")
		assert.True(t, integrationRule.OnTransitiveChange, "integration tests should invalidate on transitive change")
		assert.True(t, integrationRule.OnBuildChange, "integration tests should invalidate on build change")
		assert.True(t, integrationRule.OnDependencyBuildChange, "integration tests should invalidate on dependency build change")
	})
}

func TestComputeDependencyBuildHash(t *testing.T) {
	t.Run("empty map returns empty string", func(t *testing.T) {
		result := ComputeDependencyBuildHash(map[string]string{})
		assert.Empty(t, result)
	})

	t.Run("nil map returns empty string", func(t *testing.T) {
		result := ComputeDependencyBuildHash(nil)
		assert.Empty(t, result)
	})

	t.Run("single dependency produces hash", func(t *testing.T) {
		result := ComputeDependencyBuildHash(map[string]string{
			"eac-core": "build-123",
		})
		assert.NotEmpty(t, result)
		assert.Len(t, result, 16) // 8 bytes = 16 hex chars
	})

	t.Run("same input produces same hash", func(t *testing.T) {
		deps := map[string]string{
			"eac-core":     "build-123",
			"eac-commands": "build-456",
		}
		hash1 := ComputeDependencyBuildHash(deps)
		hash2 := ComputeDependencyBuildHash(deps)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("different input produces different hash", func(t *testing.T) {
		hash1 := ComputeDependencyBuildHash(map[string]string{
			"eac-core": "build-123",
		})
		hash2 := ComputeDependencyBuildHash(map[string]string{
			"eac-core": "build-456",
		})
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("order independent", func(t *testing.T) {
		// Build maps in different order but same content
		deps1 := map[string]string{
			"eac-core":     "build-123",
			"eac-commands": "build-456",
		}
		deps2 := map[string]string{
			"eac-commands": "build-456",
			"eac-core":     "build-123",
		}
		assert.Equal(t, ComputeDependencyBuildHash(deps1), ComputeDependencyBuildHash(deps2))
	})
}
