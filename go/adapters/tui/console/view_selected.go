package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderSelectedHeader returns a single line for the logs panel header:
// - UoW details (Unit/Tool/Status) when a component tab is hovered/selected
// - Empty string when nothing is hovered (caller falls back to phase header)
func (m Model) renderSelectedHeader(contentWidth int) string {
	padTo := func(s string, width int) string {
		visWidth := lipgloss.Width(s)
		if visWidth >= width {
			return s
		}
		return s + strings.Repeat(" ", width-visWidth)
	}

	var line string
	if strings.HasPrefix(m.Interaction.HoveredZone, "tab:") || m.Interaction.HoveredTab != "" {
		activeComponent := m.Interaction.HoveredTab
		if activeComponent == "" {
			activeComponent = m.getEffectiveActiveTab()
		}
		line = m.renderSelectedUoW(activeComponent, contentWidth)
	}

	return padTo(line, contentWidth)
}

// renderSelectedUoW renders UoW details for the selected/hovered component tab.
func (m Model) renderSelectedUoW(activeComponent string, contentWidth int) string {
	if activeComponent == "" {
		return ""
	}

	// Fixed column widths matching Resources pane
	const (
		col1Width = 38 // Unit
		col2Width = 24 // Runner/Tool
		col3Width = 20 // Status
	)

	// Helper to truncate with ellipsis
	truncate := func(s string, maxLen int) string {
		if len(s) <= maxLen {
			return s
		}
		if maxLen <= 3 {
			return s[:maxLen]
		}
		return s[:maxLen-3] + "..."
	}

	// Helper to pad string to fixed width
	padTo := func(s string, width int) string {
		visWidth := lipgloss.Width(s)
		if visWidth >= width {
			return s
		}
		return s + strings.Repeat(" ", width-visWidth)
	}

	// Semantic styles for status display
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))   // White for labels
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))  // Gold/amber for active unit
	pendingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // Gray for pending
	runningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))  // Blue for running
	completeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Green for complete
	cachedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))  // Purple for cached
	failedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))  // Red for failed
	sep := Styles.Dim.Render(" │ ")

	var col1, col2, col3 string

	// Get status of the active component
	var statusStr string
	var statusStyle lipgloss.Style
	if state, exists := m.Execution.UoWStates[activeComponent]; exists {
		switch state.Status {
		case UoWPending:
			statusStr = "pending"
			statusStyle = pendingStyle
		case UoWRunning:
			statusStr = "running"
			statusStyle = runningStyle
			if !state.StartTime.IsZero() {
				dur := time.Since(state.StartTime)
				statusStr = fmt.Sprintf("running %s", formatElapsed(dur))
			}
		case UoWComplete:
			statusStr = "done"
			statusStyle = completeStyle
			if !state.StartTime.IsZero() && !state.EndTime.IsZero() {
				dur := state.EndTime.Sub(state.StartTime)
				statusStr = fmt.Sprintf("done %s", formatElapsed(dur))
			}
		case UoWSkipped:
			statusStr = "cached"
			statusStyle = cachedStyle
		case UoWFailed:
			statusStr = "failed"
			statusStyle = failedStyle
			if state.ExitCode != 0 {
				statusStr = fmt.Sprintf("failed (exit %d)", state.ExitCode)
			}
		}
	}

	// Determine tool label based on operation type
	toolLabel := "Tool"
	switch m.Display.RunPhaseName {
	case "building", "Building":
		toolLabel = "Builder"
	case "testing", "Testing":
		toolLabel = "Runner"
	case "linting", "Linting":
		toolLabel = "Linter"
	case "scanning", "Scanning":
		toolLabel = "Scanner"
	}

	// Use structured fields from UoWState
	var unitName, toolName string
	if state, exists := m.Execution.UoWStates[activeComponent]; exists {
		if state.Module != "" && state.Component != "" {
			unitName = state.Module + ":" + state.Component
		} else if state.Module != "" {
			unitName = state.Module
		} else {
			unitName = state.DisplayName
		}
		toolName = state.Tool
	} else {
		unitName = activeComponent
	}

	// Col1: Unit
	col1 = labelStyle.Render("Unit:") + activeStyle.Render(truncate(unitName, col1Width-5))

	// Col2: Runner/Tool
	if toolName != "" {
		col2 = labelStyle.Render(toolLabel+":") + activeStyle.Render(truncate(toolName, col2Width-len(toolLabel)-1))
	}

	// Col3: Status
	if statusStr != "" {
		col3 = labelStyle.Render("Status:") + statusStyle.Render(truncate(statusStr, col3Width-7))
	}

	col1 = padTo(col1, col1Width)
	col2 = padTo(col2, col2Width)
	col3 = padTo(col3, col3Width)

	return col1 + sep + col2 + sep + col3
}

// renderSelectedHelp renders help text for a hovered resource element.
// elementName is the human-readable name (e.g., "CPU", "Memory") resolved by the catalog.
func (m Model) renderSelectedHelp(elementName, helpText string, contentWidth int) string {
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	return labelStyle.Render(elementName+": ") + helpStyle.Render(helpText)
}
