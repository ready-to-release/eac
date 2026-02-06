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
	pane := m.panes[PhaseRun]

	// Check if active module is cached - show special header
	activeModule := m.getEffectiveActiveTab()
	var isCachedModule bool
	if activeModule != "" {
		if state, exists := m.uowStates[activeModule]; exists && state.Status == UoWSkipped {
			isCachedModule = true
		}
	}

	var left string
	if isCachedModule {
		// Special header for cached modules
		cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
		icon := "⏭"
		if m.asciiMode {
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
			name := m.runPhaseName
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
					if state, exists := m.uowStates[activeModule]; exists {
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
					m.completed,
					m.total,
				)
			}

			if (pane.Status == PhaseComplete || pane.Status == PhaseFailed) && pane.Summary != "" {
				left += ": " + Styles.Dim.Render(pane.Summary)
			}
		}
	}

	borderLen := width - lipgloss.Width(left) - 2
	if borderLen < 1 {
		borderLen = 1
	}
	b.WriteString("┌" + left + " " + Styles.Border.Render(strings.Repeat("─", borderLen)) + "┐\n")

	// Content area
	contentHeight := height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	// During init phase: show init buffer lines + animated dots
	if pane.Status == PhasePending {
		innerWidth := width - 4
		if innerWidth < 1 {
			innerWidth = 1
		}

		initLines := m.panes[PhaseInit].Buffer.Last(contentHeight)
		lineCount := 0

		for _, line := range initLines {
			icon := m.phaseIcon(PhaseComplete)
			text := render.PadOrTruncate(line.Text, innerWidth-3)
			b.WriteString(Styles.Border.Render("│") + icon + " " + text + " " + Styles.Border.Render("│") + "\n")
			lineCount++
		}

		// Animated "Initializing..." dots
		if lineCount < contentHeight {
			animLine := m.renderInitAnimatedStatus(innerWidth)
			b.WriteString(Styles.Border.Render("│") + " " + animLine + " " + Styles.Border.Render("│") + "\n")
			lineCount++
		}

		// Fill remaining with empty lines
		for lineCount < contentHeight {
			b.WriteString(Styles.Border.Render("│") + " " + strings.Repeat(" ", innerWidth) + " " + Styles.Border.Render("│") + "\n")
			lineCount++
		}

		// Footer
		footerBorderLen := width - 2
		if footerBorderLen < 1 {
			footerBorderLen = 1
		}
		b.WriteString("└" + strings.Repeat("─", footerBorderLen) + "┘")
		return b.String()
	}

	// Render special content for cached modules
	if isCachedModule {
		if state, exists := m.uowStates[activeModule]; exists {
			m.renderCachedContent(&b, activeModule, state, width, contentHeight)

			// Footer
			footerBorderLen := width - 2
			if footerBorderLen < 1 {
				footerBorderLen = 1
			}
			b.WriteString("└" + strings.Repeat("─", footerBorderLen) + "┘")
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
			lineContent := m.renderLogLine(lines[i], width-4, i)
			b.WriteString(lineContent + "\n")
		} else {
			b.WriteString(Styles.Dim.Render("│") + " " + strings.Repeat(" ", width-4) + Styles.Dim.Render("│") + "\n")
		}
	}

	// Footer with scroll indicator
	if pane.scrollOffset == 0 {
		footerBorderLen := width - 2
		if footerBorderLen < 1 {
			footerBorderLen = 1
		}
		b.WriteString("└" + strings.Repeat("─", footerBorderLen) + "┘")
	} else {
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
		indicator := fmt.Sprintf(" ↑ %d%% [%d-%d/%d] ", scrollPercent, viewStart+1, viewEnd, totalLines)
		borderLen := width - lipgloss.Width(indicator) - 2
		if borderLen < 1 {
			borderLen = 1
		}
		b.WriteString("└" + Styles.Dim.Render(indicator) + strings.Repeat("─", borderLen) + "┘")
	}

	return b.String()
}

// renderLogLine renders a single log line for the logs panel.
// lineIndex is the 0-based index within the visible content area (for selection highlighting).
func (m Model) renderLogLine(line Line, maxWidth int, lineIndex int) string {
	text := strings.ReplaceAll(line.Text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", "")

	if len(text) > maxWidth {
		if m.asciiMode {
			text = text[:maxWidth-2] + ".."
		} else {
			text = text[:maxWidth-1] + "…"
		}
	}

	iconFail, iconWarn, iconInfo := m.lineIcons()

	// Check if this line is within selection range
	isSelected := m.isLineSelected(lineIndex)

	var prefix, styled string
	switch line.Level {
	case LevelError:
		prefix = Styles.ErrorPrefix.Render("│" + iconFail)
		if isSelected {
			styled = lipgloss.NewStyle().Reverse(true).Render(text)
		} else {
			styled = Styles.Error.Render(text)
		}
	case LevelWarn:
		prefix = Styles.WarnPrefix.Render("│" + iconWarn)
		if isSelected {
			styled = lipgloss.NewStyle().Reverse(true).Render(text)
		} else {
			styled = Styles.Warn.Render(text)
		}
	default:
		prefix = Styles.InfoPrefix.Render("│" + iconInfo)
		if isSelected {
			styled = lipgloss.NewStyle().Reverse(true).Render(text)
		} else {
			styled = Styles.Info.Render(text)
		}
	}

	return prefix + " " + styled
}

// isLineSelected returns true if the given content line index is within the selection range.
func (m Model) isLineSelected(lineIndex int) bool {
	if !m.selection.Active {
		return false
	}

	startLine, endLine := m.selection.StartLine, m.selection.EndLine
	if startLine > endLine {
		startLine, endLine = endLine, startLine
	}

	return lineIndex >= startLine && lineIndex <= endLine
}

// renderCachedContent renders special content for cached/skipped modules.
func (m Model) renderCachedContent(b *strings.Builder, moniker string, state *UoWState, width, contentHeight int) {
	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	dimStyle := Styles.Dim

	// Calculate content width (accounting for borders)
	contentWidth := width - 4

	// Helper to render an empty line
	emptyLine := func() {
		b.WriteString(dimStyle.Render("│") + " " + strings.Repeat(" ", contentWidth) + dimStyle.Render("│") + "\n")
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
		b.WriteString(dimStyle.Render("│") + " " + leftPad + style.Render(text) + rightPad + dimStyle.Render("│") + "\n")
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
		b.WriteString(dimStyle.Render("│") + " " + renderedPrefix + renderedText + strings.Repeat(" ", padding) + dimStyle.Render("│") + "\n")
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
		if m.asciiMode {
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
