//go:build L0 && ov

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleTemplate_BasicExpansion(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"container": {
			Components: ModuleComponents{
				"dockerfile": &ComponentEntry{Root: "containers/{moniker}"},
				"markdown":   &ComponentEntry{Root: "containers/{moniker}"},
			},
		},
	}

	mod := &Module{
		Moniker:    "test-container",
		Name:       "Test Container",
		Template:   "container",
		Components: make(ModuleComponents),
	}

	err := ExpandModuleFromTemplate(mod, templates, nil, "", "myorg")
	require.NoError(t, err)

	// Template components should be merged
	assert.Contains(t, mod.Components, "dockerfile")
	assert.Contains(t, mod.Components, "markdown")

	// Placeholders should be substituted
	assert.Equal(t, "containers/test-container", mod.Components["dockerfile"].Root)
	assert.Equal(t, "containers/test-container", mod.Components["markdown"].Root)
}

func TestModuleTemplate_UnknownTemplate(t *testing.T) {
	templates := map[string]ModuleTemplate{}

	mod := &Module{
		Moniker:  "test-mod",
		Template: "nonexistent",
	}

	err := ExpandModuleFromTemplate(mod, templates, nil, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown template")
}

func TestModuleTemplate_ModuleOverridesTemplate(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"go-library": {
			Versioning: &ModuleVersioning{
				Scheme:      "Implicit",
				ReleaseType: "none",
			},
			Components: ModuleComponents{
				"go":       &ComponentEntry{Root: "{go_root}"},
				"markdown": &ComponentEntry{Root: "{go_root}"},
			},
		},
	}

	mod := &Module{
		Moniker:  "my-lib",
		Template: "go-library",
		Parameters: map[string]string{
			"go_root": "go/my-lib",
		},
		// Module explicitly sets different root for markdown
		Components: ModuleComponents{
			"markdown": &ComponentEntry{Root: "docs/my-lib"},
		},
	}

	err := ExpandModuleFromTemplate(mod, templates, nil, "", "")
	require.NoError(t, err)

	// Module's markdown should NOT be overwritten by template
	assert.Equal(t, "docs/my-lib", mod.Components["markdown"].Root)

	// Template's go component should be added
	assert.Contains(t, mod.Components, "go")
	assert.Equal(t, "go/my-lib", mod.Components["go"].Root)

	// Versioning should come from template
	require.NotNil(t, mod.Versioning)
	assert.Equal(t, "Implicit", mod.Versioning.Scheme)
}

func TestModuleTemplate_VersioningOverride(t *testing.T) {
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

	err := ExpandModuleFromTemplate(mod, templates, nil, "", "")
	require.NoError(t, err)

	// Module's versioning should win
	assert.Equal(t, "CalVer", mod.Versioning.Scheme)
	assert.Equal(t, "internal", mod.Versioning.ReleaseType)
}

func TestModuleTemplate_ParameterSubstitution(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"container": {
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
		},
	}

	mod := &Module{
		Moniker:    "my-app",
		Template:   "container",
		Components: make(ModuleComponents),
	}

	err := ExpandModuleFromTemplate(mod, templates, nil, "", "ready-to-release")
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

	err := ExpandModuleFromTemplate(mod, templates, nil, "", "")
	require.NoError(t, err)

	// Should have unique deps: core (from module), utils (from module), contracts (from template)
	assert.Contains(t, mod.DependsOn, "core")
	assert.Contains(t, mod.DependsOn, "utils")
	assert.Contains(t, mod.DependsOn, "contracts")
	// No duplicates
	assert.Len(t, mod.DependsOn, 3)
}

func TestModuleTemplate_InferGoRoot(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"go-library": {
			Components: ModuleComponents{
				"markdown": &ComponentEntry{Root: "{go_root}"},
				"gherkin":  &ComponentEntry{Root: "specs/{moniker}"},
			},
		},
	}

	mod := &Module{
		Moniker:  "my-lib",
		Template: "go-library",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/my-lib"}, // This provides go_root
		},
	}

	err := ExpandModuleFromTemplate(mod, templates, nil, "", "")
	require.NoError(t, err)

	// go_root should be inferred from go component
	assert.Equal(t, "go/my-lib", mod.Components["markdown"].Root)
	assert.Equal(t, "specs/my-lib", mod.Components["gherkin"].Root)
}

func TestModuleTemplate_NoTemplate(t *testing.T) {
	// Module without template should pass through unchanged
	mod := &Module{
		Moniker: "explicit-mod",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/explicit"},
		},
	}

	err := ExpandModuleFromTemplate(mod, nil, nil, "", "")
	require.NoError(t, err)

	// Should be unchanged
	assert.Equal(t, "go/explicit", mod.Components["go"].Root)
}

