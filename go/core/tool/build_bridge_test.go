package tool

import (
	"fmt"
	"io"
	"sync"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// lookPathNotFound simulates no binaries available on PATH.
func lookPathNotFound(name string) (string, error) {
	return "", fmt.Errorf("%s: not found", name)
}

// lookPathAllFound simulates all binaries available on PATH.
func lookPathAllFound(name string) (string, error) {
	return "/usr/bin/" + name, nil
}

// mockBuildHandler implements build.BuilderPort for testing.
type mockBuildHandler struct {
	name         string
	buildResult  int
	artifacts    []string
	requirements []string
}

func (m *mockBuildHandler) Name() string { return m.name }

func (m *mockBuildHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts any) int {
	return m.buildResult
}

func (m *mockBuildHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return m.artifacts
}

func (m *mockBuildHandler) Requirements() []string {
	return m.requirements
}

func (m *mockBuildHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	return nil
}

func (m *mockBuildHandler) IsContainer() bool     { return false }
func (m *mockBuildHandler) IsHostInstalled() bool { return true }

func TestBuildBridge_GetHandler_YAMLTool(t *testing.T) {
	bridge := NewBuildBridge()

	// Set up tool system using RegisterFromConfig (canonical registration)
	registry := NewRegistry()
	// Mock verifier: all tools are available
	registry.SetVerifier(func(tool *ToolDefinition) bool { return true })

	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"custom-builder": {
				Type:   ToolTypeSystem,
				Binary: "custom",
			},
		},
	}
	if err := registry.RegisterFromConfig(config); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Use canonical name for lookup
	got := bridge.GetHandler("custom-builder")
	if got == nil {
		t.Fatal("GetHandler returned nil for YAML tool")
	}
	// Tool ID is now canonical (no suffix) - type is in tool.Type field
	if got.Name() != "custom-builder" {
		t.Errorf("GetHandler().Name() = %q, want %q", got.Name(), "custom-builder")
	}
}

func TestBuildBridge_GetHandler_NotFound(t *testing.T) {
	bridge := NewBuildBridge()

	got := bridge.GetHandler("nonexistent")
	if got != nil {
		t.Error("GetHandler should return nil for nonexistent handler")
	}
}

func TestBuildBridge_GetAllHandlers(t *testing.T) {
	bridge := NewBuildBridge()

	// Set up tool system with tools using RegisterFromConfig
	registry := NewRegistry()
	// Mock verifier: all tools are available
	registry.SetVerifier(func(tool *ToolDefinition) bool { return true })

	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"mkdocs": {
				Type:   ToolTypeSystem,
				Binary: "mkdocs",
			},
			"go": {
				Type:   ToolTypeSystem,
				Binary: "go",
			},
		},
	}
	if err := registry.RegisterFromConfig(config); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	handlers := bridge.GetAllHandlers()

	// Should have: go:system, mkdocs:system
	if len(handlers) != 2 {
		t.Errorf("GetAllHandlers() returned %d handlers, want 2", len(handlers))
	}
}

func TestBuildBridge_HasHandler(t *testing.T) {
	bridge := NewBuildBridge()

	// Set up tool system using RegisterFromConfig
	registry := NewRegistry()
	// Mock verifier: all tools are available
	registry.SetVerifier(func(tool *ToolDefinition) bool { return true })

	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"mkdocs": {
				Type:   ToolTypeSystem,
				Binary: "mkdocs",
			},
		},
	}
	if err := registry.RegisterFromConfig(config); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	tests := []struct {
		name   string
		exists bool
	}{
		{"mkdocs", true},  // canonical name lookup
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bridge.HasHandler(tt.name); got != tt.exists {
				t.Errorf("HasHandler(%q) = %v, want %v", tt.name, got, tt.exists)
			}
		})
	}
}

func TestBuildBridge_GetHandlersForModule_NilModule(t *testing.T) {
	bridge := NewBuildBridge()

	handlers := bridge.GetHandlersForModule(nil)
	if handlers != nil {
		t.Error("GetHandlersForModule(nil) should return nil")
	}
}

func TestGlobalBuildBridge(t *testing.T) {
	// GlobalBuildBridge should return same instance
	bridge1 := GlobalBuildBridge()
	bridge2 := GlobalBuildBridge()

	if bridge1 != bridge2 {
		t.Error("GlobalBuildBridge should return singleton")
	}

	if bridge1 == nil {
		t.Error("GlobalBuildBridge should not return nil")
	}
}

