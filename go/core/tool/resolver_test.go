package tool

import (
	"os"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

func TestNewResolver(t *testing.T) {
	r := NewRegistry()
	resolver := NewResolver(r)
	if resolver == nil {
		t.Fatal("NewResolver returned nil")
	}
}

func TestDefaultResolver_LayeredResolution(t *testing.T) {
	// Setup registry with tools
	registry := NewRegistry()
	registry.Register(&ToolDefinition{ID: "default-tool", Type: ToolTypeSystem, Binary: "default"})
	registry.Register(&ToolDefinition{ID: "project-tool", Type: ToolTypeSystem, Binary: "project"})
	registry.Register(&ToolDefinition{ID: "env-tool", Type: ToolTypeSystem, Binary: "env"})
	registry.Register(&ToolDefinition{ID: "cli-tool", Type: ToolTypeSystem, Binary: "cli"})

	resolver := NewResolver(registry)

	// Load configuration layers
	resolver.LoadDefaults(map[string]*ToolAssignment{
		"go": {Builder: "default-tool"},
	})
	resolver.LoadProjectConfig(map[string]*ToolAssignment{
		"go": {Builder: "project-tool"},
	})
	resolver.LoadEnvironmentConfig("ci", map[string]*ToolAssignment{
		"go": {Builder: "env-tool"},
	})

	t.Run("defaults only", func(t *testing.T) {
		// Create fresh resolver with only defaults
		r := NewResolver(registry)
		r.LoadDefaults(map[string]*ToolAssignment{
			"go": {Builder: "default-tool"},
		})

		tool, err := r.Resolve("go", core.ActionBuild)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tool.ID != "default-tool" {
			t.Errorf("expected default-tool, got %s", tool.ID)
		}
	})

	t.Run("project overrides defaults", func(t *testing.T) {
		// Project config without environment
		r := NewResolver(registry)
		r.LoadDefaults(map[string]*ToolAssignment{
			"go": {Builder: "default-tool"},
		})
		r.LoadProjectConfig(map[string]*ToolAssignment{
			"go": {Builder: "project-tool"},
		})

		tool, err := r.Resolve("go", core.ActionBuild)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tool.ID != "project-tool" {
			t.Errorf("expected project-tool, got %s", tool.ID)
		}
	})

	t.Run("environment overrides project", func(t *testing.T) {
		resolver.SetEnvironment("ci")
		tool, err := resolver.Resolve("go", core.ActionBuild)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tool.ID != "env-tool" {
			t.Errorf("expected env-tool, got %s", tool.ID)
		}
	})

	t.Run("CLI overrides all", func(t *testing.T) {
		resolver.SetEnvironment("ci")
		resolver.SetOverride("go", core.ActionBuild, "cli-tool")

		tool, err := resolver.Resolve("go", core.ActionBuild)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tool.ID != "cli-tool" {
			t.Errorf("expected cli-tool, got %s", tool.ID)
		}
	})

	t.Run("ClearOverrides", func(t *testing.T) {
		resolver.SetOverride("go", core.ActionBuild, "cli-tool")
		resolver.ClearOverrides()

		resolver.SetEnvironment("ci")
		tool, err := resolver.Resolve("go", core.ActionBuild)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tool.ID != "env-tool" {
			t.Errorf("after ClearOverrides, expected env-tool, got %s", tool.ID)
		}
	})
}

func TestDefaultResolver_NoToolConfigured(t *testing.T) {
	registry := NewRegistry()
	resolver := NewResolver(registry)

	_, err := resolver.Resolve("nonexistent", core.ActionBuild)
	if err == nil {
		t.Error("expected error for unconfigured component type")
	}
}

func TestDefaultResolver_ToolNotInRegistry(t *testing.T) {
	registry := NewRegistry()
	resolver := NewResolver(registry)
	resolver.LoadDefaults(map[string]*ToolAssignment{
		"go": {Builder: "nonexistent-tool"},
	})

	_, err := resolver.Resolve("go", core.ActionBuild)
	if err == nil {
		t.Error("expected error for tool not in registry")
	}
}

func TestDefaultResolver_ResolveAll(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&ToolDefinition{ID: "builder", Type: ToolTypeSystem, Binary: "b"})
	registry.Register(&ToolDefinition{ID: "linter", Type: ToolTypeSystem, Binary: "l"})
	registry.Register(&ToolDefinition{ID: "tester", Type: ToolTypeSystem, Binary: "t"})

	resolver := NewResolver(registry)
	resolver.LoadDefaults(map[string]*ToolAssignment{
		"go": {
			Builder: "builder",
			Linter:  "linter",
			Tester:  "tester",
		},
	})

	tools := resolver.ResolveAll("go")
	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}

	if tools[core.ActionBuild] == nil || tools[core.ActionBuild].ID != "builder" {
		t.Error("builder not resolved correctly")
	}
	if tools[core.ActionLint] == nil || tools[core.ActionLint].ID != "linter" {
		t.Error("linter not resolved correctly")
	}
	if tools[core.ActionTest] == nil || tools[core.ActionTest].ID != "tester" {
		t.Error("tester not resolved correctly")
	}
}

