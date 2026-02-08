// Package mocha provides a test runner adapter for TypeScript mocha unit tests.
package mocha

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/adapters/npm"
	"github.com/ready-to-release/eac/go/clibase/ctrf"
	"github.com/ready-to-release/eac/go/clibase/testrunners"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	testrunners.Register(&MochaRunner{})
	testrunners.RegisterDescriptor(&testrunners.TestTypeDescriptor{
		TestType:      "mocha",
		IsBDD:         false,
		ComponentType: "typescript",
		MonikerStyle:  "file",
	})
}

// MochaRunner handles TypeScript mocha unit test execution.
type MochaRunner struct{}

// TestTypes returns the test types this runner handles.
func (r *MochaRunner) TestTypes() []string {
	return []string{"mocha"}
}

// IsBDD returns false because mocha is a unit test framework.
func (r *MochaRunner) IsBDD() bool {
	return false
}

// GetTestInfo extracts structured test metadata from a mocha test reference.
func (r *MochaRunner) GetTestInfo(test testing.TestReference, workspaceRoot string, cfg *config.EACConfig) *testrunners.TestInfo {
	relPath, err := filepath.Rel(workspaceRoot, test.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)
	relDir := filepath.ToSlash(filepath.Dir(relPath))

	info := &testrunners.TestInfo{Language: "ts"}

	info.ModuleMoniker = findTsModuleForPath(relDir, cfg)
	if info.ModuleMoniker == "" {
		return nil
	}

	info.TestRoot = relDir
	info.PackageKey = relDir
	info.DisplayName = relDir

	return info
}

// findTsModuleForPath finds the module moniker for a given TypeScript path.
func findTsModuleForPath(relPath string, cfg *config.EACConfig) string {
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

// FindTestRoot finds the module root for a mocha test file.
func (r *MochaRunner) FindTestRoot(testPath string, cfg *config.EACConfig) string {
	return ""
}

// BuildPackagePath constructs the package path for test grouping.
func (r *MochaRunner) BuildPackagePath(testRoot, testPath string) string {
	return testRoot
}

// Execute runs TypeScript mocha tests for a package.
func (r *MochaRunner) Execute(pkgPath string, tests []testing.TestReference, logWriter io.Writer, cfg testrunners.RunConfig) testrunners.RunResult {
	start := time.Now()
	result := testrunners.RunResult{
		PackageName:   pkgPath,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	moduleRoot := filepath.Dir(filepath.Join(cfg.WorkspaceRoot, pkgPath))

	// Check if package.json exists
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "No package.json found at %s\n", packageJSON)
		result.PackageFailed = true
		return result
	}

	// Prepare isolated npm environment
	isolation := npm.NewNpmIsolation(cfg.WorkspaceRoot)
	env, err := isolation.PrepareIsolatedEnv(moduleRoot, cfg.ModuleMoniker)
	if err != nil {
		fmt.Fprintf(logWriter, "Failed to prepare isolated environment: %v\n", err)
		result.PackageFailed = true
		return result
	}

	fmt.Fprintf(logWriter, "Using isolated environment: %s\n", env.WorkDir)

	// Run npm ci/install
	packageLock := filepath.Join(moduleRoot, "package-lock.json")
	var npmCmd string
	if _, err := os.Stat(packageLock); err == nil {
		npmCmd = "ci"
	} else {
		npmCmd = "install"
	}

	fmt.Fprintf(logWriter, "Installing npm dependencies (npm %s)...\n", npmCmd)
	npm.NpmInstallMu.Lock()
	installToolDef := tool.GlobalRegistry().GetOrAdhoc("npm")
	installExecCtx := &tool.ExecutionContext{
		ModuleRoot:    env.WorkDir,
		FullEnv:       env.Env,
		ArgsOverrides: []string{npmCmd},
	}
	installResult, installErr := tool.GlobalExecutor().Execute(context.Background(), installToolDef, installExecCtx)
	npm.NpmInstallMu.Unlock()
	if installResult != nil {
		fmt.Fprintf(logWriter, "%s%s\n", installResult.Stdout, installResult.Stderr)
	}
	if installErr != nil || (installResult != nil && installResult.ExitCode != 0) {
		fmt.Fprintf(logWriter, "npm %s failed: %v\n", npmCmd, installErr)
		result.PackageFailed = true
		return result
	}

	// Build npm test command with JSON reporter
	args := []string{"test", "--", "--reporter", "json"}

	fmt.Fprintf(logWriter, "=== Testing TypeScript mocha tests ===\n")
	fmt.Fprintf(logWriter, "Module root: %s (isolated: %s)\n", moduleRoot, env.WorkDir)
	fmt.Fprintf(logWriter, "Command: npm %s\n\n", strings.Join(args, " "))

	// Use tool.BuildCommand for pipe-based stdout/stderr streaming
	npmTool := tool.GlobalRegistry().GetOrAdhoc("npm")
	cmd := tool.BuildCommand(context.Background(), npmTool, &tool.ExecutionContext{
		ModuleRoot:    env.WorkDir,
		FullEnv:       append(env.Env, "CLIE_TEST_LOGGING_ACTIVE=true"),
		ArgsOverrides: args,
	})

	stdout, pipeErr := cmd.StdoutPipe()
	if pipeErr != nil {
		fmt.Fprintf(logWriter, "Failed to create stdout pipe: %v\n", pipeErr)
		result.PackageFailed = true
		return result
	}
	stderr, pipeErr := cmd.StderrPipe()
	if pipeErr != nil {
		fmt.Fprintf(logWriter, "Failed to create stderr pipe: %v\n", pipeErr)
		result.PackageFailed = true
		return result
	}

	runErr := cmd.Start()
	if runErr != nil {
		fmt.Fprintf(logWriter, "Failed to start mocha: %v\n", runErr)
		result.PackageFailed = true
		return result
	}

	jsonOutput, _ := io.ReadAll(stdout)  //nolint:errcheck // partial data still useful
	stderrOutput, _ := io.ReadAll(stderr) //nolint:errcheck // partial data still useful

	runErr = cmd.Wait()

	if len(stderrOutput) > 0 {
		fmt.Fprintf(logWriter, "%s\n", stderrOutput)
	}

	// Convert mocha JSON to CTRF and save
	if len(jsonOutput) > 0 && cfg.OutputDir != "" {
		if ctrfReport := convertMochaJSONToCTRF(jsonOutput); ctrfReport != nil {
			if ctrfData, err := ctrfReport.ToJSON(); err == nil {
				jsonPath := filepath.Join(cfg.OutputDir, "unit.json")
				_ = os.WriteFile(jsonPath, ctrfData, 0o644) //nolint:errcheck // best-effort artifact save
				fmt.Fprintf(logWriter, "CTRF JSON saved to unit.json (%d bytes)\n", len(ctrfData))
			}
		}
	}

	if runErr != nil {
		result.PackageFailed = true
		result.TestsFailed = len(tests)
		fmt.Fprintf(logWriter, "mocha tests failed\n")
	} else {
		result.TestsPassed = len(tests)
		fmt.Fprintf(logWriter, "mocha tests passed\n")
	}

	result.TestsTotal = len(tests)
	result.Duration = time.Since(start)

	return result
}

