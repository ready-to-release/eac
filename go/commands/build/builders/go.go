// go.go - Build handler for Go build system
//
// Handler registration, type definitions, interface methods, and tool execution.
// Build logic is split across sibling files:
//   - go_build.go:    Build orchestration (buildGoModule, buildLibrary, buildTestModule, buildSingleBinary)
//   - go_cross.go:    Cross-compilation (buildCrossCompiledFromArtifacts)
//   - go_artifacts.go: Artifact listing
//   - go_version.go:  Version injection, ldflags, changelog parsing
//   - go_checksum.go: SHA256 checksum generation
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	build "github.com/ready-to-release/eac/contracts/runner/0.1.0/build"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&GoHandler{})
}

// GoHandler builds Go modules (libraries, CLIs, tests).
type GoHandler struct{}

// goToolOptions controls which tool (system go or container go-oci) to use.
type goToolOptions struct {
	// UseContainer forces use of container tool (go-oci) instead of system go.
	UseContainer bool

	// RaceDetection enables race detector (requires CGO, uses container).
	RaceDetection bool

	// CGOEnabled forces CGO support (uses container for reliable cross-platform CGO).
	CGOEnabled bool
}

// resolveToolName determines whether to use system 'go' or container 'go-oci'.
// Container is used when CGO or race detection is needed.
func (h *GoHandler) resolveToolName(opts goToolOptions) string {
	// Use container for specific scenarios
	if opts.UseContainer || opts.RaceDetection || opts.CGOEnabled {
		return "go-oci"
	}
	// Default: native go through tool system
	return "go"
}

// executeTool runs the go tool with given args and env overrides.
// Uses the tool system instead of direct exec.Command.
func (h *GoHandler) executeTool(ctx context.Context, toolName, moduleRoot, workspaceRoot string,
	logWriter io.Writer, args []string, envOverrides map[string]string) int {

	registry := tool.GlobalRegistry()
	executor := tool.GlobalExecutor()

	// Get tool definition
	toolDef, ok := registry.Get(toolName)
	if !ok {
		Logln(logWriter, "Tool not found: %s", toolName)
		return 1
	}

	// Build placeholders for mount path resolution
	placeholders := map[string]string{
		"{workspace}": workspaceRoot,
		"{module}":    moduleRoot,
	}

	// Build execution context
	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		LogWriter:     logWriter,
		Operation:     core.ActionBuild,
		ArgsOverrides: args,
		EnvOverrides:  envOverrides,
		Placeholders:  placeholders,
	}

	result, err := executor.Execute(ctx, toolDef, execCtx)
	if err != nil {
		Logln(logWriter, "Tool execution error: %v", err)
		return 1
	}

	return result.ExitCode
}

// executeGoTool is a package-level helper that executes go commands via the tool system.
// It uses the default "go" system tool unless specific scenarios (CGO, race) require "go-oci".
// When the module is not listed in go.work, GOWORK=off is set automatically so that
// ./... pattern resolution works within the standalone module.
func executeGoTool(moduleRoot, workspaceRoot string, logWriter io.Writer, args []string, envOverrides map[string]string) int {
	if !isModuleInGoWork(moduleRoot, workspaceRoot) {
		if envOverrides == nil {
			envOverrides = map[string]string{}
		}
		if _, ok := envOverrides["GOWORK"]; !ok {
			envOverrides["GOWORK"] = "off"
		}
	}
	h := &GoHandler{}
	return h.executeTool(context.Background(), "go", moduleRoot, workspaceRoot, logWriter, args, envOverrides)
}

// isModuleInGoWork checks whether moduleRoot is listed in the workspace go.work file.
func isModuleInGoWork(moduleRoot, workspaceRoot string) bool {
	goWorkPath := filepath.Join(workspaceRoot, "go.work")
	data, err := os.ReadFile(goWorkPath)
	if err != nil {
		return false // No go.work = standalone mode, no issue
	}

	// Normalize moduleRoot for comparison
	absModule, err := filepath.Abs(moduleRoot)
	if err != nil {
		return false
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "./") || strings.HasPrefix(line, ".\\") {
			usePath := filepath.Join(workspaceRoot, filepath.FromSlash(line))
			absUse, err := filepath.Abs(usePath)
			if err != nil {
				continue
			}
			if strings.EqualFold(absModule, absUse) {
				return true
			}
		}
	}
	return false
}

// executeGoToolWithEnv is like executeGoTool but for cross-compilation with GOOS/GOARCH.
func executeGoToolWithEnv(moduleRoot, workspaceRoot string, logWriter io.Writer, args []string, goos, goarch string) int {
	envOverrides := map[string]string{
		"GOOS":        goos,
		"GOARCH":      goarch,
		"CGO_ENABLED": "0",
	}
	return executeGoTool(moduleRoot, workspaceRoot, logWriter, args, envOverrides)
}

func (h *GoHandler) Name() string { return "go" }

func (h *GoHandler) Capabilities() []string { return []string{"go_module", "cross_compile"} }

func (h *GoHandler) Requirements() []string { return []string{"go"} }

func (h *GoHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	componentRoot := filepath.Join(workspaceRoot, module.GetComponentRoot(component))
	// Walk up from component root to find go.mod (may be in parent directory)
	goMod := findGoMod(componentRoot, workspaceRoot)
	if goMod == "" {
		return fmt.Errorf("go.mod not found for component %s (searched from %s)", component, componentRoot)
	}
	return nil
}

// IsContainer returns false as Go builds run using the local go toolchain.
func (h *GoHandler) IsContainer() bool { return false }

// IsHostInstalled returns true as Go builds use the local go toolchain.
func (h *GoHandler) IsHostInstalled() bool { return true }

// findGoMod walks up from dir to find go.mod, stopping at workspaceRoot.
func findGoMod(dir, workspaceRoot string) string {
	for {
		goMod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			return goMod
		}
		parent := filepath.Dir(dir)
		if parent == dir || len(dir) < len(workspaceRoot) {
			return ""
		}
		dir = parent
	}
}

func (h *GoHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		return nil
	}
	return listGoModuleArtifacts(concrete, workspaceRoot)
}

func (h *GoHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, rawOpts any) int {
	opts, _ := rawOpts.(BuildOptions)
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		Logln(logWriter, "Error: invalid module type")
		return 1
	}
	return buildGoModule(concrete, workspaceRoot, outputDir, logWriter, opts)
}

// isScriptOnlyModule returns true if module only has script components (pwsh/bash).
// These modules are expected to not have Go packages.
func isScriptOnlyModule(module *modules.ModuleContract) bool {
	return module.HasComponent("pwsh") || module.HasComponent("bash")
}

var _ build.BuilderPort = (*GoHandler)(nil)
