// Package gotest provides a test runner adapter for Go tests (gotest and godog).
package gotest

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	godogadapter "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/clibase/testrunners"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
)

var goRunnerLog = logging.C()

func init() {
	goRunner := &GoTestRunner{}
	testrunners.Register(goRunner)
	testrunners.RegisterFallback(goRunner) // Go runner is the fallback for unknown types

	// Register descriptor for gotest type
	testrunners.RegisterDescriptor(&testrunners.TestTypeDescriptor{
		TestType:      "gotest",
		IsBDD:         false,
		ComponentType: "go",
		MonikerStyle:  "file",
		DefaultInferences: []testrunners.Inference{
			{
				TestTypes:   []string{"gotest"},
				ThenAddTags: []string{"@deps:go"},
				Description: "Go tests require Go toolchain",
			},
			{
				TestTypes:   []string{"gotest"},
				ThenAddTags: []string{"@L1"},
				Description: "Go unit tests default to L1 (unit test level)",
			},
		},
	})
}

// GoTestRunner handles Go test execution for both unit tests (gotest) and BDD tests (godog).
type GoTestRunner struct{}

// TestTypes returns the test types this runner handles.
func (r *GoTestRunner) TestTypes() []string {
	return []string{"gotest", "godog"}
}

// IsBDD returns true because this runner handles godog BDD tests.
func (r *GoTestRunner) IsBDD() bool {
	return true
}

