package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandModuleGroups_BasicExpansion(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "mkdocs-render-oci", ModuleGroup: "oci-tools"},
			{Moniker: "drawio-oci", ModuleGroup: "oci-tools"},
			{Moniker: "mermaid-oci", ModuleGroup: "oci-tools"},
			{Moniker: "docs", DependsOn: []string{"oci-tools"}},
		},
	}

	err := cfg.expandModuleGroups()
	require.NoError(t, err)

	docs := cfg.GetByMoniker("docs")
	require.NotNil(t, docs)

	assert.ElementsMatch(t, []string{"mkdocs-render-oci", "drawio-oci", "mermaid-oci"}, docs.DependsOn)
}

func TestExpandModuleGroups_MixedGroupAndMoniker(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "pdf-oci", ModuleGroup: "oci-tools"},
			{Moniker: "drawio-oci", ModuleGroup: "oci-tools"},
			{Moniker: "core"},
			{Moniker: "docs", DependsOn: []string{"core", "oci-tools"}},
		},
	}

	err := cfg.expandModuleGroups()
	require.NoError(t, err)

	docs := cfg.GetByMoniker("docs")
	require.NotNil(t, docs)

	assert.ElementsMatch(t, []string{"core", "pdf-oci", "drawio-oci"}, docs.DependsOn)
}

func TestExpandModuleGroups_MultipleGroups(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "contract-a", ModuleGroup: "contracts"},
			{Moniker: "contract-b", ModuleGroup: "contracts"},
			{Moniker: "adapter-x", ModuleGroup: "adapters"},
			{Moniker: "adapter-y", ModuleGroup: "adapters"},
			{Moniker: "cli", DependsOn: []string{"contracts", "adapters"}},
		},
	}

	err := cfg.expandModuleGroups()
	require.NoError(t, err)

	cli := cfg.GetByMoniker("cli")
	require.NotNil(t, cli)

	assert.ElementsMatch(t, []string{"contract-a", "contract-b", "adapter-x", "adapter-y"}, cli.DependsOn)
}

func TestExpandModuleGroups_Deduplication(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "pdf-oci", ModuleGroup: "oci-tools"},
			{Moniker: "docs", DependsOn: []string{"pdf-oci", "oci-tools"}},
		},
	}

	err := cfg.expandModuleGroups()
	require.NoError(t, err)

	docs := cfg.GetByMoniker("docs")
	require.NotNil(t, docs)

	// pdf-oci appears both as direct dep and via group; should only appear once
	assert.Equal(t, []string{"pdf-oci"}, docs.DependsOn)
}

func TestExpandModuleGroups_GroupNameCollidesWithMoniker(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "oci-tools", ModuleGroup: ""},        // Module with this moniker
			{Moniker: "pdf-oci", ModuleGroup: "oci-tools"}, // Group with same name
		},
	}

	err := cfg.expandModuleGroups()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "oci-tools")
	assert.Contains(t, err.Error(), "collides")
}

func TestExpandModuleGroups_EmptyDependsOn(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "pdf-oci", ModuleGroup: "oci-tools"},
			{Moniker: "core", DependsOn: []string{}},
		},
	}

	err := cfg.expandModuleGroups()
	require.NoError(t, err)

	core := cfg.GetByMoniker("core")
	require.NotNil(t, core)
	assert.Empty(t, core.DependsOn)
}

func TestExpandModuleGroups_NilDependsOn(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "pdf-oci", ModuleGroup: "oci-tools"},
			{Moniker: "core"},
		},
	}

	err := cfg.expandModuleGroups()
	require.NoError(t, err)

	core := cfg.GetByMoniker("core")
	require.NotNil(t, core)
	assert.Nil(t, core.DependsOn)
}

func TestExpandModuleGroups_NoGroups(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "core"},
			{Moniker: "cli", DependsOn: []string{"core"}},
		},
	}

	err := cfg.expandModuleGroups()
	require.NoError(t, err)

	cli := cfg.GetByMoniker("cli")
	require.NotNil(t, cli)
	assert.Equal(t, []string{"core"}, cli.DependsOn)
}

func TestExpandModuleGroups_EmptyGroup(t *testing.T) {
	// A group name referenced in depends_on but no modules have that group
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "core"},
			{Moniker: "cli", DependsOn: []string{"core", "nonexistent-group"}},
		},
	}

	err := cfg.expandModuleGroups()
	require.NoError(t, err)

	cli := cfg.GetByMoniker("cli")
	require.NotNil(t, cli)

	// nonexistent-group is neither a moniker nor a group, kept as-is (passthrough)
	assert.ElementsMatch(t, []string{"core", "nonexistent-group"}, cli.DependsOn)
}

func TestExpandModuleGroups_SelfReferenceViaGroup(t *testing.T) {
	// A module depends on its own group - should not include itself
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "pdf-oci", ModuleGroup: "oci-tools", DependsOn: []string{"oci-tools"}},
			{Moniker: "drawio-oci", ModuleGroup: "oci-tools"},
		},
	}

	err := cfg.expandModuleGroups()
	require.NoError(t, err)

	pdfOCI := cfg.GetByMoniker("pdf-oci")
	require.NotNil(t, pdfOCI)

	// Should expand to other group members, excluding self
	assert.Equal(t, []string{"drawio-oci"}, pdfOCI.DependsOn)
}

func TestExpandModuleGroups_PreservesOrder(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "a-oci", ModuleGroup: "oci-tools"},
			{Moniker: "b-oci", ModuleGroup: "oci-tools"},
			{Moniker: "c-oci", ModuleGroup: "oci-tools"},
			{Moniker: "docs", DependsOn: []string{"oci-tools"}},
		},
	}

	err := cfg.expandModuleGroups()
	require.NoError(t, err)

	docs := cfg.GetByMoniker("docs")
	require.NotNil(t, docs)

	// Should preserve the order modules appear in the config
	assert.Equal(t, []string{"a-oci", "b-oci", "c-oci"}, docs.DependsOn)
}
