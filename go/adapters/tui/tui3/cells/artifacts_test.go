package cells

import "testing"

func TestArtifactsCell_Render(t *testing.T) {
	tests := []struct {
		name      string
		artifacts []ArtifactInfo
		expected  string
	}{
		{
			name:      "no artifacts",
			artifacts: nil,
			expected:  "artifacts: none",
		},
		{
			name: "single artifact",
			artifacts: []ArtifactInfo{
				{Name: "eac.exe", Size: 12400000},
			},
			expected: "artifacts: eac.exe (12.4MB)",
		},
		{
			name: "small artifact",
			artifacts: []ArtifactInfo{
				{Name: "config.json", Size: 1500},
			},
			expected: "artifacts: config.json (1.5KB)",
		},
		{
			name: "multiple artifacts",
			artifacts: []ArtifactInfo{
				{Name: "eac.exe", Size: 12400000},
				{Name: "eac.dll", Size: 5000000},
			},
			expected: "artifacts: eac.exe (12.4MB), eac.dll (5.0MB)",
		},
		{
			name: "bytes size",
			artifacts: []ArtifactInfo{
				{Name: "tiny.txt", Size: 500},
			},
			expected: "artifacts: tiny.txt (500B)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewArtifactsCell()
			c.SetArtifacts(tt.artifacts)
			got := stripANSI(c.Render(80, 1))
			if got != tt.expected {
				t.Errorf("ArtifactsCell.Render() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArtifactsCell_ZoneID(t *testing.T) {
	c := NewArtifactsCell()
	if got := c.ZoneID(); got != "" {
		t.Errorf("ArtifactsCell.ZoneID() = %q, want %q", got, "")
	}
}

func TestArtifactsCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*ArtifactsCell)(nil)
}
