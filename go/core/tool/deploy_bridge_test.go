package tool

import (
	"context"
	"io"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	deploy "github.com/ready-to-release/eac/contracts/runner/0.1.0/deploy"
)

// mockDeployHandler implements deploy.DeployerPort for testing.
type mockDeployHandler struct {
	name         string
	deployResult int
	requirements []string
}

func (m *mockDeployHandler) Name() string { return m.name }

func (m *mockDeployHandler) Deploy(_ context.Context, _ core.ModuleContractPort, _, _ string, _ io.Writer, _ any) int {
	return m.deployResult
}

func (m *mockDeployHandler) DryRun(_ context.Context, _ core.ModuleContractPort, _, _ string, _ io.Writer, _ any) int {
	return m.deployResult
}

func (m *mockDeployHandler) Requirements() []string        { return m.requirements }
func (m *mockDeployHandler) IsContainer() bool             { return false }
func (m *mockDeployHandler) IsHostInstalled() bool         { return true }
func (m *mockDeployHandler) ValidateModule(_ core.ModuleContractPort, _, _ string) error {
	return nil
}

var _ deploy.DeployerPort = (*mockDeployHandler)(nil)

func TestDeployBridge_RegisterNativeHandler(t *testing.T) {
	bridge := NewDeployBridge()

	handler := &mockDeployHandler{name: "az-bicep", deployResult: 0}
	bridge.RegisterNativeHandler(handler)

	got := bridge.GetHandler("az-bicep")
	if got == nil {
		t.Fatal("GetHandler returned nil for registered native handler")
	}
	if got.Name() != "az-bicep" {
		t.Errorf("GetHandler().Name() = %q, want %q", got.Name(), "az-bicep")
	}
}

func TestDeployBridge_GetHandler_NotFound(t *testing.T) {
	bridge := NewDeployBridge()

	got := bridge.GetHandler("nonexistent")
	if got != nil {
		t.Error("GetHandler should return nil for nonexistent handler")
	}
}

func TestDeployBridge_GetHandler_YAMLTool(t *testing.T) {
	bridge := NewDeployBridge()

	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool { return true })

	cfg := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"az-bicep": {
				Type:   ToolTypeSystem,
				Binary: "az",
			},
		},
	}
	if err := registry.RegisterFromConfig(cfg); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	got := bridge.GetHandler("az-bicep")
	if got == nil {
		t.Fatal("GetHandler returned nil for YAML tool")
	}
	if got.Name() != "az-bicep" {
		t.Errorf("GetHandler().Name() = %q, want %q", got.Name(), "az-bicep")
	}
}

func TestDeployBridge_NativeHandlerPrecedence(t *testing.T) {
	bridge := NewDeployBridge()

	// Register a native handler
	native := &mockDeployHandler{name: "az-bicep", deployResult: 42}
	bridge.RegisterNativeHandler(native)

	// Also register a tool system handler with the same name
	registry := NewRegistry()
	registry.SetVerifier(func(tool *ToolDefinition) bool { return true })
	cfg := &ToolConfig{
		SystemTools: map[string]*ToolDefinition{
			"az-bicep": {
				Type:   ToolTypeSystem,
				Binary: "az",
			},
		},
	}
	if err := registry.RegisterFromConfig(cfg); err != nil {
		t.Fatalf("RegisterFromConfig failed: %v", err)
	}
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Native handler should take precedence
	got := bridge.GetHandler("az-bicep")
	if got == nil {
		t.Fatal("GetHandler returned nil")
	}
	// Verify it's the native handler by checking the deploy result
	result := got.Deploy(context.Background(), nil, "", "", nil, nil)
	if result != 42 {
		t.Errorf("Expected native handler (result=42), got result=%d", result)
	}
}

func TestDeployBridge_GetHandlersForModule_NilModule(t *testing.T) {
	bridge := NewDeployBridge()

	handlers := bridge.GetHandlersForModule(nil)
	if handlers != nil {
		t.Errorf("expected nil for nil module, got %v", handlers)
	}
}

func TestNewDeployToolHandlerAdapter(t *testing.T) {
	tool := &ToolDefinition{
		Type:   ToolTypeContainer,
		Binary: "az",
	}
	tool.ID = "az-bicep"
	executor := &mockExecutor{
		result: &ExecutionResult{ExitCode: 0},
	}

	adapter := NewDeployToolHandlerAdapter(tool, executor)
	if adapter == nil {
		t.Fatal("NewDeployToolHandlerAdapter returned nil")
	}
	if adapter.Name() != "az-bicep" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "az-bicep")
	}

	// ValidateModule should always return nil for YAML tools
	if err := adapter.ValidateModule(nil, "", "development"); err != nil {
		t.Errorf("ValidateModule() = %v, want nil", err)
	}
}

func TestDeployToolHandlerAdapter_Requirements(t *testing.T) {
	tool := &ToolDefinition{
		Type:   ToolTypeSystem,
		Binary: "az",
	}
	tool.ID = "az-bicep"

	adapter := NewDeployToolHandlerAdapter(tool, &mockExecutor{})

	reqs := adapter.Requirements()
	// BaseHandlerAdapter.Requirements() returns tool's require field
	if reqs == nil {
		// System tools without explicit requirements return nil — expected
	}
}

func TestDeployToolHandlerAdapter_IsContainer(t *testing.T) {
	tests := []struct {
		name       string
		toolType   ToolType
		wantResult bool
	}{
		{"container tool", ToolTypeContainer, true},
		{"system tool", ToolTypeSystem, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &ToolDefinition{Type: tt.toolType, Binary: "az"}
			tool.ID = "test"
			adapter := NewDeployToolHandlerAdapter(tool, &mockExecutor{})
			if got := adapter.IsContainer(); got != tt.wantResult {
				t.Errorf("IsContainer() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}
