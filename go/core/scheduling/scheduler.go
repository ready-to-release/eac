package scheduling

import "github.com/ready-to-release/eac/go/core/workunit"

// Stats provides scheduling statistics for monitoring.
type Stats struct {
	Total     int // Total items originally added to scheduler
	Pending   int // Items still in queue (not yet popped)
	Ready     int // Items ready to run (deps satisfied, still in queue)
	Blocked   int // Items waiting on dependencies (still in queue)
	Completed int // Items that have been marked complete
	Failed    int // Items that have been marked failed
}

// WorkScheduler provides pull-based work scheduling with LPT ordering.
//
// In-flight tracking is IMPLICIT:
//   - When Next() or WaitForReady() returns an item, it's removed from the queue (now "in-flight")
//   - The item stays in-flight until MarkComplete() or MarkFailed() is called
//   - Dependencies are only unblocked when MarkComplete() is called
//
// Thread-safe for concurrent access from multiple workers.
type WorkScheduler interface {
	// Next returns the heaviest ready work unit (LPT scheduling).
	// "Ready" means all dependencies are complete.
	// Returns nil immediately if no items are ready.
	// Does NOT block - use WaitForReady for blocking behavior.
	//
	// The returned item is removed from the queue (now "in-flight").
	// Caller MUST call MarkComplete or MarkFailed when done.
	Next() *workunit.UnitSpec

	// WaitForReady blocks until a ready item is available, then returns it.
	// Returns nil when queue is exhausted (all items completed) or scheduler is closed.
	//
	// The returned item is removed from the queue (now "in-flight").
	// Caller MUST call MarkComplete or MarkFailed when done.
	WaitForReady() *workunit.UnitSpec

	// MarkComplete signals successful completion of a unit.
	// This unblocks any units that depend on this one.
	// Must be called exactly once per item returned by Next/WaitForReady.
	MarkComplete(id workunit.UnitID)

	// MarkFailed signals failure of a unit.
	// Dependent units will report HasFailedDependency() = true.
	// Must be called exactly once per item returned by Next/WaitForReady.
	MarkFailed(id workunit.UnitID)

	// HasFailedDependency returns true if any dependency of the unit failed.
	// Used to skip execution and propagate failures.
	HasFailedDependency(id workunit.UnitID) bool

	// Stats returns current scheduler statistics.
	Stats() Stats

	// Close releases scheduler resources and wakes any blocked WaitForReady calls.
	Close()
}
