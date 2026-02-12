// Package behave provides a test runner adapter for Python behave BDD tests.
package behave

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/adapters/pip"
	"github.com/ready-to-release/eac/go/clibase/testrunners"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	RegisterWith(testrunners.DefaultRegistry())
}

// RegisterWith registers the BehaveRunner and its descriptor with the given registry.
// This enables testing with an isolated registry instead of relying on global state.
func RegisterWith(reg *testrunners.Registry) {
	reg.Register(&BehaveRunner{})
	reg.RegisterDescriptor(&testrunners.TestTypeDescriptor{
		TestType:      "behave",
		IsBDD:         true,
		ComponentType: "gherkin",
		MonikerStyle:  "feature",
		FeatureTestTypeResolver: func(info testrunners.FeatureModuleInfo) bool {
			return info.HasPython
		},
		DefaultInferences: []testrunners.Inference{
			{
				TestTypes:   []string{"behave"},
				ThenAddTags: []string{"@deps:python"},
				Description: "Behave tests require Python toolchain",
			},
			{
				TestTypes:   []string{"behave"},
				ThenAddTags: []string{"@L2"},
				Description: "Behave BDD tests are L2 by default",
			},
		},
	})
}

// BehaveRunner handles Python behave BDD test execution.
type BehaveRunner struct{}

// TestTypes returns the test types this runner handles.
func (r *BehaveRunner) TestTypes() []string {
	return []string{"behave"}
}

// IsBDD returns true because behave is a BDD test framework.
func (r *BehaveRunner) IsBDD() bool {
	return true
}

// GetTestInfo extracts structured test metadata from a behave test reference.
func (r *BehaveRunner) GetTestInfo(ref testing.TestReference, workspaceRoot string, cfg any) *testrunners.TestInfo {
	eacCfg := cfg.(*config.EACConfig)

	relPath, err := filepath.Rel(workspaceRoot, ref.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)

	info := &testrunners.TestInfo{Language: "python"}

	// Extract module moniker from specs path
	specsPrefix := eacCfg.Repository.Paths.SpecsRoot + "/"
	specRelPath := strings.TrimPrefix(relPath, specsPrefix)
	specRelPath = filepath.ToSlash(specRelPath)

	// Get module moniker from first path component
	parts := strings.Split(specRelPath, "/")
	if len(parts) == 0 {
		return nil
	}
	info.ModuleMoniker = parts[0]

	// Verify module exists
	if eacCfg.Repository.GetByMoniker(info.ModuleMoniker) == nil {
		return nil
	}

	// Find test root
	info.TestRoot = r.FindTestRoot(relPath, cfg)
	if info.TestRoot == "" {
		return nil
	}

	// Build package key and display name
	featureFolderName := extractFeatureFolderName(relPath)
	info.PackageKey = featureFolderName + ":" + info.TestRoot + ":" + relPath
	info.DisplayName = featureFolderName + ":" + info.TestRoot

	return info
}

// FindTestRoot finds the module root for a behave feature file.
func (r *BehaveRunner) FindTestRoot(featurePath string, cfg any) string {
	eacCfg := cfg.(*config.EACConfig)

	// Extract module moniker from specs path
	specsPrefix := eacCfg.Repository.Paths.SpecsRoot + "/"
	relPath := strings.TrimPrefix(filepath.ToSlash(featurePath), specsPrefix)
	relPath = strings.TrimPrefix(relPath, strings.ReplaceAll(specsPrefix, "/", "\\"))
	relPath = filepath.ToSlash(relPath)

	// Get module moniker (first path component)
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		return ""
	}
	moniker := parts[0]

	// Look up the module by moniker
	module, ok := eacCfg.Repository.GetModule(moniker)
	if !ok {
		return ""
	}

	// Return the module's Python component root
	pyRoot := module.Components.GetComponentRoot("python")
	if pyRoot == "" {
		return ""
	}
	return filepath.ToSlash(pyRoot)
}

