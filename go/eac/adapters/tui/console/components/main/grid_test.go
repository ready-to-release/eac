package mainarea

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/adapters/tui/console/components/shared"
)

func TestNewUnitGrid(t *testing.T) {
	tests := []struct {
		name        string
		columns     int
		wantColumns int
	}{
		{
			name:        "standard 3 columns",
			columns:     3,
			wantColumns: 3,
		},
		{
			name:        "1 column",
			columns:     1,
			wantColumns: 1,
		},
		{
			name:        "5 columns",
			columns:     5,
			wantColumns: 5,
		},
		{
			name:        "zero columns defaults to 3",
			columns:     0,
			wantColumns: 3,
		},
		{
			name:        "negative columns defaults to 3",
			columns:     -1,
			wantColumns: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := NewUnitGrid(tt.columns)
			if grid == nil {
				t.Fatal("NewUnitGrid() returned nil")
			}
			if grid.columns != tt.wantColumns {
				t.Errorf("NewUnitGrid().columns = %d, want %d", grid.columns, tt.wantColumns)
			}
			if grid.selectedIndex != -1 {
				t.Errorf("NewUnitGrid().selectedIndex = %d, want -1", grid.selectedIndex)
			}
			if grid.hoveredIndex != -1 {
				t.Errorf("NewUnitGrid().hoveredIndex = %d, want -1", grid.hoveredIndex)
			}
		})
	}
}

func TestUnitGrid_SetUnits(t *testing.T) {
	grid := NewUnitGrid(3)

	units := []*shared.UnitState{
		{Moniker: "unit1", Status: shared.UnitPending},
		{Moniker: "unit2", Status: shared.UnitRunning},
		{Moniker: "unit3", Status: shared.UnitSuccess},
	}

	grid.SetUnits(units)

	if len(grid.units) != 3 {
		t.Errorf("SetUnits() resulted in %d units, want 3", len(grid.units))
	}
}

func TestUnitGrid_SetUnits_ClampsSelection(t *testing.T) {
	grid := NewUnitGrid(3)

	// Set 5 units and select the last one
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "unit1"}, {Moniker: "unit2"}, {Moniker: "unit3"},
		{Moniker: "unit4"}, {Moniker: "unit5"},
	})
	grid.Select(4) // Select last unit

	if grid.selectedIndex != 4 {
		t.Fatalf("Select(4) failed, got selectedIndex %d", grid.selectedIndex)
	}

	// Now set fewer units - selection should be clamped
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "unit1"}, {Moniker: "unit2"},
	})

	if grid.selectedIndex != 1 {
		t.Errorf("SetUnits() didn't clamp selection, got %d, want 1", grid.selectedIndex)
	}
}

func TestUnitGrid_AddUnit(t *testing.T) {
	grid := NewUnitGrid(3)

	// First unit should auto-select
	grid.AddUnit(&shared.UnitState{Moniker: "unit1", Status: shared.UnitPending})

	if len(grid.units) != 1 {
		t.Errorf("AddUnit() resulted in %d units, want 1", len(grid.units))
	}
	if grid.selectedIndex != 0 {
		t.Errorf("First AddUnit() should auto-select, got selectedIndex %d", grid.selectedIndex)
	}

	// Second unit should not change selection
	grid.AddUnit(&shared.UnitState{Moniker: "unit2", Status: shared.UnitRunning})

	if len(grid.units) != 2 {
		t.Errorf("AddUnit() resulted in %d units, want 2", len(grid.units))
	}
	if grid.selectedIndex != 0 {
		t.Errorf("Second AddUnit() changed selection, got selectedIndex %d", grid.selectedIndex)
	}
}

func TestUnitGrid_UpdateUnit(t *testing.T) {
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "unit1", Status: shared.UnitPending},
		{Moniker: "unit2", Status: shared.UnitPending},
	})

	grid.UpdateUnit("unit1", shared.UnitSuccess)

	if grid.units[0].Status != shared.UnitSuccess {
		t.Errorf("UpdateUnit() didn't update status, got %v", grid.units[0].Status)
	}
	if grid.units[1].Status != shared.UnitPending {
		t.Errorf("UpdateUnit() modified wrong unit, got %v", grid.units[1].Status)
	}
}

func TestUnitGrid_UpdateUnit_NotFound(t *testing.T) {
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "unit1", Status: shared.UnitPending},
	})

	// Updating non-existent unit should not panic or change anything
	grid.UpdateUnit("nonexistent", shared.UnitSuccess)

	if len(grid.units) != 1 {
		t.Errorf("UpdateUnit() changed unit count, got %d", len(grid.units))
	}
	if grid.units[0].Status != shared.UnitPending {
		t.Errorf("UpdateUnit() modified existing unit, got %v", grid.units[0].Status)
	}
}

