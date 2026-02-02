package testing

import (
	"testing"
)

func TestNewMockRegistry(t *testing.T) {
	registry := NewMockRegistry()

	if registry == nil {
		t.Fatal("NewMockRegistry returned nil")
	}

	if registry.Count() != 0 {
		t.Errorf("expected empty registry, got %d modules", registry.Count())
	}
}

func TestNewMockRegistry_WithModule(t *testing.T) {
	registry := NewMockRegistry(
		WithModule("test-module"),
	)

	if registry.Count() != 1 {
		t.Errorf("expected 1 module, got %d", registry.Count())
	}

	module, exists := registry.Get("test-module")
	if !exists {
		t.Fatal("module 'test-module' not found in registry")
	}

	if module.Moniker != "test-module" {
		t.Errorf("expected moniker 'test-module', got '%s'", module.Moniker)
	}
}

func TestNewMockRegistry_WithVersioning(t *testing.T) {
	registry := NewMockRegistry(
		WithModule("versioned-module", WithVersioning()),
	)

	module, exists := registry.Get("versioned-module")
	if !exists {
		t.Fatal("module not found")
	}

	if module.Versioning == nil {
		t.Error("expected Versioning to be set, got nil")
	}

	if module.Versioning.Scheme != "CalVer" {
		t.Errorf("expected default versioning scheme 'CalVer', got '%s'", module.Versioning.Scheme)
	}
}

func TestNewMockRegistry_WithChangelog(t *testing.T) {
	registry := NewMockRegistry(
		WithModule("changelog-module",
			WithVersioning(),
			WithChangelog("docs/CHANGELOG.md"),
		),
	)

	module, exists := registry.Get("changelog-module")
	if !exists {
		t.Fatal("module not found")
	}

	if module.Versioning == nil {
		t.Fatal("Versioning not set")
	}

	if module.Versioning.Changelog != "docs/CHANGELOG.md" {
		t.Errorf("expected changelog 'docs/CHANGELOG.md', got '%s'", module.Versioning.Changelog)
	}
}

func TestNewMockRegistry_MultipleModules(t *testing.T) {
	registry := NewMockRegistry(
		WithModule("module-1", WithVersioning()),
		WithModule("module-2"),
		WithModule("module-3", WithVersioning(), WithChangelog("CHANGES.md")),
	)

	if registry.Count() != 3 {
		t.Errorf("expected 3 modules, got %d", registry.Count())
	}

	// Verify module-1
	m1, exists := registry.Get("module-1")
	if !exists {
		t.Error("module-1 not found")
	}
	if m1.Versioning == nil {
		t.Error("module-1 should have versioning")
	}

	// Verify module-2
	m2, exists := registry.Get("module-2")
	if !exists {
		t.Error("module-2 not found")
	}
	if m2.Versioning != nil {
		t.Error("module-2 should not have versioning")
	}

	// Verify module-3
	m3, exists := registry.Get("module-3")
	if !exists {
		t.Error("module-3 not found")
	}
	if m3.Versioning == nil || m3.Versioning.Changelog != "CHANGES.md" {
		t.Error("module-3 should have versioning with CHANGES.md changelog")
	}
}

func TestNewMockRegistry_WithDependencies(t *testing.T) {
	registry := NewMockRegistry(
		WithModule("core"),
		WithModule("api", WithDependsOn("core")),
		WithModule("ui", WithDependsOn("core", "api")),
	)

	// Verify api depends on core
	api, _ := registry.Get("api")
	if len(api.DependsOn) != 1 || api.DependsOn[0] != "core" {
		t.Errorf("expected api to depend on ['core'], got %v", api.DependsOn)
	}

	// Verify ui depends on core and api
	ui, _ := registry.Get("ui")
	if len(ui.DependsOn) != 2 {
		t.Errorf("expected ui to have 2 dependencies, got %d", len(ui.DependsOn))
	}
}

func TestNewMockRegistry_WithComponent(t *testing.T) {
	registry := NewMockRegistry(
		WithModule("go-module",
			WithComponent("go", "go/mymodule"),
		),
	)

	module, exists := registry.Get("go-module")
	if !exists {
		t.Fatal("module not found")
	}

	if !module.HasComponent("go") {
		t.Error("module should have go component")
	}

	root := module.GetComponentRoot("go")
	if root != "go/mymodule" {
		t.Errorf("expected root 'go/mymodule', got '%s'", root)
	}
}
