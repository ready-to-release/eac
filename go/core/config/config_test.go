//go:build L0 && ov

package config

import (
	"errors"
	"testing"

	"github.com/ready-to-release/eac/go/core/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultOptions(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.NotEmpty(t, cfg.RepoRoot)
	assert.NotEmpty(t, cfg.ConfigRoot)
	assert.NotNil(t, cfg.Repository)
	assert.NotEmpty(t, cfg.Repository.Modules)
	assert.NotNil(t, cfg.Environments)
	assert.NotNil(t, cfg.TestingTags)
	assert.NotNil(t, cfg.TestSuites)
}

func TestLoad_WithExplicitRepoRoot(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	opts := LoadOptions{
		RepoRoot:        repoRoot,
		ValidateSchemas: true,
	}

	cfg, err := Load(opts)
	require.NoError(t, err)
	assert.Equal(t, repoRoot, cfg.RepoRoot)
}

func TestLoad_LazyLoad(t *testing.T) {
	opts := LoadOptions{
		ValidateSchemas: true,
		LazyLoad:        true,
	}

	cfg, err := Load(opts)
	require.NoError(t, err)

	// Configs should not be loaded yet
	assert.Nil(t, cfg.Repository)
	assert.Nil(t, cfg.Environments)
	assert.Nil(t, cfg.TestingTags)
	assert.Nil(t, cfg.TestSuites)

	// Load on demand
	err = cfg.LoadRepository(true)
	require.NoError(t, err)
	assert.NotNil(t, cfg.Repository)
}

func TestLoad_WithoutSchemaValidation(t *testing.T) {
	opts := LoadOptions{
		ValidateSchemas: false,
	}

	cfg, err := Load(opts)
	require.NoError(t, err)
	assert.NotNil(t, cfg.Repository)
	assert.NotEmpty(t, cfg.Repository.Modules)
}

func TestRepositoryConfig_ModuleDefaults(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	// Check that defaults are applied
	for _, m := range cfg.Repository.Modules {
		assert.NotEmpty(t, m.Components, "module %s should have components", m.Moniker)
		assert.NotEmpty(t, m.Description, "module %s should have description", m.Moniker)
		assert.NotNil(t, m.DependsOn, "module %s should have depends_on", m.Moniker)
	}
}

func TestRepositoryConfig_GetModule(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	m, ok := cfg.Repository.GetModule("core")
	assert.True(t, ok)
	assert.Equal(t, "core", m.Moniker)

	_, ok = cfg.Repository.GetModule("nonexistent")
	assert.False(t, ok)
}

func TestEnvironmentsConfig_GetEnvironment(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	env, ok := cfg.Environments.GetEnvironment("local01")
	assert.True(t, ok)
	assert.Equal(t, "L2", env.Level)

	_, ok = cfg.Environments.GetEnvironment("nonexistent")
	assert.False(t, ok)
}

func TestEnvironmentsConfig_GetEnvironmentsByLevel(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	l0Envs := cfg.Environments.GetEnvironmentsByLevel("L0")
	assert.NotEmpty(t, l0Envs)
	for _, env := range l0Envs {
		assert.Equal(t, "L0", env.Level)
	}
}

func TestEnvironmentsConfig_Validate(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	err = cfg.Environments.Validate()
	assert.NoError(t, err)
}

func TestTestingTagsConfig_Initialize(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	// Should be initialized
	assert.NotNil(t, cfg.TestingTags.compiledPatterns)
	assert.NotNil(t, cfg.TestingTags.tagLookup)
	assert.NotNil(t, cfg.TestingTags.tagsByType)
}

func TestTestingTagsConfig_GetTag(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	// Exact match
	tag, ok := cfg.TestingTags.GetTag("@L0")
	assert.True(t, ok)
	assert.Equal(t, "taxonomy-level", tag.Type)

	// Pattern match
	tag, ok = cfg.TestingTags.GetTag("@skip:wip")
	assert.True(t, ok)
	assert.Equal(t, "execution_control", tag.Type)
}

func TestTestingTagsConfig_IsKnownTag(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	assert.True(t, cfg.TestingTags.IsKnownTag("@L0"))
	assert.True(t, cfg.TestingTags.IsKnownTag("@skip:wip"))
	assert.True(t, cfg.TestingTags.IsKnownTag("@deps:docker"))
	assert.False(t, cfg.TestingTags.IsKnownTag("@unknown"))
}

func TestTestingTagsConfig_GetTaxonomyLevelTags(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	tags := cfg.TestingTags.GetTaxonomyLevelTags()
	assert.Contains(t, tags, "@L0")
	assert.Contains(t, tags, "@L1")
	assert.Contains(t, tags, "@L2")
	assert.Contains(t, tags, "@L3")
	assert.Contains(t, tags, "@L4")
	assert.Contains(t, tags, "@HE2E")
}

func TestTestingTagsConfig_GetValidSkipReasons(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	reasons := cfg.TestingTags.GetValidSkipReasons()
	assert.Contains(t, reasons, "wip")
	assert.Contains(t, reasons, "broken")
	assert.Contains(t, reasons, "flaky")
}

func TestTestingTagsConfig_GetSkipTags(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	tags := cfg.TestingTags.GetSkipTags()
	assert.Contains(t, tags, "@skip:wip")
	assert.True(t, len(tags) > 1, "should have multiple skip tags")
}

func TestValidateAll(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	err = cfg.ValidateAll()
	assert.NoError(t, err)
}

