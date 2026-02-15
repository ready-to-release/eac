//go:build L0 && ov

package config

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleTemplate_VersioningFromTemplate(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"go-library": {
			Versioning: &ModuleVersioning{
				Scheme:      "Implicit",
				ReleaseType: "none",
			},
		},
	}

	mod := &Module{
		Moniker:  "my-lib",
		Template: "go-library",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/my-lib"},
		},
	}

	err := ExpandModuleFromTemplate(mod, templates, nil)
	require.NoError(t, err)

	// Versioning should come from template
	require.NotNil(t, mod.Versioning)
	assert.Equal(t, "Implicit", mod.Versioning.Scheme)
	assert.Equal(t, "none", mod.Versioning.ReleaseType)

	// User-declared component should be preserved
	assert.Contains(t, mod.Components, "go")
	assert.Equal(t, "go/my-lib", mod.Components["go"].Root)
}

func TestModuleTemplate_UnknownTemplate(t *testing.T) {
	templates := map[string]ModuleTemplate{}

	mod := &Module{
		Moniker:  "test-mod",
		Template: "nonexistent",
	}

	err := ExpandModuleFromTemplate(mod, templates, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown template")
}

func TestModuleTemplate_ModuleVersioningWins(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"go-cli": {
			Versioning: &ModuleVersioning{
				Scheme:      "SemVer",
				Changelog:   "release/{moniker}/CHANGELOG.md",
				ReleaseType: "published",
			},
		},
	}

	mod := &Module{
		Moniker:  "my-cli",
		Template: "go-cli",
		// Module explicitly overrides versioning
		Versioning: &ModuleVersioning{
			Scheme:      "CalVer",
			ReleaseType: "internal",
		},
	}

	err := ExpandModuleFromTemplate(mod, templates, nil)
	require.NoError(t, err)

	// Module's versioning should win
	assert.Equal(t, "CalVer", mod.Versioning.Scheme)
	assert.Equal(t, "internal", mod.Versioning.ReleaseType)
}

func TestModuleTemplate_ParameterSubstitution(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"container": {},
	}

	mod := &Module{
		Moniker:  "my-app",
		Template: "container",
		Components: ModuleComponents{
			"dockerfile": &ComponentEntry{
				Root: "containers/{moniker}",
				DockerBuild: &DockerBuildConfig{
					Container:  "{moniker}",
					Context:    "containers/{moniker}",
					Dockerfile: "containers/{moniker}/Dockerfile",
					Tags: []string{
						"ghcr.io/{owner}/{moniker}:latest",
						"ghcr.io/{owner}/{moniker}:sha-{short_sha}",
					},
				},
			},
		},
	}

	vars := map[string]string{
		"moniker": "my-app",
		"owner":   "ready-to-release",
	}
	err := ExpandModuleFromTemplate(mod, templates, vars)
	require.NoError(t, err)

	dbc := mod.Components["dockerfile"].DockerBuild
	require.NotNil(t, dbc)

	assert.Equal(t, "my-app", dbc.Container)
	assert.Equal(t, "containers/my-app", dbc.Context)
	assert.Equal(t, "containers/my-app/Dockerfile", dbc.Dockerfile)
	assert.Contains(t, dbc.Tags, "ghcr.io/ready-to-release/my-app:latest")
	// short_sha is not provided, so placeholder remains
	assert.Contains(t, dbc.Tags, "ghcr.io/ready-to-release/my-app:sha-{short_sha}")
}

func TestModuleTemplate_DependsOnMerge(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"go-library": {
			DependsOn: []string{"contracts", "core"},
		},
	}

	mod := &Module{
		Moniker:   "my-lib",
		Template:  "go-library",
		DependsOn: []string{"core", "utils"}, // "core" overlaps with template
	}

	err := ExpandModuleFromTemplate(mod, templates, nil)
	require.NoError(t, err)

	// Should have unique deps: core (from module), utils (from module), contracts (from template)
	assert.Contains(t, mod.DependsOn, "core")
	assert.Contains(t, mod.DependsOn, "utils")
	assert.Contains(t, mod.DependsOn, "contracts")
	// No duplicates
	assert.Len(t, mod.DependsOn, 3)
}

func TestModuleTemplate_NoTemplate(t *testing.T) {
	// Module without template should pass through unchanged
	mod := &Module{
		Moniker: "explicit-mod",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/explicit"},
		},
	}

	err := ExpandModuleFromTemplate(mod, nil, nil)
	require.NoError(t, err)

	// Should be unchanged
	assert.Equal(t, "go/explicit", mod.Components["go"].Root)
}

