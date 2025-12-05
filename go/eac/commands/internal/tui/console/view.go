package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the console window.
func (m Model) View() string {
	// Final render: plain text without ANSI colors
	if m.quitting {
		return m.viewFinal()
	}
	if m.usePanes {
		return m.viewPanes()
	}
	return m.viewLegacy()
}

// viewFinal renders a clean plain-text summary (no ANSI escape codes)
// Outputs enough blank lines first to overwrite the styled TUI content
func (m Model) viewFinal() string {
	var b strings.Builder

	// Clear the TUI area by outputting blank lines to match height
	// This overwrites the styled pane content that was displayed
	for i := 0; i < m.height; i++ {
		b.WriteString(strings.Repeat(" ", m.width) + "\n")
	}

	// Move cursor back up to overwrite the blank lines with our summary
	// Use ANSI escape: \033[<n>A moves cursor up n lines
	b.WriteString(fmt.Sprintf("\033[%dA", m.height))

	elapsed := time.Since(m.startTime).Round(time.Millisecond * 100)

	// Determine overall status
	hasFailure := false
	for _, pane := range m.panes {
		if pane.Status == PhaseFailed {
			hasFailure = true
			break
		}
	}

	// Simple status line
	icon := "+"
	status := "complete"
	if hasFailure {
		icon = "x"
		status = "failed"
	}

	b.WriteString(fmt.Sprintf("[%s] %s: %d/%d %s (%s)\n",
		icon, m.phase, m.completed, m.total, status, formatElapsed(elapsed)))

	// Show phase summaries if available
	for _, phase := range []Phase{PhaseInit, PhaseRun, PhaseEnd} {
		pane := m.panes[phase]
		if pane.Summary != "" {
			phaseIcon := "+"
			if pane.Status == PhaseFailed {
				phaseIcon = "x"
			}
			b.WriteString(fmt.Sprintf("  [%s] %s: %s\n", phaseIcon, phase.String(), pane.Summary))
		}
	}

	// Show last error if any
	if m.lastError != nil {
		b.WriteString(fmt.Sprintf("  [!] %s\n", m.lastError.Text))
	}

	return b.String()
}

// viewPanes renders the 3-pane layout
func (m Model) viewPanes() string {
	var b strings.Builder

	// Calculate heights for each pane
	initH, runH, endH := m.calculatePaneHeights()

	// Render Init pane
	b.WriteString(m.renderPaneHeader(PhaseInit))
	if initH > 0 {
		b.WriteString("\n")
		b.WriteString(m.renderPaneContent(PhaseInit, initH))
	}
	b.WriteString("\n")

	// Render Run pane
	b.WriteString(m.renderPaneHeader(PhaseRun))
	if runH > 0 {
		b.WriteString("\n")
		b.WriteString(m.renderPaneContent(PhaseRun, runH))
	}
	b.WriteString("\n")

	// Render End pane
	b.WriteString(m.renderPaneHeader(PhaseEnd))
	if endH > 0 {
		b.WriteString("\n")
		b.WriteString(m.renderPaneContent(PhaseEnd, endH))
	}

	return b.String()
}

// renderPaneHeader renders a pane's header line
func (m Model) renderPaneHeader(phase Phase) string {
	pane := m.panes[phase]

	// Build icon and name
	icon := pane.Status.Icon()
	name := phase.String()

	// Style based on status
	var iconStyle, nameStyle lipgloss.Style
	switch pane.Status {
	case PhaseActive:
		iconStyle = Styles.Running
		nameStyle = Styles.Phase
	case PhaseComplete:
		iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Green
		nameStyle = Styles.Dim
	case PhaseFailed:
		iconStyle = Styles.Error
		nameStyle = Styles.Error
	default:
		iconStyle = Styles.Dim
		nameStyle = Styles.Dim
	}

	// Build header content
	left := iconStyle.Render(icon) + " " + nameStyle.Render(name)

	// Add status info for Run phase when active
	if phase == PhaseRun && pane.Status == PhaseActive {
		elapsed := time.Since(m.startTime).Round(time.Millisecond * 100)

		// Build status indicators
		var status string
		if len(m.running) > 0 {
			if len(m.running) <= 2 {
				status = strings.Join(m.running, ", ")
			} else {
				status = fmt.Sprintf("%s +%d more", m.running[0], len(m.running)-1)
			}
		}

		left = fmt.Sprintf("%s %s │ %d/%d",
			left,
			Styles.Time.Render(formatElapsed(elapsed)),
			m.completed,
			m.total,
		)

		if status != "" {
			left += " " + Styles.Running.Render(IconRunning+" "+status)
		}
	}

	// Add summary for completed phases
	if (pane.Status == PhaseComplete || pane.Status == PhaseFailed) && pane.Summary != "" {
		left += ": " + Styles.Dim.Render(pane.Summary)
	}

	// Add "waiting..." for pending phases
	if pane.Status == PhasePending {
		left += ": " + Styles.Dim.Render("waiting...")
	}

	// Border line
	borderLen := m.width - lipgloss.Width(left) - 2
	if borderLen < 3 {
		borderLen = 3
	}
	border := Styles.Border.Render("─" + strings.Repeat("─", borderLen))

	return "┌" + left + " " + border + "┐"
}