// TestPeekRepositoryType tests the type detection function
func TestPeekRepositoryType(t *testing.T) {
	t.Run("detects mono type", func(t *testing.T) {
		// Use the real config root which should have repository.yml
		cfg, err := Load(LoadOptions{ValidateSchemas: false, LazyLoad: true})
		require.NoError(t, err)

		repoType, err := peekRepositoryType(cfg.ConfigRoot)
		require.NoError(t, err)
		// This repo is configured as mono
		assert.Equal(t, "mono", repoType)
	})

	t.Run("returns empty for nonexistent directory", func(t *testing.T) {
		repoType, err := peekRepositoryType("/nonexistent/path")
		require.NoError(t, err) // Should not error, just return empty
		assert.Empty(t, repoType)
	})
}

// TestLoadRepositoryTypeDefaults tests loading type-specific defaults
func TestLoadRepositoryTypeDefaults(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	t.Run("loads mono defaults", func(t *testing.T) {
		cfg, err := LoadRepositoryTypeDefaults(repoRoot, "mono")
		require.NoError(t, err)
		require.NotNil(t, cfg)
		// Mono should have max_branch_age_days = 7
		assert.Equal(t, 7, cfg.Repository.MaxBranchAgeDays)
	})

	t.Run("loads poly defaults", func(t *testing.T) {
		cfg, err := LoadRepositoryTypeDefaults(repoRoot, "poly")
		require.NoError(t, err)
		require.NotNil(t, cfg)
		// Poly should have unrestricted versioning
		assert.Equal(t, "unrestricted", cfg.Repository.Versioning.Constraint)
	})

	t.Run("loads adjacent defaults", func(t *testing.T) {
		cfg, err := LoadRepositoryTypeDefaults(repoRoot, "adjacent")
		require.NoError(t, err)
		require.NotNil(t, cfg)
		// Adjacent should have merge strategy
		assert.Equal(t, "merge", cfg.Repository.PR.MergeStrategy)
	})

	t.Run("returns ErrNoDefaults for unknown type", func(t *testing.T) {
		cfg, err := LoadRepositoryTypeDefaults(repoRoot, "unknown")
		require.True(t, errors.Is(err, ErrNoDefaults))
		assert.Nil(t, cfg)
	})

	t.Run("returns ErrNoDefaults for empty type", func(t *testing.T) {
		cfg, err := LoadRepositoryTypeDefaults(repoRoot, "")
		require.True(t, errors.Is(err, ErrNoDefaults))
		assert.Nil(t, cfg)
	})
}

// TestLoadRepository_TypeSpecificMerge tests the type-specific merge order
func TestLoadRepository_TypeSpecificMerge(t *testing.T) {
	t.Run("applies type-specific defaults", func(t *testing.T) {
		// Load config - this repo is type=mono
		cfg, err := Load(DefaultLoadOptions())
		require.NoError(t, err)

		// Mono type defaults should be applied (max_branch_age_days=7)
		// unless overridden in user config
		assert.NotZero(t, cfg.Repository.Repository.MaxBranchAgeDays)
	})

	t.Run("user config overrides type defaults", func(t *testing.T) {
		// Load config - user config values should win
		cfg, err := Load(DefaultLoadOptions())
		require.NoError(t, err)

		// Repository type should be mono (from user config)
		assert.Equal(t, "mono", cfg.Repository.Repository.Type)
	})
}

func TestModule_ShouldAggregateFromDependencies(t *testing.T) {
	tests := []struct {
		name     string
		module   Module
		expected bool
	}{
		{
			name:     "no dependencies",
			module:   Module{Moniker: "test"},
			expected: false,
		},
		{
			name: "library with dependencies - should NOT aggregate",
			module: Module{
				Moniker:   "my-lib",
				DependsOn: []string{"other-lib"},
			},
			expected: false,
		},
		{
			name: "container module - should aggregate",
			module: Module{
				Moniker:   "my-container",
				DependsOn: []string{"my-lib"},
				Components: ModuleComponents{
					"container": &ComponentEntry{Type: "container"},
				},
			},
			expected: true,
		},
		{
			name: "container module (dockerfile type) - should aggregate",
			module: Module{
				Moniker:   "eac-ext",
				DependsOn: []string{"eac-cli"},
				Components: ModuleComponents{
					"container": &ComponentEntry{Type: "container"},
				},
			},
			expected: true,
		},
		{
			name: "bundle release type - should aggregate",
			module: Module{
				Moniker:   "clie-eac-bundle",
				DependsOn: []string{"clie", "eac-ext"},
				Versioning: &ModuleVersioning{
					ReleaseType: "bundle",
				},
			},
			expected: true,
		},
		{
			name: "published release type with deps - should NOT aggregate",
			module: Module{
				Moniker:   "my-app",
				DependsOn: []string{"my-lib"},
				Versioning: &ModuleVersioning{
					ReleaseType: "published",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.module.ShouldAggregateFromDependencies()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBlueprintsPopulatesComponentKinds verifies the loading pipeline invariant:
// LoadRepository (via LoadBlueprints) must populate ComponentKinds before
// EnsureComponentKinds is called. This guards against regressions where the
// load ordering might change.
func TestBlueprintsPopulatesComponentKinds(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	// Verify LoadBlueprints populated ComponentKinds
	require.NotNil(t, cfg.ComponentKinds)
	require.NotEmpty(t, cfg.ComponentKinds.Kinds, "ComponentKinds should be populated by LoadBlueprints")

	// Verify well-known kinds exist
	assert.NotNil(t, cfg.ComponentKinds.Get("go"), "go component kind should exist")
	assert.NotNil(t, cfg.ComponentKinds.Get("typescript"), "typescript component kind should exist")
	assert.NotNil(t, cfg.ComponentKinds.Get("container"), "container component kind should exist")
}

