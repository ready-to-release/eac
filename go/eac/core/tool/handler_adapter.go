// Package tool provides handler adapters that bridge the new tool system
// with existing handler interfaces.
package tool

import (
	"context"
	"io"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// BuildHandler is the interface for build handlers.
// This matches the existing builders.Handler interface to enable seamless integration.
type BuildHandler interface {
	// Name returns the handler identifier (e.g., "go", "mkdocs", "docker")
	Name() string

	// Build executes the build for a module.
	// Returns exit code (0 = success, non-zero = failure).
	Build(module *modules.ModuleContract, workspaceRoot, outputDir string,
		logWriter io.Writer, opts BuildOptions) int

	// ListArtifacts returns artifact paths that would be produced.
	// Paths are relative to the module's output directory.
	ListArtifacts(module *modules.ModuleContract, workspaceRoot string) []string

	// Requirements returns system dependencies required by this handler.
	// Used for early validation (e.g., ["go", "docker"]).
	Requirements() []string

	// ValidateModule checks if a module's configuration is valid for a specific component.
	// Returns nil if valid, or an error describing the problem.
	ValidateModule(module *modules.ModuleContract, workspaceRoot, component string) error
}

// BuildOptions contains flags for controlling the build process.
// This matches the existing builders.BuildOptions structure.
type BuildOptions struct {
	TidyFirst          bool     // Run go mod tidy before building
	Version            string   // Version to inject via ldflags
	DryRun             bool     // Simulate build without actually running it
	RequestedArtifacts []string // Specific artifact IDs to build (empty = default artifacts, "*" = all)
	Component          string   // Specific component to build (empty = all components)
}

// ToolHandlerAdapter wraps a ToolDefinition to implement BuildHandler.
// This allows YAML-defined tools to be used with the existing handler system.
type ToolHandlerAdapter struct {
	tool     *ToolDefinition
	executor Executor
}

// NewToolHandlerAdapter creates a handler adapter for a tool definition.
func NewToolHandlerAdapter(tool *ToolDefinition, executor Executor) *ToolHandlerAdapter {
	return &ToolHandlerAdapter{
		tool:     tool,
		executor: executor,
	}
}

// Name returns the tool ID.
func (a *ToolHandlerAdapter) Name() string {
	return a.tool.ID
}

// Build executes the tool for a build operation.
func (a *ToolHandlerAdapter) Build(module *modules.ModuleContract, workspaceRoot, outputDir string,
	logWriter io.Writer, opts BuildOptions) int {

	// Build execution context
	execCtx := &ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    module.GetComponentRoot(opts.Component),
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     OperationBuild,
		Placeholders: map[string]string{
			"{workspace}": workspaceRoot,
			"{module}":    filepath.Join(workspaceRoot, module.GetComponentRoot(opts.Component)),
			"{output}":    outputDir,
		},
	}

	// Add version to args if specified
	if opts.Version != "" && a.tool.Type == ToolTypeSystem {
		execCtx.ArgsOverrides = append(execCtx.ArgsOverrides, "-ldflags", "-X main.Version="+opts.Version)
	}

	// Execute the tool
	ctx := context.Background()
	result, err := a.executor.Execute(ctx, a.tool, execCtx)
	if err != nil {
		if logWriter != nil {
			io.WriteString(logWriter, "Error executing tool: "+err.Error()+"\n")
		}
		return 1
	}

	return result.ExitCode
}

// ListArtifacts returns an empty list - YAML-defined tools don't specify artifacts.
// Complex artifact handling should use native handlers.
func (a *ToolHandlerAdapter) ListArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	// YAML-defined tools don't specify artifacts
	// This is intentionally minimal - complex builds should use native handlers
	return nil
}

// Requirements returns the tool's requirements.
func (a *ToolHandlerAdapter) Requirements() []string {
	return a.tool.Requirements
}

// ValidateModule validates the tool can run - always returns nil for YAML tools.
// Complex validation should use native handlers.
func (a *ToolHandlerAdapter) ValidateModule(module *modules.ModuleContract, workspaceRoot, component string) error {
	// Basic validation - just check if this is a container tool that Docker is required
	if a.tool.Type == ToolTypeContainer {
		// Could check Docker availability here
		return nil
	}
	return nil
}

// LintHandler is the interface for lint handlers.
type LintHandler interface {
	Name() string
	Lint(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts LintOptions) int
	Requirements() []string
	ValidateModule(moduleRoot, workspaceRoot string) error
}

// LintOptions contains flags for controlling the lint process.
type LintOptions struct {
	Fix       bool     // Auto-fix issues where possible
	Config    string   // Override config file path
	Files     []string // Specific files to lint (for file-based linting)
	InputMode string   // How files are passed: "packages", "files", or "directory"
}

