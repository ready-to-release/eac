package cells

import "strings"

// ContainersCell displays container status as dots.
type ContainersCell struct {
	active    []string
	planned   []string
	asciiMode bool
}

// NewContainersCell creates a new containers cell.
func NewContainersCell() *ContainersCell {
	return &ContainersCell{}
}

// SetContainers sets the active and planned container lists.
func (c *ContainersCell) SetContainers(active, planned []string) {
	c.active = active
	c.planned = planned
}

// SetASCIIMode enables or disables ASCII-only output.
func (c *ContainersCell) SetASCIIMode(ascii bool) {
	c.asciiMode = ascii
}

// Render returns the containers display.
func (c *ContainersCell) Render(width, height int) string {
	filled, empty := "●", "○"
	if c.asciiMode {
		filled, empty = "*", "o"
	}

	// Build a set of active containers
	activeSet := make(map[string]bool)
	for _, name := range c.active {
		activeSet[name] = true
	}

	var dots strings.Builder
	for _, name := range c.planned {
		if activeSet[name] {
			dots.WriteString(filled)
		} else {
			dots.WriteString(empty)
		}
	}

	return "Containers:" + dots.String()
}

// ZoneID returns the mouse zone identifier.
func (c *ContainersCell) ZoneID() string {
	return "res-jobs"
}
