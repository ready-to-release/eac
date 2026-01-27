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
	TabFailed    lipgloss.Style // Failed module tab
	TabDim       lipgloss.Style // Muted/inactive tab
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

	// Tab styles - 3D raised/pressed effect via background colors only
	// Active tab: pressed/sunken look - dark background (appears recessed)
	TabActive: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).  // Bright white text
		Background(lipgloss.Color("17")).  // Dark blue - pressed into content
		Padding(0, 1),
	// Inactive tabs: raised look - lighter background (appears raised)
	TabPending: lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")). // Light gray text - waiting
		Background(lipgloss.Color("242")). // Light gray bg - raised
		Padding(0, 1),
	TabRunning: lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")). // Orange text
		Background(lipgloss.Color("242")). // Light gray bg - raised
		Padding(0, 1),
	TabComplete: lipgloss.NewStyle().
		Foreground(lipgloss.Color("71")).  // Green text
		Background(lipgloss.Color("242")). // Light gray bg - raised
		Padding(0, 1),
	TabFailed: lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red text
		Background(lipgloss.Color("242")). // Light gray bg - raised
		Bold(true).
		Padding(0, 1),
	TabDim: lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")). // Muted gray text
		Background(lipgloss.Color("242")). // Light gray bg - raised
		Padding(0, 1),
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
