package shared

import "github.com/charmbracelet/lipgloss"

// Colors used throughout the TUI components.
var (
	ColorDim     = lipgloss.Color("240")
	ColorGrey    = lipgloss.Color("245")
	ColorWhite   = lipgloss.Color("15")
	ColorYellow  = lipgloss.Color("226")
	ColorOrange  = lipgloss.Color("208")
	ColorGreen   = lipgloss.Color("40")
	ColorCyan    = lipgloss.Color("39")
	ColorRed     = lipgloss.Color("196")
	ColorBlue    = lipgloss.Color("24")
	ColorDarkRed = lipgloss.Color("52")
	ColorDarkBg  = lipgloss.Color("236")
	ColorDimBg   = lipgloss.Color("234")
)

// Base styles used by components.
var (
	// Text styles
	DimStyle    = lipgloss.NewStyle().Foreground(ColorDim)
	BoldStyle   = lipgloss.NewStyle().Bold(true)
	ErrorStyle  = lipgloss.NewStyle().Foreground(ColorRed)
	WarnStyle   = lipgloss.NewStyle().Foreground(ColorOrange)
	SuccessStyle = lipgloss.NewStyle().Foreground(ColorGreen)

	// Border styles
	BorderStyle = lipgloss.NewStyle().Foreground(ColorDim)
)

// StatusBorderColor returns the border color for a unit status.
func StatusBorderColor(status UnitStatus) lipgloss.Color {
	switch status {
	case UnitPending:
		return ColorDim
	case UnitQueued:
		return ColorYellow
	case UnitRunning:
		return ColorOrange
	case UnitSuccess:
		return ColorGreen
	case UnitSkipped:
		return ColorCyan
	case UnitFailed:
		return ColorRed
	default:
		return ColorDim
	}
}

// StatusDot returns a colored dot for the given status.
func StatusDot(status UnitStatus, asciiMode bool) string {
	filled, empty := "●", "○"
	if asciiMode {
		filled, empty = "*", "o"
	}

	switch status {
	case UnitPending:
		return DimStyle.Render(empty)
	case UnitQueued:
		return lipgloss.NewStyle().Foreground(ColorYellow).Render(filled)
	case UnitRunning:
		return lipgloss.NewStyle().Foreground(ColorOrange).Render(filled)
	case UnitSuccess:
		return lipgloss.NewStyle().Foreground(ColorGreen).Render(filled)
	case UnitSkipped:
		return lipgloss.NewStyle().Foreground(ColorCyan).Render(filled)
	case UnitFailed:
		return lipgloss.NewStyle().Foreground(ColorRed).Render(filled)
	default:
		return DimStyle.Render("?")
	}
}
