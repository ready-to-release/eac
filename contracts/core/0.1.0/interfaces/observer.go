package interfaces

import (
	"io"
	"time"
)

// ExecutionPhase represents the phase of execution.
type ExecutionPhase int

const (
	PhaseInit ExecutionPhase = iota
	PhaseRun
	PhaseSummary
)

// String returns the display name for a phase.
func (p ExecutionPhase) String() string {
	switch p {
	case PhaseInit:
		return "Initialization"
	case PhaseRun:
		return "Run"
	case PhaseSummary:
		return "Summary"
	default:
		return "Unknown"
	}
}

// UnitState represents the state of a work unit.
type UnitState int

const (
	UnitStateQueued UnitState = iota
	UnitStateRunning
	UnitStateCompleted
	UnitStateFailed
	UnitStateCached
)

// String returns the display name for a unit state.
func (s UnitState) String() string {
	switch s {
	case UnitStateQueued:
		return "queued"
	case UnitStateRunning:
		return "running"
	case UnitStateCompleted:
		return "completed"
	case UnitStateFailed:
		return "failed"
	case UnitStateCached:
		return "cached"
	default:
		return "unknown"
	}
}

// ExecutionEvent is the base interface for all execution events.
// Events are immutable value types that describe what happened.
type ExecutionEvent interface {
	EventType() string
	Timestamp() time.Time
}

// PhaseStartedEvent is emitted when an execution phase begins.
type PhaseStartedEvent struct {
	Phase     ExecutionPhase
	Time      time.Time
	TotalWork int // Total units of work (for Run phase)
}

func (e PhaseStartedEvent) EventType() string    { return "phase_started" }
func (e PhaseStartedEvent) Timestamp() time.Time { return e.Time }

// PhaseCompletedEvent is emitted when an execution phase ends.
type PhaseCompletedEvent struct {
	Phase    ExecutionPhase
	Time     time.Time
	Success  bool
	Summary  string
	Duration time.Duration
}

func (e PhaseCompletedEvent) EventType() string    { return "phase_completed" }
func (e PhaseCompletedEvent) Timestamp() time.Time { return e.Time }

// UnitQueuedEvent is emitted when a work unit is scheduled for execution.
type UnitQueuedEvent struct {
	Time        time.Time
	ID          string // Globally unique ID (Longname: context:module:component:tool)
	DisplayName string // Short display name (module:component)
	Module      string
	Component   string
	Handler     string
	Weight      int
	Layer       int // 0 if not using layers
}

func (e UnitQueuedEvent) EventType() string    { return "unit_queued" }
func (e UnitQueuedEvent) Timestamp() time.Time { return e.Time }

// UnitStartedEvent is emitted when a work unit begins execution.
type UnitStartedEvent struct {
	Time time.Time
	ID   string
}

func (e UnitStartedEvent) EventType() string    { return "unit_started" }
func (e UnitStartedEvent) Timestamp() time.Time { return e.Time }

// UnitCompletedEvent is emitted when a work unit finishes execution.
type UnitCompletedEvent struct {
	Time      time.Time
	ID        string
	ExitCode  int           // 0=success, >0=failure, -1=cached
	Duration  time.Duration
	LogPath   string
	CacheTime time.Time // When cached artifact was built (zero if not cached)
}

func (e UnitCompletedEvent) EventType() string    { return "unit_completed" }
func (e UnitCompletedEvent) Timestamp() time.Time { return e.Time }

// ProgressUpdateEvent is emitted periodically during execution.
type ProgressUpdateEvent struct {
	Time         time.Time
	Running      []string // IDs of currently running units
	Completed    int
	Total        int
	CurrentLayer int // 0 if not using layers
	TotalLayers  int // 0 if not using layers

	// Capacity tracking (three-value model)
	Roof           int // Hard ceiling - actual peak allocation (workers spawned at start)
	PressureTarget int // Dynamic optimal capacity (may be < Roof under memory pressure)
}

func (e ProgressUpdateEvent) EventType() string    { return "progress_update" }
func (e ProgressUpdateEvent) Timestamp() time.Time { return e.Time }

// ResourceStatusEvent is emitted when resource usage changes.
// This replaces the TUI-specific LockStatus with a generic representation.
type ResourceStatusEvent struct {
	Time      time.Time
	Resources []ResourceInfo
}

