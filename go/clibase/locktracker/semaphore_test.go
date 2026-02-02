//go:build L1
// +build L1

package locktracker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewTrackedSemaphore Tests
// =============================================================================

// TestNewTrackedSemaphore_CreatesWithCorrectCapacity verifies constructor creates semaphore with capacity.
func TestNewTrackedSemaphore_CreatesWithCorrectCapacity(t *testing.T) {
	tests := []struct {
		name     string
		semName  string
		capacity int64
	}{
		{
			name:     "capacity of 1",
			semName:  "single-worker",
			capacity: 1,
		},
		{
			name:     "capacity of 5",
			semName:  "worker-pool",
			capacity: 5,
		},
		{
			name:     "capacity of 100",
			semName:  "large-pool",
			capacity: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sem := NewTrackedSemaphore(tt.semName, tt.capacity)
			require.NotNil(t, sem, "NewTrackedSemaphore should return non-nil semaphore")

			// Clean up
			defer sem.Close()

			capacity, used, waiting := sem.Stats()
			assert.Equal(t, tt.capacity, capacity, "Capacity should match")
			assert.Equal(t, int64(0), used, "Initial used should be 0")
			assert.Equal(t, int64(0), waiting, "Initial waiting should be 0")
		})
	}
}

// TestNewTrackedSemaphore_RegistersWithRegistry verifies semaphore is registered in the registry.
func TestNewTrackedSemaphore_RegistersWithRegistry(t *testing.T) {
	reg := NewRegistry()

	// Create semaphore with custom registry
	sem := NewTrackedSemaphoreWithRegistry("test-sem", 5, reg)
	require.NotNil(t, sem)
	defer sem.Close()

	// Verify it's registered
	snapshot := reg.Snapshot()
	assert.Equal(t, 1, len(snapshot), "Semaphore should be registered")

	// Verify lock info
	var lockInfo LockInfo
	for _, info := range snapshot {
		lockInfo = info
		break
	}

	assert.Equal(t, LockTypeSemaphore, lockInfo.Type)
	assert.Equal(t, "test-sem", lockInfo.Name)
	assert.Equal(t, int64(5), lockInfo.Capacity)
	assert.Equal(t, int64(0), lockInfo.Used)
	assert.Equal(t, int64(0), lockInfo.Waiting)
}

// TestNewTrackedSemaphore_UniqueIDs verifies each semaphore gets a unique ID.
func TestNewTrackedSemaphore_UniqueIDs(t *testing.T) {
	reg := NewRegistry()

	sem1 := NewTrackedSemaphoreWithRegistry("pool-1", 5, reg)
	sem2 := NewTrackedSemaphoreWithRegistry("pool-2", 10, reg)
	sem3 := NewTrackedSemaphoreWithRegistry("pool-1", 5, reg) // Same name, different ID

	defer sem1.Close()
	defer sem2.Close()
	defer sem3.Close()

	snapshot := reg.Snapshot()
	assert.Equal(t, 3, len(snapshot), "Should have 3 registered semaphores")

	// Verify all IDs are unique
	ids := make(map[string]bool)
	for id := range snapshot {
		if ids[id] {
			t.Errorf("Duplicate ID found: %s", id)
		}
		ids[id] = true
	}
}

// =============================================================================
// Acquire Tests
// =============================================================================

// TestTrackedSemaphore_Acquire_BlocksUntilSlotAvailable verifies Acquire blocks when full.
func TestTrackedSemaphore_Acquire_BlocksUntilSlotAvailable(t *testing.T) {
	sem := NewTrackedSemaphore("test", 1)
	defer sem.Close()

	ctx := context.Background()

	// Acquire the only slot
	err := sem.Acquire(ctx)
	require.NoError(t, err)

	// Try to acquire again - should block
	acquired := make(chan bool, 1)
	go func() {
		err := sem.Acquire(ctx)
		acquired <- (err == nil)
	}()

	// Give goroutine time to start blocking
	time.Sleep(50 * time.Millisecond)

	// Should not have acquired yet
	select {
	case <-acquired:
		t.Fatal("Acquire should block when semaphore is full")
	default:
		// Good, still blocking
	}

	// Release the slot
	sem.Release()

	// Now the goroutine should acquire
	select {
	case success := <-acquired:
		assert.True(t, success, "Acquire should succeed after release")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Acquire should unblock after release")
	}

	// Clean up
	sem.Release()
}

