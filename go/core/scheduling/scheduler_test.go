package scheduling

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDependencyScheduler(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp2"}, Weight: 2},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	assert.NotNil(t, s)
	assert.Equal(t, 2, s.Len())
}

func TestDependencyScheduler_LPT_Ordering(t *testing.T) {
	// LPT = Longest Processing Time First
	// Items should pop in weight descending order
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "light"}, Weight: 1},
		{ID: workunit.UnitID{Module: "mod1", Component: "heavy"}, Weight: 8},
		{ID: workunit.UnitID{Module: "mod1", Component: "medium"}, Weight: 4},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Should get heaviest first
	item1 := s.WaitForReady()
	require.NotNil(t, item1)
	assert.Equal(t, "heavy", item1.ID.Component)
	assert.Equal(t, 8, item1.Weight)
	s.MarkComplete(item1.ID)

	// Then medium
	item2 := s.WaitForReady()
	require.NotNil(t, item2)
	assert.Equal(t, "medium", item2.ID.Component)
	assert.Equal(t, 4, item2.Weight)
	s.MarkComplete(item2.ID)

	// Then light
	item3 := s.WaitForReady()
	require.NotNil(t, item3)
	assert.Equal(t, "light", item3.ID.Component)
	assert.Equal(t, 1, item3.Weight)
	s.MarkComplete(item3.ID)

	// Queue should be empty
	assert.Equal(t, 0, s.Len())
}

func TestDependencyScheduler_DependencyRespect(t *testing.T) {
	// comp2 depends on comp1
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			Weight:    8, // Higher weight but blocked
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Even though comp2 has higher weight, comp1 should pop first because comp2 is blocked
	item1 := s.WaitForReady()
	require.NotNil(t, item1)
	assert.Equal(t, "comp1", item1.ID.Component, "comp1 should pop first (comp2 is blocked)")

	// Mark comp1 complete
	s.MarkComplete(item1.ID)

	// Now comp2 should be ready
	item2 := s.WaitForReady()
	require.NotNil(t, item2)
	assert.Equal(t, "comp2", item2.ID.Component)
}

func TestDependencyScheduler_FailurePropagation(t *testing.T) {
	// comp2 depends on comp1
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			Weight:    2,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Pop and fail comp1
	item1 := s.WaitForReady()
	require.NotNil(t, item1)
	assert.Equal(t, "comp1", item1.ID.Component)
	s.MarkFailed(item1.ID)

	// comp2 should become ready (deps complete, even if failed)
	item2 := s.WaitForReady()
	require.NotNil(t, item2)
	assert.Equal(t, "comp2", item2.ID.Component)

	// comp2 should report that it has a failed dependency
	assert.True(t, s.HasFailedDependency(item2.ID))
}

func TestDependencyScheduler_InFlightTracking(t *testing.T) {
	// Verify in-flight semantics:
	// - Item returned by WaitForReady() is removed from queue
	// - Item stays "in-flight" until MarkComplete/MarkFailed
	// - Dependencies only unblock on MarkComplete
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			Weight:    2,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Initially 2 items in queue
	assert.Equal(t, 2, s.Len())

	// Pop comp1 - now it's in-flight
	item1 := s.WaitForReady()
	require.NotNil(t, item1)
	assert.Equal(t, "comp1", item1.ID.Component)

	// Only 1 item in queue now (comp2 is still there, blocked)
	assert.Equal(t, 1, s.Len())

	// comp2 is still blocked (comp1 in-flight, not complete)
	item2 := s.Next() // Non-blocking
	assert.Nil(t, item2, "comp2 should still be blocked")

	// Mark comp1 complete - now comp2 unblocks
	s.MarkComplete(item1.ID)

	// comp2 should now be ready
	item2 = s.Next()
	require.NotNil(t, item2)
	assert.Equal(t, "comp2", item2.ID.Component)
}

func TestDependencyScheduler_CycleDetection(t *testing.T) {
	// Create a cycle: A -> B -> C -> A
	work := []workunit.UnitSpec{
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "A"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "C"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "B"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "C"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "B"}},
		},
	}

	_, err := NewDependencyScheduler(work)
	require.Error(t, err)

	var cycleErr *CircularDependencyError
	require.ErrorAs(t, err, &cycleErr)
	assert.Len(t, cycleErr.Units, 3)
}

