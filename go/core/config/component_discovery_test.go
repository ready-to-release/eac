package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFile creates a file with the given content, creating parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}

// TestDiscoverComponents_DirExists verifies the dir_exists check mode.
func TestDiscoverComponents_DirExists(t *testing.T) {
	repoRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, "go", "mylib"), 0755))

	rules := []ComponentDiscoveryRule{
		{
			Component:  "assets",
			DeriveFrom: []string{"go"},
			Check:      "dir_exists",
		},
	}

	mod := &Module{
		Moniker: "mylib",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/mylib"},
		},
	}

	vars := map[string]string{"moniker": "mylib"}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "assets")
	assert.Equal(t, "go/mylib", mod.Components["assets"].Root)
}

// TestDiscoverComponents_RequiredFile verifies the required_file check mode.
func TestDiscoverComponents_RequiredFile(t *testing.T) {
	repoRoot := t.TempDir()
	designDir := filepath.Join(repoRoot, "specs", "mymod", ".design")
	require.NoError(t, os.MkdirAll(designDir, 0755))
	writeFile(t, filepath.Join(designDir, "workspace.dsl"), "workspace {}")

	rules := []ComponentDiscoveryRule{
		{
			Component:    "structurizr",
			Path:         "{specs_root}/{moniker}/{design_dir}",
			Check:        "required_file",
			RequiredFile: "{workspace_dsl}",
		},
	}

	mod := &Module{
		Moniker:    "mymod",
		Components: make(ModuleComponents),
	}

	vars := map[string]string{
		"moniker":       "mymod",
		"specs_root":    "specs",
		"design_dir":    ".design",
		"workspace_dsl": "workspace.dsl",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "structurizr")
	assert.Equal(t, "specs/mymod/.design", mod.Components["structurizr"].Root)
}

// TestDiscoverComponents_HasFilesInSubdirs verifies the has_files_in_subdirs check mode.
func TestDiscoverComponents_HasFilesInSubdirs(t *testing.T) {
	repoRoot := t.TempDir()
	featureDir := filepath.Join(repoRoot, "specs", "core", "cache-invalidation")
	require.NoError(t, os.MkdirAll(featureDir, 0755))
	writeFile(t, filepath.Join(featureDir, "specification.feature"), "Feature: Cache")

	rules := []ComponentDiscoveryRule{
		{
			Component:    "gherkin",
			Path:         "{specs_root}/{moniker}",
			Check:        "has_files_in_subdirs",
			CheckPattern: "{specification}",
		},
	}

	mod := &Module{
		Moniker:    "core",
		Components: make(ModuleComponents),
	}

	vars := map[string]string{
		"moniker":       "core",
		"specs_root":    "specs",
		"specification": "specification.feature",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "gherkin")
	assert.Equal(t, "specs/core", mod.Components["gherkin"].Root)
}

// TestDiscoverComponents_HasFilesInSubdirs_NoFiles verifies that has_files_in_subdirs
// returns false when no matching files exist in subdirectories.
func TestDiscoverComponents_HasFilesInSubdirs_NoFiles(t *testing.T) {
	repoRoot := t.TempDir()
	// Create the specs dir but no subdirectory with the feature file
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, "specs", "empty"), 0755))

	rules := []ComponentDiscoveryRule{
		{
			Component:    "gherkin",
			Path:         "{specs_root}/{moniker}",
			Check:        "has_files_in_subdirs",
			CheckPattern: "{specification}",
		},
	}

	mod := &Module{
		Moniker:    "empty",
		Components: make(ModuleComponents),
	}

	vars := map[string]string{
		"moniker":       "empty",
		"specs_root":    "specs",
		"specification": "specification.feature",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.NotContains(t, mod.Components, "gherkin")
}

