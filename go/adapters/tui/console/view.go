package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

// stableLampsCount is the fixed number of lamps for all resource displays.
// This ensures consistent visual width across CPU, Mem, Docker mem, and Tools.
const stableLampsCount = 16

// pressureLampZones calculates non-uniform color zone boundaries for pressure lamps.
// Distribution favors green (safe) with fewer red (critical) lamps:
// - Green: 3/8 of total (37.5%) - large safe zone
// - Yellow: fills remainder after other zones - warning zone
// - Orange: 3/16 of total (18.75%) - elevated pressure
// - Red: 1/8 of total (12.5%) - critical zone
// For 16 lamps: 6 green, 5 yellow, 3 orange, 2 red
func pressureLampZones(totalLamps int) (greenEnd, yellowEnd, orangeEnd int) {
	greenCount := totalLamps * 3 / 8   // 6 for 16 lamps
	redCount := totalLamps / 8         // 2 for 16 lamps
	orangeCount := totalLamps * 3 / 16 // 3 for 16 lamps
	yellowCount := totalLamps - greenCount - redCount - orangeCount // remainder (5 for 16)

	greenEnd = greenCount
	yellowEnd = greenEnd + yellowCount
	orangeEnd = yellowEnd + orangeCount
	// red fills the rest (totalLamps - orangeEnd)
	return
}

// getPressureColor returns a lipgloss style for the given pressure level.
// The color matches the scheduler lamp color zones (green/yellow/orange/red).
func getPressureColor(running, capacity int) lipgloss.Style {
	if capacity <= 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("255")) // white if no capacity
	}

	// Calculate which lamp position this pressure corresponds to
	activeDots := (running * stableLampsCount) / capacity
	if activeDots > stableLampsCount {
		activeDots = stableLampsCount
	}

	greenEnd, yellowEnd, orangeEnd := pressureLampZones(stableLampsCount)

	// Color based on position (use last active lamp, or 0 if none)
	pos := activeDots
	if pos > 0 {
		pos-- // 0-indexed position of last active lamp
	}

	switch {
	case pos < greenEnd:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("40")) // green
	case pos < yellowEnd:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // yellow
	case pos < orangeEnd:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // orange
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red
	}
}

// LayoutMetrics contains pre-calculated layout dimensions for the TUI.
// This is the SINGLE SOURCE OF TRUTH for layout calculations.
// Both rendering (viewPanes) and mouse handling (detectTabAt) MUST use this
// to ensure click detection aligns with rendered output.
type LayoutMetrics struct {
	InitLines       int // Lines used by init pane (always 1)
	ResourcesLines  int // Lines used by resources pane (0 if not shown)
	SelectedLines   int // Lines used by selected pane (0 if not shown)
	ComponentsStart int // Y coordinate where components panel content starts
	SummaryLines    int // Lines reserved for summary pane (0 if not shown)
	RemainingHeight int // Height available for side-by-side layout
}

