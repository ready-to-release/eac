package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// View renders the console window.
func (m Model) View() string {
	// When quitting in alt screen mode, return empty (screen will be restored)
	// Plain-text summary is printed after program.Run() completes
	if m.quitting {
		return ""
	}
	// Wrap with zone.Scan to enable mouse click detection on marked zones
	return zone.Scan(m.viewPanes())
}

// ViewFinal renders a clean summary after TUI exit.
// Shows Init pane and summary data (which has authoritative results from orchestrator).
func (m Model) ViewFinal() string {
	var b strings.Builder

	// Render Init pane (frozen state)
	b.WriteString(m.renderPaneHeaderPlain(PhaseInit))
	b.WriteString("\n")
	b.WriteString(m.renderPaneContentPlainExpanded(PhaseInit, 10))
	b.WriteString("\n")
	b.WriteString(m.renderPaneFooterPlain(PhaseInit))
	b.WriteString("\n")

	// Render summary data if available (this has authoritative success/failure info)
	if m.summaryData != nil {
		// Details first (includes module table from summary.go)
		if len(m.summaryData.Details) > 0 {
			b.WriteString("\n")
			for _, line := range m.summaryData.Details {
				// Strip markdown table pipes for cleaner terminal output
				cleanLine := stripMarkdownPipes(line)
				b.WriteString(fmt.Sprintf("%s\n", cleanLine))
			}
		}

		// Summary status at the end
		icon := "✓"
		statusText := "PASSED"
		if !m.summaryData.Success {
			icon = "✗"
			statusText = "FAILED"
		}
		b.WriteString(fmt.Sprintf("\n%s %s (%s)\n", icon, statusText, formatElapsed(m.summaryData.TotalTime)))

		// Run phase summary
		if m.summaryData.RunSummary != "" {
			b.WriteString(fmt.Sprintf("  %s\n", m.summaryData.RunSummary))
		}
	} else if m.panes[PhaseRun].Status != PhasePending {
		// Fallback: use module states if no summary data (shouldn't normally happen)
		tabs := m.GetVisibleTabs()
		if len(tabs) > 0 {
			b.WriteString("\n")
			b.WriteString(m.renderModuleResultsTable(tabs))
		}
	}

	return b.String()
}

// stripMarkdownPipes removes leading/trailing pipe characters from markdown table lines.
// Converts "| col1 | col2 |" to "col1   col2" for cleaner terminal display.
func stripMarkdownPipes(line string) string {
	// Handle separator lines (dashes)
	if strings.Contains(line, "---") {
		// Convert separator to plain dashes without pipes
		line = strings.TrimPrefix(line, "|")
		line = strings.TrimSuffix(line, "|")
		line = strings.ReplaceAll(line, "|", " ")
		return strings.TrimSpace(line)
	}

	// Handle data lines
	if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
		// Remove outer pipes
		line = strings.TrimPrefix(line, "|")
		line = strings.TrimSuffix(line, "|")
		// Replace inner pipes with spaces for alignment
		line = strings.ReplaceAll(line, " | ", "   ")
		return strings.TrimSpace(line)
	}

	return line
}

