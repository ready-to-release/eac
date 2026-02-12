package scheduling

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Helper to build UnitIDs concisely (matches existing test patterns)
// =============================================================================

func uid(module, comp string) workunit.UnitID {
	return workunit.UnitID{Module: module, ComponentType: comp, ComponentName: comp}
}

func spec(module, comp string, weight int, deps ...workunit.UnitID) workunit.UnitSpec {
	return workunit.UnitSpec{
		ID:        uid(module, comp),
		Weight:    weight,
		DependsOn: deps,
	}
}

// =============================================================================
// IsReady tests
// =============================================================================

func TestDependencyTracker_IsReady_NoDependencies(t *testing.T) {
	tests := []struct {
		name string
		work []workunit.UnitSpec
		id   workunit.UnitID
		want bool
	}{
		{
			name: "unit with no dependencies is ready",
			work: []workunit.UnitSpec{spec("mod1", "A", 1)},
			id:   uid("mod1", "A"),
			want: true,
		},
		{
			name: "unknown unit with no tracked dependencies is ready",
			work: []workunit.UnitSpec{spec("mod1", "A", 1)},
			id:   uid("mod1", "unknown"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := NewDependencyTracker(tt.work)
			assert.Equal(t, tt.want, dt.IsReady(tt.id))
		})
	}
}

func TestDependencyTracker_IsReady_SingleDependency(t *testing.T) {
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod1", "B", 2, uid("mod1", "A")),
	}
	dt := NewDependencyTracker(work)

	// B depends on A, A is not complete
	assert.False(t, dt.IsReady(uid("mod1", "B")))

	// Complete A
	dt.MarkComplete(uid("mod1", "A"))
	assert.True(t, dt.IsReady(uid("mod1", "B")))
}

func TestDependencyTracker_IsReady_MultipleDependencies(t *testing.T) {
	// C depends on both A and B
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod1", "B", 1),
		spec("mod1", "C", 2, uid("mod1", "A"), uid("mod1", "B")),
	}
	dt := NewDependencyTracker(work)

	// C not ready initially
	assert.False(t, dt.IsReady(uid("mod1", "C")))

	// Complete only A - C still not ready (B incomplete)
	dt.MarkComplete(uid("mod1", "A"))
	assert.False(t, dt.IsReady(uid("mod1", "C")))

	// Complete B - now C is ready
	dt.MarkComplete(uid("mod1", "B"))
	assert.True(t, dt.IsReady(uid("mod1", "C")))
}

func TestDependencyTracker_IsReady_FailedDepCountsAsComplete(t *testing.T) {
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod1", "B", 2, uid("mod1", "A")),
	}
	dt := NewDependencyTracker(work)

	// Fail A (failed deps are marked complete)
	dt.MarkFailed(uid("mod1", "A"))

	// B should be ready because A is complete (even though failed)
	assert.True(t, dt.IsReady(uid("mod1", "B")))
}

// =============================================================================
// MarkComplete / MarkFailed tests
// =============================================================================

func TestDependencyTracker_MarkComplete_IncreasesCompletedCount(t *testing.T) {
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod1", "B", 1),
	}
	dt := NewDependencyTracker(work)

	assert.Equal(t, 0, dt.CompletedCount())

	dt.MarkComplete(uid("mod1", "A"))
	assert.Equal(t, 1, dt.CompletedCount())

	dt.MarkComplete(uid("mod1", "B"))
	assert.Equal(t, 2, dt.CompletedCount())
}

func TestDependencyTracker_MarkFailed_IncreasesFailedAndCompletedCounts(t *testing.T) {
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod1", "B", 1),
	}
	dt := NewDependencyTracker(work)

	assert.Equal(t, 0, dt.FailedCount())
	assert.Equal(t, 0, dt.CompletedCount())

	dt.MarkFailed(uid("mod1", "A"))
	assert.Equal(t, 1, dt.FailedCount())
	assert.Equal(t, 1, dt.CompletedCount(), "failed units are also marked complete")

	dt.MarkComplete(uid("mod1", "B"))
	assert.Equal(t, 1, dt.FailedCount())
	assert.Equal(t, 2, dt.CompletedCount())
}