func TestDependencyScheduler_ConcurrentAccess(t *testing.T) {
	// Create many items with unique component names
	work := make([]workunit.UnitSpec, 100)
	for i := 0; i < 100; i++ {
		work[i] = workunit.UnitSpec{
			ID:     workunit.UnitID{Module: "mod1", Component: fmt.Sprintf("comp%d", i)},
			Weight: i%10 + 1,
			Index:  i,
		}
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Multiple goroutines popping concurrently
	var wg sync.WaitGroup
	popped := make(chan *workunit.UnitSpec, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				item := s.WaitForReady()
				if item == nil {
					return
				}
				popped <- item
				s.MarkComplete(item.ID)
			}
		}()
	}

	// Close scheduler after a brief delay to let workers drain
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.Close()
	}()

	wg.Wait()
	close(popped)

	// Should have popped all items
	count := 0
	for range popped {
		count++
	}
	assert.Equal(t, 100, count, "all items should be popped")
}

func TestDependencyScheduler_Stats(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			Weight:    2,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp3"},
			Weight:    3,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	stats := s.Stats()
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 3, stats.Pending)
	assert.Equal(t, 1, stats.Ready)   // Only comp1 is ready
	assert.Equal(t, 2, stats.Blocked) // comp2 and comp3 blocked on comp1
	assert.Equal(t, 0, stats.Completed)
	assert.Equal(t, 0, stats.Failed)
}

func TestDependencyScheduler_Stats_AfterCompletion(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			Weight:    2,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Pop comp1
	item := s.WaitForReady()
	require.NotNil(t, item)
	s.MarkComplete(item.ID)

	stats := s.Stats()
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 1, stats.Pending)   // Only comp2 remains
	assert.Equal(t, 1, stats.Ready)     // comp2 is now ready
	assert.Equal(t, 0, stats.Blocked)   // Nothing blocked
	assert.Equal(t, 1, stats.Completed) // comp1 completed
	assert.Equal(t, 0, stats.Failed)
}

func TestDependencyScheduler_EmptyWork(t *testing.T) {
	work := []workunit.UnitSpec{}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, 0, s.Len())

	// Next should return nil immediately for empty scheduler
	item := s.Next()
	assert.Nil(t, item)

	// WaitForReady should return nil immediately for empty scheduler
	item = s.WaitForReady()
	assert.Nil(t, item)
}

func TestDependencyScheduler_Close(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Close the scheduler
	s.Close()

	// WaitForReady should return nil after close
	item := s.WaitForReady()
	assert.Nil(t, item)
}

func TestDependencyScheduler_PreservesIndex(t *testing.T) {
	// Index field should be preserved for result ordering
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1, Index: 0},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp2"}, Weight: 8, Index: 1},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp3"}, Weight: 4, Index: 2},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Pop in LPT order but verify indices preserved
	item1 := s.WaitForReady()
	assert.Equal(t, 1, item1.Index) // comp2 (weight 8)
	s.MarkComplete(item1.ID)

	item2 := s.WaitForReady()
	assert.Equal(t, 2, item2.Index) // comp3 (weight 4)
	s.MarkComplete(item2.ID)

	item3 := s.WaitForReady()
	assert.Equal(t, 0, item3.Index) // comp1 (weight 1)
	s.MarkComplete(item3.ID)
}

func TestDependencyScheduler_Next_NonBlocking(t *testing.T) {
	// All items are blocked - Next should return nil immediately
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			Weight:    2,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Pop comp1 (the only ready item)
	item1 := s.Next()
	require.NotNil(t, item1)
	assert.Equal(t, "comp1", item1.ID.Component)

	// Next should return nil immediately (comp2 is blocked, comp1 is in-flight)
	item2 := s.Next()
	assert.Nil(t, item2, "Next should return nil when no items are ready")
}

