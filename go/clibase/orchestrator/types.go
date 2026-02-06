package orchestrator

import (
	"context"
	"io"
	"sort"
	"time"
)

// SummaryBuilder defines the interface for incremental summary building.
// This interface allows the scheduler to send results to the summary builder
// without creating an import cycle with cmdframework.
type SummaryBuilder interface {
	// AddResult adds a unit result to the builder (thread-safe).
	AddResult(result UnitResult)
}

// WorkItem represents a single unit of work to be processed.
type WorkItem struct {
	// Moniker is the unique identifier for this work item
	Moniker string
	// Index is the original position in the input slice (for ordering results)
	Index int
}

// WorkResult represents the outcome of processing a work item.
type WorkResult struct {
	// Moniker is the unique identifier for this work item
	Moniker string
	// Index is the original position in the input slice
	Index int
	// ExitCode is the exit code from the worker function (0 = success)
	ExitCode int
	// LogPath is the relative path to the detailed log file
	LogPath string
	// Warnings are any warning messages collected during processing
	Warnings []string
	// Errors are any error messages collected during processing
	Errors []string
	// Duration is the time taken to process this work item
	Duration time.Duration
	// Type is the module/work type (e.g., "go", "container", "typescript", "static")
	Type string
}

// WorkerFunc is a function that processes a single work item.
// It receives a context (for cancellation/timeout), the moniker, and a log writer.
// All output should go to the provided logWriter (not stdout/stderr).
// Returns an exit code (0 for success).
type WorkerFunc func(ctx context.Context, moniker string, logWriter io.Writer) int

// Config holds orchestrator configuration options.
type Config struct {
	// WorkspaceRoot is the root directory of the repository
	WorkspaceRoot string
	// OutputBaseDir is the base directory for all output (e.g., paths.OutBuildRelPath or paths.OutTestRelPath)
	OutputBaseDir string
	// LogFileName is the name of the log file for each module (e.g., "build.log" or "test.log")
	LogFileName string
	// ActionVerb is the present continuous verb for status messages (e.g., "building", "testing")
	ActionVerb string
	// MaxConcurrency is the maximum number of concurrent workers (0 = number of CPUs)
	// For component-level parallelism, this is used as the weight capacity
	MaxConcurrency int
	// StatusUpdateInterval is how often to show status updates (default: 500ms)
	StatusUpdateInterval int // in milliseconds
	// ComponentTypesDisplay maps moniker to type string (e.g., "go", "container") for display
	ComponentTypesDisplay map[string]string
	// ShowTimings enables the timing summary section (use --timings flag)
	ShowTimings bool
	// DryRun skips actual execution and preserves existing output artifacts
	DryRun bool
	// TUI enables the TUI console for real-time output display
	TUI bool
	// TUIHeight is the height of the TUI console window (default: tui.DefaultHeight)
	TUIHeight int
	// TUIASCIIMode uses ASCII-only characters for TUI (default: false)
	// Set to true via --ascii flag for terminals that don't render Unicode well
	TUIASCIIMode bool
	// TUI3Demo enables the experimental tui3 layout (default: false)
	// Set to true via --demo flag to preview the new TUI design
	TUI3Demo bool
	// SkipTUIDelay skips TUI exit delay (exit immediately when done)
	SkipTUIDelay bool
	// Turbo is the parallelism multiplier (1.0x, 1.25x, 2.0x, etc.)
	// Higher values reduce memory per slot, increasing concurrent builds.
	// Default is 1.25x when turbo is enabled, 1.0x when disabled.
	Turbo float64
}

// ConfigUpdate holds configuration fields that can be updated after creation.
// Only non-zero/non-nil fields are applied. This enables creating an
// orchestrator with minimal config and updating it once full config is loaded.
type ConfigUpdate struct {
	WorkspaceRoot         string
	OutputBaseDir         string
	LogFileName           string
	MaxConcurrency        int
	Turbo                 float64
	ComponentTypesDisplay map[string]string
	ShowTimings           bool
	DryRun                bool
}

