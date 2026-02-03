package orchestrator

import (
	"container/heap"
	"sync"

	"github.com/ready-to-release/eac/go/core/execution"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// WorkQueue is a thread-safe priority queue for work units.
// Items are ordered by weight descending (LPT scheduling - Longest Processing Time First).
// Only items with satisfied dependencies are considered ready.
type WorkQueue struct {
	mu               sync.Mutex
	items            workHeap
	deps             *DepsTracker
	notEmpty         *sync.Cond // signaled when items added or deps satisfied
	closed           bool
	overCapacityItem *workunit.UnitID // Track active over-capacity execution (weight > totalCapacity)
}

// QueueStats holds queue statistics for display.
type QueueStats struct {
	Total   int // total items in queue
	Ready   int // items ready to run (deps satisfied)
	Blocked int // items waiting on dependencies
}

// NewWorkQueue creates a new work queue with dependency tracking.
// Items are ordered by weight descending (heaviest first).
//
// Uses LayerModeNone by default: component layers + DependsOn are enforced,
// but module layer ordering is not. For strict layer enforcement, use
// NewWorkQueueWithPolicy with LayerModeStrict.
//
// Panics if the work units contain circular dependencies.
func NewWorkQueue(work []workunit.UnitSpec) *WorkQueue {
	// Use LayerModeNone as default - this enforces component layers and DependsOn
	// but does not enforce module layer ordering (parallel across modules).
	q, err := NewWorkQueueWithPolicy(work, execution.LayerModeNone)
	if err != nil {
		// Circular dependencies are a configuration error - fail fast.
		panic("NewWorkQueue: " + err.Error())
	}
	return q
}

// NewWorkQueueWithPolicy creates a work queue that uses LayerPolicy for scheduling.
// The LayerPolicy determines when units are ready based on layer constraints:
//   - LayerModeStrict: module layers + component layers + DependsOn
//   - LayerModeNone: component layers + DependsOn (module layers not enforced)
//
// Returns an error if the plan computation fails (e.g., circular dependencies).
func NewWorkQueueWithPolicy(work []workunit.UnitSpec, mode execution.LayerMode) (*WorkQueue, error) {
	// Compute the execution plan using the LayerPlanner
	planner := execution.NewLayerPlanner(mode)
	plan, err := planner.ComputePlan(work)
	if err != nil {
		return nil, err
	}

	// Create DepsTracker and configure with LayerPolicy
	deps := NewDepsTracker(work)
	deps.SetLayerPolicy(planner, plan)

	q := &WorkQueue{
		items: make(workHeap, 0, len(work)),
		deps:  deps,
	}
	q.notEmpty = sync.NewCond(&q.mu)

	// Add all items to heap
	for _, w := range work {
		heap.Push(&q.items, w)
	}

	return q, nil
}

// HasLayerPolicy returns true if the queue is using LayerPolicy for scheduling.
func (q *WorkQueue) HasLayerPolicy() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.deps.HasLayerPolicy()
}

// PopReady blocks until a ready item is available, then returns it.
// An item is ready when all its DependsOn components have completed.
// Returns nil when queue is empty or closed.
// Note: Use PopReadyWithBudget for bin-packing scheduling.
func (q *WorkQueue) PopReady() *workunit.UnitSpec {
	return q.PopReadyWithBudget(0, 0) // 0 = unlimited budget, 0 = unlimited capacity
}

// PopReadyWithBudget blocks until a ready item is available that fits within the budget.
// If budget <= 0, returns the heaviest ready item (pure LPT).
// If budget > 0, returns the heaviest ready item with weight <= budget (bin-packing).
// If totalCapacity > 0 and item weight > totalCapacity, marks item as over-capacity (serial execution).
// Returns nil when queue is empty or closed.
func (q *WorkQueue) PopReadyWithBudget(budget int, totalCapacity int) *workunit.UnitSpec {
	q.mu.Lock()
	defer q.mu.Unlock()

	for {
		// Queue exhausted - all items have been popped
		if len(q.items) == 0 {
			return nil
		}

		// Find highest-weight ready item that fits budget
		if spec := q.findReadyWithBudget(budget, totalCapacity); spec != nil {
			return spec
		}

		// No ready items but queue not empty - items blocked on deps
		// If queue is closed, we have a deadlock (shouldn't happen)
		if q.closed {
			return nil
		}

		// Wait for deps to complete (MarkComplete broadcasts)
		q.notEmpty.Wait()
	}
}

