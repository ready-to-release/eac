//go:build L0 && ov

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestModuleTypesConfig_Get tests type lookup
func TestModuleTypesConfig_Get(t *testing.T) {
	t.Run("returns correct type definition", func(t *testing.T) {
		cfg := &ModuleTypesConfig{
			Types: []ModuleTypeDef{
				{Name: "go", BuildDeps: []string{"go"}},
				{Name: "python", BuildDeps: []string{"python"}},
			},
		}

		result := cfg.Get("go")

		assert.NotNil(t, result)
		assert.Equal(t, "go", result.Name)
		assert.Equal(t, []string{"go"}, result.BuildDeps)
	})

	t.Run("returns nil for unknown type", func(t *testing.T) {
		cfg := &ModuleTypesConfig{
			Types: []ModuleTypeDef{
				{Name: "go", BuildDeps: []string{"go"}},
			},
		}

		result := cfg.Get("unknown-type")

		assert.Nil(t, result)
	})

	t.Run("handles empty types list", func(t *testing.T) {
		cfg := &ModuleTypesConfig{
			Types: []ModuleTypeDef{},
		}

		result := cfg.Get("any-type")

		assert.Nil(t, result)
	})

	t.Run("builds type map lazily", func(t *testing.T) {
		cfg := &ModuleTypesConfig{
			Types: []ModuleTypeDef{
				{Name: "go", BuildDeps: []string{"go"}},
			},
		}

		// typeMap should be nil initially
		assert.Nil(t, cfg.typeMap)

		// After Get, typeMap should be built
		_ = cfg.Get("go")
		assert.NotNil(t, cfg.typeMap)
	})
}

// TestModuleTypesConfig_HasCapability tests capability checking
func TestModuleTypesConfig_HasCapability(t *testing.T) {
	cfg := &ModuleTypesConfig{
		Types: []ModuleTypeDef{
			{
				Name:         "go",
				Capabilities: []string{"testing", "documentation"},
			},
			{
				Name:         "static",
				Capabilities: []string{},
			},
		},
	}

	t.Run("returns true when type has capability", func(t *testing.T) {
		assert.True(t, cfg.HasCapability("go", "testing"))
		assert.True(t, cfg.HasCapability("go", "documentation"))
	})

	t.Run("returns false when type lacks capability", func(t *testing.T) {
		assert.False(t, cfg.HasCapability("go", "deployment"))
	})

	t.Run("returns false for unknown type", func(t *testing.T) {
		assert.False(t, cfg.HasCapability("unknown-type", "testing"))
	})

	t.Run("returns false for type with no capabilities", func(t *testing.T) {
		assert.False(t, cfg.HasCapability("static", "testing"))
	})
}

// TestModuleTypesConfig_GetBuildDeps tests build deps lookup
func TestModuleTypesConfig_GetBuildDeps(t *testing.T) {
	cfg := &ModuleTypesConfig{
		Types: []ModuleTypeDef{
			{Name: "go", BuildDeps: []string{"go"}},
			{Name: "container", BuildDeps: []string{"docker"}},
			{Name: "static", BuildDeps: []string{}},
		},
	}

	t.Run("returns correct build deps", func(t *testing.T) {
		assert.Equal(t, []string{"go"}, cfg.GetBuildDeps("go"))
		assert.Equal(t, []string{"docker"}, cfg.GetBuildDeps("container"))
	})

	t.Run("returns nil for unknown type", func(t *testing.T) {
		assert.Nil(t, cfg.GetBuildDeps("unknown"))
	})

	t.Run("returns empty slice for type with no build deps", func(t *testing.T) {
		assert.Equal(t, []string{}, cfg.GetBuildDeps("static"))
	})
}

// TestModuleTypesConfig_GetPrimaryBuildDep tests primary build dep lookup
func TestModuleTypesConfig_GetPrimaryBuildDep(t *testing.T) {
	cfg := &ModuleTypesConfig{
		Types: []ModuleTypeDef{
			{Name: "go", BuildDeps: []string{"go"}},
			{Name: "container", BuildDeps: []string{"docker"}},
			{Name: "static", BuildDeps: []string{}},
		},
	}

	t.Run("returns first build dep", func(t *testing.T) {
		assert.Equal(t, "go", cfg.GetPrimaryBuildDep("go"))
		assert.Equal(t, "docker", cfg.GetPrimaryBuildDep("container"))
	})

	t.Run("returns empty for unknown type", func(t *testing.T) {
		assert.Equal(t, "", cfg.GetPrimaryBuildDep("unknown"))
	})

	t.Run("returns empty for type with no build deps", func(t *testing.T) {
		assert.Equal(t, "", cfg.GetPrimaryBuildDep("static"))
	})
}

// TestModuleTypesConfig_GetTypesWithCapability tests capability filtering
func TestModuleTypesConfig_GetTypesWithCapability(t *testing.T) {
	cfg := &ModuleTypesConfig{
		Types: []ModuleTypeDef{
			{Name: "go", Capabilities: []string{"testing", "docs"}},
			{Name: "typescript", Capabilities: []string{"testing", "npm"}},
			{Name: "container", Capabilities: []string{"docs"}},
			{Name: "static", Capabilities: []string{}},
		},
	}

	t.Run("returns types with capability", func(t *testing.T) {
		result := cfg.GetTypesWithCapability("testing")
		assert.ElementsMatch(t, []string{"go", "typescript"}, result)
	})

	t.Run("returns single type", func(t *testing.T) {
		result := cfg.GetTypesWithCapability("npm")
		assert.Equal(t, []string{"typescript"}, result)
	})

	t.Run("returns empty for unknown capability", func(t *testing.T) {
		result := cfg.GetTypesWithCapability("deployment")
		assert.Empty(t, result)
	})
}

