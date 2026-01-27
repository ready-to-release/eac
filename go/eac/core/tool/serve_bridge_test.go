package tool

import (
	"testing"
)

func TestNewServeBridge(t *testing.T) {
	bridge := NewServeBridge()

	if bridge == nil {
		t.Fatal("NewServeBridge returned nil")
	}
	if bridge.serverTools == nil {
		t.Error("serverTools map not initialized")
	}
}

func TestServeBridge_GetServer_YAMLTool(t *testing.T) {
	bridge := NewServeBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:    "static-site", // Matches default mapping
		Type:  ToolTypeContainer,
		Image: "nginx:alpine",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Should return YAML tool server
	server := bridge.GetServer(ServerStaticSite)
	if server == nil {
		t.Fatal("GetServer returned nil for YAML tool server")
	}
}

func TestServeBridge_GetServer_NotFound(t *testing.T) {
	bridge := NewServeBridge()

	// No tool system configured
	server := bridge.GetServer(ServerStaticSite)
	if server != nil {
		t.Error("GetServer should return nil when no server available")
	}
}

func TestServeBridge_GetServer_CustomMapping(t *testing.T) {
	bridge := NewServeBridge()

	// Register custom mapping
	bridge.SetServerToolMapping(ServerMkDocsLive, "custom-mkdocs-server")

	// Set up tool system with the custom tool
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:    "custom-mkdocs-server",
		Type:  ToolTypeContainer,
		Image: "custom/mkdocs",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Should use custom mapping
	server := bridge.GetServer(ServerMkDocsLive)
	if server == nil {
		t.Fatal("GetServer returned nil for custom mapped server")
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
		ID:    "structurizr",
		Type:  ToolTypeContainer,
		Image: "structurizr/lite",
	})

	server := bridge.GetServer(ServerStructurizr)
	if server == nil {
		t.Error("Tool system not properly configured")
	}
}

