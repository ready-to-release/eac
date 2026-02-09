package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandComponentGroups_BasicExpansion(t *testing.T) {
	m := &Module{
		Moniker: "my-docs",
		Components: ModuleComponents{
			"asset-blueprints": &ComponentEntry{Root: "docs/assets/blueprints", ComponentGroup: "assets"},
			"asset-images":     &ComponentEntry{Root: "docs/assets/images", ComponentGroup: "assets"},
			"site":             &ComponentEntry{Root: "docs/site", DependsOn: []string{"assets"}},
		},
	}

	m.Components.applyComponentGroupDefaults()
	err := m.expandComponentGroups()
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"asset-blueprints", "asset-images"}, m.Components["site"].DependsOn)
}

func TestExpandComponentGroups_MixedGroupAndDirect(t *testing.T) {
	m := &Module{
		Moniker: "my-docs",
		Components: ModuleComponents{
			"asset-blueprints": &ComponentEntry{Root: "docs/assets/blueprints", ComponentGroup: "assets"},
			"asset-images":     &ComponentEntry{Root: "docs/assets/images", ComponentGroup: "assets"},
			"base":             &ComponentEntry{Root: "docs/base"},
			"site":             &ComponentEntry{Root: "docs/site", DependsOn: []string{"base", "assets"}},
		},
	}

	m.Components.applyComponentGroupDefaults()
	err := m.expandComponentGroups()
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"base", "asset-blueprints", "asset-images"}, m.Components["site"].DependsOn)
}

func TestExpandComponentGroups_Deduplication(t *testing.T) {
	m := &Module{
		Moniker: "my-docs",
		Components: ModuleComponents{
			"asset-blueprints": &ComponentEntry{Root: "docs/assets/blueprints", ComponentGroup: "assets"},
			"site":             &ComponentEntry{Root: "docs/site", DependsOn: []string{"asset-blueprints", "assets"}},
		},
	}

	m.Components.applyComponentGroupDefaults()
	err := m.expandComponentGroups()
	require.NoError(t, err)

	// asset-blueprints appears both as direct dep and via group; should only appear once
	assert.Equal(t, []string{"asset-blueprints"}, m.Components["site"].DependsOn)
}

func TestExpandComponentGroups_CollisionError(t *testing.T) {
	m := &Module{
		Moniker: "my-docs",
		Components: ModuleComponents{
			"assets":       &ComponentEntry{Root: "docs/assets"},                                       // component named "assets"
			"asset-images": &ComponentEntry{Root: "docs/assets/images", ComponentGroup: "assets"}, // group named "assets"
		},
	}

	m.Components.applyComponentGroupDefaults()
	err := m.expandComponentGroups()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "assets")
	assert.Contains(t, err.Error(), "collides")
}

func TestExpandComponentGroups_SelfReferenceViaGroup(t *testing.T) {
	m := &Module{
		Moniker: "my-docs",
		Components: ModuleComponents{
			"asset-blueprints": &ComponentEntry{Root: "docs/assets/blueprints", ComponentGroup: "assets", DependsOn: []string{"assets"}},
			"asset-images":     &ComponentEntry{Root: "docs/assets/images", ComponentGroup: "assets"},
		},
	}

	m.Components.applyComponentGroupDefaults()
	err := m.expandComponentGroups()
	require.NoError(t, err)

	// Should expand to other group members, excluding self
	assert.Equal(t, []string{"asset-images"}, m.Components["asset-blueprints"].DependsOn)
}

func TestExpandComponentGroups_DefaultGroupNaming(t *testing.T) {
	m := &Module{
		Moniker: "my-docs",
		Components: ModuleComponents{
			"go":   &ComponentEntry{Root: "go/core"},
			"docs": &ComponentEntry{Root: "docs"},
		},
	}

	m.Components.applyComponentGroupDefaults()

	// Default: component_group == component name
	assert.Equal(t, "go", m.Components["go"].ComponentGroup)
	assert.Equal(t, "docs", m.Components["docs"].ComponentGroup)

	// Should not cause any errors (self-named groups are excluded from groups map)
	err := m.expandComponentGroups()
	require.NoError(t, err)
}

func TestExpandComponentGroups_NilComponents(t *testing.T) {
	m := &Module{
		Moniker:    "empty",
		Components: nil,
	}

	err := m.expandComponentGroups()
	require.NoError(t, err)
}

