package console

import "time"

// lineMsg is sent when a new output line is received.
type lineMsg Line

// batchLineMsg is sent when multiple output lines are received in a batch.
// This reduces Bubbletea Update+View cycles by processing up to 50 lines per message.
type batchLineMsg []Line

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

	// Capacity tracking (three-value model)
	Roof           int // Hard ceiling - actual peak allocation (workers spawned at start)
	PressureTarget int // Dynamic optimal capacity (may be < Roof under memory pressure)

	// Detailed lock tracking info (from locktracker.Registry)
	Locks []LockStatus // Individual lock states

	// Container tools (e.g., "mkdocs-build", "go-lint")
	ActiveContainerTools []string
	UsedContainerTools   []string
	// System tools (e.g., "go", "docker")
	ActiveSystemTools []string
	UsedSystemTools   []string

	// Docker memory metrics
	DockerMemPercent float64 // Docker memory pool usage percentage (0-100)
	DockerAvailable  bool    // Whether Docker is available

	// Container instance counts (for "Containers" lamps)
	RunningContainerCount int // Currently running container instances (lit lamps)
	TotalContainerCount   int // Total container instances started (total lamps shown)

	// System tool instance counts (for "Native" lamps)
	RunningSystemCount int // Currently running system tool invocations (lit lamps)
	TotalSystemCount   int // Total system tool invocations started (total lamps)
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

// InitSummaryMsg delivers init summary for structured display.
type InitSummaryMsg struct {
	Summary *InitSummary
}

// UoWStartMsg is sent when a module starts execution (for tab tracking).
// This creates the tab in pending state (scheduled but waiting for slot).
type UoWStartMsg struct {
	Moniker     string // Full ID for matching (Longname)
	DisplayName string // Context-aware name for tab display
	Weight      int    // Scheduling weight/pressure for this module
}

// UoWRunningMsg is sent when a module acquires its execution slot.
// This transitions the tab from pending to running state.
type UoWRunningMsg struct {
	Moniker string
}

// UoWCompleteMsg is sent when a module completes execution.
type UoWCompleteMsg struct {
	Moniker   string
	ExitCode  int
	CacheTime time.Time // For cached modules: when the artifact was last built (zero = unknown)
	LogPath   string    // Path to build log file (if available)
}

// TabSelectMsg is sent when user clicks on a tab.
type TabSelectMsg struct {
	Moniker string // Empty string = aggregate view
}

// TabDecayMsg is sent periodically to clean up decayed tabs.
type TabDecayMsg struct{}

// ConfigReadyMsg delivers configuration metadata before full init summary.
// This enables the TUI to show command context, parallelism mode, etc.
// before module resolution completes.
type ConfigReadyMsg struct {
	CommandName      string
	ExecutionContext string
	ParallelismMode  string
	EffectiveWorkers int
	WeightedCapacity int
	OutputDir        string
}

// MarqueeTickMsg is sent periodically to animate hovered tab name scrolling.
type MarqueeTickMsg struct{}
