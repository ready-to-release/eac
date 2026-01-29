package mainarea

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ready-to-release/eac/go/eac/commands/internal/tui/console/components/shared"
)

// FocusArea represents which panel has focus.
type FocusArea int

const (
	FocusGrid FocusArea = iota
	FocusOutput
)

// LayoutConstraints defines the layout parameters.
type LayoutConstraints struct {
	MinGridColumns   int
	MaxGridColumns   int
	DefaultGridRatio float64 // Grid width as fraction of total (0.0-1.0)
	MinGridWidth     int
	MinOutputWidth   int
}

// DefaultConstraints returns sensible default layout constraints.
func DefaultConstraints() LayoutConstraints {
	return LayoutConstraints{
		MinGridColumns:   2,
		MaxGridColumns:   4,
		DefaultGridRatio: 0.35, // 35% grid, 65% output
		MinGridWidth:     30,
		MinOutputWidth:   40,
	}
}

// Area coordinates the grid and output panels in the main area.
type Area struct {
	grid        *UnitGrid
	output      *OutputPanel
	constraints LayoutConstraints
	focus       FocusArea
	width       int
	height      int
	compactMode bool // Use compact single-row tabs instead of bordered cells
}

// NewArea creates a new main area with grid and output panels.
func NewArea(width, height int, columns int) *Area {
	constraints := DefaultConstraints()

	gridWidth := int(float64(width) * constraints.DefaultGridRatio)
	outputWidth := width - gridWidth - 1 // -1 for separator

	return &Area{
		grid:        NewUnitGrid(columns),
		output:      NewOutputPanel(outputWidth, height),
		constraints: constraints,
		focus:       FocusGrid,
		width:       width,
		height:      height,
		compactMode: true, // Default to compact single-row tabs
	}
}

// SetCompactMode enables or disables compact tab rendering.
func (a *Area) SetCompactMode(compact bool) {
	a.compactMode = compact
}

// ToggleCompactMode toggles between compact and bordered grid rendering.
func (a *Area) ToggleCompactMode() {
	a.compactMode = !a.compactMode
}

// SetUnits updates the units in the grid.
func (a *Area) SetUnits(units []*shared.UnitState) {
	a.grid.SetUnits(units)
}

// AddUnit adds a unit to the grid.
func (a *Area) AddUnit(unit *shared.UnitState) {
	a.grid.AddUnit(unit)
}

// UpdateUnit updates a unit's status.
func (a *Area) UpdateUnit(moniker string, status shared.UnitStatus) {
	a.grid.UpdateUnit(moniker, status)
}

// SelectUnit selects a unit by index.
func (a *Area) SelectUnit(index int) {
	a.grid.Select(index)
}

// SelectUnitByMoniker selects a unit by moniker.
func (a *Area) SelectUnitByMoniker(moniker string) {
	a.grid.SelectByMoniker(moniker)
}

// SelectedUnit returns the currently selected unit.
func (a *Area) SelectedUnit() *shared.UnitState {
	return a.grid.SelectedUnit()
}

// SetOutputLines sets the output panel lines.
func (a *Area) SetOutputLines(unitName string, lines []Line) {
	a.output.SetUnit(unitName, lines)
}

// AppendOutputLine appends a line to the output.
func (a *Area) AppendOutputLine(line Line) {
	a.output.AppendLine(line)
}

// SetFocus sets which panel has focus.
func (a *Area) SetFocus(focus FocusArea) {
	a.focus = focus
	a.grid.Focus(focus == FocusGrid)
	a.output.Focus(focus == FocusOutput)
}

// ToggleFocus switches focus between grid and output.
func (a *Area) ToggleFocus() {
	if a.focus == FocusGrid {
		a.SetFocus(FocusOutput)
	} else {
		a.SetFocus(FocusGrid)
	}
}

// Update handles messages for the main area.
func (a *Area) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			a.ToggleFocus()
			return nil
		case "c":
			// Toggle compact mode
			a.ToggleCompactMode()
			return nil
		}
	}

	// Route to focused panel
	if a.focus == FocusGrid {
		return a.grid.Update(msg)
	}
	return a.output.Update(msg)
}

// Render renders the main area with grid and output side by side.
func (a *Area) Render(width, height int) string {
	a.width = width
	a.height = height

	// Calculate widths
	gridWidth := int(float64(width) * a.constraints.DefaultGridRatio)
	if gridWidth < a.constraints.MinGridWidth {
		gridWidth = a.constraints.MinGridWidth
	}

	outputWidth := width - gridWidth - 1 // -1 for separator
	if outputWidth < a.constraints.MinOutputWidth {
		outputWidth = a.constraints.MinOutputWidth
		gridWidth = width - outputWidth - 1
	}

	// Render panels
	var gridView string
	if a.compactMode {
		gridView = a.grid.RenderCompact(gridWidth, height)
	} else {
		gridView = a.grid.RenderSimple(gridWidth, height)
	}
	outputView := a.output.Render(outputWidth, height)

	// Separator
	separator := lipgloss.NewStyle().
		Foreground(shared.ColorDim).
		Render(strings.Repeat("│\n", height))

	// Join horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		gridView,
		separator,
		outputView,
	)
}

// Grid returns the grid component for direct access.
func (a *Area) Grid() *UnitGrid {
	return a.grid
}

// Output returns the output component for direct access.
func (a *Area) Output() *OutputPanel {
	return a.output
}

// Navigate navigates the grid in the given direction.
func (a *Area) Navigate(dir shared.Direction) {
	if a.focus == FocusGrid {
		a.grid.Navigate(dir)
	}
}
