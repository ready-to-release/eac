package tool

import (
	"context"
	"io"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// ScanHandler is the interface for security scan handlers.
type ScanHandler interface {
	Name() string
	Scan(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts ScanOptions) (int, []byte)
	Requirements() []string
	IsContainer() bool
	IsHostInstalled() bool
}

// ScanOptions contains flags for controlling the scan process.
type ScanOptions struct {
	ScanType   string // Type of scan (sbom, vuln, secrets, sast)
	OutputPath string // Path for scan results
	Severity   string // Minimum severity level
}

// ScanHandlerAdapter wraps a ToolDefinition to implement ScanHandler.
type ScanHandlerAdapter struct {
	BaseHandlerAdapter
}

// NewScanHandlerAdapter creates a handler adapter for scanning.
func NewScanHandlerAdapter(tool *ToolDefinition, executor Executor) *ScanHandlerAdapter {
	return &ScanHandlerAdapter{
		BaseHandlerAdapter: BaseHandlerAdapter{Tool: tool, Executor: executor},
	}
}

// Scan executes the tool for a scan operation.
func (a *ScanHandlerAdapter) Scan(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts ScanOptions) (int, []byte) {
	execCtx := &ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     core.ActionScan,
		Placeholders: map[string]string{
			"{workspace}": workspaceRoot,
			"{module}":    moduleRoot,
			"{output}":    outputDir,
		},
	}

	ctx := context.Background()
	result, err := a.Executor.Execute(ctx, a.Tool, execCtx)
	if err != nil {
		if logWriter != nil {
			io.WriteString(logWriter, "Error executing scanner: "+err.Error()+"\n")
		}
		return 1, nil
	}

	return result.ExitCode, result.Stdout
}