// TestTrackedSemaphore_Acquire_UpdatesRegistry verifies Acquire updates the registry.
func TestTrackedSemaphore_Acquire_UpdatesRegistry(t *testing.T) {
	reg := NewRegistry()
	sem := NewTrackedSemaphoreWithRegistry("test", 5, reg)
	defer sem.Close()

	ctx := context.Background()

	// Initial state
	_, used, _ := sem.Stats()
	assert.Equal(t, int64(0), used)

	// Acquire slots
	for i := 1; i <= 3; i++ {
		err := sem.Acquire(ctx)
		require.NoError(t, err)

		_, used, _ := sem.Stats()
		assert.Equal(t, int64(i), used, "Used should be %d after %d acquires", i, i)
	}

	// Verify registry is updated
	summary := reg.Summary()
	assert.Equal(t, int64(3), summary.TotalUsed)
}

// TestTrackedSemaphore_Acquire_ContextCancellation verifies Acquire respects context cancellation.
func TestTrackedSemaphore_Acquire_ContextCancellation(t *testing.T) {
	sem := NewTrackedSemaphore("test", 1)
	defer sem.Close()

	// Fill the semaphore
	ctx := context.Background()
	err := sem.Acquire(ctx)
	require.NoError(t, err)

	// Try to acquire with cancellable context
	cancelCtx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)
	go func() {
		errChan <- sem.Acquire(cancelCtx)
	}()

	// Give goroutine time to start waiting
	time.Sleep(50 * time.Millisecond)

	// Cancel the context
	cancel()

	// Should receive context.Canceled error
	select {
	case err := <-errChan:
		assert.ErrorIs(t, err, context.Canceled, "Acquire should return context.Canceled")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Acquire should return after context cancellation")
	}

	// Clean up
	sem.Release()
}

// TestTrackedSemaphore_Acquire_ContextTimeout verifies Acquire respects context timeout.
func TestTrackedSemaphore_Acquire_ContextTimeout(t *testing.T) {
	sem := NewTrackedSemaphore("test", 1)
	defer sem.Close()

	// Fill the semaphore
	ctx := context.Background()
	err := sem.Acquire(ctx)
	require.NoError(t, err)

	// Try to acquire with timeout context
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err = sem.Acquire(timeoutCtx)
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.DeadlineExceeded, "Acquire should return context.DeadlineExceeded")
	assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond, "Should wait at least until timeout")
	assert.Less(t, elapsed, 200*time.Millisecond, "Should not wait too long after timeout")

	// Clean up
	sem.Release()
}

// TestTrackedSemaphore_Acquire_MultipleAcquires verifies multiple acquires up to capacity.
func TestTrackedSemaphore_Acquire_MultipleAcquires(t *testing.T) {
	const capacity = 5
	sem := NewTrackedSemaphore("test", capacity)
	defer sem.Close()

	ctx := context.Background()

	// Acquire all slots
	for i := 0; i < capacity; i++ {
		err := sem.Acquire(ctx)
		require.NoError(t, err, "Acquire %d should succeed", i)
	}

	_, used, _ := sem.Stats()
	assert.Equal(t, int64(capacity), used, "All slots should be used")

	// Verify next acquire would block
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := sem.Acquire(timeoutCtx)
	assert.ErrorIs(t, err, context.DeadlineExceeded, "Acquire should timeout when full")

	// Clean up
	for i := 0; i < capacity; i++ {
		sem.Release()
	}
}

// =============================================================================
// Release Tests
// =============================================================================

// TestTrackedSemaphore_Release_FreesSlot verifies Release frees a slot.
func TestTrackedSemaphore_Release_FreesSlot(t *testing.T) {
	sem := NewTrackedSemaphore("test", 3)
	defer sem.Close()

	ctx := context.Background()

	// Acquire all slots
	for i := 0; i < 3; i++ {
		err := sem.Acquire(ctx)
		require.NoError(t, err)
	}

	_, used, _ := sem.Stats()
	assert.Equal(t, int64(3), used)

	// Release one slot
	sem.Release()

	_, used, _ = sem.Stats()
	assert.Equal(t, int64(2), used, "Used should decrease after release")
}

