package tool

import (
	"testing"
)

func TestNewToolSystemForTesting(t *testing.T) {
	ts := NewToolSystemForTesting()
	if ts == nil {
		t.Fatal("expected non-nil ToolSystem")
	}
	if ts.Registry == nil {
		t.Fatal("expected non-nil Registry")
	}
	if ts.Executor == nil {
		t.Fatal("expected non-nil Executor")
	}
	if ts.Config == nil {
		t.Fatal("expected non-nil Config")
	}
	if ts.Resolver == nil {
		t.Fatal("expected non-nil Resolver")
	}
	if ts.BuildBridge == nil {
		t.Fatal("expected non-nil BuildBridge")
	}
	if ts.LintBridge == nil {
		t.Fatal("expected non-nil LintBridge")
	}
	if ts.TestBridge == nil {
		t.Fatal("expected non-nil TestBridge")
	}
	if ts.ScanBridge == nil {
		t.Fatal("expected non-nil ScanBridge")
	}
	if ts.ServeBridge == nil {
		t.Fatal("expected non-nil ServeBridge")
	}
	if ts.HandlerBridge == nil {
		t.Fatal("expected non-nil HandlerBridge")
	}
	if ts.Categories == nil {
		t.Fatal("expected non-nil Categories")
	}
}

func TestToolSystem_GetOrAdhoc(t *testing.T) {
	ts := NewToolSystemForTesting()

	// Ad-hoc tool for unregistered binary
	tool := ts.GetOrAdhoc("git")
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.ID != "git" {
		t.Fatalf("expected ID 'git', got %q", tool.ID)
	}
	if tool.Type != ToolTypeSystem {
		t.Fatalf("expected system type, got %q", tool.Type)
	}

	// Register a tool and verify it's returned
	registered := &ToolDefinition{ID: "custom", Type: ToolTypeSystem, Binary: "custom"}
	if err := ts.Registry.Register(registered); err != nil {
		t.Fatalf("failed to register: %v", err)
	}
	got := ts.GetOrAdhoc("custom")
	if got == nil {
		t.Fatal("expected non-nil tool")
	}
	if got.Binary != "custom" {
		t.Fatalf("expected binary 'custom', got %q", got.Binary)
	}
}

func TestToolSystem_GetToolImage(t *testing.T) {
	ts := NewToolSystemForTesting()

	// Unregistered tool returns empty
	if img := ts.GetToolImage("nonexistent"); img != "" {
		t.Fatalf("expected empty string, got %q", img)
	}

	// Register container tool with image
	tool := &ToolDefinition{
		ID:    "scanner",
		Type:  ToolTypeContainer,
		Image: "ghcr.io/org/scanner",
		Tag:   "v1.0",
	}
	if err := ts.Registry.Register(tool); err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	img := ts.GetToolImage("scanner")
	if img != "ghcr.io/org/scanner:v1.0" {
		t.Fatalf("expected 'ghcr.io/org/scanner:v1.0', got %q", img)
	}
}

func TestToolSystem_GetToolImageWithDefault(t *testing.T) {
	ts := NewToolSystemForTesting()

	// Unregistered tool returns default
	img := ts.GetToolImageWithDefault("nonexistent", "default:latest")
	if img != "default:latest" {
		t.Fatalf("expected 'default:latest', got %q", img)
	}
}

func TestToolSystem_GetTestTypeComponentType(t *testing.T) {
	ts := NewToolSystemForTesting()

	tests := []struct {
		testType string
		want     string
	}{
		{"gotest", "go"},
		{"godog", "gherkin"},
		{"mocha", "typescript"},
		{"unknown", "go"}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.testType, func(t *testing.T) {
			got := ts.GetTestTypeComponentType(tt.testType)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}

	// Test with config mapping override
	ts.Config.TestTypeMapping = map[string]string{
		"gotest": "override-type",
	}
	got := ts.GetTestTypeComponentType("gotest")
	if got != "override-type" {
		t.Fatalf("expected 'override-type', got %q", got)
	}
}

func TestToolSystem_FilterPlatformSupported(t *testing.T) {
	ts := NewToolSystemForTesting()

	// Unregistered tools are included
	result := ts.FilterPlatformSupported([]string{"unknown-a", "unknown-b"})
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestToolSystem_Categories(t *testing.T) {
	ts := NewToolSystemForTesting()

	// Verify scanner category resolution
	toolID := ts.ScannerToolIDForCategory("sbom")
	if toolID != ToolTrivySBOM {
		t.Fatalf("expected %q, got %q", ToolTrivySBOM, toolID)
	}

	// Unknown category returns empty
	unknown := ts.ScannerToolIDForCategory("nonexistent")
	if unknown != "" {
		t.Fatalf("expected empty, got %q", unknown)
	}

	// Server type resolution
	serverID := ts.ServerToolIDForType(ToolStaticSite)
	if serverID != ToolStaticSite {
		t.Fatalf("expected %q, got %q", ToolStaticSite, serverID)
	}
}

func TestToolSystem_Close(t *testing.T) {
	ts := NewToolSystemForTesting()
	if err := ts.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
