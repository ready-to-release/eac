package tool

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	container "github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces"
)

func TestCredentials_GlobalHostEnvForwarding(t *testing.T) {
	// Set host env vars
	t.Setenv("TEST_GITHUB_TOKEN", "secret-token-123")
	t.Setenv("TEST_NPM_TOKEN", "npm-secret-456")

	var capturedEnv map[string]string
	mock := newMockContainerPort()
	mock.executeFunc = func(ctx context.Context, config *container.ContainerConfig) (*container.ContainerResult, error) {
		capturedEnv = config.Env
		return &container.ContainerResult{ExitCode: 0, Duration: time.Millisecond}, nil
	}
	mock.images["test-image:v1"] = true

	e := NewExecutorWithContainer(mock)
	e.SetCredentials(&CredentialsConfig{
		HostEnv: []string{"TEST_GITHUB_TOKEN", "TEST_NPM_TOKEN"},
	})

	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "test-image",
		Tag:   "v1",
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: t.TempDir(),
		ModuleRoot:    t.TempDir(),
	}

	_, err := e.Execute(context.Background(), td, execCtx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if capturedEnv["TEST_GITHUB_TOKEN"] != "secret-token-123" {
		t.Errorf("TEST_GITHUB_TOKEN = %q, want %q", capturedEnv["TEST_GITHUB_TOKEN"], "secret-token-123")
	}
	if capturedEnv["TEST_NPM_TOKEN"] != "npm-secret-456" {
		t.Errorf("TEST_NPM_TOKEN = %q, want %q", capturedEnv["TEST_NPM_TOKEN"], "npm-secret-456")
	}
}

func TestCredentials_CIEnvForwarding(t *testing.T) {
	t.Setenv("CI", "true")
	t.Setenv("GITHUB_ACTIONS", "true")

	var capturedEnv map[string]string
	mock := newMockContainerPort()
	mock.executeFunc = func(ctx context.Context, config *container.ContainerConfig) (*container.ContainerResult, error) {
		capturedEnv = config.Env
		return &container.ContainerResult{ExitCode: 0, Duration: time.Millisecond}, nil
	}
	mock.images["test-image:v1"] = true

	e := NewExecutorWithContainer(mock)
	e.SetCredentials(&CredentialsConfig{
		CIEnv: []string{"CI", "GITHUB_ACTIONS"},
	})

	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "test-image",
		Tag:   "v1",
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: t.TempDir(),
		ModuleRoot:    t.TempDir(),
	}

	_, err := e.Execute(context.Background(), td, execCtx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if capturedEnv["CI"] != "true" {
		t.Errorf("CI = %q, want %q", capturedEnv["CI"], "true")
	}
	if capturedEnv["GITHUB_ACTIONS"] != "true" {
		t.Errorf("GITHUB_ACTIONS = %q, want %q", capturedEnv["GITHUB_ACTIONS"], "true")
	}
}

