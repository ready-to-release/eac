//go:build L1
// +build L1

package locktracker

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewRegistry Tests
// =============================================================================

// TestNewRegistry_CreatesEmptyRegistry verifies that NewRegistry creates an empty registry.
func TestNewRegistry_CreatesEmptyRegistry(t *testing.T) {
	reg := NewRegistry()

	require.NotNil(t, reg, "NewRegistry should return non-nil registry")

	// Verify it's empty
	snapshot := reg.Snapshot()
	assert.Equal(t, 0, len(snapshot), "New registry should have no locks")

	summary := reg.Summary()
	assert.Equal(t, 0, summary.Total, "New registry should have Total of 0")
	assert.Equal(t, int64(0), summary.TotalCapacity, "New registry should have TotalCapacity of 0")
	assert.Equal(t, int64(0), summary.TotalUsed, "New registry should have TotalUsed of 0")
	assert.Equal(t, int64(0), summary.TotalWaiting, "New registry should have TotalWaiting of 0")
}

// TestNewRegistry_MultipleInstances verifies that NewRegistry creates independent instances.
func TestNewRegistry_MultipleInstances(t *testing.T) {
	reg1 := NewRegistry()
	reg2 := NewRegistry()

	require.NotNil(t, reg1)
	require.NotNil(t, reg2)

	// Add lock to reg1 only
	lockInfo := LockInfo{
		ID:       "test-lock",
		Type:     LockTypeMutex,
		Name:     "test",
		Capacity: 1,
		Used:     1,
	}
	reg1.Register(&lockInfo)

	// Verify reg1 has the lock
	assert.Equal(t, 1, len(reg1.Snapshot()), "reg1 should have 1 lock")

	// Verify reg2 is still empty
	assert.Equal(t, 0, len(reg2.Snapshot()), "reg2 should have 0 locks")
}

// =============================================================================
// Register Tests
// =============================================================================

// TestRegistry_Register_AddsLock verifies that Register adds a lock to the registry.
func TestRegistry_Register_AddsLock(t *testing.T) {
	reg := NewRegistry()

	lockInfo := LockInfo{
		ID:         "mutex-1",
		Type:       LockTypeMutex,
		Name:       "config-mutex",
		AcquiredAt: time.Now(),
		Holder:     "goroutine-1",
		Capacity:   1,
		Used:       1,
		Waiting:    0,
	}

	reg.Register(&lockInfo)

	snapshot := reg.Snapshot()
	assert.Equal(t, 1, len(snapshot), "Registry should have 1 lock after Register")

	// Verify the lock info is correct
	registered, exists := snapshot["mutex-1"]
	require.True(t, exists, "Lock should exist in snapshot")
	assert.Equal(t, lockInfo.ID, registered.ID)
	assert.Equal(t, lockInfo.Type, registered.Type)
	assert.Equal(t, lockInfo.Name, registered.Name)
	assert.Equal(t, lockInfo.Holder, registered.Holder)
	assert.Equal(t, lockInfo.Capacity, registered.Capacity)
	assert.Equal(t, lockInfo.Used, registered.Used)
}

// TestRegistry_Register_TriggersEventAcquired verifies that Register triggers an EventAcquired notification.
func TestRegistry_Register_TriggersEventAcquired(t *testing.T) {
	reg := NewRegistry()

	// Subscribe to events
	eventChan := make(chan LockEvent, 10)
	unsubscribe := reg.Subscribe(eventChan)
	defer unsubscribe()

	lockInfo := LockInfo{
		ID:       "sem-1",
		Type:     LockTypeSemaphore,
		Name:     "worker-pool",
		Capacity: 5,
		Used:     1,
	}

	reg.Register(&lockInfo)

	// Verify event was received
	select {
	case event := <-eventChan:
		assert.Equal(t, EventAcquired, event.Type, "Event type should be EventAcquired")
		assert.Equal(t, lockInfo.ID, event.Lock.ID, "Event should contain the registered lock")
		assert.False(t, event.Timestamp.IsZero(), "Event should have a timestamp")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected EventAcquired event, but none received")
	}
}

