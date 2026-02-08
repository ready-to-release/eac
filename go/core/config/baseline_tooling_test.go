package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAndStripRoot_RootAlone_Stripped(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "mkdocs-render-oci", DependsOn: []string{"root"}},
			{Moniker: "core"},
		},
	}

	err := cfg.validateAndStripRoot()
	require.NoError(t, err)

	m := cfg.GetByMoniker("mkdocs-render-oci")
	require.NotNil(t, m)
	assert.Empty(t, m.DependsOn, "root should be stripped, leaving empty depends_on")
}

func TestValidateAndStripRoot_RootWithOtherDeps_Fails(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "bad-module", DependsOn: []string{"root", "core"}},
			{Moniker: "core"},
		},
	}

	err := cfg.validateAndStripRoot()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad-module")
	assert.Contains(t, err.Error(), "root")
	assert.Contains(t, err.Error(), "only entry")
}

func TestValidateAndStripRoot_NoRoot_Unchanged(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "core"},
			{Moniker: "cli", DependsOn: []string{"core"}},
		},
	}

	err := cfg.validateAndStripRoot()
	require.NoError(t, err)

	cli := cfg.GetByMoniker("cli")
	require.NotNil(t, cli)
	assert.Equal(t, []string{"core"}, cli.DependsOn)
}

func TestValidateAndStripRoot_EmptyDeps_Unchanged(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "core", DependsOn: []string{}},
		},
	}

	err := cfg.validateAndStripRoot()
	require.NoError(t, err)
	assert.Empty(t, cfg.GetByMoniker("core").DependsOn)
}

func TestValidateAndStripRoot_NilDeps_Unchanged(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "core"},
		},
	}

	err := cfg.validateAndStripRoot()
	require.NoError(t, err)
	assert.Nil(t, cfg.GetByMoniker("core").DependsOn)
}

func TestValidateAndStripRoot_RootAsMoniker_Fails(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "root"},
		},
	}

	err := cfg.validateAndStripRoot()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reserved")
}

func TestValidateAndStripRoot_MultipleRootModules(t *testing.T) {
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "mkdocs-render-oci", DependsOn: []string{"root"}},
			{Moniker: "drawio-oci", DependsOn: []string{"root"}},
			{Moniker: "core"},
		},
	}

	err := cfg.validateAndStripRoot()
	require.NoError(t, err)

	assert.Empty(t, cfg.GetByMoniker("mkdocs-render-oci").DependsOn)
	assert.Empty(t, cfg.GetByMoniker("drawio-oci").DependsOn)
}

func TestExpandModuleGroups_WithRootSentinel(t *testing.T) {
	// Full integration: root modules + group expansion
	cfg := &RepositoryConfig{
		Modules: []Module{
			{Moniker: "mkdocs-render-oci", ModuleGroup: "oci-tools", DependsOn: []string{"root"}},
			{Moniker: "drawio-oci", ModuleGroup: "oci-tools", DependsOn: []string{"root"}},
			{Moniker: "docs", DependsOn: []string{"oci-tools"}},
		},
	}

	err := cfg.expandModuleGroups()
	require.NoError(t, err)

	// Root modules have no deps after stripping
	assert.Empty(t, cfg.GetByMoniker("mkdocs-render-oci").DependsOn)
	assert.Empty(t, cfg.GetByMoniker("drawio-oci").DependsOn)

	// Consumer still gets expanded group deps
	docs := cfg.GetByMoniker("docs")
	assert.ElementsMatch(t, []string{"mkdocs-render-oci", "drawio-oci"}, docs.DependsOn)
}