func TestConventionDiscovery_GherkinFound(t *testing.T) {
	// Create temp directory structure
	repoRoot := t.TempDir()
	specsDir := filepath.Join(repoRoot, "specs", "my-module")
	require.NoError(t, os.MkdirAll(specsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(specsDir, "specification.feature"), []byte("Feature: Test"), 0644))

	conv := DefaultDiscoveryConventions()
	mod := &Module{
		Moniker:      "my-module",
		AutoDiscover: true,
		Components:   make(ModuleComponents),
	}

	discoverComponents(mod, conv, repoRoot)

	assert.Contains(t, mod.Components, "gherkin")
	assert.Equal(t, "specs/my-module", mod.Components["gherkin"].Root)
}

func TestConventionDiscovery_StructurizrFound(t *testing.T) {
	repoRoot := t.TempDir()
	designDir := filepath.Join(repoRoot, "specs", "my-module", ".design")
	require.NoError(t, os.MkdirAll(designDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(designDir, "workspace.dsl"), []byte("workspace {}"), 0644))

	conv := DefaultDiscoveryConventions()
	mod := &Module{
		Moniker:      "my-module",
		AutoDiscover: true,
		Components:   make(ModuleComponents),
	}

	discoverComponents(mod, conv, repoRoot)

	assert.Contains(t, mod.Components, "structurizr")
	assert.Equal(t, "specs/my-module/.design", mod.Components["structurizr"].Root)
}

func TestConventionDiscovery_GherkinStepsFound(t *testing.T) {
	repoRoot := t.TempDir()
	// gherkin-steps are discovered based on go component location + /specs subdirectory
	goDir := filepath.Join(repoRoot, "go", "my-module")
	specsDir := filepath.Join(goDir, "specs")
	require.NoError(t, os.MkdirAll(specsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(specsDir, "godog_test.go"), []byte("package test"), 0644))

	conv := DefaultDiscoveryConventions()
	mod := &Module{
		Moniker:      "my-module",
		AutoDiscover: true,
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/my-module"},
		},
	}

	discoverComponents(mod, conv, repoRoot)

	assert.Contains(t, mod.Components, "gherkin-steps")
	assert.Equal(t, "go/my-module/specs", mod.Components["gherkin-steps"].Root)
}

func TestConventionDiscovery_MarkdownDerived(t *testing.T) {
	repoRoot := t.TempDir()
	goDir := filepath.Join(repoRoot, "go", "my-lib")
	require.NoError(t, os.MkdirAll(goDir, 0755))

	conv := DefaultDiscoveryConventions()
	mod := &Module{
		Moniker:      "my-lib",
		AutoDiscover: true,
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/my-lib"},
		},
	}

	discoverComponents(mod, conv, repoRoot)

	// Markdown should be derived from go component
	assert.Contains(t, mod.Components, "markdown")
	assert.Equal(t, "go/my-lib", mod.Components["markdown"].Root)
}

func TestConventionDiscovery_NotFoundWhenMissing(t *testing.T) {
	repoRoot := t.TempDir()
	// Don't create any directories

	conv := DefaultDiscoveryConventions()
	mod := &Module{
		Moniker:      "my-module",
		AutoDiscover: true,
		Components:   make(ModuleComponents),
	}

	discoverComponents(mod, conv, repoRoot)

	// Nothing should be discovered
	assert.Empty(t, mod.Components)
}

func TestConventionDiscovery_DoesNotOverwriteExisting(t *testing.T) {
	repoRoot := t.TempDir()
	specsDir := filepath.Join(repoRoot, "specs", "my-module")
	require.NoError(t, os.MkdirAll(specsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(specsDir, "specification.feature"), []byte("Feature: Test"), 0644))

	conv := DefaultDiscoveryConventions()
	mod := &Module{
		Moniker:      "my-module",
		AutoDiscover: true,
		Components: ModuleComponents{
			"gherkin": &ComponentEntry{Root: "custom/path"}, // Already defined
		},
	}

	discoverComponents(mod, conv, repoRoot)

	// Should NOT overwrite existing component
	assert.Equal(t, "custom/path", mod.Components["gherkin"].Root)
}

func TestResolveConventionalComponents_NilMarker(t *testing.T) {
	repoRoot := t.TempDir()
	specsDir := filepath.Join(repoRoot, "specs", "my-module")
	require.NoError(t, os.MkdirAll(specsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(specsDir, "specification.feature"), []byte("Feature: Test"), 0644))

	conv := DefaultDiscoveryConventions()
	mod := &Module{
		Moniker: "my-module",
		Components: ModuleComponents{
			"gherkin": nil, // nil marker means "use convention"
		},
	}

	resolveConventionalComponents(mod, conv, repoRoot)

	// Should resolve nil to conventional path
	require.NotNil(t, mod.Components["gherkin"])
	assert.Equal(t, "specs/my-module", mod.Components["gherkin"].Root)
}

