package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultDeriveName(t *testing.T) {
	tests := []struct {
		name     string
		rootPath string
		want     string
	}{
		{"last segment simple", "go/adapters/godog", "godog"},
		{"last segment nested", "go/commands/base", "base"},
		{"single segment", "core", "core"},
		{"strips leading dots", "specs/docs/.design", "design"},
		{"empty string", "", ""},
		{"slash only", "/", ""},
		{"replaces underscores", "go/my_adapter", "my-adapter"},
		{"truncates to 16 chars", "go/very-long-component-name-here", "very-long-compon"},
		{"version path", "contracts/core/0.1.0", "0.1.0"},
		{"trailing slash normalized", "go/core/", "core"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultDeriveName(tt.rootPath)
			if got != tt.want {
				t.Errorf("DefaultDeriveName(%q) = %q, want %q", tt.rootPath, got, tt.want)
			}
		})
	}
}

func TestModuleComponents_UnmarshalYAML_ListFormat(t *testing.T) {
	yamlInput := `
- type: go
  root: go/adapters/godog
- type: go
  root: go/adapters/ai
- type: assets
  root: go/adapters/ai
  name: assets
`
	var mc ModuleComponents
	err := yaml.Unmarshal([]byte(yamlInput), &mc)
	if err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	if _, ok := mc["godog"]; !ok {
		t.Error("expected component 'godog' (derived from go/adapters/godog)")
	}
	if _, ok := mc["ai"]; !ok {
		t.Error("expected component 'ai' (derived from go/adapters/ai)")
	}
	if _, ok := mc["assets"]; !ok {
		t.Error("expected component 'assets' (explicit name override)")
	}

	if mc["godog"].Type != "go" {
		t.Errorf("godog type = %q, want 'go'", mc["godog"].Type)
	}
	if mc["assets"].Type != "assets" {
		t.Errorf("assets type = %q, want 'assets'", mc["assets"].Type)
	}
	if mc["assets"].Name != "assets" {
		t.Errorf("assets name = %q, want 'assets'", mc["assets"].Name)
	}
}

func TestModuleComponents_UnmarshalYAML_MapFormatRejected(t *testing.T) {
	yamlInput := `
go: go/core
assets: go/core
`
	var mc ModuleComponents
	err := yaml.Unmarshal([]byte(yamlInput), &mc)
	if err == nil {
		t.Fatal("expected error for map format, got nil")
	}
}

func TestModuleComponents_UnmarshalYAML_ListFormat_TypeFallback(t *testing.T) {
	yamlInput := `
- type: container
- type: container-assets
`
	var mc ModuleComponents
	err := yaml.Unmarshal([]byte(yamlInput), &mc)
	if err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	if _, ok := mc["container"]; !ok {
		t.Error("expected component 'container' (type fallback)")
	}
	if _, ok := mc["container-assets"]; !ok {
		t.Error("expected component 'container-assets' (type fallback)")
	}
}

func TestModuleComponents_UnmarshalYAML_ListFormat_DuplicateError(t *testing.T) {
	yamlInput := `
- type: go
  root: go/foo/bar
- type: assets
  root: go/baz/bar
`
	var mc ModuleComponents
	err := yaml.Unmarshal([]byte(yamlInput), &mc)
	if err == nil {
		t.Fatal("expected error for duplicate derived name 'bar', got nil")
	}
}

func TestModuleComponents_UnmarshalYAML_ListFormat_NameOverrideResolvesDuplicate(t *testing.T) {
	yamlInput := `
- type: go
  root: go/foo/bar
- type: assets
  root: go/baz/bar
  name: baz-assets
`
	var mc ModuleComponents
	err := yaml.Unmarshal([]byte(yamlInput), &mc)
	if err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	if _, ok := mc["bar"]; !ok {
		t.Error("expected component 'bar'")
	}
	if _, ok := mc["baz-assets"]; !ok {
		t.Error("expected component 'baz-assets'")
	}
}

