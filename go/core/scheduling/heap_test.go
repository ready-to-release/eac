package scheduling

import (
	"container/heap"
	"testing"

	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// unitHeap interface method tests
// =============================================================================

func TestUnitHeap_Len(t *testing.T) {
	tests := []struct {
		name string
		h    unitHeap
		want int
	}{
		{
			name: "empty heap",
			h:    unitHeap{},
			want: 0,
		},
		{
			name: "single element",
			h:    unitHeap{{ID: uid("mod1", "A"), Weight: 1}},
			want: 1,
		},
		{
			name: "multiple elements",
			h: unitHeap{
				{ID: uid("mod1", "A"), Weight: 1},
				{ID: uid("mod1", "B"), Weight: 2},
				{ID: uid("mod1", "C"), Weight: 3},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.h.Len())
		})
	}
}

func TestUnitHeap_Less_MaxHeapOrdering(t *testing.T) {
	// Less returns true when i has higher weight than j (max-heap / LPT)
	tests := []struct {
		name     string
		weightI  int
		weightJ  int
		wantLess bool
	}{
		{
			name:     "higher weight comes first",
			weightI:  10,
			weightJ:  5,
			wantLess: true,
		},
		{
			name:     "lower weight does not come first",
			weightI:  5,
			weightJ:  10,
			wantLess: false,
		},
		{
			name:     "equal weights are not less",
			weightI:  5,
			weightJ:  5,
			wantLess: false,
		},
		{
			name:     "zero weight vs positive",
			weightI:  0,
			weightJ:  1,
			wantLess: false,
		},
		{
			name:     "positive weight vs zero",
			weightI:  1,
			weightJ:  0,
			wantLess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := unitHeap{
				{ID: uid("mod1", "A"), Weight: tt.weightI},
				{ID: uid("mod1", "B"), Weight: tt.weightJ},
			}
			assert.Equal(t, tt.wantLess, h.Less(0, 1))
		})
	}
}

func TestUnitHeap_Swap(t *testing.T) {
	h := unitHeap{
		{ID: uid("mod1", "A"), Weight: 1},
		{ID: uid("mod1", "B"), Weight: 2},
	}

	h.Swap(0, 1)

	assert.Equal(t, "B", h[0].ID.ComponentName)
	assert.Equal(t, "A", h[1].ID.ComponentName)
	assert.Equal(t, 2, h[0].Weight)
	assert.Equal(t, 1, h[1].Weight)
}

func TestUnitHeap_Push(t *testing.T) {
	h := &unitHeap{}

	h.Push(workunit.UnitSpec{ID: uid("mod1", "A"), Weight: 5})

	assert.Equal(t, 1, h.Len())
	assert.Equal(t, "A", (*h)[0].ID.ComponentName)
	assert.Equal(t, 5, (*h)[0].Weight)
}

func TestUnitHeap_Push_Multiple(t *testing.T) {
	h := &unitHeap{}

	h.Push(workunit.UnitSpec{ID: uid("mod1", "A"), Weight: 1})
	h.Push(workunit.UnitSpec{ID: uid("mod1", "B"), Weight: 2})
	h.Push(workunit.UnitSpec{ID: uid("mod1", "C"), Weight: 3})

	assert.Equal(t, 3, h.Len())
}

func TestUnitHeap_Pop(t *testing.T) {
	h := &unitHeap{
		{ID: uid("mod1", "A"), Weight: 1},
		{ID: uid("mod1", "B"), Weight: 2},
	}

	// Pop removes the last element (container/heap handles re-ordering)
	result := h.Pop()
	popped, ok := result.(workunit.UnitSpec)

	require.True(t, ok, "Pop should return a workunit.UnitSpec")
	assert.Equal(t, "B", popped.ID.ComponentName)
	assert.Equal(t, 2, popped.Weight)
	assert.Equal(t, 1, h.Len())
}

func TestUnitHeap_Pop_SingleElement(t *testing.T) {
	h := &unitHeap{
		{ID: uid("mod1", "A"), Weight: 42},
	}

	result := h.Pop()
	popped := result.(workunit.UnitSpec)

	assert.Equal(t, "A", popped.ID.ComponentName)
	assert.Equal(t, 42, popped.Weight)
	assert.Equal(t, 0, h.Len())
}

// =============================================================================
// container/heap integration tests
// =============================================================================

func TestUnitHeap_HeapInit_MaxHeapProperty(t *testing.T) {
	// After heap.Init, the root should be the max-weight element
	h := &unitHeap{
		{ID: uid("mod1", "light"), Weight: 1},
		{ID: uid("mod1", "heavy"), Weight: 10},
		{ID: uid("mod1", "medium"), Weight: 5},
	}

	heap.Init(h)

	// Root of max-heap should be the heaviest
	assert.Equal(t, 10, (*h)[0].Weight, "heap root should be the max-weight element")
}