// TestDiscoverComponents_HasMatchingFiles verifies the has_matching_files check mode.
func TestDiscoverComponents_HasMatchingFiles(t *testing.T) {
	repoRoot := t.TempDir()
	containerDir := filepath.Join(repoRoot, "containers", "test-oci")
	require.NoError(t, os.MkdirAll(containerDir, 0755))
	writeFile(t, filepath.Join(containerDir, "config.txt"), "some config")

	rules := []ComponentDiscoveryRule{
		{
			Component:     "testdata",
			OnlyTemplates: []string{"*container*"},
			Path:          "{containers_root}/{moniker}",
			Check:         "has_matching_files",
			Patterns:      []string{"*.txt", "*.conf", "*.html", "*.sh"},
			SetPatterns:   true,
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
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "testdata")
	entry := mod.Components["testdata"]
	assert.Equal(t, "containers/test-oci", entry.Root)
	require.NotNil(t, entry.Patterns)
	assert.Equal(t, []string{"*.txt", "*.conf", "*.html", "*.sh"}, entry.Patterns.Source)
}

// TestDiscoverComponents_TemplateFilter verifies that only_templates filters correctly.
func TestDiscoverComponents_TemplateFilter(t *testing.T) {
	repoRoot := t.TempDir()
	containerDir := filepath.Join(repoRoot, "containers", "mylib")
	require.NoError(t, os.MkdirAll(containerDir, 0755))
	writeFile(t, filepath.Join(containerDir, "script.py"), "print('hello')")

	rules := []ComponentDiscoveryRule{
		{
			Component:     "assets",
			OnlyTemplates: []string{"*container*"},
			Path:          "{containers_root}/{moniker}",
			Check:         "dir_exists",
		},
	}

	// Non-container template should be skipped
	mod := &Module{
		Moniker:    "mylib",
		Template:   "go-library",
		Components: make(ModuleComponents),
	}

	vars := map[string]string{
		"moniker":         "mylib",
		"containers_root": "containers",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.NotContains(t, mod.Components, "assets",
		"rule with only_templates=[*container*] should skip go-library")
}

// TestDiscoverComponents_TemplateFilter_Matches verifies that matching template passes filter.
func TestDiscoverComponents_TemplateFilter_Matches(t *testing.T) {
	repoRoot := t.TempDir()
	containerDir := filepath.Join(repoRoot, "containers", "test-oci")
	require.NoError(t, os.MkdirAll(containerDir, 0755))
	writeFile(t, filepath.Join(containerDir, "script.py"), "print('hello')")

	rules := []ComponentDiscoveryRule{
		{
			Component:     "assets",
			OnlyTemplates: []string{"*container*"},
			Path:          "{containers_root}/{moniker}",
			Check:         "dir_exists",
		},
	}

	mod := &Module{
		Moniker:    "test-oci",
		Template:   "container-multiarch",
		Components: make(ModuleComponents),
	}

	vars := map[string]string{
		"moniker":         "test-oci",
		"containers_root": "containers",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "assets",
		"container-multiarch should match *container*")
}

// TestDiscoverComponents_NilMarkerResolved verifies nil component markers get resolved.
func TestDiscoverComponents_NilMarkerResolved(t *testing.T) {
	repoRoot := t.TempDir()
	specsDir := filepath.Join(repoRoot, "specs", "mymod", "feature1")
	require.NoError(t, os.MkdirAll(specsDir, 0755))
	writeFile(t, filepath.Join(specsDir, "specification.feature"), "Feature: Test")

	rules := []ComponentDiscoveryRule{
		{
			Component:    "gherkin",
			Path:         "{specs_root}/{moniker}",
			Check:        "has_files_in_subdirs",
			CheckPattern: "{specification}",
		},
	}

	mod := &Module{
		Moniker: "mymod",
		Components: ModuleComponents{
			"gherkin": nil, // nil marker
		},
	}

	vars := map[string]string{
		"moniker":       "mymod",
		"specs_root":    "specs",
		"specification": "specification.feature",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	require.NotNil(t, mod.Components["gherkin"])
	assert.Equal(t, "specs/mymod", mod.Components["gherkin"].Root)
}

// TestDiscoverComponents_NilMarkerRemoved verifies nil markers are removed when path doesn't exist.
func TestDiscoverComponents_NilMarkerRemoved(t *testing.T) {
	repoRoot := t.TempDir()
	// Don't create any specs directory

	rules := []ComponentDiscoveryRule{
		{
			Component:    "gherkin",
			Path:         "{specs_root}/{moniker}",
			Check:        "has_files_in_subdirs",
			CheckPattern: "{specification}",
		},
	}

	mod := &Module{
		Moniker: "mymod",
		Components: ModuleComponents{
			"gherkin": nil, // nil marker for non-existent path
		},
	}

	vars := map[string]string{
		"moniker":       "mymod",
		"specs_root":    "specs",
		"specification": "specification.feature",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.NotContains(t, mod.Components, "gherkin",
		"nil marker should be removed when path doesn't exist")
}

// TestDiscoverComponents_DoesNotOverrideExisting verifies existing components are not overwritten.
func TestDiscoverComponents_DoesNotOverrideExisting(t *testing.T) {
	repoRoot := t.TempDir()
	specsDir := filepath.Join(repoRoot, "specs", "mymod", "feature1")
	require.NoError(t, os.MkdirAll(specsDir, 0755))
	writeFile(t, filepath.Join(specsDir, "specification.feature"), "Feature: Test")

	rules := []ComponentDiscoveryRule{
		{
			Component:    "gherkin",
			Path:         "{specs_root}/{moniker}",
			Check:        "has_files_in_subdirs",
			CheckPattern: "{specification}",
		},
	}

	mod := &Module{
		Moniker: "mymod",
		Components: ModuleComponents{
			"gherkin": &ComponentEntry{Root: "custom/path"}, // Already defined with Root
		},
	}

	vars := map[string]string{
		"moniker":       "mymod",
		"specs_root":    "specs",
		"specification": "specification.feature",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Equal(t, "custom/path", mod.Components["gherkin"].Root,
		"existing component should not be overwritten")
}

// TestDiscoverComponents_VariableExpansion verifies that variables are expanded in paths.
func TestDiscoverComponents_VariableExpansion(t *testing.T) {
	repoRoot := t.TempDir()
	designDir := filepath.Join(repoRoot, "custom-specs", "mymod", "architecture")
	require.NoError(t, os.MkdirAll(designDir, 0755))
	writeFile(t, filepath.Join(designDir, "model.dsl"), "workspace {}")

	rules := []ComponentDiscoveryRule{
		{
			Component:    "structurizr",
			Path:         "{specs_root}/{moniker}/{design_dir}",
			Check:        "required_file",
			RequiredFile: "{workspace_dsl}",
		},
	}

	mod := &Module{
		Moniker:    "mymod",
		Components: make(ModuleComponents),
	}

	vars := map[string]string{
		"moniker":       "mymod",
		"specs_root":    "custom-specs",
		"design_dir":    "architecture",
		"workspace_dsl": "model.dsl",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "structurizr")
	assert.Equal(t, "custom-specs/mymod/architecture", mod.Components["structurizr"].Root)
}

// TestDiscoverComponents_DeriveFromWithSubdirs verifies derive_from + derive_subdirs logic.
func TestDiscoverComponents_DeriveFromWithSubdirs(t *testing.T) {
	repoRoot := t.TempDir()
	specsDir := filepath.Join(repoRoot, "go", "mymod", "specs")
	require.NoError(t, os.MkdirAll(specsDir, 0755))
	writeFile(t, filepath.Join(specsDir, "godog_test.go"), "package specs")

	rules := []ComponentDiscoveryRule{
		{
			Component:     "godog",
			DeriveFrom:    []string{"go", "typescript"},
			DeriveSubdirs: map[string]string{"go": "specs", "typescript": "features"},
			FallbackPath:  "go/eac/specs/{moniker}",
			Check:         "required_file",
			RequiredFile:  "{godog_test}",
		},
	}

	mod := &Module{
		Moniker: "mymod",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/mymod"},
		},
	}

	vars := map[string]string{
		"moniker":    "mymod",
		"godog_test": "godog_test.go",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "godog")
	assert.Equal(t, "go/mymod/specs", mod.Components["godog"].Root)
}

// TestDiscoverComponents_DeriveFromFallback verifies fallback path when no source component exists.
func TestDiscoverComponents_DeriveFromFallback(t *testing.T) {
	repoRoot := t.TempDir()
	fallbackDir := filepath.Join(repoRoot, "go", "eac", "specs", "standalone")
	require.NoError(t, os.MkdirAll(fallbackDir, 0755))
	writeFile(t, filepath.Join(fallbackDir, "godog_test.go"), "package specs")

	rules := []ComponentDiscoveryRule{
		{
			Component:     "godog",
			DeriveFrom:    []string{"go", "typescript"},
			DeriveSubdirs: map[string]string{"go": "specs", "typescript": "features"},
			FallbackPath:  "go/eac/specs/{moniker}",
			Check:         "required_file",
			RequiredFile:  "{godog_test}",
		},
	}

	// Module with no go or typescript component
	mod := &Module{
		Moniker:    "standalone",
		Components: make(ModuleComponents),
	}

	vars := map[string]string{
		"moniker":    "standalone",
		"godog_test": "godog_test.go",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "godog")
	assert.Equal(t, "go/eac/specs/standalone", mod.Components["godog"].Root)
}

// TestDiscoverComponents_NilComponentsMap verifies graceful handling of nil components.
func TestDiscoverComponents_NilComponentsMap(t *testing.T) {
	repoRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, "specs", "mymod", "feat1"), 0755))
	writeFile(t, filepath.Join(repoRoot, "specs", "mymod", "feat1", "specification.feature"), "Feature: Test")

	rules := []ComponentDiscoveryRule{
		{
			Component:    "gherkin",
			Path:         "{specs_root}/{moniker}",
			Check:        "has_files_in_subdirs",
			CheckPattern: "{specification}",
		},
	}

	mod := &Module{
		Moniker:    "mymod",
		Components: nil, // nil map
	}

	vars := map[string]string{
		"moniker":       "mymod",
		"specs_root":    "specs",
		"specification": "specification.feature",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	require.NotNil(t, mod.Components)
	assert.Contains(t, mod.Components, "gherkin")
}

// TestDiscoverComponents_EmptyRules verifies no changes with empty rules.
func TestDiscoverComponents_EmptyRules(t *testing.T) {
	mod := &Module{
		Moniker:    "mymod",
		Components: make(ModuleComponents),
	}

	discoverComponentsFromRules(mod, nil, t.TempDir(), nil)
	assert.Empty(t, mod.Components)
}

// TestDiscoverComponents_MultipleRulesOrdered verifies rules are applied in order.
func TestDiscoverComponents_MultipleRulesOrdered(t *testing.T) {
	repoRoot := t.TempDir()

	// Set up filesystem for gherkin and structurizr
	specsDir := filepath.Join(repoRoot, "specs", "core", "feature1")
	require.NoError(t, os.MkdirAll(specsDir, 0755))
	writeFile(t, filepath.Join(specsDir, "specification.feature"), "Feature: Test")

	designDir := filepath.Join(repoRoot, "specs", "core", ".design")
	require.NoError(t, os.MkdirAll(designDir, 0755))
	writeFile(t, filepath.Join(designDir, "workspace.dsl"), "workspace {}")

	goDir := filepath.Join(repoRoot, "go", "core")
	require.NoError(t, os.MkdirAll(goDir, 0755))

	rules := []ComponentDiscoveryRule{
		{
			Component:    "gherkin",
			Path:         "{specs_root}/{moniker}",
			Check:        "has_files_in_subdirs",
			CheckPattern: "{specification}",
		},
		{
			Component:    "structurizr",
			Path:         "{specs_root}/{moniker}/{design_dir}",
			Check:        "required_file",
			RequiredFile: "{workspace_dsl}",
		},
		{
			Component:  "assets",
			DeriveFrom: []string{"go", "typescript", "dockerfile"},
			Check:      "dir_exists",
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
		"specification": "specification.feature",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "gherkin")
	assert.Contains(t, mod.Components, "structurizr")
	assert.Contains(t, mod.Components, "assets")
	assert.Equal(t, "specs/core", mod.Components["gherkin"].Root)
	assert.Equal(t, "specs/core/.design", mod.Components["structurizr"].Root)
	assert.Equal(t, "go/core", mod.Components["assets"].Root)
}

// TestBuildDiscoveryVars verifies discovery variable construction.
func TestBuildDiscoveryVars(t *testing.T) {
	mod := &Module{
		Moniker: "mymod",
		Components: ModuleComponents{
			"go":         &ComponentEntry{Root: "go/mymod"},
			"typescript": &ComponentEntry{Root: "ts/mymod"},
		},
	}

	repoCfg := &RepositoryConfig{
		Paths: PathsConfig{
			SpecsRoot:      "specs",
			ContainersRoot: "containers",
		},
		Conventions: ConventionsConfig{
			DesignDir:     ".design",
			WorkspaceDSL:  "workspace.dsl",
			Specification: "specification.feature",
			GodogTest:     "godog_test.go",
		},
	}

	vars := buildDiscoveryVars(mod, repoCfg)

	assert.Equal(t, "mymod", vars["moniker"])
	assert.Equal(t, "specs", vars["specs_root"])
	assert.Equal(t, "containers", vars["containers_root"])
	assert.Equal(t, ".design", vars["design_dir"])
	assert.Equal(t, "workspace.dsl", vars["workspace_dsl"])
	assert.Equal(t, "specification.feature", vars["specification"])
	assert.Equal(t, "godog_test.go", vars["godog_test"])
	assert.Equal(t, "go/mymod", vars["go_root"])
	assert.Equal(t, "ts/mymod", vars["typescript_root"])
}

// TestExpandVars verifies variable expansion in strings.
func TestExpandVars(t *testing.T) {
	vars := map[string]string{
		"moniker":    "mymod",
		"specs_root": "specs",
		"design_dir": ".design",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"{specs_root}/{moniker}", "specs/mymod"},
		{"{specs_root}/{moniker}/{design_dir}", "specs/mymod/.design"},
		{"no-vars", "no-vars"},
		{"", ""},
		{"{unknown}", "{unknown}"},
	}

	for _, tc := range tests {
		result := expandVars(tc.input, vars)
		assert.Equal(t, tc.expected, result, "input: %s", tc.input)
	}
}

// TestRuleMatchesTemplate verifies template glob matching.
func TestRuleMatchesTemplate(t *testing.T) {
	tests := []struct {
		name      string
		templates []string
		modTmpl   string
		want      bool
	}{
		{"no filter matches all", nil, "anything", true},
		{"empty filter matches all", []string{}, "anything", true},
		{"exact match", []string{"container"}, "container", true},
		{"glob match", []string{"*container*"}, "container", true},
		{"glob match multiarch", []string{"*container*"}, "container-multiarch", true},
		{"no match", []string{"*container*"}, "go-library", false},
		{"multiple patterns first matches", []string{"*container*", "*docs*"}, "container", true},
		{"multiple patterns second matches", []string{"*container*", "*docs*"}, "docs-site", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ruleMatchesTemplate(tt.templates, tt.modTmpl)
			assert.Equal(t, tt.want, got)
		})
	}
}

// =============================================================================
// Tests for DiscoverContainerComponents
// =============================================================================

// TestDiscoverContainerComponents_Basic verifies basic container component discovery.
func TestDiscoverContainerComponents_Basic(t *testing.T) {
	repoRoot := t.TempDir()

	// Create container directories with Dockerfiles
	for _, name := range []string{"pdf-oci", "mkdocs-oci", "go-oci"} {
		dir := filepath.Join(repoRoot, "containers", name)
		require.NoError(t, os.MkdirAll(dir, 0755))
		writeFile(t, filepath.Join(dir, "Dockerfile"), "FROM alpine")
	}

	mod := &Module{
		Moniker: "oci-tools",
		DiscoverComponents: &DiscoverComponentsConfig{
			Type:      "containers",
			LocalOnly: []string{"pdf-oci", "go-oci"},
		},
		Components: make(ModuleComponents),
	}

	DiscoverContainerComponents(mod, repoRoot, "containers", "myorg")

	// Should discover 3 components
	assert.Len(t, mod.Components, 3)

	// Check pdf-oci (local_only → push=false)
	pdfComp := mod.Components["pdf-oci"]
	require.NotNil(t, pdfComp)
	assert.Equal(t, "dockerfile", pdfComp.Type)
	assert.Equal(t, "containers/pdf-oci", pdfComp.Root)
	require.NotNil(t, pdfComp.DockerBuild)
	assert.False(t, pdfComp.DockerBuild.ShouldPush())
	assert.Empty(t, pdfComp.DockerBuild.Tags)
	assert.Empty(t, pdfComp.DockerBuild.Registry)
	assert.Nil(t, pdfComp.DockerBuild.Cache)

	// Check mkdocs-oci (push=true)
	mkdocsComp := mod.Components["mkdocs-oci"]
	require.NotNil(t, mkdocsComp)
	assert.Equal(t, "dockerfile", mkdocsComp.Type)
	assert.Equal(t, "containers/mkdocs-oci", mkdocsComp.Root)
	require.NotNil(t, mkdocsComp.DockerBuild)
	assert.True(t, mkdocsComp.DockerBuild.ShouldPush())
	assert.Contains(t, mkdocsComp.DockerBuild.Tags, "ghcr.io/myorg/mkdocs-oci:latest")
	assert.Equal(t, "ghcr.io", mkdocsComp.DockerBuild.Registry)
	require.NotNil(t, mkdocsComp.DockerBuild.Cache)
	assert.Equal(t, "gha", mkdocsComp.DockerBuild.Cache.Type)

	// Check go-oci (local_only → push=false)
	goComp := mod.Components["go-oci"]
	require.NotNil(t, goComp)
	assert.False(t, goComp.DockerBuild.ShouldPush())
}

// TestDiscoverContainerComponents_SkipsExisting verifies existing components are not overwritten.
func TestDiscoverContainerComponents_SkipsExisting(t *testing.T) {
	repoRoot := t.TempDir()

	// Create container with Dockerfile
	dir := filepath.Join(repoRoot, "containers", "custom-oci")
	require.NoError(t, os.MkdirAll(dir, 0755))
	writeFile(t, filepath.Join(dir, "Dockerfile"), "FROM alpine")

	mod := &Module{
		Moniker: "oci-tools",
		DiscoverComponents: &DiscoverComponentsConfig{
			Type: "containers",
		},
		Components: ModuleComponents{
			"custom-oci": &ComponentEntry{Root: "custom/path"}, // Already defined
		},
	}

	DiscoverContainerComponents(mod, repoRoot, "containers", "myorg")

	// Should not override existing component
	assert.Equal(t, "custom/path", mod.Components["custom-oci"].Root)
}

// TestDiscoverContainerComponents_EmptyDir verifies graceful handling of empty containers dir.
func TestDiscoverContainerComponents_EmptyDir(t *testing.T) {
	repoRoot := t.TempDir()

	mod := &Module{
		Moniker: "oci-tools",
		DiscoverComponents: &DiscoverComponentsConfig{
			Type: "containers",
		},
		Components: make(ModuleComponents),
	}

	DiscoverContainerComponents(mod, repoRoot, "containers", "myorg")

	assert.Empty(t, mod.Components)
}

// TestDiscoverContainerComponents_NoDockerfile verifies dirs without Dockerfile are skipped.
func TestDiscoverContainerComponents_NoDockerfile(t *testing.T) {
	repoRoot := t.TempDir()

	// Create directory without Dockerfile
	dir := filepath.Join(repoRoot, "containers", "no-docker")
	require.NoError(t, os.MkdirAll(dir, 0755))
	writeFile(t, filepath.Join(dir, "README.md"), "# Not a container")

	mod := &Module{
		Moniker: "oci-tools",
		DiscoverComponents: &DiscoverComponentsConfig{
			Type: "containers",
		},
		Components: make(ModuleComponents),
	}

	DiscoverContainerComponents(mod, repoRoot, "containers", "myorg")

	assert.Empty(t, mod.Components)
}

// TestDiscoverContainerComponents_NoLocalOnly verifies all containers push when no local_only set.
func TestDiscoverContainerComponents_NoLocalOnly(t *testing.T) {
	repoRoot := t.TempDir()

	for _, name := range []string{"alpha-oci", "beta-oci"} {
		dir := filepath.Join(repoRoot, "containers", name)
		require.NoError(t, os.MkdirAll(dir, 0755))
		writeFile(t, filepath.Join(dir, "Dockerfile"), "FROM alpine")
	}

	mod := &Module{
		Moniker: "oci-tools",
		DiscoverComponents: &DiscoverComponentsConfig{
			Type: "containers",
		},
		Components: make(ModuleComponents),
	}

	DiscoverContainerComponents(mod, repoRoot, "containers", "myorg")

	for _, name := range []string{"alpha-oci", "beta-oci"} {
		comp := mod.Components[name]
		require.NotNil(t, comp, "component %s should exist", name)
		assert.True(t, comp.DockerBuild.ShouldPush(), "component %s should push", name)
	}
}

// TestDiscoverContainerComponents_DockerBuildConfig verifies complete docker_build structure.
func TestDiscoverContainerComponents_DockerBuildConfig(t *testing.T) {
	repoRoot := t.TempDir()

	dir := filepath.Join(repoRoot, "containers", "test-oci")
	require.NoError(t, os.MkdirAll(dir, 0755))
	writeFile(t, filepath.Join(dir, "Dockerfile"), "FROM alpine")

	mod := &Module{
		Moniker: "oci-tools",
		DiscoverComponents: &DiscoverComponentsConfig{
			Type: "containers",
		},
		Components: make(ModuleComponents),
	}

	DiscoverContainerComponents(mod, repoRoot, "containers", "testowner")

	comp := mod.Components["test-oci"]
	require.NotNil(t, comp)

	dbc := comp.DockerBuild
	require.NotNil(t, dbc)
	assert.Equal(t, "test-oci", dbc.Container)
	assert.Equal(t, "containers/test-oci", dbc.Context)
	assert.Equal(t, "containers/test-oci/Dockerfile", dbc.Dockerfile)
	assert.Equal(t, []string{"linux/amd64"}, dbc.Platforms)
	assert.True(t, dbc.ShouldPush())
	assert.Equal(t, "ghcr.io", dbc.Registry)
	assert.Contains(t, dbc.Tags, "ghcr.io/testowner/test-oci:latest")
	assert.Contains(t, dbc.Tags, "ghcr.io/testowner/test-oci:sha-{short_sha}")
	require.NotNil(t, dbc.Cache)
	assert.Equal(t, "gha", dbc.Cache.Type)

	// Component should have source patterns
	require.NotNil(t, comp.Patterns)
	assert.Equal(t, []string{"**/*"}, comp.Patterns.Source)
}

// TestPreClaimContainerNames verifies container names are pre-claimed for namespace exclusion.
func TestPreClaimContainerNames(t *testing.T) {
	repoRoot := t.TempDir()

	// Create container directories with Dockerfiles
	for _, name := range []string{"pdf-oci", "mkdocs-oci", "go-oci"} {
		dir := filepath.Join(repoRoot, "containers", name)
		require.NoError(t, os.MkdirAll(dir, 0755))
		writeFile(t, filepath.Join(dir, "Dockerfile"), "FROM alpine")
	}

	// Create a directory without Dockerfile (should not be claimed)
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, "containers", "no-docker"), 0755))

	claimed := make(map[string]bool)
	preClaimContainerNames(claimed, repoRoot, "containers")

	assert.True(t, claimed["pdf-oci"])
	assert.True(t, claimed["mkdocs-oci"])
	assert.True(t, claimed["go-oci"])
	assert.False(t, claimed["no-docker"], "directories without Dockerfile should not be claimed")
}

