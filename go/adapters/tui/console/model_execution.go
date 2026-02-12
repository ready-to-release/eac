package console

import (
	"fmt"
	"strings"
	"time"
)

// ExecutionState tracks execution progress, pane state, capacity, and channels.
type ExecutionState struct {
	// 3-pane state
	Panes       [3]*Pane     // Init, Run, Summary panes
	ActivePhase Phase        // Currently active phase
	SummaryData *SummaryData // Structured data for Summary pane
	InitSummary *InitSummary // Structured data for Init pane

	// Results buffer for post-execution output
	ResultsBuffer *RingBuffer // Output that appears after Run phase completes

	// Run phase state (orchestrator status)
	Running   []string // Currently running module monikers
	Completed int      // Completed count
	Total     int      // Total modules

	// Capacity tracking (three-value model) - Host scheduler
	Roof           int // Hard ceiling - actual peak allocation (workers spawned at start)
	PressureTarget int // Dynamic optimal capacity (may be < roof under memory pressure)

	// Docker scheduler capacity tracking
	DockerRunning        int // Currently running container weight
	DockerPressureTarget int // Docker scheduler pressure target
	DockerRoof           int // Docker scheduler max capacity
	DockerWaiting        int // Jobs waiting for docker scheduler

	StartTime time.Time // Execution start
	LastError *Line     // Most recent error (sticky display)

	// Lock tracking info (from locktracker.Registry)
	Locks []LockStatus // Individual lock states

	// Per-module tab tracking for Run phase
	UoWStates  map[string]*UoWState // Per-module state (running, completed, failed)
	UoWOrder   []string             // Order in which modules started (for tab ordering)
	NextUoWIdx int                  // Counter for assigning unique indices to modules

	// Note: UoW counters (uowTotal, uowDone, uowCached, uowFailed) were removed.
	// Use DeriveCounts() to compute statistics from UoWStates.

	// Channels for async updates
	LineChan   <-chan Line     // Incoming output lines
	StatusChan <-chan Status   // Status updates
	DoneChan   <-chan struct{} // Termination signal - closes to unblock listeners

	// Done state
	LinesDone   bool
	StatusDone  bool
	AllWorkDone bool // True after AfterExecute hooks complete (AllWorkDoneMsg received)
}

// SummaryData holds structured information for the Summary pane.
type SummaryData struct {
	Success     bool          // Overall success/failure
	TotalTime   time.Duration // Total execution time
	InitSummary string        // Init phase summary text
	RunSummary  string        // Run phase summary text
	Details     []string      // Detail lines (artifacts, errors, stats, etc.)
	NextSteps   string        // Suggested next action
}

// InitSummary holds structured init summary data for the Init pane.
// This is populated from initsummary.Summary and sent via InitSummaryMsg.
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

// PlannedTool represents a tool that will be used during execution.
type PlannedTool struct {
	Name        string // Tool identifier (e.g., "go", "godog", "trivy")
	IsContainer bool   // true = runs in container, false = runs on system
}

// ExecutionModule represents a module and its units of work.
type ExecutionModule struct {
	Name string     // Module name (e.g., "eac")
	UoWs []UoWEntry // Units of work with ID and display name
}

