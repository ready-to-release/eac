package console

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/ready-to-release/eac/go/adapters/tui/console/render"
)

// InteractionState tracks user interaction, exit/freeze state, and display preferences.
type InteractionState struct {
	// Display preferences
	Paused           bool        // Pause scrolling (for review)
	ErrorMode        bool        // Show only errors
	MouseMode        bool        // Mouse mode: true=scrolling enabled, false=text selection enabled
	ViewMode         ViewMode    // Components view mode: TabGrid or Tree
	TabViewMode      TabViewMode // Which label text to show in UoW tabs (Name/Module/Type/Tool/Mode/State)
	TabsScrollOffset int         // Scroll offset for tabs/tree panel
	TabWidth         int         // User-controlled individual tab width (Up/Down, min/max limits)
	PaneWidthCols    int         // Logical pane width in column units (Left/Right adjusts)

	// Tab hover/selection
	ActiveTab        string // Currently selected tab ("" = aggregate view)
	HoveredTab       string // Tab currently under mouse cursor
	HoveredTabScroll int    // Marquee scroll offset for hovered tab name
	HoveredZone      string // Currently hovered zone ID (empty = none, "tab:moniker" for tabs, "res-*" for resources)
	MaxTabs          int    // Maximum visible tabs before scrolling/hiding

	// User interaction tracking (for delayed exit)
	LastUserInteraction time.Time // Reset on scroll or tab click
	UserHasInteracted   bool      // True if user ever interacted with mouse
	ForceExit           bool      // User pressed ESC to force immediate exit
	ExitCountdownSecs   int       // Seconds remaining until user timer expires (10, 9, 8...)
	FreezeTimeoutSecs   int       // Timeout duration in seconds (10 default, 120 when Freeze clicked)
	SkipTUIDelay        bool      // Skip all user interaction tracking, exit immediately when done
	AllRunnersCompleted time.Time // When all runners finished

	// Exit/finalization state
	ExitRequested      bool         // True when we want to exit (waiting for summary if needed)
	PendingSummaryData *SummaryData // Summary data received from builder

	// Quitting state - triggers plain-text final render
	Quitting bool
}

// ResourceState holds cached system metrics, layout metrics, and widget catalog.
type ResourceState struct {
	// Cached system metrics (updated periodically, not on every render)
	CachedCPUPercent       []float64 // Per-core CPU usage percentages
	CachedMemPercent       float64   // Memory usage percentage
	CachedDockerMemPercent float64   // Docker memory pool usage percentage
	DockerAvailable        bool      // Whether Docker is available
	LastMetricsUpdate      time.Time // When metrics were last updated

	// Cached layout metrics (recomputed on tick/resize, not every View/mouse event)
	CachedLayoutMetrics *render.LayoutMetrics

	// Cached pane heights (invalidated alongside layout metrics)
	CachedPaneHeights *PaneHeights

	// Widget catalog (initialized in NewModel)
	Catalog *WidgetCatalog

	// Text selection state (mouse drag to copy)
	Selection SelectionState
}

// SelectionState tracks mouse text selection in the logs pane.
type SelectionState struct {
	Active    bool // Currently selecting (mouse down)
	StartX    int  // Mouse press X (screen coords)
	StartY    int  // Mouse press Y (screen coords)
	EndX      int  // Current/end X (screen coords)
	EndY      int  // Current/end Y (screen coords)
	StartLine int  // Buffer line index at selection start
	EndLine   int  // Buffer line index at selection end
}

// ViewMode represents the display mode for the components pane.
type ViewMode int

const (
	ViewModeTabGrid ViewMode = iota // Tab grid view (default)
	ViewModeTree                    // Execution tree view
)

// TabViewMode controls what text appears in the label portion of UoW tabs.
type TabViewMode int

const (
	TabViewName   TabViewMode = iota // Component display name (default)
	TabViewModule                    // Module moniker (core, eac)
	TabViewType                      // Component type (go-lib, dockerfile, godog)
	TabViewTool                      // Tool/runner/builder name (go, buildx, mocha)
	TabViewExec                      // Execution mode: "container" or "host"
	TabViewState                     // Execution state (pending, run 4s, done 12s, cached, fail:1)
)

const tabViewModeCount = 6

// String returns the display name for a tab view mode.
func (v TabViewMode) String() string {
	switch v {
	case TabViewName:
		return "Name"
	case TabViewModule:
		return "Module"
	case TabViewType:
		return "Type"
	case TabViewTool:
		return "Tool"
	case TabViewExec:
		return "Mode"
	case TabViewState:
		return "State"
	default:
		return "Name"
	}
}

// Tab sizing constants for dynamic widget sizing.
const (
	tabWidthMin     = 10 // Minimum: 7 chars label + 3 badge
	tabWidthMax     = 30 // Maximum: comfortable reading width
	tabWidthDefault = 15 // Default: matches current behavior
	tabGap          = 1  // Gap between tabs (space character)
	tabBorder       = 2  // Left + right panel borders
)

// SetActiveTab sets the currently active tab.
// Tabs persist until FIFO removal, so any visible tab can be selected.
func (m *Model) SetActiveTab(moniker string) {
	// Empty moniker = aggregate view (always valid)
	if moniker == "" {
		m.Interaction.ActiveTab = ""
		return
	}

	// Validate moniker exists in visible tabs
	_, exists := m.Execution.UoWStates[moniker]
	if !exists {
		return
	}
	m.Interaction.ActiveTab = moniker
}

