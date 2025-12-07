package console

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Reset scroll offsets when window size changes
		if m.usePanes {
			for _, pane := range m.panes {
				if pane != nil && !pane.autoScroll {
					// If user was scrolled up, try to maintain position, but cap at new max
					initH, runH, summaryH := m.calculatePaneHeights()
					paneHeight := initH
					if pane.Phase == PhaseRun {
						paneHeight = runH
					} else if pane.Phase == PhaseSummary {
						paneHeight = summaryH
					}
					pane.UpdateMaxScroll(paneHeight)
				}
			}
		}
		return m, nil

	case lineMsg:
		line := Line(msg)
		// Write to active phase's buffer in pane mode
		if m.usePanes {
			pane := m.panes[m.activePhase]
			pane.Buffer.Push(line)
			// If pane is scrolled up (not auto-scrolling), increment offset to keep view locked
			if !pane.autoScroll && pane.scrollOffset > 0 {
				pane.scrollOffset++
			}
		}
		// Also write to legacy buffer for backward compatibility
		m.buffer.Push(line)

		// Track errors for sticky display
		if line.Level == LevelError {
			m.lastError = &line
		}

		// Continue listening for more lines
		if !m.linesDone {
			return m, m.listenForLines()
		}
		return m, nil

	case PhaseLineMsg:
		// Write to specific phase's buffer
		if m.usePanes && m.panes[msg.Phase] != nil {
			pane := m.panes[msg.Phase]
			pane.Buffer.Push(msg.Line)
			// If pane is scrolled up (not auto-scrolling), increment offset to keep view locked
			if !pane.autoScroll && pane.scrollOffset > 0 {
				pane.scrollOffset++
			}
		}
		m.buffer.Push(msg.Line)
		return m, nil

	case ResultLineMsg:
		// Write to results buffer
		m.resultsBuffer.Push(msg.Line)
		return m, nil

	case SummaryDataMsg:
		// Set summary data and activate Summary pane
		m.summaryData = msg.Data
		// Automatically activate Summary pane
		if m.activePhase != PhaseSummary {
			// Mark current phase as complete
			if m.panes[m.activePhase].Status == PhaseActive {
				m.panes[m.activePhase].Status = PhaseComplete
				m.panes[m.activePhase].EndTime = time.Now()
			}
			// Activate Summary pane
			m.activePhase = PhaseSummary
			m.panes[PhaseSummary].Status = PhaseActive
			m.panes[PhaseSummary].StartTime = time.Now()
		}
		// Wait for user to press any key before exiting
		m.waitingForExit = true
		// Start auto-exit timer (0.5 seconds)
		return m, m.autoExitTimer()

	case PhaseUpdateMsg:
		if msg.Phase < Phase(len(m.panes)) && m.panes[msg.Phase] != nil {
			// Only update status if it's set (non-zero)
			if msg.Status != 0 {
				m.panes[msg.Phase].Status = msg.Status
			}
			// Update summary if provided
			if msg.Summary != "" {
				m.panes[msg.Phase].Summary = msg.Summary
			}

			// Track timing and active phase
			if msg.Status == PhaseActive {
				// Mark previous phase as complete if it was active
				if m.activePhase != msg.Phase && m.activePhase < Phase(len(m.panes)) && m.panes[m.activePhase].Status == PhaseActive {
					m.panes[m.activePhase].Status = PhaseComplete
					m.panes[m.activePhase].EndTime = time.Now()
				}
				m.panes[msg.Phase].StartTime = time.Now()
				m.activePhase = msg.Phase
			} else if msg.Status == PhaseComplete || msg.Status == PhaseFailed {
				m.panes[msg.Phase].EndTime = time.Now()
			}
		}
		return m, nil

	case statusMsg:
		status := Status(msg)
		m.phase = status.Phase
		m.running = status.Running
		m.completed = status.Completed
		m.total = status.Total

		if !m.statusDone {
			return m, m.listenForStatus()
		}
		return m, nil

	case tickMsg:
		// Never auto-quit - always wait for user to press a key
		// Quitting only happens via handleKey() when waitingForExit is true
		return m, m.tickCmd()

	case linesDoneMsg:
		m.linesDone = true
		return m, nil

	case statusDoneMsg:
		m.statusDone = true
		return m, nil

	case completedMsg:
		// A module completed - update display
		m.completed++
		// Remove from running list
		var newRunning []string
		for _, r := range m.running {
			if r != msg.Moniker {
				newRunning = append(newRunning, r)
			}
		}
		m.running = newRunning
		return m, nil

	case autoExitTimerMsg:
		// Auto-exit if: still waiting for exit AND not already quitting
		if m.waitingForExit && !m.quitting {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If waiting for user to exit, any key press quits
	if m.waitingForExit {
		m.quitting = true
		return m, tea.Quit
	}

	// Normal key handling
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case " ", "p":
		// Toggle pause
		m.paused = !m.paused
	case "e":
		// Toggle error-only mode
		m.errorMode = !m.errorMode
	case "c":
		// Clear last error
		m.lastError = nil
	}
	return m, nil
}

