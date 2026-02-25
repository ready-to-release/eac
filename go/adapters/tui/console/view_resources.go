package console

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderTopBar renders a 5-row dashboard bar at the top of the TUI.
// Row 1: Header with freeze button
// Row 2: (left empty) │ Host / Docker column headers (right)
// Row 3: Data (status text, counters, slots) left │ Pressure lamps (right)
// Row 4: CPU left │ Mem lamps (right)
// Row 5: Footer
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

	// === Far-right columns: Host (left col) │ Docker (right col) ===
	hostPressureLamps := m.Resources.Catalog.RenderWidget("res-host", snap)
	hostMemLamps := m.Resources.Catalog.RenderWidget("res-mem", snap)
	dockerPressureLamps := m.Resources.Catalog.RenderWidget("res-docker", snap)
	dockerMemLamps := m.Resources.Catalog.RenderWidget("res-dmem", snap)

	// Build each column's 3 rows: header, pressure, mem
	colGap := "  "
	headerLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)

	hostHeader := headerLabel.Render("Host")
	dockerHeader := headerLabel.Render("Docker")
	hostPressure := white.Render(fmt.Sprintf("%-10s", "Pressure:")) + hostPressureLamps
	dockerPressure := white.Render(fmt.Sprintf("%-10s", "Pressure:")) + dockerPressureLamps
	hostMem := white.Render(fmt.Sprintf("%-10s", "Mem:")) + hostMemLamps
	dockerMem := white.Render(fmt.Sprintf("%-10s", "Mem:")) + dockerMemLamps

	// Compute per-column visual widths, then align
	hostColWidth := lipgloss.Width(hostPressure)
	if w := lipgloss.Width(hostMem); w > hostColWidth {
		hostColWidth = w
	}
	if w := lipgloss.Width(hostHeader); w > hostColWidth {
		hostColWidth = w
	}
	dockerColWidth := lipgloss.Width(dockerPressure)
	if w := lipgloss.Width(dockerMem); w > dockerColWidth {
		dockerColWidth = w
	}
	if w := lipgloss.Width(dockerHeader); w > dockerColWidth {
		dockerColWidth = w
	}

	padRight := func(s string, width int) string {
		if w := lipgloss.Width(s); w < width {
			return s + strings.Repeat(" ", width-w)
		}
		return s
	}

	// 3 right-side rows: header, pressure, mem
	rightRow1 := padRight(hostHeader, hostColWidth) + colGap + padRight(dockerHeader, dockerColWidth)
	rightRow2 := padRight(hostPressure, hostColWidth) + colGap + padRight(dockerPressure, dockerColWidth)
	rightRow3 := padRight(hostMem, hostColWidth) + colGap + padRight(dockerMem, dockerColWidth)

	rightWidth := lipgloss.Width(rightRow1)
	if w := lipgloss.Width(rightRow2); w > rightWidth {
		rightWidth = w
	}
	if w := lipgloss.Width(rightRow3); w > rightWidth {
		rightWidth = w
	}
	rightRow1 = padRight(rightRow1, rightWidth)
	rightRow2 = padRight(rightRow2, rightWidth)
	rightRow3 = padRight(rightRow3, rightWidth)

	sepWidth := lipgloss.Width(sep)
	leftZoneWidth := contentWidth - rightWidth - sepWidth
	if leftZoneWidth < 10 {
		leftZoneWidth = 10
	}

	// === Row 2: Headers (left empty) │ Host / Docker headers (right) ===
	emptyLeft := strings.Repeat(" ", leftZoneWidth)
	result.WriteString(Styles.Border.Render("│") + " " + emptyLeft + sep + rightRow1 + " " + Styles.Border.Render("│") + "\n")

	// === Row 3: Data (left) │ Pressure lamps (right) ===
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
	result.WriteString(Styles.Border.Render("│") + " " + leftCells + sep + rightRow2 + " " + Styles.Border.Render("│") + "\n")

	// === Row 4: CPU (left) │ Mem lamps (right) ===
	cpuLeft := cpuCell
	cpuWidth := lipgloss.Width(cpuLeft)
	if cpuWidth < leftZoneWidth {
		cpuLeft += strings.Repeat(" ", leftZoneWidth-cpuWidth)
	}
	result.WriteString(Styles.Border.Render("│") + " " + cpuLeft + sep + rightRow3 + " " + Styles.Border.Render("│") + "\n")

	// === Row 4: Footer ===
	footerLen := m.Display.Width - 2
	if footerLen < 1 {
		footerLen = 1
	}
	result.WriteString("└" + Styles.Border.Render(strings.Repeat("─", footerLen)) + "┘")

	return result.String()
}