// viewPanes renders the layout with panes appearing progressively.
// When Run phase is active, Components (tabs/tree) and Run (logs) are side-by-side.
func (m Model) viewPanes() string {
	var b strings.Builder

	// Track lines used
	usedLines := 0

	// Render Init pane (always visible)
	if m.initSummary != nil {
		// Compact single line when initialization is complete
		b.WriteString(m.renderInitPaneCompact())
		b.WriteString("\n")
		usedLines += 1
	} else {
		// Loading state - show animated dots
		b.WriteString(m.renderInitPaneLoading())
		b.WriteString("\n")
		usedLines += 1
	}

	// Render Resources pane between Init and Run (shows locks and system metrics)
	if resourcesPane := m.renderResourcesPane(); resourcesPane != "" {
		b.WriteString(resourcesPane)
		b.WriteString("\n")
		usedLines += strings.Count(resourcesPane, "\n") + 1
	}

	// Render Run pane only if it actually started (not still pending)
	if m.panes[PhaseRun].Status != PhasePending {
		// Show lamp countdown row if user has interacted and countdown is running
		if m.userHasInteracted && m.exitCountdownSecs > 0 && !m.allRunnersCompleted.IsZero() {
			b.WriteString(m.renderExitLampRow())
			b.WriteString("\n")
			usedLines += 1
		}

		// Calculate remaining height for side-by-side panels
		// Reserve space for summary if it will be shown
		summaryLines := 0
		if m.summaryData != nil {
			summaryLines = 6 // header + 4 content + footer
		}

		// Fill remaining terminal height
		remainingHeight := m.height - usedLines - summaryLines
		if remainingHeight < 5 {
			remainingHeight = 5
		}

		// Only show waiting state in left pane when actively exiting:
		// exitRequested is set when all runners done AND (user never interacted OR user timer expired)
		showWaitingState := m.exitRequested && m.summaryData == nil

		tabs := m.GetVisibleTabs()
		// Side-by-side layout: Components (tabs/tree) on LEFT, Logs on RIGHT
		// Pass waiting state to show in left pane when waiting for user idle
		sideBySide := m.renderSideBySideLayoutWithState(tabs, remainingHeight, showWaitingState)
		b.WriteString(sideBySide)
		b.WriteString("\n")
	}

	// Render Summary pane (only when summary data is available)
	if m.summaryData != nil {
		summaryH := 4
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

// renderSideBySideLayout renders Components (left) and Logs (right) side by side.
func (m Model) renderSideBySideLayout(tabs []*ModuleState, height int) string {
	return m.renderSideBySideLayoutWithState(tabs, height, false)
}

// renderSideBySideLayoutWithState renders Components (left) and Logs (right) side by side.
// If showWaitingState is true, the left panel shows the current waiting state instead of tabs.
func (m Model) renderSideBySideLayoutWithState(tabs []*ModuleState, height int, showWaitingState bool) string {
	// Fixed width for components panel, logs take remaining space
	const componentsWidth = 62 // Fixed width for 3 columns of tabs
	logsWidth := m.width - componentsWidth - 1 // -1 for separator
	if logsWidth < 40 {
		logsWidth = 40
	}

	// Render left panel (Components - tabs or tree based on viewMode)
	// If waiting for summary, show waiting state instead
	var leftPanel string
	if showWaitingState {
		leftPanel = m.renderWaitingStatePanel(componentsWidth, height)
	} else if m.viewMode == ViewModeTree {
		leftPanel = m.renderTreePanel(tabs, componentsWidth, height)
	} else {
		leftPanel = m.renderTabGridPanel(tabs, componentsWidth, height)
	}

	// Render right panel (Logs)
	rightPanel := m.renderLogsPanel(logsWidth, height)

	// Join horizontally with separator
	leftLines := strings.Split(leftPanel, "\n")
	rightLines := strings.Split(rightPanel, "\n")

	// Ensure both have the same number of lines
	for len(leftLines) < height {
		leftLines = append(leftLines, strings.Repeat(" ", componentsWidth))
	}
	for len(rightLines) < height {
		rightLines = append(rightLines, strings.Repeat(" ", logsWidth))
	}

	// Build the combined view
	var b strings.Builder
	separator := Styles.Border.Render("│")

	for i := 0; i < height; i++ {
		left := leftLines[i]
		right := rightLines[i]

		// Pad left panel to exact width
		leftWidth := lipgloss.Width(left)
		if leftWidth < componentsWidth {
			left += strings.Repeat(" ", componentsWidth-leftWidth)
		}

		b.WriteString(left)
		b.WriteString(separator)
		b.WriteString(right)

		if i < height-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderTabGridPanel renders the tab grid as a panel for side-by-side layout.
func (m Model) renderTabGridPanel(tabs []*ModuleState, width, height int) string {
	var b strings.Builder

	// Header: ┌─ Components: 3 running, 5/10 done ─────┐
	var running, completed, skipped, failed int
	for _, tab := range tabs {
		switch tab.Status {
		case ModuleRunning:
			running++
		case ModuleComplete:
			completed++
		case ModuleSkipped:
			skipped++
		case ModuleFailed:
			failed++
		}
	}

	title := "Components"
	var statusParts []string
	if running > 0 {
		statusParts = append(statusParts, fmt.Sprintf("%d running", running))
	}
	if completed > 0 || skipped > 0 || failed > 0 {
		statusParts = append(statusParts, fmt.Sprintf("%d/%d done", completed+skipped+failed, len(tabs)))
	}
	if len(statusParts) > 0 {
		title += ": " + strings.Join(statusParts, ", ")
	}

	// Mode indicator
	modeIndicator := "[Tabs]"

	headerLeft := "┌─ " + Styles.Dim.Render(title) + " "
	headerRight := Styles.Dim.Render(modeIndicator) + " ─┐"
	headerBorderLen := width - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight)
	if headerBorderLen < 1 {
		headerBorderLen = 1
	}
	b.WriteString(headerLeft + Styles.Border.Render(strings.Repeat("─", headerBorderLen)) + headerRight + "\n")

	// Tab content area (height - 2 for header and footer)
	contentHeight := height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Render all tab content (for scrolling)
	tabContent := m.renderTabGridContent(tabs, width-2, contentHeight+m.tabsScrollOffset+50)
	tabLines := strings.Split(tabContent, "\n")

	// Apply scroll offset
	scrollOffset := m.tabsScrollOffset
	if scrollOffset > len(tabLines)-contentHeight {
		scrollOffset = len(tabLines) - contentHeight
		if scrollOffset < 0 {
			scrollOffset = 0
		}
	}

	for i := 0; i < contentHeight; i++ {
		lineIdx := i + scrollOffset
		if lineIdx < len(tabLines) {
			line := tabLines[lineIdx]
			lineWidth := lipgloss.Width(line)
			padding := width - 2 - lineWidth
			if padding < 0 {
				padding = 0
			}
			b.WriteString(Styles.Border.Render("│") + line + strings.Repeat(" ", padding) + Styles.Border.Render("│") + "\n")
		} else {
			b.WriteString(Styles.Border.Render("│") + strings.Repeat(" ", width-2) + Styles.Border.Render("│") + "\n")
		}
	}

	// Footer
	footerBorderLen := width - 2
	if footerBorderLen < 1 {
		footerBorderLen = 1
	}
	b.WriteString("└" + Styles.Border.Render(strings.Repeat("─", footerBorderLen)) + "┘")

	return b.String()
}

// renderInitPaneLoading renders a loading indicator while initialization is in progress.
func (m Model) renderInitPaneLoading() string {
	// Animated dots based on elapsed time
	elapsed := time.Since(m.startTime)
	dotCount := int(elapsed.Seconds()*2) % 4 // 0, 1, 2, 3 cycling

	var dots string
	if m.asciiMode {
		dots = strings.Repeat(".", dotCount+1)
	} else {
		dots = strings.Repeat("·", dotCount+1)
	}

	icon := "▶"
	if m.asciiMode {
		icon = ">"
	}

	// Check if we're waiting for any locks
	waitingMessage := m.getWaitingForLocksMessage()
	var statusText string
	if waitingMessage != "" {
		statusText = waitingMessage
	} else {
		statusText = "Initializing"
	}

	left := Styles.Running.Render(icon) + " " + Styles.Phase.Render(statusText) + Styles.Dim.Render(dots)

	// Single line with border
	borderLen := m.width - lipgloss.Width(left) - 4
	if borderLen < 1 {
		borderLen = 1
	}

	return "─" + left + " " + Styles.Border.Render(strings.Repeat("─", borderLen)) + "─"
}

// WaitingState describes what the TUI is currently waiting for.
type WaitingState int

const (
	WaitingNone           WaitingState = iota // Not waiting for anything
	WaitingRenderSummary                      // Waiting for summary to be rendered
	WaitingUserExitTimer                      // Waiting for user exit timer
	WaitingBothSummaryAndTimer                // Both (shouldn't happen normally)
)

// getWaitingState returns the current waiting state.
// This follows the three-thread model:
// - User timer: countdown from last interaction (informational)
// - Summary renderer: background builder producing pendingSummaryData
// - Exit decision: when all done AND (never interacted OR timer expired)
func (m Model) getWaitingState() WaitingState {
	hasSummary := m.summaryData != nil
	hasPendingSummary := m.pendingSummaryData != nil

	// exitRequested is set when we want to exit (all done AND user timer expired/never interacted)
	if m.exitRequested {
		if !hasPendingSummary && !hasSummary {
			// Waiting for summary builder to finish
			return WaitingRenderSummary
		}
		// Summary ready or shown - no waiting state (will quit immediately)
		return WaitingNone
	}

	// Not actively exiting yet - check if showing user timer countdown
	allDone := !m.allRunnersCompleted.IsZero()
	if allDone && m.userHasInteracted && m.exitCountdownSecs > 0 {
		// User has interacted and is counting down - show informational timer
		return WaitingUserExitTimer
	}

	return WaitingNone
}

// getWaitingStateMessage returns a human-readable message for the current waiting state.
func (m Model) getWaitingStateMessage() string {
	switch m.getWaitingState() {
	case WaitingRenderSummary:
		return "rendering summary"
	case WaitingUserExitTimer:
		// User has interacted - show countdown as informational (not blocking yet)
		return fmt.Sprintf("exit in %ds", m.exitCountdownSecs)
	case WaitingBothSummaryAndTimer:
		return fmt.Sprintf("rendering + waiting %ds", m.exitCountdownSecs)
	default:
		return ""
	}
}

// getWaitingForLocksMessage returns a message if any locks are being waited on.
// Returns empty string if no locks are waiting.
func (m Model) getWaitingForLocksMessage() string {
	var waitingLocks []string
	for _, lock := range m.locks {
		if lock.Waiting > 0 {
			// Extract just the identifier from "type:identifier" format
			name := lock.Name
			if idx := strings.Index(name, ":"); idx >= 0 {
				name = name[idx+1:]
			}
			waitingLocks = append(waitingLocks, name)
		}
	}

	if len(waitingLocks) == 0 {
		return ""
	}

	if len(waitingLocks) == 1 {
		return fmt.Sprintf("Waiting for lock: %s", waitingLocks[0])
	}
	return fmt.Sprintf("Waiting for locks: %s", strings.Join(waitingLocks, ", "))
}

// renderWaitingStatePanel renders the waiting state in the left panel.
func (m Model) renderWaitingStatePanel(width, height int) string {
	// Animated dots
	elapsed := time.Since(m.startTime)
	dotCount := int(elapsed.Seconds()*2) % 4

	dots := strings.Repeat(".", dotCount+1)
	if !m.asciiMode {
		dots = strings.Repeat("·", dotCount+1)
	}

	icon := "◐"
	if m.asciiMode {
		// Rotating ASCII spinner
		spinChars := []string{"|", "/", "-", "\\"}
		icon = spinChars[int(elapsed.Seconds()*4)%4]
	}

	// Get the waiting state message
	stateMsg := m.getWaitingStateMessage()
	if stateMsg == "" {
		stateMsg = "waiting"
	}

	msg := Styles.Running.Render(icon) + " " + Styles.Phase.Render(stateMsg) + Styles.Dim.Render(dots)

	// Build panel with centered message
	var result strings.Builder

	// Center vertically
	topPadding := (height - 1) / 2
	for i := 0; i < topPadding; i++ {
		result.WriteString(strings.Repeat(" ", width))
		result.WriteString("\n")
	}

	// Center horizontally
	msgWidth := lipgloss.Width(msg)
	leftPadding := (width - msgWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}
	result.WriteString(strings.Repeat(" ", leftPadding))
	result.WriteString(msg)
	// Pad to full width
	rightPadding := width - leftPadding - msgWidth
	if rightPadding > 0 {
		result.WriteString(strings.Repeat(" ", rightPadding))
	}

	// Fill remaining lines
	for i := topPadding + 1; i < height; i++ {
		result.WriteString("\n")
		result.WriteString(strings.Repeat(" ", width))
	}

	return result.String()
}

// renderCreatingSummary shows a centered "Rendering summary..." message (full width version).
func (m Model) renderCreatingSummary(height int) string {
	// Animated dots
	elapsed := time.Since(m.startTime)
	dotCount := int(elapsed.Seconds()*2) % 4

	dots := strings.Repeat(".", dotCount+1)
	if !m.asciiMode {
		dots = strings.Repeat("·", dotCount+1)
	}

	icon := "◐"
	if m.asciiMode {
		// Rotating ASCII spinner
		spinChars := []string{"|", "/", "-", "\\"}
		icon = spinChars[int(elapsed.Seconds()*4)%4]
	}

	msg := Styles.Running.Render(icon) + " " + Styles.Phase.Render("Rendering summary") + Styles.Dim.Render(dots)

	// Center vertically and horizontally
	var result strings.Builder
	topPadding := (height - 1) / 2
	for i := 0; i < topPadding; i++ {
		result.WriteString("\n")
	}

	msgWidth := lipgloss.Width(msg)
	leftPadding := (m.width - msgWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}
	result.WriteString(strings.Repeat(" ", leftPadding))
	result.WriteString(msg)

	for i := topPadding + 1; i < height; i++ {
		result.WriteString("\n")
	}

	return result.String()
}

// renderExitLampRow renders a single-line visual countdown with 30 lamps turning off right to left.
func (m Model) renderExitLampRow() string {
	const totalLamps = 30

	// Calculate lamps remaining: 10 seconds = 30 lamps, so 3 lamps per second
	lampsLit := m.exitCountdownSecs * 3
	if lampsLit > totalLamps {
		lampsLit = totalLamps
	}
	if lampsLit < 0 {
		lampsLit = 0
	}

	// Lamp characters
	lampOn := "●"
	lampOff := "○"
	if m.asciiMode {
		lampOn = "O"
		lampOff = "."
	}

	// Build lamp string: lit lamps on left, off lamps on right
	var lamps strings.Builder
	// Color based on time remaining
	var color lipgloss.Color
	if lampsLit > 20 {
		color = lipgloss.Color("34") // Green - plenty of time
	} else if lampsLit > 10 {
		color = lipgloss.Color("214") // Yellow/orange - getting low
	} else {
		color = lipgloss.Color("196") // Red - almost out
	}

	for i := 0; i < totalLamps; i++ {
		if i < lampsLit {
			lamps.WriteString(lipgloss.NewStyle().Foreground(color).Render(lampOn))
		} else {
			lamps.WriteString(Styles.Dim.Render(lampOff))
		}
	}

	// Build status message - show precise waiting state
	var msg string
	if m.summaryData != nil && m.summaryData.Success {
		msg = "Complete"
	} else if m.summaryData != nil {
		msg = "Failed"
	} else {
		msg = "Done"
	}

	// Status icon
	icon := "✓"
	iconColor := lipgloss.Color("34") // Green
	if m.summaryData != nil && !m.summaryData.Success {
		icon = "✗"
		iconColor = lipgloss.Color("196") // Red
	}
	if m.asciiMode {
		icon = "V"
		if m.summaryData != nil && !m.summaryData.Success {
			icon = "X"
		}
	}

	// Get precise waiting state
	waitingText := m.getWaitingStateMessage()

	left := lipgloss.NewStyle().Foreground(iconColor).Bold(true).Render(icon) +
		" " + Styles.Phase.Render(msg) +
		"  " + lamps.String() +
		"  " + Styles.Dim.Render(waitingText)

	// Single line with border fill
	borderLen := m.width - lipgloss.Width(left) - 4
	if borderLen < 1 {
		borderLen = 1
	}

	return "─" + left + " " + Styles.Border.Render(strings.Repeat("─", borderLen)) + "─"
}

// renderInitPaneCompact renders a compact single-line Init pane when initialization is complete.
func (m Model) renderInitPaneCompact() string {
	pane := m.panes[PhaseInit]
	icon := m.phaseIcon(pane.Status)

	var iconStyle lipgloss.Style
	switch pane.Status {
	case PhaseComplete:
		iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("71")) // Green
	case PhaseFailed:
		iconStyle = Styles.Error
	default:
		iconStyle = Styles.Dim
	}

	// Build compact summary: "✓ Initialization: 33 components, 3 layers, 15 workers"
	left := iconStyle.Render(icon) + " " + Styles.Dim.Render("Initialization")

	if m.initSummary != nil {
		var parts []string
		if m.initSummary.ComponentCount > 0 {
			parts = append(parts, fmt.Sprintf("%d components", m.initSummary.ComponentCount))
		}
		if m.initSummary.LayerCount > 0 {
			parts = append(parts, fmt.Sprintf("%d layers", m.initSummary.LayerCount))
		}
		if m.initSummary.EffectiveWorkers > 0 {
			parts = append(parts, fmt.Sprintf("%d workers", m.initSummary.EffectiveWorkers))
		}
		if len(parts) > 0 {
			left += ": " + Styles.Dim.Render(strings.Join(parts, ", "))
		}
	}

	// Single line with border
	borderLen := m.width - lipgloss.Width(left) - 4
	if borderLen < 1 {
		borderLen = 1
	}

	return "─" + left + " " + Styles.Border.Render(strings.Repeat("─", borderLen)) + "─"
}

// renderTabGridContent renders the tab grid as compact single-line tabs.
// Format: [◦name①] [▸name②] [✓name③] - no borders, maximum density.
func (m Model) renderTabGridContent(tabs []*ModuleState, width, height int) string {
	if len(tabs) == 0 {
		return Styles.Dim.Render("No components")
	}

	// Calculate how many tabs fit per row
	// Each tab: [icon name weight] = ~18 chars
	const tabWidth = 18
	tabsPerRow := width / tabWidth
	if tabsPerRow < 1 {
		tabsPerRow = 1
	}
	if tabsPerRow > 6 {
		tabsPerRow = 6 // Cap at 6 for readability
	}

	effectiveActiveTab := m.getEffectiveActiveTab()

	// Color map: each state has background color and text color (black/white for contrast)
	// Design: red=fail, green=success, blue=cache, grey=pending, orange=running
	type cellColors struct {
		bg   lipgloss.Color
		text lipgloss.Color
	}
	colorMap := map[ModuleStatus]cellColors{
		ModulePending:  {bg: lipgloss.Color("240"), text: lipgloss.Color("0")},   // Grey bg, black text
		ModuleRunning:  {bg: lipgloss.Color("208"), text: lipgloss.Color("0")},   // Orange bg, black text
		ModuleComplete: {bg: lipgloss.Color("34"), text: lipgloss.Color("0")},    // Green bg, black text
		ModuleSkipped:  {bg: lipgloss.Color("27"), text: lipgloss.Color("255")},  // Blue bg, white text
		ModuleFailed:   {bg: lipgloss.Color("196"), text: lipgloss.Color("255")}, // Red bg, white text
	}

	// Helper to render a single compact tab
	renderCompactTab := func(state *ModuleState) string {
		isActive := state.Moniker == effectiveActiveTab
		isHovered := state.Moniker == m.hoveredTab && !isActive

		// Weight indicator (circled number)
		var weightStr string
		if m.asciiMode {
			weightStr = fmt.Sprintf("%d", state.Weight)
		} else {
			weightStr = weightDigit(state.Weight)
		}

		// Fixed label width: tabWidth - spaces(3) - weight(2)
		labelWidth := tabWidth - 5
		if labelWidth < 4 {
			labelWidth = 4
		}

		// Get the full name for marquee effect
		fullName := state.Moniker
		label := fullName

		// Marquee scrolling for hovered tab
		if isHovered && len(fullName) > labelWidth {
			// Delay before scrolling starts (10 ticks = 1 second to read visible part)
			// Then advance position every 4 ticks (400ms per char)
			const startDelay = 10
			if m.hoveredTabScroll > startDelay {
				effectiveScroll := (m.hoveredTabScroll - startDelay) / 4
				scrollPos := effectiveScroll % (len(fullName) + 3) // +3 for gap before wrap
				if scrollPos < len(fullName) {
					// Show scrolled portion
					label = fullName[scrollPos:]
					if len(label) < labelWidth {
						// Add gap and wrap around
						label = label + "   " + fullName
					}
				} else {
					// In the gap, show start of name
					gapPos := scrollPos - len(fullName)
					label = strings.Repeat(" ", 3-gapPos) + fullName
				}
			}
		}

		// Truncate/pad label to fixed width
		labelLen := lipgloss.Width(label)
		if labelLen > labelWidth {
			label = label[:labelWidth]
		} else if labelLen < labelWidth {
			// Pad with dots for non-hovered, spaces for hovered
			if isHovered {
				label = label + strings.Repeat(" ", labelWidth-labelLen)
			} else {
				label = label + strings.Repeat(".", labelWidth-labelLen)
			}
		}

		// Get colors from map
		colors := colorMap[state.Status]
		if colors.bg == "" {
			colors = cellColors{bg: lipgloss.Color("236"), text: lipgloss.Color("255")}
		}

		// Build tab content: " name weight"
		content := " " + label + " " + weightStr

		// Pad to fixed width
		contentWidth := lipgloss.Width(content)
		if contentWidth < tabWidth {
			content += strings.Repeat(" ", tabWidth-contentWidth)
		}

		// Apply styling with full background
		var style lipgloss.Style
		if isActive {
			// Selected: white bg, black text, bold
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("255")).
				Bold(true)
		} else if isHovered {
			// Hovered: slightly lighter bg
			style = lipgloss.NewStyle().
				Foreground(colors.text).
				Background(lipgloss.Color("250"))
		} else {
			// Normal: use color map
			style = lipgloss.NewStyle().
				Foreground(colors.text).
				Background(colors.bg)
			if state.Status == ModuleRunning {
				style = style.Bold(true)
			}
		}

		styledTab := style.Render(content)
		return zone.Mark(state.Moniker, styledTab)
	}

	// Helper to render a row of tabs
	renderTabRow := func(layerTabs []*ModuleState, rowStart int) []string {
		var tabParts []string

		for colIdx := 0; colIdx < tabsPerRow; colIdx++ {
			tabIdx := rowStart + colIdx
			if tabIdx < len(layerTabs) {
				tabParts = append(tabParts, renderCompactTab(layerTabs[tabIdx]))
			}
		}

		// Single row output
		return []string{strings.Join(tabParts, " ")}
	}

	var rows []string
	usedLayerGrouping := false

	// Group tabs by layer if ExecutionTree is available
	if m.initSummary != nil && len(m.initSummary.ExecutionTree) > 0 {
		// Build a map from moniker to tab state for quick lookup
		tabMap := make(map[string]*ModuleState)
		for _, t := range tabs {
			tabMap[t.Moniker] = t
		}

		// Count how many tabs we can match to layers
		matchedTabs := 0
		for _, layer := range m.initSummary.ExecutionTree {
			for _, module := range layer.Modules {
				for _, comp := range module.Components {
					moniker := module.Name + ":" + comp
					if _, ok := tabMap[moniker]; ok {
						matchedTabs++
					}
				}
			}
		}

		// Only use layer grouping if we matched at least half the tabs
		if matchedTabs > 0 && matchedTabs >= len(tabs)/2 {
			usedLayerGrouping = true
			for layerIdx, layer := range m.initSummary.ExecutionTree {
				// Layer header
				layerHeader := Styles.Dim.Render(fmt.Sprintf("── Layer %d ──", layerIdx+1))
				rows = append(rows, layerHeader)

				// Collect tabs for this layer
				var layerTabs []*ModuleState
				for _, module := range layer.Modules {
					for _, comp := range module.Components {
						moniker := module.Name + ":" + comp
						if state, ok := tabMap[moniker]; ok {
							layerTabs = append(layerTabs, state)
						}
					}
				}

				// Render tabs in rows of 3
				for rowStart := 0; rowStart < len(layerTabs); rowStart += tabsPerRow {
					tabRows := renderTabRow(layerTabs, rowStart)
					rows = append(rows, tabRows...)
				}
			}
		}
	}

	// Fallback: flat list (no layer grouping or monikers didn't match)
	if !usedLayerGrouping {
		for rowStart := 0; rowStart < len(tabs); rowStart += tabsPerRow {
			tabRows := renderTabRow(tabs, rowStart)
			rows = append(rows, tabRows...)
		}
	}

	// Limit to available height
	if len(rows) > height {
		rows = rows[:height]
	}

	return strings.Join(rows, "\n")
}

// renderTreePanel renders the execution tree as a panel for side-by-side layout.
func (m Model) renderTreePanel(tabs []*ModuleState, width, height int) string {
	var b strings.Builder

	// Header with tree mode indicator
	title := "Components"
	modeIndicator := "[Tree]"

	headerLeft := "┌─ " + Styles.Dim.Render(title) + " "
	headerRight := Styles.Dim.Render(modeIndicator) + " ─┐"
	headerBorderLen := width - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight)
	if headerBorderLen < 1 {
		headerBorderLen = 1
	}
	b.WriteString(headerLeft + Styles.Border.Render(strings.Repeat("─", headerBorderLen)) + headerRight + "\n")

	// Tree content area
	contentHeight := height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Render all tree content (for scrolling)
	treeContent := m.renderTreeContent(tabs, width-2, contentHeight+m.tabsScrollOffset+50)
	treeLines := strings.Split(treeContent, "\n")

	// Apply scroll offset
	scrollOffset := m.tabsScrollOffset
	if scrollOffset > len(treeLines)-contentHeight {
		scrollOffset = len(treeLines) - contentHeight
		if scrollOffset < 0 {
			scrollOffset = 0
		}
	}

	for i := 0; i < contentHeight; i++ {
		lineIdx := i + scrollOffset
		if lineIdx < len(treeLines) {
			line := treeLines[lineIdx]
			lineWidth := lipgloss.Width(line)
			padding := width - 2 - lineWidth
			if padding < 0 {
				padding = 0
			}
			b.WriteString(Styles.Border.Render("│") + line + strings.Repeat(" ", padding) + Styles.Border.Render("│") + "\n")
		} else {
			b.WriteString(Styles.Border.Render("│") + strings.Repeat(" ", width-2) + Styles.Border.Render("│") + "\n")
		}
	}

	// Footer
	footerBorderLen := width - 2
	if footerBorderLen < 1 {
		footerBorderLen = 1
	}
	b.WriteString("└" + Styles.Border.Render(strings.Repeat("─", footerBorderLen)) + "┘")

	return b.String()
}