// TestPreClaimContainerNames_EmptyDir verifies graceful handling when containers dir doesn't exist.
func TestPreClaimContainerNames_EmptyDir(t *testing.T) {
	repoRoot := t.TempDir()

	claimed := make(map[string]bool)
	preClaimContainerNames(claimed, repoRoot, "containers")

	assert.Empty(t, claimed)
}

// =============================================================================
// Tests for Go template discovery rules (testdata, test-assets, impl-assets)
// =============================================================================

// TestDiscoverComponents_GoTestdata verifies that the testdata rule discovers
// {go_root}/testdata for modules with go-* templates using dir_exists check.
func TestDiscoverComponents_GoTestdata(t *testing.T) {
	repoRoot := t.TempDir()
	testdataDir := filepath.Join(repoRoot, "go", "mymod", "testdata")
	require.NoError(t, os.MkdirAll(testdataDir, 0755))

	rules := []ComponentDiscoveryRule{
		{
			Component:     "testdata",
			OnlyTemplates: []string{"go-*"},
			DeriveFrom:    []string{"go"},
			DeriveSubdirs: map[string]string{"go": "testdata"},
			Check:         "dir_exists",
		},
	}

	mod := &Module{
		Moniker:  "mymod",
		Template: "go-exe",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/mymod"},
		},
	}

	vars := map[string]string{
		"moniker": "mymod",
		"go_root": "go/mymod",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "testdata")
	assert.Equal(t, "go/mymod/testdata", mod.Components["testdata"].Root)
}

