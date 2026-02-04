package demo

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ready-to-release/eac/go/adapters/tui/demo/cells"
)

// Update handles messages and returns updated model and commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TickMsg:
		// Update system metrics (CPU/memory)
		m.updateMetrics()
		// Update cells with current state
		m.updateCells()
		// Schedule next tick
		cmds = append(cmds, m.tickCmd())

	case UnitStartMsg:
		m.AddUnit(msg.Moniker, msg.DisplayName, cells.UnitPending, msg.Weight)
		m.updateCells() // Sync cells immediately for responsive UI

	case UnitRunningMsg:
		m.UpdateUnit(msg.Moniker, cells.UnitRunning)
		m.updateCells() // Sync cells immediately for responsive UI

	case UnitCompleteMsg:
		status := statusFromExitCode(msg.ExitCode)
		m.UpdateUnit(msg.Moniker, status)
		m.updateCells() // Sync cells immediately for responsive UI

	case OutputLineMsg:
		m.AppendOutput(msg.Line)

	case MetricsUpdateMsg:
		m.SetCPUPercents(msg.CPUPercents)
		m.SetMemPercent(msg.MemPercent)
		m.SetDockerMemPercent(msg.DockerMemPercent)
		m.SetDockerAvailable(msg.DockerAvailable)

	case ToolActiveMsg:
		if msg.IsContainer {
			m.activeContainerTools = addToSlice(m.activeContainerTools, msg.Name)
		} else {
			m.activeSystem = addToSlice(m.activeSystem, msg.Name)
		}

	case ToolInactiveMsg:
		if msg.IsContainer {
			m.activeContainerTools = removeFromSlice(m.activeContainerTools, msg.Name)
		} else {
			m.activeSystem = removeFromSlice(m.activeSystem, msg.Name)
		}

	case SummaryMsg:
		m.SetSummary(msg.Data)
		m.updateCells() // Sync cells immediately

	case SetModulesMsg:
		m.SetModules(msg.Modules)
		m.updateCells() // Sync cells immediately

	case QuitMsg:
		m.quitting = true
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}

// statusFromExitCode converts an exit code to a unit status.
func statusFromExitCode(exitCode int) cells.UnitStatus {
	if exitCode == 0 {
		return cells.UnitComplete
	} else if exitCode < 0 {
		return cells.UnitSkipped
	}
	return cells.UnitFailed
}

// addToSlice adds a string to a slice if not already present.
func addToSlice(slice []string, s string) []string {
	for _, item := range slice {
		if item == s {
			return slice
		}
	}
	return append(slice, s)
}

// removeFromSlice removes a string from a slice.
func removeFromSlice(slice []string, s string) []string {
	var result []string
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
