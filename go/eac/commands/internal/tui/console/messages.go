package console

import "time"

// lineMsg is sent when a new output line is received.
type lineMsg Line

// statusMsg is sent when the orchestrator status changes.
type statusMsg Status

// tickMsg is sent periodically for time updates.
type tickMsg time.Time

// linesDoneMsg indicates the line channel is closed.
type linesDoneMsg struct{}

// statusDoneMsg indicates the status channel is closed.
type statusDoneMsg struct{}

// completedMsg indicates a module has completed.
type completedMsg struct {
	Moniker  string
	ExitCode int
	Duration time.Duration
}

// Status represents a status update from the orchestrator.
type Status struct {
	Phase       string
	Running     []string
	Completed   int
	Total       int
	Layer       int // Current layer being executed (1-indexed, 0 = not using layers)
	TotalLayers int // Total number of layers (0 = not using layers)

	// Detailed lock tracking info (from locktracker.Registry)
	Locks []LockStatus // Individual lock states
}

// LockStatus represents the state of a single lock.
type LockStatus struct {
	Name     string // Lock name (e.g., "component-scheduler", "module:books")
	Type     string // Lock type: "semaphore", "weighted", "filelock"
	Capacity int    // Total capacity (for semaphores/weighted)
	Used     int    // Currently in use
	Waiting  int    // Waiting for this lock
}

// PhaseUpdateMsg is sent when a phase changes state (exported for tui package).
type PhaseUpdateMsg struct {
	Phase   Phase       // Which phase this update is for
	Status  PhaseStatus // New status (active, complete, failed)
	Summary string      // Summary text for collapsed view
}

// PhaseLineMsg is sent when a line should go to a specific phase's buffer (exported for tui package).
type PhaseLineMsg struct {
	Phase Phase
	Line  Line
}

// ResultLineMsg is sent when a line should go to the results buffer (exported for tui package).
type ResultLineMsg struct {
	Line Line
}

// SummaryDataMsg is sent to populate and activate the Summary pane (exported for tui package).
type SummaryDataMsg struct {
	Data *SummaryData
}

// ModuleStartMsg is sent when a module starts execution (for tab tracking).
// This creates the tab in pending state (scheduled but waiting for slot).
type ModuleStartMsg struct {
	Moniker string
	Weight  int // Scheduling weight/pressure for this module
}

// ModuleRunningMsg is sent when a module acquires its execution slot.
// This transitions the tab from pending to running state.
type ModuleRunningMsg struct {
	Moniker string
}

// ModuleCompleteMsg is sent when a module completes execution.
type ModuleCompleteMsg struct {
	Moniker  string
	ExitCode int
}

// TabSelectMsg is sent when user clicks on a tab.
type TabSelectMsg struct {
	Moniker string // Empty string = aggregate view
}

// TabDecayMsg is sent periodically to clean up decayed tabs.
type TabDecayMsg struct{}
