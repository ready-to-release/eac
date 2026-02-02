package preflight

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ready-to-release/eac/go/adapters/tui/console/components/shared"
)

// Layout defines the pre-flight section layout dimensions.
type Layout struct {
	HeaderHeight  int
	ContentHeight int
	SelectedWidth int // As fraction of total width (0.0-1.0)
}

// DefaultLayout returns the default pre-flight layout.
func DefaultLayout() Layout {
	return Layout{
		HeaderHeight:  2,
		ContentHeight: 8,
		SelectedWidth: 30, // 30% for selected panel
	}
}

// Section coordinates the pre-flight components: header, selected, and tree.
type Section struct {
	header   *Header
	selected *SelectedPanel
	tree     *TreePanel
	layout   Layout

	collapsed bool
	focused   bool
	width     int
	height    int
}

// NewSection creates a new pre-flight section.
func NewSection(width, height int) *Section {
	layout := DefaultLayout()

	treeWidth := width * 70 / 100 // 70% for tree
	treeHeight := layout.ContentHeight

	return &Section{
		header:   NewHeader(),
		selected: NewSelectedPanel(),
		tree:     NewTreePanel(treeWidth, treeHeight),
		layout:   layout,
		width:    width,
		height:   height,
	}
}

// SetPhase updates the current phase.
func (s *Section) SetPhase(phase Phase, name string) {
	s.header.SetPhase(phase, name)
}

// SetPaused updates the paused state.
func (s *Section) SetPaused(paused bool) {
	s.header.SetPaused(paused)
}

// SetSelectedUnit updates the selected unit details.
func (s *Section) SetSelectedUnit(unit *shared.UnitDetails) {
	s.selected.SetUnit(unit)
}

// SetLayers updates the execution tree layers.
func (s *Section) SetLayers(layers []shared.ExecutionLayer) {
	s.tree.SetLayers(layers)
}

// SetCollapsed sets whether the section is collapsed.
func (s *Section) SetCollapsed(collapsed bool) {
	s.collapsed = collapsed
}

// IsCollapsed returns true if the section is collapsed.
func (s *Section) IsCollapsed() bool {
	return s.collapsed
}

// Update handles messages for the pre-flight section.
func (s *Section) Update(msg tea.Msg) tea.Cmd {
	if s.collapsed {
		return nil
	}

	// Route to tree panel if it has focus
	return s.tree.Update(msg)
}

// Render renders the pre-flight section.
func (s *Section) Render(width, height int) string {
	s.width = width
	s.height = height

	if s.collapsed {
		return s.renderCollapsed(width)
	}

	return s.renderExpanded(width, height)
}

// renderCollapsed renders a minimal collapsed view.
func (s *Section) renderCollapsed(width int) string {
	return s.header.Render(width, s.layout.HeaderHeight)
}

// renderExpanded renders the full pre-flight section.
func (s *Section) renderExpanded(width, height int) string {
	var b strings.Builder

	// Header (full width)
	headerView := s.header.Render(width, s.layout.HeaderHeight)
	b.WriteString(headerView)
	b.WriteString("\n")

	// Content area: selected (left) | tree (right)
	contentHeight := height - s.layout.HeaderHeight - 1 // -1 for newline
	if contentHeight < 1 {
		contentHeight = 1
	}

	selectedWidth := width * s.layout.SelectedWidth / 100
	treeWidth := width - selectedWidth - 3 // -3 for separator

	selectedView := s.selected.Render(selectedWidth, contentHeight)
	treeView := s.tree.Render(treeWidth, contentHeight)

	// Horizontal join
	contentView := shared.HorizontalJoin(selectedView, treeView, contentHeight)
	b.WriteString(contentView)

	return b.String()
}

// RenderWithProgress renders with progress indicator.
func (s *Section) RenderWithProgress(width, height int, completed, total int) string {
	s.width = width
	s.height = height

	if s.collapsed {
		return s.header.RenderWithProgress(width, s.layout.HeaderHeight, completed, total)
	}

	var b strings.Builder

	// Header with progress
	headerView := s.header.RenderWithProgress(width, s.layout.HeaderHeight, completed, total)
	b.WriteString(headerView)
	b.WriteString("\n")

	// Content area
	contentHeight := height - s.layout.HeaderHeight - 1
	if contentHeight < 1 {
		contentHeight = 1
	}

	selectedWidth := width * s.layout.SelectedWidth / 100
	treeWidth := width - selectedWidth - 3

	selectedView := s.selected.Render(selectedWidth, contentHeight)
	treeView := s.tree.Render(treeWidth, contentHeight)

	contentView := shared.HorizontalJoin(selectedView, treeView, contentHeight)
	b.WriteString(contentView)

	return b.String()
}

// Focus sets focus on the section.
func (s *Section) Focus(focused bool) {
	s.focused = focused
	// When section gains focus, focus the tree by default
	s.tree.Focus(focused)
}

// IsFocused returns true if the section has focus.
func (s *Section) IsFocused() bool {
	return s.focused
}

// Height returns the total height of the section.
func (s *Section) Height() int {
	if s.collapsed {
		return s.layout.HeaderHeight
	}
	return s.layout.HeaderHeight + s.layout.ContentHeight
}
