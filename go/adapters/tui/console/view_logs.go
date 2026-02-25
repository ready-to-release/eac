package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/ready-to-release/eac/go/adapters/tui/console/render"
)

// renderLogsPanel renders the logs panel for side-by-side layout.
func (m Model) renderLogsPanel(width, height int) string {
	var b strings.Builder
	pane := m.Execution.Panes[PhaseRun]

	// Check if active module is cached - show special header
	activeModule := m.getEffectiveActiveTab()
	var isCachedModule bool
	if activeModule != "" {
		if state, exists := m.Execution.UoWStates[activeModule]; exists && state.Status == UoWSkipped {
			isCachedModule = true
		}
	}

	var left string
	if isCachedModule {
		// Special header for cached modules
		cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
		icon := "⏭"
		if m.Display.AsciiMode {
			icon = "="
		}
		left = cyanStyle.Render(icon) + " " + cyanStyle.Render("Cached") + ": " + cyanStyle.Bold(true).Render(activeModule)
	} else {
		// Try Selected content as header (UoW details when tab is hovered)
		headerContentWidth := width - 4
		if headerContentWidth < 10 {
			headerContentWidth = 10
		}
		selectedHeader := m.renderSelectedHeader(headerContentWidth)
		if strings.TrimSpace(selectedHeader) != "" {
			left = selectedHeader
		} else {
			// Fallback: phase-based header
			icon := m.phaseIcon(pane.Status)
			name := m.Display.RunPhaseName
			if name == "" {
				name = "Run"
			}

			var iconStyle, nameStyle lipgloss.Style
			switch pane.Status {
			case PhaseActive:
				iconStyle = Styles.Running
				nameStyle = Styles.Phase
			case PhaseComplete:
				iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("71"))
				nameStyle = Styles.Dim
			case PhaseFailed:
				iconStyle = Styles.Error
				nameStyle = Styles.Error
			default:
				iconStyle = Styles.Dim
				nameStyle = Styles.Dim
			}

			left = iconStyle.Render(icon) + " " + nameStyle.Render(name)

			if pane.Status == PhaseActive {
				var moduleElapsed time.Duration
				if activeModule != "" {
					left += ": " + Styles.Phase.Render(activeModule)
					if state, exists := m.Execution.UoWStates[activeModule]; exists {
						if state.Status == UoWRunning {
							moduleElapsed = time.Since(state.StartTime).Round(time.Millisecond * 100)
						} else if state.Status == UoWComplete || state.Status == UoWSkipped || state.Status == UoWFailed {
							moduleElapsed = state.EndTime.Sub(state.StartTime).Round(time.Millisecond * 100)
						}
					}
				}

				left = fmt.Sprintf("%s %s %d/%d",
					left,
					Styles.Time.Render(formatElapsed(moduleElapsed)),
					m.Execution.Completed,
					m.Execution.Total,
				)
			}

			if (pane.Status == PhaseComplete || pane.Status == PhaseFailed) && pane.Summary != "" {
				left += ": " + Styles.Dim.Render(pane.Summary)
			}
		}
	}

	// Header line (borderless)
	headerWidth := width - lipgloss.Width(left)
	if headerWidth < 0 {
		headerWidth = 0
	}
	b.WriteString(left + strings.Repeat(" ", headerWidth) + "\n")

	// Content area (1 line for header, no footer)
	contentHeight := height - 1
	if contentHeight < 1 {
		contentHeight = 1
	}

	// During init phase: show init buffer lines + animated dots
	if pane.Status == PhasePending {
		innerWidth := width - 2
		if innerWidth < 1 {
			innerWidth = 1
		}

		initLines := m.Execution.Panes[PhaseInit].Buffer.Last(contentHeight)
		lineCount := 0

		for _, line := range initLines {
			icon := m.phaseIcon(PhaseComplete)
			text := render.PadOrTruncate(line.Text, innerWidth-3)
			b.WriteString(icon + " " + text + "\n")
			lineCount++
		}

		// Animated "Initializing..." dots
		if lineCount < contentHeight {
			animLine := m.renderInitAnimatedStatus(innerWidth)
			b.WriteString(" " + animLine + "\n")
			lineCount++
		}

		// Fill remaining with empty lines
		for lineCount < contentHeight {
			b.WriteString(strings.Repeat(" ", width) + "\n")
			lineCount++
		}

		return b.String()
	}

	// Render special content for cached modules
	if isCachedModule {
		if state, exists := m.Execution.UoWStates[activeModule]; exists {
			m.renderCachedContent(&b, activeModule, state, width, contentHeight)
			return b.String()
		}
	}

	var buffer *RingBuffer
	if activeModule != "" {
		if moduleBuffer := m.GetActiveUoWBuffer(); moduleBuffer != nil {
			buffer = moduleBuffer
		} else {
			buffer = pane.Buffer
		}
	} else {
		buffer = pane.Buffer
	}

	pane.UpdateMaxScrollForBuffer(buffer, contentHeight)

	var lines []Line
	if pane.scrollOffset == 0 || pane.autoScroll {
		lines = buffer.Last(contentHeight)
		pane.scrollOffset = 0
	} else {
		lines = buffer.GetRange(pane.scrollOffset, contentHeight)
	}

	for i := 0; i < contentHeight; i++ {
		if i < len(lines) {
			lineContent := m.renderLogLine(lines[i], 0, i, true)
			b.WriteString(lineContent + "\n")
		} else {
			b.WriteString(strings.Repeat(" ", width) + "\n")
		}
	}

	// Scroll indicator (inline, no border chrome)
	if pane.scrollOffset > 0 {
		totalLines := buffer.Count()
		scrollPercent := 0
		if pane.maxScroll > 0 {
			scrollPercent = (pane.scrollOffset * 100) / pane.maxScroll
		}
		viewStart := totalLines - pane.scrollOffset - contentHeight
		if viewStart < 0 {
			viewStart = 0
		}
		viewEnd := viewStart + contentHeight
		if viewEnd > totalLines {
			viewEnd = totalLines
		}
		indicator := fmt.Sprintf(" ↑ %d%% [%d-%d/%d]", scrollPercent, viewStart+1, viewEnd, totalLines)
		b.WriteString(" " + Styles.Dim.Render(indicator))
	}

	return b.String()
}

