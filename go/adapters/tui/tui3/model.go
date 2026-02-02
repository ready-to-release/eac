// Package tui3 provides a redesigned TUI implementation with modular cell components.
package tui3

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/ready-to-release/eac/go/adapters/tui/tui3/cells"
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
	startTime   time.Time
	command     string   // "build", "test", "lint"
	modules     []string // Module list for command cell
	layer       int      // Current execution layer
	totalLayers int      // Total layers

	// UoW tracking
	uowTotal    int
	uowRunning  int
	uowCapacity int
	uowDone     int
	uowCached   int
	uowFailed   int

	// Module/unit state
	units    []cells.SelectorUnit
	layers   [][]string // Layer organization for selector
	selected string     // Currently selected unit

	// Output lines
	outputLines []cells.OutputLine

	// Tools tracking
	plannedContainers []string
	activeContainers  []string
	plannedSystem     []string
	activeSystem      []string

	// System metrics (cached)
	cpuPercents       []float64
	memPercent        float64
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
	Containers *cells.ContainersCell
	UoWStats   *cells.UoWStatsCell
	Tools      *cells.ToolsCell
	Layer      *cells.LayerCell
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
		Containers: cells.NewContainersCell(),
		UoWStats:   cells.NewUoWStatsCell(),
		Tools:      cells.NewToolsCell(),
		Layer:      cells.NewLayerCell(),
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

// updateCells syncs all cell state from model state.
func (m *Model) updateCells() {
	// Derive UoW stats from unit state (single source of truth)
	// Running is tracked as total weight (not count) for resource display
	var running, done, cached, failed int
	for _, unit := range m.units {
		switch unit.Status {
		case cells.UnitRunning:
			// Use weight for running units so display matches capacity
			w := unit.Weight
			if w <= 0 {
				w = 1
			}
			running += w
		case cells.UnitComplete:
			done++
		case cells.UnitSkipped:
			cached++
		case cells.UnitFailed:
			failed++
		}
	}
	m.uowTotal = len(m.units)
	m.uowRunning = running
	m.uowDone = done
	m.uowCached = cached
	m.uowFailed = failed

	// Timer
	m.cells.Timer.SetElapsed(time.Since(m.startTime))

	// CPU/Memory
	m.cells.CPU.SetPercents(m.cpuPercents)
	m.cells.Mem.SetPercent(m.memPercent)

	// Containers
	m.cells.Containers.SetContainers(m.activeContainers, m.plannedContainers)

	// UoW Stats
	m.cells.UoWStats.SetStats(m.uowTotal, m.uowRunning, m.uowCapacity, m.uowDone, m.uowCached, m.uowFailed)

	// Tools
	m.cells.Tools.SetTools(m.plannedContainers, m.activeContainers, m.plannedSystem, m.activeSystem)

	// Layer
	m.cells.Layer.SetLayer(m.layer, m.totalLayers)

	// Command
	m.cells.Command.SetCommand(m.command)
	m.cells.Command.SetModules(m.modules)

	// Selector
	m.cells.Selector.SetUnits(m.units)
	m.cells.Selector.SetLayers(m.layers)
	m.cells.Selector.SetSelected(m.selected)

	// Output
	m.cells.Output.SetLines(m.outputLines)

	// Summary - update with derived counts from units (single source of truth)
	if m.summaryData != nil {
		m.summaryData.Total = m.uowTotal
		m.summaryData.Succeeded = done
		m.summaryData.Failed = failed
		m.summaryData.Skipped = cached
	}
	m.cells.Summary.SetData(m.summaryData)
}

// metricsUpdateInterval is how often to refresh CPU/memory metrics.
// These gopsutil calls are expensive (100-500ms on Windows), so we cache them.
const metricsUpdateInterval = 500 * time.Millisecond

// updateMetrics refreshes the cached CPU and memory metrics.
// Should be called from the tick handler, not from View().
func (m *Model) updateMetrics() {
	// Skip if updated recently
	if time.Since(m.lastMetricsUpdate) < metricsUpdateInterval {
		return
	}

	// Update CPU metrics (this is the slow call)
	if perCore, err := cpu.Percent(0, true); err == nil {
		m.cpuPercents = perCore
	}

	// Update memory metrics
	if memInfo, err := mem.VirtualMemory(); err == nil {
		m.memPercent = memInfo.UsedPercent
	}

	m.lastMetricsUpdate = time.Now()
}

// SetModules sets the module list for display.
func (m *Model) SetModules(modules []string) {
	m.modules = modules
}

// SetLayer sets the current layer.
func (m *Model) SetLayer(current, total int) {
	m.layer = current
	m.totalLayers = total
}

// SetUoWStats sets the UoW statistics.
func (m *Model) SetUoWStats(total, running, capacity, done, cached, failed int) {
	m.uowTotal = total
	m.uowRunning = running
	m.uowCapacity = capacity
	m.uowDone = done
	m.uowCached = cached
	m.uowFailed = failed
}

// AddUnit adds a unit to the selector, or updates if already exists.
func (m *Model) AddUnit(moniker, displayName string, status cells.UnitStatus, weight int) {
	// Check if unit already exists to prevent duplicate registration
	for i := range m.units {
		if m.units[i].Moniker == moniker {
			m.units[i].DisplayName = displayName
			m.units[i].Status = status
			if weight > 0 {
				m.units[i].Weight = weight
			}
			return
		}
	}
	m.units = append(m.units, cells.SelectorUnit{
		Moniker:     moniker,
		DisplayName: displayName,
		Status:      status,
		Weight:      weight,
	})
}

// UpdateUnit updates a unit's status.
func (m *Model) UpdateUnit(moniker string, status cells.UnitStatus) {
	for i := range m.units {
		if m.units[i].Moniker == moniker {
			m.units[i].Status = status
			return
		}
	}
}

// SetLayers sets the layer organization for the selector.
func (m *Model) SetLayers(layers [][]string) {
	m.layers = layers
}

// SetSelected sets the currently selected unit.
func (m *Model) SetSelected(moniker string) {
	m.selected = moniker
}

// AppendOutput adds an output line.
func (m *Model) AppendOutput(line cells.OutputLine) {
	m.outputLines = append(m.outputLines, line)
}

// SetCPUPercents sets the CPU usage percentages.
func (m *Model) SetCPUPercents(percents []float64) {
	m.cpuPercents = percents
}

// SetMemPercent sets the memory usage percentage.
func (m *Model) SetMemPercent(percent float64) {
	m.memPercent = percent
}

// SetPlannedTools sets the planned tool lists.
func (m *Model) SetPlannedTools(containers, system []string) {
	m.plannedContainers = containers
	m.plannedSystem = system
}

// SetActiveTools sets the active tool lists.
func (m *Model) SetActiveTools(containers, system []string) {
	m.activeContainers = containers
	m.activeSystem = system
}

// SetSummary sets the summary data.
func (m *Model) SetSummary(data *cells.SummaryData) {
	m.summaryData = data
}
