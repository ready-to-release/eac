package cells

import "strings"

const dockerMemTotalDots = 11

// DockerMemCell displays Docker memory pool usage as an 11-dot bar.
type DockerMemCell struct {
	percent   float64
	available bool
	asciiMode bool
}

// NewDockerMemCell creates a new Docker memory cell.
func NewDockerMemCell() *DockerMemCell {
	return &DockerMemCell{}
}

// SetPercent sets the Docker memory usage percentage (0-100).
func (c *DockerMemCell) SetPercent(percent float64) {
	c.percent = percent
}

// SetAvailable sets whether Docker is available.
func (c *DockerMemCell) SetAvailable(available bool) {
	c.available = available
}

// SetASCIIMode enables or disables ASCII-only output.
func (c *DockerMemCell) SetASCIIMode(ascii bool) {
	c.asciiMode = ascii
}

// Render returns the Docker memory usage display.
func (c *DockerMemCell) Render(width, height int) string {
	filled, empty := "●", "○"
	if c.asciiMode {
		filled, empty = "*", "o"
	}

	// When Docker is unavailable, show all empty dots
	if !c.available {
		return "DockerMem:" + strings.Repeat(empty, dockerMemTotalDots)
	}

	// Calculate filled dots (out of 11)
	filledCount := int(c.percent / 100.0 * dockerMemTotalDots)
	if filledCount > dockerMemTotalDots {
		filledCount = dockerMemTotalDots
	}

	var dots strings.Builder
	for i := 0; i < dockerMemTotalDots; i++ {
		if i < filledCount {
			dots.WriteString(filled)
		} else {
			dots.WriteString(empty)
		}
	}

	return "DockerMem:" + dots.String()
}

// ZoneID returns the mouse zone identifier.
func (c *DockerMemCell) ZoneID() string {
	return "res-dmem"
}
