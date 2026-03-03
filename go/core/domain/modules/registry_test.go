//go:build L1 && ov
// +build L1,ov

package modules

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
)

func createTestModule(moniker, compType string) *ModuleContract {
	base := domain.BaseContract{
		Moniker: moniker,
		Description: "Test " + moniker,
		Components: config.ModuleComponents{
			compType: &config.ComponentEntry{Root: "test/" + moniker},
		},
	}
	return NewModuleContract(base, "/workspace")
}

func createTestModuleWithPackages(moniker string, components map[string]string) *ModuleContract {
	comps := make(config.ModuleComponents)
	for name, root := range components {
		comps[name] = &config.ComponentEntry{Root: root}
	}
	base := domain.BaseContract{
		Moniker:     moniker,
		Description: "Test " + moniker,
		Components:  comps,
	}
	return NewModuleContract(base, "/workspace")
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")

	if registry.Version() != "0.1.0" {
		t.Errorf("Expected version '0.1.0', got '%s'", registry.Version())
	}

	if registry.WorkspaceRoot() != "/workspace" {
		t.Errorf("Expected workspace root '/workspace', got '%s'", registry.WorkspaceRoot())
	}

	if registry.Count() != 0 {
		t.Errorf("Expected count 0, got %d", registry.Count())
	}
}

func TestRegistry_Add(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")
	module := createTestModule("test-module", "test-type")

	err := registry.Add(module)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if registry.Count() != 1 {
		t.Errorf("Expected count 1, got %d", registry.Count())
	}
}

func TestRegistry_Add_Duplicate(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")
	module := createTestModule("test-module", "test-type")

	_ = registry.Add(module)
	err := registry.Add(module)

	if err == nil {
		t.Error("Expected error when adding duplicate module")
	}
}

