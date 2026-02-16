package scheduling

import (
	"container/heap"
	"sync"

	"github.com/ready-to-release/eac/go/core/workunit"
)

// DependencyScheduler implements WorkScheduler with dependency-aware LPT scheduling.
// Items are ordered by weight descending (heaviest first).
// Only items with all dependencies complete are considered ready.
type DependencyScheduler struct {
	mu       sync.Mutex
	items    unitHeap           // Max-heap by weight
	tracker  *DependencyTracker // Dependency state
	notReady *sync.Cond         // Signal when items may become ready
	closed   bool
	total    int // Original count for stats
}

// SchedulerOption configures optional behavior for NewDependencyScheduler.
type SchedulerOption func(*schedulerConfig)

type schedulerConfig struct {
	skipValidation bool
}

// WithSkipValidation skips cycle detection during scheduler construction.
// Use when the caller has already validated that the dependency graph is acyclic
// (e.g., via a prior module-level topological sort).
func WithSkipValidation() SchedulerOption {
	return func(c *schedulerConfig) {
		c.skipValidation = true
	}
}

// NewDependencyScheduler creates a scheduler for the given work units.
// Returns error if circular dependencies are detected (unless WithSkipValidation is used).
func NewDependencyScheduler(units []workunit.UnitSpec, opts ...SchedulerOption) (*DependencyScheduler, error) {
	var cfg schedulerConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	if !cfg.skipValidation {
		if err := validateNoDuplicates(units); err != nil {
			return nil, err
		}
		if err := validateNoCycles(units); err != nil {
			return nil, err
		}
	}

	ds := &DependencyScheduler{
		items:   make(unitHeap, 0, len(units)),
		tracker: NewDependencyTracker(units),
		total:   len(units),
	}
	ds.notReady = sync.NewCond(&ds.mu)

	// Add all items to heap
	for _, u := range units {
		heap.Push(&ds.items, u)
	}

	return ds, nil
}

// Next returns the heaviest ready item (LPT scheduling).
// Returns nil immediately if no items are ready - does not block.
// The returned item is removed from the queue (now "in-flight").
func (ds *DependencyScheduler) Next() *workunit.UnitSpec {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	return ds.findHeaviestReady()
}

// WaitForReady blocks until a ready item is available, then returns it.
// Returns nil when queue is exhausted or scheduler is closed.
// The returned item is removed from the queue (now "in-flight").
func (ds *DependencyScheduler) WaitForReady() *workunit.UnitSpec {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for {
		// Check termination conditions
		if ds.closed {
			return nil
		}

		// Queue exhausted - all items have been popped and completed
		if len(ds.items) == 0 {
			return nil
		}

		// Try to find a ready item
		if item := ds.findHeaviestReady(); item != nil {
			return item
		}

		// No ready items but queue not empty - items blocked on deps
		// Wait for deps to complete (MarkComplete broadcasts)
		ds.notReady.Wait()
	}
}

// findHeaviestReady finds and removes the heaviest ready item (LPT).
// Caller must hold lock.
func (ds *DependencyScheduler) findHeaviestReady() *workunit.UnitSpec {
	// Scan heap for heaviest ready item
	// Note: heap property ensures items are roughly ordered by weight,
	// but we need to scan for the heaviest READY item specifically
	best := -1
	bestWeight := -1

	for i, item := range ds.items {
		if !ds.tracker.IsReady(item.ID) {
			continue
		}

		if item.Weight > bestWeight {
			best = i
			bestWeight = item.Weight
		}
	}

	if best == -1 {
		return nil
	}

	item := heap.Remove(&ds.items, best).(workunit.UnitSpec)
	return &item
}

// MarkComplete signals completion and wakes waiters.
func (ds *DependencyScheduler) MarkComplete(id workunit.UnitID) {
	ds.mu.Lock()
	ds.tracker.MarkComplete(id)
	ds.notReady.Broadcast()
	ds.mu.Unlock()
}

// MarkFailed signals failure and wakes waiters.
func (ds *DependencyScheduler) MarkFailed(id workunit.UnitID) {
	ds.mu.Lock()
	ds.tracker.MarkFailed(id)
	ds.notReady.Broadcast()
	ds.mu.Unlock()
}

// MarkFailedCascade signals failure and proactively fails all transitive dependents.
// Returns the specs that were cascade-failed (removed from the queue).
// This prevents dependent UoWs from being scheduled, saving worker slots.
func (ds *DependencyScheduler) MarkFailedCascade(id workunit.UnitID) []workunit.UnitSpec {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Mark the original item as failed
	ds.tracker.MarkFailed(id)

	// Cascade to all transitive dependents
	cascadedIDs := ds.tracker.CascadeFail(id)

	// Build set for fast lookup
	cascadedSet := make(map[string]bool, len(cascadedIDs))
	for _, cid := range cascadedIDs {
		cascadedSet[cid.Longname()] = true
	}

	// Remove cascade-failed items from heap (walk backwards for stable indices)
	var cascadedSpecs []workunit.UnitSpec
	for i := len(ds.items) - 1; i >= 0; i-- {
		if cascadedSet[ds.items[i].ID.Longname()] {
			removed := heap.Remove(&ds.items, i).(workunit.UnitSpec)
			cascadedSpecs = append(cascadedSpecs, removed)
		}
	}

	// Wake waiters — some may now see the scheduler exhausted
	ds.notReady.Broadcast()

	return cascadedSpecs
}

// HasFailedDependency checks if any dependency failed.
func (ds *DependencyScheduler) HasFailedDependency(id workunit.UnitID) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	return ds.tracker.HasFailedDependency(id)
}

// FailedDependencyModules returns the unique module names of failed dependencies.
func (ds *DependencyScheduler) FailedDependencyModules(id workunit.UnitID) []string {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	return ds.tracker.FailedDependencyModules(id)
}

// Stats returns current statistics.
func (ds *DependencyScheduler) Stats() Stats {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ready := 0
	blocked := 0
	for _, item := range ds.items {
		if ds.tracker.IsReady(item.ID) {
			ready++
		} else {
			blocked++
		}
	}

	return Stats{
		Total:     ds.total,
		Pending:   len(ds.items),
		Ready:     ready,
		Blocked:   blocked,
		Completed: ds.tracker.CompletedCount() - ds.tracker.FailedCount(),
		Failed:    ds.tracker.FailedCount(),
	}
}

// Close shuts down the scheduler and wakes any blocked WaitForReady calls.
func (ds *DependencyScheduler) Close() {
	ds.mu.Lock()
	ds.closed = true
	ds.notReady.Broadcast()
	ds.mu.Unlock()
}

// Len returns the number of items still in the queue (not yet popped).
func (ds *DependencyScheduler) Len() int {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	return len(ds.items)
}
