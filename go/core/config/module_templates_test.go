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

	err := ExpandModuleFromTemplate(mod, templates, nil, "", nil, nil)
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

	err := ExpandModuleFromTemplate(mod, templates, nil, "", nil, nil)
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

	err := ExpandModuleFromTemplate(mod, templates, nil, "", nil, nil)
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
	err := ExpandModuleFromTemplate(mod, templates, nil, "", vars, nil)
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

	err := ExpandModuleFromTemplate(mod, templates, nil, "", nil, nil)
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

	err := ExpandModuleFromTemplate(mod, nil, nil, "", nil, nil)
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
// Tests for Discovery Rule Resolution
// =============================================================================

// TestDiscoverList_ResolvesNamedRules verifies that template discover lists
// resolve to actual rules from the named rules map.
func TestDiscoverList_ResolvesNamedRules(t *testing.T) {
	namedRules := map[string]ComponentDiscoveryRule{
		"assets-from-go": {
			Component:      "assets",
			DeriveFromType: []string{"go", "typescript"},
			Check:          "dir_exists",
		},
		"gherkin-specs": {
			Component:    "gherkin",
			Path:         "{specs_root}/{moniker}",
			Check:        "has_files_in_subdirs",
			CheckPattern: "{specification}",
		},
	}

	rules := resolveDiscoverList([]string{"assets-from-go", "gherkin-specs"}, namedRules)
	require.Len(t, rules, 2)

	assert.Equal(t, "assets", rules[0].Component)
	assert.Equal(t, "gherkin", rules[1].Component)
}

// TestDiscoverList_UnknownRulesSkipped verifies that unknown rule names are skipped.
func TestDiscoverList_UnknownRulesSkipped(t *testing.T) {
	namedRules := map[string]ComponentDiscoveryRule{
		"assets-from-go": {Component: "assets"},
	}

	rules := resolveDiscoverList([]string{"assets-from-go", "nonexistent-rule"}, namedRules)
	require.Len(t, rules, 1)
	assert.Equal(t, "assets", rules[0].Component)
}

// TestDiscoverList_NilNamedRules verifies that nil rules map returns nil.
func TestDiscoverList_NilNamedRules(t *testing.T) {
	rules := resolveDiscoverList([]string{"assets-from-go"}, nil)
	assert.Nil(t, rules)
}

// TestDiscoverList_DiscoveryPassCreatesComponent verifies that a template's
// discover list creates components when filesystem checks pass.
func TestDiscoverList_DiscoveryPassCreatesComponent(t *testing.T) {
	repoRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, "go", "mylib", "specs"), 0755))
	writeFile(t, filepath.Join(repoRoot, "go", "mylib", "specs", "godog_test.go"), "package specs")

	namedRules := map[string]ComponentDiscoveryRule{
		"godog-from-go": {
			Component:      "godog",
			DeriveFromType: []string{"go"},
			DeriveSubdir:   "specs",
			FallbackPath:   "go/eac/specs/{moniker}",
			Check:          "required_file",
			RequiredFile:   "godog_test.go",
		},
	}

	templates := map[string]ModuleTemplate{
		"go-library": {
			Discover: []string{"godog-from-go"},
		},
	}

	mod := &Module{
		Moniker:  "mylib",
		Template: "go-library",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/mylib"},
		},
	}

	vars := map[string]string{"moniker": "mylib"}
	err := ExpandModuleFromTemplate(mod, templates, nil, repoRoot, vars, namedRules)
	require.NoError(t, err)

	assert.Contains(t, mod.Components, "godog")
	assert.Equal(t, "go/mylib/specs", mod.Components["godog"].Root)
}

// TestDiscoverList_DiscoveryFailSkipsComponent verifies that a discovery rule
// skips the component when its filesystem check fails.
func TestDiscoverList_DiscoveryFailSkipsComponent(t *testing.T) {
	repoRoot := t.TempDir()
	// Don't create the specs directory

	namedRules := map[string]ComponentDiscoveryRule{
		"godog-from-go": {
			Component:      "godog",
			DeriveFromType: []string{"go"},
			DeriveSubdir:   "specs",
			Check:          "required_file",
			RequiredFile:   "godog_test.go",
		},
	}

	templates := map[string]ModuleTemplate{
		"go-library": {
			Discover: []string{"godog-from-go"},
		},
	}

	mod := &Module{
		Moniker:  "mylib",
		Template: "go-library",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/mylib"},
		},
	}

	vars := map[string]string{"moniker": "mylib"}
	err := ExpandModuleFromTemplate(mod, templates, nil, repoRoot, vars, namedRules)
	require.NoError(t, err)

	assert.NotContains(t, mod.Components, "godog",
		"godog component should not be created when specs directory does not exist")
}

