package workunit

import "time"

// ExitCodeTimeout is the exit code used when a worker is killed due to timeout.
// Matches the Unix timeout(1) convention.
const ExitCodeTimeout = 124

// UnitResult represents the outcome of executing a unit of work.
type UnitResult struct {
	ID       UnitID            // The work unit that was executed
	ExitCode int               // 0=success, -1=cached/skipped, 124=timeout, >0=failure
	Duration time.Duration     // How long execution took
	LogPath  string            // Path to the execution log
	Metrics  map[string]any    // Context-specific metrics (tests_passed, issues_found, etc.)
}

// Success returns true if the unit executed successfully.
func (r UnitResult) Success() bool {
	return r.ExitCode == 0
}

// Cached returns true if the unit was skipped due to cache hit.
func (r UnitResult) Cached() bool {
	return r.ExitCode == -1
}

// Failed returns true if the unit execution failed.
func (r UnitResult) Failed() bool {
	return r.ExitCode > 0
}

// TimedOut returns true if the unit was killed due to timeout.
func (r UnitResult) TimedOut() bool {
	return r.ExitCode == ExitCodeTimeout
}
