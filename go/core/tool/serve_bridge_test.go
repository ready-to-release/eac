package tool

import (
	"testing"
)

func TestNewServeBridge(t *testing.T) {
	bridge := NewServeBridge()

	if bridge == nil {
		t.Fatal("NewServeBridge returned nil")
	}
}

func TestServeBridge_GetServerByToolID_YAMLTool(t *testing.T) {
	bridge := NewServeBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:        "nginx-oci",
		Type:      ToolTypeContainer,
		LocalPath: "containers/nginx-oci",
		Serve: &ServeConfig{
			ContainerPort: 8080,
		},
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Should return server for servable tool
	server := bridge.GetServerByToolID("nginx-oci")
	if server == nil {
		t.Fatal("GetServerByToolID returned nil for servable tool")
	}
}

func TestServeBridge_GetServerByToolID_NonServable(t *testing.T) {
	bridge := NewServeBridge()

	// Set up tool system with non-servable tool
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:        "trivy-sbom",
		Type:      ToolTypeContainer,
		LocalPath: "containers/trivy",
		// No Serve config - not servable
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Should return nil for non-servable tool
	server := bridge.GetServerByToolID("trivy-sbom")
	if server != nil {
		t.Error("GetServerByToolID should return nil for non-servable tool")
	}
}

func TestServeBridge_GetServerByToolID_NotFound(t *testing.T) {
	bridge := NewServeBridge()

	// No tool system configured
	server := bridge.GetServerByToolID("nginx-oci")
	if server != nil {
		t.Error("GetServerByToolID should return nil when no server available")
	}
}

func TestServeBridge_SetToolSystem(t *testing.T) {
	bridge := NewServeBridge()

	registry := NewRegistry()
	resolver := NewResolver(registry)
	executor := &mockExecutor{}

	bridge.SetToolSystem(registry, resolver, executor)

	// Verify tool system is set by registering a tool and retrieving it
	registry.Register(&ToolDefinition{
		ID:        "structurizr-lite",
		Type:      ToolTypeContainer,
		LocalPath: "containers/structurizr",
		Serve: &ServeConfig{
			ContainerPort: 8080,
		},
	})

	server := bridge.GetServerByToolID("structurizr-lite")
	if server == nil {
		t.Error("Tool system not properly configured")
	}
}

func TestServeBridge_HasServerByToolID(t *testing.T) {
	bridge := NewServeBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:        "nginx-oci",
		Type:      ToolTypeContainer,
		LocalPath: "containers/nginx-oci",
		Serve: &ServeConfig{
			ContainerPort: 8000,
		},
	})
	registry.Register(&ToolDefinition{
		ID:        "mkdocs-live",
		Type:      ToolTypeContainer,
		LocalPath: "containers/mkdocs-live",
		Serve: &ServeConfig{
			ContainerPort: 8000,
		},
	})
	registry.Register(&ToolDefinition{
		ID:        "trivy-sbom", // Not servable
		Type:      ToolTypeContainer,
		LocalPath: "containers/trivy",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	tests := []struct {
		toolID string
		exists bool
	}{
		{"nginx-oci", true},
		{"mkdocs-live", true},
		{"trivy-sbom", false},  // exists but not servable
		{"nonexistent", false}, // doesn't exist
	}

	for _, tt := range tests {
		t.Run(tt.toolID, func(t *testing.T) {
			if got := bridge.HasServerByToolID(tt.toolID); got != tt.exists {
				t.Errorf("HasServerByToolID(%q) = %v, want %v", tt.toolID, got, tt.exists)
			}
		})
	}
}

func TestServeBridge_GetAllServableTools(t *testing.T) {
	bridge := NewServeBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:        "nginx-oci",
		Type:      ToolTypeContainer,
		LocalPath: "containers/nginx-oci",
		Serve: &ServeConfig{
			ContainerPort: 8000,
		},
	})
	registry.Register(&ToolDefinition{
		ID:        "mkdocs-live",
		Type:      ToolTypeContainer,
		LocalPath: "containers/mkdocs-live",
		Serve: &ServeConfig{
			ContainerPort: 8000,
		},
	})
	registry.Register(&ToolDefinition{
		ID:        "structurizr-lite",
		Type:      ToolTypeContainer,
		LocalPath: "containers/structurizr",
		Serve: &ServeConfig{
			ContainerPort: 8080,
		},
	})
	registry.Register(&ToolDefinition{
		ID:        "trivy-sbom", // Not servable
		Type:      ToolTypeContainer,
		LocalPath: "containers/trivy",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	tools := bridge.GetAllServableTools()

	// Should include only servable tools: nginx-oci, mkdocs-live, structurizr-lite
	if len(tools) != 3 {
		t.Errorf("GetAllServableTools() returned %d tools, want 3", len(tools))
	}

	// Verify all returned tools are servable
	for _, tool := range tools {
		if !tool.IsServable() {
			t.Errorf("Tool %q should be servable", tool.ID)
		}
	}
}

func TestGlobalServeBridge(t *testing.T) {
	// GlobalServeBridge should return same instance
	bridge1 := GlobalServeBridge()
	bridge2 := GlobalServeBridge()

	if bridge1 != bridge2 {
		t.Error("GlobalServeBridge should return singleton")
	}

	if bridge1 == nil {
		t.Error("GlobalServeBridge should not return nil")
	}
}

func TestServeBridge_Concurrent(t *testing.T) {
	bridge := NewServeBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:        "nginx-oci",
		Type:      ToolTypeContainer,
		LocalPath: "containers/nginx-oci",
		Serve: &ServeConfig{
			ContainerPort: 8000,
		},
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Test concurrent access
	done := make(chan bool)

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.GetServerByToolID("nginx-oci")
		}
		done <- true
	}()

	// Concurrent HasServerByToolID
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.HasServerByToolID("mkdocs-live")
		}
		done <- true
	}()

	// Concurrent GetAllServableTools
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.GetAllServableTools()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}

