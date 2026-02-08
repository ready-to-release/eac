package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/ready-to-release/eac/go/adapters/tui/console/render"
)

// detailPaneHeight is the fixed height (in lines) of the detail pane when visible.
// 6 lines: header, labels1, values1, labels2, values2, footer.
const detailPaneHeight = 6

// Status styles for the detail pane, defined once to avoid repeated allocation.
var detailStyle = struct {
	Pending  lipgloss.Style
	Running  lipgloss.Style
	Complete lipgloss.Style
	Skipped  lipgloss.Style
	Failed   lipgloss.Style
}{
	Pending:  lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
	Running:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
	Complete: lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
	Skipped:  lipgloss.NewStyle().Foreground(lipgloss.Color("75")),
	Failed:   lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
}

// detailStyleForStatus returns the style for a given UoW status.
func detailStyleForStatus(status UoWStatus) lipgloss.Style {
	switch status {
	case UoWPending:
		return detailStyle.Pending
	case UoWRunning:
		return detailStyle.Running
	case UoWComplete:
		return detailStyle.Complete
	case UoWSkipped:
		return detailStyle.Skipped
	case UoWFailed:
		return detailStyle.Failed
	default:
		return Styles.Phase
	}
}

// renderDetailPane renders the detail pane showing selected UoW metadata as fixed-width widget cells.
// Returns exactly detailPaneHeight lines. width is the total panel width including borders.
func (m Model) renderDetailPane(width int) string {
	var b strings.Builder

	// Resolve the active UoW state
	activeMoniker := m.getEffectiveActiveTab()
	if m.Interaction.HoveredTab != "" {
		activeMoniker = m.Interaction.HoveredTab
	}

	var state *UoWState
	if activeMoniker != "" {
		if s, exists := m.Execution.UoWStates[activeMoniker]; exists {
			state = s
		}
	}

	// Border characters
	tl, tr, bl, br, hz, vt := "┌", "┐", "└", "┘", "─", "│"
	if m.Display.AsciiMode {
		tl, tr, bl, br, hz, vt = "+", "+", "+", "+", "-", "|"
	}

	innerWidth := width - 4 // 2 border chars + 2 padding spaces
	if innerWidth < 20 {
		innerWidth = 20
	}

	dimBorder := Styles.Dim.Render

	// Helper: render a padded content line with borders
	contentLine := func(content string) string {
		visWidth := lipgloss.Width(content)
		pad := innerWidth - visWidth
		if pad < 0 {
			pad = 0
		}
		return dimBorder(vt) + " " + content + strings.Repeat(" ", pad) + " " + dimBorder(vt)
	}

	// --- Line 0: Header with UoW moniker ---
	headerText := " Detail "
	if activeMoniker != "" {
		headerText = " " + activeMoniker + " "
	}
	maxHeaderLen := width - 4
	if len(headerText) > maxHeaderLen {
		headerText = headerText[:maxHeaderLen-1] + "…"
	}
	headerStyle := Styles.Phase
	if state != nil {
		headerStyle = detailStyleForStatus(state.Status)
	}
	borderLen := width - lipgloss.Width(headerText) - 2
	if borderLen < 1 {
		borderLen = 1
	}
	b.WriteString(dimBorder(tl) + headerStyle.Render(headerText) + dimBorder(strings.Repeat(hz, borderLen)) + dimBorder(tr) + "\n")

	// Style helpers
	labelStyle := Styles.Dim
	amberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	// Default to empty-state placeholders; override when state is available
	moduleName, componentName, uowID, toolName, typeName := "---", "---", "---", "---", "---"
	stateName, durationStr, phaseStr, progressStr := "---", "---", "---", "---"
	stateStyle, durationStyle := Styles.Dim, Styles.Dim

	if state != nil {
		moduleName = state.Module
		componentName = state.Component
		toolName = state.Tool

		parts := strings.Split(state.Moniker, ":")
		if len(parts) > 0 {
			uowID = parts[len(parts)-1]
		}

		// Type: container or host
		if m.Display.AsciiMode {
			if state.Container {
				typeName = "docker"
			} else {
				typeName = "host"
			}
		} else {
			if state.Container {
				typeName = "\U0001F433" // whale emoji
			} else {
				typeName = "\U0001F5A5" // desktop emoji
			}
		}

		stateName, stateStyle = m.detailStatusText(state)
		durationStr, durationStyle = m.detailDuration(state)

		phaseStr = m.Display.RunPhaseName
		if phaseStr == "" {
			phaseStr = "Run"
		}
		progressStr = fmt.Sprintf("%d/%d", m.Execution.Completed, m.Execution.Total)
	}

	// --- UoW progress lamps (10-lamp gradient) ---
	const detailLampCount = 10
	var progressLamps string
	if state != nil && m.Execution.Total > 0 {
		activeLamps := m.Execution.Completed * detailLampCount / m.Execution.Total
		if m.Execution.Completed >= m.Execution.Total {
			activeLamps = detailLampCount
		}
		progressLamps = render.RenderProgressGradientLamps(activeLamps, detailLampCount, m.Display.AsciiMode)
	} else {
		progressLamps = render.RenderProgressGradientLamps(0, detailLampCount, m.Display.AsciiMode)
	}

	// --- Cell rendering helpers ---
	type cell struct {
		label string
		value string
		width int
		style lipgloss.Style
	}

	renderCellLabels := func(cells []cell) string {
		parts := make([]string, 0, len(cells))
		for i := range cells {
			parts = append(parts, labelStyle.Render(render.PadOrTruncate(cells[i].label, cells[i].width)))
		}
		return strings.Join(parts, "")
	}

	renderCellValues := func(cells []cell) string {
		parts := make([]string, 0, len(cells))
		for i := range cells {
			val := cells[i].value
			visLen := lipgloss.Width(val)
			if visLen > cells[i].width {
				// For pre-rendered ANSI content (like lamps), truncate is not safe;
				// for plain text, truncate and add ellipsis
				if len(val) == visLen {
					val = val[:cells[i].width-1] + "…"
				}
			}
			styled := cells[i].style.Render(val)
			visWidth := lipgloss.Width(styled)
			pad := cells[i].width - visWidth
			if pad < 0 {
				pad = 0
			}
			parts = append(parts, styled+strings.Repeat(" ", pad))
		}
		return strings.Join(parts, "")
	}

	// Row 1: Module, Component, Type, Tool, UoW ID
	row1 := []cell{
		{"Module", moduleName, 14, amberStyle},
		{"Component", componentName, 14, amberStyle},
		{"Type", typeName, 14, whiteStyle},
		{"Tool", toolName, 12, amberStyle},
		{"UoW ID", uowID, 12, amberStyle},
	}

	// Row 2: Progress, lamps(no label), State, Duration, Phase
	row2 := []cell{
		{"Progress", progressStr, 14, whiteStyle},
		{"", progressLamps, 14, whiteStyle},
		{"State", stateName, 14, stateStyle},
		{"Duration", durationStr, 12, durationStyle},
		{"Phase", phaseStr, 12, Styles.Phase},
	}

	// Lines 1-2: Row 1 labels + values
	b.WriteString(contentLine(renderCellLabels(row1)) + "\n")
	b.WriteString(contentLine(renderCellValues(row1)) + "\n")
	// Lines 3-4: Row 2 labels + values
	b.WriteString(contentLine(renderCellLabels(row2)) + "\n")
	b.WriteString(contentLine(renderCellValues(row2)) + "\n")

	// --- Line 5: Footer ---
	footerBorderLen := width - 2
	if footerBorderLen < 1 {
		footerBorderLen = 1
	}
	b.WriteString(dimBorder(bl) + dimBorder(strings.Repeat(hz, footerBorderLen)) + dimBorder(br))

	return b.String()
}

