//go:build L1

package orchestrator

import (
	"sync"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/locktracker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWeightedSemaphore_RegistersWithRegistry(t *testing.T) {
	registry := locktracker.NewRegistry()
	ws := NewWeightedSemaphoreWithRegistry("test-semaphore", 8, registry)
	defer ws.Close()

	snapshot := registry.Snapshot()
	require.Len(t, snapshot, 1)

	var info locktracker.LockInfo
	for _, v := range snapshot {
		info = v
	}

	assert.Equal(t, locktracker.LockTypeWeighted, info.Type)
	assert.Equal(t, "test-semaphore", info.Name)
	assert.Equal(t, int64(8), info.Capacity)
	assert.Equal(t, int64(0), info.Used)
	assert.Equal(t, int64(0), info.Waiting)
}

func TestWeightedSemaphore_Acquire_UpdatesRegistry(t *testing.T) {
	registry := locktracker.NewRegistry()
	ws := NewWeightedSemaphoreWithRegistry("test-semaphore", 8, registry)
	defer ws.Close()

	ws.Acquire(3)

	snapshot := registry.Snapshot()
	require.Len(t, snapshot, 1)

	var info locktracker.LockInfo
	for _, v := range snapshot {
		info = v
	}

	assert.Equal(t, int64(3), info.Used)
	assert.Equal(t, int64(0), info.Waiting)
}

func TestWeightedSemaphore_Release_UpdatesRegistry(t *testing.T) {
	registry := locktracker.NewRegistry()
	ws := NewWeightedSemaphoreWithRegistry("test-semaphore", 8, registry)
	defer ws.Close()

	ws.Acquire(3)
	ws.Release(3)

	snapshot := registry.Snapshot()
	require.Len(t, snapshot, 1)

	var info locktracker.LockInfo
	for _, v := range snapshot {
		info = v
	}

	assert.Equal(t, int64(0), info.Used)
}

func TestWeightedSemaphore_WaitingCount(t *testing.T) {
	registry := locktracker.NewRegistry()
	ws := NewWeightedSemaphoreWithRegistry("test-semaphore", 2, registry)
	defer ws.Close()

	// Fill the semaphore
	ws.Acquire(2)

	// Start a goroutine that will block waiting
	waitStarted := make(chan struct{})
	waitDone := make(chan struct{})
	go func() {
		close(waitStarted)
		ws.Acquire(1) // Will block
		close(waitDone)
	}()

	// Wait for the goroutine to start waiting
	<-waitStarted
	time.Sleep(50 * time.Millisecond) // Give time for waiting count to increment

	// Check waiting count
	snapshot := registry.Snapshot()
	var info locktracker.LockInfo
	for _, v := range snapshot {
		info = v
	}
	assert.Equal(t, int64(1), info.Waiting, "waiting count should be 1")

	// Release to unblock
	ws.Release(1)
	<-waitDone

	// Waiting should now be 0
	snapshot = registry.Snapshot()
	for _, v := range snapshot {
		info = v
	}
	assert.Equal(t, int64(0), info.Waiting, "waiting count should be 0 after unblock")
}

func TestWeightedSemaphore_Close_UnregistersFromRegistry(t *testing.T) {
	registry := locktracker.NewRegistry()
	ws := NewWeightedSemaphoreWithRegistry("test-semaphore", 8, registry)

	snapshot := registry.Snapshot()
	require.Len(t, snapshot, 1)

	ws.Close()

	snapshot = registry.Snapshot()
	assert.Len(t, snapshot, 0, "registry should be empty after Close")
}

func TestWeightedSemaphore_Events(t *testing.T) {
	registry := locktracker.NewRegistry()

	// Subscribe to events
	eventCh := make(chan locktracker.LockEvent, 10)
	unsubscribe := registry.Subscribe(eventCh)
	defer unsubscribe()

	ws := NewWeightedSemaphoreWithRegistry("test-semaphore", 8, registry)

	// Should receive EventAcquired from registration
	select {
	case event := <-eventCh:
		assert.Equal(t, locktracker.EventAcquired, event.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected EventAcquired event")
	}

	// Acquire and expect capacity changed event
	ws.Acquire(2)
	select {
	case event := <-eventCh:
		assert.Equal(t, locktracker.EventCapacityChanged, event.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected EventCapacityChanged event after Acquire")
	}

	// Release and expect capacity changed event
	ws.Release(2)
	select {
	case event := <-eventCh:
		assert.Equal(t, locktracker.EventCapacityChanged, event.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected EventCapacityChanged event after Release")
	}

	// Close and expect released event
	ws.Close()
	select {
	case event := <-eventCh:
		assert.Equal(t, locktracker.EventReleased, event.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected EventReleased event after Close")
	}
}

func TestWeightedSemaphore_ConcurrentOperations(t *testing.T) {
	registry := locktracker.NewRegistry()
	ws := NewWeightedSemaphoreWithRegistry("test-semaphore", 10, registry)
	defer ws.Close()

	var wg sync.WaitGroup
	numGoroutines := 20
	opsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				ws.Acquire(1)
				// Simulate some work
				time.Sleep(time.Microsecond)
				ws.Release(1)
			}
		}()
	}

	wg.Wait()

	// After all operations, used should be 0
	snapshot := registry.Snapshot()
	var info locktracker.LockInfo
	for _, v := range snapshot {
		info = v
	}
	assert.Equal(t, int64(0), info.Used, "used should be 0 after all operations")
	assert.Equal(t, int64(0), info.Waiting, "waiting should be 0 after all operations")
}

func TestNewWeightedSemaphore_BackwardCompatible(t *testing.T) {
	// The original constructor should still work
	ws := NewWeightedSemaphore(8)

	ws.Acquire(3)
	assert.Equal(t, 3, ws.Used())

	ws.Release(3)
	assert.Equal(t, 0, ws.Used())
}

func TestWeightedSemaphore_SetCapacity_IncreasesCapacity(t *testing.T) {
	registry := locktracker.NewRegistry()
	ws := NewWeightedSemaphoreWithRegistry("test-semaphore", 4, registry)
	defer ws.Close()

	// Initial capacity should be 4
	assert.Equal(t, 4, ws.Capacity())

	// Increase capacity
	ws.SetCapacity(8)
	assert.Equal(t, 8, ws.Capacity())

	// Registry should reflect new capacity
	snapshot := registry.Snapshot()
	var info locktracker.LockInfo
	for _, v := range snapshot {
		info = v
	}
	assert.Equal(t, int64(8), info.Capacity)
}

func TestWeightedSemaphore_SetCapacity_DecreasesCapacity(t *testing.T) {
	registry := locktracker.NewRegistry()
	ws := NewWeightedSemaphoreWithRegistry("test-semaphore", 8, registry)
	defer ws.Close()

	// Decrease capacity
	ws.SetCapacity(4)
	assert.Equal(t, 4, ws.Capacity())

	// Registry should reflect new capacity
	snapshot := registry.Snapshot()
	var info locktracker.LockInfo
	for _, v := range snapshot {
		info = v
	}
	assert.Equal(t, int64(4), info.Capacity)
}

func TestWeightedSemaphore_SetCapacity_WakesWaiters(t *testing.T) {
	registry := locktracker.NewRegistry()
	ws := NewWeightedSemaphoreWithRegistry("test-semaphore", 2, registry)
	defer ws.Close()

	// Fill the semaphore
	ws.Acquire(2)

	// Start a goroutine that will block waiting
	acquired := make(chan struct{})
	go func() {
		ws.Acquire(1) // Will block until capacity increases
		close(acquired)
	}()

	// Give time for the goroutine to start waiting
	time.Sleep(50 * time.Millisecond)

	// Increase capacity - should wake up the waiter
	ws.SetCapacity(3)

	// The waiter should now be able to acquire
	select {
	case <-acquired:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("SetCapacity should have woken up waiting goroutine")
	}
}

func TestWeightedSemaphore_SetCapacity_MinimumCapacity(t *testing.T) {
	ws := NewWeightedSemaphore(8)

	// Try to set capacity to 0
	ws.SetCapacity(0)
	assert.Equal(t, 1, ws.Capacity(), "capacity should be at least 1")

	// Try to set negative capacity
	ws.SetCapacity(-5)
	assert.Equal(t, 1, ws.Capacity(), "capacity should be at least 1")
}

func TestWeightedSemaphore_SetCapacity_NeverBelowUsed(t *testing.T) {
	registry := locktracker.NewRegistry()
	ws := NewWeightedSemaphoreWithRegistry("test-semaphore", 8, registry)
	defer ws.Close()

	// Acquire 4 slots
	ws.Acquire(4)
	assert.Equal(t, 4, ws.Used())

	// Try to reduce capacity below used - should clamp to used
	ws.SetCapacity(2)
	assert.Equal(t, 4, ws.Capacity(), "capacity should not go below used (4)")

	// Registry should also show correct capacity
	snapshot := registry.Snapshot()
	var info locktracker.LockInfo
	for _, v := range snapshot {
		info = v
	}
	assert.Equal(t, int64(4), info.Capacity, "registry should show clamped capacity")
	assert.Equal(t, int64(4), info.Used, "registry should show used")
}