// calculateLayoutMetrics computes layout dimensions based on current model state.
// This method ensures consistent calculations across rendering and mouse handling.
func (m Model) calculateLayoutMetrics() LayoutMetrics {
	metrics := LayoutMetrics{}

	// Init pane - always 1 line (compact or loading)
	metrics.InitLines = 1

	// Resources pane - render to count actual lines
	if resourcesPane := m.renderResourcesPane(); resourcesPane != "" {
		metrics.ResourcesLines = strings.Count(resourcesPane, "\n") + 1
	}

	// Selected pane - render to count actual lines
	if selectedPane := m.renderSelectedPane(); selectedPane != "" {
		metrics.SelectedLines = strings.Count(selectedPane, "\n") + 1
	}

	// Summary pane reservation
	if m.summaryData != nil {
		metrics.SummaryLines = 6 // header + 4 content + footer
	}

	// Components panel starts after init + resources + selected + panel header
	// The +1 accounts for the panel header line (┌─ Unit (of work) ─┐)
	metrics.ComponentsStart = metrics.InitLines + metrics.ResourcesLines + metrics.SelectedLines + 1

	// Remaining height for side-by-side layout
	usedLines := metrics.InitLines + metrics.ResourcesLines + metrics.SelectedLines
	metrics.RemainingHeight = m.height - usedLines - metrics.SummaryLines
	if metrics.RemainingHeight < 5 {
		metrics.RemainingHeight = 5
	}

	return metrics
}

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
// Shows only essential summary data - no pane chrome.
func (m Model) ViewFinal() string {
	var b strings.Builder

	// Render summary data if available
	if m.summaryData != nil {
		// Details (includes module table from summary.go)
		if len(m.summaryData.Details) > 0 {
			for _, line := range m.summaryData.Details {
				cleanLine := stripMarkdownPipes(line)
				b.WriteString(fmt.Sprintf("%s\n", cleanLine))
			}
		}

		// Summary status
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
// IMPORTANT: Layout calculations use calculateLayoutMetrics() as single source of truth.
func (m Model) viewPanes() string {
	var b strings.Builder

	// Use shared layout metrics for consistency with mouse handling
	metrics := m.calculateLayoutMetrics()

	// Render Init pane (always visible)
	if m.initSummary != nil {
		// Compact single line when initialization is complete
		b.WriteString(m.renderInitPaneCompact())
		b.WriteString("\n")
	} else {
		// Loading state - show animated dots
		b.WriteString(m.renderInitPaneLoading())
		b.WriteString("\n")
	}

	// Render Resources pane between Init and Run (shows locks and system metrics)
	if resourcesPane := m.renderResourcesPane(); resourcesPane != "" {
		b.WriteString(resourcesPane)
		b.WriteString("\n")
	}

	// Render Selected pane (shows details of selected/hovered component)
	if selectedPane := m.renderSelectedPane(); selectedPane != "" {
		b.WriteString(selectedPane)
		b.WriteString("\n")
	}

	// Render Run pane only if it actually started (not still pending)
	if m.panes[PhaseRun].Status != PhasePending {
		tabs := m.GetVisibleTabs()
		// Side-by-side layout: Components (tabs/tree) on LEFT, Logs on RIGHT
		// Use metrics.RemainingHeight for consistent height calculation
		sideBySide := m.renderSideBySideLayout(tabs, metrics.RemainingHeight)
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
func (m Model) renderSideBySideLayout(tabs []*UoWState, height int) string {
	// Dynamic width based on tab columns
	componentsWidth := m.ComponentsWidth()
	logsWidth := m.width - componentsWidth - 1 // -1 for separator
	if logsWidth < 40 {
		logsWidth = 40
	}

	// Render left panel (Components - always use tab grid view)
	leftPanel := m.renderTabGridPanel(tabs, componentsWidth, height)

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
func (m Model) renderTabGridPanel(tabs []*UoWState, width, height int) string {
	var b strings.Builder

	// Header: ┌─ Components: 3 running, 5/10 done ─────┐
	var running, completed, skipped, failed int
	for _, tab := range tabs {
		switch tab.Status {
		case UoWRunning:
			running++
		case UoWComplete:
			completed++
		case UoWSkipped:
			skipped++
		case UoWFailed:
			failed++
		}
	}

	title := "Units (of work)"

	headerLeft := "┌─ " + Styles.Dim.Render(title) + " "
	headerRight := "─┐"
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

	// Build compact summary: "✓ Initialization: 33 units, 15 workers"
	left := iconStyle.Render(icon) + " " + Styles.Dim.Render("Initialization")

	if m.initSummary != nil {
		var parts []string
		if m.initSummary.UoWCount > 0 {
			parts = append(parts, fmt.Sprintf("%d units", m.initSummary.UoWCount))
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
func (m Model) renderTabGridContent(tabs []*UoWState, width, height int) string {
	if len(tabs) == 0 {
		return Styles.Dim.Render("No components")
	}

	// Use configured tab columns (adjustable with left/right arrows)
	// Each tab: [name weight] = ~15 chars
	const tabWidth = 15
	tabsPerRow := m.tabColumns
	if tabsPerRow < 1 {
		tabsPerRow = 1
	}
	if tabsPerRow > 6 {
		tabsPerRow = 6
	}

	effectiveActiveTab := m.getEffectiveActiveTab()

	// Color map for badge and name areas
	// Badge gets full color, name area gets subtle background for visual continuity
	type badgeColors struct {
		bg       lipgloss.Color // Badge background
		text     lipgloss.Color // Badge text
		bgActive lipgloss.Color // Active/selected badge background (lighter)
		nameBg   lipgloss.Color // Subtle name area background
	}
	colorMap := map[UoWStatus]badgeColors{
		UoWPending:  {bg: lipgloss.Color("238"), text: lipgloss.Color("250"), bgActive: lipgloss.Color("245"), nameBg: lipgloss.Color("236")},
		UoWRunning:  {bg: lipgloss.Color("208"), text: lipgloss.Color("232"), bgActive: lipgloss.Color("220"), nameBg: lipgloss.Color("94")},
		UoWComplete: {bg: lipgloss.Color("34"), text: lipgloss.Color("232"), bgActive: lipgloss.Color("46"), nameBg: lipgloss.Color("22")},
		UoWSkipped:  {bg: lipgloss.Color("31"), text: lipgloss.Color("232"), bgActive: lipgloss.Color("45"), nameBg: lipgloss.Color("23")},
		UoWFailed:   {bg: lipgloss.Color("160"), text: lipgloss.Color("255"), bgActive: lipgloss.Color("196"), nameBg: lipgloss.Color("52")},
	}

	// Helper to render a single compact tab
	// Fixed width tabs: name on left, badge (weight) on right with colored background
	renderCompactTab := func(state *UoWState) string {
		isActive := state.Moniker == effectiveActiveTab
		isHovered := state.Moniker == m.hoveredTab && !isActive

		// Badge is 3 chars: " N " for single digit, "NN " for double
		const badgeWidth = 3
		// Name area: tabWidth - badgeWidth
		labelWidth := tabWidth - badgeWidth
		if labelWidth < 4 {
			labelWidth = 4
		}

		// Use DisplayName for tab labels
		displayName := state.DisplayName
		label := displayName

		// Marquee scrolling for hovered tab
		if isHovered && len(displayName) > labelWidth {
			const startDelay = 10
			if m.hoveredTabScroll > startDelay {
				effectiveScroll := (m.hoveredTabScroll - startDelay) / 4
				scrollPos := effectiveScroll % (len(displayName) + 3)
				if scrollPos < len(displayName) {
					label = displayName[scrollPos:]
					if len(label) < labelWidth {
						label = label + "   " + displayName
					}
				} else {
					gapPos := scrollPos - len(displayName)
					label = strings.Repeat(" ", 3-gapPos) + displayName
				}
			}
		}

		// Truncate label to fixed visual width (remove chars from end until it fits)
		for lipgloss.Width(label) > labelWidth && len(label) > 0 {
			label = label[:len(label)-1]
		}
		// Remove trailing colon if present (looks cleaner)
		if strings.HasSuffix(label, ":") {
			label = label[:len(label)-1]
		}
		// Always pad to fixed width (ensures column alignment)
		if lipgloss.Width(label) < labelWidth {
			label = label + strings.Repeat(" ", labelWidth-lipgloss.Width(label))
		}

		// Get badge colors from map
		colors := colorMap[state.Status]
		if colors.bg == "" {
			colors = badgeColors{bg: lipgloss.Color("238"), text: lipgloss.Color("250"), bgActive: lipgloss.Color("245")}
		}

		// Choose badge background based on active state
		badgeBg := colors.bg
		if isActive {
			badgeBg = colors.bgActive
		}

		// Format weight number for badge: " N " or "NN " (always 3 chars)
		var weightStr string
		if state.Weight < 10 {
			weightStr = fmt.Sprintf(" %d ", state.Weight)
		} else {
			weightStr = fmt.Sprintf("%2d ", state.Weight)
		}

		// Name style: white/light text with subtle status-colored background
		// When hovered: use the badge background color with black text
		nameBgColor := colors.nameBg
		nameFgColor := lipgloss.Color("252") // Light grey/white
		if isHovered {
			nameBgColor = colors.bg
			nameFgColor = lipgloss.Color("232") // Black text on bright background
		}
		nameStyle := lipgloss.NewStyle().
			Foreground(nameFgColor).
			Background(nameBgColor)
		namePart := nameStyle.Render(label)

		// Badge style: colored background, contrasting text
		badgeStyle := lipgloss.NewStyle().
			Foreground(colors.text).
			Background(badgeBg)
		if state.Status == UoWRunning {
			badgeStyle = badgeStyle.Bold(true)
		}
		badgePart := badgeStyle.Render(weightStr)

		// Combine: name (left) + badge (right) = fixed tabWidth
		styledTab := namePart + badgePart
		return zone.Mark(state.Moniker, styledTab)
	}

	// Helper to render a row of tabs
	renderTabRow := func(tabList []*UoWState, rowStart int) []string {
		var tabParts []string

		for colIdx := 0; colIdx < tabsPerRow; colIdx++ {
			tabIdx := rowStart + colIdx
			if tabIdx < len(tabList) {
				tabParts = append(tabParts, renderCompactTab(tabList[tabIdx]))
			}
		}

		// Skip empty rows
		if len(tabParts) == 0 {
			return nil
		}

		// Single row output
		return []string{strings.Join(tabParts, " ")}
	}

	var rows []string

	// Flat list of all tabs
	for rowStart := 0; rowStart < len(tabs); rowStart += tabsPerRow {
		tabRows := renderTabRow(tabs, rowStart)
		rows = append(rows, tabRows...)
	}

	// Limit to available height
	if len(rows) > height {
		rows = rows[:height]
	}

	return strings.Join(rows, "\n")
}

// renderTreePanel renders the execution tree as a panel for side-by-side layout.
func (m Model) renderTreePanel(tabs []*UoWState, width, height int) string {
	var b strings.Builder

	// Header with tree mode indicator
	title := "Units (of work)"
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
func (m Model) renderTreeContent(tabs []*UoWState, width, height int) string {
	if m.initSummary == nil || len(m.initSummary.ExecutionTree) == 0 {
		// Fallback: group tabs by module
		return m.renderTabsAsTree(tabs, width, height)
	}

	// Check if monikers match ExecutionTree format
	tabMap := make(map[string]*UoWState)
	for _, t := range tabs {
		tabMap[t.Moniker] = t
	}
	matchedTabs := 0
	for _, module := range m.initSummary.ExecutionTree {
		for _, uow := range module.UoWs {
			if _, ok := tabMap[uow.ID]; ok {
				matchedTabs++
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

	for modIdx, module := range m.initSummary.ExecutionTree {
		isLastMod := modIdx == len(m.initSummary.ExecutionTree)-1

		// Module line with tree branch
		modBranch := "├─"
		if isLastMod {
			modBranch = "└─"
		}
		modLine := Styles.Dim.Render(modBranch) + " " + Styles.Phase.Render(module.Name)
		lines = append(lines, modLine)

		// Units of work under this module
		for uowIdx, uow := range module.UoWs {
			isLastUoW := uowIdx == len(module.UoWs)-1

			// Find matching tab for this unit of work
			moniker := uow.ID
			var tabState *UoWState
			for _, t := range tabs {
				if t.Moniker == moniker {
					tabState = t
					break
				}
			}

			// UoW line
			uowBranch := "│   ├─"
			if isLastMod {
				uowBranch = "    ├─"
			}
			if isLastUoW {
				if isLastMod {
					uowBranch = "    └─"
				} else {
					uowBranch = "│   └─"
				}
			}

			// Status icon and styling
			icon := "○"
			var style lipgloss.Style = Styles.Dim
			if tabState != nil {
				icon = m.statusIcon(tabState.Status)
				switch tabState.Status {
				case UoWRunning:
					style = Styles.TabRunning
				case UoWComplete:
					style = Styles.TabComplete
				case UoWSkipped:
					style = Styles.TabSkipped
				case UoWFailed:
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

			displayName := uow.DisplayName
			maxNameLen := width - len(uowBranch) - 4 - lipgloss.Width(weightStr)
			if len(displayName) > maxNameLen && maxNameLen > 3 {
				displayName = displayName[:maxNameLen-1] + "…"
			}

			uowContent := icon + " " + displayName + weightStr

			// Highlight active unit of work
			if moniker == effectiveActiveTab {
				style = style.Bold(true).Reverse(true)
			}

			uowLine := Styles.Dim.Render(uowBranch) + " " + style.Render(uowContent)
			lines = append(lines, uowLine)
		}
	}

	// Limit to available height
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// renderTabsAsTree renders tabs grouped by module as a tree (fallback when no ExecutionTree).
func (m Model) renderTabsAsTree(tabs []*UoWState, width, height int) string {
	if len(tabs) == 0 {
		return Styles.Dim.Render("No components")
	}

	// Group tabs by module (first part of moniker before ":")
	moduleGroups := make(map[string][]*UoWState)
	var uowOrder []string

	for _, tab := range tabs {
		parts := strings.SplitN(tab.Moniker, ":", 2)
		moduleName := parts[0]
		if _, exists := moduleGroups[moduleName]; !exists {
			uowOrder = append(uowOrder, moduleName)
		}
		moduleGroups[moduleName] = append(moduleGroups[moduleName], tab)
	}

	var lines []string
	effectiveActiveTab := m.getEffectiveActiveTab()

	for modIdx, moduleName := range uowOrder {
		isLastMod := modIdx == len(uowOrder)-1

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
			case UoWRunning:
				style = Styles.TabRunning
			case UoWComplete:
				style = Styles.TabComplete
			case UoWSkipped:
				style = Styles.TabSkipped
			case UoWFailed:
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

// renderTabBar renders tab bar for module switching.
// Dynamically expands to multiple rows if tabs overflow.
func (m Model) renderTabBar(tabs []*UoWState) string {
	const tabWidth = 20 // Fixed width for uniform tabs

	// Build list of all tab entries (modules first, then "All" at the end)
	var allTabs []tabEntry

	// Add module tabs first
	for _, state := range tabs {
		icon := m.statusIcon(state.Status)

		// Use DisplayName for tab labels
		label := state.DisplayName

		// Weight as circled digit or suffix based on mode (with trailing space for background coverage)
		var weightStr string
		if m.asciiMode {
			weightStr = fmt.Sprintf(" w%d ", state.Weight)
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
		case UoWPending:
			style = Styles.TabPending
		case UoWRunning:
			style = Styles.TabRunning
		case UoWComplete:
			style = Styles.TabComplete
		case UoWSkipped:
			style = Styles.TabSkipped
		case UoWFailed:
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
func (m Model) statusIcon(status UoWStatus) string {
	if m.asciiMode {
		switch status {
		case UoWPending:
			return "o"
		case UoWRunning:
			return ">"
		case UoWComplete:
			return "V"
		case UoWSkipped:
			return "="
		case UoWFailed:
			return "X"
		default:
			return "?"
		}
	}
	// Unicode mode
	switch status {
	case UoWPending:
		return "◦"
	case UoWRunning:
		return "▶"
	case UoWComplete:
		return "✓"
	case UoWSkipped:
		return "⏭"
	case UoWFailed:
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
			if state, exists := m.uowStates[activeModule]; exists {
				if state.Status == UoWRunning {
					moduleElapsed = time.Since(state.StartTime).Round(time.Millisecond * 100)
				} else if state.Status == UoWComplete || state.Status == UoWSkipped || state.Status == UoWFailed {
					// Show final duration for completed modules
					moduleElapsed = state.EndTime.Sub(state.StartTime).Round(time.Millisecond * 100)
				}
			}
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
// Layout: 3 rows x 5 columns
// Row 1: CPU lamps          | Time         | (empty)        | Mem lamps       | Docker Mem lamps
// Row 2: Host scheduler     | Host weight  | cached/done/X  | Native tools    | Containers lamps
// Row 3: Docker scheduler   | Docker weight| progress count | Tools placeholder| Container tools
// Returns empty string if no scheduler is active.
func (m Model) renderResourcesPane() string {
	// Get component scheduler stats (the single weighted scheduler)
	var pressureCap int

	for _, lock := range m.locks {
		if lock.Name == "component-scheduler" {
			pressureCap = lock.Capacity
			break
		}
	}

	// Don't show if no scheduler is active
	if pressureCap == 0 {
		return ""
	}

	var result strings.Builder

	// Fixed column widths for 5-column layout
	const (
		col1Width = 24 // Scheduler lamps (CPU, Host, Docker)
		col2Width = 11 // Slots (Time, Host weight, Docker weight) - left aligned
		col3Width = 12 // Progress (empty, cached/done/failed, progress count)
		col4Width = 24 // Host resources (Mem, Native, placeholder)
		col5Width = 28 // Docker resources (Mem, Containers, Tools)
	)

	// Label widths for lamp alignment within columns
	// All labels in a column are padded to the longest label length
	const (
		col1LabelWidth = 7  // "Docker:" is longest (7 chars)
		col4LabelWidth = 7  // "Native:" is longest (7 chars)
		col5LabelWidth = 11 // "Containers:" is longest (11 chars)
	)

	// Helper to pad string to fixed width
	padTo := func(s string, width int) string {
		visWidth := lipgloss.Width(s)
		if visWidth >= width {
			return s
		}
		return s + strings.Repeat(" ", width-visWidth)
	}

	// Helper to right-align string in fixed width
	rightAlign := func(s string, width int) string {
		visWidth := lipgloss.Width(s)
		if visWidth >= width {
			return s
		}
		return strings.Repeat(" ", width-visWidth) + s
	}

	// Freeze button (rendered first so we can calculate remaining width)
	var freezeBtn string
	if m.exitCountdownSecs > 0 && m.userHasInteracted {
		mins := m.exitCountdownSecs / 60
		secs := m.exitCountdownSecs % 60
		var btnStyle lipgloss.Style
		if m.exitCountdownSecs > 60 {
			btnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)
		} else if m.exitCountdownSecs > 20 {
			btnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		} else {
			btnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		}
		freezeBtn = btnStyle.Render(fmt.Sprintf("[%d:%02d]", mins, secs))
	} else {
		freezeBtn = Styles.Dim.Render("[...]")
	}
	freezeBtn = zone.Mark("freeze-button", freezeBtn)

	// Header with column titles: ┌ Scheduler │ Slots │ Progress │ Host │ Docker ─── [...] ┐
	headerSepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("236")) // Very dark
	headerSep := headerSepStyle.Render(" │ ")
	col1Header := padTo(Styles.Dim.Render("Scheduler"), col1Width)
	col2Header := padTo(Styles.Dim.Render("Slots"), col2Width)
	col3Header := padTo(Styles.Dim.Render("Progress"), col3Width)
	col4Header := padTo(Styles.Dim.Render("Host"), col4Width)
	col5Header := Styles.Dim.Render("Docker")
	headerTitles := col1Header + headerSep + col2Header + headerSep + col3Header + headerSep + col4Header + headerSep + col5Header
	headerLeft := "┌ " + headerTitles + " "

	headerBorderLen := m.width - lipgloss.Width(headerLeft) - lipgloss.Width(freezeBtn) - 3
	if headerBorderLen < 3 {
		headerBorderLen = 3
	}
	result.WriteString(headerLeft + Styles.Border.Render(strings.Repeat("─", headerBorderLen)) + " " + freezeBtn + " ┐\n")

	// Styles
	white := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))  // Yellow for waiting warning
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))     // Red for high waiting
	veryDark := lipgloss.NewStyle().Foreground(lipgloss.Color("236")) // Very dark for separators
	sep := veryDark.Render(" │ ")

	// Content width (inside borders)
	contentWidth := m.width - 4

	// === Get capacity info ===
	capInfo := m.GetCapacityInfo()
	if capInfo.Roof == 0 {
		capInfo.Roof = pressureCap
	}
	if capInfo.PressureTarget == 0 {
		capInfo.PressureTarget = pressureCap
	}

	// Use derived counts (single source of truth)
	counts := m.DeriveCounts()
	running := counts.Running
	done := counts.Done
	cached := counts.Cached
	failed := counts.Failed

	// === Line 1: CPU | Time | (empty) | Mem | Docker Mem ===
	// CPU lamps (padded label to align with "Docker:")
	cpuDots := m.renderCPUDots()
	cpuStr := zone.Mark("res-cpu", white.Render(fmt.Sprintf("%-*s", col1LabelWidth, "CPU:"))+cpuDots)
	col1Line1 := padTo(cpuStr, col1Width)

	// Timer
	elapsed := time.Since(m.startTime)
	mins := int(elapsed.Minutes())
	secs := int(elapsed.Seconds()) % 60
	timerStr := zone.Mark("res-timer", white.Render(fmt.Sprintf("Time:%02d:%02d", mins, secs)))
	col2Line1 := padTo(timerStr, col2Width)

	// Empty column
	col3Line1 := padTo("", col3Width)

	// Host memory lamps (padded label to align with "Native:")
	memDots := m.renderMemDots()
	memStr := zone.Mark("res-mem", white.Render(fmt.Sprintf("%-*s", col4LabelWidth, "Mem:"))+memDots)
	col4Line1 := padTo(memStr, col4Width)

	// Docker memory lamps (padded label to align with "Containers:")
	dockerMemDots := m.renderDockerMemDots()
	dockerMemStr := zone.Mark("res-dmem", white.Render(fmt.Sprintf("%-*s", col5LabelWidth, "Mem:"))+dockerMemDots)
	col5Line1 := padTo(dockerMemStr, col5Width)

	line1 := col1Line1 + sep + col2Line1 + sep + col3Line1 + sep + col4Line1 + sep + col5Line1
	line1 = padTo(line1, contentWidth)
	result.WriteString(Styles.Border.Render("│") + " " + line1 + " " + Styles.Border.Render("│") + "\n")

	// === Line 2: Host scheduler | Host weight | counters | Native tools | Containers ===
	// Host scheduler lamps with green/yellow/orange/red gradient (padded label to align with "Docker:")
	hostLamps := m.renderSchedulerPressureLamps(running, capInfo.PressureTarget)
	hostLampsStr := white.Render(fmt.Sprintf("%-*s", col1LabelWidth, "Host:")) + hostLamps
	hostStr := zone.Mark("res-host", hostLampsStr)
	col1Line2 := padTo(hostStr, col1Width)

	// Host weight counter - color matches pressure lamps
	hostPressureColor := getPressureColor(running, capInfo.PressureTarget)
	hostWeightStr := hostPressureColor.Render(fmt.Sprintf("%d/%d", running, capInfo.PressureTarget))
	// Check if there are jobs waiting for host scheduler
	var hostWaiting int
	for _, lock := range m.locks {
		if lock.Name == "component-scheduler" {
			hostWaiting = lock.Waiting
			break
		}
	}
	if hostWaiting > 0 {
		wStyle := yellow
		if hostWaiting > 5 {
			wStyle = red
		}
		hostWeightStr += wStyle.Render(fmt.Sprintf(" W:%d", hostWaiting))
	}
	col2Line2 := padTo(zone.Mark("res-host-weight", hostWeightStr), col2Width)

	// Progress counters (cached/done/failed)
	blueBg := lipgloss.NewStyle().Background(lipgloss.Color("39")).Foreground(lipgloss.Color("0"))   // Cached (blue)
	greenBg := lipgloss.NewStyle().Background(lipgloss.Color("40")).Foreground(lipgloss.Color("0")) // Done (green)
	redBg := lipgloss.NewStyle().Background(lipgloss.Color("196")).Foreground(lipgloss.Color("0"))  // Failed (red)
	countersStr := blueBg.Render(fmt.Sprintf("%3d", cached)) + " " + greenBg.Render(fmt.Sprintf("%3d", done)) + " " + redBg.Render(fmt.Sprintf("%3d", failed))
	col3Line2 := zone.Mark("res-counters", rightAlign(countersStr, col3Width))

	// Native/system tools (padded label to align with "Native:" - already max width)
	nativeToolsStr := zone.Mark("res-native", white.Render(fmt.Sprintf("%-*s", col4LabelWidth, "Native:"))+m.renderToolsDots())
	col4Line2 := padTo(nativeToolsStr, col4Width)

	// Running Docker containers count (padded label to align with "Containers:" - already max width)
	containerCountDots := m.renderRunningContainersDots()
	containerCountStr := zone.Mark("res-jobs", white.Render(fmt.Sprintf("%-*s", col5LabelWidth, "Containers:"))+containerCountDots)
	col5Line2 := padTo(containerCountStr, col5Width)

	line2 := col1Line2 + sep + col2Line2 + sep + col3Line2 + sep + col4Line2 + sep + col5Line2
	line2 = padTo(line2, contentWidth)
	result.WriteString(Styles.Border.Render("│") + " " + line2 + " " + Styles.Border.Render("│") + "\n")

	// === Line 3: Docker scheduler | Docker weight | progress | placeholder | Container tools ===
	// Docker scheduler lamps
	dockerRoof := m.dockerRoof
	dockerRunning := m.dockerRunning
	dockerPressureTarget := m.dockerPressureTarget
	if dockerRoof == 0 {
		// No docker scheduler, show empty/grey lamps
		dockerRoof = 1 // avoid div by zero
	}
	if dockerPressureTarget == 0 {
		dockerPressureTarget = dockerRoof
	}
	// Docker scheduler lamps with green/yellow/orange/red gradient (padded label - "Docker:" is already max width for col1)
	dockerLamps := m.renderSchedulerPressureLamps(dockerRunning, dockerPressureTarget)
	dockerLampsStr := white.Render(fmt.Sprintf("%-*s", col1LabelWidth, "Docker:")) + dockerLamps
	dockerStr := zone.Mark("res-docker", dockerLampsStr)
	col1Line3 := padTo(dockerStr, col1Width)

	// Docker weight counter - color matches pressure lamps
	dockerPressureColor := getPressureColor(dockerRunning, dockerPressureTarget)
	dockerWeightStr := dockerPressureColor.Render(fmt.Sprintf("%d/%d", dockerRunning, dockerRoof))
	if m.dockerWaiting > 0 {
		wStyle := yellow
		if m.dockerWaiting > 5 {
			wStyle = red
		}
		dockerWeightStr += wStyle.Render(fmt.Sprintf(" W:%d", m.dockerWaiting))
	}
	col2Line3 := padTo(zone.Mark("res-docker-weight", dockerWeightStr), col2Width)

	// Progress: finalized/total
	total := 0
	if m.initSummary != nil && m.initSummary.UoWCount > 0 {
		total = m.initSummary.UoWCount
	} else {
		total = len(m.uowOrder)
	}
	finalized := done + cached + failed
	progressStr := white.Render(fmt.Sprintf("%d/%d", finalized, total))
	col3Line3 := rightAlign(progressStr, col3Width)

	// Placeholder for col4 (padded label to align with "Native:")
	veryDark = lipgloss.NewStyle().Foreground(lipgloss.Color("236")) // Very dark for unused slots
	col4Line3 := padTo(white.Render(fmt.Sprintf("%-*s", col4LabelWidth, "Tools:"))+veryDark.Render(strings.Repeat("○", stableLampsCount)), col4Width)

	// Container tools (padded label to align with "Containers:")
	containerDots := m.renderPlannedContainerDots()
	containerToolsStr := zone.Mark("res-ctools", white.Render(fmt.Sprintf("%-*s", col5LabelWidth, "Tools:"))+containerDots)
	col5Line3 := padTo(containerToolsStr, col5Width)

	line3 := col1Line3 + sep + col2Line3 + sep + col3Line3 + sep + col4Line3 + sep + col5Line3
	line3 = padTo(line3, contentWidth)
	result.WriteString(Styles.Border.Render("│") + " " + line3 + " " + Styles.Border.Render("│") + "\n")

	// Footer: └─────────────────────────────────────┘
	footerBorderLen := m.width - 2
	if footerBorderLen < 3 {
		footerBorderLen = 3
	}
	result.WriteString("└" + Styles.Border.Render(strings.Repeat("─", footerBorderLen)) + "┘")

	return result.String()
}

// renderSelectedPane renders a context-aware information display:
// - UoW details when hovering over component tabs
// - Help text when hovering over Resources pane elements
// Always renders (for stable layout) but with empty content when nothing is hovered.
// Returns empty string if scheduler is not active.
func (m Model) renderSelectedPane() string {
	// Only show when scheduler is active (same condition as Resources pane)
	hasScheduler := false
	for _, lock := range m.locks {
		if lock.Name == "component-scheduler" && lock.Capacity > 0 {
			hasScheduler = true
			break
		}
	}
	if !hasScheduler {
		return ""
	}

	var result strings.Builder

	// Helper to pad string to fixed width
	padTo := func(s string, width int) string {
		visWidth := lipgloss.Width(s)
		if visWidth >= width {
			return s
		}
		return s + strings.Repeat(" ", width-visWidth)
	}

	// Header: ┌─ Selected ─────────────────────────┐
	title := "Selected"
	headerLeft := "┌─ " + Styles.Dim.Render(title) + " "
	headerBorderLen := m.width - lipgloss.Width(headerLeft) - 1
	if headerBorderLen < 3 {
		headerBorderLen = 3
	}
	result.WriteString(headerLeft + Styles.Border.Render(strings.Repeat("─", headerBorderLen)) + "┐\n")

	// Content width (inside borders)
	contentWidth := m.width - 4

	var content string

	// Determine what to show based on hoveredZone
	if strings.HasPrefix(m.hoveredZone, "tab:") || m.hoveredTab != "" {
		// Show UoW details for hovered/selected component tab
		activeComponent := m.hoveredTab
		if activeComponent == "" {
			activeComponent = m.getEffectiveActiveTab()
		}
		content = m.renderSelectedUoW(activeComponent, contentWidth)
	} else if helpText, ok := HelpTextMap[m.hoveredZone]; ok {
		// Show help text for hovered resource element
		content = m.renderSelectedHelp(m.hoveredZone, helpText, contentWidth)
	}
	// else: empty content (no hover)

	content = padTo(content, contentWidth)
	result.WriteString(Styles.Border.Render("│") + " " + content + " " + Styles.Border.Render("│") + "\n")

	// Footer: └─────────────────────────────────────┘
	footerBorderLen := m.width - 2
	if footerBorderLen < 3 {
		footerBorderLen = 3
	}
	result.WriteString("└" + Styles.Border.Render(strings.Repeat("─", footerBorderLen)) + "┘")

	return result.String()
}

// renderSelectedUoW renders UoW details for the selected/hovered component tab.
func (m Model) renderSelectedUoW(activeComponent string, contentWidth int) string {
	if activeComponent == "" {
		return ""
	}

	// Fixed column widths matching Resources pane
	const (
		col1Width = 38 // Unit
		col2Width = 24 // Runner/Tool
		col3Width = 20 // Status
	)

	// Helper to truncate with ellipsis
	truncate := func(s string, maxLen int) string {
		if len(s) <= maxLen {
			return s
		}
		if maxLen <= 3 {
			return s[:maxLen]
		}
		return s[:maxLen-3] + "..."
	}

	// Helper to pad string to fixed width
	padTo := func(s string, width int) string {
		visWidth := lipgloss.Width(s)
		if visWidth >= width {
			return s
		}
		return s + strings.Repeat(" ", width-visWidth)
	}

	// Semantic styles for status display
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))   // White for labels
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))  // Gold/amber for active unit
	pendingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // Gray for pending
	runningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))  // Blue for running
	completeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Green for complete
	cachedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))  // Purple for cached
	failedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))  // Red for failed
	sep := Styles.Dim.Render(" │ ")

	var col1, col2, col3 string

	// Get status of the active component
	var statusStr string
	var statusStyle lipgloss.Style
	if state, exists := m.uowStates[activeComponent]; exists {
		switch state.Status {
		case UoWPending:
			statusStr = "pending"
			statusStyle = pendingStyle
		case UoWRunning:
			statusStr = "running"
			statusStyle = runningStyle
			if !state.StartTime.IsZero() {
				dur := time.Since(state.StartTime)
				statusStr = fmt.Sprintf("running %s", formatElapsed(dur))
			}
		case UoWComplete:
			statusStr = "done"
			statusStyle = completeStyle
			if !state.StartTime.IsZero() && !state.EndTime.IsZero() {
				dur := state.EndTime.Sub(state.StartTime)
				statusStr = fmt.Sprintf("done %s", formatElapsed(dur))
			}
		case UoWSkipped:
			statusStr = "cached"
			statusStyle = cachedStyle
		case UoWFailed:
			statusStr = "failed"
			statusStyle = failedStyle
			if state.ExitCode != 0 {
				statusStr = fmt.Sprintf("failed (exit %d)", state.ExitCode)
			}
		}
	}

	// Parse component string into parts
	parts := strings.Split(activeComponent, ":")

	// Determine tool label based on operation type
	toolLabel := "Tool"
	switch m.runPhaseName {
	case "building", "Building":
		toolLabel = "Builder"
	case "testing", "Testing":
		toolLabel = "Runner"
	case "linting", "Linting":
		toolLabel = "Linter"
	case "scanning", "Scanning":
		toolLabel = "Scanner"
	}

	// Extract unit name and tool name
	var unitName, toolName string
	if len(parts) >= 2 {
		if len(parts) > 2 {
			unitName = parts[0] + ":" + strings.Join(parts[1:len(parts)-1], ":")
		} else {
			unitName = activeComponent
		}
		toolName = parts[len(parts)-1]
	} else {
		unitName = activeComponent
		toolName = ""
	}

	// Col1: Unit
	col1 = labelStyle.Render("Unit:") + activeStyle.Render(truncate(unitName, col1Width-5))

	// Col2: Runner/Tool
	if toolName != "" {
		col2 = labelStyle.Render(toolLabel+":") + activeStyle.Render(truncate(toolName, col2Width-len(toolLabel)-1))
	}

	// Col3: Status
	if statusStr != "" {
		col3 = labelStyle.Render("Status:") + statusStyle.Render(truncate(statusStr, col3Width-7))
	}

	col1 = padTo(col1, col1Width)
	col2 = padTo(col2, col2Width)
	col3 = padTo(col3, col3Width)

	return col1 + sep + col2 + sep + col3
}

// renderSelectedHelp renders help text for a hovered resource element.
func (m Model) renderSelectedHelp(zoneID, helpText string, contentWidth int) string {
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	// Map zone IDs to friendly element names
	elementNames := map[string]string{
		"res-timer":        "Timer",
		"res-cpu":          "CPU",
		"res-mem":          "Memory",
		"res-dmem":         "Docker Mem",
		"res-host":         "Host",
		"res-docker":       "Docker",
		"res-host-weight":  "Host Weight",
		"res-docker-weight": "Docker Weight",
		"res-counters":     "Counters",
		"res-jobs":         "Containers",
		"res-native":       "Native",
		"res-ctools":       "Tools",
		"freeze-button":    "Freeze",
	}

	elementName := elementNames[zoneID]
	if elementName == "" {
		// Fallback: derive from zone ID
		elementName = strings.TrimPrefix(zoneID, "res-")
		if len(elementName) > 0 {
			elementName = strings.ToUpper(elementName[:1]) + elementName[1:]
		}
	}

	return labelStyle.Render(elementName+": ") + helpStyle.Render(helpText)
}

// renderComponentsPane renders the Components pane containing the tab bar.
// This provides a bordered container for component tabs similar to other panes.
func (m Model) renderComponentsPane(tabs []*UoWState) string {
	if len(tabs) == 0 {
		return ""
	}

	var result strings.Builder

	// Count running, completed, skipped, failed for header summary
	var running, completed, skipped, failed int
	for _, tab := range tabs {
		switch tab.Status {
		case UoWRunning:
			running++
		case UoWComplete:
			completed++
		case UoWSkipped:
			skipped++
		case UoWFailed:
			failed++
		}
	}

	// Header: ┌─ Units (of work) ─────────────────┐
	title := "Units (of work)"

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
func (m Model) renderTabBarContent(tabs []*UoWState) string {
	const tabWidth = 20 // Fixed width for uniform tabs

	// Build list of all tab entries (modules first, then "All" at the end)
	var allTabs []tabEntry

	// Add module tabs first
	for _, state := range tabs {
		icon := m.statusIcon(state.Status)

		// Use DisplayName for tab labels
		label := state.DisplayName

		// Weight as circled digit or suffix based on mode (with trailing space for background coverage)
		var weightStr string
		if m.asciiMode {
			weightStr = fmt.Sprintf(" w%d ", state.Weight)
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
		case UoWPending:
			style = Styles.TabPending
		case UoWRunning:
			style = Styles.TabRunning
		case UoWComplete:
			style = Styles.TabComplete
		case UoWSkipped:
			style = Styles.TabSkipped
		case UoWFailed:
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

// renderCPUDots returns colored dots representing CPU usage normalized to stableLampsCount.
// Normalizes 8-64 cores to 16 lamps by grouping cores and averaging their usage.
// Color progression: green (low) → yellow → orange → red (high)
// Uses cached metrics updated by UpdateCachedMetrics() to avoid blocking gopsutil calls.
func (m Model) renderCPUDots() string {
	// Use cached values (updated periodically by tick handler)
	perCore := m.cachedCPUPercent
	if len(perCore) == 0 {
		// No data yet - show placeholder
		dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		emptyChar := "○"
		if m.asciiMode {
			emptyChar = "o"
		}
		return dim.Render(strings.Repeat(emptyChar, stableLampsCount))
	}

	// Normalize cores to stableLampsCount lamps
	// Each lamp represents a range of cores (or fraction of a core)
	normalized := make([]float64, stableLampsCount)
	coreCount := len(perCore)

	if coreCount <= stableLampsCount {
		// Fewer cores than lamps: each core maps to multiple lamps
		// E.g., 8 cores → 16 lamps = each core controls 2 lamps (pairs)
		lampsPerCore := float64(stableLampsCount) / float64(coreCount)
		for i, pct := range perCore {
			startLamp := int(float64(i) * lampsPerCore)
			endLamp := int(float64(i+1) * lampsPerCore)
			if endLamp > stableLampsCount {
				endLamp = stableLampsCount
			}
			for j := startLamp; j < endLamp; j++ {
				normalized[j] = pct
			}
		}
	} else {
		// More cores than lamps: group cores and average
		coresPerLamp := float64(coreCount) / float64(stableLampsCount)
		for i := 0; i < stableLampsCount; i++ {
			startCore := int(float64(i) * coresPerLamp)
			endCore := int(float64(i+1) * coresPerLamp)
			if endCore > coreCount {
				endCore = coreCount
			}
			sum := 0.0
			count := 0
			for j := startCore; j < endCore; j++ {
				sum += perCore[j]
				count++
			}
			if count > 0 {
				normalized[i] = sum / float64(count)
			}
		}
	}

	// Colors for thresholds
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("71"))    // Active (low usage)
	dimGreen := lipgloss.NewStyle().Foreground(lipgloss.Color("22")) // Idle
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))  // Medium
	orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))  // High
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))     // Maxed

	filledChar, emptyChar := "●", "○"
	if m.asciiMode {
		filledChar, emptyChar = "*", "o"
	}

	var dots strings.Builder
	for _, pct := range normalized {
		switch {
		case pct < 5:
			dots.WriteString(dimGreen.Render(emptyChar)) // Idle - unfilled green
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

// renderMemDots returns 16 dots representing memory usage.
// Color progression uses non-uniform zones: more green (safe), fewer red (critical).
// Unfilled dots use the color of their threshold zone.
// Uses cached metrics updated by UpdateCachedMetrics() to avoid blocking gopsutil calls.
func (m Model) renderMemDots() string {
	const totalDots = stableLampsCount

	// Use cached value (updated periodically by tick handler)
	usedPct := m.cachedMemPercent
	if usedPct == 0 && m.lastMetricsUpdate.IsZero() {
		// No metrics yet
		return strings.Repeat("-", totalDots)
	}
	// Map percentage to dots: 0% = 0 active, 100% = 16 active
	activeDots := int(usedPct / 100.0 * float64(totalDots))
	if activeDots > totalDots {
		activeDots = totalDots
	}
	if activeDots < 0 {
		activeDots = 0
	}

	// Choose characters based on mode
	filledChar, emptyChar := "●", "○"
	if m.asciiMode {
		filledChar, emptyChar = "*", "o"
	}

	// Colors for each threshold zone (non-uniform distribution)
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("71"))      // safe (filled)
	dimGreen := lipgloss.NewStyle().Foreground(lipgloss.Color("22"))   // safe (unfilled)
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))    // warning (filled)
	dimYellow := lipgloss.NewStyle().Foreground(lipgloss.Color("58"))  // warning (unfilled)
	orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))    // elevated (filled)
	dimOrange := lipgloss.NewStyle().Foreground(lipgloss.Color("130")) // elevated (unfilled)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))       // critical (filled)
	dimRed := lipgloss.NewStyle().Foreground(lipgloss.Color("52"))     // critical (unfilled)

	// Get zone boundaries (6 green, 5 yellow, 3 orange, 2 red for 16 lamps)
	greenEnd, yellowEnd, orangeEnd := pressureLampZones(totalDots)

	var dots strings.Builder
	for i := 0; i < totalDots; i++ {
		isFilled := i < activeDots
		char := emptyChar
		if isFilled {
			char = filledChar
		}

		// Color based on position (non-uniform zones)
		switch {
		case i < greenEnd:
			if isFilled {
				dots.WriteString(green.Render(char))
			} else {
				dots.WriteString(dimGreen.Render(char))
			}
		case i < yellowEnd:
			if isFilled {
				dots.WriteString(yellow.Render(char))
			} else {
				dots.WriteString(dimYellow.Render(char))
			}
		case i < orangeEnd:
			if isFilled {
				dots.WriteString(orange.Render(char))
			} else {
				dots.WriteString(dimOrange.Render(char))
			}
		default:
			if isFilled {
				dots.WriteString(red.Render(char))
			} else {
				dots.WriteString(dimRed.Render(char))
			}
		}
	}

	return dots.String()
}

// renderSchedulerPressureLamps renders scheduler pressure lamps with green/yellow/orange/red gradient.
// Shows current pressure utilization: running / pressureTarget as a percentage.
// Uses non-uniform color zones: more green (safe), fewer red (critical).
// Filled lamps = current usage, unfilled lamps = available headroom.
func (m Model) renderSchedulerPressureLamps(running, pressureTarget int) string {
	const totalDots = stableLampsCount

	// Handle edge cases
	if pressureTarget <= 0 {
		dim := lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
		emptyChar := "○"
		if m.asciiMode {
			emptyChar = "o"
		}
		return dim.Render(strings.Repeat(emptyChar, totalDots))
	}

	// Calculate usage percentage and map to dots
	usagePct := float64(running) / float64(pressureTarget) * 100.0
	activeDots := int(usagePct / 100.0 * float64(totalDots))
	if activeDots > totalDots {
		activeDots = totalDots
	}
	if activeDots < 0 {
		activeDots = 0
	}

	// Choose characters based on mode
	filledChar, emptyChar := "●", "○"
	if m.asciiMode {
		filledChar, emptyChar = "*", "o"
	}

	// Colors for each threshold zone (non-uniform distribution)
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("71"))      // safe (filled)
	dimGreen := lipgloss.NewStyle().Foreground(lipgloss.Color("22"))   // safe (unfilled)
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))    // warning (filled)
	dimYellow := lipgloss.NewStyle().Foreground(lipgloss.Color("58"))  // warning (unfilled)
	orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))    // elevated (filled)
	dimOrange := lipgloss.NewStyle().Foreground(lipgloss.Color("130")) // elevated (unfilled)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))       // critical (filled)
	dimRed := lipgloss.NewStyle().Foreground(lipgloss.Color("52"))     // critical (unfilled)

	// Get zone boundaries (6 green, 5 yellow, 3 orange, 2 red for 16 lamps)
	greenEnd, yellowEnd, orangeEnd := pressureLampZones(totalDots)

	var dots strings.Builder
	for i := 0; i < totalDots; i++ {
		isFilled := i < activeDots
		char := emptyChar
		if isFilled {
			char = filledChar
		}

		// Color based on position (non-uniform zones)
		switch {
		case i < greenEnd:
			if isFilled {
				dots.WriteString(green.Render(char))
			} else {
				dots.WriteString(dimGreen.Render(char))
			}
		case i < yellowEnd:
			if isFilled {
				dots.WriteString(yellow.Render(char))
			} else {
				dots.WriteString(dimYellow.Render(char))
			}
		case i < orangeEnd:
			if isFilled {
				dots.WriteString(orange.Render(char))
			} else {
				dots.WriteString(dimOrange.Render(char))
			}
		default:
			if isFilled {
				dots.WriteString(red.Render(char))
			} else {
				dots.WriteString(dimRed.Render(char))
			}
		}
	}

	return dots.String()
}

