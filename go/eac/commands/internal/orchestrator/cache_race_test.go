package orchestrator

import (
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTuiMarkCompleted_AlwaysIncrements verifies that tuiMarkCompleted always
// increments the counter, regardless of whether the item was early-cached.
// This tests the fix for the race condition where background cache detection
// and workers both tried to count items.
func TestTuiMarkCompleted_AlwaysIncrements(t *testing.T) {
	// Create minimal scheduler state for testing
	us := &UnitScheduler{
		tuiTotal:     3,
		tuiCompleted: 0,
		tuiRunning:   []string{},
	}

	// Test 1: Normal completion (not early cached)
	us.tuiMarkCompleted("mod1:comp1", 0)
	us.tuiMu.Lock()
	assert.Equal(t, 1, us.tuiCompleted, "counter should increment for normal completion")
	us.tuiMu.Unlock()

	// Test 2: Cached completion (early cached flag set)
	// After the fix, this should ALSO increment because workers are sole source of truth
	us.earlyCached.Store("mod1:comp2", EarlyCacheInfo{
		Module:    "mod1",
		Component: "comp2",
	})
	us.tuiMarkCompleted("mod1:comp2", -1) // -1 = cached
	us.tuiMu.Lock()
	assert.Equal(t, 2, us.tuiCompleted, "counter should increment even for early-cached items")
	us.tuiMu.Unlock()

	// Test 3: Another normal completion
	us.tuiMarkCompleted("mod1:comp3", 0)
	us.tuiMu.Lock()
	assert.Equal(t, 3, us.tuiCompleted, "counter should be 3 after all completions")
	us.tuiMu.Unlock()
}

// TestTuiMarkCompleted_RemovesFromRunning verifies that tuiMarkCompleted
// properly removes items from the running list.
func TestTuiMarkCompleted_RemovesFromRunning(t *testing.T) {
	us := &UnitScheduler{
		tuiTotal:     2,
		tuiCompleted: 0,
		tuiRunning:   []string{"mod1:comp1", "mod1:comp2"},
	}

	// Complete first item
	us.tuiMarkCompleted("mod1:comp1", 0)

	us.tuiMu.Lock()
	assert.Equal(t, 1, us.tuiCompleted)
	assert.Equal(t, []string{"mod1:comp2"}, us.tuiRunning, "comp1 should be removed from running")
	us.tuiMu.Unlock()

	// Complete second item
	us.tuiMarkCompleted("mod1:comp2", 0)

	us.tuiMu.Lock()
	assert.Equal(t, 2, us.tuiCompleted)
	assert.Empty(t, us.tuiRunning, "running list should be empty")
	us.tuiMu.Unlock()
}

// TestBackgroundCacheDetection_DoesNotIncrementCounter verifies that
// StartBackgroundCacheDetection does NOT increment tuiCompleted.
// Workers are the sole source of truth for completion counting.
func TestBackgroundCacheDetection_DoesNotIncrementCounter(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		WorkspaceRoot:  tmpDir,
		OutputBaseDir:  "out/test",
		LogFileName:    "test.log",
		ActionVerb:     "building",
		MaxConcurrency: 2,
	}

	// Create scheduler and immediately stop background goroutines
	us := NewUnitScheduler(config, nil, nil)
	us.StopCapacityTicker()
	defer us.Close()

	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "mod1", Component: "comp1"}, Weight: 1, Index: 0},
		{ID: workunit.UnitID{Module: "mod1", Component: "comp2"}, Weight: 1, Index: 1},
	}

	us.InitializeWork(work)

	// All items will be detected as cached
	cachedModules := map[string]bool{"mod1": true}
	cacheTimes := map[string]time.Time{"mod1": time.Now()}

	verifier := func(wsRoot string, spec workunit.UnitSpec, cached map[string]bool, times map[string]time.Time) (bool, time.Time) {
		return true, times["mod1"] // All items are cached
	}

	// Run background cache detection synchronously by waiting for it to complete
	var wg sync.WaitGroup
	wg.Add(1)

	// Start detection
	us.StartBackgroundCacheDetection(work, cachedModules, cacheTimes, verifier)

	// Wait a bit for the background goroutines to complete
	time.Sleep(150 * time.Millisecond)
	wg.Done()

	// Check that tuiCompleted was NOT incremented by background detection
	us.tuiMu.Lock()
	completed := us.tuiCompleted
	us.tuiMu.Unlock()

	// After fix: background detection should NOT increment counter
	assert.Equal(t, 0, completed, "background cache detection should NOT increment tuiCompleted")

	// Verify items were marked as early cached (visual state is fine)
	_, ok1 := us.earlyCached.Load("mod1:comp1")
	_, ok2 := us.earlyCached.Load("mod1:comp2")
	assert.True(t, ok1, "comp1 should be marked as early cached")
	assert.True(t, ok2, "comp2 should be marked as early cached")
}

// TestRunComponents_CounterMatchesTotal is an integration test that verifies
// the counter equals total after all work completes, including with cache detection.
// This test may trigger pre-existing races in infrastructure code when run with -race.
func TestRunComponents_CounterMatchesTotal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	config := &Config{
		WorkspaceRoot:  tmpDir,
		OutputBaseDir:  "out/test",
		LogFileName:    "test.log",
		ActionVerb:     "building",
		MaxConcurrency: 2,
	}

	us := NewUnitScheduler(config, nil, nil)
	us.StopCapacityTicker()
	defer us.Close()

	// Create 10 items, half of which will be "cached"
	work := make([]workunit.UnitSpec, 10)
	for i := 0; i < 10; i++ {
		work[i] = workunit.UnitSpec{
			ID:     workunit.UnitID{Module: "mod1", Component: string(rune('a' + i))},
			Weight: 1,
			Index:  i,
		}
	}

	us.InitializeWork(work)

	// Set up cache detection that marks even-indexed items as cached
	var detectionCount int32
	cachedModules := map[string]bool{"mod1": true}
	verifier := func(wsRoot string, spec workunit.UnitSpec, cached map[string]bool, times map[string]time.Time) (bool, time.Time) {
		atomic.AddInt32(&detectionCount, 1)
		// Small delay to increase chance of race with worker
		time.Sleep(5 * time.Millisecond)
		// Even indices are "cached"
		return spec.Index%2 == 0, time.Now()
	}

	us.SetCacheDetection(verifier, cachedModules)

	// Track how many workers actually executed vs short-circuited
	var executedCount int32
	worker := func(module, component string, logWriter io.Writer) int {
		atomic.AddInt32(&executedCount, 1)
		time.Sleep(10 * time.Millisecond) // Simulate work
		return 0
	}

	// Run components
	results := us.RunUnits(work, worker)

	// Verify all items processed
	require.Len(t, results, 10, "should have 10 results")

	// Verify counter matches total
	us.tuiMu.Lock()
	completed := us.tuiCompleted
	total := us.tuiTotal
	us.tuiMu.Unlock()

	assert.Equal(t, total, completed, "tuiCompleted should equal tuiTotal after all work done")
	assert.Equal(t, 10, completed, "all 10 items should be counted as completed")
}