// BuildPackagePath constructs the package path for test grouping.
func (r *BehaveRunner) BuildPackagePath(testRoot, featurePath string) string {
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

// parsedPackagePath holds the parsed components of a behave package path.
type parsedPackagePath struct {
	displayName    string
	relPkgPath     string
	relFeatureFile string
}

// parsePackagePath parses a package path in the format "featureName:moduleRoot:featurePath"
// or "moduleRoot" into its component parts.
func parsePackagePath(pkgPath string) parsedPackagePath {
	parts := strings.Split(pkgPath, ":")
	switch len(parts) {
	case 3:
		return parsedPackagePath{
			displayName:    parts[0] + ":" + parts[1],
			relPkgPath:     parts[1],
			relFeatureFile: parts[2],
		}
	case 1:
		return parsedPackagePath{
			displayName: parts[0],
			relPkgPath:  parts[0],
		}
	default:
		return parsedPackagePath{
			displayName: pkgPath,
			relPkgPath:  pkgPath,
		}
	}
}

// collectBehaveResults parses behave JSON output and populates the result with CTRF data.
// Returns true if results were successfully parsed.
func collectBehaveResults(result *testrunners.RunResult, outputDir string, logWriter io.Writer) bool {
	if outputDir == "" {
		return false
	}
	jsonPath := filepath.Join(outputDir, "behave.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return false
	}
	ctrfReport := convertBehaveJSONToCTRF(data)
	if ctrfReport == nil {
		return false
	}
	result.TestsPassed = ctrfReport.Results.Summary.Passed
	result.TestsFailed = ctrfReport.Results.Summary.Failed
	result.TestsSkipped = ctrfReport.Results.Summary.Skipped
	result.TestsTotal = ctrfReport.Results.Summary.Tests

	if ctrfData, err := ctrfReport.ToJSON(); err == nil {
		ctrfPath := filepath.Join(outputDir, "bdd.json")
		_ = os.WriteFile(ctrfPath, ctrfData, 0o644)
		fmt.Fprintf(logWriter, "CTRF JSON saved to bdd.json (%d bytes)\n", len(ctrfData))
	}
	return true
}

// executionFailed returns true if the execution result indicates failure.
func executionFailed(execResult *tool.ExecutionResult, runErr error) bool {
	return runErr != nil || (execResult != nil && execResult.ExitCode != 0)
}

// Execute runs Python behave tests for a package.
func (r *BehaveRunner) Execute(pkgPath string, tests []testing.TestReference, logWriter io.Writer, cfg testrunners.RunConfig) testrunners.RunResult {
	start := time.Now()
	pkg := parsePackagePath(pkgPath)

	result := testrunners.RunResult{
		PackageName:   pkg.displayName,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	moduleRoot := filepath.Join(cfg.WorkspaceRoot, pkg.relPkgPath)

	// Check if pyproject.toml exists
	pyprojectToml := filepath.Join(moduleRoot, "pyproject.toml")
	if _, err := os.Stat(pyprojectToml); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "No pyproject.toml found at %s\n", pyprojectToml)
		result.PackageFailed = true
		return result
	}

	// Resolve python tool via the tool system (auto/system/container)
	pythonTool, ok := tool.GlobalRegistry().Get("python")
	if !ok {
		fmt.Fprintf(logWriter, "Python tool not found in registry\n")
		result.PackageFailed = true
		return result
	}

	runCtx := cfg.Ctx
	if runCtx == nil {
		runCtx = context.Background()
	}

	var execResult *tool.ExecutionResult
	var runErr error

	if pythonTool.Type == tool.ToolTypeContainer {
		execResult, runErr = r.executeContainer(runCtx, pythonTool, moduleRoot, pkg.relFeatureFile, logWriter, cfg)
	} else {
		execResult, runErr = r.executeSystem(runCtx, pythonTool, moduleRoot, pkg.relFeatureFile, logWriter, cfg)
	}

	if execResult != nil {
		fmt.Fprintf(logWriter, "%s%s\n", execResult.Stdout, execResult.Stderr)
	}

	// Parse JSON results, falling back to synthetic counts
	if !collectBehaveResults(&result, cfg.OutputDir, logWriter) {
		failed := executionFailed(execResult, runErr)
		result.TestsTotal = 1
		if failed {
			result.TestsFailed = 1
		} else {
			result.TestsPassed = 1
		}
	}

	if executionFailed(execResult, runErr) {
		result.PackageFailed = true
		fmt.Fprintf(logWriter, "behave tests failed\n")
	} else {
		fmt.Fprintf(logWriter, "behave tests passed\n")
	}

	result.Duration = time.Since(start)
	return result
}