func TestDependencyTracker_MarkComplete_Idempotent(t *testing.T) {
	work := []workunit.UnitSpec{spec("mod1", "A", 1)}
	dt := NewDependencyTracker(work)

	dt.MarkComplete(uid("mod1", "A"))
	assert.Equal(t, 1, dt.CompletedCount())

	// Mark again - count should remain 1 (boolean map)
	dt.MarkComplete(uid("mod1", "A"))
	assert.Equal(t, 1, dt.CompletedCount())
}

// =============================================================================
// HasFailedDependency tests
// =============================================================================

func TestDependencyTracker_HasFailedDependency_NoFailures(t *testing.T) {
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod1", "B", 2, uid("mod1", "A")),
	}
	dt := NewDependencyTracker(work)

	assert.False(t, dt.HasFailedDependency(uid("mod1", "B")))

	dt.MarkComplete(uid("mod1", "A"))
	assert.False(t, dt.HasFailedDependency(uid("mod1", "B")))
}

func TestDependencyTracker_HasFailedDependency_WithFailure(t *testing.T) {
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod1", "B", 1),
		spec("mod1", "C", 2, uid("mod1", "A"), uid("mod1", "B")),
	}
	dt := NewDependencyTracker(work)

	dt.MarkComplete(uid("mod1", "A"))
	dt.MarkFailed(uid("mod1", "B"))

	assert.True(t, dt.HasFailedDependency(uid("mod1", "C")))
}

func TestDependencyTracker_HasFailedDependency_UnitWithNoDeps(t *testing.T) {
	work := []workunit.UnitSpec{spec("mod1", "A", 1)}
	dt := NewDependencyTracker(work)

	assert.False(t, dt.HasFailedDependency(uid("mod1", "A")),
		"unit with no dependencies can never have a failed dependency")
}

// =============================================================================
// FailedDependencyModules tests
// =============================================================================

func TestDependencyTracker_FailedDependencyModules(t *testing.T) {
	tests := []struct {
		name        string
		work        []workunit.UnitSpec
		failIDs     []workunit.UnitID
		queryID     workunit.UnitID
		wantModules []string
	}{
		{
			name: "no failures returns nil",
			work: []workunit.UnitSpec{
				spec("mod1", "A", 1),
				spec("mod1", "B", 2, uid("mod1", "A")),
			},
			failIDs:     nil,
			queryID:     uid("mod1", "B"),
			wantModules: nil,
		},
		{
			name: "single failed dependency module",
			work: []workunit.UnitSpec{
				spec("mod1", "A", 1),
				spec("mod2", "B", 2, uid("mod1", "A")),
			},
			failIDs:     []workunit.UnitID{uid("mod1", "A")},
			queryID:     uid("mod2", "B"),
			wantModules: []string{"mod1"},
		},
		{
			name: "multiple failed deps in same module deduplicates",
			work: []workunit.UnitSpec{
				spec("mod1", "A", 1),
				spec("mod1", "B", 1),
				spec("mod2", "C", 2, uid("mod1", "A"), uid("mod1", "B")),
			},
			failIDs:     []workunit.UnitID{uid("mod1", "A"), uid("mod1", "B")},
			queryID:     uid("mod2", "C"),
			wantModules: []string{"mod1"},
		},
		{
			name: "failed deps across different modules",
			work: []workunit.UnitSpec{
				spec("mod1", "A", 1),
				spec("mod2", "B", 1),
				spec("mod3", "C", 2, uid("mod1", "A"), uid("mod2", "B")),
			},
			failIDs:     []workunit.UnitID{uid("mod1", "A"), uid("mod2", "B")},
			queryID:     uid("mod3", "C"),
			wantModules: []string{"mod1", "mod2"},
		},
		{
			name: "only failed deps are reported not successful ones",
			work: []workunit.UnitSpec{
				spec("mod1", "A", 1),
				spec("mod2", "B", 1),
				spec("mod3", "C", 2, uid("mod1", "A"), uid("mod2", "B")),
			},
			failIDs:     []workunit.UnitID{uid("mod1", "A")}, // only A fails
			queryID:     uid("mod3", "C"),
			wantModules: []string{"mod1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := NewDependencyTracker(tt.work)
			for _, id := range tt.failIDs {
				dt.MarkFailed(id)
			}

			got := dt.FailedDependencyModules(tt.queryID)

			if tt.wantModules == nil {
				assert.Nil(t, got)
			} else {
				assert.ElementsMatch(t, tt.wantModules, got)
			}
		})
	}
}

