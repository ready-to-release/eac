// go_build.go - Build orchestration for Go modules.
//
// Contains the main build dispatch (buildGoModule) and build strategies
// for libraries, test modules, and single-binary executables.
// Cross-compilation is in go_cross.go.
package builders

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/tool"
)

// buildGoModule builds any Go module based on per-module artifact definitions.
// Behavior is driven by artifacts defined in repository.yml:
//   - No artifacts: library (compile-only verification)
//   - Single executable: builds binary for current platform
//   - Multiple executables: cross-compiled binaries
//   - Test artifacts: runs tests and captures results
func buildGoModule(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	// Use the actual component name from build options (supports bundle sub-components like "ai", "docker")
	// Falls back to "go" for standard single-component modules
	componentName := opts.Component
	if componentName == "" {
		componentName = "go"
	}
	moduleRoot := filepath.Join(workspaceRoot, module.GetComponentRoot(componentName))

	Logln(logWriter, "\n=== Building go: %s ===", module.Moniker)

	// Check if module has this go component
	hasGoModule := module.HasComponent(componentName)

	// Skip if not a go module - this is expected for script-only modules (pwsh, bash)
	// that have test-impl components triggering Go builds
	if !hasGoModule {
		if !isScriptOnlyModule(module) {
			Logln(logWriter, "  Skipping: module '%s' has no Go package", module.Moniker)
		}
		return 0
	}

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		Logln(logWriter, "Running: go mod tidy")
		if exitCode := executeGoTool(moduleRoot, workspaceRoot, logWriter, []string{"mod", "tidy"}, nil); exitCode != 0 {
			Logln(logWriter, "❌ go mod tidy failed")
			return exitCode
		}
	}

	// Step 2: go generate
	Logln(logWriter, "Running: go generate ./...")
	if exitCode := executeGoTool(moduleRoot, workspaceRoot, logWriter, []string{"generate", "./..."}, nil); exitCode != 0 {
		Logln(logWriter, "❌ go generate failed")
		return exitCode
	}

	// Step 3: Build based on per-module artifact definitions
	if !module.HasBuildArtifacts() {
		// No artifacts = library (compile-only verification)
		return buildLibrary(module, moduleRoot, workspaceRoot, outputDir, logWriter)
	}

	// Check artifact types
	hasExecutables := module.HasExecutableArtifacts()
	hasTests := module.HasTestArtifacts()

	if hasTests {
		// Test module - run tests and capture results
		return buildTestModule(module, moduleRoot, workspaceRoot, outputDir, logWriter, opts)
	}

	if hasExecutables {
		execArtifacts := module.GetArtifactsByType("executable")
		if len(execArtifacts) == 1 {
			// Single executable - build for current platform
			return buildSingleBinaryFromArtifact(module, moduleRoot, workspaceRoot, outputDir, logWriter, opts, execArtifacts[0])
		}
		// Multiple executables - cross-compile
		return buildCrossCompiledFromArtifacts(module, moduleRoot, workspaceRoot, outputDir, logWriter, opts, execArtifacts)
	}

	// Fallback: library behavior
	return buildLibrary(module, moduleRoot, workspaceRoot, outputDir, logWriter)
}

// buildLibrary builds a library module (compile-only verification).
func buildLibrary(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer) int {
	Logln(logWriter, "Running: go build ./...")
	if exitCode := executeGoTool(moduleRoot, workspaceRoot, logWriter, []string{"build", "./..."}, nil); exitCode != 0 {
		Logln(logWriter, "❌ go build failed")
		return exitCode
	}

	Logln(logWriter, "✅ Library module built successfully")
	return 0
}

// buildTestModule runs tests and captures results.
func buildTestModule(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "Running: go test ./... -json")

	// First verify it compiles
	if exitCode := executeGoTool(moduleRoot, workspaceRoot, logWriter, []string{"build", "./..."}, nil); exitCode != 0 {
		Logln(logWriter, "❌ go build failed")
		return exitCode
	}

	// Run tests with JSON output for results capture
	// StdoutWriter redirects test JSON output to results file,
	// StderrWriter sends diagnostic output to logWriter
	resultsPath := filepath.Join(outputDir, "results.json")
	resultsFile, err := os.Create(resultsPath)
	if err != nil {
		Logln(logWriter, "❌ Failed to create results file: %v", err)
		return 1
	}
	defer resultsFile.Close()

	registry := tool.GlobalRegistry()
	executor := tool.GlobalExecutor()
	toolDef := registry.GetOrAdhoc("go")

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		LogWriter:     logWriter,
		StdoutWriter:  resultsFile,
		StderrWriter:  logWriter,
		Operation:     core.ActionBuild,
		ArgsOverrides: []string{"test", "./...", "-json"},
		Placeholders: map[string]string{
			"{workspace}": workspaceRoot,
			"{module}":    moduleRoot,
		},
	}

	result, err := executor.Execute(context.Background(), toolDef, execCtx)
	if err != nil {
		// Test failures are expected - capture the results anyway
		Logln(logWriter, "⚠️  Tests completed with failures")
	} else if result.ExitCode != 0 {
		Logln(logWriter, "⚠️  Tests completed with failures")
	}

	Logln(logWriter, "✅ Test module built and results captured")
	return 0
}

// buildSingleBinaryFromArtifact builds a single binary from a per-module artifact definition.
func buildSingleBinaryFromArtifact(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions, artifact config.ModuleArtifact) int {
	// Resolve artifact pattern to binary name
	resolver := config.NewArtifactResolverWithPlatform(module.Moniker, "", runtime.GOOS, runtime.GOARCH)
	binaryName := resolver.ResolvePattern(artifact.Pattern)
	binaryPath := filepath.Join(outputDir, binaryName)

	Logln(logWriter, "Running: go build -o %s", binaryPath)
	exitCode := executeGoTool(moduleRoot, workspaceRoot, logWriter, []string{"build", "-o", binaryPath}, nil)
	if exitCode == 0 {
		Logln(logWriter, "✅ Built executable: %s", binaryName)
	}
	return exitCode
}