// renderPaneContent renders the content lines for a pane
func (m Model) renderPaneContent(phase Phase, height int) string {
	var b strings.Builder
	pane := m.panes[phase]

	// Get lines from this pane's buffer
	lines := pane.Buffer.Last(height)

	for i := 0; i < height; i++ {
		if i < len(lines) {
			b.WriteString(m.renderLine(lines[i]))
		} else {
			b.WriteString(Styles.Dim.Render("│ "))
		}
		if i < height-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// viewLegacy renders the original single-buffer view (for backward compatibility)
func (m Model) viewLegacy() string {
	var b strings.Builder

	// Calculate available lines for output
	outputLines := m.height
	if m.showHeader {
		outputLines--
		b.WriteString(m.renderHeader())
		b.WriteString("\n")
	}

	// Get lines to display
	lines := m.getDisplayLines(outputLines)

	for i, line := range lines {
		b.WriteString(m.renderLine(line))
		if i < len(lines)-1 {
			b.WriteString("\n")
		}
	}

	// Pad with empty lines if needed
	for i := len(lines); i < outputLines; i++ {
		if i > 0 || len(lines) > 0 {
			b.WriteString("\n")
		}
		b.WriteString(Styles.Dim.Render("│ "))
	}

	return b.String()
}

func (m Model) renderHeader() string {
	elapsed := time.Since(m.startTime).Round(time.Millisecond * 100)

	// Build status indicators
	var status string
	if len(m.running) > 0 {
		if len(m.running) <= 2 {
			status = strings.Join(m.running, ", ")
		} else {
			status = fmt.Sprintf("%s +%d more", m.running[0], len(m.running)-1)
		}
	}

	// Format: [Phase] 12.3s | 3/10 | running: module-a, module-b
	left := fmt.Sprintf("%s %s │ %d/%d",
		Styles.Phase.Render(m.phase),
		Styles.Time.Render(formatElapsed(elapsed)),
		m.completed,
		m.total,
	)

	right := ""
	if status != "" {
		right = Styles.Running.Render(IconRunning + " " + status)
	}
	if m.paused {
		right = Styles.Paused.Render(IconPaused+" PAUSED") + " " + right
	}
	if m.errorMode {
		right = Styles.Error.Render("ERRORS ONLY") + " " + right
	}

	// Fit to width
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 0 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}

func (m Model) renderLine(line Line) string {
	// Truncate to fit width
	maxLen := m.width - 4 // Account for prefix
	if maxLen < 10 {
		maxLen = 10
	}
	text := line.Text
	if len(text) > maxLen {
		text = text[:maxLen-1] + "…"
	}

	// Style based on level
	var prefix, styled string
	switch line.Level {
	case LevelError:
		prefix = Styles.ErrorPrefix.Render("│" + IconFail)
		styled = Styles.Error.Render(text)
	case LevelWarn:
		prefix = Styles.WarnPrefix.Render("│" + IconWarn)
		styled = Styles.Warn.Render(text)
	default:
		prefix = Styles.InfoPrefix.Render("│" + IconInfo)
		styled = Styles.Info.Render(text)
	}

	return prefix + " " + styled
}

func (m Model) getDisplayLines(count int) []Line {
	if m.errorMode {
		// Filter to errors only
		return m.buffer.LastByLevel(count, LevelError)
	}

	return m.buffer.Last(count)
}

// formatElapsed formats a duration for display.
func formatElapsed(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}
