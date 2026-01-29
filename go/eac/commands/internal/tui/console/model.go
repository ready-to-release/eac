package console

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

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
	activeContainers  []string // Currently active container tools (orange)
	usedContainers    []string // All container tools ever used
	activeSystemTools []string // Currently active system tools (orange)
	usedSystemTools   []string // All system tools ever used

	// Per-module tab tracking for Run phase
	moduleStates    map[string]*ModuleState // Per-module state (running, completed, failed)
	moduleOrder     []string                // Order in which modules started (for tab ordering)
	nextModuleIdx   int                     // Counter for assigning unique indices to modules
	activeTab       string                  // Currently selected tab ("" = aggregate view)
	hoveredTab      string                  // Tab currently under mouse cursor
	hoveredTabScroll int                    // Marquee scroll offset for hovered tab name
	maxTabs         int                     // Maximum visible tabs before scrolling/hiding

	// Channels for async updates
	lineChan   <-chan Line   // Incoming output lines
	statusChan <-chan Status // Status updates

	// Display preferences
	paused    bool     // Pause scrolling (for review)
	errorMode bool     // Show only errors
	mouseMode bool     // Mouse mode: true=scrolling enabled, false=text selection enabled
	viewMode         ViewMode // Components view mode: TabGrid or Tree
	tabsScrollOffset int      // Scroll offset for tabs/tree panel

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
}

// ModuleState tracks per-module execution state for tab display.
type ModuleState struct {
	Moniker   string       // Module identifier
	Index     int          // 1-based index in execution order (for tab display)
	Weight    int          // Scheduling weight/pressure (shown in tab)
	Buffer    *RingBuffer  // Module-specific output buffer
	Status    ModuleStatus // Running, Complete, Failed
	StartTime time.Time    // When module started
	EndTime   time.Time    // When module finished (zero if running)
	ExitCode  int          // Exit code (only valid when complete/failed)
	DecayTime time.Time    // When tab should disappear (zero = don't decay)
	CacheTime time.Time    // For cached modules: when the artifact was last built
	LogPath   string       // Path to build log file (if available)
}

// ModuleStatus represents the execution state of a module.
type ModuleStatus int

const (
	ModulePending  ModuleStatus = iota // Scheduled, waiting for slot
	ModuleRunning                      // Actively executing
	ModuleComplete                     // Finished successfully
	ModuleSkipped                      // Skipped (cached, unchanged)
	ModuleFailed                       // Finished with error
)

// ViewMode represents the display mode for the components pane.
type ViewMode int

const (
	ViewModeTabGrid ViewMode = iota // Tab grid view (default)
	ViewModeTree                    // Execution tree view
)

