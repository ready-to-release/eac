package cells

import (
	"strings"
	"testing"
)

func TestToolsCell_Render(t *testing.T) {
	tests := []struct {
		name              string
		plannedContainerTools []string
		activeContainerTools  []string
		plannedSystem     []string
		activeSystem      []string
		wantTotal         int // total dots
	}{
		{
			name:              "no tools",
			plannedContainerTools: []string{},
			activeContainerTools:  []string{},
			plannedSystem:     []string{},
			activeSystem:      []string{},
			wantTotal:         0,
		},
		{
			name:              "container tools only",
			plannedContainerTools: []string{"go", "trivy"},
			activeContainerTools:  []string{"go"},
			plannedSystem:     []string{},
			activeSystem:      []string{},
			wantTotal:         2,
		},
		{
			name:              "system tools only",
			plannedContainerTools: []string{},
			activeContainerTools:  []string{},
			plannedSystem:     []string{"git", "curl"},
			activeSystem:      []string{"git"},
			wantTotal:         2,
		},
		{
			name:              "mixed tools",
			plannedContainerTools: []string{"go", "trivy"},
			activeContainerTools:  []string{"go"},
			plannedSystem:     []string{"git"},
			activeSystem:      []string{},
			wantTotal:         3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewToolsCell()
			c.SetTools(tt.plannedContainerTools, tt.activeContainerTools, tt.plannedSystem, tt.activeSystem)
			got := stripANSI(c.Render(40, 1))

			total := strings.Count(got, "●") + strings.Count(got, "○")
			if total != tt.wantTotal {
				t.Errorf("ToolsCell total dots = %d, want %d (got: %q)", total, tt.wantTotal, got)
			}
		})
	}
}

func TestToolsCell_HasLabel(t *testing.T) {
	c := NewToolsCell()
	c.SetTools([]string{"go"}, []string{"go"}, []string{}, []string{})
	got := stripANSI(c.Render(40, 1))

	if !strings.HasPrefix(got, "Tools:") {
		t.Errorf("ToolsCell.Render() should start with 'Tools:', got %q", got)
	}
}

func TestToolsCell_ZoneID(t *testing.T) {
	c := NewToolsCell()
	if got := c.ZoneID(); got != "res-tools" {
		t.Errorf("ToolsCell.ZoneID() = %q, want %q", got, "res-tools")
	}
}

func TestToolsCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*ToolsCell)(nil)
}