// renderLogLine renders a single log line for the logs panel.
// lineIndex is the 0-based index within the visible content area (for selection highlighting).
// maxWidth <= 0 means no truncation (open-right mode).
// borderless removes the left │ border from the prefix.
func (m Model) renderLogLine(line Line, maxWidth int, lineIndex int, borderless bool) string {
	text := strings.ReplaceAll(line.Text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", "")

	if maxWidth > 0 && len(text) > maxWidth {
		if m.Display.AsciiMode {
			text = text[:maxWidth-2] + ".."
		} else {
			text = text[:maxWidth-1] + "…"
		}
	}

	iconFail, iconWarn, iconInfo := m.lineIcons()

	// Check if this line is within selection range
	isSelected := m.isLineSelected(lineIndex)

	border := "│"
	if borderless {
		border = ""
	}

	var prefix string
	var textStyle lipgloss.Style
	switch line.Level {
	case LevelError:
		prefix = Styles.ErrorPrefix.Render(border + iconFail)
		textStyle = Styles.Error
	case LevelWarn:
		prefix = Styles.WarnPrefix.Render(border + iconWarn)
		textStyle = Styles.Warn
	default:
		prefix = Styles.InfoPrefix.Render(border + iconInfo)
		textStyle = Styles.Info
	}

	if isSelected {
		textStyle = lipgloss.NewStyle().Reverse(true)
	}

	return prefix + " " + textStyle.Render(text)
}

// isLineSelected returns true if the given content line index is within the selection range.
func (m Model) isLineSelected(lineIndex int) bool {
	if !m.Resources.Selection.Active {
		return false
	}

	startLine, endLine := m.Resources.Selection.StartLine, m.Resources.Selection.EndLine
	if startLine > endLine {
		startLine, endLine = endLine, startLine
	}

	return lineIndex >= startLine && lineIndex <= endLine
}

// renderLogsPaneHeadless renders the logs panel without any borders.
// Used below the detail pane where the header info is redundant.
// Logs flow freely in the available space — no header, footer, or side borders.
func (m Model) renderLogsPaneHeadless(width, height int) string {
	var b strings.Builder
	pane := m.Execution.Panes[PhaseRun]

	activeModule := m.getEffectiveActiveTab()

	// Full height for content — no borders to subtract
	contentHeight := height
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Cached module special content
	if activeModule != "" {
		if state, exists := m.Execution.UoWStates[activeModule]; exists && state.Status == UoWSkipped {
			m.renderCachedContent(&b, activeModule, state, width, contentHeight)
			return b.String()
		}
	}

	var buffer *RingBuffer
	if activeModule != "" {
		if moduleBuffer := m.GetActiveUoWBuffer(); moduleBuffer != nil {
			buffer = moduleBuffer
		} else {
			buffer = pane.Buffer
		}
	} else {
		buffer = pane.Buffer
	}

	pane.UpdateMaxScrollForBuffer(buffer, contentHeight)

	var lines []Line
	if pane.scrollOffset == 0 || pane.autoScroll {
		lines = buffer.Last(contentHeight)
		pane.scrollOffset = 0
	} else {
		lines = buffer.GetRange(pane.scrollOffset, contentHeight)
	}

	for i := 0; i < contentHeight; i++ {
		if i < len(lines) {
			b.WriteString(m.renderLogLine(lines[i], 0, i, true)) // borderless, no truncation
		}
		if i < contentHeight-1 {
			b.WriteString("\n")
		}
	}

	// Scroll indicator (inline, no border chrome)
	if pane.scrollOffset > 0 {
		totalLines := buffer.Count()
		scrollPercent := 0
		if pane.maxScroll > 0 {
			scrollPercent = (pane.scrollOffset * 100) / pane.maxScroll
		}
		viewStart := totalLines - pane.scrollOffset - contentHeight
		if viewStart < 0 {
			viewStart = 0
		}
		viewEnd := viewStart + contentHeight
		if viewEnd > totalLines {
			viewEnd = totalLines
		}
		indicator := fmt.Sprintf(" ↑ %d%% [%d-%d/%d]", scrollPercent, viewStart+1, viewEnd, totalLines)
		b.WriteString(" " + Styles.Dim.Render(indicator))
	}

	return b.String()
}

// renderCachedContent renders special content for cached/skipped modules.
func (m Model) renderCachedContent(b *strings.Builder, moniker string, state *UoWState, width, contentHeight int) {
	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	dimStyle := Styles.Dim

	// Full width, no borders
	contentWidth := width

	// Helper to render an empty line
	emptyLine := func() {
		b.WriteString(strings.Repeat(" ", contentWidth) + "\n")
	}

	// Helper to render a centered line
	centerLine := func(text string, style lipgloss.Style) {
		textLen := lipgloss.Width(text)
		if textLen >= contentWidth {
			text = text[:contentWidth-1] + "…"
			textLen = contentWidth
		}
		padding := (contentWidth - textLen) / 2
		leftPad := strings.Repeat(" ", padding)
		rightPad := strings.Repeat(" ", contentWidth-textLen-padding)
		b.WriteString(leftPad + style.Render(text) + rightPad + "\n")
	}

	// Helper to render a left-aligned line with prefix
	prefixLine := func(prefix, text string, prefixStyle, textStyle lipgloss.Style) {
		renderedPrefix := prefixStyle.Render(prefix)
		renderedText := textStyle.Render(text)
		totalLen := lipgloss.Width(renderedPrefix) + lipgloss.Width(renderedText)
		padding := contentWidth - totalLen
		if padding < 0 {
			padding = 0
		}
		b.WriteString(renderedPrefix + renderedText + strings.Repeat(" ", padding) + "\n")
	}

	lineCount := 0

	// Empty lines at top for visual balance
	for i := 0; i < 2 && lineCount < contentHeight; i++ {
		emptyLine()
		lineCount++
	}

	// Cache icon and status
	if lineCount < contentHeight {
		icon := "⏭"
		if m.Display.AsciiMode {
			icon = "="
		}
		centerLine(icon+"  CACHED", cyanStyle.Bold(true))
		lineCount++
	}

	// Empty line
	if lineCount < contentHeight {
		emptyLine()
		lineCount++
	}

	// Module name
	if lineCount < contentHeight {
		centerLine(moniker, cyanStyle)
		lineCount++
	}

	// Empty lines
	for i := 0; i < 2 && lineCount < contentHeight; i++ {
		emptyLine()
		lineCount++
	}

	// Separator
	if lineCount < contentHeight {
		sepLen := contentWidth - 8
		if sepLen < 4 {
			sepLen = 4
		}
		sep := strings.Repeat("─", sepLen/2)
		centerLine(sep+" ◆ "+sep, dimStyle)
		lineCount++
	}

	// Empty line
	if lineCount < contentHeight {
		emptyLine()
		lineCount++
	}

	// Explanation text
	explanations := []string{
		"This component was skipped because",
		"no changes were detected since the",
		"last successful build.",
		"",
		"The cached result from a previous",
		"run is being used instead.",
	}

	for _, line := range explanations {
		if lineCount >= contentHeight {
			break
		}
		if line == "" {
			emptyLine()
		} else {
			prefixLine("    ", line, dimStyle, dimStyle)
		}
		lineCount++
	}

	// Empty line
	if lineCount < contentHeight {
		emptyLine()
		lineCount++
	}

	// Cache info: when the artifact was built
	if !state.CacheTime.IsZero() && lineCount < contentHeight {
		builtTime := state.CacheTime.Format("2006-01-02 15:04:05")
		prefixLine("    Last built: ", builtTime, dimStyle, cyanStyle)
		lineCount++
	}

	// Log path link if available
	if state.LogPath != "" && lineCount < contentHeight {
		emptyLine()
		lineCount++
		if lineCount < contentHeight {
			prefixLine("    Build log: ", state.LogPath, dimStyle, cyanStyle)
			lineCount++
		}
	}

	// Fill remaining space with empty lines
	for lineCount < contentHeight {
		emptyLine()
		lineCount++
	}
}
