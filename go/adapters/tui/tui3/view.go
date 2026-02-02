package tui3

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout constants
const (
	resourcesHeight = 2 // Two rows for resources
	commandHeight   = 1 // One row for command
	summaryHeight   = 1 // One row for summary when shown
	borderChrome    = 2 // Top and bottom borders
)

// View renders the TUI.
func (m Model) View() string {
	if m.quitting {
		return m.renderFinalView()
	}

	// Calculate layout dimensions
	layout := m.calculateLayout()

	// Build the view from sections
	var sections []string

	// Section 1: Resources (2 rows)
	sections = append(sections, m.renderResourcesPane(layout))

	// Section 2: Command (1 row)
	sections = append(sections, m.renderCommandPane(layout))

	// Section 3-6: Main area (selector + right stack)
	sections = append(sections, m.renderMainArea(layout))

	// Section 7: Summary (if available)
	if m.summaryData != nil {
		sections = append(sections, m.renderSummaryPane(layout))
	}

	return strings.Join(sections, "\n")
}

// Layout holds calculated dimensions for rendering.
type Layout struct {
	Width           int
	Height          int
	ResourcesHeight int
	CommandHeight   int
	MainHeight      int
	SummaryHeight   int
	SelectorWidth   int
	RightStackWidth int
}

// calculateLayout computes the layout dimensions.
func (m Model) calculateLayout() Layout {
	l := Layout{
		Width:           m.width,
		Height:          m.height,
		ResourcesHeight: resourcesHeight + borderChrome,
		CommandHeight:   commandHeight + borderChrome,
	}

	// Summary height if data is present
	if m.summaryData != nil {
		l.SummaryHeight = summaryHeight + borderChrome
	}

	// Main area gets remaining height
	l.MainHeight = m.height - l.ResourcesHeight - l.CommandHeight - l.SummaryHeight
	if l.MainHeight < 5 {
		l.MainHeight = 5
	}

	// Selector is ~35% of width, right stack is ~65%
	l.SelectorWidth = m.width * 35 / 100
	if l.SelectorWidth < 20 {
		l.SelectorWidth = 20
	}
	l.RightStackWidth = m.width - l.SelectorWidth - 1 // -1 for separator

	return l
}

// renderResourcesPane renders the top resources section (cells A-G).
func (m Model) renderResourcesPane(l Layout) string {
	// Row 1: Timer, CPU, Memory, Containers
	row1Parts := []string{
		m.cells.Timer.Render(8, 1),
		m.cells.CPU.Render(30, 1),
		m.cells.Mem.Render(20, 1),
		m.cells.Containers.Render(20, 1),
	}
	row1 := strings.Join(row1Parts, " │ ")

	// Row 2: UoW Stats, Tools, Layer
	row2Parts := []string{
		m.cells.UoWStats.Render(40, 1),
		m.cells.Tools.Render(15, 1),
		m.cells.Layer.Render(12, 1),
	}
	row2 := strings.Join(row2Parts, " │ ")

	// Wrap with simple border
	return m.wrapPane("Resources", row1+"\n"+row2, l.Width)
}

// renderCommandPane renders the command section (cell H).
func (m Model) renderCommandPane(l Layout) string {
	content := m.cells.Command.Render(l.Width-4, 1)
	return m.wrapPane("Command", content, l.Width)
}

// renderMainArea renders the main area with selector and right stack.
func (m Model) renderMainArea(l Layout) string {
	// Left: Selector
	selectorContent := m.cells.Selector.Render(l.SelectorWidth-2, l.MainHeight-2)
	selector := m.wrapPane("Units", selectorContent, l.SelectorWidth)

	// Right: Stack of Selected, Resources, Output
	rightStack := m.renderRightStack(l)

	// Join side by side
	return lipgloss.JoinHorizontal(lipgloss.Top, selector, rightStack)
}

// renderRightStack renders the right side stack (cells J, K-M, N).
func (m Model) renderRightStack(l Layout) string {
	var sections []string

	// Selected (1 line)
	selectedContent := m.cells.Selected.Render(l.RightStackWidth-4, 1)
	sections = append(sections, wrapLine("Selected", selectedContent, l.RightStackWidth))

	// UoW Resources: Deps, Cache, Artifacts (1 line combined)
	depsContent := m.cells.Deps.Render(l.RightStackWidth/3, 1)
	cacheContent := m.cells.Cache.Render(l.RightStackWidth/3, 1)
	artifactsContent := m.cells.Artifacts.Render(l.RightStackWidth/3, 1)
	resourcesLine := depsContent + " " + cacheContent + " " + artifactsContent
	sections = append(sections, wrapLine("Resources", resourcesLine, l.RightStackWidth))

	// Output (remaining height)
	outputHeight := l.MainHeight - 4 - 2 // -4 for selected and resources, -2 for borders
	if outputHeight < 3 {
		outputHeight = 3
	}
	outputContent := m.cells.Output.Render(l.RightStackWidth-4, outputHeight)
	sections = append(sections, m.wrapPane("Output", outputContent, l.RightStackWidth))

	return strings.Join(sections, "\n")
}

// renderSummaryPane renders the summary section (cell O).
func (m Model) renderSummaryPane(l Layout) string {
	content := m.cells.Summary.Render(l.Width-4, 1)
	return m.wrapPane("Summary", content, l.Width)
}

// wrapPane wraps content in a titled border.
func (m Model) wrapPane(title, content string, width int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(width - 2)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	header := titleStyle.Render("─ " + title + " ")
	borderLen := width - lipgloss.Width(header) - 4
	if borderLen < 0 {
		borderLen = 0
	}
	header += strings.Repeat("─", borderLen)

	return header + "\n" + style.Render(content)
}

// wrapLine wraps a single line with a label.
func wrapLine(label, content string, width int) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	return labelStyle.Render(label+": ") + content
}

// renderFinalView renders the plain-text final output when quitting.
func (m Model) renderFinalView() string {
	if m.summaryData == nil {
		return "Execution complete."
	}

	return m.cells.Summary.Render(m.width, 3)
}
