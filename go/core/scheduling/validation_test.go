package scheduling

import (
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateNoCycles_EmptyWork(t *testing.T) {
	err := validateNoCycles([]workunit.UnitSpec{})
	assert.NoError(t, err)
}

func TestValidateNoCycles_SingleUnitNoCycle(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}},
	}
	err := validateNoCycles(work)
	assert.NoError(t, err)
}

func TestValidateNoCycles_LinearDependencyNoCycle(t *testing.T) {
	// A -> B (B depends on A) - no cycle
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}},
		{
			ID:        workunit.UnitID{Module: "m", Component: "b", Tool: "t", Action: core.ActionBuild},
			DependsOn: []workunit.UnitID{{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}},
		},
	}
	err := validateNoCycles(work)
	assert.NoError(t, err)
}

func TestValidateNoCycles_SimpleCycle_ReportsUnits(t *testing.T) {
	// A -> B -> A (cycle)
	work := []workunit.UnitSpec{
		{
			ID:        workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild},
			DependsOn: []workunit.UnitID{{Module: "m", Component: "b", Tool: "t", Action: core.ActionBuild}},
		},
		{
			ID:        workunit.UnitID{Module: "m", Component: "b", Tool: "t", Action: core.ActionBuild},
			DependsOn: []workunit.UnitID{{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}},
		},
	}
	err := validateNoCycles(work)
	require.Error(t, err)

	var cycleErr *CircularDependencyError
	require.ErrorAs(t, err, &cycleErr)

	// CRITICAL: The error must contain units, not be empty
	assert.Len(t, cycleErr.Units, 2, "cycle error should contain 2 units, not empty array")
	assert.Contains(t, cycleErr.Units, "build:m:a:t")
	assert.Contains(t, cycleErr.Units, "build:m:b:t")
}

func TestValidateNoCycles_TripleCycle_ReportsAllMembers(t *testing.T) {
	// A -> B -> C -> A (all three should be reported)
	work := []workunit.UnitSpec{
		{
			ID:        workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild},
			DependsOn: []workunit.UnitID{{Module: "m", Component: "c", Tool: "t", Action: core.ActionBuild}},
		},
		{
			ID:        workunit.UnitID{Module: "m", Component: "b", Tool: "t", Action: core.ActionBuild},
			DependsOn: []workunit.UnitID{{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}},
		},
		{
			ID:        workunit.UnitID{Module: "m", Component: "c", Tool: "t", Action: core.ActionBuild},
			DependsOn: []workunit.UnitID{{Module: "m", Component: "b", Tool: "t", Action: core.ActionBuild}},
		},
	}
	err := validateNoCycles(work)
	require.Error(t, err)

	var cycleErr *CircularDependencyError
	require.ErrorAs(t, err, &cycleErr)
	assert.Len(t, cycleErr.Units, 3, "all 3 cycle members should be reported")
}

func TestValidateNoCycles_ExternalDependencyIgnored(t *testing.T) {
	// Unit depends on something not in work set - should be processed fine
	work := []workunit.UnitSpec{
		{
			ID:        workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild},
			DependsOn: []workunit.UnitID{{Module: "external", Component: "x", Tool: "t", Action: core.ActionBuild}},
		},
	}
	err := validateNoCycles(work)
	assert.NoError(t, err, "external dependencies should be ignored")
}

func TestValidateNoCycles_MixedCycleAndNonCycle(t *testing.T) {
	// D has no deps (should process fine)
	// A -> B -> C -> A (cycle - these 3 should be reported)
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "m", Component: "d", Tool: "t", Action: core.ActionBuild}},
		{
			ID:        workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild},
			DependsOn: []workunit.UnitID{{Module: "m", Component: "c", Tool: "t", Action: core.ActionBuild}},
		},
		{
			ID:        workunit.UnitID{Module: "m", Component: "b", Tool: "t", Action: core.ActionBuild},
			DependsOn: []workunit.UnitID{{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}},
		},
		{
			ID:        workunit.UnitID{Module: "m", Component: "c", Tool: "t", Action: core.ActionBuild},
			DependsOn: []workunit.UnitID{{Module: "m", Component: "b", Tool: "t", Action: core.ActionBuild}},
		},
	}
	err := validateNoCycles(work)
	require.Error(t, err)

	var cycleErr *CircularDependencyError
	require.ErrorAs(t, err, &cycleErr)

	// Only the cycle members should be reported (not D)
	assert.Len(t, cycleErr.Units, 3, "only cycle members should be reported")
	assert.NotContains(t, cycleErr.Units, "build:m:d:t", "D is not in the cycle")
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"empty slice", []string{}, "[]"},
		{"single item", []string{"a"}, "[a]"},
		{"multiple items", []string{"a", "b", "c"}, "[a, b, c]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinStrings(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCircularDependencyError_Error(t *testing.T) {
	err := &CircularDependencyError{Units: []string{"a", "b"}}
	msg := err.Error()
	assert.Contains(t, msg, "circular dependency")
	assert.Contains(t, msg, "a")
	assert.Contains(t, msg, "b")
}

