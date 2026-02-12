// Package dotnet provides a test runner adapter for .NET xUnit tests.
package dotnet

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/adapters/nuget"
	"github.com/ready-to-release/eac/go/clibase/testrunners"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	testrunners.Register(&DotnetTestRunner{})
	testrunners.RegisterDescriptor(&testrunners.TestTypeDescriptor{
		TestType:      "dotnet",
		IsBDD:         false,
		ComponentType: "dotnet",
		MonikerStyle:  "file",
		DefaultInferences: []testrunners.Inference{
			{
				TestTypes:   []string{"dotnet"},
				ThenAddTags: []string{"@deps:dotnet"},
				Description: ".NET tests require .NET SDK",
			},
			{
				TestTypes:   []string{"dotnet"},
				ThenAddTags: []string{"@L1"},
				Description: ".NET unit tests default to L1 (unit test level)",
			},
		},
	})
}

// DotnetTestRunner handles .NET xUnit test execution.
type DotnetTestRunner struct{}

// TestTypes returns the test types this runner handles.
func (r *DotnetTestRunner) TestTypes() []string {
	return []string{"dotnet"}
}

// IsBDD returns false because this handles xUnit unit tests, not BDD.
func (r *DotnetTestRunner) IsBDD() bool {
	return false
}

// GetTestInfo extracts structured test metadata from a .NET test reference.
func (r *DotnetTestRunner) GetTestInfo(ref testing.TestReference, workspaceRoot string, cfg any) *testrunners.TestInfo {
	eacCfg := cfg.(*config.EACConfig)

	relPath, err := filepath.Rel(workspaceRoot, ref.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)
	relDir := filepath.ToSlash(filepath.Dir(relPath))

	info := &testrunners.TestInfo{Language: "cs"}

	info.ModuleMoniker = findDotnetModuleForPath(relDir, eacCfg)
	if info.ModuleMoniker == "" {
		return nil
	}

	info.TestRoot = relDir
	info.PackageKey = relDir
	info.DisplayName = relDir

	return info
}

// findDotnetModuleForPath finds the module moniker for a given .NET path.
func findDotnetModuleForPath(relPath string, cfg *config.EACConfig) string {
	for i := range cfg.Repository.Modules {
		module := &cfg.Repository.Modules[i]
		for _, entry := range module.Components {
			if entry == nil || entry.Root == "" {
				continue
			}
			pkgRoot := filepath.ToSlash(entry.Root)
			if strings.HasPrefix(relPath, pkgRoot+"/") || relPath == pkgRoot {
				return module.Moniker
			}
		}
	}
	return ""
}

// FindTestRoot returns empty string - .NET tests live in the same project structure.
func (r *DotnetTestRunner) FindTestRoot(testPath string, cfg any) string {
	return ""
}

// BuildPackagePath constructs the package path for test grouping.
func (r *DotnetTestRunner) BuildPackagePath(testRoot, testPath string) string {
	return testRoot
}