// TestTrackedSemaphore_Release_UpdatesRegistry verifies Release updates the registry.
func TestTrackedSemaphore_Release_UpdatesRegistry(t *testing.T) {
	reg := NewRegistry()
	sem := NewTrackedSemaphoreWithRegistry("test", 5, reg)
	defer sem.Close()

	ctx := context.Background()

	// Acquire some slots
	for i := 0; i < 3; i++ {
		err := sem.Acquire(ctx)
		require.NoError(t, err)
	}

	summary := reg.Summary()
	assert.Equal(t, int64(3), summary.TotalUsed)

	// Release all
	for i := 0; i < 3; i++ {
		sem.Release()
	}

	summary = reg.Summary()
	assert.Equal(t, int64(0), summary.TotalUsed)
}

// TestTrackedSemaphore_Release_UnblocksWaiting verifies Release unblocks waiting goroutines.
func TestTrackedSemaphore_Release_UnblocksWaiting(t *testing.T) {
	sem := NewTrackedSemaphore("test", 1)
	defer sem.Close()

	ctx := context.Background()

	// Acquire the slot
	err := sem.Acquire(ctx)
	require.NoError(t, err)

	// Start a goroutine that will wait
	acquired := make(chan struct{})
	go func() {
		err := sem.Acquire(ctx)
		if err == nil {
			close(acquired)
		}
	}()

	// Give time for goroutine to start waiting
	time.Sleep(50 * time.Millisecond)

	// Release
	sem.Release()

	// Goroutine should acquire
	select {
	case <-acquired:
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Waiting goroutine should be unblocked by release")
	}

	// Clean up
	sem.Release()
}

// =============================================================================
// TryAcquire Tests
// =============================================================================

// TestTrackedSemaphore_TryAcquire_ReturnsTrueWhenAvailable verifies TryAcquire returns true when slot available.
func TestTrackedSemaphore_TryAcquire_ReturnsTrueWhenAvailable(t *testing.T) {
	sem := NewTrackedSemaphore("test", 3)
	defer sem.Close()

	// Should succeed
	result := sem.TryAcquire()
	assert.True(t, result, "TryAcquire should return true when slots available")

	_, used, _ := sem.Stats()
	assert.Equal(t, int64(1), used)

	// Clean up
	sem.Release()
}

// TestTrackedSemaphore_TryAcquire_ReturnsFalseWhenFull verifies TryAcquire returns false when full.
func TestTrackedSemaphore_TryAcquire_ReturnsFalseWhenFull(t *testing.T) {
	sem := NewTrackedSemaphore("test", 1)
	defer sem.Close()

	// Fill the semaphore
	result := sem.TryAcquire()
	require.True(t, result)

	// Should fail
	result = sem.TryAcquire()
	assert.False(t, result, "TryAcquire should return false when semaphore is full")

	// Used should still be 1
	_, used, _ := sem.Stats()
	assert.Equal(t, int64(1), used)

	// Clean up
	sem.Release()
}

// TestTrackedSemaphore_TryAcquire_DoesNotBlock verifies TryAcquire returns immediately.
func TestTrackedSemaphore_TryAcquire_DoesNotBlock(t *testing.T) {
	sem := NewTrackedSemaphore("test", 1)
	defer sem.Close()

	// Fill the semaphore
	result := sem.TryAcquire()
	require.True(t, result)

	// TryAcquire should return immediately, not block
	start := time.Now()
	result = sem.TryAcquire()
	elapsed := time.Since(start)

	assert.False(t, result)
	assert.Less(t, elapsed, 10*time.Millisecond, "TryAcquire should return immediately")

	// Clean up
	sem.Release()
}

