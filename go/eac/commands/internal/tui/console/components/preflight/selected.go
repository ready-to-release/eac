package preflight

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ready-to-release/eac/go/eac/commands/internal/tui/console/components/shared"
)

// SelectedPanel displays details about the currently selected unit.
type SelectedPanel struct {
	unit    *shared.UnitDetails
	focused bool
}

// NewSelectedPanel creates a new selected unit panel.
func NewSelectedPanel() *SelectedPanel {
	return &SelectedPanel{}
}

// SetUnit updates the displayed unit details.
func (s *SelectedPanel) SetUnit(unit *shared.UnitDetails) {
	s.unit = unit
}

// Render renders the selected unit panel.
func (s *SelectedPanel) Render(width, height int) string {
	if s.unit == nil {
		return s.renderEmpty(height)
	}

	var lines []string

	// Unit name (bold)
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(shared.ColorWhite)
	lines = append(lines, nameStyle.Render(shared.Truncate(s.unit.Moniker, width-2)))
	lines = append(lines, "")

	// Details
	lines = append(lines, shared.LabelValue("Module", s.unit.Module))
	lines = append(lines, shared.LabelValue("Component", s.unit.Component))
	lines = append(lines, shared.LabelValue("Handler", s.unit.Handler))

	// Status with color
	statusStyle := s.statusStyle(s.unit.Status)
	lines = append(lines, shared.DimStyle.Render("Status: ")+statusStyle.Render(s.unit.Status.String()))

	// Weight
	lines = append(lines, shared.LabelValue("Weight", fmt.Sprintf("%d", s.unit.Weight)))

	// Duration (only if running or completed)
	if s.unit.Status == shared.UnitRunning || s.unit.Status >= shared.UnitSuccess {
		lines = append(lines, shared.LabelValue("Duration", shared.FormatDuration(s.unit.Duration)))
	}

	// Dependencies
	if len(s.unit.Dependencies) > 0 {
		lines = append(lines, "")
		lines = append(lines, shared.DimStyle.Render("Dependencies:"))
		for _, dep := range s.unit.Dependencies {
			lines = append(lines, "  "+shared.DimStyle.Render("- ")+dep)
		}
	}

	return shared.PadToHeight(strings.Join(lines, "\n"), height)
}

// renderEmpty renders the panel when no unit is selected.
func (s *SelectedPanel) renderEmpty(height int) string {
	content := shared.DimStyle.Render("No unit selected\n\nClick a unit in the grid\nor use arrow keys to navigate")
	return shared.PadToHeight(content, height)
}

// statusStyle returns the appropriate style for a unit status.
func (s *SelectedPanel) statusStyle(status shared.UnitStatus) lipgloss.Style {
	switch status {
	case shared.UnitRunning:
		return lipgloss.NewStyle().Foreground(shared.ColorOrange).Bold(true)
	case shared.UnitSuccess:
		return lipgloss.NewStyle().Foreground(shared.ColorGreen)
	case shared.UnitSkipped:
		return lipgloss.NewStyle().Foreground(shared.ColorCyan)
	case shared.UnitFailed:
		return lipgloss.NewStyle().Foreground(shared.ColorRed).Bold(true)
	default:
		return shared.DimStyle
	}
}

// Focus sets whether the panel has focus.
func (s *SelectedPanel) Focus(focused bool) {
	s.focused = focused
}

// IsFocused returns true if the panel has focus.
func (s *SelectedPanel) IsFocused() bool {
	return s.focused
}

// RenderWithBorder renders the panel with a border.
func (s *SelectedPanel) RenderWithBorder(width, height int, title string) string {
	innerWidth := width - 4
	innerHeight := height - 2

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	content := s.Render(innerWidth, innerHeight)

	borderColor := shared.ColorDim
	if s.focused {
		borderColor = shared.ColorCyan
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(innerWidth).
		Height(innerHeight)

	return style.Render(content)
}