func TestDependencyScheduler_ComplexDependencyGraph(t *testing.T) {
	// Diamond dependency pattern:
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "A"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "B"},
			Weight:    2,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "C"},
			Weight:    3,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
		{
			ID:     workunit.UnitID{Module: "mod1", Component: "D"},
			Weight: 4,
			DependsOn: []workunit.UnitID{
				{Module: "mod1", Component: "B"},
				{Module: "mod1", Component: "C"},
			},
		},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Only A is initially ready
	stats := s.Stats()
	assert.Equal(t, 1, stats.Ready)
	assert.Equal(t, 3, stats.Blocked)

	// Pop A
	itemA := s.WaitForReady()
	require.NotNil(t, itemA)
	assert.Equal(t, "A", itemA.ID.Component)
	s.MarkComplete(itemA.ID)

	// B and C should now be ready, D still blocked
	stats = s.Stats()
	assert.Equal(t, 2, stats.Ready)
	assert.Equal(t, 1, stats.Blocked)

	// Pop B and C (C has higher weight, should come first)
	itemC := s.WaitForReady()
	require.NotNil(t, itemC)
	assert.Equal(t, "C", itemC.ID.Component)
	s.MarkComplete(itemC.ID)

	itemB := s.WaitForReady()
	require.NotNil(t, itemB)
	assert.Equal(t, "B", itemB.ID.Component)
	s.MarkComplete(itemB.ID)

	// D should now be ready
	itemD := s.WaitForReady()
	require.NotNil(t, itemD)
	assert.Equal(t, "D", itemD.ID.Component)
	s.MarkComplete(itemD.ID)

	// Queue should be empty
	assert.Equal(t, 0, s.Len())
}

// --- Heap tests ---

func TestUnitHeap_LessFunction(t *testing.T) {
	h := unitHeap{
		{ID: workunit.UnitID{Component: "a"}, Weight: 5},
		{ID: workunit.UnitID{Component: "b"}, Weight: 10},
	}

	// Less returns true if i should come before j
	// For max-heap (LPT), higher weight should come first
	assert.True(t, h.Less(1, 0), "weight 10 should come before weight 5")
	assert.False(t, h.Less(0, 1), "weight 5 should not come before weight 10")
}

// --- DependencyTracker tests ---

func TestDependencyTracker_IsReady(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := NewDependencyTracker(work)

	// comp1 has no deps - always ready
	assert.True(t, dt.IsReady(workunit.UnitID{Module: "mod1", Component: "comp1"}))

	// comp2 depends on comp1 - not ready yet
	assert.False(t, dt.IsReady(workunit.UnitID{Module: "mod1", Component: "comp2"}))

	// Mark comp1 complete
	dt.MarkComplete(workunit.UnitID{Module: "mod1", Component: "comp1"})

	// Now comp2 is ready
	assert.True(t, dt.IsReady(workunit.UnitID{Module: "mod1", Component: "comp2"}))
}

func TestDependencyTracker_HasFailedDependency(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := NewDependencyTracker(work)

	// Initially no failures
	assert.False(t, dt.HasFailedDependency(workunit.UnitID{Module: "mod1", Component: "comp2"}))

	// Mark comp1 as failed
	dt.MarkFailed(workunit.UnitID{Module: "mod1", Component: "comp1"})

	// comp2 should report failed dependency
	assert.True(t, dt.HasFailedDependency(workunit.UnitID{Module: "mod1", Component: "comp2"}))

	// comp2 should also be ready now (failed deps are marked complete)
	assert.True(t, dt.IsReady(workunit.UnitID{Module: "mod1", Component: "comp2"}))
}

// --- CascadeFail tests ---

func TestDependencyTracker_CascadeFail_Chain(t *testing.T) {
	// Chain: A -> B -> C (C depends on B, B depends on A)
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "A"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "B"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "C"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "B"}},
		},
	}

	dt := NewDependencyTracker(work)

	// Mark A as failed
	dt.MarkFailed(workunit.UnitID{Module: "mod1", Component: "A"})

	// Cascade from A — should get B and C
	cascaded := dt.CascadeFail(workunit.UnitID{Module: "mod1", Component: "A"})
	assert.Len(t, cascaded, 2)

	names := make(map[string]bool)
	for _, c := range cascaded {
		names[c.Component] = true
	}
	assert.True(t, names["B"])
	assert.True(t, names["C"])

	// Both should be marked failed and completed
	assert.True(t, dt.HasFailedDependency(workunit.UnitID{Module: "mod1", Component: "B"}))
	assert.True(t, dt.IsReady(workunit.UnitID{Module: "mod1", Component: "C"})) // completed = ready
}

func TestDependencyTracker_CascadeFail_Diamond(t *testing.T) {
	// Diamond: A -> B, A -> C, B -> D, C -> D
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "A"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "B"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "C"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
		{
			ID: workunit.UnitID{Module: "mod1", Component: "D"},
			DependsOn: []workunit.UnitID{
				{Module: "mod1", Component: "B"},
				{Module: "mod1", Component: "C"},
			},
		},
	}

	dt := NewDependencyTracker(work)
	dt.MarkFailed(workunit.UnitID{Module: "mod1", Component: "A"})

	cascaded := dt.CascadeFail(workunit.UnitID{Module: "mod1", Component: "A"})
	// B, C, and D should all be cascade-failed (D via B or C)
	assert.Len(t, cascaded, 3, "diamond should cascade to all 3 dependents")

	names := make(map[string]bool)
	for _, c := range cascaded {
		names[c.Component] = true
	}
	assert.True(t, names["B"])
	assert.True(t, names["C"])
	assert.True(t, names["D"])
}

