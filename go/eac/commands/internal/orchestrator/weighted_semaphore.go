package orchestrator

import (
	"sync"
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
}

// NewWeightedSemaphore creates a semaphore with the given capacity.
func NewWeightedSemaphore(capacity int) *WeightedSemaphore {
	ws := &WeightedSemaphore{
		capacity: capacity,
		used:     0,
	}
	ws.cond = sync.NewCond(&ws.mu)
	return ws
}

// Acquire blocks until the requested weight can be allocated.
// If weight exceeds capacity, it will still proceed (blocking all others).
func (ws *WeightedSemaphore) Acquire(weight int) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Handle edge case: if weight > capacity, we still allow it
	// but it blocks everything else until done
	if weight > ws.capacity {
		// Wait for all current work to complete
		for ws.used > 0 {
			ws.cond.Wait()
		}
		ws.used = weight
		return
	}

	// Normal case: wait until we have enough capacity
	for ws.used+weight > ws.capacity {
		ws.cond.Wait()
	}
	ws.used += weight
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
	ws.cond.Broadcast()
	ws.mu.Unlock()
}

// Capacity returns the total capacity.
func (ws *WeightedSemaphore) Capacity() int {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.capacity
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