func TestSubstituteParams(t *testing.T) {
	params := map[string]string{
		"moniker": "my-app",
		"owner":   "myorg",
		"go_root": "go/my-app",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"containers/{moniker}", "containers/my-app"},
		{"{owner}/{moniker}", "myorg/my-app"},
		{"{go_root}/internal", "go/my-app/internal"},
		{"no-placeholders", "no-placeholders"},
		{"", ""},
		{"{unknown}", "{unknown}"}, // Unknown placeholders left as-is
	}

	for _, tc := range tests {
		result := substituteParams(tc.input, params)
		assert.Equal(t, tc.expected, result, "input: %s", tc.input)
	}
}

func TestBuildModuleParams(t *testing.T) {
	mod := &Module{
		Moniker: "my-module",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/my-module"},
		},
	}

	discoveryVars := map[string]string{
		"owner":   "myorg",
		"go_root": "go/my-module",
	}
	params := buildModuleParams(mod, discoveryVars)

	assert.Equal(t, "my-module", params["moniker"])
	assert.Equal(t, "myorg", params["owner"])
	assert.Equal(t, "go/my-module", params["go_root"]) // From discovery vars
}

// =============================================================================
// Tests for Template Component Merging
// =============================================================================

func boolPtr(b bool) *bool { return &b }

// TestTemplateComponents_DockerBuildFromTemplate verifies that when a template
// provides DockerBuild and the module component doesn't, the template wins.
func TestTemplateComponents_DockerBuildFromTemplate(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"container": {
			Components: map[string]*ComponentEntry{
				"dockerfile": {
					DockerBuild: &DockerBuildConfig{
						Container:  "{moniker}",
						Context:    "containers/{moniker}",
						Dockerfile: "containers/{moniker}/Dockerfile",
						Platforms:  []string{"linux/amd64"},
						Tags: []string{
							"ghcr.io/{owner}/{moniker}:latest",
							"ghcr.io/{owner}/{moniker}:sha-{short_sha}",
						},
						Push:     boolPtr(true),
						Registry: "ghcr.io",
						Cache:    &DockerCacheConfig{Type: "gha"},
					},
				},
			},
		},
	}

	mod := &Module{
		Moniker:  "my-oci",
		Template: "container",
		Components: ModuleComponents{
			"dockerfile": &ComponentEntry{Root: "containers/my-oci"},
		},
	}

	vars := map[string]string{
		"moniker": "my-oci",
		"owner":   "ready-to-release",
	}
	err := ExpandModuleFromTemplate(mod, templates, vars)
	require.NoError(t, err)

	dbc := mod.Components["dockerfile"].DockerBuild
	require.NotNil(t, dbc, "DockerBuild should come from template")

	// After parameter substitution, placeholders should be resolved
	assert.Equal(t, "my-oci", dbc.Container)
	assert.Equal(t, "containers/my-oci", dbc.Context)
	assert.Equal(t, "containers/my-oci/Dockerfile", dbc.Dockerfile)
	assert.Equal(t, []string{"linux/amd64"}, dbc.Platforms)
	assert.Contains(t, dbc.Tags, "ghcr.io/ready-to-release/my-oci:latest")
	assert.NotNil(t, dbc.Push)
	assert.True(t, *dbc.Push)
	assert.Equal(t, "ghcr.io", dbc.Registry)
	require.NotNil(t, dbc.Cache)
	assert.Equal(t, "gha", dbc.Cache.Type)

	// Module root should be preserved
	assert.Equal(t, "containers/my-oci", mod.Components["dockerfile"].Root)
}

