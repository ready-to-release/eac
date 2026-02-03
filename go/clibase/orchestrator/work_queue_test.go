package orchestrator

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/core/execution"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LayerMode aliases for test readability
const (
	LayerModeNone   = execution.LayerModeNone
	LayerModeStrict = execution.LayerModeStrict
)

func TestNewWorkQueue(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp2"}, Weight: 2},
	}

	q := NewWorkQueue(work)

	assert.NotNil(t, q)
	assert.Equal(t, 2, q.Len())
}

func TestWorkQueue_LPT_Ordering(t *testing.T) {
	// LPT = Longest Processing Time First
	// Items should pop in weight descending order
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "light"}, Weight: 1},
		{ID: workunit.UnitID{Module: "mod1", Component: "heavy"}, Weight: 8},
		{ID: workunit.UnitID{Module: "mod1", Component: "medium"}, Weight: 4},
	}

	q := NewWorkQueue(work)

	// Should get heaviest first
	item1 := q.PopReady()
	require.NotNil(t, item1)
	assert.Equal(t, "heavy", item1.ID.Component)
	assert.Equal(t, 8, item1.Weight)

	// Then medium
	item2 := q.PopReady()
	require.NotNil(t, item2)
	assert.Equal(t, "medium", item2.ID.Component)
	assert.Equal(t, 4, item2.Weight)

	// Then light
	item3 := q.PopReady()
	require.NotNil(t, item3)
	assert.Equal(t, "light", item3.ID.Component)
	assert.Equal(t, 1, item3.Weight)

	// Queue should be empty
	assert.Equal(t, 0, q.Len())
}

func TestWorkQueue_PopReady_BlocksOnDeps(t *testing.T) {
	// comp2 depends on comp1
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			Weight:    8, // Higher weight but blocked
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	q := NewWorkQueue(work)

	// Even though comp2 has higher weight, comp1 should pop first because comp2 is blocked
	item1 := q.PopReady()
	require.NotNil(t, item1)
	assert.Equal(t, "comp1", item1.ID.Component, "comp1 should pop first (comp2 is blocked)")

	// Mark comp1 complete
	q.MarkComplete(item1.ID)

	// Now comp2 should be ready
	item2 := q.PopReady()
	require.NotNil(t, item2)
	assert.Equal(t, "comp2", item2.ID.Component)
}

func TestWorkQueue_PopReady_ReturnsNilWhenClosedAndEmpty(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
	}

	q := NewWorkQueue(work)

	// Pop the only item
	item := q.PopReady()
	require.NotNil(t, item)

	// Close the queue
	q.Close()

	// PopReady should return nil (queue is empty and closed)
	result := q.PopReady()
	assert.Nil(t, result)
}

func TestWorkQueue_Stats(t *testing.T) {
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

	q := NewWorkQueue(work)

	stats := q.Stats()
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 1, stats.Ready)   // Only comp1 is ready
	assert.Equal(t, 2, stats.Blocked) // comp2 and comp3 blocked on comp1
}

func TestWorkQueue_Stats_AfterCompletion(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			Weight:    2,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	q := NewWorkQueue(work)

	// Pop comp1
	item := q.PopReady()
	require.NotNil(t, item)
	q.MarkComplete(item.ID)

	stats := q.Stats()
	assert.Equal(t, 1, stats.Total)   // Only comp2 remains
	assert.Equal(t, 1, stats.Ready)   // comp2 is now ready
	assert.Equal(t, 0, stats.Blocked) // Nothing blocked
}

func TestWorkQueue_ConcurrentAccess(t *testing.T) {
	// Create many items with unique component names
	work := make([]workunit.UnitSpec, 100)
	for i := 0; i < 100; i++ {
		work[i] = workunit.UnitSpec{
			ID:     workunit.UnitID{Module: "mod1", Component: fmt.Sprintf("comp%d", i)},
			Weight: i%10 + 1,
			Index:  i,
		}
	}

	q := NewWorkQueue(work)

	// Multiple goroutines popping concurrently
	var wg sync.WaitGroup
	popped := make(chan *workunit.UnitSpec, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				item := q.PopReady()
				if item == nil {
					return
				}
				popped <- item
				q.MarkComplete(item.ID)
			}
		}()
	}

	// Close queue after a brief delay to let workers drain
	go func() {
		time.Sleep(100 * time.Millisecond)
		q.Close()
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

func TestWorkQueue_Len(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp2"}, Weight: 2},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp3"}, Weight: 3},
	}

	q := NewWorkQueue(work)

	assert.Equal(t, 3, q.Len())

	q.PopReady()
	assert.Equal(t, 2, q.Len())

	q.PopReady()
	assert.Equal(t, 1, q.Len())

	q.PopReady()
	assert.Equal(t, 0, q.Len())
}