func TestBuildBridge_GetHandlersForModule(t *testing.T) {
	bridge := NewBuildBridge()

	// Create module
	module := modules.NewModuleContract(domain.BaseContract{
		Moniker: "test-module",
		Components: config.ModuleComponents{
			"go": &config.ComponentEntry{Root: "."},
		},
	}, "/workspace")

	// Without config.Global() set, we expect empty results
	handlers := bridge.GetHandlersForModule(module)
	if len(handlers) != 0 {
		t.Logf("Got %d handlers (expected 0 without global config)", len(handlers))
	}
}

func TestNewBuildBridge(t *testing.T) {
	bridge := NewBuildBridge()

	if bridge == nil {
		t.Fatal("NewBuildBridge returned nil")
	}
}

func TestBuildBridge_SetToolSystem(t *testing.T) {
	bridge := NewBuildBridge()

	registry := NewRegistry()
	// Mock verifier: all tools are available
	registry.SetVerifier(func(tool *ToolDefinition) bool { return true })

	resolver := NewResolver(registry)
	executor := &mockExecutor{}

	bridge.SetToolSystem(registry, resolver, executor)

	// Verify tool system is set by registering a tool and retrieving it
	// Use RegisterFromConfig for canonical registration
	config := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"test-tool": {
				Type:   ToolTypeSystem,
				Binary: "test",
			},
		},
	}
	if err := registry.RegisterFromConfig(config); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}

	handler := bridge.GetHandler("test-tool")
	if handler == nil {
		t.Error("Tool system not properly configured")
	}
}

func TestBuildBridge_NativeHandlerSkippedWhenRequirementsNotMet(t *testing.T) {
	bridge := NewBuildBridge()
	bridge.lookPathFunc = lookPathNotFound // npm not on PATH

	// Register native handler that requires "npm"
	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "npm-build",
		requirements: []string{"npm"},
	})

	// Create registry with only container tool (npm not available via verifier)
	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		return false // Nothing available on system
	})
	if err := registry.RegisterFromConfig(&ToolConfig{
		ContainerTools: map[string]*ToolDefinition{
			"npm-build": {Type: ToolTypeContainer, Image: "node", Tag: "22-alpine"},
		},
	}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	handler := bridge.GetHandler("npm-build")
	if handler == nil {
		t.Fatal("expected container handler fallback, got nil")
	}
	if !handler.IsContainer() {
		t.Error("expected container handler, got native")
	}
}

func TestBuildBridge_NativeHandlerUsedWhenRequirementsMet(t *testing.T) {
	bridge := NewBuildBridge()
	bridge.lookPathFunc = lookPathAllFound // npm IS on PATH

	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "npm-build",
		requirements: []string{"npm"},
	})

	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		return true // Everything available
	})
	if err := registry.RegisterFromConfig(&ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"npm": {Type: ToolTypeSystem, Binary: "npm"},
		},
		ContainerTools: map[string]*ToolDefinition{
			"npm-build": {Type: ToolTypeContainer, Image: "node", Tag: "22-alpine"},
		},
	}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	handler := bridge.GetHandler("npm-build")
	if handler == nil {
		t.Fatal("expected native handler, got nil")
	}
	if handler.IsContainer() {
		t.Error("expected native handler, got container")
	}
}

func TestBuildBridge_ContainerModeSkipsNativeHandlers(t *testing.T) {
	bridge := NewBuildBridge()

	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "npm-build",
		requirements: []string{"npm"},
	})

	registry := NewRegistry()
	if err := registry.RegisterFromConfig(&ToolConfig{
		ExecutorMode: ExecutorModeContainer,
		ContainerTools: map[string]*ToolDefinition{
			"npm-build": {Type: ToolTypeContainer, Image: "node", Tag: "22-alpine"},
		},
	}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	handler := bridge.GetHandler("npm-build")
	if handler == nil {
		t.Fatal("expected container handler in container mode, got nil")
	}
	if !handler.IsContainer() {
		t.Error("expected container handler in container mode, got native")
	}
}

