package tool

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/domain"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
)

func TestNewScanBridge(t *testing.T) {
	bridge := NewScanBridge()

	if bridge == nil {
		t.Fatal("NewScanBridge returned nil")
	}
}

func TestScanBridge_GetScannerByToolID_YAMLTool(t *testing.T) {
	bridge := NewScanBridge()

	// Set up tool system - use LocalPath for test containers (no version pinning needed)
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:           "trivy-sbom",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Should return scanner for tool
	scanner := bridge.GetScannerByToolID("trivy-sbom")
	if scanner == nil {
		t.Fatal("GetScannerByToolID returned nil for scanner tool")
	}
}

func TestScanBridge_GetScannerByToolID_NotFound(t *testing.T) {
	bridge := NewScanBridge()

	// No tool system configured
	scanner := bridge.GetScannerByToolID("trivy-sbom")
	if scanner != nil {
		t.Error("GetScannerByToolID should return nil when no scanner available")
	}
}

func TestScanBridge_SetToolSystem(t *testing.T) {
	bridge := NewScanBridge()

	registry := NewRegistry()
	resolver := NewResolver(registry)
	executor := &mockExecutor{}

	bridge.SetToolSystem(registry, resolver, executor)

	// Verify tool system is set by registering a tool and retrieving it
	registry.Register(&ToolDefinition{
		ID:           "trivy-vuln",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})

	scanner := bridge.GetScannerByToolID("trivy-vuln")
	if scanner == nil {
		t.Error("Tool system not properly configured")
	}
}

