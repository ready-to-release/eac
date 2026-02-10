// Package reqnroll provides a test runner adapter for .NET Reqnroll BDD tests.
package reqnroll

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	dotnetadapter "github.com/ready-to-release/eac/go/adapters/dotnet"
	"github.com/ready-to-release/eac/go/adapters/nuget"
	"github.com/ready-to-release/eac/go/clibase/testrunners"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	testrunners.Register(&ReqnrollRunner{})
	testrunners.RegisterDescriptor(&testrunners.TestTypeDescriptor{
		TestType:      "reqnroll",
		IsBDD:         true,
		ComponentType: "gherkin",
		MonikerStyle:  "feature",
		FeatureTestTypeResolver: func(info testrunners.FeatureModuleInfo) bool {
			return info.HasDotnet
		},
		DefaultInferences: []testrunners.Inference{
			{
				TestTypes:   []string{"reqnroll"},
				ThenAddTags: []string{"@deps:dotnet"},
				Description: "Reqnroll tests require .NET SDK",
			},
		},
	})
}

// ReqnrollRunner handles .NET Reqnroll BDD test execution.
type ReqnrollRunner struct{}

// TestTypes returns the test types this runner handles.
func (r *ReqnrollRunner) TestTypes() []string {
	return []string{"reqnroll"}
}

// IsBDD returns true because Reqnroll is a BDD test framework.
func (r *ReqnrollRunner) IsBDD() bool {
	return true
}

// GetTestInfo extracts structured test metadata from a Reqnroll test reference.
func (r *ReqnrollRunner) GetTestInfo(ref testing.TestReference, workspaceRoot string, cfg any) *testrunners.TestInfo {
	eacCfg := cfg.(*config.EACConfig)

	relPath, err := filepath.Rel(workspaceRoot, ref.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)

	info := &testrunners.TestInfo{Language: "cs"}

	specsPrefix := eacCfg.Repository.Paths.SpecsRoot + "/"
	specRelPath := strings.TrimPrefix(relPath, specsPrefix)
	specRelPath = filepath.ToSlash(specRelPath)

	parts := strings.Split(specRelPath, "/")
	if len(parts) == 0 {
		return nil
	}
	info.ModuleMoniker = parts[0]

	if eacCfg.Repository.GetByMoniker(info.ModuleMoniker) == nil {
		return nil
	}

	info.TestRoot = r.FindTestRoot(relPath, cfg)
	if info.TestRoot == "" {
		return nil
	}

	featureFolderName := extractFeatureFolderName(relPath)
	info.PackageKey = featureFolderName + ":" + info.TestRoot + ":" + relPath
	info.DisplayName = featureFolderName + ":" + info.TestRoot

	return info
}

// FindTestRoot finds the .NET test project directory for a Reqnroll feature file.
func (r *ReqnrollRunner) FindTestRoot(featurePath string, cfg any) string {
	eacCfg := cfg.(*config.EACConfig)

	specsPrefix := eacCfg.Repository.Paths.SpecsRoot + "/"
	relPath := strings.TrimPrefix(filepath.ToSlash(featurePath), specsPrefix)
	relPath = filepath.ToSlash(relPath)

	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		return ""
	}
	moniker := parts[0]

	module := eacCfg.Repository.GetByMoniker(moniker)
	if module == nil {
		return ""
	}

	// Return the module's dotnet component root
	for _, comp := range module.Components {
		if comp != nil && comp.Root != "" {
			return filepath.ToSlash(comp.Root)
		}
	}
	return ""
}

// BuildPackagePath constructs the package path for test grouping.
func (r *ReqnrollRunner) BuildPackagePath(testRoot, featurePath string) string {
	if testRoot == "" {
		return ""
	}
	if featurePath != "" {
		featureFolderName := extractFeatureFolderName(featurePath)
		return featureFolderName + ":" + testRoot + ":" + featurePath
	}
	return testRoot
}

// extractFeatureFolderName extracts the feature folder name from a feature path.
func extractFeatureFolderName(featurePath string) string {
	featurePath = filepath.ToSlash(featurePath)
	dir := filepath.Dir(featurePath)
	return filepath.Base(dir)
}

