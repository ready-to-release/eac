package flags

import (
	"testing"
)

func TestRebuildUnconsumedArgs(t *testing.T) {
	tests := []struct {
		name       string
		original   []string
		remaining  []string
		positional []string
		want       []string
	}{
		{
			name:       "preserves order with component flag",
			original:   []string{"docs", "--no-tui", "--component", "site"},
			remaining:  []string{"--component"},
			positional: []string{"docs", "site"},
			want:       []string{"docs", "--component", "site"},
		},
		{
			name:       "preserves order with version flag",
			original:   []string{"core", "--version", "1.0.0", "--no-tui"},
			remaining:  []string{"--version"},
			positional: []string{"core", "1.0.0"},
			want:       []string{"core", "--version", "1.0.0"},
		},
		{
			name:       "no remaining flags",
			original:   []string{"docs", "--no-tui"},
			remaining:  nil,
			positional: []string{"docs"},
			want:       []string{"docs"},
		},
		{
			name:       "equals syntax preserved as single token",
			original:   []string{"docs", "--component=site", "--no-tui"},
			remaining:  []string{"--component=site"},
			positional: []string{"docs"},
			want:       []string{"docs", "--component=site"},
		},
		{
			name:       "multiple components",
			original:   []string{"docs", "--component", "site", "--component", "pdf"},
			remaining:  []string{"--component", "--component"},
			positional: []string{"docs", "site", "pdf"},
			want:       []string{"docs", "--component", "site", "--component", "pdf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RebuildUnconsumedArgs(tt.original, tt.remaining, tt.positional)
			if len(got) != len(tt.want) {
				t.Errorf("RebuildUnconsumedArgs() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("RebuildUnconsumedArgs()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
