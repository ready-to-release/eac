//go:build L0 && ov

// Tests type-based defaults applied when loading contracts from .r2r/eac
package modules_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRepoWithTypes creates a temporary repository with both modules.yml and module-types.yml
func createTestRepoWithTypes(t *testing.T, modulesContent, typesContent string) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Create .r2r/eac directory
	eacDir := filepath.Join(tmpDir, ".r2r", "eac")
	if err := os.MkdirAll(eacDir, 0755); err != nil {
		t.Fatalf("failed to create EAC directory: %v", err)
	}

	// Create fake .git directory so repository root detection works
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Write modules.yml
	modulesPath := filepath.Join(eacDir, "modules.yml")
	if err := os.WriteFile(modulesPath, []byte(modulesContent), 0644); err != nil {
		t.Fatalf("failed to write modules.yml: %v", err)
	}

	// Write module-types.yml if provided
	if typesContent != "" {
		typesPath := filepath.Join(eacDir, "module-types.yml")
		if err := os.WriteFile(typesPath, []byte(typesContent), 0644); err != nil {
			t.Fatalf("failed to write module-types.yml: %v", err)
		}
	}

	return tmpDir
}

// Standard fixture for module-types.yml
const standardModuleTypesYAML = `
types:
  - name: go
    build_deps: [go]
    defaults:
      files:
        source: ["**/*.go", "**/*.go.txt"]
        config: ["go.mod", "go.sum"]
        changelog: CHANGELOG.md
      repo:
        specs: ["specs/{moniker}/**"]
        test_impl: "{root}/tests"
        design: "specs/{moniker}/.design"

  - name: no-specs-type
    build_deps: []
    defaults:
      repo:
        specs: []

  - name: custom-type
    build_deps: []
    defaults:
      files:
        source: ["**/*.custom"]
        config: ["config.yml"]
        assets: ["**/*.txt"]
      repo:
        specs: ["custom/{moniker}/**"]
`

// TestTypeDefaults_Go tests Go type defaults application
func TestTypeDefaults_Go(t *testing.T) {
	modulesContent := `
modules:
  - moniker: my-lib
    name: My Library
    type: go
    files:
      root: src/lib
`
	repoRoot := createTestRepoWithTypes(t, modulesContent, standardModuleTypesYAML)

	registry, err := loadTestModules(repoRoot)
	require.NoError(t, err, "failed to load modules")

	module, found := registry.Get("my-lib")
	require.True(t, found, "module not found")

	// Type defaults should be applied
	assert.Equal(t, []string{"**/*.go", "**/*.go.txt"}, module.Files.Source, "source should come from type defaults")
	assert.Equal(t, []string{"go.mod", "go.sum"}, module.Files.Config, "config should come from type defaults")
	assert.Equal(t, "CHANGELOG.md", module.Files.Changelog, "changelog should come from type defaults")

	// Variable substitution should be applied
	assert.Equal(t, []string{"specs/my-lib/**"}, module.Files.Repo.Specs, "specs should have {moniker} substituted")
	assert.Equal(t, "src/lib/tests", module.Files.Repo.TestImpl, "test_impl should have {root} substituted")
	assert.Equal(t, "specs/my-lib/.design", module.Files.Repo.Design, "design should have {moniker} substituted")
}

// TestTypeDefaults_CustomType tests custom type defaults application
func TestTypeDefaults_CustomType(t *testing.T) {
	modulesContent := `
modules:
  - moniker: my-custom
    name: My Custom Module
    type: custom-type
    files:
      root: src/custom
`
	repoRoot := createTestRepoWithTypes(t, modulesContent, standardModuleTypesYAML)

	registry, err := loadTestModules(repoRoot)
	require.NoError(t, err, "failed to load modules")

	module, found := registry.Get("my-custom")
	require.True(t, found, "module not found")

	// Custom type defaults should be applied
	assert.Equal(t, []string{"**/*.custom"}, module.Files.Source)
	assert.Equal(t, []string{"config.yml"}, module.Files.Config)
	assert.Equal(t, []string{"**/*.txt"}, module.Files.Assets)
	assert.Equal(t, []string{"custom/my-custom/**"}, module.Files.Repo.Specs)

	// Generic defaults for fields not in type defaults
	assert.Equal(t, "CHANGELOG.md", module.Files.Changelog)
}

// TestTypeDefaults_TypeWithEmptySpecs tests that type with specs: [] gives empty specs
func TestTypeDefaults_TypeWithEmptySpecs(t *testing.T) {
	modulesContent := `
modules:
  - moniker: no-specs-module
    name: No Specs Module
    type: no-specs-type
    files:
      root: src/nospecs
`
	repoRoot := createTestRepoWithTypes(t, modulesContent, standardModuleTypesYAML)

	registry, err := loadTestModules(repoRoot)
	require.NoError(t, err, "failed to load modules")

	module, found := registry.Get("no-specs-module")
	require.True(t, found, "module not found")

	// Type has specs: [], module should get empty specs (not generic default)
	assert.NotNil(t, module.Files.Repo.Specs, "specs should not be nil")
	assert.Empty(t, module.Files.Repo.Specs, "specs should be empty from type default")
}