func TestModuleComponents_GetFirstByType(t *testing.T) {
	mc := ModuleComponents{
		"core":      &ComponentEntry{Type: "go", Root: "go/core"},
		"assets":    &ComponentEntry{Type: "assets", Root: "go/core"},
		"container": &ComponentEntry{Root: "containers/test"},
	}

	t.Run("finds by explicit type", func(t *testing.T) {
		name, entry := mc.GetFirstByType("go")
		if name != "core" || entry == nil {
			t.Errorf("GetFirstByType('go') = (%q, %v), want ('core', non-nil)", name, entry)
		}
	})

	t.Run("finds by implicit type (name=type)", func(t *testing.T) {
		name, entry := mc.GetFirstByType("container")
		if name != "container" || entry == nil {
			t.Errorf("GetFirstByType('container') = (%q, %v), want ('container', non-nil)", name, entry)
		}
	})

	t.Run("returns nil for missing type", func(t *testing.T) {
		name, entry := mc.GetFirstByType("typescript")
		if name != "" || entry != nil {
			t.Errorf("GetFirstByType('typescript') = (%q, %v), want ('', nil)", name, entry)
		}
	})

	t.Run("nil map returns nil", func(t *testing.T) {
		var nilMC ModuleComponents
		name, entry := nilMC.GetFirstByType("go")
		if name != "" || entry != nil {
			t.Errorf("nil.GetFirstByType('go') = (%q, %v), want ('', nil)", name, entry)
		}
	})
}

// --- Facet expansion tests ---

func TestExpandFacets_SpecsFacet(t *testing.T) {
	yamlInput := `
- type: go
  root: go/adapters/godog
  specs:
    - "**/*.feature"
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	// Parent component
	if _, ok := mc["godog"]; !ok {
		t.Fatal("expected parent component 'godog'")
	}

	// Synthetic specs component
	synth, ok := mc["godog~specs"]
	if !ok {
		t.Fatal("expected synthetic component 'godog~specs'")
	}
	if synth.Type != "gherkin" {
		t.Errorf("godog~specs type = %q, want 'gherkin'", synth.Type)
	}
	if synth.Root != "go/adapters/godog" {
		t.Errorf("godog~specs root = %q, want 'go/adapters/godog'", synth.Root)
	}
	if synth.ParentComponent != "godog" {
		t.Errorf("godog~specs ParentComponent = %q, want 'godog'", synth.ParentComponent)
	}
	if synth.FacetName != "specs" {
		t.Errorf("godog~specs FacetName = %q, want 'specs'", synth.FacetName)
	}
	if synth.Patterns == nil || len(synth.Patterns.Source) != 1 || synth.Patterns.Source[0] != "**/*.feature" {
		t.Errorf("godog~specs patterns = %v, want [**/*.feature]", synth.Patterns)
	}
}

func TestExpandFacets_DesignFacet(t *testing.T) {
	yamlInput := `
- type: go
  root: go/core
  design:
    - "**/*.dsl"
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	synth, ok := mc["core~design"]
	if !ok {
		t.Fatal("expected synthetic component 'core~design'")
	}
	if synth.Type != "structurizr" {
		t.Errorf("core~design type = %q, want 'structurizr'", synth.Type)
	}
	if synth.FacetName != "design" {
		t.Errorf("core~design FacetName = %q, want 'design'", synth.FacetName)
	}
}

func TestExpandFacets_DocsFacet(t *testing.T) {
	yamlInput := `
- type: go
  root: go/core
  docs:
    - "**/*.md"
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	synth, ok := mc["core~docs"]
	if !ok {
		t.Fatal("expected synthetic component 'core~docs'")
	}
	if synth.Type != "docs-assets" {
		t.Errorf("core~docs type = %q, want 'docs-assets'", synth.Type)
	}
}

func TestExpandFacets_MultipleFacets(t *testing.T) {
	yamlInput := `
- type: go
  root: go/adapters/godog
  specs:
    - "**/*.feature"
  design:
    - "**/*.dsl"
  docs:
    - "**/*.md"
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	// Should have 4 components: parent + 3 facets
	if len(mc) != 4 {
		t.Errorf("expected 4 components, got %d", len(mc))
	}
	if _, ok := mc["godog"]; !ok {
		t.Error("expected parent 'godog'")
	}
	if _, ok := mc["godog~specs"]; !ok {
		t.Error("expected 'godog~specs'")
	}
	if _, ok := mc["godog~design"]; !ok {
		t.Error("expected 'godog~design'")
	}
	if _, ok := mc["godog~docs"]; !ok {
		t.Error("expected 'godog~docs'")
	}
}