func TestUnitGrid_Select(t *testing.T) {
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "unit1"}, {Moniker: "unit2"}, {Moniker: "unit3"},
	})

	tests := []struct {
		name      string
		index     int
		wantIndex int
	}{
		{"select first", 0, 0},
		{"select middle", 1, 1},
		{"select last", 2, 2},
		{"select negative stays unchanged", -1, 2}, // Previous selection was 2
		{"select out of bounds stays unchanged", 10, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid.Select(tt.index)
			if grid.selectedIndex != tt.wantIndex {
				t.Errorf("Select(%d) set selectedIndex to %d, want %d", tt.index, grid.selectedIndex, tt.wantIndex)
			}
		})
	}
}

func TestUnitGrid_SelectByMoniker(t *testing.T) {
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "module/comp1"}, {Moniker: "module/comp2"}, {Moniker: "other/comp3"},
	})

	tests := []struct {
		name      string
		moniker   string
		wantIndex int
	}{
		{"select first by moniker", "module/comp1", 0},
		{"select middle by moniker", "module/comp2", 1},
		{"select last by moniker", "other/comp3", 2},
		{"nonexistent moniker stays unchanged", "nonexistent", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid.SelectByMoniker(tt.moniker)
			if grid.selectedIndex != tt.wantIndex {
				t.Errorf("SelectByMoniker(%q) set selectedIndex to %d, want %d", tt.moniker, grid.selectedIndex, tt.wantIndex)
			}
		})
	}
}

func TestUnitGrid_SelectedUnit(t *testing.T) {
	tests := []struct {
		name        string
		units       []*shared.UnitState
		selectIndex int
		wantNil     bool
		wantMoniker string
	}{
		{
			name:        "empty grid returns nil",
			units:       []*shared.UnitState{},
			selectIndex: -1,
			wantNil:     true,
		},
		{
			name: "no selection returns nil",
			units: []*shared.UnitState{
				{Moniker: "unit1"},
			},
			selectIndex: -1,
			wantNil:     true,
		},
		{
			name: "valid selection returns unit",
			units: []*shared.UnitState{
				{Moniker: "unit1"}, {Moniker: "unit2"},
			},
			selectIndex: 1,
			wantMoniker: "unit2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := NewUnitGrid(3)
			grid.SetUnits(tt.units)
			if tt.selectIndex >= 0 {
				grid.Select(tt.selectIndex)
			}

			got := grid.SelectedUnit()

			if tt.wantNil {
				if got != nil {
					t.Errorf("SelectedUnit() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("SelectedUnit() = nil, want unit with moniker %q", tt.wantMoniker)
				} else if got.Moniker != tt.wantMoniker {
					t.Errorf("SelectedUnit().Moniker = %q, want %q", got.Moniker, tt.wantMoniker)
				}
			}
		})
	}
}

func TestUnitGrid_SelectedIndex(t *testing.T) {
	grid := NewUnitGrid(3)

	if grid.SelectedIndex() != -1 {
		t.Errorf("New grid SelectedIndex() = %d, want -1", grid.SelectedIndex())
	}

	grid.SetUnits([]*shared.UnitState{{Moniker: "unit1"}, {Moniker: "unit2"}})
	grid.Select(1)

	if grid.SelectedIndex() != 1 {
		t.Errorf("After Select(1), SelectedIndex() = %d, want 1", grid.SelectedIndex())
	}
}

func TestUnitGrid_Navigate_EmptyGrid(t *testing.T) {
	grid := NewUnitGrid(3)

	// Navigation on empty grid should not panic
	directions := []shared.Direction{shared.DirUp, shared.DirDown, shared.DirLeft, shared.DirRight}
	for _, dir := range directions {
		grid.Navigate(dir)
		if grid.selectedIndex != -1 {
			t.Errorf("Navigate() on empty grid changed selectedIndex to %d", grid.selectedIndex)
		}
	}
}

func TestUnitGrid_Navigate_InitializesSelection(t *testing.T) {
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "unit1"}, {Moniker: "unit2"},
	})
	// Don't auto-select by setting units after construction
	grid.selectedIndex = -1

	// First navigation should initialize selection to 0
	grid.Navigate(shared.DirDown)

	if grid.selectedIndex != 0 {
		t.Errorf("First Navigate() should initialize selection to 0, got %d", grid.selectedIndex)
	}
}