// TestDiscoverList_ModuleComponentNotOverwritten verifies that a discovery rule
// does not overwrite a user-declared component.
func TestDiscoverList_ModuleComponentNotOverwritten(t *testing.T) {
	repoRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, "go", "mymod"), 0755))

	namedRules := map[string]ComponentDiscoveryRule{
		"assets-from-go": {
			Component:      "assets",
			DeriveFromType: []string{"go"},
			Check:          "dir_exists",
		},
	}

	templates := map[string]ModuleTemplate{
		"go-library": {
			Discover: []string{"assets-from-go"},
		},
	}

	mod := &Module{
		Moniker:  "mymod",
		Template: "go-library",
		Components: ModuleComponents{
			"go":       &ComponentEntry{Root: "go/mymod"},
			"assets": &ComponentEntry{Root: "docs/custom"}, // User-declared
		},
	}

	vars := map[string]string{"moniker": "mymod"}
	err := ExpandModuleFromTemplate(mod, templates, nil, repoRoot, vars, namedRules)
	require.NoError(t, err)

	// User's assets root should NOT be overwritten by discovery
	assert.Equal(t, "docs/custom", mod.Components["assets"].Root)
}

// TestDiscoverList_SetPatternsFromDiscovery verifies that set_patterns copies
// discovery patterns to the component's Patterns.Source.
func TestDiscoverList_SetPatternsFromDiscovery(t *testing.T) {
	repoRoot := t.TempDir()
	containerDir := filepath.Join(repoRoot, "containers", "test-oci")
	require.NoError(t, os.MkdirAll(containerDir, 0755))
	writeFile(t, filepath.Join(containerDir, "config.txt"), "some config")

	namedRules := map[string]ComponentDiscoveryRule{
		"assets-from-container": {
			Component: "assets",
			Path:      "{containers_root}/{moniker}",
			Check:     "dir_exists",
		},
	}

	templates := map[string]ModuleTemplate{
		"container": {
			Discover: []string{"assets-from-container"},
		},
	}

	mod := &Module{
		Moniker:    "test-oci",
		Template:   "container",
		Components: make(ModuleComponents),
	}

	vars := map[string]string{
		"moniker":         "test-oci",
		"containers_root": "containers",
	}
	err := ExpandModuleFromTemplate(mod, templates, nil, repoRoot, vars, namedRules)
	require.NoError(t, err)

	assert.Contains(t, mod.Components, "assets")
	entry := mod.Components["assets"]
	assert.Equal(t, "containers/test-oci", entry.Root)
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
	err := ExpandModuleFromTemplate(mod, templates, nil, "", vars, nil)
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
	err := ExpandModuleFromTemplate(mod, templates, nil, "", vars, nil)
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

	err := ExpandModuleFromTemplate(mod, templates, nil, "", nil, nil)
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

	err := ExpandModuleFromTemplate(mod, templates, nil, "", nil, nil)
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

	// Verify go-library template has discover rules
	assert.Contains(t, goLib.Discover, "assets-from-go")
	assert.Contains(t, goLib.Discover, "gherkin-specs")
	assert.Contains(t, goLib.Discover, "structurizr-design")
	assert.Contains(t, goLib.Discover, "godog-from-go")
	assert.NotContains(t, goLib.Discover, "markdown-from-go", "old discovery rule should be removed")
	assert.NotContains(t, goLib.Discover, "yaml-from-go", "old discovery rule should be removed")

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

	// Verify script-installer has discover rules
	scriptInstaller := cfg.Templates["script-installer"]
	assert.Contains(t, scriptInstaller.Discover, "gherkin-specs")
	assert.Contains(t, scriptInstaller.Discover, "godog-from-go")

	// Verify component blueprints are loaded
	require.NotNil(t, cfg.ComponentBlueprints)
	assert.Contains(t, cfg.ComponentBlueprints, "structurizr-per-component")
	assert.Contains(t, cfg.ComponentBlueprints, "gherkin-per-component")
	assert.Contains(t, cfg.ComponentBlueprints, "godog-per-component")

	// Verify structurizr-per-component blueprint structure
	structBP := cfg.ComponentBlueprints["structurizr-per-component"]
	assert.Equal(t, "structurizr-{name}", structBP.Component)
	assert.Equal(t, "structurizr", structBP.Type)
	assert.Equal(t, "{specs_root}/{name}/{design_dir}", structBP.Root)
	require.NotNil(t, structBP.Discovery)
	assert.Equal(t, "required_file", structBP.Discovery.Check)

	// Verify go-library template has blueprints
	assert.Contains(t, goLib.Blueprints, "structurizr-per-component")
	assert.Contains(t, goLib.Blueprints, "gherkin-per-component")
	assert.Contains(t, goLib.Blueprints, "godog-per-component")

	// Verify go-contracts template has blueprints (no godog)
	goContracts := cfg.Templates["go-contracts"]
	assert.Contains(t, goContracts.Blueprints, "structurizr-per-component")
	assert.Contains(t, goContracts.Blueprints, "gherkin-per-component")
	assert.NotContains(t, goContracts.Blueprints, "godog-per-component")

	// Verify discovery rules are loaded
	require.NotNil(t, cfg.DiscoveryRules)
	assert.Contains(t, cfg.DiscoveryRules, "assets-from-go")
	assert.Contains(t, cfg.DiscoveryRules, "gherkin-specs")
	assert.Contains(t, cfg.DiscoveryRules, "structurizr-design")
	assert.Contains(t, cfg.DiscoveryRules, "godog-from-go")
	assert.Contains(t, cfg.DiscoveryRules, "assets-from-container")

	// Verify discovery rule structure
	mdRule := cfg.DiscoveryRules["assets-from-go"]
	assert.Equal(t, "assets", mdRule.Component)
	assert.Equal(t, []string{"go", "typescript", "dockerfile"}, mdRule.DeriveFromType)
	assert.Equal(t, "dir_exists", mdRule.Check)

	gherkinRule := cfg.DiscoveryRules["gherkin-specs"]
	assert.Equal(t, "gherkin", gherkinRule.Component)
	assert.Equal(t, "{specs_root}/{moniker}", gherkinRule.Path)
	assert.Equal(t, "has_files_in_subdirs", gherkinRule.Check)
}

