package tool

import (
	"bytes"
	"context"
	"testing"
)

func TestDeployToolHandlerAdapter_DryRunPassesWhatIf(t *testing.T) {
	var capturedCtx *ExecutionContext
	capturingExecutor := &capturingMockExecutor{
		captureCtx: func(ctx *ExecutionContext) { capturedCtx = ctx },
	}

	td := &ToolDefinition{
		Type:    ToolTypeContainer,
		Binary:  "az",
		Command: []string{"sh", "/entrypoint.sh"},
	}
	td.ID = "az-bicep"

	adapter := NewDeployToolHandlerAdapter(td, capturingExecutor)
	module := createTestModule("infra")
	opts := DeployOptions{
		Environment: "development",
		DryRun:      true,
		Component:   "networking",
	}

	var buf bytes.Buffer
	exitCode := adapter.DryRun(context.Background(), module, "/workspace", "/out", &buf, opts)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if capturedCtx == nil {
		t.Fatal("executor was not called")
	}

	// Verify {dry-run-flag} placeholder is set
	if capturedCtx.Placeholders["{dry-run-flag}"] != "--what-if" {
		t.Errorf("expected {dry-run-flag} = '--what-if', got %q", capturedCtx.Placeholders["{dry-run-flag}"])
	}

	// Verify ArgsOverrides includes --what-if
	found := false
	for _, arg := range capturedCtx.ArgsOverrides {
		if arg == "--what-if" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ArgsOverrides to contain '--what-if', got %v", capturedCtx.ArgsOverrides)
	}
}

func TestDeployToolHandlerAdapter_DeployNoWhatIf(t *testing.T) {
	var capturedCtx *ExecutionContext
	capturingExecutor := &capturingMockExecutor{
		captureCtx: func(ctx *ExecutionContext) { capturedCtx = ctx },
	}

	td := &ToolDefinition{
		Type:    ToolTypeContainer,
		Binary:  "az",
		Command: []string{"sh", "/entrypoint.sh"},
	}
	td.ID = "az-bicep"

	adapter := NewDeployToolHandlerAdapter(td, capturingExecutor)
	module := createTestModule("infra")
	opts := DeployOptions{
		Environment: "development",
		DryRun:      false,
		Component:   "networking",
	}

	var buf bytes.Buffer
	exitCode := adapter.Deploy(context.Background(), module, "/workspace", "/out", &buf, opts)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	// Verify {dry-run-flag} placeholder is empty for real deploy
	if capturedCtx.Placeholders["{dry-run-flag}"] != "" {
		t.Errorf("expected {dry-run-flag} = '', got %q", capturedCtx.Placeholders["{dry-run-flag}"])
	}

	// Verify ArgsOverrides does NOT contain --what-if
	for _, arg := range capturedCtx.ArgsOverrides {
		if arg == "--what-if" {
			t.Error("ArgsOverrides should not contain '--what-if' for real deploy")
		}
	}
}

// capturingMockExecutor captures the ExecutionContext for assertions.
type capturingMockExecutor struct {
	captureCtx func(*ExecutionContext)
}

func (m *capturingMockExecutor) Execute(_ context.Context, _ *ToolDefinition, execCtx *ExecutionContext) (*ExecutionResult, error) {
	if m.captureCtx != nil {
		m.captureCtx(execCtx)
	}
	return &ExecutionResult{ExitCode: 0}, nil
}

func (m *capturingMockExecutor) Validate(_ *ToolDefinition) error { return nil }
func (m *capturingMockExecutor) Close() error                    { return nil }