// handleMouse handles mouse events for pane scrolling.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Mouse scrolling:
	// - Scroll wheel → scroll panes
	// - Shift+Click → select text (standard terminal behavior, bypasses mouse mode)

	// Wheel events: handle scrolling
	if msg.Type == tea.MouseWheelUp || msg.Type == tea.MouseWheelDown {

		// Only handle scrolling in pane mode
		if !m.usePanes {
			return m, nil
		}

		// Determine which pane the mouse is over
		paneIdx := m.getPaneAtPosition(msg.Y)
		if paneIdx < 0 || paneIdx >= len(m.panes) {
			return m, nil // Mouse not over any pane
		}

		pane := m.panes[paneIdx]
		if pane == nil {
			return m, nil
		}

		// Calculate pane height for this specific pane
		initH, runH, summaryH := m.calculatePaneHeights()
		paneHeight := initH
		if paneIdx == 1 { // Run pane
			paneHeight = runH
		} else if paneIdx == 2 { // Summary pane
			paneHeight = summaryH
		}

		// Update max scroll based on current buffer size
		pane.UpdateMaxScroll(paneHeight)

		// Scroll the pane
		scrollAmount := 3 // Lines to scroll per wheel tick
		if msg.Type == tea.MouseWheelUp {
			pane.ScrollUp(scrollAmount)
		} else {
			pane.ScrollDown(scrollAmount)
		}

		return m, nil
	}

	return m, nil
}

// getPaneAtPosition determines which pane (0, 1, or 2) is at the given Y coordinate.
// Returns -1 if not over any pane content area.
// All panes are always visible with dynamic heights based on terminal size.
func (m Model) getPaneAtPosition(y int) int {
	if !m.usePanes {
		return -1
	}

	// Calculate pane boundaries dynamically based on terminal height
	initH, runH, summaryH := m.calculatePaneHeights()

	// Pane layout (each pane has header + content + footer):
	// Init pane
	currentLine := 1
	initContentStart := currentLine
	initContentEnd := currentLine + initH - 1
	currentLine += initH + 1 // +1 for footer

	// Run pane (dynamic height, fills available space)
	currentLine++ // header
	runContentStart := currentLine
	runContentEnd := currentLine + runH - 1
	currentLine += runH + 1 // +1 for footer

	// Summary pane
	currentLine++ // header
	summaryContentStart := currentLine
	summaryContentEnd := currentLine + summaryH - 1

	// Check which pane content area the Y-coordinate falls into
	// Add tolerance of ±1 to handle edge cases with terminal scrolling
	if y >= initContentStart-1 && y <= initContentEnd+1 {
		return 0 // Init pane content
	} else if y >= runContentStart-1 && y <= runContentEnd+1 {
		return 1 // Run pane content
	} else if y >= summaryContentStart-1 && y <= summaryContentEnd+1 {
		return 2 // Summary pane content
	}

	return -1 // Header/footer or outside panes
}

// IsDone returns true if both line and status channels are done.
func (m Model) IsDone() bool {
	return m.linesDone && m.statusDone
}
