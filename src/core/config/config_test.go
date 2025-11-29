//go:build L0 && ov

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultOptions(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.NotEmpty(t, cfg.RepoRoot)
	assert.NotEmpty(t, cfg.ConfigRoot)
	assert.NotNil(t, cfg.Modules)
	assert.NotNil(t, cfg.Environments)
	assert.NotNil(t, cfg.TestingTags)
	assert.NotNil(t, cfg.TestSuites)
}

func TestLoad_WithExplicitRepoRoot(t *testing.T) {
	// Use findRepositoryRoot (the local implementation)
	repoRoot, err := findRepositoryRoot("")
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
	assert.Nil(t, cfg.Modules)
	assert.Nil(t, cfg.Environments)
	assert.Nil(t, cfg.TestingTags)
	assert.Nil(t, cfg.TestSuites)

	// Load on demand
	err = cfg.LoadModules(true)
	require.NoError(t, err)
	assert.NotNil(t, cfg.Modules)
}

func TestLoad_WithoutSchemaValidation(t *testing.T) {
	opts := LoadOptions{
		ValidateSchemas: false,
	}

	cfg, err := Load(opts)
	require.NoError(t, err)
	assert.NotNil(t, cfg.Modules)
}

func TestModulesConfig_Defaults(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	// Check that defaults are applied
	for _, m := range cfg.Modules.Modules {
		assert.NotEmpty(t, m.Type, "module %s should have type", m.Moniker)
		assert.NotEmpty(t, m.Parent, "module %s should have parent", m.Moniker)
		assert.NotEmpty(t, m.Description, "module %s should have description", m.Moniker)
		assert.NotNil(t, m.DependsOn, "module %s should have depends_on", m.Moniker)
	}
}

func TestModulesConfig_GetModule(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	m, ok := cfg.Modules.GetModule("src-core")
	assert.True(t, ok)
	assert.Equal(t, "src-core", m.Moniker)

	_, ok = cfg.Modules.GetModule("nonexistent")
	assert.False(t, ok)
}

