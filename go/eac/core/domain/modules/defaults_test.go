//go:build L0 && ov

// Tests default values applied when loading contracts from .r2r/eac
package modules_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/domain"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"gopkg.in/yaml.v3"
)

// loadTestModules loads modules from a test repo without schema validation
func loadTestModules(repoRoot string) (*modules.Registry, error) {
	return modules.LoadFromWorkspaceNoValidation(repoRoot)
}

// createTestRepo creates a temporary repository with minimal repository.yml for testing defaults
func createTestRepo(t *testing.T, modulesContent string) string {
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

	// Write repository.yml
	modulesPath := filepath.Join(eacDir, "repository.yml")
	if err := os.WriteFile(modulesPath, []byte(modulesContent), 0644); err != nil {
		t.Fatalf("failed to write repository.yml: %v", err)
	}

	return tmpDir
}

// TestModuleDefaults_Description tests that description defaults to name when not specified
func TestModuleDefaults_Description(t *testing.T) {
	content := `
modules:
  - moniker: test-module
    name: My Test Module Name
    type: test-type
    files:
      root: src/test
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("test-module")
	if !found {
		t.Fatal("module not found")
	}

	expected := "My Test Module Name"
	if module.Description != expected {
		t.Errorf("Description default: expected %q, got %q", expected, module.Description)
	}
}

// TestModuleDefaults_DependsOn tests that depends_on defaults to empty slice when not specified
func TestModuleDefaults_DependsOn(t *testing.T) {
	content := `
modules:
  - moniker: test-module
    name: Test Module
    type: test-type
    files:
      root: src/test
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("test-module")
	if !found {
		t.Fatal("module not found")
	}

	if module.DependsOn == nil {
		t.Error("DependsOn default: expected non-nil empty slice, got nil")
	}
	if len(module.DependsOn) != 0 {
		t.Errorf("DependsOn default: expected empty slice, got %v", module.DependsOn)
	}
}

// TestModuleDefaults_FilesChangelog tests that files.changelog defaults to "CHANGELOG.md"
func TestModuleDefaults_FilesChangelog(t *testing.T) {
	content := `
modules:
  - moniker: test-module
    name: Test Module
    type: test-type
    files:
      root: src/test
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("test-module")
	if !found {
		t.Fatal("module not found")
	}

	expected := "CHANGELOG.md"
	if module.Files.Changelog != expected {
		t.Errorf("Files.Changelog default: expected %q, got %q", expected, module.Files.Changelog)
	}
}

// TestModuleDefaults_FilesRepoSpecs tests that files.repo.specs defaults to ["specs/<moniker>/**"]
func TestModuleDefaults_FilesRepoSpecs(t *testing.T) {
	content := `
modules:
  - moniker: test-module
    name: Test Module
    type: test-type
    files:
      root: src/test
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("test-module")
	if !found {
		t.Fatal("module not found")
	}

	if module.Files.Repo.Specs == nil {
		t.Fatal("Files.Repo.Specs default: expected non-nil slice, got nil")
	}
	if len(module.Files.Repo.Specs) != 1 {
		t.Fatalf("Files.Repo.Specs default: expected 1 element, got %d", len(module.Files.Repo.Specs))
	}

	expected := "specs/test-module/**"
	if module.Files.Repo.Specs[0] != expected {
		t.Errorf("Files.Repo.Specs default: expected %q, got %q", expected, module.Files.Repo.Specs[0])
	}
}

// TestModuleDefaults_FilesSource tests that files.source defaults to nil/empty when not specified
func TestModuleDefaults_FilesSource(t *testing.T) {
	content := `
modules:
  - moniker: test-module
    name: Test Module
    type: test-type
    files:
      root: src/test
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("test-module")
	if !found {
		t.Fatal("module not found")
	}

	// Source should be nil or empty when not specified (no default applied)
	if len(module.Files.Source) != 0 {
		t.Errorf("Files.Source default: expected empty, got %v", module.Files.Source)
	}
}

// TestModuleDefaults_FilesConfig tests that files.config defaults to nil/empty when not specified
func TestModuleDefaults_FilesConfig(t *testing.T) {
	content := `
modules:
  - moniker: test-module
    name: Test Module
    type: test-type
    files:
      root: src/test
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("test-module")
	if !found {
		t.Fatal("module not found")
	}

	if len(module.Files.Config) != 0 {
		t.Errorf("Files.Config default: expected empty, got %v", module.Files.Config)
	}
}

// TestModuleDefaults_FilesAssets tests that files.assets defaults to nil/empty when not specified
func TestModuleDefaults_FilesAssets(t *testing.T) {
	content := `
modules:
  - moniker: test-module
    name: Test Module
    type: test-type
    files:
      root: src/test
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("test-module")
	if !found {
		t.Fatal("module not found")
	}

	if len(module.Files.Assets) != 0 {
		t.Errorf("Files.Assets default: expected empty, got %v", module.Files.Assets)
	}
}

// TestModuleDefaults_FilesTests tests that files.tests defaults to nil/empty when not specified
func TestModuleDefaults_FilesTests(t *testing.T) {
	content := `
modules:
  - moniker: test-module
    name: Test Module
    type: test-type
    files:
      root: src/test
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("test-module")
	if !found {
		t.Fatal("module not found")
	}

	if len(module.Files.Tests) != 0 {
		t.Errorf("Files.Tests default: expected empty, got %v", module.Files.Tests)
	}
}