func TestExpandFacets_EmptyFacetsNoExpansion(t *testing.T) {
	yamlInput := `
- type: go
  root: go/core
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	if len(mc) != 1 {
		t.Errorf("expected 1 component (no facets), got %d", len(mc))
	}
}

func TestExpandFacets_IsFacetComponent(t *testing.T) {
	yamlInput := `
- type: go
  root: go/core
  specs:
    - "**/*.feature"
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	if mc["core"].IsFacetComponent() {
		t.Error("parent 'core' should not be a facet component")
	}
	if !mc["core~specs"].IsFacetComponent() {
		t.Error("'core~specs' should be a facet component")
	}
}

func TestExpandFacets_ConflictDetection(t *testing.T) {
	yamlInput := `
- type: go
  root: go/core
  specs:
    - "**/*.feature"
- type: gherkin
  name: core~specs
`
	var mc ModuleComponents
	err := yaml.Unmarshal([]byte(yamlInput), &mc)
	if err == nil {
		t.Fatal("expected error for facet name conflict, got nil")
	}
}

func TestExtractComponentOrder_ListFormat(t *testing.T) {
	yamlInput := `
moniker: test
name: Test Module
components:
  - type: go
    root: go/adapters/godog
  - type: go
    root: go/adapters/ai
  - type: assets
    root: go/adapters/ai
    name: assets
`
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlInput), &node); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	moduleNode := node.Content[0]
	order := extractComponentOrder(moduleNode)

	expected := []string{"godog", "ai", "assets"}
	if len(order) != len(expected) {
		t.Fatalf("extractComponentOrder returned %v, want %v", order, expected)
	}
	for i, name := range expected {
		if order[i] != name {
			t.Errorf("order[%d] = %q, want %q", i, order[i], name)
		}
	}
}

func TestExtractComponentOrder_WithFacets(t *testing.T) {
	yamlInput := `
moniker: test
name: Test Module
components:
  - type: go
    root: go/adapters/godog
    specs:
      - "**/*.feature"
  - type: go
    root: go/adapters/ai
`
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlInput), &node); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	moduleNode := node.Content[0]
	order := extractComponentOrder(moduleNode)

	expected := []string{"godog", "godog~specs", "ai"}
	if len(order) != len(expected) {
		t.Fatalf("extractComponentOrder returned %v, want %v", order, expected)
	}
	for i, name := range expected {
		if order[i] != name {
			t.Errorf("order[%d] = %q, want %q", i, order[i], name)
		}
	}
}

func TestModule_UnmarshalYAML_ListFormat(t *testing.T) {
	yamlInput := `
moniker: adapters
name: Adapters
components:
  - type: go
    root: go/adapters/godog
  - type: go
    root: go/adapters/ai
  - type: assets
    root: go/adapters/ai
    name: assets
`
	var m Module
	if err := yaml.Unmarshal([]byte(yamlInput), &m); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	if _, ok := m.Components["godog"]; !ok {
		t.Error("expected component 'godog'")
	}
	if _, ok := m.Components["ai"]; !ok {
		t.Error("expected component 'ai'")
	}
	if _, ok := m.Components["assets"]; !ok {
		t.Error("expected component 'assets'")
	}

	expectedOrder := []string{"godog", "ai", "assets"}
	if len(m.ComponentOrder) != len(expectedOrder) {
		t.Fatalf("ComponentOrder = %v, want %v", m.ComponentOrder, expectedOrder)
	}
	for i, name := range expectedOrder {
		if m.ComponentOrder[i] != name {
			t.Errorf("ComponentOrder[%d] = %q, want %q", i, m.ComponentOrder[i], name)
		}
	}
}

