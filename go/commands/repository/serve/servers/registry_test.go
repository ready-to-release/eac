package servers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/tool"
)

// findRepoRoot finds the repository root by looking for go.work
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root (go.work)")
		}
		dir = parent
	}
}

func TestHasServer(t *testing.T) {
	// Initialize tool system from default config
	repoRoot := findRepoRoot(t)
	registry, _, _, err := tool.InitializeFromConfig(repoRoot, "")
	if err != nil {
		t.Fatalf("failed to initialize tool system: %v", err)
	}
	tool.SetGlobalRegistry(registry)

	// Should have static-site from tool-config.yml
	if !HasServer(tool.ToolStaticSite) {
		t.Error("HasServer should return true for static-site server")
	}
}

func TestGetAllServableTools(t *testing.T) {
	// Initialize tool system from default config
	repoRoot := findRepoRoot(t)
	registry, _, _, err := tool.InitializeFromConfig(repoRoot, "")
	if err != nil {
		t.Fatalf("failed to initialize tool system: %v", err)
	}
	tool.SetGlobalRegistry(registry)

	tools := GetAllServableTools()
	if len(tools) == 0 {
		t.Error("GetAllServableTools should return at least one tool")
	}

	// Verify static-site is in the list
	found := false
	for _, tool := range tools {
		if tool.GetServerType() == "static-site" {
			found = true
			break
		}
	}
	if !found {
		t.Error("static-site should be in GetAllServableTools")
	}
}

// TestServerToolConstants removed - use tool.ToolXxx constants from go/eac/core/tool/ids.go.
// See go/eac/core/tool/ids_test.go for server constant tests.

func TestAllServersRegistered(t *testing.T) {
	// Initialize tool system from default config
	repoRoot := findRepoRoot(t)
	registry, _, _, err := tool.InitializeFromConfig(repoRoot, "")
	if err != nil {
		t.Fatalf("failed to initialize tool system: %v", err)
	}
	tool.SetGlobalRegistry(registry)

	// All three servers should be available via tool-config.yml
	servers := []string{
		tool.ToolStaticSite,
		tool.ToolMkDocsLive,
		tool.ToolStructurizrLite,
	}

	for _, toolID := range servers {
		t.Run(toolID, func(t *testing.T) {
			if !HasServer(toolID) {
				t.Errorf("Server %q should be registered", toolID)
			}
			server := GetServer(toolID)
			if server == nil {
				t.Errorf("GetServer(%q) should not return nil", toolID)
			}
		})
	}
}

func TestGetServerToolByID(t *testing.T) {
	// Initialize tool system from default config
	repoRoot := findRepoRoot(t)
	registry, _, _, err := tool.InitializeFromConfig(repoRoot, "")
	if err != nil {
		t.Fatalf("failed to initialize tool system: %v", err)
	}
	tool.SetGlobalRegistry(registry)

	serverTool := GetServerToolByID(tool.ToolStaticSite)
	if serverTool == nil {
		t.Fatal("GetServerToolByID should return a tool for static-site")
	}
	if !serverTool.IsServable() {
		t.Error("static-site tool should be servable")
	}
}
