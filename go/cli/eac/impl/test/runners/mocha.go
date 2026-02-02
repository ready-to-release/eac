// mocha.go - Test runner for TypeScript mocha unit tests
//
// MIGRATION NOTE: This runner uses direct npm subprocess execution for test
// execution with JSON output parsing. For container-based testing:
//
// 1. The test bridge path (GetTestFunc) now uses npm-test from
//    tool-config.yml for simple pass/fail test execution.
//
// 2. This runner is still used for detailed test execution with:
//    - JSON reporter output capture
//    - CTRF format conversion for test reporting
//    - Log file management
//
// Full containerization of this runner would require extending the tool executor
// to support structured output capture (JSON parsing from container stdout).
package runners

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/cli/eac/impl/test/internal/ctrf"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/platform"
	"github.com/ready-to-release/eac/go/core/testing"
)

func init() {
	Register(&MochaRunner{})
}

// MochaRunner handles TypeScript mocha unit test execution.
type MochaRunner struct{}

// TestTypes returns the test types this runner handles.
func (r *MochaRunner) TestTypes() []string {
	return []string{"mocha"}
}

// GetTestInfo extracts structured test metadata from a mocha test reference.
func (r *MochaRunner) GetTestInfo(test testing.TestReference, workspaceRoot string, cfg *config.EACConfig) *TestInfo {
	// Calculate relative path from workspace root
	relPath, err := filepath.Rel(workspaceRoot, test.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)
	relDir := filepath.ToSlash(filepath.Dir(relPath))

	info := &TestInfo{Language: "ts"}

	// Find the module this path belongs to
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
		// Check all package roots for a match
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
// Mocha tests are typically in a test/ directory within the module.
// Returns the parent directory of the test directory.
func (r *MochaRunner) FindTestRoot(testPath string, cfg *config.EACConfig) string {
	// Mocha tests don't need special test root finding
	// The test directory path is used directly
	return ""
}

// BuildPackagePath constructs the package path for test grouping.
// For mocha tests, we group by the test directory.
func (r *MochaRunner) BuildPackagePath(testRoot, testPath string) string {
	// For mocha tests, the testRoot is the test directory path itself
	return testRoot
}

// Execute runs TypeScript mocha tests for a package.
// logWriter is provided by the orchestrator (UoW manages log files).
func (r *MochaRunner) Execute(pkgPath string, tests []testing.TestReference, logWriter io.Writer, cfg RunConfig) RunResult {
	start := time.Now()
	result := RunResult{
		PackageName:   pkgPath,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	// pkgPath is the test directory (e.g., "typescript/vscode-commit/test")
	// We need to find the module root (parent of test directory)
	moduleRoot := filepath.Dir(filepath.Join(cfg.WorkspaceRoot, pkgPath))

	// Check if package.json exists
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "No package.json found at %s\n", packageJSON)
		result.PackageFailed = true
		return result
	}

	// Prepare isolated npm environment to avoid Windows EPERM errors
	// and interference between parallel test runs
	isolation := NewNpmIsolation(cfg.WorkspaceRoot)
	env, err := isolation.PrepareIsolatedEnv(moduleRoot, cfg.ModuleMoniker)
	if err != nil {
		fmt.Fprintf(logWriter, "Failed to prepare isolated environment: %v\n", err)
		result.PackageFailed = true
		return result
	}

	fmt.Fprintf(logWriter, "Using isolated environment: %s\n", env.WorkDir)

	// Always run npm ci (or npm install if no package-lock.json)
	// This ensures dependencies are installed with correct platform binaries
	packageLock := filepath.Join(moduleRoot, "package-lock.json")
	var npmCmd string
	if _, err := os.Stat(packageLock); err == nil {
		npmCmd = "ci"
	} else {
		npmCmd = "install"
	}

	fmt.Fprintf(logWriter, "Installing npm dependencies (npm %s)...\n", npmCmd)
	installName, installArgs := platform.WrapCommand("npm", npmCmd)
	installCmd := exec.Command(installName, installArgs...)
	installCmd.Dir = env.WorkDir // Run in isolated directory
	installCmd.Env = env.Env     // Use environment with NPM_CONFIG_CACHE
	installOutput, installErr := installCmd.CombinedOutput()
	fmt.Fprintf(logWriter, "%s\n", installOutput)
	if installErr != nil {
		fmt.Fprintf(logWriter, "npm %s failed: %v\n", npmCmd, installErr)
		result.PackageFailed = true
		return result
	}

	// Build npm test command with JSON reporter for structured output
	args := []string{"test", "--", "--reporter", "json"}

	// Log command
	fmt.Fprintf(logWriter, "=== Testing TypeScript mocha tests ===\n")
	fmt.Fprintf(logWriter, "Module root: %s (isolated: %s)\n", moduleRoot, env.WorkDir)
	fmt.Fprintf(logWriter, "Command: npm %s\n\n", strings.Join(args, " "))

	// Execute npm test in isolated directory
	wrappedName, wrappedArgs := platform.WrapCommand("npm", args...)
	cmd := exec.Command(wrappedName, wrappedArgs...)
	cmd.Dir = env.WorkDir // Run in isolated directory
	cmd.Env = append(env.Env, "R2R_TEST_LOGGING_ACTIVE=true")

	// Capture stdout (JSON) and stderr separately
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
		fmt.Fprintf(logWriter, "Failed to start: %v\n", runErr)
		result.PackageFailed = true
		return result
	}

	// Read stdout (JSON output) - ignore errors as partial data may be useful
	jsonOutput, _ := io.ReadAll(stdout) //nolint:errcheck // partial data still useful
	// Read stderr (error messages)
	stderrOutput, _ := io.ReadAll(stderr) //nolint:errcheck // partial data still useful

	runErr = cmd.Wait()

	// Write stderr to log file
	if len(stderrOutput) > 0 {
		fmt.Fprintf(logWriter, "%s\n", stderrOutput)
	}

	// Convert mocha JSON to CTRF and save (writes to UoW output directory)
	if len(jsonOutput) > 0 && cfg.OutputDir != "" {
		if ctrfReport := convertMochaJSONToCTRF(jsonOutput); ctrfReport != nil {
			if ctrfData, err := ctrfReport.ToJSON(); err == nil {
				jsonPath := filepath.Join(cfg.OutputDir, "unit.json")
				_ = os.WriteFile(jsonPath, ctrfData, 0o644) //nolint:errcheck // best-effort artifact save
				fmt.Fprintf(logWriter, "CTRF JSON saved to unit.json (%d bytes)\n", len(ctrfData))
			}
		}
	}

	// Parse results
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
	result := int64(ms + 0.5) // Round to nearest
	if result == 0 {
		return 1 // Minimum 1ms for non-zero durations
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

	// Add passed tests
	for _, t := range mocha.Passes {
		report.AddTest(ctrf.Test{
			Name:     t.FullTitle,
			Status:   ctrf.StatusPassed,
			Duration: mochaDurationMs(t.Duration),
		})
	}

	// Add failed tests
	for _, t := range mocha.Failures {
		report.AddTest(ctrf.Test{
			Name:     t.FullTitle,
			Status:   ctrf.StatusFailed,
			Duration: mochaDurationMs(t.Duration),
			Message:  t.Err.Message,
			Trace:    t.Err.Stack,
		})
	}

	// Add pending tests
	for _, t := range mocha.Pending {
		report.AddTest(ctrf.Test{
			Name:     t.FullTitle,
			Status:   ctrf.StatusPending,
			Duration: mochaDurationMs(t.Duration),
		})
	}

	// Set actual test times from mocha stats
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