func (e ResourceStatusEvent) EventType() string    { return "resource_status" }
func (e ResourceStatusEvent) Timestamp() time.Time { return e.Time }

// ResourceInfo describes a resource constraint.
type ResourceInfo struct {
	Name     string // e.g., "cpu-scheduler", "module:books"
	Type     string // "semaphore", "weighted", "filelock"
	Capacity int
	Used     int
	Waiting  int
}

// ToolStatusEvent is emitted when tool usage changes.
type ToolStatusEvent struct {
	Time             time.Time
	ActiveContainers []string
	UsedContainers   []string
	ActiveSystem     []string
	UsedSystem       []string
}

func (e ToolStatusEvent) EventType() string    { return "tool_status" }
func (e ToolStatusEvent) Timestamp() time.Time { return e.Time }

// OutputLineEvent is emitted for each line of output from a unit.
type OutputLineEvent struct {
	Time   time.Time
	Source string // Unit ID or "system"
	Text   string
	Level  OutputLevel
}

func (e OutputLineEvent) EventType() string    { return "output_line" }
func (e OutputLineEvent) Timestamp() time.Time { return e.Time }

// OutputLevel represents output severity.
type OutputLevel int

const (
	OutputLevelInfo OutputLevel = iota
	OutputLevelWarn
	OutputLevelError
)

// SummaryReadyEvent is emitted when execution summary is available.
type SummaryReadyEvent struct {
	Time      time.Time
	Success   bool
	TotalTime time.Duration
	Passed    int
	Failed    int
	Skipped   int
	Cached    int
	Details   []string
	NextSteps string
}

func (e SummaryReadyEvent) EventType() string    { return "summary_ready" }
func (e SummaryReadyEvent) Timestamp() time.Time { return e.Time }

// InitSummaryEvent provides structured init phase information.
// This replaces tui.InitSummary with a generic representation.
type InitSummaryEvent struct {
	Time             time.Time
	Command          string
	ExecutionContext string // local/CI/container
	RequestedModules int
	ResolvedModules  int
	TotalUnits       int
	Layers           []LayerInfo
	Parallelism      ParallelismInfo
	Flags            FlagsInfo
}

func (e InitSummaryEvent) EventType() string    { return "init_summary" }
func (e InitSummaryEvent) Timestamp() time.Time { return e.Time }

// LayerInfo describes an execution layer.
type LayerInfo struct {
	Modules []ModuleInfo
}

// ModuleInfo describes a module within a layer.
type ModuleInfo struct {
	Name  string
	Units []UnitInfo
}

// UnitInfo describes a unit of work.
type UnitInfo struct {
	ID          string
	DisplayName string
	Weight      int
}

// ParallelismInfo describes parallelism configuration.
type ParallelismInfo struct {
	Mode             string // "ci" or "devbox"
	EffectiveWorkers int
	TurboBoost       int
	WeightedCapacity int
}

// FlagsInfo captures relevant execution flags.
type FlagsInfo struct {
	TidyFirst    bool
	ForceRebuild bool
	DryRun       bool
	UseTUI       bool
	SkipDeps     bool
	SkipDepm     bool
}

// ExecutionObserver receives execution events.
// Implementations must be thread-safe - events may arrive concurrently.
type ExecutionObserver interface {
	// OnEvent is called for each execution event.
	// Implementations should be non-blocking.
	OnEvent(event ExecutionEvent)
}

// WriterFactory creates writers for capturing unit output.
// This allows observers to intercept output streams.
type WriterFactory interface {
	// NewWriter creates a writer for a unit's output.
	// The returned writer tees to both the observer and logWriter.
	// Callers must Close the returned writer when done.
	NewWriter(unitID string, logWriter io.Writer) io.WriteCloser
}

// ObservableExecutor emits events during execution.
// Implemented by Orchestrator and UnitScheduler.
type ObservableExecutor interface {
	// AddObserver registers an observer to receive events.
	AddObserver(observer ExecutionObserver)

	// RemoveObserver unregisters an observer.
	RemoveObserver(observer ExecutionObserver)

	// SetWriterFactory sets the factory for creating output writers.
	// If not set, output goes only to log files.
	SetWriterFactory(factory WriterFactory)
}