// executeSystem runs behave using a local system Python with venv isolation.
func (r *BehaveRunner) executeSystem(ctx context.Context, pythonTool *tool.ToolDefinition, moduleRoot, relFeatureFile string, logWriter io.Writer, cfg testrunners.RunConfig) (*tool.ExecutionResult, error) {
	// Prepare isolated pip environment
	isolation := pip.NewPipIsolation(cfg.WorkspaceRoot)
	env, err := isolation.PrepareIsolatedEnv(moduleRoot, cfg.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare isolated environment: %v", err)
	}

	fmt.Fprintf(logWriter, "Using isolated environment: %s\n", env.WorkDir)

	// Create venv if needed
	if _, err := os.Stat(env.PythonBin); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "Creating virtual environment at %s...\n", env.VenvDir)
		var venvResult *tool.ExecutionResult
		var venvErr error
		_ = pip.WithInstallLock(func() error {
			venvExecCtx := &tool.ExecutionContext{
				ModuleRoot:    env.WorkDir,
				ArgsOverrides: []string{"-m", "venv", env.VenvDir},
			}
			venvResult, venvErr = tool.GlobalExecutor().Execute(ctx, pythonTool, venvExecCtx)
			return venvErr
		})
		if venvResult != nil && (len(venvResult.Stdout) > 0 || len(venvResult.Stderr) > 0) {
			fmt.Fprintf(logWriter, "%s%s\n", venvResult.Stdout, venvResult.Stderr)
		}
		if venvErr != nil || (venvResult != nil && venvResult.ExitCode != 0) {
			return venvResult, fmt.Errorf("venv creation failed: %v", venvErr)
		}
	}

	// Install dependencies using the resolved python tool with venv env
	// Skip reinstallation when deps marker indicates pyproject.toml hasn't changed
	if env.DepsUpToDate {
		fmt.Fprintf(logWriter, "Dependencies up-to-date (cached venv), skipping pip install\n")
	} else {
		fmt.Fprintf(logWriter, "Installing Python dependencies...\n")
		var installResult *tool.ExecutionResult
		var installErr error
		_ = pip.WithInstallLock(func() error {
			pipExecCtx := &tool.ExecutionContext{
				ModuleRoot:    env.WorkDir,
				FullEnv:       env.Env,
				ArgsOverrides: []string{"-m", "pip", "install", "-e", ".[dev,test]", "--quiet"},
			}
			installResult, installErr = tool.GlobalExecutor().Execute(ctx, pythonTool, pipExecCtx)
			return installErr
		})
		if installResult != nil && (len(installResult.Stdout) > 0 || len(installResult.Stderr) > 0) {
			fmt.Fprintf(logWriter, "%s%s\n", installResult.Stdout, installResult.Stderr)
		}
		if installErr != nil || (installResult != nil && installResult.ExitCode != 0) {
			return installResult, fmt.Errorf("pip install failed: %v", installErr)
		}
		pip.MarkDepsInstalled(env.VenvDir, moduleRoot)
	}

	// Build behave command
	behaveArgs := []string{"-m", "behave", "--format", "json"}
	if cfg.OutputDir != "" {
		jsonPath := filepath.Join(cfg.OutputDir, "behave.json")
		behaveArgs = append(behaveArgs, "--outfile", jsonPath)
	}

	// Add tag filters
	if len(cfg.SuiteTagFilter.Selectors) > 0 {
		translator := &BehaveTagTranslator{}
		tagArgs := translator.TranslateTagFilter(cfg.SuiteTagFilter)
		behaveArgs = append(behaveArgs, tagArgs...)
	}

	// Add the specific feature file if provided
	if relFeatureFile != "" {
		featurePath := filepath.Join(cfg.WorkspaceRoot, relFeatureFile)
		relPath, relErr := filepath.Rel(env.WorkDir, featurePath)
		if relErr == nil {
			behaveArgs = append(behaveArgs, relPath)
		}
	}

	fmt.Fprintf(logWriter, "=== Testing Python behave specs ===\n")
	fmt.Fprintf(logWriter, "Module root: %s (isolated: %s)\n", moduleRoot, env.WorkDir)
	fmt.Fprintf(logWriter, "Command: python %s\n\n", strings.Join(behaveArgs, " "))

	// Execute behave using the resolved tool with venv env
	behaveExecCtx := &tool.ExecutionContext{
		ModuleRoot:    env.WorkDir,
		FullEnv:       env.Env,
		ArgsOverrides: behaveArgs,
	}
	return tool.GlobalExecutor().Execute(ctx, pythonTool, behaveExecCtx)
}