// renderTreeContent renders the execution tree content without borders.
func (m Model) renderTreeContent(tabs []*ModuleState, width, height int) string {
	if m.initSummary == nil || len(m.initSummary.ExecutionTree) == 0 {
		// Fallback: group tabs by module
		return m.renderTabsAsTree(tabs, width, height)
	}

	// Check if monikers match ExecutionTree format (same logic as tab grid)
	tabMap := make(map[string]*ModuleState)
	for _, t := range tabs {
		tabMap[t.Moniker] = t
	}
	matchedTabs := 0
	for _, layer := range m.initSummary.ExecutionTree {
		for _, module := range layer.Modules {
			for _, comp := range module.Components {
				moniker := module.Name + ":" + comp
				if _, ok := tabMap[moniker]; ok {
					matchedTabs++
				}
			}
		}
	}
	// If less than half match, fall back to module grouping
	if matchedTabs == 0 || matchedTabs < len(tabs)/2 {
		return m.renderTabsAsTree(tabs, width, height)
	}

	// Build tree from ExecutionTree
	var lines []string
	effectiveActiveTab := m.getEffectiveActiveTab()

	for layerIdx, layer := range m.initSummary.ExecutionTree {
		// Layer header
		layerLine := Styles.Dim.Render(fmt.Sprintf("Layer %d", layerIdx+1))
		lines = append(lines, layerLine)

		for modIdx, module := range layer.Modules {
			isLastMod := modIdx == len(layer.Modules)-1

			// Module line with tree branch
			modBranch := "├─"
			if isLastMod {
				modBranch = "└─"
			}
			modLine := Styles.Dim.Render(modBranch) + " " + Styles.Phase.Render(module.Name)
			lines = append(lines, modLine)

			// Components under this module
			for compIdx, comp := range module.Components {
				isLastComp := compIdx == len(module.Components)-1

				// Find matching tab for this component
				moniker := module.Name + ":" + comp
				var tabState *ModuleState
				for _, t := range tabs {
					if t.Moniker == moniker {
						tabState = t
						break
					}
				}

				// Component line
				compBranch := "│   ├─"
				if isLastMod {
					compBranch = "    ├─"
				}
				if isLastComp {
					if isLastMod {
						compBranch = "    └─"
					} else {
						compBranch = "│   └─"
					}
				}

				// Status icon and styling
				icon := "○"
				var style lipgloss.Style = Styles.Dim
				if tabState != nil {
					icon = m.statusIcon(tabState.Status)
					switch tabState.Status {
					case ModuleRunning:
						style = Styles.TabRunning
					case ModuleComplete:
						style = Styles.TabComplete
					case ModuleSkipped:
						style = Styles.TabSkipped
					case ModuleFailed:
						style = Styles.TabFailed
					default:
						style = Styles.TabPending
					}
				}

				// Weight indicator
				weightStr := ""
				if tabState != nil && tabState.Weight > 0 {
					if m.asciiMode {
						weightStr = fmt.Sprintf(" w%d", tabState.Weight)
					} else {
						weightStr = " " + weightDigit(tabState.Weight)
					}
				}

				compName := comp
				maxNameLen := width - len(compBranch) - 4 - lipgloss.Width(weightStr)
				if len(compName) > maxNameLen && maxNameLen > 3 {
					compName = compName[:maxNameLen-1] + "…"
				}

				compContent := icon + " " + compName + weightStr

				// Highlight active component
				if moniker == effectiveActiveTab {
					style = style.Bold(true).Reverse(true)
				}

				compLine := Styles.Dim.Render(compBranch) + " " + style.Render(compContent)
				lines = append(lines, compLine)
			}
		}

		// Spacing between layers
		if layerIdx < len(m.initSummary.ExecutionTree)-1 {
			lines = append(lines, "")
		}
	}

	// Limit to available height
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// renderTabsAsTree renders tabs grouped by module as a tree (fallback when no ExecutionTree).
func (m Model) renderTabsAsTree(tabs []*ModuleState, width, height int) string {
	if len(tabs) == 0 {
		return Styles.Dim.Render("No components")
	}

	// Group tabs by module (first part of moniker before ":")
	moduleGroups := make(map[string][]*ModuleState)
	var moduleOrder []string

	for _, tab := range tabs {
		parts := strings.SplitN(tab.Moniker, ":", 2)
		moduleName := parts[0]
		if _, exists := moduleGroups[moduleName]; !exists {
			moduleOrder = append(moduleOrder, moduleName)
		}
		moduleGroups[moduleName] = append(moduleGroups[moduleName], tab)
	}

	var lines []string
	effectiveActiveTab := m.getEffectiveActiveTab()

	for modIdx, moduleName := range moduleOrder {
		isLastMod := modIdx == len(moduleOrder)-1

		// Module line
		modBranch := "├─"
		if isLastMod {
			modBranch = "└─"
		}
		modLine := Styles.Dim.Render(modBranch) + " " + Styles.Phase.Render(moduleName)
		lines = append(lines, modLine)

		// Components under this module
		components := moduleGroups[moduleName]
		for compIdx, tab := range components {
			isLastComp := compIdx == len(components)-1

			compBranch := "│   ├─"
			if isLastMod {
				compBranch = "    ├─"
			}
			if isLastComp {
				if isLastMod {
					compBranch = "    └─"
				} else {
					compBranch = "│   └─"
				}
			}

			icon := m.statusIcon(tab.Status)
			var style lipgloss.Style
			switch tab.Status {
			case ModuleRunning:
				style = Styles.TabRunning
			case ModuleComplete:
				style = Styles.TabComplete
			case ModuleSkipped:
				style = Styles.TabSkipped
			case ModuleFailed:
				style = Styles.TabFailed
			default:
				style = Styles.TabPending
			}

			// Component name (without module prefix)
			compName := tab.Moniker
			if parts := strings.SplitN(tab.Moniker, ":", 2); len(parts) > 1 {
				compName = parts[1]
			}

			weightStr := ""
			if tab.Weight > 0 {
				if m.asciiMode {
					weightStr = fmt.Sprintf(" w%d", tab.Weight)
				} else {
					weightStr = " " + weightDigit(tab.Weight)
				}
			}

			maxNameLen := width - len(compBranch) - 4 - lipgloss.Width(weightStr)
			if len(compName) > maxNameLen && maxNameLen > 3 {
				compName = compName[:maxNameLen-1] + "…"
			}

			compContent := icon + " " + compName + weightStr

			// Highlight active component
			if tab.Moniker == effectiveActiveTab {
				style = style.Bold(true).Reverse(true)
			}

			compLine := Styles.Dim.Render(compBranch) + " " + style.Render(compContent)
			lines = append(lines, compLine)
		}
	}

	// Limit to available height
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// renderLogsPanel renders the logs panel for side-by-side layout.
func (m Model) renderLogsPanel(width, height int) string {
	var b strings.Builder
	pane := m.panes[PhaseRun]

	// Check if active module is cached - show special header
	activeModule := m.getEffectiveActiveTab()
	var isCachedModule bool
	if activeModule != "" {
		if state, exists := m.moduleStates[activeModule]; exists && state.Status == ModuleSkipped {
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
		// Normal header: ▶ Building: module:component 1.2s 5/10
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
				if state, exists := m.moduleStates[activeModule]; exists {
					if state.Status == ModuleRunning {
						moduleElapsed = time.Since(state.StartTime).Round(time.Millisecond * 100)
					} else if state.Status == ModuleComplete || state.Status == ModuleSkipped || state.Status == ModuleFailed {
						moduleElapsed = state.EndTime.Sub(state.StartTime).Round(time.Millisecond * 100)
					}
				}
			}

			if m.totalLayers > 0 && m.layer > 0 {
				left += " " + Styles.Dim.Render(fmt.Sprintf("(layer %d/%d)", m.layer, m.totalLayers))
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

	// Render special content for cached modules
	if isCachedModule {
		if state, exists := m.moduleStates[activeModule]; exists {
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
		if moduleBuffer := m.GetActiveModuleBuffer(); moduleBuffer != nil {
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
			lineContent := m.renderLogLine(lines[i], width-4)
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
func (m Model) renderLogLine(line Line, maxWidth int) string {
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

	var prefix, styled string
	switch line.Level {
	case LevelError:
		prefix = Styles.ErrorPrefix.Render("│" + iconFail)
		styled = Styles.Error.Render(text)
	case LevelWarn:
		prefix = Styles.WarnPrefix.Render("│" + iconWarn)
		styled = Styles.Warn.Render(text)
	default:
		prefix = Styles.InfoPrefix.Render("│" + iconInfo)
		styled = Styles.Info.Render(text)
	}

	return prefix + " " + styled
}

// renderCachedContent renders special content for cached/skipped modules.
func (m Model) renderCachedContent(b *strings.Builder, moniker string, state *ModuleState, width, contentHeight int) {
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

// renderTabBar renders tab bar for module switching.
// Dynamically expands to multiple rows if tabs overflow.
func (m Model) renderTabBar(tabs []*ModuleState) string {
	const tabWidth = 20 // Fixed width for uniform tabs

	// Build list of all tab entries (modules first, then "All" at the end)
	var allTabs []tabEntry

	// Add module tabs first
	for _, state := range tabs {
		icon := m.statusIcon(state.Status)

		// Use full module:component name
		label := state.Moniker

		// Weight as circled digit or suffix based on mode
		var weightStr string
		if m.asciiMode {
			weightStr = fmt.Sprintf(" w%d", state.Weight)
		} else {
			// Unicode mode: show circled digit with trailing space for background coverage
			weightStr = " " + weightDigit(state.Weight) + " "
		}

		// Calculate space for name
		// Use lipgloss.Width for correct Unicode width (circled digits are 2 chars wide)
		iconWidth := 1
		weightWidth := lipgloss.Width(weightStr)
		nameSpace := tabWidth - iconWidth - 1 - weightWidth - 1
		if nameSpace < 4 {
			nameSpace = 4
		}

		if len(label) > nameSpace {
			if m.asciiMode {
				label = label[:nameSpace-2] + ".."
			} else {
				label = label[:nameSpace-1] + "…"
			}
		}

		// Build tab text
		tabText := fmt.Sprintf("%s %s", icon, label)
		padLen := tabWidth - lipgloss.Width(tabText) - lipgloss.Width(weightStr)
		if padLen > 0 {
			tabText += strings.Repeat(" ", padLen)
		}
		tabText += weightStr

		// Style based on selection, hover, and status
		effectiveActiveTab := m.getEffectiveActiveTab()
		isActive := state.Moniker == effectiveActiveTab
		isHovered := state.Moniker == m.hoveredTab

		// Get base style from status
		var style lipgloss.Style
		switch state.Status {
		case ModulePending:
			style = Styles.TabPending
		case ModuleRunning:
			style = Styles.TabRunning
		case ModuleComplete:
			style = Styles.TabComplete
		case ModuleSkipped:
			style = Styles.TabSkipped
		case ModuleFailed:
			style = Styles.TabFailed
		default:
			style = Styles.TabDim
		}

		// Apply hover effect (dimmed highlight background)
		if isHovered && !isActive {
			style = style.Background(lipgloss.Color("238"))
		}

		// If active, use reverse video effect for clear selection indicator
		if isActive {
			style = style.Bold(true).Reverse(true)
		}

		allTabs = append(allTabs, tabEntry{text: tabText, style: style, zoneID: state.Moniker})
	}

	// Fixed number of tabs per row
	const tabsPerRow = 8

	// Split into rows
	var rows [][]tabEntry
	for i := 0; i < len(allTabs); i += tabsPerRow {
		end := i + tabsPerRow
		if end > len(allTabs) {
			end = len(allTabs)
		}
		rows = append(rows, allTabs[i:end])
	}

	// Fill last row with empty filler tabs for visual coherence
	if len(rows) > 0 {
		lastRow := rows[len(rows)-1]
		for len(lastRow) < tabsPerRow {
			// Create empty filler tab (no name, no weight)
			fillerText := strings.Repeat(" ", tabWidth)
			lastRow = append(lastRow, tabEntry{
				text:  fillerText,
				style: Styles.TabDim,
			})
		}
		rows[len(rows)-1] = lastRow
	}

	// Render all rows
	var result strings.Builder
	for i, row := range rows {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(m.renderTabRow(row, tabWidth))
	}
	return result.String()
}

// statusIcon returns the appropriate icon for a module status.
func (m Model) statusIcon(status ModuleStatus) string {
	if m.asciiMode {
		switch status {
		case ModulePending:
			return "o"
		case ModuleRunning:
			return ">"
		case ModuleComplete:
			return "V"
		case ModuleSkipped:
			return "="
		case ModuleFailed:
			return "X"
		default:
			return "?"
		}
	}
	// Unicode mode
	switch status {
	case ModulePending:
		return "◦"
	case ModuleRunning:
		return "▶"
	case ModuleComplete:
		return "✓"
	case ModuleSkipped:
		return "⏭"
	case ModuleFailed:
		return "✗"
	default:
		return "?"
	}
}

// padTab pads or truncates a tab label to the given width.
func (m Model) padTab(s string, width int) string {
	w := lipgloss.Width(s)
	if w > width {
		if m.asciiMode {
			return s[:width-2] + ".."
		}
		return s[:width-1] + "…"
	}
	return s + strings.Repeat(" ", width-w)
}

// renderTabRow renders a single row of tabs as lego-style blocks.
// Uniform width tabs with minimal separators.
func (m Model) renderTabRow(tabs []tabEntry, tabWidth int) string {
	if len(tabs) == 0 {
		remaining := m.width - 2
		if remaining < 0 {
			remaining = 0
		}
		return Styles.Border.Render("│") + strings.Repeat(" ", remaining) + Styles.Border.Render("│")
	}

	var b strings.Builder

	// Left border
	b.WriteString(Styles.Border.Render("│"))

	usedWidth := 2 // outer borders

	for i, tab := range tabs {
		// Thin separator between tabs (not before first)
		if i > 0 {
			b.WriteString(Styles.Border.Render("│"))
			usedWidth++
		}

		// Render tab content with style, wrapped in zone for mouse detection
		styledTab := tab.style.Render(tab.text)
		if tab.zoneID != "" {
			styledTab = zone.Mark(tab.zoneID, styledTab)
		}
		b.WriteString(styledTab)
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

// tabEntry represents a single tab for rendering.
type tabEntry struct {
	text   string
	style  lipgloss.Style
	zoneID string // Zone ID for mouse click detection (empty for filler tabs)
}

// renderPaneHeader renders a pane's header line.
func (m Model) renderPaneHeader(phase Phase) string {
	pane := m.panes[phase]

	// Build icon and name
	icon := m.phaseIcon(pane.Status)
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
		iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("71")) // Green
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
		// Show active module:component (from selected tab)
		activeModule := m.getEffectiveActiveTab()
		var moduleElapsed time.Duration
		if activeModule != "" {
			left += ": " + Styles.Phase.Render(activeModule)
			// Get elapsed time for this specific module
			if state, exists := m.moduleStates[activeModule]; exists {
				if state.Status == ModuleRunning {
					moduleElapsed = time.Since(state.StartTime).Round(time.Millisecond * 100)
				} else if state.Status == ModuleComplete || state.Status == ModuleSkipped || state.Status == ModuleFailed {
					// Show final duration for completed modules
					moduleElapsed = state.EndTime.Sub(state.StartTime).Round(time.Millisecond * 100)
				}
			}
		}

		// Add layer info if using layered execution
		if m.totalLayers > 0 && m.layer > 0 {
			left += " " + Styles.Dim.Render(fmt.Sprintf("(layer %d/%d)", m.layer, m.totalLayers))
		}

		// Show module elapsed time and progress count
		left = fmt.Sprintf("%s %s %d/%d",
			left,
			Styles.Time.Render(formatElapsed(moduleElapsed)),
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

// renderResourcesPane renders a bordered pane showing resource pressure and system metrics.
// Returns empty string if no scheduler is active.
func (m Model) renderResourcesPane() string {
	// Get component scheduler stats (the single weighted scheduler)
	var pressureUsed, pressureCap int

	for _, lock := range m.locks {
		if lock.Name == "component-scheduler" {
			pressureUsed = lock.Used
			pressureCap = lock.Capacity
			break
		}
	}

	// Don't show if no scheduler is active
	if pressureCap == 0 {
		return ""
	}

	var result strings.Builder

	// Header: ┌─ Resources ──────────────────────┐
	title := "Resources"
	headerLeft := "┌─ " + Styles.Dim.Render(title) + " "

	headerBorderLen := m.width - lipgloss.Width(headerLeft) - 1
	if headerBorderLen < 3 {
		headerBorderLen = 3
	}
	result.WriteString(headerLeft + Styles.Border.Render(strings.Repeat("─", headerBorderLen)) + "┐\n")

	// Content line: timer | mouse mode | pressure (weighted scheduler) | cpu dots | memory %
	// White style for labels
	white := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	// Timer with fixed width MM:SS format (white)
	elapsed := time.Since(m.startTime)
	mins := int(elapsed.Minutes())
	secs := int(elapsed.Seconds()) % 60
	timerStr := white.Render(fmt.Sprintf("%02d:%02d", mins, secs))

	// Mouse mode indicator: when OFF, show "Select" to indicate text selection is available
	// Press 'm' to toggle
	var mouseModeStr string
	if !m.mouseMode {
		cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
		mouseModeStr = cyan.Render("[m]Select")
	}

	pressureStr := white.Render("Pressure: ") + Styles.Dim.Render(fmt.Sprintf("%2d/%-2d", pressureUsed, pressureCap))

	// Per-core CPU usage as colored dots
	cpuStr := white.Render("CPU: ") + m.renderCPUDots()

	// System memory usage as dots (11 dots, red for used, dim for free)
	memStr := white.Render("Memory: ") + m.renderMemDots()

	// Active containers with pressure dots (colored by weight)
	containerDotsStr := m.renderActiveContainerDots()

	sep := Styles.Dim.Render(" │ ")
	content := timerStr
	if mouseModeStr != "" {
		content += sep + mouseModeStr
	}
	content += sep + pressureStr + sep + cpuStr + sep + memStr
	if containerDotsStr != "" {
		content += sep + white.Render("Jobs: ") + containerDotsStr
	}

	// Add "waiting for locks" indicator at end of line if any locks are being waited on
	if waitingMsg := m.getWaitingForLocksMessage(); waitingMsg != "" {
		// Highlight waiting status with yellow/orange
		waitingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		// Animated dots
		elapsed := time.Since(m.startTime)
		dotCount := int(elapsed.Seconds()*2) % 4
		dots := strings.Repeat(".", dotCount+1)
		content += sep + waitingStyle.Render("⏳ "+waitingMsg+dots)
	}

	// Add exit countdown if active
	if m.exitCountdownSecs > 0 {
		exitStr := white.Render("Exit: ") + Styles.Dim.Render(fmt.Sprintf("%d", m.exitCountdownSecs))
		content += sep + exitStr
	}

	// Content line 1: │ content                            │
	contentPadding := m.width - lipgloss.Width(content) - 4
	if contentPadding < 0 {
		contentPadding = 0
	}
	result.WriteString(Styles.Border.Render("│") + " " + content + strings.Repeat(" ", contentPadding) + " " + Styles.Border.Render("│") + "\n")

	// Content line 2: tools row (show tools grouped by type - containers | system)
	hasContainers := len(m.usedContainers) > 0
	hasSystem := len(m.usedSystemTools) > 0
	if hasContainers || hasSystem {
		// Horizontal separator between metrics and tools
		sepLen := m.width - 2
		if sepLen < 3 {
			sepLen = 3
		}
		result.WriteString("├" + Styles.Border.Render(strings.Repeat("─", sepLen)) + "┤\n")

		// Render each tool with appropriate color (orange=active, light grey/white=was active)
		orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
		lightGrey := lipgloss.NewStyle().Foreground(lipgloss.Color("250")) // Light grey for previously-active tools
		white := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))     // White for labels

		// Build active tool sets for quick lookup
		activeContainerSet := make(map[string]bool)
		for _, t := range m.activeContainers {
			activeContainerSet[t] = true
		}
		activeSystemSet := make(map[string]bool)
		for _, t := range m.activeSystemTools {
			activeSystemSet[t] = true
		}

		var toolsParts []string

		// System tools group (first)
		if hasSystem {
			var systemParts []string
			for _, tool := range m.usedSystemTools {
				if activeSystemSet[tool] {
					systemParts = append(systemParts, orange.Render(tool))
				} else {
					systemParts = append(systemParts, lightGrey.Render(tool))
				}
			}
			systemList := strings.Join(systemParts, Styles.Dim.Render(", "))
			toolsParts = append(toolsParts, white.Render("System: ")+systemList)
		}

		// Containers group (second)
		if hasContainers {
			var containerParts []string
			for _, tool := range m.usedContainers {
				if activeContainerSet[tool] {
					containerParts = append(containerParts, orange.Render(tool))
				} else {
					containerParts = append(containerParts, lightGrey.Render(tool))
				}
			}
			containersList := strings.Join(containerParts, Styles.Dim.Render(", "))
			toolsParts = append(toolsParts, white.Render("Containers: ")+containersList)
		}

		// Join groups with separator
		toolsContent := strings.Join(toolsParts, Styles.Dim.Render(" │ "))

		// Calculate padding (need to use lipgloss.Width for styled content)
		toolsPadding := m.width - lipgloss.Width(toolsContent) - 4
		if toolsPadding < 0 {
			toolsPadding = 0
		}
		result.WriteString(Styles.Border.Render("│") + " " + toolsContent + strings.Repeat(" ", toolsPadding) + " " + Styles.Border.Render("│") + "\n")
	}

	// Content line 3: active component row (shows hovered or active tab name)
	activeComponent := ""
	if m.hoveredTab != "" {
		activeComponent = m.hoveredTab
	} else if activeTab := m.getEffectiveActiveTab(); activeTab != "" {
		activeComponent = activeTab
	}
	if activeComponent != "" {
		// Horizontal separator
		sepLen := m.width - 2
		if sepLen < 3 {
			sepLen = 3
		}
		result.WriteString("├" + Styles.Border.Render(strings.Repeat("─", sepLen)) + "┤\n")

		white := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
		orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
		dim := Styles.Dim

		// Parse component string into parts: module:component:handler
		// Format: "module:component:handler" or "module:component/path:handler"
		parts := strings.Split(activeComponent, ":")
		var componentContent string

		// Determine tool label based on operation type
		toolLabel := "Tool"
		isTesting := false
		switch m.runPhaseName {
		case "building", "Building":
			toolLabel = "Builder"
		case "testing", "Testing":
			toolLabel = "Runner"
			isTesting = true
		case "linting", "Linting":
			toolLabel = "Linter"
		case "scanning", "Scanning":
			toolLabel = "Scanner"
		}

		// Special handling for test display
		// Test formats:
		// - gotest: "module:path:gotest" (3 parts)
		// - godog:  "module:featureName:testRoot:specPath:godog" (5 parts)
		if isTesting && len(parts) >= 3 {
			module := parts[0]
			runner := parts[len(parts)-1] // Last part is always the runner (gotest/godog)

			if runner == "godog" && len(parts) >= 5 {
				// BDD test: module:featureName:testRoot:specPath:godog
				featureName := parts[1]
				testRoot := parts[2]
				specPath := strings.Join(parts[3:len(parts)-1], ":")
				componentContent = white.Render("Unit: ") + orange.Render(module+":"+featureName) +
					dim.Render(" │ ") + white.Render(toolLabel+": ") + orange.Render(runner) +
					dim.Render(" │ ") + white.Render("Code: ") + orange.Render(testRoot) +
					dim.Render(" │ ") + white.Render("Spec: ") + orange.Render(specPath)
			} else {
				// Unit test: module:path:gotest (or similar)
				codePath := strings.Join(parts[1:len(parts)-1], ":")
				componentContent = white.Render("Unit: ") + orange.Render(module+":"+codePath) +
					dim.Render(" │ ") + white.Render(toolLabel+": ") + orange.Render(runner)
			}
		} else if len(parts) >= 3 {
			// Non-test: module:component:handler
			module := parts[0]
			component := parts[1]
			handler := parts[len(parts)-1]
			if len(parts) > 3 {
				component = strings.Join(parts[1:len(parts)-1], ":")
			}
			componentContent = white.Render("Unit: ") + orange.Render(module+":"+component) +
				dim.Render(" │ ") + white.Render(toolLabel+": ") + orange.Render(handler)
		} else if len(parts) == 2 {
			componentContent = white.Render("Unit: ") + orange.Render(parts[0]+":"+parts[1])
		} else {
			componentContent = white.Render("Unit: ") + orange.Render(activeComponent)
		}

		componentPadding := m.width - lipgloss.Width(componentContent) - 4
		if componentPadding < 0 {
			componentPadding = 0
		}
		result.WriteString(Styles.Border.Render("│") + " " + componentContent + strings.Repeat(" ", componentPadding) + " " + Styles.Border.Render("│") + "\n")
	}

	// Footer: └─────────────────────────────────────┘
	footerBorderLen := m.width - 2
	if footerBorderLen < 3 {
		footerBorderLen = 3
	}
	result.WriteString("└" + Styles.Border.Render(strings.Repeat("─", footerBorderLen)) + "┘")

	return result.String()
}

// renderComponentsPane renders the Components pane containing the tab bar.
// This provides a bordered container for component tabs similar to other panes.
func (m Model) renderComponentsPane(tabs []*ModuleState) string {
	if len(tabs) == 0 {
		return ""
	}

	var result strings.Builder

	// Count running, completed, skipped, failed for header summary
	var running, completed, skipped, failed int
	for _, tab := range tabs {
		switch tab.Status {
		case ModuleRunning:
			running++
		case ModuleComplete:
			completed++
		case ModuleSkipped:
			skipped++
		case ModuleFailed:
			failed++
		}
	}

	// Header: ┌─ Components: 3 running, 5/10 done ─────────────────┐
	title := "Components"
	var statusParts []string
	if running > 0 {
		statusParts = append(statusParts, fmt.Sprintf("%d running", running))
	}
	if completed > 0 || skipped > 0 || failed > 0 {
		statusParts = append(statusParts, fmt.Sprintf("%d/%d done", completed+skipped+failed, len(tabs)))
	}
	if len(statusParts) > 0 {
		title += ": " + strings.Join(statusParts, ", ")
	}

	headerLeft := "┌─ " + Styles.Dim.Render(title) + " "
	headerBorderLen := m.width - lipgloss.Width(headerLeft) - 1
	if headerBorderLen < 3 {
		headerBorderLen = 3
	}
	result.WriteString(headerLeft + Styles.Border.Render(strings.Repeat("─", headerBorderLen)) + "┐\n")

	// Tab bar content (may be multiple rows)
	tabBar := m.renderTabBarContent(tabs)
	result.WriteString(tabBar)

	// Footer: └─────────────────────────────────────────────────────┘
	footerBorderLen := m.width - 2
	if footerBorderLen < 3 {
		footerBorderLen = 3
	}
	result.WriteString("\n└" + Styles.Border.Render(strings.Repeat("─", footerBorderLen)) + "┘")

	return result.String()
}

// renderTabBarContent renders the tab rows without outer borders (used inside Components pane).
func (m Model) renderTabBarContent(tabs []*ModuleState) string {
	const tabWidth = 20 // Fixed width for uniform tabs

	// Build list of all tab entries (modules first, then "All" at the end)
	var allTabs []tabEntry

	// Add module tabs first
	for _, state := range tabs {
		icon := m.statusIcon(state.Status)

		// Use full module:component name
		label := state.Moniker

		// Weight as circled digit or suffix based on mode
		var weightStr string
		if m.asciiMode {
			weightStr = fmt.Sprintf(" w%d", state.Weight)
		} else {
			// Unicode mode: show circled digit with trailing space for background coverage
			weightStr = " " + weightDigit(state.Weight) + " "
		}

		// Calculate space for name
		// Use lipgloss.Width for correct Unicode width (circled digits are 2 chars wide)
		iconWidth := 1
		weightWidth := lipgloss.Width(weightStr)
		nameSpace := tabWidth - iconWidth - 1 - weightWidth - 1
		if nameSpace < 4 {
			nameSpace = 4
		}

		if len(label) > nameSpace {
			if m.asciiMode {
				label = label[:nameSpace-2] + ".."
			} else {
				label = label[:nameSpace-1] + "…"
			}
		}

		// Build tab text
		tabText := fmt.Sprintf("%s %s", icon, label)
		padLen := tabWidth - lipgloss.Width(tabText) - lipgloss.Width(weightStr)
		if padLen > 0 {
			tabText += strings.Repeat(" ", padLen)
		}
		tabText += weightStr

		// Style based on selection, hover, and status
		effectiveActiveTab := m.getEffectiveActiveTab()
		isActive := state.Moniker == effectiveActiveTab
		isHovered := state.Moniker == m.hoveredTab

		// Get base style from status
		var style lipgloss.Style
		switch state.Status {
		case ModulePending:
			style = Styles.TabPending
		case ModuleRunning:
			style = Styles.TabRunning
		case ModuleComplete:
			style = Styles.TabComplete
		case ModuleSkipped:
			style = Styles.TabSkipped
		case ModuleFailed:
			style = Styles.TabFailed
		default:
			style = Styles.TabDim
		}

		// Apply hover effect (dimmed highlight background)
		if isHovered && !isActive {
			style = style.Background(lipgloss.Color("238"))
		}

		// If active, use reverse video effect for clear selection indicator
		if isActive {
			style = style.Bold(true).Reverse(true)
		}

		allTabs = append(allTabs, tabEntry{text: tabText, style: style, zoneID: state.Moniker})
	}

	// Fixed number of tabs per row
	const tabsPerRow = 8

	// Split into rows
	var rows [][]tabEntry
	for i := 0; i < len(allTabs); i += tabsPerRow {
		end := i + tabsPerRow
		if end > len(allTabs) {
			end = len(allTabs)
		}
		rows = append(rows, allTabs[i:end])
	}

	// Fill last row with empty filler tabs for visual coherence
	if len(rows) > 0 {
		lastRow := rows[len(rows)-1]
		for len(lastRow) < tabsPerRow {
			// Create empty filler tab (no name, no weight)
			fillerText := strings.Repeat(" ", tabWidth)
			lastRow = append(lastRow, tabEntry{
				text:  fillerText,
				style: Styles.TabDim,
			})
		}
		rows[len(rows)-1] = lastRow
	}

	// Render all rows
	var result strings.Builder
	for i, row := range rows {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(m.renderTabRow(row, tabWidth))
	}
	return result.String()
}

// renderCPUDots returns colored dots representing per-core CPU usage.
// Green = low, Yellow = medium, Orange = very high, Red = maxed
func (m Model) renderCPUDots() string {
	perCore, err := cpu.Percent(0, true) // true = per CPU
	if err != nil || len(perCore) == 0 {
		return "----"
	}

	var dots strings.Builder
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("71"))   // Green
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
	orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))    // Red
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))    // Dim gray

	// Choose characters based on mode
	filledChar, emptyChar := "●", "○"
	if m.asciiMode {
		filledChar, emptyChar = "*", "o"
	}

	for _, pct := range perCore {
		switch {
		case pct < 5:
			dots.WriteString(dim.Render(emptyChar))
		case pct < 50:
			dots.WriteString(green.Render(filledChar))
		case pct < 80:
			dots.WriteString(yellow.Render(filledChar))
		case pct < 95:
			dots.WriteString(orange.Render(filledChar))
		default:
			dots.WriteString(red.Render(filledChar))
		}
	}

	return dots.String()
}

// renderMemDots returns 11 dots representing memory usage.
// Color progression: green (low) → yellow → orange → red (high)
func (m Model) renderMemDots() string {
	const totalDots = 11

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return strings.Repeat("-", totalDots)
	}

	// Calculate how many dots should be active based on usage
	usedPct := memInfo.UsedPercent
	// Map percentage to dots: 0% = 1 active, 100% = 11 active
	activeDots := 1 + int(usedPct/10.0+0.5)
	if activeDots > totalDots {
		activeDots = totalDots
	}
	if activeDots < 1 {
		activeDots = 1
	}

	green := lipgloss.NewStyle().Foreground(lipgloss.Color("71"))   // 0-30%
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // 30-60%
	orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // 60-90%
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))    // 90-100%
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Choose characters based on mode
	filledChar, emptyChar := "●", "○"
	if m.asciiMode {
		filledChar, emptyChar = "*", "o"
	}

	var dots strings.Builder
	for i := 0; i < totalDots; i++ {
		if i >= activeDots {
			// Inactive dots are dim
			dots.WriteString(dim.Render(emptyChar))
		} else {
			// Active dots colored by position (threshold)
			// Dots 0-2 (0-30%): green, 3-5 (30-60%): yellow, 6-8 (60-90%): orange, 9-10 (90-100%): red
			switch {
			case i < 3:
				dots.WriteString(green.Render(filledChar))
			case i < 6:
				dots.WriteString(yellow.Render(filledChar))
			case i < 9:
				dots.WriteString(orange.Render(filledChar))
			default:
				dots.WriteString(red.Render(filledChar))
			}
		}
	}

	return dots.String()
}

// renderActiveContainerDots returns colored dots for each running job.
// Each dot represents an active job, colored by its weight/pressure:
// Green = weight 1-2 (light), Yellow = weight 3-4, Orange = weight 5-6, Red = weight 7+
func (m Model) renderActiveContainerDots() string {
	// Collect running modules with their weights
	var runningJobs []int // Weights of running jobs
	for _, moniker := range m.moduleOrder {
		state := m.moduleStates[moniker]
		if state != nil && state.Status == ModuleRunning {
			runningJobs = append(runningJobs, state.Weight)
		}
	}

	if len(runningJobs) == 0 {
		return ""
	}

	green := lipgloss.NewStyle().Foreground(lipgloss.Color("71"))   // Weight 1-2
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Weight 3-4
	orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Weight 5-6
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))    // Weight 7+

	// Choose character based on mode
	filledChar := "●"
	if m.asciiMode {
		filledChar = "*"
	}

	var dots strings.Builder
	for _, weight := range runningJobs {
		switch {
		case weight <= 2:
			dots.WriteString(green.Render(filledChar))
		case weight <= 4:
			dots.WriteString(yellow.Render(filledChar))
		case weight <= 6:
			dots.WriteString(orange.Render(filledChar))
		default:
			dots.WriteString(red.Render(filledChar))
		}
	}

	return dots.String()
}