func TestRegistry_Add_EmptyMoniker(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")
	base := domain.BaseContract{
		Moniker: "",
		Description: "Test",
	}
	module := NewModuleContract(base, "")

	err := registry.Add(module)
	if err == nil {
		t.Error("Expected error when adding module with empty moniker")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")
	module := createTestModule("test-module", "test-type")
	_ = registry.Add(module)

	retrieved, exists := registry.Get("test-module")
	if !exists {
		t.Error("Expected module to exist")
	}

	if retrieved.GetMoniker() != "test-module" {
		t.Errorf("Expected moniker 'test-module', got '%s'", retrieved.GetMoniker())
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")

	_, exists := registry.Get("nonexistent")
	if exists {
		t.Error("Expected module to not exist")
	}
}

func TestRegistry_Has(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")
	module := createTestModule("test-module", "test-type")
	_ = registry.Add(module)

	if !registry.Has("test-module") {
		t.Error("Expected Has to return true")
	}

	if registry.Has("nonexistent") {
		t.Error("Expected Has to return false for nonexistent module")
	}
}

func TestRegistry_All(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")

	modules := []*ModuleContract{
		createTestModule("module1", "type1"),
		createTestModule("module2", "type2"),
		createTestModule("module3", "type3"),
	}

	for _, m := range modules {
		_ = registry.Add(m)
	}

	all := registry.All()
	if len(all) != 3 {
		t.Errorf("Expected 3 modules, got %d", len(all))
	}
}

func TestRegistry_AllMonikers(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")

	_ = registry.Add(createTestModule("module-c", "type"))
	_ = registry.Add(createTestModule("module-a", "type"))
	_ = registry.Add(createTestModule("module-b", "type"))

	monikers := registry.AllMonikers()

	if len(monikers) != 3 {
		t.Fatalf("Expected 3 monikers, got %d", len(monikers))
	}

	// Should be sorted alphabetically
	expected := []string{"module-a", "module-b", "module-c"}
	for i, moniker := range monikers {
		if moniker != expected[i] {
			t.Errorf("Moniker %d: expected '%s', got '%s'", i, expected[i], moniker)
		}
	}
}

func TestRegistry_FilterByComponent(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")

	// Create modules with various package combinations
	_ = registry.Add(createTestModuleWithPackages("module1", map[string]string{"go": "go/mod1", "specs": "specs/mod1"}))
	_ = registry.Add(createTestModuleWithPackages("module2", map[string]string{"typescript": "ts/mod2"}))
	_ = registry.Add(createTestModuleWithPackages("module3", map[string]string{"go": "go/mod3", "design": "specs/mod3/.design"}))
	_ = registry.Add(createTestModuleWithPackages("module4", map[string]string{"book": "docs"}))

	// Filter by 'go' component
	goModules := registry.FilterByComponent("go")
	if len(goModules) != 2 {
		t.Errorf("Expected 2 modules with 'go' component, got %d", len(goModules))
	}
	for _, m := range goModules {
		if !m.HasComponent("go") {
			t.Errorf("Module %s should have 'go' component", m.Moniker)
		}
	}

	// Filter by 'specs' component
	specsModules := registry.FilterByComponent("specs")
	if len(specsModules) != 1 {
		t.Errorf("Expected 1 module with 'specs' component, got %d", len(specsModules))
	}
}

func TestRegistry_FilterByAnyComponent(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")

	_ = registry.Add(createTestModuleWithPackages("module1", map[string]string{"go": "go/mod1"}))
	_ = registry.Add(createTestModuleWithPackages("module2", map[string]string{"typescript": "ts/mod2"}))
	_ = registry.Add(createTestModuleWithPackages("module3", map[string]string{"book": "docs"}))

	// Filter by any of 'go' or 'typescript'
	codeModules := registry.FilterByAnyComponent("go", "typescript")
	if len(codeModules) != 2 {
		t.Errorf("Expected 2 modules with 'go' or 'typescript' component, got %d", len(codeModules))
	}
}

func TestRegistry_FindByRoot(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")

	base1 := domain.BaseContract{
		Moniker: "module1",
		Components: config.ModuleComponents{
			"go": &config.ComponentEntry{Root: "src/test"},
		},
	}
	base2 := domain.BaseContract{
		Moniker: "module2",
		Components: config.ModuleComponents{
			"go": &config.ComponentEntry{Root: "src/test"},
		},
	}
	base3 := domain.BaseContract{
		Moniker: "module3",
		Components: config.ModuleComponents{
			"go": &config.ComponentEntry{Root: "src/other"},
		},
	}

	_ = registry.Add(NewModuleContract(base1, ""))
	_ = registry.Add(NewModuleContract(base2, ""))
	_ = registry.Add(NewModuleContract(base3, ""))

	modules := registry.FindByRoot("src/test")

	if len(modules) != 2 {
		t.Errorf("Expected 2 modules with root 'src/test', got %d", len(modules))
	}
}

func TestRegistry_GetDependencyGraph(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")

	base1 := domain.BaseContract{
		Moniker:   "module1",
		DependsOn: []string{"module2", "module3"},
	}
	base2 := domain.BaseContract{
		Moniker:   "module2",
		DependsOn: []string{"module3"},
	}
	base3 := domain.BaseContract{
		Moniker: "module3",
	}

	_ = registry.Add(NewModuleContract(base1, ""))
	_ = registry.Add(NewModuleContract(base2, ""))
	_ = registry.Add(NewModuleContract(base3, ""))

	graph := registry.GetDependencyGraph()

	if len(graph["module1"]) != 2 {
		t.Errorf("Expected module1 to have 2 dependencies, got %d", len(graph["module1"]))
	}

	if len(graph["module2"]) != 1 {
		t.Errorf("Expected module2 to have 1 dependency, got %d", len(graph["module2"]))
	}

	if len(graph["module3"]) != 0 {
		t.Errorf("Expected module3 to have 0 dependencies, got %d", len(graph["module3"]))
	}
}

func TestRegistry_GetReverseDependencyGraph(t *testing.T) {
	registry := NewRegistry("0.1.0", "/workspace")

	base1 := domain.BaseContract{
		Moniker:   "module1",
		DependsOn: []string{"module3"},
	}
	base2 := domain.BaseContract{
		Moniker:   "module2",
		DependsOn: []string{"module3"},
	}
	base3 := domain.BaseContract{
		Moniker: "module3",
	}

	_ = registry.Add(NewModuleContract(base1, ""))
	_ = registry.Add(NewModuleContract(base2, ""))
	_ = registry.Add(NewModuleContract(base3, ""))

	reverseGraph := registry.GetReverseDependencyGraph()

	// module3 should be depended on by module1 and module2
	if len(reverseGraph["module3"]) != 2 {
		t.Errorf("Expected module3 to be depended on by 2 modules, got %d", len(reverseGraph["module3"]))
	}

	// module1 and module2 should have no dependents
	if len(reverseGraph["module1"]) != 0 {
		t.Errorf("Expected module1 to have 0 dependents, got %d", len(reverseGraph["module1"]))
	}
}
