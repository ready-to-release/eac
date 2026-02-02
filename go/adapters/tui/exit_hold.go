package tui

import (
	"context"
	"sync"
	"time"
)

// ExitHoldController provides controlled exit timing for TUI adapters.
// TUIs can signal "hold exit" when the user is interacting with output,
// and the command framework waits until holds are released or timeout.
type ExitHoldController struct {
	mu       sync.Mutex
	holds    int
	released chan struct{}
}

// NewExitHoldController creates a new ExitHoldController.
// Initial state is released (no holds active).
func NewExitHoldController() *ExitHoldController {
	c := &ExitHoldController{
		released: make(chan struct{}),
	}
	close(c.released) // Start in released state (no holds)
	return c
}

// HoldExit signals that exit should be delayed.
// Returns a release function that must be called when done.
// The release function is safe to call multiple times.
func (c *ExitHoldController) HoldExit() func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If this is the first hold, create a new channel
	if c.holds == 0 {
		c.released = make(chan struct{})
	}
	c.holds++

	released := false
	return func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		if released {
			return // Already released, idempotent
		}
		released = true

		c.holds--
		if c.holds == 0 {
			close(c.released)
		}
	}
}

// WaitForRelease blocks until all holds are released OR timeout expires.
// Returns true if released cleanly, false on timeout or context cancellation.
// If no holds are active, returns immediately with true.
func (c *ExitHoldController) WaitForRelease(ctx context.Context, timeout time.Duration) bool {
	c.mu.Lock()
	if c.holds == 0 {
		c.mu.Unlock()
		return true
	}
	ch := c.released
	c.mu.Unlock()

	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		return false
	case <-ctx.Done():
		return false
	}
}

// HasActiveHolds returns true if there are any active holds.
func (c *ExitHoldController) HasActiveHolds() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.holds > 0
}
