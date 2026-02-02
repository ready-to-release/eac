package cells

import "strings"

const memTotalDots = 11

// MemCell displays memory usage as an 11-dot bar.
type MemCell struct {
	percent   float64
	asciiMode bool
}

// NewMemCell creates a new memory cell.
func NewMemCell() *MemCell {
	return &MemCell{}
}

// SetPercent sets the memory usage percentage (0-100).
func (c *MemCell) SetPercent(percent float64) {
	c.percent = percent
}

// SetASCIIMode enables or disables ASCII-only output.
func (c *MemCell) SetASCIIMode(ascii bool) {
	c.asciiMode = ascii
}

// Render returns the memory usage display.
func (c *MemCell) Render(width, height int) string {
	filled, empty := "●", "○"
	if c.asciiMode {
		filled, empty = "*", "o"
	}

	// Calculate filled dots (out of 11)
	filledCount := int(c.percent / 100.0 * memTotalDots)
	if filledCount > memTotalDots {
		filledCount = memTotalDots
	}

	var dots strings.Builder
	for i := 0; i < memTotalDots; i++ {
		if i < filledCount {
			dots.WriteString(filled)
		} else {
			dots.WriteString(empty)
		}
	}

	return "Mem:" + dots.String()
}

// ZoneID returns the mouse zone identifier.
func (c *MemCell) ZoneID() string {
	return "res-mem"
}
