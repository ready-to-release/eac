package cells

import (
	"testing"
	"time"
)

func TestTimerCell_Render(t *testing.T) {
	tests := []struct {
		name     string
		elapsed  time.Duration
		ascii    bool
		expected string
	}{
		{"zero", 0, false, "00:00"},
		{"seconds only", 45 * time.Second, false, "00:45"},
		{"minutes and seconds", 5*time.Minute + 30*time.Second, false, "05:30"},
		{"hour plus", 65 * time.Minute, false, "65:00"},
		{"exact minute", 1 * time.Minute, false, "01:00"},
		{"large duration", 120 * time.Minute, false, "120:00"},
		{"subsecond ignored", 500 * time.Millisecond, false, "00:00"},
		{"seconds plus ms", 45*time.Second + 500*time.Millisecond, false, "00:45"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewTimerCell()
			c.SetElapsed(tt.elapsed)
			c.SetASCIIMode(tt.ascii)
			got := stripANSI(c.Render(10, 1))
			if got != tt.expected {
				t.Errorf("TimerCell.Render() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTimerCell_ZoneID(t *testing.T) {
	c := NewTimerCell()
	if got := c.ZoneID(); got != "res-timer" {
		t.Errorf("TimerCell.ZoneID() = %q, want %q", got, "res-timer")
	}
}

func TestTimerCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*TimerCell)(nil)
}
