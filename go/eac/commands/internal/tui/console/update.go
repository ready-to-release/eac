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
		for _, pane := range m.panes {
			if pane != nil && !pane.autoScroll {
				// If user was scrolled up, try to maintain position, but cap at new max
				initH, runH, summaryH := m.calculatePaneHeights()
				paneHeight := initH
				switch pane.Phase {
				case PhaseRun:
					paneHeight = runH
				case PhaseSummary:
					paneHeight = summaryH
				}
				pane.UpdateMaxScroll(paneHeight)
			}
		}
		return m, nil

	case lineMsg:
		line := Line(msg)
		// Write to active phase's buffer
		pane := m.panes[m.activePhase]
		pane.Buffer.Push(line)
		// If pane is scrolled up (not auto-scrolling), increment offset to keep view locked
		if !pane.autoScroll && pane.scrollOffset > 0 {
			pane.scrollOffset++
		}

		// Route to per-module buffer if this is a module output line
		// (Source is the module moniker, not "system" or phase name)
		if line.Source != "" && line.Source != "system" && m.activePhase == PhaseRun {
			// Check if this is a known module (has a state)
			if state, exists := m.moduleStates[line.Source]; exists {
				state.Buffer.Push(line)
			} else {
				// Create module state on first output (module started)
				// Weight defaults to 1 if not set via ModuleStartMsg
				state := m.GetOrCreateModuleState(line.Source, 0)
				state.Buffer.Push(line)
			}
		}

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
		if m.panes[msg.Phase] != nil {
			pane := m.panes[msg.Phase]
			pane.Buffer.Push(msg.Line)
			// If pane is scrolled up (not auto-scrolling), increment offset to keep view locked
			if !pane.autoScroll && pane.scrollOffset > 0 {
				pane.scrollOffset++
			}
		}
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
			// Mark current phase as complete or failed based on success
			if m.panes[m.activePhase].Status == PhaseActive {
				if msg.Data != nil && !msg.Data.Success {
					m.panes[m.activePhase].Status = PhaseFailed
				} else {
					m.panes[m.activePhase].Status = PhaseComplete
				}
				m.panes[m.activePhase].EndTime = time.Now()
			}
			// Activate Summary pane
			m.activePhase = PhaseSummary
			m.panes[PhaseSummary].Status = PhaseActive
			m.panes[PhaseSummary].StartTime = time.Now()
		}
		// Mark summary pane as complete/failed and quit immediately
		if msg.Data != nil && !msg.Data.Success {
			m.panes[PhaseSummary].Status = PhaseFailed
		} else {
			m.panes[PhaseSummary].Status = PhaseComplete
		}
		m.panes[PhaseSummary].EndTime = time.Now()
		m.quitting = true
		return m, tea.Quit

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
			switch msg.Status {
			case PhaseActive:
				// Mark previous phase as complete if it was active
				if m.activePhase != msg.Phase && m.activePhase < Phase(len(m.panes)) && m.panes[m.activePhase].Status == PhaseActive {
					m.panes[m.activePhase].Status = PhaseComplete
					m.panes[m.activePhase].EndTime = time.Now()
				}
				m.panes[msg.Phase].StartTime = time.Now()
				m.activePhase = msg.Phase
			case PhaseComplete, PhaseFailed:
				m.panes[msg.Phase].EndTime = time.Now()
			}
		}
		return m, nil

	case statusMsg:
		status := Status(msg)

		// Build sets for comparison
		oldRunningSet := make(map[string]bool)
		for _, moniker := range m.running {
			oldRunningSet[moniker] = true
		}
		newRunningSet := make(map[string]bool)
		for _, moniker := range status.Running {
			newRunningSet[moniker] = true
		}

		// Mark NEW modules as running (appeared in status.Running but weren't before)
		for _, moniker := range status.Running {
			if !oldRunningSet[moniker] {
				// New running module - create state if needed and mark as running
				state := m.GetOrCreateModuleState(moniker, 0)
				state.Status = ModuleRunning
				state.StartTime = time.Now()
			}
		}

		// Mark COMPLETED modules (were running but now gone)
		// Only update if still in Running status - ModuleCompleteMsg may have already set the real status
		for _, moniker := range m.running {
			if !newRunningSet[moniker] {
				// This module was running and is now gone
				if state, exists := m.moduleStates[moniker]; exists {
					// Only set to complete if still running (ModuleCompleteMsg sets actual status with exit code)
					if state.Status == ModuleRunning {
						state.Status = ModuleComplete
					}
					if state.EndTime.IsZero() {
						state.EndTime = time.Now()
					}
				}
			}
		}

		m.running = status.Running
		m.completed = status.Completed
		m.total = status.Total
		m.layer = status.Layer
		m.totalLayers = status.TotalLayers

		// Update lock tracking info
		m.locks = status.Locks

		if !m.statusDone {
			return m, m.listenForStatus()
		}
		return m, nil

	case tickMsg:
		// Clean up decayed tabs on each tick
		m.CleanupDecayedTabs()
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

		// Mark module as complete in tab tracking
		exitCode := msg.ExitCode
		m.MarkModuleComplete(msg.Moniker, exitCode)
		return m, nil

	case ModuleStartMsg:
		// Create module state when module starts (pending state)
		m.GetOrCreateModuleState(msg.Moniker, msg.Weight)
		return m, nil

	case ModuleRunningMsg:
		// Mark module as running (slot acquired)
		m.MarkModuleRunning(msg.Moniker)
		return m, nil

	case ModuleCompleteMsg:
		// Mark module as complete (alternative to completedMsg)
		m.MarkModuleComplete(msg.Moniker, msg.ExitCode)
		return m, nil

	case TabSelectMsg:
		// Switch to selected tab
		m.SetActiveTab(msg.Moniker)
		// Reset scroll to bottom when switching tabs
		if m.panes[PhaseRun] != nil {
			m.panes[PhaseRun].scrollOffset = 0
			m.panes[PhaseRun].autoScroll = true
		}
		return m, nil

	case TabDecayMsg:
		// Clean up decayed tabs
		m.CleanupDecayedTabs()
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
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

	// Tab navigation shortcuts
	case "tab":
		// Cycle to next tab
		m.cycleTab(1)
	case "shift+tab":
		// Cycle to previous tab
		m.cycleTab(-1)
	case "a":
		// Switch to "All" (aggregate view)
		m.SetActiveTab("")
		if m.panes[PhaseRun] != nil {
			m.panes[PhaseRun].scrollOffset = 0
			m.panes[PhaseRun].autoScroll = true
		}
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Direct tab selection (1-9)
		idx := int(msg.String()[0] - '1') // Convert "1" to 0, "2" to 1, etc.
		tabs := m.GetVisibleTabs()
		if idx < len(tabs) {
			m.SetActiveTab(tabs[idx].Moniker)
			if m.panes[PhaseRun] != nil {
				m.panes[PhaseRun].scrollOffset = 0
				m.panes[PhaseRun].autoScroll = true
			}
		}
	case "0":
		// "0" = switch to aggregate view
		m.SetActiveTab("")
		if m.panes[PhaseRun] != nil {
			m.panes[PhaseRun].scrollOffset = 0
			m.panes[PhaseRun].autoScroll = true
		}
	}
	return m, nil
}

