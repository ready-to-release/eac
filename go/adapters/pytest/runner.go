// Package pytest provides a test runner adapter for Python pytest unit tests.
package pytest

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

// RegisterWith registers the PytestRunner and its descriptor with the given registry.
// This enables testing with an isolated registry instead of relying on global state.
func RegisterWith(reg *testrunners.Registry) {
	reg.Register(&PytestRunner{})
	reg.RegisterDescriptor(&testrunners.TestTypeDescriptor{
		TestType:      "pytest",
		IsBDD:         false,
		ComponentType: "python",
		MonikerStyle:  "file",
		DefaultInferences: []testrunners.Inference{
			{
				TestTypes:   []string{"pytest"},
				ThenAddTags: []string{"@deps:python"},
				Description: "Python tests require Python toolchain",
			},
			{
				TestTypes:   []string{"pytest"},
				ThenAddTags: []string{"@L1"},
				Description: "Pytest unit tests are L1 by default",
			},
		},
	})
}

// PytestRunner handles Python pytest unit test execution.
type PytestRunner struct{}

// TestTypes returns the test types this runner handles.
func (r *PytestRunner) TestTypes() []string {
	return []string{"pytest"}
}

// IsBDD returns false because pytest is a unit test framework.
func (r *PytestRunner) IsBDD() bool {
	return false
}

// GetTestInfo extracts structured test metadata from a pytest test reference.
func (r *PytestRunner) GetTestInfo(ref testing.TestReference, workspaceRoot string, cfg any) *testrunners.TestInfo {
	eacCfg := cfg.(*config.EACConfig)

	relPath, err := filepath.Rel(workspaceRoot, ref.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)
	relDir := filepath.ToSlash(filepath.Dir(relPath))

	info := &testrunners.TestInfo{Language: "python"}

	info.ModuleMoniker = findPythonModuleForPath(relDir, eacCfg)
	if info.ModuleMoniker == "" {
		return nil
	}

	info.TestRoot = relDir
	info.PackageKey = relDir
	info.DisplayName = relDir

	return info
}

