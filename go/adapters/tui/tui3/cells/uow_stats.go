package cells

import "fmt"

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
