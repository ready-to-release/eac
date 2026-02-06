package tool

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if r.Count() != 0 {
		t.Error("new registry should be empty")
	}
}

func TestDefaultRegistry_Register(t *testing.T) {
	r := NewRegistry()

	t.Run("valid tool", func(t *testing.T) {
		tool := &ToolDefinition{
			ID:     "test-tool",
			Type:   ToolTypeSystem,
			Binary: "echo",
		}
		err := r.Register(tool)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if r.Count() != 1 {
			t.Error("registry should have 1 tool")
		}
	})

	t.Run("nil tool", func(t *testing.T) {
		err := r.Register(nil)
		if err == nil {
			t.Error("expected error for nil tool")
		}
	})

	t.Run("invalid tool", func(t *testing.T) {
		tool := &ToolDefinition{
			ID:   "invalid",
			Type: ToolTypeSystem,
			// Missing binary
		}
		err := r.Register(tool)
		if err == nil {
			t.Error("expected error for invalid tool")
		}
	})

	t.Run("overwrites existing", func(t *testing.T) {
		tool1 := &ToolDefinition{
			ID:     "overwrite-test:system", // Use suffix for direct register
			Type:   ToolTypeSystem,
			Binary: "original",
		}
		tool2 := &ToolDefinition{
			ID:     "overwrite-test:system",
			Type:   ToolTypeSystem,
			Binary: "updated",
		}

		r.Register(tool1)
		r.Register(tool2)

		// Use exact suffix lookup
		got, ok := r.Get("overwrite-test:system")
		if !ok {
			t.Fatal("tool should exist")
		}
		if got.Binary != "updated" {
			t.Error("tool should have been overwritten")
		}
	})
}

func TestDefaultRegistry_Get(t *testing.T) {
	r := NewRegistry()
	// Mock verifier so system tools are considered available
	r.SetVerifier(func(tool *ToolDefinition) bool { return true })

	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"get-test": {
				Type:   ToolTypeSystem,
				Binary: "test",
			},
		},
	}
	r.RegisterFromConfig(config)

	t.Run("existing tool via canonical lookup", func(t *testing.T) {
		got, ok := r.Get("get-test")
		if !ok {
			t.Fatal("tool should be found via canonical lookup")
		}
		// Tool ID is now canonical (no suffix)
		if got.ID != "get-test" {
			t.Errorf("ID = %q, want %q", got.ID, "get-test")
		}
		// Type is preserved in the Type field
		if got.Type != ToolTypeSystem {
			t.Errorf("Type = %q, want %q", got.Type, ToolTypeSystem)
		}
	})

	t.Run("existing tool via exact suffix", func(t *testing.T) {
		got, ok := r.Get("get-test:system")
		if !ok {
			t.Fatal("tool should be found via exact suffix")
		}
		// Even when looked up with suffix, tool.ID is canonical
		if got.ID != "get-test" {
			t.Errorf("ID = %q, want %q", got.ID, "get-test")
		}
	})

	t.Run("nonexistent tool", func(t *testing.T) {
		_, ok := r.Get("nonexistent")
		if ok {
			t.Error("should not find nonexistent tool")
		}
	})

	t.Run("returns clone", func(t *testing.T) {
		got1, _ := r.Get("get-test")
		got2, _ := r.Get("get-test")

		got1.Binary = "modified"
		if got2.Binary == "modified" {
			t.Error("Get should return independent clones")
		}
	})
}

func TestDefaultRegistry_GetAll(t *testing.T) {
	r := NewRegistry()
	// Mock verifier so system tools are considered available
	r.SetVerifier(func(tool *ToolDefinition) bool { return true })

	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"tool1": {Type: ToolTypeSystem, Binary: "t1"},
		},
		ContainerTools: map[string]*ToolDefinition{
			"tool2": {Type: ToolTypeContainer, LocalPath: "containers/test"},
		},
	}
	r.RegisterFromConfig(config)

	all := r.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 tools, got %d", len(all))
	}

	// Verify tools are returned by canonical name (not suffixed)
	if _, ok := all["tool1"]; !ok {
		t.Error("GetAll should return tools by canonical name")
	}

	// Verify it's a copy
	if tool1, ok := all["tool1"]; ok {
		tool1.Binary = "modified"
		original, _ := r.Get("tool1")
		if original.Binary == "modified" {
			t.Error("GetAll should return copies")
		}
	}
}

