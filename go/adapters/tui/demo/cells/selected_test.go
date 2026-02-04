package cells

import "testing"

func TestSelectedCell_Render(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		primary  string
		tool     string
		status   string
		expected string
	}{
		{
			name:     "no selection",
			count:    0,
			primary:  "",
			tool:     "",
			status:   "",
			expected: "(no selection)",
		},
		{
			name:     "single selection",
			count:    1,
			primary:  "core:go",
			tool:     "go",
			status:   "running 2.3s",
			expected: "core:go │ go │ running 2.3s",
		},
		{
			name:     "multi selection",
			count:    3,
			primary:  "",
			tool:     "",
			status:   "",
			expected: "3 rows",
		},
		{
			name:     "many selected",
			count:    15,
			primary:  "",
			tool:     "",
			status:   "",
			expected: "15 rows",
		},
		{
			name:     "single with different details",
			count:    1,
			primary:  "contracts",
			tool:     "structurizr",
			status:   "complete 1.2s",
			expected: "contracts │ structurizr │ complete 1.2s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewSelectedCell()
			c.SetSelection(tt.count, tt.primary, tt.tool, tt.status)
			got := stripANSI(c.Render(60, 1))
			if got != tt.expected {
				t.Errorf("SelectedCell.Render() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSelectedCell_ZoneID(t *testing.T) {
	c := NewSelectedCell()
	// Selected cell doesn't have a zone ID (not clickable)
	if got := c.ZoneID(); got != "" {
		t.Errorf("SelectedCell.ZoneID() = %q, want %q", got, "")
	}
}

func TestSelectedCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*SelectedCell)(nil)
}
