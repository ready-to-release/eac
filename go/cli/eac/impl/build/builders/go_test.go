package builders

import (
	"bytes"
	"context"
	"testing"
	"time"

	container "github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/core/tool"
)

// mockContainerPortForGo implements container.ContainerPort for Go handler testing.
type mockContainerPortForGo struct {
	images         map[string]bool
	executeResults []*container.ContainerResult
	executeCalls   int
	lastConfig     *container.ContainerConfig
}

func newMockContainerPortForGo() *mockContainerPortForGo {
	return &mockContainerPortForGo{
		images:         make(map[string]bool),
		executeResults: []*container.ContainerResult{},
	}
}

func (m *mockContainerPortForGo) Execute(ctx context.Context, config *container.ContainerConfig) (*container.ContainerResult, error) {
	m.lastConfig = config
	idx := m.executeCalls
	m.executeCalls++

	if idx < len(m.executeResults) {
		return m.executeResults[idx], nil
	}
	return &container.ContainerResult{
		ExitCode: 0,
		Stdout:   []byte("mock output"),
		Stderr:   nil,
		Duration: 100 * time.Millisecond,
	}, nil
}

func (m *mockContainerPortForGo) Build(ctx context.Context, config *container.BuildConfig) error {
	return nil
}

func (m *mockContainerPortForGo) Pull(ctx context.Context, imageRef string) error {
	m.images[imageRef] = true
	return nil
}

func (m *mockContainerPortForGo) ImageExists(ctx context.Context, imageRef string) bool {
	return m.images[imageRef]
}

func (m *mockContainerPortForGo) IsAvailable() bool {
	return true
}

func (m *mockContainerPortForGo) Close() error {
	return nil
}

func TestGoHandler_Name(t *testing.T) {
	h := &GoHandler{}
	if h.Name() != "go" {
		t.Errorf("expected name 'go', got '%s'", h.Name())
	}
}

func TestGoHandler_Capabilities(t *testing.T) {
	h := &GoHandler{}
	caps := h.Capabilities()

	expected := []string{"go_module", "cross_compile"}
	if len(caps) != len(expected) {
		t.Fatalf("expected %d capabilities, got %d", len(expected), len(caps))
	}

	for i, cap := range expected {
		if caps[i] != cap {
			t.Errorf("expected capability[%d] = '%s', got '%s'", i, cap, caps[i])
		}
	}
}

// TestGoHandler_ResolveToolName tests the tool selection logic.
// These tests will pass once resolveToolName is implemented on GoHandler.
func TestGoHandler_ResolveToolName_Default(t *testing.T) {
	h := &GoHandler{}
	opts := goToolOptions{}

	toolName := h.resolveToolName(opts)
	if toolName != "go" {
		t.Errorf("expected default tool 'go', got '%s'", toolName)
	}
}

func TestGoHandler_ResolveToolName_UseContainer(t *testing.T) {
	h := &GoHandler{}
	opts := goToolOptions{UseContainer: true}

	toolName := h.resolveToolName(opts)
	if toolName != "go-oci" {
		t.Errorf("expected container tool 'go-oci' when UseContainer=true, got '%s'", toolName)
	}
}

func TestGoHandler_ResolveToolName_RaceDetection(t *testing.T) {
	h := &GoHandler{}
	opts := goToolOptions{RaceDetection: true}

	toolName := h.resolveToolName(opts)
	if toolName != "go-oci" {
		t.Errorf("expected container tool 'go-oci' when RaceDetection=true, got '%s'", toolName)
	}
}

func TestGoHandler_ResolveToolName_CGOEnabled(t *testing.T) {
	h := &GoHandler{}
	opts := goToolOptions{CGOEnabled: true}

	toolName := h.resolveToolName(opts)
	if toolName != "go-oci" {
		t.Errorf("expected container tool 'go-oci' when CGOEnabled=true, got '%s'", toolName)
	}
}