// =============================================================================
// GetDependsOn tests (table-driven, extends coverage from scheduler_test.go)
// =============================================================================

func TestDependencyTracker_GetDependsOn_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		work     []workunit.UnitSpec
		queryID  workunit.UnitID
		wantDeps []workunit.UnitID
	}{
		{
			name:     "unit with no dependencies",
			work:     []workunit.UnitSpec{spec("mod1", "A", 1)},
			queryID:  uid("mod1", "A"),
			wantDeps: nil,
		},
		{
			name: "unit with multiple dependencies",
			work: []workunit.UnitSpec{
				spec("mod1", "A", 1),
				spec("mod1", "B", 1),
				spec("mod1", "C", 2, uid("mod1", "A"), uid("mod1", "B")),
			},
			queryID:  uid("mod1", "C"),
			wantDeps: []workunit.UnitID{uid("mod1", "A"), uid("mod1", "B")},
		},
		{
			name:     "unknown unit returns nil",
			work:     []workunit.UnitSpec{spec("mod1", "A", 1)},
			queryID:  uid("mod1", "unknown"),
			wantDeps: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := NewDependencyTracker(tt.work)
			got := dt.GetDependsOn(tt.queryID)
			assert.Equal(t, tt.wantDeps, got)
		})
	}
}

// =============================================================================
// GetBlocks tests (reverse dependencies, table-driven, extends scheduler_test.go)
// =============================================================================

func TestDependencyTracker_GetBlocks_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		work       []workunit.UnitSpec
		queryID    workunit.UnitID
		wantBlocks []string // component names for easy assertion
	}{
		{
			name:       "unit with no dependents",
			work:       []workunit.UnitSpec{spec("mod1", "A", 1)},
			queryID:    uid("mod1", "A"),
			wantBlocks: nil,
		},
		{
			name: "unit blocks one dependent",
			work: []workunit.UnitSpec{
				spec("mod1", "A", 1),
				spec("mod1", "B", 2, uid("mod1", "A")),
			},
			queryID:    uid("mod1", "A"),
			wantBlocks: []string{"B"},
		},
		{
			name: "unit blocks multiple dependents",
			work: []workunit.UnitSpec{
				spec("mod1", "A", 1),
				spec("mod1", "B", 2, uid("mod1", "A")),
				spec("mod1", "C", 3, uid("mod1", "A")),
				spec("mod1", "D", 4, uid("mod1", "A")),
			},
			queryID:    uid("mod1", "A"),
			wantBlocks: []string{"B", "C", "D"},
		},
		{
			name:       "unknown unit has no blocks",
			work:       []workunit.UnitSpec{spec("mod1", "A", 1)},
			queryID:    uid("mod1", "unknown"),
			wantBlocks: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := NewDependencyTracker(tt.work)
			blocks := dt.GetBlocks(tt.queryID)

			if tt.wantBlocks == nil {
				assert.Empty(t, blocks)
			} else {
				gotNames := make([]string, len(blocks))
				for i, b := range blocks {
					gotNames[i] = b.ComponentName
				}
				assert.ElementsMatch(t, tt.wantBlocks, gotNames)
			}
		})
	}
}

// =============================================================================
// CascadeFail tests
// =============================================================================