func TestDefaultRegistry_ListByType(t *testing.T) {
	r := NewRegistry()
	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"sys1": {Type: ToolTypeSystem, Binary: "s1"},
			"sys2": {Type: ToolTypeSystem, Binary: "s2"},
		},
		ContainerTools: map[string]*ToolDefinition{
			"cont1": {Type: ToolTypeContainer, LocalPath: "containers/test"},
		},
	}
	r.RegisterFromConfig(config)

	systemTools := r.ListByType(ToolTypeSystem)
	if len(systemTools) != 2 {
		t.Errorf("expected 2 system tools, got %d", len(systemTools))
	}

	containerTools := r.ListByType(ToolTypeContainer)
	if len(containerTools) != 1 {
		t.Errorf("expected 1 container tool, got %d", len(containerTools))
	}
}

func TestDefaultRegistry_Has(t *testing.T) {
	r := NewRegistry()
	r.Register(&ToolDefinition{ID: "exists", Type: ToolTypeSystem, Binary: "x"})

	if !r.Has("exists") {
		t.Error("Has should return true for existing tool")
	}
	if r.Has("nonexistent") {
		t.Error("Has should return false for nonexistent tool")
	}
}

func TestDefaultRegistry_Remove(t *testing.T) {
	r := NewRegistry()
	r.Register(&ToolDefinition{ID: "remove-me", Type: ToolTypeSystem, Binary: "x"})

	if !r.Remove("remove-me") {
		t.Error("Remove should return true for existing tool")
	}
	if r.Has("remove-me") {
		t.Error("tool should be removed")
	}
	if r.Remove("remove-me") {
		t.Error("Remove should return false for already removed tool")
	}
}

func TestDefaultRegistry_Clear(t *testing.T) {
	r := NewRegistry()
	r.Register(&ToolDefinition{ID: "t1", Type: ToolTypeSystem, Binary: "x"})
	r.Register(&ToolDefinition{ID: "t2", Type: ToolTypeSystem, Binary: "y"})

	r.Clear()
	if r.Count() != 0 {
		t.Error("registry should be empty after Clear")
	}
}

func TestDefaultRegistry_ConcurrentAccess(t *testing.T) {
	r := NewRegistry()

	var wg sync.WaitGroup
	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r.Register(&ToolDefinition{
				ID:     "concurrent-tool",
				Type:   ToolTypeSystem,
				Binary: "test",
			})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Get("concurrent-tool")
			r.GetAll()
			r.ListByType(ToolTypeSystem)
			r.Count()
		}()
	}

	wg.Wait()
}

