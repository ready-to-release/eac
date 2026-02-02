package tui

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
)

// observerMockConsole is a test mock that records all method calls.
type observerMockConsole struct {
	phaseCalls          []Phase
	phaseCompleteCalls  []phaseCompleteCall
	uowStartCalls    []moduleStartCall
	uowRunningCalls  []string
	uowCompleteCalls []moduleCompleteCall
	statusCalls         []Status
	lineCalls           []Line
	summaryCalls        []*SummaryData
	initSummaryCalls    []*InitSummary
	resultLines         []string
	phaseLines          []phaseLineCall
}

type phaseCompleteCall struct {
	phase   Phase
	success bool
	summary string
}

type moduleStartCall struct {
	moniker     string
	displayName string
	weight      int
}

type moduleCompleteCall struct {
	moniker   string
	exitCode  int
	cacheTime time.Time
	logPath   string
}

type phaseLineCall struct {
	phase Phase
	text  string
}

func newMockConsole() *observerMockConsole {
	return &observerMockConsole{}
}

// Console interface implementations
func (m *observerMockConsole) Start(_ context.Context) error { return nil }
func (m *observerMockConsole) StartAsync(_ context.Context)  {}
func (m *observerMockConsole) Wait()                            {}
func (m *observerMockConsole) Stop()                            {}
func (m *observerMockConsole) NewWriter(source string, logWriter io.Writer) io.Writer {
	return logWriter // Just pass through for testing
}
func (m *observerMockConsole) SendLine(line Line)              { m.lineCalls = append(m.lineCalls, line) }
func (m *observerMockConsole) WriteResult(text string)         { m.resultLines = append(m.resultLines, text) }
func (m *observerMockConsole) UpdateStatus(status Status)      { m.statusCalls = append(m.statusCalls, status) }
func (m *observerMockConsole) SetPhase(phase Phase)            { m.phaseCalls = append(m.phaseCalls, phase) }
func (m *observerMockConsole) CompletePhase(phase Phase, success bool, summary string) {
	m.phaseCompleteCalls = append(m.phaseCompleteCalls, phaseCompleteCall{phase, success, summary})
}
func (m *observerMockConsole) WriteToPhase(phase Phase, text string) {
	m.phaseLines = append(m.phaseLines, phaseLineCall{phase, text})
}
func (m *observerMockConsole) SetPhaseSummary(phase Phase, summary string) {}
func (m *observerMockConsole) StartUoW(moniker, displayName string, weight int) {
	m.uowStartCalls = append(m.uowStartCalls, moduleStartCall{moniker, displayName, weight})
}
func (m *observerMockConsole) MarkUoWRunning(moniker string) {
	m.uowRunningCalls = append(m.uowRunningCalls, moniker)
}
func (m *observerMockConsole) MarkUoWComplete(moniker string, exitCode int) {
	m.uowCompleteCalls = append(m.uowCompleteCalls, moduleCompleteCall{moniker, exitCode, time.Time{}, ""})
}
func (m *observerMockConsole) MarkUoWCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string) {
	m.uowCompleteCalls = append(m.uowCompleteCalls, moduleCompleteCall{moniker, exitCode, cacheTime, logPath})
}
func (m *observerMockConsole) SendSummary(data *SummaryData) {
	m.summaryCalls = append(m.summaryCalls, data)
}
func (m *observerMockConsole) SetInitSummary(summary *InitSummary) {
	m.initSummaryCalls = append(m.initSummaryCalls, summary)
}

func TestTUIObserverImplementsExecutionObserver(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	// Verify it implements ExecutionObserver
	var _ interfaces.ExecutionObserver = observer
}

func TestTUIObserverImplementsWriterFactory(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	// Verify it implements WriterFactory
	var _ interfaces.WriterFactory = observer
}