// LintHandlerAdapter wraps a ToolDefinition to implement LintHandler.
type LintHandlerAdapter struct {
	tool     *ToolDefinition
	executor Executor
}

// NewLintHandlerAdapter creates a handler adapter for linting.
func NewLintHandlerAdapter(tool *ToolDefinition, executor Executor) *LintHandlerAdapter {
	return &LintHandlerAdapter{
		tool:     tool,
		executor: executor,
	}
}

// Name returns the tool ID.
func (a *LintHandlerAdapter) Name() string {
	return a.tool.ID
}

// Lint executes the tool for a lint operation.
func (a *LintHandlerAdapter) Lint(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts LintOptions) int {
	execCtx := &ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     OperationLint,
		Files:         opts.Files,
		Placeholders: map[string]string{
			"{workspace}": workspaceRoot,
			"{module}":    filepath.Join(workspaceRoot, moduleRoot),
			"{output}":    outputDir,
		},
	}

	// Add fix flag if supported
	if opts.Fix {
		execCtx.ArgsOverrides = append(execCtx.ArgsOverrides, "--fix")
	}

	ctx := context.Background()
	result, err := a.executor.Execute(ctx, a.tool, execCtx)
	if err != nil {
		if logWriter != nil {
			io.WriteString(logWriter, "Error executing linter: "+err.Error()+"\n")
		}
		return 1
	}

	return result.ExitCode
}

// Requirements returns the tool's requirements.
func (a *LintHandlerAdapter) Requirements() []string {
	return a.tool.Requirements
}

// ValidateModule always returns nil for YAML tools.
func (a *LintHandlerAdapter) ValidateModule(moduleRoot, workspaceRoot string) error {
	return nil
}

// TestHandler is the interface for test handlers.
type TestHandler interface {
	Name() string
	Test(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts TestOptions) int
	Requirements() []string
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
	tool     *ToolDefinition
	executor Executor
}

// NewTestHandlerAdapter creates a handler adapter for testing.
func NewTestHandlerAdapter(tool *ToolDefinition, executor Executor) *TestHandlerAdapter {
	return &TestHandlerAdapter{
		tool:     tool,
		executor: executor,
	}
}

// Name returns the tool ID.
func (a *TestHandlerAdapter) Name() string {
	return a.tool.ID
}

// Test executes the tool for a test operation.
func (a *TestHandlerAdapter) Test(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts TestOptions) int {
	execCtx := &ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    module.GetComponentRoot(""),
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     OperationTest,
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
	result, err := a.executor.Execute(ctx, a.tool, execCtx)
	if err != nil {
		if logWriter != nil {
			io.WriteString(logWriter, "Error executing tests: "+err.Error()+"\n")
		}
		return 1
	}

	return result.ExitCode
}

// Requirements returns the tool's requirements.
func (a *TestHandlerAdapter) Requirements() []string {
	return a.tool.Requirements
}

// ScanHandler is the interface for security scan handlers.
type ScanHandler interface {
	Name() string
	Scan(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts ScanOptions) (int, []byte)
	Requirements() []string
}

// ScanOptions contains flags for controlling the scan process.
type ScanOptions struct {
	ScanType   string // Type of scan (sbom, vuln, secrets, sast)
	OutputPath string // Path for scan results
	Severity   string // Minimum severity level
}

// ScanHandlerAdapter wraps a ToolDefinition to implement ScanHandler.
type ScanHandlerAdapter struct {
	tool     *ToolDefinition
	executor Executor
}

// NewScanHandlerAdapter creates a handler adapter for scanning.
func NewScanHandlerAdapter(tool *ToolDefinition, executor Executor) *ScanHandlerAdapter {
	return &ScanHandlerAdapter{
		tool:     tool,
		executor: executor,
	}
}

// Name returns the tool ID.
func (a *ScanHandlerAdapter) Name() string {
	return a.tool.ID
}

// Scan executes the tool for a scan operation.
func (a *ScanHandlerAdapter) Scan(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts ScanOptions) (int, []byte) {
	execCtx := &ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     OperationScan,
		Placeholders: map[string]string{
			"{workspace}": workspaceRoot,
			"{module}":    filepath.Join(workspaceRoot, moduleRoot),
			"{output}":    outputDir,
		},
	}

	ctx := context.Background()
	result, err := a.executor.Execute(ctx, a.tool, execCtx)
	if err != nil {
		if logWriter != nil {
			io.WriteString(logWriter, "Error executing scanner: "+err.Error()+"\n")
		}
		return 1, nil
	}

	return result.ExitCode, result.Stdout
}

// Requirements returns the tool's requirements.
func (a *ScanHandlerAdapter) Requirements() []string {
	return a.tool.Requirements
}
