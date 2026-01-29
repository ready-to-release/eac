package orchestrator

import (
	"io"
	"sort"
	"time"
)

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

// WorkerFunc is a function that processes a single work item
// It receives the moniker and should return an exit code (0 for success)
// All output should go to the provided logWriter (not stdout/stderr).
type WorkerFunc func(moniker string, logWriter io.Writer) int

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
	// ModuleTypes maps moniker to type string (e.g., "go", "container") for display
	ModuleTypes map[string]string
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
	// Turbo is the parallelism multiplier (1.0x, 1.25x, 2.0x, etc.)
	// Higher values reduce memory per slot, increasing concurrent builds.
	// Default is 1.25x when turbo is enabled, 1.0x when disabled.
	Turbo float64
}

// ComponentWork represents a single component build work item.
// This is the unit of work for component-level parallelism.
type ComponentWork struct {
	// Module is the module moniker (e.g., "eac-core")
	Module string
	// Component is the component name (e.g., "go", "typescript", "book")
	Component string
	// ComponentType is the component type from component-types.yml (may differ from Component for named components)
	ComponentType string
	// Handler is the name of the build handler (e.g., "go", "npm", "mkdocs")
	Handler string
	// IsContainer indicates if the handler runs in a Docker container (vs system tool)
	IsContainer bool
	// Weight is the resource weight for scheduling (base weight × amp, 1=light, 4=heavy)
	Weight int
	// BuildAfter lists component types that must complete before this one (same module)
	BuildAfter []string
	// Index is used for result ordering
	Index int
	// Cached indicates this component is pre-determined to be cached (skip execution overhead)
	Cached bool
}

// ComponentResult represents the outcome of building a single component.
type ComponentResult struct {
	// Module is the module moniker
	Module string
	// Component is the component name
	Component string
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

// ComponentWorkerFunc processes a single component build.
// It receives the module, component name, and log writer.
// Returns an exit code (0 for success).
type ComponentWorkerFunc func(module, component string, logWriter io.Writer) int

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

// ComponentResultSet aggregates component results for a single module.
type ComponentResultSet struct {
	// Module is the module moniker
	Module string
	// Components contains the results for each component in the module
	Components []ComponentResult
	// Status is the derived module status based on component results
	Status ModuleStatus
	// Duration is the total time taken to process all components
	Duration time.Duration
}

// DeriveStatus computes the module status from component results.
// - If any component has ExitCode > 0 -> ModuleStatusFailed
// - If any component has ExitCode < -1 -> ModuleStatusRunning (pending/in-progress)
// - If all components have ExitCode == -1 -> ModuleStatusSkipped (cached)
// - If all components have ExitCode <= 0 (mix of 0 and -1) -> ModuleStatusSuccess
// - If no components -> ModuleStatusPending.
func (rs *ComponentResultSet) DeriveStatus() ModuleStatus {
	if len(rs.Components) == 0 {
		return ModuleStatusPending
	}

	hasRunning := false
	allSkipped := true
	for _, c := range rs.Components {
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

// GetSortedComponents returns a copy of the components sorted alphabetically by Component name.
// The original slice is not modified.
func (rs *ComponentResultSet) GetSortedComponents() []ComponentResult {
	if len(rs.Components) == 0 {
		return []ComponentResult{}
	}

	// Create a copy to avoid modifying the original
	sorted := make([]ComponentResult, len(rs.Components))
	copy(sorted, rs.Components)

	// Sort alphabetically by Component name
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Component < sorted[j].Component
	})

	return sorted
}
