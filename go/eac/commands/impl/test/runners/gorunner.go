// gorunner.go - Test runner for Go tests (gotest and godog)
package runners

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/impl/test/internal/runner"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

func init() {
	goRunner := &GoRunner{}
	Register(goRunner)
	RegisterFallback(goRunner) // Go runner is the fallback for unknown types
}

// GoRunner handles Go test execution for both unit tests (gotest) and BDD tests (godog).
type GoRunner struct{}

// TestTypes returns the test types this runner handles.
func (r *GoRunner) TestTypes() []string {
	return []string{"gotest", "godog"}
}

// FindTestRoot finds the test runner package for a godog feature file.
// For gotest, returns empty string (tests are in the same directory as source).
// For godog, returns the path to the directory containing godog_test.go.
func (r *GoRunner) FindTestRoot(featurePath string, cfg *config.EACConfig) string {
	// gotest doesn't need a separate test root
	if !strings.HasSuffix(featurePath, ".feature") {
		return ""
	}

	// Extract relative path from specs root
	specsPrefix := cfg.Repository.Paths.SpecsRoot + "/"
	relPath := strings.TrimPrefix(filepath.ToSlash(featurePath), specsPrefix)
	relPath = strings.TrimPrefix(relPath, strings.ReplaceAll(specsPrefix, "/", "\\"))
	relPath = filepath.ToSlash(relPath)

	// Get path components
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		return ""
	}

	// Try progressively deeper paths to find godog_test.go
	moniker := parts[0]
	basePath := cfg.Repository.TestImplPath(moniker)

	// Check if godog_test.go exists at base path
	workspaceRoot := cfg.RepoRoot
	if fileExists(filepath.Join(workspaceRoot, basePath, "godog_test.go")) {
		return basePath
	}

	// Try adding subdirectories (skip the filename at the end)
	for i := 1; i < len(parts)-1; i++ {
		subPath := filepath.Join(basePath, strings.Join(parts[1:i+1], "/"))
		subPath = filepath.ToSlash(subPath)
		if fileExists(filepath.Join(workspaceRoot, subPath, "godog_test.go")) {
			return subPath
		}
	}

	// No test runner found
	return ""
}

// BuildPackagePath constructs the package path for test grouping.
// For godog BDD tests, returns "featureFolderName:testRoot" for cleaner display.
// For gotest, returns the directory path.
func (r *GoRunner) BuildPackagePath(testRoot string, featurePath string) string {
	if featurePath != "" && testRoot != "" {
		// Extract feature folder name from path like:
		// "specs/repository/no-build-tags-in-steps/specification.feature"
		// -> "no-build-tags-in-steps"
		featureFolderName := extractFeatureFolderName(featurePath)
		// Store full feature path after second colon for Execute() to use
		return featureFolderName + ":" + testRoot + ":" + featurePath
	}
	return testRoot
}

// extractFeatureFolderName extracts the feature folder name from a feature path.
// Input: "specs/repository/no-build-tags-in-steps/specification.feature"
// Output: "no-build-tags-in-steps"
func extractFeatureFolderName(featurePath string) string {
	featurePath = filepath.ToSlash(featurePath)
	// Remove the filename (specification.feature)
	dir := filepath.Dir(featurePath)
	// Get the last directory component (feature folder name)
	return filepath.Base(dir)
}