// Icon returns the icon for a module status (ASCII-safe).
func (s ModuleStatus) Icon() string {
	switch s {
	case ModulePending:
		return "o" // Waiting
	case ModuleRunning:
		return ">" // Active
	case ModuleComplete:
		return "V" // Done
	case ModuleSkipped:
		return "=" // Cached/Skipped
	case ModuleFailed:
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
func (s ModuleStatus) Colors() StatusColors {
	switch s {
	case ModulePending:
		return StatusColors{Border: "238", Text: "245", Bg: "234"} // Gray
	case ModuleRunning:
		return StatusColors{Border: "214", Text: "214", Bg: "94"} // Orange/Yellow
	case ModuleComplete:
		return StatusColors{Border: "40", Text: "40", Bg: "22"} // Green
	case ModuleSkipped:
		return StatusColors{Border: "75", Text: "75", Bg: "23"} // Cyan/Blue (cached)
	case ModuleFailed:
		return StatusColors{Border: "196", Text: "196", Bg: "52"} // Red
	default:
		return StatusColors{Border: "238", Text: "245", Bg: "234"} // Gray
	}
}

// StatusFromExitCode returns the appropriate ModuleStatus for an exit code.
// exitCode == 0: Complete (success)
// exitCode < 0: Skipped (cached)
// exitCode > 0: Failed
func StatusFromExitCode(exitCode int) ModuleStatus {
	if exitCode == 0 {
		return ModuleComplete
	} else if exitCode < 0 {
		return ModuleSkipped
	}
	return ModuleFailed
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
		moduleStates:      make(map[string]*ModuleState), // Per-module state tracking
		moduleOrder:       make([]string, 0),             // Tab ordering
		activeTab:         "",                            // Start with aggregate view
		maxTabs:           36,                            // Maximum visible tabs (6 rows × 6 tabs/row)
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

	// Component count
	ComponentCount int

	// Execution tree - for visual tree rendering
	ExecutionTree         []ExecutionLayer // Full tree: layers → modules → components
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

// GetOrCreateModuleState gets or creates a module state for the given moniker.
func (m *Model) GetOrCreateModuleState(moniker string, weight int) *ModuleState {
	if state, exists := m.moduleStates[moniker]; exists {
		// Update weight if provided (in case it changed)
		if weight > 0 {
			state.Weight = weight
		}
		return state
	}

	// DEBUG: Log module registration for TUI caching investigation
	log.Debugf("[TUI-CACHE] GetOrCreateModuleState: registering module=%s weight=%d", moniker, weight)

	// Increment counter first to get unique index
	m.nextModuleIdx++

	// Default weight to 1 if not provided
	if weight <= 0 {
		weight = 1
	}

	// Create new module state with its own buffer
	// StartTime is left zero - will be set when MarkModuleRunning is called
	state := &ModuleState{
		Moniker: moniker,
		Index:   m.nextModuleIdx, // Unique 1-based index from counter
		Weight:  weight,
		Buffer:  NewRingBuffer(200),
		Status:  ModulePending, // Start as pending until slot acquired
	}
	m.moduleStates[moniker] = state
	m.moduleOrder = append(m.moduleOrder, moniker)

	return state
}

// MarkModuleRunning marks a module as actively running (slot acquired).
// This sets the StartTime to now, so duration tracking reflects actual execution time.
func (m *Model) MarkModuleRunning(moniker string) {
	state, exists := m.moduleStates[moniker]
	if !exists {
		return
	}
	state.Status = ModuleRunning
	state.StartTime = time.Now() // Start timing when execution actually begins
}

// MarkModuleComplete marks a module as completed with the given exit code.
// Exit codes: 0=success(green), <0=skipped/cached(blue), >0=failed(red)
// If the tab is currently selected, it stays visible until user switches away.
func (m *Model) MarkModuleComplete(moniker string, exitCode int) {
	state, exists := m.moduleStates[moniker]
	if !exists {
		// Module not registered - create it now so we can set the correct status
		// This handles race conditions where complete arrives before start
		log.Debugf("[TUI-CACHE] MarkModuleComplete: module=%s exitCode=%d - creating state", moniker, exitCode)
		state = m.GetOrCreateModuleState(moniker, 1)
	}

	// Always set the status from exit code - this is the authoritative source
	newStatus := StatusFromExitCode(exitCode)
	log.Debugf("[TUI-CACHE] MarkModuleComplete: module=%s exitCode=%d -> %v", moniker, exitCode, newStatus)

	state.EndTime = time.Now()
	state.ExitCode = exitCode
	state.Status = newStatus

	// Tabs stay visible - only removed via FIFO when over limit
}

// removeModuleFromTabs removes a module from the tab display.
func (m *Model) removeModuleFromTabs(moniker string) {
	// Remove from order
	var newOrder []string
	for _, name := range m.moduleOrder {
		if name != moniker {
			newOrder = append(newOrder, name)
		}
	}
	m.moduleOrder = newOrder

	// Remove from states
	delete(m.moduleStates, moniker)
}

// GetVisibleTabs returns the tabs that should be displayed.
// Tabs are shown in execution order (moduleOrder). All tabs are kept - no deletion.
// maxTabs is ignored to preserve full visibility of execution state.
func (m *Model) GetVisibleTabs() []*ModuleState {
	var tabs []*ModuleState

	for _, moniker := range m.moduleOrder {
		state := m.moduleStates[moniker]
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
	_, exists := m.moduleStates[moniker]
	if !exists {
		return
	}
	m.activeTab = moniker
}

// GetActiveModuleBuffer returns the buffer for the active tab.
// If no tab is selected, returns the first tab's buffer.
func (m *Model) GetActiveModuleBuffer() *RingBuffer {
	activeMoniker := m.getEffectiveActiveTab()
	if activeMoniker == "" {
		return nil // No tabs yet
	}
	if state, exists := m.moduleStates[activeMoniker]; exists {
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
	if len(m.moduleOrder) > 0 {
		return m.moduleOrder[0]
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
