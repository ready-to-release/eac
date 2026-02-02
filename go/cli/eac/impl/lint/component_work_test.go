//go:build L1 && ov
// +build L1,ov

package lint

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/workunit"
)

// TestCountLintComponents tests the CountLintComponents function.
func TestCountLintComponents(t *testing.T) {
	tests := []struct {
		name   string
		layers [][]workunit.UnitSpec
		want   int
	}{
		{
			name:   "nil layers returns 0",
			layers: nil,
			want:   0,
		},
		{
			name:   "empty layers returns 0",
			layers: [][]workunit.UnitSpec{},
			want:   0,
		},
		{
			name: "single layer with single component",
			layers: [][]workunit.UnitSpec{
				{
					{ID: workunit.UnitID{Module: "mod-a", Component: "go-lint"}},
				},
			},
			want: 1,
		},
		{
			name: "single layer with multiple components",
			layers: [][]workunit.UnitSpec{
				{
					{ID: workunit.UnitID{Module: "mod-a", Component: "go-lint"}},
					{ID: workunit.UnitID{Module: "mod-b", Component: "go-lint"}},
					{ID: workunit.UnitID{Module: "mod-c", Component: "eslint"}},
				},
			},
			want: 3,
		},
		{
			name: "multiple layers with multiple components",
			layers: [][]workunit.UnitSpec{
				{
					{ID: workunit.UnitID{Module: "mod-a", Component: "go-lint"}},
					{ID: workunit.UnitID{Module: "mod-b", Component: "go-lint"}},
				},
				{
					{ID: workunit.UnitID{Module: "mod-c", Component: "eslint"}},
				},
			},
			want: 3,
		},
		{
			name: "empty inner layer",
			layers: [][]workunit.UnitSpec{
				{},
				{
					{ID: workunit.UnitID{Module: "mod-a", Component: "go-lint"}},
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