// TestRegistry_Register_MultipleLocks verifies that multiple locks can be registered.
func TestRegistry_Register_MultipleLocks(t *testing.T) {
	reg := NewRegistry()

	locks := []LockInfo{
		{ID: "lock-1", Type: LockTypeMutex, Name: "mutex-1", Capacity: 1, Used: 1},
		{ID: "lock-2", Type: LockTypeSemaphore, Name: "sem-1", Capacity: 5, Used: 3},
		{ID: "lock-3", Type: LockTypeFileLock, Name: "file-1", Capacity: 1, Used: 1},
		{ID: "lock-4", Type: LockTypeRWMutex, Name: "rwmutex-1", Capacity: 1, Used: 1},
		{ID: "lock-5", Type: LockTypeWeighted, Name: "weighted-1", Capacity: 10, Used: 7},
	}

	for _, lock := range locks {
		reg.Register(&lock)
	}

	snapshot := reg.Snapshot()
	assert.Equal(t, 5, len(snapshot), "Registry should have 5 locks")

	// Verify all locks are present
	for _, lock := range locks {
		registered, exists := snapshot[lock.ID]
		require.True(t, exists, "Lock %s should exist", lock.ID)
		assert.Equal(t, lock.Type, registered.Type)
		assert.Equal(t, lock.Name, registered.Name)
	}
}

// TestRegistry_Register_UpdatesExisting verifies that registering with the same ID updates the lock.
func TestRegistry_Register_UpdatesExisting(t *testing.T) {
	reg := NewRegistry()

	// Register initial lock
	lockInfo := LockInfo{
		ID:       "lock-1",
		Type:     LockTypeSemaphore,
		Name:     "pool",
		Capacity: 5,
		Used:     2,
		Waiting:  0,
	}
	reg.Register(&lockInfo)

	// Update with same ID
	updated := LockInfo{
		ID:       "lock-1",
		Type:     LockTypeSemaphore,
		Name:     "pool",
		Capacity: 5,
		Used:     4,
		Waiting:  1,
	}
	reg.Register(&updated)

	snapshot := reg.Snapshot()
	assert.Equal(t, 1, len(snapshot), "Registry should still have 1 lock")

	registered := snapshot["lock-1"]
	assert.Equal(t, int64(4), registered.Used, "Used should be updated")
	assert.Equal(t, int64(1), registered.Waiting, "Waiting should be updated")
}

// =============================================================================
// Unregister Tests
// =============================================================================

// TestRegistry_Unregister_RemovesLock verifies that Unregister removes a lock from the registry.
func TestRegistry_Unregister_RemovesLock(t *testing.T) {
	reg := NewRegistry()

	lockInfo := LockInfo{
		ID:       "lock-1",
		Type:     LockTypeMutex,
		Name:     "test",
		Capacity: 1,
		Used:     1,
	}
	reg.Register(&lockInfo)

	// Verify lock exists
	assert.Equal(t, 1, len(reg.Snapshot()))

	// Unregister
	reg.Unregister("lock-1")

	// Verify lock is removed
	snapshot := reg.Snapshot()
	assert.Equal(t, 0, len(snapshot), "Registry should have 0 locks after Unregister")
	_, exists := snapshot["lock-1"]
	assert.False(t, exists, "Lock should not exist after Unregister")
}

// TestRegistry_Unregister_TriggersEventReleased verifies that Unregister triggers an EventReleased notification.
func TestRegistry_Unregister_TriggersEventReleased(t *testing.T) {
	reg := NewRegistry()

	lockInfo := LockInfo{
		ID:       "lock-1",
		Type:     LockTypeMutex,
		Name:     "test",
		Capacity: 1,
		Used:     1,
	}
	reg.Register(&lockInfo)

	// Subscribe after registration to only capture release event
	eventChan := make(chan LockEvent, 10)
	unsubscribe := reg.Subscribe(eventChan)
	defer unsubscribe()

	reg.Unregister("lock-1")

	// Verify event was received
	select {
	case event := <-eventChan:
		assert.Equal(t, EventReleased, event.Type, "Event type should be EventReleased")
		assert.Equal(t, lockInfo.ID, event.Lock.ID, "Event should contain the unregistered lock")
		assert.False(t, event.Timestamp.IsZero(), "Event should have a timestamp")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected EventReleased event, but none received")
	}
}

// TestRegistry_Unregister_NonexistentLock verifies that Unregister handles nonexistent locks gracefully.
func TestRegistry_Unregister_NonexistentLock(t *testing.T) {
	reg := NewRegistry()

	// Should not panic
	reg.Unregister("nonexistent-lock")

	// Registry should still be empty
	assert.Equal(t, 0, len(reg.Snapshot()))
}