func TestClone_PreservesFacetFields(t *testing.T) {
	mc := ModuleComponents{
		"core": &ComponentEntry{
			Type:   "go",
			Root:   "go/core",
			Specs:  &FacetDeclaration{Patterns: []string{"**/*.feature"}},
			Design: &FacetDeclaration{Root: "specs/core/.design", Patterns: []string{"**/*.dsl"}},
			Docs:   &FacetDeclaration{Patterns: []string{"**/*.md"}},
		},
		"core~specs": &ComponentEntry{
			Type:            "gherkin",
			Root:            "go/core",
			ParentComponent: "core",
			FacetName:       "specs",
		},
	}

	clone := mc.Clone()

	if clone["core"].Specs.Patterns[0] != "**/*.feature" {
		t.Errorf("clone specs = %v, want [**/*.feature]", clone["core"].Specs.Patterns)
	}
	if clone["core"].Design.Root != "specs/core/.design" {
		t.Errorf("clone design root = %q, want 'specs/core/.design'", clone["core"].Design.Root)
	}
	if clone["core"].Design.Patterns[0] != "**/*.dsl" {
		t.Errorf("clone design = %v, want [**/*.dsl]", clone["core"].Design.Patterns)
	}
	if clone["core"].Docs.Patterns[0] != "**/*.md" {
		t.Errorf("clone docs = %v, want [**/*.md]", clone["core"].Docs.Patterns)
	}
	if clone["core~specs"].ParentComponent != "core" {
		t.Errorf("clone parent = %q, want 'core'", clone["core~specs"].ParentComponent)
	}
	if clone["core~specs"].FacetName != "specs" {
		t.Errorf("clone facet = %q, want 'specs'", clone["core~specs"].FacetName)
	}

	// Verify deep copy
	mc["core"].Specs.Patterns[0] = "modified"
	if clone["core"].Specs.Patterns[0] == "modified" {
		t.Error("clone should be deep copy, but specs was shared")
	}
	mc["core"].Design.Root = "modified"
	if clone["core"].Design.Root == "modified" {
		t.Error("clone should be deep copy, but design root was shared")
	}
}

// --- Rooted facet tests ---

func TestFacetDeclaration_UnmarshalYAML_List(t *testing.T) {
	input := `["**/*.feature", "**/*.go"]`
	var fd FacetDeclaration
	if err := yaml.Unmarshal([]byte(input), &fd); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}
	if fd.Root != "" {
		t.Errorf("Root = %q, want empty", fd.Root)
	}
	if len(fd.Patterns) != 2 || fd.Patterns[0] != "**/*.feature" {
		t.Errorf("Patterns = %v, want [**/*.feature, **/*.go]", fd.Patterns)
	}
}

func TestFacetDeclaration_UnmarshalYAML_Mapping(t *testing.T) {
	input := `
root: specs/vscode-commit/.design
patterns:
  - "workspace.dsl"
  - "**/*.dsl"
`
	var fd FacetDeclaration
	if err := yaml.Unmarshal([]byte(input), &fd); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}
	if fd.Root != "specs/vscode-commit/.design" {
		t.Errorf("Root = %q, want 'specs/vscode-commit/.design'", fd.Root)
	}
	if len(fd.Patterns) != 2 || fd.Patterns[0] != "workspace.dsl" {
		t.Errorf("Patterns = %v, want [workspace.dsl, **/*.dsl]", fd.Patterns)
	}
}

func TestFacetDeclaration_IsEmpty(t *testing.T) {
	var nilFD *FacetDeclaration
	if !nilFD.IsEmpty() {
		t.Error("nil FacetDeclaration should be empty")
	}
	if !(&FacetDeclaration{}).IsEmpty() {
		t.Error("FacetDeclaration with no patterns should be empty")
	}
	if (&FacetDeclaration{Patterns: []string{"*.dsl"}}).IsEmpty() {
		t.Error("FacetDeclaration with patterns should not be empty")
	}
}