// TestTypeDefaults_NoModuleTypes tests backward compatibility when module-types.yml doesn't exist
func TestTypeDefaults_NoModuleTypes(t *testing.T) {
	modulesContent := `
modules:
  - moniker: legacy-module
    name: Legacy Module
    type: some-type
    files:
      root: src/legacy
`
	// Create repo without module-types.yml
	repoRoot := createTestRepoWithTypes(t, modulesContent, "")

	registry, err := loadTestModules(repoRoot)
	require.NoError(t, err, "modules should load even without module-types.yml")

	module, found := registry.Get("legacy-module")
	require.True(t, found, "module not found")

	// Should use generic defaults
	assert.Equal(t, []string{"specs/legacy-module/**"}, module.Files.Repo.Specs)
	assert.Equal(t, "CHANGELOG.md", module.Files.Changelog)
	// Source/Config/Assets should be empty (no generic defaults for these)
	assert.Empty(t, module.Files.Source)
	assert.Empty(t, module.Files.Config)
}

// TestTypeDefaults_UnknownType tests that unknown types fall back to generic defaults
func TestTypeDefaults_UnknownType(t *testing.T) {
	modulesContent := `
modules:
  - moniker: unknown-type-module
    name: Unknown Type Module
    type: unknown-type-xyz
    files:
      root: src/unknown
`
	repoRoot := createTestRepoWithTypes(t, modulesContent, standardModuleTypesYAML)

	registry, err := loadTestModules(repoRoot)
	require.NoError(t, err, "should load with unknown type")

	module, found := registry.Get("unknown-type-module")
	require.True(t, found, "module not found")

	// Should use generic defaults
	assert.Equal(t, []string{"specs/unknown-type-module/**"}, module.Files.Repo.Specs)
	assert.Equal(t, "CHANGELOG.md", module.Files.Changelog)
	assert.Empty(t, module.Files.Source)
	assert.Empty(t, module.Files.Config)
}

// TestTypeDefaults_ExplicitOverridesTypeDefaults tests that explicit values take precedence
func TestTypeDefaults_ExplicitOverridesTypeDefaults(t *testing.T) {
	modulesContent := `
modules:
  - moniker: explicit-module
    name: Explicit Module
    type: go
    files:
      root: src/explicit
      source:
        - "custom/*.go"
      config:
        - custom.mod
      changelog: HISTORY.md
      repo:
        specs:
          - my/custom/specs/**
        test_impl: custom/tests
        design: custom/design
`
	repoRoot := createTestRepoWithTypes(t, modulesContent, standardModuleTypesYAML)

	registry, err := loadTestModules(repoRoot)
	require.NoError(t, err, "failed to load modules")

	module, found := registry.Get("explicit-module")
	require.True(t, found, "module not found")

	// Explicit values should be preserved, not overwritten by type defaults
	assert.Equal(t, []string{"custom/*.go"}, module.Files.Source)
	assert.Equal(t, []string{"custom.mod"}, module.Files.Config)
	assert.Equal(t, "HISTORY.md", module.Files.Changelog)
	assert.Equal(t, []string{"my/custom/specs/**"}, module.Files.Repo.Specs)
	assert.Equal(t, "custom/tests", module.Files.Repo.TestImpl)
	assert.Equal(t, "custom/design", module.Files.Repo.Design)
}

// TestTypeDefaults_ExplicitEmptySpecsPreserved tests that explicit specs: [] is preserved
func TestTypeDefaults_ExplicitEmptySpecsPreserved(t *testing.T) {
	modulesContent := `
modules:
  - moniker: empty-specs-module
    name: Empty Specs Module
    type: go
    files:
      root: src/emptyspecs
      repo:
        specs: []
`
	repoRoot := createTestRepoWithTypes(t, modulesContent, standardModuleTypesYAML)

	registry, err := loadTestModules(repoRoot)
	require.NoError(t, err, "failed to load modules")

	module, found := registry.Get("empty-specs-module")
	require.True(t, found, "module not found")

	// Explicit empty specs should be preserved, not replaced with type default
	assert.NotNil(t, module.Files.Repo.Specs)
	assert.Empty(t, module.Files.Repo.Specs, "explicit empty specs should not be replaced with type default")
}