// cycleTab cycles through tabs in the given direction (+1 = next, -1 = prev).
func (m *Model) cycleTab(direction int) {
	tabs := m.GetVisibleTabs()
	if len(tabs) == 0 {
		return
	}

	// Build list: [aggregate view] + [module tabs]
	allTabs := make([]string, 0, len(tabs)+1)
	allTabs = append(allTabs, "") // Aggregate view
	for _, t := range tabs {
		allTabs = append(allTabs, t.Moniker)
	}

	// Find current index
	currentIdx := 0
	for i, moniker := range allTabs {
		if moniker == m.activeTab {
			currentIdx = i
			break
		}
	}

	// Calculate new index with wrap-around
	newIdx := (currentIdx + direction + len(allTabs)) % len(allTabs)
	m.SetActiveTab(allTabs[newIdx])

	// Reset scroll
	if m.panes[PhaseRun] != nil {
		m.panes[PhaseRun].scrollOffset = 0
		m.panes[PhaseRun].autoScroll = true
	}
}

// handleMouse handles mouse events for pane scrolling and tab clicks.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Mouse interactions:
	// - Scroll wheel → scroll panes (handled first to prevent any accidental tab switching)
	// - Click on tab bar → switch tabs
	// - Shift+Click → select text (standard terminal behavior, bypasses mouse mode)

	// Handle wheel events FIRST - scrolling should never change tabs
	if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
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
		switch paneIdx {
		case 1: // Run pane
			paneHeight = runH
		case 2: // Summary pane
			paneHeight = summaryH
		}

		// Determine which buffer is being displayed (for Run pane with active tab)
		buffer := pane.Buffer
		if paneIdx == 1 && m.activeTab != "" { // Run pane with module tab selected
			if moduleBuffer := m.GetActiveModuleBuffer(); moduleBuffer != nil {
				buffer = moduleBuffer
			}
		}

		// Update max scroll based on the ACTIVE buffer
		pane.UpdateMaxScrollForBuffer(buffer, paneHeight)

		// Scroll the pane - use Button to determine direction
		scrollAmount := 3 // Lines to scroll per wheel tick
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			pane.ScrollUp(scrollAmount)
		case tea.MouseButtonWheelDown:
			pane.ScrollDown(scrollAmount)
		}

		return m, nil
	}

	// Handle left mouse button click for tab selection (use Press for responsiveness)
	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
		// Check if click is on the tab bar (always shown when Run phase is active)
		if m.panes[PhaseRun].Status != PhasePending {
			tabBarStartY, tabBarEndY := m.getTabBarYRange()
			if msg.Y >= tabBarStartY && msg.Y <= tabBarEndY {
				// Click is on the tab bar - determine which tab was clicked
				row := msg.Y - tabBarStartY
				selectedTab := m.getTabAtPosition(msg.X, row)
				m.SetActiveTab(selectedTab)
				// Reset scroll
				if m.panes[PhaseRun] != nil {
					m.panes[PhaseRun].scrollOffset = 0
					m.panes[PhaseRun].autoScroll = true
				}
				return m, nil
			}
		}
	}

	return m, nil
}

