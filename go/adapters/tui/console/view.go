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
	if m.Resources.CachedLayoutMetrics != nil {
		return *m.Resources.CachedLayoutMetrics
	}
	return m.computeLayoutMetrics()
}

// computeLayoutMetrics computes layout dimensions based on current model state.
func (m Model) computeLayoutMetrics() render.LayoutMetrics {
	metrics := render.LayoutMetrics{}
	metrics.SummaryLines = 7    // top bar (4) + bottom bar (3)
	metrics.ComponentsStart = 5 // top bar (4 rows, last has no trailing \n) + panel header (1)
	metrics.RemainingHeight = m.Display.Height - metrics.SummaryLines
	if metrics.RemainingHeight < 5 {
		metrics.RemainingHeight = 5
	}
	// Show detail pane when run phase is active and terminal has enough space.
	// Requires: detailPaneHeight (6) + minimum 5 log lines = 11 lines.
	if m.Execution.Panes[PhaseRun] != nil && m.Execution.Panes[PhaseRun].Status != PhasePending &&
		metrics.RemainingHeight >= detailPaneHeight+5 {
		metrics.DetailPaneHeight = detailPaneHeight
	}
	return metrics
}

// invalidateLayoutMetrics forces recomputation on next calculateLayoutMetrics() call.
// Also recomputes and caches pane heights.
func (m *Model) invalidateLayoutMetrics() {
	metrics := m.computeLayoutMetrics()
	m.Resources.CachedLayoutMetrics = &metrics
	initH, runH, summaryH := m.computePaneHeights()
	m.Resources.CachedPaneHeights = &PaneHeights{InitH: initH, RunH: runH, SummaryH: summaryH}
}

// View renders the console window.
func (m Model) View() string {
	// When quitting in alt screen mode, return empty (screen will be restored)
	// Plain-text summary is printed after program.Run() completes
	if m.Interaction.Quitting {
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
	if m.Execution.SummaryData != nil {
		// Details (includes module table from summary.go)
		for _, line := range m.Execution.SummaryData.Details {
			cleanLine := stripMarkdownPipes(line)
			b.WriteString(fmt.Sprintf("%s\n", cleanLine))
		}

		// Summary status
		icon := "✓"
		statusText := "PASSED"
		if !m.Execution.SummaryData.Success {
			icon = "✗"
			statusText = "FAILED"
		}
		b.WriteString(fmt.Sprintf("\n%s %s (%s)\n", icon, statusText, formatElapsed(m.Execution.SummaryData.TotalTime)))

		// Run phase summary
		if m.Execution.SummaryData.RunSummary != "" {
			b.WriteString(fmt.Sprintf("  %s\n", m.Execution.SummaryData.RunSummary))
		}
	}

	return b.String()
}

func stripMarkdownPipes(line string) string { return render.StripMarkdownPipes(line) }

// viewPanes renders the 3-section layout: top bar, side-by-side (middle), bottom bar.
// Uses the same render path for both init and active phases — init data is zero/OFF,
// active-phase renderers handle empty state gracefully.
// IMPORTANT: Layout calculations use calculateLayoutMetrics() as single source of truth.
func (m Model) viewPanes() string {
	var b strings.Builder
	metrics := m.calculateLayoutMetrics()

	// === Top (fixed): Dashboard bar (4 rows) ===
	b.WriteString(m.renderTopBar())
	b.WriteString("\n")

	// === Middle (dynamic): Side-by-side fills RemainingHeight ===
	tabs := m.GetVisibleTabs()
	sideBySide := m.renderSideBySideLayout(tabs, metrics)
	b.WriteString(sideBySide)
	b.WriteString("\n")

	// === Bottom (fixed): Help bar (3 rows) ===
	b.WriteString(m.renderBottomBar())

	return b.String()
}

// renderSideBySideLayout renders Components (left) and Logs (right) side by side.
// When DetailPaneHeight > 0, the right panel is split: detail pane on top, headless logs below.
// Otherwise falls back to single logs panel with header.
func (m Model) renderSideBySideLayout(tabs []*UoWState, metrics render.LayoutMetrics) string {
	height := metrics.RemainingHeight
	componentsWidth := m.ComponentsWidth()
	rightWidth := m.Display.Width - componentsWidth
	if rightWidth < 40 {
		rightWidth = 40
	}

	// Left: tab grid (full height)
	leftPanel := m.renderTabGridPanel(tabs, componentsWidth, height)

	// Right: detail pane + headless logs, or single logs panel
	var rightPanel string
	if metrics.DetailPaneHeight > 0 {
		detailPanel := m.renderDetailPane(rightWidth)
		logsHeight := height - metrics.DetailPaneHeight
		if logsHeight < 2 {
			logsHeight = 2
		}
		logsPanel := m.renderLogsPaneHeadless(rightWidth, logsHeight)
		rightPanel = detailPanel + "\n" + logsPanel
	} else {
		rightPanel = m.renderLogsPanel(rightWidth, height)
	}

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
	return render.UoWStatusIcon(int(status), m.Display.AsciiMode)
}

func weightDigit(weight int) string       { return render.WeightDigit(weight) }
func getModuleName(moniker string) string { return render.GetModuleName(moniker) }

func (m Model) lineIcons() (fail, warn, info string) {
	return render.LineIcons(m.Display.AsciiMode)
}

func (m Model) phaseIcon(status PhaseStatus) string {
	return render.PhaseIcon(int(status), m.Display.AsciiMode)
}