// TestRegistry_Unregister_MultipleLocks verifies that multiple locks can be unregistered.
func TestRegistry_Unregister_MultipleLocks(t *testing.T) {
	reg := NewRegistry()

	// Register multiple locks
	for i := 1; i <= 5; i++ {
		reg.Register(&LockInfo{
			ID:       "lock-" + string(rune('0'+i)),
			Type:     LockTypeMutex,
			Capacity: 1,
			Used:     1,
		})
	}

	assert.Equal(t, 5, len(reg.Snapshot()))

	// Unregister some
	reg.Unregister("lock-1")
	reg.Unregister("lock-3")
	reg.Unregister("lock-5")

	snapshot := reg.Snapshot()
	assert.Equal(t, 2, len(snapshot), "Registry should have 2 locks remaining")

	_, exists1 := snapshot["lock-1"]
	_, exists2 := snapshot["lock-2"]
	_, exists3 := snapshot["lock-3"]
	_, exists4 := snapshot["lock-4"]
	_, exists5 := snapshot["lock-5"]

	assert.False(t, exists1)
	assert.True(t, exists2)
	assert.False(t, exists3)
	assert.True(t, exists4)
	assert.False(t, exists5)
}

// =============================================================================
// UpdateSemaphore Tests
// =============================================================================

// TestRegistry_UpdateSemaphore_UpdatesCounts verifies that UpdateSemaphore updates used/waiting counts.
func TestRegistry_UpdateSemaphore_UpdatesCounts(t *testing.T) {
	reg := NewRegistry()

	// Register a semaphore
	lockInfo := LockInfo{
		ID:       "sem-1",
		Type:     LockTypeSemaphore,
		Name:     "worker-pool",
		Capacity: 10,
		Used:     0,
		Waiting:  0,
	}
	reg.Register(&lockInfo)

	// Update counts
	reg.UpdateSemaphore("sem-1", 5, 2)

	snapshot := reg.Snapshot()
	sem := snapshot["sem-1"]
	assert.Equal(t, int64(5), sem.Used, "Used should be updated to 5")
	assert.Equal(t, int64(2), sem.Waiting, "Waiting should be updated to 2")
}

// TestRegistry_UpdateSemaphore_TriggersCapacityChanged verifies that UpdateSemaphore triggers EventCapacityChanged.
func TestRegistry_UpdateSemaphore_TriggersCapacityChanged(t *testing.T) {
	reg := NewRegistry()

	lockInfo := LockInfo{
		ID:       "sem-1",
		Type:     LockTypeSemaphore,
		Name:     "pool",
		Capacity: 5,
		Used:     1,
		Waiting:  0,
	}
	reg.Register(&lockInfo)

	// Subscribe after registration
	eventChan := make(chan LockEvent, 10)
	unsubscribe := reg.Subscribe(eventChan)
	defer unsubscribe()

	reg.UpdateSemaphore("sem-1", 3, 1)

	// Verify event was received
	select {
	case event := <-eventChan:
		assert.Equal(t, EventCapacityChanged, event.Type, "Event type should be EventCapacityChanged")
		assert.Equal(t, "sem-1", event.Lock.ID, "Event should contain the updated lock")
		assert.Equal(t, int64(3), event.Lock.Used, "Event should have updated Used")
		assert.Equal(t, int64(1), event.Lock.Waiting, "Event should have updated Waiting")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected EventCapacityChanged event, but none received")
	}
}

// TestRegistry_UpdateSemaphore_NonexistentLock verifies behavior for nonexistent lock.
func TestRegistry_UpdateSemaphore_NonexistentLock(t *testing.T) {
	reg := NewRegistry()

	// Should not panic
	reg.UpdateSemaphore("nonexistent", 5, 2)

	// Registry should still be empty
	assert.Equal(t, 0, len(reg.Snapshot()))
}