func TestUnitGrid_Navigate_SingleUnit(t *testing.T) {
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{{Moniker: "unit1"}})
	grid.Select(0)

	// All directions should stay at 0 for single unit
	directions := []shared.Direction{shared.DirUp, shared.DirDown, shared.DirLeft, shared.DirRight}
	for _, dir := range directions {
		grid.Navigate(dir)
		if grid.selectedIndex != 0 {
			t.Errorf("Navigate(%v) on single unit changed selection to %d, want 0", dir, grid.selectedIndex)
		}
	}
}

func TestUnitGrid_Navigate_Horizontal(t *testing.T) {
	// Create a grid with 6 units in 3 columns:
	// [0] [1] [2]
	// [3] [4] [5]
	// Note: Left/Right navigation is linear (wraps across rows)
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "u0"}, {Moniker: "u1"}, {Moniker: "u2"},
		{Moniker: "u3"}, {Moniker: "u4"}, {Moniker: "u5"},
	})

	tests := []struct {
		name       string
		startIndex int
		direction  shared.Direction
		wantIndex  int
	}{
		// Left navigation (linear, wraps across rows)
		{"left from middle", 1, shared.DirLeft, 0},
		{"left from right edge", 2, shared.DirLeft, 1},
		{"left from left edge (stays)", 0, shared.DirLeft, 0},
		{"left from second row middle", 4, shared.DirLeft, 3},
		{"left from second row left wraps to previous row", 3, shared.DirLeft, 2},

		// Right navigation (linear, wraps across rows)
		{"right from left edge", 0, shared.DirRight, 1},
		{"right from middle", 1, shared.DirRight, 2},
		{"right from right edge wraps to next row", 2, shared.DirRight, 3},
		{"right from last unit (stays)", 5, shared.DirRight, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid.Select(tt.startIndex)
			grid.Navigate(tt.direction)
			if grid.selectedIndex != tt.wantIndex {
				t.Errorf("Navigate(%v) from %d = %d, want %d", tt.direction, tt.startIndex, grid.selectedIndex, tt.wantIndex)
			}
		})
	}
}

func TestUnitGrid_Navigate_Vertical(t *testing.T) {
	// Create a grid with 6 units in 3 columns:
	// [0] [1] [2]
	// [3] [4] [5]
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "u0"}, {Moniker: "u1"}, {Moniker: "u2"},
		{Moniker: "u3"}, {Moniker: "u4"}, {Moniker: "u5"},
	})

	tests := []struct {
		name       string
		startIndex int
		direction  shared.Direction
		wantIndex  int
	}{
		// Down navigation
		{"down from top row", 0, shared.DirDown, 3},
		{"down from top middle", 1, shared.DirDown, 4},
		{"down from bottom row (stays)", 3, shared.DirDown, 3},
		{"down from bottom middle (stays)", 4, shared.DirDown, 4},

		// Up navigation
		{"up from bottom row", 3, shared.DirUp, 0},
		{"up from bottom middle", 4, shared.DirUp, 1},
		{"up from top row (stays)", 0, shared.DirUp, 0},
		{"up from top middle (stays)", 1, shared.DirUp, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid.Select(tt.startIndex)
			grid.Navigate(tt.direction)
			if grid.selectedIndex != tt.wantIndex {
				t.Errorf("Navigate(%v) from %d = %d, want %d", tt.direction, tt.startIndex, grid.selectedIndex, tt.wantIndex)
			}
		})
	}
}

func TestUnitGrid_Navigate_IncompleteLastRow(t *testing.T) {
	// Create a grid with 5 units in 3 columns:
	// [0] [1] [2]
	// [3] [4]
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "u0"}, {Moniker: "u1"}, {Moniker: "u2"},
		{Moniker: "u3"}, {Moniker: "u4"},
	})

	tests := []struct {
		name       string
		startIndex int
		direction  shared.Direction
		wantIndex  int
	}{
		// Down navigation with incomplete row
		{"down from top right to nowhere (stays)", 2, shared.DirDown, 2},
		{"down from top left to bottom left", 0, shared.DirDown, 3},

		// Up from incomplete row
		{"up from last row", 4, shared.DirUp, 1},

		// Right at edge of incomplete row
		{"right from last unit (stays)", 4, shared.DirRight, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid.Select(tt.startIndex)
			grid.Navigate(tt.direction)
			if grid.selectedIndex != tt.wantIndex {
				t.Errorf("Navigate(%v) from %d = %d, want %d", tt.direction, tt.startIndex, grid.selectedIndex, tt.wantIndex)
			}
		})
	}
}

