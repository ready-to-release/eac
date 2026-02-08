// Package display defines TUI display interfaces and data types for clibase.
//
// This package decouples clibase from adapters/tui by providing interfaces
// and data types that clibase uses. The adapters/tui package implements
// these interfaces, and concrete implementations are injected at runtime.
//
// Dependency direction: adapters/tui -> clibase/display (not the reverse).
package display

import (
	"context"
	"io"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// Console is the interface for TUI console implementations.
// This mirrors the contract Console interface but is defined in clibase
// to avoid the layering violation of clibase importing adapters.
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
type ConsoleFactory func(config Config) Console

// Config configures TUI console behavior.
type Config struct {
	// Display
	Height     int  // Total height (default: DefaultHeight)
	BufferSize int  // Line buffer size (default: DefaultBufferSize)
	ASCIIMode  bool // Use ASCII-only characters
	TUI3Demo   bool // Use experimental tui3 layout

	// Behavior
	SkipTUIDelay bool // Skip exit delay

	// Command Context
	RunPhaseName string // Custom name for Run phase (e.g., "building")
	CommandName  string // Name of the command being run
	ActionType   core.ActionType // ActionBuild, ActionTest, ActionLint, ActionScan, etc.
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

// InitSummary holds structured init summary data for the Init pane.
type InitSummary struct {
	Command           string
	ExecutionContext   string
	RequestedModules  int
	CalculatedModules int
	AddedDepm         int
	UoWCount          int
	ExecutionTree     []ExecutionModule
	ParallelismMode   string
	EffectiveWorkers  int
	TurboBoost        int
	WeightedCapacity  int
	Flags             InitSummaryFlags
	DepsVerified      bool
	DepsSkipped       bool
	DepsAvailable     []string
	DepsMissing       []string
	DepmVerified      bool
	DepmSkipped       bool
	DepmResolved      int
	DepmExisting      int
	DepmTotal         int
	DepmMissing       []string
	IncrementalEnabled  bool
	IncrementalChanged  int
	IncrementalUpToDate int
	IncrementalFresh    bool
	TestSuiteName     string
	TestSelected      int
	TestDiscovered    int
	TestOSFiltered    int
	OutputDir         string
	PlannedTools      []PlannedTool
}

// ExecutionModule represents a module and its UoWs.
type ExecutionModule struct {
	Name string
	UoWs []UoWEntry
}

// UoWEntry represents a unit of work with its globally unique ID.
type UoWEntry struct {
	ID            string
	DisplayName   string
	Weight        int
	Tags          workunit.TagSummary
	Module        string
	Component     string
	Tool          string
	ComponentType string
	Container     bool
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
	Name        string
	IsContainer bool
}

// PlannedWorkItem describes predicted work from module/component discovery.
type PlannedWorkItem struct {
	ID            string
	DisplayName   string
	Weight        int
	Module        string
	Component     string
	ComponentType string
}

// UoWEnrichment carries incremental enrichment data for an existing planned tab.
type UoWEnrichment struct {
	ID          string
	Tool        string
	Container   bool
	Weight      int
	CacheStatus CacheHit
	CacheTime   time.Time
	DependsOn   []string
}

// CacheHit describes the cache status of a work unit.
type CacheHit int

const (
	CacheUnknown  CacheHit = iota
	CacheMiss
	CacheHitFresh
)

// Default TUI configuration values.
const (
	// DefaultHeight is the initial TUI height before WindowSizeMsg arrives.
	DefaultHeight = 40

	// DefaultBufferSize is the default line buffer capacity.
	DefaultBufferSize = 1000
)

// TUIBootstrap holds the factory functions for creating TUI components.
// This enables dependency injection from adapters/tui into clibase/cmdframework
// without clibase needing to import the adapter package.
type TUIBootstrap struct {
	// NewForCommand creates a Console for the given command path and config.
	NewForCommand func(commandPath string, config Config) Console

	// NewObserver creates an observer that translates execution events to TUI ops.
	// The observer must implement both interfaces.ExecutionObserver and interfaces.WriterFactory.
	NewObserver func(console Console) (observer interface{}, writerFactory interface{})

	// NewHooks creates TUI hooks for controlled interaction.
	// Returns an interfaces.TUIHooks implementation.
	NewHooks func(console Console) interface{}

	// UnwrapConsole extracts the inner Console from a wrapper (e.g., parallel.Console).
	// This is needed because some Console implementations wrap another Console.
	// Returns the input console if no unwrapping is needed.
	UnwrapConsole func(console Console) Console
}

// global TUI bootstrap - set by adapters/tui at init time.
var globalBootstrap *TUIBootstrap

// SetTUIBootstrap registers the TUI factory functions.
// Called by adapters/tui during initialization.
func SetTUIBootstrap(b *TUIBootstrap) {
	globalBootstrap = b
}

// GetTUIBootstrap returns the registered TUI bootstrap, or nil if none set.
func GetTUIBootstrap() *TUIBootstrap {
	return globalBootstrap
}
