package mainarea

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shahar3/bubble-grid/frame"
	"github.com/shahar3/bubble-grid/grid"

	"github.com/ready-to-release/eac/go/eac/adapters/tui/console/components/shared"
)

// UnitGrid displays units in a grid layout using bubble-grid.
type UnitGrid struct {
	units         []*shared.UnitState
	selectedIndex int
	hoveredIndex  int
	columns       int
	width         int
	height        int
	focused       bool
}

// NewUnitGrid creates a new unit grid with the specified number of columns.
func NewUnitGrid(columns int) *UnitGrid {
	if columns < 1 {
		columns = 3
	}
	return &UnitGrid{
		columns:       columns,
		selectedIndex: -1,
		hoveredIndex:  -1,
	}
}

// SetUnits updates the units displayed in the grid.
func (g *UnitGrid) SetUnits(units []*shared.UnitState) {
	g.units = units
	// Clamp selected index
	if g.selectedIndex >= len(units) {
		g.selectedIndex = len(units) - 1
	}
}

// AddUnit adds a unit to the grid.
func (g *UnitGrid) AddUnit(unit *shared.UnitState) {
	g.units = append(g.units, unit)
	// Auto-select first unit
	if g.selectedIndex < 0 && len(g.units) == 1 {
		g.selectedIndex = 0
	}
}

// UpdateUnit updates a unit by moniker.
func (g *UnitGrid) UpdateUnit(moniker string, status shared.UnitStatus) {
	for _, unit := range g.units {
		if unit.Moniker == moniker {
			unit.Status = status
			return
		}
	}
}

// Select selects a unit by index.
func (g *UnitGrid) Select(index int) {
	if index >= 0 && index < len(g.units) {
		g.selectedIndex = index
	}
}

// SelectByMoniker selects a unit by moniker.
func (g *UnitGrid) SelectByMoniker(moniker string) {
	for i, unit := range g.units {
		if unit.Moniker == moniker {
			g.selectedIndex = i
			return
		}
	}
}

// Hover sets the hovered unit index.
func (g *UnitGrid) Hover(index int) {
	g.hoveredIndex = index
}

// Navigate moves selection in the given direction.
func (g *UnitGrid) Navigate(dir shared.Direction) {
	if len(g.units) == 0 {
		return
	}

	// Initialize selection if none
	if g.selectedIndex < 0 {
		g.selectedIndex = 0
		return
	}

	switch dir {
	case shared.DirUp:
		if g.selectedIndex >= g.columns {
			g.selectedIndex -= g.columns
		}
	case shared.DirDown:
		if g.selectedIndex+g.columns < len(g.units) {
			g.selectedIndex += g.columns
		}
	case shared.DirLeft:
		if g.selectedIndex > 0 {
			g.selectedIndex--
		}
	case shared.DirRight:
		if g.selectedIndex < len(g.units)-1 {
			g.selectedIndex++
		}
	}
}

// SelectedUnit returns the currently selected unit, or nil if none.
func (g *UnitGrid) SelectedUnit() *shared.UnitState {
	if g.selectedIndex >= 0 && g.selectedIndex < len(g.units) {
		return g.units[g.selectedIndex]
	}
	return nil
}

// SelectedIndex returns the currently selected index.
func (g *UnitGrid) SelectedIndex() int {
	return g.selectedIndex
}

// Update handles messages for the grid.
func (g *UnitGrid) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		g.width = msg.Width
		g.height = msg.Height
	case tea.KeyMsg:
		if !g.focused {
			return nil
		}
		switch msg.String() {
		case "up", "k":
			g.Navigate(shared.DirUp)
		case "down", "j":
			g.Navigate(shared.DirDown)
		case "left", "h":
			g.Navigate(shared.DirLeft)
		case "right", "l":
			g.Navigate(shared.DirRight)
		}
	}
	return nil
}

// Render renders the grid.
func (g *UnitGrid) Render(width, height int) string {
	g.width = width
	g.height = height

	if len(g.units) == 0 {
		return g.renderEmpty(width, height)
	}

	// Create a bubble-grid
	bg := grid.NewStackedGrid()
	bg.SetSize(width, height)

	// Add units as framed items
	for i, unit := range g.units {
		col := i % g.columns
		cell := g.createCell(i, unit)

		// Create frame with status-colored border
		framed := frame.NewFrame(cell)
		framed = framed.ChangeBorderColor(shared.StatusBorderColor(unit.Status))

		// Highlight selected cell
		if i == g.selectedIndex {
			framed = framed.ChangeBorderColor(lipgloss.Color("15")) // Bright white
		}

		bg.AddItem(framed, grid.ItemOptions{Column: col})
	}

	return bg.Render()
}

// createCell creates a unit cell for the grid.
func (g *UnitGrid) createCell(index int, unit *shared.UnitState) *unitCell {
	return &unitCell{
		unit:       unit,
		isSelected: index == g.selectedIndex,
		isHovered:  index == g.hoveredIndex,
	}
}