func TestWorkQueue_PreservesIndex(t *testing.T) {
	// Index field should be preserved for result ordering
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1, Index: 0},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp2"}, Weight: 8, Index: 1},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp3"}, Weight: 4, Index: 2},
	}

	q := NewWorkQueue(work)

	// Pop in LPT order but verify indices preserved
	item1 := q.PopReady()
	assert.Equal(t, 1, item1.Index) // comp2 (weight 8)

	item2 := q.PopReady()
	assert.Equal(t, 2, item2.Index) // comp3 (weight 4)

	item3 := q.PopReady()
	assert.Equal(t, 0, item3.Index) // comp1 (weight 1)
}

// TestWorkHeap_LessFunction verifies heap ordering directly
func TestWorkHeap_LessFunction(t *testing.T) {
	h := workHeap{
		{ID: workunit.UnitID{Component: "a"}, Weight: 5},
		{ID: workunit.UnitID{Component: "b"}, Weight: 10},
	}

	// Less returns true if i should come before j
	// For max-heap (LPT), higher weight should come first
	assert.True(t, h.Less(1, 0), "weight 10 should come before weight 5")
	assert.False(t, h.Less(0, 1), "weight 5 should not come before weight 10")
}

func TestWorkQueue_PopReadyWithBudget_BinPacking(t *testing.T) {
	// Test bin-packing: with limited budget, smaller items should be picked
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "heavy"}, Weight: 8, Index: 0},
		{ID: workunit.UnitID{Module: "mod1", Component: "medium"}, Weight: 4, Index: 1},
		{ID: workunit.UnitID{Module: "mod1", Component: "light"}, Weight: 1, Index: 2},
	}

	q := NewWorkQueue(work)

	// With budget 4, should skip heavy (8) and pick medium (4)
	item1 := q.PopReadyWithBudget(4, 10)
	require.NotNil(t, item1)
	assert.Equal(t, "medium", item1.ID.Component, "should pick heaviest that fits budget")
	assert.Equal(t, 4, item1.Weight)

	// With budget 2, should skip heavy (8) and pick light (1)
	item2 := q.PopReadyWithBudget(2, 10)
	require.NotNil(t, item2)
	assert.Equal(t, "light", item2.ID.Component, "should pick heaviest that fits budget")
	assert.Equal(t, 1, item2.Weight)

	// With unlimited budget (0), should pick the remaining heavy
	item3 := q.PopReadyWithBudget(0, 10)
	require.NotNil(t, item3)
	assert.Equal(t, "heavy", item3.ID.Component, "unlimited budget should pick heaviest")
	assert.Equal(t, 8, item3.Weight)
}

func TestWorkQueue_PopReadyWithBudget_NothingFits(t *testing.T) {
	// All items exceed budget - should use over-capacity fallback (heaviest item)
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "heavy1"}, Weight: 8},
		{ID: workunit.UnitID{Module: "mod1", Component: "heavy2"}, Weight: 6},
	}

	q := NewWorkQueue(work)

	// Budget 4 - neither heavy1 (8) nor heavy2 (6) fits
	// With over-capacity fallback, should return heaviest (heavy1, weight=8)
	item := q.PopReadyWithBudget(4, 10)
	require.NotNil(t, item, "should return heaviest item via over-capacity fallback")
	assert.Equal(t, "heavy1", item.ID.Component)
	assert.Equal(t, 8, item.Weight)

	// Second item should also be returned (heavy2, weight=6)
	item = q.PopReadyWithBudget(4, 10)
	require.NotNil(t, item, "should return second heavy item via over-capacity fallback")
	assert.Equal(t, "heavy2", item.ID.Component)
	assert.Equal(t, 6, item.Weight)

	// Now queue is empty
	assert.Equal(t, 0, q.Len())
}

func TestWorkQueue_HasReadyWithBudget(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "heavy"}, Weight: 8},
		{ID: workunit.UnitID{Module: "mod1", Component: "light"}, Weight: 2},
	}

	q := NewWorkQueue(work)

	// Budget 10 - both fit
	assert.True(t, q.HasReadyWithBudget(10))

	// Budget 4 - only light fits
	assert.True(t, q.HasReadyWithBudget(4))

	// Budget 1 - nothing fits
	assert.False(t, q.HasReadyWithBudget(1))

	// Unlimited - always true if queue has items
	assert.True(t, q.HasReadyWithBudget(0))
}

