package cells

import "testing"

func TestDepsCell_Render(t *testing.T) {
	tests := []struct {
		name     string
		deps     []DependencyInfo
		width    int
		expected string
	}{
		{
			name:     "no deps",
			deps:     nil,
			width:    30,
			expected: "deps: none",
		},
		{
			name: "single dep ready",
			deps: []DependencyInfo{
				{Moniker: "contracts", Status: "ready"},
			},
			width:    40,
			expected: "deps: contracts (ready)",
		},
		{
			name: "multiple deps",
			deps: []DependencyInfo{
				{Moniker: "contracts", Status: "ready"},
				{Moniker: "godog", Status: "building"},
			},
			width:    50,
			expected: "deps: contracts (ready), godog (building)",
		},
		{
			name: "truncate long deps",
			deps: []DependencyInfo{
				{Moniker: "contracts", Status: "ready"},
				{Moniker: "godog", Status: "building"},
				{Moniker: "core", Status: "pending"},
			},
			width: 40,
			// "deps: contracts (ready), godog (building)" = 46 > 36 (40-4)
			// So we truncate before adding the second dep
			expected: "deps: contracts (ready) ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewDepsCell()
			c.SetDeps(tt.deps)
			got := stripANSI(c.Render(tt.width, 1))
			if got != tt.expected {
				t.Errorf("DepsCell.Render() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDepsCell_ZoneID(t *testing.T) {
	c := NewDepsCell()
	if got := c.ZoneID(); got != "" {
		t.Errorf("DepsCell.ZoneID() = %q, want %q", got, "")
	}
}

func TestDepsCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*DepsCell)(nil)
}
