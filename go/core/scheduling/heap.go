package scheduling

import "github.com/ready-to-release/eac/go/core/workunit"

// unitHeap implements container/heap.Interface for max-heap by weight.
// Higher weight items are popped first (LPT - Longest Processing Time First).
type unitHeap []workunit.UnitSpec

func (h unitHeap) Len() int { return len(h) }

// Less returns true if item i should come before item j.
// For max heap (LPT), higher weight should come first.
func (h unitHeap) Less(i, j int) bool { return h[i].Weight > h[j].Weight }

func (h unitHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

// Push adds an item to the heap.
func (h *unitHeap) Push(x any) {
	*h = append(*h, x.(workunit.UnitSpec))
}

// Pop removes and returns the last item (heap will fix up).
func (h *unitHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