// TestRegistry_UpdateSemaphore_MultipleUpdates verifies multiple updates work correctly.
func TestRegistry_UpdateSemaphore_MultipleUpdates(t *testing.T) {
	reg := NewRegistry()

	lockInfo := LockInfo{
		ID:       "sem-1",
		Type:     LockTypeSemaphore,
		Name:     "pool",
		Capacity: 10,
		Used:     0,
		Waiting:  0,
	}
	reg.Register(&lockInfo)

	updates := []struct {
		used    int64
		waiting int64
	}{
		{1, 0},
		{5, 2},
		{10, 5},
		{7, 3},
		{0, 0},
	}

	for _, u := range updates {
		reg.UpdateSemaphore("sem-1", u.used, u.waiting)

		snapshot := reg.Snapshot()
		sem := snapshot["sem-1"]
		assert.Equal(t, u.used, sem.Used)
		assert.Equal(t, u.waiting, sem.Waiting)
	}
}

// =============================================================================
// Snapshot Tests
// =============================================================================

// TestRegistry_Snapshot_ReturnsCopy verifies that Snapshot returns a copy, not a reference.
func TestRegistry_Snapshot_ReturnsCopy(t *testing.T) {
	reg := NewRegistry()

	lockInfo := LockInfo{
		ID:       "lock-1",
		Type:     LockTypeMutex,
		Name:     "test",
		Capacity: 1,
		Used:     1,
	}
	reg.Register(&lockInfo)

	// Get snapshot
	snapshot1 := reg.Snapshot()
	assert.Equal(t, 1, len(snapshot1))

	// Modify snapshot1 (should not affect registry)
	delete(snapshot1, "lock-1")
	assert.Equal(t, 0, len(snapshot1))

	// Get another snapshot
	snapshot2 := reg.Snapshot()
	assert.Equal(t, 1, len(snapshot2), "Registry should be unaffected by snapshot modification")
}

// TestRegistry_Snapshot_ReturnsDeepCopy verifies that LockInfo values are copied.
func TestRegistry_Snapshot_ReturnsDeepCopy(t *testing.T) {
	reg := NewRegistry()

	lockInfo := LockInfo{
		ID:       "lock-1",
		Type:     LockTypeSemaphore,
		Name:     "test",
		Capacity: 10,
		Used:     5,
		Waiting:  2,
	}
	reg.Register(&lockInfo)

	// Get snapshot and modify it
	snapshot := reg.Snapshot()
	snapshot["lock-1"] = LockInfo{
		ID:      "lock-1",
		Used:    9999,
		Waiting: 9999,
	}

	// Get another snapshot and verify original values
	snapshot2 := reg.Snapshot()
	assert.Equal(t, int64(5), snapshot2["lock-1"].Used, "Original Used should be unchanged")
	assert.Equal(t, int64(2), snapshot2["lock-1"].Waiting, "Original Waiting should be unchanged")
}

// TestRegistry_Snapshot_EmptyRegistry verifies Snapshot on empty registry.
func TestRegistry_Snapshot_EmptyRegistry(t *testing.T) {
	reg := NewRegistry()

	snapshot := reg.Snapshot()
	require.NotNil(t, snapshot, "Snapshot should never be nil")
	assert.Equal(t, 0, len(snapshot))
}

// =============================================================================
// Summary Tests
// =============================================================================

// TestRegistry_Summary_AggregatesStatistics verifies that Summary returns aggregated statistics.
func TestRegistry_Summary_AggregatesStatistics(t *testing.T) {
	reg := NewRegistry()

	locks := []LockInfo{
		{ID: "mutex-1", Type: LockTypeMutex, Capacity: 1, Used: 1, Waiting: 0},
		{ID: "mutex-2", Type: LockTypeMutex, Capacity: 1, Used: 1, Waiting: 0},
		{ID: "sem-1", Type: LockTypeSemaphore, Capacity: 10, Used: 5, Waiting: 2},
		{ID: "sem-2", Type: LockTypeSemaphore, Capacity: 5, Used: 3, Waiting: 1},
		{ID: "file-1", Type: LockTypeFileLock, Capacity: 1, Used: 1, Waiting: 0},
	}

	for _, lock := range locks {
		reg.Register(&lock)
	}

	summary := reg.Summary()

	assert.Equal(t, 5, summary.Total, "Total should be 5")
	assert.Equal(t, 2, summary.ByType[LockTypeMutex], "Should have 2 mutexes")
	assert.Equal(t, 2, summary.ByType[LockTypeSemaphore], "Should have 2 semaphores")
	assert.Equal(t, 1, summary.ByType[LockTypeFileLock], "Should have 1 filelock")
	assert.Equal(t, int64(18), summary.TotalCapacity, "TotalCapacity should be 1+1+10+5+1=18")
	assert.Equal(t, int64(11), summary.TotalUsed, "TotalUsed should be 1+1+5+3+1=11")
	assert.Equal(t, int64(3), summary.TotalWaiting, "TotalWaiting should be 0+0+2+1+0=3")
}

