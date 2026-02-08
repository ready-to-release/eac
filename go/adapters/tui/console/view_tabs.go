package console

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderTabGridPanel renders the tab grid as a panel for side-by-side layout.
func (m Model) renderTabGridPanel(tabs []*UoWState, width, height int) string {
	var b strings.Builder

	// Header: ┌ ●●●●●○○○○○○○ ──────────────────────┐
	snap := m.buildWidgetSnapshot()
	progressLamps := m.Resources.Catalog.RenderWidget("res-progress", snap)
	headerLeft := "┌ " + progressLamps + " "
	headerRight := "─┐"
	headerBorderLen := width - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight)
	if headerBorderLen < 1 {
		headerBorderLen = 1
	}
	b.WriteString(headerLeft + Styles.Border.Render(strings.Repeat("─", headerBorderLen)) + headerRight + "\n")

	// Tab content area (height - 2 for header and footer)
	contentHeight := height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Render all tab content (for scrolling)
	tabContent := m.renderTabGridContent(tabs, width-2, contentHeight+m.Interaction.TabsScrollOffset+50)
	tabLines := strings.Split(tabContent, "\n")

	// Apply scroll offset
	scrollOffset := m.Interaction.TabsScrollOffset
	if scrollOffset > len(tabLines)-contentHeight {
		scrollOffset = len(tabLines) - contentHeight
		if scrollOffset < 0 {
			scrollOffset = 0
		}
	}

	for i := 0; i < contentHeight; i++ {
		lineIdx := i + scrollOffset
		if lineIdx < len(tabLines) {
			line := tabLines[lineIdx]
			lineWidth := lipgloss.Width(line)
			padding := width - 2 - lineWidth
			if padding < 0 {
				padding = 0
			}
			b.WriteString(Styles.Border.Render("│") + line + strings.Repeat(" ", padding) + Styles.Border.Render("│") + "\n")
		} else {
			b.WriteString(Styles.Border.Render("│") + strings.Repeat(" ", width-2) + Styles.Border.Render("│") + "\n")
		}
	}

	// Footer
	footerBorderLen := width - 2
	if footerBorderLen < 1 {
		footerBorderLen = 1
	}
	b.WriteString("└" + Styles.Border.Render(strings.Repeat("─", footerBorderLen)) + "┘")

	return b.String()
}

// renderTabGridContent renders the tab grid as compact single-line tabs.
// Uses the catalog tab widget with dynamic sizing based on panel width.
func (m Model) renderTabGridContent(tabs []*UoWState, width, height int) string {
	if len(tabs) == 0 {
		if m.Execution.InitSummary == nil {
			return Styles.Dim.Render("  waiting for components...")
		}
		return Styles.Dim.Render("No components")
	}

	// Dynamic tab sizing based on tab width and available panel width
	sizing := ComputeTabSizing(width+2, m.Interaction.TabWidth, m.Interaction.HoveredTabScroll, m.Display.AsciiMode)
	sizing.ViewMode = m.Interaction.TabViewMode
	tabsPerRow := sizing.TabColumns

	effectiveActiveTab := m.getEffectiveActiveTab()

	// Render a single tab via catalog
	renderTab := func(state *UoWState) string {
		instance := TabInstance{
			Moniker:       state.Moniker,
			DisplayName:   state.DisplayName,
			Status:        state.Status,
			Weight:        state.Weight,
			IsActive:      state.Moniker == effectiveActiveTab,
			IsHovered:     state.Moniker == m.Interaction.HoveredTab && state.Moniker != effectiveActiveTab,
			Module:        state.Module,
			Component:     state.Component,
			Tool:          state.Tool,
			ComponentType: state.ComponentType,
			Container:     state.Container,
			StartTime:     state.StartTime,
			EndTime:       state.EndTime,
			ExitCode:      state.ExitCode,
		}
		return m.Resources.Catalog.RenderTab(instance, sizing)
	}

	var rows []string
	for rowStart := 0; rowStart < len(tabs); rowStart += tabsPerRow {
		var tabParts []string
		for colIdx := 0; colIdx < tabsPerRow; colIdx++ {
			tabIdx := rowStart + colIdx
			if tabIdx < len(tabs) {
				tabParts = append(tabParts, renderTab(tabs[tabIdx]))
			}
		}
		if len(tabParts) > 0 {
			rows = append(rows, strings.Join(tabParts, " "))
		}
	}

	if len(rows) > height {
		rows = rows[:height]
	}

	return strings.Join(rows, "\n")
}