// getTabBarYRange returns the Y coordinate range of the tab bar rows.
// Returns (startY, endY) where both are inclusive.
func (m Model) getTabBarYRange() (int, int) {
	// Layout (0-indexed Y coordinates):
	// 0: Init header
	// 1 to initH: Init content
	// initH+1: Init footer
	// initH+2: Resources header (if shown)
	// initH+3: Resources content
	// initH+4: Resources footer
	// initH+5 to initH+5+tabRows-1: Tab bar rows
	// Then: Run header, Run content, etc.
	initH, _, _ := m.calculatePaneHeights()

	// Check if resources pane is shown (scheduler active)
	resourcesShown := false
	for _, lock := range m.locks {
		if lock.Name == "component-scheduler" && lock.Capacity > 0 {
			resourcesShown = true
			break
		}
	}

	startY := initH + 2 // After init footer
	if resourcesShown {
		startY += 3 // Resources pane: header + content + footer
	}

	// Calculate number of tab rows
	tabs := m.GetVisibleTabs()
	numTabs := len(tabs) + 1 // +1 for "All" tab
	const tabsPerRow = 6
	numRows := (numTabs + tabsPerRow - 1) / tabsPerRow
	if numRows < 1 {
		numRows = 1
	}

	endY := startY + numRows - 1
	return startY, endY
}

// getTabAtPosition determines which tab was clicked based on X coordinate and row.
func (m Model) getTabAtPosition(x, row int) string {
	tabs := m.GetVisibleTabs()

	// Fixed tab layout: 20 chars per tab, 6 tabs per row
	// │tab1              │tab2              │...│All               │
	// Border (1) + tab content (20) + separator (1) per tab
	const tabWidth = 20
	const tabsPerRow = 6

	// Build all tabs (modules only, no "All" tab)
	allTabs := make([]string, 0, len(tabs))
	for _, state := range tabs {
		allTabs = append(allTabs, state.Moniker)
	}

	// Calculate which tabs are in this row
	startIdx := row * tabsPerRow
	if startIdx >= len(allTabs) {
		return m.activeTab // Click on empty row
	}

	// Position calculation within the row:
	// │ = 1 char border
	// Each tab is tabWidth chars
	// │ separator between tabs
	currentX := 1 // After left border

	endIdx := startIdx + tabsPerRow
	if endIdx > len(allTabs) {
		endIdx = len(allTabs)
	}

	for i := startIdx; i < endIdx; i++ {
		// Check if click is within this tab
		if x >= currentX && x < currentX+tabWidth {
			return allTabs[i]
		}
		currentX += tabWidth

		// Add separator width (except after last tab)
		if i < endIdx-1 {
			currentX++ // │ separator
		}
	}

	// Click not on any recognized tab - keep current selection
	return m.activeTab
}

// getPaneAtPosition determines which pane (0, 1, or 2) is at the given Y coordinate.
// Returns -1 if not over any pane content area.
// All panes are always visible with dynamic heights based on terminal size.
func (m Model) getPaneAtPosition(y int) int {
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
