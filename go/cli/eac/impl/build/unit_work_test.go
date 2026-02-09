package build

import (
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func TestCountUnits(t *testing.T) {
	tests := []struct {
		name  string
		specs []workunit.UnitSpec
		want  int
	}{
		{
			name:  "nil slice",
			specs: nil,
			want:  0,
		},
		{
			name:  "empty slice",
			specs: []workunit.UnitSpec{},
			want:  0,
		},
		{
			name:  "three specs",
			specs: make([]workunit.UnitSpec, 3),
			want:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountUnits(tt.specs)
			if got != tt.want {
				t.Errorf("CountUnits() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFilterByComponents(t *testing.T) {
	makeSpec := func(module, compName string, deps ...workunit.UnitID) workunit.UnitSpec {
		return workunit.UnitSpec{
			ID: workunit.UnitID{
				Action:        core.ActionBuild,
				Module:        module,
				ComponentType: "go",
				ComponentName: compName,
				Tool:          "go",
			},
			ComponentType: "go",
			DependsOn:     deps,
		}
	}

	t.Run("empty filter returns all", func(t *testing.T) {
		specs := []workunit.UnitSpec{makeSpec("mod", "go"), makeSpec("mod", "docs")}
		got := filterByComponents(specs, nil)
		if len(got) != 2 {
			t.Fatalf("expected 2 specs, got %d", len(got))
		}
	})

	t.Run("filters by component name", func(t *testing.T) {
		specs := []workunit.UnitSpec{
			makeSpec("mod", "go"),
			makeSpec("mod", "docs"),
			makeSpec("mod", "gherkin"),
		}
		got := filterByComponents(specs, []string{"go", "gherkin"})
		if len(got) != 2 {
			t.Fatalf("expected 2 specs, got %d", len(got))
		}
		if got[0].ID.ComponentName != "go" {
			t.Errorf("spec[0].ComponentName = %q, want %q", got[0].ID.ComponentName, "go")
		}
		if got[1].ID.ComponentName != "gherkin" {
			t.Errorf("spec[1].ComponentName = %q, want %q", got[1].ID.ComponentName, "gherkin")
		}
	})

	t.Run("re-indexes filtered specs", func(t *testing.T) {
		specs := []workunit.UnitSpec{
			makeSpec("mod", "go"),
			makeSpec("mod", "docs"),
			makeSpec("mod", "gherkin"),
		}
		specs[0].Index = 0
		specs[1].Index = 1
		specs[2].Index = 2

		got := filterByComponents(specs, []string{"go", "gherkin"})
		if got[0].Index != 0 {
			t.Errorf("spec[0].Index = %d, want 0", got[0].Index)
		}
		if got[1].Index != 1 {
			t.Errorf("spec[1].Index = %d, want 1", got[1].Index)
		}
	})

	t.Run("removes dangling DependsOn references", func(t *testing.T) {
		goSpec := makeSpec("mod", "go")
		docsID := workunit.UnitID{
			Action:        core.ActionBuild,
			Module:        "mod",
			ComponentType: "docs",
			ComponentName: "docs",
			Tool:          "mdbook",
		}
		gherkinSpec := makeSpec("mod", "gherkin", goSpec.ID, docsID)

		specs := []workunit.UnitSpec{goSpec, gherkinSpec}
		got := filterByComponents(specs, []string{"go", "gherkin"})

		if len(got) != 2 {
			t.Fatalf("expected 2 specs, got %d", len(got))
		}
		// gherkin should only have go dep left (docs was filtered out)
		if len(got[1].DependsOn) != 1 {
			t.Fatalf("expected 1 dep, got %d", len(got[1].DependsOn))
		}
		if got[1].DependsOn[0].ComponentName != "go" {
			t.Errorf("remaining dep component = %q, want %q", got[1].DependsOn[0].ComponentName, "go")
		}
	})

	t.Run("no matches returns empty", func(t *testing.T) {
		specs := []workunit.UnitSpec{makeSpec("mod", "go")}
		got := filterByComponents(specs, []string{"nonexistent"})
		if len(got) != 0 {
			t.Fatalf("expected 0 specs, got %d", len(got))
		}
	})
}
