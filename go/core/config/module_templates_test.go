//go:build L0 && ov

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleTemplate_BasicExpansion(t *testing.T) {
	templates := map[string]ModuleTemplate{
		"container-template": {
			Components: ModuleComponents{
				"dockerfile": &ComponentEntry{Root: "containers/{moniker}"},
				"markdown":   &ComponentEntry{Root: "containers/{moniker}"},
			},
		},
	}

	mod := &Module{
		Moniker:    "test-container",
		Name:       "Test Container",
		Template:   "container-template",
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
		"go-library-template": {
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
		Template: "go-library-template",
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
		"go-cli-template": {
			Versioning: &ModuleVersioning{
				Scheme:      "SemVer",
				Changelog:   "release/{moniker}/CHANGELOG.md",
				ReleaseType: "published",
			},
		},
	}

	mod := &Module{
		Moniker:  "my-cli",
		Template: "go-cli-template",
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
		"container-template": {
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
		Template:   "container-template",
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
		"go-library-template": {
			DependsOn: []string{"contracts", "core"},
		},
	}

	mod := &Module{
		Moniker:   "my-lib",
		Template:  "go-library-template",
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
		"go-library-template": {
			Components: ModuleComponents{
				"markdown": &ComponentEntry{Root: "{go_root}"},
				"gherkin":  &ComponentEntry{Root: "specs/{moniker}"},
			},
		},
	}

	mod := &Module{
		Moniker:  "my-lib",
		Template: "go-library-template",
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
		"container-template": {
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
		Template: "container-template",
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
		"go-library-template": {
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
		Template: "go-library-template",
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
	assert.Contains(t, cfg.Templates, "go-library-template")
	assert.Contains(t, cfg.Templates, "go-cli-template")
	assert.Contains(t, cfg.Templates, "container-template")
	assert.Contains(t, cfg.Templates, "container-multiarch-template")
	assert.Contains(t, cfg.Templates, "docs-site-template")

	// Verify go-library template structure
	goLib := cfg.Templates["go-library-template"]
	require.NotNil(t, goLib.Versioning)
	assert.Equal(t, "Implicit", goLib.Versioning.Scheme)
	assert.Contains(t, goLib.Components, "go")
	assert.Contains(t, goLib.Components, "markdown")

	// Verify container template structure
	container := cfg.Templates["container-template"]
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
		Template:   "container-template",
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

// =============================================================================
// Container Auxiliary Components Discovery Tests
// =============================================================================

// TestDiscoverContainerAuxiliaryComponents_TestdataDiscovery verifies that testdata
// components are discovered when container directories contain matching files.
func TestDiscoverContainerAuxiliaryComponents_TestdataDiscovery(t *testing.T) {
	tests := []struct {
		name     string
		files    []string // Files to create in containers/{moniker}/
		wantComp bool     // Should testdata component be discovered
	}{
		{
			name:     "discovers testdata with .txt files",
			files:    []string{"config.txt", "sample.txt"},
			wantComp: true,
		},
		{
			name:     "discovers testdata with .conf files",
			files:    []string{"nginx.conf", "app.conf"},
			wantComp: true,
		},
		{
			name:     "discovers testdata with .html files",
			files:    []string{"index.html", "template.html"},
			wantComp: true,
		},
		{
			name:     "discovers testdata with .sh files",
			files:    []string{"entrypoint.sh", "setup.sh"},
			wantComp: true,
		},
		{
			name:     "discovers testdata with mixed matching files",
			files:    []string{"config.txt", "nginx.conf", "entrypoint.sh"},
			wantComp: true,
		},
		{
			name:     "no testdata with only Dockerfile",
			files:    []string{"Dockerfile"},
			wantComp: false,
		},
		{
			name:     "no testdata with only non-matching files",
			files:    []string{"Dockerfile", "go.mod", "mkdocs.yml"},
			wantComp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			containerDir := filepath.Join(repoRoot, "containers", "test-oci")
			require.NoError(t, os.MkdirAll(containerDir, 0755))

			// Create test files
			for _, f := range tt.files {
				require.NoError(t, os.WriteFile(
					filepath.Join(containerDir, f),
					[]byte("test content"),
					0644,
				))
			}

			mod := &Module{
				Moniker:    "test-oci",
				Template:   "container-template",
				Components: make(ModuleComponents),
			}

			discoverContainerAuxiliaryComponents(mod, repoRoot)

			if tt.wantComp {
				assert.Contains(t, mod.Components, "testdata",
					"expected testdata component to be discovered")
				if comp, ok := mod.Components["testdata"]; ok {
					assert.Equal(t, "containers/test-oci", comp.Root)
					assert.NotNil(t, comp.Patterns)
					assert.Contains(t, comp.Patterns.Source, "*.txt")
					assert.Contains(t, comp.Patterns.Source, "*.conf")
					assert.Contains(t, comp.Patterns.Source, "*.html")
					assert.Contains(t, comp.Patterns.Source, "*.sh")
				}
			} else {
				assert.NotContains(t, mod.Components, "testdata",
					"expected testdata component NOT to be discovered")
			}
		})
	}
}

// TestDiscoverContainerAuxiliaryComponents_ScriptsDiscovery verifies that containercode
// components are discovered when container directories contain Python or JavaScript files.
func TestDiscoverContainerAuxiliaryComponents_ScriptsDiscovery(t *testing.T) {
	tests := []struct {
		name     string
		files    []string // Files to create in containers/{moniker}/
		wantComp bool     // Should containercode component be discovered
	}{
		{
			name:     "discovers containercode with .py files",
			files:    []string{"mkdocs_macros.py", "helpers.py"},
			wantComp: true,
		},
		{
			name:     "discovers containercode with .js files",
			files:    []string{"batch-render.js", "utils.js"},
			wantComp: true,
		},
		{
			name:     "discovers containercode with mixed .py and .js files",
			files:    []string{"main.py", "helper.js"},
			wantComp: true,
		},
		{
			name:     "no containercode with only Dockerfile",
			files:    []string{"Dockerfile"},
			wantComp: false,
		},
		{
			name:     "no containercode with only config files",
			files:    []string{"Dockerfile", "mkdocs.yml", "requirements.txt"},
			wantComp: false,
		},
		{
			name:     "no containercode with only testdata files",
			files:    []string{"Dockerfile", "config.txt", "nginx.conf"},
			wantComp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			containerDir := filepath.Join(repoRoot, "containers", "render-oci")
			require.NoError(t, os.MkdirAll(containerDir, 0755))

			// Create test files
			for _, f := range tt.files {
				require.NoError(t, os.WriteFile(
					filepath.Join(containerDir, f),
					[]byte("test content"),
					0644,
				))
			}

			mod := &Module{
				Moniker:    "render-oci",
				Template:   "container-template",
				Components: make(ModuleComponents),
			}

			discoverContainerAuxiliaryComponents(mod, repoRoot)

			if tt.wantComp {
				assert.Contains(t, mod.Components, "containercode",
					"expected containercode component to be discovered")
				if comp, ok := mod.Components["containercode"]; ok {
					assert.Equal(t, "containers/render-oci", comp.Root)
					assert.NotNil(t, comp.Patterns)
					assert.Contains(t, comp.Patterns.Source, "*.py")
					assert.Contains(t, comp.Patterns.Source, "*.js")
				}
			} else {
				assert.NotContains(t, mod.Components, "containercode",
					"expected containercode component NOT to be discovered")
			}
		})
	}
}

// TestDiscoverContainerAuxiliaryComponents_NoOverride verifies that existing
// component definitions are not overwritten by auto-discovery.
func TestDiscoverContainerAuxiliaryComponents_NoOverride(t *testing.T) {
	t.Run("does not override existing testdata component", func(t *testing.T) {
		repoRoot := t.TempDir()
		containerDir := filepath.Join(repoRoot, "containers", "test-oci")
		require.NoError(t, os.MkdirAll(containerDir, 0755))

		// Create files that would trigger testdata discovery
		require.NoError(t, os.WriteFile(
			filepath.Join(containerDir, "config.txt"),
			[]byte("test"),
			0644,
		))

		mod := &Module{
			Moniker:  "test-oci",
			Template: "container-template",
			Components: ModuleComponents{
				"testdata": &ComponentEntry{
					Root: "custom/testdata/path",
					Patterns: &ComponentPatterns{
						Source: []string{"custom/*.data"},
					},
				},
			},
		}

		discoverContainerAuxiliaryComponents(mod, repoRoot)

		// Existing component should NOT be overwritten
		assert.Equal(t, "custom/testdata/path", mod.Components["testdata"].Root)
		assert.Equal(t, []string{"custom/*.data"}, mod.Components["testdata"].Patterns.Source)
	})

	t.Run("does not override existing containercode component", func(t *testing.T) {
		repoRoot := t.TempDir()
		containerDir := filepath.Join(repoRoot, "containers", "render-oci")
		require.NoError(t, os.MkdirAll(containerDir, 0755))

		// Create files that would trigger containercode discovery
		require.NoError(t, os.WriteFile(
			filepath.Join(containerDir, "script.py"),
			[]byte("test"),
			0644,
		))

		mod := &Module{
			Moniker:  "render-oci",
			Template: "container-template",
			Components: ModuleComponents{
				"containercode": &ComponentEntry{
					Root: "scripts/custom",
					Patterns: &ComponentPatterns{
						Source: []string{"**/*.custom"},
					},
				},
			},
		}

		discoverContainerAuxiliaryComponents(mod, repoRoot)

		// Existing component should NOT be overwritten
		assert.Equal(t, "scripts/custom", mod.Components["containercode"].Root)
		assert.Equal(t, []string{"**/*.custom"}, mod.Components["containercode"].Patterns.Source)
	})

	t.Run("does not override nil component marker", func(t *testing.T) {
		repoRoot := t.TempDir()
		containerDir := filepath.Join(repoRoot, "containers", "test-oci")
		require.NoError(t, os.MkdirAll(containerDir, 0755))

		// Create files that would trigger testdata discovery
		require.NoError(t, os.WriteFile(
			filepath.Join(containerDir, "config.txt"),
			[]byte("test"),
			0644,
		))

		mod := &Module{
			Moniker:  "test-oci",
			Template: "container-template",
			Components: ModuleComponents{
				"testdata": nil, // Explicit nil marker means "use convention"
			},
		}

		discoverContainerAuxiliaryComponents(mod, repoRoot)

		// nil marker should be treated as "component exists" - no override
		assert.Nil(t, mod.Components["testdata"])
	})
}

// TestDiscoverContainerAuxiliaryComponents_NoMatchingFiles verifies that components
// are not added when no matching files exist in the container directory.
func TestDiscoverContainerAuxiliaryComponents_NoMatchingFiles(t *testing.T) {
	t.Run("no components discovered when directory is empty", func(t *testing.T) {
		repoRoot := t.TempDir()
		containerDir := filepath.Join(repoRoot, "containers", "empty-oci")
		require.NoError(t, os.MkdirAll(containerDir, 0755))

		mod := &Module{
			Moniker:    "empty-oci",
			Template:   "container-template",
			Components: make(ModuleComponents),
		}

		discoverContainerAuxiliaryComponents(mod, repoRoot)

		assert.NotContains(t, mod.Components, "testdata")
		assert.NotContains(t, mod.Components, "containercode")
	})

	t.Run("no components discovered when only Dockerfile exists", func(t *testing.T) {
		repoRoot := t.TempDir()
		containerDir := filepath.Join(repoRoot, "containers", "minimal-oci")
		require.NoError(t, os.MkdirAll(containerDir, 0755))

		require.NoError(t, os.WriteFile(
			filepath.Join(containerDir, "Dockerfile"),
			[]byte("FROM alpine"),
			0644,
		))

		mod := &Module{
			Moniker:    "minimal-oci",
			Template:   "container-template",
			Components: make(ModuleComponents),
		}

		discoverContainerAuxiliaryComponents(mod, repoRoot)

		assert.NotContains(t, mod.Components, "testdata")
		assert.NotContains(t, mod.Components, "containercode")
	})

	t.Run("no components discovered when container directory does not exist", func(t *testing.T) {
		repoRoot := t.TempDir()
		// Don't create container directory

		mod := &Module{
			Moniker:    "nonexistent-oci",
			Template:   "container-template",
			Components: make(ModuleComponents),
		}

		discoverContainerAuxiliaryComponents(mod, repoRoot)

		assert.NotContains(t, mod.Components, "testdata")
		assert.NotContains(t, mod.Components, "containercode")
	})

	t.Run("no components discovered with only yaml and json files", func(t *testing.T) {
		repoRoot := t.TempDir()
		containerDir := filepath.Join(repoRoot, "containers", "config-oci")
		require.NoError(t, os.MkdirAll(containerDir, 0755))

		// Only create files that should NOT match testdata or containercode patterns
		files := []string{"Dockerfile", "mkdocs.yml", "config.json"}
		for _, f := range files {
			require.NoError(t, os.WriteFile(
				filepath.Join(containerDir, f),
				[]byte("content"),
				0644,
			))
		}

		mod := &Module{
			Moniker:    "config-oci",
			Template:   "container-template",
			Components: make(ModuleComponents),
		}

		discoverContainerAuxiliaryComponents(mod, repoRoot)

		assert.NotContains(t, mod.Components, "testdata")
		assert.NotContains(t, mod.Components, "containercode")
	})

	t.Run("requirements.txt does trigger testdata discovery", func(t *testing.T) {
		// Note: This documents current behavior. requirements.txt matches *.txt pattern.
		// If this is undesired, the pattern should be changed to exclude it.
		repoRoot := t.TempDir()
		containerDir := filepath.Join(repoRoot, "containers", "py-oci")
		require.NoError(t, os.MkdirAll(containerDir, 0755))

		require.NoError(t, os.WriteFile(
			filepath.Join(containerDir, "requirements.txt"),
			[]byte("flask==2.0"),
			0644,
		))

		mod := &Module{
			Moniker:    "py-oci",
			Template:   "container-template",
			Components: make(ModuleComponents),
		}

		discoverContainerAuxiliaryComponents(mod, repoRoot)

		// requirements.txt matches *.txt - this is current behavior
		// Document it explicitly so we know this is intentional
		assert.Contains(t, mod.Components, "testdata",
			"requirements.txt matches *.txt pattern - this is expected behavior")
	})
}

// TestDiscoverContainerAuxiliaryComponents_NonContainerTemplate verifies that
// auxiliary component discovery only applies to container templates.
func TestDiscoverContainerAuxiliaryComponents_NonContainerTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantSkip bool // Should discovery be skipped
	}{
		{
			name:     "container template triggers discovery",
			template: "container-template",
			wantSkip: false,
		},
		{
			name:     "container-multiarch template triggers discovery",
			template: "container-multiarch-template",
			wantSkip: false,
		},
		{
			name:     "go-library template skips discovery",
			template: "go-library-template",
			wantSkip: true,
		},
		{
			name:     "go-cli template skips discovery",
			template: "go-cli-template",
			wantSkip: true,
		},
		{
			name:     "documentation-site template skips discovery",
			template: "docs-site-template",
			wantSkip: true,
		},
		{
			name:     "empty template skips discovery",
			template: "",
			wantSkip: true,
		},
		{
			name:     "custom-container template triggers discovery",
			template: "custom-container",
			wantSkip: false,
		},
		{
			name:     "my-container-build template triggers discovery",
			template: "my-container-build",
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			containerDir := filepath.Join(repoRoot, "containers", "test-oci")
			require.NoError(t, os.MkdirAll(containerDir, 0755))

			// Create files that would trigger discovery
			require.NoError(t, os.WriteFile(
				filepath.Join(containerDir, "config.txt"),
				[]byte("test"),
				0644,
			))
			require.NoError(t, os.WriteFile(
				filepath.Join(containerDir, "script.py"),
				[]byte("test"),
				0644,
			))

			mod := &Module{
				Moniker:    "test-oci",
				Template:   tt.template,
				Components: make(ModuleComponents),
			}

			discoverContainerAuxiliaryComponents(mod, repoRoot)

			if tt.wantSkip {
				assert.NotContains(t, mod.Components, "testdata",
					"non-container template should skip testdata discovery")
				assert.NotContains(t, mod.Components, "containercode",
					"non-container template should skip containercode discovery")
			} else {
				assert.Contains(t, mod.Components, "testdata",
					"container template should discover testdata")
				assert.Contains(t, mod.Components, "containercode",
					"container template should discover containercode")
			}
		})
	}
}

// TestDiscoverContainerAuxiliaryComponents_BothComponentsDiscovered verifies that
// both testdata and containercode can be discovered from the same container.
func TestDiscoverContainerAuxiliaryComponents_BothComponentsDiscovered(t *testing.T) {
	repoRoot := t.TempDir()
	containerDir := filepath.Join(repoRoot, "containers", "full-oci")
	require.NoError(t, os.MkdirAll(containerDir, 0755))

	// Create files for both components
	testdataFiles := []string{"entrypoint.sh", "nginx.conf", "index.html"}
	scriptFiles := []string{"mkdocs_macros.py", "batch-render.js"}
	otherFiles := []string{"Dockerfile", "requirements.txt", "mkdocs.yml"}

	for _, f := range append(append(testdataFiles, scriptFiles...), otherFiles...) {
		require.NoError(t, os.WriteFile(
			filepath.Join(containerDir, f),
			[]byte("content"),
			0644,
		))
	}

	mod := &Module{
		Moniker:    "full-oci",
		Template:   "container-template",
		Components: make(ModuleComponents),
	}

	discoverContainerAuxiliaryComponents(mod, repoRoot)

	// Both components should be discovered
	assert.Contains(t, mod.Components, "testdata")
	assert.Contains(t, mod.Components, "containercode")

	// Verify testdata component
	testdataComp := mod.Components["testdata"]
	require.NotNil(t, testdataComp)
	assert.Equal(t, "containers/full-oci", testdataComp.Root)

	// Verify containercode component
	codeComp := mod.Components["containercode"]
	require.NotNil(t, codeComp)
	assert.Equal(t, "containers/full-oci", codeComp.Root)
}

// TestDiscoverContainerAuxiliaryComponents_FilesInSubdirectories verifies that
// files in subdirectories are also considered for discovery.
func TestDiscoverContainerAuxiliaryComponents_FilesInSubdirectories(t *testing.T) {
	t.Run("discovers testdata from subdirectory", func(t *testing.T) {
		repoRoot := t.TempDir()
		containerDir := filepath.Join(repoRoot, "containers", "nested-oci")
		subDir := filepath.Join(containerDir, "config")
		require.NoError(t, os.MkdirAll(subDir, 0755))

		// Create file in subdirectory
		require.NoError(t, os.WriteFile(
			filepath.Join(subDir, "app.conf"),
			[]byte("test"),
			0644,
		))

		mod := &Module{
			Moniker:    "nested-oci",
			Template:   "container-template",
			Components: make(ModuleComponents),
		}

		discoverContainerAuxiliaryComponents(mod, repoRoot)

		assert.Contains(t, mod.Components, "testdata",
			"should discover testdata from subdirectory files")
	})

	t.Run("discovers containercode from subdirectory", func(t *testing.T) {
		repoRoot := t.TempDir()
		containerDir := filepath.Join(repoRoot, "containers", "nested-oci")
		subDir := filepath.Join(containerDir, "scripts")
		require.NoError(t, os.MkdirAll(subDir, 0755))

		// Create file in subdirectory
		require.NoError(t, os.WriteFile(
			filepath.Join(subDir, "helper.py"),
			[]byte("test"),
			0644,
		))

		mod := &Module{
			Moniker:    "nested-oci",
			Template:   "container-template",
			Components: make(ModuleComponents),
		}

		discoverContainerAuxiliaryComponents(mod, repoRoot)

		assert.Contains(t, mod.Components, "containercode",
			"should discover containercode from subdirectory files")
	})
}

// TestDiscoverContainerAuxiliaryComponents_NilComponentsMap verifies that
// discovery handles nil Components map gracefully.
func TestDiscoverContainerAuxiliaryComponents_NilComponentsMap(t *testing.T) {
	repoRoot := t.TempDir()
	containerDir := filepath.Join(repoRoot, "containers", "nil-map-oci")
	require.NoError(t, os.MkdirAll(containerDir, 0755))

	require.NoError(t, os.WriteFile(
		filepath.Join(containerDir, "config.txt"),
		[]byte("test"),
		0644,
	))

	mod := &Module{
		Moniker:    "nil-map-oci",
		Template:   "container-template",
		Components: nil, // nil map
	}

	// Should not panic
	discoverContainerAuxiliaryComponents(mod, repoRoot)

	// Components map should be initialized and populated
	require.NotNil(t, mod.Components)
	assert.Contains(t, mod.Components, "testdata")
}