// renderEmpty renders an empty grid placeholder.
func (g *UnitGrid) renderEmpty(width, height int) string {
	content := shared.DimStyle.Render("No units")
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// Focus sets whether the grid has focus.
func (g *UnitGrid) Focus(focused bool) {
	g.focused = focused
}

// IsFocused returns true if the grid has focus.
func (g *UnitGrid) IsFocused() bool {
	return g.focused
}

// unitCell implements grid.Item and grid.Sizer for unit cells.
type unitCell struct {
	unit       *shared.UnitState
	isSelected bool
	isHovered  bool
	width      int
	height     int
}

// Render implements grid.Item.
func (c *unitCell) Render() string {
	icon := c.unit.Status.Icon()
	name := shared.Truncate(c.unit.Moniker, c.width-6)
	if c.width == 0 {
		name = c.unit.Moniker
	}

	line1 := fmt.Sprintf("%s %s", icon, name)
	line2 := shared.DimStyle.Render(fmt.Sprintf("   %s w%d", c.unit.Handler, c.unit.Weight))

	content := line1 + "\n" + line2

	style := c.statusStyle()
	if c.isSelected {
		style = style.Reverse(true).Bold(true)
	} else if c.isHovered {
		style = style.Background(lipgloss.Color("238"))
	}

	if c.width > 0 && c.height > 0 {
		style = style.Width(c.width).Height(c.height)
	}

	return style.Render(content)
}

// SetSize implements grid.Sizer.
func (c *unitCell) SetSize(width, height int) grid.Sizer {
	c.width = width
	c.height = height
	return c
}

// statusStyle returns the base style for the unit status.
func (c *unitCell) statusStyle() lipgloss.Style {
	switch c.unit.Status {
	case shared.UnitRunning:
		return lipgloss.NewStyle().Foreground(shared.ColorOrange)
	case shared.UnitSuccess:
		return lipgloss.NewStyle().Foreground(shared.ColorGreen)
	case shared.UnitSkipped:
		return lipgloss.NewStyle().Foreground(shared.ColorCyan)
	case shared.UnitFailed:
		return lipgloss.NewStyle().Foreground(shared.ColorRed)
	default:
		return lipgloss.NewStyle().Foreground(shared.ColorGrey)
	}
}

// RenderSimple renders a simpler grid without bubble-grid (fallback).
func (g *UnitGrid) RenderSimple(width, height int) string {
	if len(g.units) == 0 {
		return g.renderEmpty(width, height)
	}

	cellWidth := width / g.columns
	if cellWidth < 10 {
		cellWidth = 10
	}

	var rows []string
	var currentRow []string

	for i, unit := range g.units {
		cell := g.renderSimpleCell(i, unit, cellWidth)
		currentRow = append(currentRow, cell)

		if len(currentRow) == g.columns || i == len(g.units)-1 {
			// Pad incomplete row
			for len(currentRow) < g.columns {
				currentRow = append(currentRow, strings.Repeat(" ", cellWidth))
			}
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = nil
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderSimpleCell renders a single cell without bubble-grid.
func (g *UnitGrid) renderSimpleCell(index int, unit *shared.UnitState, width int) string {
	icon := unit.Status.Icon()
	name := shared.Truncate(unit.Moniker, width-8)

	content := fmt.Sprintf("%s %s", icon, name)

	style := lipgloss.NewStyle().
		Width(width - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(shared.StatusBorderColor(unit.Status))

	if index == g.selectedIndex {
		style = style.BorderForeground(lipgloss.Color("15")).Bold(true)
	}

	return style.Render(content)
}

// RenderCompact renders tabs in a compact single-row format without borders.
// Format: [◦name₁] [▸name₂] [✓name₃] - uses Unicode status icons and subscript indices.
func (g *UnitGrid) RenderCompact(width, height int) string {
	if len(g.units) == 0 {
		return g.renderEmpty(width, height)
	}

	// Calculate how many tabs fit per row
	// Format: "[◦name] " = ~15 chars minimum
	tabWidth := 16
	tabsPerRow := width / tabWidth
	if tabsPerRow < 1 {
		tabsPerRow = 1
	}

	var rows []string
	var currentRow []string

	for i, unit := range g.units {
		tab := g.renderCompactTab(i, unit, tabWidth-2)
		currentRow = append(currentRow, tab)

		if len(currentRow) >= tabsPerRow || i == len(g.units)-1 {
			rows = append(rows, strings.Join(currentRow, " "))
			currentRow = nil
		}

		// Limit rows to fit height
		if len(rows) >= height {
			break
		}
	}

	return strings.Join(rows, "\n")
}

// renderCompactTab renders a single compact tab.
func (g *UnitGrid) renderCompactTab(index int, unit *shared.UnitState, maxNameLen int) string {
	icon := unit.Status.UnicodeIcon()
	name := shared.Truncate(unit.Moniker, maxNameLen)

	// Color the icon based on status
	iconStyle := lipgloss.NewStyle().Foreground(shared.StatusBorderColor(unit.Status))
	coloredIcon := iconStyle.Render(icon)

	// Subscript index (①②③... or just numbers)
	indexStr := subscriptIndex(index + 1)

	// Build the tab content
	content := fmt.Sprintf("%s%s%s", coloredIcon, name, indexStr)

	// Highlight selected tab
	if index == g.selectedIndex {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Bold(true).
			Render("[" + content + "]")
	}

	return "[" + content + "]"
}

// subscriptIndex returns a subscript number representation.
// Uses circled numbers ①②③... for 1-20, then falls back to plain numbers.
func subscriptIndex(n int) string {
	circled := []string{
		"", "①", "②", "③", "④", "⑤", "⑥", "⑦", "⑧", "⑨", "⑩",
		"⑪", "⑫", "⑬", "⑭", "⑮", "⑯", "⑰", "⑱", "⑲", "⑳",
	}
	if n > 0 && n < len(circled) {
		return circled[n]
	}
	return fmt.Sprintf("(%d)", n)
}
