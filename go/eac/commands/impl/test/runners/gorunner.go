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

	"github.com/ready-to-release/eac/go/eac/commands/impl/test/internal/ctrf"
	"github.com/ready-to-release/eac/go/eac/commands/impl/test/internal/runner"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

var goRunnerLog = logging.C()

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

// GetTestInfo extracts structured test metadata from a Go test reference.
func (r *GoRunner) GetTestInfo(test testing.TestReference, workspaceRoot string, cfg *config.EACConfig) *TestInfo {
	// Calculate relative path from workspace root
	relPath, err := filepath.Rel(workspaceRoot, test.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)

	info := &TestInfo{Language: "go"}

	if test.Type == "godog" {
		// BDD test: extract from specs path
		specsPrefix := cfg.Repository.Paths.SpecsRoot + "/"
		specRelPath := strings.TrimPrefix(relPath, specsPrefix)
		specRelPath = filepath.ToSlash(specRelPath)

		// Get module moniker from first path component
		parts := strings.Split(specRelPath, "/")
		if len(parts) == 0 {
			return nil
		}
		info.ModuleMoniker = parts[0]

		// Verify module exists
		if cfg.Repository.GetByMoniker(info.ModuleMoniker) == nil {
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
	} else {
		// Unit test (gotest): extract module from path using module_mapping
		absDir := filepath.Dir(test.FilePath)
		relDir, err := filepath.Rel(workspaceRoot, absDir)
		if err != nil {
			return nil
		}
		relDir = filepath.ToSlash(relDir)

		// Find the module this path belongs to
		info.ModuleMoniker = findModuleForPath(relDir, cfg)
		if info.ModuleMoniker == "" {
			return nil
		}

		info.TestRoot = relDir
		info.PackageKey = relDir
		info.DisplayName = relDir
	}

	return info
}

// findModuleForPath finds the module moniker for a given relative path.
func findModuleForPath(relPath string, cfg *config.EACConfig) string {
	// Iterate through modules to find the one that owns this path
	for i := range cfg.Repository.Modules {
		module := &cfg.Repository.Modules[i]
		// Check all package roots
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

// FindTestRoot finds the test runner package for a godog feature file.
// For gotest, returns empty string (tests are in the same directory as source).
// For godog, returns the path to the directory containing godog_test.go.
// Returns empty string if module not found or no godog_test.go exists - caller must handle.
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
		goRunnerLog.Debugf("FindTestRoot: empty path parts for %s", featurePath)
		return ""
	}

	// Extract moniker from first path component
	moniker := parts[0]

	// Verify module exists - fail early if not
	if cfg.Repository.GetByMoniker(moniker) == nil {
		goRunnerLog.Debugf("FindTestRoot: unknown module %s for %s", moniker, featurePath)
		return "" // Unknown module, no fallback guessing
	}

	// Get test impl path from module contract
	basePath := cfg.Repository.TestImplPath(moniker)
	if basePath == "" {
		goRunnerLog.Debugf("FindTestRoot: no test-impl path for module %s", moniker)
		return ""
	}

	// Check if godog_test.go exists at base path
	workspaceRoot := cfg.RepoRoot
	baseCheck := filepath.Join(workspaceRoot, basePath, "godog_test.go")
	if fileExists(baseCheck) {
		goRunnerLog.Debugf("FindTestRoot: found godog_test.go at basePath %s", basePath)
		return basePath
	}

	// Try adding subdirectories (skip the filename at the end)
	for i := 1; i < len(parts)-1; i++ {
		subPath := filepath.Join(basePath, strings.Join(parts[1:i+1], "/"))
		subPath = filepath.ToSlash(subPath)
		subCheck := filepath.Join(workspaceRoot, subPath, "godog_test.go")
		if fileExists(subCheck) {
			goRunnerLog.Debugf("FindTestRoot: found godog_test.go at subPath %s", subPath)
			return subPath
		}
	}

	// No test runner found - no fallback
	goRunnerLog.Debugf("FindTestRoot: no godog_test.go found for %s (basePath=%s, parts=%v)", featurePath, basePath, parts)
	return ""
}

// BuildPackagePath constructs the package path for test grouping.
// For godog BDD tests, returns "featureFolderName:testRoot" for cleaner display.
// For gotest, returns the directory path.
func (r *GoRunner) BuildPackagePath(testRoot, featurePath string) string {
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
// Output: "no-build-tags-in-steps".
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
		relPkgPath = parts[1]     // testRoot
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

	result := RunResult{
		PackageName:   displayName,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	actualPkgDir := filepath.Join(cfg.WorkspaceRoot, relPkgPath)

	// Use pre-created OutputDir if set, otherwise create based on module path
	var logDir string
	if cfg.OutputDir != "" {
		logDir = cfg.OutputDir
	} else {
		outputPath := cfg.ModuleOutputPath
		if outputPath == "" {
			outputPath = pkgPath
		}
		logDir = filepath.Join(cfg.TestRunDir, sanitizePathForLog(outputPath))
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
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

	// Disable file logging in test subprocesses to prevent polluting out/commands.log
	cmd.Env = append(cmd.Env, "R2R_TEST_LOGGING_ACTIVE=true")

	// Set godog environment variables if this is a godog test
	isGodogTest := fileExists(filepath.Join(actualPkgDir, "godog_test.go"))
	if isGodogTest {
		cmd.Env = append(cmd.Env, "GODOG_FORMAT=progress")
		if cfg.SuiteTagFilter != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_SUITE_TAGS=%s", cfg.SuiteTagFilter))
		}

		// Always set report output for godog tests (cucumber.json format)
		cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_OUTPUT_DIR=%s", logDir))

		if relFeatureFile != "" {
			relFeaturePath, relErr := filepath.Rel(actualPkgDir, filepath.Join(cfg.WorkspaceRoot, relFeatureFile))
			if relErr == nil {
				relFeaturePath = filepath.ToSlash(relFeaturePath)
				cmd.Env = append(cmd.Env, fmt.Sprintf("GODOG_PATHS=%s", relFeaturePath))
			}
		}
	}

	// Run tests with streaming output
	testResult, runErr := streamingRunner.Run(cmd)

	// Save CTRF JSON output for unit tests (non-godog)
	// Godog tests save cucumber.json via GODOG_OUTPUT_DIR
	if !isGodogTest {
		events := streamingRunner.GetEvents()
		if len(events) > 0 {
			report := convertGoTestEventsToCTRF(events)
			if ctrfData, err := report.ToJSON(); err == nil {
				jsonPath := filepath.Join(logDir, "unit.json")
				_ = os.WriteFile(jsonPath, ctrfData, 0o644) //nolint:errcheck // best-effort artifact save
			}
		}
	}

	result.TestsPassed = testResult.TestsPassed
	result.TestsFailed = testResult.TestsFailed
	result.TestsSkipped = testResult.TestsSkipped
	result.TestsTotal = testResult.TestsTotal
	result.PackageFailed = testResult.PackageFailed || runErr != nil
	result.Duration = time.Since(start)

	return result
}

// fileExists checks if a file exists.
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
	cmd.Env = append(cmd.Env, "R2R_TEST_LOGGING_ACTIVE=true")

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Fprintf(logWriter, "go generate output:\n%s\n", string(output))
	}
	return err
}

