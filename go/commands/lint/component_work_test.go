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
		name  string
		units []workunit.UnitSpec
		want  int
	}{
		{
			name:  "nil units returns 0",
			units: nil,
			want:  0,
		},
		{
			name:  "empty units returns 0",
			units: []workunit.UnitSpec{},
			want:  0,
		},
		{
			name: "single component",
			units: []workunit.UnitSpec{
				{ID: workunit.UnitID{Module: "mod-a", ComponentType: "go-lint", ComponentName: "go-lint"}},
			},
			want: 1,
		},
		{
			name: "multiple components",
			units: []workunit.UnitSpec{
				{ID: workunit.UnitID{Module: "mod-a", ComponentType: "go-lint", ComponentName: "go-lint"}},
				{ID: workunit.UnitID{Module: "mod-b", ComponentType: "go-lint", ComponentName: "go-lint"}},
				{ID: workunit.UnitID{Module: "mod-c", ComponentType: "eslint", ComponentName: "eslint"}},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountLintComponents(tt.units)
			if got != tt.want {
				t.Errorf("CountLintComponents() = %d, want %d", got, tt.want)
			}
		})
	}
}