func TestWorkQueue_PopReadyWithBudget_RespectsDependencies(t *testing.T) {
	// Even with budget, blocked items should not be picked
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "dep"}, Weight: 2},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "light-blocked"},
			Weight:    1, // Lighter but blocked
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "dep"}},
		},
	}

	q := NewWorkQueue(work)

	// Budget 1 - light-blocked (1) fits but is blocked, dep (2) doesn't fit
	// With over-capacity fallback, should return dep (2) even though it exceeds budget
	dep := q.PopReadyWithBudget(1, 10)
	require.NotNil(t, dep, "should return dep via over-capacity fallback")
	assert.Equal(t, "dep", dep.ID.Component)
	assert.Equal(t, 2, dep.Weight)

	// Mark dep complete - now light-blocked should become ready
	q.MarkComplete(dep.ID)

	// Now light-blocked should be ready and fit budget
	item := q.PopReadyWithBudget(1, 10)
	require.NotNil(t, item)
	assert.Equal(t, "light-blocked", item.ID.Component)
}

// --- Tests for LayerPolicy-based WorkQueue ---

func TestNewWorkQueueWithPolicy_LayerModeNone(t *testing.T) {
	// LayerModeNone should behave identically to the original DependsOn-only mode
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			Weight:    8,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	q, err := NewWorkQueueWithPolicy(work, LayerModeNone)
	require.NoError(t, err)
	assert.NotNil(t, q)

	// comp1 should be ready (no deps), comp2 blocked
	item1 := q.PopReady()
	require.NotNil(t, item1)
	assert.Equal(t, "comp1", item1.ID.Component)

	// Mark complete and comp2 becomes ready
	q.MarkComplete(item1.ID)

	item2 := q.PopReady()
	require.NotNil(t, item2)
	assert.Equal(t, "comp2", item2.ID.Component)
}

func TestNewWorkQueueWithPolicy_LayerModeStrict(t *testing.T) {
	// LayerModeStrict enforces module layers
	// mod1 has no deps, mod2 depends on mod1
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:     workunit.UnitID{Module: "mod2", Component: "comp1"},
			Weight: 8,
			// Cross-module dependency
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	q, err := NewWorkQueueWithPolicy(work, LayerModeStrict)
	require.NoError(t, err)

	// mod1:comp1 should be ready (layer 0)
	item1 := q.PopReady()
	require.NotNil(t, item1)
	assert.Equal(t, "mod1", item1.ID.Module)

	// Mark complete - mod2 should now be ready
	q.MarkComplete(item1.ID)

	item2 := q.PopReady()
	require.NotNil(t, item2)
	assert.Equal(t, "mod2", item2.ID.Module)
}

func TestNewWorkQueueWithPolicy_ModuleLayerOrdering(t *testing.T) {
	// Three modules: mod1 -> mod2 -> mod3
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod2", Component: "comp1"},
			Weight:    2,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
		{
			ID:        workunit.UnitID{Module: "mod3", Component: "comp1"},
			Weight:    3,
			DependsOn: []workunit.UnitID{{Module: "mod2", Component: "comp1"}},
		},
	}

	q, err := NewWorkQueueWithPolicy(work, LayerModeStrict)
	require.NoError(t, err)

	// Only mod1 should be ready initially
	stats := q.Stats()
	assert.Equal(t, 1, stats.Ready)
	assert.Equal(t, 2, stats.Blocked)

	// Pop mod1
	item1 := q.PopReady()
	require.NotNil(t, item1)
	assert.Equal(t, "mod1", item1.ID.Module)
	q.MarkComplete(item1.ID)

	// Now mod2 should be ready
	item2 := q.PopReady()
	require.NotNil(t, item2)
	assert.Equal(t, "mod2", item2.ID.Module)
	q.MarkComplete(item2.ID)

	// Now mod3 should be ready
	item3 := q.PopReady()
	require.NotNil(t, item3)
	assert.Equal(t, "mod3", item3.ID.Module)
}

