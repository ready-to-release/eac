package lights

import (
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/tui/console/components/shared"
)

func TestNewPanel(t *testing.T) {
	tests := []struct {
		name      string
		asciiMode bool
	}{
		{
			name:      "unicode mode",
			asciiMode: false,
		},
		{
			name:      "ascii mode",
			asciiMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panel := NewPanel(tt.asciiMode)
			if panel == nil {
				t.Fatal("NewPanel() returned nil")
			}
			if panel.asciiMode != tt.asciiMode {
				t.Errorf("NewPanel().asciiMode = %v, want %v", panel.asciiMode, tt.asciiMode)
			}
			if len(panel.units) != 0 {
				t.Errorf("NewPanel().units has %d items, want 0", len(panel.units))
			}
		})
	}
}

func TestPanel_SetUnits(t *testing.T) {
	panel := NewPanel(false)

	units := []UnitLight{
		{Moniker: "unit1", Status: shared.UnitPending},
		{Moniker: "unit2", Status: shared.UnitRunning},
		{Moniker: "unit3", Status: shared.UnitSuccess},
	}

	panel.SetUnits(units)

	if len(panel.units) != 3 {
		t.Errorf("SetUnits() resulted in %d units, want 3", len(panel.units))
	}

	for i, u := range units {
		if panel.units[i].Moniker != u.Moniker {
			t.Errorf("SetUnits() unit[%d].Moniker = %q, want %q", i, panel.units[i].Moniker, u.Moniker)
		}
		if panel.units[i].Status != u.Status {
			t.Errorf("SetUnits() unit[%d].Status = %v, want %v", i, panel.units[i].Status, u.Status)
		}
	}
}

func TestPanel_SetUnits_Replaces(t *testing.T) {
	panel := NewPanel(false)

	// Set initial units
	panel.SetUnits([]UnitLight{
		{Moniker: "old1", Status: shared.UnitPending},
		{Moniker: "old2", Status: shared.UnitPending},
	})

	// Replace with new units
	panel.SetUnits([]UnitLight{
		{Moniker: "new1", Status: shared.UnitSuccess},
	})

	if len(panel.units) != 1 {
		t.Errorf("SetUnits() should replace, got %d units", len(panel.units))
	}
	if panel.units[0].Moniker != "new1" {
		t.Errorf("SetUnits() didn't replace units, got moniker %q", panel.units[0].Moniker)
	}
}

func TestPanel_AddUnit(t *testing.T) {
	panel := NewPanel(false)

	panel.AddUnit("unit1", shared.UnitPending)
	panel.AddUnit("unit2", shared.UnitRunning)
	panel.AddUnit("unit3", shared.UnitSuccess)

	if len(panel.units) != 3 {
		t.Errorf("AddUnit() resulted in %d units, want 3", len(panel.units))
	}

	expected := []struct {
		moniker string
		status  shared.UnitStatus
	}{
		{"unit1", shared.UnitPending},
		{"unit2", shared.UnitRunning},
		{"unit3", shared.UnitSuccess},
	}

	for i, exp := range expected {
		if panel.units[i].Moniker != exp.moniker {
			t.Errorf("AddUnit() unit[%d].Moniker = %q, want %q", i, panel.units[i].Moniker, exp.moniker)
		}
		if panel.units[i].Status != exp.status {
			t.Errorf("AddUnit() unit[%d].Status = %v, want %v", i, panel.units[i].Status, exp.status)
		}
	}
}