func TestDefaultRegistry_RegisterFromYAML(t *testing.T) {
	// Create temporary YAML file
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "tool-config.yml")

	yamlContent := `
system-tools:
  yaml-tool:
    type: system
    binary: echo
    description: Tool from YAML
    args: ["hello"]
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test YAML: %v", err)
	}

	r := NewRegistry()
	err := r.RegisterFromYAML(yamlPath)
	if err != nil {
		t.Fatalf("RegisterFromYAML failed: %v", err)
	}

	// Tools are stored with :system suffix
	tool, ok := r.Get("yaml-tool:system")
	if !ok {
		t.Fatal("tool should be registered from YAML with :system suffix")
	}
	if tool.Description != "Tool from YAML" {
		t.Errorf("Description = %q, want %q", tool.Description, "Tool from YAML")
	}
	if tool.Binary != "echo" {
		t.Errorf("Binary = %q, want %q", tool.Binary, "echo")
	}
}

func TestDefaultRegistry_RegisterFromYAML_FileNotFound(t *testing.T) {
	r := NewRegistry()
	err := r.RegisterFromYAML("/nonexistent/path.yml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestDefaultRegistry_RegisterFromConfig(t *testing.T) {
	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"config-tool": {
				Type:   ToolTypeSystem,
				Binary: "test",
			},
		},
	}

	r := NewRegistry()
	err := r.RegisterFromConfig(config)
	if err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}

	// Tools are accessible via both canonical and suffixed keys
	tool, ok := r.Get("config-tool:system")
	if !ok {
		t.Fatal("tool should be accessible via :system suffix")
	}
	// But tool.ID is canonical (no suffix)
	if tool.ID != "config-tool" {
		t.Errorf("ID = %q, want %q", tool.ID, "config-tool")
	}

	// Also accessible via canonical name
	tool2, ok := r.Get("config-tool")
	if !ok {
		t.Fatal("tool should be accessible via canonical name")
	}
	if tool2.ID != "config-tool" {
		t.Errorf("ID = %q, want %q", tool2.ID, "config-tool")
	}
}

func TestDefaultRegistry_Validate(t *testing.T) {
	r := NewRegistry()
	r.Register(&ToolDefinition{ID: "valid", Type: ToolTypeSystem, Binary: "echo"})

	t.Run("existing valid tool", func(t *testing.T) {
		err := r.Validate("valid")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nonexistent tool", func(t *testing.T) {
		err := r.Validate("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent tool")
		}
	})
}

func TestDefaultRegistry_ValidateAll(t *testing.T) {
	r := NewRegistry()
	r.Register(&ToolDefinition{ID: "valid1", Type: ToolTypeSystem, Binary: "echo"})
	r.Register(&ToolDefinition{ID: "valid2", Type: ToolTypeContainer, Image: "alpine", Tag: "3.20"})

	errs := r.ValidateAll()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

// =============================================================================
// Canonical Lookup Tests
// =============================================================================

func TestDefaultRegistry_CanonicalLookup_ExactSuffix(t *testing.T) {
	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"npm-build": {
				Type:        ToolTypeSystem,
				Binary:      "npm",
				Args:        []string{"run", "build"},
				Description: "NPM build (system)",
			},
		},
		ContainerTools: map[string]*ToolDefinition{
			"npm-build": {
				Type:        ToolTypeContainer,
				Image:       "node",
				Tag:         "22-alpine",
				Description: "NPM build (container)",
			},
		},
	}

	r := NewRegistry()
	err := r.RegisterFromConfig(config)
	if err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}

	t.Run("exact system lookup", func(t *testing.T) {
		tool, ok := r.Get("npm-build:system")
		if !ok {
			t.Fatal("should find tool with exact :system suffix")
		}
		if tool.Type != ToolTypeSystem {
			t.Errorf("Type = %q, want %q", tool.Type, ToolTypeSystem)
		}
		if tool.Binary != "npm" {
			t.Errorf("Binary = %q, want %q", tool.Binary, "npm")
		}
	})

	t.Run("exact container lookup", func(t *testing.T) {
		tool, ok := r.Get("npm-build:container")
		if !ok {
			t.Fatal("should find tool with exact :container suffix")
		}
		if tool.Type != ToolTypeContainer {
			t.Errorf("Type = %q, want %q", tool.Type, ToolTypeContainer)
		}
		if tool.Image != "node" {
			t.Errorf("Image = %q, want %q", tool.Image, "node")
		}
	})
}

func TestDefaultRegistry_CanonicalLookup_AutoBinding(t *testing.T) {
	// Mock system tool that's NOT available (binary doesn't exist)
	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"fake-tool": {
				Type:        ToolTypeSystem,
				Binary:      "nonexistent-binary-xyz-12345",
				Description: "System tool (unavailable)",
			},
		},
		ContainerTools: map[string]*ToolDefinition{
			"fake-tool": {
				Type:        ToolTypeContainer,
				Image:       "alpine",
				Tag:         "3.20",
				Description: "Container fallback",
			},
		},
		ToolBindings: map[string]ToolBinding{}, // auto is default
	}

	r := NewRegistry()
	err := r.RegisterFromConfig(config)
	if err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}

	t.Run("canonical lookup falls back to container when system unavailable", func(t *testing.T) {
		tool, ok := r.Get("fake-tool")
		if !ok {
			t.Fatal("should find tool via canonical lookup")
		}
		// Since system binary doesn't exist, should fall back to container
		if tool.Type != ToolTypeContainer {
			t.Errorf("Type = %q, want %q (system should be unavailable, fallback to container)", tool.Type, ToolTypeContainer)
		}
		if tool.Image != "alpine" {
			t.Errorf("Image = %q, want %q", tool.Image, "alpine")
		}
	})
}

func TestDefaultRegistry_CanonicalLookup_SystemBinding(t *testing.T) {
	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"docker": {
				Type:        ToolTypeSystem,
				Binary:      "docker",
				Description: "Docker (system)",
			},
		},
		ContainerTools: map[string]*ToolDefinition{
			"docker": {
				Type:        ToolTypeContainer,
				Image:       "docker",
				Tag:         "dind",
				Description: "Docker in Docker",
			},
		},
		ToolBindings: map[string]ToolBinding{
			"docker": ToolBindingSystem, // Force system
		},
	}

	r := NewRegistry()
	err := r.RegisterFromConfig(config)
	if err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}

	t.Run("system binding forces system tool", func(t *testing.T) {
		tool, ok := r.Get("docker")
		if !ok {
			t.Fatal("should find tool via canonical lookup")
		}
		if tool.Type != ToolTypeSystem {
			t.Errorf("Type = %q, want %q (binding forces system)", tool.Type, ToolTypeSystem)
		}
	})
}

func TestDefaultRegistry_CanonicalLookup_ContainerBinding(t *testing.T) {
	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"golangci-lint": {
				Type:        ToolTypeSystem,
				Binary:      "golangci-lint",
				Description: "Linter (system)",
			},
		},
		ContainerTools: map[string]*ToolDefinition{
			"golangci-lint": {
				Type:        ToolTypeContainer,
				Image:       "golangci/golangci-lint",
				Tag:         "v1.62.2",
				Description: "Linter (container)",
			},
		},
		ToolBindings: map[string]ToolBinding{
			"golangci-lint": ToolBindingContainer, // Force container
		},
	}

	r := NewRegistry()
	err := r.RegisterFromConfig(config)
	if err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}

	t.Run("container binding forces container tool", func(t *testing.T) {
		tool, ok := r.Get("golangci-lint")
		if !ok {
			t.Fatal("should find tool via canonical lookup")
		}
		if tool.Type != ToolTypeContainer {
			t.Errorf("Type = %q, want %q (binding forces container)", tool.Type, ToolTypeContainer)
		}
	})
}

func TestDefaultRegistry_CanonicalLookup_OnlyContainer(t *testing.T) {
	// When only container exists, auto should use it
	config := &ToolConfig{
		ContainerTools: map[string]*ToolDefinition{
			"semgrep": {
				Type:        ToolTypeContainer,
				Image:       "semgrep/semgrep",
				Tag:         "1.145.1", // Must use explicit version, not "latest"
				Description: "SAST scanner",
			},
		},
	}

	r := NewRegistry()
	err := r.RegisterFromConfig(config)
	if err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}

	t.Run("canonical lookup finds container when no system exists", func(t *testing.T) {
		tool, ok := r.Get("semgrep")
		if !ok {
			t.Fatal("should find tool via canonical lookup")
		}
		if tool.Type != ToolTypeContainer {
			t.Errorf("Type = %q, want %q", tool.Type, ToolTypeContainer)
		}
	})
}

func TestDefaultRegistry_CanonicalLookup_OnlySystem(t *testing.T) {
	// When only system exists and it's available, auto should use it
	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"echo-test": {
				Type:        ToolTypeSystem,
				Binary:      "echo", // Available on all systems
				Description: "Echo command",
			},
		},
	}

	r := NewRegistry()
	err := r.RegisterFromConfig(config)
	if err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}

	t.Run("canonical lookup finds available system tool", func(t *testing.T) {
		tool, ok := r.Get("echo-test")
		if !ok {
			t.Fatal("should find tool via canonical lookup")
		}
		if tool.Type != ToolTypeSystem {
			t.Errorf("Type = %q, want %q", tool.Type, ToolTypeSystem)
		}
	})
}

func TestDefaultRegistry_CanonicalLookup_NotFound(t *testing.T) {
	r := NewRegistry()

	t.Run("canonical lookup returns not found for unknown tool", func(t *testing.T) {
		_, ok := r.Get("nonexistent-tool")
		if ok {
			t.Error("should not find nonexistent tool")
		}
	})
}

func TestDefaultRegistry_CanonicalLookup_RequirementsRecursive(t *testing.T) {
	// Test that system tool availability checks requirements recursively
	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"npm": {
				Type:        ToolTypeSystem,
				Binary:      "nonexistent-npm-xyz", // Not available
				Description: "Node package manager",
			},
			"tsc": {
				Type:         ToolTypeSystem,
				Binary:       "nonexistent-tsc-xyz", // Not available
				Requirements: []string{"npm"},
				Description:  "TypeScript compiler",
			},
			"npm-build": {
				Type:         ToolTypeSystem,
				Binary:       "nonexistent-npm-xyz",
				Args:         []string{"run", "build"},
				Requirements: []string{"npm", "tsc"},
				Description:  "NPM build with TypeScript",
			},
		},
		ContainerTools: map[string]*ToolDefinition{
			"npm-build": {
				Type:        ToolTypeContainer,
				Image:       "node",
				Tag:         "22-alpine",
				Description: "NPM build (container fallback)",
			},
		},
	}

	r := NewRegistry()
	err := r.RegisterFromConfig(config)
	if err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}

	t.Run("falls back to container when system requirements unavailable", func(t *testing.T) {
		tool, ok := r.Get("npm-build")
		if !ok {
			t.Fatal("should find tool via canonical lookup")
		}
		// System npm-build requires npm and tsc which are unavailable
		// Should fall back to container
		if tool.Type != ToolTypeContainer {
			t.Errorf("Type = %q, want %q (requirements unavailable, should fallback)", tool.Type, ToolTypeContainer)
		}
	})
}

// =============================================================================
// DefaultBinding (ExecutorMode) Tests
// =============================================================================

func TestDefaultRegistry_DefaultBinding_Container(t *testing.T) {
	config := &ToolConfig{
		ExecutorMode: ExecutorModeContainer,
		SystemTools: map[string]*ToolDefinition{
			"my-tool": {Type: ToolTypeSystem, Binary: "echo"},
		},
		ContainerTools: map[string]*ToolDefinition{
			"my-tool": {Type: ToolTypeContainer, Image: "alpine", Tag: "3.20"},
		},
	}

	r := NewRegistry()
	r.SetVerifier(func(tool *ToolDefinition) bool { return true })
	r.RegisterFromConfig(config)

	tool, ok := r.Get("my-tool")
	if !ok {
		t.Fatal("should find tool")
	}
	if tool.Type != ToolTypeContainer {
		t.Errorf("Type = %q, want %q (executor-mode=container should force container)", tool.Type, ToolTypeContainer)
	}
}

func TestDefaultRegistry_DefaultBinding_System(t *testing.T) {
	config := &ToolConfig{
		ExecutorMode: ExecutorModeSystem,
		SystemTools: map[string]*ToolDefinition{
			"my-tool": {Type: ToolTypeSystem, Binary: "echo"},
		},
		ContainerTools: map[string]*ToolDefinition{
			"my-tool": {Type: ToolTypeContainer, Image: "alpine", Tag: "3.20"},
		},
	}

	r := NewRegistry()
	r.RegisterFromConfig(config)

	tool, ok := r.Get("my-tool")
	if !ok {
		t.Fatal("should find tool")
	}
	if tool.Type != ToolTypeSystem {
		t.Errorf("Type = %q, want %q (executor-mode=system should force system)", tool.Type, ToolTypeSystem)
	}
}

func TestDefaultRegistry_DefaultBinding_PerToolOverride(t *testing.T) {
	// executor-mode is container, but per-tool binding forces system for "go"
	config := &ToolConfig{
		ExecutorMode: ExecutorModeContainer,
		SystemTools: map[string]*ToolDefinition{
			"go":      {Type: ToolTypeSystem, Binary: "go"},
			"my-tool": {Type: ToolTypeSystem, Binary: "echo"},
		},
		ContainerTools: map[string]*ToolDefinition{
			"go":      {Type: ToolTypeContainer, Image: "golang", Tag: "1.22"},
			"my-tool": {Type: ToolTypeContainer, Image: "alpine", Tag: "3.20"},
		},
		ToolBindings: map[string]ToolBinding{
			"go": ToolBindingSystem, // Per-tool override
		},
	}

	r := NewRegistry()
	r.RegisterFromConfig(config)

	// "go" should resolve to system (per-tool binding overrides executor-mode)
	goTool, ok := r.Get("go")
	if !ok {
		t.Fatal("should find go tool")
	}
	if goTool.Type != ToolTypeSystem {
		t.Errorf("go Type = %q, want %q (per-tool binding should override executor-mode)", goTool.Type, ToolTypeSystem)
	}

	// "my-tool" should resolve to container (follows executor-mode)
	myTool, ok := r.Get("my-tool")
	if !ok {
		t.Fatal("should find my-tool")
	}
	if myTool.Type != ToolTypeContainer {
		t.Errorf("my-tool Type = %q, want %q (should follow executor-mode=container)", myTool.Type, ToolTypeContainer)
	}
}

func TestDefaultRegistry_DefaultBinding_Auto(t *testing.T) {
	// Explicit auto mode should behave like the default
	config := &ToolConfig{
		ExecutorMode: ExecutorModeAuto,
		SystemTools: map[string]*ToolDefinition{
			"my-tool": {Type: ToolTypeSystem, Binary: "echo"},
		},
		ContainerTools: map[string]*ToolDefinition{
			"my-tool": {Type: ToolTypeContainer, Image: "alpine", Tag: "3.20"},
		},
	}

	r := NewRegistry()
	r.SetVerifier(func(tool *ToolDefinition) bool { return true })
	r.RegisterFromConfig(config)

	// Auto mode with available system tool should prefer system
	tool, ok := r.Get("my-tool")
	if !ok {
		t.Fatal("should find tool")
	}
	if tool.Type != ToolTypeSystem {
		t.Errorf("Type = %q, want %q (auto mode should prefer available system tool)", tool.Type, ToolTypeSystem)
	}
}

func TestDefaultRegistry_DefaultBinding_ContainerMode_SystemOnlyTool(t *testing.T) {
	// executor-mode=container but tool only has system variant → not found (strict)
	config := &ToolConfig{
		ExecutorMode: ExecutorModeContainer,
		SystemTools: map[string]*ToolDefinition{
			"system-only": {Type: ToolTypeSystem, Binary: "echo"},
		},
	}

	r := NewRegistry()
	r.RegisterFromConfig(config)

	_, ok := r.Get("system-only")
	if ok {
		t.Error("should NOT find tool when executor-mode=container and only system variant exists")
	}
}

func TestDefaultRegistry_DefaultBinding_SystemMode_ContainerOnlyTool(t *testing.T) {
	// executor-mode=system but tool only has container variant → not found (strict)
	config := &ToolConfig{
		ExecutorMode: ExecutorModeSystem,
		ContainerTools: map[string]*ToolDefinition{
			"container-only": {Type: ToolTypeContainer, Image: "alpine", Tag: "3.20"},
		},
	}

	r := NewRegistry()
	r.RegisterFromConfig(config)

	_, ok := r.Get("container-only")
	if ok {
		t.Error("should NOT find tool when executor-mode=system and only container variant exists")
	}
}

func TestDefaultRegistry_DefaultBinding_AutoMode_FallbackPreserved(t *testing.T) {
	// In auto mode, direct-ID fallback should still work (for ad-hoc tools)
	r := NewRegistry()
	r.Register(&ToolDefinition{ID: "adhoc-tool", Type: ToolTypeSystem, Binary: "echo"})

	tool, ok := r.Get("adhoc-tool")
	if !ok {
		t.Fatal("auto mode should find ad-hoc tools via direct-ID fallback")
	}
	if tool.Binary != "echo" {
		t.Errorf("Binary = %q, want %q", tool.Binary, "echo")
	}
}

func TestDefaultRegistry_DefaultBinding_ExplicitSuffixBypassesMode(t *testing.T) {
	// Even with executor-mode=system, explicit :container suffix should work
	config := &ToolConfig{
		ExecutorMode: ExecutorModeSystem,
		SystemTools: map[string]*ToolDefinition{
			"my-tool": {Type: ToolTypeSystem, Binary: "echo"},
		},
		ContainerTools: map[string]*ToolDefinition{
			"my-tool": {Type: ToolTypeContainer, Image: "alpine", Tag: "3.20"},
		},
	}

	r := NewRegistry()
	r.RegisterFromConfig(config)

	// Explicit suffix should bypass executor-mode
	tool, ok := r.Get("my-tool:container")
	if !ok {
		t.Fatal("should find tool via explicit :container suffix")
	}
	if tool.Type != ToolTypeContainer {
		t.Errorf("Type = %q, want %q", tool.Type, ToolTypeContainer)
	}
}

func TestDefaultRegistry_GetExecutorMode(t *testing.T) {
	tests := []struct {
		name string
		mode ExecutorMode
		want ExecutorMode
	}{
		{"auto", ExecutorModeAuto, ExecutorModeAuto},
		{"container", ExecutorModeContainer, ExecutorModeContainer},
		{"system", ExecutorModeSystem, ExecutorModeSystem},
		{"empty defaults to auto", "", ExecutorModeAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ToolConfig{ExecutorMode: tt.mode}
			r := NewRegistry()
			r.RegisterFromConfig(config)

			got := r.GetExecutorMode()
			if got != tt.want {
				t.Errorf("GetExecutorMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