func TestTUIObserverPhaseStartedEvent(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	now := time.Now()
	observer.OnEvent(interfaces.PhaseStartedEvent{
		Phase:     interfaces.PhaseRun,
		Time:      now,
		TotalWork: 10,
	})

	if len(mock.phaseCalls) != 1 {
		t.Fatalf("Expected 1 SetPhase call, got %d", len(mock.phaseCalls))
	}
	if mock.phaseCalls[0] != PhaseRun {
		t.Errorf("Expected PhaseRun, got %v", mock.phaseCalls[0])
	}
}

func TestTUIObserverPhaseCompletedEvent(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	now := time.Now()
	observer.OnEvent(interfaces.PhaseCompletedEvent{
		Phase:    interfaces.PhaseInit,
		Time:     now,
		Success:  true,
		Summary:  "Initialization complete",
		Duration: 5 * time.Second,
	})

	if len(mock.phaseCompleteCalls) != 1 {
		t.Fatalf("Expected 1 CompletePhase call, got %d", len(mock.phaseCompleteCalls))
	}
	call := mock.phaseCompleteCalls[0]
	if call.phase != PhaseInit {
		t.Errorf("Expected PhaseInit, got %v", call.phase)
	}
	if !call.success {
		t.Error("Expected success=true")
	}
	if call.summary != "Initialization complete" {
		t.Errorf("Expected 'Initialization complete', got %q", call.summary)
	}
}

func TestTUIObserverUnitQueuedEvent(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	now := time.Now()
	observer.OnEvent(interfaces.UnitQueuedEvent{
		Time:        now,
		ID:          "build:core:go:go",
		DisplayName: "core:go",
		Module:      "core",
		Component:   "go",
		Handler:     "go",
		Weight:      4,
		Layer:       1,
	})

	if len(mock.uowStartCalls) != 1 {
		t.Fatalf("Expected 1 StartUoW call, got %d", len(mock.uowStartCalls))
	}
	call := mock.uowStartCalls[0]
	if call.moniker != "build:core:go:go" {
		t.Errorf("Expected moniker 'build:core:go:go', got %q", call.moniker)
	}
	if call.displayName != "core:go" {
		t.Errorf("Expected displayName 'core:go', got %q", call.displayName)
	}
	if call.weight != 4 {
		t.Errorf("Expected weight 4, got %d", call.weight)
	}
}

func TestTUIObserverUnitStartedEvent(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	now := time.Now()
	observer.OnEvent(interfaces.UnitStartedEvent{
		Time: now,
		ID:   "build:core:go:go",
	})

	if len(mock.uowRunningCalls) != 1 {
		t.Fatalf("Expected 1 MarkUoWRunning call, got %d", len(mock.uowRunningCalls))
	}
	if mock.uowRunningCalls[0] != "build:core:go:go" {
		t.Errorf("Expected 'build:core:go:go', got %q", mock.uowRunningCalls[0])
	}
}

func TestTUIObserverUnitCompletedEvent(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	now := time.Now()
	cacheTime := now.Add(-time.Hour)
	observer.OnEvent(interfaces.UnitCompletedEvent{
		Time:      now,
		ID:        "build:core:go:go",
		ExitCode:  -1,
		Duration:  2 * time.Second,
		LogPath:   "out/build/core/go/build.log",
		CacheTime: cacheTime,
	})

	if len(mock.uowCompleteCalls) != 1 {
		t.Fatalf("Expected 1 MarkUoWComplete call, got %d", len(mock.uowCompleteCalls))
	}
	call := mock.uowCompleteCalls[0]
	if call.moniker != "build:core:go:go" {
		t.Errorf("Expected 'build:core:go:go', got %q", call.moniker)
	}
	if call.exitCode != -1 {
		t.Errorf("Expected exitCode -1, got %d", call.exitCode)
	}
	if !call.cacheTime.Equal(cacheTime) {
		t.Errorf("Expected cacheTime %v, got %v", cacheTime, call.cacheTime)
	}
	if call.logPath != "out/build/core/go/build.log" {
		t.Errorf("Expected logPath 'out/build/core/go/build.log', got %q", call.logPath)
	}
}

