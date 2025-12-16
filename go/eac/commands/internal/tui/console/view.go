package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the console window.
func (m Model) View() string {
	// When quitting in alt screen mode, return empty (screen will be restored)
	// Plain-text summary is printed after program.Run() completes
	if m.quitting {
		return ""
	}
	return m.viewPanes()
}

// ViewFinal renders a clean plain-text version of all panes (no ANSI escape codes)
// Used for post-exit summary when using alt screen mode.
// Shows expanded content: up to 40 lines for Init, up to 30 for Run, all Summary details.
func (m Model) ViewFinal() string {
	var b strings.Builder

	// Final render uses expanded heights to show full content
	const (
		finalInitHeight    = 40 // Show full initialization output
		finalRunHeight     = 30 // Show recent run output
		finalSummaryHeight = 50 // Show all summary details including module table
	)

	// Render Init pane with expanded height
	b.WriteString(m.renderPaneHeaderPlain(PhaseInit))
	b.WriteString("\n")
	b.WriteString(m.renderPaneContentPlainExpanded(PhaseInit, finalInitHeight))
	b.WriteString("\n")
	b.WriteString(m.renderPaneFooterPlain(PhaseInit))
	b.WriteString("\n")

	// Render Run pane only if it actually ran (not still pending)
	if m.panes[PhaseRun].Status != PhasePending {
		b.WriteString(m.renderPaneHeaderPlain(PhaseRun))
		b.WriteString("\n")
		b.WriteString(m.renderPaneContentPlainExpanded(PhaseRun, finalRunHeight))
		b.WriteString("\n")
		b.WriteString(m.renderPaneFooterPlain(PhaseRun))
		b.WriteString("\n")
	}

	// Render Summary pane (only if data exists)
	if m.summaryData != nil {
		b.WriteString(m.renderPaneHeaderPlain(PhaseSummary))
		b.WriteString("\n")
		b.WriteString(m.renderSummaryContentPlainExpanded())
		b.WriteString("\n")
		b.WriteString(m.renderPaneFooterPlain(PhaseSummary))
		b.WriteString("\n")
	}

	return b.String()
}

// viewPanes renders the 3-pane layout with panes appearing progressively
func (m Model) viewPanes() string {
	var b strings.Builder

	// Calculate heights for each pane
	initH, runH, summaryH := m.calculatePaneHeights()

	// Render Init pane (always visible)
	b.WriteString(m.renderPaneHeader(PhaseInit))
	b.WriteString("\n")
	b.WriteString(m.renderPaneContent(PhaseInit, initH))
	b.WriteString("\n")
	b.WriteString(m.renderPaneFooter(PhaseInit, initH))
	b.WriteString("\n")

	// Render Run pane only if it actually started (not still pending)
	if m.panes[PhaseRun].Status != PhasePending {
		b.WriteString(m.renderPaneHeader(PhaseRun))
		b.WriteString("\n")

		// Always render tab bar (All tab is always visible)
		tabs := m.GetVisibleTabs()
		b.WriteString(m.renderTabBar(tabs))
		b.WriteString("\n")
		// Reduce run pane content height by 1 for the tab bar
		runH--

		b.WriteString(m.renderPaneContent(PhaseRun, runH))
		b.WriteString("\n")
		b.WriteString(m.renderPaneFooter(PhaseRun, runH))
		b.WriteString("\n")
	}

	// Render Summary pane (only when summary data is available)
	if m.summaryData != nil {
		b.WriteString(m.renderPaneHeader(PhaseSummary))
		b.WriteString("\n")
		b.WriteString(m.renderPaneContent(PhaseSummary, summaryH))
		b.WriteString("\n")
		b.WriteString(m.renderPaneFooter(PhaseSummary, summaryH))
		b.WriteString("\n")
	}

	// No results section in pane view - results appear after TUI exits
	// This ensures the view height stays constant (prevents cursor misalignment in inline mode)

	return b.String()
}