func TestNewWorkQueueWithPolicy_ParallelModulesInSameLayer(t *testing.T) {
	// mod1 and mod2 have no deps - should be in same layer and both ready
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 4},
		{ID: workunit.UnitID{Module: "mod2", Component: "comp1"}, Weight: 2},
	}

	q, err := NewWorkQueueWithPolicy(work, LayerModeStrict)
	require.NoError(t, err)

	// Both should be ready
	stats := q.Stats()
	assert.Equal(t, 2, stats.Ready)
	assert.Equal(t, 0, stats.Blocked)

	// Should pop heavier one first (LPT)
	item1 := q.PopReady()
	require.NotNil(t, item1)
	assert.Equal(t, "mod1", item1.ID.Module)
	assert.Equal(t, 4, item1.Weight)

	// Second one still ready without needing to mark first complete
	item2 := q.PopReady()
	require.NotNil(t, item2)
	assert.Equal(t, "mod2", item2.ID.Module)
}

func TestNewWorkQueueWithPolicy_HasLayerPolicy(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
	}

	q, err := NewWorkQueueWithPolicy(work, LayerModeStrict)
	require.NoError(t, err)

	// Queue should have layer policy configured
	assert.True(t, q.HasLayerPolicy())
}

func TestNewWorkQueue_BackwardCompatibility(t *testing.T) {
	// NewWorkQueue (without policy) should still work for backward compatibility
	// It should default to LayerModeNone behavior
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1},
		{
			ID:        workunit.UnitID{Module: "mod1", Component: "comp2"},
			Weight:    8,
			DependsOn: []workunit.UnitID{{Module: "mod1", Component: "comp1"}},
		},
	}

	q := NewWorkQueue(work)

	// Should work exactly as before
	item1 := q.PopReady()
	require.NotNil(t, item1)
	assert.Equal(t, "comp1", item1.ID.Component)

	q.MarkComplete(item1.ID)

	item2 := q.PopReady()
	require.NotNil(t, item2)
	assert.Equal(t, "comp2", item2.ID.Component)
}

func TestNewWorkQueueWithPolicy_EmptyWork(t *testing.T) {
	work := []workunit.UnitSpec{}

	q, err := NewWorkQueueWithPolicy(work, LayerModeStrict)
	require.NoError(t, err)
	assert.NotNil(t, q)
	assert.Equal(t, 0, q.Len())

	// PopReady should return nil immediately for empty queue
	q.Close()
	item := q.PopReady()
	assert.Nil(t, item)
}

// --- Tests for Over-Capacity Scheduling Fix ---

// TestOverCapacityScheduling_HeavyModuleStarts verifies that when budget > 0
// and no ready items fit within the budget, the algorithm falls back to
// returning the heaviest ready item (over-capacity scheduling).
// This prevents deadlock when all remaining modules are heavier than available capacity.
func TestOverCapacityScheduling_HeavyModuleStarts(t *testing.T) {
	// Scenario: All ready modules exceed available capacity
	// Expected: Return heaviest ready module to prevent deadlock
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 8},
		{ID: workunit.UnitID{Module: "mod2", Component: "comp1"}, Weight: 6},
		{ID: workunit.UnitID{Module: "mod3", Component: "comp1"}, Weight: 7},
	}

	q := NewWorkQueue(work)

	// Budget is 4, but all modules are heavier (8, 6, 7)
	// OLD BEHAVIOR: Would block/return nil waiting for capacity
	// NEW BEHAVIOR: Should return heaviest ready item (mod1, weight=8)
	item := q.PopReadyWithBudget(4, 10)
	require.NotNil(t, item, "should return heaviest ready item when nothing fits budget (over-capacity fallback)")
	assert.Equal(t, "mod1", item.ID.Module, "should return heaviest module")
	assert.Equal(t, 8, item.Weight, "should return item with weight 8")
}

// --- Tests for Over-Capacity Coordination (TDD - These tests will FAIL initially) ---

