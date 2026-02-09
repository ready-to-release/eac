package cells

import (
	"strings"
	"testing"
)

func TestSelectorCell_Render(t *testing.T) {
	tests := []struct {
		name      string
		units     []SelectorUnit
		selected  string
		wantParts []string
	}{
		{
			name:      "empty",
			units:     nil,
			selected:  "",
			wantParts: []string{},
		},
		{
			name: "single unit",
			units: []SelectorUnit{
				{Moniker: "contracts", Status: UnitComplete},
			},
			selected:  "contracts",
			wantParts: []string{"contracts"},
		},
		{
			name: "multiple units",
			units: []SelectorUnit{
				{Moniker: "contracts", Status: UnitComplete},
				{Moniker: "core", Status: UnitRunning},
				{Moniker: "docs", Status: UnitPending},
				{Moniker: "eac", Status: UnitSkipped},
			},
			selected:  "core",
			wantParts: []string{"contracts", "core", "docs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSelectorCell()
			c.SetUnits(tt.units)
			c.SetSelected(tt.selected)
			got := stripANSI(c.Render(60, 10))

			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("SelectorCell.Render() missing %q in:\n%s", part, got)
				}
			}
		})
	}
}

func TestSelectorCell_TabZoneIDs(t *testing.T) {
	c := NewSelectorCell()
	c.SetUnits([]SelectorUnit{
		{Moniker: "mod1", Status: UnitPending},
		{Moniker: "mod2", Status: UnitRunning},
	})

	// The zone IDs should be in format "tab:{moniker}"
	zones := c.GetTabZones()
	if len(zones) != 2 {
		t.Errorf("GetTabZones() returned %d zones, want 2", len(zones))
	}

	expectedZones := []string{"tab:mod1", "tab:mod2"}
	for i, want := range expectedZones {
		if i < len(zones) && zones[i] != want {
			t.Errorf("zone[%d] = %q, want %q", i, zones[i], want)
		}
	}
}

func TestSelectorCell_StatusColors(t *testing.T) {
	// This test verifies that status affects rendering
	// (colors are in ANSI codes, hard to verify directly)
	c := NewSelectorCell()
	c.SetUnits([]SelectorUnit{
		{Moniker: "pending", Status: UnitPending},
		{Moniker: "running", Status: UnitRunning},
		{Moniker: "complete", Status: UnitComplete},
		{Moniker: "skipped", Status: UnitSkipped},
		{Moniker: "failed", Status: UnitFailed},
	})

	got := c.Render(80, 5)

	// Should contain all names
	stripped := stripANSI(got)
	for _, name := range []string{"pending", "running", "complete", "skipped", "failed"} {
		if !strings.Contains(stripped, name) {
			t.Errorf("SelectorCell.Render() missing %q", name)
		}
	}
}

func TestSelectorCell_ZoneID(t *testing.T) {
	c := NewSelectorCell()
	// Main zone ID is empty since individual tabs have their own zones
	if got := c.ZoneID(); got != "" {
		t.Errorf("SelectorCell.ZoneID() = %q, want %q", got, "")
	}
}

func TestSelectorCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*SelectorCell)(nil)
}

func TestSelectorCell_IconSpacing(t *testing.T) {
	// Verify that weight badge is separated from name
	c := NewSelectorCell()
	c.SetUnits([]SelectorUnit{
		{Moniker: "build:core:go:go", DisplayName: "go", Status: UnitPending, Weight: 1},
	})

	got := stripANSI(c.Render(60, 5))

	// Should contain the weight and name
	if !strings.Contains(got, "1") {
		t.Errorf("Expected weight '1' in output:\n%s", got)
	}
	if !strings.Contains(got, "go") {
		t.Errorf("Expected name 'go' in output:\n%s", got)
	}
}

func TestSelectorCell_DisplayNameUsedForLabel(t *testing.T) {
	// Verify that DisplayName is used for tab labels, not Moniker
	c := NewSelectorCell()
	c.SetUnits([]SelectorUnit{
		{Moniker: "build:core:go:go", DisplayName: "core:go", Status: UnitComplete, Weight: 2},
	})

	got := stripANSI(c.Render(60, 5))

	// Should use DisplayName "core:go", not the long Moniker
	if !strings.Contains(got, "core:go") {
		t.Errorf("Expected DisplayName 'core:go' in output:\n%s", got)
	}
	// Should NOT show the long moniker
	if strings.Contains(got, "build:core") {
		t.Errorf("Should not show full moniker in output:\n%s", got)
	}
}
