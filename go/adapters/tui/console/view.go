package console

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/ready-to-release/eac/go/adapters/tui/console/render"
)

const stableLampsCount = render.StableLampsCount

func getPressureColor(running, capacity int) lipgloss.Style {
	return render.GetPressureColor(running, capacity)
}

// calculateLayoutMetrics returns cached layout dimensions, recomputing only when invalidated.
func (m Model) calculateLayoutMetrics() render.LayoutMetrics {
	if m.cachedLayoutMetrics != nil {
		return *m.cachedLayoutMetrics
	}
	return m.computeLayoutMetrics()
}

// computeLayoutMetrics computes layout dimensions based on current model state.
func (m Model) computeLayoutMetrics() render.LayoutMetrics {
	metrics := render.LayoutMetrics{}
	metrics.SummaryLines = 5                          // header + data + help + lamps + footer
	metrics.ComponentsStart = 1                       // panel header only (no resources pane)
	metrics.RemainingHeight = m.height - metrics.SummaryLines
	if metrics.RemainingHeight < 5 {
		metrics.RemainingHeight = 5
	}
	return metrics
}

// invalidateLayoutMetrics forces recomputation on next calculateLayoutMetrics() call.
func (m *Model) invalidateLayoutMetrics() {
	metrics := m.computeLayoutMetrics()
	m.cachedLayoutMetrics = &metrics
}

// View renders the console window.
func (m Model) View() string {
	// When quitting in alt screen mode, return empty (screen will be restored)
	// Plain-text summary is printed after program.Run() completes
	if m.quitting {
		return ""
	}
	// Wrap with zone.Scan to enable mouse click detection on marked zones
	return zone.Scan(m.viewPanes())
}

// ViewFinal renders a clean summary after TUI exit.
// Shows only essential summary data - no pane chrome.
func (m Model) ViewFinal() string {
	var b strings.Builder

	// Render summary data if available
	if m.summaryData != nil {
		// Details (includes module table from summary.go)
		for _, line := range m.summaryData.Details {
			cleanLine := stripMarkdownPipes(line)
			b.WriteString(fmt.Sprintf("%s\n", cleanLine))
		}

		// Summary status
		icon := "✓"
		statusText := "PASSED"
		if !m.summaryData.Success {
			icon = "✗"
			statusText = "FAILED"
		}
		b.WriteString(fmt.Sprintf("\n%s %s (%s)\n", icon, statusText, formatElapsed(m.summaryData.TotalTime)))

		// Run phase summary
		if m.summaryData.RunSummary != "" {
			b.WriteString(fmt.Sprintf("  %s\n", m.summaryData.RunSummary))
		}
	}

	return b.String()
}

func stripMarkdownPipes(line string) string { return render.StripMarkdownPipes(line) }

// viewPanes renders the 2-section layout: side-by-side (top), status bar (bottom).
// Uses the same render path for both init and active phases — init data is zero/OFF,
// active-phase renderers handle empty state gracefully.
// IMPORTANT: Layout calculations use calculateLayoutMetrics() as single source of truth.
func (m Model) viewPanes() string {
	var b strings.Builder
	metrics := m.calculateLayoutMetrics()

	// === Top (dynamic): Side-by-side fills RemainingHeight ===
	tabs := m.GetVisibleTabs()
	sideBySide := m.renderSideBySideLayout(tabs, metrics.RemainingHeight)
	b.WriteString(sideBySide)
	b.WriteString("\n")

	// === Bottom (fixed): Status bar (5 rows) ===
	b.WriteString(m.renderStatusBar())

	return b.String()
}

// renderSideBySideLayout renders Components (left) and Logs (right) side by side.
// Logs header shows Selected content (UoW details) when a tab is hovered, otherwise phase header.
func (m Model) renderSideBySideLayout(tabs []*UoWState, height int) string {
	componentsWidth := m.ComponentsWidth()
	rightWidth := m.width - componentsWidth
	if rightWidth < 40 {
		rightWidth = 40
	}

	// Left: tab grid (full height)
	leftPanel := m.renderTabGridPanel(tabs, componentsWidth, height)

	// Right: logs (full height, header shows Selected content)
	rightPanel := m.renderLogsPanel(rightWidth, height)

	// Join horizontally
	leftLines := strings.Split(leftPanel, "\n")
	rightLines := strings.Split(rightPanel, "\n")

	for len(leftLines) < height {
		leftLines = append(leftLines, strings.Repeat(" ", componentsWidth))
	}
	for len(rightLines) < height {
		rightLines = append(rightLines, strings.Repeat(" ", rightWidth))
	}

	var b strings.Builder
	for i := 0; i < height; i++ {
		left := leftLines[i]
		right := rightLines[i]

		leftWidth := lipgloss.Width(left)
		if leftWidth < componentsWidth {
			left += strings.Repeat(" ", componentsWidth-leftWidth)
		}

		b.WriteString(left)
		b.WriteString(right)

		if i < height-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// --- Thin delegations to render package ---

func (m Model) statusIcon(status UoWStatus) string {
	return render.UoWStatusIcon(int(status), m.asciiMode)
}

func weightDigit(weight int) string       { return render.WeightDigit(weight) }
func getModuleName(moniker string) string { return render.GetModuleName(moniker) }

func (m Model) lineIcons() (fail, warn, info string) {
	return render.LineIcons(m.asciiMode)
}

func (m Model) phaseIcon(status PhaseStatus) string {
	return render.PhaseIcon(int(status), m.asciiMode)
}