// TestTrackedSemaphore_TryAcquire_UpdatesRegistry verifies TryAcquire updates registry on success.
func TestTrackedSemaphore_TryAcquire_UpdatesRegistry(t *testing.T) {
	reg := NewRegistry()
	sem := NewTrackedSemaphoreWithRegistry("test", 3, reg)
	defer sem.Close()

	// TryAcquire multiple times
	for i := 1; i <= 3; i++ {
		result := sem.TryAcquire()
		assert.True(t, result)

		summary := reg.Summary()
		assert.Equal(t, int64(i), summary.TotalUsed)
	}

	// Should fail now
	result := sem.TryAcquire()
	assert.False(t, result)

	// Used should still be 3
	summary := reg.Summary()
	assert.Equal(t, int64(3), summary.TotalUsed)

	// Clean up
	for i := 0; i < 3; i++ {
		sem.Release()
	}
}

// =============================================================================
// Stats Tests
// =============================================================================

// TestTrackedSemaphore_Stats_ReturnsCorrectValues verifies Stats returns correct values.
func TestTrackedSemaphore_Stats_ReturnsCorrectValues(t *testing.T) {
	sem := NewTrackedSemaphore("test", 5)
	defer sem.Close()

	ctx := context.Background()

	// Initial state
	capacity, used, waiting := sem.Stats()
	assert.Equal(t, int64(5), capacity)
	assert.Equal(t, int64(0), used)
	assert.Equal(t, int64(0), waiting)

	// After some acquires
	for i := 0; i < 3; i++ {
		err := sem.Acquire(ctx)
		require.NoError(t, err)
	}

	capacity, used, waiting = sem.Stats()
	assert.Equal(t, int64(5), capacity)
	assert.Equal(t, int64(3), used)
	assert.Equal(t, int64(0), waiting)

	// Clean up
	for i := 0; i < 3; i++ {
		sem.Release()
	}
}

// TestTrackedSemaphore_Stats_WaitingCount verifies waiting count is tracked.
func TestTrackedSemaphore_Stats_WaitingCount(t *testing.T) {
	sem := NewTrackedSemaphore("test", 1)
	defer sem.Close()

	ctx := context.Background()

	// Fill the semaphore
	err := sem.Acquire(ctx)
	require.NoError(t, err)

	// Start multiple goroutines waiting
	const numWaiters = 3
	started := make(chan struct{}, numWaiters)
	acquired := make(chan struct{}, numWaiters)

	for i := 0; i < numWaiters; i++ {
		go func() {
			started <- struct{}{}
			err := sem.Acquire(ctx)
			if err == nil {
				acquired <- struct{}{}
				// Hold for a bit then release
				time.Sleep(10 * time.Millisecond)
				sem.Release()
			}
		}()
	}

	// Wait for all goroutines to start
	for i := 0; i < numWaiters; i++ {
		<-started
	}

	// Give time for goroutines to enter waiting state
	time.Sleep(50 * time.Millisecond)

	// Check waiting count
	_, _, waiting := sem.Stats()
	assert.GreaterOrEqual(t, waiting, int64(1), "Should have at least 1 waiter")
	assert.LessOrEqual(t, waiting, int64(numWaiters), "Should have at most %d waiters", numWaiters)

	// Release and let waiters through
	sem.Release()

	// Wait for all to complete
	for i := 0; i < numWaiters; i++ {
		select {
		case <-acquired:
			// Good
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Timeout waiting for goroutines to complete")
		}
	}
}

// =============================================================================
// Close Tests
// =============================================================================

// TestTrackedSemaphore_Close_UnregistersFromRegistry verifies Close unregisters the semaphore.
func TestTrackedSemaphore_Close_UnregistersFromRegistry(t *testing.T) {
	reg := NewRegistry()
	sem := NewTrackedSemaphoreWithRegistry("test", 5, reg)

	// Verify it's registered
	assert.Equal(t, 1, len(reg.Snapshot()))

	// Close
	sem.Close()

	// Verify it's unregistered
	assert.Equal(t, 0, len(reg.Snapshot()), "Semaphore should be unregistered after Close")
}

