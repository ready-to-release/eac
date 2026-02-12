package lint

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/workunit"
)

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
				{ID: workunit.UnitID{Module: "mod-a", ComponentType: "go", ComponentName: "go", Tool: "golangci-lint"}},
			},
			want: 1,
		},
		{
			name: "multiple components across modules",
			units: []workunit.UnitSpec{
				{ID: workunit.UnitID{Module: "mod-a", ComponentType: "go", ComponentName: "go", Tool: "golangci-lint"}},
				{ID: workunit.UnitID{Module: "mod-b", ComponentType: "go", ComponentName: "go", Tool: "golangci-lint"}},
				{ID: workunit.UnitID{Module: "mod-c", ComponentType: "docs", ComponentName: "docs", Tool: "markdownlint"}},
			},
			want: 3,
		},
		{
			name: "same module multiple components",
			units: []workunit.UnitSpec{
				{ID: workunit.UnitID{Module: "mod-a", ComponentType: "go", ComponentName: "go", Tool: "golangci-lint"}},
				{ID: workunit.UnitID{Module: "mod-a", ComponentType: "go", ComponentName: "go", Tool: "staticcheck"}},
				{ID: workunit.UnitID{Module: "mod-a", ComponentType: "docs", ComponentName: "docs", Tool: "markdownlint"}},
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
