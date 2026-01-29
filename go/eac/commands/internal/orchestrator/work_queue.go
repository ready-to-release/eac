package orchestrator

import (
	"container/heap"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

// WorkQueue is a thread-safe priority queue for work units.
// Items are ordered by weight descending (LPT scheduling - Longest Processing Time First).
// Only items with satisfied dependencies are considered ready.
type WorkQueue struct {
	mu       sync.Mutex
	items    workHeap
	deps     *DepsTracker
	notEmpty *sync.Cond // signaled when items added or deps satisfied
	closed   bool
}

// QueueStats holds queue statistics for display.
type QueueStats struct {
	Total   int // total items in queue
	Ready   int // items ready to run (deps satisfied)
	Blocked int // items waiting on dependencies
}

// NewWorkQueue creates a new work queue with dependency tracking.
// Items are ordered by weight descending (heaviest first).
func NewWorkQueue(work []workunit.UnitSpec) *WorkQueue {
	q := &WorkQueue{
		items: make(workHeap, 0, len(work)),
		deps:  NewDepsTracker(work),
	}
	q.notEmpty = sync.NewCond(&q.mu)

	// Add all items to heap
	for _, w := range work {
		heap.Push(&q.items, w)
	}

	return q
}

// PopReady blocks until a ready item is available, then returns it.
// An item is ready when all its DependsOn components have completed.
// Returns nil when queue is empty or closed.
func (q *WorkQueue) PopReady() *workunit.UnitSpec {
	q.mu.Lock()
	defer q.mu.Unlock()

	for {
		// Queue exhausted - all items have been popped
		if len(q.items) == 0 {
			return nil
		}

		// Find highest-weight ready item
		if spec := q.findReady(); spec != nil {
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

// findReady finds and removes the highest-weight item with satisfied deps.
// Must be called with mu held.
func (q *WorkQueue) findReady() *workunit.UnitSpec {
	// Scan heap for first ready item
	// Note: The heap maintains weight ordering, so scanning from start
	// gives us items in roughly weight order, but heap property only
	// guarantees parent > children, not strict sorted order.
	// We scan to find the highest-weight ready item.
	bestIdx := -1
	bestWeight := -1

	for i, item := range q.items {
		if q.deps.IsReady(item.ID) {
			if item.Weight > bestWeight {
				bestIdx = i
				bestWeight = item.Weight
			}
		}
	}

	if bestIdx == -1 {
		return nil
	}

	// Remove from heap and return
	item := heap.Remove(&q.items, bestIdx).(workunit.UnitSpec)
	return &item
}

// MarkComplete marks a component as done, potentially unblocking dependents.
func (q *WorkQueue) MarkComplete(id workunit.UnitID) {
	q.mu.Lock()
	q.deps.MarkComplete(id)
	q.notEmpty.Broadcast() // wake PopReady to check for newly ready items
	q.mu.Unlock()
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