// TestTemplateComponents_ModuleDockerBuildOverrides verifies that when a module
// provides its own DockerBuild (e.g., push: false), the module's fields take
// precedence while template fields fill in as defaults.
func TestTemplateComponents_ModuleDockerBuildOverrides(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"container": {
			Components: map[string]*ComponentEntry{
				"dockerfile": {
					DockerBuild: &DockerBuildConfig{
						Container:  "{moniker}",
						Context:    "containers/{moniker}",
						Dockerfile: "containers/{moniker}/Dockerfile",
						Platforms:  []string{"linux/amd64"},
						Tags: []string{
							"ghcr.io/{owner}/{moniker}:latest",
						},
						Push:     boolPtr(true),
						Registry: "ghcr.io",
						Cache:    &DockerCacheConfig{Type: "gha"},
					},
				},
			},
		},
	}

	mod := &Module{
		Moniker:  "drawio-oci",
		Template: "container",
		Components: ModuleComponents{
			"dockerfile": &ComponentEntry{
				Root: "containers/drawio-oci",
				DockerBuild: &DockerBuildConfig{
					Push: boolPtr(false),
				},
			},
		},
	}

	vars := map[string]string{
		"moniker": "drawio-oci",
		"owner":   "ready-to-release",
	}
	err := ExpandModuleFromTemplate(mod, templates, vars)
	require.NoError(t, err)

	dbc := mod.Components["dockerfile"].DockerBuild
	require.NotNil(t, dbc)

	// Module's push: false should override template's push: true
	require.NotNil(t, dbc.Push)
	assert.False(t, *dbc.Push)

	// Template defaults should fill in the remaining fields (after substitution)
	assert.Equal(t, "drawio-oci", dbc.Container)
	assert.Equal(t, "containers/drawio-oci", dbc.Context)
	assert.Equal(t, "containers/drawio-oci/Dockerfile", dbc.Dockerfile)
	assert.Equal(t, []string{"linux/amd64"}, dbc.Platforms)
	assert.Contains(t, dbc.Tags, "ghcr.io/ready-to-release/drawio-oci:latest")
	assert.Equal(t, "ghcr.io", dbc.Registry)
	require.NotNil(t, dbc.Cache)
	assert.Equal(t, "gha", dbc.Cache.Type)
}

// TestTemplateComponents_SkipsNonExistentModuleComponent verifies that template
// components are NOT auto-created in the module — they only provide defaults
// for components the module already declares.
func TestTemplateComponents_SkipsNonExistentModuleComponent(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"container": {
			Components: map[string]*ComponentEntry{
				"dockerfile": {
					DockerBuild: &DockerBuildConfig{
						Container: "{moniker}",
						Push:      boolPtr(true),
					},
				},
			},
		},
	}

	// Module has NO "dockerfile" component
	mod := &Module{
		Moniker:  "my-lib",
		Template: "container",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/my-lib"},
		},
	}

	err := ExpandModuleFromTemplate(mod, templates, nil)
	require.NoError(t, err)

	// Template's "dockerfile" component should NOT be created
	assert.NotContains(t, mod.Components, "dockerfile")
	// Existing components should be preserved
	assert.Contains(t, mod.Components, "go")
}

// TestTemplateComponents_RootNotOverridden verifies that the template
// never overrides a module component's Root.
func TestTemplateComponents_RootNotOverridden(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"container": {
			Components: map[string]*ComponentEntry{
				"dockerfile": {
					Root: "should-not-apply",
					DockerBuild: &DockerBuildConfig{
						Container: "{moniker}",
						Push:      boolPtr(true),
					},
				},
			},
		},
	}

	mod := &Module{
		Moniker:  "my-oci",
		Template: "container",
		Components: ModuleComponents{
			"dockerfile": &ComponentEntry{Root: "containers/my-oci"},
		},
	}

	err := ExpandModuleFromTemplate(mod, templates, nil)
	require.NoError(t, err)

	// Module root should be preserved, not overridden by template
	assert.Equal(t, "containers/my-oci", mod.Components["dockerfile"].Root)
}

// Integration test: Load real templates from defaults file
func TestLoadBlueprintsDefaults_Integration(t *testing.T) {
	// Find repo root
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	// Load templates
	cfg, err := LoadBlueprintsDefaults(repoRoot)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify expected templates exist
	assert.Contains(t, cfg.Templates, "go-library")
	assert.Contains(t, cfg.Templates, "go-cli")
	assert.Contains(t, cfg.Templates, "go-exe")
	assert.Contains(t, cfg.Templates, "go-contracts")
	assert.Contains(t, cfg.Templates, "container")
	assert.Contains(t, cfg.Templates, "container-multiarch")
	assert.Contains(t, cfg.Templates, "docs-site")

	// Verify go-library template structure
	goLib := cfg.Templates["go-library"]
	require.NotNil(t, goLib.Versioning)
	assert.Equal(t, "Implicit", goLib.Versioning.Scheme)

	// Verify container template has docker_build component defaults
	container := cfg.Templates["container"]
	require.NotNil(t, container.Components, "container template should have component defaults")
	require.Contains(t, container.Components, "dockerfile")
	require.NotNil(t, container.Components["dockerfile"].DockerBuild)
	assert.Equal(t, "{moniker}", container.Components["dockerfile"].DockerBuild.Container)
	assert.Equal(t, "ghcr.io", container.Components["dockerfile"].DockerBuild.Registry)
	require.NotNil(t, container.Components["dockerfile"].DockerBuild.Cache)
	assert.Equal(t, "gha", container.Components["dockerfile"].DockerBuild.Cache.Type)

	// Verify new templates exist
	assert.Contains(t, cfg.Templates, "script-installer")
	assert.Contains(t, cfg.Templates, "scripts-bundle")
}