// findPythonModuleForPath finds the module moniker for a given Python path.
func findPythonModuleForPath(relPath string, cfg *config.EACConfig) string {
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

// FindTestRoot finds the module root for a pytest test file.
func (r *PytestRunner) FindTestRoot(testPath string, cfg any) string {
	return ""
}

// BuildPackagePath constructs the package path for test grouping.
func (r *PytestRunner) BuildPackagePath(testRoot, testPath string) string {
	return testRoot
}

// collectPytestResults parses pytest JSON output and populates the result with CTRF data.
// Returns true if results were successfully parsed.
func collectPytestResults(result *testrunners.RunResult, outputDir string, logWriter io.Writer) bool {
	if outputDir == "" {
		return false
	}
	jsonPath := filepath.Join(outputDir, "results.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return false
	}
	ctrfReport := convertPytestJSONToCTRF(data)
	if ctrfReport == nil {
		return false
	}
	result.TestsPassed = ctrfReport.Results.Summary.Passed
	result.TestsFailed = ctrfReport.Results.Summary.Failed
	result.TestsSkipped = ctrfReport.Results.Summary.Skipped
	result.TestsTotal = ctrfReport.Results.Summary.Tests

	if ctrfData, err := ctrfReport.ToJSON(); err == nil {
		ctrfPath := filepath.Join(outputDir, "unit.json")
		_ = os.WriteFile(ctrfPath, ctrfData, 0o644)
		fmt.Fprintf(logWriter, "CTRF JSON saved to unit.json (%d bytes)\n", len(ctrfData))
	}
	return true
}

// pytestExecutionFailed returns true if the execution result indicates failure.
func pytestExecutionFailed(execResult *tool.ExecutionResult, runErr error) bool {
	return runErr != nil || (execResult != nil && execResult.ExitCode != 0)
}

// Execute runs Python pytest tests for a package.
func (r *PytestRunner) Execute(pkgPath string, tests []testing.TestReference, logWriter io.Writer, cfg testrunners.RunConfig) testrunners.RunResult {
	start := time.Now()
	result := testrunners.RunResult{
		PackageName:   pkgPath,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	moduleRoot := filepath.Dir(filepath.Join(cfg.WorkspaceRoot, pkgPath))

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
		execResult, runErr = r.executeContainer(runCtx, pythonTool, moduleRoot, tests, logWriter, cfg)
	} else {
		execResult, runErr = r.executeSystem(runCtx, pythonTool, moduleRoot, tests, logWriter, cfg)
	}

	if execResult != nil {
		fmt.Fprintf(logWriter, "%s%s\n", execResult.Stdout, execResult.Stderr)
	}

	// Parse JSON results; if unavailable, report 1 passed or 1 failed as conservative fallback.
	// len(tests) is the test-file count, not the individual test count, so using it
	// would inflate metrics. Real counts come from pytest JSON when available.
	if !collectPytestResults(&result, cfg.OutputDir, logWriter) {
		if pytestExecutionFailed(execResult, runErr) {
			result.TestsTotal = 1
			result.TestsFailed = 1
		} else {
			result.TestsTotal = 1
			result.TestsPassed = 1
		}
	}

	if pytestExecutionFailed(execResult, runErr) {
		result.PackageFailed = true
		fmt.Fprintf(logWriter, "pytest tests failed\n")
	} else {
		fmt.Fprintf(logWriter, "pytest tests passed\n")
	}

	result.Duration = time.Since(start)
	return result
}

// executeSystem runs pytest using a local system Python with venv isolation.
func (r *PytestRunner) executeSystem(ctx context.Context, pythonTool *tool.ToolDefinition, moduleRoot string, _ []testing.TestReference, logWriter io.Writer, cfg testrunners.RunConfig) (*tool.ExecutionResult, error) {
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

	// Build pytest command with JSON report
	pytestArgs := []string{"-m", "pytest", "--tb=short", "-q"}
	if cfg.OutputDir != "" {
		junitPath := filepath.Join(cfg.OutputDir, "results.xml")
		pytestArgs = append(pytestArgs, "--junitxml="+junitPath)
		jsonPath := filepath.Join(cfg.OutputDir, "results.json")
		pytestArgs = append(pytestArgs, "--json-report", "--json-report-file="+jsonPath)
	}

	fmt.Fprintf(logWriter, "=== Testing Python pytest tests ===\n")
	fmt.Fprintf(logWriter, "Module root: %s (isolated: %s)\n", moduleRoot, env.WorkDir)
	fmt.Fprintf(logWriter, "Command: python %s\n\n", strings.Join(pytestArgs, " "))

	// Execute pytest using the resolved tool with venv env (PATH ensures venv python)
	pytestExecCtx := &tool.ExecutionContext{
		ModuleRoot:    env.WorkDir,
		FullEnv:       env.Env,
		ArgsOverrides: pytestArgs,
	}
	return tool.GlobalExecutor().Execute(ctx, pythonTool, pytestExecCtx)
}

// executeContainer runs pytest inside a container, skipping venv entirely.
func (r *PytestRunner) executeContainer(ctx context.Context, pythonTool *tool.ToolDefinition, moduleRoot string, _ []testing.TestReference, logWriter io.Writer, cfg testrunners.RunConfig) (*tool.ExecutionResult, error) {
	fmt.Fprintf(logWriter, "=== Testing Python pytest tests (container) ===\n")
	fmt.Fprintf(logWriter, "Module root: %s\n", moduleRoot)

	// Build a single shell command: install deps + run pytest
	pytestArgs := []string{"-m", "pytest", "--tb=short", "-q"}
	if cfg.OutputDir != "" {
		junitPath := filepath.Join("/output", "results.xml")
		pytestArgs = append(pytestArgs, "--junitxml="+junitPath)
		jsonPath := filepath.Join("/output", "results.json")
		pytestArgs = append(pytestArgs, "--json-report", "--json-report-file="+jsonPath)
	}

	// Install deps and run pytest in one invocation
	installCmd := "pip install -e '.[dev,test]' --quiet"
	pytestCmd := "python " + strings.Join(pytestArgs, " ")
	shellCmd := installCmd + " && " + pytestCmd

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

// Ensure PytestRunner implements the interface (compile-time check).
var _ testrunners.TestTypeRunner = (*PytestRunner)(nil)
