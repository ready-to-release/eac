package console

import (
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
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

		// Route to per-UoW buffer only - no aggregate stream
		// Each unit of work's output is isolated to its own buffer
		if line.Source != "" && line.Source != "system" && m.activePhase == PhaseRun {
			// Check if this is a known UoW (has a state)
			if state, exists := m.uowStates[line.Source]; exists {
				state.Buffer.Push(line)
			} else {
				// Create UoW state on first output
				state := m.GetOrCreateUoWState(line.Source, "", 0)
				state.Buffer.Push(line)
			}
		} else {
			// System/phase lines go to the pane's buffer (init, summary phases)
			pane := m.panes[m.activePhase]
			pane.Buffer.Push(line)
		}

		// If pane is scrolled up (not auto-scrolling), increment offset to keep view locked
		// Only increment if this line affects the currently viewed buffer
		pane := m.panes[m.activePhase]
		if pane != nil && !pane.autoScroll && pane.scrollOffset > 0 {
			// Get effective tab (defaults to first UoW if none selected)
			effectiveTab := m.getEffectiveActiveTab()
			if line.Source == effectiveTab {
				// Only lines for the active UoW increment scroll
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
			for _, module := range msg.Summary.ExecutionTree {
				for _, uow := range module.UoWs {
					// Use full ID (Longname) for matching, DisplayName for tab labels
					m.GetOrCreateUoWState(uow.ID, uow.DisplayName, uow.Weight)
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
					m.plannedContainerTools = append(m.plannedContainerTools, tool.Name)
				} else {
					m.plannedSystemTools = append(m.plannedSystemTools, tool.Name)
				}
			}
		}

		// Initialize capacity tracking from WeightedCapacity
		// This ensures roof/pressureTarget have sensible defaults before first progress event
		// Fixes race condition where first render happens before ProgressUpdateEvent on slow machines
		if msg.Summary != nil && msg.Summary.WeightedCapacity > 0 {
			m.roof = msg.Summary.WeightedCapacity
			m.pressureTarget = msg.Summary.WeightedCapacity
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

		// Ensure TUI is visible for minimum display time before auto-exit
		// This prevents the TUI from flashing and disappearing on fast builds
		elapsed := time.Since(m.startTime)
		if elapsed < m.minDisplayTime && !m.userHasInteracted {
			// Not enough time has passed - delay exit until tick handler
			m.exitRequested = true
			return m, nil
		}

		// If user hasn't interacted and minimum time has passed, exit
		if !m.userHasInteracted {
			m.exitRequested = true
			m.activateSummary()
			m.quitting = true
			return m, tea.Quit
		}

		// User has interacted - start the countdown timer
		// Set exitRequested so tick handler knows to proceed with countdown
		m.exitRequested = true
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
				state := m.GetOrCreateUoWState(moniker, "", 0)
				state.Status = UoWRunning
				state.StartTime = time.Now()
			}
		}

		// Track EndTime for modules that were running but now gone
		// DON'T set status here - only UoWCompleteMsg sets terminal status
		// (it has the exit code to determine Complete vs Skipped vs Failed)
		for _, moniker := range m.running {
			if !newRunningSet[moniker] {
				// This module was running and is now gone
				if state, exists := m.uowStates[moniker]; exists {
					if state.EndTime.IsZero() {
						state.EndTime = time.Now()
					}
				}
			}
		}

		m.running = status.Running
		m.completed = status.Completed
		m.total = status.Total

		// Update capacity tracking (three-value model)
		// Only update if non-zero to preserve values across partial status updates
		if status.Roof > 0 {
			m.roof = status.Roof
		}
		if status.PressureTarget > 0 {
			m.pressureTarget = status.PressureTarget
		}

		// Track when all runners completed
		if m.total > 0 && m.completed >= m.total && m.allRunnersCompleted.IsZero() {
			m.allRunnersCompleted = time.Now()
		}

		// Update lock tracking info (preserve existing if new is empty for visual stability)
		if len(status.Locks) > 0 {
			m.locks = status.Locks

			// Extract docker-scheduler stats for visibility
			for _, lock := range status.Locks {
				if lock.Name == "docker-scheduler" {
					m.dockerRunning = lock.Used
					m.dockerRoof = lock.Capacity
					m.dockerWaiting = lock.Waiting
					// Docker scheduler uses capacity as pressure target (no separate pressure tracking)
					m.dockerPressureTarget = lock.Capacity
					break
				}
			}
		}

		// Update active tools only if provided (preserve across partial status updates)
		// Different observer events send partial Status - don't let progress updates wipe tool status
		if len(status.ActiveContainerTools) > 0 || len(status.UsedContainerTools) > 0 {
			m.activeContainerTools = status.ActiveContainerTools
			// Track seen containers (containers that have been used at least once)
			for _, c := range status.ActiveContainerTools {
				if !containsString(m.seenContainerTools, c) {
					m.seenContainerTools = append(m.seenContainerTools, c)
				}
			}
		}
		if len(status.ActiveSystemTools) > 0 || len(status.UsedSystemTools) > 0 {
			m.activeSystemTools = status.ActiveSystemTools
			// Track seen system tools (tools that have been used at least once)
			for _, t := range status.ActiveSystemTools {
				if !containsString(m.seenSystemTools, t) {
					m.seenSystemTools = append(m.seenSystemTools, t)
				}
			}
		}

		// Update Docker memory metrics only if explicitly set (preserve across partial updates)
		// dockerAvailable is "sticky" - once Docker becomes available, it stays available
		// This handles slow machines where Docker pool may initialize after first status events
		if status.DockerAvailable {
			m.cachedDockerMemPercent = status.DockerMemPercent
			m.dockerAvailable = true // sticky - never reset to false
		}

		// Update container instance counts (for Containers lamps)
		// Only update if provided - totalContainerCount should never decrease
		if status.TotalContainerCount > 0 || status.RunningContainerCount > 0 {
			m.runningContainerCount = status.RunningContainerCount
			// Total should only increase (lamps persist)
			if status.TotalContainerCount > m.totalContainerCount {
				m.totalContainerCount = status.TotalContainerCount
			}
		}

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
			m.panes[PhaseRun].CheckAutoScrollResume(m.autoScrollResume)
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
			// Ensure TUI is visible for minimum time before auto-exit (unless user interacted)
			elapsed := time.Since(m.startTime)
			if elapsed < m.minDisplayTime && !m.userHasInteracted {
				// Not enough time - keep ticking
				return m, m.tickCmd()
			}

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
		m.MarkUoWComplete(msg.Moniker, exitCode)
		return m, nil

	case UoWStartMsg:
		// Create module state when module starts (pending state)
		m.GetOrCreateUoWState(msg.Moniker, msg.DisplayName, msg.Weight)
		return m, nil

	case UoWRunningMsg:
		// Mark module as running (slot acquired)
		m.MarkUoWRunning(msg.Moniker)
		return m, nil

	case UoWCompleteMsg:
		// Mark module as complete (alternative to completedMsg)
		m.MarkUoWComplete(msg.Moniker, msg.ExitCode)

		// Store cache info for skipped (cached) modules
		if state, exists := m.uowStates[msg.Moniker]; exists {
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
			for _, moniker := range m.uowOrder {
				if state, exists := m.uowStates[moniker]; exists {
					if state.Status == UoWRunning {
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
		// Reserved - no aggregate view
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
		// Reserved - no aggregate view
	case "t":
		// Tree view removed - only tab grid view is supported
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
	for _, state := range m.uowStates {
		if state.Status == UoWRunning || state.Status == UoWPending {
			// Mark as skipped (blue) - the safest default since by the time
			// summary arrives, all work is complete. Running/pending state
			// indicates the completion message is still in the queue.
			// Note: counts are derived from state via DeriveCounts(), no counter update needed
			state.Status = UoWSkipped
			state.ExitCode = -1
			if state.EndTime.IsZero() {
				state.EndTime = time.Now()
			}
		}
	}
}

// cycleTab cycles through tabs in the given direction (+1 = next, -1 = prev).
func (m *Model) cycleTab(direction int) {
	tabs := m.GetVisibleTabs()
	if len(tabs) == 0 {
		return
	}

	// Build list of UoW tabs only (no aggregate view)
	allTabs := make([]string, 0, len(tabs))
	for _, t := range tabs {
		allTabs = append(allTabs, t.Moniker)
	}

	// Find current index (use effective tab for empty activeTab)
	effectiveTab := m.getEffectiveActiveTab()
	currentIdx := 0
	for i, moniker := range allTabs {
		if moniker == effectiveTab {
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
		// Only reset exit sequence if summary hasn't been received yet
		// Once summary is received, we're in countdown mode - don't reset
		if m.pendingSummaryData == nil {
			m.exitRequested = false
		}
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
		// Y = InitLines - 1: Resources header (contains freeze-button)
		// Y = InitLines:     Row 1 (CPU, Time, empty, Mem, Docker Mem)
		// Y = InitLines + 1: Row 2 (Host, Host weight, counters, Native, Containers)
		// Y = InitLines + 2: Row 3 (Docker, Docker weight, progress, Tools placeholder, Container tools)
		// Y = InitLines + 3: Resources footer

		resourcesStartY := metrics.InitLines - 1 // Header line
		row1Y := metrics.InitLines
		row2Y := metrics.InitLines + 1
		row3Y := metrics.InitLines + 2

		// Column boundaries (matching renderResourcesPane 5-column layout)
		// col1Width=24, col2Width=11, col3Width=12, col4Width=24, col5Width=28
		// Plus 3 chars for each separator " │ " and 2 chars for border "│ "
		const (
			borderLeft = 2                             // "│ "
			col1End    = borderLeft + 24               // End of column 1
			col2End    = col1End + 3 + 11              // + separator + column 2
			col3End    = col2End + 3 + 12              // + separator + column 3
			col4End    = col3End + 3 + 24              // + separator + column 4
			col5End    = col4End + 3 + 28              // + separator + column 5
		)

		// Check header line for freeze button
		if y == resourcesStartY {
			// Freeze button is at the right end of the header
			if x > m.width-20 { // Approximate freeze button location
				return "freeze-button"
			}
			return ""
		}

		// Row 1: CPU | Time | (empty) | Mem | Docker Mem
		if y == row1Y {
			if x < col1End {
				return "res-cpu"
			} else if x < col2End {
				return "res-timer"
			} else if x < col3End {
				return "" // empty column
			} else if x < col4End {
				return "res-mem"
			} else if x < col5End {
				return "res-dmem"
			}
			return ""
		}

		// Row 2: Host | Host weight | counters | Native | Containers
		if y == row2Y {
			if x < col1End {
				return "res-host"
			} else if x < col2End {
				return "res-host-weight"
			} else if x < col3End {
				return "res-counters"
			} else if x < col4End {
				return "res-native"
			} else if x < col5End {
				return "res-jobs"
			}
			return ""
		}

		// Row 3: Docker | Docker weight | progress | Tools placeholder | Container tools
		if y == row3Y {
			if x < col1End {
				return "res-docker"
			} else if x < col2End {
				return "res-docker-weight"
			} else if x < col3End {
				return "" // progress (no help text)
			} else if x < col4End {
				return "" // placeholder (no help text)
			} else if x < col5End {
				return "res-ctools"
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

		// Tab grid mode - compact single-line tabs
		// Use configured tab columns (matches renderTabGridContent)
		const tabWidth = 15
		tabsPerRow := m.tabColumns
		if tabsPerRow < 1 {
			tabsPerRow = 1
		}
		if tabsPerRow > 6 {
			tabsPerRow = 6
		}

		// Calculate column from X (accounting for left border and tab width)
		col := (x - 1) / tabWidth
		if col < 0 || col >= tabsPerRow {
			return ""
		}

		// Flat list - 1 line per row (compact)
		tabRow := row
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
			// Tab grid: count lines (1 line per row of compact tabs)
			// Uses configured tab columns (adjustable with left/right arrows)
			tabsPerRow := m.tabColumns
			if tabsPerRow < 1 {
				tabsPerRow = 1
			}
			if tabsPerRow > 6 {
				tabsPerRow = 6
			}
			rows := (len(tabs) + tabsPerRow - 1) / tabsPerRow
			maxScroll := rows
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
			if moduleBuffer := m.GetActiveUoWBuffer(); moduleBuffer != nil {
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

	// Handle text selection in logs pane (click-drag-release to copy)
	componentsWidth := m.ComponentsWidth()
	metrics := m.calculateLayoutMetrics()
	logsStartX := componentsWidth + 1
	logsStartY := metrics.ComponentsStart
	logsHeight := metrics.RemainingHeight - 2 // -2 for header/footer

	// Mouse press in logs pane - start selection
	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
		if msg.X >= logsStartX && msg.Y > logsStartY && msg.Y <= logsStartY+logsHeight {
			m.selection = SelectionState{
				Active:    true,
				StartX:    msg.X,
				StartY:    msg.Y,
				EndX:      msg.X,
				EndY:      msg.Y,
				StartLine: msg.Y - logsStartY - 1, // Convert to content line index
				EndLine:   msg.Y - logsStartY - 1,
			}
			return m, nil
		}
	}

	// Mouse motion while selecting - update end position
	if msg.Action == tea.MouseActionMotion && m.selection.Active {
		m.selection.EndX = msg.X
		m.selection.EndY = msg.Y
		m.selection.EndLine = msg.Y - logsStartY - 1
		return m, nil
	}

	// Mouse release - copy selection to clipboard
	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease && m.selection.Active {
		// Extract selected text from buffer
		selectedText := m.extractSelectedText(logsStartX, logsHeight)
		if selectedText != "" {
			_ = clipboard.WriteAll(selectedText)
		}
		m.selection = SelectionState{} // Clear selection
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

// extractSelectedText extracts text from the logs buffer based on selection state.
func (m Model) extractSelectedText(logsStartX, logsHeight int) string {
	if !m.selection.Active {
		return ""
	}

	// Get the active buffer
	pane := m.panes[PhaseRun]
	if pane == nil {
		return ""
	}

	var buffer *RingBuffer
	if m.activeTab != "" {
		if moduleBuffer := m.GetActiveUoWBuffer(); moduleBuffer != nil {
			buffer = moduleBuffer
		} else {
			buffer = pane.Buffer
		}
	} else {
		buffer = pane.Buffer
	}

	if buffer == nil {
		return ""
	}

	// Get visible lines (same logic as renderLogsPanel)
	var lines []Line
	if pane.scrollOffset == 0 || pane.autoScroll {
		lines = buffer.Last(logsHeight)
	} else {
		lines = buffer.GetRange(pane.scrollOffset, logsHeight)
	}

	// Normalize selection (start <= end)
	startLine, endLine := m.selection.StartLine, m.selection.EndLine
	if startLine > endLine {
		startLine, endLine = endLine, startLine
	}
	startX, endX := m.selection.StartX, m.selection.EndX
	if m.selection.StartY > m.selection.EndY || (m.selection.StartY == m.selection.EndY && m.selection.StartX > m.selection.EndX) {
		startX, endX = endX, startX
	}

	// Clamp to valid range
	if startLine < 0 {
		startLine = 0
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}
	if startLine > endLine || endLine < 0 {
		return ""
	}

	// Extract text from selected lines
	var result strings.Builder
	for i := startLine; i <= endLine; i++ {
		if i >= len(lines) {
			break
		}
		text := lines[i].Text

		// For single-line selection, extract substring
		if startLine == endLine {
			// Adjust X coordinates relative to text start (after prefix)
			textStartX := logsStartX + 3 // "│X " prefix
			charStart := startX - textStartX
			charEnd := endX - textStartX

			if charStart < 0 {
				charStart = 0
			}
			if charEnd > len(text) {
				charEnd = len(text)
			}
			if charStart < charEnd && charStart < len(text) {
				result.WriteString(text[charStart:charEnd])
			}
		} else {
			// Multi-line: full lines for middle, partial for first/last
			if i == startLine {
				textStartX := logsStartX + 3
				charStart := startX - textStartX
				if charStart < 0 {
					charStart = 0
				}
				if charStart < len(text) {
					result.WriteString(text[charStart:])
				}
				result.WriteString("\n")
			} else if i == endLine {
				textStartX := logsStartX + 3
				charEnd := endX - textStartX
				if charEnd > len(text) {
					charEnd = len(text)
				}
				if charEnd > 0 {
					result.WriteString(text[:charEnd])
				}
			} else {
				result.WriteString(text)
				result.WriteString("\n")
			}
		}
	}

	return result.String()
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

// containsString checks if a slice contains a string.
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
