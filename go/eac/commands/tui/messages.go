// Package tui provides a text user interface abstraction for CLI commands.
// It defines interfaces and types that allow multiple TUI implementations
// to be bound to specific commands via a registry pattern.
package tui

import "time"

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
	Phase       string
	Running     []string
	Completed   int
	Total       int
	Layer       int // Current layer being executed (1-indexed, 0 = not using layers)
	TotalLayers int // Total number of layers (0 = not using layers)

	// Detailed lock tracking info (from locktracker.Registry)
	Locks []LockStatus // Individual lock states

	// Container tools (e.g., "mkdocs-build", "go-lint")
	ActiveContainers []string
	UsedContainers   []string
	// System tools (e.g., "go", "docker")
	ActiveSystemTools []string
	UsedSystemTools   []string
}

// LockStatus represents the state of a single lock.
type LockStatus struct {
	Name     string // Lock name (e.g., "component-scheduler", "module:books")
	Type     string // Lock type: "semaphore", "weighted", "filelock"
	Capacity int    // Total capacity (for semaphores/weighted)
	Used     int    // Currently in use
	Waiting  int    // Waiting for this lock
}

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

// PhaseStatus represents the status of a phase.
type PhaseStatus int

const (
	// PhasePending means the phase has not started.
	PhasePending PhaseStatus = iota
	// PhaseActive means the phase is currently running.
	PhaseActive
	// PhaseComplete means the phase finished successfully.
	PhaseComplete
	// PhaseFailed means the phase finished with errors.
	PhaseFailed
)

// SummaryData holds structured summary information for the Summary pane.
type SummaryData struct {
	Success     bool          // Overall success/failure
	TotalTime   time.Duration // Total execution time
	InitSummary string        // Init phase summary text
	RunSummary  string        // Run phase summary text
	Details     []string      // Detail lines (artifacts, errors, stats, etc.)
	NextSteps   string        // Suggested next action
}

// InitSummary holds structured init summary data for the Init pane.
type InitSummary struct {
	// Command being executed
	Command string // "build", "test", "lint", "scan"

	// Execution context
	ExecutionContext string // local/CI/container

	// Module counts
	RequestedModules  int // What user asked for
	CalculatedModules int // Final list after dependency resolution
	AddedDepm         int // Module dependencies added

	// UoW count - total units of work to schedule
	UoWCount int

	// Execution tree - for visual tree rendering
	ExecutionTree         []ExecutionLayer // Full tree: layers -> modules -> components
	LayerCount            int              // Number of execution layers
	LayerSizes            []int            // Number of modules per layer
	ComponentsPerModLayer []int            // Number of components per layer
	FlatExecution         bool             // True if running all layers in parallel

	// Parallelism
	ParallelismMode  string // "ci" or "devbox"
	EffectiveWorkers int    // Final worker count
	TurboBoost       int    // Additional workers (0 if not turbo)
	WeightedCapacity int    // Component scheduler capacity

	// Flags
	Flags InitSummaryFlags

	// Deps status
	DepsVerified  bool
	DepsSkipped   bool
	DepsAvailable []string // Available system deps
	DepsMissing   []string // Missing system deps

	// Depm status
	DepmVerified bool
	DepmSkipped  bool
	DepmResolved int      // Number of resolved module deps (to be built)
	DepmExisting int      // Number of existing module deps (already built)
	DepmTotal    int      // Total module deps
	DepmMissing  []string // Missing module deps

	// Incremental (build only)
	IncrementalEnabled  bool
	IncrementalChanged  int
	IncrementalUpToDate int
	IncrementalFresh    bool

	// Test-specific
	TestSuiteName  string
	TestSelected   int
	TestDiscovered int
	TestOSFiltered int

	// Output directory
	OutputDir string

	// Tools that will be used (for pre-populating tool lamps)
	PlannedTools []PlannedTool
}

// ExecutionLayer represents a single execution layer with its modules.
type ExecutionLayer struct {
	Modules []ExecutionModule
}

// ExecutionModule represents a module and its components within a layer.
type ExecutionModule struct {
	Name       string   // Module name (e.g., "eac-commands")
	Components []string // Component names (e.g., ["godog", "impl/build"])
}

// InitSummaryFlags captures relevant flags for display.
type InitSummaryFlags struct {
	TidyFirst    bool
	ForceRebuild bool
	DryRun       bool
	UseTUI       bool
	SkipDeps     bool
	SkipDepm     bool
}

// PlannedTool represents a tool that will be used during execution.
type PlannedTool struct {
	Name        string // Tool identifier (e.g., "go", "godog", "trivy")
	IsContainer bool   // true = runs in container, false = runs on system
}

// SubcommandInfo describes a subcommand for command selection.
// Used by parent commands to define their available subcommands.
// Convert to CommandOption using SubcommandToOption for use with the selector.
type SubcommandInfo struct {
	Name        string
	Description string
	Aliases     []string
}
