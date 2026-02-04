package cells

import (
	"strings"
	"testing"
)

func TestCPUCell_Render(t *testing.T) {
	tests := []struct {
		name     string
		percents []float64
		ascii    bool
		wantDots int // filled dots
	}{
		{"idle 4 cores", []float64{0, 0, 0, 0}, false, 0},
		{"half 4 cores", []float64{50, 50, 50, 50}, false, 2},
		{"full 4 cores", []float64{100, 100, 100, 100}, false, 4},
		{"mixed", []float64{25, 75, 50, 100}, false, 2}, // 62.5% avg = 2-3 dots
		{"8 cores idle", []float64{0, 0, 0, 0, 0, 0, 0, 0}, false, 0},
		{"8 cores full", []float64{100, 100, 100, 100, 100, 100, 100, 100}, false, 8},
		{"single core", []float64{50}, false, 0}, // 50% of 1 = 0.5, rounds to 0 or 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCPUCell()
			c.SetPercents(tt.percents)
			c.SetASCIIMode(tt.ascii)
			got := c.Render(40, 1)

			// Count filled dots
			stripped := stripANSI(got)
			filled := strings.Count(stripped, "●")
			if tt.ascii {
				filled = strings.Count(stripped, "*")
			}

			// Allow ±1 variance due to rounding
			if filled < tt.wantDots-1 || filled > tt.wantDots+1 {
				t.Errorf("CPUCell filled dots = %d, want ~%d (got: %q)", filled, tt.wantDots, stripped)
			}
		})
	}
}

func TestCPUCell_HasLabel(t *testing.T) {
	c := NewCPUCell()
	c.SetPercents([]float64{50, 50, 50, 50})
	got := stripANSI(c.Render(40, 1))

	if !strings.HasPrefix(got, "CPU:") {
		t.Errorf("CPUCell.Render() should start with 'CPU:', got %q", got)
	}
}

func TestCPUCell_ASCIIMode(t *testing.T) {
	c := NewCPUCell()
	c.SetPercents([]float64{100, 100})
	c.SetASCIIMode(true)
	got := stripANSI(c.Render(40, 1))

	// Should use * for filled and o for empty
	if !strings.Contains(got, "*") {
		t.Errorf("CPUCell ASCII mode should use *, got %q", got)
	}
	if strings.Contains(got, "●") {
		t.Errorf("CPUCell ASCII mode should not use ●, got %q", got)
	}
}

func TestCPUCell_ZoneID(t *testing.T) {
	c := NewCPUCell()
	if got := c.ZoneID(); got != "res-cpu" {
		t.Errorf("CPUCell.ZoneID() = %q, want %q", got, "res-cpu")
	}
}

func TestCPUCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*CPUCell)(nil)
}
