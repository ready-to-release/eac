// Package iobuffer provides I/O utility types for bounded buffering.
package iobuffer

import (
	"bytes"
	"fmt"
)

// LimitedBuffer is a buffer that limits the amount of data it can hold.
// It implements io.Writer and prevents memory exhaustion from unbounded output
// (e.g., Docker container stdout/stderr).
type LimitedBuffer struct {
	buf   bytes.Buffer
	limit int64
	total int64
}

// NewLimitedBuffer creates a LimitedBuffer with the given byte limit.
func NewLimitedBuffer(limit int64) *LimitedBuffer {
	return &LimitedBuffer{limit: limit}
}

// Write implements io.Writer with size limit.
func (lb *LimitedBuffer) Write(p []byte) (n int, err error) {
	// Check if we're already over the limit
	if lb.total >= lb.limit {
		return 0, fmt.Errorf("buffer limit exceeded (%d bytes)", lb.limit)
	}

	// Calculate how much we can write
	remaining := lb.limit - lb.total
	toWrite := int64(len(p))
	willExceed := toWrite > remaining

	if willExceed {
		toWrite = remaining
	}

	// Write what we can
	n, err = lb.buf.Write(p[:toWrite])
	lb.total += int64(n)

	// If we exceeded the limit (not just reached it), return an error
	if willExceed {
		return n, fmt.Errorf("buffer limit exceeded (%d bytes)", lb.limit)
	}

	return n, err
}

// String returns the buffer contents as a string.
func (lb *LimitedBuffer) String() string {
	return lb.buf.String()
}

// Total returns the total number of bytes written (including partial writes).
func (lb *LimitedBuffer) Total() int64 {
	return lb.total
}
