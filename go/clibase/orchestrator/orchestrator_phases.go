package orchestrator

import (
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/display"
)

// SetPhase switches to a specific phase (Init, Run, End).
// This emits a PhaseStartedEvent to all registered observers.
func (o *Orchestrator) SetPhase(phase display.Phase) {
	o.emit(core.PhaseStartedEvent{
		Phase:     core.ExecutionPhase(phase),
		Time:      time.Now(),
		TotalWork: o.tuiTotal,
	})
}

// SetPhaseSummary sets the summary text for a phase (shown when collapsed).
// Note: This is handled by observers via phase events.
func (o *Orchestrator) SetPhaseSummary(phase display.Phase, summary string) {
	// Phase summary is now handled via PhaseCompletedEvent
	// This method is kept for backward compatibility
}

// CompletePhase marks a phase as complete.
// This emits a PhaseCompletedEvent to all registered observers.
func (o *Orchestrator) CompletePhase(phase display.Phase, success bool, summary string) {
	o.emit(core.PhaseCompletedEvent{
		Phase:   core.ExecutionPhase(phase),
		Time:    time.Now(),
		Success: success,
		Summary: summary,
	})
}

// WriteToPhase writes a line to a specific phase's buffer.
// This emits an OutputLineEvent to all registered observers.
func (o *Orchestrator) WriteToPhase(phase display.Phase, text string) {
	o.emit(core.OutputLineEvent{
		Time:   time.Now(),
		Source: phase.String(),
		Text:   text,
		Level:  core.OutputLevelInfo,
	})
}
