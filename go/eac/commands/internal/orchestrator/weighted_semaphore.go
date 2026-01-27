package orchestrator

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ready-to-release/eac/go/eac/commands/internal/locktracker"
)

// WeightedSemaphore controls concurrent access based on resource weight.
// Unlike a counting semaphore that treats all work as equal, this allows
// different work items to consume different amounts of capacity.
//
// Example with capacity=8:
//   - 8 weight-1 tasks can run simultaneously
//   - 2 weight-4 tasks can run simultaneously
//   - 1 weight-4 + 4 weight-1 tasks can run simultaneously
type WeightedSemaphore struct {
	mu       sync.Mutex
	cond     *sync.Cond
	capacity int
	used     int
	waiting  int // Count of goroutines waiting to acquire

	// Lock tracking
	id       string
	name     string
	registry *locktracker.Registry
	closed   bool
}

// NewWeightedSemaphore creates a semaphore with the given capacity.
// This creates a semaphore without lock tracking. Use NewWeightedSemaphoreWithRegistry
// for semaphores that should be tracked.
func NewWeightedSemaphore(capacity int) *WeightedSemaphore {
	ws := &WeightedSemaphore{
		capacity: capacity,
		used:     0,
	}
	ws.cond = sync.NewCond(&ws.mu)
	return ws
}

// NewWeightedSemaphoreWithRegistry creates a semaphore with lock tracking.
// The semaphore is registered with the provided registry and will report
// capacity usage and waiting counts in real-time.
func NewWeightedSemaphoreWithRegistry(name string, capacity int, registry *locktracker.Registry) *WeightedSemaphore {
	ws := &WeightedSemaphore{
		capacity: capacity,
		used:     0,
		id:       uuid.New().String(),
		name:     name,
		registry: registry,
	}
	ws.cond = sync.NewCond(&ws.mu)

	// Register with the lock tracker
	ws.registry.Register(&locktracker.LockInfo{
		ID:       ws.id,
		Type:     locktracker.LockTypeWeighted,
		Name:     name,
		Capacity: int64(capacity),
		Used:     0,
		Waiting:  0,
	})

	return ws
}

// Acquire blocks until the requested weight can be allocated.
// If weight exceeds capacity, it will still proceed (blocking all others).
func (ws *WeightedSemaphore) Acquire(weight int) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Track waiting if we'll need to wait
	needsWait := false

	// Handle edge case: if weight > capacity, we still allow it
	// but it blocks everything else until done
	if weight > ws.capacity {
		// Wait for all current work to complete
		if ws.used > 0 {
			needsWait = true
			ws.waiting++
			ws.updateRegistry()
		}
		for ws.used > 0 {
			ws.cond.Wait()
		}
		if needsWait {
			ws.waiting--
		}
		ws.used = weight
		ws.updateRegistry()
		return
	}

	// Normal case: wait until we have enough capacity
	if ws.used+weight > ws.capacity {
		needsWait = true
		ws.waiting++
		ws.updateRegistry()
	}
	for ws.used+weight > ws.capacity {
		ws.cond.Wait()
	}
	if needsWait {
		ws.waiting--
	}
	ws.used += weight
	ws.updateRegistry()
}

// TryAcquire attempts to acquire the requested weight without blocking.
// Returns true if acquired, false if not enough capacity.
func (ws *WeightedSemaphore) TryAcquire(weight int) bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.used+weight <= ws.capacity {
		ws.used += weight
		return true
	}
	return false
}

// Release returns the specified weight to the pool.
func (ws *WeightedSemaphore) Release(weight int) {
	ws.mu.Lock()
	ws.used -= weight
	if ws.used < 0 {
		ws.used = 0 // Safety: don't go negative
	}
	ws.updateRegistry()
	ws.cond.Broadcast()
	ws.mu.Unlock()
}

// Capacity returns the total capacity.
func (ws *WeightedSemaphore) Capacity() int {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.capacity
}

// SetCapacity dynamically adjusts the semaphore capacity.
// This is used for adaptive resource-based scheduling where capacity
// is recalculated based on available system resources.
// The new capacity takes effect immediately, but existing acquired
// slots are not forcibly released.
// Capacity will not be reduced below the currently used slots to avoid
// displaying invalid states like "3/2".
func (ws *WeightedSemaphore) SetCapacity(newCapacity int) {
	if newCapacity < 1 {
		newCapacity = 1
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Don't reduce capacity below currently used slots
	// This prevents displaying invalid states like "3/2"
	if newCapacity < ws.used {
		newCapacity = ws.used
	}

	oldCapacity := ws.capacity
	ws.capacity = newCapacity

	// Update registry with new capacity
	if ws.registry != nil {
		ws.registry.UpdateSemaphoreCapacity(ws.id, int64(newCapacity))
	}

	// If capacity increased, wake up waiters who might now be able to proceed
	if newCapacity > oldCapacity {
		ws.cond.Broadcast()
	}
}

// Used returns the currently used capacity.
func (ws *WeightedSemaphore) Used() int {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.used
}

// Available returns the currently available capacity.
func (ws *WeightedSemaphore) Available() int {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.capacity - ws.used
}

// Close unregisters the semaphore from tracking.
// Should be called when the semaphore is no longer needed.
func (ws *WeightedSemaphore) Close() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed || ws.registry == nil {
		return
	}
	ws.closed = true
	ws.registry.Unregister(ws.id)
}

// updateRegistry sends current state to the lock tracker.
// Must be called with mu held.
func (ws *WeightedSemaphore) updateRegistry() {
	if ws.registry == nil {
		return
	}
	ws.registry.UpdateSemaphore(ws.id, int64(ws.used), int64(ws.waiting))
}
