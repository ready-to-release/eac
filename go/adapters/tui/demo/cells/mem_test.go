package cells

import (
	"strings"
	"testing"
)

func TestMemCell_Render(t *testing.T) {
	tests := []struct {
		name     string
		percent  float64
		wantDots int // filled out of 11
	}{
		{"0%", 0, 0},
		{"50%", 50, 5},
		{"100%", 100, 11},
		{"27%", 27, 3},
		{"90%", 90, 10},
		{"10%", 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMemCell()
			c.SetPercent(tt.percent)
			got := c.Render(40, 1)

			stripped := stripANSI(got)
			filled := strings.Count(stripped, "●")

			// Allow ±1 variance
			if filled < tt.wantDots-1 || filled > tt.wantDots+1 {
				t.Errorf("MemCell filled = %d, want ~%d (got: %q)", filled, tt.wantDots, stripped)
			}
		})
	}
}

func TestMemCell_HasLabel(t *testing.T) {
	c := NewMemCell()
	c.SetPercent(50)
	got := stripANSI(c.Render(40, 1))

	if !strings.HasPrefix(got, "Mem:") {
		t.Errorf("MemCell.Render() should start with 'Mem:', got %q", got)
	}
}

func TestMemCell_TotalDots(t *testing.T) {
	c := NewMemCell()
	c.SetPercent(50)
	got := stripANSI(c.Render(40, 1))

	// Should have 11 total dots (filled + empty)
	total := strings.Count(got, "●") + strings.Count(got, "○")
	if total != 11 {
		t.Errorf("MemCell total dots = %d, want 11 (got: %q)", total, got)
	}
}

func TestMemCell_ZoneID(t *testing.T) {
	c := NewMemCell()
	if got := c.ZoneID(); got != "res-mem" {
		t.Errorf("MemCell.ZoneID() = %q, want %q", got, "res-mem")
	}
}

func TestMemCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*MemCell)(nil)
}