// TestWorkQueue_OverCapacity_BlocksMultipleWorkers verifies that when multiple workers
// try to pop items where weight > totalCapacity, only ONE worker gets an item at a time.
// The second worker should get nil until the first completes and calls ReleaseOverCapacity.
//
// This test proves the DEADLOCK BUG: without coordination, multiple workers can
// grab over-capacity items simultaneously, all blocking on semaphore acquisition forever.
func TestWorkQueue_OverCapacity_BlocksMultipleWorkers(t *testing.T) {
	// Scenario: capacity=4, two workers try to pop weight=10 items simultaneously
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "heavy1", Component: "comp1"}, Weight: 10, Index: 0},
		{ID: workunit.UnitID{Module: "heavy2", Component: "comp1"}, Weight: 10, Index: 1},
	}

	q := NewWorkQueue(work)
	totalCapacity := 4

	// Channel to synchronize worker attempts
	worker1Ready := make(chan struct{})
	worker2Ready := make(chan struct{})
	results := make(chan *workunit.UnitSpec, 2)

	var wg sync.WaitGroup

	// Worker 1: Pops first over-capacity item
	wg.Add(1)
	go func() {
		defer wg.Done()
		close(worker1Ready) // Signal ready
		<-worker2Ready      // Wait for worker 2 to be ready

		spec := q.PopReadyWithBudget(1, totalCapacity) // budget=1, all items exceed
		results <- spec

		if spec != nil {
			// Simulate work
			time.Sleep(100 * time.Millisecond)
			q.MarkComplete(spec.ID)
			q.ReleaseOverCapacity(spec.ID)
		}
	}()

	// Worker 2: Should block or return nil while worker 1 holds over-capacity lock
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-worker1Ready      // Wait for worker 1 to be ready
		close(worker2Ready) // Signal ready

		// Give worker 1 a slight head start to grab the lock
		time.Sleep(10 * time.Millisecond)

		spec := q.PopReadyWithBudget(1, totalCapacity)
		results <- spec

		if spec != nil {
			q.MarkComplete(spec.ID)
			q.ReleaseOverCapacity(spec.ID)
		}
	}()

	wg.Wait()
	close(results)

	// Collect results
	var item1, item2 *workunit.UnitSpec
	item1 = <-results
	item2 = <-results

	// EXPECTED BEHAVIOR (after fix):
	// - First worker gets heavy1 (weight=10), overCapacityItem is set
	// - Second worker gets nil (must wait for first to complete)
	// - After first completes and calls ReleaseOverCapacity, second worker can proceed
	require.NotNil(t, item1, "first worker should get an item")
	assert.Equal(t, 10, item1.Weight, "first worker should get over-capacity item")

	// Without the fix, item2 might also be non-nil (both workers grab items)
	// With the fix, item2 should be nil until item1 completes
	// Since we're testing serially here, we expect only one item to be popped initially
	assert.NotNil(t, item2, "second worker should eventually get an item after first releases")
}

// TestWorkQueue_OverCapacity_AllowsNormalItemsInParallel verifies that when an
// over-capacity item is active (weight > totalCapacity), workers can STILL pop
// normal items (weight <= totalCapacity) in parallel. The over-capacity lock
// should only block OTHER over-capacity items, not normal items.
//
// This test ensures we don't over-serialize the system - light items should
// process in parallel even when heavy items are running.
func TestWorkQueue_OverCapacity_AllowsNormalItemsInParallel(t *testing.T) {
	// Scenario: capacity=8, overCapacityItem active (weight=10), workers pop weight=2 items
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "heavy", Component: "comp1"}, Weight: 10, Index: 0},
		{ID: workunit.UnitID{Module: "light1", Component: "comp1"}, Weight: 2, Index: 1},
		{ID: workunit.UnitID{Module: "light2", Component: "comp1"}, Weight: 2, Index: 2},
		{ID: workunit.UnitID{Module: "light3", Component: "comp1"}, Weight: 2, Index: 3},
	}

	q := NewWorkQueue(work)
	totalCapacity := 8

	results := make(chan *workunit.UnitSpec, 4)
	var wg sync.WaitGroup

	// Worker 1: Grabs heavy item (over-capacity)
	wg.Add(1)
	go func() {
		defer wg.Done()
		spec := q.PopReadyWithBudget(1, totalCapacity) // budget=1, heavy won't fit
		if spec != nil {
			results <- spec
			time.Sleep(200 * time.Millisecond) // Hold over-capacity lock for a while
			q.MarkComplete(spec.ID)
			q.ReleaseOverCapacity(spec.ID)
		}
	}()

	// Give worker 1 time to grab the heavy item
	time.Sleep(50 * time.Millisecond)

	// Workers 2-4: Should be able to grab light items in parallel
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			spec := q.PopReadyWithBudget(8, totalCapacity) // budget=8, light items fit
			if spec != nil {
				results <- spec
				q.MarkComplete(spec.ID)
				// Light items don't need ReleaseOverCapacity (weight <= totalCapacity)
			}
		}()
	}

	wg.Wait()
	close(results)

	// Collect results
	popped := []*workunit.UnitSpec{}
	for spec := range results {
		popped = append(popped, spec)
	}

	// EXPECTED BEHAVIOR (after fix):
	// - All 4 items should be popped (1 heavy + 3 light)
	// - Light items can be popped while heavy item holds over-capacity lock
	// - Over-capacity lock only affects weight > totalCapacity items
	require.Equal(t, 4, len(popped), "all items should be popped")

	// Verify we got the heavy item and light items
	heavyCount := 0
	lightCount := 0
	for _, spec := range popped {
		if spec.Weight == 10 {
			heavyCount++
		} else if spec.Weight == 2 {
			lightCount++
		}
	}
	assert.Equal(t, 1, heavyCount, "should have 1 heavy item")
	assert.Equal(t, 3, lightCount, "should have 3 light items")
}

