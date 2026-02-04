package cells

import (
	"fmt"
	"time"
)

// TimerCell displays elapsed time in MM:SS format.
type TimerCell struct {
	elapsed   time.Duration
	asciiMode bool
}

// NewTimerCell creates a new timer cell.
func NewTimerCell() *TimerCell {
	return &TimerCell{}
}

// SetElapsed sets the elapsed duration to display.
func (c *TimerCell) SetElapsed(d time.Duration) {
	c.elapsed = d
}

// SetASCIIMode enables or disables ASCII-only output.
func (c *TimerCell) SetASCIIMode(ascii bool) {
	c.asciiMode = ascii
}

// Render returns the timer display.
func (c *TimerCell) Render(width, height int) string {
	totalSeconds := int(c.elapsed.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("Time:%02d:%02d", minutes, seconds)
}

// ZoneID returns the mouse zone identifier.
func (c *TimerCell) ZoneID() string {
	return "res-timer"
}