// GetTestInfo extracts structured test metadata from a Go test reference.
func (r *GoTestRunner) GetTestInfo(test testing.TestReference, workspaceRoot string, cfg *config.EACConfig) *testrunners.TestInfo {
	// Calculate relative path from workspace root
	relPath, err := filepath.Rel(workspaceRoot, test.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)

	info := &testrunners.TestInfo{Language: "go"}

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

// FindTestRoot finds the test runner package for a godog feature file.
// For gotest, returns empty string (tests are in the same directory as source).
// For godog, returns the path to the directory containing godog_test.go.
// Returns empty string if module not found or no godog_test.go exists - caller must handle.
func (r *GoTestRunner) FindTestRoot(featurePath string, cfg *config.EACConfig) string {
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

	// Check if godog test file exists at base path
	workspaceRoot := cfg.RepoRoot
	godogTestFile := cfg.Repository.Conventions.GodogTest
	baseCheck := filepath.Join(workspaceRoot, basePath, godogTestFile)
	if fileExists(baseCheck) {
		goRunnerLog.Debugf("FindTestRoot: found %s at basePath %s", godogTestFile, basePath)
		return basePath
	}

	// Try adding subdirectories (skip the filename at the end)
	for i := 1; i < len(parts)-1; i++ {
		subPath := filepath.Join(basePath, strings.Join(parts[1:i+1], "/"))
		subPath = filepath.ToSlash(subPath)
		subCheck := filepath.Join(workspaceRoot, subPath, godogTestFile)
		if fileExists(subCheck) {
			goRunnerLog.Debugf("FindTestRoot: found %s at subPath %s", godogTestFile, subPath)
			return subPath
		}
	}

	// No test runner found - no fallback
	goRunnerLog.Debugf("FindTestRoot: no %s found for %s (basePath=%s, parts=%v)", godogTestFile, featurePath, basePath, parts)
	return ""
}

// BuildPackagePath constructs the package path for test grouping.
// For godog BDD tests, returns "featureFolderName:testRoot" for cleaner display.
// For gotest, returns the directory path.
func (r *GoTestRunner) BuildPackagePath(testRoot, featurePath string) string {
	if featurePath != "" && testRoot != "" {
		featureFolderName := extractFeatureFolderName(featurePath)
		return featureFolderName + ":" + testRoot + ":" + featurePath
	}
	return testRoot
}

// Execute runs Go tests for a package and returns results.
// logWriter is provided by the orchestrator (UoW manages log files).
func (r *GoTestRunner) Execute(pkgPath string, tests []testing.TestReference, logWriter io.Writer, cfg testrunners.RunConfig) testrunners.RunResult {
	start := time.Now()

	// Parse package path - new format: "featureName:testRoot:featurePath" or "testRoot"
	var displayName, relPkgPath, relFeatureFile string
	parts := strings.Split(pkgPath, ":")
	if len(parts) == 3 {
		// BDD format: featureName:testRoot:featurePath
		displayName = parts[0] + ":" + parts[1]
		relPkgPath = parts[1]
		relFeatureFile = parts[2]
	} else if len(parts) == 1 {
		// Unit test format: just the package path
		displayName = parts[0]
		relPkgPath = parts[0]
	} else {
		// Fallback for unexpected format
		displayName = pkgPath
		relPkgPath = pkgPath
	}

	result := testrunners.RunResult{
		PackageName:   displayName,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	actualPkgDir := filepath.Join(cfg.WorkspaceRoot, relPkgPath)

	// UoW creates the log file - runner just writes to logWriter
	streamingRunner := testrunners.NewStreamingRunner(logWriter, logWriter)

	// Use worker context for subprocess cancellation
	runCtx := cfg.Ctx
	if runCtx == nil {
		runCtx = context.Background()
	}

	// Run go generate to ensure embedded files exist
	if err := runGoGenerate(runCtx, actualPkgDir, logWriter); err != nil {
		fmt.Fprintf(logWriter, "Warning: go generate failed: %v\n", err)
	}

	// Translate TagFilter to godog string for build tag extraction and env var
	translator := &godogadapter.GodogTagTranslator{}
	suiteTagFilterStr := translator.TranslateTagFilter(cfg.SuiteTagFilter)

	// Build go test command
	goTestArgs := []string{"test", "-json", "-v"}

	// Extract Go build tags from suite filter
	if buildTags := extractGoBuildTags(suiteTagFilterStr); buildTags != "" {
		goTestArgs = append(goTestArgs, "-tags", buildTags)
	}

	// Add coverage if enabled
	if cfg.Coverage && cfg.OutputDir != "" {
		coverageFile := filepath.Join(cfg.OutputDir, "coverage.out")
		goTestArgs = append(goTestArgs, "-cover", "-coverprofile="+coverageFile)
	}

	// Add package path
	goTestArgs = append(goTestArgs, ".")

	// Build godog-specific environment variables
	isGodogTest := fileExists(filepath.Join(actualPkgDir, "godog_test.go"))
	testRunID := filepath.Base(cfg.TestRunDir)

	envOverrides := map[string]string{
		"R2R_TEST_RUN_ID":         testRunID,
		"R2R_TEST_LOGGING_ACTIVE": "true",
	}

	if isGodogTest {
		envOverrides["GODOG_FORMAT"] = "progress"
		if suiteTagFilterStr != "" {
			envOverrides["GODOG_SUITE_TAGS"] = suiteTagFilterStr
		}
		if cfg.OutputDir != "" {
			envOverrides["GODOG_OUTPUT_DIR"] = cfg.OutputDir
		}
		if relFeatureFile != "" {
			relFeaturePath, relErr := filepath.Rel(actualPkgDir, filepath.Join(cfg.WorkspaceRoot, relFeatureFile))
			if relErr == nil {
				envOverrides["GODOG_PATHS"] = filepath.ToSlash(relFeaturePath)
			}
		}
	}

	// Use tool.BuildCommand for streaming JSON test event parsing
	goTool := tool.GlobalRegistry().GetOrAdhoc("go")
	cmd := tool.BuildCommand(runCtx, goTool, &tool.ExecutionContext{
		WorkspaceRoot: cfg.WorkspaceRoot,
		ModuleRoot:    actualPkgDir,
		ArgsOverrides: goTestArgs,
		EnvOverrides:  envOverrides,
	})

	// Run tests with streaming output
	testResult, runErr := streamingRunner.Run(cmd)

	// Save CTRF JSON output for unit tests (non-godog)
	if !isGodogTest && cfg.OutputDir != "" {
		events := streamingRunner.GetEvents()
		if len(events) > 0 {
			report := convertGoTestEventsToCTRF(events)
			if ctrfData, err := report.ToJSON(); err == nil {
				jsonPath := filepath.Join(cfg.OutputDir, "unit.json")
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

// Ensure GoTestRunner implements the interface (compile-time check).
var _ testrunners.TestTypeRunner = (*GoTestRunner)(nil)
