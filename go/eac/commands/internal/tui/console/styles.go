package console

import "github.com/charmbracelet/lipgloss"

// Styles holds all the lipgloss styles for the console.
var Styles = struct {
	// Header styles
	Phase   lipgloss.Style
	Time    lipgloss.Style
	Running lipgloss.Style
	Paused  lipgloss.Style
	Counter lipgloss.Style

	// Line prefix styles
	InfoPrefix  lipgloss.Style
	WarnPrefix  lipgloss.Style
	ErrorPrefix lipgloss.Style

	// Line content styles
	Info  lipgloss.Style
	Warn  lipgloss.Style
	Error lipgloss.Style
	Dim   lipgloss.Style

	// Border styles
	Border lipgloss.Style
}{
	Phase:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),  // Cyan
	Time:    lipgloss.NewStyle().Foreground(lipgloss.Color("245")),            // Gray
	Running: lipgloss.NewStyle().Foreground(lipgloss.Color("214")),            // Orange
	Paused:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226")), // Yellow
	Counter: lipgloss.NewStyle().Foreground(lipgloss.Color("250")),            // Light gray

	InfoPrefix:  lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
	WarnPrefix:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
	ErrorPrefix: lipgloss.NewStyle().Foreground(lipgloss.Color("196")),

	Info:  lipgloss.NewStyle(),
	Warn:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
	Error: lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
	Dim:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")),

	Border: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
}

// Icons for different states.
const (
	IconPass    = "✓"
	IconFail    = "✗"
	IconWarn    = "!"
	IconInfo    = " "
	IconRunning = "▶"
	IconPaused  = "⏸"
)