// renderTabBar renders the horizontal tab bar for module switching
func (m Model) renderTabBar(tabs []*ModuleState) string {
	var b strings.Builder

	// Left border
	b.WriteString(Styles.Border.Render("│"))

	// Render "All" tab first (aggregate view)
	allLabel := "All"
	if m.activeTab == "" {
		b.WriteString(Styles.TabActive.Render(allLabel))
	} else {
		b.WriteString(Styles.TabComplete.Render(allLabel))
	}

	// Calculate available width for module tabs
	// Account for borders, "All" tab, and separators
	usedWidth := 2 + lipgloss.Width(allLabel) + 2 // borders + all tab

	// Render module tabs
	for _, state := range tabs {
		// Tab separator
		b.WriteString(Styles.TabSeparator.Render(" "))
		usedWidth++

		// Truncate long module names
		label := state.Moniker
		maxLabelLen := 12
		if len(label) > maxLabelLen {
			label = label[:maxLabelLen-1] + "…"
		}

		// Add status icon
		icon := state.Status.Icon()
		tabText := icon + " " + label

		// Check if we have room for this tab
		tabWidth := lipgloss.Width(tabText) + 2 // +2 for padding
		if usedWidth+tabWidth > m.width-4 {
			// Show overflow indicator
			b.WriteString(Styles.Dim.Render("…"))
			break
		}

		// Style based on selection and status
		var style lipgloss.Style
		if state.Moniker == m.activeTab {
			style = Styles.TabActive
		} else {
			switch state.Status {
			case ModuleRunning:
				style = Styles.TabRunning
			case ModuleComplete:
				style = Styles.TabComplete
			case ModuleFailed:
				style = Styles.TabFailed
			}
		}

		b.WriteString(style.Render(tabText))
		usedWidth += tabWidth
	}

	// Fill remaining space and right border
	remaining := m.width - usedWidth - 1
	if remaining > 0 {
		b.WriteString(strings.Repeat(" ", remaining))
	}
	b.WriteString(Styles.Border.Render("│"))

	return b.String()
}