// findModuleRoot walks up from dir to find the directory containing go.mod.
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

// sanitizePathForLog converts a package path to a safe directory name.
func sanitizePathForLog(pkgPath string) string {
	safe := strings.ReplaceAll(pkgPath, ":", "_")
	safe = strings.ReplaceAll(safe, "\\", "/")
	return safe
}

// extractGoBuildTags extracts Go build tags from a suite tag filter.
// Input: "@L0,@L1 && ~@skip:wip" or "@L0,@L1,@L2" or "@deps:gh-token"
// Output: "L0,L1" or "L0,L1,L2" or "L0,L1,deps_gh_token" (comma-separated Go build tags)
// Note: @deps:<name> tags are translated to deps_<name> (colon->underscore, hyphen->underscore).
func extractGoBuildTags(suiteTagFilter string) string {
	if suiteTagFilter == "" {
		return ""
	}

	var tags []string

	// Look for L-level tags (@L0, @L1, @L2, @L3, @L4)
	// Only include tags that are NOT negated (preceded by ~)
	for _, level := range []string{"L0", "L1", "L2", "L3", "L4"} {
		tag := "@" + level
		idx := 0
		for {
			pos := strings.Index(suiteTagFilter[idx:], tag)
			if pos == -1 {
				break
			}
			absPos := idx + pos
			// Check if this occurrence is negated (preceded by ~)
			if absPos == 0 || suiteTagFilter[absPos-1] != '~' {
				tags = append(tags, level)
				break
			}
			idx = absPos + len(tag)
		}
	}

	// Look for @deps:<name> tags and translate to deps_<name>
	// Go build tags can't contain : or -, so we replace them with _
	depsPrefix := "@deps:"
	idx := 0
	for {
		pos := strings.Index(suiteTagFilter[idx:], depsPrefix)
		if pos == -1 {
			break
		}
		start := idx + pos + len(depsPrefix)
		// Find the end of the dependency name (stop at space, comma, &, or end)
		end := start
		for end < len(suiteTagFilter) {
			c := suiteTagFilter[end]
			if c == ' ' || c == ',' || c == '&' || c == ')' {
				break
			}
			end++
		}
		if end > start {
			depName := suiteTagFilter[start:end]
			// Translate to valid Go build tag: replace - with _
			goBuildTag := "deps_" + strings.ReplaceAll(depName, "-", "_")
			tags = append(tags, goBuildTag)
		}
		idx = end
	}

	if len(tags) == 0 {
		return ""
	}

	return strings.Join(tags, ",")
}