// =============================================================================
// Tests for MergeBlueprintsConfig
// =============================================================================

func TestMergeBlueprintsConfig_BothNil(t *testing.T) {
	result := MergeBlueprintsConfig(nil, nil)
	require.NotNil(t, result)
	assert.Empty(t, result.ComponentBlueprints)
	assert.Empty(t, result.DiscoveryRules)
	assert.Empty(t, result.Templates)
	assert.Empty(t, result.ArtifactMatrices)
}

func TestMergeBlueprintsConfig_BaseNil(t *testing.T) {
	override := &BlueprintsConfig{
		DiscoveryRules: map[string]ComponentDiscoveryRule{
			"assets-from-go": {Component: "assets"},
		},
	}
	result := MergeBlueprintsConfig(nil, override)
	assert.Equal(t, override, result)
}

func TestMergeBlueprintsConfig_OverrideNil(t *testing.T) {
	base := &BlueprintsConfig{
		Templates: map[string]ModuleTemplate{
			"go-lib": {Discover: []string{"assets-from-go"}},
		},
	}
	result := MergeBlueprintsConfig(base, nil)
	assert.Equal(t, base, result)
}

func TestMergeBlueprintsConfig_ComponentBlueprints(t *testing.T) {
	base := &BlueprintsConfig{
		ComponentBlueprints: map[string]ComponentBlueprint{
			"structurizr-per-component": {Component: "structurizr-{name}", Type: "structurizr"},
		},
	}
	override := &BlueprintsConfig{
		ComponentBlueprints: map[string]ComponentBlueprint{
			"structurizr-per-component": {Component: "structurizr-{name}", Type: "custom"}, // Override
			"gherkin-per-component":     {Component: "gherkin-{name}", Type: "gherkin"},     // Add new
		},
	}

	result := MergeBlueprintsConfig(base, override)

	// Overridden blueprint
	assert.Equal(t, "custom", result.ComponentBlueprints["structurizr-per-component"].Type)
	// New blueprint added
	assert.Equal(t, "gherkin", result.ComponentBlueprints["gherkin-per-component"].Type)
}