func TestTUIObserverProgressUpdateEvent(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	now := time.Now()
	observer.OnEvent(interfaces.ProgressUpdateEvent{
		Time:         now,
		Running:      []string{"build:core:go:go"},
		Completed:    5,
		Total:        10,
		CurrentLayer: 2,
		TotalLayers:  3,
	})

	if len(mock.statusCalls) != 1 {
		t.Fatalf("Expected 1 UpdateStatus call, got %d", len(mock.statusCalls))
	}
	status := mock.statusCalls[0]
	if len(status.Running) != 1 || status.Running[0] != "build:core:go:go" {
		t.Errorf("Unexpected running: %v", status.Running)
	}
	if status.Completed != 5 {
		t.Errorf("Expected Completed 5, got %d", status.Completed)
	}
	if status.Total != 10 {
		t.Errorf("Expected Total 10, got %d", status.Total)
	}
	if status.Layer != 2 {
		t.Errorf("Expected Layer 2, got %d", status.Layer)
	}
	if status.TotalLayers != 3 {
		t.Errorf("Expected TotalLayers 3, got %d", status.TotalLayers)
	}
}

func TestTUIObserverResourceStatusEvent(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	now := time.Now()
	observer.OnEvent(interfaces.ResourceStatusEvent{
		Time: now,
		Resources: []interfaces.ResourceInfo{
			{Name: "cpu-scheduler", Type: "weighted", Capacity: 8, Used: 6, Waiting: 2},
		},
	})

	if len(mock.statusCalls) != 1 {
		t.Fatalf("Expected 1 UpdateStatus call, got %d", len(mock.statusCalls))
	}
	status := mock.statusCalls[0]
	if len(status.Locks) != 1 {
		t.Fatalf("Expected 1 lock, got %d", len(status.Locks))
	}
	lock := status.Locks[0]
	if lock.Name != "cpu-scheduler" {
		t.Errorf("Expected Name 'cpu-scheduler', got %q", lock.Name)
	}
	if lock.Capacity != 8 {
		t.Errorf("Expected Capacity 8, got %d", lock.Capacity)
	}
}

func TestTUIObserverOutputLineEvent(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	now := time.Now()
	observer.OnEvent(interfaces.OutputLineEvent{
		Time:   now,
		Source: "build:core:go:go",
		Text:   "Building...",
		Level:  interfaces.OutputLevelWarn,
	})

	if len(mock.lineCalls) != 1 {
		t.Fatalf("Expected 1 SendLine call, got %d", len(mock.lineCalls))
	}
	line := mock.lineCalls[0]
	if line.Source != "build:core:go:go" {
		t.Errorf("Expected Source 'build:core:go:go', got %q", line.Source)
	}
	if line.Text != "Building..." {
		t.Errorf("Expected Text 'Building...', got %q", line.Text)
	}
	if line.Level != LevelWarn {
		t.Errorf("Expected LevelWarn, got %v", line.Level)
	}
}

func TestTUIObserverSummaryReadyEvent(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	now := time.Now()
	observer.OnEvent(interfaces.SummaryReadyEvent{
		Time:      now,
		Success:   true,
		TotalTime: 30 * time.Second,
		Passed:    8,
		Failed:    0,
		Skipped:   1,
		Cached:    2,
		Details:   []string{"All tests passed"},
		NextSteps: "Run: eac release",
	})

	if len(mock.summaryCalls) != 1 {
		t.Fatalf("Expected 1 SendSummary call, got %d", len(mock.summaryCalls))
	}
	data := mock.summaryCalls[0]
	if !data.Success {
		t.Error("Expected Success=true")
	}
	if data.TotalTime != 30*time.Second {
		t.Errorf("Expected TotalTime 30s, got %v", data.TotalTime)
	}
	if data.NextSteps != "Run: eac release" {
		t.Errorf("Expected NextSteps 'Run: eac release', got %q", data.NextSteps)
	}
}