// Execute runs Go tests for a package and returns results.
func (r *GoRunner) Execute(pkgPath string, tests []testing.TestReference, tuiWriter io.Writer, cfg RunConfig) RunResult {
	start := time.Now()

	// Parse package path - new format: "featureName:testRoot:featurePath" or "testRoot"
	var displayName, relPkgPath, relFeatureFile string
	parts := strings.Split(pkgPath, ":")
	if len(parts) == 3 {
		// BDD format: featureName:testRoot:featurePath
		// Display as "featureName:testRoot" (without full feature path)
		displayName = parts[0] + ":" + parts[1]
		relPkgPath = parts[1]    // testRoot
		relFeatureFile = parts[2] // full feature path
	} else if len(parts) == 1 {
		// Unit test format: just the package path
		displayName = parts[0]
		relPkgPath = parts[0]
	} else {
		// Fallback for unexpected format
		displayName = pkgPath
		relPkgPath = pkgPath
	}

	result := RunResult{PackageName: displayName}

	actualPkgDir := filepath.Join(cfg.WorkspaceRoot, relPkgPath)

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

	// Create streaming test runner
	streamingRunner := runner.NewStreamingRunner(tuiWriter, logFile)

	// Run go generate to ensure embedded files exist (e.g., from contracts)
	// This is needed because test jobs may run on fresh checkouts without build artifacts
	if err := runGoGenerate(actualPkgDir, logFile); err != nil {
		fmt.Fprintf(logFile, "Warning: go generate failed: %v\n", err)
		// Don't fail - go generate might not be needed for all packages
	}

	// Build go test command
	goTestArgs := []string{"test", "-json", "-v", "-parallel", fmt.Sprintf("%d", cfg.Parallelism)}

	// Extract Go build tags from suite filter (e.g., "@L0,@L1" -> "L0,L1")
	// This ensures test files with //go:build constraints are compiled
	if buildTags := extractGoBuildTags(cfg.SuiteTagFilter); buildTags != "" {
		goTestArgs = append(goTestArgs, "-tags", buildTags)
	}

	// Add coverage if enabled
	if cfg.Coverage {
		coverageFile := filepath.Join(logDir, "coverage.out")
		goTestArgs = append(goTestArgs, "-cover", "-coverprofile="+coverageFile)
	}

	// Add package path
	goTestArgs = append(goTestArgs, ".")

	cmd := exec.Command("go", goTestArgs...)
	cmd.Dir = actualPkgDir
	cmd.Env = os.Environ()

	// Set test run ID for nested commands
	testRunID := filepath.Base(cfg.TestRunDir)
	cmd.Env = append(cmd.Env, fmt.Sprintf("R2R_TEST_RUN_ID=%s", testRunID))

	// Set godog environment variables if this is a godog test
	isGodogTest := fileExists(filepath.Join(actualPkgDir, "godog_test.go"))
	if isGodogTest {
		cmd.Env = append(cmd.Env, "GODOG_FORMAT=progress")
		if cfg.SuiteTagFilter != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_SUITE_TAGS=%s", cfg.SuiteTagFilter))
		}
		if relFeatureFile != "" {
			relFeaturePath, _ := filepath.Rel(actualPkgDir, filepath.Join(cfg.WorkspaceRoot, relFeatureFile))
			relFeaturePath = filepath.ToSlash(relFeaturePath)
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_PATHS=%s", relFeaturePath))

			// Set report output for feature files
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_OUTPUT_DIR=%s", logDir))
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_REPORT_FORMAT=%s", cfg.ReportFormat))
		}
	}

	// Run tests with streaming output
	testResult, runErr := streamingRunner.Run(cmd)

	result.TestsPassed = testResult.TestsPassed
	result.TestsFailed = testResult.TestsFailed
	result.TestsSkipped = testResult.TestsSkipped
	result.TestsTotal = testResult.TestsTotal
	result.PackageFailed = testResult.PackageFailed || runErr != nil
	result.Duration = time.Since(start)

	return result
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// runGoGenerate runs go generate for a package directory.
// This ensures embedded files from contracts are available for testing.
func runGoGenerate(pkgDir string, logWriter io.Writer) error {
	// Find the module root by walking up to find go.mod
	moduleRoot := findModuleRoot(pkgDir)
	if moduleRoot == "" {
		return nil // No go.mod found, skip
	}

	cmd := exec.Command("go", "generate", "./...")
	cmd.Dir = moduleRoot
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Fprintf(logWriter, "go generate output:\n%s\n", string(output))
	}
	return err
}

// findModuleRoot walks up from dir to find the directory containing go.mod
func findModuleRoot(dir string) string {
	for {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "" // Reached root
		}
		dir = parent
	}
}

// sanitizePathForLog converts a package path to a safe directory name
func sanitizePathForLog(pkgPath string) string {
	safe := strings.ReplaceAll(pkgPath, ":", "_")
	safe = strings.ReplaceAll(safe, "\\", "/")
	return safe
}

// extractGoBuildTags extracts Go build tags from a suite tag filter.
// Input: "@L0,@L1 && ~@skip:wip" or "@L0,@L1,@L2"
// Output: "L0,L1" or "L0,L1,L2" (comma-separated Go build tags)
func extractGoBuildTags(suiteTagFilter string) string {
	if suiteTagFilter == "" {
		return ""
	}

	var tags []string

	// Look for L-level tags (@L0, @L1, @L2, @L3, @L4)
	for _, level := range []string{"L0", "L1", "L2", "L3", "L4"} {
		if strings.Contains(suiteTagFilter, "@"+level) {
			tags = append(tags, level)
		}
	}

	if len(tags) == 0 {
		return ""
	}

	return strings.Join(tags, ",")
}
