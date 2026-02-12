package demo

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/ready-to-release/eac/go/adapters/tui/demo/cells"
)

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
	m.cells.DockerMem.SetPercent(m.dockerMemPercent)
	m.cells.DockerMem.SetAvailable(m.dockerAvailable)

	// Containers
	m.cells.Containers.SetContainers(m.activeContainerTools, m.plannedContainerTools)

	// UoW Stats
	m.cells.UoWStats.SetStats(m.uowTotal, m.uowRunning, m.uowCapacity, m.uowDone, m.uowCached, m.uowFailed)

	// Tools
	m.cells.Tools.SetTools(m.plannedContainerTools, m.activeContainerTools, m.plannedSystem, m.activeSystem)

	// Command
	m.cells.Command.SetCommand(m.command)
	m.cells.Command.SetModules(m.modules)

	// Selector
	m.cells.Selector.SetUnits(m.units)
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

// SetDockerMemPercent sets the Docker memory usage percentage.
func (m *Model) SetDockerMemPercent(percent float64) {
	m.dockerMemPercent = percent
}

// SetDockerAvailable sets whether Docker is available.
func (m *Model) SetDockerAvailable(available bool) {
	m.dockerAvailable = available
}

// SetPlannedTools sets the planned tool lists.
func (m *Model) SetPlannedTools(containers, system []string) {
	m.plannedContainerTools = containers
	m.plannedSystem = system
}

// SetActiveTools sets the active tool lists.
func (m *Model) SetActiveTools(containers, system []string) {
	m.activeContainerTools = containers
	m.activeSystem = system
}

// SetSummary sets the summary data.
func (m *Model) SetSummary(data *cells.SummaryData) {
	m.summaryData = data
}
