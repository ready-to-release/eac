// Package tui defines the interface contract for terminal user interfaces.
package tui

import (
	"context"
	"io"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Console is the primary interface for all TUI implementations.
// Both parallel task TUIs (build, test, lint, scan) and simple TUIs
// must implement this interface.
type Console interface {
	// Lifecycle
	Start(ctx context.Context) error
	StartAsync(ctx context.Context)
	Wait()
	Stop()

	// Output
	NewWriter(source string, logWriter io.Writer) io.Writer
	SendLine(line Line)
	WriteResult(text string)

	// Status Updates
	UpdateStatus(status Status)

	// Configuration
	// StatusRefreshInterval returns how often the orchestrator should send status updates.
	// The orchestrator should set up a ticker and call UpdateStatus at this interval
	// to keep the TUI responsive even during idle periods.
	// Returns 0 if no periodic refresh is needed.
	StatusRefreshInterval() time.Duration

	// Phase Management
	SetPhase(phase Phase)
	CompletePhase(phase Phase, success bool, summary string)
	WriteToPhase(phase Phase, text string)
	SetPhaseSummary(phase Phase, summary string)

	// UoW/Task Tracking (for parallel TUIs)
	StartUoW(moniker, displayName string, weight int)
	MarkUoWRunning(moniker string)
	MarkUoWComplete(moniker string, exitCode int)
	MarkUoWCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string)

	// Early Configuration (progressive display)
	SendConfigReady(commandName, executionContext, parallelismMode string,
		effectiveWorkers, weightedCapacity int, outputDir string)

	// Summary
	SendSummary(data *SummaryData)
	SetInitSummary(summary *InitSummary)

	// Fast Init + Streaming Updates + Hold-Until-Done
	SendPlannedWork(items []PlannedWorkItem)
	EnrichUoW(enrichment UoWEnrichment)
	SignalAllWorkDone()
}

// ConsoleFactory creates Console instances with the given configuration.
// This allows TUI implementations to be instantiated lazily with runtime config.
type ConsoleFactory func(config Config) Console

// Registry holds command-to-TUI bindings.
type Registry interface {
	Register(pattern string, factory ConsoleFactory)
	NewForCommand(commandPath string, config Config) Console
	MustHaveDefault()
	ListBindings() []string
}

// Config configures TUI console behavior.
// All fields have sensible defaults; zero values are valid.
type Config struct {
	// Display
	Height     int  // Total height (default: DefaultHeight)
	BufferSize int  // Line buffer size (default: DefaultBufferSize)
	ASCIIMode  bool // Use ASCII-only characters
	TUI3Demo   bool // Use experimental tui3 layout

	// Behavior
	SkipTUIDelay bool // Skip exit delay

	// Command Context
	RunPhaseName string          // Custom name for Run phase (e.g., "building")
	CommandName  string          // Name of the command being run
	ActionType   core.ActionType // ActionBuild, ActionTest, ActionLint, ActionScan, etc.

	// Advanced Configuration (optional)
	TUIConfig *TUIConfig // Timeout and layout configuration (nil = use defaults)
}

// TUIConfig holds all configurable TUI parameters.
// This is passed to console.NewModel to avoid global state and enable testing.
type TUIConfig struct {
	// Timeouts
	MetricsInterval       time.Duration // How often to refresh CPU/memory metrics (default: 500ms)
	StatusRefreshInterval time.Duration // How often orchestrator should send status updates (default: 250ms)
	MinDisplayTime        time.Duration // Minimum time to show completion state (default: 1500ms)
	ExitCountdown         time.Duration // Exit countdown duration (default: 10s)
	FreezeCountdown       time.Duration // Extended countdown when Freeze clicked (default: 120s)
	AutoScrollResume      time.Duration // Auto-scroll resume delay (default: 8s)

	// Layout
	MaxTabs           int // Maximum visible tabs before scrolling (default: 36)
	DefaultColumns    int // Default number of tab columns (default: 4)
	MinColumns        int // Minimum tab columns (default: 2)
	MaxColumns        int // Maximum tab columns (default: 6)
	BufferSizePane    int // Buffer size for each pane (default: 500)
	BufferSizeResults int // Buffer size for results (default: 100)
	BufferSizeUoW     int // Buffer size per UoW (default: 200)
}

