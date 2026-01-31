// Package mainarea provides components for the main area of the TUI.
package mainarea

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ready-to-release/eac/go/eac/adapters/tui/console/components/shared"
)

// Line represents a single output line with metadata.
type Line struct {
	Text      string
	Source    string // Module moniker or "system"
	Level     Level
	Timestamp time.Time
}

// Level represents the severity level of an output line.
type Level int

const (
	LevelInfo Level = iota
	LevelWarn
	LevelError
)

// Buffer interface for line storage (to avoid circular deps).
type Buffer interface {
	All() []Line
	Count() int
}

// OutputPanel displays live output with scrolling using bubbles/viewport.
type OutputPanel struct {
	vp         viewport.Model
	lines      []Line
	autoScroll bool
	unitName   string
	focused    bool
	width      int
	height     int
}

// NewOutputPanel creates a new output panel.
func NewOutputPanel(width, height int) *OutputPanel {
	vp := viewport.New(width, height-2) // -2 for header/footer
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 3

	return &OutputPanel{
		vp:         vp,
		autoScroll: true,
		width:      width,
		height:     height,
	}
}

// SetUnit switches to a different unit's output.
func (o *OutputPanel) SetUnit(name string, lines []Line) {
	o.unitName = name
	o.lines = lines
	o.refreshContent()
	if o.autoScroll {
		o.vp.GotoBottom()
	}
}

// AppendLine adds a line and refreshes the display.
func (o *OutputPanel) AppendLine(line Line) {
	o.lines = append(o.lines, line)
	o.refreshContent()

	if o.autoScroll {
		o.vp.GotoBottom()
	}
}

// SetLines replaces all lines.
func (o *OutputPanel) SetLines(lines []Line) {
	o.lines = lines
	o.refreshContent()
	if o.autoScroll {
		o.vp.GotoBottom()
	}
}

// refreshContent updates the viewport content.
func (o *OutputPanel) refreshContent() {
	var formatted []string
	for _, line := range o.lines {
		formatted = append(formatted, o.formatLine(line))
	}
	o.vp.SetContent(strings.Join(formatted, "\n"))
}

// formatLine formats a single line with prefix and styling.
func (o *OutputPanel) formatLine(line Line) string {
	var prefix string
	var style lipgloss.Style

	switch line.Level {
	case LevelError:
		prefix = shared.ErrorStyle.Render("|X ")
		style = shared.ErrorStyle
	case LevelWarn:
		prefix = shared.WarnStyle.Render("|! ")
		style = shared.WarnStyle
	default:
		prefix = shared.DimStyle.Render("|  ")
		style = lipgloss.NewStyle()
	}

	return prefix + style.Render(line.Text)
}

// Update handles messages for the output panel.
func (o *OutputPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "G": // Shift+G = go to bottom, enable auto-scroll
			o.vp.GotoBottom()
			o.autoScroll = true
			return nil
		case "g": // g = go to top, disable auto-scroll
			o.vp.GotoTop()
			o.autoScroll = false
			return nil
		}
	case tea.MouseMsg:
		// Any scroll disables auto-scroll
		if msg.Type == tea.MouseWheelUp || msg.Type == tea.MouseWheelDown {
			o.autoScroll = false
		}
	case tea.WindowSizeMsg:
		o.width = msg.Width
		o.height = msg.Height
		o.vp.Width = msg.Width
		o.vp.Height = msg.Height - 2
	}

	// Let viewport handle standard navigation
	o.vp, cmd = o.vp.Update(msg)

	// Re-enable auto-scroll if we're at bottom
	if o.vp.AtBottom() {
		o.autoScroll = true
	}

	return cmd
}

// Render renders the output panel.
func (o *OutputPanel) Render(width, height int) string {
	o.width = width
	o.height = height
	o.vp.Width = width
	o.vp.Height = height - 2

	var b strings.Builder

	// Header
	b.WriteString(o.renderHeader(width))
	b.WriteString("\n")

	// Viewport content
	b.WriteString(o.vp.View())
	b.WriteString("\n")

	// Footer
	b.WriteString(o.renderFooter(width))

	return b.String()
}

// renderHeader renders the panel header.
func (o *OutputPanel) renderHeader(width int) string {
	title := fmt.Sprintf(" Output: %s ", o.unitName)
	if o.unitName == "" {
		title = " Output "
	}
	if o.autoScroll {
		title += "[auto] "
	}

	titleLen := lipgloss.Width(title)
	borderLen := width - titleLen - 2
	if borderLen < 0 {
		borderLen = 0
	}

	border := shared.BorderStyle.Render(strings.Repeat("-", borderLen))
	return shared.BorderStyle.Render("+") + title + border + shared.BorderStyle.Render("+")
}

// renderFooter renders the panel footer with scroll position.
func (o *OutputPanel) renderFooter(width int) string {
	var indicator string

	if o.vp.AtTop() {
		indicator = " TOP "
	} else if o.vp.AtBottom() {
		indicator = " END "
	} else {
		percent := int(o.vp.ScrollPercent() * 100)
		indicator = fmt.Sprintf(" %d%% ", percent)
	}

	indicatorLen := len(indicator)
	borderLen := width - indicatorLen - 2
	if borderLen < 0 {
		borderLen = 0
	}

	border := shared.BorderStyle.Render(strings.Repeat("-", borderLen))
	return shared.BorderStyle.Render("+") + border + indicator + shared.BorderStyle.Render("+")
}

// Focus sets whether the panel has focus.
func (o *OutputPanel) Focus(focused bool) {
	o.focused = focused
}

// IsFocused returns true if the panel has focus.
func (o *OutputPanel) IsFocused() bool {
	return o.focused
}

// PageDown scrolls down one page.
func (o *OutputPanel) PageDown() {
	o.vp.ViewDown()
	o.autoScroll = o.vp.AtBottom()
}

// PageUp scrolls up one page.
func (o *OutputPanel) PageUp() {
	o.vp.ViewUp()
	o.autoScroll = false
}

// GotoTop scrolls to the top.
func (o *OutputPanel) GotoTop() {
	o.vp.GotoTop()
	o.autoScroll = false
}

// GotoBottom scrolls to the bottom and enables auto-scroll.
func (o *OutputPanel) GotoBottom() {
	o.vp.GotoBottom()
	o.autoScroll = true
}

// IsAutoScrolling returns true if auto-scroll is enabled.
func (o *OutputPanel) IsAutoScrolling() bool {
	return o.autoScroll
}

// SetAutoScroll sets the auto-scroll state.
func (o *OutputPanel) SetAutoScroll(enabled bool) {
	o.autoScroll = enabled
	if enabled {
		o.vp.GotoBottom()
	}
}