// weightDigit returns a colored circled digit for weight display.
// Uses filled circled digits: ❶ ❷ ❸ ❹ ❺ ❻ ❼ ❽ ❾
func weightDigit(weight int) string {
	digits := []string{"❶", "❷", "❸", "❹", "❺", "❻", "❼", "❽", "❾"}
	if weight < 1 {
		return ""
	}
	if weight > 9 {
		weight = 9
	}
	return digits[weight-1]
}

// padOrTruncate pads or truncates a string to exactly the specified width.
func padOrTruncate(s string, width int) string {
	if len(s) > width {
		if width > 1 {
			return s[:width-1] + "…"
		}
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// padOrTruncateASCII pads or truncates using ASCII-only characters (no Unicode ellipsis).
func padOrTruncateASCII(s string, width int) string {
	if len(s) > width {
		if width > 2 {
			return s[:width-2] + ".."
		}
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// padRight pads a string to the right to reach the specified width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}


// renderPaneContent renders the content lines for a pane.
func (m Model) renderPaneContent(phase Phase, height int) string {
	var b strings.Builder
	pane := m.panes[phase]

	// Special handling for Summary pane: render structured data
	if phase == PhaseSummary && m.summaryData != nil {
		return m.renderSummaryContent(height)
	}

	// For Run phase, check if we're showing a specific module's buffer
	// Use effective active tab (defaults to first tab if none explicitly selected)
	var buffer *RingBuffer
	effectiveTab := m.getEffectiveActiveTab()
	if phase == PhaseRun && effectiveTab != "" {
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

// renderPaneFooter renders the bottom border for a pane.
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

// renderResults renders the results section (rolling output after Run pane).
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
		if m.asciiMode {
			text = text[:maxLen-2] + ".."
		} else {
			text = text[:maxLen-1] + "…"
		}
	}

	// Get icons based on mode
	iconFail, iconWarn, iconInfo := m.lineIcons()

	// Style based on level
	var prefix, styled string
	switch line.Level {
	case LevelError:
		prefix = Styles.ErrorPrefix.Render("│" + iconFail)
		styled = Styles.Error.Render(text)
	case LevelWarn:
		prefix = Styles.WarnPrefix.Render("│" + iconWarn)
		styled = Styles.Warn.Render(text)
	default:
		prefix = Styles.InfoPrefix.Render("│" + iconInfo)
		styled = Styles.Info.Render(text)
	}

	return prefix + " " + styled
}

// lineIcons returns the icons for fail, warn, and info based on mode.
func (m Model) lineIcons() (fail, warn, info string) {
	if m.asciiMode {
		return "X", "!", " "
	}
	return "✗", "!", " "
}

// phaseIcon returns the appropriate icon for a phase status.
func (m Model) phaseIcon(status PhaseStatus) string {
	if m.asciiMode {
		switch status {
		case PhasePending:
			return "o"
		case PhaseActive:
			return ">"
		case PhaseComplete:
			return "V"
		case PhaseFailed:
			return "X"
		default:
			return "?"
		}
	}
	// Unicode mode
	switch status {
	case PhasePending:
		return "○"
	case PhaseActive:
		return "▶"
	case PhaseComplete:
		return "✓"
	case PhaseFailed:
		return "✗"
	default:
		return "?"
	}
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

// renderSummaryContent renders the Summary pane's structured content.
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
	contentLines = append(contentLines, primaryStatus, "") // Primary status + blank line

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

// renderPaneHeaderPlain renders a pane's header line in plain text (no ANSI styling).
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

// renderPaneContentPlain renders the content lines for a pane in plain text (no ANSI styling).
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

// renderPaneFooterPlain renders the bottom border for a pane in plain text (no ANSI styling).
func (m Model) renderPaneFooterPlain(phase Phase) string {
	borderLen := m.width - 2
	if borderLen < 3 {
		borderLen = 3
	}
	return "└" + strings.Repeat("─", borderLen) + "┘"
}

// renderSummaryContentPlain renders the Summary pane's structured content in plain text.
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
	contentLines = append(contentLines, primaryStatus, "") // Primary status + blank line

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
	contentLines = append(contentLines, primaryStatus, "") // Primary status + blank line

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

// renderModuleResultsTable renders a table of module results for final output.
// Columns: Status | Module:Component | Duration | Log Path
func (m Model) renderModuleResultsTable(tabs []*ModuleState) string {
	if len(tabs) == 0 {
		return ""
	}

	var b strings.Builder

	// Find max module name length for alignment
	maxNameLen := 20
	for _, state := range tabs {
		if len(state.Moniker) > maxNameLen {
			maxNameLen = len(state.Moniker)
		}
	}
	if maxNameLen > 35 {
		maxNameLen = 35
	}

	// Determine output dir and log filename based on run phase name
	outDir := "out/build"
	logFile := "build.log"
	if m.runPhaseName != "" {
		switch m.runPhaseName {
		case "Testing", "testing":
			outDir = "out/test"
			logFile = "test.log"
		case "Scanning", "scanning":
			outDir = "out/scan"
			logFile = "scan.log"
		case "Linting", "linting":
			outDir = "out/lint"
			logFile = "lint.log"
		}
	}

	// Calculate total table width
	tableWidth := 7 + 1 + (maxNameLen + 2) + 1 + 10 + 1 + 42 // stat(7) + borders + name + borders + duration + borders + log

	// Table title row
	title := " Results "
	titlePadding := tableWidth - len(title) - 2
	leftPad := titlePadding / 2
	rightPad := titlePadding - leftPad
	b.WriteString("┌")
	b.WriteString(strings.Repeat("─", leftPad))
	b.WriteString(title)
	b.WriteString(strings.Repeat("─", rightPad))
	b.WriteString("┐\n")

	// Column headers (extra space in Stat column to match icon width)
	b.WriteString("│ Stat  │ ")
	b.WriteString(padRight("Module:Component", maxNameLen))
	b.WriteString(" │ Duration │ ")
	b.WriteString(padRight("Log", 40))
	b.WriteString(" │\n")
	b.WriteString("├───────┼")
	b.WriteString(strings.Repeat("─", maxNameLen+2))
	b.WriteString("┼──────────┼")
	b.WriteString(strings.Repeat("─", 42))
	b.WriteString("┤\n")

	// Count results
	var passed, skipped, failed, running, pending int

	// Table rows
	for _, state := range tabs {
		// Status icon (extra space to account for Unicode width variation)
		var icon string
		switch state.Status {
		case ModuleComplete:
			icon = "  ✓  "
			passed++
		case ModuleSkipped:
			icon = "  ⏭  "
			skipped++
		case ModuleFailed:
			icon = "  ✗  "
			failed++
		case ModuleRunning:
			icon = "  ▶  "
			running++
		default:
			icon = "  ◦  "
			pending++
		}

		// Module name (truncate if needed, rune-safe)
		name := state.Moniker
		if len(name) > maxNameLen {
			runes := []rune(name)
			if len(runes) > maxNameLen-1 {
				name = string(runes[:maxNameLen-1]) + "…"
			}
		}

		// Duration
		var duration string
		if state.Status == ModuleRunning {
			duration = fmt.Sprintf("%6.1fs", time.Since(state.StartTime).Seconds())
		} else if !state.EndTime.IsZero() {
			duration = fmt.Sprintf("%6.1fs", state.EndTime.Sub(state.StartTime).Seconds())
		} else {
			duration = "    -  "
		}

		// Log path: out/<phase>/<module>/<component>/<phase>.log
		// Moniker format is "module:component"
		logPath := outDir + "/" + strings.Replace(state.Moniker, ":", "/", 1) + "/" + logFile
		if len(logPath) > 40 {
			// Truncate safely (log paths are ASCII, but use rune-safe truncation)
			runes := []rune(logPath)
			if len(runes) > 39 {
				logPath = string(runes[:39]) + "…"
			}
		}

		b.WriteString("│")
		b.WriteString(icon)
		b.WriteString(" │ ")
		b.WriteString(padRight(name, maxNameLen))
		b.WriteString(" │ ")
		b.WriteString(duration)
		b.WriteString(" │ ")
		b.WriteString(padRight(logPath, 40))
		b.WriteString(" │\n")
	}

	// Table footer with summary
	b.WriteString("└───────┴")
	b.WriteString(strings.Repeat("─", maxNameLen+2))
	b.WriteString("┴──────────┴")
	b.WriteString(strings.Repeat("─", 42))
	b.WriteString("┘\n")

	// Summary line - check for overall failure (phases or summary data)
	overallFailed := false
	if m.panes[PhaseInit].Status == PhaseFailed || m.panes[PhaseRun].Status == PhaseFailed {
		overallFailed = true
	}
	if m.summaryData != nil && !m.summaryData.Success {
		overallFailed = true
	}
	if failed > 0 {
		overallFailed = true
	}

	if overallFailed {
		b.WriteString(fmt.Sprintf("  Total: %d  ✗ FAILED", len(tabs)))
		if passed > 0 {
			b.WriteString(fmt.Sprintf("  (✓ %d passed", passed))
			if failed > 0 {
				b.WriteString(fmt.Sprintf(", ✗ %d failed", failed))
			}
			b.WriteString(")")
		} else if failed > 0 {
			b.WriteString(fmt.Sprintf("  ✗ %d failed", failed))
		}
	} else {
		b.WriteString(fmt.Sprintf("  Total: %d  ✓ %d passed", len(tabs), passed))
	}
	if skipped > 0 {
		b.WriteString(fmt.Sprintf("  ⏭ %d cached", skipped))
	}
	if running > 0 {
		b.WriteString(fmt.Sprintf("  ▶ %d running", running))
	}
	if pending > 0 {
		b.WriteString(fmt.Sprintf("  ◦ %d pending", pending))
	}
	b.WriteString("\n")

	return b.String()
}