// renderDockerMemDots renders Docker memory pool usage as colored dots.
// Always shows 16 dots - filled based on usage percentage.
// Uses non-uniform color zones: more green (safe), fewer red (critical).
// Unfilled dots use the color of their threshold zone.
func (m Model) renderDockerMemDots() string {
	const totalDots = stableLampsCount

	// Choose characters based on mode
	filledChar, emptyChar := "●", "○"
	if m.asciiMode {
		filledChar, emptyChar = "*", "o"
	}

	// Use cached Docker memory percentage (0 if not tracked)
	usedPct := m.cachedDockerMemPercent

	// Map percentage to dots: 0% = 0 active, 100% = 16 active
	activeDots := int(usedPct / 100.0 * float64(totalDots))
	if activeDots > totalDots {
		activeDots = totalDots
	}

	// Colors for each threshold zone (non-uniform distribution)
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("71"))      // safe (filled)
	dimGreen := lipgloss.NewStyle().Foreground(lipgloss.Color("22"))   // safe (unfilled)
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))    // warning (filled)
	dimYellow := lipgloss.NewStyle().Foreground(lipgloss.Color("58"))  // warning (unfilled)
	orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))    // elevated (filled)
	dimOrange := lipgloss.NewStyle().Foreground(lipgloss.Color("130")) // elevated (unfilled)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))       // critical (filled)
	dimRed := lipgloss.NewStyle().Foreground(lipgloss.Color("52"))     // critical (unfilled)

	// Get zone boundaries (6 green, 5 yellow, 3 orange, 2 red for 16 lamps)
	greenEnd, yellowEnd, orangeEnd := pressureLampZones(totalDots)

	var dots strings.Builder
	for i := 0; i < totalDots; i++ {
		isFilled := i < activeDots
		char := emptyChar
		if isFilled {
			char = filledChar
		}

		// Color based on position (non-uniform zones)
		switch {
		case i < greenEnd:
			if isFilled {
				dots.WriteString(green.Render(char))
			} else {
				dots.WriteString(dimGreen.Render(char))
			}
		case i < yellowEnd:
			if isFilled {
				dots.WriteString(yellow.Render(char))
			} else {
				dots.WriteString(dimYellow.Render(char))
			}
		case i < orangeEnd:
			if isFilled {
				dots.WriteString(orange.Render(char))
			} else {
				dots.WriteString(dimOrange.Render(char))
			}
		default:
			if isFilled {
				dots.WriteString(red.Render(char))
			} else {
				dots.WriteString(dimRed.Render(char))
			}
		}
	}

	return dots.String()
}