func TestServeOptions(t *testing.T) {
	// Test that ServeOptions struct works correctly
	opts := ServeOptions{
		LiveReload:  true,
		OpenBrowser: false,
		WatchPaths:  []string{"docs/", "mkdocs.yml"},
	}

	if !opts.LiveReload {
		t.Error("LiveReload should be true")
	}
	if opts.OpenBrowser {
		t.Error("OpenBrowser should be false")
	}
	if len(opts.WatchPaths) != 2 {
		t.Errorf("WatchPaths should have 2 elements, got %d", len(opts.WatchPaths))
	}
}

func TestServeResult(t *testing.T) {
	// Test that ServeResult struct works correctly
	result := &ServeResult{
		Port:    8080,
		URL:     "http://localhost:8080",
		PID:     12345,
		Running: true,
	}

	if result.Port != 8080 {
		t.Errorf("Port = %d, want 8080", result.Port)
	}
	if result.URL != "http://localhost:8080" {
		t.Errorf("URL = %q, want http://localhost:8080", result.URL)
	}
	if result.PID != 12345 {
		t.Errorf("PID = %d, want 12345", result.PID)
	}
	if !result.Running {
		t.Error("Running should be true")
	}
}

func TestServeBridge_NilToolSystem(t *testing.T) {
	bridge := NewServeBridge()

	// Don't set tool system - should handle nil gracefully
	// No servers should be available
	if bridge.HasServerByToolID("nginx-oci") {
		t.Error("HasServerByToolID should return false without tool system")
	}

	// GetServerByToolID should return nil
	server := bridge.GetServerByToolID("mkdocs-live")
	if server != nil {
		t.Error("GetServerByToolID should return nil without tool system")
	}
}

func TestServeBridge_ToolSystemWithExecutorOnly(t *testing.T) {
	bridge := NewServeBridge()

	// Set tool system without executor - tool-based servers shouldn't work
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:        "mkdocs-live",
		Type:      ToolTypeContainer,
		LocalPath: "containers/mkdocs-live",
		Serve: &ServeConfig{
			ContainerPort: 8000,
		},
	})
	bridge.SetToolSystem(registry, nil, nil) // nil executor

	// Server should NOT work without executor
	server := bridge.GetServerByToolID("mkdocs-live")
	if server != nil {
		t.Error("GetServerByToolID should return nil without executor")
	}
}

func TestServeBridge_GetAllServableTools_Empty(t *testing.T) {
	bridge := NewServeBridge()

	// No tool system configured
	tools := bridge.GetAllServableTools()
	if tools != nil {
		t.Errorf("GetAllServableTools() should return nil, got %d items", len(tools))
	}
}

func TestServeBridge_GetServerToolByID(t *testing.T) {
	bridge := NewServeBridge()

	// Set up tool system
	registry := NewRegistry()
	staticSite := &ToolDefinition{
		ID:        "nginx-oci",
		Type:      ToolTypeContainer,
		LocalPath: "containers/nginx-oci",
		Serve: &ServeConfig{
			ContainerPort: 8000,
		},
	}
	registry.Register(staticSite)
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Should return the tool definition
	tool := bridge.GetServerToolByID("nginx-oci")
	if tool == nil {
		t.Fatal("GetServerToolByID returned nil for existing tool")
	}
	if !tool.IsServable() {
		t.Error("Returned tool should be servable")
	}

	// Non-existent tool should return nil
	if bridge.GetServerToolByID("nonexistent") != nil {
		t.Error("GetServerToolByID should return nil for non-existent tool")
	}
}
