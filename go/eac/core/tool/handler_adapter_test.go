package tool

import (
	"bytes"
	"context"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/adapters"
	"github.com/ready-to-release/eac/go/eac/core/domain"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"github.com/ready-to-release/eac/contracts/eac-core-interfaces"
)

// createTestModule creates a ModuleContractPort for testing.
func createTestModule(moniker string) interfaces.ModuleContractPort {
	m := modules.NewModuleContract(domain.BaseContract{
		Moniker: moniker,
	}, "/workspace")
	return adapters.AdaptModule(m)
}

// mockExecutor implements Executor for testing.
type mockExecutor struct {
	result *ExecutionResult
	err    error
}

func (m *mockExecutor) Execute(ctx context.Context, tool *ToolDefinition, execCtx *ExecutionContext) (*ExecutionResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &ExecutionResult{ExitCode: 0}, nil
}

func (m *mockExecutor) Validate(tool *ToolDefinition) error {
	return nil
}

func (m *mockExecutor) Close() error {
	return nil
}

func TestToolHandlerAdapter_Name(t *testing.T) {
	tool := &ToolDefinition{
		ID:     "test-tool",
		Type:   ToolTypeSystem,
		Binary: "test",
	}
	adapter := NewToolHandlerAdapter(tool, &mockExecutor{})

	if adapter.Name() != "test-tool" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "test-tool")
	}
}

func TestToolHandlerAdapter_Build(t *testing.T) {
	tool := &ToolDefinition{
		ID:     "go-build",
		Type:   ToolTypeSystem,
		Binary: "go",
		Args:   []string{"build"},
	}

	t.Run("successful build", func(t *testing.T) {
		executor := &mockExecutor{
			result: &ExecutionResult{ExitCode: 0},
		}
		adapter := NewToolHandlerAdapter(tool, executor)

		module := createTestModule("test-module")

		var logBuf bytes.Buffer
		opts := BuildOptions{
			Component: "go",
		}

		exitCode := adapter.Build(module, "/workspace", "/output", &logBuf, opts)
		if exitCode != 0 {
			t.Errorf("Build() = %d, want 0", exitCode)
		}
	})

	t.Run("failed build", func(t *testing.T) {
		executor := &mockExecutor{
			result: &ExecutionResult{ExitCode: 1},
		}
		adapter := NewToolHandlerAdapter(tool, executor)

		module := createTestModule("test-module")

		var logBuf bytes.Buffer
		opts := BuildOptions{}

		exitCode := adapter.Build(module, "/workspace", "/output", &logBuf, opts)
		if exitCode != 1 {
			t.Errorf("Build() = %d, want 1", exitCode)
		}
	})

	t.Run("executor error", func(t *testing.T) {
		executor := &mockExecutor{
			err: context.DeadlineExceeded,
		}
		adapter := NewToolHandlerAdapter(tool, executor)

		module := createTestModule("test-module")

		var logBuf bytes.Buffer
		opts := BuildOptions{}

		exitCode := adapter.Build(module, "/workspace", "/output", &logBuf, opts)
		if exitCode != 1 {
			t.Errorf("Build() = %d, want 1 (error case)", exitCode)
		}
	})
}

func TestToolHandlerAdapter_Requirements(t *testing.T) {
	tool := &ToolDefinition{
		ID:           "test-tool",
		Type:         ToolTypeSystem,
		Binary:       "test",
		Requirements: []string{"test", "docker"},
	}
	adapter := NewToolHandlerAdapter(tool, &mockExecutor{})

	reqs := adapter.Requirements()
	if len(reqs) != 2 {
		t.Errorf("Requirements() returned %d items, want 2", len(reqs))
	}
	if reqs[0] != "test" || reqs[1] != "docker" {
		t.Errorf("Requirements() = %v, want [test docker]", reqs)
	}
}

