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
	Phase     string
	Running   []string
	Completed int
	Total     int
}

// PhaseUpdateMsg is sent when a phase changes state (exported for tui package)
type PhaseUpdateMsg struct {
	Phase   Phase       // Which phase this update is for
	Status  PhaseStatus // New status (active, complete, failed)
	Summary string      // Summary text for collapsed view
}

// PhaseLineMsg is sent when a line should go to a specific phase's buffer (exported for tui package)
type PhaseLineMsg struct {
	Phase Phase
	Line  Line
}

// ResultLineMsg is sent when a line should go to the results buffer (exported for tui package)
type ResultLineMsg struct {
	Line Line
}

// SummaryDataMsg is sent to populate and activate the Summary pane (exported for tui package)
type SummaryDataMsg struct {
	Data *SummaryData
}

// autoExitTimerMsg is sent when the auto-exit timer expires
type autoExitTimerMsg struct{}
