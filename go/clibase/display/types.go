package display

import "time"

// Phase represents the three execution phases.
type Phase int

const (
	// PhaseInit is the initialization phase (discovery, setup, dependency verification).
	PhaseInit Phase = iota
	// PhaseRun is the execution phase (parallel workers running).
	PhaseRun
	// PhaseSummary is the summary phase (final results, statistics, next steps).
	PhaseSummary
)

// Phase display names.
const (
	PhaseNameInitialization = "Initialization"
	PhaseNameRun            = "Run"
	PhaseNameSummary        = "Summary"
	PhaseNameUnknown        = "Unknown"
)

// String returns the display name for a phase.
func (p Phase) String() string {
	switch p {
	case PhaseInit:
		return PhaseNameInitialization
	case PhaseRun:
		return PhaseNameRun
	case PhaseSummary:
		return PhaseNameSummary
	default:
		return PhaseNameUnknown
	}
}

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

// Line represents a line of output from a command.
type Line struct {
	Text      string
	Source    string // Module moniker or "system"
	Level     Level
	Timestamp time.Time
}

// Status represents a status update from the orchestrator.
type Status struct {
	Phase     string
	Running   []string
	Completed int
	Total     int

	Roof           int
	PressureTarget int

	Locks []LockStatus

	ActiveContainerTools []string
	UsedContainerTools   []string
	ActiveSystemTools    []string
	UsedSystemTools      []string

	DockerMemPercent float64
	DockerAvailable  bool

	RunningContainerCount int
	TotalContainerCount   int

	RunningSystemCount int
	TotalSystemCount   int
}

// LockStatus represents the state of a single lock.
type LockStatus struct {
	Name     string
	Type     string
	Capacity int
	Used     int
	Waiting  int
}

// SummaryData holds structured summary information for the Summary pane.
type SummaryData struct {
	Success     bool
	TotalTime   time.Duration
	InitSummary string
	RunSummary  string
	Details     []string
	NextSteps   string
}
