package cells

import "fmt"

// SelectedCell displays information about the currently selected unit(s).
type SelectedCell struct {
	count   int
	primary string
	tool    string
	status  string
}

// NewSelectedCell creates a new selected cell.
func NewSelectedCell() *SelectedCell {
	return &SelectedCell{}
}

// SetSelection sets the selection state.
// count is the number of selected items.
// primary, tool, and status are used for single selection display.
func (c *SelectedCell) SetSelection(count int, primary, tool, status string) {
	c.count = count
	c.primary = primary
	c.tool = tool
	c.status = status
}

// Render returns the selection display.
func (c *SelectedCell) Render(width, height int) string {
	if c.count == 0 {
		return "(no selection)"
	}

	if c.count > 1 {
		return fmt.Sprintf("%d rows", c.count)
	}

	// Single selection: show details
	return fmt.Sprintf("%s │ %s │ %s", c.primary, c.tool, c.status)
}

// ZoneID returns empty string (not an interactive zone).
func (c *SelectedCell) ZoneID() string {
	return ""
}