func TestDependencyTracker_CascadeFail(t *testing.T) {
	tests := []struct {
		name          string
		work          []workunit.UnitSpec
		failID        workunit.UnitID
		wantCascaded  []string // component names
		wantFailCount int
	}{
		{
			name:          "no dependents returns empty",
			work:          []workunit.UnitSpec{spec("mod1", "A", 1), spec("mod1", "B", 1)},
			failID:        uid("mod1", "A"),
			wantCascaded:  nil,
			wantFailCount: 1, // only the original
		},
		{
			name: "linear chain cascades",
			work: []workunit.UnitSpec{
				spec("mod1", "A", 1),
				spec("mod1", "B", 2, uid("mod1", "A")),
				spec("mod1", "C", 3, uid("mod1", "B")),
			},
			failID:        uid("mod1", "A"),
			wantCascaded:  []string{"B", "C"},
			wantFailCount: 3,
		},
		{
			name: "diamond dependency cascades without duplicates",
			work: []workunit.UnitSpec{
				spec("mod1", "A", 1),
				spec("mod1", "B", 2, uid("mod1", "A")),
				spec("mod1", "C", 3, uid("mod1", "A")),
				spec("mod1", "D", 4, uid("mod1", "B"), uid("mod1", "C")),
			},
			failID:        uid("mod1", "A"),
			wantCascaded:  []string{"B", "C", "D"},
			wantFailCount: 4,
		},
		{
			name: "independent chains are not affected",
			work: []workunit.UnitSpec{
				spec("mod1", "A", 1),
				spec("mod1", "B", 2, uid("mod1", "A")),
				spec("mod2", "C", 1),
				spec("mod2", "D", 2, uid("mod2", "C")),
			},
			failID:        uid("mod1", "A"),
			wantCascaded:  []string{"B"},
			wantFailCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := NewDependencyTracker(tt.work)
			dt.MarkFailed(tt.failID)

			cascaded := dt.CascadeFail(tt.failID)

			if tt.wantCascaded == nil {
				assert.Empty(t, cascaded)
			} else {
				gotNames := make([]string, len(cascaded))
				for i, c := range cascaded {
					gotNames[i] = c.ComponentName
				}
				assert.ElementsMatch(t, tt.wantCascaded, gotNames)
			}
			assert.Equal(t, tt.wantFailCount, dt.FailedCount())
		})
	}
}

func TestDependencyTracker_CascadeFail_AlreadyFailedSkipped(t *testing.T) {
	// If B is already failed before cascade from A, B is skipped entirely.
	// Because CascadeFail skips already-failed nodes (to avoid infinite loops
	// on diamond deps), B's subtree (C) is NOT reached through A's cascade.
	// The caller is expected to have cascaded B's dependents when B was originally failed.
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod1", "B", 2, uid("mod1", "A")),
		spec("mod1", "C", 3, uid("mod1", "B")),
	}
	dt := NewDependencyTracker(work)

	// Fail B first (with its own cascade)
	dt.MarkFailed(uid("mod1", "B"))
	cascadedFromB := dt.CascadeFail(uid("mod1", "B"))
	require.Len(t, cascadedFromB, 1, "C should cascade from B")
	assert.Equal(t, "C", cascadedFromB[0].ComponentName)

	// Now fail A
	dt.MarkFailed(uid("mod1", "A"))
	cascadedFromA := dt.CascadeFail(uid("mod1", "A"))

	// B and C are already failed, so nothing new cascades from A
	assert.Empty(t, cascadedFromA, "already-failed nodes are skipped in cascade")
}

// =============================================================================
// Lazy reverse map (ensureBlocks) tests
// =============================================================================

func TestDependencyTracker_EnsureBlocks_LazyConstruction(t *testing.T) {
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod1", "B", 2, uid("mod1", "A")),
	}
	dt := NewDependencyTracker(work)

	// blocks not built yet
	assert.Nil(t, dt.blocks)

	// Forward operations don't trigger build
	dt.IsReady(uid("mod1", "A"))
	dt.MarkComplete(uid("mod1", "A"))
	dt.HasFailedDependency(uid("mod1", "B"))
	assert.Nil(t, dt.blocks, "forward-only operations should not build the reverse map")
}