// DefaultTUIConfig returns the default TUI configuration.
// These values match the current hardcoded values in the console.
func DefaultTUIConfig() *TUIConfig {
	return &TUIConfig{
		// Timeouts
		MetricsInterval:       500 * time.Millisecond,
		StatusRefreshInterval: 250 * time.Millisecond,
		MinDisplayTime:        1500 * time.Millisecond,
		ExitCountdown:         10 * time.Second,
		FreezeCountdown:       120 * time.Second,
		AutoScrollResume:      8 * time.Second,

		// Layout
		MaxTabs:           36,
		DefaultColumns:    4,
		MinColumns:        2,
		MaxColumns:        6,
		BufferSizePane:    500,
		BufferSizeResults: 100,
		BufferSizeUoW:     200,
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

	// Capacity tracking for the three-value model:
	// - Running: sum of weights currently executing (derived from UoW states)
	// - Roof: hard ceiling - actual peak allocation (workers spawned at start)
	// - PressureTarget: dynamic optimal capacity based on RAM/CPU
	//
	// Display: "Cap: ●●●●●●●●●●●●●●●●●●●●●●●● 24/24 [▼16]"
	//          where 24/24 = running/roof, [▼16] = pressure target
	Roof           int // Hard ceiling - peak allocation (set at scheduler start)
	PressureTarget int // Dynamic optimal capacity (may be < Roof under memory pressure)

	// Detailed lock tracking info (from locktracker.Registry)
	Locks []LockStatus // Individual lock states

	// Container tools (e.g., "mkdocs-build", "go-lint")
	ActiveContainerTools []string
	UsedContainerTools   []string
	// System tools (e.g., "go", "docker")
	ActiveSystemTools []string
	UsedSystemTools   []string

	// Docker memory metrics (from dual-pool scheduler)
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

	// Execution tree - modules and their components
	ExecutionTree []ExecutionModule

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

// UoWEntry represents a unit of work with its globally unique ID.
type UoWEntry struct {
	ID          string          // UnitIDPort.Longname() - the canonical key (context:module:component:tool[:extra])
	DisplayName string          // UnitIDPort.Shortname() - for tabs (module:component or spec name)
	Weight      int             // Scheduling weight for resource allocation
	Tags        core.TagSummary // Classified tag data (test UoWs only)

	// Structured identity (decomposed from UnitSpec/UnitID)
	Module        string // Module moniker (e.g., "core", "eac")
	Component     string // Component name (e.g., "go", "gherkin")
	Tool          string // Tool/handler name (e.g., "go", "godog", "buildx")
	ComponentType string // From component-types.yml (e.g., "go", "gherkin", "dockerfile")
	Container     bool   // true = runs in container, false = runs on host
}

// ExecutionModule represents a module and its UoWs.
type ExecutionModule struct {
	Name string     // Module name (e.g., "eac-commands")
	UoWs []UoWEntry // Units of work with their globally unique IDs
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

// SubcommandInfo describes a subcommand for display/conversion purposes.
type SubcommandInfo struct {
	Name        string
	Description string
	Aliases     []string
}

// Default TUI configuration values.
const (
	// DefaultHeight is the initial TUI height before WindowSizeMsg arrives.
	// In alt-screen mode, the actual terminal height is used instead.
	DefaultHeight = 40

	// DefaultBufferSize is the default line buffer capacity.
	DefaultBufferSize = 1000
)

// WithDefaults returns a copy of Config with default values applied.
func (c Config) WithDefaults() Config {
	if c.Height <= 0 {
		c.Height = DefaultHeight
	}
	if c.BufferSize <= 0 {
		c.BufferSize = DefaultBufferSize
	}
	return c
}

// PlannedWorkItem describes predicted work from module/component discovery.
// Used by SendPlannedWork to create grey skeleton tabs before tool resolution.
type PlannedWorkItem struct {
	ID            string // Predicted Longname: context:module:component:toolID
	DisplayName   string // module:component (matches UnitID.DisplayName() format)
	Weight        int    // Approximate weight from component type config
	Module        string // Module moniker
	Component     string // Component name
	ComponentType string // From component-types.yml
}

// UoWEnrichment carries incremental enrichment data for an existing planned tab.
type UoWEnrichment struct {
	ID          string    // Must match existing tab ID (Longname)
	Tool        string    // Resolved tool handler name
	Container   bool      // Runs in container (only applied when Tool is set)
	Weight      int       // Actual weight from tool resolution (0 = keep existing)
	CacheStatus CacheHit  // Cache status from incremental detection
	CacheTime   time.Time // When cached artifact was built
	DependsOn   []string  // Dependency edges (UoW IDs)
}

// CacheHit describes the cache status of a work unit.
type CacheHit int

const (
	CacheUnknown  CacheHit = iota // Not yet determined
	CacheMiss                     // Must execute
	CacheHitFresh                 // Cached and up-to-date
)

// Binding represents a resolved command-to-TUI binding.
type Binding struct {
	Pattern   string
	IsDefault bool // true for "*" binding
}
