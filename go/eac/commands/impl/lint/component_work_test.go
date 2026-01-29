//go:build L1 && ov
// +build L1,ov

package lint

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
)

// TestCountLintComponents tests the CountLintComponents function.
func TestCountLintComponents(t *testing.T) {
	tests := []struct {
		name   string
		layers [][]orchestrator.ComponentWork
		want   int
	}{
		{
			name:   "nil layers returns 0",
			layers: nil,
			want:   0,
		},
		{
			name:   "empty layers returns 0",
			layers: [][]orchestrator.ComponentWork{},
			want:   0,
		},
		{
			name: "single layer with single component",
			layers: [][]orchestrator.ComponentWork{
				{
					{Module: "mod-a", Component: "go-lint"},
				},
			},
			want: 1,
		},
		{
			name: "single layer with multiple components",
			layers: [][]orchestrator.ComponentWork{
				{
					{Module: "mod-a", Component: "go-lint"},
					{Module: "mod-b", Component: "go-lint"},
					{Module: "mod-c", Component: "eslint"},
				},
			},
			want: 3,
		},
		{
			name: "multiple layers with multiple components",
			layers: [][]orchestrator.ComponentWork{
				{
					{Module: "mod-a", Component: "go-lint"},
					{Module: "mod-b", Component: "go-lint"},
				},
				{
					{Module: "mod-c", Component: "eslint"},
				},
			},
			want: 3,
		},
		{
			name: "empty inner layer",
			layers: [][]orchestrator.ComponentWork{
				{},
				{
					{Module: "mod-a", Component: "go-lint"},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountLintComponents(tt.layers)
			if got != tt.want {
				t.Errorf("CountLintComponents() = %d, want %d", got, tt.want)
			}
		})
	}
}
