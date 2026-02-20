package tool

import (
	"context"
	"io"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

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
	// Compute relative module path for container workdir resolution.
	// {module} is the absolute host path (for bind mounts).
	// {module-rel} is the workspace-relative path (for container workdir).
	moduleRel, _ := filepath.Rel(workspaceRoot, moduleRoot)
	execCtx := &ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     core.ActionLint,
		Files:         opts.Files,
		Placeholders: map[string]string{
			"{workspace}":  workspaceRoot,
			"{module}":     moduleRoot,
			"{module-rel}": filepath.ToSlash(moduleRel),
			"{output}":     outputDir,
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