func TestServeBridge_HasServer(t *testing.T) {
	bridge := NewServeBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:    "static-site",
		Type:  ToolTypeContainer,
		Image: "nginx:alpine",
	})
	registry.Register(&ToolDefinition{
		ID:    "mkdocs-live",
		Type:  ToolTypeContainer,
		Image: "squidfunk/mkdocs-material",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	tests := []struct {
		serverType ServerType
		exists     bool
	}{
		{ServerStaticSite, true},   // yaml
		{ServerMkDocsLive, true},   // yaml
		{ServerStructurizr, false}, // not registered
	}

	for _, tt := range tests {
		t.Run(string(tt.serverType), func(t *testing.T) {
			if got := bridge.HasServer(tt.serverType); got != tt.exists {
				t.Errorf("HasServer(%q) = %v, want %v", tt.serverType, got, tt.exists)
			}
		})
	}
}

func TestServeBridge_GetAllServerTypes(t *testing.T) {
	bridge := NewServeBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:    "static-site",
		Type:  ToolTypeContainer,
		Image: "nginx:alpine",
	})
	registry.Register(&ToolDefinition{
		ID:    "mkdocs-live",
		Type:  ToolTypeContainer,
		Image: "squidfunk/mkdocs-material",
	})
	registry.Register(&ToolDefinition{
		ID:    "structurizr",
		Type:  ToolTypeContainer,
		Image: "structurizr/lite",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	serverTypes := bridge.GetAllServerTypes()

	// Should include: static-site, mkdocs-live, structurizr
	if len(serverTypes) < 3 {
		t.Errorf("GetAllServerTypes() returned %d types, want at least 3", len(serverTypes))
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

func TestParseServerType(t *testing.T) {
	tests := []struct {
		input    string
		expected ServerType
		ok       bool
	}{
		{"static-site", ServerStaticSite, true},
		{"mkdocs-live", ServerMkDocsLive, true},
		{"structurizr", ServerStructurizr, true},
		{"unknown", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseServerType(tt.input)
			if ok != tt.ok {
				t.Errorf("ParseServerType(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if got != tt.expected {
				t.Errorf("ParseServerType(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestServerTypeConstants(t *testing.T) {
	// Verify all server type constants have expected values
	tests := []struct {
		constant ServerType
		value    string
	}{
		{ServerStaticSite, "static-site"},
		{ServerMkDocsLive, "mkdocs-live"},
		{ServerStructurizr, "structurizr"},
	}

	for _, tt := range tests {
		if string(tt.constant) != tt.value {
			t.Errorf("Server constant %v = %q, want %q", tt.constant, tt.constant, tt.value)
		}
	}
}

func TestServeBridge_DefaultServerMappings(t *testing.T) {
	bridge := NewServeBridge()

	// Verify default mappings exist
	expected := map[ServerType]string{
		ServerStaticSite:  "static-site",
		ServerMkDocsLive:  "mkdocs-live",
		ServerStructurizr: "structurizr",
	}

	for serverType, expectedToolID := range expected {
		toolID := bridge.GetServerToolID(serverType)
		if toolID != expectedToolID {
			t.Errorf("Default mapping for %v = %q, want %q", serverType, toolID, expectedToolID)
		}
	}
}

func TestServeBridge_Concurrent(t *testing.T) {
	bridge := NewServeBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:    "static-site",
		Type:  ToolTypeContainer,
		Image: "nginx:alpine",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Test concurrent access
	done := make(chan bool)

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.GetServer(ServerStaticSite)
		}
		done <- true
	}()

	// Concurrent HasServer
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.HasServer(ServerMkDocsLive)
		}
		done <- true
	}()

	// Concurrent GetAllServerTypes
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.GetAllServerTypes()
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

func TestServeBridge_SetServerToolMapping(t *testing.T) {
	bridge := NewServeBridge()

	// Set a custom mapping
	bridge.SetServerToolMapping(ServerMkDocsLive, "my-custom-mkdocs")

	// Verify it was set
	toolID := bridge.GetServerToolID(ServerMkDocsLive)
	if toolID != "my-custom-mkdocs" {
		t.Errorf("GetServerToolID(ServerMkDocsLive) = %q, want %q", toolID, "my-custom-mkdocs")
	}

	// Other mappings should be unchanged
	staticToolID := bridge.GetServerToolID(ServerStaticSite)
	if staticToolID != "static-site" {
		t.Errorf("GetServerToolID(ServerStaticSite) = %q, want %q", staticToolID, "static-site")
	}
}

func TestServeBridge_GetServerToolID_NotFound(t *testing.T) {
	bridge := NewServeBridge()

	// Query a non-existent server type (cast to avoid compile error)
	toolID := bridge.GetServerToolID(ServerType("unknown-server"))
	if toolID != "" {
		t.Errorf("GetServerToolID for unknown type should return empty string, got %q", toolID)
	}
}

func TestServeBridge_NilToolSystem(t *testing.T) {
	bridge := NewServeBridge()

	// Don't set tool system - should handle nil gracefully
	// No servers should be available
	if bridge.HasServer(ServerStaticSite) {
		t.Error("HasServer should return false without tool system")
	}

	// GetServer should return nil
	server := bridge.GetServer(ServerMkDocsLive)
	if server != nil {
		t.Error("GetServer should return nil without tool system")
	}
}

func TestServeBridge_ToolSystemWithExecutorOnly(t *testing.T) {
	bridge := NewServeBridge()

	// Set tool system without executor - tool-based servers shouldn't work
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:    "mkdocs-live",
		Type:  ToolTypeContainer,
		Image: "squidfunk/mkdocs-material",
	})
	bridge.SetToolSystem(registry, nil, nil) // nil executor

	// YAML-based server should NOT work without executor
	server := bridge.GetServer(ServerMkDocsLive)
	if server != nil {
		t.Error("GetServer should return nil without executor")
	}
}

func TestServeBridge_GetAllServerTypes_Empty(t *testing.T) {
	bridge := NewServeBridge()

	// No tool system configured
	serverTypes := bridge.GetAllServerTypes()
	if len(serverTypes) != 0 {
		t.Errorf("GetAllServerTypes() should return empty slice, got %d items", len(serverTypes))
	}
}
