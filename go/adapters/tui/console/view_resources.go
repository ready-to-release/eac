package console

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderTopBar renders a 4-row dashboard bar at the top of the TUI.
// Row 1: Header with freeze button
// Row 2: Data (status text, counters, slots) left │ Host Mem + Host lamps right
// Row 3: CPU left │ Docker Mem + Pressure lamps right
// Row 4: Footer
func (m Model) renderTopBar() string {
	var result strings.Builder

	snap := m.buildWidgetSnapshot()
	veryDark := lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
	white := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	sep := veryDark.Render(" │ ")

	// Widget-rendered cells
	counters := m.Resources.Catalog.RenderWidget("res-counters", snap)
	progressCount := m.Resources.Catalog.RenderWidget("progress-count", snap)
	statusText := m.Resources.Catalog.RenderWidget("status-text", snap)
	cpuCell := m.Resources.Catalog.RenderWidget("res-cpu", snap)
	slotsCell := m.Resources.Catalog.RenderWidget("res-slots", snap)

	contentWidth := m.Display.Width - 4 // inside "│ " + " │"
	if contentWidth < 1 {
		contentWidth = 1
	}

	// === Row 1: Header with freeze button ===
	freezeBtn := m.Resources.Catalog.RenderWidget("freeze-button", snap)
	headerLeft := "┌─ " + Styles.Dim.Render("Status") + " "
	headerBorderLen := m.Display.Width - lipgloss.Width(headerLeft) - lipgloss.Width(freezeBtn) - 3
	if headerBorderLen < 1 {
		headerBorderLen = 1
	}
	result.WriteString(headerLeft + Styles.Border.Render(strings.Repeat("─", headerBorderLen)) + " " + freezeBtn + " ┐\n")

	// === Far-right column: resource lamp pairs (host top, docker bottom) ===
	memLamps := m.Resources.Catalog.RenderWidget("res-mem", snap)
	hostLamps := m.Resources.Catalog.RenderWidget("res-host", snap)
	dmemLamps := m.Resources.Catalog.RenderWidget("res-dmem", snap)
	dockerLamps := m.Resources.Catalog.RenderWidget("res-docker", snap)

	// Use %-10s so "Host:" and "Pressure:" align to the same column width
	hostRight := white.Render(fmt.Sprintf("%-5s", "Mem:")) + memLamps + "  " + white.Render(fmt.Sprintf("%-10s", "Host:")) + hostLamps
	dockerRight := white.Render(fmt.Sprintf("%-5s", "Mem:")) + dmemLamps + "  " + white.Render(fmt.Sprintf("%-10s", "Pressure:")) + dockerLamps

	// Ensure both right-column rows are the same width
	rightWidth := lipgloss.Width(hostRight)
	if w := lipgloss.Width(dockerRight); w > rightWidth {
		rightWidth = w
	}
	if lipgloss.Width(hostRight) < rightWidth {
		hostRight += strings.Repeat(" ", rightWidth-lipgloss.Width(hostRight))
	}
	if lipgloss.Width(dockerRight) < rightWidth {
		dockerRight += strings.Repeat(" ", rightWidth-lipgloss.Width(dockerRight))
	}

	sepWidth := lipgloss.Width(sep)
	leftZoneWidth := contentWidth - rightWidth - sepWidth
	if leftZoneWidth < 10 {
		leftZoneWidth = 10
	}

	// === Row 2: Data (left) │ Host lamps (right) ===
	leftCells := statusText + sep + counters + "  " + progressCount + sep + slotsCell
	leftWidth := lipgloss.Width(leftCells)
	if leftWidth > leftZoneWidth {
		// Truncate status text to fit
		fixedRight := sep + counters + "  " + progressCount + sep + slotsCell
		statusMaxWidth := leftZoneWidth - lipgloss.Width(fixedRight)
		if statusMaxWidth < 3 {
			statusMaxWidth = 3
		}
		if lipgloss.Width(statusText) > statusMaxWidth {
			statusText = Styles.Dim.Render("…")
		}
		leftCells = statusText + fixedRight
		leftWidth = lipgloss.Width(leftCells)
	}
	if leftWidth < leftZoneWidth {
		leftCells += strings.Repeat(" ", leftZoneWidth-leftWidth)
	}
	result.WriteString(Styles.Border.Render("│") + " " + leftCells + sep + hostRight + " " + Styles.Border.Render("│") + "\n")

	// === Row 3: CPU (left) │ Docker lamps (right) ===
	cpuLeft := cpuCell
	cpuWidth := lipgloss.Width(cpuLeft)
	if cpuWidth < leftZoneWidth {
		cpuLeft += strings.Repeat(" ", leftZoneWidth-cpuWidth)
	}
	result.WriteString(Styles.Border.Render("│") + " " + cpuLeft + sep + dockerRight + " " + Styles.Border.Render("│") + "\n")

	// === Row 4: Footer ===
	footerLen := m.Display.Width - 2
	if footerLen < 1 {
		footerLen = 1
	}
	result.WriteString("└" + Styles.Border.Render(strings.Repeat("─", footerLen)) + "┘")

	return result.String()
}

// renderBottomBar renders a 3-row help bar at the bottom of the TUI.
// Row 1: Header
// Row 2: Help text for hovered element (or placeholder)
// Row 3: Footer
func (m Model) renderBottomBar() string {
	var result strings.Builder

	contentWidth := m.Display.Width - 4 // inside "│ " + " │"
	if contentWidth < 1 {
		contentWidth = 1
	}

	// === Row 1: Header ===
	headerBorderLen := m.Display.Width - 2
	if headerBorderLen < 1 {
		headerBorderLen = 1
	}
	result.WriteString("┌" + Styles.Border.Render(strings.Repeat("─", headerBorderLen)) + "┐\n")

	// === Row 2: Help text ===
	var helpLine string
	if helpText, ok := m.Resources.Catalog.HelpText(m.Interaction.HoveredZone); ok {
		elementName := m.Resources.Catalog.ElementName(m.Interaction.HoveredZone)
		helpLine = m.renderSelectedHelp(elementName, helpText, contentWidth)
	} else {
		helpLine = Styles.Dim.Render("hover over a UI element to understand it")
	}
	if helpVisWidth := lipgloss.Width(helpLine); helpVisWidth < contentWidth {
		helpLine += strings.Repeat(" ", contentWidth-helpVisWidth)
	}
	result.WriteString(Styles.Border.Render("│") + " " + helpLine + " " + Styles.Border.Render("│") + "\n")

	// === Row 3: Footer ===
	footerLen := m.Display.Width - 2
	if footerLen < 1 {
		footerLen = 1
	}
	result.WriteString("└" + Styles.Border.Render(strings.Repeat("─", footerLen)) + "┘")

	return result.String()
}
