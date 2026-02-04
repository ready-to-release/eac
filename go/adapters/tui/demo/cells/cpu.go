package cells

import "strings"

// CPUCell displays CPU usage as dots, one per core.
type CPUCell struct {
	percents  []float64
	asciiMode bool
}

// NewCPUCell creates a new CPU cell.
func NewCPUCell() *CPUCell {
	return &CPUCell{}
}

// SetPercents sets per-core CPU percentages.
func (c *CPUCell) SetPercents(percents []float64) {
	c.percents = percents
}

// SetASCIIMode enables or disables ASCII-only output.
func (c *CPUCell) SetASCIIMode(ascii bool) {
	c.asciiMode = ascii
}

// Render returns the CPU usage display.
func (c *CPUCell) Render(width, height int) string {
	filled, empty := "●", "○"
	if c.asciiMode {
		filled, empty = "*", "o"
	}

	numCores := len(c.percents)
	if numCores == 0 {
		return "CPU:"
	}

	// Calculate average usage to determine filled dots
	var total float64
	for _, p := range c.percents {
		total += p
	}
	avgPercent := total / float64(numCores)

	// Number of filled dots based on average
	filledCount := int(avgPercent / 100.0 * float64(numCores))

	var dots strings.Builder
	for i := 0; i < numCores; i++ {
		if i < filledCount {
			dots.WriteString(filled)
		} else {
			dots.WriteString(empty)
		}
	}

	return "CPU:" + dots.String()
}

// ZoneID returns the mouse zone identifier.
func (c *CPUCell) ZoneID() string {
	return "res-cpu"
}
