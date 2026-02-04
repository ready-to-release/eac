package cells

import (
	"strings"
	"testing"
)

func TestDockerMemCell_Render(t *testing.T) {
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
			c := NewDockerMemCell()
			c.SetPercent(tt.percent)
			c.SetAvailable(true)
			got := c.Render(40, 1)

			stripped := stripANSI(got)
			filled := strings.Count(stripped, "●")

			// Allow ±1 variance due to rounding
			if filled < tt.wantDots-1 || filled > tt.wantDots+1 {
				t.Errorf("DockerMemCell filled = %d, want ~%d (got: %q)", filled, tt.wantDots, stripped)
			}
		})
	}
}

func TestDockerMemCell_HasLabel(t *testing.T) {
	c := NewDockerMemCell()
	c.SetPercent(50)
	c.SetAvailable(true)
	got := stripANSI(c.Render(40, 1))

	if !strings.HasPrefix(got, "DockerMem:") {
		t.Errorf("DockerMemCell.Render() should start with 'DockerMem:', got %q", got)
	}
}

func TestDockerMemCell_TotalDots(t *testing.T) {
	c := NewDockerMemCell()
	c.SetPercent(50)
	c.SetAvailable(true)
	got := stripANSI(c.Render(40, 1))

	// Should have 11 total dots (filled + empty)
	total := strings.Count(got, "●") + strings.Count(got, "○")
	if total != 11 {
		t.Errorf("DockerMemCell total dots = %d, want 11 (got: %q)", total, got)
	}
}

func TestDockerMemCell_UnavailableShowsEmpty(t *testing.T) {
	c := NewDockerMemCell()
	c.SetPercent(50) // Percent is set but should be ignored
	c.SetAvailable(false)
	got := stripANSI(c.Render(40, 1))

	// When unavailable, should show all empty dots
	filled := strings.Count(got, "●")
	empty := strings.Count(got, "○")

	if filled != 0 {
		t.Errorf("DockerMemCell unavailable should have 0 filled, got %d", filled)
	}
	if empty != 11 {
		t.Errorf("DockerMemCell unavailable should have 11 empty, got %d (output: %q)", empty, got)
	}
}

func TestDockerMemCell_ASCIIMode(t *testing.T) {
	c := NewDockerMemCell()
	c.SetPercent(50)
	c.SetAvailable(true)
	c.SetASCIIMode(true)
	got := stripANSI(c.Render(40, 1))

	// Should use ASCII characters
	if strings.Contains(got, "●") || strings.Contains(got, "○") {
		t.Errorf("DockerMemCell ASCII mode should not have Unicode dots, got %q", got)
	}

	// Extract just the dots part (after "DockerMem:")
	dots := strings.TrimPrefix(got, "DockerMem:")
	filled := strings.Count(dots, "*")
	empty := strings.Count(dots, "o")
	total := filled + empty

	if total != 11 {
		t.Errorf("DockerMemCell ASCII mode total dots = %d, want 11 (got: %q)", total, got)
	}
}

func TestDockerMemCell_ASCIIModeUnavailable(t *testing.T) {
	c := NewDockerMemCell()
	c.SetAvailable(false)
	c.SetASCIIMode(true)
	got := stripANSI(c.Render(40, 1))

	// Extract just the dots part (after "DockerMem:")
	dots := strings.TrimPrefix(got, "DockerMem:")

	// When unavailable in ASCII mode, should show all 'o'
	filled := strings.Count(dots, "*")
	empty := strings.Count(dots, "o")

	if filled != 0 {
		t.Errorf("DockerMemCell ASCII unavailable should have 0 filled, got %d", filled)
	}
	if empty != 11 {
		t.Errorf("DockerMemCell ASCII unavailable should have 11 empty, got %d (output: %q)", empty, got)
	}
}

func TestDockerMemCell_ZoneID(t *testing.T) {
	c := NewDockerMemCell()
	if got := c.ZoneID(); got != "res-dmem" {
		t.Errorf("DockerMemCell.ZoneID() = %q, want %q", got, "res-dmem")
	}
}

func TestDockerMemCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*DockerMemCell)(nil)
}

func TestDockerMemCell_OverflowPercent(t *testing.T) {
	c := NewDockerMemCell()
	c.SetPercent(150) // Over 100%
	c.SetAvailable(true)
	got := stripANSI(c.Render(40, 1))

	// Should cap at 11 filled dots
	filled := strings.Count(got, "●")
	if filled > 11 {
		t.Errorf("DockerMemCell should cap at 11 filled, got %d", filled)
	}
}
