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

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case lineMsg:
		line := Line(msg)
		// Write to active phase's buffer in pane mode
		if m.usePanes {
			m.panes[m.activePhase].Buffer.Push(line)
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
			m.panes[msg.Phase].Buffer.Push(msg.Line)
		}
		m.buffer.Push(msg.Line)
		return m, nil

	case PhaseUpdateMsg:
		if m.panes[msg.Phase] != nil {
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
				if m.activePhase != msg.Phase && m.panes[m.activePhase].Status == PhaseActive {
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
		// Auto-quit when both channels are done (after one final render)
		if m.linesDone && m.statusDone {
			m.quitting = true // Triggers plain-text final render
			return m, tea.Quit
		}
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
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

// IsDone returns true if both line and status channels are done.
func (m Model) IsDone() bool {
	return m.linesDone && m.statusDone
}
