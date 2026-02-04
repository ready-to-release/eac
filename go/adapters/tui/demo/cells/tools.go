package cells

import "strings"

// ToolsCell displays tool status as dots.
// Container tools are shown first (blue), then system tools (orange).
type ToolsCell struct {
	plannedContainerTools []string
	activeContainerTools  []string
	plannedSystem     []string
	activeSystem      []string
	asciiMode         bool
}

// NewToolsCell creates a new tools cell.
func NewToolsCell() *ToolsCell {
	return &ToolsCell{}
}

// SetTools sets the planned and active tool lists.
func (c *ToolsCell) SetTools(plannedContainerTools, activeContainerTools, plannedSystem, activeSystem []string) {
	c.plannedContainerTools = plannedContainerTools
	c.activeContainerTools = activeContainerTools
	c.plannedSystem = plannedSystem
	c.activeSystem = activeSystem
}

// SetASCIIMode enables or disables ASCII-only output.
func (c *ToolsCell) SetASCIIMode(ascii bool) {
	c.asciiMode = ascii
}

// Render returns the tools display.
func (c *ToolsCell) Render(width, height int) string {
	filled, empty := "●", "○"
	if c.asciiMode {
		filled, empty = "*", "o"
	}

	// Build sets for active tools
	activeContainerSet := make(map[string]bool)
	for _, name := range c.activeContainerTools {
		activeContainerSet[name] = true
	}
	activeSystemSet := make(map[string]bool)
	for _, name := range c.activeSystem {
		activeSystemSet[name] = true
	}

	var dots strings.Builder

	// Container tools first
	for _, name := range c.plannedContainerTools {
		if activeContainerSet[name] {
			dots.WriteString(filled)
		} else {
			dots.WriteString(empty)
		}
	}

	// System tools
	for _, name := range c.plannedSystem {
		if activeSystemSet[name] {
			dots.WriteString(filled)
		} else {
			dots.WriteString(empty)
		}
	}

	return "Tools:" + dots.String()
}

// ZoneID returns the mouse zone identifier.
func (c *ToolsCell) ZoneID() string {
	return "res-tools"
}