func TestBuildBridge_SystemModeAlwaysUsesNativeHandler(t *testing.T) {
	bridge := NewBuildBridge()
	bridge.lookPathFunc = lookPathNotFound // npm not on PATH, but system mode forces native

	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "npm-build",
		requirements: []string{"npm"},
	})

	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		return false // npm not available
	})
	if err := registry.RegisterFromConfig(&ToolConfig{
		ExecutorMode: ExecutorModeSystem,
		SystemTools: map[string]*ToolDefinition{
			"npm-build": {Type: ToolTypeSystem, Binary: "npm"},
		},
	}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	handler := bridge.GetHandler("npm-build")
	if handler == nil {
		t.Fatal("expected native handler in system mode, got nil")
	}
	if handler.IsContainer() {
		t.Error("system mode should force native handler, got container")
	}
}

func TestBuildBridge_NativeHandlerNoRequirementsAlwaysReturned(t *testing.T) {
	bridge := NewBuildBridge()

	// Handler with no requirements (e.g., DependencyGraphHandler)
	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "dependency-graph",
		requirements: nil,
	})

	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		return false // Nothing available
	})
	if err := registry.RegisterFromConfig(&ToolConfig{}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	handler := bridge.GetHandler("dependency-graph")
	if handler == nil {
		t.Fatal("expected native handler with no requirements, got nil")
	}
	if handler.IsContainer() {
		t.Error("expected native handler, got container")
	}
}

func TestBuildBridge_PerToolBindingOverride(t *testing.T) {
	bridge := NewBuildBridge()

	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "npm-build",
		requirements: []string{"npm"},
	})

	// Global mode is auto, but npm-build has per-tool container binding
	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		return true // npm IS available on system
	})
	if err := registry.RegisterFromConfig(&ToolConfig{
		ExecutorMode: ExecutorModeAuto,
		SystemTools: map[string]*ToolDefinition{
			"npm": {Type: ToolTypeSystem, Binary: "npm"},
		},
		ContainerTools: map[string]*ToolDefinition{
			"npm-build": {Type: ToolTypeContainer, Image: "node", Tag: "22-alpine"},
		},
		ToolBindings: map[string]ToolBinding{
			"npm-build": ToolBindingContainer, // Force container for npm-build
		},
	}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	handler := bridge.GetHandler("npm-build")
	if handler == nil {
		t.Fatal("expected container handler via per-tool binding, got nil")
	}
	if !handler.IsContainer() {
		t.Error("per-tool container binding should skip native handler")
	}
}

func TestBuildBridge_GetAllHandlers_RespectsAvailabilityGating(t *testing.T) {
	bridge := NewBuildBridge()
	bridge.lookPathFunc = lookPathNotFound // npm not on PATH

	// Register two native handlers: one with met requirements, one without
	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "go-build",
		requirements: nil, // No requirements = always available
	})
	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "npm-build",
		requirements: []string{"npm"},
	})

	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		return false // Nothing available on system
	})
	if err := registry.RegisterFromConfig(&ToolConfig{
		ContainerTools: map[string]*ToolDefinition{
			"npm-build": {Type: ToolTypeContainer, Image: "node", Tag: "22-alpine"},
		},
	}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	handlers := bridge.GetAllHandlers()

	// go-build: native (no requirements)
	if h, ok := handlers["go-build"]; !ok {
		t.Error("go-build should be in GetAllHandlers()")
	} else if h.IsContainer() {
		t.Error("go-build should be native (no requirements)")
	}

	// npm-build: container (requirements not met → native skipped)
	if h, ok := handlers["npm-build"]; !ok {
		t.Error("npm-build should be in GetAllHandlers() via container fallback")
	} else if !h.IsContainer() {
		t.Error("npm-build should be container (requirements not met)")
	}
}

func TestBuildBridge_HasHandler_RespectsAvailabilityGating(t *testing.T) {
	bridge := NewBuildBridge()
	bridge.lookPathFunc = lookPathNotFound // npm not on PATH

	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "npm-build",
		requirements: []string{"npm"},
	})

	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		return false // npm not available
	})
	if err := registry.RegisterFromConfig(&ToolConfig{
		ContainerTools: map[string]*ToolDefinition{
			"npm-build": {Type: ToolTypeContainer, Image: "node", Tag: "22-alpine"},
		},
	}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Should still report true (container fallback available)
	if !bridge.HasHandler("npm-build") {
		t.Error("HasHandler should return true (container fallback exists)")
	}
}

func TestBuildBridge_HasHandler_FalseWhenNeitherAvailable(t *testing.T) {
	bridge := NewBuildBridge()
	bridge.lookPathFunc = lookPathNotFound // npm not on PATH

	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "npm-build",
		requirements: []string{"npm"},
	})

	// Registry with nothing for npm-build
	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		return false
	})
	if err := registry.RegisterFromConfig(&ToolConfig{}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Native requirements not met AND no registry fallback → false
	if bridge.HasHandler("npm-build") {
		t.Error("HasHandler should return false (no available handler)")
	}
}

