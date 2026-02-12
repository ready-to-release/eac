package tool

import (
	"context"
	"io"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// TestHandler is the interface for test handlers.
type TestHandler interface {
	Name() string
	Test(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts TestOptions) int
	Requirements() []string
	IsContainer() bool
	IsHostInstalled() bool
}

// TestOptions contains flags for controlling the test process.
type TestOptions struct {
	Verbose    bool   // Verbose output
	Filter     string // Test filter pattern
	CoverageOn bool   // Enable coverage
	Timeout    string // Test timeout
}

// TestHandlerAdapter wraps a ToolDefinition to implement TestHandler.
type TestHandlerAdapter struct {
	BaseHandlerAdapter
}

// NewTestHandlerAdapter creates a handler adapter for testing.
func NewTestHandlerAdapter(tool *ToolDefinition, executor Executor) *TestHandlerAdapter {
	return &TestHandlerAdapter{
		BaseHandlerAdapter: BaseHandlerAdapter{Tool: tool, Executor: executor},
	}
}

// Test executes the tool for a test operation.
func (a *TestHandlerAdapter) Test(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts TestOptions) int {
	execCtx := &ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    module.GetComponentRoot(""),
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     core.ActionTest,
		Placeholders: map[string]string{
			"{workspace}": workspaceRoot,
			"{module}":    filepath.Join(workspaceRoot, module.GetComponentRoot("")),
			"{output}":    outputDir,
		},
	}

	// Add test-specific args
	if opts.Verbose {
		execCtx.ArgsOverrides = append(execCtx.ArgsOverrides, "-v")
	}
	if opts.Filter != "" {
		execCtx.ArgsOverrides = append(execCtx.ArgsOverrides, "-run", opts.Filter)
	}

	ctx := context.Background()
	result, err := a.Executor.Execute(ctx, a.Tool, execCtx)
	if err != nil {
		if logWriter != nil {
			io.WriteString(logWriter, "Error executing tests: "+err.Error()+"\n")
		}
		return 1
	}

	return result.ExitCode
}