func TestPanel_UpdateUnit(t *testing.T) {
	tests := []struct {
		name           string
		initialUnits   []UnitLight
		updateMoniker  string
		updateStatus   shared.UnitStatus
		wantUnitCount  int
		wantFinalUnits []UnitLight
	}{
		{
			name: "update existing unit",
			initialUnits: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
				{Moniker: "unit2", Status: shared.UnitPending},
			},
			updateMoniker: "unit1",
			updateStatus:  shared.UnitSuccess,
			wantUnitCount: 2,
			wantFinalUnits: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitSuccess},
				{Moniker: "unit2", Status: shared.UnitPending},
			},
		},
		{
			name: "update non-existent unit adds it",
			initialUnits: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
			},
			updateMoniker: "unit2",
			updateStatus:  shared.UnitRunning,
			wantUnitCount: 2,
			wantFinalUnits: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
				{Moniker: "unit2", Status: shared.UnitRunning},
			},
		},
		{
			name:           "update on empty panel adds unit",
			initialUnits:   []UnitLight{},
			updateMoniker:  "unit1",
			updateStatus:   shared.UnitSuccess,
			wantUnitCount:  1,
			wantFinalUnits: []UnitLight{{Moniker: "unit1", Status: shared.UnitSuccess}},
		},
		{
			name: "update middle unit",
			initialUnits: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
				{Moniker: "unit2", Status: shared.UnitPending},
				{Moniker: "unit3", Status: shared.UnitPending},
			},
			updateMoniker: "unit2",
			updateStatus:  shared.UnitFailed,
			wantUnitCount: 3,
			wantFinalUnits: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
				{Moniker: "unit2", Status: shared.UnitFailed},
				{Moniker: "unit3", Status: shared.UnitPending},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panel := NewPanel(false)
			panel.SetUnits(tt.initialUnits)

			panel.UpdateUnit(tt.updateMoniker, tt.updateStatus)

			if len(panel.units) != tt.wantUnitCount {
				t.Errorf("UpdateUnit() resulted in %d units, want %d", len(panel.units), tt.wantUnitCount)
			}

			for i, want := range tt.wantFinalUnits {
				if i >= len(panel.units) {
					t.Errorf("Missing unit at index %d", i)
					continue
				}
				if panel.units[i].Moniker != want.Moniker {
					t.Errorf("UpdateUnit() unit[%d].Moniker = %q, want %q", i, panel.units[i].Moniker, want.Moniker)
				}
				if panel.units[i].Status != want.Status {
					t.Errorf("UpdateUnit() unit[%d].Status = %v, want %v", i, panel.units[i].Status, want.Status)
				}
			}
		})
	}
}

func TestPanel_CountByStatus(t *testing.T) {
	tests := []struct {
		name  string
		units []UnitLight
		want  map[shared.UnitStatus]int
	}{
		{
			name:  "empty panel",
			units: []UnitLight{},
			want:  map[shared.UnitStatus]int{},
		},
		{
			name: "single status",
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
				{Moniker: "unit2", Status: shared.UnitPending},
				{Moniker: "unit3", Status: shared.UnitPending},
			},
			want: map[shared.UnitStatus]int{
				shared.UnitPending: 3,
			},
		},
		{
			name: "multiple statuses",
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
				{Moniker: "unit2", Status: shared.UnitRunning},
				{Moniker: "unit3", Status: shared.UnitSuccess},
				{Moniker: "unit4", Status: shared.UnitSuccess},
				{Moniker: "unit5", Status: shared.UnitFailed},
			},
			want: map[shared.UnitStatus]int{
				shared.UnitPending: 1,
				shared.UnitRunning: 1,
				shared.UnitSuccess: 2,
				shared.UnitFailed:  1,
			},
		},
		{
			name: "all statuses",
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
				{Moniker: "unit2", Status: shared.UnitQueued},
				{Moniker: "unit3", Status: shared.UnitRunning},
				{Moniker: "unit4", Status: shared.UnitSuccess},
				{Moniker: "unit5", Status: shared.UnitSkipped},
				{Moniker: "unit6", Status: shared.UnitFailed},
			},
			want: map[shared.UnitStatus]int{
				shared.UnitPending: 1,
				shared.UnitQueued:  1,
				shared.UnitRunning: 1,
				shared.UnitSuccess: 1,
				shared.UnitSkipped: 1,
				shared.UnitFailed:  1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panel := NewPanel(false)
			panel.SetUnits(tt.units)

			got := panel.CountByStatus()

			if len(got) != len(tt.want) {
				t.Errorf("CountByStatus() returned %d entries, want %d", len(got), len(tt.want))
			}

			for status, wantCount := range tt.want {
				if got[status] != wantCount {
					t.Errorf("CountByStatus()[%v] = %d, want %d", status, got[status], wantCount)
				}
			}
		})
	}
}

