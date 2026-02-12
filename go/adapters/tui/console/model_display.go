package console

import "time"

// DisplayConfig holds display configuration values.
type DisplayConfig struct {
	Height                int           // Total height (default: tui.DefaultHeight)
	Width                 int           // Terminal width
	RunPhaseName          string        // Custom name for Run phase (e.g., "building", "testing")
	AsciiMode             bool          // Use ASCII-only characters (--ascii flag)
	MetricsUpdateInterval time.Duration // How often to refresh CPU/memory metrics
	MinDisplayTime        time.Duration // Minimum time to show completion state
	AutoScrollResume      time.Duration // Auto-scroll resume delay
	BufferSizeUoW         int           // Buffer size per UoW
}

// PaneHeights holds the cached pane height values.
type PaneHeights struct {
	InitH    int
	RunH     int
	SummaryH int
}

// calculatePaneHeights returns cached pane heights if available, otherwise computes them.
func (m Model) calculatePaneHeights() (initH, runH, summaryH int) {
	if m.Resources.CachedPaneHeights != nil {
		return m.Resources.CachedPaneHeights.InitH, m.Resources.CachedPaneHeights.RunH, m.Resources.CachedPaneHeights.SummaryH
	}
	return m.computePaneHeights()
}

// computePaneHeights determines how many lines each pane gets.
// Init and Summary are fixed, Run fills ALL remaining terminal space.
func (m Model) computePaneHeights() (initH, runH, summaryH int) {
	// Fixed content heights for Init and Summary panes
	const initHeight = 6
	const summaryHeight = 4

	// Chrome lines (non-content):
	// - Init: header (1) + footer (1) = 2
	// - Resources pane: header (1) + content (1) + footer (1) = 3
	// - Run: header (1) + footer (1) = 2
	// Total base chrome = 7 lines
	chromeLines := 7

	// Add Summary chrome only if it will be rendered
	if m.Execution.SummaryData != nil {
		chromeLines += 2 // Summary header (1) + footer (1)
	}

	// Account for structured Init pane (8 lines total when initSummary is available)
	// instead of buffer-based init (2 chrome + 6 content)
	structuredInitHeight := 0
	if m.Execution.InitSummary != nil {
		structuredInitHeight = 8 // header + 4 rows + 2 separators + footer
	}

	// Account for Execution Tree pane (variable height based on modules)
	execTreeHeight := 0
	if m.Execution.InitSummary != nil && len(m.Execution.InitSummary.ExecutionTree) > 0 {
		// header (1) + modules + footer (1)
		execTreeHeight = 2 // header + footer
		for _, module := range m.Execution.InitSummary.ExecutionTree {
			execTreeHeight += 1 + len(module.UoWs) // module header + UoWs
		}
	}

	// Use actual terminal height (m.Display.Height is updated by WindowSizeMsg in alt-screen mode)
	terminalHeight := m.Display.Height
	if terminalHeight < 20 {
		terminalHeight = 20 // Minimum usable height
	}

	// Allocate heights
	if m.Execution.InitSummary != nil {
		// Using structured init - don't count initH in content, it's fully rendered
		initH = 0 // Not used for structured rendering
	} else {
		initH = initHeight
	}

	if m.Execution.SummaryData != nil {
		summaryH = summaryHeight
	} else {
		summaryH = 0 // Don't reserve space for Summary if not showing
	}

	// Calculate Run pane height
	if m.Execution.InitSummary != nil {
		// Structured mode: subtract fixed pane heights
		runH = terminalHeight - chromeLines - structuredInitHeight - execTreeHeight - summaryH
	} else {
		// Buffer mode: use original calculation
		runH = terminalHeight - chromeLines - initH - summaryH
	}

	// Minimum Run pane content height
	if runH < 5 {
		runH = 5
	}

	return initH, runH, summaryH
}