func TestUnitHeap_HeapPushPop_LPTOrder(t *testing.T) {
	// Use container/heap Push/Pop to verify LPT (heaviest first) ordering
	h := &unitHeap{}
	heap.Init(h)

	heap.Push(h, workunit.UnitSpec{ID: uid("mod1", "light"), Weight: 1})
	heap.Push(h, workunit.UnitSpec{ID: uid("mod1", "heavy"), Weight: 10})
	heap.Push(h, workunit.UnitSpec{ID: uid("mod1", "medium"), Weight: 5})
	heap.Push(h, workunit.UnitSpec{ID: uid("mod1", "feather"), Weight: 0})

	// Pop should return in descending weight order (LPT)
	weights := make([]int, 0, 4)
	for h.Len() > 0 {
		item := heap.Pop(h).(workunit.UnitSpec)
		weights = append(weights, item.Weight)
	}

	assert.Equal(t, []int{10, 5, 1, 0}, weights, "items should pop in descending weight order (LPT)")
}

func TestUnitHeap_HeapPop_PreservesMetadata(t *testing.T) {
	// Verify that Index and other fields survive the heap round-trip
	h := &unitHeap{}
	heap.Init(h)

	original := workunit.UnitSpec{
		ID:     uid("mod1", "test"),
		Weight: 42,
		Index:  7,
	}
	heap.Push(h, original)

	popped := heap.Pop(h).(workunit.UnitSpec)

	assert.Equal(t, "test", popped.ID.ComponentName)
	assert.Equal(t, 42, popped.Weight)
	assert.Equal(t, 7, popped.Index, "Index field should be preserved through heap operations")
}

func TestUnitHeap_HeapRemove_MiddleElement(t *testing.T) {
	h := &unitHeap{
		{ID: uid("mod1", "A"), Weight: 10},
		{ID: uid("mod1", "B"), Weight: 5},
		{ID: uid("mod1", "C"), Weight: 1},
	}
	heap.Init(h)

	// Remove element at index 1
	removed := heap.Remove(h, 1).(workunit.UnitSpec)
	assert.NotNil(t, removed)
	assert.Equal(t, 2, h.Len(), "heap should have 2 elements after removal")

	// Remaining elements should still pop in LPT order
	first := heap.Pop(h).(workunit.UnitSpec)
	second := heap.Pop(h).(workunit.UnitSpec)
	assert.Greater(t, first.Weight, second.Weight,
		"remaining elements should maintain max-heap property")
}

func TestUnitHeap_HeapPush_MaintainsHeapProperty(t *testing.T) {
	h := &unitHeap{}
	heap.Init(h)

	// Push in random order
	weights := []int{3, 7, 1, 9, 4, 6, 2, 8, 5}
	for i, w := range weights {
		heap.Push(h, workunit.UnitSpec{
			ID:     uid("mod1", string(rune('A'+i))),
			Weight: w,
		})
	}

	// Pop should yield strictly descending order
	prev := int(^uint(0) >> 1) // max int
	for h.Len() > 0 {
		item := heap.Pop(h).(workunit.UnitSpec)
		assert.LessOrEqual(t, item.Weight, prev,
			"heap should maintain descending order on pop")
		prev = item.Weight
	}
}

func TestUnitHeap_EqualWeights_AllPopped(t *testing.T) {
	// When all weights are equal, every element should still be popped
	h := &unitHeap{}
	heap.Init(h)

	for i := 0; i < 5; i++ {
		heap.Push(h, workunit.UnitSpec{
			ID:     uid("mod1", string(rune('A'+i))),
			Weight: 3,
			Index:  i,
		})
	}

	count := 0
	for h.Len() > 0 {
		item := heap.Pop(h).(workunit.UnitSpec)
		assert.Equal(t, 3, item.Weight)
		count++
	}
	assert.Equal(t, 5, count, "all equal-weight items should be popped")
}

func TestUnitHeap_EmptyHeap_LenIsZero(t *testing.T) {
	h := &unitHeap{}
	heap.Init(h)

	assert.Equal(t, 0, h.Len())
}

func TestUnitHeap_SingleElement_PushAndPop(t *testing.T) {
	h := &unitHeap{}
	heap.Init(h)

	heap.Push(h, workunit.UnitSpec{ID: uid("mod1", "solo"), Weight: 99})
	assert.Equal(t, 1, h.Len())

	item := heap.Pop(h).(workunit.UnitSpec)
	assert.Equal(t, "solo", item.ID.ComponentName)
	assert.Equal(t, 99, item.Weight)
	assert.Equal(t, 0, h.Len())
}
