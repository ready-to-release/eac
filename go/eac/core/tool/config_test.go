package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadToolConfig(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	contractsDir := filepath.Join(tmpDir, "contracts", "eac-core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create contracts dir: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write default config (must include docker and go as bootstrap tools)
	defaultConfig := `
tools:
  docker:
    type: system
    binary: docker
  go:
    type: system
    binary: go
  default-tool:
    type: system
    binary: default-binary
    description: Default tool

component-tools:
  golang:
    builder: default-tool
`
	if err := os.WriteFile(filepath.Join(contractsDir, "tool-config.yml"), []byte(defaultConfig), 0644); err != nil {
		t.Fatalf("failed to write default config: %v", err)
	}

	t.Run("loads defaults only", func(t *testing.T) {
		config, err := LoadToolConfig(tmpDir, configDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		tool, ok := config.Tools["default-tool"]
		if !ok {
			t.Fatal("default-tool should be loaded")
		}
		if tool.Binary != "default-binary" {
			t.Errorf("Binary = %q, want %q", tool.Binary, "default-binary")
		}

		assignment := config.ComponentTools["golang"]
		if assignment == nil || assignment.Builder != "default-tool" {
			t.Error("component-tools assignment not loaded correctly")
		}
	})

	t.Run("merges project override", func(t *testing.T) {
		// Write project override
		projectConfig := `
tools:
  project-tool:
    type: system
    binary: project-binary

component-tools:
  golang:
    linter: project-tool
  typescript:
    builder: project-tool
`
		if err := os.WriteFile(filepath.Join(configDir, "tool-config.yml"), []byte(projectConfig), 0644); err != nil {
			t.Fatalf("failed to write project config: %v", err)
		}

		config, err := LoadToolConfig(tmpDir, configDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have both tools
		if _, ok := config.Tools["default-tool"]; !ok {
			t.Error("default-tool should still exist")
		}
		if _, ok := config.Tools["project-tool"]; !ok {
			t.Error("project-tool should be added")
		}

		// Golang component should have merged assignments
		golangAssignment := config.ComponentTools["golang"]
		if golangAssignment == nil {
			t.Fatal("golang assignment should exist")
		}
		if golangAssignment.Builder != "default-tool" {
			t.Errorf("Builder = %q, should be preserved from defaults", golangAssignment.Builder)
		}
		if golangAssignment.Linter != "project-tool" {
			t.Errorf("Linter = %q, should be from project", golangAssignment.Linter)
		}

		// TypeScript should be from project
		tsAssignment := config.ComponentTools["typescript"]
		if tsAssignment == nil || tsAssignment.Builder != "project-tool" {
			t.Error("typescript assignment not added correctly")
		}
	})
}

func TestMergeToolAssignment(t *testing.T) {
	base := &ToolAssignment{
		Builder: "base-builder",
		Linter:  "base-linter",
	}

	override := &ToolAssignment{
		Linter:  "override-linter",
		Scanner: "override-scanner",
	}

	mergeToolAssignment(base, override)

	if base.Builder != "base-builder" {
		t.Error("Builder should be preserved")
	}
	if base.Linter != "override-linter" {
		t.Error("Linter should be overridden")
	}
	if base.Scanner != "override-scanner" {
		t.Error("Scanner should be added")
	}
}

func TestInitializeFromConfig(t *testing.T) {
	// Create minimal config
	tmpDir := t.TempDir()
	contractsDir := filepath.Join(tmpDir, "contracts", "eac-core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	config := `
tools:
  docker:
    type: system
    binary: docker
  go:
    type: system
    binary: go
  echo-tool:
    type: system
    binary: echo

component-tools:
  test:
    builder: echo-tool
`
	if err := os.WriteFile(filepath.Join(contractsDir, "tool-config.yml"), []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	registry, resolver, err := InitializeFromConfig(tmpDir, configDir)
	if err != nil {
		t.Fatalf("InitializeFromConfig failed: %v", err)
	}

	// Check registry
	if !registry.Has("echo-tool") {
		t.Error("registry should have echo-tool")
	}

	// Check resolver
	tool, err := resolver.Resolve("test", OperationBuild)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if tool.ID != "echo-tool" {
		t.Errorf("resolved tool ID = %q, want %q", tool.ID, "echo-tool")
	}
}

func TestLoadToolConfig_NoDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".eac")
	os.MkdirAll(configDir, 0755)

	// Should return empty config, not error
	config, err := LoadToolConfig(tmpDir, configDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("config should not be nil")
	}

	if config.Tools == nil {
		t.Error("Tools map should be initialized")
	}
}

func TestDefaultsRoot_ContainerMode(t *testing.T) {
	// Test container mode
	os.Setenv("R2R_CONTAINER_ROOT", "/container/root")
	defer os.Unsetenv("R2R_CONTAINER_ROOT")

	root := defaultsRoot("/local/root")
	if root != "/container/root" {
		t.Errorf("defaultsRoot = %q, want %q", root, "/container/root")
	}
}

func TestDefaultsRoot_LocalMode(t *testing.T) {
	os.Unsetenv("R2R_CONTAINER_ROOT")

	root := defaultsRoot("/local/root")
	if root != "/local/root" {
		t.Errorf("defaultsRoot = %q, want %q", root, "/local/root")
	}
}

func TestValidation_DuplicateToolIDs(t *testing.T) {
	tmpDir := t.TempDir()
	contractsDir := filepath.Join(tmpDir, "contracts", "eac-core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Config with duplicate tool ID
	config := `
tools:
  my-tool:
    type: system
    binary: first
  other-tool:
    type: system
    binary: other
  my-tool:
    type: container
    image: second
    tag: "1.0"
`
	if err := os.WriteFile(filepath.Join(contractsDir, "tool-config.yml"), []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := LoadToolConfig(tmpDir, configDir)
	if err == nil {
		t.Fatal("expected error for duplicate tool IDs")
	}

	errStr := err.Error()
	// yaml.v3 detects duplicates with: "mapping key \"my-tool\" already defined at line X"
	if !contains(errStr, "already defined") {
		t.Errorf("error should mention 'already defined', got: %v", err)
	}
	if !contains(errStr, "my-tool") {
		t.Errorf("error should mention tool name 'my-tool', got: %v", err)
	}
}

func TestValidation_BootstrapToolMustBeSystem(t *testing.T) {
	tmpDir := t.TempDir()
	contractsDir := filepath.Join(tmpDir, "contracts", "eac-core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// docker is a bootstrap tool - must be system type
	config := `
tools:
  docker:
    type: container
    image: docker
    tag: "latest"
  go:
    type: system
    binary: go
`
	if err := os.WriteFile(filepath.Join(contractsDir, "tool-config.yml"), []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := LoadToolConfig(tmpDir, configDir)
	if err == nil {
		t.Fatal("expected error for bootstrap tool as container")
	}

	errStr := err.Error()
	if !contains(errStr, "docker") {
		t.Errorf("error should mention 'docker', got: %v", err)
	}
	if !contains(errStr, "bootstrap") {
		t.Errorf("error should mention 'bootstrap', got: %v", err)
	}
}

func TestValidation_UnknownToolReference(t *testing.T) {
	tmpDir := t.TempDir()
	contractsDir := filepath.Join(tmpDir, "contracts", "eac-core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	config := `
tools:
  docker:
    type: system
    binary: docker
  go:
    type: system
    binary: go
  real-tool:
    type: system
    binary: real

component-tools:
  golang:
    builder: nonexistent-tool
`
	if err := os.WriteFile(filepath.Join(contractsDir, "tool-config.yml"), []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := LoadToolConfig(tmpDir, configDir)
	if err == nil {
		t.Fatal("expected error for unknown tool reference")
	}

	errStr := err.Error()
	if !contains(errStr, "nonexistent-tool") {
		t.Errorf("error should mention the unknown tool, got: %v", err)
	}
	if !contains(errStr, "component-tools") {
		t.Errorf("error should mention component-tools, got: %v", err)
	}
}

func TestValidation_UnknownRequirement(t *testing.T) {
	tmpDir := t.TempDir()
	contractsDir := filepath.Join(tmpDir, "contracts", "eac-core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	config := `
tools:
  docker:
    type: system
    binary: docker
  go:
    type: system
    binary: go
  my-tool:
    type: container
    image: test
    tag: "1.0"
    requirements: [nonexistent-req]
`
	if err := os.WriteFile(filepath.Join(contractsDir, "tool-config.yml"), []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := LoadToolConfig(tmpDir, configDir)
	if err == nil {
		t.Fatal("expected error for unknown requirement")
	}

	errStr := err.Error()
	if !contains(errStr, "nonexistent-req") {
		t.Errorf("error should mention the unknown requirement, got: %v", err)
	}
	if !contains(errStr, "requires") {
		t.Errorf("error should mention 'requires', got: %v", err)
	}
}


func TestValidation_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	contractsDir := filepath.Join(tmpDir, "contracts", "eac-core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Valid config following all conventions
	config := `
tools:
  docker:
    type: system
    binary: docker
  go:
    type: system
    binary: go
  npm-build:
    type: container
    image: node
    tag: "22-alpine"
    requirements: [docker]

component-tools:
  go:
    builder: go
  typescript:
    builder: npm-build
`
	if err := os.WriteFile(filepath.Join(contractsDir, "tool-config.yml"), []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := LoadToolConfig(tmpDir, configDir)
	if err != nil {
		t.Fatalf("valid config should not error: %v", err)
	}

	if len(cfg.Tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(cfg.Tools))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
