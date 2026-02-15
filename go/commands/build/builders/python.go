// python.go - Build handler for Python build system
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&PythonHandler{})
}

// PythonHandler builds Python modules (libraries, CLIs, test-only).
type PythonHandler struct{}

func (h *PythonHandler) Name() string { return "python" }

func (h *PythonHandler) Capabilities() []string { return []string{"python_module"} }

func (h *PythonHandler) Requirements() []string { return []string{"python"} }

func (h *PythonHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	componentRoot := filepath.Join(workspaceRoot, module.GetComponentRoot(component))
	// Check for pyproject.toml or setup.py
	if _, err := os.Stat(filepath.Join(componentRoot, "pyproject.toml")); err == nil {
		return nil
	}
	if _, err := os.Stat(filepath.Join(componentRoot, "setup.py")); err == nil {
		return nil
	}
	return fmt.Errorf("neither pyproject.toml nor setup.py found for component %s (searched at %s)", component, componentRoot)
}

// IsContainer returns false as Python builds run using the local Python toolchain.
func (h *PythonHandler) IsContainer() bool { return false }

// IsHostInstalled returns true as Python builds use the local Python toolchain.
func (h *PythonHandler) IsHostInstalled() bool { return true }

func (h *PythonHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		return nil
	}
	return listPythonModuleArtifacts(concrete)
}

func (h *PythonHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		Logln(logWriter, "Error: invalid module type")
		return 1
	}
	return buildPythonModule(concrete, workspaceRoot, outputDir, logWriter, opts)
}

// listPythonModuleArtifacts returns artifacts produced by a Python module build.
func listPythonModuleArtifacts(module *modules.ModuleContract) []string {
	if !module.HasComponent("python") {
		return nil
	}

	// Check for per-module artifact definitions
	if module.HasBuildArtifacts() {
		var artifacts []string
		for _, artifact := range module.GetBuildArtifacts() {
			artifacts = append(artifacts, artifact.Pattern)
		}
		return artifacts
	}

	// Default: wheel + sdist
	return []string{"dist/*.whl", "dist/*.tar.gz"}
}

// buildPythonModule builds a Python module.
func buildPythonModule(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	componentName := opts.Component
	if componentName == "" {
		componentName = "python"
	}
	moduleRoot := filepath.Join(workspaceRoot, module.GetComponentRoot(componentName))

	Logln(logWriter, "\n=== Building python: %s ===", module.Moniker)

	if !module.HasComponent(componentName) {
		Logln(logWriter, "  Skipping: module '%s' has no Python package", module.Moniker)
		return 0
	}

	// Check for pyproject.toml
	pyprojectPath := filepath.Join(moduleRoot, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); os.IsNotExist(err) {
		Logln(logWriter, "No pyproject.toml found at %s", moduleRoot)
		return 1
	}

	// Step 1: Compile check (verify all .py files are valid)
	Logln(logWriter, "Running: python -m py_compile check")
	exitCode := executePythonTool(moduleRoot, workspaceRoot, outputDir, logWriter,
		[]string{"-m", "compileall", "-q", "src/"})
	if exitCode != 0 {
		Logln(logWriter, "Python compile check failed")
		return exitCode
	}

	// Step 2: Build wheel + sdist if build tool is available
	Logln(logWriter, "Running: python -m build")
	exitCode = executePythonTool(moduleRoot, workspaceRoot, outputDir, logWriter,
		[]string{"-m", "build", "--outdir", outputDir})
	if exitCode != 0 {
		// Fall back to compile-only if python -m build is not available
		Logln(logWriter, "python -m build failed (build package may not be installed)")
		Logln(logWriter, "Falling back to compile-only verification")
		return 0
	}

	Logln(logWriter, "Python module built successfully")
	return 0
}

// executePythonTool runs a Python command via the tool system.
func executePythonTool(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, args []string) int {
	registry := tool.GlobalRegistry()
	executor := tool.GlobalExecutor()

	toolDef, ok := registry.Get("python")
	if !ok {
		Logln(logWriter, "Python tool not found in registry")
		return 1
	}

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     core.ActionBuild,
		ArgsOverrides: args,
		Placeholders: map[string]string{
			"{workspace}": workspaceRoot,
			"{module}":    moduleRoot,
			"{output}":    outputDir,
		},
	}

	result, err := executor.Execute(context.Background(), toolDef, execCtx)
	if err != nil {
		Logln(logWriter, "Python tool execution error: %v", err)
		return 1
	}

	return result.ExitCode
}