// renderRunningContainersDots renders container instance lamps.
// INVARIANT: Lamp index maps directly to container instance index.
//   - Lamp 0 = container instance 0 (first spawned)
//   - Lamp 1 = container instance 1 (second spawned)
//   - etc.
// FIFO semantics: containers complete in spawn order, so completed containers
// are always at lower indices than running containers.
// When container N completes, lamp N goes from lit to dim.
//
// Visual layout (16 lamps):
//   ○○○○●●●●○○○○○○○○
//   └──┘└──┘└────────┘
//   done run  unused
func (m Model) renderRunningContainersDots() string {
	const maxDots = stableLampsCount

	// 3 distinct colors:
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))    // Running = bright cyan filled
	dimCyan := lipgloss.NewStyle().Foreground(lipgloss.Color("30")) // Completed = dim cyan unfilled
	black := lipgloss.NewStyle().Foreground(lipgloss.Color("236"))  // Never-used = near-black

	filledChar, emptyChar := "●", "○"
	if m.asciiMode {
		filledChar, emptyChar = "*", "o"
	}

	total := m.totalContainerCount
	running := m.runningContainerCount
	completed := total - running

	// Clamp to max display
	if total > maxDots {
		total = maxDots
	}
	if running > total {
		running = total
	}
	if completed < 0 {
		completed = 0
	}

	// If no containers have run, show all black (never-used)
	if total == 0 {
		return black.Render(strings.Repeat(emptyChar, maxDots))
	}

	var dots strings.Builder

	// Render lamps in index order (0 to maxDots-1)
	// Lamp i represents container instance i
	for i := 0; i < maxDots; i++ {
		if i < completed {
			// Container i has completed (FIFO: first spawned = first completed)
			dots.WriteString(dimCyan.Render(emptyChar))
		} else if i < total {
			// Container i is still running
			dots.WriteString(cyan.Render(filledChar))
		} else {
			// Slot i has never been used
			dots.WriteString(black.Render(emptyChar))
		}
	}

	return dots.String()
}

