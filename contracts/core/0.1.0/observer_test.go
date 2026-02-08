package core

import (
	"testing"
	"time"
)

func TestExecutionPhaseString(t *testing.T) {
	tests := []struct {
		phase    ExecutionPhase
		expected string
	}{
		{PhaseInit, "Initialization"},
		{PhaseRun, "Run"},
		{PhaseSummary, "Summary"},
		{ExecutionPhase(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.phase.String(); got != tt.expected {
				t.Errorf("ExecutionPhase(%d).String() = %q, want %q", tt.phase, got, tt.expected)
			}
		})
	}
}

func TestUnitStateString(t *testing.T) {
	tests := []struct {
		state    UnitState
		expected string
	}{
		{UnitStateQueued, "queued"},
		{UnitStateRunning, "running"},
		{UnitStateCompleted, "completed"},
		{UnitStateFailed, "failed"},
		{UnitStateCached, "cached"},
		{UnitState(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("UnitState(%d).String() = %q, want %q", tt.state, got, tt.expected)
			}
		})
	}
}

func TestPhaseStartedEvent(t *testing.T) {
	now := time.Now()
	e := PhaseStartedEvent{
		Phase:     PhaseRun,
		Time:      now,
		TotalWork: 10,
	}

	if got := e.EventType(); got != "phase_started" {
		t.Errorf("EventType() = %q, want %q", got, "phase_started")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

func TestPhaseCompletedEvent(t *testing.T) {
	now := time.Now()
	e := PhaseCompletedEvent{
		Phase:    PhaseRun,
		Time:     now,
		Success:  true,
		Summary:  "All passed",
		Duration: 5 * time.Second,
	}

	if got := e.EventType(); got != "phase_completed" {
		t.Errorf("EventType() = %q, want %q", got, "phase_completed")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

func TestUnitQueuedEvent(t *testing.T) {
	now := time.Now()
	e := UnitQueuedEvent{
		Time:        now,
		ID:          "build:eac-core:go:go",
		DisplayName: "eac-core:go",
		Weight:      4,
	}

	if got := e.EventType(); got != "unit_queued" {
		t.Errorf("EventType() = %q, want %q", got, "unit_queued")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

func TestEventImplementsInterface(t *testing.T) {
	now := time.Now()

	events := []ExecutionEvent{
		PhaseStartedEvent{Time: now},
		PhaseCompletedEvent{Time: now},
		UnitQueuedEvent{Time: now},
		UnitStartedEvent{Time: now},
		UnitCompletedEvent{Time: now},
		ProgressUpdateEvent{Time: now},
		ResourceStatusEvent{Time: now},
		ToolStatusEvent{Time: now},
		OutputLineEvent{Time: now},
		SummaryReadyEvent{Time: now},
		InitSummaryEvent{Time: now},
	}

	for _, e := range events {
		if e.EventType() == "" {
			t.Errorf("Event %T has empty EventType()", e)
		}
		if e.Timestamp().IsZero() {
			t.Errorf("Event %T has zero Timestamp()", e)
		}
	}
}