func TestDependencyTracker_CascadeFail_Partial(t *testing.T) {
	// Two independent chains: A->B and C->D
	// Fail A — only B should cascade, not C or D
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "A"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "B"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
		{ID: workunit.UnitID{Module: "mod2", Component: "C"}},
		{
			ID:        workunit.UnitID{Module: "mod2", Component: "D"},
			DependsOn: []workunit.UnitID{{Module: "mod2", Component: "C"}},
		},
	}

	dt := NewDependencyTracker(work)
	dt.MarkFailed(workunit.UnitID{Module: "mod1", Component: "A"})

	cascaded := dt.CascadeFail(workunit.UnitID{Module: "mod1", Component: "A"})
	assert.Len(t, cascaded, 1, "only B should cascade")
	assert.Equal(t, "B", cascaded[0].Component)
}

func TestDependencyTracker_CascadeFail_NoDependents(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "A"}},
		{ID: workunit.UnitID{Module: "mod1", Component: "B"}},
	}

	dt := NewDependencyTracker(work)
	dt.MarkFailed(workunit.UnitID{Module: "mod1", Component: "A"})

	cascaded := dt.CascadeFail(workunit.UnitID{Module: "mod1", Component: "A"})
	assert.Empty(t, cascaded, "no dependents means no cascade")
}

// --- MarkFailedCascade scheduler tests ---

func TestDependencyScheduler_MarkFailedCascade_Chain(t *testing.T) {
	// A -> B -> C, fail A — B and C removed from queue
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "A"}, Weight: 1, Index: 0},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "B"},
			Weight:    2,
			Index:     1,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "C"},
			Weight:    3,
			Index:     2,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "B"}},
		},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Pop A
	itemA := s.WaitForReady()
	require.NotNil(t, itemA)
	assert.Equal(t, "A", itemA.ID.Component)

	// Cascade fail A
	cascaded := s.MarkFailedCascade(itemA.ID)

	// B and C should be cascade-failed
	assert.Len(t, cascaded, 2)
	names := make(map[string]bool)
	for _, c := range cascaded {
		names[c.ID.Component] = true
	}
	assert.True(t, names["B"])
	assert.True(t, names["C"])

	// Queue should be empty
	assert.Equal(t, 0, s.Len())

	// WaitForReady should return nil (exhausted)
	item := s.WaitForReady()
	assert.Nil(t, item)
}

func TestDependencyScheduler_MarkFailedCascade_PreservesIndex(t *testing.T) {
	// Verify cascaded specs preserve their original Index
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "A"}, Weight: 1, Index: 5},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "B"},
			Weight:    2,
			Index:     10,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	itemA := s.WaitForReady()
	require.NotNil(t, itemA)

	cascaded := s.MarkFailedCascade(itemA.ID)
	require.Len(t, cascaded, 1)
	assert.Equal(t, 10, cascaded[0].Index, "Index should be preserved")
}

func TestDependencyScheduler_MarkFailedCascade_InFlightNotCascaded(t *testing.T) {
	// A -> B, both popped. Fail A — B is in-flight (already popped), not in queue.
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "A"}, Weight: 1, Index: 0},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "B"},
			Weight:    2,
			Index:     1,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// Pop A
	itemA := s.WaitForReady()
	require.NotNil(t, itemA)
	// Complete A so B becomes ready
	s.MarkComplete(itemA.ID)

	// Pop B (now in-flight)
	itemB := s.WaitForReady()
	require.NotNil(t, itemB)
	assert.Equal(t, "B", itemB.ID.Component)

	// Now fail A via cascade — B is already popped, so cascade should return empty
	cascaded := s.MarkFailedCascade(itemA.ID)
	assert.Empty(t, cascaded, "in-flight items should not be cascade-failed")

	// But HasFailedDependency should still detect the failure
	assert.True(t, s.HasFailedDependency(itemB.ID))
}