// UnitResult represents the outcome of executing a single work unit.
type UnitResult struct {
	// Module is the module moniker
	Module string
	// Component is the component name
	Component string
	// Handler is the build/test handler type (e.g., "go", "gotest", "godog")
	Handler string
	// ExitCode is the exit code from the build (0 = success)
	ExitCode int
	// Duration is the time taken to build this component
	Duration time.Duration
	// Errors are any error messages collected during the build
	Errors []string
	// Warnings are any warning messages collected during the build
	Warnings []string
	// LogPath is the relative path to the component's log file
	LogPath string

	// Test-specific fields (0 for non-test commands)
	TestsTotal   int
	TestsPassed  int
	TestsFailed  int
	TestsSkipped int
}

// UnitWorkerFunc processes a single work unit.
// It receives a context (for cancellation/timeout), the module, component name, and log writer.
// Returns an exit code (0 for success).
type UnitWorkerFunc func(ctx context.Context, module, component string, logWriter io.Writer) int

// ModuleStatus represents the execution status of a module.
type ModuleStatus int

const (
	// ModuleStatusPending indicates the module has not started processing.
	ModuleStatusPending ModuleStatus = iota
	// ModuleStatusRunning indicates the module is currently being processed.
	ModuleStatusRunning
	// ModuleStatusSuccess indicates the module completed successfully.
	ModuleStatusSuccess
	// ModuleStatusSkipped indicates the module was skipped (cached/unchanged).
	ModuleStatusSkipped
	// ModuleStatusFailed indicates the module failed.
	ModuleStatusFailed
)

// String returns the string representation of a ModuleStatus.
func (s ModuleStatus) String() string {
	switch s {
	case ModuleStatusPending:
		return "pending"
	case ModuleStatusRunning:
		return "running"
	case ModuleStatusSuccess:
		return "success"
	case ModuleStatusSkipped:
		return "cached"
	case ModuleStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// ModuleResultSet aggregates unit results for a single module.
type ModuleResultSet struct {
	// Module is the module moniker
	Module string
	// Units contains the results for each work unit in the module
	Units []UnitResult
	// Status is the derived module status based on component results
	Status ModuleStatus
	// Duration is the total time taken to process all components
	Duration time.Duration
}

// DeriveStatus computes the module status from unit results.
// - If any unit has ExitCode > 0 -> ModuleStatusFailed
// - If any unit has ExitCode < -1 -> ModuleStatusRunning (pending/in-progress)
// - If all units have ExitCode == -1 -> ModuleStatusSkipped (cached)
// - If all units have ExitCode <= 0 (mix of 0 and -1) -> ModuleStatusSuccess
// - If no units -> ModuleStatusPending.
func (rs *ModuleResultSet) DeriveStatus() ModuleStatus {
	if len(rs.Units) == 0 {
		return ModuleStatusPending
	}

	hasRunning := false
	allSkipped := true
	for _, c := range rs.Units {
		if c.ExitCode > 0 {
			return ModuleStatusFailed
		}
		if c.ExitCode < -1 {
			hasRunning = true
		}
		if c.ExitCode != -1 {
			allSkipped = false
		}
	}

	if hasRunning {
		return ModuleStatusRunning
	}

	if allSkipped {
		return ModuleStatusSkipped
	}

	return ModuleStatusSuccess
}

// GetSortedUnits returns a copy of the units sorted alphabetically by Component name.
// The original slice is not modified.
func (rs *ModuleResultSet) GetSortedUnits() []UnitResult {
	if len(rs.Units) == 0 {
		return []UnitResult{}
	}

	// Create a copy to avoid modifying the original
	sorted := make([]UnitResult, len(rs.Units))
	copy(sorted, rs.Units)

	// Sort alphabetically by Component name
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Component < sorted[j].Component
	})

	return sorted
}