func TestBuildBridge_NativeHandlerSkippedWhenRegistryRejectsVersion(t *testing.T) {
	bridge := NewBuildBridge()
	bridge.lookPathFunc = lookPathAllFound // npm IS on PATH (tier 1 passes)

	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "npm-build",
		requirements: []string{"npm"},
	})

	// Registry has npm:system but verifier rejects it (wrong version)
	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		return false // npm version check fails (tier 2 rejects)
	})
	if err := registry.RegisterFromConfig(&ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"npm": {Type: ToolTypeSystem, Binary: "npm"},
		},
		ContainerTools: map[string]*ToolDefinition{
			"npm-build": {Type: ToolTypeContainer, Image: "node", Tag: "22-alpine"},
		},
	}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	handler := bridge.GetHandler("npm-build")
	if handler == nil {
		t.Fatal("expected container handler fallback, got nil")
	}
	if !handler.IsContainer() {
		t.Error("expected container handler when registry rejects version, got native")
	}
}

func TestBuildBridge_PreWarmRequirements(t *testing.T) {
	bridge := NewBuildBridge()
	bridge.lookPathFunc = lookPathAllFound

	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "npm-build",
		requirements: []string{"npm"},
	})
	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "go-build",
		requirements: []string{"go"},
	})

	var verified []string
	var mu sync.Mutex

	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		mu.Lock()
		verified = append(verified, tool.ID)
		mu.Unlock()
		return true
	})
	if err := registry.RegisterFromConfig(&ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"npm": {Type: ToolTypeSystem, Binary: "npm"},
			"go":  {Type: ToolTypeSystem, Binary: "go"},
		},
	}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	bridge.PreWarmRequirements()

	mu.Lock()
	defer mu.Unlock()
	if len(verified) != 2 {
		t.Errorf("expected 2 verifications, got %d: %v", len(verified), verified)
	}

	// After pre-warm, IsAvailable should return cached results
	if !registry.IsAvailable("npm") {
		t.Error("npm should be available after pre-warm")
	}
	if !registry.IsAvailable("go") {
		t.Error("go should be available after pre-warm")
	}
}

func TestBuildBridge_PreWarmRequirements_SkipsMissingBinaries(t *testing.T) {
	bridge := NewBuildBridge()
	bridge.lookPathFunc = func(name string) (string, error) {
		if name == "go" {
			return "/usr/bin/go", nil
		}
		return "", fmt.Errorf("%s: not found", name)
	}

	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "npm-build",
		requirements: []string{"npm"}, // Not on PATH
	})
	bridge.RegisterNativeHandler(&mockBuildHandler{
		name:         "go-build",
		requirements: []string{"go"}, // On PATH
	})

	var verified []string
	var mu sync.Mutex

	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		mu.Lock()
		verified = append(verified, tool.ID)
		mu.Unlock()
		return true
	})
	if err := registry.RegisterFromConfig(&ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"npm": {Type: ToolTypeSystem, Binary: "npm"},
			"go":  {Type: ToolTypeSystem, Binary: "go"},
		},
	}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	bridge.PreWarmRequirements()

	mu.Lock()
	defer mu.Unlock()
	// Only "go" should be verified (npm was fast-rejected by LookPath)
	if len(verified) != 1 {
		t.Errorf("expected 1 verification (npm skipped), got %d: %v", len(verified), verified)
	}
	if len(verified) > 0 && verified[0] != "go" {
		t.Errorf("expected 'go' to be verified, got %q", verified[0])
	}
}

func TestRegistry_IsAvailable_UsesCachePath(t *testing.T) {
	registry := NewRegistry()

	callCount := 0
	registry.SetVerifier(func(tool *ToolDefinition) bool {
		callCount++
		return true
	})

	if err := registry.RegisterFromConfig(&ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"npm": {Type: ToolTypeSystem, Binary: "npm"},
		},
	}); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}

	// First call should verify exactly once (no double verification)
	if !registry.IsAvailable("npm") {
		t.Error("npm should be available")
	}
	if callCount != 1 {
		t.Errorf("expected 1 verification call, got %d (double-verify bug?)", callCount)
	}

	// Second call should use cache (no additional verification)
	if !registry.IsAvailable("npm") {
		t.Error("npm should still be available (cached)")
	}
	if callCount != 1 {
		t.Errorf("expected 1 verification call (cached), got %d", callCount)
	}
}