// TestModuleDefaults_FilesExclude tests that files.exclude defaults to nil/empty when not specified
func TestModuleDefaults_FilesExclude(t *testing.T) {
	content := `
modules:
  - moniker: test-module
    name: Test Module
    type: test-type
    files:
      root: src/test
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("test-module")
	if !found {
		t.Fatal("module not found")
	}

	if len(module.Files.Exclude) != 0 {
		t.Errorf("Files.Exclude default: expected empty, got %v", module.Files.Exclude)
	}
}

// TestModuleDefaults_FilesRepoOther tests that files.repo.other defaults to nil/empty when not specified
func TestModuleDefaults_FilesRepoOther(t *testing.T) {
	content := `
modules:
  - moniker: test-module
    name: Test Module
    type: test-type
    files:
      root: src/test
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("test-module")
	if !found {
		t.Fatal("module not found")
	}

	if len(module.Files.Repo.Other) != 0 {
		t.Errorf("Files.Repo.Other default: expected empty, got %v", module.Files.Repo.Other)
	}
}

// TestModuleDefaults_FilesRepoExclude tests that files.repo.exclude defaults to nil/empty when not specified
func TestModuleDefaults_FilesRepoExclude(t *testing.T) {
	content := `
modules:
  - moniker: test-module
    name: Test Module
    type: test-type
    files:
      root: src/test
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("test-module")
	if !found {
		t.Fatal("module not found")
	}

	if len(module.Files.Repo.Exclude) != 0 {
		t.Errorf("Files.Repo.Exclude default: expected empty, got %v", module.Files.Repo.Exclude)
	}
}

// TestModuleDefaults_AllFieldsMinimal tests that a minimal module definition gets all defaults applied
func TestModuleDefaults_AllFieldsMinimal(t *testing.T) {
	// Minimal module with only required fields
	content := `
modules:
  - moniker: minimal-module
    name: Minimal Module
    files:
      root: src/minimal
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("minimal-module")
	if !found {
		t.Fatal("module not found")
	}

	// Test all defaults are applied
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Moniker", module.Moniker, "minimal-module"},
		{"Name", module.Name, "Minimal Module"},
		{"Description", module.Description, "Minimal Module"},
		{"Files.Root", module.Files.Root, "src/minimal"},
		{"Files.Changelog", module.Files.Changelog, "CHANGELOG.md"},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.got)
		}
	}

	// Test slice defaults
	if module.DependsOn == nil {
		t.Error("DependsOn should not be nil")
	}
	if len(module.DependsOn) != 0 {
		t.Errorf("DependsOn should be empty, got %v", module.DependsOn)
	}

	if len(module.Files.Repo.Specs) != 1 || module.Files.Repo.Specs[0] != "specs/minimal-module/**" {
		t.Errorf("Files.Repo.Specs should be [specs/minimal-module/**], got %v", module.Files.Repo.Specs)
	}
}

// TestModuleDefaults_ExplicitValuesNotOverwritten tests that explicit values are not overwritten by defaults
func TestModuleDefaults_ExplicitValuesNotOverwritten(t *testing.T) {
	content := `
modules:
  - moniker: explicit-module
    name: Explicit Module
    type: custom-type
    description: Custom description
    depends_on:
      - dep1
      - dep2
    files:
      root: src/explicit
      source:
        - "*.go"
      config:
        - go.mod
      changelog: HISTORY.md
      repo:
        specs:
          - custom/specs/**
  - moniker: dep1
    name: Dependency One
    type: dep-type
    files:
      root: src/dep1
  - moniker: dep2
    name: Dependency Two
    type: dep-type
    files:
      root: src/dep2
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("explicit-module")
	if !found {
		t.Fatal("module not found")
	}

	// Verify explicit values are preserved
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Type", module.Type, "custom-type"},
		{"Description", module.Description, "Custom description"},
		{"Files.Changelog", module.Files.Changelog, "HISTORY.md"},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.got)
		}
	}

	if len(module.DependsOn) != 2 {
		t.Errorf("DependsOn should have 2 elements, got %v", module.DependsOn)
	}

	if len(module.Files.Repo.Specs) != 1 || module.Files.Repo.Specs[0] != "custom/specs/**" {
		t.Errorf("Files.Repo.Specs should be [custom/specs/**], got %v", module.Files.Repo.Specs)
	}
}

// TestModuleDefaults_OmittedFieldsGetDefaults tests that omitted fields get defaults applied
func TestModuleDefaults_OmittedFieldsGetDefaults(t *testing.T) {
	// Note: Fields must be omitted (not empty string) to get defaults.
	// Schema validation rejects empty strings for fields with pattern constraints.
	// Only description and changelog can be empty strings (no pattern constraint).
	content := `
modules:
  - moniker: omitted-fields-module
    name: Omitted Fields Module
    description: ""
    files:
      root: src/test
      changelog: ""
`
	repoRoot := createTestRepo(t, content)

	registry, err := loadTestModules(repoRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	module, found := registry.Get("omitted-fields-module")
	if !found {
		t.Fatal("module not found")
	}

	// Empty description should default to name
	if module.Description != "Omitted Fields Module" {
		t.Errorf("Description should default to name when empty, got %q", module.Description)
	}
	// Empty changelog should default
	if module.Files.Changelog != "CHANGELOG.md" {
		t.Errorf("Files.Changelog should default when empty string, got %q", module.Files.Changelog)
	}
}

// Compile-time check that contracts package is used (for the yaml import)
var _ = domain.BaseContract{}
var _ = yaml.Node{}
