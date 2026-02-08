package tool

import (
	"path/filepath"
	"testing"

	tools "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
)

func TestNewToolLoader(t *testing.T) {
	// Use the actual repo root to load the new tools.yml
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		t.Skip("Could not find repo root")
	}

	configRoot := filepath.Join(repoRoot, ".eac")

	loader, err := NewToolLoader(repoRoot, configRoot)
	if err != nil {
		t.Fatalf("NewToolLoader() error = %v", err)
	}

	// Verify bootstrap namespace is loaded
	if !loader.IsNamespaceLoaded(tools.NSBootstrap) {
		t.Error("Expected bootstrap namespace to be loaded")
	}

	// Verify bootstrap tools are available
	bootstrapTools := []string{"go", "docker", "git"}
	for _, id := range bootstrapTools {
		tool, ok := loader.GetTool(id)
		if !ok {
			t.Errorf("Expected bootstrap tool %q to be loaded", id)
			continue
		}
		if tool.ID() != id {
			t.Errorf("Tool ID = %q, want %q", tool.ID(), id)
		}
	}
}

func TestToolLoader_EnsureNamespace(t *testing.T) {
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		t.Skip("Could not find repo root")
	}

	configRoot := filepath.Join(repoRoot, ".eac")

	loader, err := NewToolLoader(repoRoot, configRoot)
	if err != nil {
		t.Fatalf("NewToolLoader() error = %v", err)
	}

	// Go namespace should not be loaded initially
	if loader.IsNamespaceLoaded(tools.NSGo) {
		t.Error("Expected go namespace to not be loaded initially")
	}

	// Load go namespace
	if err := loader.EnsureNamespace(tools.NSGo); err != nil {
		t.Errorf("EnsureNamespace(go) error = %v", err)
	}

	// Now go namespace should be loaded
	if !loader.IsNamespaceLoaded(tools.NSGo) {
		t.Error("Expected go namespace to be loaded after EnsureNamespace")
	}

	// Verify go tools are available
	goTools := loader.ListToolsByNamespace(tools.NSGo)
	if len(goTools) == 0 {
		t.Error("Expected go namespace to have tools")
	}
}

func TestToolLoader_EnsureNamespaceForComponent(t *testing.T) {
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		t.Skip("Could not find repo root")
	}

	configRoot := filepath.Join(repoRoot, ".eac")

	loader, err := NewToolLoader(repoRoot, configRoot)
	if err != nil {
		t.Fatalf("NewToolLoader() error = %v", err)
	}

	tests := []struct {
		componentType string
		wantNamespace tools.Namespace
	}{
		{"go", tools.NSGo},
		{"go-feature", tools.NSGo},
		{"godog", tools.NSGo},
		{"typescript", tools.NSNode},
		{"javascript", tools.NSNode},
		{"book", tools.NSDocs},
		{"design", tools.NSDocs},
		{"unknown", tools.NSBootstrap},
	}

	for _, tt := range tests {
		t.Run(tt.componentType, func(t *testing.T) {
			// Reset loader state for each test
			loader.mu.Lock()
			loader.loaded = make(map[tools.Namespace]bool)
			loader.loaded[tools.NSBootstrap] = true // Keep bootstrap
			loader.mu.Unlock()

			if err := loader.EnsureNamespaceForComponent(tt.componentType); err != nil {
				t.Errorf("EnsureNamespaceForComponent(%q) error = %v", tt.componentType, err)
			}

			if !loader.IsNamespaceLoaded(tt.wantNamespace) {
				t.Errorf("Expected namespace %q to be loaded for component %q", tt.wantNamespace, tt.componentType)
			}
		})
	}
}

func TestToolLoader_GetBinding(t *testing.T) {
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		t.Skip("Could not find repo root")
	}

	configRoot := filepath.Join(repoRoot, ".eac")

	loader, err := NewToolLoader(repoRoot, configRoot)
	if err != nil {
		t.Fatalf("NewToolLoader() error = %v", err)
	}

	tests := []struct {
		toolID string
		want   tools.Binding
	}{
		{"go", tools.BindingSystem},
		{"docker", tools.BindingSystem},
		{"git", tools.BindingSystem},
		{"golangci-lint", tools.BindingAuto}, // Not explicitly bound
		{"unknown", tools.BindingAuto},       // Unknown tool defaults to auto
	}

	for _, tt := range tests {
		t.Run(tt.toolID, func(t *testing.T) {
			got := loader.GetBinding(tt.toolID)
			if got != tt.want {
				t.Errorf("GetBinding(%q) = %v, want %v", tt.toolID, got, tt.want)
			}
		})
	}
}

func TestToolLoader_GetComponentTools(t *testing.T) {
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		t.Skip("Could not find repo root")
	}

	configRoot := filepath.Join(repoRoot, ".eac")

	loader, err := NewToolLoader(repoRoot, configRoot)
	if err != nil {
		t.Fatalf("NewToolLoader() error = %v", err)
	}

	// Test Go component tools
	assignment, ok := loader.GetComponentTools("go")
	if !ok {
		t.Fatal("Expected go component tools to exist")
	}

	if assignment.Builder() != "go" {
		t.Errorf("Go builder = %q, want %q", assignment.Builder(), "go")
	}

	if assignment.Linter() != "golangci-lint" {
		t.Errorf("Go linter = %q, want %q", assignment.Linter(), "golangci-lint")
	}

	scanners := assignment.Scanners()
	if len(scanners) == 0 {
		t.Error("Expected go component to have scanners")
	}
}

func TestToolLoader_LoadAll(t *testing.T) {
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		t.Skip("Could not find repo root")
	}

	configRoot := filepath.Join(repoRoot, ".eac")

	loader, err := NewToolLoader(repoRoot, configRoot)
	if err != nil {
		t.Fatalf("NewToolLoader() error = %v", err)
	}

	// Load all namespaces
	if err := loader.LoadAll(); err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	// All namespaces should be loaded
	namespaces := []tools.Namespace{tools.NSBootstrap, tools.NSGo, tools.NSDocs, tools.NSNode, tools.NSSecurity}
	for _, ns := range namespaces {
		if !loader.IsNamespaceLoaded(ns) {
			t.Errorf("Expected namespace %q to be loaded after LoadAll", ns)
		}
	}

	// Should have many tools loaded
	allTools := loader.ListTools()
	if len(allTools) < 20 {
		t.Errorf("Expected at least 20 tools after LoadAll, got %d", len(allTools))
	}
}

func TestComponentToNamespace(t *testing.T) {
	tests := []struct {
		componentType string
		want          tools.Namespace
	}{
		{"go", tools.NSGo},
		{"go-feature", tools.NSGo},
		{"godog", tools.NSGo},
		{"node", tools.NSNode},
		{"typescript", tools.NSNode},
		{"javascript", tools.NSNode},
		{"book", tools.NSDocs},
		{"base-site", tools.NSDocs},
		{config.ComponentTypePdfRender, tools.NSDocs},
		{"structurizr", tools.NSDocs},
		{"design", tools.NSDocs},
		{"unknown", tools.NSBootstrap},
		{"", tools.NSBootstrap},
	}

	for _, tt := range tests {
		t.Run(tt.componentType, func(t *testing.T) {
			got := tools.ComponentToNamespace(tt.componentType)
			if got != tt.want {
				t.Errorf("ComponentToNamespace(%q) = %v, want %v", tt.componentType, got, tt.want)
			}
		})
	}
}