func TestExpandFacets_RootedDesignFacet(t *testing.T) {
	yamlInput := `
- type: typescript
  root: typescript/vscode-commit
  design:
    root: specs/vscode-commit/.design
    patterns:
      - "workspace.dsl"
      - "**/*.dsl"
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	synth, ok := mc["vscode-commit~design"]
	if !ok {
		t.Fatal("expected synthetic component 'vscode-commit~design'")
	}
	if synth.Type != "structurizr" {
		t.Errorf("type = %q, want 'structurizr'", synth.Type)
	}
	// Root should be the FACET's root, not the parent's root
	if synth.Root != "specs/vscode-commit/.design" {
		t.Errorf("root = %q, want 'specs/vscode-commit/.design'", synth.Root)
	}
	if synth.ParentComponent != "vscode-commit" {
		t.Errorf("ParentComponent = %q, want 'vscode-commit'", synth.ParentComponent)
	}
	if len(synth.Patterns.Source) != 2 || synth.Patterns.Source[0] != "workspace.dsl" {
		t.Errorf("patterns = %v, want [workspace.dsl, **/*.dsl]", synth.Patterns.Source)
	}
}

func TestExpandFacets_MixedSimpleAndRooted(t *testing.T) {
	yamlInput := `
- type: typescript
  root: typescript/vscode-commit
  specs:
    - "features/**/*.steps.ts"
  design:
    root: specs/vscode-commit/.design
    patterns:
      - "**/*.dsl"
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	// specs inherits parent root
	specs := mc["vscode-commit~specs"]
	if specs == nil {
		t.Fatal("expected 'vscode-commit~specs'")
	}
	if specs.Root != "typescript/vscode-commit" {
		t.Errorf("specs root = %q, want 'typescript/vscode-commit'", specs.Root)
	}

	// design has independent root
	design := mc["vscode-commit~design"]
	if design == nil {
		t.Fatal("expected 'vscode-commit~design'")
	}
	if design.Root != "specs/vscode-commit/.design" {
		t.Errorf("design root = %q, want 'specs/vscode-commit/.design'", design.Root)
	}
}

func TestExtractComponentOrder_WithRootedFacets(t *testing.T) {
	yamlInput := `
moniker: test
name: Test Module
components:
  - type: typescript
    root: typescript/vscode-commit
    design:
      root: specs/vscode-commit/.design
      patterns:
        - "**/*.dsl"
  - type: assets
    root: typescript/vscode-commit
    name: assets
`
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlInput), &node); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	moduleNode := node.Content[0]
	order := extractComponentOrder(moduleNode)

	expected := []string{"vscode-commit", "vscode-commit~design", "assets"}
	if len(order) != len(expected) {
		t.Fatalf("extractComponentOrder returned %v, want %v", order, expected)
	}
	for i, name := range expected {
		if order[i] != name {
			t.Errorf("order[%d] = %q, want %q", i, order[i], name)
		}
	}
}

// --- Companion expansion tests ---

func TestExpandCompanions_WithField(t *testing.T) {
	yamlInput := `
- type: go
  root: go/core
  with: [assets, markdown, yaml]
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	// Should have 4 components: parent + 3 companions
	if len(mc) != 4 {
		t.Errorf("expected 4 components, got %d", len(mc))
	}
	if _, ok := mc["core"]; !ok {
		t.Error("expected parent component 'core'")
	}

	for _, compType := range []string{"assets", "markdown", "yaml"} {
		comp, ok := mc[compType]
		if !ok {
			t.Errorf("expected companion component %q", compType)
			continue
		}
		if comp.Type != compType {
			t.Errorf("%s type = %q, want %q", compType, comp.Type, compType)
		}
		if comp.Root != "go/core" {
			t.Errorf("%s root = %q, want 'go/core'", compType, comp.Root)
		}
	}
}

func TestExpandCompanions_SkipsExisting(t *testing.T) {
	yamlInput := `
- type: go
  root: go/core
  with: [assets, markdown]
- type: assets
  root: go/core/custom
  name: assets
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	// Should have 3: core, assets (explicit), markdown (companion)
	if len(mc) != 3 {
		t.Errorf("expected 3 components, got %d", len(mc))
	}

	// assets should retain its explicit root, not be overwritten
	if mc["assets"].Root != "go/core/custom" {
		t.Errorf("assets root = %q, want 'go/core/custom' (explicit should win)", mc["assets"].Root)
	}

	// markdown should be auto-created
	if _, ok := mc["markdown"]; !ok {
		t.Error("expected companion component 'markdown'")
	}
}