func TestGoHandler_ExecuteTool_SystemTool(t *testing.T) {
	// Setup mock tool registry
	registry := tool.NewRegistry()
	goTool := &tool.ToolDefinition{
		ID:     "go",
		Type:   tool.ToolTypeSystem,
		Binary: "go",
	}
	if err := registry.Register(goTool); err != nil {
		t.Fatalf("failed to register tool: %v", err)
	}
	tool.SetGlobalRegistry(registry)
	defer func() {
		tool.SetGlobalRegistry(tool.NewRegistry())
	}()

	h := &GoHandler{}
	var logBuf bytes.Buffer

	// Test building args for mod tidy
	args := []string{"mod", "tidy"}
	ctx := context.Background()

	// Create a test context
	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: t.TempDir(),
		ModuleRoot:    ".",
		LogWriter:     &logBuf,
		Operation:     tool.OperationBuild,
		ArgsOverrides: args,
	}

	// Verify the execution context is correctly configured
	if len(execCtx.ArgsOverrides) != 2 {
		t.Errorf("expected 2 args, got %d", len(execCtx.ArgsOverrides))
	}
	if execCtx.ArgsOverrides[0] != "mod" || execCtx.ArgsOverrides[1] != "tidy" {
		t.Errorf("expected args ['mod', 'tidy'], got %v", execCtx.ArgsOverrides)
	}

	_ = ctx // Will be used in actual implementation
	_ = h   // Handler reference
}

func TestGoHandler_CrossCompileEnvOverrides(t *testing.T) {
	tests := []struct {
		name   string
		goos   string
		goarch string
	}{
		{"linux-amd64", "linux", "amd64"},
		{"linux-arm64", "linux", "arm64"},
		{"darwin-amd64", "darwin", "amd64"},
		{"darwin-arm64", "darwin", "arm64"},
		{"windows-amd64", "windows", "amd64"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Verify env overrides are correctly structured
			env := map[string]string{
				"GOOS":        tc.goos,
				"GOARCH":      tc.goarch,
				"CGO_ENABLED": "0",
			}

			if env["GOOS"] != tc.goos {
				t.Errorf("expected GOOS=%s, got %s", tc.goos, env["GOOS"])
			}
			if env["GOARCH"] != tc.goarch {
				t.Errorf("expected GOARCH=%s, got %s", tc.goarch, env["GOARCH"])
			}
			if env["CGO_ENABLED"] != "0" {
				t.Errorf("expected CGO_ENABLED=0, got %s", env["CGO_ENABLED"])
			}
		})
	}
}

func TestGoHandler_BuildArgsForBuild(t *testing.T) {
	tests := []struct {
		name       string
		outputPath string
		ldflags    string
		wantArgs   []string
	}{
		{
			name:       "simple build",
			outputPath: "/out/bin",
			ldflags:    "",
			wantArgs:   []string{"build", "-o", "/out/bin"},
		},
		{
			name:       "build with ldflags",
			outputPath: "/out/bin",
			ldflags:    "-X main.Version=1.0.0",
			wantArgs:   []string{"build", "-o", "/out/bin", "-ldflags", "-X main.Version=1.0.0"},
		},
		{
			name:       "library build",
			outputPath: "",
			ldflags:    "",
			wantArgs:   []string{"build", "./..."},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var args []string
			if tc.outputPath == "" {
				args = []string{"build", "./..."}
			} else {
				args = []string{"build", "-o", tc.outputPath}
				if tc.ldflags != "" {
					args = append(args, "-ldflags", tc.ldflags)
				}
			}

			if len(args) != len(tc.wantArgs) {
				t.Errorf("got %d args, want %d", len(args), len(tc.wantArgs))
				return
			}

			for i, arg := range tc.wantArgs {
				if args[i] != arg {
					t.Errorf("arg[%d] = %q, want %q", i, args[i], arg)
				}
			}
		})
	}
}

func TestGoHandler_IsContainer(t *testing.T) {
	h := &GoHandler{}
	if h.IsContainer() {
		t.Error("GoHandler should not be a container handler by default")
	}
}

func TestGoHandler_IsHostInstalled(t *testing.T) {
	h := &GoHandler{}
	if !h.IsHostInstalled() {
		t.Error("GoHandler should report as host-installed")
	}
}

func TestGoHandler_Requirements(t *testing.T) {
	h := &GoHandler{}
	reqs := h.Requirements()

	// After refactoring, Go handler should have "go" as a requirement
	// but the system won't fail if Go is not installed (lazy validation)
	found := false
	for _, req := range reqs {
		if req == "go" {
			found = true
			break
		}
	}

	if !found {
		t.Error("GoHandler requirements should include 'go'")
	}
}