// renderPlannedContainerDots renders container lamps with fixed width (16 max).
// Shows containers as they are used: filled when running, unfilled when stopped.
// Order: left=old (idle), middle=active, right=unused
// Unused slots are black/dim. This tracks the count of individual containers used.
func (m Model) renderPlannedContainerDots() string {
	const fixedLamps = stableLampsCount

	sharpBlue := lipgloss.NewStyle().Foreground(lipgloss.Color("39")) // Active container (filled)
	lightBlue := lipgloss.NewStyle().Foreground(lipgloss.Color("75")) // Inactive container (unfilled blue)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("236"))      // Empty/unused slot (very dark)

	filledChar := "●"
	emptyChar := "○"
	if m.asciiMode {
		filledChar = "*"
		emptyChar = "o"
	}

	// Calculate active container tools from running UoWs
	activeTools := m.getActiveContainerTools()

	// Separate seen containers into inactive (idle) and active
	var inactiveCount, activeCount int
	for _, container := range m.seenContainerTools {
		if activeTools[container] {
			activeCount++
		} else {
			inactiveCount++
		}
	}

	// Clamp to max lamps
	totalSeen := inactiveCount + activeCount
	if totalSeen > fixedLamps {
		// Proportionally reduce if over limit
		if inactiveCount > fixedLamps {
			inactiveCount = fixedLamps
			activeCount = 0
		} else {
			activeCount = fixedLamps - inactiveCount
		}
	}

	var dots strings.Builder
	lampCount := 0

	// 1. Inactive/idle containers first (left) - old
	for i := 0; i < inactiveCount && lampCount < fixedLamps; i++ {
		dots.WriteString(lightBlue.Render(emptyChar))
		lampCount++
	}

	// 2. Active containers (middle) - currently running
	for i := 0; i < activeCount && lampCount < fixedLamps; i++ {
		dots.WriteString(sharpBlue.Render(filledChar))
		lampCount++
	}

	// 3. Unused slots (right) - never used
	for lampCount < fixedLamps {
		dots.WriteString(dim.Render(emptyChar))
		lampCount++
	}

	return dots.String()
}