func TestExpandComponentGroups_EmptyDependsOn(t *testing.T) {
	m := &Module{
		Moniker: "my-docs",
		Components: ModuleComponents{
			"asset-blueprints": &ComponentEntry{Root: "docs/assets/blueprints", ComponentGroup: "assets"},
			"site":             &ComponentEntry{Root: "docs/site", DependsOn: []string{}},
		},
	}

	m.Components.applyComponentGroupDefaults()
	err := m.expandComponentGroups()
	require.NoError(t, err)

	assert.Empty(t, m.Components["site"].DependsOn)
}

func TestExpandComponentGroups_NilEntry(t *testing.T) {
	m := &Module{
		Moniker: "my-docs",
		Components: ModuleComponents{
			"ghost": nil,
			"site":  &ComponentEntry{Root: "docs/site"},
		},
	}

	m.Components.applyComponentGroupDefaults()
	err := m.expandComponentGroups()
	require.NoError(t, err)
}

func TestExpandComponentGroups_MultipleGroups(t *testing.T) {
	m := &Module{
		Moniker: "my-docs",
		Components: ModuleComponents{
			"asset-blueprints": &ComponentEntry{Root: "a", ComponentGroup: "assets"},
			"asset-images":     &ComponentEntry{Root: "b", ComponentGroup: "assets"},
			"tool-a":           &ComponentEntry{Root: "c", ComponentGroup: "tools"},
			"tool-b":           &ComponentEntry{Root: "d", ComponentGroup: "tools"},
			"site":             &ComponentEntry{Root: "e", DependsOn: []string{"assets", "tools"}},
		},
	}

	m.Components.applyComponentGroupDefaults()
	err := m.expandComponentGroups()
	require.NoError(t, err)

	assert.ElementsMatch(t,
		[]string{"asset-blueprints", "asset-images", "tool-a", "tool-b"},
		m.Components["site"].DependsOn,
	)
}

func TestExpandComponentGroups_NonexistentGroupPassthrough(t *testing.T) {
	m := &Module{
		Moniker: "my-docs",
		Components: ModuleComponents{
			"site": &ComponentEntry{Root: "docs/site", DependsOn: []string{"nonexistent-group"}},
		},
	}

	m.Components.applyComponentGroupDefaults()
	err := m.expandComponentGroups()
	require.NoError(t, err)

	// nonexistent-group is neither a component name nor a group, kept as-is
	assert.Equal(t, []string{"nonexistent-group"}, m.Components["site"].DependsOn)
}

func TestExpandComponentGroups_PreservesOrder(t *testing.T) {
	m := &Module{
		Moniker: "my-docs",
		Components: ModuleComponents{
			"a-asset": &ComponentEntry{Root: "a", ComponentGroup: "assets"},
			"b-asset": &ComponentEntry{Root: "b", ComponentGroup: "assets"},
			"c-asset": &ComponentEntry{Root: "c", ComponentGroup: "assets"},
			"site":    &ComponentEntry{Root: "d", DependsOn: []string{"assets"}},
		},
	}

	m.Components.applyComponentGroupDefaults()
	err := m.expandComponentGroups()
	require.NoError(t, err)

	// Members sorted alphabetically for deterministic order
	assert.Equal(t, []string{"a-asset", "b-asset", "c-asset"}, m.Components["site"].DependsOn)
}

func TestExpandAllComponentGroups_CrossModuleIsolation(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{
				Moniker: "mod-a",
				Components: ModuleComponents{
					"x": &ComponentEntry{Root: "a/x", ComponentGroup: "shared"},
					"y": &ComponentEntry{Root: "a/y", DependsOn: []string{"shared"}},
				},
			},
			{
				Moniker: "mod-b",
				Components: ModuleComponents{
					"p": &ComponentEntry{Root: "b/p", ComponentGroup: "shared"},
					"q": &ComponentEntry{Root: "b/q", DependsOn: []string{"shared"}},
				},
			},
		},
	}

	err := cfg.expandAllComponentGroups()
	require.NoError(t, err)

	// mod-a: "shared" expands to ["x"] only
	assert.Equal(t, []string{"x"}, cfg.Modules[0].Components["y"].DependsOn)

	// mod-b: "shared" expands to ["p"] only (not x from mod-a)
	assert.Equal(t, []string{"p"}, cfg.Modules[1].Components["q"].DependsOn)
}

