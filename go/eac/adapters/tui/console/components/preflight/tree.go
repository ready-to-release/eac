// Package preflight provides components for the pre-flight section of the TUI.
package preflight

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ready-to-release/eac/go/eac/adapters/tui/console/components/shared"
)

// TreePanel displays the execution tree using a scrollable viewport.
type TreePanel struct {
	vp      viewport.Model
	layers  []shared.ExecutionLayer
	focused bool
}

// NewTreePanel creates a new tree panel with the given dimensions.
func NewTreePanel(width, height int) *TreePanel {
	vp := viewport.New(width, height)
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 1 // Fine-grained scroll for tree
	return &TreePanel{vp: vp}
}

// SetLayers updates the execution layers displayed in the tree.
func (t *TreePanel) SetLayers(layers []shared.ExecutionLayer) {
	t.layers = layers
	t.vp.SetContent(t.renderContent())
}

// renderContent renders the tree content as a string.
func (t *TreePanel) renderContent() string {
	if len(t.layers) == 0 {
		return shared.DimStyle.Render("No execution tree available")
	}

	var lines []string

	for layerIdx, layer := range t.layers {
		// Layer header
		lines = append(lines, shared.LayerHeader(layerIdx+1))

		for modIdx, module := range layer.Modules {
			isLast := modIdx == len(layer.Modules)-1
			branch := shared.TreeBranch(isLast)

			modLine := fmt.Sprintf("  %s %s", branch, module.Name)
			if len(module.Components) > 0 {
				compStr := strings.Join(module.Components, ", ")
				modLine += shared.DimStyle.Render(" - ") + compStr
			}
			lines = append(lines, modLine)
		}

		// Add spacing between layers
		if layerIdx < len(t.layers)-1 {
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}

// Update handles messages for the tree panel.
func (t *TreePanel) Update(msg tea.Msg) tea.Cmd {
	if !t.focused {
		return nil
	}

	var cmd tea.Cmd
	t.vp, cmd = t.vp.Update(msg)
	return cmd
}

// Render renders the tree panel.
func (t *TreePanel) Render(width, height int) string {
	t.vp.Width = width
	t.vp.Height = height
	return t.vp.View()
}

// Focus sets whether the tree panel has keyboard focus.
func (t *TreePanel) Focus(focused bool) {
	t.focused = focused
}

// IsFocused returns true if the tree panel has keyboard focus.
func (t *TreePanel) IsFocused() bool {
	return t.focused
}

// ScrollPercent returns the current scroll position as a percentage.
func (t *TreePanel) ScrollPercent() float64 {
	return t.vp.ScrollPercent()
}

// AtTop returns true if the viewport is at the top.
func (t *TreePanel) AtTop() bool {
	return t.vp.AtTop()
}

// AtBottom returns true if the viewport is at the bottom.
func (t *TreePanel) AtBottom() bool {
	return t.vp.AtBottom()
}

// ScrollIndicator returns a scroll indicator string.
func (t *TreePanel) ScrollIndicator() string {
	if t.vp.AtTop() {
		return "TOP"
	}
	if t.vp.AtBottom() {
		return "END"
	}
	return fmt.Sprintf("%d%%", int(t.vp.ScrollPercent()*100))
}

// borderStyle returns the border style based on focus state.
func (t *TreePanel) borderStyle() lipgloss.Style {
	if t.focused {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(shared.ColorCyan)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(shared.ColorDim)
}

// RenderWithBorder renders the tree panel with a border and title.
func (t *TreePanel) RenderWithBorder(width, height int, title string) string {
	// Account for border (2 chars each side) and title line
	innerWidth := width - 4
	innerHeight := height - 2

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	content := t.Render(innerWidth, innerHeight)

	style := t.borderStyle().
		Width(innerWidth).
		Height(innerHeight)

	if title != "" {
		style = style.BorderTop(true)
	}

	return style.Render(content)
}
