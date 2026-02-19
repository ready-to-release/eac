package console

import (
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
)

// detectResourceZoneAt checks if the given mouse message is within a resource zone.
// Returns the zone ID if found, empty string otherwise.
func (m Model) detectResourceZoneAt(msg tea.MouseMsg) string {
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

// detectTabAt determines which tab (if any) is at the given x,y coordinates.
// Uses calculateLayoutMetrics() as single source of truth for Y offset.
func (m Model) detectTabAt(x, y int) string {
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
	// ComponentsStart already includes: top bar (4) + newline (1) + panel header (1)
	metrics := m.calculateLayoutMetrics()

	// Content starts at ComponentsStart (0-indexed Y coordinate)
	contentStartY := metrics.ComponentsStart

	// Check if Y is within content area
	row := y - contentStartY
	if row < 0 {
		return ""
	}

	// Clamp scroll offset if it's become invalid (content may have changed)
	tabsScrollOffset := m.Interaction.TabsScrollOffset
	if tabsScrollOffset < 0 {
		tabsScrollOffset = 0
	}

	// Account for scroll offset - add it to get the actual content row
	row += tabsScrollOffset

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
			hoveredTab := m.detectTabAt(msg.X, msg.Y)
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
		resourceZone := m.detectResourceZoneAt(msg)
		if resourceZone != "" {
			if m.Interaction.HoveredZone != resourceZone {
				m.Interaction.HoveredZone = resourceZone
				m.Interaction.HoveredTab = ""      // Clear tab hover when over resource zone
				m.Interaction.HoveredTabScroll = 0 // Reset marquee
			}
			return m, nil
		}

		// If not over a resource zone, check for tab hover
		hoveredTab := m.detectTabAt(msg.X, msg.Y)
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
		if tab := m.detectTabAt(msg.X, msg.Y); tab != "" {
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