func TestToolHandlerAdapter_ValidateModule(t *testing.T) {
	tool := &ToolDefinition{
		ID:     "test-tool",
		Type:   ToolTypeSystem,
		Binary: "test",
	}
	adapter := NewToolHandlerAdapter(tool, &mockExecutor{})

	module := createTestModule("test-module")

	// Should always return nil for YAML tools
	err := adapter.ValidateModule(module, "/workspace", "go")
	if err != nil {
		t.Errorf("ValidateModule() returned error: %v", err)
	}
}

func TestToolHandlerAdapter_ListArtifacts(t *testing.T) {
	tool := &ToolDefinition{
		ID:     "test-tool",
		Type:   ToolTypeSystem,
		Binary: "test",
	}
	adapter := NewToolHandlerAdapter(tool, &mockExecutor{})

	module := createTestModule("test-module")

	// Should return nil for YAML tools
	artifacts := adapter.ListArtifacts(module, "/workspace")
	if artifacts != nil {
		t.Errorf("ListArtifacts() = %v, want nil", artifacts)
	}
}

func TestLintHandlerAdapter_Lint(t *testing.T) {
	tool := &ToolDefinition{
		ID:     "go-lint",
		Type:   ToolTypeSystem,
		Binary: "golangci-lint",
		Args:   []string{"run"},
	}

	t.Run("successful lint", func(t *testing.T) {
		executor := &mockExecutor{
			result: &ExecutionResult{ExitCode: 0},
		}
		adapter := NewLintHandlerAdapter(tool, executor)

		var logBuf bytes.Buffer
		opts := LintOptions{}

		exitCode := adapter.Lint("go/module", "/workspace", "/output", &logBuf, opts)
		if exitCode != 0 {
			t.Errorf("Lint() = %d, want 0", exitCode)
		}
	})

	t.Run("lint with fix", func(t *testing.T) {
		executor := &mockExecutor{
			result: &ExecutionResult{ExitCode: 0},
		}
		adapter := NewLintHandlerAdapter(tool, executor)

		var logBuf bytes.Buffer
		opts := LintOptions{Fix: true}

		exitCode := adapter.Lint("go/module", "/workspace", "/output", &logBuf, opts)
		if exitCode != 0 {
			t.Errorf("Lint() = %d, want 0", exitCode)
		}
	})

	t.Run("lint failures", func(t *testing.T) {
		executor := &mockExecutor{
			result: &ExecutionResult{ExitCode: 1},
		}
		adapter := NewLintHandlerAdapter(tool, executor)

		var logBuf bytes.Buffer
		opts := LintOptions{}

		exitCode := adapter.Lint("go/module", "/workspace", "/output", &logBuf, opts)
		if exitCode != 1 {
			t.Errorf("Lint() = %d, want 1", exitCode)
		}
	})
}

func TestLintHandlerAdapter_Name(t *testing.T) {
	tool := &ToolDefinition{
		ID:     "go-lint",
		Type:   ToolTypeSystem,
		Binary: "golangci-lint",
	}
	adapter := NewLintHandlerAdapter(tool, &mockExecutor{})

	if adapter.Name() != "go-lint" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "go-lint")
	}
}

func TestTestHandlerAdapter_Test(t *testing.T) {
	tool := &ToolDefinition{
		ID:     "go-test",
		Type:   ToolTypeSystem,
		Binary: "go",
		Args:   []string{"test", "./..."},
	}

	t.Run("successful test", func(t *testing.T) {
		executor := &mockExecutor{
			result: &ExecutionResult{ExitCode: 0},
		}
		adapter := NewTestHandlerAdapter(tool, executor)

		module := createTestModule("test-module")

		var logBuf bytes.Buffer
		opts := TestOptions{}

		exitCode := adapter.Test(module, "/workspace", "/output", &logBuf, opts)
		if exitCode != 0 {
			t.Errorf("Test() = %d, want 0", exitCode)
		}
	})

	t.Run("test with verbose and filter", func(t *testing.T) {
		executor := &mockExecutor{
			result: &ExecutionResult{ExitCode: 0},
		}
		adapter := NewTestHandlerAdapter(tool, executor)

		module := createTestModule("test-module")

		var logBuf bytes.Buffer
		opts := TestOptions{
			Verbose: true,
			Filter:  "TestFoo",
		}

		exitCode := adapter.Test(module, "/workspace", "/output", &logBuf, opts)
		if exitCode != 0 {
			t.Errorf("Test() = %d, want 0", exitCode)
		}
	})

	t.Run("test failures", func(t *testing.T) {
		executor := &mockExecutor{
			result: &ExecutionResult{ExitCode: 1},
		}
		adapter := NewTestHandlerAdapter(tool, executor)

		module := createTestModule("test-module")

		var logBuf bytes.Buffer
		opts := TestOptions{}

		exitCode := adapter.Test(module, "/workspace", "/output", &logBuf, opts)
		if exitCode != 1 {
			t.Errorf("Test() = %d, want 1", exitCode)
		}
	})
}

