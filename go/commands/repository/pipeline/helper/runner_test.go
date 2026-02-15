package pipelinerunner

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

func TestTopologicalSort_LinearChain(t *testing.T) {
	registry := modules.NewRegistry("0.1.0", "/workspace")
	registry.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "core",
		DependsOn: nil,
	}, "/workspace"))
	registry.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "clie",
		DependsOn: []string{"core"},
	}, "/workspace"))

	sorted, err := topologicalSort(registry, []string{"clie", "core"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(sorted))
	}
	if sorted[0] != "core" || sorted[1] != "clie" {
		t.Errorf("expected [core, clie], got %v", sorted)
	}
}

func TestTopologicalSort_CircularDependency(t *testing.T) {
	registry := modules.NewRegistry("0.1.0", "/workspace")
	registry.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "module-a",
		DependsOn: []string{"module-c"},
	}, "/workspace"))
	registry.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "module-b",
		DependsOn: []string{"module-a"},
	}, "/workspace"))
	registry.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "module-c",
		DependsOn: []string{"module-b"},
	}, "/workspace"))

	_, err := topologicalSort(registry, []string{"module-a", "module-b", "module-c"})
	if err == nil {
		t.Fatal("expected circular dependency error, got nil")
	}
	if err != errCircularDependency {
		t.Errorf("expected errCircularDependency, got: %v", err)
	}
}

func TestTopologicalSort_NoDependencies(t *testing.T) {
	registry := modules.NewRegistry("0.1.0", "/workspace")
	registry.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker: "a",
	}, "/workspace"))
	registry.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker: "b",
	}, "/workspace"))

	sorted, err := topologicalSort(registry, []string{"a", "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(sorted))
	}
}

func TestTopologicalSort_DiamondDependency(t *testing.T) {
	registry := modules.NewRegistry("0.1.0", "/workspace")
	registry.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker: "base",
	}, "/workspace"))
	registry.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "left",
		DependsOn: []string{"base"},
	}, "/workspace"))
	registry.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "right",
		DependsOn: []string{"base"},
	}, "/workspace"))
	registry.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "top",
		DependsOn: []string{"left", "right"},
	}, "/workspace"))

	sorted, err := topologicalSort(registry, []string{"top", "left", "right", "base"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 4 {
		t.Fatalf("expected 4 modules, got %d", len(sorted))
	}
	// base must come before left and right, both must come before top
	indexOf := func(s string) int {
		for i, v := range sorted {
			if v == s {
				return i
			}
		}
		return -1
	}
	if indexOf("base") >= indexOf("left") {
		t.Errorf("base should come before left: %v", sorted)
	}
	if indexOf("base") >= indexOf("right") {
		t.Errorf("base should come before right: %v", sorted)
	}
	if indexOf("left") >= indexOf("top") {
		t.Errorf("left should come before top: %v", sorted)
	}
	if indexOf("right") >= indexOf("top") {
		t.Errorf("right should come before top: %v", sorted)
	}
}