func TestValidateNoCycles_DuplicateUnitIDs_NoCycle(t *testing.T) {
	// If work slice contains duplicates, should not falsely report a cycle
	// This tests the fix for comparing against unitSet length instead of work length
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}},
		{ID: workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}}, // duplicate
	}
	err := validateNoCycles(work)
	// Should NOT report an error - duplicates don't create a cycle
	assert.NoError(t, err, "duplicate unit IDs should not cause false cycle detection")
}

// =============================================================================
// Duplicate Detection Tests (TDD: validateNoDuplicates doesn't exist yet)
// =============================================================================

func TestDuplicateUnitIDError_Error(t *testing.T) {
	err := &DuplicateUnitIDError{
		FirstIndex:  0,
		SecondIndex: 3,
		Longname:    "build:core:go:go",
	}
	msg := err.Error()

	assert.Contains(t, msg, "duplicate")
	assert.Contains(t, msg, "0")
	assert.Contains(t, msg, "3")
	assert.Contains(t, msg, "build:core:go:go")

	// Verify the exact format
	expected := "duplicate UnitID at indices 0 and 3: build:core:go:go"
	assert.Equal(t, expected, msg)
}

func TestValidateNoDuplicates_EmptySlice(t *testing.T) {
	err := validateNoDuplicates([]workunit.UnitSpec{})
	assert.NoError(t, err, "empty slice should return nil")
}

func TestValidateNoDuplicates_UniqueWorkUnits(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}},
		{ID: workunit.UnitID{Module: "m", Component: "b", Tool: "t", Action: core.ActionBuild}},
		{ID: workunit.UnitID{Module: "m", Component: "c", Tool: "t", Action: core.ActionBuild}},
	}
	err := validateNoDuplicates(work)
	assert.NoError(t, err, "unique work units should return nil")
}

func TestValidateNoDuplicates_AdjacentDuplicates(t *testing.T) {
	// Two units with same Longname at indices 0 and 1
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}},
		{ID: workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}}, // duplicate
	}
	err := validateNoDuplicates(work)
	require.Error(t, err, "should detect duplicate")

	var dupErr *DuplicateUnitIDError
	require.ErrorAs(t, err, &dupErr)

	assert.Equal(t, 0, dupErr.FirstIndex, "first index should be 0")
	assert.Equal(t, 1, dupErr.SecondIndex, "second index should be 1")
	assert.Equal(t, "build:m:a:t", dupErr.Longname)
}

func TestValidateNoDuplicates_FindsFirstDuplicatePair(t *testing.T) {
	// Three units: A, A, B (indices 0, 1, 2)
	// Should report indices 0 and 1 (the first duplicate pair)
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}}, // index 0
		{ID: workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}}, // index 1 (dup of 0)
		{ID: workunit.UnitID{Module: "m", Component: "b", Tool: "t", Action: core.ActionBuild}}, // index 2
	}
	err := validateNoDuplicates(work)
	require.Error(t, err, "should detect duplicate")

	var dupErr *DuplicateUnitIDError
	require.ErrorAs(t, err, &dupErr)

	assert.Equal(t, 0, dupErr.FirstIndex, "should report first occurrence at index 0")
	assert.Equal(t, 1, dupErr.SecondIndex, "should report second occurrence at index 1")
	assert.Equal(t, "build:m:a:t", dupErr.Longname)
}

func TestValidateNoDuplicates_NonAdjacentDuplicates(t *testing.T) {
	// Four units: A, B, C, A (indices 0, 1, 2, 3)
	// Should report indices 0 and 3
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}}, // index 0
		{ID: workunit.UnitID{Module: "m", Component: "b", Tool: "t", Action: core.ActionBuild}}, // index 1
		{ID: workunit.UnitID{Module: "m", Component: "c", Tool: "t", Action: core.ActionBuild}}, // index 2
		{ID: workunit.UnitID{Module: "m", Component: "a", Tool: "t", Action: core.ActionBuild}}, // index 3 (dup of 0)
	}
	err := validateNoDuplicates(work)
	require.Error(t, err, "should detect non-adjacent duplicate")

	var dupErr *DuplicateUnitIDError
	require.ErrorAs(t, err, &dupErr)

	assert.Equal(t, 0, dupErr.FirstIndex, "should report first occurrence at index 0")
	assert.Equal(t, 3, dupErr.SecondIndex, "should report second occurrence at index 3")
	assert.Equal(t, "build:m:a:t", dupErr.Longname)
}