func TestExpandCompanions_NoWithField(t *testing.T) {
	yamlInput := `
- type: go
  root: go/core
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	if len(mc) != 1 {
		t.Errorf("expected 1 component (no companions), got %d", len(mc))
	}
}

func TestExtractComponentOrder_WithCompanions(t *testing.T) {
	yamlInput := `
moniker: test
name: Test Module
components:
  - type: go
    root: go/core
    with: [assets, markdown, yaml]
`
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlInput), &node); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	moduleNode := node.Content[0]
	order := extractComponentOrder(moduleNode)

	expected := []string{"core", "assets", "markdown", "yaml"}
	if len(order) != len(expected) {
		t.Fatalf("extractComponentOrder returned %v, want %v", order, expected)
	}
	for i, name := range expected {
		if order[i] != name {
			t.Errorf("order[%d] = %q, want %q", i, order[i], name)
		}
	}
}

func TestClone_PreservesWithField(t *testing.T) {
	mc := ModuleComponents{
		"core": &ComponentEntry{
			Type: "go",
			Root: "go/core",
			With: []string{"assets", "markdown"},
		},
	}

	clone := mc.Clone()

	if len(clone["core"].With) != 2 || clone["core"].With[0] != "assets" {
		t.Errorf("clone With = %v, want [assets, markdown]", clone["core"].With)
	}

	// Verify deep copy
	mc["core"].With[0] = "modified"
	if clone["core"].With[0] == "modified" {
		t.Error("clone should be deep copy, but With was shared")
	}
}

// --- BDD runner expansion tests ---

func TestExpandBDDRunners_TypeScriptSpecs(t *testing.T) {
	yamlInput := `
- type: typescript
  root: typescript/vscode-commit
  specs:
    root: typescript/vscode-commit/features
    patterns:
      - "**/*.feature"
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	m := &Module{
		Moniker:    "vscode-commit",
		Components: mc,
	}

	compTypes := &ComponentKindsConfig{
		Kinds: map[string]*ComponentType{
			"typescript": {BDDRunner: "cucumberjs"},
			"cucumberjs": {DefaultRoot: "features"},
		},
	}

	m.expandBDDRunners(compTypes)

	runner, ok := m.Components["cucumberjs"]
	if !ok {
		t.Fatal("expected auto-created cucumberjs component")
	}
	if runner.Type != "cucumberjs" {
		t.Errorf("runner type = %q, want 'cucumberjs'", runner.Type)
	}
	// Runner uses parent component root (step implementations live in source tree),
	// NOT specs facet root (which points to where .feature files are).
	if runner.Root != "typescript/vscode-commit" {
		t.Errorf("runner root = %q, want 'typescript/vscode-commit'", runner.Root)
	}
	if !runner.AutoBDDRunner {
		t.Error("runner should have AutoBDDRunner=true")
	}
}

func TestExpandBDDRunners_GoSpecs(t *testing.T) {
	yamlInput := `
- type: go
  root: go/core
  specs:
    - "**/*.feature"
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	m := &Module{
		Moniker:    "core",
		Components: mc,
	}

	compTypes := &ComponentKindsConfig{
		Kinds: map[string]*ComponentType{
			"go":    {BDDRunner: "godog"},
			"godog": {DefaultRoot: "specs"},
		},
	}

	m.expandBDDRunners(compTypes)

	runner, ok := m.Components["godog"]
	if !ok {
		t.Fatal("expected auto-created godog component")
	}
	if runner.Type != "godog" {
		t.Errorf("runner type = %q, want 'godog'", runner.Type)
	}
	// specs: facet has no root, so runner root = parent component root
	if runner.Root != "go/core" {
		t.Errorf("runner root = %q, want 'go/core'", runner.Root)
	}
}

func TestExpandBDDRunners_ExplicitRunnerWins(t *testing.T) {
	yamlInput := `
- type: typescript
  root: typescript/vscode-commit
  specs:
    - "**/*.feature"
- type: cucumberjs
  root: typescript/vscode-commit/custom-features
  name: cucumberjs
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	m := &Module{
		Moniker:    "vscode-commit",
		Components: mc,
	}

	compTypes := &ComponentKindsConfig{
		Kinds: map[string]*ComponentType{
			"typescript": {BDDRunner: "cucumberjs"},
			"cucumberjs": {DefaultRoot: "features"},
		},
	}

	componentCountBefore := len(m.Components)
	m.expandBDDRunners(compTypes)

	// No new components should be added since cucumberjs type already exists
	if len(m.Components) != componentCountBefore {
		t.Errorf("expected %d components (no new additions), got %d", componentCountBefore, len(m.Components))
	}

	// Explicit cucumberjs should be preserved
	runner := m.Components["cucumberjs"]
	if runner == nil {
		t.Fatal("expected cucumberjs component")
	}
	if runner.Root != "typescript/vscode-commit/custom-features" {
		t.Errorf("runner root = %q, want 'typescript/vscode-commit/custom-features' (explicit should win)", runner.Root)
	}
	if runner.AutoBDDRunner {
		t.Error("explicit runner should not have AutoBDDRunner=true")
	}
}

