package cells

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SelectorCell displays a grid of selectable units organized by layers.
type SelectorCell struct {
	units     []SelectorUnit
	unitMap   map[string]*SelectorUnit
	layers    [][]string
	selected  string
	hovered   string
	asciiMode bool
}

// NewSelectorCell creates a new selector cell.
func NewSelectorCell() *SelectorCell {
	return &SelectorCell{
		unitMap: make(map[string]*SelectorUnit),
	}
}

// SetUnits sets the list of units.
// Makes a copy to avoid stale pointers when the source slice reallocates.
func (c *SelectorCell) SetUnits(units []SelectorUnit) {
	// Copy to avoid sharing backing array with caller
	c.units = make([]SelectorUnit, len(units))
	copy(c.units, units)

	c.unitMap = make(map[string]*SelectorUnit)
	for i := range c.units {
		c.unitMap[c.units[i].Moniker] = &c.units[i]
	}
}

// SetLayers sets the layer organization.
func (c *SelectorCell) SetLayers(layers [][]string) {
	c.layers = layers
}

// SetSelected sets the currently selected unit moniker.
func (c *SelectorCell) SetSelected(moniker string) {
	c.selected = moniker
}

// SetHovered sets the currently hovered unit moniker.
func (c *SelectorCell) SetHovered(moniker string) {
	c.hovered = moniker
}

// SetASCIIMode enables or disables ASCII-only output.
func (c *SelectorCell) SetASCIIMode(ascii bool) {
	c.asciiMode = ascii
}

// Render returns the selector display.
func (c *SelectorCell) Render(width, height int) string {
	if len(c.layers) == 0 {
		return ""
	}

	var lines []string

	for layerIdx, layer := range c.layers {
		// Layer header
		header := fmt.Sprintf("─ Module Layer %d ─", layerIdx+1)
		lines = append(lines, header)

		// Build tabs for this layer
		var tabs []string
		for _, moniker := range layer {
			tab := c.renderTab(moniker)
			tabs = append(tabs, tab)
		}

		// Join tabs with wrapping
		tabLine := c.wrapTabs(tabs, width)
		lines = append(lines, tabLine...)

		// Check if we've exceeded height
		if len(lines) >= height {
			break
		}
	}

	// Truncate to height
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// renderTab renders a single tab with weight number badge and name.
func (c *SelectorCell) renderTab(moniker string) string {
	unit := c.unitMap[moniker]
	if unit == nil {
		return fmt.Sprintf("[%s]", moniker)
	}

	// Use DisplayName for tab label if available, otherwise fall back to Moniker
	name := unit.DisplayName
	if name == "" {
		name = moniker
	}

	// Truncate long names
	if len(name) > 10 {
		name = name[:8] + ".."
	}

	isSelected := moniker == c.selected
	isHovered := moniker == c.hovered && !isSelected

	// Get badge colors (number area only)
	fg, bg, fgActive, bgActive := unit.Status.BadgeColors()

	// Choose colors based on selection state
	badgeFg, badgeBg := fg, bg
	if isSelected {
		badgeFg, badgeBg = fgActive, bgActive
	}

	// Format weight number, right-aligned with 1 space padding on each side
	// Weight badge is always 3 chars wide: " N " for single digit, "NN " for double
	weightStr := fmt.Sprintf("%d", unit.Weight)
	if unit.Weight < 10 {
		weightStr = " " + weightStr + " " // " N "
	} else {
		weightStr = weightStr + " " // "NN "
	}

	// Badge style: weight number with colored background
	badgeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(badgeFg)).
		Background(lipgloss.Color(badgeBg))
	badge := badgeStyle.Render(weightStr)

	// Name style: white text with subtle status-colored background
	// When hovered: use badge bg color with black text
	nameBg := unit.Status.NameBgColor()
	nameFg := "252" // Light grey/white
	if isHovered {
		nameBg = bg
		nameFg = "232" // Black text on bright background
	}
	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(nameFg)).
		Background(lipgloss.Color(nameBg))
	nameContent := nameStyle.Render(name)

	return badge + nameContent
}

// wrapTabs wraps tabs to fit within width, aligning them in fixed-width columns.
func (c *SelectorCell) wrapTabs(tabs []string, width int) []string {
	if len(tabs) == 0 {
		return nil
	}

	// Find the maximum visual width of all tabs
	maxTabWidth := 0
	for _, tab := range tabs {
		w := lipgloss.Width(tab)
		if w > maxTabWidth {
			maxTabWidth = w
		}
	}

	// Add 1 for spacing between columns
	colWidth := maxTabWidth + 1

	// Calculate how many columns fit
	numCols := width / colWidth
	if numCols < 1 {
		numCols = 1
	}

	var lines []string
	var currentLine strings.Builder
	colIdx := 0

	for i, tab := range tabs {
		// Pad tab to column width
		visualWidth := lipgloss.Width(tab)
		padding := maxTabWidth - visualWidth
		paddedTab := tab + strings.Repeat(" ", padding)

		if colIdx > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(paddedTab)
		colIdx++

		// Start new row when we've filled all columns
		if colIdx >= numCols {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			colIdx = 0
		}

		// Handle last tab if it doesn't fill a complete row
		if i == len(tabs)-1 && colIdx > 0 {
			lines = append(lines, currentLine.String())
		}
	}

	return lines
}

// GetTabZones returns the zone IDs for all tabs.
func (c *SelectorCell) GetTabZones() []string {
	var zones []string
	for _, unit := range c.units {
		zones = append(zones, fmt.Sprintf("tab:%s", unit.Moniker))
	}
	return zones
}

// ZoneID returns empty string (tabs have individual zones).
func (c *SelectorCell) ZoneID() string {
	return ""
}
