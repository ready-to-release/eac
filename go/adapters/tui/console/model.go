package console

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/ready-to-release/eac/go/core/logging"
)

var log = logging.C()

// HelpTextMap maps zone IDs to help descriptions for the Resources pane elements.
var HelpTextMap = map[string]string{
	"res-timer":    "Elapsed time since execution started",
	"res-cpu":      "CPU usage per core (green=low, yellow=medium, red=high)",
	"res-mem":      "Memory usage (green=low, yellow=medium, red=high)",
	"res-jobs":     "Active containers running in parallel",
	"res-uow":      "Unit of Work counts: running/done/cached/failed",
	"res-tools":    "Tool lamps: blue=container, orange=system (filled=active)",
	"res-layer":    "Current execution layer / total layers",
	"freeze-button": "Click to pause auto-exit countdown for 2 minutes",
}

// Model is the Bubbletea model for the console window.
// Displays build/test output in a 3-pane view (Init/Run/Summary).
type Model struct {
	// Configuration
	height       int    // Total height (default: tui.DefaultHeight)
	width        int    // Terminal width
	runPhaseName string // Custom name for Run phase (e.g., "building", "testing")
	asciiMode    bool   // Use ASCII-only characters (--ascii flag)

	// 3-pane state
	panes       [3]*Pane     // Init, Run, Summary panes
	activePhase Phase        // Currently active phase
	summaryData *SummaryData // Structured data for Summary pane
	initSummary *InitSummary // Structured data for Init pane

	// Results buffer for post-execution output
	resultsBuffer *RingBuffer // Output that appears after Run phase completes

	// Run phase state (orchestrator status)
	running     []string  // Currently running module monikers
	completed   int       // Completed count
	total       int       // Total modules
	layer       int       // Current layer being executed (1-indexed, 0 = not using layers)
	totalLayers int       // Total number of layers (0 = not using layers)
	startTime   time.Time // Execution start
	lastError   *Line     // Most recent error (sticky display)

	// Lock tracking info (from locktracker.Registry)
	locks []LockStatus // Individual lock states

	// Tools tracking (containers vs system)
	plannedContainers  []string // All container tools (from PlannedTools at init)
	plannedSystemTools []string // All system tools (from PlannedTools at init)
	activeContainers   []string // Currently active container tools
	activeSystemTools  []string // Currently active system tools

	// Per-module tab tracking for Run phase
	uowStates    map[string]*UoWState // Per-module state (running, completed, failed)
	uowOrder     []string                // Order in which modules started (for tab ordering)
	nextUoWIdx   int                     // Counter for assigning unique indices to modules
	activeTab       string                  // Currently selected tab ("" = aggregate view)
	hoveredTab      string                  // Tab currently under mouse cursor
	hoveredTabScroll int                    // Marquee scroll offset for hovered tab name
	hoveredZone     string                  // Currently hovered zone ID (empty = none, "tab:moniker" for tabs, "res-*" for resources)
	maxTabs         int                     // Maximum visible tabs before scrolling/hiding

	// Cumulative UoW counters (only increment, never recalculated)
	uowTotal   int // Total UoW expected (set once)
	uowRunning int // Currently running (can go up/down)
	uowDone    int // Completed successfully (only goes up)
	uowCached  int // Skipped/cached (only goes up)
	uowFailed  int // Failed (only goes up)

	// Channels for async updates
	lineChan   <-chan Line   // Incoming output lines
	statusChan <-chan Status // Status updates

	// Display preferences
	paused    bool     // Pause scrolling (for review)
	errorMode bool     // Show only errors
	mouseMode bool     // Mouse mode: true=scrolling enabled, false=text selection enabled
	viewMode         ViewMode // Components view mode: TabGrid or Tree
	tabsScrollOffset int      // Scroll offset for tabs/tree panel
	tabColumns       int      // Number of columns for tab grid (2-6, default 4)

	// Done state
	linesDone  bool
	statusDone bool

	// User interaction tracking (for delayed exit)
	lastUserInteraction time.Time // Reset on scroll or tab click
	userHasInteracted   bool      // True if user ever interacted with mouse
	forceExit           bool      // User pressed ESC to force immediate exit
	exitCountdownSecs   int       // Seconds remaining until user timer expires (10, 9, 8...)
	freezeTimeoutSecs   int       // Timeout duration in seconds (10 default, 120 when Freeze clicked)
	skipTUIDelay            bool      // Skip all user interaction tracking, exit immediately when done
	allRunnersCompleted time.Time // When all runners finished

	// Exit/finalization state
	exitRequested      bool         // True when we want to exit (waiting for summary if needed)
	pendingSummaryData *SummaryData // Summary data received from builder

	// Quitting state - triggers plain-text final render
	quitting bool

	// Cached system metrics (updated periodically, not on every render)
	cachedCPUPercent   []float64 // Per-core CPU usage percentages
	cachedMemPercent   float64   // Memory usage percentage
	lastMetricsUpdate  time.Time // When metrics were last updated

	// Text selection state (mouse drag to copy)
	selection SelectionState
}