// TestDiscoverComponents_GoTestAssets verifies that the test-assets rule discovers
// {go_root}/specs/assets for modules with go-* templates using dir_exists check.
func TestDiscoverComponents_GoTestAssets(t *testing.T) {
	repoRoot := t.TempDir()
	assetsDir := filepath.Join(repoRoot, "go", "mymod", "specs", "assets")
	require.NoError(t, os.MkdirAll(assetsDir, 0755))

	rules := []ComponentDiscoveryRule{
		{
			Component:     "test-assets",
			OnlyTemplates: []string{"go-*"},
			DeriveFrom:    []string{"go"},
			DeriveSubdirs: map[string]string{"go": "specs/assets"},
			Check:         "dir_exists",
		},
	}

	mod := &Module{
		Moniker:  "mymod",
		Template: "go-exe",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/mymod"},
		},
	}

	vars := map[string]string{
		"moniker": "mymod",
		"go_root": "go/mymod",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "test-assets")
	assert.Equal(t, "go/mymod/specs/assets", mod.Components["test-assets"].Root)
}

// TestDiscoverComponents_GoImplAssets verifies that the impl-assets rule discovers
// {go_root}/impl for modules with go-exe using has_matching_files check.
func TestDiscoverComponents_GoImplAssets(t *testing.T) {
	repoRoot := t.TempDir()
	implDir := filepath.Join(repoRoot, "go", "cli", "eac", "impl", "build", "assets")
	require.NoError(t, os.MkdirAll(implDir, 0755))
	writeFile(t, filepath.Join(implDir, "file.txt"), "asset content")

	rules := []ComponentDiscoveryRule{
		{
			Component:     "impl-assets",
			OnlyTemplates: []string{"go-exe"},
			DeriveFrom:    []string{"go"},
			DeriveSubdirs: map[string]string{"go": "impl"},
			Check:         "has_matching_files",
			Patterns:      []string{"*.txt", "*.json"},
		},
	}

	mod := &Module{
		Moniker:  "eac",
		Template: "go-exe",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/cli/eac"},
		},
	}

	vars := map[string]string{
		"moniker": "eac",
		"go_root": "go/cli/eac",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.Contains(t, mod.Components, "impl-assets")
	assert.Equal(t, "go/cli/eac/impl", mod.Components["impl-assets"].Root)
}