func TestDefaultResolver_ResolveMultiple(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&ToolDefinition{ID: "linter1", Type: ToolTypeSystem, Binary: "l1"})
	registry.Register(&ToolDefinition{ID: "linter2", Type: ToolTypeSystem, Binary: "l2"})
	registry.Register(&ToolDefinition{ID: "linter3", Type: ToolTypeSystem, Binary: "l3"})

	resolver := NewResolver(registry)
	resolver.LoadDefaults(map[string]*ToolAssignment{
		"go": {
			Linters: []string{"linter1", "linter2", "linter3"},
		},
	})

	tools, err := resolver.ResolveMultiple("go", core.ActionLint)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 3 {
		t.Errorf("expected 3 linters, got %d", len(tools))
	}
}

func TestDefaultResolver_DetectEnvironment(t *testing.T) {
	resolver := NewResolver(NewRegistry())

	// Test local detection (no CI env vars)
	env := resolver.DetectEnvironment()
	// Note: This will return "ci" if run in CI, "windows" on Windows, "local" otherwise
	if env != "local" && env != "ci" && env != "windows" {
		t.Errorf("unexpected environment: %s", env)
	}

	// Test CI detection
	os.Setenv("CI", "true")
	defer os.Unsetenv("CI")

	env = resolver.DetectEnvironment()
	if env != "ci" {
		t.Errorf("expected ci environment, got %s", env)
	}
}

func TestDefaultResolver_HasTool(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&ToolDefinition{ID: "builder", Type: ToolTypeSystem, Binary: "b"})

	resolver := NewResolver(registry)
	resolver.LoadDefaults(map[string]*ToolAssignment{
		"go": {Builder: "builder"},
	})

	if !resolver.HasTool("go", core.ActionBuild) {
		t.Error("HasTool should return true for configured tool")
	}
	if resolver.HasTool("go", core.ActionLint) {
		t.Error("HasTool should return false for unconfigured operation")
	}
	if resolver.HasTool("nonexistent", core.ActionBuild) {
		t.Error("HasTool should return false for unconfigured component type")
	}
}

func TestDefaultResolver_ListConfiguredComponents(t *testing.T) {
	registry := NewRegistry()
	resolver := NewResolver(registry)

	resolver.LoadDefaults(map[string]*ToolAssignment{
		"go":         {},
		"typescript": {},
	})
	resolver.LoadProjectConfig(map[string]*ToolAssignment{
		"python": {},
	})
	resolver.LoadEnvironmentConfig("ci", map[string]*ToolAssignment{
		"rust": {},
	})

	components := resolver.ListConfiguredComponents()
	if len(components) != 4 {
		t.Errorf("expected 4 components, got %d: %v", len(components), components)
	}

	// Check all expected components are present
	componentSet := make(map[string]bool)
	for _, c := range components {
		componentSet[c] = true
	}
	for _, expected := range []string{"go", "typescript", "python", "rust"} {
		if !componentSet[expected] {
			t.Errorf("missing component: %s", expected)
		}
	}
}

func TestDefaultResolver_LoadFromConfig(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&ToolDefinition{ID: "tool1", Type: ToolTypeSystem, Binary: "t1"})
	registry.Register(&ToolDefinition{ID: "tool2", Type: ToolTypeSystem, Binary: "t2"})
	registry.Register(&ToolDefinition{ID: "ci-tool", Type: ToolTypeSystem, Binary: "ci"})

	config := &ToolConfig{
		ComponentTools: map[string]*ToolAssignment{
			"go": {Builder: "tool1"},
		},
		Environments: map[string]*EnvironmentConfig{
			"ci": {
				ComponentTools: map[string]*ToolAssignment{
					"go": {Builder: "ci-tool"},
				},
			},
		},
	}

	resolver := NewResolver(registry)
	resolver.LoadFromConfig(config, false) // Load as project config

	// Without environment set, should get project config
	tool, err := resolver.Resolve("go", core.ActionBuild)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tool.ID != "tool1" {
		t.Errorf("expected tool1, got %s", tool.ID)
	}

	// With ci environment, should get ci-tool
	resolver.SetEnvironment("ci")
	tool, err = resolver.Resolve("go", core.ActionBuild)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tool.ID != "ci-tool" {
		t.Errorf("expected ci-tool, got %s", tool.ID)
	}
}

func TestDefaultResolver_PartialAssignments(t *testing.T) {
	// Test that partial assignments in higher layers only override specified fields
	registry := NewRegistry()
	registry.Register(&ToolDefinition{ID: "default-builder", Type: ToolTypeSystem, Binary: "db"})
	registry.Register(&ToolDefinition{ID: "default-linter", Type: ToolTypeSystem, Binary: "dl"})
	registry.Register(&ToolDefinition{ID: "project-linter", Type: ToolTypeSystem, Binary: "pl"})

	resolver := NewResolver(registry)
	resolver.LoadDefaults(map[string]*ToolAssignment{
		"go": {
			Builder: "default-builder",
			Linter:  "default-linter",
		},
	})
	resolver.LoadProjectConfig(map[string]*ToolAssignment{
		"go": {
			Linter: "project-linter", // Only override linter
		},
	})

	// Builder should come from defaults (project didn't specify)
	// Linter should come from project
	buildTool, err := resolver.Resolve("go", core.ActionBuild)
	if err != nil {
		t.Fatalf("unexpected error resolving builder: %v", err)
	}
	if buildTool.ID != "default-builder" {
		t.Errorf("builder should be from defaults, got %s", buildTool.ID)
	}

	lintTool, err := resolver.Resolve("go", core.ActionLint)
	if err != nil {
		t.Fatalf("unexpected error resolving linter: %v", err)
	}
	if lintTool.ID != "project-linter" {
		t.Errorf("linter should be from project, got %s", lintTool.ID)
	}
}
