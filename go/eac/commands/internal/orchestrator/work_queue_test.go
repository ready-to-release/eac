package orchestrator

import (
	"sync"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	// Create many items
	work := make([]workunit.UnitSpec, 100)
	for i := 0; i < 100; i++ {
		work[i] = workunit.UnitSpec{
			ID:     workunit.UnitID{Module: "mod1", Component: string(rune('a' + i%26))},
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