// SelectionState tracks mouse text selection in the logs pane.
type SelectionState struct {
	Active    bool // Currently selecting (mouse down)
	StartX    int  // Mouse press X (screen coords)
	StartY    int  // Mouse press Y (screen coords)
	EndX      int  // Current/end X (screen coords)
	EndY      int  // Current/end Y (screen coords)
	StartLine int  // Buffer line index at selection start
	EndLine   int  // Buffer line index at selection end
}

// UoWState tracks per-module execution state for tab display.
type UoWState struct {
	Moniker     string       // Full ID for matching (Longname: context:module:component:tool)
	DisplayName string       // Short name for display (Shortname: module:component)
	Index       int          // 1-based index in execution order (for tab display)
	Weight      int          // Scheduling weight/pressure (shown in tab)
	Buffer      *RingBuffer  // Module-specific output buffer
	Status      UoWStatus // Running, Complete, Failed
	StartTime   time.Time    // When module started
	EndTime     time.Time    // When module finished (zero if running)
	ExitCode    int          // Exit code (only valid when complete/failed)
	DecayTime   time.Time    // When tab should disappear (zero = don't decay)
	CacheTime   time.Time    // For cached modules: when the artifact was last built
	LogPath     string       // Path to build log file (if available)
}

// UoWStatus represents the execution state of a module.
type UoWStatus int

const (
	UoWPending  UoWStatus = iota // Scheduled, waiting for slot
	UoWRunning                      // Actively executing
	UoWComplete                     // Finished successfully
	UoWSkipped                      // Skipped (cached, unchanged)
	UoWFailed                       // Finished with error
)

// ViewMode represents the display mode for the components pane.
type ViewMode int

