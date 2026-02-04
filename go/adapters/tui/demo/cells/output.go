package cells

import (
	"strings"
)

// OutputCell displays log output with scrolling support.
type OutputCell struct {
	lines        []OutputLine
	scrollOffset int  // Lines from the bottom (0 = show most recent)
	autoScroll   bool // Auto-scroll to bottom on new lines
	asciiMode    bool
}

// NewOutputCell creates a new output cell.
func NewOutputCell() *OutputCell {
	return &OutputCell{
		autoScroll: true,
	}
}

// SetLines sets all output lines.
func (c *OutputCell) SetLines(lines []OutputLine) {
	c.lines = lines
	if c.autoScroll {
		c.scrollOffset = 0
	}
}

// AppendLine adds a line to the output.
func (c *OutputCell) AppendLine(line OutputLine) {
	c.lines = append(c.lines, line)
	if c.autoScroll {
		c.scrollOffset = 0
	}
}

// SetScrollOffset sets the scroll offset (lines from bottom).
func (c *OutputCell) SetScrollOffset(offset int) {
	c.scrollOffset = offset
	if offset > 0 {
		c.autoScroll = false
	}
}

// SetAutoScroll enables or disables auto-scroll.
func (c *OutputCell) SetAutoScroll(enabled bool) {
	c.autoScroll = enabled
	if enabled {
		c.scrollOffset = 0
	}
}

// IsAutoScrolling returns true if auto-scroll is enabled.
func (c *OutputCell) IsAutoScrolling() bool {
	return c.autoScroll
}

// SetASCIIMode enables or disables ASCII-only output.
func (c *OutputCell) SetASCIIMode(ascii bool) {
	c.asciiMode = ascii
}

// Render returns the output display.
func (c *OutputCell) Render(width, height int) string {
	if len(c.lines) == 0 {
		return ""
	}

	// Calculate which lines to show
	visibleCount := height
	if visibleCount > len(c.lines) {
		visibleCount = len(c.lines)
	}

	// Get the range of lines to display
	// scrollOffset = 0 means show the most recent lines
	startIdx := len(c.lines) - visibleCount - c.scrollOffset
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + visibleCount
	if endIdx > len(c.lines) {
		endIdx = len(c.lines)
	}

	// Format visible lines
	var formatted []string
	for i := startIdx; i < endIdx; i++ {
		formatted = append(formatted, c.formatLine(c.lines[i], width))
	}

	// Pad to fill height
	for len(formatted) < height {
		formatted = append(formatted, "")
	}

	return strings.Join(formatted, "\n")
}

// formatLine formats a single line with level prefix.
func (c *OutputCell) formatLine(line OutputLine, width int) string {
	var prefix string

	switch line.Level {
	case LineLevelError:
		prefix = "|X "
	case LineLevelWarn:
		prefix = "|! "
	default:
		prefix = "|  "
	}

	text := line.Text
	maxLen := width - len(prefix)
	if maxLen > 0 && len(text) > maxLen {
		text = text[:maxLen-3] + "..."
	}

	return prefix + text
}

// ScrollUp scrolls up by the given number of lines.
func (c *OutputCell) ScrollUp(lines int) {
	c.scrollOffset += lines
	maxOffset := len(c.lines) - 1
	if c.scrollOffset > maxOffset {
		c.scrollOffset = maxOffset
	}
	c.autoScroll = false
}

// ScrollDown scrolls down by the given number of lines.
func (c *OutputCell) ScrollDown(lines int) {
	c.scrollOffset -= lines
	if c.scrollOffset < 0 {
		c.scrollOffset = 0
		c.autoScroll = true
	}
}

// GotoTop scrolls to the top (oldest lines).
func (c *OutputCell) GotoTop() {
	c.scrollOffset = len(c.lines) - 1
	c.autoScroll = false
}

// GotoBottom scrolls to the bottom (newest lines).
func (c *OutputCell) GotoBottom() {
	c.scrollOffset = 0
	c.autoScroll = true
}

// ZoneID returns empty string (not an interactive zone).
func (c *OutputCell) ZoneID() string {
	return ""
}
