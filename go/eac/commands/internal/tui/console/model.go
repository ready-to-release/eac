package console

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Model is the Bubbletea model for the console window.
// Displays build/test output in a 3-pane view (Init/Run/End).
type Model struct {
	// Configuration
	height     int  // Total height (default: tui.DefaultHeight)
	width      int  // Terminal width
	showHeader bool // Show status header

	// 3-pane state
	panes       [3]*Pane // Init, Run, End panes
	activePhase Phase    // Currently active phase

	// Legacy single-buffer support (for backward compatibility)
	buffer *RingBuffer // Shared buffer when not using panes

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

	// Done state
	linesDone  bool
	statusDone bool

	// Feature flag for 3-pane mode
	usePanes bool

	// Quitting state - triggers plain-text final render
	quitting bool
}

// NewModel creates a new console model.
func NewModel(height int, showHeader bool, lineChan <-chan Line, statusChan <-chan Status) Model {
	if height <= 0 {
		height = 5
	}

	// Create panes with appropriate buffer sizes
	bufferSize := 500 // Per-pane buffer
	panes := [3]*Pane{
		NewPane(PhaseInit, bufferSize),
		NewPane(PhaseRun, bufferSize),
		NewPane(PhaseEnd, bufferSize),
	}

	return Model{
		height:      height,
		width:       80, // Default, will be updated on WindowSizeMsg
		showHeader:  showHeader,
		buffer:      NewRingBuffer(1000), // Legacy buffer
		panes:       panes,
		activePhase: PhaseInit, // Start with Init phase
		lineChan:    lineChan,
		statusChan:  statusChan,
		startTime:   time.Now(),
		phase:       "Starting",
		usePanes:    true, // Enable 3-pane mode by default
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
func (m Model) calculatePaneHeights() (initH, runH, endH int) {
	available := m.height

	// Reserve 1 line for each pane header
	headerLines := 3
	contentLines := available - headerLines
	if contentLines < 3 {
		contentLines = 3
	}

	// Distribute based on active phase
	// Init always gets 3 lines minimum to show context info
	const initMinHeight = 3

	switch m.activePhase {
	case PhaseInit:
		// Init gets most space, others get 0 content
		initH = contentLines
		runH = 0
		endH = 0
	case PhaseRun:
		// Init keeps 3 lines, Run gets the rest
		initH = initMinHeight
		runH = contentLines - initMinHeight
		if runH < 3 {
			runH = 3
		}
		endH = 0
	case PhaseEnd:
		// Init keeps 3 lines, Run gets 3 lines, End gets exactly 1 line (just result)
		// Detailed summary rolls off to console after TUI stops
		initH = initMinHeight
		runH = 3
		endH = 1
	}

	return
}

// UsePanes returns whether 3-pane mode is enabled
func (m Model) UsePanes() bool {
	return m.usePanes
}

// SetUsePanes enables or disables 3-pane mode
func (m *Model) SetUsePanes(use bool) {
	m.usePanes = use
}
