package console

import "time"

// Phase represents the three execution phases
type Phase int

const (
	// PhaseInit is the initialization phase (discovery, setup, dependency verification)
	PhaseInit Phase = iota
	// PhaseRun is the execution phase (parallel workers running)
	PhaseRun
	// PhaseEnd is the completion phase (summary, results)
	PhaseEnd
)

// String returns the display name for a phase
func (p Phase) String() string {
	switch p {
	case PhaseInit:
		return "Init"
	case PhaseRun:
		return "Run"
	case PhaseEnd:
		return "End"
	default:
		return "Unknown"
	}
}

// PhaseStatus represents the status of a phase
type PhaseStatus int

const (
	// PhasePending means the phase has not started
	PhasePending PhaseStatus = iota
	// PhaseActive means the phase is currently running
	PhaseActive
	// PhaseComplete means the phase finished successfully
	PhaseComplete
	// PhaseFailed means the phase finished with errors
	PhaseFailed
)

// Icon returns the icon for a phase status
func (s PhaseStatus) Icon() string {
	switch s {
	case PhasePending:
		return "○"
	case PhaseActive:
		return "▶"
	case PhaseComplete:
		return "✓"
	case PhaseFailed:
		return "✗"
	default:
		return "?"
	}
}

// Pane represents a single pane in the 3-pane layout
type Pane struct {
	Phase     Phase
	Status    PhaseStatus
	Summary   string      // Collapsed summary text (shown when not active)
	Buffer    *RingBuffer // Output lines for this pane
	StartTime time.Time   // When this phase started
	EndTime   time.Time   // When this phase ended (zero if not complete)
}

// NewPane creates a new pane for the given phase
func NewPane(phase Phase, bufferSize int) *Pane {
	return &Pane{
		Phase:  phase,
		Status: PhasePending,
		Buffer: NewRingBuffer(bufferSize),
	}
}

// Duration returns how long this phase has been running or took to complete
func (p *Pane) Duration() time.Duration {
	if p.StartTime.IsZero() {
		return 0
	}
	if !p.EndTime.IsZero() {
		return p.EndTime.Sub(p.StartTime)
	}
	return time.Since(p.StartTime)
}

// IsExpanded returns true if the pane should be expanded (active or recently completed)
func (p *Pane) IsExpanded() bool {
	return p.Status == PhaseActive
}

// HeaderText returns the header text for the pane
func (p *Pane) HeaderText() string {
	icon := p.Status.Icon()
	name := p.Phase.String()

	if p.Status == PhaseComplete || p.Status == PhaseFailed {
		if p.Summary != "" {
			return icon + " " + name + ": " + p.Summary
		}
	}

	if p.Status == PhasePending {
		return icon + " " + name + ": waiting..."
	}

	return icon + " " + name
}