// =============================================================================
// Tests for MergeBlueprintsConfig
// =============================================================================

func TestMergeBlueprintsConfig_BothNil(t *testing.T) {
	result := MergeBlueprintsConfig(nil, nil)
	require.NotNil(t, result)
	assert.Empty(t, result.Templates)
	assert.Empty(t, result.ArtifactMatrices)
}

func TestMergeBlueprintsConfig_BaseNil(t *testing.T) {
	override := &BlueprintsConfig{
		Templates: map[string]ModuleTemplate{
			"go-lib": {},
		},
	}
	result := MergeBlueprintsConfig(nil, override)
	assert.Equal(t, override, result)
}

func TestMergeBlueprintsConfig_OverrideNil(t *testing.T) {
	base := &BlueprintsConfig{
		Templates: map[string]ModuleTemplate{
			"go-lib": {},
		},
	}
	result := MergeBlueprintsConfig(base, nil)
	assert.Equal(t, base, result)
}

func TestMergeBlueprintsConfig_OverrideKeys(t *testing.T) {
	base := &BlueprintsConfig{
		Templates: map[string]ModuleTemplate{
			"go-lib": {},
		},
		ArtifactMatrices: map[string]*ArtifactMatrix{
			"cross-platform": {Entries: []ArtifactMatrixEntry{{ID: "linux-amd64"}}},
		},
	}

	override := &BlueprintsConfig{
		Templates: map[string]ModuleTemplate{
			"new-template": {}, // Add new
		},
		ArtifactMatrices: map[string]*ArtifactMatrix{
			"cross-platform": {Entries: []ArtifactMatrixEntry{{ID: "override"}}}, // Override existing
		},
	}

	result := MergeBlueprintsConfig(base, override)

	// Base template preserved
	assert.Contains(t, result.Templates, "go-lib")
	// New template added
	assert.Contains(t, result.Templates, "new-template")

	// Overridden matrix
	assert.Equal(t, "override", result.ArtifactMatrices["cross-platform"].Entries[0].ID)
}

func TestMergeBlueprintsConfig_EmptyMaps(t *testing.T) {
	base := &BlueprintsConfig{
		Templates: map[string]ModuleTemplate{
			"go-lib": {},
		},
	}
	override := &BlueprintsConfig{} // All maps nil

	result := MergeBlueprintsConfig(base, override)

	// Base should be preserved
	assert.Contains(t, result.Templates, "go-lib")
}

// TestExpandModuleTemplates_NeverCreatesUndeclaredComponents is a regression test
// ensuring the pipeline never auto-creates components that weren't explicitly declared.
// This guards against reintroduction of the removed discovery/blueprint creation logic.
func TestExpandModuleTemplates_NeverCreatesUndeclaredComponents(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"container": {
			Components: map[string]*ComponentEntry{
				"dockerfile": {
					DockerBuild: &DockerBuildConfig{
						Container: "{moniker}",
						Push:      boolPtr(true),
					},
				},
				"container-assets": {
					Root: "containers/{moniker}",
				},
			},
		},
	}

	// Module only declares "go" — NOT "dockerfile" or "container-assets"
	mod := &Module{
		Moniker:  "my-lib",
		Template: "container",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/my-lib"},
		},
	}

	err := ExpandModuleFromTemplate(mod, templates, nil)
	require.NoError(t, err)

	// ONLY the explicitly declared "go" component should exist
	assert.Len(t, mod.Components, 1, "only explicitly declared components should exist")
	assert.Contains(t, mod.Components, "go")
	assert.NotContains(t, mod.Components, "dockerfile", "undeclared template component must not be created")
	assert.NotContains(t, mod.Components, "container-assets", "undeclared template component must not be created")
}

// TestLoadBlueprintsDefaults_ArtifactMatrices verifies that artifact matrices are loaded.
func TestLoadBlueprintsDefaults_ArtifactMatrices(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := LoadBlueprintsDefaults(repoRoot)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify artifact matrices are loaded
	require.NotNil(t, cfg.ArtifactMatrices)
	assert.Contains(t, cfg.ArtifactMatrices, "cross-platform")
	assert.Contains(t, cfg.ArtifactMatrices, "cross-platform-upx")
	assert.Contains(t, cfg.ArtifactMatrices, "single-platform")

	// Verify cross-platform matrix has expected entries
	cp := cfg.ArtifactMatrices["cross-platform"]
	require.NotNil(t, cp)
	assert.Len(t, cp.Entries, 5)
}