func TestDependencyTracker_EnsureBlocks_Idempotent(t *testing.T) {
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod1", "B", 2, uid("mod1", "A")),
	}
	dt := NewDependencyTracker(work)

	// First call builds it
	blocks1 := dt.GetBlocks(uid("mod1", "A"))
	assert.NotNil(t, dt.blocks)
	assert.Nil(t, dt.work, "work should be released after build")

	// Second call returns same result (idempotent, no panic from nil work)
	blocks2 := dt.GetBlocks(uid("mod1", "A"))
	assert.Equal(t, len(blocks1), len(blocks2))
}

func TestDependencyTracker_EnsureBlocks_WorkReleasedAfterBuild(t *testing.T) {
	work := []workunit.UnitSpec{spec("mod1", "A", 1)}
	dt := NewDependencyTracker(work)

	assert.NotNil(t, dt.work, "work slice retained before blocks built")

	dt.GetBlocks(uid("mod1", "A"))
	assert.Nil(t, dt.work, "work slice released after blocks built")
}

// =============================================================================
// CompletedCount / FailedCount tests
// =============================================================================

func TestDependencyTracker_Counts(t *testing.T) {
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod1", "B", 1),
		spec("mod1", "C", 1),
	}
	dt := NewDependencyTracker(work)

	assert.Equal(t, 0, dt.CompletedCount())
	assert.Equal(t, 0, dt.FailedCount())

	dt.MarkComplete(uid("mod1", "A"))
	assert.Equal(t, 1, dt.CompletedCount())
	assert.Equal(t, 0, dt.FailedCount())

	dt.MarkFailed(uid("mod1", "B"))
	assert.Equal(t, 2, dt.CompletedCount(), "failed units also count as completed")
	assert.Equal(t, 1, dt.FailedCount())

	dt.MarkComplete(uid("mod1", "C"))
	assert.Equal(t, 3, dt.CompletedCount())
	assert.Equal(t, 1, dt.FailedCount())
}

// =============================================================================
// Empty tracker edge cases
// =============================================================================

func TestDependencyTracker_EmptyWork(t *testing.T) {
	dt := NewDependencyTracker([]workunit.UnitSpec{})

	assert.Equal(t, 0, dt.CompletedCount())
	assert.Equal(t, 0, dt.FailedCount())

	// Operations on unknown IDs should not panic
	assert.True(t, dt.IsReady(uid("mod1", "A")))
	assert.False(t, dt.HasFailedDependency(uid("mod1", "A")))
	assert.Nil(t, dt.GetDependsOn(uid("mod1", "A")))
	assert.Empty(t, dt.GetBlocks(uid("mod1", "A")))
	assert.Nil(t, dt.FailedDependencyModules(uid("mod1", "A")))
	assert.Empty(t, dt.CascadeFail(uid("mod1", "A")))
}

// =============================================================================
// Cross-module dependency tracking
// =============================================================================

func TestDependencyTracker_CrossModuleDependencies(t *testing.T) {
	// mod2:B depends on mod1:A (cross-module dependency)
	work := []workunit.UnitSpec{
		spec("mod1", "A", 1),
		spec("mod2", "B", 2, uid("mod1", "A")),
	}
	dt := NewDependencyTracker(work)

	assert.False(t, dt.IsReady(uid("mod2", "B")))

	dt.MarkComplete(uid("mod1", "A"))
	assert.True(t, dt.IsReady(uid("mod2", "B")))

	// Verify reverse map tracks cross-module
	blocks := dt.GetBlocks(uid("mod1", "A"))
	require.Len(t, blocks, 1)
	assert.Equal(t, "mod2", blocks[0].Module)
	assert.Equal(t, "B", blocks[0].ComponentName)
}