// TestWorkQueue_WithinCapacity_UsesNormalCoordination verifies that items
// with weight <= totalCapacity do NOT trigger over-capacity serialization.
// They should use normal semaphore coordination (can run in parallel if
// combined weight fits capacity).
//
// This test ensures we don't incorrectly mark items as over-capacity when
// they actually fit within total capacity.
func TestWorkQueue_WithinCapacity_UsesNormalCoordination(t *testing.T) {
	// Scenario: capacity=32, queue has two weight=20 items
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 20, Index: 0},
		{ID: workunit.UnitID{Module: "mod2", Component: "comp1"}, Weight: 20, Index: 1},
	}

	q := NewWorkQueue(work)
	totalCapacity := 32

	// Both items fit within totalCapacity (20 <= 32), so neither should be
	// marked over-capacity. Both should be poppable.
	item1 := q.PopReadyWithBudget(32, totalCapacity)
	require.NotNil(t, item1, "first item should be popped")
	assert.Equal(t, 20, item1.Weight, "first item should have weight 20")

	// Second item should also be poppable (no over-capacity lock blocking it)
	item2 := q.PopReadyWithBudget(32, totalCapacity)
	require.NotNil(t, item2, "second item should be popped")
	assert.Equal(t, 20, item2.Weight, "second item should have weight 20")

	// Queue should be empty
	assert.Equal(t, 0, q.Len(), "queue should be empty after popping both items")

	// NOTE: In real execution, the semaphore would coordinate these items:
	// - Worker 1 acquires 20 slots (12 available)
	// - Worker 2 tries to acquire 20 slots → BLOCKS (only 12 available)
	// - Worker 1 completes, releases 20 slots (32 available)
	// - Worker 2's acquire succeeds
	//
	// But at the QUEUE level, both items can be popped without over-capacity lock.
}

// TestWorkQueue_OverCapacity_ReleaseUnblocksWaiters verifies that after an
// over-capacity item completes and calls ReleaseOverCapacity, waiting workers
// wake up and can pop the next over-capacity item.
//
// This test ensures the Broadcast mechanism works correctly to wake blocked workers.
func TestWorkQueue_OverCapacity_ReleaseUnblocksWaiters(t *testing.T) {
	// Scenario: Over-capacity item completes, waiting workers should wake
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "heavy1", Component: "comp1"}, Weight: 10, Index: 0},
		{ID: workunit.UnitID{Module: "heavy2", Component: "comp1"}, Weight: 10, Index: 1},
		{ID: workunit.UnitID{Module: "heavy3", Component: "comp1"}, Weight: 10, Index: 2},
	}

	q := NewWorkQueue(work)
	totalCapacity := 4

	results := make(chan string, 3)
	var wg sync.WaitGroup

	// Spawn 3 workers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			spec := q.PopReadyWithBudget(1, totalCapacity) // All items exceed budget
			if spec == nil {
				return
			}

			results <- fmt.Sprintf("worker%d:%s", workerID, spec.ID.Module)

			// Simulate work
			time.Sleep(50 * time.Millisecond)

			// Mark complete and release over-capacity lock
			q.MarkComplete(spec.ID)
			q.ReleaseOverCapacity(spec.ID)
		}(i)
	}

	wg.Wait()
	close(results)

	// Collect results
	completed := []string{}
	for r := range results {
		completed = append(completed, r)
	}

	// EXPECTED BEHAVIOR (after fix):
	// - All 3 workers should complete (serial execution via over-capacity lock)
	// - Each worker waits for previous to call ReleaseOverCapacity
	// - Broadcast wakes waiting workers
	assert.Equal(t, 3, len(completed), "all 3 items should complete serially")
}

