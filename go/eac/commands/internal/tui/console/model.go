package console

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Model is the Bubbletea model for the console window.
// Displays build/test output in a 3-pane view (Init/Run/Summary).
type Model struct {
	// Configuration
	height       int    // Total height (default: tui.DefaultHeight)
	width        int    // Terminal width
	runPhaseName string // Custom name for Run phase (e.g., "building", "testing")

	// 3-pane state
	panes       [3]*Pane     // Init, Run, Summary panes
	activePhase Phase        // Currently active phase
	summaryData *SummaryData // Structured data for Summary pane

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

	// Per-module tab tracking for Run phase
	moduleStates  map[string]*ModuleState // Per-module state (running, completed, failed)
	moduleOrder   []string                // Order in which modules started (for tab ordering)
	nextModuleIdx int                     // Counter for assigning unique indices to modules
	activeTab     string                  // Currently selected tab ("" = aggregate view)
	maxTabs       int                     // Maximum visible tabs before scrolling/hiding

	// Channels for async updates
	lineChan   <-chan Line   // Incoming output lines
	statusChan <-chan Status // Status updates

	// Display preferences
	paused    bool // Pause scrolling (for review)
	errorMode bool // Show only errors
	mouseMode bool // Mouse mode: true=scrolling enabled, false=text selection enabled

	// Done state
	linesDone  bool
	statusDone bool

	// Quitting state - triggers plain-text final render
	quitting bool
}

// ModuleState tracks per-module execution state for tab display.
type ModuleState struct {
	Moniker   string       // Module identifier
	Index     int          // 1-based index in execution order (for tab display)
	Buffer    *RingBuffer  // Module-specific output buffer
	Status    ModuleStatus // Running, Complete, Failed
	StartTime time.Time    // When module started
	EndTime   time.Time    // When module finished (zero if running)
	ExitCode  int          // Exit code (only valid when complete/failed)
	DecayTime time.Time    // When tab should disappear (zero = don't decay)
}

// ModuleStatus represents the execution state of a module.
type ModuleStatus int

const (
	ModuleRunning ModuleStatus = iota
	ModuleComplete
	ModuleFailed
)

// Icon returns the icon for a module status.
func (s ModuleStatus) Icon() string {
	switch s {
	case ModuleRunning:
		return "▶"
	case ModuleComplete:
		return "✓"
	case ModuleFailed:
		return "✗"
	default:
		return "?"
	}
}

// NewModel creates a new console model.
func NewModel(height int, runPhaseName string, lineChan <-chan Line, statusChan <-chan Status) Model {
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
		height:        height,
		width:         80, // Default, will be updated on WindowSizeMsg
		runPhaseName:  runPhaseName,
		resultsBuffer: NewRingBuffer(100), // Results buffer
		panes:         panes,
		activePhase:   PhaseInit, // Start with Init phase
		lineChan:      lineChan,
		statusChan:    statusChan,
		startTime:     time.Now(),
		mouseMode:     true,                          // Start with mouse ON (scrolling enabled)
		moduleStates:  make(map[string]*ModuleState), // Per-module state tracking
		moduleOrder:   make([]string, 0),             // Tab ordering
		activeTab:     "",                            // Start with aggregate view
		maxTabs:       8,                             // Maximum visible tabs
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
// Dynamic heights: Init and Summary are fixed, Run fills remaining space.
func (m Model) calculatePaneHeights() (initH, runH, summaryH int) {
	// Total lines needed for headers and footers (3 headers + 3 footers)
	const headerFooterLines = 6

	// Fixed heights for Init and Summary panes
	const initHeight = 5
	const summaryHeight = 10

	// Minimum height for Run pane
	const minRunHeight = 5

	// Calculate available space for content
	availableForContent := m.height - headerFooterLines

	// Allocate heights: Init and Summary fixed, Run gets the rest
	initH = initHeight
	summaryH = summaryHeight
	runH = availableForContent - initH - summaryH

	// Ensure Run pane meets minimum height
	if runH < minRunHeight {
		runH = minRunHeight
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

// GetOrCreateModuleState gets or creates a module state for the given moniker.
func (m *Model) GetOrCreateModuleState(moniker string) *ModuleState {
	if state, exists := m.moduleStates[moniker]; exists {
		return state
	}

	// Increment counter first to get unique index
	m.nextModuleIdx++

	// Create new module state with its own buffer
	state := &ModuleState{
		Moniker:   moniker,
		Index:     m.nextModuleIdx, // Unique 1-based index from counter
		Buffer:    NewRingBuffer(200),
		Status:    ModuleRunning,
		StartTime: time.Now(),
	}
	m.moduleStates[moniker] = state
	m.moduleOrder = append(m.moduleOrder, moniker)

	return state
}

// MarkModuleComplete marks a module as completed
// If the tab is currently selected, it stays visible until user switches away.
func (m *Model) MarkModuleComplete(moniker string, exitCode int) {
	state, exists := m.moduleStates[moniker]
	if !exists {
		return
	}

	state.EndTime = time.Now()
	state.ExitCode = exitCode
	if exitCode == 0 {
		state.Status = ModuleComplete
	} else {
		state.Status = ModuleFailed
	}

	// Only remove from tabs if not currently selected
	// If selected, user can continue viewing it until they switch away
	if m.activeTab != moniker {
		m.removeModuleFromTabs(moniker)
	}
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

// GetVisibleTabs returns the tabs that should be displayed
// Shows running modules + the active tab (even if completed).
func (m *Model) GetVisibleTabs() []*ModuleState {
	var tabs []*ModuleState

	for _, moniker := range m.moduleOrder {
		state := m.moduleStates[moniker]
		if state == nil {
			continue
		}

		// Show running modules OR the currently selected tab (even if completed)
		if state.Status == ModuleRunning || moniker == m.activeTab {
			tabs = append(tabs, state)
		}
	}

	// Limit to maxTabs if too many
	if len(tabs) > m.maxTabs {
		tabs = tabs[:m.maxTabs]
	}

	return tabs
}

// SetActiveTab sets the currently active tab
// When switching away from a completed tab, it will be removed.
func (m *Model) SetActiveTab(moniker string) {
	oldTab := m.activeTab

	// Empty moniker = aggregate view (always valid)
	if moniker == "" {
		m.activeTab = ""
	} else {
		// Validate moniker exists (allow completed tabs that are still selected)
		state, exists := m.moduleStates[moniker]
		if !exists {
			return
		}
		// Only allow switching to running modules or the current active tab
		if state.Status != ModuleRunning && moniker != m.activeTab {
			return
		}
		m.activeTab = moniker
	}

	// If we switched away from a completed/failed tab, remove it now
	if oldTab != "" && oldTab != m.activeTab {
		if state, exists := m.moduleStates[oldTab]; exists {
			if state.Status == ModuleComplete || state.Status == ModuleFailed {
				m.removeModuleFromTabs(oldTab)
			}
		}
	}
}

// GetActiveModuleBuffer returns the buffer for the active tab, or nil for aggregate view.
func (m *Model) GetActiveModuleBuffer() *RingBuffer {
	if m.activeTab == "" {
		return nil // Aggregate view - use Run pane buffer
	}
	if state, exists := m.moduleStates[m.activeTab]; exists {
		return state.Buffer
	}
	return nil
}

// CleanupDecayedTabs is a no-op now (tabs removed instantly on completion).
func (m *Model) CleanupDecayedTabs() {
	// Tabs are now removed instantly when modules complete
	// This function is kept for compatibility but does nothing
}
