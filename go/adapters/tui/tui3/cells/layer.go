package cells

import "fmt"

// LayerCell displays the current execution layer indicator.
type LayerCell struct {
	current int
	total   int
}

// NewLayerCell creates a new layer cell.
func NewLayerCell() *LayerCell {
	return &LayerCell{}
}

// SetLayer sets the current and total layer values.
func (c *LayerCell) SetLayer(current, total int) {
	c.current = current
	c.total = total
}

// Render returns the layer display.
func (c *LayerCell) Render(width, height int) string {
	if c.total == 0 {
		return ""
	}
	return fmt.Sprintf("Layer: %d/%d", c.current, c.total)
}

// ZoneID returns the mouse zone identifier.
func (c *LayerCell) ZoneID() string {
	return "res-layer"
}
