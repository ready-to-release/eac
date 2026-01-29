// Package lights provides a single-row status lights panel for the TUI.
package lights

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ready-to-release/eac/go/eac/commands/internal/tui/console/components/shared"
)

// UnitLight represents a single unit's status in the lights panel.
type UnitLight struct {
	Moniker string
	Status  shared.UnitStatus
}

// Panel displays a single row of status lights, one per unit.
type Panel struct {
	units     []UnitLight
	asciiMode bool
}

// NewPanel creates a new lights panel.
func NewPanel(asciiMode bool) *Panel {
	return &Panel{
		asciiMode: asciiMode,
	}
}

// SetUnits updates the units displayed in the lights panel.
func (p *Panel) SetUnits(units []UnitLight) {
	p.units = units
}

// AddUnit adds a unit to the lights panel.
func (p *Panel) AddUnit(moniker string, status shared.UnitStatus) {
	p.units = append(p.units, UnitLight{
		Moniker: moniker,
		Status:  status,
	})
}

// UpdateUnit updates the status of a unit by moniker.
func (p *Panel) UpdateUnit(moniker string, status shared.UnitStatus) {
	for i := range p.units {
		if p.units[i].Moniker == moniker {
			p.units[i].Status = status
			return
		}
	}
	// If not found, add it
	p.AddUnit(moniker, status)
}

// Render renders the lights panel.
func (p *Panel) Render(width, height int) string {
	if len(p.units) == 0 {
		return p.renderEmpty(width)
	}

	var dots strings.Builder
	for _, unit := range p.units {
		dots.WriteString(shared.StatusDot(unit.Status, p.asciiMode))
	}

	content := dots.String()

	// Center in available width
	contentWidth := lipgloss.Width(content)
	padding := (width - contentWidth) / 2
	if padding > 0 {
		content = strings.Repeat(" ", padding) + content
	}

	// Wrap with simple border
	return p.borderWrap(content, width)
}

// renderEmpty renders an empty lights panel.
func (p *Panel) renderEmpty(width int) string {
	content := shared.DimStyle.Render("No units")
	contentWidth := lipgloss.Width(content)
	padding := (width - contentWidth) / 2
	if padding > 0 {
		content = strings.Repeat(" ", padding) + content
	}
	return p.borderWrap(content, width)
}

// borderWrap adds a simple top/bottom border.
func (p *Panel) borderWrap(content string, width int) string {
	border := shared.BorderStyle.Render(strings.Repeat("─", width))
	return border + "\n" + content + "\n" + border
}

// CountByStatus returns the count of units with each status.
func (p *Panel) CountByStatus() map[shared.UnitStatus]int {
	counts := make(map[shared.UnitStatus]int)
	for _, unit := range p.units {
		counts[unit.Status]++
	}
	return counts
}

// Summary returns a compact summary string like "5/10 (2 running)".
func (p *Panel) Summary() string {
	counts := p.CountByStatus()
	total := len(p.units)
	done := counts[shared.UnitSuccess] + counts[shared.UnitSkipped] + counts[shared.UnitFailed]
	running := counts[shared.UnitRunning]

	if running > 0 {
		return strings.Join([]string{
			lipgloss.NewStyle().Foreground(shared.ColorGreen).Render(
				strings.Repeat("V", done),
			),
			lipgloss.NewStyle().Foreground(shared.ColorOrange).Render(
				strings.Repeat(">", running),
			),
			shared.DimStyle.Render(
				strings.Repeat("o", total-done-running),
			),
		}, "")
	}

	return shared.DimStyle.Render(strings.Repeat("o", total-done)) +
		lipgloss.NewStyle().Foreground(shared.ColorGreen).Render(strings.Repeat("V", done))
}
