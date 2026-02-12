// Package demo provides a redesigned TUI implementation with modular cell components.
package demo

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"

	"github.com/ready-to-release/eac/go/adapters/tui/demo/cells"
)

// Model is the Bubbletea model for the tui3 console.
type Model struct {
	// Display configuration
	width     int
	height    int
	asciiMode bool

	// Cells - all the visual components
	cells Cells

	// Core state
	startTime time.Time
	command   string   // "build", "test", "lint"
	modules   []string // Module list for command cell

	// UoW tracking
	uowTotal    int
	uowRunning  int
	uowCapacity int
	uowDone     int
	uowCached   int
	uowFailed   int

	// Module/unit state
	units    []cells.SelectorUnit
	selected string // Currently selected unit

	// Output lines
	outputLines []cells.OutputLine

	// Tools tracking
	plannedContainerTools []string
	activeContainerTools  []string
	plannedSystem         []string
	activeSystem          []string

	// System metrics (cached)
	cpuPercents       []float64
	memPercent        float64
	dockerMemPercent  float64
	dockerAvailable   bool
	lastMetricsUpdate time.Time

	// Summary data
	summaryData *cells.SummaryData

	// Quitting state
	quitting bool
}

// Cells holds all cell component instances.
type Cells struct {
	Timer      *cells.TimerCell
	CPU        *cells.CPUCell
	Mem        *cells.MemCell
	DockerMem  *cells.DockerMemCell
	Containers *cells.ContainersCell
	UoWStats   *cells.UoWStatsCell
	Tools      *cells.ToolsCell
	Command    *cells.CommandCell
	Selector   *cells.SelectorCell
	Selected   *cells.SelectedCell
	Deps       *cells.DepsCell
	Cache      *cells.CacheCell
	Artifacts  *cells.ArtifactsCell
	Output     *cells.OutputCell
	Summary    *cells.SummaryCell
}

// NewModel creates a new tui3 model.
func NewModel(width, height int, command string, asciiMode bool) Model {
	// Initialize zone manager for mouse click tracking
	zone.NewGlobal()

	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	m := Model{
		width:       width,
		height:      height,
		asciiMode:   asciiMode,
		command:     command,
		startTime:   time.Now(),
		uowCapacity: 16, // Default concurrency
	}

	// Initialize all cells
	m.initCells()

	return m
}

// initCells creates all cell instances.
func (m *Model) initCells() {
	m.cells = Cells{
		Timer:      cells.NewTimerCell(),
		CPU:        cells.NewCPUCell(),
		Mem:        cells.NewMemCell(),
		DockerMem:  cells.NewDockerMemCell(),
		Containers: cells.NewContainersCell(),
		UoWStats:   cells.NewUoWStatsCell(),
		Tools:      cells.NewToolsCell(),
		Command:    cells.NewCommandCell(),
		Selector:   cells.NewSelectorCell(),
		Selected:   cells.NewSelectedCell(),
		Deps:       cells.NewDepsCell(),
		Cache:      cells.NewCacheCell(),
		Artifacts:  cells.NewArtifactsCell(),
		Output:     cells.NewOutputCell(),
		Summary:    cells.NewSummaryCell(),
	}

	// Set ASCII mode on cells that support it
	m.cells.Timer.SetASCIIMode(m.asciiMode)
	m.cells.CPU.SetASCIIMode(m.asciiMode)
	m.cells.Mem.SetASCIIMode(m.asciiMode)
	m.cells.DockerMem.SetASCIIMode(m.asciiMode)
	m.cells.Containers.SetASCIIMode(m.asciiMode)
	m.cells.UoWStats.SetASCIIMode(m.asciiMode)
	m.cells.Tools.SetASCIIMode(m.asciiMode)
	m.cells.Selector.SetASCIIMode(m.asciiMode)
	m.cells.Output.SetASCIIMode(m.asciiMode)
}

// Init initializes the model and returns initial commands.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
	)
}

// tickCmd returns a tick every 100ms for time updates.
func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