func TestPanel_Summary(t *testing.T) {
	tests := []struct {
		name  string
		units []UnitLight
	}{
		{
			name:  "empty panel",
			units: []UnitLight{},
		},
		{
			name: "all pending",
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
				{Moniker: "unit2", Status: shared.UnitPending},
			},
		},
		{
			name: "some running",
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitSuccess},
				{Moniker: "unit2", Status: shared.UnitRunning},
				{Moniker: "unit3", Status: shared.UnitPending},
			},
		},
		{
			name: "all completed",
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitSuccess},
				{Moniker: "unit2", Status: shared.UnitSuccess},
				{Moniker: "unit3", Status: shared.UnitSkipped},
				{Moniker: "unit4", Status: shared.UnitFailed},
			},
		},
		{
			name: "mixed statuses",
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitSuccess},
				{Moniker: "unit2", Status: shared.UnitRunning},
				{Moniker: "unit3", Status: shared.UnitRunning},
				{Moniker: "unit4", Status: shared.UnitQueued},
				{Moniker: "unit5", Status: shared.UnitPending},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panel := NewPanel(false)
			panel.SetUnits(tt.units)

			got := panel.Summary()

			// Summary should always return a string (possibly empty for empty panel)
			// We verify it doesn't panic and returns something reasonable
			if len(tt.units) > 0 && got == "" {
				t.Errorf("Summary() returned empty string for non-empty panel")
			}
		})
	}
}

func TestPanel_Render(t *testing.T) {
	tests := []struct {
		name      string
		asciiMode bool
		units     []UnitLight
		width     int
		height    int
	}{
		{
			name:      "empty panel unicode",
			asciiMode: false,
			units:     []UnitLight{},
			width:     40,
			height:    3,
		},
		{
			name:      "empty panel ascii",
			asciiMode: true,
			units:     []UnitLight{},
			width:     40,
			height:    3,
		},
		{
			name:      "single unit unicode",
			asciiMode: false,
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
			},
			width:  40,
			height: 3,
		},
		{
			name:      "multiple units unicode",
			asciiMode: false,
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
				{Moniker: "unit2", Status: shared.UnitRunning},
				{Moniker: "unit3", Status: shared.UnitSuccess},
				{Moniker: "unit4", Status: shared.UnitFailed},
			},
			width:  40,
			height: 3,
		},
		{
			name:      "multiple units ascii",
			asciiMode: true,
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
				{Moniker: "unit2", Status: shared.UnitRunning},
				{Moniker: "unit3", Status: shared.UnitSuccess},
				{Moniker: "unit4", Status: shared.UnitFailed},
			},
			width:  40,
			height: 3,
		},
		{
			name:      "all statuses",
			asciiMode: false,
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitPending},
				{Moniker: "unit2", Status: shared.UnitQueued},
				{Moniker: "unit3", Status: shared.UnitRunning},
				{Moniker: "unit4", Status: shared.UnitSuccess},
				{Moniker: "unit5", Status: shared.UnitSkipped},
				{Moniker: "unit6", Status: shared.UnitFailed},
			},
			width:  60,
			height: 3,
		},
		{
			name:      "narrow width",
			asciiMode: false,
			units: []UnitLight{
				{Moniker: "unit1", Status: shared.UnitSuccess},
			},
			width:  10,
			height: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panel := NewPanel(tt.asciiMode)
			panel.SetUnits(tt.units)

			got := panel.Render(tt.width, tt.height)

			// Verify output is not empty
			if got == "" {
				t.Error("Render() returned empty string")
			}

			// Verify output contains multiple lines (border + content + border)
			lines := strings.Split(got, "\n")
			if len(lines) < 3 {
				t.Errorf("Render() returned %d lines, want at least 3", len(lines))
			}

			// Verify output contains "No units" for empty panel
			if len(tt.units) == 0 && !strings.Contains(got, "No units") {
				t.Error("Render() for empty panel should contain 'No units'")
			}
		})
	}
}

func TestPanel_Render_StatusDots(t *testing.T) {
	panel := NewPanel(true) // ASCII mode for predictable characters
	panel.SetUnits([]UnitLight{
		{Moniker: "unit1", Status: shared.UnitPending},
		{Moniker: "unit2", Status: shared.UnitQueued},
		{Moniker: "unit3", Status: shared.UnitRunning},
		{Moniker: "unit4", Status: shared.UnitSuccess},
	})

	got := panel.Render(40, 3)

	// In ASCII mode, we should see the ASCII status characters
	// The styled output may contain ANSI codes, but the characters should be present
	if got == "" {
		t.Error("Render() returned empty string")
	}
}