// TestRegistry_Summary_EmptyRegistry verifies Summary on empty registry.
func TestRegistry_Summary_EmptyRegistry(t *testing.T) {
	reg := NewRegistry()

	summary := reg.Summary()

	assert.Equal(t, 0, summary.Total)
	assert.NotNil(t, summary.ByType, "ByType should not be nil")
	assert.Equal(t, 0, len(summary.ByType))
	assert.Equal(t, int64(0), summary.TotalCapacity)
	assert.Equal(t, int64(0), summary.TotalUsed)
	assert.Equal(t, int64(0), summary.TotalWaiting)
}

// TestRegistry_Summary_UpdatesAfterChanges verifies Summary reflects changes.
func TestRegistry_Summary_UpdatesAfterChanges(t *testing.T) {
	reg := NewRegistry()

	// Add locks
	reg.Register(&LockInfo{ID: "lock-1", Type: LockTypeMutex, Capacity: 1, Used: 1})
	reg.Register(&LockInfo{ID: "lock-2", Type: LockTypeSemaphore, Capacity: 5, Used: 3, Waiting: 1})

	summary1 := reg.Summary()
	assert.Equal(t, 2, summary1.Total)
	assert.Equal(t, int64(6), summary1.TotalCapacity)
	assert.Equal(t, int64(4), summary1.TotalUsed)
	assert.Equal(t, int64(1), summary1.TotalWaiting)

	// Remove a lock
	reg.Unregister("lock-1")

	summary2 := reg.Summary()
	assert.Equal(t, 1, summary2.Total)
	assert.Equal(t, int64(5), summary2.TotalCapacity)
	assert.Equal(t, int64(3), summary2.TotalUsed)
	assert.Equal(t, int64(1), summary2.TotalWaiting)

	// Update semaphore
	reg.UpdateSemaphore("lock-2", 5, 3)

	summary3 := reg.Summary()
	assert.Equal(t, int64(5), summary3.TotalUsed)
	assert.Equal(t, int64(3), summary3.TotalWaiting)
}

// =============================================================================
// Subscribe/Unsubscribe Tests
// =============================================================================

