package interfaces

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
		Module:      "eac-core",
		Component:   "go",
		Handler:     "go",
		Weight:      4,
	}

	if got := e.EventType(); got != "unit_queued" {
		t.Errorf("EventType() = %q, want %q", got, "unit_queued")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

func TestUnitStartedEvent(t *testing.T) {
	now := time.Now()
	e := UnitStartedEvent{
		Time: now,
		ID:   "build:eac-core:go:go",
	}

	if got := e.EventType(); got != "unit_started" {
		t.Errorf("EventType() = %q, want %q", got, "unit_started")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

func TestUnitCompletedEvent(t *testing.T) {
	now := time.Now()
	cacheTime := now.Add(-time.Hour)
	e := UnitCompletedEvent{
		Time:      now,
		ID:        "build:eac-core:go:go",
		ExitCode:  0,
		Duration:  2 * time.Second,
		LogPath:   "out/build/eac-core/go/build.log",
		CacheTime: cacheTime,
	}

	if got := e.EventType(); got != "unit_completed" {
		t.Errorf("EventType() = %q, want %q", got, "unit_completed")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

func TestProgressUpdateEvent(t *testing.T) {
	now := time.Now()
	e := ProgressUpdateEvent{
		Time:      now,
		Running:   []string{"build:eac-core:go:go", "build:eac-cli:go:go"},
		Completed: 5,
		Total:     10,
	}

	if got := e.EventType(); got != "progress_update" {
		t.Errorf("EventType() = %q, want %q", got, "progress_update")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

func TestResourceStatusEvent(t *testing.T) {
	now := time.Now()
	e := ResourceStatusEvent{
		Time: now,
		Resources: []ResourceInfo{
			{Name: "cpu-scheduler", Type: "weighted", Capacity: 8, Used: 6, Waiting: 2},
		},
	}

	if got := e.EventType(); got != "resource_status" {
		t.Errorf("EventType() = %q, want %q", got, "resource_status")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

func TestToolStatusEvent(t *testing.T) {
	now := time.Now()
	e := ToolStatusEvent{
		Time:             now,
		ActiveContainerTools: []string{"docker"},
		UsedContainerTools:   []string{"docker", "podman"},
		ActiveSystem:     []string{"go"},
		UsedSystem:       []string{"go", "golangci-lint"},
	}

	if got := e.EventType(); got != "tool_status" {
		t.Errorf("EventType() = %q, want %q", got, "tool_status")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

func TestOutputLineEvent(t *testing.T) {
	now := time.Now()
	e := OutputLineEvent{
		Time:   now,
		Source: "build:eac-core:go:go",
		Text:   "Building...",
		Level:  OutputLevelInfo,
	}

	if got := e.EventType(); got != "output_line" {
		t.Errorf("EventType() = %q, want %q", got, "output_line")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

func TestSummaryReadyEvent(t *testing.T) {
	now := time.Now()
	e := SummaryReadyEvent{
		Time:      now,
		Success:   true,
		TotalTime: 30 * time.Second,
		Passed:    8,
		Failed:    0,
		Skipped:   1,
		Cached:    2,
		Details:   []string{"All tests passed"},
		NextSteps: "Run: eac release",
	}

	if got := e.EventType(); got != "summary_ready" {
		t.Errorf("EventType() = %q, want %q", got, "summary_ready")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

func TestInitSummaryEvent(t *testing.T) {
	now := time.Now()
	e := InitSummaryEvent{
		Time:             now,
		Command:          "build",
		ExecutionContext: "local",
		RequestedModules: 2,
		ResolvedModules:  5,
		TotalUnits:       12,
		Modules: []ModuleInfo{
			{Name: "eac-core", Units: []UnitInfo{
				{ID: "build:eac-core:go:go", DisplayName: "eac-core:go", Weight: 4},
			}},
		},
		Parallelism: ParallelismInfo{
			Mode:             "devbox",
			EffectiveWorkers: 8,
			TurboBoost:       125,
			WeightedCapacity: 32,
		},
		Flags: FlagsInfo{
			TidyFirst:    true,
			ForceRebuild: false,
			DryRun:       false,
			UseTUI:       true,
			SkipDeps:     false,
			SkipDepm:     false,
		},
	}

	if got := e.EventType(); got != "init_summary" {
		t.Errorf("EventType() = %q, want %q", got, "init_summary")
	}
	if got := e.Timestamp(); !got.Equal(now) {
		t.Errorf("Timestamp() = %v, want %v", got, now)
	}
}

// TestEventImplementsInterface verifies all event types implement ExecutionEvent.
func TestEventImplementsInterface(t *testing.T) {
	now := time.Now()

	// This test will fail to compile if any event type doesn't implement ExecutionEvent
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

// mockObserver is a test implementation of ExecutionObserver.
type mockObserver struct {
	events []ExecutionEvent
}

func (m *mockObserver) OnEvent(event ExecutionEvent) {
	m.events = append(m.events, event)
}

func TestMockObserverReceivesEvents(t *testing.T) {
	obs := &mockObserver{}
	now := time.Now()

	// Simulate sending events to observer
	obs.OnEvent(PhaseStartedEvent{Phase: PhaseInit, Time: now})
	obs.OnEvent(UnitQueuedEvent{Time: now, ID: "test:module:component:tool"})

	if len(obs.events) != 2 {
		t.Errorf("Observer received %d events, want 2", len(obs.events))
	}
}
