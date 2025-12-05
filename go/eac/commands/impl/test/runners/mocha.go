// mocha.go - Test runner for TypeScript mocha unit tests
package runners

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/platform"
	"github.com/ready-to-release/eac/go/eac/core/testing"
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
func (r *MochaRunner) BuildPackagePath(testRoot string, testPath string) string {
	// For mocha tests, the testRoot is the test directory path itself
	return testRoot
}

// Execute runs TypeScript mocha tests for a package.
func (r *MochaRunner) Execute(pkgPath string, tests []testing.TestReference, tuiWriter io.Writer, cfg RunConfig) RunResult {
	start := time.Now()
	result := RunResult{PackageName: pkgPath}

	// pkgPath is the test directory (e.g., "typescript/vscode-ext-commit/test")
	// We need to find the module root (parent of test directory)
	moduleRoot := filepath.Dir(filepath.Join(cfg.WorkspaceRoot, pkgPath))

	// Create log directory using module-based output path if available
	outputPath := cfg.ModuleOutputPath
	if outputPath == "" {
		outputPath = sanitizePathForLog(pkgPath)
	}
	logDir := filepath.Join(cfg.TestRunDir, outputPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(tuiWriter, "Failed to create log directory: %v\n", err)
		result.PackageFailed = true
		return result
	}

	// Create log file
	logFilePath := filepath.Join(logDir, "test.log")
	logFile, err := os.Create(logFilePath)
	if err != nil {
		fmt.Fprintf(tuiWriter, "Failed to create log file: %v\n", err)
		result.PackageFailed = true
		return result
	}
	defer logFile.Close()
	result.LogFilePath = logFilePath

	// Check if package.json exists
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		fmt.Fprintf(tuiWriter, "No package.json found\n")
		fmt.Fprintf(logFile, "No package.json found at %s\n", packageJSON)
		result.PackageFailed = true
		return result
	}

	// Build npm test command
	args := []string{"test"}

	// Log command
	fmt.Fprintf(logFile, "=== Testing TypeScript mocha tests ===\n")
	fmt.Fprintf(logFile, "Module root: %s\n", moduleRoot)
	fmt.Fprintf(logFile, "Command: npm %s\n\n", strings.Join(args, " "))

	// Execute npm test
	wrappedName, wrappedArgs := platform.WrapCommand("npm", args...)
	cmd := exec.Command(wrappedName, wrappedArgs...)
	cmd.Dir = moduleRoot
	cmd.Env = os.Environ()

	// Capture output
	output, runErr := cmd.CombinedOutput()
	fmt.Fprintf(logFile, "%s\n", output)

	// Parse results
	if runErr != nil {
		result.PackageFailed = true
		result.TestsFailed = len(tests)
		fmt.Fprintf(tuiWriter, "mocha tests failed\n")
	} else {
		result.TestsPassed = len(tests)
		fmt.Fprintf(tuiWriter, "mocha tests passed\n")
	}

	result.TestsTotal = len(tests)
	result.Duration = time.Since(start)

	return result
}
