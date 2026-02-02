package cells

import (
	"strings"
	"testing"
	"time"
)

func TestSummaryCell_Render(t *testing.T) {
	tests := []struct {
		name      string
		data      *SummaryData
		contains  []string
		notEmpty  bool
	}{
		{
			name:      "nil data",
			data:      nil,
			notEmpty:  false,
			contains:  []string{},
		},
		{
			name: "all success",
			data: &SummaryData{
				Total:     10,
				Succeeded: 10,
				Failed:    0,
				Skipped:   0,
				Duration:  2 * time.Minute,
			},
			notEmpty: true,
			contains: []string{"10", "2m"},
		},
		{
			name: "mixed results",
			data: &SummaryData{
				Total:     35,
				Succeeded: 30,
				Failed:    2,
				Skipped:   3,
				Duration:  5*time.Minute + 30*time.Second,
			},
			notEmpty: true,
			contains: []string{"35", "30", "2", "3", "5m"},
		},
		{
			name: "short duration",
			data: &SummaryData{
				Total:     5,
				Succeeded: 5,
				Failed:    0,
				Skipped:   0,
				Duration:  45 * time.Second,
			},
			notEmpty: true,
			contains: []string{"5", "45"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSummaryCell()
			c.SetData(tt.data)
			got := stripANSI(c.Render(80, 3))

			if tt.notEmpty && got == "" {
				t.Error("SummaryCell.Render() returned empty for non-nil data")
			}
			if !tt.notEmpty && got != "" {
				t.Errorf("SummaryCell.Render() returned %q for nil data", got)
			}

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("SummaryCell.Render() missing %q in %q", want, got)
				}
			}
		})
	}
}

func TestSummaryCell_ZoneID(t *testing.T) {
	c := NewSummaryCell()
	if got := c.ZoneID(); got != "" {
		t.Errorf("SummaryCell.ZoneID() = %q, want %q", got, "")
	}
}

func TestSummaryCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*SummaryCell)(nil)
}
