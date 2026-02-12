package tool

import (
	"context"
	"io"

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