// TestDiscoverComponents_GoTemplateFilterSkipsNonGo verifies that go-* template filter
// skips modules with non-Go templates like container.
func TestDiscoverComponents_GoTemplateFilterSkipsNonGo(t *testing.T) {
	repoRoot := t.TempDir()
	testdataDir := filepath.Join(repoRoot, "go", "mymod", "testdata")
	require.NoError(t, os.MkdirAll(testdataDir, 0755))

	rules := []ComponentDiscoveryRule{
		{
			Component:     "testdata",
			OnlyTemplates: []string{"go-*"},
			DeriveFrom:    []string{"go"},
			DeriveSubdirs: map[string]string{"go": "testdata"},
			Check:         "dir_exists",
		},
	}

	mod := &Module{
		Moniker:  "mymod",
		Template: "container",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/mymod"},
		},
	}

	vars := map[string]string{
		"moniker": "mymod",
		"go_root": "go/mymod",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.NotContains(t, mod.Components, "testdata",
		"rule with only_templates=[go-*] should skip container")
}

// TestDiscoverComponents_GoTestdata_NoDir verifies that the testdata rule does NOT
// create a component when the testdata directory does not exist.
func TestDiscoverComponents_GoTestdata_NoDir(t *testing.T) {
	repoRoot := t.TempDir()
	// Create the go module directory but NOT testdata/
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, "go", "mymod"), 0755))

	rules := []ComponentDiscoveryRule{
		{
			Component:     "testdata",
			OnlyTemplates: []string{"go-*"},
			DeriveFrom:    []string{"go"},
			DeriveSubdirs: map[string]string{"go": "testdata"},
			Check:         "dir_exists",
		},
	}

	mod := &Module{
		Moniker:  "mymod",
		Template: "go-exe",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/mymod"},
		},
	}

	vars := map[string]string{
		"moniker": "mymod",
		"go_root": "go/mymod",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.NotContains(t, mod.Components, "testdata",
		"testdata component should not be created when directory does not exist")
}