// Execute runs Reqnroll BDD tests for a package.
func (r *ReqnrollRunner) Execute(pkgPath string, tests []testing.TestReference, logWriter io.Writer, cfg testrunners.RunConfig) testrunners.RunResult {
	start := time.Now()

	var displayName, relPkgPath, relFeatureFile string
	parts := strings.Split(pkgPath, ":")
	if len(parts) == 3 {
		displayName = parts[0] + ":" + parts[1]
		relPkgPath = parts[1]
		relFeatureFile = parts[2]
	} else if len(parts) == 1 {
		displayName = parts[0]
		relPkgPath = parts[0]
	} else {
		displayName = pkgPath
		relPkgPath = pkgPath
	}

	result := testrunners.RunResult{
		PackageName:   displayName,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	moduleRoot := filepath.Join(cfg.WorkspaceRoot, relPkgPath)

	// Prepare isolated NuGet environment
	isolation := nuget.NewNuGetIsolation(cfg.WorkspaceRoot)
	env, err := isolation.PrepareIsolatedEnv(moduleRoot, cfg.OutputDir)
	if err != nil {
		fmt.Fprintf(logWriter, "Failed to prepare isolated environment: %v\n", err)
		result.PackageFailed = true
		return result
	}

	// Restore (serialized)
	nuget.NuGetRestoreMu.Lock()
	restoreToolDef := tool.GlobalRegistry().GetOrAdhoc("dotnet")
	restoreResult, restoreErr := tool.GlobalExecutor().Execute(context.Background(), restoreToolDef, &tool.ExecutionContext{
		ModuleRoot:    env.WorkDir,
		FullEnv:       env.Env,
		ArgsOverrides: []string{"restore"},
	})
	nuget.NuGetRestoreMu.Unlock()
	if restoreResult != nil {
		fmt.Fprintf(logWriter, "%s%s\n", restoreResult.Stdout, restoreResult.Stderr)
	}
	if restoreErr != nil || (restoreResult != nil && restoreResult.ExitCode != 0) {
		fmt.Fprintf(logWriter, "dotnet restore failed: %v\n", restoreErr)
		result.PackageFailed = true
		return result
	}

	// Build test arguments
	trxPath := filepath.Join(cfg.OutputDir, "results.trx")
	testArgs := []string{
		"test",
		"--no-restore",
		"--logger", fmt.Sprintf("trx;LogFileName=%s", trxPath),
		"--results-directory", cfg.OutputDir,
	}

	// Filter to specific feature file if provided
	if relFeatureFile != "" {
		featurePath := filepath.Join(cfg.WorkspaceRoot, relFeatureFile)
		relPath, relErr := filepath.Rel(env.WorkDir, featurePath)
		if relErr == nil {
			featureName := strings.TrimSuffix(filepath.Base(relPath), ".feature")
			testArgs = append(testArgs, "--filter", fmt.Sprintf("FeatureTitle~%s", featureName))
		}
	}

	runCtx := cfg.Ctx
	if runCtx == nil {
		runCtx = context.Background()
	}

	fmt.Fprintf(logWriter, "=== Testing .NET Reqnroll BDD specs ===\n")
	fmt.Fprintf(logWriter, "Module root: %s (isolated: %s)\n", moduleRoot, env.WorkDir)
	fmt.Fprintf(logWriter, "Command: dotnet %s\n\n", strings.Join(testArgs, " "))

	dotnetTool := tool.GlobalRegistry().GetOrAdhoc("dotnet")
	execResult, runErr := tool.GlobalExecutor().Execute(runCtx, dotnetTool, &tool.ExecutionContext{
		ModuleRoot:    env.WorkDir,
		FullEnv:       env.Env,
		ArgsOverrides: testArgs,
	})
	if execResult != nil {
		fmt.Fprintf(logWriter, "%s%s\n", execResult.Stdout, execResult.Stderr)
	}

	// Parse TRX results (reuse dotnet adapter's TRX parser)
	if cfg.OutputDir != "" {
		if trxData, readErr := os.ReadFile(trxPath); readErr == nil {
			if ctrfReport := dotnetadapter.ConvertTRXToCTRF(trxData); ctrfReport != nil {
				if ctrfData, jsonErr := ctrfReport.ToJSON(); jsonErr == nil {
					jsonPath := filepath.Join(cfg.OutputDir, "cucumber.json")
					_ = os.WriteFile(jsonPath, ctrfData, 0o644)
				}
				result.TestsPassed = ctrfReport.Results.Summary.Passed
				result.TestsFailed = ctrfReport.Results.Summary.Failed
				result.TestsSkipped = ctrfReport.Results.Summary.Skipped
				result.TestsTotal = ctrfReport.Results.Summary.Tests
			}
		}
	}

	failed := runErr != nil || (execResult != nil && execResult.ExitCode != 0)
	if failed {
		result.PackageFailed = true
		if result.TestsTotal == 0 {
			result.TestsFailed = len(tests)
			result.TestsTotal = len(tests)
		}
	} else if result.TestsTotal == 0 {
		result.TestsPassed = len(tests)
		result.TestsTotal = len(tests)
	}

	result.Duration = time.Since(start)
	return result
}

var _ testrunners.TestTypeRunner = (*ReqnrollRunner)(nil)
