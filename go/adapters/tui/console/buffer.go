// Package console provides a TUI console component for displaying build/test output.
package console

import (
	"sync"
	"time"
)

// Level represents the severity level of an output line.
type Level int

const (
	// LevelInfo is for normal informational output.
	LevelInfo Level = iota
	// LevelWarn is for warning messages.
	LevelWarn
	// LevelError is for error messages.
	LevelError
)

// Line represents a single output line with metadata.
type Line struct {
	Text      string
	Source    string // Module moniker or "system"
	Level     Level
	Timestamp time.Time
}

// RingBuffer stores the last N lines with thread-safe access.
// Uses a circular buffer pattern for efficient memory usage.
type RingBuffer struct {
	mu       sync.RWMutex
	lines    []Line
	capacity int
	head     int // Next write position
	count    int // Current number of lines
}

// NewRingBuffer creates a new ring buffer with the specified capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 100
	}
	return &RingBuffer{
		lines:    make([]Line, capacity),
		capacity: capacity,
	}
}

// Push adds a line to the buffer, overwriting oldest if full.
func (rb *RingBuffer) Push(line Line) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.lines[rb.head] = line
	rb.head = (rb.head + 1) % rb.capacity
	if rb.count < rb.capacity {
		rb.count++
	}
}

// Last returns the last n lines in chronological order.
func (rb *RingBuffer) Last(n int) []Line {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if n > rb.count {
		n = rb.count
	}
	if n <= 0 {
		return nil
	}

	result := make([]Line, n)
	start := (rb.head - n + rb.capacity) % rb.capacity

	for i := 0; i < n; i++ {
		result[i] = rb.lines[(start+i)%rb.capacity]
	}
	return result
}

// GetRange returns count lines starting from offset from the end.
// offset=0 means most recent lines (same as Last())
// offset=10 means skip the 10 most recent lines, then return count lines
// This is used for scrolling: higher offset = older content.
func (rb *RingBuffer) GetRange(offset, count int) []Line {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 || offset >= rb.count {
		return []Line{}
	}

	// Calculate the actual number of lines we can return
	available := rb.count - offset
	if count > available {
		count = available
	}

	if count <= 0 {
		return []Line{}
	}

	result := make([]Line, count)

	// Start index: go back (offset + count) from head, then forward offset
	// This gives us the range [rb.count - offset - count, rb.count - offset)
	startIdx := (rb.head - offset - count + rb.capacity) % rb.capacity

	for i := 0; i < count; i++ {
		idx := (startIdx + i) % rb.capacity
		result[i] = rb.lines[idx]
	}

	return result
}

// Count returns the current number of lines in the buffer.
func (rb *RingBuffer) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// Clear removes all lines from the buffer.
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.head = 0
	rb.count = 0
}

// All returns all lines in chronological order.
func (rb *RingBuffer) All() []Line {
	return rb.Last(rb.Count())
}

// LastByLevel returns the last n lines filtered by level.
func (rb *RingBuffer) LastByLevel(n int, level Level) []Line {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	// Collect all lines of the specified level
	var filtered []Line
	start := 0
	if rb.count == rb.capacity {
		start = rb.head
	}

	for i := 0; i < rb.count; i++ {
		idx := (start + i) % rb.capacity
		if rb.lines[idx].Level == level {
			filtered = append(filtered, rb.lines[idx])
		}
	}

	// Return last n
	if n > len(filtered) {
		n = len(filtered)
	}
	if n <= 0 {
		return nil
	}

	return filtered[len(filtered)-n:]
}