// TestFindFirstComponentByType_Deterministic verifies that findFirstComponentByType
// always returns the same component when multiple components share the same type.
// This is critical for hash stability: non-deterministic selection causes different
// glob patterns across process invocations, breaking cache invalidation.
func TestFindFirstComponentByType_Deterministic(t *testing.T) {
	// Simulate the adapters module with many Go-typed components.
	// Map iteration order in Go is non-deterministic, so without sorting
	// this would return different results across runs.
	components := ModuleComponents{
		"godog-eac":    &ComponentEntry{Type: "go", Root: "go/adapters/godog"},
		"ai-eac":       &ComponentEntry{Type: "go", Root: "go/adapters/ai"},
		"docker-eac":   &ComponentEntry{Type: "go", Root: "go/adapters/docker"},
		"eac-to-eac":   &ComponentEntry{Type: "go", Root: "go/adapters/eac"},
		"tui-eac":      &ComponentEntry{Type: "go", Root: "go/adapters/tui"},
		"cucumber-eac": &ComponentEntry{Type: "go", Root: "go/adapters/cucumber"},
		"gotest-eac":   &ComponentEntry{Type: "go", Root: "go/adapters/gotest"},
		"mocha-eac":    &ComponentEntry{Type: "go", Root: "go/adapters/mocha"},
		"npm-eac":      &ComponentEntry{Type: "go", Root: "go/adapters/npm"},
		"pip-eac":      &ComponentEntry{Type: "go", Root: "go/adapters/pip"},
		"pytest-eac":   &ComponentEntry{Type: "go", Root: "go/adapters/pytest"},
		"behave-eac":   &ComponentEntry{Type: "go", Root: "go/adapters/behave"},
		"nuget-eac":    &ComponentEntry{Type: "go", Root: "go/adapters/nuget"},
		"dotnet-eac":   &ComponentEntry{Type: "go", Root: "go/adapters/dotnet"},
		"reqnroll-eac": &ComponentEntry{Type: "go", Root: "go/adapters/reqnroll"},
	}

	// Run 100 times to catch non-determinism (map randomization varies per iteration)
	var firstName string
	for i := 0; i < 100; i++ {
		name, entry := findFirstComponentByType(components, "go")
		require.NotNil(t, entry, "should find a go component")

		if i == 0 {
			firstName = name
			// With sorted names, "ai-eac" should always win (alphabetically first)
			assert.Equal(t, "ai-eac", name, "alphabetically first go component should be selected")
		} else {
			assert.Equal(t, firstName, name,
				"findFirstComponentByType must return the same component every time (iteration %d)", i)
		}
	}
}

