// Package testutil provides shared test utilities for command testing.
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// FixtureConfig loads EAC config from a test fixture directory.
// The fixture must contain the standard .r2r/eac/ structure.
func FixtureConfig(t *testing.T, fixturePath string) *config.EACConfig {
	t.Helper()

	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("failed to get absolute path for fixture %s: %v", fixturePath, err)
	}

	// Clear cache to ensure fresh load
	config.ClearCache()

	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        absPath,
		ValidateSchemas: false, // Skip validation in tests for speed
	})
	if err != nil {
		t.Fatalf("failed to load fixture config from %s: %v", absPath, err)
	}

	return cfg
}

// FixtureConfigValidated loads EAC config with schema validation.
func FixtureConfigValidated(t *testing.T, fixturePath string) *config.EACConfig {
	t.Helper()

	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("failed to get absolute path for fixture %s: %v", fixturePath, err)
	}

	// Clear cache to ensure fresh load
	config.ClearCache()

	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        absPath,
		ValidateSchemas: true,
	})
	if err != nil {
		t.Fatalf("failed to load fixture config from %s: %v", absPath, err)
	}

	return cfg
}

// WithWorkingDir temporarily changes to a directory and restores it after the test.
func WithWorkingDir(t *testing.T, dir string) func() {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to directory %s: %v", dir, err)
	}

	return func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("failed to restore working directory to %s: %v", original, err)
		}
	}
}

// TempDir creates a temporary directory that is automatically cleaned up after the test.
// Returns the path to the created directory.
func TempDir(t *testing.T, pattern string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// WriteFixtureFile writes a file within a fixture directory.
func WriteFixtureFile(t *testing.T, baseDir, relPath, content string) {
	t.Helper()

	fullPath := filepath.Join(baseDir, relPath)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", fullPath, err)
	}
}

// SetupMinimalFixture creates a minimal valid EAC config fixture in a temp directory.
// Returns the path to the fixture root.
func SetupMinimalFixture(t *testing.T) string {
	t.Helper()

	dir := TempDir(t, "eac-test-*")

	// Create minimal repository.yml
	WriteFixtureFile(t, dir, ".r2r/eac/repository.yml", `
name: test-repo
description: Test repository for unit tests
versioning:
  scheme: SemVer
modules: []
`)

	// Create minimal module-types.yml
	WriteFixtureFile(t, dir, ".r2r/eac/module-types.yml", `
module_types: []
`)

	return dir
}

// SetupFixtureWithModules creates a fixture with specified modules.
func SetupFixtureWithModules(t *testing.T, modules []ModuleSpec) string {
	t.Helper()

	dir := TempDir(t, "eac-test-*")

	// Build modules YAML
	modulesYAML := "modules:\n"
	for _, m := range modules {
		modulesYAML += "  - moniker: " + m.Moniker + "\n"
		modulesYAML += "    name: " + m.Name + "\n"
		modulesYAML += "    type: " + m.Type + "\n"
		if m.Description != "" {
			modulesYAML += "    description: " + m.Description + "\n"
		}
	}

	// Create repository.yml
	WriteFixtureFile(t, dir, ".r2r/eac/repository.yml", `
name: test-repo
description: Test repository for unit tests
versioning:
  scheme: SemVer
`+modulesYAML)

	// Create minimal module-types.yml with referenced types
	typesYAML := "module_types:\n"
	seenTypes := make(map[string]bool)
	for _, m := range modules {
		if !seenTypes[m.Type] {
			typesYAML += "  - name: " + m.Type + "\n"
			typesYAML += "    description: " + m.Type + " module type\n"
			seenTypes[m.Type] = true
		}
	}

	WriteFixtureFile(t, dir, ".r2r/eac/module-types.yml", typesYAML)

	return dir
}

// ModuleSpec defines a module for test fixtures.
type ModuleSpec struct {
	Moniker     string
	Name        string
	Type        string
	Description string
}

// ClearConfigCache clears the global config cache.
// Call this in test setup when you need to reload config.
func ClearConfigCache() {
	config.ClearCache()
}
