package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func computeTestOrder(t *testing.T, cfg *RepositoryConfig, ct *ComponentTypesConfig) *DisplayOrder {
	t.Helper()
	require.NoError(t, cfg.expandModuleGroups())
	cfg.computeDisplayOrder(ct)
	require.NotNil(t, cfg.DisplayOrder)
	return cfg.DisplayOrder
}

func TestDisplayOrder_BaselineModulesFirst(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "core"},
			{Moniker: "mkdocs-render-oci", ModuleGroup: "oci-tools", DependsOn: []string{"root"}},
			{Moniker: "drawio-oci", ModuleGroup: "oci-tools", DependsOn: []string{"root"}},
			{Moniker: "cli", DependsOn: []string{"core"}},
		},
	}
	order := computeTestOrder(t, cfg, nil)

	// Baseline modules should be depth -1
	assert.Equal(t, -1, order.Depth["mkdocs-render-oci"])
	assert.Equal(t, -1, order.Depth["drawio-oci"])
	assert.True(t, order.IsBaseline["mkdocs-render-oci"])
	assert.True(t, order.IsBaseline["drawio-oci"])

	// Regular modules at their computed depth
	assert.Equal(t, 0, order.Depth["core"])
	assert.Equal(t, 1, order.Depth["cli"])

	// Baseline first, then depth 0, then depth 1
	assert.Equal(t, []string{"mkdocs-render-oci", "drawio-oci", "core", "cli"}, order.Modules)
}

func TestDisplayOrder_DepthPropagation(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "contracts"},
			{Moniker: "core", DependsOn: []string{"contracts"}},
			{Moniker: "clibase", DependsOn: []string{"core"}},
			{Moniker: "cli", DependsOn: []string{"clibase", "core"}},
		},
	}
	order := computeTestOrder(t, cfg, nil)
	assert.Equal(t, 0, order.Depth["contracts"])
	assert.Equal(t, 1, order.Depth["core"])
	assert.Equal(t, 2, order.Depth["clibase"])
	assert.Equal(t, 3, order.Depth["cli"])

	assert.Equal(t, []string{"contracts", "core", "clibase", "cli"}, order.Modules)
}

func TestDisplayOrder_GroupedBeforeUngrouped(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "templates"},
			{Moniker: "contracts-core", ModuleGroup: "contracts"},
			{Moniker: "contracts-ai", ModuleGroup: "contracts"},
		},
	}
	order := computeTestOrder(t, cfg, nil)
	// All depth 0; grouped modules before ungrouped, declaration order within group
	assert.Equal(t, []string{"contracts-core", "contracts-ai", "templates"}, order.Modules)
}

func TestDisplayOrder_DeclarationOrderTiebreaker(t *testing.T) {
	// All same depth, same group (ungrouped) — should keep YAML declaration order
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "zebra"},
			{Moniker: "alpha"},
			{Moniker: "middle"},
		},
	}
	order := computeTestOrder(t, cfg, nil)

	// Declaration order preserved (not alphabetical)
	assert.Equal(t, []string{"zebra", "alpha", "middle"}, order.Modules)
}

func TestDisplayOrder_GroupDeclarationOrder(t *testing.T) {
	// Same group, same depth — should use declaration order within group
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "z-tool", ModuleGroup: "tools"},
			{Moniker: "a-tool", ModuleGroup: "tools"},
			{Moniker: "m-tool", ModuleGroup: "tools"},
		},
	}
	order := computeTestOrder(t, cfg, nil)

	assert.Equal(t, []string{"z-tool", "a-tool", "m-tool"}, order.Modules)
}

func TestDisplayOrder_ComponentOrder_DependsOn(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{
				Moniker: "docs",
				Components: ModuleComponents{
					"site":    &ComponentEntry{},
					"pdf":     &ComponentEntry{DependsOn: []string{"site"}},
					"drawio":  &ComponentEntry{},
				},
				ComponentOrder: []string{"pdf", "site", "drawio"},
			},
		},
	}
	order := computeTestOrder(t, cfg, nil)
	comps := order.Components["docs"]
	require.NotNil(t, comps)

	// pdf depends on site, so site must come before pdf.
	// drawio has no deps, so it goes by declaration order.
	// Declaration order is: pdf(0), site(1), drawio(2)
	// Round 1: ready = [site(1), drawio(2)] → site, drawio
	// Round 2: ready = [pdf(0)] → pdf
	assert.Equal(t, []string{"site", "drawio", "pdf"}, comps)
}