// renderPaneHeader renders a pane's header line
func (m Model) renderPaneHeader(phase Phase) string {
	pane := m.panes[phase]

	// Build icon and name
	icon := pane.Status.Icon()
	name := phase.String()

	// Use custom run phase name if provided
	if phase == PhaseRun && m.runPhaseName != "" {
		name = m.runPhaseName
	}

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

		// Add layer info if using layered execution
		if m.totalLayers > 0 && m.layer > 0 {
			left += " " + Styles.Dim.Render(fmt.Sprintf("(layer %d/%d)", m.layer, m.totalLayers))
		}

		// Show elapsed time and progress count (running modules shown via tabs)
		left = fmt.Sprintf("%s %s %d/%d",
			left,
			Styles.Time.Render(formatElapsed(elapsed)),
			m.completed,
			m.total,
		)
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

	// Special handling for Summary pane: render structured data
	if phase == PhaseSummary && m.summaryData != nil {
		return m.renderSummaryContent(height)
	}

	// For Run phase, check if we're showing a specific module's buffer
	var buffer *RingBuffer
	if phase == PhaseRun && m.activeTab != "" {
		// Show selected module's buffer
		if moduleBuffer := m.GetActiveModuleBuffer(); moduleBuffer != nil {
			buffer = moduleBuffer
		} else {
			buffer = pane.Buffer // Fallback to pane buffer
		}
	} else {
		buffer = pane.Buffer
	}

	// Update max scroll based on the ACTIVE buffer (not always pane buffer)
	pane.UpdateMaxScrollForBuffer(buffer, height)

	// Get lines based on scroll offset
	var lines []Line
	if pane.scrollOffset == 0 || pane.autoScroll {
		// At bottom or auto-scrolling: show most recent lines
		lines = buffer.Last(height)
		pane.scrollOffset = 0 // Ensure it stays at 0
	} else {
		// Scrolled up: show lines at offset
		lines = buffer.GetRange(pane.scrollOffset, height)
	}

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

// renderPaneFooter renders the bottom border for a pane
func (m Model) renderPaneFooter(phase Phase, height int) string {
	pane := m.panes[phase]

	// If not scrolled, show simple footer
	if pane.scrollOffset == 0 {
		borderLen := m.width - 2
		if borderLen < 3 {
			borderLen = 3
		}
		return "└" + strings.Repeat("─", borderLen) + "┘"
	}

	// Scrolled: show scroll indicator
	totalLines := pane.Buffer.Count()

	// Calculate scroll percentage (inverted: 100% at top, 0% at bottom)
	scrollPercent := 0
	if pane.maxScroll > 0 {
		scrollPercent = (pane.scrollOffset * 100) / pane.maxScroll
	}

	// Position in buffer: which line range we're viewing
	viewStart := totalLines - pane.scrollOffset - height
	if viewStart < 0 {
		viewStart = 0
	}
	viewEnd := viewStart + height
	if viewEnd > totalLines {
		viewEnd = totalLines
	}

	// Build scroll indicator: "↑ 45% [250-260/500]"
	indicator := fmt.Sprintf(" ↑ %d%% [%d-%d/%d] ", scrollPercent, viewStart+1, viewEnd, totalLines)

	// Calculate remaining border length
	borderLen := m.width - lipgloss.Width(indicator) - 2
	if borderLen < 3 {
		borderLen = 3
	}

	return "└" + Styles.Dim.Render(indicator) + strings.Repeat("─", borderLen) + "┘"
}

// renderResults renders the results section (rolling output after Run pane)
func (m Model) renderResults() string {
	lines := m.resultsBuffer.Last(20) // Show last 20 result lines
	if len(lines) == 0 {
		return ""
	}

	var b strings.Builder
	for _, line := range lines {
		b.WriteString(m.renderLine(line))
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) renderLine(line Line) string {
	// Strip newlines to prevent multi-line rendering (keeps pane height fixed)
	text := strings.ReplaceAll(line.Text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", "")

	// Truncate to fit width
	maxLen := m.width - 4 // Account for prefix
	if maxLen < 10 {
		maxLen = 10
	}
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

// formatElapsed formats a duration for display.
func formatElapsed(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}

// renderSummaryContent renders the Summary pane's structured content
func (m Model) renderSummaryContent(height int) string {
	var b strings.Builder
	data := m.summaryData

	// Build content lines
	var contentLines []string

	// Line 1: Primary status with timing
	icon := "✓"
	statusText := "Complete"
	if !data.Success {
		icon = "✗"
		statusText = "Failed"
	}
	primaryStatus := fmt.Sprintf("%s %s (%s)", icon, statusText, formatElapsed(data.TotalTime))
	contentLines = append(contentLines, primaryStatus)
	contentLines = append(contentLines, "") // Blank line

	// Lines 3-4: Phase summaries
	if data.InitSummary != "" {
		initIcon := m.panes[PhaseInit].Status.Icon()
		contentLines = append(contentLines, fmt.Sprintf("%s %s: %s", initIcon, PhaseNameInitialization, data.InitSummary))
	}
	if data.RunSummary != "" {
		runIcon := m.panes[PhaseRun].Status.Icon()
		runName := m.runPhaseName
		if runName == "" {
			runName = "Run"
		}
		contentLines = append(contentLines, fmt.Sprintf("%s %s: %s", runIcon, runName, data.RunSummary))
	}

	// Add blank line before details
	if len(data.Details) > 0 {
		contentLines = append(contentLines, "")
	}

	// Add detail lines (artifacts, errors, stats)
	contentLines = append(contentLines, data.Details...)

	// Add blank line before next steps
	if data.NextSteps != "" && len(contentLines) > 0 {
		contentLines = append(contentLines, "")
	}

	// Last line: Next steps
	if data.NextSteps != "" {
		contentLines = append(contentLines, data.NextSteps)
	}

	// Render lines into pane format (only render actual content, no trailing blanks)
	numLines := len(contentLines)
	if numLines > height {
		numLines = height
	}
	for i := 0; i < numLines; i++ {
		// Render as info-level line with proper styling
		text := contentLines[i]
		// Truncate to fit width
		maxLen := m.width - 4 // Account for prefix
		if maxLen < 10 {
			maxLen = 10
		}
		if len(text) > maxLen {
			text = text[:maxLen-1] + "…"
		}

		prefix := Styles.InfoPrefix.Render("│" + IconInfo)
		styled := Styles.Info.Render(text)
		b.WriteString(prefix + " " + styled)

		if i < numLines-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
// renderPaneHeaderPlain renders a pane's header line in plain text (no ANSI styling)
func (m Model) renderPaneHeaderPlain(phase Phase) string {
	pane := m.panes[phase]

	// Build icon and name
	icon := ""
	switch pane.Status {
	case PhaseActive:
		icon = "▶"
	case PhaseComplete:
		icon = "✓"
	case PhaseFailed:
		icon = "✗"
	default:
		icon = "·"
	}

	name := phase.String()
	if phase == PhaseRun && m.runPhaseName != "" {
		name = m.runPhaseName
	}

	// Build header content
	left := icon + " " + name

	// Add status info for Run phase when active (running modules shown via tabs)
	if phase == PhaseRun && pane.Status == PhaseActive {
		elapsed := time.Since(m.startTime).Round(time.Millisecond * 100)
		left = fmt.Sprintf("%s %s │ %d/%d", left, formatElapsed(elapsed), m.completed, m.total)
	}

	// Add summary for completed phases
	if (pane.Status == PhaseComplete || pane.Status == PhaseFailed) && pane.Summary != "" {
		left += ": " + pane.Summary
	}

	// Add "waiting..." for pending phases
	if pane.Status == PhasePending {
		left += ": waiting..."
	}

	// Border line
	borderLen := m.width - len(left) - 2
	if borderLen < 3 {
		borderLen = 3
	}
	border := strings.Repeat("─", borderLen)

	return "┌" + left + " " + border + "┐"
}

// renderPaneContentPlain renders the content lines for a pane in plain text (no ANSI styling)
func (m Model) renderPaneContentPlain(phase Phase, height int) string {
	var b strings.Builder
	pane := m.panes[phase]

	// Special handling for Summary pane: render structured data
	if phase == PhaseSummary && m.summaryData != nil {
		return m.renderSummaryContentPlain(height)
	}

	// Get last lines from buffer
	lines := pane.Buffer.Last(height)

	for i := 0; i < height; i++ {
		if i < len(lines) {
			// Render line without styling, strip newlines to prevent multi-line rendering
			text := strings.ReplaceAll(lines[i].Text, "\n", " ")
			text = strings.ReplaceAll(text, "\r", "")
			maxLen := m.width - 4
			if maxLen < 10 {
				maxLen = 10
			}
			if len(text) > maxLen {
				text = text[:maxLen-1] + "…"
			}
			b.WriteString("│  " + text)
		} else {
			b.WriteString("│ ")
		}
		if i < height-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderPaneFooterPlain renders the bottom border for a pane in plain text (no ANSI styling)
func (m Model) renderPaneFooterPlain(phase Phase) string {
	borderLen := m.width - 2
	if borderLen < 3 {
		borderLen = 3
	}
	return "└" + strings.Repeat("─", borderLen) + "┘"
}

// renderSummaryContentPlain renders the Summary pane's structured content in plain text
func (m Model) renderSummaryContentPlain(height int) string {
	var b strings.Builder
	data := m.summaryData

	// Build content lines
	var contentLines []string

	// Line 1: Primary status with timing
	icon := "✓"
	statusText := "Complete"
	if !data.Success {
		icon = "✗"
		statusText = "Failed"
	}
	primaryStatus := fmt.Sprintf("%s %s (%s)", icon, statusText, formatElapsed(data.TotalTime))
	contentLines = append(contentLines, primaryStatus)
	contentLines = append(contentLines, "") // Blank line

	// Lines 3-4: Phase summaries
	if data.InitSummary != "" {
		initIcon := "✓"
		if m.panes[PhaseInit].Status == PhaseFailed {
			initIcon = "✗"
		}
		contentLines = append(contentLines, fmt.Sprintf("%s %s: %s", initIcon, PhaseNameInitialization, data.InitSummary))
	}
	if data.RunSummary != "" {
		runIcon := "✓"
		if m.panes[PhaseRun].Status == PhaseFailed {
			runIcon = "✗"
		}
		runName := m.runPhaseName
		if runName == "" {
			runName = "Run"
		}
		contentLines = append(contentLines, fmt.Sprintf("%s %s: %s", runIcon, runName, data.RunSummary))
	}

	// Add blank line before details
	if len(data.Details) > 0 {
		contentLines = append(contentLines, "")
	}

	// Add detail lines
	contentLines = append(contentLines, data.Details...)

	// Add blank line before next steps
	if data.NextSteps != "" && len(contentLines) > 0 {
		contentLines = append(contentLines, "")
	}

	// Last line: Next steps
	if data.NextSteps != "" {
		contentLines = append(contentLines, data.NextSteps)
	}

	// Render lines into pane format (only render actual content, no trailing blanks)
	numLines := len(contentLines)
	if numLines > height {
		numLines = height
	}
	for i := 0; i < numLines; i++ {
		text := contentLines[i]
		maxLen := m.width - 4
		if maxLen < 10 {
			maxLen = 10
		}
		if len(text) > maxLen {
			text = text[:maxLen-1] + "…"
		}
		b.WriteString("│  " + text)

		if i < numLines-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderPaneContentPlainExpanded renders buffer content for final display.
// Unlike renderPaneContentPlain, this only renders actual lines (no blank padding).
func (m Model) renderPaneContentPlainExpanded(phase Phase, maxHeight int) string {
	var b strings.Builder
	pane := m.panes[phase]

	// Get all lines from buffer, capped at maxHeight
	allLines := pane.Buffer.All()
	lines := allLines
	if len(lines) > maxHeight {
		lines = lines[len(lines)-maxHeight:]
	}

	for i, line := range lines {
		// Render line without styling, strip newlines to prevent multi-line rendering
		text := strings.ReplaceAll(line.Text, "\n", " ")
		text = strings.ReplaceAll(text, "\r", "")
		maxLen := m.width - 4
		if maxLen < 10 {
			maxLen = 10
		}
		if len(text) > maxLen {
			text = text[:maxLen-1] + "…"
		}
		b.WriteString("│  " + text)
		if i < len(lines)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderSummaryContentPlainExpanded renders all Summary content without height limit.
func (m Model) renderSummaryContentPlainExpanded() string {
	var b strings.Builder
	data := m.summaryData

	// Build content lines
	var contentLines []string

	// Line 1: Primary status with timing
	icon := "✓"
	statusText := "Complete"
	if !data.Success {
		icon = "✗"
		statusText = "Failed"
	}
	primaryStatus := fmt.Sprintf("%s %s (%s)", icon, statusText, formatElapsed(data.TotalTime))
	contentLines = append(contentLines, primaryStatus)
	contentLines = append(contentLines, "") // Blank line

	// Lines 3-4: Phase summaries
	if data.InitSummary != "" {
		initIcon := "✓"
		if m.panes[PhaseInit].Status == PhaseFailed {
			initIcon = "✗"
		}
		contentLines = append(contentLines, fmt.Sprintf("%s %s: %s", initIcon, PhaseNameInitialization, data.InitSummary))
	}
	if data.RunSummary != "" {
		runIcon := "✓"
		if m.panes[PhaseRun].Status == PhaseFailed {
			runIcon = "✗"
		}
		runName := m.runPhaseName
		if runName == "" {
			runName = "Run"
		}
		contentLines = append(contentLines, fmt.Sprintf("%s %s: %s", runIcon, runName, data.RunSummary))
	}

	// Add blank line before details
	if len(data.Details) > 0 {
		contentLines = append(contentLines, "")
	}

	// Add all detail lines (no truncation)
	contentLines = append(contentLines, data.Details...)

	// Add blank line before next steps
	if data.NextSteps != "" && len(contentLines) > 0 {
		contentLines = append(contentLines, "")
	}

	// Last line: Next steps
	if data.NextSteps != "" {
		contentLines = append(contentLines, data.NextSteps)
	}

	// Render all lines (no height limit)
	for i, text := range contentLines {
		maxLen := m.width - 4
		if maxLen < 10 {
			maxLen = 10
		}
		if len(text) > maxLen {
			text = text[:maxLen-1] + "…"
		}
		b.WriteString("│  " + text)
		if i < len(contentLines)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