func TestDependencyScheduler_MarkFailedCascade_NoDependents(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "A"}, Weight: 1, Index: 0},
		{ID: workunit.UnitID{Module: "mod1", Component: "B"}, Weight: 2, Index: 1},
	}

	s, err := NewDependencyScheduler(work)
	require.NoError(t, err)

	// LPT: B (weight 2) pops first
	itemB := s.WaitForReady()
	require.NotNil(t, itemB)
	assert.Equal(t, "B", itemB.ID.Component)

	// Fail B — A is independent, should not cascade
	cascaded := s.MarkFailedCascade(itemB.ID)
	assert.Empty(t, cascaded, "independent items should not cascade")

	// A should still be available
	itemA := s.WaitForReady()
	require.NotNil(t, itemA)
	assert.Equal(t, "A", itemA.ID.Component)
}

func TestDependencyTracker_GetDependsOn(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := NewDependencyTracker(work)

	deps := dt.GetDependsOn(workunit.UnitID{Module: "mod1", Component: "comp2"})
	require.Len(t, deps, 1)
	assert.Equal(t, "comp1", deps[0].Component)
}

func TestDependencyTracker_GetBlocks(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp3"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := NewDependencyTracker(work)

	blocks := dt.GetBlocks(workunit.UnitID{Module: "mod1", Component: "comp1"})
	require.Len(t, blocks, 2)

	// Check that both comp2 and comp3 depend on comp1
	components := []string{blocks[0].Component, blocks[1].Component}
	assert.Contains(t, components, "comp2")
	assert.Contains(t, components, "comp3")
}

// --- WithSkipValidation tests ---

func TestNewDependencyScheduler_WithSkipValidation(t *testing.T) {
	// Normal case: acyclic graph with SkipValidation
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			Weight:    2,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	s, err := NewDependencyScheduler(work, WithSkipValidation())
	require.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, 2, s.Len())
}

func TestNewDependencyScheduler_WithSkipValidation_CyclesNotDetected(t *testing.T) {
	// WithSkipValidation skips cycle detection - cyclic graph succeeds construction
	work := []workunit.UnitSpec{
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "A"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "B"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "B"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
	}

	// Without SkipValidation - should fail
	_, err := NewDependencyScheduler(work)
	require.Error(t, err)

	// With SkipValidation - should succeed (caller guarantees acyclic)
	s, err := NewDependencyScheduler(work, WithSkipValidation())
	require.NoError(t, err)
	assert.NotNil(t, s)
}

// --- Lazy reverse map tests ---

func TestDependencyTracker_LazyBlocks_NotBuiltUntilNeeded(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := NewDependencyTracker(work)

	// blocks should be nil (not yet built)
	assert.Nil(t, dt.blocks, "reverse map should be lazy - not built at construction time")

	// Forward map should work immediately
	assert.True(t, dt.IsReady(workunit.UnitID{Module: "mod1", Component: "comp1"}))
	assert.False(t, dt.IsReady(workunit.UnitID{Module: "mod1", Component: "comp2"}))
}

func TestDependencyTracker_LazyBlocks_BuiltOnGetBlocks(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	dt := NewDependencyTracker(work)
	assert.Nil(t, dt.blocks)

	// GetBlocks triggers lazy construction
	blocks := dt.GetBlocks(workunit.UnitID{Module: "mod1", Component: "comp1"})
	assert.NotNil(t, dt.blocks, "reverse map should be built after GetBlocks")
	require.Len(t, blocks, 1)
	assert.Equal(t, "comp2", blocks[0].Component)
}

func TestDependencyTracker_LazyBlocks_BuiltOnCascadeFail(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "A"}},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "B"},
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "A"}},
		},
	}

	dt := NewDependencyTracker(work)
	assert.Nil(t, dt.blocks)

	dt.MarkFailed(workunit.UnitID{Module: "mod1", Component: "A"})

	// CascadeFail triggers lazy construction
	cascaded := dt.CascadeFail(workunit.UnitID{Module: "mod1", Component: "A"})
	assert.NotNil(t, dt.blocks, "reverse map should be built after CascadeFail")
	require.Len(t, cascaded, 1)
	assert.Equal(t, "B", cascaded[0].Component)
}

func TestDependencyTracker_LazyBlocks_WorkReleasedAfterBuild(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}},
	}

	dt := NewDependencyTracker(work)
	assert.NotNil(t, dt.work, "work should be retained before blocks built")

	// Trigger build
	dt.GetBlocks(workunit.UnitID{Module: "mod1", Component: "comp1"})
	assert.Nil(t, dt.work, "work should be released after blocks built")
}