func TestModulesConfig_GetCatchAllModule(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	m, ok := cfg.Modules.GetCatchAllModule()
	assert.True(t, ok)
	assert.Equal(t, "unordered", m.Moniker)
	assert.True(t, m.Flags.CatchAll)
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

func TestTestingTagsConfig_BuildGodogSkipTagFilter(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	filter := cfg.TestingTags.BuildGodogSkipTagFilter()
	assert.Contains(t, filter, "~@skip:wip")
	assert.Contains(t, filter, "&&")
}

func TestValidateAll(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	err = cfg.ValidateAll()
	assert.NoError(t, err)
}

// TestLoad_ModuleTypesLoaded verifies ModuleTypes is populated after Load
func TestLoad_ModuleTypesLoaded(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	// ModuleTypes should be loaded
	assert.NotNil(t, cfg.ModuleTypes, "ModuleTypes should be loaded")
	assert.NotEmpty(t, cfg.ModuleTypes.Types, "ModuleTypes should have types")

	// Type lookup should work
	goLib := cfg.ModuleTypes.Get("go-library")
	assert.NotNil(t, goLib, "should find go-library type")
	assert.Equal(t, "go", goLib.BuildSystem)
}

// TestLoad_TypeDefaultsApplied verifies type defaults are applied after Load
func TestLoad_TypeDefaultsApplied(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	// Find a go-library module (src-core)
	srcCore, ok := cfg.Modules.GetModule("src-core")
	require.True(t, ok, "src-core module should exist")
	assert.Equal(t, "go-library", srcCore.Type)

	// Go-library type defaults should be applied
	goLibType := cfg.ModuleTypes.Get("go-library")
	require.NotNil(t, goLibType)

	if goLibType.Defaults != nil && goLibType.Defaults.Files != nil {
		// If type has source defaults, they should be applied (unless explicit in modules.yml)
		if goLibType.Defaults.Files.Source != nil {
			// src-core has explicit source in modules.yml, so check that format
			assert.NotEmpty(t, srcCore.Files.Source)
		}
	}
}

// TestModulesConfig_ApplyTypeDefaults tests direct ApplyTypeDefaults call
func TestModulesConfig_ApplyTypeDefaults(t *testing.T) {
	t.Run("applies type defaults", func(t *testing.T) {
		modules := &ModulesConfig{
			Modules: []Module{
				{
					Moniker: "test-mod",
					Name:    "Test Module",
					Type:    "go-library",
					Files: Files{
						Root: "src/test",
					},
				},
			},
		}

		types := &ModuleTypesConfig{
			Types: []ModuleTypeDef{
				{
					Name:        "go-library",
					BuildSystem: "go",
					Defaults: &TypeDefaults{
						Files: &FilesDefaults{
							Source: []string{"**/*.go"},
							Config: []string{"go.mod"},
						},
						Repo: &RepoDefaults{
							Specs:    []string{"specs/{moniker}/**"},
							TestImpl: "{root}/tests",
						},
					},
				},
			},
		}

		modules.ApplyTypeDefaults(types)

		m := modules.Modules[0]
		assert.Equal(t, []string{"**/*.go"}, m.Files.Source)
		assert.Equal(t, []string{"go.mod"}, m.Files.Config)
		assert.Equal(t, []string{"specs/test-mod/**"}, m.Files.Repo.Specs)
		assert.Equal(t, "src/test/tests", m.Files.Repo.TestImpl)
	})

	t.Run("preserves explicit values", func(t *testing.T) {
		modules := &ModulesConfig{
			Modules: []Module{
				{
					Moniker: "explicit-mod",
					Name:    "Explicit Module",
					Type:    "go-library",
					Files: Files{
						Root:   "src/explicit",
						Source: []string{"custom/*.go"},
						Config: []string{"custom.mod"},
						Repo: RepoFiles{
							Specs:    []string{"my/specs/**"},
							TestImpl: "my/tests",
						},
					},
				},
			},
		}

		types := &ModuleTypesConfig{
			Types: []ModuleTypeDef{
				{
					Name: "go-library",
					Defaults: &TypeDefaults{
						Files: &FilesDefaults{
							Source: []string{"**/*.go"},
							Config: []string{"go.mod"},
						},
						Repo: &RepoDefaults{
							Specs:    []string{"specs/{moniker}/**"},
							TestImpl: "{root}/tests",
						},
					},
				},
			},
		}

		modules.ApplyTypeDefaults(types)

		m := modules.Modules[0]
		// Explicit values should be preserved
		assert.Equal(t, []string{"custom/*.go"}, m.Files.Source)
		assert.Equal(t, []string{"custom.mod"}, m.Files.Config)
		assert.Equal(t, []string{"my/specs/**"}, m.Files.Repo.Specs)
		assert.Equal(t, "my/tests", m.Files.Repo.TestImpl)
	})

	t.Run("handles nil types", func(t *testing.T) {
		modules := &ModulesConfig{
			Modules: []Module{
				{
					Moniker: "test-mod",
					Name:    "Test Module",
					Type:    "go-library",
					Files: Files{
						Root: "src/test",
					},
				},
			},
		}

		// Should not panic with nil types
		modules.ApplyTypeDefaults(nil)

		m := modules.Modules[0]
		// Generic defaults should be applied
		assert.Equal(t, "CHANGELOG.md", m.Files.Changelog)
		assert.Equal(t, []string{"specs/test-mod/**"}, m.Files.Repo.Specs)
	})

	t.Run("preserves explicit empty specs", func(t *testing.T) {
		modules := &ModulesConfig{
			Modules: []Module{
				{
					Moniker: "no-specs-mod",
					Name:    "No Specs Module",
					Type:    "go-library",
					Files: Files{
						Root: "src/nospecs",
						Repo: RepoFiles{
							Specs: []string{}, // Explicit empty
						},
					},
				},
			},
		}

		types := &ModuleTypesConfig{
			Types: []ModuleTypeDef{
				{
					Name: "go-library",
					Defaults: &TypeDefaults{
						Repo: &RepoDefaults{
							Specs: []string{"specs/{moniker}/**"},
						},
					},
				},
			},
		}

		modules.ApplyTypeDefaults(types)

		m := modules.Modules[0]
		// Explicit empty should be preserved
		assert.NotNil(t, m.Files.Repo.Specs)
		assert.Empty(t, m.Files.Repo.Specs)
	})
}

// TestModuleTypesConfig_Integration tests type config methods with real data
func TestModuleTypesConfig_Integration(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)
	require.NotNil(t, cfg.ModuleTypes)

	t.Run("Get returns type definition", func(t *testing.T) {
		goLib := cfg.ModuleTypes.Get("go-library")
		require.NotNil(t, goLib)
		assert.Equal(t, "go", goLib.BuildSystem)
	})

	t.Run("Get returns nil for unknown type", func(t *testing.T) {
		unknown := cfg.ModuleTypes.Get("unknown-xyz")
		assert.Nil(t, unknown)
	})

	t.Run("GetBuildSystem returns correct system", func(t *testing.T) {
		buildSys := cfg.ModuleTypes.GetBuildSystem("go-library")
		assert.Equal(t, "go", buildSys)
	})

	t.Run("GetTypesByBuildSystem finds go types", func(t *testing.T) {
		goTypes := cfg.ModuleTypes.GetTypesByBuildSystem("go")
		assert.NotEmpty(t, goTypes)
		assert.Contains(t, goTypes, "go-library")
	})
}