func TestResolveConventionalComponents_NilMarkerRemoved(t *testing.T) {
	repoRoot := t.TempDir()
	// Don't create specs directory

	conv := DefaultDiscoveryConventions()
	mod := &Module{
		Moniker: "my-module",
		Components: ModuleComponents{
			"gherkin": nil, // nil marker for non-existent path
		},
	}

	resolveConventionalComponents(mod, conv, repoRoot)

	// nil marker should be removed since path doesn't exist
	assert.NotContains(t, mod.Components, "gherkin")
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
		Parameters: map[string]string{
			"custom_param": "custom_value",
		},
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/my-module"},
		},
	}

	params := buildModuleParams(mod, "myorg")

	assert.Equal(t, "my-module", params["moniker"])
	assert.Equal(t, "myorg", params["owner"])
	assert.Equal(t, "custom_value", params["custom_param"])
	assert.Equal(t, "go/my-module", params["go_root"]) // Inferred from go component
}

func TestDefaultDiscoveryConventions(t *testing.T) {
	conv := DefaultDiscoveryConventions()

	require.NotNil(t, conv.Gherkin)
	assert.Equal(t, "specs/{moniker}", conv.Gherkin.PathPattern)
	assert.Equal(t, "specification.feature", conv.Gherkin.RequiredFile)

	require.NotNil(t, conv.Structurizr)
	assert.Equal(t, "specs/{moniker}/.design", conv.Structurizr.PathPattern)
	assert.Equal(t, "workspace.dsl", conv.Structurizr.RequiredFile)

	require.NotNil(t, conv.GherkinSteps)
	assert.Equal(t, "go/eac/specs/{moniker}", conv.GherkinSteps.FallbackPattern)
	assert.Equal(t, "godog_test.go", conv.GherkinSteps.RequiredFile)

	require.NotNil(t, conv.Markdown)
	assert.Contains(t, conv.Markdown.DeriveFrom, "go")
}

func TestCloneComponentEntry(t *testing.T) {
	original := &ComponentEntry{
		Type: "go",
		Root: "go/my-lib",
		Patterns: &ComponentPatterns{
			Source: []string{"**/*.go"},
			Tests:  []string{"**/*_test.go"},
		},
		Build: &ModuleBuild{
			Handler: "go",
		},
		Amp: &AmpConfig{
			Build: 2.0,
			Test:  1.5,
		},
	}

	clone := cloneComponentEntry(original)

	// Should be equal but different instances
	assert.Equal(t, original.Type, clone.Type)
	assert.Equal(t, original.Root, clone.Root)
	assert.Equal(t, original.Patterns.Source, clone.Patterns.Source)
	assert.Equal(t, original.Amp.Build, clone.Amp.Build)

	// Modifying clone should not affect original
	clone.Root = "changed"
	clone.Patterns.Source[0] = "changed"
	clone.Amp.Build = 9.9

	assert.Equal(t, "go/my-lib", original.Root)
	assert.Equal(t, "**/*.go", original.Patterns.Source[0])
	assert.Equal(t, 2.0, original.Amp.Build)
}

func TestCloneComponentEntry_Nil(t *testing.T) {
	clone := cloneComponentEntry(nil)
	assert.Nil(t, clone)
}

