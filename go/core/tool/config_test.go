package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadToolConfig(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	contractsDir := filepath.Join(tmpDir, "contracts", "core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create contracts dir: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write default config (must include docker and go as bootstrap tools)
	defaultConfig := `
system-tools:
  docker:
    binary: docker
  go:
    binary: go
  default-tool:
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

		tool, ok := config.SystemTools["default-tool"]
		if !ok {
			t.Fatal("default-tool should be loaded in SystemTools")
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
system-tools:
  project-tool:
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
		if _, ok := config.SystemTools["default-tool"]; !ok {
			t.Error("default-tool should still exist")
		}
		if _, ok := config.SystemTools["project-tool"]; !ok {
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
	contractsDir := filepath.Join(tmpDir, "contracts", "core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	config := `
system-tools:
  docker:
    binary: docker
  go:
    binary: go
  echo-tool:
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

	// Check registry - tools are accessible via both canonical name and suffixed alias
	if !registry.Has("echo-tool:system") {
		t.Error("registry should have echo-tool:system alias")
	}
	if !registry.Has("echo-tool") {
		t.Error("registry should have echo-tool canonical key")
	}

	// Check resolver - resolved tool has CANONICAL ID (no suffix)
	tool, err := resolver.Resolve("test", OperationBuild)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	// Tool ID is now canonical - type suffix removed
	if tool.ID != "echo-tool" {
		t.Errorf("resolved tool ID = %q, want %q", tool.ID, "echo-tool")
	}
	// Type is still available in the Type field
	if tool.Type != ToolTypeSystem {
		t.Errorf("resolved tool Type = %q, want %q", tool.Type, ToolTypeSystem)
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

	if config.SystemTools == nil {
		t.Error("SystemTools map should be initialized")
	}
	if config.ContainerTools == nil {
		t.Error("ContainerTools map should be initialized")
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
	contractsDir := filepath.Join(tmpDir, "contracts", "core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Config with duplicate tool ID within system-tools
	config := `
system-tools:
  my-tool:
    binary: first
  other-tool:
    binary: other
  my-tool:
    binary: second
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

func TestValidation_BootstrapToolMustExist(t *testing.T) {
	tmpDir := t.TempDir()
	contractsDir := filepath.Join(tmpDir, "contracts", "core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Missing docker bootstrap tool - only has go
	config := `
system-tools:
  go:
    binary: go
`
	if err := os.WriteFile(filepath.Join(contractsDir, "tool-config.yml"), []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := LoadToolConfig(tmpDir, configDir)
	if err == nil {
		t.Fatal("expected error for missing bootstrap tool")
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
	contractsDir := filepath.Join(tmpDir, "contracts", "core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	config := `
system-tools:
  docker:
    binary: docker
  go:
    binary: go
  real-tool:
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
	contractsDir := filepath.Join(tmpDir, "contracts", "core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	config := `
system-tools:
  docker:
    binary: docker
  go:
    binary: go

container-tools:
  my-tool:
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
	contractsDir := filepath.Join(tmpDir, "contracts", "core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Valid config with system-tools and container-tools
	config := `
system-tools:
  docker:
    binary: docker
  go:
    binary: go

container-tools:
  npm-build:
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

	if len(cfg.SystemTools) != 2 {
		t.Errorf("expected 2 system tools, got %d", len(cfg.SystemTools))
	}
	if len(cfg.ContainerTools) != 1 {
		t.Errorf("expected 1 container tool, got %d", len(cfg.ContainerTools))
	}
}

func TestValidation_SameToolInBothTables(t *testing.T) {
	tmpDir := t.TempDir()
	contractsDir := filepath.Join(tmpDir, "contracts", "core", "0.1.0", "defaults")
	configDir := filepath.Join(tmpDir, ".eac")

	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Same canonical name in both system-tools and container-tools (valid!)
	config := `
system-tools:
  docker:
    binary: docker
  go:
    binary: go
  npm-build:
    binary: npm
    args: [run, build]
    requirements: [npm]
  npm:
    binary: npm

container-tools:
  npm-build:
    image: node
    tag: "22-alpine"
    requirements: [docker]

tool-bindings:
  docker: system
  go: system

component-tools:
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

	// Should have npm-build in both tables
	if _, ok := cfg.SystemTools["npm-build"]; !ok {
		t.Error("npm-build should exist in SystemTools")
	}
	if _, ok := cfg.ContainerTools["npm-build"]; !ok {
		t.Error("npm-build should exist in ContainerTools")
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