func TestBuildBridge_GetToolForComponent(t *testing.T) {
	tests := []struct {
		name          string
		componentType string
		setupRegistry func(r *DefaultRegistry)
		setupResolver func(r Registry) *DefaultResolver
		wantToolID    string
		wantCPUs      int
		wantNil       bool
	}{
		{
			name:          "returns tool from registry by direct lookup",
			componentType: "go-build", // Direct tool name
			setupRegistry: func(r *DefaultRegistry) {
				config := &ToolConfig{
					SystemTools: map[string]*ToolDefinition{
						"go-build": {
							Type:   ToolTypeSystem,
							Binary: "go",
							Resources: &ToolResources{
								CPUs: 2,
							},
						},
					},
				}
				if err := r.RegisterFromConfig(config); err != nil {
					t.Fatalf("RegisterFromConfig failed: %v", err)
				}
			},
			setupResolver: nil,
			wantToolID:    "go-build",
			wantCPUs:      2,
		},
		{
			name:          "returns tool with resolver component-tools mapping",
			componentType: "typescript",
			setupRegistry: func(r *DefaultRegistry) {
				config := &ToolConfig{
					ContainerTools: map[string]*ToolDefinition{
						"npm-build": {
							Type:      ToolTypeContainer,
							Image:     "node",
							Tag:       "22-alpine",
							Container: "node",
							Resources: &ToolResources{
								CPUs:   4,
								Memory: "4g",
							},
						},
					},
					ComponentTools: map[string]*ToolAssignment{
						"typescript": {
							Builder: "npm-build",
						},
					},
				}
				if err := r.RegisterFromConfig(config); err != nil {
					t.Fatalf("RegisterFromConfig failed: %v", err)
				}
			},
			setupResolver: func(r Registry) *DefaultResolver {
				resolver := NewResolver(r)
				// Load component-tools mapping
				resolver.LoadProjectConfig(map[string]*ToolAssignment{
					"typescript": {Builder: "npm-build"},
				})
				return resolver
			},
			wantToolID: "npm-build",
			wantCPUs:   4,
		},
		{
			name:          "returns nil for unknown component type",
			componentType: "unknown",
			setupRegistry: func(r *DefaultRegistry) {},
			setupResolver: nil,
			wantNil:       true,
		},
		{
			name:          "returns tool with default resources (CPUs=0)",
			componentType: "simple-tool",
			setupRegistry: func(r *DefaultRegistry) {
				config := &ToolConfig{
					SystemTools: map[string]*ToolDefinition{
						"simple-tool": {
							Type:   ToolTypeSystem,
							Binary: "simple",
							// No Resources specified
						},
					},
				}
				if err := r.RegisterFromConfig(config); err != nil {
					t.Fatalf("RegisterFromConfig failed: %v", err)
				}
			},
			setupResolver: nil,
			wantToolID:    "simple-tool",
			wantCPUs:      0, // nil resources = 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewBuildBridge()
			registry := NewRegistry()
			registry.SetVerifier(func(tool *ToolDefinition) bool { return true })

			if tt.setupRegistry != nil {
				tt.setupRegistry(registry)
			}

			var resolver *DefaultResolver
			if tt.setupResolver != nil {
				resolver = tt.setupResolver(registry)
			}

			bridge.SetToolSystem(registry, resolver, &mockExecutor{})

			got := bridge.GetToolForComponent(tt.componentType)

			if tt.wantNil {
				if got != nil {
					t.Errorf("GetToolForComponent(%q) = %v, want nil", tt.componentType, got)
				}
				return
			}

			if got == nil {
				t.Fatalf("GetToolForComponent(%q) = nil, want tool", tt.componentType)
			}

			if got.ID != tt.wantToolID {
				t.Errorf("GetToolForComponent(%q).ID = %q, want %q", tt.componentType, got.ID, tt.wantToolID)
			}

			gotCPUs := 0
			if got.Resources != nil {
				gotCPUs = got.Resources.CPUs
			}
			if gotCPUs != tt.wantCPUs {
				t.Errorf("GetToolForComponent(%q).Resources.CPUs = %d, want %d", tt.componentType, gotCPUs, tt.wantCPUs)
			}
		})
	}
}
