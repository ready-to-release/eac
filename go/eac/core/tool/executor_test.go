package tool

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// mockConn implements net.Conn for testing.
type mockConn struct {
	net.Conn
}

func (m *mockConn) Close() error { return nil }
func (m *mockConn) Read(b []byte) (n int, err error) { return 0, io.EOF }
func (m *mockConn) Write(b []byte) (n int, err error) { return len(b), nil }

// mockDockerClient implements DockerClient for testing.
type mockDockerClient struct {
	images          []image.Summary
	pullErr         error
	createErr       error
	startErr        error
	waitResponse    container.WaitResponse
	waitErr         error
	attachResponse  *types.HijackedResponse
	attachErr       error
	containerOutput []byte
}

func (m *mockDockerClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	return m.images, nil
}

func (m *mockDockerClient) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	if m.pullErr != nil {
		return nil, m.pullErr
	}
	return io.NopCloser(bytes.NewReader(nil)), nil
}

func (m *mockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	if m.createErr != nil {
		return container.CreateResponse{}, m.createErr
	}
	return container.CreateResponse{ID: "test-container-id"}, nil
}

func (m *mockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return m.startErr
}

func (m *mockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return nil
}

func (m *mockDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	waitCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)

	if m.waitErr != nil {
		errCh <- m.waitErr
	} else {
		waitCh <- m.waitResponse
	}

	return waitCh, errCh
}

func (m *mockDockerClient) ContainerAttach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error) {
	if m.attachErr != nil {
		return types.HijackedResponse{}, m.attachErr
	}
	// Return a mock response with container output
	// Use a pipe to simulate the HijackedResponse Reader which expects a *bufio.Reader
	pr, pw := io.Pipe()
	go func() {
		pw.Write(m.containerOutput)
		pw.Close()
	}()
	return types.HijackedResponse{
		Conn:   &mockConn{},
		Reader: bufio.NewReader(pr),
	}, nil
}

func (m *mockDockerClient) Close() error {
	return nil
}

func TestNewExecutor(t *testing.T) {
	e := NewExecutor()
	if e == nil {
		t.Fatal("NewExecutor returned nil")
	}
}

func TestDefaultExecutor_ExecuteSystem(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows: echo behaves differently")
	}

	e := NewExecutor()
	defer e.Close()

	tool := &ToolDefinition{
		ID:     "echo-test",
		Type:   ToolTypeSystem,
		Binary: "echo",
		Args:   []string{"hello", "world"},
	}

	ctx := context.Background()
	execCtx := &ExecutionContext{
		WorkspaceRoot: ".",
	}

	result, err := e.Execute(ctx, tool, execCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	expected := "hello world"
	stdout := string(bytes.TrimSpace(result.Stdout))
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestDefaultExecutor_ExecuteSystem_WithEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows: sh not available")
	}

	e := NewExecutor()
	defer e.Close()

	// Use a command that outputs an environment variable
	// On Windows, we can use cmd /c echo %VAR%
	// On Unix, we use sh -c 'echo $VAR'
	tool := &ToolDefinition{
		ID:     "env-test",
		Type:   ToolTypeSystem,
		Binary: "sh",
		Args:   []string{"-c", "echo $TEST_VAR"},
		Env:    map[string]string{"TEST_VAR": "tool-env"},
	}

	ctx := context.Background()
	execCtx := &ExecutionContext{
		WorkspaceRoot: ".",
		EnvOverrides:  map[string]string{"TEST_VAR": "context-override"},
	}

	result, err := e.Execute(ctx, tool, execCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Context override should take precedence
	stdout := string(bytes.TrimSpace(result.Stdout))
	if stdout != "context-override" {
		t.Errorf("stdout = %q, want %q", stdout, "context-override")
	}
}