// detailStatusText returns display text and style for a UoW status.
func (m Model) detailStatusText(state *UoWState) (string, lipgloss.Style) {
	icon := m.statusIcon(state.Status)
	style := detailStyleForStatus(state.Status)
	switch state.Status {
	case UoWPending:
		return icon + " WAIT", style
	case UoWRunning:
		return icon + " RUN", style
	case UoWComplete:
		return icon + " DONE", style
	case UoWSkipped:
		return icon + " CACHED", style
	case UoWFailed:
		return icon + " FAIL", style
	default:
		return icon + " ???", Styles.Dim
	}
}

// detailDuration returns the formatted duration and style for a UoW.
func (m Model) detailDuration(state *UoWState) (string, lipgloss.Style) {
	style := detailStyleForStatus(state.Status)
	switch state.Status {
	case UoWRunning:
		if !state.StartTime.IsZero() {
			dur := time.Since(state.StartTime).Round(100 * time.Millisecond)
			return formatElapsed(dur), style
		}
		return "0s", style
	case UoWComplete:
		if !state.StartTime.IsZero() && !state.EndTime.IsZero() {
			dur := state.EndTime.Sub(state.StartTime).Round(100 * time.Millisecond)
			return formatElapsed(dur), style
		}
		return "done", style
	case UoWSkipped:
		return "cached", style
	case UoWFailed:
		if !state.StartTime.IsZero() && !state.EndTime.IsZero() {
			dur := state.EndTime.Sub(state.StartTime).Round(100 * time.Millisecond)
			return formatElapsed(dur), style
		}
		return "failed", style
	default:
		return "---", Styles.Dim
	}
}