// Execute runs .NET xUnit tests for a package and returns results.
func (r *DotnetTestRunner) Execute(pkgPath string, tests []testing.TestReference, logWriter io.Writer, cfg testrunners.RunConfig) testrunners.RunResult {
	start := time.Now()
	result := testrunners.RunResult{
		PackageName:   pkgPath,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	moduleRoot := filepath.Join(cfg.WorkspaceRoot, pkgPath)

	// Verify a .csproj or .sln exists
	if !hasDotnetProject(moduleRoot) {
		fmt.Fprintf(logWriter, "No .csproj or .sln found at %s\n", moduleRoot)
		result.PackageFailed = true
		return result
	}

	// Prepare isolated NuGet environment
	isolation := nuget.NewNuGetIsolation(cfg.WorkspaceRoot)
	env, err := isolation.PrepareIsolatedEnv(moduleRoot, cfg.OutputDir)
	if err != nil {
		fmt.Fprintf(logWriter, "Failed to prepare isolated environment: %v\n", err)
		result.PackageFailed = true
		return result
	}

	fmt.Fprintf(logWriter, "Using isolated environment: %s\n", env.WorkDir)

	// Step 1: dotnet restore (serialized)
	fmt.Fprintf(logWriter, "Restoring NuGet packages...\n")
	nuget.NuGetRestoreMu.Lock()
	restoreToolDef := tool.GlobalRegistry().GetOrAdhoc("dotnet")
	restoreExecCtx := &tool.ExecutionContext{
		ModuleRoot:    env.WorkDir,
		FullEnv:       env.Env,
		ArgsOverrides: []string{"restore"},
	}
	restoreResult, restoreErr := tool.GlobalExecutor().Execute(context.Background(), restoreToolDef, restoreExecCtx)
	nuget.NuGetRestoreMu.Unlock()
	if restoreResult != nil {
		fmt.Fprintf(logWriter, "%s%s\n", restoreResult.Stdout, restoreResult.Stderr)
	}
	if restoreErr != nil || (restoreResult != nil && restoreResult.ExitCode != 0) {
		fmt.Fprintf(logWriter, "dotnet restore failed: %v\n", restoreErr)
		result.PackageFailed = true
		return result
	}

	// Step 2: dotnet test with TRX output
	trxPath := filepath.Join(cfg.OutputDir, "results.trx")
	testArgs := []string{
		"test",
		"--no-restore",
		"--logger", fmt.Sprintf("trx;LogFileName=%s", trxPath),
		"--results-directory", cfg.OutputDir,
	}

	runCtx := cfg.Ctx
	if runCtx == nil {
		runCtx = context.Background()
	}

	fmt.Fprintf(logWriter, "=== Testing .NET xUnit tests ===\n")
	fmt.Fprintf(logWriter, "Module root: %s (isolated: %s)\n", moduleRoot, env.WorkDir)
	fmt.Fprintf(logWriter, "Command: dotnet %s\n\n", strings.Join(testArgs, " "))

	dotnetTool := tool.GlobalRegistry().GetOrAdhoc("dotnet")
	execResult, runErr := tool.GlobalExecutor().Execute(runCtx, dotnetTool, &tool.ExecutionContext{
		ModuleRoot:    env.WorkDir,
		FullEnv:       env.Env,
		ArgsOverrides: testArgs,
		LogWriter:     logWriter,
	})
	if execResult != nil {
		fmt.Fprintf(logWriter, "%s%s\n", execResult.Stdout, execResult.Stderr)
	}

	// Parse TRX and convert to CTRF
	if cfg.OutputDir != "" {
		if trxData, readErr := os.ReadFile(trxPath); readErr == nil {
			if ctrfReport := ConvertTRXToCTRF(trxData); ctrfReport != nil {
				if ctrfData, jsonErr := ctrfReport.ToJSON(); jsonErr == nil {
					jsonPath := filepath.Join(cfg.OutputDir, "unit.json")
					_ = os.WriteFile(jsonPath, ctrfData, 0o644)
					fmt.Fprintf(logWriter, "CTRF JSON saved to unit.json (%d bytes)\n", len(ctrfData))

					result.TestsPassed = ctrfReport.Results.Summary.Passed
					result.TestsFailed = ctrfReport.Results.Summary.Failed
					result.TestsSkipped = ctrfReport.Results.Summary.Skipped
					result.TestsTotal = ctrfReport.Results.Summary.Tests
				}
			}
		}
	}

	failed := runErr != nil || (execResult != nil && execResult.ExitCode != 0)
	if failed || result.TestsFailed > 0 {
		result.PackageFailed = true
	}

	// Fallback counts if TRX parsing failed
	if result.TestsTotal == 0 {
		if failed {
			result.TestsFailed = len(tests)
			result.PackageFailed = true
		} else {
			result.TestsPassed = len(tests)
		}
		result.TestsTotal = len(tests)
	}

	result.Duration = time.Since(start)
	return result
}

// hasDotnetProject checks if a .csproj or .sln exists in the directory.
func hasDotnetProject(dir string) bool {
	patterns := []string{"*.csproj", "*.fsproj", "*.sln"}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(dir, pattern))
		if len(matches) > 0 {
			return true
		}
	}
	return false
}

// Ensure DotnetTestRunner implements the interface (compile-time check).
var _ testrunners.TestTypeRunner = (*DotnetTestRunner)(nil)