// durationMs converts seconds to milliseconds, ensuring minimum 1ms for non-zero durations.
func durationMs(seconds float64) int64 {
	if seconds <= 0 {
		return 0
	}
	ms := int64(seconds * 1000)
	if ms == 0 {
		return 1 // Round up sub-millisecond to 1ms
	}
	return ms
}

// convertGoTestEventsToCTRF converts go test -json events to CTRF format.
func convertGoTestEventsToCTRF(events []runner.TestEvent) *ctrf.Report {
	report := ctrf.NewReport("go-test")

	// Track test state: map test name -> output lines and elapsed time
	type testState struct {
		output  []string
		elapsed float64
	}
	tests := make(map[string]*testState)

	// Track start/stop times from events
	var startTime, stopTime time.Time
	for _, event := range events {
		if event.Time != "" {
			if t, err := time.Parse(time.RFC3339Nano, event.Time); err == nil {
				if startTime.IsZero() || t.Before(startTime) {
					startTime = t
				}
				if t.After(stopTime) {
					stopTime = t
				}
			}
		}

		// Skip package-level events (no test name)
		if event.Test == "" {
			continue
		}

		// Initialize test state if needed
		if tests[event.Test] == nil {
			tests[event.Test] = &testState{}
		}

		switch event.Action {
		case "output":
			tests[event.Test].output = append(tests[event.Test].output, event.Output)
		case "pass":
			tests[event.Test].elapsed = event.Elapsed
			report.AddTest(ctrf.Test{
				Name:     event.Test,
				Status:   ctrf.StatusPassed,
				Duration: durationMs(event.Elapsed),
				Suite:    event.Package,
			})
		case "fail":
			tests[event.Test].elapsed = event.Elapsed
			report.AddTest(ctrf.Test{
				Name:     event.Test,
				Status:   ctrf.StatusFailed,
				Duration: durationMs(event.Elapsed),
				Suite:    event.Package,
				Trace:    strings.Join(tests[event.Test].output, ""),
			})
		case "skip":
			report.AddTest(ctrf.Test{
				Name:     event.Test,
				Status:   ctrf.StatusSkipped,
				Duration: durationMs(event.Elapsed),
				Suite:    event.Package,
			})
		}
	}

	// Set actual test times
	if !startTime.IsZero() && !stopTime.IsZero() {
		report.SetTimes(startTime.UnixMilli(), stopTime.UnixMilli())
	} else {
		report.Finalize()
	}
	return report
}