// mochaReport represents mocha's native JSON output format.
type mochaReport struct {
	Stats struct {
		Suites   int    `json:"suites"`
		Tests    int    `json:"tests"`
		Passes   int    `json:"passes"`
		Pending  int    `json:"pending"`
		Failures int    `json:"failures"`
		Start    string `json:"start"`
		End      string `json:"end"`
		Duration int    `json:"duration"`
	} `json:"stats"`
	Tests    []mochaTest `json:"tests"`
	Passes   []mochaTest `json:"passes"`
	Failures []mochaTest `json:"failures"`
	Pending  []mochaTest `json:"pending"`
}

type mochaTest struct {
	Title        string   `json:"title"`
	FullTitle    string   `json:"fullTitle"`
	Duration     float64  `json:"duration"`
	CurrentRetry int      `json:"currentRetry"`
	Err          mochaErr `json:"err"`
}

type mochaErr struct {
	Message string `json:"message"`
	Stack   string `json:"stack"`
}

// mochaDurationMs converts mocha duration (already in ms) ensuring minimum 1ms for non-zero.
func mochaDurationMs(ms float64) int64 {
	if ms <= 0 {
		return 0
	}
	result := int64(ms + 0.5)
	if result == 0 {
		return 1
	}
	return result
}

// convertMochaJSONToCTRF converts mocha's native JSON output to CTRF format.
func convertMochaJSONToCTRF(jsonData []byte) *ctrf.Report {
	var mocha mochaReport
	if err := json.Unmarshal(jsonData, &mocha); err != nil {
		return nil
	}

	report := ctrf.NewReport("mocha")

	for _, t := range mocha.Passes {
		report.AddTest(ctrf.Test{
			Name:     t.FullTitle,
			Status:   ctrf.StatusPassed,
			Duration: mochaDurationMs(t.Duration),
		})
	}

	for _, t := range mocha.Failures {
		report.AddTest(ctrf.Test{
			Name:     t.FullTitle,
			Status:   ctrf.StatusFailed,
			Duration: mochaDurationMs(t.Duration),
			Message:  t.Err.Message,
			Trace:    t.Err.Stack,
		})
	}

	for _, t := range mocha.Pending {
		report.AddTest(ctrf.Test{
			Name:     t.FullTitle,
			Status:   ctrf.StatusPending,
			Duration: mochaDurationMs(t.Duration),
		})
	}

	if mocha.Stats.Start != "" && mocha.Stats.End != "" {
		if startTime, err := time.Parse(time.RFC3339, mocha.Stats.Start); err == nil {
			if endTime, err := time.Parse(time.RFC3339, mocha.Stats.End); err == nil {
				report.SetTimes(startTime.UnixMilli(), endTime.UnixMilli())
			}
		}
	} else {
		report.Finalize()
	}

	return report
}
