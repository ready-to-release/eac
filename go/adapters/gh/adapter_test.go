package gh

import (
	"runtime"
	"testing"

	"github.com/ready-to-release/eac/go/core/tool"
)

func TestNew(t *testing.T) {
	ts := tool.NewToolSystemForTesting()
	adapter := New(ts, "/workspace")

	if adapter == nil {
		t.Fatal("expected non-nil adapter")
	}
	if adapter.workDir != "/workspace" {
		t.Fatalf("expected workDir '/workspace', got %q", adapter.workDir)
	}
	if adapter.ts != ts {
		t.Fatal("expected ToolSystem to be injected")
	}
}

func TestAdapter_ImplementsCLIExecutor(t *testing.T) {
	// Compile-time check is in adapter.go via var _ github.CLIExecutor = (*Adapter)(nil)
	// This test just verifies New returns a usable adapter.
	ts := tool.NewToolSystemForTesting()
	adapter := New(ts, ".")

	// Register a mock "gh" tool. Use "go" as the binary since it's guaranteed
	// to exist in Go test environments. On Windows, "echo" is a shell builtin
	// and not available as a standalone executable.
	ts.Registry.Register(&tool.ToolDefinition{
		ID:     "gh",
		Type:   tool.ToolTypeSystem,
		Binary: "go",
	})

	// Execute with "version" arg — "go version" always succeeds and produces output
	output, err := adapter.Exec("version")
	if err != nil {
		// On Windows CI, go may not be in PATH in all contexts; skip gracefully
		if runtime.GOOS == "windows" {
			t.Skipf("skipping: go binary not accessible: %v", err)
		}
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output) == 0 {
		t.Fatal("expected non-empty output from go version")
	}
}
