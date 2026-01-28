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

// viewPanes renders the 3-pane layout with panes appearing progressively.
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

	// Render Resources pane between Init and Run (shows locks and system metrics)
	if resourcesPane := m.renderResourcesPane(); resourcesPane != "" {
		b.WriteString(resourcesPane)
		b.WriteString("\n")
	}

	// Render Run pane only if it actually started (not still pending)
	if m.panes[PhaseRun].Status != PhasePending {
		// Render Components pane (contains tab bar) between Resources and Run content
		tabs := m.GetVisibleTabs()
		componentsPane := m.renderComponentsPane(tabs)
		b.WriteString(componentsPane)
		b.WriteString("\n")

		// Count components pane rows (header + tab rows + footer)
		componentsPaneRows := strings.Count(componentsPane, "\n") + 1
		runH -= componentsPaneRows

		b.WriteString(m.renderPaneHeader(PhaseRun))
		b.WriteString("\n")
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

	// Content line: timer | pressure (weighted scheduler) | cpu dots | memory %
	// White style for labels
	white := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	// Timer with fixed width MM:SS format (white)
	elapsed := time.Since(m.startTime)
	mins := int(elapsed.Minutes())
	secs := int(elapsed.Seconds()) % 60
	timerStr := white.Render(fmt.Sprintf("%02d:%02d", mins, secs))

	pressureStr := white.Render("Pressure: ") + Styles.Dim.Render(fmt.Sprintf("%2d/%-2d", pressureUsed, pressureCap))

	// Per-core CPU usage as colored dots
	cpuStr := white.Render("CPU: ") + m.renderCPUDots()

	// System memory usage as dots (11 dots, red for used, dim for free)
	memStr := white.Render("Memory: ") + m.renderMemDots()

	// Active containers with pressure dots (colored by weight)
	containerDotsStr := m.renderActiveContainerDots()

	sep := Styles.Dim.Render(" │ ")
	content := timerStr + sep + pressureStr + sep + cpuStr + sep + memStr
	if containerDotsStr != "" {
		content += sep + white.Render("Jobs: ") + containerDotsStr
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

		// Determine labels based on operation type
		componentLabel := "Component"
		toolLabel := "Tool"
		switch m.runPhaseName {
		case "building", "Building":
			componentLabel = "Component"
			toolLabel = "Builder"
		case "testing", "Testing":
			componentLabel = "Test"
			toolLabel = "Runner"
		case "linting", "Linting":
			componentLabel = "Component"
			toolLabel = "Linter"
		case "scanning", "Scanning":
			componentLabel = "Component"
			toolLabel = "Scanner"
		}

		if len(parts) >= 3 {
			// Full format: module:component:handler
			module := parts[0]
			component := parts[1]
			handler := parts[len(parts)-1]
			// If there are more than 3 parts, component might have colons in path
			if len(parts) > 3 {
				component = strings.Join(parts[1:len(parts)-1], ":")
			}
			componentContent = white.Render("Module: ") + orange.Render(module) +
				dim.Render(" │ ") + white.Render(componentLabel+": ") + orange.Render(component) +
				dim.Render(" │ ") + white.Render(toolLabel+": ") + orange.Render(handler)
		} else if len(parts) == 2 {
			// Format: module:component
			componentContent = white.Render("Module: ") + orange.Render(parts[0]) +
				dim.Render(" │ ") + white.Render(componentLabel+": ") + orange.Render(parts[1])
		} else {
			// Fallback: just show as-is
			componentContent = white.Render("Active: ") + orange.Render(activeComponent)
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
