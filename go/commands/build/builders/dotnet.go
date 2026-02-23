// dotnet.go - Build handler for .NET build system
package builders

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	build "github.com/ready-to-release/eac/contracts/runner/0.1.0/build"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&DotnetHandler{})
}

// DotnetHandler builds .NET modules (libraries, CLIs, web apps).
type DotnetHandler struct{}

func (h *DotnetHandler) Name() string { return "dotnet" }

func (h *DotnetHandler) Capabilities() []string { return []string{"dotnet_module", "cross_compile"} }

func (h *DotnetHandler) Requirements() []string { return []string{"dotnet"} }

func (h *DotnetHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	componentRoot := filepath.Join(workspaceRoot, module.GetComponentRoot(component))
	patterns := []string{"*.csproj", "*.fsproj", "*.sln"}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(componentRoot, pattern))
		if len(matches) > 0 {
			return nil
		}
	}
	// Walk parent directories up to workspace root for .sln
	dir := componentRoot
	for {
		matches, _ := filepath.Glob(filepath.Join(dir, "*.sln"))
		if len(matches) > 0 {
			return nil
		}
		parent := filepath.Dir(dir)
		if parent == dir || len(dir) < len(workspaceRoot) {
			break
		}
		dir = parent
	}
	return fmt.Errorf("no .csproj, .fsproj, or .sln found for component %s (searched from %s)", component, componentRoot)
}

// IsContainer returns false as .NET builds run using the local dotnet toolchain.
func (h *DotnetHandler) IsContainer() bool { return false }

// IsHostInstalled returns true as .NET builds use the local dotnet toolchain.
func (h *DotnetHandler) IsHostInstalled() bool { return true }

func (h *DotnetHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		return nil
	}
	if !concrete.HasBuildArtifacts() {
		return nil
	}
	return listDotnetArtifacts(concrete)
}

func (h *DotnetHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, rawOpts any) int {
	opts, _ := rawOpts.(BuildOptions)
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		Logln(logWriter, "Error: invalid module type")
		return 1
	}
	return buildDotnetModule(concrete, workspaceRoot, outputDir, logWriter, opts)
}

// executeDotnetTool runs the dotnet tool with given args.
func executeDotnetTool(moduleRoot, workspaceRoot string, logWriter io.Writer, args []string, envOverrides map[string]string) int {
	registry := tool.GlobalRegistry()
	executor := tool.GlobalExecutor()

	toolDef := registry.GetOrAdhoc("dotnet")
	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		LogWriter:     logWriter,
		Operation:     core.ActionBuild,
		ArgsOverrides: args,
		EnvOverrides:  envOverrides,
		Placeholders: map[string]string{
			"{workspace}": workspaceRoot,
			"{module}":    moduleRoot,
		},
	}

	result, err := executor.Execute(context.Background(), toolDef, execCtx)
	if err != nil {
		Logln(logWriter, "Tool execution error: %v", err)
		return 1
	}
	return result.ExitCode
}

// buildDotnetModule builds a .NET module based on per-module artifact definitions.
func buildDotnetModule(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	componentName := opts.Component
	if componentName == "" {
		componentName = "dotnet"
	}
	moduleRoot := filepath.Join(workspaceRoot, module.GetComponentRoot(componentName))

	Logln(logWriter, "\n=== Building dotnet: %s ===", module.Moniker)

	if !module.HasComponent(componentName) {
		Logln(logWriter, "  Skipping: module '%s' has no .NET component", module.Moniker)
		return 0
	}

	// Step 1: dotnet restore
	Logln(logWriter, "Running: dotnet restore")
	if exitCode := executeDotnetTool(moduleRoot, workspaceRoot, logWriter, []string{"restore"}, nil); exitCode != 0 {
		Logln(logWriter, "dotnet restore failed")
		return exitCode
	}

	// Step 2: Build based on artifact definitions
	if !module.HasBuildArtifacts() {
		return buildDotnetLibrary(moduleRoot, workspaceRoot, outputDir, logWriter)
	}

	if module.HasExecutableArtifacts() {
		return buildDotnetPublish(module, moduleRoot, workspaceRoot, outputDir, logWriter, opts)
	}

	return buildDotnetLibrary(moduleRoot, workspaceRoot, outputDir, logWriter)
}

// buildDotnetLibrary builds a library module (compile-only verification).
func buildDotnetLibrary(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer) int {
	Logln(logWriter, "Running: dotnet build --no-restore")
	args := []string{"build", "--no-restore", "--configuration", "Release"}
	if exitCode := executeDotnetTool(moduleRoot, workspaceRoot, logWriter, args, nil); exitCode != 0 {
		Logln(logWriter, "dotnet build failed")
		return exitCode
	}
	Logln(logWriter, "Library module built successfully")
	return 0
}

// buildDotnetPublish builds a publishable .NET application.
func buildDotnetPublish(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	execArtifacts := module.GetArtifactsByType("executable")

	for _, artifact := range execArtifacts {
		rid := dotnetRIDFromArtifactID(artifact.ID)
		if rid == "" {
			continue
		}

		resolver := config.NewArtifactResolver(module.Moniker, "")
		outputName := resolver.ResolvePattern(artifact.Pattern)
		publishDir := filepath.Join(outputDir, rid)

		Logln(logWriter, "Publishing: %s -> %s", rid, outputName)

		args := []string{
			"publish",
			"--no-restore",
			"--configuration", "Release",
			"--runtime", rid,
			"--self-contained", "true",
			"--output", publishDir,
		}

		if exitCode := executeDotnetTool(moduleRoot, workspaceRoot, logWriter, args, nil); exitCode != 0 {
			Logln(logWriter, "dotnet publish failed for %s", rid)
			return exitCode
		}
	}

	Logln(logWriter, "All publish targets built successfully")
	return 0
}

// dotnetRIDFromArtifactID converts an artifact ID to a .NET Runtime Identifier.
func dotnetRIDFromArtifactID(id string) string {
	knownRIDs := map[string]string{
		"linux-x64":   "linux-x64",
		"linux-arm64": "linux-arm64",
		"win-x64":     "win-x64",
		"win-arm64":   "win-arm64",
		"osx-x64":     "osx-x64",
		"osx-arm64":   "osx-arm64",
	}
	if rid, ok := knownRIDs[id]; ok {
		return rid
	}
	goStyleRIDs := map[string]string{
		"linux-amd64":   "linux-x64",
		"windows-amd64": "win-x64",
		"darwin-amd64":  "osx-x64",
		"darwin-arm64":  "osx-arm64",
	}
	if rid, ok := goStyleRIDs[id]; ok {
		return rid
	}
	// If the ID itself looks like a RID (contains a dash), try it directly
	if strings.Contains(id, "-") {
		return id
	}
	return ""
}

// listDotnetArtifacts returns the artifacts that would be produced.
func listDotnetArtifacts(module *modules.ModuleContract) []string {
	var artifacts []string
	for _, artifact := range module.GetBuildArtifacts() {
		resolver := config.NewArtifactResolverWithPlatform(module.Moniker, "", runtime.GOOS, runtime.GOARCH)
		name := resolver.ResolvePattern(artifact.Pattern)
		artifacts = append(artifacts, name)
	}
	return artifacts
}

var _ build.BuilderPort = (*DotnetHandler)(nil)
