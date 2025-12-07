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
	showHeader   bool   // Show status header
	runPhaseName string // Custom name for Run phase (e.g., "building", "testing")

	// 3-pane state
	panes       [3]*Pane     // Init, Run, Summary panes
	activePhase Phase        // Currently active phase
	summaryData *SummaryData // Structured data for Summary pane

	// Legacy single-buffer support (for backward compatibility)
	buffer *RingBuffer // Shared buffer when not using panes

	// Results buffer for post-execution output
	resultsBuffer *RingBuffer // Output that appears after Run phase completes

	// Run phase state (orchestrator status)
	running   []string  // Currently running module monikers
	completed int       // Completed count
	total     int       // Total modules
	phase     string    // Current phase name (legacy)
	startTime time.Time // Execution start
	lastError *Line     // Most recent error (sticky display)

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

	// Feature flag for 3-pane mode
	usePanes bool

	// Quitting state - triggers plain-text final render
	quitting bool

	// Waiting for user to press any key before exiting
	waitingForExit bool
}

// NewModel creates a new console model.
func NewModel(height int, showHeader bool, runPhaseName string, lineChan <-chan Line, statusChan <-chan Status) Model {
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
		showHeader:    showHeader,
		runPhaseName:  runPhaseName,
		buffer:        NewRingBuffer(1000), // Legacy buffer
		resultsBuffer: NewRingBuffer(100),  // Results buffer
		panes:         panes,
		activePhase:   PhaseInit, // Start with Init phase
		lineChan:      lineChan,
		statusChan:    statusChan,
		startTime:     time.Now(),
		phase:         "Starting",
		usePanes:      true, // Enable 2-pane mode by default
		mouseMode:     true, // Start with mouse ON (scrolling enabled)
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

// autoExitTimer returns a command that fires after 0.5 seconds to auto-exit.
func (m Model) autoExitTimer() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return autoExitTimerMsg{}
	})
}

// SetTotal sets the total number of modules.
func (m *Model) SetTotal(total int) {
	m.total = total
}

// SetPhase sets the current phase name (legacy).
func (m *Model) SetPhase(phase string) {
	m.phase = phase
}

// SetActivePhase switches to a new phase
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

// GetActivePane returns the currently active pane
func (m *Model) GetActivePane() *Pane {
	return m.panes[m.activePhase]
}

// WriteToPhase writes a line to a specific phase's buffer
func (m *Model) WriteToPhase(phase Phase, line Line) {
	if m.usePanes && m.panes[phase] != nil {
		m.panes[phase].Buffer.Push(line)
	}
	// Also write to legacy buffer for backward compatibility
	m.buffer.Push(line)
}

// SetPhaseSummary sets the summary text for a phase (shown when collapsed)
func (m *Model) SetPhaseSummary(phase Phase, summary string) {
	if m.panes[phase] != nil {
		m.panes[phase].Summary = summary
	}
}

// CompletePhase marks a phase as complete with success/failure status
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
// Dynamic heights: Init and Summary are fixed, Run fills remaining space
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

	return
}

// UsePanes returns whether 2-pane mode is enabled
func (m Model) UsePanes() bool {
	return m.usePanes
}

// SetUsePanes enables or disables 2-pane mode
func (m *Model) SetUsePanes(use bool) {
	m.usePanes = use
}

// WriteResult writes a line to the results buffer
func (m *Model) WriteResult(line Line) {
	m.resultsBuffer.Push(line)
}

// SummaryData holds structured information for the Summary pane
type SummaryData struct {
	Success     bool          // Overall success/failure
	TotalTime   time.Duration // Total execution time
	InitSummary string        // Init phase summary text
	RunSummary  string        // Run phase summary text
	Details     []string      // Detail lines (artifacts, errors, stats, etc.)
	NextSteps   string        // Suggested next action
}

// SetSummaryData updates the summary data for the Summary pane
func (m *Model) SetSummaryData(data *SummaryData) {
	m.summaryData = data
}
