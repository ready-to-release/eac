package console

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderStatusBar renders a 5-row status bar at the bottom of the TUI.
// Row 1: Header with freeze button
// Row 2: Data (status text, counters, slots, CPU)
// Row 3: Help text for hovered element (or placeholder)
// Row 4: Resource lamps (Mem/Host/Docker, 2-column layout)
// Row 5: Footer
func (m Model) renderStatusBar() string {
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

	// === Row 2: Data (unchanged layout) ===
	rightCells := sep + counters + "  " + progressCount + sep + slotsCell + sep + cpuCell
	rightWidth := lipgloss.Width(rightCells)
	statusMaxWidth := contentWidth - rightWidth
	if statusMaxWidth < 10 {
		statusMaxWidth = 10
	}
	if lipgloss.Width(statusText) > statusMaxWidth {
		statusText = Styles.Dim.Render("…")
	}
	dataLine := statusText + rightCells
	if visWidth := lipgloss.Width(dataLine); visWidth < contentWidth {
		dataLine += strings.Repeat(" ", contentWidth-visWidth)
	}
	result.WriteString(Styles.Border.Render("│") + " " + dataLine + " " + Styles.Border.Render("│") + "\n")

	// === Row 3: Help text ===
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

	// === Row 4: Resource lamps (2-column: Host │ Docker) ===
	memLamps := m.Resources.Catalog.RenderWidget("res-mem", snap)
	hostLamps := m.Resources.Catalog.RenderWidget("res-host", snap)
	dmemLamps := m.Resources.Catalog.RenderWidget("res-dmem", snap)
	dockerLamps := m.Resources.Catalog.RenderWidget("res-docker", snap)

	col1 := white.Render(fmt.Sprintf("%-5s", "Mem:")) + memLamps + " " + white.Render(fmt.Sprintf("%-5s", "Host:")) + hostLamps
	col2 := white.Render(fmt.Sprintf("%-5s", "Mem:")) + dmemLamps + " " + white.Render(fmt.Sprintf("%-7s", "Docker:")) + dockerLamps
	lampsLine := col1 + sep + col2

	if lampsVisWidth := lipgloss.Width(lampsLine); lampsVisWidth < contentWidth {
		lampsLine += strings.Repeat(" ", contentWidth-lampsVisWidth)
	}
	result.WriteString(Styles.Border.Render("│") + " " + lampsLine + " " + Styles.Border.Render("│") + "\n")

	// === Row 5: Footer ===
	footerLen := m.Display.Width - 2
	if footerLen < 1 {
		footerLen = 1
	}
	result.WriteString("└" + Styles.Border.Render(strings.Repeat("─", footerLen)) + "┘")

	return result.String()
}