// TestTrackedSemaphore_Close_TriggersReleasedEvent verifies Close triggers EventReleased.
func TestTrackedSemaphore_Close_TriggersReleasedEvent(t *testing.T) {
	reg := NewRegistry()
	sem := NewTrackedSemaphoreWithRegistry("test", 5, reg)

	eventChan := make(chan LockEvent, 10)
	unsubscribe := reg.Subscribe(eventChan)
	defer unsubscribe()

	// Drain the EventAcquired from registration
	select {
	case <-eventChan:
	default:
	}

	// Close
	sem.Close()

	// Verify EventReleased was received
	select {
	case event := <-eventChan:
		assert.Equal(t, EventReleased, event.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected EventReleased after Close")
	}
}

// TestTrackedSemaphore_Close_Idempotent verifies Close can be called multiple times safely.
func TestTrackedSemaphore_Close_Idempotent(t *testing.T) {
	sem := NewTrackedSemaphore("test", 5)

	// Multiple closes should not panic
	sem.Close()
	sem.Close()
	sem.Close()
}

// =============================================================================
// Concurrent Operations Tests
// =============================================================================

// TestTrackedSemaphore_ConcurrentAcquireRelease verifies concurrent acquire/release operations.
func TestTrackedSemaphore_ConcurrentAcquireRelease(t *testing.T) {
	const capacity = 10
	const numGoroutines = 50
	const operationsPerGoroutine = 100

	sem := NewTrackedSemaphore("test", capacity)
	defer sem.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	var opsCompleted int64

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				err := sem.Acquire(ctx)
				if err != nil {
					continue
				}

				// Small delay to increase contention
				time.Sleep(time.Microsecond)

				sem.Release()
				atomic.AddInt64(&opsCompleted, 1)
			}
		}()
	}

	wg.Wait()

	// Verify all operations completed (none blocked indefinitely)
	assert.Equal(t, int64(numGoroutines*operationsPerGoroutine), opsCompleted,
		"All operations should complete")

	// Verify semaphore is back to initial state
	_, used, waiting := sem.Stats()
	assert.Equal(t, int64(0), used, "Used should be 0 after all operations")
	assert.Equal(t, int64(0), waiting, "Waiting should be 0 after all operations")
}

// TestTrackedSemaphore_ConcurrentTryAcquire verifies concurrent TryAcquire operations.
func TestTrackedSemaphore_ConcurrentTryAcquire(t *testing.T) {
	const capacity = 5
	const numGoroutines = 20

	sem := NewTrackedSemaphore("test", capacity)
	defer sem.Close()

	var wg sync.WaitGroup
	successCount := int64(0)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if sem.TryAcquire() {
				atomic.AddInt64(&successCount, 1)
				time.Sleep(50 * time.Millisecond)
				sem.Release()
			}
		}()
	}

	wg.Wait()

	// At least some should succeed, at most capacity at a time
	assert.Greater(t, successCount, int64(0), "Some TryAcquire should succeed")
	assert.GreaterOrEqual(t, successCount, int64(capacity), "At least capacity goroutines should succeed")
}

// TestTrackedSemaphore_ConcurrentStats verifies Stats is thread-safe.
func TestTrackedSemaphore_ConcurrentStats(t *testing.T) {
	const capacity = 5
	const numGoroutines = 20

	sem := NewTrackedSemaphore("test", capacity)
	defer sem.Close()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Start some background operations
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < 50; j++ {
				if sem.TryAcquire() {
					time.Sleep(time.Microsecond)
					sem.Release()
				}
			}
		}()
	}

	// Concurrent Stats calls
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				capacity, used, waiting := sem.Stats()
				// Just verify values are reasonable
				assert.GreaterOrEqual(t, capacity, int64(0))
				assert.GreaterOrEqual(t, used, int64(0))
				assert.LessOrEqual(t, used, capacity)
				assert.GreaterOrEqual(t, waiting, int64(0))
			}
		}()
	}

	// Concurrent Acquire with context
	for i := 0; i < numGoroutines/4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
			defer cancel()

			if sem.Acquire(timeoutCtx) == nil {
				sem.Release()
			}
		}()
	}

	wg.Wait()
}

