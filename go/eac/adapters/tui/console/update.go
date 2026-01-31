package console

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/ready-to-release/eac/go/eac/core/config"
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

		// If pane is scrolled up (not auto-scrolling), increment offset to keep view locked
		// Only increment if this line affects the currently viewed buffer
		if !pane.autoScroll && pane.scrollOffset > 0 {
			if m.activeTab == "" {
				// Viewing "All" - any line increments scroll
				pane.scrollOffset++
			} else if line.Source == m.activeTab {
				// Viewing specific module - only lines for that module increment scroll
				pane.scrollOffset++
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
			// Only increment if viewing the aggregate buffer (PhaseLineMsg goes to pane.Buffer, not module buffers)
			if !pane.autoScroll && pane.scrollOffset > 0 {
				// For Run pane, only increment if viewing aggregate (not a specific module tab)
				if msg.Phase != PhaseRun || m.activeTab == "" {
					pane.scrollOffset++
				}
			}
		}
		return m, nil

	case ResultLineMsg:
		// Write to results buffer
		m.resultsBuffer.Push(msg.Line)
		return m, nil

	case InitSummaryMsg:
		// Store init summary for structured rendering
		m.initSummary = msg.Summary

		// Pre-register all tabs from ExecutionTree
		// This ensures tabs are constant after init - only state changes, no additions
		if msg.Summary != nil && len(msg.Summary.ExecutionTree) > 0 {
			for _, layer := range msg.Summary.ExecutionTree {
				for _, module := range layer.Modules {
					for _, comp := range module.Components {
						// Reconstruct full moniker: module:component[:handler]
						// comp may be "component" or "component:handler"
						moniker := module.Name + ":" + comp
						// Pre-register with pending status and default weight
						m.GetOrCreateModuleState(moniker, 1)
					}
				}
			}

			// Calculate initial tab columns to fit all UoWs without scrolling
			m.tabColumns = m.calculateOptimalTabColumns()
		}

		// Pre-populate tool lamps from PlannedTools
		// All tools shown from start (inactive), light up when active
		if msg.Summary != nil && len(msg.Summary.PlannedTools) > 0 {
			for _, tool := range msg.Summary.PlannedTools {
				if tool.IsContainer {
					m.plannedContainers = append(m.plannedContainers, tool.Name)
				} else {
					m.plannedSystemTools = append(m.plannedSystemTools, tool.Name)
				}
			}
		}
		return m, nil

	case SummaryDataMsg:
		// Summary renderer has finished - store the data
		m.pendingSummaryData = msg.Data

		// Receiving summary means all runners are done - mark complete if not already
		if m.allRunnersCompleted.IsZero() {
			m.allRunnersCompleted = time.Now()
		}

		// If skipTUIDelay is set, exit immediately without user interaction tracking
		if m.skipTUIDelay {
			m.activateSummary()
			m.quitting = true
			return m, tea.Quit
		}

		// If user hasn't interacted, exit immediately (fast path - prioritize speed)
		if !m.userHasInteracted {
			m.exitRequested = true
			m.activateSummary()
			m.quitting = true
			return m, tea.Quit
		}

		// User has interacted - let them read the output
		// DON'T set exitRequested here - let the tick handler manage the 10-second countdown
		// and only set exitRequested when the timer expires
		return m, nil

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

		// Track EndTime for modules that were running but now gone
		// DON'T set status here - only ModuleCompleteMsg sets terminal status
		// (it has the exit code to determine Complete vs Skipped vs Failed)
		for _, moniker := range m.running {
			if !newRunningSet[moniker] {
				// This module was running and is now gone
				if state, exists := m.moduleStates[moniker]; exists {
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

		// Track when all runners completed
		if m.total > 0 && m.completed >= m.total && m.allRunnersCompleted.IsZero() {
			m.allRunnersCompleted = time.Now()
		}

		// Update lock tracking info
		m.locks = status.Locks

		// Update active tools (planned tools set at init, only active state changes)
		m.activeContainers = status.ActiveContainers
		m.activeSystemTools = status.ActiveSystemTools

		if !m.statusDone {
			return m, m.listenForStatus()
		}
		return m, nil

	case MarqueeTickMsg:
		// Animate hovered tab name scrolling (marquee effect)
		// Fast ticks (100ms) but only advance scroll every 4th tick for smooth slow movement
		if m.hoveredTab != "" {
			m.hoveredTabScroll++
			// Continue ticking while hovering
			return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
				return MarqueeTickMsg{}
			})
		}
		return m, nil

	case tickMsg:
		// Update cached system metrics (CPU/memory) - do this early so View() has fresh values
		// These gopsutil calls are expensive, so we cache them and only update periodically
		m.UpdateCachedMetrics()

		// Clean up decayed tabs on each tick
		m.CleanupDecayedTabs()

		// Check if auto-scroll should resume after timeout
		if m.panes[PhaseRun] != nil {
			m.panes[PhaseRun].CheckAutoScrollResume(config.TUIAutoScrollResumeTimeout())
		}

		// === THREE-THREAD MODEL ===
		//
		// Thread 1: User timer - countdown from last mouse interaction
		// Thread 2: Summary renderer - background builder producing pendingSummaryData
		// Thread 3: Exit decision - when to show "rendering summary" and quit
		//
		// Key insight: User interacting during execution = investigating output.
		// We delay exit to give them time to read, resetting the 10s timer on each interaction.

		// 1. User timer: track countdown from last mouse interaction
		userTimerExpired := false
		if m.userHasInteracted {
			elapsed := time.Since(m.lastUserInteraction)
			remaining := m.freezeTimeoutSecs - int(elapsed.Seconds())
			if remaining < 0 {
				remaining = 0
			}
			m.exitCountdownSecs = remaining
			userTimerExpired = remaining == 0
		}

		// 2. Summary renderer: runs in background (handled by SummaryDataMsg)
		// pendingSummaryData is set when builder finishes - nothing to do here

		// 3. Exit decision: should we start the exit sequence?
		// Conditions: all runners completed AND (user never interacted OR user timer expired)
		allRunnersDone := !m.allRunnersCompleted.IsZero()
		shouldExit := allRunnersDone && (!m.userHasInteracted || userTimerExpired)

		if shouldExit && !m.exitRequested {
			m.exitRequested = true
		}

		// 4. Finalization: if exit requested, either quit (if summary ready) or show "rendering summary"
		if m.exitRequested {
			if m.pendingSummaryData != nil {
				// Summary is ready - activate and quit
				m.activateSummary()
				m.quitting = true
				return m, tea.Quit
			}
			// Summary not ready yet - keep ticking (view will show "rendering summary")
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

		// Store cache info for skipped (cached) modules
		if state, exists := m.moduleStates[msg.Moniker]; exists {
			if !msg.CacheTime.IsZero() {
				state.CacheTime = msg.CacheTime
			}
			if msg.LogPath != "" {
				state.LogPath = msg.LogPath
			}
		}

		// Auto-select next running tab if:
		// 1. The completed module was the effective active tab (including default first tab)
		// 2. User hasn't interacted (would indicate they're manually navigating)
		effectiveTab := m.getEffectiveActiveTab()
		if msg.Moniker == effectiveTab && !m.userHasInteracted {
			// Find next running module to select
			for _, moniker := range m.moduleOrder {
				if state, exists := m.moduleStates[moniker]; exists {
					if state.Status == ModuleRunning {
						m.activeTab = moniker
						// Reset scroll for new tab
						if m.panes[PhaseRun] != nil {
							m.panes[PhaseRun].scrollOffset = 0
							m.panes[PhaseRun].autoScroll = true
						}
						break
					}
				}
			}
		}
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
	case "esc":
		// Force immediate exit (bypass user interaction delay)
		m.forceExit = true
		m.quitting = true
		return m, tea.Quit
	case "enter":
		// Enter skips the freeze countdown (sets timer to 0, normal exit behavior continues)
		if m.userHasInteracted && m.exitCountdownSecs > 0 {
			m.exitCountdownSecs = 0
		}
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
	case "t":
		// Toggle between TabGrid and Tree view
		if m.viewMode == ViewModeTabGrid {
			m.viewMode = ViewModeTree
		} else {
			m.viewMode = ViewModeTabGrid
		}
		// Reset scroll offset when switching views (different layouts)
		m.tabsScrollOffset = 0
	case "m":
		// Toggle mouse mode: ON = scrolling/clicking, OFF = text selection
		m.mouseMode = !m.mouseMode
		if m.mouseMode {
			return m, tea.EnableMouseAllMotion
		}
		return m, tea.DisableMouse
	case "left":
		// Decrease tab columns (min 2)
		if m.tabColumns > 2 {
			m.tabColumns--
			m.tabsScrollOffset = 0 // Reset scroll when changing layout
		}
	case "right":
		// Increase tab columns (max 6)
		if m.tabColumns < 6 {
			m.tabColumns++
			m.tabsScrollOffset = 0 // Reset scroll when changing layout
		}
	}
	return m, nil
}

// activateSummary transitions from pending summary to active summary display.
func (m *Model) activateSummary() {
	if m.pendingSummaryData == nil {
		return
	}

	m.summaryData = m.pendingSummaryData
	m.pendingSummaryData = nil

	// Finalize any tabs still in non-terminal states before exiting.
	// By the time summary arrives, all workers have completed - any remaining
	// running/pending tabs are due to message queue timing. Mark them as
	// skipped (blue) to avoid showing misleading orange tabs at exit.
	m.finalizeAllTabs()

	// Activate Summary pane
	if m.activePhase != PhaseSummary {
		// Mark current phase as complete or failed based on success
		if m.panes[m.activePhase].Status == PhaseActive {
			if m.summaryData != nil && !m.summaryData.Success {
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

	// Mark summary pane as complete/failed
	if m.summaryData != nil && !m.summaryData.Success {
		m.panes[PhaseSummary].Status = PhaseFailed
	} else {
		m.panes[PhaseSummary].Status = PhaseComplete
	}
	m.panes[PhaseSummary].EndTime = time.Now()
}

// finalizeAllTabs ensures all module tabs are in terminal states before exit.
// Any tabs still in running/pending state are marked as skipped (blue).
// This handles message queue timing where completion messages haven't been
// processed yet when the summary arrives.
func (m *Model) finalizeAllTabs() {
	for _, state := range m.moduleStates {
		if state.Status == ModuleRunning || state.Status == ModulePending {
			// Mark as skipped (blue) - the safest default since by the time
			// summary arrives, all work is complete. Running/pending state
			// indicates the completion message is still in the queue.
			if state.Status == ModuleRunning && m.uowRunning > 0 {
				m.uowRunning--
			}
			state.Status = ModuleSkipped
			state.ExitCode = -1
			if state.EndTime.IsZero() {
				state.EndTime = time.Now()
			}
			m.uowCached++
		}
	}
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
	// Track ANY mouse interaction (scroll or click) - this delays auto-exit
	// so user can investigate output without losing context
	if msg.Button != tea.MouseButtonNone && msg.Action != tea.MouseActionMotion {
		m.userHasInteracted = true
		m.lastUserInteraction = time.Now()
		m.exitRequested = false // Reset exit sequence - user is still investigating
	}

	// Mouse interactions:
	// - Scroll wheel → scroll panes (handled first to prevent any accidental tab switching)
	// - Click on tab bar → switch tabs
	// - Shift+Click → select text (standard terminal behavior, bypasses mouse mode)

	// Resource zone detection helper for Resources pane
	// IMPORTANT: Uses calculateLayoutMetrics() as single source of truth for Y offset.
	// This ensures hover detection aligns with rendered output regardless of layout changes.
	detectResourceZoneAt := func(x, y int) string {
		metrics := m.calculateLayoutMetrics()

		// Resources pane only exists if it has lines
		if metrics.ResourcesLines == 0 {
			return ""
		}

		// Resources pane layout (0-indexed Y coordinates):
		// Y = InitLines:     Resources header (contains freeze-button)
		// Y = InitLines + 1: Content line 1 (timer, cpu, mem, jobs)
		// Y = InitLines + 2: Content line 2 (uow, tools, layer)
		// Y = InitLines + 3: Resources footer

		resourcesStartY := metrics.InitLines // Header line
		contentLine1Y := metrics.InitLines + 1
		contentLine2Y := metrics.InitLines + 2

		// Column boundaries (matching renderResourcesPane)
		const (
			col1End = 38        // timer+CPU or UoW
			col2End = 38 + 3 + 24 // + separator + Mem/Tools
			col3End = 38 + 3 + 24 + 3 + 20 // + separator + Jobs/Layer
		)

		// Check header line for freeze button
		if y == resourcesStartY {
			// Freeze button is at the right end of the header
			// It's marked with zone, so just check if we're on header line and right side
			if x > m.width-20 { // Approximate freeze button location
				return "freeze-button"
			}
			return ""
		}

		// Check content line 1: timer, cpu, mem, jobs
		if y == contentLine1Y {
			if x < col1End {
				// Col1 contains both timer and CPU
				// Timer is first ~6 chars, CPU follows
				if x < 8 {
					return "res-timer"
				}
				return "res-cpu"
			} else if x < col2End {
				return "res-mem"
			} else if x < col3End {
				return "res-jobs"
			}
			return ""
		}

		// Check content line 2: uow, tools, layer
		if y == contentLine2Y {
			if x < col1End {
				return "res-uow"
			} else if x < col2End {
				return "res-tools"
			} else if x < col3End {
				return "res-layer"
			}
			return ""
		}

		return ""
	}

	// Tab detection helper for side-by-side layout (works for both tab grid and tree view)
	// IMPORTANT: Uses calculateLayoutMetrics() as single source of truth for Y offset.
	// This ensures click detection aligns with rendered output regardless of layout changes.
	detectTabAt := func(x, y int) string {
		if m.panes[PhaseRun].Status == PhasePending {
			return ""
		}
		tabs := m.GetVisibleTabs()
		if len(tabs) == 0 {
			return ""
		}

		// Dynamic components panel width (matches renderSideBySideLayout)
		componentsWidth := m.ComponentsWidth()

		// Check if X is within the components panel (left side)
		if x >= componentsWidth {
			return "" // Click is on the logs panel (right side)
		}

		// Use shared layout metrics - single source of truth for Y offset calculation
		// ComponentsStart already includes: init + resources + selected + panel header
		metrics := m.calculateLayoutMetrics()

		// Content starts at ComponentsStart (0-indexed Y coordinate)
		// Subtract 1 because mouse Y coordinates are 0-indexed in terminal
		contentStartY := metrics.ComponentsStart - 1

		// Check if Y is within content area
		row := y - contentStartY
		if row < 0 {
			return ""
		}

		// Clamp scroll offset if it's become invalid (content may have changed)
		if m.tabsScrollOffset < 0 {
			m.tabsScrollOffset = 0
		}

		// Account for scroll offset - add it to get the actual content row
		row += m.tabsScrollOffset

		// Tree view mode - each line maps to a component (need to track line->moniker mapping)
		if m.viewMode == ViewModeTree {
			// Build line-to-moniker mapping (matches renderTreeContent logic)
			var lineToMoniker []string

			// Check if we should use ExecutionTree or fallback (same logic as renderTreeContent)
			useExecutionTree := false
			if m.initSummary != nil && len(m.initSummary.ExecutionTree) > 0 {
				matchedTabs := 0
				tabMap := make(map[string]bool)
				for _, tab := range tabs {
					tabMap[tab.Moniker] = true
				}
				for _, layer := range m.initSummary.ExecutionTree {
					for _, module := range layer.Modules {
						for _, comp := range module.Components {
							moniker := module.Name + ":" + comp
							if tabMap[moniker] {
								matchedTabs++
							}
						}
					}
				}
				useExecutionTree = matchedTabs > 0 && matchedTabs >= len(tabs)/2
			}

			if useExecutionTree {
				// Tree from ExecutionTree
				for layerIdx, layer := range m.initSummary.ExecutionTree {
					lineToMoniker = append(lineToMoniker, "") // Layer header line
					for _, module := range layer.Modules {
						lineToMoniker = append(lineToMoniker, "") // Module line
						for _, comp := range module.Components {
							moniker := module.Name + ":" + comp
							lineToMoniker = append(lineToMoniker, moniker) // Component line
						}
					}
					if layerIdx < len(m.initSummary.ExecutionTree)-1 {
						lineToMoniker = append(lineToMoniker, "") // Spacing between layers
					}
				}
			} else {
				// Fallback: group tabs by module
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

				for _, moduleName := range moduleOrder {
					lineToMoniker = append(lineToMoniker, "") // Module line
					for _, tab := range moduleGroups[moduleName] {
						lineToMoniker = append(lineToMoniker, tab.Moniker) // Component line
					}
				}
			}

			if row < len(lineToMoniker) {
				return lineToMoniker[row]
			}
			return ""
		}

		// Tab grid mode - compact single-line tabs, grouped by layer
		// Use configured tab columns (matches renderTabGridContent)
		const tabWidth = 15
		tabsPerRow := m.tabColumns
		if tabsPerRow < 1 {
			tabsPerRow = 1
		}
		if tabsPerRow > 6 {
			tabsPerRow = 6
		}

		// Build a map from moniker to tab state for quick lookup
		tabMap := make(map[string]*ModuleState)
		for _, t := range tabs {
			tabMap[t.Moniker] = t
		}

		// Calculate column from X (accounting for left border and tab width)
		col := (x - 1) / tabWidth
		if col < 0 || col >= tabsPerRow {
			return ""
		}

		// Check if layer grouping is used (same logic as renderTabGridContent)
		useLayerGrouping := false
		if m.initSummary != nil && len(m.initSummary.ExecutionTree) > 0 {
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
			useLayerGrouping = matchedTabs > 0 && matchedTabs >= len(tabs)/2
		}

		// Handle layer-grouped layout
		if useLayerGrouping {
			currentLine := 0
			for _, layer := range m.initSummary.ExecutionTree {
				// Layer header takes 1 line
				currentLine++

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

				// Each row of tabs takes 1 line (compact layout)
				numRows := (len(layerTabs) + tabsPerRow - 1) / tabsPerRow
				layerEndLine := currentLine + numRows

				if row >= currentLine && row < layerEndLine {
					// Click is within this layer's tab area
					localRow := row - currentLine
					tabIdx := localRow*tabsPerRow + col
					if tabIdx >= 0 && tabIdx < len(layerTabs) {
						return layerTabs[tabIdx].Moniker
					}
					return ""
				}

				currentLine = layerEndLine
			}
			return ""
		}

		// Fallback: flat list (no layer grouping)
		tabRow := row // 1 line per row (compact)
		numTabRows := (len(tabs) + tabsPerRow - 1) / tabsPerRow
		if tabRow >= numTabRows {
			return ""
		}

		tabIdx := tabRow*tabsPerRow + col
		if tabIdx >= 0 && tabIdx < len(tabs) {
			return tabs[tabIdx].Moniker
		}
		return ""
	}

	// Handle wheel events FIRST - scrolling should never change tabs
	if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
		// Dynamic components panel width (matches renderSideBySideLayout)
		componentsWidth := m.ComponentsWidth()
		scrollAmount := 3 // Lines to scroll per wheel tick

		// Check if mouse is over tabs pane (left) or logs pane (right)
		if msg.X < componentsWidth {
			// Scrolling over tabs/tree pane (left side)
			// Calculate max scroll based on content
			tabs := m.GetVisibleTabs()
			maxScroll := 0
			if m.viewMode == ViewModeTree {
				// Tree view: count lines
				if m.initSummary != nil && len(m.initSummary.ExecutionTree) > 0 {
					for layerIdx, layer := range m.initSummary.ExecutionTree {
						maxScroll++ // Layer header
						for _, module := range layer.Modules {
							maxScroll++ // Module line
							maxScroll += len(module.Components)
						}
						if layerIdx < len(m.initSummary.ExecutionTree)-1 {
							maxScroll++ // Spacing
						}
					}
				} else {
					// Fallback count
					maxScroll = len(tabs) * 2
				}
			} else {
				// Tab grid: count lines (layer header + 1 line per row of compact tabs)
				// Uses configured tab columns (adjustable with left/right arrows)
				tabsPerRow := m.tabColumns
				if tabsPerRow < 1 {
					tabsPerRow = 1
				}
				if tabsPerRow > 6 {
					tabsPerRow = 6
				}
				if m.initSummary != nil && len(m.initSummary.ExecutionTree) > 0 {
					for _, layer := range m.initSummary.ExecutionTree {
						maxScroll++ // Layer header
						layerComps := 0
						for _, module := range layer.Modules {
							layerComps += len(module.Components)
						}
						rows := (layerComps + tabsPerRow - 1) / tabsPerRow
						maxScroll += rows // 1 line per row (compact)
					}
				} else {
					rows := (len(tabs) + tabsPerRow - 1) / tabsPerRow
					maxScroll = rows
				}
			}
			// Leave some visible content (at least 5 lines)
			maxScroll -= 5
			if maxScroll < 0 {
				maxScroll = 0
			}

			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.tabsScrollOffset -= scrollAmount
				if m.tabsScrollOffset < 0 {
					m.tabsScrollOffset = 0
				}
			case tea.MouseButtonWheelDown:
				m.tabsScrollOffset += scrollAmount
				if m.tabsScrollOffset > maxScroll {
					m.tabsScrollOffset = maxScroll
				}
			}
			// Update hover state after scroll - the tab under the cursor may have changed
			// (detectTabAt accounts for scroll offset, so it will find the correct tab)
			hoveredTab := detectTabAt(msg.X, msg.Y)
			if hoveredTab != m.hoveredTab {
				m.hoveredTab = hoveredTab
				m.hoveredTabScroll = 0 // Reset marquee scroll on hover change
				// Update hoveredZone for tab
				if hoveredTab != "" {
					m.hoveredZone = "tab:" + hoveredTab
				} else {
					m.hoveredZone = ""
				}
				if hoveredTab != "" {
					return m, tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
						return MarqueeTickMsg{}
					})
				}
			}
			return m, nil
		}

		// Scrolling over logs pane (right side) - directly use Run pane
		// In side-by-side layout, right side is always the Run pane logs
		pane := m.panes[PhaseRun]
		if pane == nil {
			return m, nil
		}

		// Use shared layout metrics - single source of truth for height calculation
		metrics := m.calculateLayoutMetrics()
		paneHeight := metrics.RemainingHeight - 2 // -2 for panel header/footer
		if paneHeight < 5 {
			paneHeight = 5
		}

		// Determine which buffer is being displayed (for Run pane with active tab)
		buffer := pane.Buffer
		if m.activeTab != "" {
			if moduleBuffer := m.GetActiveModuleBuffer(); moduleBuffer != nil {
				buffer = moduleBuffer
			}
		}

		// Update max scroll based on the ACTIVE buffer
		pane.UpdateMaxScrollForBuffer(buffer, paneHeight)

		// Scroll the pane - use Button to determine direction
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			pane.ScrollUp(scrollAmount)
		case tea.MouseButtonWheelDown:
			pane.ScrollDown(scrollAmount)
		}

		return m, nil
	}

	// Handle mouse motion for hover effect
	if msg.Action == tea.MouseActionMotion {
		// Check Resources pane zones first (help text display)
		// Uses layout-aware detection for accurate Y positioning
		resourceZone := detectResourceZoneAt(msg.X, msg.Y)
		if resourceZone != "" {
			if m.hoveredZone != resourceZone {
				m.hoveredZone = resourceZone
				m.hoveredTab = ""           // Clear tab hover when over resource zone
				m.hoveredTabScroll = 0      // Reset marquee
			}
			return m, nil
		}

		// If not over a resource zone, check for tab hover
		hoveredTab := detectTabAt(msg.X, msg.Y)
		if hoveredTab != m.hoveredTab {
			m.hoveredTab = hoveredTab
			m.hoveredTabScroll = 0 // Reset marquee scroll on hover change
			// Update hoveredZone for tab
			if hoveredTab != "" {
				m.hoveredZone = "tab:" + hoveredTab
			} else {
				m.hoveredZone = ""
			}
			// Start marquee ticker if hovering a tab
			if hoveredTab != "" {
				return m, tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
					return MarqueeTickMsg{}
				})
			}
		} else if hoveredTab == "" && m.hoveredZone != "" {
			// Clear hoveredZone if not over any resource zone or tab
			m.hoveredZone = ""
		}
		return m, nil
	}

	// Handle left mouse button click for tab selection and freeze button
	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease {
		// Check for Freeze button click first
		if zone.Get("freeze-button").InBounds(msg) {
			// Activate freeze: set 2-minute timeout and reset timer
			m.freezeTimeoutSecs = 120
			m.userHasInteracted = true
			m.lastUserInteraction = time.Now()
			m.exitRequested = false // Cancel any pending exit
			return m, nil
		}

		// Then check for tab clicks
		if tab := detectTabAt(msg.X, msg.Y); tab != "" {
			if tab != m.activeTab {
				m.SetActiveTab(tab)
				if m.panes[PhaseRun] != nil {
					m.panes[PhaseRun].scrollOffset = 0
					m.panes[PhaseRun].autoScroll = true
				}
			}
			// Reset tabs scroll to ensure continued click detection works
			m.tabsScrollOffset = 0
			return m, nil
		}
	}

	return m, nil
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
