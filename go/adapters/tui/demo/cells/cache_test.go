package cells

import "testing"

func TestCacheCell_Render(t *testing.T) {
	tests := []struct {
		name     string
		hit      bool
		reason   string
		expected string
	}{
		{
			name:     "cache hit",
			hit:      true,
			reason:   "",
			expected: "cache: hit",
		},
		{
			name:     "cache miss no reason",
			hit:      false,
			reason:   "",
			expected: "cache: miss",
		},
		{
			name:     "cache miss with reason",
			hit:      false,
			reason:   "content changed",
			expected: "cache: miss (content changed)",
		},
		{
			name:     "cache miss deps changed",
			hit:      false,
			reason:   "deps changed",
			expected: "cache: miss (deps changed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCacheCell()
			c.SetCacheInfo(tt.hit, tt.reason)
			got := stripANSI(c.Render(40, 1))
			if got != tt.expected {
				t.Errorf("CacheCell.Render() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCacheCell_ZoneID(t *testing.T) {
	c := NewCacheCell()
	if got := c.ZoneID(); got != "" {
		t.Errorf("CacheCell.ZoneID() = %q, want %q", got, "")
	}
}

func TestCacheCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*CacheCell)(nil)
}