// TestTrackedSemaphore_WaitingIncrementsWhileBlocked verifies waiting count during blocking.
func TestTrackedSemaphore_WaitingIncrementsWhileBlocked(t *testing.T) {
	sem := NewTrackedSemaphore("test", 1)
	defer sem.Close()

	ctx := context.Background()

	// Fill the semaphore
	err := sem.Acquire(ctx)
	require.NoError(t, err)

	// Start a goroutine that will block
	started := make(chan struct{})
	go func() {
		close(started)
		// This will block
		timeoutCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()
		sem.Acquire(timeoutCtx)
	}()

	// Wait for goroutine to start
	<-started

	// Give time to enter waiting state
	time.Sleep(50 * time.Millisecond)

	// Check waiting count
	_, _, waiting := sem.Stats()
	assert.GreaterOrEqual(t, waiting, int64(1), "Waiting count should be at least 1 while blocked")

	// Clean up - release the held slot
	sem.Release()

	// Give time for the blocked goroutine to complete
	time.Sleep(100 * time.Millisecond)
}

// TestTrackedSemaphore_WaitingDecrementsWhenAcquired verifies waiting count decreases after acquire.
func TestTrackedSemaphore_WaitingDecrementsWhenAcquired(t *testing.T) {
	sem := NewTrackedSemaphore("test", 1)
	defer sem.Close()

	ctx := context.Background()

	// Fill the semaphore
	err := sem.Acquire(ctx)
	require.NoError(t, err)

	// Start a goroutine that will block then acquire
	acquired := make(chan struct{})
	go func() {
		err := sem.Acquire(ctx)
		if err == nil {
			close(acquired)
			// Hold for a bit then release
			time.Sleep(10 * time.Millisecond)
			sem.Release()
		}
	}()

	// Give time to enter waiting state
	time.Sleep(50 * time.Millisecond)

	// Release to allow acquire
	sem.Release()

	// Wait for acquire to complete
	select {
	case <-acquired:
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Goroutine should have acquired")
	}

	// Give time for waiting count to update
	time.Sleep(20 * time.Millisecond)

	// Waiting count should be 0 now
	_, _, waiting := sem.Stats()
	assert.Equal(t, int64(0), waiting, "Waiting should be 0 after acquire completes")
}

// TestTrackedSemaphore_StressTest verifies semaphore under high stress.
func TestTrackedSemaphore_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const capacity = 10
	const numGoroutines = 100
	const operationsPerGoroutine = 500

	sem := NewTrackedSemaphore("stress-test", capacity)
	defer sem.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	var totalAcquires int64
	var totalReleases int64

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Mix of Acquire and TryAcquire
				if j%3 == 0 {
					if sem.TryAcquire() {
						atomic.AddInt64(&totalAcquires, 1)
						atomic.AddInt64(&totalReleases, 1)
						sem.Release()
					}
				} else {
					timeoutCtx, cancel := context.WithTimeout(ctx, time.Millisecond)
					if sem.Acquire(timeoutCtx) == nil {
						atomic.AddInt64(&totalAcquires, 1)
						atomic.AddInt64(&totalReleases, 1)
						sem.Release()
					}
					cancel()
				}
			}
		}()
	}

	wg.Wait()

	// Verify invariants
	_, used, waiting := sem.Stats()
	assert.Equal(t, int64(0), used, "Used should be 0 after stress test")
	assert.Equal(t, int64(0), waiting, "Waiting should be 0 after stress test")
	assert.Equal(t, totalAcquires, totalReleases, "Acquires should equal releases")

	t.Logf("Total successful operations: %d", totalAcquires)
}

// TestTrackedSemaphore_RaceDetector verifies no race conditions (run with -race flag).
func TestTrackedSemaphore_RaceDetector(t *testing.T) {
	sem := NewTrackedSemaphore("race-test", 5)
	defer sem.Close()

	ctx := context.Background()
	var wg sync.WaitGroup

	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire
			if sem.TryAcquire() {
				// Stats
				_, _, _ = sem.Stats()

				// Release
				sem.Release()
			}

			// Blocking acquire
			timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
			defer cancel()

			if sem.Acquire(timeoutCtx) == nil {
				// Stats again
				_, _, _ = sem.Stats()

				sem.Release()
			}
		}()
	}

	wg.Wait()
}
