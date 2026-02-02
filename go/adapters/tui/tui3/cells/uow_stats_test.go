package cells

import (
	"strings"
	"testing"
)

func TestUoWStatsCell_Render(t *testing.T) {
	tests := []struct {
		name       string
		total      int
		running    int
		capacity   int
		done       int
		cached     int
		failed     int
		ascii      bool
		wantPrefix string // Check prefix matches
	}{
		{
			name:       "all pending",
			total:      35,
			running:    0,
			capacity:   16,
			done:       0,
			cached:     0,
			failed:     0,
			ascii:      false,
			wantPrefix: "UoW: 35/35",
		},
		{
			name:       "some running",
			total:      35,
			running:    2,
			capacity:   16,
			done:       0,
			cached:     0,
			failed:     0,
			ascii:      false,
			wantPrefix: "UoW: 33/35",
		},
		{
			name:       "mixed states",
			total:      35,
			running:    2,
			capacity:   16,
			done:       10,
			cached:     5,
			failed:     1,
			ascii:      false,
			wantPrefix: "UoW: 17/35", // 35 - 10 - 5 - 1 - 2 = 17 remaining
		},
		{
			name:       "all complete",
			total:      10,
			running:    0,
			capacity:   4,
			done:       8,
			cached:     2,
			failed:     0,
			ascii:      false,
			wantPrefix: "UoW: 0/10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewUoWStatsCell()
			c.SetStats(tt.total, tt.running, tt.capacity, tt.done, tt.cached, tt.failed)
			c.SetASCIIMode(tt.ascii)
			got := stripANSI(c.Render(80, 1))

			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("UoWStatsCell.Render() = %q, want prefix %q", got, tt.wantPrefix)
			}
		})
	}
}

func TestUoWStatsCell_Format(t *testing.T) {
	c := NewUoWStatsCell()
	c.SetStats(35, 2, 16, 10, 5, 1)
	c.SetASCIIMode(false)
	got := stripANSI(c.Render(80, 1))

	// Should contain all parts (with Unicode)
	// %2d formats running with leading space for single digits
	expected := "UoW: 17/35 | ▶  2/16 | ✓ 10 ⏭  5 ✗  1"
	if got != expected {
		t.Errorf("UoWStatsCell.Render() = %q, want %q", got, expected)
	}
}

func TestUoWStatsCell_ASCIIMode(t *testing.T) {
	c := NewUoWStatsCell()
	c.SetStats(35, 2, 16, 10, 5, 1)
	c.SetASCIIMode(true)
	got := stripANSI(c.Render(80, 1))

	// Should contain ASCII characters instead of Unicode
	// %2d formats running with leading space for single digits
	expected := "UoW: 17/35 | >  2/16 | V 10 =  5 X  1"
	if got != expected {
		t.Errorf("UoWStatsCell.Render() ASCII = %q, want %q", got, expected)
	}
}

func TestUoWStatsCell_ZoneID(t *testing.T) {
	c := NewUoWStatsCell()
	if got := c.ZoneID(); got != "res-uow" {
		t.Errorf("UoWStatsCell.ZoneID() = %q, want %q", got, "res-uow")
	}
}

func TestUoWStatsCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*UoWStatsCell)(nil)
}
