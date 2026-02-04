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

func TestUoWStatsCell_RenderLampsOnly(t *testing.T) {
	tests := []struct {
		name         string
		running      int
		capacity     int
		wantFilled   int
		wantTotal    int
		wantHasLabel bool
	}{
		{
			name:         "no work",
			running:      0,
			capacity:     10,
			wantFilled:   0,
			wantTotal:    10,
			wantHasLabel: true,
		},
		{
			name:         "half capacity",
			running:      5,
			capacity:     10,
			wantFilled:   5,
			wantTotal:    10,
			wantHasLabel: true,
		},
		{
			name:         "full capacity",
			running:      8,
			capacity:     8,
			wantFilled:   8,
			wantTotal:    8,
			wantHasLabel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewUoWStatsCell()
			c.SetStats(10, tt.running, tt.capacity, 0, 0, 0)
			got := stripANSI(c.RenderLampsOnly(40, 1))

			// Should start with "Work:" label
			if tt.wantHasLabel && !strings.HasPrefix(got, "Work:") {
				t.Errorf("RenderLampsOnly() should start with 'Work:', got %q", got)
			}

			// Count filled and empty dots
			filled := strings.Count(got, "●")
			empty := strings.Count(got, "○")
			total := filled + empty

			if filled != tt.wantFilled {
				t.Errorf("RenderLampsOnly() filled = %d, want %d (got: %q)", filled, tt.wantFilled, got)
			}
			if total != tt.wantTotal {
				t.Errorf("RenderLampsOnly() total dots = %d, want %d (got: %q)", total, tt.wantTotal, got)
			}
		})
	}
}

func TestUoWStatsCell_RenderLampsOnly_ASCIIMode(t *testing.T) {
	c := NewUoWStatsCell()
	c.SetStats(10, 3, 8, 0, 0, 0)
	c.SetASCIIMode(true)
	got := stripANSI(c.RenderLampsOnly(40, 1))

	// Should start with "Work:" label
	if !strings.HasPrefix(got, "Work:") {
		t.Errorf("RenderLampsOnly() ASCII should start with 'Work:', got %q", got)
	}

	// Extract just the dots part
	dots := strings.TrimPrefix(got, "Work:")

	// Should use ASCII characters
	if strings.Contains(dots, "●") || strings.Contains(dots, "○") {
		t.Errorf("RenderLampsOnly() ASCII should not have Unicode dots, got %q", got)
	}

	filled := strings.Count(dots, "*")
	empty := strings.Count(dots, "o")

	if filled != 3 {
		t.Errorf("RenderLampsOnly() ASCII filled = %d, want 3", filled)
	}
	if filled+empty != 8 {
		t.Errorf("RenderLampsOnly() ASCII total = %d, want 8", filled+empty)
	}
}

func TestUoWStatsCell_RenderCountsOnly(t *testing.T) {
	tests := []struct {
		name    string
		total   int
		done    int
		cached  int
		failed  int
		wantPre string // Expected prefix or substring
	}{
		{
			name:    "basic counts",
			total:   8,
			done:    3,
			cached:  0,
			failed:  0,
			wantPre: "3/8",
		},
		{
			name:    "with cached",
			total:   10,
			done:    5,
			cached:  2,
			failed:  0,
			wantPre: "7/10", // done + cached = completed
		},
		{
			name:    "with failures",
			total:   8,
			done:    3,
			cached:  1,
			failed:  2,
			wantPre: "4/8", // done + cached = 4 completed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewUoWStatsCell()
			c.SetStats(tt.total, 0, 8, tt.done, tt.cached, tt.failed)
			got := stripANSI(c.RenderCountsOnly(40, 1))

			// Should contain the completed/total ratio
			if !strings.Contains(got, tt.wantPre) {
				t.Errorf("RenderCountsOnly() = %q, want to contain %q", got, tt.wantPre)
			}
		})
	}
}

func TestUoWStatsCell_RenderCountsOnly_Format(t *testing.T) {
	c := NewUoWStatsCell()
	c.SetStats(8, 2, 16, 3, 0, 0)
	got := stripANSI(c.RenderCountsOnly(40, 1))

	// Should contain: completed/total P:pending F:failed
	// done=3, cached=0, so completed=3
	// pending = total - done - cached - failed - running = 8 - 3 - 0 - 0 - 2 = 3
	if !strings.Contains(got, "3/8") {
		t.Errorf("RenderCountsOnly() should contain '3/8', got %q", got)
	}
	if !strings.Contains(got, "P:") {
		t.Errorf("RenderCountsOnly() should contain 'P:', got %q", got)
	}
	if !strings.Contains(got, "F:") {
		t.Errorf("RenderCountsOnly() should contain 'F:', got %q", got)
	}
}