func TestScanBridge_HasScannerByToolID(t *testing.T) {
	bridge := NewScanBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:           "trivy-sbom",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})
	registry.Register(&ToolDefinition{
		ID:           "trivy-vuln",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	tests := []struct {
		toolID string
		exists bool
	}{
		{"trivy-sbom", true},
		{"trivy-vuln", true},
		{"trivy-secrets", false}, // not registered
	}

	for _, tt := range tests {
		t.Run(tt.toolID, func(t *testing.T) {
			if got := bridge.HasScannerByToolID(tt.toolID); got != tt.exists {
				t.Errorf("HasScannerByToolID(%q) = %v, want %v", tt.toolID, got, tt.exists)
			}
		})
	}
}

func TestScanBridge_GetAllScannerTools(t *testing.T) {
	bridge := NewScanBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:           "trivy-sbom",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})
	registry.Register(&ToolDefinition{
		ID:           "trivy-vuln",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})
	registry.Register(&ToolDefinition{
		ID:           "trivy-secrets",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})
	registry.Register(&ToolDefinition{
		ID:        "static-site", // Not a scanner
		Type:      ToolTypeContainer,
		LocalPath: "containers/static-site",
		Serve: &ServeConfig{
			ContainerPort: 8080,
		},
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	scannerTools := bridge.GetAllScannerTools()

	// Should include only scanner tools: trivy-sbom, trivy-vuln, trivy-secrets
	if len(scannerTools) != 3 {
		t.Errorf("GetAllScannerTools() returned %d tools, want 3", len(scannerTools))
	}

	// Verify all returned tools are scanners
	for _, tool := range scannerTools {
		if tool.GetScannerCategory() == "" {
			t.Errorf("Tool %q should have a scanner category", tool.ID)
		}
	}
}

func TestScanBridge_GetScannersByCategory(t *testing.T) {
	bridge := NewScanBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:           "trivy-sbom",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})
	registry.Register(&ToolDefinition{
		ID:           "trivy-vuln",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})
	registry.Register(&ToolDefinition{
		ID:           "semgrep",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/semgrep",
		OutputFormat: "json",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Should find sbom scanners
	sbomScanners := bridge.GetScannersByCategory("sbom")
	if len(sbomScanners) != 1 {
		t.Errorf("GetScannersByCategory('sbom') returned %d tools, want 1", len(sbomScanners))
	}

	// Should find sast scanners
	sastScanners := bridge.GetScannersByCategory("sast")
	if len(sastScanners) != 1 {
		t.Errorf("GetScannersByCategory('sast') returned %d tools, want 1", len(sastScanners))
	}

	// Should return empty for unknown category
	unknownScanners := bridge.GetScannersByCategory("unknown")
	if len(unknownScanners) != 0 {
		t.Errorf("GetScannersByCategory('unknown') returned %d tools, want 0", len(unknownScanners))
	}
}

func TestScanBridge_GetScannersForModule_NilModule(t *testing.T) {
	bridge := NewScanBridge()

	scanners := bridge.GetScannersForModule(nil, nil)
	if scanners != nil {
		t.Error("GetScannersForModule(nil, nil) should return nil")
	}
}

func TestScanBridge_GetScannersForModule_NilComponentTypes(t *testing.T) {
	bridge := NewScanBridge()

	module := modules.NewModuleContract(domain.BaseContract{
		Moniker: "test-module",
	}, "/workspace")

	scanners := bridge.GetScannersForModule(module, nil)
	if scanners != nil {
		t.Error("GetScannersForModule with nil componentTypes should return nil")
	}
}

func TestScanBridge_GetScannersForModule(t *testing.T) {
	bridge := NewScanBridge()

	// Create a module with a go component
	module := modules.NewModuleContract(domain.BaseContract{
		Moniker: "test-module",
		Components: domain.ModuleComponents{
			"main": &domain.ComponentEntry{
				Root: ".",
				Type: "go",
			},
		},
	}, "/workspace")

	// Create component types config with scanners (tool IDs now)
	componentTypes := &config.ComponentTypesConfig{
		ComponentTypes: map[string]*config.ComponentType{
			"go": {
				Scanners: []string{"trivy-sbom", "trivy-vuln", "trivy-secrets"},
			},
		},
	}

	scanners := bridge.GetScannersForModule(module, componentTypes)

	// Should return scanner tool IDs for go component
	if len(scanners) == 0 {
		t.Fatal("GetScannersForModule returned empty scanner list")
	}

	// Verify expected scanners
	found := make(map[string]bool)
	for _, toolID := range scanners {
		found[toolID] = true
	}

	if !found["trivy-sbom"] {
		t.Error("trivy-sbom should be in scanner list")
	}
	if !found["trivy-vuln"] {
		t.Error("trivy-vuln should be in scanner list")
	}
	if !found["trivy-secrets"] {
		t.Error("trivy-secrets should be in scanner list")
	}
}

func TestScanBridge_GetScannersForModule_MultipleComponents(t *testing.T) {
	bridge := NewScanBridge()

	// Create a module with multiple components
	module := modules.NewModuleContract(domain.BaseContract{
		Moniker: "test-module",
		Components: domain.ModuleComponents{
			"main": &domain.ComponentEntry{
				Root: ".",
				Type: "go",
			},
			"config": &domain.ComponentEntry{
				Root: "config",
				Type: "dockerfile",
			},
		},
	}, "/workspace")

	// Create component types config with different scanners
	componentTypes := &config.ComponentTypesConfig{
		ComponentTypes: map[string]*config.ComponentType{
			"go": {
				Scanners: []string{"trivy-sbom", "trivy-vuln"},
			},
			"dockerfile": {
				Scanners: []string{"trivy-iac", "trivy-secrets"},
			},
		},
	}

	scanners := bridge.GetScannersForModule(module, componentTypes)

	// Should return union of scanners from both component types
	// Go: trivy-sbom, trivy-vuln
	// Dockerfile: trivy-iac, trivy-secrets
	// Total unique: 4
	if len(scanners) != 4 {
		t.Errorf("GetScannersForModule returned %d scanners, want 4", len(scanners))
	}
}

func TestGlobalScanBridge(t *testing.T) {
	// GlobalScanBridge should return same instance
	bridge1 := GlobalScanBridge()
	bridge2 := GlobalScanBridge()

	if bridge1 != bridge2 {
		t.Error("GlobalScanBridge should return singleton")
	}

	if bridge1 == nil {
		t.Error("GlobalScanBridge should not return nil")
	}
}

func TestScanBridge_Concurrent(t *testing.T) {
	bridge := NewScanBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:           "trivy-sbom",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Test concurrent access
	done := make(chan bool)

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.GetScannerByToolID("trivy-sbom")
		}
		done <- true
	}()

	// Concurrent HasScannerByToolID
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.HasScannerByToolID("trivy-vuln")
		}
		done <- true
	}()

	// Concurrent GetAllScannerTools
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.GetAllScannerTools()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}

func TestScanBridge_IsContainer(t *testing.T) {
	bridge := NewScanBridge()

	// Set up tool system with both container and system tools
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:           "trivy-sbom",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})
	registry.Register(&ToolDefinition{
		ID:     "go-lint",
		Type:   ToolTypeSystem,
		Binary: "golangci-lint",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	if !bridge.IsContainer("trivy-sbom") {
		t.Error("trivy-sbom should be a container tool")
	}
	if bridge.IsContainer("go-lint") {
		t.Error("go-lint should not be a container tool")
	}
	if bridge.IsContainer("nonexistent") {
		t.Error("nonexistent tool should return false")
	}
}

func TestScanBridge_GetScanHandler(t *testing.T) {
	bridge := NewScanBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:           "trivy-sbom",
		Type:         ToolTypeContainer,
		LocalPath:    "containers/trivy",
		OutputFormat: "json",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	handler := bridge.GetScanHandler("trivy-sbom")
	if handler == nil {
		t.Error("GetScanHandler should return a handler for valid tool")
	}

	// Non-existent tool should return nil
	if bridge.GetScanHandler("nonexistent") != nil {
		t.Error("GetScanHandler should return nil for non-existent tool")
	}
}
