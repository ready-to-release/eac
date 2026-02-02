package cells

import (
	"strings"
	"testing"
)

func TestContainersCell_Render(t *testing.T) {
	tests := []struct {
		name      string
		active    []string
		planned   []string
		wantFill  int
		wantEmpty int
	}{
		{
			name:      "no containers",
			active:    []string{},
			planned:   []string{},
			wantFill:  0,
			wantEmpty: 0,
		},
		{
			name:      "all planned inactive",
			active:    []string{},
			planned:   []string{"go", "trivy", "zap"},
			wantFill:  0,
			wantEmpty: 3,
		},
		{
			name:      "one active",
			active:    []string{"go"},
			planned:   []string{"go", "trivy", "zap"},
			wantFill:  1,
			wantEmpty: 2,
		},
		{
			name:      "all active",
			active:    []string{"go", "trivy", "zap"},
			planned:   []string{"go", "trivy", "zap"},
			wantFill:  3,
			wantEmpty: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContainersCell()
			c.SetContainers(tt.active, tt.planned)
			got := stripANSI(c.Render(40, 1))

			filled := strings.Count(got, "●")
			empty := strings.Count(got, "○")

			if filled != tt.wantFill {
				t.Errorf("ContainersCell filled = %d, want %d (got: %q)", filled, tt.wantFill, got)
			}
			if empty != tt.wantEmpty {
				t.Errorf("ContainersCell empty = %d, want %d (got: %q)", empty, tt.wantEmpty, got)
			}
		})
	}
}

func TestContainersCell_HasLabel(t *testing.T) {
	c := NewContainersCell()
	c.SetContainers([]string{"go"}, []string{"go", "trivy"})
	got := stripANSI(c.Render(40, 1))

	if !strings.HasPrefix(got, "Containers:") {
		t.Errorf("ContainersCell.Render() should start with 'Containers:', got %q", got)
	}
}

func TestContainersCell_ZoneID(t *testing.T) {
	c := NewContainersCell()
	if got := c.ZoneID(); got != "res-jobs" {
		t.Errorf("ContainersCell.ZoneID() = %q, want %q", got, "res-jobs")
	}
}

func TestContainersCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*ContainersCell)(nil)
}