func TestCredentials_PerToolHostEnv(t *testing.T) {
	t.Setenv("TEST_TOOL_SPECIFIC", "tool-value")

	var capturedEnv map[string]string
	mock := newMockContainerPort()
	mock.executeFunc = func(ctx context.Context, config *container.ContainerConfig) (*container.ContainerResult, error) {
		capturedEnv = config.Env
		return &container.ContainerResult{ExitCode: 0, Duration: time.Millisecond}, nil
	}
	mock.images["test-image:v1"] = true

	e := NewExecutorWithContainer(mock)
	// No global credentials — only per-tool

	td := &ToolDefinition{
		ID:      "test-tool",
		Type:    ToolTypeContainer,
		Image:   "test-image",
		Tag:     "v1",
		HostEnv: []string{"TEST_TOOL_SPECIFIC"},
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: t.TempDir(),
		ModuleRoot:    t.TempDir(),
	}

	_, err := e.Execute(context.Background(), td, execCtx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if capturedEnv["TEST_TOOL_SPECIFIC"] != "tool-value" {
		t.Errorf("TEST_TOOL_SPECIFIC = %q, want %q", capturedEnv["TEST_TOOL_SPECIFIC"], "tool-value")
	}
}

func TestCredentials_Precedence(t *testing.T) {
	// Global credential sets FOO
	t.Setenv("FOO", "from-host")

	var capturedEnv map[string]string
	mock := newMockContainerPort()
	mock.executeFunc = func(ctx context.Context, config *container.ContainerConfig) (*container.ContainerResult, error) {
		capturedEnv = config.Env
		return &container.ContainerResult{ExitCode: 0, Duration: time.Millisecond}, nil
	}
	mock.images["test-image:v1"] = true

	e := NewExecutorWithContainer(mock)
	e.SetCredentials(&CredentialsConfig{
		HostEnv: []string{"FOO"},
	})

	// Tool.Env overrides the forwarded value
	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "test-image",
		Tag:   "v1",
		Env:   map[string]string{"FOO": "from-yaml"},
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: t.TempDir(),
		ModuleRoot:    t.TempDir(),
	}

	_, err := e.Execute(context.Background(), td, execCtx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Static YAML env should override forwarded host value
	if capturedEnv["FOO"] != "from-yaml" {
		t.Errorf("FOO = %q, want %q (static YAML should override forwarded)", capturedEnv["FOO"], "from-yaml")
	}
}

func TestCredentials_EnvOverridesPrecedence(t *testing.T) {
	t.Setenv("BAR", "from-host")

	var capturedEnv map[string]string
	mock := newMockContainerPort()
	mock.executeFunc = func(ctx context.Context, config *container.ContainerConfig) (*container.ContainerResult, error) {
		capturedEnv = config.Env
		return &container.ContainerResult{ExitCode: 0, Duration: time.Millisecond}, nil
	}
	mock.images["test-image:v1"] = true

	e := NewExecutorWithContainer(mock)
	e.SetCredentials(&CredentialsConfig{
		HostEnv: []string{"BAR"},
	})

	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "test-image",
		Tag:   "v1",
		Env:   map[string]string{"BAR": "from-yaml"},
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: t.TempDir(),
		ModuleRoot:    t.TempDir(),
		EnvOverrides:  map[string]string{"BAR": "from-override"},
	}

	_, err := e.Execute(context.Background(), td, execCtx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// EnvOverrides should override both forwarded and static YAML
	if capturedEnv["BAR"] != "from-override" {
		t.Errorf("BAR = %q, want %q (EnvOverrides should win)", capturedEnv["BAR"], "from-override")
	}
}

func TestCredentials_UnsetEnvNotForwarded(t *testing.T) {
	// Ensure these are NOT set
	os.Unsetenv("DEFINITELY_NOT_SET_VAR")

	var capturedEnv map[string]string
	mock := newMockContainerPort()
	mock.executeFunc = func(ctx context.Context, config *container.ContainerConfig) (*container.ContainerResult, error) {
		capturedEnv = config.Env
		return &container.ContainerResult{ExitCode: 0, Duration: time.Millisecond}, nil
	}
	mock.images["test-image:v1"] = true

	e := NewExecutorWithContainer(mock)
	e.SetCredentials(&CredentialsConfig{
		HostEnv: []string{"DEFINITELY_NOT_SET_VAR"},
	})

	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "test-image",
		Tag:   "v1",
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: t.TempDir(),
		ModuleRoot:    t.TempDir(),
	}

	_, err := e.Execute(context.Background(), td, execCtx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if _, exists := capturedEnv["DEFINITELY_NOT_SET_VAR"]; exists {
		t.Error("Unset env var should not appear in container env")
	}
}

func TestCredentials_NilCredentialsIsHarmless(t *testing.T) {
	var capturedEnv map[string]string
	mock := newMockContainerPort()
	mock.executeFunc = func(ctx context.Context, config *container.ContainerConfig) (*container.ContainerResult, error) {
		capturedEnv = config.Env
		return &container.ContainerResult{ExitCode: 0, Duration: time.Millisecond}, nil
	}
	mock.images["test-image:v1"] = true

	e := NewExecutorWithContainer(mock)
	// No credentials set (nil)

	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "test-image",
		Tag:   "v1",
		Env:   map[string]string{"STATIC": "value"},
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: t.TempDir(),
		ModuleRoot:    t.TempDir(),
	}

	_, err := e.Execute(context.Background(), td, execCtx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Static env should still work without credentials
	if capturedEnv["STATIC"] != "value" {
		t.Errorf("STATIC = %q, want %q", capturedEnv["STATIC"], "value")
	}
}

func TestCredentials_DebugLogging(t *testing.T) {
	t.Setenv("TEST_LOG_VAR", "value")

	var logBuf bytes.Buffer
	mock := newMockContainerPort()
	mock.executeFunc = func(ctx context.Context, config *container.ContainerConfig) (*container.ContainerResult, error) {
		return &container.ContainerResult{ExitCode: 0, Duration: time.Millisecond}, nil
	}
	mock.images["test-image:v1"] = true

	e := NewExecutorWithContainer(mock)
	e.SetCredentials(&CredentialsConfig{
		HostEnv: []string{"TEST_LOG_VAR"},
	})

	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "test-image",
		Tag:   "v1",
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: t.TempDir(),
		ModuleRoot:    t.TempDir(),
		LogWriter:     &logBuf,
	}

	_, err := e.Execute(context.Background(), td, execCtx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	logOutput := logBuf.String()
	// Should log the variable name
	if !bytes.Contains([]byte(logOutput), []byte("TEST_LOG_VAR")) {
		t.Errorf("debug log should contain forwarded var name, got: %s", logOutput)
	}
	// Should NOT log the variable value
	if bytes.Contains([]byte(logOutput), []byte("value")) {
		t.Errorf("debug log should NOT contain var value, got: %s", logOutput)
	}
}