func TestDefaultExecutor_ExecuteSystem_WithArgsOverrides(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows: echo behaves differently")
	}

	e := NewExecutor()
	defer e.Close()

	tool := &ToolDefinition{
		ID:     "args-test",
		Type:   ToolTypeSystem,
		Binary: "echo",
		Args:   []string{"base"},
	}

	ctx := context.Background()
	execCtx := &ExecutionContext{
		WorkspaceRoot: ".",
		ArgsOverrides: []string{"override"},
	}

	result, err := e.Execute(ctx, tool, execCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have both base args and overrides
	stdout := string(bytes.TrimSpace(result.Stdout))
	if stdout != "base override" {
		t.Errorf("stdout = %q, want %q", stdout, "base override")
	}
}

func TestDefaultExecutor_ExecuteSystem_Placeholders(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows: echo behaves differently")
	}

	e := NewExecutor()
	defer e.Close()

	tool := &ToolDefinition{
		ID:     "placeholder-test",
		Type:   ToolTypeSystem,
		Binary: "echo",
		Args:   []string{"{workspace}", "{module}"},
	}

	ctx := context.Background()
	execCtx := &ExecutionContext{
		WorkspaceRoot: ".",
		Placeholders: map[string]string{
			"{workspace}": "/home/user/project",
			"{module}":    "mymodule",
		},
	}

	result, err := e.Execute(ctx, tool, execCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stdout := string(bytes.TrimSpace(result.Stdout))
	expected := "/home/user/project mymodule"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
}

func TestDefaultExecutor_ExecuteSystem_NonZeroExit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows: sh not available")
	}

	e := NewExecutor()
	defer e.Close()

	tool := &ToolDefinition{
		ID:     "exit-test",
		Type:   ToolTypeSystem,
		Binary: "sh",
		Args:   []string{"-c", "exit 42"},
	}

	ctx := context.Background()
	execCtx := &ExecutionContext{
		WorkspaceRoot: ".",
	}

	result, err := e.Execute(ctx, tool, execCtx)
	if err != nil {
		t.Fatalf("unexpected error for non-zero exit: %v", err)
	}

	if result.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", result.ExitCode)
	}
}

func TestDefaultExecutor_ExecuteSystem_BinaryNotFound(t *testing.T) {
	e := NewExecutor()
	defer e.Close()

	tool := &ToolDefinition{
		ID:     "notfound-test",
		Type:   ToolTypeSystem,
		Binary: "nonexistent-binary-12345",
	}

	ctx := context.Background()
	execCtx := &ExecutionContext{
		WorkspaceRoot: ".",
	}

	_, err := e.Execute(ctx, tool, execCtx)
	if err == nil {
		t.Error("expected error for nonexistent binary")
	}
}

func TestDefaultExecutor_Validate_SystemTool(t *testing.T) {
	e := NewExecutor()
	defer e.Close()

	t.Run("existing binary", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows: echo is not a standalone binary")
		}
		tool := &ToolDefinition{
			ID:     "echo",
			Type:   ToolTypeSystem,
			Binary: "echo", // Should exist on most systems
		}
		err := e.Validate(tool)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nonexistent binary", func(t *testing.T) {
		tool := &ToolDefinition{
			ID:     "nonexistent",
			Type:   ToolTypeSystem,
			Binary: "nonexistent-binary-xyz",
		}
		err := e.Validate(tool)
		if err == nil {
			t.Error("expected error for nonexistent binary")
		}
	})

	t.Run("nil tool", func(t *testing.T) {
		err := e.Validate(nil)
		if err == nil {
			t.Error("expected error for nil tool")
		}
	})
}

func TestDefaultExecutor_Execute_NilInputs(t *testing.T) {
	e := NewExecutor()
	defer e.Close()

	ctx := context.Background()

	t.Run("nil tool", func(t *testing.T) {
		_, err := e.Execute(ctx, nil, &ExecutionContext{})
		if err == nil {
			t.Error("expected error for nil tool")
		}
	})

	t.Run("nil context", func(t *testing.T) {
		tool := &ToolDefinition{
			ID:     "test",
			Type:   ToolTypeSystem,
			Binary: "echo",
		}
		_, err := e.Execute(ctx, tool, nil)
		if err == nil {
			t.Error("expected error for nil context")
		}
	})
}

