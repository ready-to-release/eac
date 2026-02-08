// Package tool provides handler adapters that bridge the new tool system
// with existing handler core.
package tool

import (
	"context"
	"io"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/environments"
)

// BuildHandler is the interface for build handlers.
// This matches the existing builders.Handler interface to enable seamless integration.
type BuildHandler interface {
	// Name returns the handler identifier (e.g., "go", "mkdocs", "docker")
	Name() string

	// Build executes the build for a module.
	// Returns exit code (0 = success, non-zero = failure).
	Build(module core.ModuleContractPort, workspaceRoot, outputDir string,
		logWriter io.Writer, opts BuildOptions) int

	// ListArtifacts returns artifact paths that would be produced.
	// Paths are relative to the module's output directory.
	ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string

	// Requirements returns system dependencies required by this handler.
	// Used for early validation (e.g., ["go", "docker"]).
	Requirements() []string

	// ValidateModule checks if a module's configuration is valid for a specific component.
	// Returns nil if valid, or an error describing the problem.
	ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error

	// IsContainer returns true if this handler runs in a Docker container.
	IsContainer() bool

	// IsHostInstalled returns true if this handler runs using host-installed tools (not containers).
	IsHostInstalled() bool
}

// BuildOptions contains flags for controlling the build process.
// This matches the existing builders.BuildOptions structure.
type BuildOptions struct {
	TidyFirst          bool                       // Run go mod tidy before building
	Version            string                     // Version to inject via ldflags
	DryRun             bool                       // Simulate build without actually running it
	RequestedArtifacts []string                   // Specific artifact IDs to build (empty = default artifacts, "*" = all)
	Component          string                     // Specific component to build (empty = all components)
	Reproducible       bool                       // Force rebuild of MkDocs HTML even if staging unchanged (CI mode)
	ForceRebuild       bool                       // Force full rebuild, ignore cache (--rebuild flag)
	Weight             int                        // Resource multiplier for container builds (default: 1)
	CacheConfig        *cache.Config              // Cache control configuration (--skip-cache flag)
	ArtifactsMode      environments.ArtifactsMode // Artifact scope mode (all, devbox)
}

// BaseHandlerAdapter provides common handler adapter functionality.
// Embed this in specific handler adapters to avoid repeating Name/Requirements/IsContainer/IsHostInstalled.
type BaseHandlerAdapter struct {
	Tool     *ToolDefinition
	Executor Executor
}

// Name returns the tool ID.
func (b *BaseHandlerAdapter) Name() string {
	return b.Tool.ID
}

// Requirements returns the tool's requirements.
func (b *BaseHandlerAdapter) Requirements() []string {
	return b.Tool.Requirements
}

// IsContainer returns true if this handler runs in a Docker container.
func (b *BaseHandlerAdapter) IsContainer() bool {
	return b.Tool.Type == ToolTypeContainer
}

// IsHostInstalled returns true if this handler runs using host-installed tools.
func (b *BaseHandlerAdapter) IsHostInstalled() bool {
	return b.Tool.Type == ToolTypeSystem
}

// ToolHandlerAdapter wraps a ToolDefinition to implement BuildHandler.
// This allows YAML-defined tools to be used with the existing handler system.
type ToolHandlerAdapter struct {
	BaseHandlerAdapter
}

// NewToolHandlerAdapter creates a handler adapter for a tool definition.
func NewToolHandlerAdapter(tool *ToolDefinition, executor Executor) *ToolHandlerAdapter {
	return &ToolHandlerAdapter{
		BaseHandlerAdapter: BaseHandlerAdapter{Tool: tool, Executor: executor},
	}
}

// Build executes the tool for a build operation.
func (a *ToolHandlerAdapter) Build(module core.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, opts BuildOptions) int {

	// Build execution context
	// Note: {module} is the RELATIVE path for use in container workdir (e.g., /docs/{module})
	// while ModuleRoot is the relative path from workspace for host operations
	moduleRelPath := module.GetComponentRoot(opts.Component)
	execCtx := &ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRelPath,
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     core.ActionBuild,
		Placeholders: map[string]string{
			"{workspace}": workspaceRoot,
			"{module}":    moduleRelPath, // Relative path for container workdir
			"{output}":    outputDir,
		},
	}

	// Add version to args if specified
	if opts.Version != "" && a.Tool.Type == ToolTypeSystem {
		execCtx.ArgsOverrides = append(execCtx.ArgsOverrides, "-ldflags", "-X main.Version="+opts.Version)
	}

	// Execute the tool
	ctx := context.Background()
	result, err := a.Executor.Execute(ctx, a.Tool, execCtx)
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
func (a *ToolHandlerAdapter) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	// YAML-defined tools don't specify artifacts
	// This is intentionally minimal - complex builds should use native handlers
	return nil
}

// ValidateModule validates the tool can run - always returns nil for YAML tools.
// Complex validation should use native handlers.
func (a *ToolHandlerAdapter) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	// Basic validation - just check if this is a container tool that Docker is required
	if a.Tool.Type == ToolTypeContainer {
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
	IsContainer() bool
	IsHostInstalled() bool
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
	BaseHandlerAdapter
}

// NewLintHandlerAdapter creates a handler adapter for linting.
func NewLintHandlerAdapter(tool *ToolDefinition, executor Executor) *LintHandlerAdapter {
	return &LintHandlerAdapter{
		BaseHandlerAdapter: BaseHandlerAdapter{Tool: tool, Executor: executor},
	}
}

// Lint executes the tool for a lint operation.
func (a *LintHandlerAdapter) Lint(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts LintOptions) int {
	execCtx := &ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     core.ActionLint,
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
	result, err := a.Executor.Execute(ctx, a.Tool, execCtx)
	if err != nil {
		if logWriter != nil {
			io.WriteString(logWriter, "Error executing linter: "+err.Error()+"\n")
		}
		return 1
	}

	return result.ExitCode
}

// ValidateModule always returns nil for YAML tools.
func (a *LintHandlerAdapter) ValidateModule(moduleRoot, workspaceRoot string) error {
	return nil
}

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
			"{module}":    filepath.Join(workspaceRoot, moduleRoot),
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