func TestExpandBDDRunners_NoSpecs_UsesConvention(t *testing.T) {
	yamlInput := `
- type: typescript
  root: typescript/vscode-commit
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	m := &Module{
		Moniker:    "vscode-commit",
		Components: mc,
	}

	compTypes := &ComponentKindsConfig{
		Kinds: map[string]*ComponentType{
			"typescript": {BDDRunner: "cucumberjs"},
			"cucumberjs": {DefaultRoot: "features"},
		},
	}

	m.expandBDDRunners(compTypes)

	// BDD runner auto-created using parent component root
	runner, ok := m.Components["cucumberjs"]
	if !ok {
		t.Fatal("expected auto-created cucumberjs component using convention")
	}
	if runner.Root != "typescript/vscode-commit" {
		t.Errorf("runner root = %q, want 'typescript/vscode-commit' (parent root)", runner.Root)
	}
	if !runner.AutoBDDRunner {
		t.Error("runner should have AutoBDDRunner=true")
	}
}

func TestExpandBDDRunners_NoBDDRunner(t *testing.T) {
	yamlInput := `
- type: container
  root: containers/test
  specs:
    - "**/*.feature"
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	m := &Module{
		Moniker:    "test",
		Components: mc,
	}

	compTypes := &ComponentKindsConfig{
		Kinds: map[string]*ComponentType{
			"container": {}, // No BDDRunner
		},
	}

	m.expandBDDRunners(compTypes)

	// Only original components should exist (container + container~specs synthetic)
	for name, entry := range m.Components {
		if entry != nil && entry.AutoBDDRunner {
			t.Errorf("unexpected auto-BDD runner component %q", name)
		}
	}
}

func TestExpandBDDRunners_NilCompTypes(t *testing.T) {
	m := &Module{
		Moniker: "test",
		Components: ModuleComponents{
			"go": &ComponentEntry{
				Type:  "go",
				Root:  "go/core",
				Specs: &FacetDeclaration{Patterns: []string{"**/*.feature"}},
			},
		},
	}

	// Should not panic
	m.expandBDDRunners(nil)
}

func TestExpandBDDRunners_AppendsToComponentOrder(t *testing.T) {
	yamlInput := `
- type: typescript
  root: typescript/vscode-commit
  specs:
    - "**/*.feature"
`
	var mc ModuleComponents
	if err := yaml.Unmarshal([]byte(yamlInput), &mc); err != nil {
		t.Fatalf("UnmarshalYAML failed: %v", err)
	}

	m := &Module{
		Moniker:        "vscode-commit",
		Components:     mc,
		ComponentOrder: []string{"vscode-commit", "vscode-commit~specs"},
	}

	compTypes := &ComponentKindsConfig{
		Kinds: map[string]*ComponentType{
			"typescript": {BDDRunner: "cucumberjs"},
			"cucumberjs": {DefaultRoot: "features"},
		},
	}

	m.expandBDDRunners(compTypes)

	// cucumberjs should be appended to ComponentOrder
	found := false
	for _, name := range m.ComponentOrder {
		if name == "cucumberjs" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ComponentOrder = %v, expected 'cucumberjs' to be appended", m.ComponentOrder)
	}
}

