package render

// LayoutMetrics contains pre-calculated layout dimensions for the TUI.
// This is the SINGLE SOURCE OF TRUTH for layout calculations.
// Both rendering (viewPanes) and mouse handling (detectTabAt) MUST use this
// to ensure click detection aligns with rendered output.
type LayoutMetrics struct {
	ComponentsStart int // Y coordinate where components panel content starts (always 1)
	SummaryLines    int // Lines reserved for status bar (5: header + data + help + lamps + footer)
	RemainingHeight int // Height available for side-by-side layout
	DetailPaneHeight int // Height of the detail pane (9 when active, 0 when collapsed)
}