// TestFindFirstComponentByType_ExplicitType verifies selection with explicit Type field.
func TestFindFirstComponentByType_ExplicitType(t *testing.T) {
	components := ModuleComponents{
		"core":      &ComponentEntry{Type: "go", Root: "go/core"},
		"contracts": &ComponentEntry{Type: "go", Root: "contracts/core"},
		"web":       &ComponentEntry{Type: "typescript", Root: "ts/web"},
	}

	name, entry := findFirstComponentByType(components, "go")
	require.NotNil(t, entry)
	// "contracts" is alphabetically before "core"
	assert.Equal(t, "contracts", name)
	assert.Equal(t, "contracts/core", entry.Root)
}

// TestFindFirstComponentByType_NilAndEmpty verifies edge cases.
func TestFindFirstComponentByType_NilAndEmpty(t *testing.T) {
	name, entry := findFirstComponentByType(nil, "go")
	assert.Equal(t, "", name)
	assert.Nil(t, entry)

	name, entry = findFirstComponentByType(ModuleComponents{}, "go")
	assert.Equal(t, "", name)
	assert.Nil(t, entry)
}

// TestDiscoverComponents_GoImplAssets_WrongTemplate verifies that the impl-assets rule
// does NOT create a component when the module template is go-library because
// only_templates requires go-exe.
func TestDiscoverComponents_GoImplAssets_WrongTemplate(t *testing.T) {
	repoRoot := t.TempDir()
	implDir := filepath.Join(repoRoot, "go", "mylib", "impl")
	require.NoError(t, os.MkdirAll(implDir, 0755))
	writeFile(t, filepath.Join(implDir, "data.json"), `{"key":"value"}`)

	rules := []ComponentDiscoveryRule{
		{
			Component:     "impl-assets",
			OnlyTemplates: []string{"go-exe"},
			DeriveFrom:    []string{"go"},
			DeriveSubdirs: map[string]string{"go": "impl"},
			Check:         "has_matching_files",
			Patterns:      []string{"*.txt", "*.json"},
		},
	}

	mod := &Module{
		Moniker:  "mylib",
		Template: "go-library",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/mylib"},
		},
	}

	vars := map[string]string{
		"moniker": "mylib",
		"go_root": "go/mylib",
	}
	discoverComponentsFromRules(mod, rules, repoRoot, vars)

	assert.NotContains(t, mod.Components, "impl-assets",
		"impl-assets should not be created for go-library (only go-exe)")
}