// GetActiveUoWBuffer returns the buffer for the active tab.
// If no tab is selected, returns the first tab's buffer.
func (m *Model) GetActiveUoWBuffer() *RingBuffer {
	activeMoniker := m.getEffectiveActiveTab()
	if activeMoniker == "" {
		return nil // No tabs yet
	}
	if state, exists := m.Execution.UoWStates[activeMoniker]; exists {
		return state.Buffer
	}
	return nil
}

// getEffectiveActiveTab returns the active tab, defaulting to first tab if none selected.
func (m *Model) getEffectiveActiveTab() string {
	if m.Interaction.ActiveTab != "" {
		return m.Interaction.ActiveTab
	}
	// Default to first tab
	if len(m.Execution.UoWOrder) > 0 {
		return m.Execution.UoWOrder[0]
	}
	return ""
}

// UpdateCachedMetrics refreshes the cached CPU and memory metrics.
// Should be called from the tick handler, not from View().
// The update interval is configured via TUIConfig.MetricsInterval.
func (m *Model) UpdateCachedMetrics() {
	// Only skip when actually quitting - keep metrics live while TUI is visible
	if m.Interaction.Quitting {
		return
	}

	// Skip if updated recently (use config-driven interval)
	if time.Since(m.Resources.LastMetricsUpdate) < m.Display.MetricsUpdateInterval {
		return
	}

	// Update CPU metrics (this is the slow call)
	if perCore, err := cpu.Percent(0, true); err == nil {
		m.Resources.CachedCPUPercent = perCore
	}

	// Update memory metrics
	if memInfo, err := mem.VirtualMemory(); err == nil {
		m.Resources.CachedMemPercent = memInfo.UsedPercent
	}

	m.Resources.LastMetricsUpdate = time.Now()
}

// calculateOptimalPaneCols determines the minimum number of columns (1-10)
// needed to display all UoWs without scrolling in the visible panel area.
// Simple math: cols = ceil(totalUoWs / uowPerColumn)
func (m *Model) calculateOptimalPaneCols() int {
	// Use UoWCount from initSummary - this is the actual count of scheduled work items
	// (number of scheduled work items, not just components)
	totalUoWs := 0
	if m.Execution.InitSummary != nil {
		totalUoWs = m.Execution.InitSummary.UoWCount
	}

	// Get visible rows from layout metrics
	metrics := m.calculateLayoutMetrics()
	// Panel content height = remaining height - header(1) - footer(1)
	uowPerCol := metrics.RemainingHeight - 2
	if uowPerCol < 1 {
		uowPerCol = 1
	}

	// cols = ceil(totalUoWs / uowPerCol)
	cols := (totalUoWs + uowPerCol - 1) / uowPerCol

	// Clamp to valid range
	if cols < 2 {
		cols = 2
	}
	maxCols := m.maxPaneWidthCols()
	if cols > maxCols {
		cols = maxCols
	}

	return cols
}

// maxPaneWidth returns the single source of truth for how wide the
// components panel is allowed to be: the smaller of (terminal - 40) and
// 60% of terminal, with a floor of 20.
func (m Model) maxPaneWidth() int {
	const minLogsWidth = 40
	// Two constraints — take the tighter one
	byLogs := m.Display.Width - minLogsWidth
	byPercent := m.Display.Width * 60 / 100
	max := byLogs
	if byPercent < max {
		max = byPercent
	}
	if max < 20 {
		max = 20
	}
	return max
}

// ComponentsWidth returns the width of the components panel.
// Derived from paneWidthCols (column count) and tabWidth (per-tab width).
// When the max-width constraint reduces available space, the column count
// is recomputed so the panel is sized to exactly fit — no empty gap.
func (m Model) ComponentsWidth() int {
	cols := m.Interaction.PaneWidthCols
	if cols < 1 {
		cols = 1
	}
	width := cols*m.Interaction.TabWidth + (cols-1)*tabGap + tabBorder
	if max := m.maxPaneWidth(); width > max {
		// Recompute how many columns actually fit, then size to match exactly
		cols = (max - tabBorder + tabGap) / (m.Interaction.TabWidth + tabGap)
		if cols < 1 {
			cols = 1
		}
		width = cols*m.Interaction.TabWidth + (cols-1)*tabGap + tabBorder
	}
	if width < 20 {
		width = 20
	}
	return width
}

// maxPaneWidthCols returns the maximum useful number of columns.
// Capped by panel max width (consistent with ComponentsWidth) and UoW count.
func (m Model) maxPaneWidthCols() int {
	maxWidth := m.maxPaneWidth()
	if maxWidth < m.Interaction.TabWidth+tabBorder {
		return 1
	}
	maxCols := (maxWidth - tabBorder + tabGap) / (m.Interaction.TabWidth + tabGap)
	if maxCols < 1 {
		maxCols = 1
	}
	if maxCols > 10 {
		maxCols = 10
	}
	// Never more columns than UoWs — extra columns would just be empty space
	uowCount := len(m.Execution.UoWOrder)
	if uowCount > 0 && maxCols > uowCount {
		maxCols = uowCount
	}
	return maxCols
}
