package tui3

import (
	"time"

	"github.com/ready-to-release/eac/go/adapters/tui/tui3/cells"
)

// TickMsg is sent periodically for time-based updates.
type TickMsg time.Time

// WindowSizeMsg is sent when the terminal is resized.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// UnitStartMsg indicates a unit has started execution.
type UnitStartMsg struct {
	Moniker     string // Globally unique ID (Longname) for matching
	DisplayName string // Short name for tab display
	Weight      int
}

// UnitRunningMsg indicates a unit has acquired a slot and is running.
type UnitRunningMsg struct {
	Moniker string
}

// UnitCompleteMsg indicates a unit has finished.
type UnitCompleteMsg struct {
	Moniker  string
	ExitCode int // 0=success, <0=skipped, >0=failed
}

// OutputLineMsg adds a line to the output.
type OutputLineMsg struct {
	Line cells.OutputLine
}

// LayerUpdateMsg updates the current layer.
type LayerUpdateMsg struct {
	Current int
	Total   int
}

// MetricsUpdateMsg updates system metrics.
type MetricsUpdateMsg struct {
	CPUPercents []float64
	MemPercent  float64
}

// ToolActiveMsg indicates a tool has become active.
type ToolActiveMsg struct {
	Name        string
	IsContainer bool
}

// ToolInactiveMsg indicates a tool has become inactive.
type ToolInactiveMsg struct {
	Name        string
	IsContainer bool
}

// SummaryMsg provides the final execution summary.
type SummaryMsg struct {
	Data *cells.SummaryData
}

// SetLayersMsg sets the layer organization for the selector.
type SetLayersMsg struct {
	Layers [][]string
}

// QuitMsg requests the TUI to exit.
type QuitMsg struct{}
