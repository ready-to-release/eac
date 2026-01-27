package servers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/tool"
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
	registry, _, err := tool.InitializeFromConfig(repoRoot, "")
	if err != nil {
		t.Fatalf("failed to initialize tool system: %v", err)
	}
	tool.SetGlobalRegistry(registry)

	// Should have static-site from tool-config.yml
	if !HasServer(ServerStaticSite) {
		t.Error("HasServer should return true for static-site server")
	}
}

func TestGetAllServerTypes(t *testing.T) {
	// Initialize tool system from default config
	repoRoot := findRepoRoot(t)
	registry, _, err := tool.InitializeFromConfig(repoRoot, "")
	if err != nil {
		t.Fatalf("failed to initialize tool system: %v", err)
	}
	tool.SetGlobalRegistry(registry)

	types := GetAllServerTypes()
	if len(types) == 0 {
		t.Error("GetAllServerTypes should return at least one type")
	}

	// Verify static-site is in the list
	found := false
	for _, st := range types {
		if st == ServerStaticSite {
			found = true
			break
		}
	}
	if !found {
		t.Error("ServerStaticSite should be in GetAllServerTypes")
	}
}

func TestGetServerToolID(t *testing.T) {
	toolID := GetServerToolID(ServerStaticSite)
	if toolID != "static-site" {
		t.Errorf("GetServerToolID(ServerStaticSite) = %q, want %q", toolID, "static-site")
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
	// Verify constants have expected values
	if ServerStaticSite != "static-site" {
		t.Errorf("ServerStaticSite = %q, want %q", ServerStaticSite, "static-site")
	}
	if ServerMkDocsLive != "mkdocs-live" {
		t.Errorf("ServerMkDocsLive = %q, want %q", ServerMkDocsLive, "mkdocs-live")
	}
	if ServerStructurizr != "structurizr" {
		t.Errorf("ServerStructurizr = %q, want %q", ServerStructurizr, "structurizr")
	}
}

func TestAllServersRegistered(t *testing.T) {
	// Initialize tool system from default config
	repoRoot := findRepoRoot(t)
	registry, _, err := tool.InitializeFromConfig(repoRoot, "")
	if err != nil {
		t.Fatalf("failed to initialize tool system: %v", err)
	}
	tool.SetGlobalRegistry(registry)

	// All three servers should be available via tool-config.yml
	servers := []ServerType{
		ServerStaticSite,
		ServerMkDocsLive,
		ServerStructurizr,
	}

	for _, st := range servers {
		t.Run(string(st), func(t *testing.T) {
			if !HasServer(st) {
				t.Errorf("Server %q should be registered", st)
			}
			server := GetServer(st)
			if server == nil {
				t.Errorf("GetServer(%q) should not return nil", st)
			}
		})
	}
}