func TestMergeBlueprintsConfig_OverrideKeys(t *testing.T) {
	base := &BlueprintsConfig{
		DiscoveryRules: map[string]ComponentDiscoveryRule{
			"assets-from-go": {Component: "assets", Check: "dir_exists"},
			"gherkin-specs":    {Component: "gherkin", Path: "{specs_root}/{moniker}"},
		},
		Templates: map[string]ModuleTemplate{
			"go-lib": {Discover: []string{"assets-from-go"}},
		},
		ArtifactMatrices: map[string]*ArtifactMatrix{
			"cross-platform": {Entries: []ArtifactMatrixEntry{{ID: "linux-amd64"}}},
		},
	}

	override := &BlueprintsConfig{
		DiscoveryRules: map[string]ComponentDiscoveryRule{
			"assets-from-go": {Component: "assets", Check: "required_file"}, // Override existing
			"new-rule":         {Component: "new"},                               // Add new
		},
		Templates: map[string]ModuleTemplate{
			"new-template": {Discover: []string{"new-rule"}}, // Add new
		},
		ArtifactMatrices: map[string]*ArtifactMatrix{
			"cross-platform": {Entries: []ArtifactMatrixEntry{{ID: "override"}}}, // Override existing
		},
	}

	result := MergeBlueprintsConfig(base, override)

	// Overridden rule
	assert.Equal(t, "required_file", result.DiscoveryRules["assets-from-go"].Check)
	// Preserved base rule
	assert.Equal(t, "gherkin", result.DiscoveryRules["gherkin-specs"].Component)
	// New rule added
	assert.Equal(t, "new", result.DiscoveryRules["new-rule"].Component)

	// Base template preserved
	assert.Contains(t, result.Templates, "go-lib")
	// New template added
	assert.Contains(t, result.Templates, "new-template")

	// Overridden matrix
	assert.Equal(t, "override", result.ArtifactMatrices["cross-platform"].Entries[0].ID)
}