// TestTypeDefaults_VariableSubstitution tests all variable substitutions
func TestTypeDefaults_VariableSubstitution(t *testing.T) {
	// Create a type that uses all three variables
	typesContent := `
types:
  - name: all-vars-type
    build_deps: []
    defaults:
      files:
        source: ["types/{type}/**"]
      repo:
        specs: ["specs/{moniker}/{type}/**"]
        test_impl: "{root}/{type}/tests"
        design: "designs/{moniker}/{type}"
`
	modulesContent := `
modules:
  - moniker: test-mod
    name: Test Module
    type: all-vars-type
    files:
      root: src/test
`
	repoRoot := createTestRepoWithTypes(t, modulesContent, typesContent)

	registry, err := loadTestModules(repoRoot)
	require.NoError(t, err, "failed to load modules")

	module, found := registry.Get("test-mod")
	require.True(t, found, "module not found")

	// Check all substitutions
	assert.Equal(t, []string{"types/all-vars-type/**"}, module.Files.Source, "{type} should be replaced")
	assert.Equal(t, []string{"specs/test-mod/all-vars-type/**"}, module.Files.Repo.Specs, "{moniker} and {type} should be replaced")
	assert.Equal(t, "src/test/all-vars-type/tests", module.Files.Repo.TestImpl, "{root} and {type} should be replaced")
	assert.Equal(t, "designs/test-mod/all-vars-type", module.Files.Repo.Design, "{moniker} and {type} should be replaced")
}

// TestTypeDefaults_PartialTypeDefaults tests when type has only some defaults
func TestTypeDefaults_PartialTypeDefaults(t *testing.T) {
	typesContent := `
types:
  - name: partial-type
    build_deps: []
    defaults:
      files:
        source: ["**/*.partial"]
        # config, assets, changelog not specified
      # repo not specified
`
	modulesContent := `
modules:
  - moniker: partial-module
    name: Partial Module
    type: partial-type
    files:
      root: src/partial
`
	repoRoot := createTestRepoWithTypes(t, modulesContent, typesContent)

	registry, err := loadTestModules(repoRoot)
	require.NoError(t, err, "failed to load modules")

	module, found := registry.Get("partial-module")
	require.True(t, found, "module not found")

	// Type default for source
	assert.Equal(t, []string{"**/*.partial"}, module.Files.Source)

	// Generic defaults for others
	assert.Equal(t, "CHANGELOG.md", module.Files.Changelog)
	assert.Equal(t, []string{"specs/partial-module/**"}, module.Files.Repo.Specs)
	assert.Equal(t, "specs/partial-module/.design", module.Files.Repo.Design)

	// Empty for fields without generic defaults (test_impl requires repository.yml)
	assert.Empty(t, module.Files.Config)
	assert.Empty(t, module.Files.Assets)
	assert.Empty(t, module.Files.Repo.TestImpl, "test_impl requires repository.yml configuration")
}

// TestTypeDefaults_MultipleModulesSameType tests multiple modules with same type
func TestTypeDefaults_MultipleModulesSameType(t *testing.T) {
	modulesContent := `
modules:
  - moniker: lib-one
    name: Library One
    type: go
    files:
      root: src/one

  - moniker: lib-two
    name: Library Two
    type: go
    files:
      root: src/two
`
	repoRoot := createTestRepoWithTypes(t, modulesContent, standardModuleTypesYAML)

	registry, err := loadTestModules(repoRoot)
	require.NoError(t, err, "failed to load modules")

	libOne, found := registry.Get("lib-one")
	require.True(t, found, "lib-one not found")

	libTwo, found := registry.Get("lib-two")
	require.True(t, found, "lib-two not found")

	// Both should have same type defaults for source/config
	assert.Equal(t, []string{"**/*.go", "**/*.go.txt"}, libOne.Files.Source)
	assert.Equal(t, []string{"**/*.go", "**/*.go.txt"}, libTwo.Files.Source)

	// But different values after substitution
	assert.Equal(t, []string{"specs/lib-one/**"}, libOne.Files.Repo.Specs)
	assert.Equal(t, []string{"specs/lib-two/**"}, libTwo.Files.Repo.Specs)

	assert.Equal(t, "src/one/tests", libOne.Files.Repo.TestImpl)
	assert.Equal(t, "src/two/tests", libTwo.Files.Repo.TestImpl)
}

// TestTypeDefaults_MixedTypes tests modules with different types
func TestTypeDefaults_MixedTypes(t *testing.T) {
	modulesContent := `
modules:
  - moniker: go-mod
    name: Go Module
    type: go
    files:
      root: src/go

  - moniker: custom-mod
    name: Custom Module
    type: custom-type
    files:
      root: src/custom

  - moniker: no-specs-mod
    name: No Specs Module
    type: no-specs-type
    files:
      root: src/nospecs
`
	repoRoot := createTestRepoWithTypes(t, modulesContent, standardModuleTypesYAML)

	registry, err := loadTestModules(repoRoot)
	require.NoError(t, err, "failed to load modules")

	goMod, _ := registry.Get("go-mod")
	customMod, _ := registry.Get("custom-mod")
	noSpecsMod, _ := registry.Get("no-specs-mod")

	// Go type
	assert.Equal(t, []string{"**/*.go", "**/*.go.txt"}, goMod.Files.Source)
	assert.Equal(t, []string{"go.mod", "go.sum"}, goMod.Files.Config)

	// Custom type
	assert.Equal(t, []string{"**/*.custom"}, customMod.Files.Source)
	assert.Equal(t, []string{"config.yml"}, customMod.Files.Config)

	// No specs type
	assert.Empty(t, noSpecsMod.Files.Repo.Specs)
}