func TestDefaultExecutor_Execute_UnknownType(t *testing.T) {
	e := NewExecutor()
	defer e.Close()

	tool := &ToolDefinition{
		ID:   "unknown",
		Type: "unknown",
	}

	ctx := context.Background()
	execCtx := &ExecutionContext{
		WorkspaceRoot: ".",
	}

	_, err := e.Execute(ctx, tool, execCtx)
	if err == nil {
		t.Error("expected error for unknown tool type")
	}
}

func TestDefaultExecutor_LogWriter(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows: echo behaves differently")
	}

	e := NewExecutor()
	defer e.Close()

	tool := &ToolDefinition{
		ID:     "log-test",
		Type:   ToolTypeSystem,
		Binary: "echo",
		Args:   []string{"logged output"},
	}

	var logBuf bytes.Buffer
	ctx := context.Background()
	execCtx := &ExecutionContext{
		WorkspaceRoot: ".",
		LogWriter:     &logBuf,
	}

	_, err := e.Execute(ctx, tool, execCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logOutput := logBuf.String()
	if !bytes.Contains([]byte(logOutput), []byte("logged output")) {
		t.Errorf("log writer should receive output, got: %q", logOutput)
	}
}

func TestDefaultExecutor_ExecuteContainer_WithMock(t *testing.T) {
	mock := &mockDockerClient{
		images: []image.Summary{
			{RepoTags: []string{"test-image:latest"}},
		},
		waitResponse:    container.WaitResponse{StatusCode: 0},
		containerOutput: []byte("container output"),
	}

	e := NewExecutorWithDocker(mock)
	defer e.Close()

	tool := &ToolDefinition{
		ID:      "container-test",
		Type:    ToolTypeContainer,
		Image:   "test-image",
		Command: []string{"echo", "hello"},
	}

	ctx := context.Background()
	execCtx := &ExecutionContext{
		WorkspaceRoot: "/workspace",
	}

	result, err := e.Execute(ctx, tool, execCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestParseMemoryLimit(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"", 0},
		{"8g", 8 * 1024 * 1024 * 1024},
		{"8G", 8 * 1024 * 1024 * 1024},
		{"8gb", 8 * 1024 * 1024 * 1024},
		{"512m", 512 * 1024 * 1024},
		{"512M", 512 * 1024 * 1024},
		{"512mb", 512 * 1024 * 1024},
		{"1024k", 1024 * 1024},
		{"1024kb", 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseMemoryLimit(tt.input)
			if got != tt.want {
				t.Errorf("parseMemoryLimit(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePlaceholders(t *testing.T) {
	placeholders := map[string]string{
		"{workspace}": "/home/user/project",
		"{module}":    "mymodule",
		"{output}":    "/out",
	}

	tests := []struct {
		input string
		want  string
	}{
		{"{workspace}/src", "/home/user/project/src"},
		{"{workspace}/{module}", "/home/user/project/mymodule"},
		{"prefix-{module}-suffix", "prefix-mymodule-suffix"},
		{"no-placeholders", "no-placeholders"},
		{"{unknown}", "{unknown}"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := resolvePlaceholders(tt.input, placeholders)
			if got != tt.want {
				t.Errorf("resolvePlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePlaceholders_NilMap(t *testing.T) {
	result := resolvePlaceholders("{workspace}", nil)
	if result != "{workspace}" {
		t.Errorf("expected unchanged string with nil placeholders, got %q", result)
	}
}

func TestExecutionResult_Duration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows: sh not available")
	}

	e := NewExecutor()
	defer e.Close()

	tool := &ToolDefinition{
		ID:     "duration-test",
		Type:   ToolTypeSystem,
		Binary: "sh",
		Args:   []string{"-c", "sleep 0.1"},
	}

	ctx := context.Background()
	execCtx := &ExecutionContext{
		WorkspaceRoot: ".",
	}

	result, err := e.Execute(ctx, tool, execCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Duration should be at least 100ms
	if result.Duration < 100*time.Millisecond {
		t.Errorf("duration too short: %v", result.Duration)
	}
}