func TestMergeBlueprintsConfig_EmptyMaps(t *testing.T) {
	base := &BlueprintsConfig{
		DiscoveryRules: map[string]ComponentDiscoveryRule{
			"assets-from-go": {Component: "assets"},
		},
	}
	override := &BlueprintsConfig{} // All maps nil

	result := MergeBlueprintsConfig(base, override)

	// Base should be preserved
	assert.Equal(t, "assets", result.DiscoveryRules["assets-from-go"].Component)
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

// =============================================================================
// Tests for Component Blueprints
// =============================================================================

// TestComponentBlueprints_StructurizrPerComponent creates temp workspace.dsl
// for 2 of 3 adapters and verifies only those get structurizr components.
func TestComponentBlueprints_StructurizrPerComponent(t *testing.T) {
	repoRoot := t.TempDir()

	// Create workspace.dsl for adapter-a and adapter-c, but NOT adapter-b
	writeFile(t, filepath.Join(repoRoot, "specs", "adapter-a", ".design", "workspace.dsl"), "workspace {}")
	writeFile(t, filepath.Join(repoRoot, "specs", "adapter-c", ".design", "workspace.dsl"), "workspace {}")

	blueprints := map[string]ComponentBlueprint{
		"structurizr-per-component": {
			Component: "structurizr-{name}",
			Type:      "structurizr",
			Root:      "{specs_root}/{name}/{design_dir}",
			Discovery: &BlueprintDiscovery{
				Check:        "required_file",
				RequiredFile: "{workspace_dsl}",
			},
		},
	}

	mod := &Module{
		Moniker: "adapters",
		Components: ModuleComponents{
			"adapter-a": &ComponentEntry{Root: "go/adapters/adapter-a"},
			"adapter-b": &ComponentEntry{Root: "go/adapters/adapter-b"},
			"adapter-c": &ComponentEntry{Root: "go/adapters/adapter-c"},
		},
	}

	vars := map[string]string{
		"moniker":       "adapters",
		"specs_root":    "specs",
		"design_dir":    ".design",
		"workspace_dsl": "workspace.dsl",
	}

	applyComponentBlueprints(mod, []string{"structurizr-per-component"}, blueprints, repoRoot, vars)

	// adapter-a and adapter-c should get structurizr components
	assert.Contains(t, mod.Components, "structurizr-adapter-a")
	assert.Equal(t, "specs/adapter-a/.design", mod.Components["structurizr-adapter-a"].Root)
	assert.Equal(t, "structurizr", mod.Components["structurizr-adapter-a"].Type)

	assert.Contains(t, mod.Components, "structurizr-adapter-c")
	assert.Equal(t, "specs/adapter-c/.design", mod.Components["structurizr-adapter-c"].Root)

	// adapter-b should NOT get a structurizr component (no workspace.dsl)
	assert.NotContains(t, mod.Components, "structurizr-adapter-b")

	// Original components should be preserved
	assert.Contains(t, mod.Components, "adapter-a")
	assert.Contains(t, mod.Components, "adapter-b")
	assert.Contains(t, mod.Components, "adapter-c")
}

// TestComponentBlueprints_GodogPerComponent verifies derive_from with {name} substitution.
func TestComponentBlueprints_GodogPerComponent(t *testing.T) {
	repoRoot := t.TempDir()

	// Create godog_test.go for adapter-a's specs directory
	writeFile(t, filepath.Join(repoRoot, "go", "adapters", "adapter-a", "specs", "godog_test.go"), "package specs")

	blueprints := map[string]ComponentBlueprint{
		"godog-per-component": {
			Component: "godog-{name}",
			Type:      "godog",
			Discovery: &BlueprintDiscovery{
				DeriveFrom:    []string{"{name}"},
				DeriveSubdirs: map[string]string{"{name}": "specs"},
				Check:         "required_file",
				RequiredFile:  "{godog_test}",
			},
		},
	}

	mod := &Module{
		Moniker: "adapters",
		Components: ModuleComponents{
			"adapter-a": &ComponentEntry{Root: "go/adapters/adapter-a"},
			"adapter-b": &ComponentEntry{Root: "go/adapters/adapter-b"},
		},
	}

	vars := map[string]string{
		"moniker":    "adapters",
		"godog_test": "godog_test.go",
	}

	applyComponentBlueprints(mod, []string{"godog-per-component"}, blueprints, repoRoot, vars)

	// adapter-a has godog_test.go → should get godog component
	assert.Contains(t, mod.Components, "godog-adapter-a")
	assert.Equal(t, "go/adapters/adapter-a/specs", mod.Components["godog-adapter-a"].Root)
	assert.Equal(t, "godog", mod.Components["godog-adapter-a"].Type)

	// adapter-b does NOT have godog_test.go → should NOT get godog component
	assert.NotContains(t, mod.Components, "godog-adapter-b")
}

// TestComponentBlueprints_DoesNotOverrideExisting verifies that a user-declared
// component is preserved and not overwritten by blueprint discovery.
func TestComponentBlueprints_DoesNotOverrideExisting(t *testing.T) {
	repoRoot := t.TempDir()

	// Create the file that would trigger discovery
	writeFile(t, filepath.Join(repoRoot, "specs", "mycomp", ".design", "workspace.dsl"), "workspace {}")

	blueprints := map[string]ComponentBlueprint{
		"structurizr-per-component": {
			Component: "structurizr-{name}",
			Type:      "structurizr",
			Root:      "{specs_root}/{name}/{design_dir}",
			Discovery: &BlueprintDiscovery{
				Check:        "required_file",
				RequiredFile: "{workspace_dsl}",
			},
		},
	}

	mod := &Module{
		Moniker: "mymod",
		Components: ModuleComponents{
			"mycomp": &ComponentEntry{Root: "go/mymod/mycomp"},
			// User already declared this component with a custom root
			"structurizr-mycomp": &ComponentEntry{Root: "custom/path/.design"},
		},
	}

	vars := map[string]string{
		"specs_root":    "specs",
		"design_dir":    ".design",
		"workspace_dsl": "workspace.dsl",
	}

	applyComponentBlueprints(mod, []string{"structurizr-per-component"}, blueprints, repoRoot, vars)

	// User-declared component should NOT be overwritten
	assert.Equal(t, "custom/path/.design", mod.Components["structurizr-mycomp"].Root)
}

// TestComponentBlueprints_SingleComponentModuleNoConflict verifies that a
// single-component module (like "core" with component "go") does not get
// spurious auxiliary components when the filesystem check fails.
func TestComponentBlueprints_SingleComponentModuleNoConflict(t *testing.T) {
	repoRoot := t.TempDir()

	// Module-level discovery finds specs/core/.design/ (via moniker).
	// Blueprint would check specs/go/.design/ (component name = "go") which doesn't exist.
	// No spurious component should be created.

	blueprints := map[string]ComponentBlueprint{
		"structurizr-per-component": {
			Component: "structurizr-{name}",
			Type:      "structurizr",
			Root:      "{specs_root}/{name}/{design_dir}",
			Discovery: &BlueprintDiscovery{
				Check:        "required_file",
				RequiredFile: "{workspace_dsl}",
			},
		},
	}

	mod := &Module{
		Moniker: "core",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/core"},
		},
	}

	vars := map[string]string{
		"moniker":       "core",
		"specs_root":    "specs",
		"design_dir":    ".design",
		"workspace_dsl": "workspace.dsl",
	}

	applyComponentBlueprints(mod, []string{"structurizr-per-component"}, blueprints, repoRoot, vars)

	// No structurizr-go component should be created (specs/go/.design/ doesn't exist)
	assert.NotContains(t, mod.Components, "structurizr-go")

	// Original component should be preserved
	assert.Contains(t, mod.Components, "go")
	assert.Equal(t, "go/core", mod.Components["go"].Root)
}