const (
	ViewModeTabGrid ViewMode = iota // Tab grid view (default)
	ViewModeTree                    // Execution tree view
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

// NewModel creates a new console model.
func NewModel(height int, runPhaseName string, lineChan <-chan Line, statusChan <-chan Status, asciiMode bool, skipTUIDelay bool) Model {
	// Initialize zone manager for mouse click tracking
	zone.NewGlobal()

	if height <= 0 {
		height = 5
	}

	// Use default if no custom run phase name provided
	if runPhaseName == "" {
		runPhaseName = "Run"
	}

	// Create panes with appropriate buffer sizes
	bufferSize := 500 // Per-pane buffer
	panes := [3]*Pane{
		NewPane(PhaseInit, bufferSize),
		NewPane(PhaseRun, bufferSize),
		NewPane(PhaseSummary, bufferSize),
	}

	return Model{
		height:            height,
		width:             80, // Default, will be updated on WindowSizeMsg
		runPhaseName:      runPhaseName,
		asciiMode:         asciiMode,
		skipTUIDelay:      skipTUIDelay,
		freezeTimeoutSecs: 10, // Default user timer (10s), set to 120 when Freeze clicked
		resultsBuffer:     NewRingBuffer(100), // Results buffer
		panes:             panes,
		activePhase:       PhaseInit, // Start with Init phase
		lineChan:          lineChan,
		statusChan:        statusChan,
		startTime:         time.Now(),
		mouseMode:         true,                          // Start with mouse ON (scrolling enabled)
		uowStates:      make(map[string]*UoWState), // Per-module state tracking
		uowOrder:       make([]string, 0),             // Tab ordering
		activeTab:         "",                            // Start with aggregate view
		maxTabs:           36,                            // Maximum visible tabs (6 rows × 6 tabs/row)
		tabColumns:        4,                             // Default tab columns (adjustable 2-6)
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.listenForLines(),
		m.listenForStatus(),
		m.tickCmd(),
	)
}

// listenForLines creates a command that waits for new output lines.
func (m Model) listenForLines() tea.Cmd {
	return func() tea.Msg {
		if m.lineChan == nil {
			return linesDoneMsg{}
		}
		line, ok := <-m.lineChan
		if !ok {
			return linesDoneMsg{}
		}
		return lineMsg(line)
	}
}

// listenForStatus creates a command that waits for status updates.
func (m Model) listenForStatus() tea.Cmd {
	return func() tea.Msg {
		if m.statusChan == nil {
			return statusDoneMsg{}
		}
		status, ok := <-m.statusChan
		if !ok {
			return statusDoneMsg{}
		}
		return statusMsg(status)
	}
}

// tickCmd returns a tick every 100ms for elapsed time updates.
func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// SetTotal sets the total number of modules.
func (m *Model) SetTotal(total int) {
	m.total = total
}

// SetActivePhase switches to a new phase.
func (m *Model) SetActivePhase(phase Phase) {
	// Mark previous phase as complete if it was active
	if m.panes[m.activePhase].Status == PhaseActive {
		m.panes[m.activePhase].Status = PhaseComplete
		m.panes[m.activePhase].EndTime = time.Now()
	}

	// Activate new phase
	m.activePhase = phase
	m.panes[phase].Status = PhaseActive
	m.panes[phase].StartTime = time.Now()
}

// GetActivePane returns the currently active pane.
func (m *Model) GetActivePane() *Pane {
	return m.panes[m.activePhase]
}

// WriteToPhase writes a line to a specific phase's buffer.
func (m *Model) WriteToPhase(phase Phase, line Line) {
	if m.panes[phase] != nil {
		m.panes[phase].Buffer.Push(line)
	}
}

// SetPhaseSummary sets the summary text for a phase (shown when collapsed).
func (m *Model) SetPhaseSummary(phase Phase, summary string) {
	if m.panes[phase] != nil {
		m.panes[phase].Summary = summary
	}
}

// CompletePhase marks a phase as complete with success/failure status.
func (m *Model) CompletePhase(phase Phase, success bool, summary string) {
	if m.panes[phase] != nil {
		if success {
			m.panes[phase].Status = PhaseComplete
		} else {
			m.panes[phase].Status = PhaseFailed
		}
		m.panes[phase].Summary = summary
		m.panes[phase].EndTime = time.Now()
	}
}

// calculatePaneHeights determines how many lines each pane gets
// Init and Summary are fixed, Run fills ALL remaining terminal space.
func (m Model) calculatePaneHeights() (initH, runH, summaryH int) {
	// Fixed content heights for Init and Summary panes
	const initHeight = 6
	const summaryHeight = 4

	// Chrome lines (non-content):
	// - Init: header (1) + footer (1) = 2
	// - Resources pane: header (1) + content (1) + footer (1) = 3
	// - Run: header (1) + footer (1) = 2
	// Total base chrome = 7 lines
	chromeLines := 7

	// Add Summary chrome only if it will be rendered
	if m.summaryData != nil {
		chromeLines += 2 // Summary header (1) + footer (1)
	}

	// Account for structured Init pane (8 lines total when initSummary is available)
	// instead of buffer-based init (2 chrome + 6 content)
	structuredInitHeight := 0
	if m.initSummary != nil {
		structuredInitHeight = 8 // header + 4 rows + 2 separators + footer
	}

	// Account for Execution Tree pane (variable height based on layers)
	execTreeHeight := 0
	if m.initSummary != nil && len(m.initSummary.ExecutionTree) > 0 {
		// header (1) + per layer: header (1) + modules + spacing (1 between layers) + footer (1)
		execTreeHeight = 2 // header + footer
		for i, layer := range m.initSummary.ExecutionTree {
			execTreeHeight += 1 + len(layer.Modules) // layer header + modules
			if i < len(m.initSummary.ExecutionTree)-1 {
				execTreeHeight++ // spacing between layers
			}
		}
	}

	// Use actual terminal height (m.height is updated by WindowSizeMsg in alt-screen mode)
	terminalHeight := m.height
	if terminalHeight < 20 {
		terminalHeight = 20 // Minimum usable height
	}

	// Allocate heights
	if m.initSummary != nil {
		// Using structured init - don't count initH in content, it's fully rendered
		initH = 0 // Not used for structured rendering
	} else {
		initH = initHeight
	}

	if m.summaryData != nil {
		summaryH = summaryHeight
	} else {
		summaryH = 0 // Don't reserve space for Summary if not showing
	}

	// Calculate Run pane height
	if m.initSummary != nil {
		// Structured mode: subtract fixed pane heights
		runH = terminalHeight - chromeLines - structuredInitHeight - execTreeHeight - summaryH
	} else {
		// Buffer mode: use original calculation
		runH = terminalHeight - chromeLines - initH - summaryH
	}

	// Minimum Run pane content height
	if runH < 5 {
		runH = 5
	}

	return initH, runH, summaryH
}

// WriteResult writes a line to the results buffer.
func (m *Model) WriteResult(line Line) {
	m.resultsBuffer.Push(line)
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

// SetSummaryData updates the summary data for the Summary pane.
func (m *Model) SetSummaryData(data *SummaryData) {
	m.summaryData = data
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

	// Execution tree - for visual tree rendering
	ExecutionTree         []ExecutionLayer // Full tree: module layers → modules → components
	LayerCount            int              // Number of module execution layers
	LayerSizes            []int            // Number of modules per layer
	ComponentsPerModLayer []int            // Number of components per module layer
	FlatExecution         bool             // True if running all layers in parallel

	// Component layers - components grouped by internal dependency order
	ComponentLayers     [][]string // Component IDs grouped by dependency layer
	ComponentLayerCount int        // Number of component layers

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

// ExecutionLayer represents a single execution layer with its modules.
type ExecutionLayer struct {
	Modules []ExecutionModule
}

// ExecutionModule represents a module and its units of work within a layer.
type ExecutionModule struct {
	Name string     // Module name (e.g., "eac-cli")
	UoWs []UoWEntry // Units of work with ID and display name
}

// UoWEntry represents a unit of work with matching and display info.
type UoWEntry struct {
	ID          string // Full moniker for matching (Longname: context:module:component:tool)
	DisplayName string // Short name for display (e.g., "go", "godog")
	Weight      int    // Scheduling weight for resource allocation
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

// GetOrCreateUoWState gets or creates a module state for the given moniker.
// displayName is the short name for tab display.
func (m *Model) GetOrCreateUoWState(moniker, displayName string, weight int) *UoWState {
	if state, exists := m.uowStates[moniker]; exists {
		// Update weight if provided (in case it changed)
		if weight > 0 {
			state.Weight = weight
		}
		// Update display name if provided
		if displayName != "" {
			state.DisplayName = displayName
		}
		return state
	}

	// DEBUG: Log module registration for TUI caching investigation
	log.Debugf("[TUI-CACHE] GetOrCreateUoWState: registering module=%s displayName=%s weight=%d", moniker, displayName, weight)

	// Increment counter first to get unique index
	m.nextUoWIdx++

	// Default weight to 1 if not provided
	if weight <= 0 {
		weight = 1
	}

	// Create new module state with its own buffer
	// StartTime is left zero - will be set when MarkUoWRunning is called
	state := &UoWState{
		Moniker:     moniker,
		DisplayName: displayName,
		Index:       m.nextUoWIdx, // Unique 1-based index from counter
		Weight:      weight,
		Buffer:      NewRingBuffer(200),
		Status:      UoWPending, // Start as pending until slot acquired
	}
	m.uowStates[moniker] = state
	m.uowOrder = append(m.uowOrder, moniker)

	// Track total UoW count
	m.uowTotal++

	return state
}

// MarkUoWRunning marks a module as actively running (slot acquired).
// This sets the StartTime to now, so duration tracking reflects actual execution time.
// If the module is already in a terminal state (complete/skipped/failed), this is a no-op
// to prevent overwriting early cache detection results.
func (m *Model) MarkUoWRunning(moniker string) {
	state, exists := m.uowStates[moniker]
	if !exists {
		return
	}
	// Don't transition to running if already in a terminal state
	// This prevents background cache detection results from being overwritten
	if state.Status == UoWComplete || state.Status == UoWSkipped || state.Status == UoWFailed {
		return
	}
	// Only count transition from pending to running
	if state.Status == UoWPending {
		m.uowRunning++
	}
	state.Status = UoWRunning
	state.StartTime = time.Now() // Start timing when execution actually begins
}

// MarkUoWComplete marks a module as completed with the given exit code.
// Exit codes: 0=success(green), <0=skipped/cached(blue), >0=failed(red)
// If the tab is currently selected, it stays visible until user switches away.
func (m *Model) MarkUoWComplete(moniker string, exitCode int) {
	state, exists := m.uowStates[moniker]
	if !exists {
		// Module not registered - create it now so we can set the correct status
		// This handles race conditions where complete arrives before start
		log.Debugf("[TUI-CACHE] MarkUoWComplete: module=%s exitCode=%d - creating state", moniker, exitCode)
		state = m.GetOrCreateUoWState(moniker, "", 1)
	}

	// Check if already in a terminal state (don't double-count)
	oldStatus := state.Status
	alreadyComplete := oldStatus == UoWComplete || oldStatus == UoWSkipped || oldStatus == UoWFailed

	// Track running count - decrement if transitioning from running
	if oldStatus == UoWRunning && m.uowRunning > 0 {
		m.uowRunning--
	}

	// Always set the status from exit code - this is the authoritative source
	newStatus := StatusFromExitCode(exitCode)
	log.Debugf("[TUI-CACHE] MarkUoWComplete: module=%s exitCode=%d -> %v (was %v)", moniker, exitCode, newStatus, oldStatus)

	state.EndTime = time.Now()
	state.ExitCode = exitCode
	state.Status = newStatus

	// Track cumulative completion counters (only increment once, when first completing)
	if !alreadyComplete {
		switch newStatus {
		case UoWComplete:
			m.uowDone++
		case UoWSkipped:
			m.uowCached++
		case UoWFailed:
			m.uowFailed++
		}
	}

	// Tabs stay visible - only removed via FIFO when over limit
}

// removeUoWFromTabs removes a module from the tab display.
func (m *Model) removeUoWFromTabs(moniker string) {
	// Remove from order
	var newOrder []string
	for _, name := range m.uowOrder {
		if name != moniker {
			newOrder = append(newOrder, name)
		}
	}
	m.uowOrder = newOrder

	// Remove from states
	delete(m.uowStates, moniker)
}

// GetVisibleTabs returns the tabs that should be displayed.
// Tabs are shown in execution order (uowOrder). All tabs are kept - no deletion.
// maxTabs is ignored to preserve full visibility of execution state.
func (m *Model) GetVisibleTabs() []*UoWState {
	var tabs []*UoWState

	for _, moniker := range m.uowOrder {
		state := m.uowStates[moniker]
		if state == nil {
			continue
		}
		tabs = append(tabs, state)
	}

	return tabs
}

// SetActiveTab sets the currently active tab
// SetActiveTab sets the currently active tab.
// Tabs persist until FIFO removal, so any visible tab can be selected.
func (m *Model) SetActiveTab(moniker string) {
	// Empty moniker = aggregate view (always valid)
	if moniker == "" {
		m.activeTab = ""
		return
	}

	// Validate moniker exists in visible tabs
	_, exists := m.uowStates[moniker]
	if !exists {
		return
	}
	m.activeTab = moniker
}

// GetActiveUoWBuffer returns the buffer for the active tab.
// If no tab is selected, returns the first tab's buffer.
func (m *Model) GetActiveUoWBuffer() *RingBuffer {
	activeMoniker := m.getEffectiveActiveTab()
	if activeMoniker == "" {
		return nil // No tabs yet
	}
	if state, exists := m.uowStates[activeMoniker]; exists {
		return state.Buffer
	}
	return nil
}

// getEffectiveActiveTab returns the active tab, defaulting to first tab if none selected.
func (m *Model) getEffectiveActiveTab() string {
	if m.activeTab != "" {
		return m.activeTab
	}
	// Default to first tab
	if len(m.uowOrder) > 0 {
		return m.uowOrder[0]
	}
	return ""
}

// CleanupDecayedTabs is a no-op now (tabs removed instantly on completion).
func (m *Model) CleanupDecayedTabs() {
	// Tabs are now removed instantly when modules complete
	// This function is kept for compatibility but does nothing
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
	if len(m.locks) == 0 {
		return nil
	}

	var parts []string

	for _, lock := range m.locks {
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

// metricsUpdateInterval is how often to refresh CPU/memory metrics.
// These gopsutil calls are expensive (100-500ms on Windows), so we cache them.
const metricsUpdateInterval = 500 * time.Millisecond

// UpdateCachedMetrics refreshes the cached CPU and memory metrics.
// Should be called from the tick handler, not from View().
func (m *Model) UpdateCachedMetrics() {
	// Skip during exit sequence - no need to update metrics when quitting
	// Also skip once all runners are done (summary is being generated)
	if m.exitRequested || m.quitting || m.pendingSummaryData != nil || !m.allRunnersCompleted.IsZero() {
		return
	}

	// Skip if updated recently
	if time.Since(m.lastMetricsUpdate) < metricsUpdateInterval {
		return
	}

	// Update CPU metrics (this is the slow call)
	if perCore, err := cpu.Percent(0, true); err == nil {
		m.cachedCPUPercent = perCore
	}

	// Update memory metrics
	if memInfo, err := mem.VirtualMemory(); err == nil {
		m.cachedMemPercent = memInfo.UsedPercent
	}

	m.lastMetricsUpdate = time.Now()
}

// calculateOptimalTabColumns determines the minimum number of columns (1-6)
// needed to display all UoWs without scrolling in the visible panel area.
// Simple math: cols = ceil(totalUoWs / uowPerColumn)
func (m *Model) calculateOptimalTabColumns() int {
	// Use UoWCount from initSummary - this is the actual count of scheduled work items
	// (number of scheduled work items, not just components)
	totalUoWs := 0
	if m.initSummary != nil {
		totalUoWs = m.initSummary.UoWCount
	}

	// Get visible rows from layout metrics
	metrics := m.calculateLayoutMetrics()
	// Panel content height = remaining height - header(1) - footer(1)
	uowPerCol := metrics.RemainingHeight - 2
	if uowPerCol < 1 {
		uowPerCol = 1
	}

	// cols = ceil(totalUoWs / uowPerCol)
	cols := (totalUoWs + uowPerCol - 1) / uowPerCol

	// Clamp to valid range (min 2 for better readability)
	if cols < 2 {
		cols = 2
	}
	if cols > 6 {
		cols = 6
	}

	return cols
}

// ComponentsWidth returns the width of the components panel based on tab columns.
func (m Model) ComponentsWidth() int {
	const tabWidth = 15
	// Width = columns * tabWidth + spaces_between_tabs + borders
	// Tabs are joined with spaces, so N columns = N-1 spaces between them
	// Borders = 2 (left + right)
	width := m.tabColumns*tabWidth + (m.tabColumns - 1) + 2
	if width < 20 {
		width = 20
	}
	return width
}