func TestTUIObserverInitSummaryEvent(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	now := time.Now()
	observer.OnEvent(interfaces.InitSummaryEvent{
		Time:             now,
		Command:          "build",
		ExecutionContext: "local",
		RequestedModules: 2,
		ResolvedModules:  5,
		TotalUnits:       12,
		Layers: []interfaces.LayerInfo{
			{Modules: []interfaces.ModuleInfo{
				{Name: "core", Units: []interfaces.UnitInfo{
					{ID: "build:core:go:go", DisplayName: "core:go", Weight: 4},
				}},
			}},
		},
		Parallelism: interfaces.ParallelismInfo{
			Mode:             "devbox",
			EffectiveWorkers: 8,
			TurboBoost:       125,
			WeightedCapacity: 32,
		},
		Flags: interfaces.FlagsInfo{
			TidyFirst: true,
			UseTUI:    true,
		},
	})

	if len(mock.initSummaryCalls) != 1 {
		t.Fatalf("Expected 1 SetInitSummary call, got %d", len(mock.initSummaryCalls))
	}
	summary := mock.initSummaryCalls[0]
	if summary.Command != "build" {
		t.Errorf("Expected Command 'build', got %q", summary.Command)
	}
	if summary.ExecutionContext != "local" {
		t.Errorf("Expected ExecutionContext 'local', got %q", summary.ExecutionContext)
	}
	if summary.RequestedModules != 2 {
		t.Errorf("Expected RequestedModules 2, got %d", summary.RequestedModules)
	}
	if summary.CalculatedModules != 5 {
		t.Errorf("Expected CalculatedModules 5, got %d", summary.CalculatedModules)
	}
	if summary.UoWCount != 12 {
		t.Errorf("Expected UoWCount 12, got %d", summary.UoWCount)
	}
	if len(summary.ExecutionTree) != 1 {
		t.Errorf("Expected 1 layer, got %d", len(summary.ExecutionTree))
	}
	if summary.ParallelismMode != "devbox" {
		t.Errorf("Expected ParallelismMode 'devbox', got %q", summary.ParallelismMode)
	}
	if summary.EffectiveWorkers != 8 {
		t.Errorf("Expected EffectiveWorkers 8, got %d", summary.EffectiveWorkers)
	}
	if !summary.Flags.TidyFirst {
		t.Error("Expected Flags.TidyFirst=true")
	}
	if !summary.Flags.UseTUI {
		t.Error("Expected Flags.UseTUI=true")
	}
}

func TestTUIObserverNewWriter(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	var buf bytes.Buffer
	writer := observer.NewWriter("test-unit", &buf)

	// Write something and verify it goes to the log writer
	n, err := writer.Write([]byte("test output\n"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 12 {
		t.Errorf("Expected 12 bytes written, got %d", n)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify the buffer received the data
	if buf.String() != "test output\n" {
		t.Errorf("Expected 'test output\\n', got %q", buf.String())
	}
}

func TestTUIObserverGetConsole(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	console := observer.Console()
	if console != mock {
		t.Error("Expected Console() to return the wrapped console")
	}
}

// unknownEvent is a test event type that implements ExecutionEvent but is not handled.
type unknownEvent struct {
	time time.Time
}

func (e unknownEvent) EventType() string    { return "unknown" }
func (e unknownEvent) Timestamp() time.Time { return e.time }

func TestTUIObserverUnknownEventIgnored(t *testing.T) {
	mock := newMockConsole()
	observer := NewTUIObserver(mock)

	// This should not panic - unknown events are simply ignored
	observer.OnEvent(unknownEvent{time: time.Now()})

	// Verify no methods were called
	if len(mock.phaseCalls) != 0 {
		t.Error("Expected no SetPhase calls")
	}
	if len(mock.statusCalls) != 0 {
		t.Error("Expected no UpdateStatus calls")
	}
}
