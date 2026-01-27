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
			ID:     "overwrite-test",
			Type:   ToolTypeSystem,
			Binary: "original",
		}
		tool2 := &ToolDefinition{
			ID:     "overwrite-test",
			Type:   ToolTypeSystem,
			Binary: "updated",
		}

		r.Register(tool1)
		r.Register(tool2)

		got, ok := r.Get("overwrite-test")
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
	tool := &ToolDefinition{
		ID:     "get-test",
		Type:   ToolTypeSystem,
		Binary: "test",
	}
	r.Register(tool)

	t.Run("existing tool", func(t *testing.T) {
		got, ok := r.Get("get-test")
		if !ok {
			t.Fatal("tool should be found")
		}
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
	r.Register(&ToolDefinition{ID: "tool1", Type: ToolTypeSystem, Binary: "t1"})
	r.Register(&ToolDefinition{ID: "tool2", Type: ToolTypeContainer, Image: "img"})

	all := r.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 tools, got %d", len(all))
	}

	// Verify it's a copy
	all["tool1"].Binary = "modified"
	original, _ := r.Get("tool1")
	if original.Binary == "modified" {
		t.Error("GetAll should return copies")
	}
}

func TestDefaultRegistry_ListByType(t *testing.T) {
	r := NewRegistry()
	r.Register(&ToolDefinition{ID: "sys1", Type: ToolTypeSystem, Binary: "s1"})
	r.Register(&ToolDefinition{ID: "sys2", Type: ToolTypeSystem, Binary: "s2"})
	r.Register(&ToolDefinition{ID: "cont1", Type: ToolTypeContainer, Image: "img1"})

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
tools:
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

	tool, ok := r.Get("yaml-tool")
	if !ok {
		t.Fatal("tool should be registered from YAML")
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
		Tools: map[string]*ToolDefinition{
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

	tool, ok := r.Get("config-tool")
	if !ok {
		t.Fatal("tool should be registered from config")
	}
	if tool.ID != "config-tool" {
		t.Error("ID should be backfilled from map key")
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
	r.Register(&ToolDefinition{ID: "valid2", Type: ToolTypeContainer, Image: "alpine"})

	errs := r.ValidateAll()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}
