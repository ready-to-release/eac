package locktracker

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// TrackedSemaphore is a channel-based semaphore that tracks its usage in a Registry.
type TrackedSemaphore struct {
	id       string
	name     string
	capacity int64
	sem      chan struct{}
	registry *Registry
	used     int64
	waiting  int64
	closed   bool
	closeMu  sync.Mutex
}

// NewTrackedSemaphore creates a new tracked semaphore with the given name and capacity.
// It registers with the global registry.
func NewTrackedSemaphore(name string, capacity int64) *TrackedSemaphore {
	return NewTrackedSemaphoreWithRegistry(name, capacity, Get())
}

// NewTrackedSemaphoreWithRegistry creates a new tracked semaphore with a specific registry.
func NewTrackedSemaphoreWithRegistry(name string, capacity int64, registry *Registry) *TrackedSemaphore {
	ts := &TrackedSemaphore{
		id:       uuid.New().String(),
		name:     name,
		capacity: capacity,
		sem:      make(chan struct{}, capacity),
		registry: registry,
	}

	// Register with the registry
	ts.registry.Register(&LockInfo{
		ID:       ts.id,
		Type:     LockTypeSemaphore,
		Name:     name,
		Capacity: capacity,
		Used:     0,
		Waiting:  0,
	})

	return ts
}

// Acquire blocks until a slot is available or the context is canceled.
func (ts *TrackedSemaphore) Acquire(ctx context.Context) error {
	// Increment waiting count
	atomic.AddInt64(&ts.waiting, 1)
	ts.updateRegistry()

	select {
	case ts.sem <- struct{}{}:
		// Got a slot
		atomic.AddInt64(&ts.waiting, -1)
		atomic.AddInt64(&ts.used, 1)
		ts.updateRegistry()
		return nil
	case <-ctx.Done():
		// Context canceled
		atomic.AddInt64(&ts.waiting, -1)
		ts.updateRegistry()
		return ctx.Err()
	}
}

// Release frees a slot in the semaphore.
func (ts *TrackedSemaphore) Release() {
	select {
	case <-ts.sem:
		atomic.AddInt64(&ts.used, -1)
		ts.updateRegistry()
	default:
		// Semaphore was empty, shouldn't happen in correct usage
	}
}

// TryAcquire attempts to acquire a slot without blocking. Returns true if successful.
func (ts *TrackedSemaphore) TryAcquire() bool {
	select {
	case ts.sem <- struct{}{}:
		atomic.AddInt64(&ts.used, 1)
		ts.updateRegistry()
		return true
	default:
		return false
	}
}

// Stats returns the current (capacity, used, waiting) counts.
func (ts *TrackedSemaphore) Stats() (capacity, used, waiting int64) {
	return ts.capacity, atomic.LoadInt64(&ts.used), atomic.LoadInt64(&ts.waiting)
}

// Used returns the current number of slots in use.
func (ts *TrackedSemaphore) Used() int64 {
	return atomic.LoadInt64(&ts.used)
}

// AcquireBlocking blocks until a slot is available (no context cancellation support).
func (ts *TrackedSemaphore) AcquireBlocking() {
	// Increment waiting count
	atomic.AddInt64(&ts.waiting, 1)
	ts.updateRegistry()

	// Block until slot available
	ts.sem <- struct{}{}

	// Got a slot
	atomic.AddInt64(&ts.waiting, -1)
	atomic.AddInt64(&ts.used, 1)
	ts.updateRegistry()
}

// Close unregisters the semaphore from tracking.
func (ts *TrackedSemaphore) Close() {
	ts.closeMu.Lock()
	defer ts.closeMu.Unlock()

	if ts.closed {
		return
	}
	ts.closed = true

	ts.registry.Unregister(ts.id)
}

// updateRegistry updates the registry with current stats.
func (ts *TrackedSemaphore) updateRegistry() {
	ts.registry.UpdateSemaphore(ts.id, atomic.LoadInt64(&ts.used), atomic.LoadInt64(&ts.waiting))
}