// TestRegistry_Subscribe_ReceivesEvents verifies that subscribers receive events.
func TestRegistry_Subscribe_ReceivesEvents(t *testing.T) {
	reg := NewRegistry()

	eventChan := make(chan LockEvent, 10)
	unsubscribe := reg.Subscribe(eventChan)
	defer unsubscribe()

	// Register a lock
	reg.Register(&LockInfo{ID: "lock-1", Type: LockTypeMutex, Capacity: 1, Used: 1})

	// Verify event received
	select {
	case event := <-eventChan:
		assert.Equal(t, EventAcquired, event.Type)
		assert.Equal(t, "lock-1", event.Lock.ID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected event but none received")
	}
}

// TestRegistry_Subscribe_MultipleSubscribers verifies that multiple subscribers all receive events.
func TestRegistry_Subscribe_MultipleSubscribers(t *testing.T) {
	reg := NewRegistry()

	chan1 := make(chan LockEvent, 10)
	chan2 := make(chan LockEvent, 10)
	chan3 := make(chan LockEvent, 10)

	unsub1 := reg.Subscribe(chan1)
	unsub2 := reg.Subscribe(chan2)
	unsub3 := reg.Subscribe(chan3)

	defer unsub1()
	defer unsub2()
	defer unsub3()

	// Register a lock
	reg.Register(&LockInfo{ID: "lock-1", Type: LockTypeMutex, Capacity: 1, Used: 1})

	// Verify all channels received the event
	channels := []chan LockEvent{chan1, chan2, chan3}
	for i, ch := range channels {
		select {
		case event := <-ch:
			assert.Equal(t, EventAcquired, event.Type, "Channel %d should receive EventAcquired", i)
			assert.Equal(t, "lock-1", event.Lock.ID)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Channel %d did not receive event", i)
		}
	}
}

// TestRegistry_Unsubscribe_StopsEvents verifies that unsubscribed channels stop receiving events.
func TestRegistry_Unsubscribe_StopsEvents(t *testing.T) {
	reg := NewRegistry()

	eventChan := make(chan LockEvent, 10)
	unsubscribe := reg.Subscribe(eventChan)

	// Register first lock
	reg.Register(&LockInfo{ID: "lock-1", Type: LockTypeMutex, Capacity: 1, Used: 1})

	// Receive the event
	select {
	case <-eventChan:
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected event for lock-1")
	}

	// Unsubscribe
	unsubscribe()

	// Register another lock
	reg.Register(&LockInfo{ID: "lock-2", Type: LockTypeMutex, Capacity: 1, Used: 1})

	// Should not receive any event
	select {
	case <-eventChan:
		t.Fatal("Should not receive event after unsubscribe")
	case <-time.After(50 * time.Millisecond):
		// Good, no event received
	}
}

// TestRegistry_Subscribe_EventTypes verifies all event types are delivered correctly.
func TestRegistry_Subscribe_EventTypes(t *testing.T) {
	reg := NewRegistry()

	eventChan := make(chan LockEvent, 10)
	unsubscribe := reg.Subscribe(eventChan)
	defer unsubscribe()

	// Register (EventAcquired)
	reg.Register(&LockInfo{ID: "sem-1", Type: LockTypeSemaphore, Capacity: 5, Used: 1})

	event1 := <-eventChan
	assert.Equal(t, EventAcquired, event1.Type)

	// UpdateSemaphore (EventCapacityChanged)
	reg.UpdateSemaphore("sem-1", 3, 1)

	event2 := <-eventChan
	assert.Equal(t, EventCapacityChanged, event2.Type)

	// Unregister (EventReleased)
	reg.Unregister("sem-1")

	event3 := <-eventChan
	assert.Equal(t, EventReleased, event3.Type)
}

// =============================================================================
// Get (Global Registry) Tests
// =============================================================================

// TestGet_ReturnsSingleton verifies that Get returns the same instance.
func TestGet_ReturnsSingleton(t *testing.T) {
	reg1 := Get()
	reg2 := Get()

	require.NotNil(t, reg1, "Get should return non-nil registry")
	require.NotNil(t, reg2, "Get should return non-nil registry")

	// Should be the same instance
	assert.Same(t, reg1, reg2, "Get should return the same singleton instance")
}

// TestGet_Persistence verifies that the global registry persists data.
func TestGet_Persistence(t *testing.T) {
	reg := Get()

	// Use a unique ID to avoid conflicts with other tests
	lockID := "global-test-lock-" + time.Now().Format("150405.000000000")

	// Clean up after test
	defer reg.Unregister(lockID)

	// Register a lock
	reg.Register(&LockInfo{
		ID:       lockID,
		Type:     LockTypeMutex,
		Name:     "test",
		Capacity: 1,
		Used:     1,
	})

	// Get registry again and verify lock exists
	reg2 := Get()
	snapshot := reg2.Snapshot()

	_, exists := snapshot[lockID]
	assert.True(t, exists, "Lock should exist in global registry")
}

// =============================================================================
// Thread Safety Tests
// =============================================================================

// TestRegistry_ConcurrentRegisterUnregister verifies thread safety of Register/Unregister.
func TestRegistry_ConcurrentRegisterUnregister(t *testing.T) {
	reg := NewRegistry()

	const numGoroutines = 50
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*operationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				lockID := "lock-" + string(rune('A'+id%26)) + "-" + string(rune('0'+j%10))
				lockInfo := LockInfo{
					ID:       lockID,
					Type:     LockTypeMutex,
					Name:     "test",
					Capacity: 1,
					Used:     1,
				}

				// Register
				reg.Register(&lockInfo)

				// Snapshot
				_ = reg.Snapshot()

				// Unregister
				reg.Unregister(lockID)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for panics (they would have been logged)
	for err := range errors {
		t.Errorf("Error during concurrent test: %v", err)
	}

	// Registry should eventually be empty (all locks unregistered)
	snapshot := reg.Snapshot()
	assert.Equal(t, 0, len(snapshot), "Registry should be empty after all operations")
}

// TestRegistry_ConcurrentSnapshot verifies thread safety of Snapshot.
func TestRegistry_ConcurrentSnapshot(t *testing.T) {
	reg := NewRegistry()

	// Pre-populate some locks
	for i := 0; i < 10; i++ {
		reg.Register(&LockInfo{
			ID:       "lock-" + string(rune('0'+i)),
			Type:     LockTypeMutex,
			Capacity: 1,
			Used:     1,
		})
	}

	const numGoroutines = 20
	const snapshotsPerGoroutine = 100

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < snapshotsPerGoroutine; j++ {
				snapshot := reg.Snapshot()
				// Just access the data
				for _, lock := range snapshot {
					_ = lock.ID
					_ = lock.Type
					_ = lock.Used
				}
			}
		}()
	}

	wg.Wait()
}