func TestModuleTemplate_DeepMergeDockerBuild(t *testing.T) {
	// Template provides base docker_build config
	templates := map[string]ModuleTemplate{
		"container": {
			Components: ModuleComponents{
				"dockerfile": &ComponentEntry{
					Root: "containers/{moniker}",
					DockerBuild: &DockerBuildConfig{
						Container:  "{moniker}",
						Context:    "containers/{moniker}",
						Dockerfile: "containers/{moniker}/Dockerfile",
						Platforms:  []string{"linux/amd64"},
						Tags: []string{
							"ghcr.io/{owner}/{moniker}:latest",
							"ghcr.io/{owner}/{moniker}:sha-{short_sha}",
						},
						Push:     true,
						Registry: "ghcr.io",
						Cache: &DockerCacheConfig{
							Type: "gha",
						},
					},
				},
			},
		},
	}

	// Module overrides only platforms (wants multi-arch)
	mod := &Module{
		Moniker:  "my-multiarch",
		Template: "container",
		Components: ModuleComponents{
			"dockerfile": &ComponentEntry{
				DockerBuild: &DockerBuildConfig{
					Platforms: []string{"linux/amd64", "linux/arm64"},
				},
			},
		},
	}

	err := ExpandModuleFromTemplate(mod, templates, nil, "", "myorg")
	require.NoError(t, err)

	// Verify deep merge
	dockerComp := mod.Components["dockerfile"]
	require.NotNil(t, dockerComp)
	require.NotNil(t, dockerComp.DockerBuild)

	dbc := dockerComp.DockerBuild
	// Override value should be used
	assert.Equal(t, []string{"linux/amd64", "linux/arm64"}, dbc.Platforms)
	// Template values should be preserved
	assert.Equal(t, "my-multiarch", dbc.Container)
	assert.Equal(t, "containers/my-multiarch", dbc.Context)
	assert.Equal(t, "containers/my-multiarch/Dockerfile", dbc.Dockerfile)
	assert.Contains(t, dbc.Tags, "ghcr.io/myorg/my-multiarch:latest")
	assert.True(t, dbc.Push)
	assert.Equal(t, "ghcr.io", dbc.Registry)
	require.NotNil(t, dbc.Cache)
	assert.Equal(t, "gha", dbc.Cache.Type)
}

func TestModuleTemplate_DeepMergePartialComponent(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"go-library": {
			Versioning: &ModuleVersioning{
				Scheme:      "Implicit",
				ReleaseType: "none",
			},
			Components: ModuleComponents{
				"go":       &ComponentEntry{Root: "{go_root}"},
				"markdown": &ComponentEntry{Root: "{go_root}"},
			},
		},
	}

	// Module provides partial component override (just type for markdown)
	mod := &Module{
		Moniker:  "my-lib",
		Template: "go-library",
		Parameters: map[string]string{
			"go_root": "go/my-lib",
		},
		Components: ModuleComponents{
			"markdown": &ComponentEntry{
				Type: "docs-markdown", // Override type, keep template's root
			},
		},
	}

	err := ExpandModuleFromTemplate(mod, templates, nil, "", "")
	require.NoError(t, err)

	// Verify component was merged
	mdComp := mod.Components["markdown"]
	require.NotNil(t, mdComp)
	assert.Equal(t, "docs-markdown", mdComp.Type) // Module's override
	assert.Equal(t, "go/my-lib", mdComp.Root)     // Template's value (substituted)
}

// Integration test: Load real templates from defaults file
func TestLoadModuleTemplates_Integration(t *testing.T) {
	// Find repo root
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	// Load templates
	cfg, err := LoadModuleTemplates(repoRoot)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify expected templates exist
	assert.Contains(t, cfg.Templates, "go-library")
	assert.Contains(t, cfg.Templates, "go-cli")
	assert.Contains(t, cfg.Templates, "container")
	assert.Contains(t, cfg.Templates, "container-multiarch")
	assert.Contains(t, cfg.Templates, "documentation-site")

	// Verify go-library template structure
	goLib := cfg.Templates["go-library"]
	require.NotNil(t, goLib.Versioning)
	assert.Equal(t, "Implicit", goLib.Versioning.Scheme)
	assert.Contains(t, goLib.Components, "go")
	assert.Contains(t, goLib.Components, "markdown")

	// Verify container template structure
	container := cfg.Templates["container"]
	assert.Contains(t, container.Components, "dockerfile")
	dockerComp := container.Components["dockerfile"]
	require.NotNil(t, dockerComp)
	require.NotNil(t, dockerComp.DockerBuild)
	assert.Equal(t, "{moniker}", dockerComp.DockerBuild.Container)
}

// Integration test: Expand a module using real templates
func TestExpandModuleTemplates_IntegrationWithRealTemplates(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := LoadModuleTemplates(repoRoot)
	require.NoError(t, err)

	mod := &Module{
		Moniker:    "test-container",
		Name:       "Test Container",
		Template:   "container",
		Components: make(ModuleComponents),
	}

	conventions := DefaultDiscoveryConventions()
	err = ExpandModuleFromTemplate(mod, cfg.Templates, conventions, repoRoot, "ready-to-release")
	require.NoError(t, err)

	// Verify expansion
	assert.Contains(t, mod.Components, "dockerfile")
	dockerComp := mod.Components["dockerfile"]
	require.NotNil(t, dockerComp)
	assert.Equal(t, "containers/test-container", dockerComp.Root)

	require.NotNil(t, dockerComp.DockerBuild)
	assert.Equal(t, "test-container", dockerComp.DockerBuild.Container)
	assert.Contains(t, dockerComp.DockerBuild.Tags, "ghcr.io/ready-to-release/test-container:latest")
}