func TestDisplayOrder_ComponentOrder_BuildAfter(t *testing.T) {
	compTypes := &ComponentTypesConfig{
		ComponentTypes: map[string]*ComponentType{
			"docs-site": {},
			"docs-pdf":  {BuildAfter: []string{"docs-site"}},
			"go":        {},
		},
	}

	cfg := &RepositoryConfig{
		Modules: []Module{
			{
				Moniker: "docs",
				Components: ModuleComponents{
					"base-site":  &ComponentEntry{Type: "docs-site"},
					"base-pdf":   &ComponentEntry{Type: "docs-pdf"},
					"go":         &ComponentEntry{},
				},
				ComponentOrder: []string{"base-pdf", "base-site", "go"},
			},
		},
	}
	comps := computeTestOrder(t, cfg, compTypes).Components["docs"]
	// base-pdf (docs-pdf type) has build_after: [docs-site], so base-site must come first.
	// go has no deps.
	// Decl order: base-pdf(0), base-site(1), go(2).
	// Round 1: ready = [base-site(1), go(2)] → base-site, go
	// Round 2: ready = [base-pdf(0)] → base-pdf
	assert.Equal(t, []string{"base-site", "go", "base-pdf"}, comps)
}

func TestDisplayOrder_ComponentOrder_DeclarationOrderFallback(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{
				Moniker: "core",
				Components: ModuleComponents{
					"go":     &ComponentEntry{},
					"godog":  &ComponentEntry{},
					"assets": &ComponentEntry{},
				},
				ComponentOrder: []string{"go", "godog", "assets"},
			},
		},
	}
	order := computeTestOrder(t, cfg, nil)

	// No dependencies — should maintain YAML declaration order
	assert.Equal(t, []string{"go", "godog", "assets"}, order.Components["core"])
}

func TestDisplayOrder_ComponentOrder_NoComponents(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "empty"},
		},
	}
	order := computeTestOrder(t, cfg, nil)

	assert.Nil(t, order.Components["empty"])
}

func TestDisplayOrder_FullScenario(t *testing.T) {
	// Simulates real repository structure:
	// oci-tools (baseline) → contracts (depth 0) → core (depth 1) → cli (depth 2)
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "core", DependsOn: []string{"contracts"}},
			{Moniker: "cli", DependsOn: []string{"core"}},
			{Moniker: "mkdocs-render-oci", ModuleGroup: "oci-tools", DependsOn: []string{"root"}},
			{Moniker: "drawio-oci", ModuleGroup: "oci-tools", DependsOn: []string{"root"}},
			{Moniker: "contracts-core", ModuleGroup: "contracts"},
			{Moniker: "contracts-ai", ModuleGroup: "contracts"},
			{Moniker: "templates"},
		},
	}
	order := computeTestOrder(t, cfg, nil)

	// Verify ordering: baseline(-1) < contracts(0) < templates(0) < core(1) < cli(2)
	// Within depth 0: grouped (contracts) before ungrouped (templates)
	expected := []string{
		"mkdocs-render-oci", "drawio-oci",  // baseline (-1), oci-tools group, decl order
		"contracts-core", "contracts-ai",    // depth 0, contracts group, decl order
		"templates",                         // depth 0, ungrouped
		"core",                              // depth 1
		"cli",                               // depth 2
	}
	assert.Equal(t, expected, order.Modules)
}

func TestDisplayOrder_BaselineRecordedInValidateAndStripRoot(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "go-oci", DependsOn: []string{"root"}},
			{Moniker: "core"},
		},
	}

	err := cfg.validateAndStripRoot()
	require.NoError(t, err)

	// baselineModules should be populated
	assert.True(t, cfg.baselineModules["go-oci"])
	assert.False(t, cfg.baselineModules["core"])

	// DependsOn should be stripped
	assert.Empty(t, cfg.GetByMoniker("go-oci").DependsOn)
}