// UoWEntry represents a unit of work with matching and display info.
type UoWEntry struct {
	ID          string // Full moniker for matching (Longname: context:module:component:tool)
	DisplayName string // Short name for display (e.g., "go", "godog")
	Weight      int    // Scheduling weight for resource allocation

	// Structured identity (from core port UnitInfo)
	Module        string // Module moniker (e.g., "core", "eac")
	Component     string // Component name (e.g., "go", "gherkin")
	Tool          string // Tool/handler name (e.g., "go", "godog", "buildx")
	ComponentType string // Component type from blueprints.yml component-kinds
	Container     bool   // true = container, false = host native
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

// UoWState tracks per-module execution state for tab display.
type UoWState struct {
	Moniker     string      // Full ID for matching (Longname: context:module:component:tool)
	DisplayName string      // Context-aware name for display
	Index       int         // 1-based index in execution order (for tab display)
	Weight      int         // Scheduling weight/pressure (shown in tab)
	Buffer      *RingBuffer // Module-specific output buffer
	Status      UoWStatus   // Running, Complete, Failed
	StartTime   time.Time   // When module started
	EndTime     time.Time   // When module finished (zero if running)
	ExitCode    int         // Exit code (only valid when complete/failed)
	DecayTime   time.Time   // When tab should disappear (zero = don't decay)
	CacheTime   time.Time   // For cached modules: when the artifact was last built
	LogPath     string      // Path to build log file (if available)

	// Structured identity (from UoWEntry, populated once on creation)
	Module        string
	Component     string
	Tool          string
	ComponentType string
	Container     bool
}

// UoWStatus represents the execution state of a module.
type UoWStatus int

const (
	UoWPending  UoWStatus = iota // Scheduled, waiting for slot
	UoWRunning                   // Actively executing
	UoWComplete                  // Finished successfully
	UoWSkipped                   // Skipped (cached, unchanged)
	UoWFailed                    // Finished with error
)

// Icon returns the icon for a module status (ASCII-safe).
func (s UoWStatus) Icon() string {
	switch s {
	case UoWPending:
		return "o" // Waiting
	case UoWRunning:
		return ">" // Active
	case UoWComplete:
		return "V" // Done
	case UoWSkipped:
		return "=" // Cached/Skipped
	case UoWFailed:
		return "X" // Error
	default:
		return "?"
	}
}

// StatusColors holds the color scheme for a module status.
type StatusColors struct {
	Border string // Border/outline color
	Text   string // Text/icon color
	Bg     string // Background color
}

// Colors returns the color scheme for a module status.
// Colors are ANSI 256-color codes.
func (s UoWStatus) Colors() StatusColors {
	switch s {
	case UoWPending:
		return StatusColors{Border: "238", Text: "245", Bg: "234"} // Gray
	case UoWRunning:
		return StatusColors{Border: "214", Text: "214", Bg: "94"} // Orange/Yellow
	case UoWComplete:
		return StatusColors{Border: "40", Text: "40", Bg: "22"} // Green
	case UoWSkipped:
		return StatusColors{Border: "75", Text: "75", Bg: "23"} // Cyan/Blue (cached)
	case UoWFailed:
		return StatusColors{Border: "196", Text: "196", Bg: "52"} // Red
	default:
		return StatusColors{Border: "238", Text: "245", Bg: "234"} // Gray
	}
}

// StatusFromExitCode returns the appropriate UoWStatus for an exit code.
// exitCode == 0: Complete (success)
// exitCode < 0: Skipped (cached)
// exitCode > 0: Failed
func StatusFromExitCode(exitCode int) UoWStatus {
	if exitCode == 0 {
		return UoWComplete
	} else if exitCode < 0 {
		return UoWSkipped
	}
	return UoWFailed
}

// SetTotal sets the total number of modules.
func (m *Model) SetTotal(total int) {
	m.Execution.Total = total
}

// SetActivePhase switches to a new phase.
func (m *Model) SetActivePhase(phase Phase) {
	// Mark previous phase as complete if it was active
	if m.Execution.Panes[m.Execution.ActivePhase].Status == PhaseActive {
		m.Execution.Panes[m.Execution.ActivePhase].Status = PhaseComplete
		m.Execution.Panes[m.Execution.ActivePhase].EndTime = time.Now()
	}

	// Activate new phase
	m.Execution.ActivePhase = phase
	m.Execution.Panes[phase].Status = PhaseActive
	m.Execution.Panes[phase].StartTime = time.Now()
}

// GetActivePane returns the currently active pane.
func (m *Model) GetActivePane() *Pane {
	return m.Execution.Panes[m.Execution.ActivePhase]
}

// WriteToPhase writes a line to a specific phase's buffer.
func (m *Model) WriteToPhase(phase Phase, line Line) {
	if m.Execution.Panes[phase] != nil {
		m.Execution.Panes[phase].Buffer.Push(line)
	}
}

// SetPhaseSummary sets the summary text for a phase (shown when collapsed).
func (m *Model) SetPhaseSummary(phase Phase, summary string) {
	if m.Execution.Panes[phase] != nil {
		m.Execution.Panes[phase].Summary = summary
	}
}

// CompletePhase marks a phase as complete with success/failure status.
func (m *Model) CompletePhase(phase Phase, success bool, summary string) {
	if m.Execution.Panes[phase] != nil {
		if success {
			m.Execution.Panes[phase].Status = PhaseComplete
		} else {
			m.Execution.Panes[phase].Status = PhaseFailed
		}
		m.Execution.Panes[phase].Summary = summary
		m.Execution.Panes[phase].EndTime = time.Now()
	}
}

// WriteResult writes a line to the results buffer.
func (m *Model) WriteResult(line Line) {
	m.Execution.ResultsBuffer.Push(line)
}

// SetSummaryData updates the summary data for the Summary pane.
func (m *Model) SetSummaryData(data *SummaryData) {
	m.Execution.SummaryData = data
}

// GetOrCreateUoWState gets or creates a module state from a UoWEntry.
// For ad-hoc creation (e.g., from progress events with only a moniker), use
// GetOrCreateUoWStateByID.
func (m *Model) GetOrCreateUoWState(entry UoWEntry) *UoWState {
	if state, exists := m.Execution.UoWStates[entry.ID]; exists {
		// Update weight if provided (in case it changed)
		if entry.Weight > 0 {
			state.Weight = entry.Weight
		}
		// Update display name if provided
		if entry.DisplayName != "" {
			state.DisplayName = entry.DisplayName
		}
		// Enrich identity if not yet set
		if state.ComponentType == "" && entry.ComponentType != "" {
			state.Module = entry.Module
			state.Component = entry.Component
			state.Tool = entry.Tool
			state.ComponentType = entry.ComponentType
			state.Container = entry.Container
		}
		return state
	}

	// DEBUG: Log module registration for TUI caching investigation
	log.Debugf("[TUI-CACHE] GetOrCreateUoWState: registering module=%s displayName=%s weight=%d", entry.ID, entry.DisplayName, entry.Weight)

	// Increment counter first to get unique index
	m.Execution.NextUoWIdx++

	// Default weight to 1 if not provided
	weight := entry.Weight
	if weight <= 0 {
		weight = 1
	}

	// Create new module state with its own buffer
	// StartTime is left zero - will be set when MarkUoWRunning is called
	// Use config-driven buffer size
	uowBufferSize := m.Display.BufferSizeUoW
	if uowBufferSize <= 0 {
		uowBufferSize = 200 // Fallback
	}
	state := &UoWState{
		Moniker:       entry.ID,
		DisplayName:   entry.DisplayName,
		Index:         m.Execution.NextUoWIdx, // Unique 1-based index from counter
		Weight:        weight,
		Buffer:        NewRingBuffer(uowBufferSize),
		Status:        UoWPending, // Start as pending until slot acquired
		Module:        entry.Module,
		Component:     entry.Component,
		Tool:          entry.Tool,
		ComponentType: entry.ComponentType,
		Container:     entry.Container,
	}
	m.Execution.UoWStates[entry.ID] = state
	m.Execution.UoWOrder = append(m.Execution.UoWOrder, entry.ID)

	return state
}

// GetOrCreateUoWStateByID is a convenience wrapper for ad-hoc creation
// from progress events that only have a moniker string.
func (m *Model) GetOrCreateUoWStateByID(moniker, displayName string, weight int) *UoWState {
	return m.GetOrCreateUoWState(UoWEntry{
		ID:          moniker,
		DisplayName: displayName,
		Weight:      weight,
	})
}

// MarkUoWRunning marks a module as actively running (slot acquired).
// This sets the StartTime to now, so duration tracking reflects actual execution time.
// If the module is already in a terminal state (complete/skipped/failed), this is a no-op
// to prevent overwriting early cache detection results.
func (m *Model) MarkUoWRunning(moniker string) {
	state, exists := m.Execution.UoWStates[moniker]
	if !exists {
		return
	}
	// Don't transition to running if already in a terminal state
	// This prevents background cache detection results from being overwritten
	if state.Status == UoWComplete || state.Status == UoWSkipped || state.Status == UoWFailed {
		return
	}
	// Weight-based tracking is derived in view.go from UoWState.Weight
	// No need to track running count here - it's derived from state
	state.Status = UoWRunning
	state.StartTime = time.Now() // Start timing when execution actually begins

}

// MarkUoWComplete marks a module as completed with the given exit code.
// Exit codes: 0=success(green), <0=skipped/cached(blue), >0=failed(red)
// If the tab is currently selected, it stays visible until user switches away.
func (m *Model) MarkUoWComplete(moniker string, exitCode int) {
	state, exists := m.Execution.UoWStates[moniker]
	if !exists {
		// Module not registered - create it now so we can set the correct status
		// This handles race conditions where complete arrives before start
		log.Debugf("[TUI-CACHE] MarkUoWComplete: module=%s exitCode=%d - creating state", moniker, exitCode)
		state = m.GetOrCreateUoWStateByID(moniker, "", 1)
	}

	// Always set the status from exit code - this is the authoritative source
	// Note: counts are derived from uowStates via DeriveCounts(), not counters
	newStatus := StatusFromExitCode(exitCode)
	log.Debugf("[TUI-CACHE] MarkUoWComplete: module=%s exitCode=%d -> %v (was %v)", moniker, exitCode, newStatus, state.Status)

	state.EndTime = time.Now()
	state.ExitCode = exitCode
	state.Status = newStatus

	// Tabs stay visible - only removed via FIFO when over limit
	// Note: Counter fields (uowDone, uowCached, uowFailed) were removed.
	// Use DeriveCounts() to get current statistics from uowStates.
}

// removeUoWFromTabs removes a module from the tab display.
func (m *Model) removeUoWFromTabs(moniker string) {
	// Remove from order
	var newOrder []string
	for _, name := range m.Execution.UoWOrder {
		if name != moniker {
			newOrder = append(newOrder, name)
		}
	}
	m.Execution.UoWOrder = newOrder

	// Remove from states
	delete(m.Execution.UoWStates, moniker)
}

// GetVisibleTabs returns the tabs that should be displayed.
// Tabs are shown in execution order (uowOrder). All tabs are kept - no deletion.
// maxTabs is ignored to preserve full visibility of execution state.
func (m *Model) GetVisibleTabs() []*UoWState {
	var tabs []*UoWState

	for _, moniker := range m.Execution.UoWOrder {
		state := m.Execution.UoWStates[moniker]
		if state == nil {
			continue
		}
		tabs = append(tabs, state)
	}

	return tabs
}

// CleanupDecayedTabs is a no-op now (tabs removed instantly on completion).
func (m *Model) CleanupDecayedTabs() {
	// Tabs are now removed instantly when modules complete
	// This function is kept for compatibility but does nothing
}

// GetCapacityInfo returns the current capacity state for display.
// Uses the three-value model: Running (from UoW states), Roof (peak allocation),
// and PressureTarget (dynamic optimal based on RAM/CPU).
func (m Model) GetCapacityInfo() CapacityInfo {
	counts := m.DeriveCounts()
	return CapacityInfo{
		Running:        counts.Running,
		Roof:           m.Execution.Roof,
		PressureTarget: m.Execution.PressureTarget,
	}
}

// formatLockInfo returns a compact string describing lock status.
// Returns empty string if no locks are active.
func (m Model) formatLockInfo() string {
	parts := m.getLockParts()
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

// getLockParts returns individual formatted lock strings.
// File locks are excluded as they're redundant with the module tabs.
func (m Model) getLockParts() []string {
	if len(m.Execution.Locks) == 0 {
		return nil
	}

	var parts []string

	for _, lock := range m.Execution.Locks {
		switch lock.Type {
		case "semaphore", "weighted":
			// Show capacity usage: "name:used/cap"
			if lock.Capacity > 0 {
				part := fmt.Sprintf("%s:%d/%d", lock.Name, lock.Used, lock.Capacity)
				if lock.Waiting > 0 {
					part += fmt.Sprintf("(w%d)", lock.Waiting)
				}
				parts = append(parts, part)
			}
		case "filelock":
			// Skip file locks - they're redundant with module tabs
			continue
		default:
			// Generic display for other lock types
			if lock.Capacity > 0 {
				parts = append(parts, fmt.Sprintf("%s:%d/%d", lock.Name, lock.Used, lock.Capacity))
			} else if lock.Used > 0 {
				parts = append(parts, fmt.Sprintf("%s:active", lock.Name))
			}
		}
	}

	return parts
}