// executeContainer runs behave inside a container, skipping venv entirely.
func (r *BehaveRunner) executeContainer(ctx context.Context, pythonTool *tool.ToolDefinition, moduleRoot, relFeatureFile string, logWriter io.Writer, cfg testrunners.RunConfig) (*tool.ExecutionResult, error) {
	fmt.Fprintf(logWriter, "=== Testing Python behave specs (container) ===\n")
	fmt.Fprintf(logWriter, "Module root: %s\n", moduleRoot)

	// Build behave command parts
	behaveArgs := []string{"-m", "behave", "--format", "json"}
	if cfg.OutputDir != "" {
		jsonPath := filepath.Join("/output", "behave.json")
		behaveArgs = append(behaveArgs, "--outfile", jsonPath)
	}

	// Add tag filters
	if len(cfg.SuiteTagFilter.Selectors) > 0 {
		translator := &BehaveTagTranslator{}
		tagArgs := translator.TranslateTagFilter(cfg.SuiteTagFilter)
		behaveArgs = append(behaveArgs, tagArgs...)
	}

	// Add specific feature file if provided
	if relFeatureFile != "" {
		featurePath := filepath.Join(cfg.WorkspaceRoot, relFeatureFile)
		relPath, relErr := filepath.Rel(moduleRoot, featurePath)
		if relErr == nil {
			behaveArgs = append(behaveArgs, relPath)
		}
	}

	// Install deps and run behave in one invocation
	installCmd := "pip install -e '.[dev,test]' --quiet"
	behaveCmd := "python " + strings.Join(behaveArgs, " ")
	shellCmd := installCmd + " && " + behaveCmd

	fmt.Fprintf(logWriter, "Command: sh -c \"%s\"\n\n", shellCmd)

	placeholders := map[string]string{
		"{workspace}": cfg.WorkspaceRoot,
		"{module}":    moduleRoot,
		"{output}":    cfg.OutputDir,
	}

	// Override command to use shell for combined install+test
	containerTool := pythonTool.Clone()
	containerTool.Command = []string{"sh", "-c", shellCmd}

	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: cfg.WorkspaceRoot,
		ModuleRoot:    moduleRoot,
		OutputDir:     cfg.OutputDir,
		Placeholders:  placeholders,
	}
	return tool.GlobalExecutor().Execute(ctx, containerTool, execCtx)
}

// Ensure BehaveRunner implements the interface (compile-time check).
var _ testrunners.TestTypeRunner = (*BehaveRunner)(nil)
