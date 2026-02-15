package config

import (
	"strings"
	"testing"
)

func testGoKinds(t *testing.T) map[string]*ComponentType {
	t.Helper()
	kinds := map[string]*ComponentType{
		"go": {NamePattern: "^go/(.+)$"},
	}
	if err := CompileNamePatterns(kinds); err != nil {
		t.Fatal(err)
	}
	return kinds
}

func TestExpandComponentRoots_PlainList(t *testing.T) {
	kinds := testGoKinds(t)

	mod := &Module{
		Moniker: "commands",
		ComponentRoots: map[string]*ComponentRootsEntry{
			"go": {Roots: []string{"go/commands/base", "go/commands/build"}},
		},
	}

	if err := expandComponentRoots(mod, kinds); err != nil {
		t.Fatal(err)
	}

	// Verify derived names (moniker "commands" prefix stripped)
	for _, want := range []string{"base", "build"} {
		entry, ok := mod.Components[want]
		if !ok {
			t.Errorf("expected component %q to exist", want)
			continue
		}
		if entry.Type != "go" {
			t.Errorf("component %q: type = %q, want %q", want, entry.Type, "go")
		}
	}

	// Verify roots
	if mod.Components["base"].Root != "go/commands/base" {
		t.Errorf("base root = %q, want %q", mod.Components["base"].Root, "go/commands/base")
	}
	if mod.Components["build"].Root != "go/commands/build" {
		t.Errorf("build root = %q, want %q", mod.Components["build"].Root, "go/commands/build")
	}

	// Verify ComponentRoots cleared
	if mod.ComponentRoots != nil {
		t.Error("ComponentRoots should be nil after expansion")
	}
}

func TestExpandComponentRoots_ObjectWithOverride(t *testing.T) {
	kinds := testGoKinds(t)

	mod := &Module{
		Moniker: "contracts",
		ComponentRoots: map[string]*ComponentRootsEntry{
			"go": {
				NamePattern: `^contracts/(.+?)/[\d.]+$`,
				Roots:       []string{"contracts/ai-provider/0.1.0", "contracts/tui/0.1.0"},
			},
		},
	}

	if err := expandComponentRoots(mod, kinds); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"ai-provider", "tui"} {
		if _, ok := mod.Components[want]; !ok {
			t.Errorf("expected component %q to exist", want)
		}
	}
}

func TestExpandComponentRoots_DefaultFallback(t *testing.T) {
	// No kind defined, no pattern → DefaultDeriveName
	mod := &Module{
		Moniker: "test",
		ComponentRoots: map[string]*ComponentRootsEntry{
			"unknown": {Roots: []string{"some/path/here"}},
		},
	}

	if err := expandComponentRoots(mod, nil); err != nil {
		t.Fatal(err)
	}

	want := "some-path-here"
	if _, ok := mod.Components[want]; !ok {
		t.Errorf("expected component %q to exist (DefaultDeriveName fallback)", want)
	}
}

func TestExpandComponentRoots_NameCollision(t *testing.T) {
	kinds := testGoKinds(t)

	mod := &Module{
		Moniker: "commands",
		Components: ModuleComponents{
			"base": &ComponentEntry{Type: "go", Root: "go/commands/base"},
		},
		ComponentRoots: map[string]*ComponentRootsEntry{
			"go": {Roots: []string{"go/commands/base"}},
		},
	}

	err := expandComponentRoots(mod, kinds)
	if err == nil {
		t.Fatal("expected error for name collision")
	}
	if !strings.Contains(err.Error(), "conflicts with existing component") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExpandComponentRoots_MixedWithExplicit(t *testing.T) {
	kinds := testGoKinds(t)

	mod := &Module{
		Moniker: "commands",
		Components: ModuleComponents{
			"assets": &ComponentEntry{Type: "assets", Root: "go/commands/repository"},
		},
		ComponentOrder: []string{"assets"},
		ComponentRoots: map[string]*ComponentRootsEntry{
			"go": {Roots: []string{"go/commands/base", "go/commands/build"}},
		},
	}

	if err := expandComponentRoots(mod, kinds); err != nil {
		t.Fatal(err)
	}

	// Explicit component still exists
	if _, ok := mod.Components["assets"]; !ok {
		t.Error("explicit component 'assets' should still exist")
	}

	// Expanded components exist (moniker "commands" prefix stripped)
	if _, ok := mod.Components["base"]; !ok {
		t.Error("expanded component 'base' should exist")
	}
	if _, ok := mod.Components["build"]; !ok {
		t.Error("expanded component 'build' should exist")
	}

	// Total components: 1 explicit + 2 expanded
	if len(mod.Components) != 3 {
		t.Errorf("expected 3 components, got %d", len(mod.Components))
	}

	// ComponentOrder should have explicit first, then expanded
	if len(mod.ComponentOrder) != 3 {
		t.Errorf("expected 3 in ComponentOrder, got %d", len(mod.ComponentOrder))
	}
	if mod.ComponentOrder[0] != "assets" {
		t.Errorf("ComponentOrder[0] = %q, want %q", mod.ComponentOrder[0], "assets")
	}
}

func TestExpandComponentRoots_EmptyNoop(t *testing.T) {
	mod := &Module{
		Moniker: "test",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/test"},
		},
	}
	original := len(mod.Components)

	if err := expandComponentRoots(mod, nil); err != nil {
		t.Fatal(err)
	}

	if len(mod.Components) != original {
		t.Errorf("expected %d components (no-op), got %d", original, len(mod.Components))
	}
}

func TestExpandComponentRoots_InvalidPattern(t *testing.T) {
	mod := &Module{
		Moniker: "test",
		ComponentRoots: map[string]*ComponentRootsEntry{
			"go": {
				NamePattern: "[invalid",
				Roots:       []string{"go/foo"},
			},
		},
	}

	err := expandComponentRoots(mod, nil)
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
	if want := "invalid name_pattern"; !strings.Contains(err.Error(), want) {
		t.Errorf("error %q should contain %q", err, want)
	}
}

func TestExpandComponentRoots_ComponentOrder(t *testing.T) {
	kinds := testGoKinds(t)

	mod := &Module{
		Moniker: "test",
		ComponentRoots: map[string]*ComponentRootsEntry{
			"go": {Roots: []string{"go/alpha", "go/beta", "go/gamma"}},
		},
	}

	if err := expandComponentRoots(mod, kinds); err != nil {
		t.Fatal(err)
	}

	expected := []string{"alpha", "beta", "gamma"}
	if len(mod.ComponentOrder) != len(expected) {
		t.Fatalf("ComponentOrder length = %d, want %d", len(mod.ComponentOrder), len(expected))
	}
	for i, want := range expected {
		if mod.ComponentOrder[i] != want {
			t.Errorf("ComponentOrder[%d] = %q, want %q", i, mod.ComponentOrder[i], want)
		}
	}
}