func TestUnitGrid_Navigate_SingleColumn(t *testing.T) {
	// Create a grid with 1 column:
	// [0]
	// [1]
	// [2]
	// Note: With 1 column, down/up behave like right/left
	grid := NewUnitGrid(1)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "u0"}, {Moniker: "u1"}, {Moniker: "u2"},
	})

	tests := []struct {
		name       string
		startIndex int
		direction  shared.Direction
		wantIndex  int
	}{
		{"down moves to next", 0, shared.DirDown, 1},
		{"down from middle", 1, shared.DirDown, 2},
		{"down from last (stays)", 2, shared.DirDown, 2},
		{"up from middle", 1, shared.DirUp, 0},
		{"up from first (stays)", 0, shared.DirUp, 0},
		{"left from middle moves to previous", 1, shared.DirLeft, 0},
		{"right moves to next", 0, shared.DirRight, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid.Select(tt.startIndex)
			grid.Navigate(tt.direction)
			if grid.selectedIndex != tt.wantIndex {
				t.Errorf("Navigate(%v) from %d = %d, want %d", tt.direction, tt.startIndex, grid.selectedIndex, tt.wantIndex)
			}
		})
	}
}

func TestUnitGrid_Navigate_WideGrid(t *testing.T) {
	// Create a grid with 5 columns, 1 row:
	// [0] [1] [2] [3] [4]
	grid := NewUnitGrid(5)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "u0"}, {Moniker: "u1"}, {Moniker: "u2"},
		{Moniker: "u3"}, {Moniker: "u4"},
	})

	tests := []struct {
		name       string
		startIndex int
		direction  shared.Direction
		wantIndex  int
	}{
		{"left from middle", 2, shared.DirLeft, 1},
		{"right from middle", 2, shared.DirRight, 3},
		{"down from any (stays)", 2, shared.DirDown, 2},
		{"up from any (stays)", 2, shared.DirUp, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid.Select(tt.startIndex)
			grid.Navigate(tt.direction)
			if grid.selectedIndex != tt.wantIndex {
				t.Errorf("Navigate(%v) from %d = %d, want %d", tt.direction, tt.startIndex, grid.selectedIndex, tt.wantIndex)
			}
		})
	}
}

func TestUnitGrid_Focus(t *testing.T) {
	grid := NewUnitGrid(3)

	if grid.IsFocused() {
		t.Error("New grid should not be focused")
	}

	grid.Focus(true)
	if !grid.IsFocused() {
		t.Error("Focus(true) should set focused to true")
	}

	grid.Focus(false)
	if grid.IsFocused() {
		t.Error("Focus(false) should set focused to false")
	}
}

func TestUnitGrid_Hover(t *testing.T) {
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "u0"}, {Moniker: "u1"}, {Moniker: "u2"},
	})

	if grid.hoveredIndex != -1 {
		t.Errorf("New grid hoveredIndex = %d, want -1", grid.hoveredIndex)
	}

	grid.Hover(1)
	if grid.hoveredIndex != 1 {
		t.Errorf("Hover(1) set hoveredIndex to %d, want 1", grid.hoveredIndex)
	}

	grid.Hover(-1)
	if grid.hoveredIndex != -1 {
		t.Errorf("Hover(-1) set hoveredIndex to %d, want -1", grid.hoveredIndex)
	}
}

func TestUnitGrid_Render_Empty(t *testing.T) {
	grid := NewUnitGrid(3)
	got := grid.Render(80, 24)

	if got == "" {
		t.Error("Render() on empty grid returned empty string")
	}
}

func TestUnitGrid_Render_WithUnits(t *testing.T) {
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "module/comp1", Status: shared.UnitPending, Handler: "go"},
		{Moniker: "module/comp2", Status: shared.UnitRunning, Handler: "docker"},
		{Moniker: "module/comp3", Status: shared.UnitSuccess, Handler: "npm"},
	})
	grid.Select(0)

	got := grid.Render(80, 24)

	if got == "" {
		t.Error("Render() returned empty string")
	}
}

func TestUnitGrid_RenderSimple(t *testing.T) {
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "module/comp1", Status: shared.UnitPending},
		{Moniker: "module/comp2", Status: shared.UnitRunning},
	})
	grid.Select(0)

	got := grid.RenderSimple(80, 24)

	if got == "" {
		t.Error("RenderSimple() returned empty string")
	}
}

func TestUnitGrid_RenderSimple_Empty(t *testing.T) {
	grid := NewUnitGrid(3)
	got := grid.RenderSimple(80, 24)

	if got == "" {
		t.Error("RenderSimple() on empty grid returned empty string")
	}
}

func TestUnitGrid_RenderSimple_NarrowWidth(t *testing.T) {
	grid := NewUnitGrid(3)
	grid.SetUnits([]*shared.UnitState{
		{Moniker: "very/long/moniker/name", Status: shared.UnitSuccess},
	})

	// Even with very narrow width, should not panic
	got := grid.RenderSimple(20, 10)

	if got == "" {
		t.Error("RenderSimple() with narrow width returned empty string")
	}
}