// findReadyWithBudget finds and removes the highest-weight ready item within budget.
// If budget <= 0, ignores budget constraint (pure LPT).
// Over-capacity fallback: If no items fit budget, returns heaviest ready item to prevent deadlock.
// If totalCapacity > 0 and item weight > totalCapacity, marks item as over-capacity for serial execution.
// Must be called with mu held.
func (q *WorkQueue) findReadyWithBudget(budget int, totalCapacity int) *workunit.UnitSpec {
	// Scan heap for best ready item that fits budget
	// Track both fitting items (within budget) and heaviest ready (fallback for over-capacity)
	bestFitting := -1        // Best item within budget
	bestFittingWeight := -1
	heaviest := -1           // Heaviest ready item (fallback)
	heaviestWeight := -1

	for i, item := range q.items {
		if !q.deps.IsReady(item.ID) {
			continue
		}

		// If over-capacity item is active, skip OTHER over-capacity items
		// but allow normal items to be popped (preserves layer parallelism)
		if q.overCapacityItem != nil {
			wouldBeOverCapacity := (totalCapacity > 0 && budget > 0 &&
				item.Weight > budget && item.Weight > totalCapacity)
			if wouldBeOverCapacity {
				continue // Skip this over-capacity item while another is active
			}
		}

		// Track heaviest ready item (fallback for over-capacity)
		if item.Weight > heaviestWeight {
			heaviest = i
			heaviestWeight = item.Weight
		}

		// Track best fitting item (within budget)
		if budget <= 0 || item.Weight <= budget {
			if item.Weight > bestFittingWeight {
				bestFitting = i
				bestFittingWeight = item.Weight
			}
		}
	}

	// Prefer fitting items, fall back to heaviest when nothing fits
	if bestFitting != -1 {
		item := heap.Remove(&q.items, bestFitting).(workunit.UnitSpec)
		return &item
	}

	// Over-capacity fallback: run heaviest item even if it exceeds budget
	if heaviest != -1 && budget > 0 {
		item := heap.Remove(&q.items, heaviest).(workunit.UnitSpec)

		// Mark as over-capacity if it exceeds TOTAL capacity (not just available budget)
		// Over-capacity items execute serially to prevent deadlock
		if totalCapacity > 0 && item.Weight > totalCapacity {
			q.overCapacityItem = &item.ID
			logging.C().Debugf("[scheduler] Over-capacity: starting weight=%d (capacity=%d) - SERIAL MODE",
				item.Weight, totalCapacity)
		} else {
			logging.C().Debugf("[scheduler] Over-capacity: starting weight=%d (available=%d)",
				item.Weight, budget)
		}

		return &item
	}

	return nil // Queue exhausted
}

// HasReadyWithBudget checks if there's a ready item that fits within budget.
// Used by scheduler to decide whether to wait for capacity or look for smaller items.
func (q *WorkQueue) HasReadyWithBudget(budget int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.items {
		if q.deps.IsReady(item.ID) {
			if budget <= 0 || item.Weight <= budget {
				return true
			}
		}
	}
	return false
}

// MarkComplete marks a component as done successfully, potentially unblocking dependents.
func (q *WorkQueue) MarkComplete(id workunit.UnitID) {
	q.mu.Lock()
	q.deps.MarkComplete(id)
	q.notEmpty.Broadcast() // wake PopReady to check for newly ready items
	q.mu.Unlock()
}

// ReleaseOverCapacity clears the over-capacity item marker and wakes waiting workers.
// Must be called after an over-capacity item (weight > totalCapacity) completes execution.
// If the provided ID doesn't match the current over-capacity item, this is a no-op.
func (q *WorkQueue) ReleaseOverCapacity(id workunit.UnitID) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.overCapacityItem != nil && q.overCapacityItem.Longname() == id.Longname() {
		q.overCapacityItem = nil
		q.notEmpty.Broadcast() // Wake workers waiting for over-capacity to clear
		logging.C().Debugf("[scheduler] Over-capacity: released lock for %s", id.Longname())
	}
}

// IsOverCapacityActive returns true if an over-capacity item is currently executing.
// Useful for testing and debugging.
func (q *WorkQueue) IsOverCapacityActive() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.overCapacityItem != nil
}

// MarkFailed marks a component as failed, which will cause dependents to fail.
func (q *WorkQueue) MarkFailed(id workunit.UnitID) {
	q.mu.Lock()
	q.deps.MarkFailed(id)
	q.notEmpty.Broadcast() // wake PopReady to check for items that should now fail
	q.mu.Unlock()
}

// HasFailedDependency returns true if any dependency of the unit has failed.
func (q *WorkQueue) HasFailedDependency(id workunit.UnitID) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.deps.HasFailedDependency(id)
}

// Close signals that no more items will be added and wakes any waiting PopReady calls.
func (q *WorkQueue) Close() {
	q.mu.Lock()
	q.closed = true
	q.notEmpty.Broadcast()
	q.mu.Unlock()
}

// Len returns current queue length.
func (q *WorkQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Stats returns queue statistics for TUI display.
func (q *WorkQueue) Stats() QueueStats {
	q.mu.Lock()
	defer q.mu.Unlock()

	ready := 0
	blocked := 0
	for _, item := range q.items {
		if q.deps.IsReady(item.ID) {
			ready++
		} else {
			blocked++
		}
	}
	return QueueStats{
		Total:   len(q.items),
		Ready:   ready,
		Blocked: blocked,
	}
}

// --- Heap implementation ---

type workHeap []workunit.UnitSpec

func (h workHeap) Len() int { return len(h) }

// Less returns true if item i should come before item j.
// For max heap (LPT), higher weight should come first.
func (h workHeap) Less(i, j int) bool { return h[i].Weight > h[j].Weight }

func (h workHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *workHeap) Push(x any) {
	*h = append(*h, x.(workunit.UnitSpec))
}

func (h *workHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
