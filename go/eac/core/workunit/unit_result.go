package workunit

import "time"

// UnitResult represents the outcome of executing a unit of work.
type UnitResult struct {
	ID       UnitID            // The work unit that was executed
	ExitCode int               // 0=success, -1=cached/skipped, >0=failure
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