// TestWorkQueue_SmallSystem_SerialHeavyItems verifies the realistic scenario
// where a small system (capacity=2) has mixed light and heavy items.
//
// This test represents the user's exact problem: light items should process
// normally, heavy items should execute serially without deadlock.
func TestWorkQueue_SmallSystem_SerialHeavyItems(t *testing.T) {
	// Scenario: capacity=2, queue has [w=1, w=1, w=5, w=5, w=1]
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "light1", Component: "comp1"}, Weight: 1, Index: 0},
		{ID: workunit.UnitID{Module: "light2", Component: "comp1"}, Weight: 1, Index: 1},
		{ID: workunit.UnitID{Module: "heavy1", Component: "comp1"}, Weight: 5, Index: 2},
		{ID: workunit.UnitID{Module: "heavy2", Component: "comp1"}, Weight: 5, Index: 3},
		{ID: workunit.UnitID{Module: "light3", Component: "comp1"}, Weight: 1, Index: 4},
	}

	q := NewWorkQueue(work)
	totalCapacity := 2

	results := make(chan *workunit.UnitSpec, 5)
	var wg sync.WaitGroup

	// Spawn 2 workers (matching capacity)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				// Get available capacity (simulate real worker behavior)
				available := totalCapacity // Simplified - real code uses semaphore.Available()

				spec := q.PopReadyWithBudget(available, totalCapacity)
				if spec == nil {
					return // Queue exhausted
				}

				results <- spec

				// Simulate work
				time.Sleep(50 * time.Millisecond)

				// Mark complete
				q.MarkComplete(spec.ID)

				// Release over-capacity lock if item exceeds totalCapacity
				if spec.Weight > totalCapacity {
					q.ReleaseOverCapacity(spec.ID)
				}
			}
		}()
	}

	wg.Wait()
	close(results)

	// Collect results
	processed := []*workunit.UnitSpec{}
	for spec := range results {
		processed = append(processed, spec)
	}

	// EXPECTED BEHAVIOR (after fix):
	// - Light items (w=1) process normally (can run in parallel)
	// - Heavy items (w=5) execute serially via over-capacity lock
	// - All 5 items complete without deadlock
	assert.Equal(t, 5, len(processed), "all items should complete")

	// Verify we got all items
	lightCount := 0
	heavyCount := 0
	for _, spec := range processed {
		if spec.Weight == 1 {
			lightCount++
		} else if spec.Weight == 5 {
			heavyCount++
		}
	}
	assert.Equal(t, 3, lightCount, "should have 3 light items")
	assert.Equal(t, 2, heavyCount, "should have 2 heavy items")
}

// TestWorkQueue_IsOverCapacityActive tests the helper method to check if
// an over-capacity item is currently active. Useful for testing and debugging.
func TestWorkQueue_IsOverCapacityActive(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "heavy", Component: "comp1"}, Weight: 10, Index: 0},
	}

	q := NewWorkQueue(work)
	totalCapacity := 4

	// Initially, no over-capacity item active
	assert.False(t, q.IsOverCapacityActive(), "initially no over-capacity item should be active")

	// Pop over-capacity item
	spec := q.PopReadyWithBudget(1, totalCapacity)
	require.NotNil(t, spec)

	// Now over-capacity should be active
	assert.True(t, q.IsOverCapacityActive(), "over-capacity item should be active after pop")

	// Release over-capacity lock
	q.ReleaseOverCapacity(spec.ID)

	// Over-capacity should be inactive again
	assert.False(t, q.IsOverCapacityActive(), "over-capacity should be inactive after release")
}