func TestScanHandlerAdapter_Scan(t *testing.T) {
	tool := &ToolDefinition{
		ID:    "trivy",
		Type:  ToolTypeContainer,
		Image: "ghcr.io/aquasecurity/trivy",
		Tag:   "0.68.2",
	}

	t.Run("successful scan", func(t *testing.T) {
		executor := &mockExecutor{
			result: &ExecutionResult{
				ExitCode: 0,
				Stdout:   []byte(`{"findings": []}`),
			},
		}
		adapter := NewScanHandlerAdapter(tool, executor)

		var logBuf bytes.Buffer
		opts := ScanOptions{ScanType: "vuln"}

		exitCode, output := adapter.Scan("go/module", "/workspace", "/output", &logBuf, opts)
		if exitCode != 0 {
			t.Errorf("Scan() exitCode = %d, want 0", exitCode)
		}
		if string(output) != `{"findings": []}` {
			t.Errorf("Scan() output = %q, want findings JSON", string(output))
		}
	})

	t.Run("scan with findings", func(t *testing.T) {
		executor := &mockExecutor{
			result: &ExecutionResult{
				ExitCode: 1, // Non-zero indicates findings
				Stdout:   []byte(`{"findings": [{"severity": "HIGH"}]}`),
			},
		}
		adapter := NewScanHandlerAdapter(tool, executor)

		var logBuf bytes.Buffer
		opts := ScanOptions{ScanType: "vuln"}

		exitCode, output := adapter.Scan("go/module", "/workspace", "/output", &logBuf, opts)
		if exitCode != 1 {
			t.Errorf("Scan() exitCode = %d, want 1", exitCode)
		}
		if len(output) == 0 {
			t.Error("Scan() should return findings output")
		}
	})

	t.Run("executor error", func(t *testing.T) {
		executor := &mockExecutor{
			err: context.DeadlineExceeded,
		}
		adapter := NewScanHandlerAdapter(tool, executor)

		var logBuf bytes.Buffer
		opts := ScanOptions{ScanType: "vuln"}

		exitCode, output := adapter.Scan("go/module", "/workspace", "/output", &logBuf, opts)
		if exitCode != 1 {
			t.Errorf("Scan() exitCode = %d, want 1 (error case)", exitCode)
		}
		if output != nil {
			t.Error("Scan() should return nil output on error")
		}
	})
}

func TestScanHandlerAdapter_Name(t *testing.T) {
	tool := &ToolDefinition{
		ID:    "trivy",
		Type:  ToolTypeContainer,
		Image: "ghcr.io/aquasecurity/trivy",
	}
	adapter := NewScanHandlerAdapter(tool, &mockExecutor{})

	if adapter.Name() != "trivy" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "trivy")
	}
}

func TestScanHandlerAdapter_Requirements(t *testing.T) {
	tool := &ToolDefinition{
		ID:           "trivy",
		Type:         ToolTypeContainer,
		Image:        "ghcr.io/aquasecurity/trivy",
		Requirements: []string{"docker"},
	}
	adapter := NewScanHandlerAdapter(tool, &mockExecutor{})

	reqs := adapter.Requirements()
	if len(reqs) != 1 || reqs[0] != "docker" {
		t.Errorf("Requirements() = %v, want [docker]", reqs)
	}
}
