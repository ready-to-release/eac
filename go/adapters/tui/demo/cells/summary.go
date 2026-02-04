package cells

import (
	"fmt"
	"time"
)

// SummaryData holds the execution summary statistics.
type SummaryData struct {
	Total     int
	Succeeded int
	Failed    int
	Skipped   int
	Duration  time.Duration
}

// SummaryCell displays the final execution summary.
type SummaryCell struct {
	data *SummaryData
}

// NewSummaryCell creates a new summary cell.
func NewSummaryCell() *SummaryCell {
	return &SummaryCell{}
}

// SetData sets the summary data.
func (c *SummaryCell) SetData(data *SummaryData) {
	c.data = data
}

// Render returns the summary display.
func (c *SummaryCell) Render(width, height int) string {
	if c.data == nil {
		return ""
	}

	// Format duration
	durStr := formatDuration(c.data.Duration)

	return fmt.Sprintf("Total: %d | Succeeded: %d | Failed: %d | Skipped: %d | Duration: %s",
		c.data.Total, c.data.Succeeded, c.data.Failed, c.data.Skipped, durStr)
}

// ZoneID returns empty string (not an interactive zone).
func (c *SummaryCell) ZoneID() string {
	return ""
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	if secs == 0 {
		return fmt.Sprintf("%dm", mins)
	}
	return fmt.Sprintf("%dm%02ds", mins, secs)
}