func TestResolveComponentDefaults_ComponentGroupFromKind(t *testing.T) {
	// ComponentGroup from component-kind is applied when component has none
	compTypes := &ComponentKindsConfig{
		Kinds: map[string]*ComponentType{
			"docs-drawio": {
				ComponentGroup: "doc-assets",
			},
			"docs-mermaid": {
				ComponentGroup: "doc-assets",
			},
			"docs-site": {
				DependsOn: []string{"doc-assets"},
			},
		},
	}

	m := &Module{
		Moniker: "docs",
		Components: ModuleComponents{
			"drawio":  &ComponentEntry{Type: "docs-drawio", Root: "docs/assets"},
			"mermaid": &ComponentEntry{Type: "docs-mermaid", Root: "docs"},
			"site":    &ComponentEntry{Type: "docs-site", Root: "docs"},
		},
	}

	m.resolveComponentRoots(compTypes)

	assert.Equal(t, "doc-assets", m.Components["drawio"].ComponentGroup)
	assert.Equal(t, "doc-assets", m.Components["mermaid"].ComponentGroup)
	assert.Equal(t, []string{"doc-assets"}, m.Components["site"].DependsOn)
}

func TestResolveComponentDefaults_UserOverrideWins(t *testing.T) {
	// User-specified ComponentGroup and DependsOn win over component-kind defaults
	compTypes := &ComponentKindsConfig{
		Kinds: map[string]*ComponentType{
			"docs-drawio": {
				ComponentGroup: "doc-assets",
			},
			"docs-site": {
				DependsOn: []string{"doc-assets"},
			},
		},
	}

	m := &Module{
		Moniker: "docs",
		Components: ModuleComponents{
			"drawio": &ComponentEntry{Type: "docs-drawio", Root: "docs/assets", ComponentGroup: "my-custom-group"},
			"site":   &ComponentEntry{Type: "docs-site", Root: "docs", DependsOn: []string{"my-direct-dep"}},
		},
	}

	m.resolveComponentRoots(compTypes)

	// User override should be preserved
	assert.Equal(t, "my-custom-group", m.Components["drawio"].ComponentGroup)
	assert.Equal(t, []string{"my-direct-dep"}, m.Components["site"].DependsOn)
}

func TestExpandComponentGroups_FromComponentKindDefaults(t *testing.T) {
	// Simulate the full pipeline: component-kind sets defaults, then group expansion runs.
	// This is what happens after ApplyComponentDefaults + expandAllComponentGroups in LoadAll.
	m := &Module{
		Moniker: "docs",
		Components: ModuleComponents{
			"drawio":  &ComponentEntry{Type: "docs-drawio", Root: "docs/assets", ComponentGroup: "doc-assets"},
			"mermaid": &ComponentEntry{Type: "docs-mermaid", Root: "docs", ComponentGroup: "doc-assets"},
			"assets":  &ComponentEntry{Root: "docs", ComponentGroup: "doc-assets"},
			"site":    &ComponentEntry{Type: "docs-site", Root: "docs", DependsOn: []string{"doc-assets"}},
			"pdf":     &ComponentEntry{Type: "docs-pdf", DependsOn: []string{"doc-assets"}},
		},
	}

	m.Components.applyComponentGroupDefaults()
	err := m.expandComponentGroups()
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"assets", "drawio", "mermaid"}, m.Components["site"].DependsOn)
	assert.ElementsMatch(t, []string{"assets", "drawio", "mermaid"}, m.Components["pdf"].DependsOn)
}

func TestExpandAllComponentGroups_Integration(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{
				Moniker: "my-docs",
				Components: ModuleComponents{
					"asset-blueprints": &ComponentEntry{Root: "docs/assets/blueprints", ComponentGroup: "assets"},
					"asset-images":     &ComponentEntry{Root: "docs/assets/images", ComponentGroup: "assets"},
					"site":             &ComponentEntry{Root: "docs/site", DependsOn: []string{"assets"}},
					"pdf":              &ComponentEntry{Root: "docs/pdf", DependsOn: []string{"assets"}},
				},
			},
		},
	}

	err := cfg.expandAllComponentGroups()
	require.NoError(t, err)

	site := cfg.Modules[0].Components["site"]
	pdf := cfg.Modules[0].Components["pdf"]

	assert.ElementsMatch(t, []string{"asset-blueprints", "asset-images"}, site.DependsOn)
	assert.ElementsMatch(t, []string{"asset-blueprints", "asset-images"}, pdf.DependsOn)
}