func TestExpandBDDRunners_NameCollision(t *testing.T) {
	// Simulate the adapters module: a Go component derived-named "godog"
	// from root go/adapters/godog should not be overwritten by the auto-created
	// godog BDD runner. The runner should use a suffixed name instead.
	mc := ModuleComponents{
		"godog": &ComponentEntry{
			Type: "go",
			Root: "go/adapters/godog",
		},
	}

	m := &Module{
		Moniker:    "adapters",
		Components: mc,
	}

	compTypes := &ComponentKindsConfig{
		Kinds: map[string]*ComponentType{
			"go":    {BDDRunner: "godog"},
			"godog": {DefaultRoot: "specs"},
		},
	}

	m.expandBDDRunners(compTypes)

	// Original Go component must be preserved
	original := m.Components["godog"]
	if original == nil {
		t.Fatal("original 'godog' Go component was overwritten")
	}
	if original.Type != "go" {
		t.Errorf("original component type = %q, want 'go'", original.Type)
	}
	if original.Root != "go/adapters/godog" {
		t.Errorf("original component root = %q, want 'go/adapters/godog'", original.Root)
	}

	// Auto-created runner should use suffixed name
	runner := m.Components["godog-runner"]
	if runner == nil {
		t.Fatal("expected auto-created 'godog-runner' component")
	}
	if runner.Type != "godog" {
		t.Errorf("runner type = %q, want 'godog'", runner.Type)
	}
	if !runner.AutoBDDRunner {
		t.Error("runner should have AutoBDDRunner=true")
	}
}

func TestClone_PreservesAutoBDDRunner(t *testing.T) {
	mc := ModuleComponents{
		"cucumberjs": &ComponentEntry{
			Type:          "cucumberjs",
			Root:          "features",
			AutoBDDRunner: true,
		},
	}

	clone := mc.Clone()

	if !clone["cucumberjs"].AutoBDDRunner {
		t.Error("clone should preserve AutoBDDRunner=true")
	}
}

func TestFindPrimaryComponent(t *testing.T) {
	t.Run("finds go component", func(t *testing.T) {
		mc := ModuleComponents{
			"core":   &ComponentEntry{Type: "go", Root: "go/core"},
			"assets": &ComponentEntry{Type: "assets", Root: "go/core"},
		}
		name, entry := mc.FindPrimaryComponent()
		if name != "core" || entry == nil {
			t.Errorf("FindPrimaryComponent() = (%q, %v), want ('core', non-nil)", name, entry)
		}
	})

	t.Run("finds typescript component", func(t *testing.T) {
		mc := ModuleComponents{
			"vscode-commit": &ComponentEntry{Type: "typescript", Root: "typescript/vscode-commit"},
		}
		name, entry := mc.FindPrimaryComponent()
		if name != "vscode-commit" || entry == nil {
			t.Errorf("FindPrimaryComponent() = (%q, %v), want ('vscode-commit', non-nil)", name, entry)
		}
	})

	t.Run("finds python component", func(t *testing.T) {
		mc := ModuleComponents{
			"myapp": &ComponentEntry{Type: "python", Root: "src/myapp"},
		}
		name, entry := mc.FindPrimaryComponent()
		if name != "myapp" || entry == nil {
			t.Errorf("FindPrimaryComponent() = (%q, %v), want ('myapp', non-nil)", name, entry)
		}
	})

	t.Run("finds dotnet component", func(t *testing.T) {
		mc := ModuleComponents{
			"inventory": &ComponentEntry{Type: "dotnet", Root: "src/Inventory"},
		}
		name, entry := mc.FindPrimaryComponent()
		if name != "inventory" || entry == nil {
			t.Errorf("FindPrimaryComponent() = (%q, %v), want ('inventory', non-nil)", name, entry)
		}
	})

	t.Run("finds rust component", func(t *testing.T) {
		mc := ModuleComponents{
			"mytool": &ComponentEntry{Type: "rust", Root: "rust/mytool"},
		}
		name, entry := mc.FindPrimaryComponent()
		if name != "mytool" || entry == nil {
			t.Errorf("FindPrimaryComponent() = (%q, %v), want ('mytool', non-nil)", name, entry)
		}
	})

	t.Run("returns nil for no primary", func(t *testing.T) {
		mc := ModuleComponents{
			"container": &ComponentEntry{Type: "container", Root: "containers/test"},
		}
		name, entry := mc.FindPrimaryComponent()
		if name != "" || entry != nil {
			t.Errorf("FindPrimaryComponent() = (%q, %v), want ('', nil)", name, entry)
		}
	})
}