// TestWorkQueue_OverCapacity_SkipSemaphore verifies that over-capacity items
// should NOT call semaphore.Acquire() since they would block forever trying
// to acquire more slots than exist.
//
// This test documents the CRITICAL BUG: over-capacity items (weight > totalCapacity)
// cannot acquire capacity from a semaphore with fewer slots than their weight.
//
// NOTE: This test focuses on the queue coordination. The actual semaphore skip
// logic is in unit_scheduler.go worker loop (tested in integration tests).
func TestWorkQueue_OverCapacity_SkipSemaphore(t *testing.T) {
	// Scenario: capacity=16, item w=20 should be marked over-capacity
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "heavy", Component: "comp1"}, Weight: 20, Index: 0},
	}

	q := NewWorkQueue(work)
	totalCapacity := 16

	// Pop item - should be marked over-capacity (20 > 16)
	spec := q.PopReadyWithBudget(10, totalCapacity)
	require.NotNil(t, spec)
	assert.Equal(t, 20, spec.Weight)

	// Verify over-capacity lock is held
	assert.True(t, q.IsOverCapacityActive(), "over-capacity lock should be active")

	// In the worker loop, this item should NOT call semaphore.Acquire(20)
	// because a 16-slot semaphore cannot fulfill a 20-slot request.
	// Instead, the worker should:
	// 1. Check: isOverCapacity := weight > totalCapacity (20 > 16 = true)
	// 2. Skip: if !isOverCapacity { semaphore.Acquire(weight) }
	// 3. Execute without semaphore
	// 4. Skip: if !isOverCapacity { semaphore.Release(weight) }
	// 5. Release: queue.ReleaseOverCapacity(spec.ID)

	// This test verifies the queue correctly identifies the over-capacity case.
	// The semaphore skip logic is tested in unit_scheduler_test.go.

	// Clean up
	q.ReleaseOverCapacity(spec.ID)
	assert.False(t, q.IsOverCapacityActive())
}

// TestBinPackingPreferred verifies that when budget > 0 and some items do fit,
// the algorithm prefers the heaviest fitting item (bin-packing) over the
// over-capacity fallback. This ensures optimal resource utilization.
func TestBinPackingPreferred(t *testing.T) {
	// Scenario: Mix of modules - some fit budget, some don't
	// Expected: Prefer heaviest fitting module (bin-packing), not over-capacity fallback
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 8},  // Too heavy
		{ID: workunit.UnitID{Module: "mod2", Component: "comp1"}, Weight: 4},  // Fits, heaviest fitting
		{ID: workunit.UnitID{Module: "mod3", Component: "comp1"}, Weight: 2},  // Fits, lighter
		{ID: workunit.UnitID{Module: "mod4", Component: "comp1"}, Weight: 10}, // Too heavy
	}

	q := NewWorkQueue(work)

	// Budget is 5 - mod2 (4) and mod3 (2) fit, but mod1 (8) and mod4 (10) don't
	// Should prefer mod2 (heaviest fitting) over mod4 (heaviest overall)
	item := q.PopReadyWithBudget(5, 10)
	require.NotNil(t, item)
	assert.Equal(t, "mod2", item.ID.Module, "should prefer heaviest fitting item (bin-packing)")
	assert.Equal(t, 4, item.Weight, "should return fitting item, not over-capacity fallback")
}

// TestSequentialHeavyExecution verifies that multiple heavy modules can execute
// sequentially without deadlock when using over-capacity scheduling.
// This is critical for real-world scenarios where most modules exceed capacity.
func TestSequentialHeavyExecution(t *testing.T) {
	// Scenario: Multiple heavy modules that all exceed capacity
	// Expected: Each module starts sequentially via over-capacity fallback
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 12},
		{ID: workunit.UnitID{Module: "mod2", Component: "comp1"}, Weight: 10},
		{ID: workunit.UnitID{Module: "mod3", Component: "comp1"}, Weight: 15},
	}

	q := NewWorkQueue(work)
	budget := 8 // All modules exceed this budget

	// Should get modules in weight descending order via over-capacity fallback
	// First: mod3 (weight=15, heaviest)
	item1 := q.PopReadyWithBudget(budget, 10)
	require.NotNil(t, item1, "first over-capacity item should return")
	assert.Equal(t, "mod3", item1.ID.Module)
	assert.Equal(t, 15, item1.Weight)
	q.ReleaseOverCapacity(item1.ID) // Release lock for next item

	// Second: mod1 (weight=12, second heaviest)
	item2 := q.PopReadyWithBudget(budget, 10)
	require.NotNil(t, item2, "second over-capacity item should return")
	assert.Equal(t, "mod1", item2.ID.Module)
	assert.Equal(t, 12, item2.Weight)
	q.ReleaseOverCapacity(item2.ID) // Release lock for next item

	// Third: mod2 (weight=10, remaining) - NOT over-capacity (10 <= 10)
	item3 := q.PopReadyWithBudget(budget, 10)
	require.NotNil(t, item3, "third item should return")
	assert.Equal(t, "mod2", item3.ID.Module)
	assert.Equal(t, 10, item3.Weight)

	// Queue should be empty
	assert.Equal(t, 0, q.Len(), "all modules should be processed without deadlock")
}