// getActiveContainerTools returns a set of container tool names that have running UoWs.
func (m Model) getActiveContainerTools() map[string]bool {
	active := make(map[string]bool)
	for _, moniker := range m.uowOrder {
		state := m.uowStates[moniker]
		if state != nil && state.Status == UoWRunning {
			// Extract tool from moniker (format: context:module:component:tool)
			tool := extractToolFromMoniker(moniker)
			if tool != "" {
				// Check if this tool is a container tool
				for _, ct := range m.plannedContainerTools {
					if ct == tool {
						active[tool] = true
						break
					}
				}
			}
		}
	}
	return active
}

// extractToolFromMoniker extracts the tool name from a moniker.
// Moniker format: context:module:component:tool (e.g., "local:books:api:go")
func extractToolFromMoniker(moniker string) string {
	parts := strings.Split(moniker, ":")
	if len(parts) >= 4 {
		return parts[len(parts)-1] // Last segment is the tool
	}
	return ""
}

// renderActiveContainerDots returns colored dots for each running job.
// Each dot represents an active job, colored by its weight/pressure:
// Green = weight 1-2 (light), Yellow = weight 3-4, Orange = weight 5-6, Red = weight 7+
func (m Model) renderActiveContainerDots() string {
	// Collect running modules with their weights
	var runningJobs []int // Weights of running jobs
	for _, moniker := range m.uowOrder {
		state := m.uowStates[moniker]
		if state != nil && state.Status == UoWRunning {
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

// renderToolsDots renders system tool lamps with fixed width (16 max).
// Shows tools as they are used: filled when running, unfilled when stopped.
// Order: left=old (idle), middle=active, right=unused
// Unused slots are black/dim. This tracks the count of individual tools used.
func (m Model) renderToolsDots() string {
	const fixedLamps = stableLampsCount

	orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))    // Active system tool (filled)
	dimOrange := lipgloss.NewStyle().Foreground(lipgloss.Color("130")) // Inactive system tool (unfilled)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("236"))       // Empty/unused slot (very dark)

	filledChar := "●"
	emptyChar := "○"
	if m.asciiMode {
		filledChar = "*"
		emptyChar = "o"
	}

	// Calculate active system tools from running UoWs
	activeSystem := m.getActiveSystemTools()

	// Separate seen tools into inactive (idle) and active
	var inactiveCount, activeCount int
	for _, tool := range m.seenSystemTools {
		if activeSystem[tool] {
			activeCount++
		} else {
			inactiveCount++
		}
	}

	// Clamp to max lamps
	totalSeen := inactiveCount + activeCount
	if totalSeen > fixedLamps {
		// Proportionally reduce if over limit
		if inactiveCount > fixedLamps {
			inactiveCount = fixedLamps
			activeCount = 0
		} else {
			activeCount = fixedLamps - inactiveCount
		}
	}

	var dots strings.Builder
	lampCount := 0

	// 1. Inactive/idle tools first (left) - old
	for i := 0; i < inactiveCount && lampCount < fixedLamps; i++ {
		dots.WriteString(dimOrange.Render(emptyChar))
		lampCount++
	}

	// 2. Active tools (middle) - currently running
	for i := 0; i < activeCount && lampCount < fixedLamps; i++ {
		dots.WriteString(orange.Render(filledChar))
		lampCount++
	}

	// 3. Unused slots (right) - never used
	for lampCount < fixedLamps {
		dots.WriteString(dim.Render(emptyChar))
		lampCount++
	}

	return dots.String()
}

// getActiveSystemTools returns a set of system tool names that have running UoWs.
func (m Model) getActiveSystemTools() map[string]bool {
	active := make(map[string]bool)
	for _, moniker := range m.uowOrder {
		state := m.uowStates[moniker]
		if state != nil && state.Status == UoWRunning {
			// Extract tool from moniker (format: context:module:component:tool)
			tool := extractToolFromMoniker(moniker)
			if tool != "" {
				// Check if this tool is a system tool
				for _, st := range m.plannedSystemTools {
					if st == tool {
						active[tool] = true
						break
					}
				}
			}
		}
	}
	return active
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

// getModuleName extracts the module name from a full moniker (module:component).
// Returns the part before the first colon, or the full string if no colon is present.
func getModuleName(moniker string) string {
	if idx := strings.Index(moniker, ":"); idx >= 0 {
		return moniker[:idx]
	}
	return moniker
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

// stripToolTypeSuffix removes :system or :container suffix from tool names.
// Type is indicated by color in the TUI, not by suffix in the name.
func stripToolTypeSuffix(toolID string) string {
	if idx := strings.LastIndex(toolID, ":"); idx != -1 {
		suffix := toolID[idx+1:]
		if suffix == "system" || suffix == "container" {
			return toolID[:idx]
		}
	}
	return toolID
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
		if moduleBuffer := m.GetActiveUoWBuffer(); moduleBuffer != nil {
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
func (m Model) renderModuleResultsTable(tabs []*UoWState) string {
	if len(tabs) == 0 {
		return ""
	}

	var b strings.Builder

	// Find max display name length for alignment
	maxNameLen := 20
	for _, state := range tabs {
		if len(state.DisplayName) > maxNameLen {
			maxNameLen = len(state.DisplayName)
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
		case UoWComplete:
			icon = "  ✓  "
			passed++
		case UoWSkipped:
			icon = "  ⏭  "
			skipped++
		case UoWFailed:
			icon = "  ✗  "
			failed++
		case UoWRunning:
			icon = "  ▶  "
			running++
		default:
			icon = "  ◦  "
			pending++
		}

		// Display name (truncate if needed, rune-safe)
		name := state.DisplayName
		if len(name) > maxNameLen {
			runes := []rune(name)
			if len(runes) > maxNameLen-1 {
				name = string(runes[:maxNameLen-1]) + "…"
			}
		}

		// Duration
		var duration string
		if state.Status == UoWRunning {
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