// TestRegistry_ConcurrentSummary verifies thread safety of Summary.
func TestRegistry_ConcurrentSummary(t *testing.T) {
	reg := NewRegistry()

	// Pre-populate some locks
	reg.Register(&LockInfo{ID: "sem-1", Type: LockTypeSemaphore, Capacity: 10, Used: 5, Waiting: 2})
	reg.Register(&LockInfo{ID: "mutex-1", Type: LockTypeMutex, Capacity: 1, Used: 1})

	const numGoroutines = 20
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				if id%2 == 0 {
					// Update semaphore
					reg.UpdateSemaphore("sem-1", int64(j%10), int64(j%5))
				}

				// Get summary
				summary := reg.Summary()
				_ = summary.Total
				_ = summary.TotalUsed
				_ = summary.TotalWaiting
			}
		}(i)
	}

	wg.Wait()
}

// TestRegistry_ConcurrentSubscribeUnsubscribe verifies thread safety of Subscribe/Unsubscribe.
func TestRegistry_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	reg := NewRegistry()

	const numGoroutines = 20
	const cyclesPerGoroutine = 50

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < cyclesPerGoroutine; j++ {
				ch := make(chan LockEvent, 1)
				unsubscribe := reg.Subscribe(ch)

				// Trigger an event
				lockID := "concurrent-lock"
				reg.Register(&LockInfo{ID: lockID, Type: LockTypeMutex, Capacity: 1, Used: 1})

				// Unsubscribe
				unsubscribe()

				// Clean up
				reg.Unregister(lockID)
			}
		}()
	}

	wg.Wait()
}

// TestRegistry_ConcurrentMixedOperations verifies thread safety with mixed operations.
func TestRegistry_ConcurrentMixedOperations(t *testing.T) {
	reg := NewRegistry()

	const numGoroutines = 30
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup

	// Goroutines that register locks
	for i := 0; i < numGoroutines/3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				lockID := "register-" + string(rune('A'+id%26)) + "-" + string(rune('0'+j%10))
				reg.Register(&LockInfo{
					ID:       lockID,
					Type:     LockTypeSemaphore,
					Capacity: 5,
					Used:     int64(j % 5),
					Waiting:  int64(j % 3),
				})
			}
		}(i)
	}

	// Goroutines that unregister locks
	for i := 0; i < numGoroutines/3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				lockID := "register-" + string(rune('A'+id%26)) + "-" + string(rune('0'+j%10))
				reg.Unregister(lockID)
			}
		}(i)
	}

	// Goroutines that read snapshot and summary
	for i := 0; i < numGoroutines/3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				_ = reg.Snapshot()
				_ = reg.Summary()
			}
		}()
	}

	wg.Wait()
}

// TestRegistry_RaceDetector verifies no race conditions (run with -race flag).
func TestRegistry_RaceDetector(t *testing.T) {
	reg := NewRegistry()

	// Subscribe to events
	eventChan := make(chan LockEvent, 100)
	unsubscribe := reg.Subscribe(eventChan)

	// Consumer goroutine
	go func() {
		for range eventChan {
			// Just consume
		}
	}()

	const numGoroutines = 10

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			lockID := "race-lock-" + string(rune('A'+id))

			// Register
			reg.Register(&LockInfo{
				ID:       lockID,
				Type:     LockTypeSemaphore,
				Capacity: 5,
				Used:     1,
			})

			// Update
			reg.UpdateSemaphore(lockID, 3, 1)

			// Snapshot
			_ = reg.Snapshot()

			// Summary
			_ = reg.Summary()

			// Unregister
			reg.Unregister(lockID)
		}(i)
	}

	wg.Wait()

	unsubscribe()
	close(eventChan)
}
