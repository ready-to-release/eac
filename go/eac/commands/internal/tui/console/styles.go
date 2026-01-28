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

	// Tab styles
	TabActive    lipgloss.Style // Currently selected tab
	TabPending   lipgloss.Style // Scheduled, waiting for slot
	TabRunning   lipgloss.Style // Running module tab
	TabComplete  lipgloss.Style // Completed module tab
	TabSkipped   lipgloss.Style // Skipped (cached) module tab
	TabFailed    lipgloss.Style // Failed module tab
	TabDim       lipgloss.Style // Muted/inactive tab
	TabHover     lipgloss.Style // Tab under mouse cursor
	TabBar       lipgloss.Style // Tab bar container
	TabSeparator lipgloss.Style // Separator between tabs
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

	// Tab styles - sharp, high contrast
	// Grey for pending (not started)
	TabActive: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).  // Bright white
		Background(lipgloss.Color("24")),  // Deep blue - selected
	TabPending: lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")). // Grey text
		Background(lipgloss.Color("236")), // Dark bg
	// Orange for running (active)
	TabRunning: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("208")). // Orange
		Background(lipgloss.Color("236")), // Dark bg
	// Green for complete (success)
	TabComplete: lipgloss.NewStyle().
		Foreground(lipgloss.Color("40")).  // Green
		Background(lipgloss.Color("236")), // Dark bg
	// Blue for skipped (cached)
	TabSkipped: lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).  // Cyan/Blue
		Background(lipgloss.Color("236")), // Dark bg
	// Red for failed
	TabFailed: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")). // Red
		Background(lipgloss.Color("52")),  // Dark red bg
	TabDim: lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // Dim grey
		Background(lipgloss.Color("234")), // Very dark bg
	TabHover: lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).  // Bright white
		Background(lipgloss.Color("240")).  // Lighter bg on hover
		Bold(true),
	TabBar: lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")),
	TabSeparator: lipgloss.NewStyle().
		Foreground(lipgloss.Color("238")),
}

// Icons for different states (ASCII-safe for Windows terminals).
const (
	IconPass    = "V"
	IconFail    = "X"
	IconWarn    = "!"
	IconInfo    = " "
	IconRunning = ">"
	IconPaused  = "="
)
