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
		m.Display.Width = msg.Width
		m.Display.Height = msg.Height
		m.invalidateLayoutMetrics()
		// Clamp pane columns to fit new terminal width
		if maxCols := m.maxPaneWidthCols(); m.Interaction.PaneWidthCols > maxCols {
			m.Interaction.PaneWidthCols = maxCols
		}
		// Reset scroll offsets when window size changes
		for _, pane := range m.Execution.Panes {
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

	case batchLineMsg:
		lines := []Line(msg)
		m.processLinesBatch(lines)

		// Continue listening for more lines
		if !m.Execution.LinesDone {
			return m, m.listenForLines()
		}
		return m, nil

	case lineMsg:
		// Single-line fallback (kept for compatibility)
		m.processLinesBatch([]Line{Line(msg)})

		// Continue listening for more lines
		if !m.Execution.LinesDone {
			return m, m.listenForLines()
		}
		return m, nil

	case PhaseLineMsg:
		// Write to specific phase's buffer
		if m.Execution.Panes[msg.Phase] != nil {
			pane := m.Execution.Panes[msg.Phase]
			pane.Buffer.Push(msg.Line)
			// If pane is scrolled up (not auto-scrolling), increment offset to keep view locked
			// Only increment if viewing the aggregate buffer (PhaseLineMsg goes to pane.Buffer, not module buffers)
			if !pane.autoScroll && pane.scrollOffset > 0 {
				// For Run pane, only increment if viewing aggregate (not a specific module tab)
				if msg.Phase != PhaseRun || m.Interaction.ActiveTab == "" {
					pane.scrollOffset++
				}
			}
		}
		return m, nil

	case ResultLineMsg:
		// Write to results buffer
		m.Execution.ResultsBuffer.Push(msg.Line)
		return m, nil

	case ConfigReadyMsg:
		m.Boot.Config = &ConfigMeta{
			CommandName:      msg.CommandName,
			ExecutionContext: msg.ExecutionContext,
			ParallelismMode:  msg.ParallelismMode,
			EffectiveWorkers: msg.EffectiveWorkers,
			WeightedCapacity: msg.WeightedCapacity,
			OutputDir:        msg.OutputDir,
		}
		if m.Boot.State < BootConfig {
			m.Boot.State = BootConfig
		}
		// Pre-set capacity values for status bar rendering
		if msg.WeightedCapacity > 0 {
			m.Execution.Roof = msg.WeightedCapacity
			m.Execution.PressureTarget = msg.WeightedCapacity
		}
		return m, nil

	case InitSummaryMsg:
		// Store init summary for structured rendering
		m.Execution.InitSummary = msg.Summary

		// Build set of resolved IDs for reconciliation with planned tabs
		resolvedIDs := make(map[string]bool)
		if msg.Summary != nil && len(msg.Summary.ExecutionTree) > 0 {
			for _, module := range msg.Summary.ExecutionTree {
				for _, uow := range module.UoWs {
					resolvedIDs[uow.ID] = true
					// Use full ID (Longname) for matching, DisplayName for tab labels
					m.GetOrCreateUoWState(uow)
				}
			}

			// Remove stale planned tabs that are not in the resolved set
			// (only remove tabs still in Pending state - running/complete tabs are authoritative)
			if len(resolvedIDs) > 0 {
				for id, state := range m.Execution.UoWStates {
					if !resolvedIDs[id] && state.Status == UoWPending {
						m.removeUoWFromTabs(id)
					}
				}
			}

			// Calculate initial tab columns to fit all UoWs without scrolling
			m.Interaction.PaneWidthCols = m.calculateOptimalPaneCols()
		}

		// Initialize capacity tracking from WeightedCapacity
		// This ensures roof/pressureTarget have sensible defaults before first progress event
		// Fixes race condition where first render happens before ProgressUpdateEvent on slow machines
		if msg.Summary != nil && msg.Summary.WeightedCapacity > 0 {
			m.Execution.Roof = msg.Summary.WeightedCapacity
			m.Execution.PressureTarget = msg.Summary.WeightedCapacity
		}
		// Advance boot state
		if m.Boot.State < BootTabs {
			m.Boot.State = BootTabs
		}
		return m, nil

	case SummaryDataMsg:
		// Summary renderer has finished - store the data
		m.Interaction.PendingSummaryData = msg.Data

		// Receiving summary means all runners are done - mark complete if not already
		if m.Interaction.AllRunnersCompleted.IsZero() {
			m.Interaction.AllRunnersCompleted = time.Now()
		}

		// Gate exit on AllWorkDone - AfterExecute hooks must complete first.
		// If AllWorkDone hasn't been signaled yet, defer exit to tick handler.
		if !m.Execution.AllWorkDone {
			m.Interaction.ExitRequested = true
			return m, nil
		}

		// If skipTUIDelay is set, exit immediately without user interaction tracking
		if m.Interaction.SkipTUIDelay {
			m.activateSummary()
			m.Interaction.Quitting = true
			return m, tea.Quit
		}

		// Ensure TUI is visible for minimum display time before auto-exit
		// This prevents the TUI from flashing and disappearing on fast builds
		elapsed := time.Since(m.Execution.StartTime)
		if elapsed < m.Display.MinDisplayTime && !m.Interaction.UserHasInteracted {
			// Not enough time has passed - delay exit until tick handler
			m.Interaction.ExitRequested = true
			return m, nil
		}

		// If user hasn't interacted and minimum time has passed, exit
		if !m.Interaction.UserHasInteracted {
			m.Interaction.ExitRequested = true
			m.activateSummary()
			m.Interaction.Quitting = true
			return m, tea.Quit
		}

		// User has interacted - start the countdown timer
		// Set exitRequested so tick handler knows to proceed with countdown
		m.Interaction.ExitRequested = true
		return m, nil

	case PlannedWorkMsg:
		// Create grey skeleton tabs from module discovery (before tool resolution)
		for _, item := range msg.Items {
			m.GetOrCreateUoWState(UoWEntry{
				ID:            item.ID,
				DisplayName:   item.DisplayName,
				Weight:        item.Weight,
				Module:        item.Module,
				Component:     item.Component,
				ComponentType: item.ComponentType,
			})
		}
		// Recalculate pane columns to fit all planned tabs
		if len(msg.Items) > 0 {
			m.Interaction.PaneWidthCols = m.calculateOptimalPaneCols()
		}
		// Advance boot state to BootPlanned
		if m.Boot.State < BootPlanned {
			m.Boot.State = BootPlanned
		}
		return m, nil

	case UoWEnrichMsg:
		// Enrich an existing planned tab with tool resolution data
		if state, exists := m.Execution.UoWStates[msg.ID]; exists {
			if msg.Tool != "" {
				state.Tool = msg.Tool
				state.Container = msg.Container
			}
			if msg.Weight > 0 {
				state.Weight = msg.Weight
			}
		}
		return m, nil

	case AllWorkDoneMsg:
		// All work including AfterExecute hooks is complete
		m.Execution.AllWorkDone = true
		// If exit was already requested and summary is ready, quit now
		if m.Interaction.ExitRequested && m.Interaction.PendingSummaryData != nil {
			m.activateSummary()
			m.Interaction.Quitting = true
			return m, tea.Quit
		}
		return m, nil

	case PhaseUpdateMsg:
		if msg.Phase < Phase(len(m.Execution.Panes)) && m.Execution.Panes[msg.Phase] != nil {
			// Only update status if it's set (non-zero)
			if msg.Status != 0 {
				m.Execution.Panes[msg.Phase].Status = msg.Status
			}
			// Update summary if provided
			if msg.Summary != "" {
				m.Execution.Panes[msg.Phase].Summary = msg.Summary
			}

			// Track timing and active phase
			switch msg.Status {
			case PhaseActive:
				// Mark previous phase as complete if it was active
				if m.Execution.ActivePhase != msg.Phase && m.Execution.ActivePhase < Phase(len(m.Execution.Panes)) && m.Execution.Panes[m.Execution.ActivePhase].Status == PhaseActive {
					m.Execution.Panes[m.Execution.ActivePhase].Status = PhaseComplete
					m.Execution.Panes[m.Execution.ActivePhase].EndTime = time.Now()
				}
				m.Execution.Panes[msg.Phase].StartTime = time.Now()
				m.Execution.ActivePhase = msg.Phase
			case PhaseComplete, PhaseFailed:
				m.Execution.Panes[msg.Phase].EndTime = time.Now()
			}
		}
		return m, nil

	case statusMsg:
		status := Status(msg)

		// Build sets for comparison
		oldRunningSet := make(map[string]bool)
		for _, moniker := range m.Execution.Running {
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
				state := m.GetOrCreateUoWStateByID(moniker, "", 0)
				state.Status = UoWRunning
				state.StartTime = time.Now()
			}
		}

		// Track EndTime for modules that were running but now gone
		// DON'T set status here - only UoWCompleteMsg sets terminal status
		// (it has the exit code to determine Complete vs Skipped vs Failed)
		for _, moniker := range m.Execution.Running {
			if !newRunningSet[moniker] {
				// This module was running and is now gone
				if state, exists := m.Execution.UoWStates[moniker]; exists {
					if state.EndTime.IsZero() {
						state.EndTime = time.Now()
					}
				}
			}
		}

		m.Execution.Running = status.Running
		m.Execution.Completed = status.Completed
		m.Execution.Total = status.Total

		// Update capacity tracking (three-value model)
		// Only update if non-zero to preserve values across partial status updates
		if status.Roof > 0 {
			m.Execution.Roof = status.Roof
		}
		if status.PressureTarget > 0 {
			m.Execution.PressureTarget = status.PressureTarget
		}

		// Track when all runners completed
		if m.Execution.Total > 0 && m.Execution.Completed >= m.Execution.Total && m.Interaction.AllRunnersCompleted.IsZero() {
			m.Interaction.AllRunnersCompleted = time.Now()
		}

		// Update lock tracking info (preserve existing if new is empty for visual stability)
		if len(status.Locks) > 0 {
			m.Execution.Locks = status.Locks

			// Extract docker-scheduler stats for visibility
			for _, lock := range status.Locks {
				if lock.Name == "docker-scheduler" {
					m.Execution.DockerRunning = lock.Used
					m.Execution.DockerRoof = lock.Capacity
					m.Execution.DockerWaiting = lock.Waiting
					// Docker scheduler uses capacity as pressure target (no separate pressure tracking)
					m.Execution.DockerPressureTarget = lock.Capacity
					break
				}
			}
		}

		// Update Docker memory metrics only if explicitly set (preserve across partial updates)
		// dockerAvailable is "sticky" - once Docker becomes available, it stays available
		// This handles slow machines where Docker pool may initialize after first status events
		if status.DockerAvailable {
			m.Resources.CachedDockerMemPercent = status.DockerMemPercent
			m.Resources.DockerAvailable = true // sticky - never reset to false
		}

		if !m.Execution.StatusDone {
			return m, m.listenForStatus()
		}
		return m, nil

	case MarqueeTickMsg:
		// Animate hovered tab name scrolling (marquee effect)
		// Fast ticks (100ms) but only advance scroll every 4th tick for smooth slow movement
		if m.Interaction.HoveredTab != "" {
			m.Interaction.HoveredTabScroll++
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

		// Advance boot state when CPU metrics become available
		if m.Boot.State == BootChrome && len(m.Resources.CachedCPUPercent) > 0 {
			m.Boot.State = BootMetrics
		}

		// Recompute layout metrics (cached for View and mouse handlers until next tick)
		m.invalidateLayoutMetrics()

		// Clean up decayed tabs on each tick
		m.CleanupDecayedTabs()

		// Check if auto-scroll should resume after timeout
		if m.Execution.Panes[PhaseRun] != nil {
			m.Execution.Panes[PhaseRun].CheckAutoScrollResume(m.Display.AutoScrollResume)
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
		if m.Interaction.UserHasInteracted {
			elapsed := time.Since(m.Interaction.LastUserInteraction)
			remaining := m.Interaction.FreezeTimeoutSecs - int(elapsed.Seconds())
			if remaining < 0 {
				remaining = 0
			}
			m.Interaction.ExitCountdownSecs = remaining
			userTimerExpired = remaining == 0
		}

		// 2. Summary renderer: runs in background (handled by SummaryDataMsg)
		// pendingSummaryData is set when builder finishes - nothing to do here

		// 3. Exit decision: should we start the exit sequence?
		// Conditions: all runners completed AND (user never interacted OR user timer expired)
		allRunnersDone := !m.Interaction.AllRunnersCompleted.IsZero()
		shouldExit := allRunnersDone && (!m.Interaction.UserHasInteracted || userTimerExpired)

		if shouldExit && !m.Interaction.ExitRequested {
			m.Interaction.ExitRequested = true
		}

		// 4. Finalization: if exit requested, either quit (if summary ready AND all work done) or show "rendering summary"
		if m.Interaction.ExitRequested {
			// Ensure TUI is visible for minimum time before auto-exit (unless user interacted)
			elapsed := time.Since(m.Execution.StartTime)
			if elapsed < m.Display.MinDisplayTime && !m.Interaction.UserHasInteracted {
				// Not enough time - keep ticking
				return m, m.tickCmd()
			}

			if m.Interaction.PendingSummaryData != nil && m.Execution.AllWorkDone {
				// Summary is ready and AfterExecute hooks are complete - activate and quit
				m.activateSummary()
				m.Interaction.Quitting = true
				return m, tea.Quit
			}
			// Summary not ready yet or AfterExecute still running - keep ticking
		}

		return m, m.tickCmd()

	case linesDoneMsg:
		m.Execution.LinesDone = true
		return m, nil

	case statusDoneMsg:
		m.Execution.StatusDone = true
		return m, nil

	case completedMsg:
		// A module completed - update display
		m.Execution.Completed++
		// Remove from running list
		var newRunning []string
		for _, r := range m.Execution.Running {
			if r != msg.Moniker {
				newRunning = append(newRunning, r)
			}
		}
		m.Execution.Running = newRunning

		// Mark module as complete in tab tracking
		exitCode := msg.ExitCode
		m.MarkUoWComplete(msg.Moniker, exitCode)
		return m, nil

	case UoWStartMsg:
		// Create module state when module starts (pending state)
		m.GetOrCreateUoWStateByID(msg.Moniker, msg.DisplayName, msg.Weight)
		return m, nil

	case UoWRunningMsg:
		// Mark module as running (slot acquired)
		m.MarkUoWRunning(msg.Moniker)
		if m.Boot.State < BootRunning {
			m.Boot.State = BootRunning
		}
		return m, nil

	case UoWCompleteMsg:
		// Mark module as complete (alternative to completedMsg)
		m.MarkUoWComplete(msg.Moniker, msg.ExitCode)

		// Store cache info for skipped (cached) modules
		if state, exists := m.Execution.UoWStates[msg.Moniker]; exists {
			if !msg.CacheTime.IsZero() {
				state.CacheTime = msg.CacheTime
			}
			if msg.LogPath != "" {
				state.LogPath = msg.LogPath
			}
		}

		// Auto-select next running tab if:
		// 1. The completed module finished successfully (exit code 0 or cached)
		// 2. The completed module was the effective active tab (including default first tab)
		// 3. User hasn't interacted (would indicate they're manually navigating)
		// Failed jobs (ExitCode > 0) keep their tab focused so the user can see the error output.
		// ExitCode <= 0 covers success (0) and dep-skipped/cached (<0).
		effectiveTab := m.getEffectiveActiveTab()
		if msg.ExitCode <= 0 && msg.Moniker == effectiveTab && !m.Interaction.UserHasInteracted {
			// Find next running module to select
			for _, moniker := range m.Execution.UoWOrder {
				if state, exists := m.Execution.UoWStates[moniker]; exists {
					if state.Status == UoWRunning {
						m.Interaction.ActiveTab = moniker
						// Reset scroll for new tab
						if m.Execution.Panes[PhaseRun] != nil {
							m.Execution.Panes[PhaseRun].scrollOffset = 0
							m.Execution.Panes[PhaseRun].autoScroll = true
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
		if m.Execution.Panes[PhaseRun] != nil {
			m.Execution.Panes[PhaseRun].scrollOffset = 0
			m.Execution.Panes[PhaseRun].autoScroll = true
		}
		return m, nil

	case TabDecayMsg:
		// Clean up decayed tabs
		m.CleanupDecayedTabs()
		return m, nil
	}

	return m, nil
}

// processLinesBatch routes a batch of lines to their appropriate buffers.
// This is called for both batchLineMsg and single lineMsg.
func (m *Model) processLinesBatch(lines []Line) {
	pane := m.Execution.Panes[m.Execution.ActivePhase]
	effectiveTab := m.getEffectiveActiveTab()
	scrollIncrement := 0

	for i := range lines {
		line := &lines[i]

		// Route to per-UoW buffer only - no aggregate stream
		if line.Source != "" && line.Source != "system" && m.Execution.ActivePhase == PhaseRun {
			if state, exists := m.Execution.UoWStates[line.Source]; exists {
				state.Buffer.Push(*line)
			} else {
				state := m.GetOrCreateUoWStateByID(line.Source, "", 0)
				state.Buffer.Push(*line)
			}
		} else {
			pane.Buffer.Push(*line)
		}

		// Count scroll increments for the active UoW
		if pane != nil && !pane.autoScroll && pane.scrollOffset > 0 {
			if line.Source == effectiveTab {
				scrollIncrement++
			}
		}

		// Track errors for sticky display
		if line.Level == LevelError {
			lineCopy := *line
			m.Execution.LastError = &lineCopy
		}
	}

	// Apply scroll offset in bulk
	if scrollIncrement > 0 && pane != nil {
		pane.scrollOffset += scrollIncrement
	}
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.Interaction.Quitting = true
		return m, tea.Quit
	case "esc":
		// Force immediate exit (bypass user interaction delay)
		m.Interaction.ForceExit = true
		m.Interaction.Quitting = true
		return m, tea.Quit
	case "enter":
		// Enter skips the freeze countdown (sets timer to 0, normal exit behavior continues)
		if m.Interaction.UserHasInteracted && m.Interaction.ExitCountdownSecs > 0 {
			m.Interaction.ExitCountdownSecs = 0
		}
	case " ", "p":
		// Toggle pause
		m.Interaction.Paused = !m.Interaction.Paused
	case "e":
		// Toggle error-only mode
		m.Interaction.ErrorMode = !m.Interaction.ErrorMode
	case "c":
		// Clear last error
		m.Execution.LastError = nil

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
			if m.Execution.Panes[PhaseRun] != nil {
				m.Execution.Panes[PhaseRun].scrollOffset = 0
				m.Execution.Panes[PhaseRun].autoScroll = true
			}
		}
	case "0":
		// Reserved - no aggregate view
	case "t":
		// Cycle tab view mode: Name → Module → Type → Tool → Mode → State → Name
		m.Interaction.TabViewMode = (m.Interaction.TabViewMode + 1) % tabViewModeCount
	case "m":
		// Toggle mouse mode: ON = scrolling/clicking, OFF = text selection
		m.Interaction.MouseMode = !m.Interaction.MouseMode
		if m.Interaction.MouseMode {
			return m, tea.EnableMouseAllMotion
		}
		return m, tea.DisableMouse
	case "up":
		// Narrow each tab by 1 char
		if m.Interaction.TabWidth > tabWidthMin {
			m.Interaction.TabWidth--
			m.Interaction.TabsScrollOffset = 0
		}
	case "down":
		// Widen each tab by 1 char
		if m.Interaction.TabWidth < tabWidthMax {
			m.Interaction.TabWidth++
			m.Interaction.TabsScrollOffset = 0
		}
	case "left":
		// Remove one column
		if m.Interaction.PaneWidthCols > 1 {
			m.Interaction.PaneWidthCols--
			m.Interaction.TabsScrollOffset = 0
		}
	case "right":
		// Add one column
		maxCols := m.maxPaneWidthCols()
		if m.Interaction.PaneWidthCols < maxCols {
			m.Interaction.PaneWidthCols++
			m.Interaction.TabsScrollOffset = 0
		}
	}
	return m, nil
}

// activateSummary transitions from pending summary to active summary display.
func (m *Model) activateSummary() {
	if m.Interaction.PendingSummaryData == nil {
		return
	}

	m.Execution.SummaryData = m.Interaction.PendingSummaryData
	m.Interaction.PendingSummaryData = nil

	// Finalize any tabs still in non-terminal states before exiting.
	// By the time summary arrives, all workers have completed - any remaining
	// running/pending tabs are due to message queue timing. Mark them as
	// skipped (blue) to avoid showing misleading orange tabs at exit.
	m.finalizeAllTabs()

	// Activate Summary pane
	if m.Execution.ActivePhase != PhaseSummary {
		// Mark current phase as complete or failed based on success
		if m.Execution.Panes[m.Execution.ActivePhase].Status == PhaseActive {
			if m.Execution.SummaryData != nil && !m.Execution.SummaryData.Success {
				m.Execution.Panes[m.Execution.ActivePhase].Status = PhaseFailed
			} else {
				m.Execution.Panes[m.Execution.ActivePhase].Status = PhaseComplete
			}
			m.Execution.Panes[m.Execution.ActivePhase].EndTime = time.Now()
		}
		// Activate Summary pane
		m.Execution.ActivePhase = PhaseSummary
		m.Execution.Panes[PhaseSummary].Status = PhaseActive
		m.Execution.Panes[PhaseSummary].StartTime = time.Now()
	}

	// Mark summary pane as complete/failed
	if m.Execution.SummaryData != nil && !m.Execution.SummaryData.Success {
		m.Execution.Panes[PhaseSummary].Status = PhaseFailed
	} else {
		m.Execution.Panes[PhaseSummary].Status = PhaseComplete
	}
	m.Execution.Panes[PhaseSummary].EndTime = time.Now()
}

// finalizeAllTabs ensures all module tabs are in terminal states before exit.
// Any tabs still in running/pending state are marked as skipped (blue).
// This handles message queue timing where completion messages haven't been
// processed yet when the summary arrives.
func (m *Model) finalizeAllTabs() {
	for _, state := range m.Execution.UoWStates {
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
	if m.Execution.Panes[PhaseRun] != nil {
		m.Execution.Panes[PhaseRun].scrollOffset = 0
		m.Execution.Panes[PhaseRun].autoScroll = true
	}
}

// handleMouse handles mouse events for pane scrolling and tab clicks.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Track ANY mouse interaction (scroll or click) - this delays auto-exit
	// so user can investigate output without losing context
	if msg.Button != tea.MouseButtonNone && msg.Action != tea.MouseActionMotion {
		m.Interaction.UserHasInteracted = true
		m.Interaction.LastUserInteraction = time.Now()
		// Only reset exit sequence if summary hasn't been received yet
		// Once summary is received, we're in countdown mode - don't reset
		if m.Interaction.PendingSummaryData == nil {
			m.Interaction.ExitRequested = false
		}
	}

	// Mouse interactions:
	// - Scroll wheel → scroll panes (handled first to prevent any accidental tab switching)
	// - Click on tab bar → switch tabs
	// - Shift+Click → select text (standard terminal behavior, bypasses mouse mode)

	// Resource zone detection helper — resource widgets are now zone-marked in the status bar.
	// bubblezone's InBounds() auto-tracks rendered positions.
	detectResourceZoneAt := func() string {
		for _, zoneID := range []string{
			"res-mem", "res-host", "res-dmem", "res-docker",
			"res-cpu", "res-counters", "res-slots",
			"progress-count", "status-text", "freeze-button",
		} {
			if zone.Get(zoneID).InBounds(msg) {
				return zoneID
			}
		}
		return ""
	}

	// Tab detection helper for side-by-side layout (works for both tab grid and tree view)
	// IMPORTANT: Uses calculateLayoutMetrics() as single source of truth for Y offset.
	// This ensures click detection aligns with rendered output regardless of layout changes.
	detectTabAt := func(x, y int) string {
		if m.Execution.Panes[PhaseRun].Status == PhasePending {
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
		contentStartY := metrics.ComponentsStart

		// Check if Y is within content area
		row := y - contentStartY
		if row < 0 {
			return ""
		}

		// Clamp scroll offset if it's become invalid (content may have changed)
		if m.Interaction.TabsScrollOffset < 0 {
			m.Interaction.TabsScrollOffset = 0
		}

		// Account for scroll offset - add it to get the actual content row
		row += m.Interaction.TabsScrollOffset

		// Tab grid mode - use same sizing as rendering
		sizing := ComputeTabSizing(componentsWidth, m.Interaction.TabWidth, 0, m.Display.AsciiMode)
		tabsPerRow := sizing.TabColumns

		// Calculate column from X (accounting for left border and tab width + gap)
		col := (x - 1) / (sizing.TabWidth + tabGap)
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
			scrollSizing := ComputeTabSizing(componentsWidth, m.Interaction.TabWidth, 0, m.Display.AsciiMode)
			rows := (len(tabs) + scrollSizing.TabColumns - 1) / scrollSizing.TabColumns
			maxScroll := rows
			// Leave some visible content (at least 5 lines)
			maxScroll -= 5
			if maxScroll < 0 {
				maxScroll = 0
			}

			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.Interaction.TabsScrollOffset -= scrollAmount
				if m.Interaction.TabsScrollOffset < 0 {
					m.Interaction.TabsScrollOffset = 0
				}
			case tea.MouseButtonWheelDown:
				m.Interaction.TabsScrollOffset += scrollAmount
				if m.Interaction.TabsScrollOffset > maxScroll {
					m.Interaction.TabsScrollOffset = maxScroll
				}
			}
			// Update hover state after scroll - the tab under the cursor may have changed
			// (detectTabAt accounts for scroll offset, so it will find the correct tab)
			hoveredTab := detectTabAt(msg.X, msg.Y)
			if hoveredTab != m.Interaction.HoveredTab {
				m.Interaction.HoveredTab = hoveredTab
				m.Interaction.HoveredTabScroll = 0 // Reset marquee scroll on hover change
				// Update hoveredZone for tab
				if hoveredTab != "" {
					m.Interaction.HoveredZone = "tab:" + hoveredTab
				} else {
					m.Interaction.HoveredZone = ""
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
		pane := m.Execution.Panes[PhaseRun]
		if pane == nil {
			return m, nil
		}

		// Use shared layout metrics - single source of truth for height calculation
		metrics := m.calculateLayoutMetrics()
		var paneHeight int
		if metrics.DetailPaneHeight > 0 {
			paneHeight = metrics.RemainingHeight - metrics.DetailPaneHeight // borderless headless pane
		} else {
			paneHeight = metrics.RemainingHeight - 2 // -2 for panel header/footer
		}
		if paneHeight < 5 {
			paneHeight = 5
		}

		// Determine which buffer is being displayed (for Run pane with active tab)
		buffer := pane.Buffer
		if m.Interaction.ActiveTab != "" {
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

	// When detail pane is active, logs start after it (borderless: no header/footer).
	// Without detail pane, logs start after ComponentsStart + header line.
	// logsStartY is set so that (msg.Y - logsStartY) gives the content line index directly.
	var logsStartY, logsHeight int
	if metrics.DetailPaneHeight > 0 {
		logsStartY = metrics.ComponentsStart + metrics.DetailPaneHeight
		logsHeight = metrics.RemainingHeight - metrics.DetailPaneHeight // borderless headless pane
	} else {
		logsStartY = metrics.ComponentsStart + 1 // +1 to skip header line
		logsHeight = metrics.RemainingHeight - 2 // -2 for header/footer
	}

	// Mouse press in logs pane - start selection
	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
		if msg.X >= logsStartX && msg.Y > logsStartY && msg.Y <= logsStartY+logsHeight {
			m.Resources.Selection = SelectionState{
				Active:    true,
				StartX:    msg.X,
				StartY:    msg.Y,
				EndX:      msg.X,
				EndY:      msg.Y,
				StartLine: msg.Y - logsStartY,
				EndLine:   msg.Y - logsStartY,
			}
			return m, nil
		}
	}

	// Mouse motion while selecting - update end position
	if msg.Action == tea.MouseActionMotion && m.Resources.Selection.Active {
		m.Resources.Selection.EndX = msg.X
		m.Resources.Selection.EndY = msg.Y
		m.Resources.Selection.EndLine = msg.Y - logsStartY
		return m, nil
	}

	// Mouse release - copy selection to clipboard
	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease && m.Resources.Selection.Active {
		// Extract selected text from buffer
		selectedText := m.extractSelectedText(logsHeight)
		if selectedText != "" {
			_ = clipboard.WriteAll(selectedText)
		}
		m.Resources.Selection = SelectionState{} // Clear selection
		return m, nil
	}

	// Handle mouse motion for hover effect
	if msg.Action == tea.MouseActionMotion {
		// Check status bar zones first (help text display)
		resourceZone := detectResourceZoneAt()
		if resourceZone != "" {
			if m.Interaction.HoveredZone != resourceZone {
				m.Interaction.HoveredZone = resourceZone
				m.Interaction.HoveredTab = ""      // Clear tab hover when over resource zone
				m.Interaction.HoveredTabScroll = 0 // Reset marquee
			}
			return m, nil
		}

		// If not over a resource zone, check for tab hover
		hoveredTab := detectTabAt(msg.X, msg.Y)
		if hoveredTab != m.Interaction.HoveredTab {
			m.Interaction.HoveredTab = hoveredTab
			m.Interaction.HoveredTabScroll = 0 // Reset marquee scroll on hover change
			// Update hoveredZone for tab
			if hoveredTab != "" {
				m.Interaction.HoveredZone = "tab:" + hoveredTab
			} else {
				m.Interaction.HoveredZone = ""
			}
			// Start marquee ticker if hovering a tab
			if hoveredTab != "" {
				return m, tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
					return MarqueeTickMsg{}
				})
			}
		} else if hoveredTab == "" && m.Interaction.HoveredZone != "" {
			// Clear hoveredZone if not over any resource zone or tab
			m.Interaction.HoveredZone = ""
		}
		return m, nil
	}

	// Handle left mouse button click for tab selection and freeze button
	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease {
		// Check for Freeze button click first
		if zone.Get("freeze-button").InBounds(msg) {
			// Activate freeze: set 2-minute timeout and reset timer
			m.Interaction.FreezeTimeoutSecs = 120
			m.Interaction.UserHasInteracted = true
			m.Interaction.LastUserInteraction = time.Now()
			m.Interaction.ExitRequested = false // Cancel any pending exit
			return m, nil
		}

		// Then check for tab clicks
		if tab := detectTabAt(msg.X, msg.Y); tab != "" {
			if tab != m.Interaction.ActiveTab {
				m.SetActiveTab(tab)
				if m.Execution.Panes[PhaseRun] != nil {
					m.Execution.Panes[PhaseRun].scrollOffset = 0
					m.Execution.Panes[PhaseRun].autoScroll = true
				}
			}
			// Reset tabs scroll to ensure continued click detection works
			m.Interaction.TabsScrollOffset = 0
			return m, nil
		}
	}

	return m, nil
}

// extractSelectedText extracts text from the logs buffer based on selection state.
func (m Model) extractSelectedText(logsHeight int) string {
	if !m.Resources.Selection.Active {
		return ""
	}

	// Get the active buffer
	pane := m.Execution.Panes[PhaseRun]
	if pane == nil {
		return ""
	}

	var buffer *RingBuffer
	if m.Interaction.ActiveTab != "" {
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
	startLine, endLine := m.Resources.Selection.StartLine, m.Resources.Selection.EndLine
	if startLine > endLine {
		startLine, endLine = endLine, startLine
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

	// Extract full line text for all selected lines.
	// Lines can overflow past terminal width (open-right mode), so X-based
	// partial selection is unreliable — always copy the complete line text.
	var result strings.Builder
	for i := startLine; i <= endLine; i++ {
		if i > startLine {
			result.WriteString("\n")
		}
		result.WriteString(lines[i].Text)
	}

	return result.String()
}

// IsDone returns true if both line and status channels are done.
func (m Model) IsDone() bool {
	return m.Execution.LinesDone && m.Execution.StatusDone
}
