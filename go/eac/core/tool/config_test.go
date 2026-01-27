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

	// Write default config
	defaultConfig := `
tools:
  default-tool:
    type: system
    binary: default-binary
    description: Default tool

component-tools:
  go:
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

		assignment := config.ComponentTools["go"]
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
  go:
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

		// Go should have merged assignments
		goAssignment := config.ComponentTools["go"]
		if goAssignment == nil {
			t.Fatal("go assignment should exist")
		}
		if goAssignment.Builder != "default-tool" {
			t.Errorf("Builder = %q, should be preserved from defaults", goAssignment.Builder)
		}
		if goAssignment.Linter != "project-tool" {
			t.Errorf("Linter = %q, should be from project", goAssignment.Linter)
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
