package cells

import "testing"

func TestLayerCell_Render(t *testing.T) {
	tests := []struct {
		name     string
		current  int
		total    int
		expected string
	}{
		{"no layers", 0, 0, ""},
		{"first of four", 1, 4, "Layer: 1/4"},
		{"last of four", 4, 4, "Layer: 4/4"},
		{"middle layer", 2, 5, "Layer: 2/5"},
		{"single layer", 1, 1, "Layer: 1/1"},
		{"many layers", 10, 20, "Layer: 10/20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewLayerCell()
			c.SetLayer(tt.current, tt.total)
			got := stripANSI(c.Render(20, 1))
			if got != tt.expected {
				t.Errorf("LayerCell.Render() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLayerCell_ZoneID(t *testing.T) {
	c := NewLayerCell()
	if got := c.ZoneID(); got != "res-layer" {
		t.Errorf("LayerCell.ZoneID() = %q, want %q", got, "res-layer")
	}
}

func TestLayerCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*LayerCell)(nil)
}