// TestComponentBlueprints_EmptyAndUnknown verifies that nil/empty/unknown
// blueprint names are silently skipped as no-ops.
func TestComponentBlueprints_EmptyAndUnknown(t *testing.T) {
	repoRoot := t.TempDir()

	blueprints := map[string]ComponentBlueprint{
		"known-blueprint": {
			Component: "aux-{name}",
			Root:      "some/{name}",
		},
	}

	mod := &Module{
		Moniker: "test",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/test"},
		},
	}

	// Test with nil blueprints map
	applyComponentBlueprints(mod, []string{"anything"}, nil, repoRoot, nil)
	assert.Len(t, mod.Components, 1, "nil blueprints map should be a no-op")

	// Test with empty blueprint names
	applyComponentBlueprints(mod, nil, blueprints, repoRoot, nil)
	assert.Len(t, mod.Components, 1, "empty blueprint names should be a no-op")

	// Test with unknown blueprint name
	applyComponentBlueprints(mod, []string{"nonexistent-blueprint"}, blueprints, repoRoot, nil)
	assert.Len(t, mod.Components, 1, "unknown blueprint name should be a no-op")
}

// TestComponentBlueprints_GherkinPerComponent verifies has_files_in_subdirs check
// for per-component gherkin discovery.
func TestComponentBlueprints_GherkinPerComponent(t *testing.T) {
	repoRoot := t.TempDir()

	// Create gherkin specs for adapter-a: specs/adapter-a/some-feature/specification.feature
	writeFile(t, filepath.Join(repoRoot, "specs", "adapter-a", "login", "specification.feature"),
		"Feature: Login")

	blueprints := map[string]ComponentBlueprint{
		"gherkin-per-component": {
			Component: "gherkin-{name}",
			Type:      "gherkin",
			Root:      "{specs_root}/{name}",
			Discovery: &BlueprintDiscovery{
				Check:        "has_files_in_subdirs",
				CheckPattern: "{specification}",
			},
		},
	}

	mod := &Module{
		Moniker: "adapters",
		Components: ModuleComponents{
			"adapter-a": &ComponentEntry{Root: "go/adapters/adapter-a"},
			"adapter-b": &ComponentEntry{Root: "go/adapters/adapter-b"},
		},
	}

	vars := map[string]string{
		"moniker":       "adapters",
		"specs_root":    "specs",
		"specification": "specification.feature",
	}

	applyComponentBlueprints(mod, []string{"gherkin-per-component"}, blueprints, repoRoot, vars)

	// adapter-a has specs → should get gherkin component
	assert.Contains(t, mod.Components, "gherkin-adapter-a")
	assert.Equal(t, "specs/adapter-a", mod.Components["gherkin-adapter-a"].Root)
	assert.Equal(t, "gherkin", mod.Components["gherkin-adapter-a"].Type)

	// adapter-b has no specs → should NOT get gherkin component
	assert.NotContains(t, mod.Components, "gherkin-adapter-b")
}