// TestModuleTypesConfig_GetTypesWithBuildDep tests build dep filtering
func TestModuleTypesConfig_GetTypesWithBuildDep(t *testing.T) {
	cfg := &ModuleTypesConfig{
		Types: []ModuleTypeDef{
			{Name: "go", BuildDeps: []string{"go"}},
			{Name: "typescript", BuildDeps: []string{"node"}},
			{Name: "container", BuildDeps: []string{"docker"}},
			{Name: "static", BuildDeps: []string{}},
		},
	}

	t.Run("returns types with build dep", func(t *testing.T) {
		result := cfg.GetTypesWithBuildDep("go")
		assert.Equal(t, []string{"go"}, result)
	})

	t.Run("returns types with docker dep", func(t *testing.T) {
		result := cfg.GetTypesWithBuildDep("docker")
		assert.Equal(t, []string{"container"}, result)
	})

	t.Run("returns empty for unknown build dep", func(t *testing.T) {
		result := cfg.GetTypesWithBuildDep("rust")
		assert.Empty(t, result)
	})
}

// TestModuleTypeDef_HasCapability tests the type-level capability check
func TestModuleTypeDef_HasCapability(t *testing.T) {
	typeDef := &ModuleTypeDef{
		Name:         "go",
		Capabilities: []string{"testing", "documentation"},
	}

	t.Run("returns true for existing capability", func(t *testing.T) {
		assert.True(t, typeDef.HasCapability("testing"))
	})

	t.Run("returns false for missing capability", func(t *testing.T) {
		assert.False(t, typeDef.HasCapability("deployment"))
	})

	t.Run("handles nil capabilities", func(t *testing.T) {
		nilCapDef := &ModuleTypeDef{Name: "test", Capabilities: nil}
		assert.False(t, nilCapDef.HasCapability("any"))
	})
}

// TestModuleTypeDef_Defaults tests defaults access
func TestModuleTypeDef_Defaults(t *testing.T) {
	t.Run("can access Files defaults", func(t *testing.T) {
		typeDef := &ModuleTypeDef{
			Name: "go",
			Defaults: &TypeDefaults{
				Files: &FilesDefaults{
					Source: []string{"**/*.go"},
					Config: []string{"go.mod", "go.sum"},
				},
			},
		}

		assert.NotNil(t, typeDef.Defaults)
		assert.NotNil(t, typeDef.Defaults.Files)
		assert.Equal(t, []string{"**/*.go"}, typeDef.Defaults.Files.Source)
		assert.Equal(t, []string{"go.mod", "go.sum"}, typeDef.Defaults.Files.Config)
	})

	t.Run("can access Repo defaults", func(t *testing.T) {
		typeDef := &ModuleTypeDef{
			Name: "go",
			Defaults: &TypeDefaults{
				Repo: &RepoDefaults{
					Specs:    []string{"specs/{moniker}/**"},
					TestImpl: "{root}/tests",
					Design:   "specs/{moniker}/.design",
				},
			},
		}

		assert.NotNil(t, typeDef.Defaults.Repo)
		assert.Equal(t, []string{"specs/{moniker}/**"}, typeDef.Defaults.Repo.Specs)
		assert.Equal(t, "{root}/tests", typeDef.Defaults.Repo.TestImpl)
		assert.Equal(t, "specs/{moniker}/.design", typeDef.Defaults.Repo.Design)
	})

	t.Run("handles nil Defaults gracefully", func(t *testing.T) {
		typeDef := &ModuleTypeDef{
			Name:     "basic-type",
			Defaults: nil,
		}

		assert.Nil(t, typeDef.Defaults)
	})

	t.Run("handles partial defaults", func(t *testing.T) {
		typeDef := &ModuleTypeDef{
			Name: "partial",
			Defaults: &TypeDefaults{
				Files: &FilesDefaults{
					Source: []string{"**/*.txt"},
				},
				// Repo is nil
			},
		}

		assert.NotNil(t, typeDef.Defaults.Files)
		assert.Nil(t, typeDef.Defaults.Repo)
	})
}

// TestTypeDefaults_Structure tests the structure hierarchy
func TestTypeDefaults_Structure(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		defaults := &TypeDefaults{
			Files: &FilesDefaults{
				Source:    []string{"**/*.go"},
				Config:    []string{"go.mod"},
				Assets:    []string{"README.md"},
				Tests:     []string{"**/*_test.go"},
				Changelog: "CHANGELOG.md",
			},
			Repo: &RepoDefaults{
				Specs:    []string{"specs/{moniker}/**"},
				TestImpl: "{root}/tests",
				Design:   "specs/{moniker}/.design",
			},
		}

		// Files
		assert.Equal(t, []string{"**/*.go"}, defaults.Files.Source)
		assert.Equal(t, []string{"go.mod"}, defaults.Files.Config)
		assert.Equal(t, []string{"README.md"}, defaults.Files.Assets)
		assert.Equal(t, []string{"**/*_test.go"}, defaults.Files.Tests)
		assert.Equal(t, "CHANGELOG.md", defaults.Files.Changelog)

		// Repo
		assert.Equal(t, []string{"specs/{moniker}/**"}, defaults.Repo.Specs)
		assert.Equal(t, "{root}/tests", defaults.Repo.TestImpl)
		assert.Equal(t, "specs/{moniker}/.design", defaults.Repo.Design)
	})
}
