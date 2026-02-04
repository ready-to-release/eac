package cells

import (
	"fmt"
	"strings"
)

// UoWStatsCell displays the unit of work statistics.
// Format: UoW: remaining/total | ▶running/cap | ✓done ⏭cached ✗failed
type UoWStatsCell struct {
	total     int
	running   int
	capacity  int
	done      int
	cached    int
	failed    int
	asciiMode bool
}

// NewUoWStatsCell creates a new UoW stats cell.
func NewUoWStatsCell() *UoWStatsCell {
	return &UoWStatsCell{}
}

// SetStats sets all statistics values.
func (c *UoWStatsCell) SetStats(total, running, capacity, done, cached, failed int) {
	c.total = total
	c.running = running
	c.capacity = capacity
	c.done = done
	c.cached = cached
	c.failed = failed
}

// SetASCIIMode enables or disables ASCII-only output.
func (c *UoWStatsCell) SetASCIIMode(ascii bool) {
	c.asciiMode = ascii
}

// Render returns the stats display.
func (c *UoWStatsCell) Render(width, height int) string {
	remaining := c.total - c.done - c.cached - c.failed - c.running

	// Choose symbols based on mode
	runIcon := "▶"
	doneIcon := "✓"
	cachedIcon := "⏭"
	failedIcon := "✗"
	if c.asciiMode {
		runIcon = ">"
		doneIcon = "V"
		cachedIcon = "="
		failedIcon = "X"
	}

	return fmt.Sprintf("UoW: %d/%d | %s %2d/%d | %s %2d %s %2d %s %2d",
		remaining, c.total,
		runIcon, c.running, c.capacity,
		doneIcon, c.done,
		cachedIcon, c.cached,
		failedIcon, c.failed,
	)
}

// ZoneID returns the mouse zone identifier.
func (c *UoWStatsCell) ZoneID() string {
	return "res-uow"
}

// RenderLampsOnly returns just the work queue lamps.
// Format: "Work:●●●○○○○○○○" showing running out of capacity.
func (c *UoWStatsCell) RenderLampsOnly(width, height int) string {
	filled, empty := "●", "○"
	if c.asciiMode {
		filled, empty = "*", "o"
	}

	var dots strings.Builder
	for i := 0; i < c.capacity; i++ {
		if i < c.running {
			dots.WriteString(filled)
		} else {
			dots.WriteString(empty)
		}
	}

	return "Work:" + dots.String()
}

// RenderCountsOnly returns just the counts without lamps.
// Format: "3/8 P:5 F:0" (completed/total P:pending F:failed)
func (c *UoWStatsCell) RenderCountsOnly(width, height int) string {
	// Completed = done + cached (successfully completed, whether from cache or fresh)
	completed := c.done + c.cached
	// Pending = total - done - cached - failed - running
	pending := c.total - c.done - c.cached - c.failed - c.running
	if pending < 0 {
		pending = 0
	}

	return fmt.Sprintf("%d/%d P:%d F:%d",
		completed, c.total,
		pending, c.failed,
	)
}
