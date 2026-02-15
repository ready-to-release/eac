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
func (r *GoTestRunner) GetTestInfo(ref testing.TestReference, workspaceRoot string, cfg any) *testrunners.TestInfo {
	eacCfg := cfg.(*config.EACConfig)

	// Calculate relative path from workspace root
	relPath, err := filepath.Rel(workspaceRoot, ref.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)

	info := &testrunners.TestInfo{Language: "go"}

	if ref.Type == "godog" {
		// BDD test: extract from specs path
		specsPrefix := eacCfg.Repository.Paths.SpecsRoot + "/"
		specRelPath := strings.TrimPrefix(relPath, specsPrefix)
		specRelPath = filepath.ToSlash(specRelPath)

		// Get module moniker from first path component
		parts := strings.Split(specRelPath, "/")
		if len(parts) == 0 {
			return nil
		}
		monikerPart := parts[0]

		// Handle compound monikers: "eac_work" → moniker="eac"
		info.ModuleMoniker = monikerPart
		if idx := strings.Index(monikerPart, "_"); idx > 0 {
			info.ModuleMoniker = monikerPart[:idx]
		}

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
	} else {
		// Unit test (gotest): extract module from path using module_mapping
		absDir := filepath.Dir(ref.FilePath)
		relDir, err := filepath.Rel(workspaceRoot, absDir)
		if err != nil {
			return nil
		}
		relDir = filepath.ToSlash(relDir)

		// Find the module this path belongs to
		info.ModuleMoniker = findModuleForPath(relDir, eacCfg)
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
//
// Supports compound monikers: spec dirs named "eac_work" split into moniker="eac", qualifier="work".
// Qualified paths use TestImplPathForQualifier to find the godog component root.
// Unqualified paths use the original TestImplPath for backward compatibility.
func (r *GoTestRunner) FindTestRoot(featurePath string, cfg any) string {
	eacCfg := cfg.(*config.EACConfig)

	// gotest doesn't need a separate test root
	if !strings.HasSuffix(featurePath, ".feature") {
		return ""
	}

	// Extract relative path from specs root
	specsPrefix := eacCfg.Repository.Paths.SpecsRoot + "/"
	relPath := strings.TrimPrefix(filepath.ToSlash(featurePath), specsPrefix)
	relPath = strings.TrimPrefix(relPath, strings.ReplaceAll(specsPrefix, "/", "\\"))
	relPath = filepath.ToSlash(relPath)

	// Get path components
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		goRunnerLog.Debugf("FindTestRoot: empty path parts for %s", featurePath)
		return ""
	}

	monikerPart := parts[0] // e.g., "eac_work" or "eac" or "core"

	// Split compound moniker on underscore: "eac_work" → moniker="eac", qualifier="work"
	moniker := monikerPart
	qualifier := ""
	if idx := strings.Index(monikerPart, "_"); idx > 0 {
		moniker = monikerPart[:idx]
		qualifier = monikerPart[idx+1:]
	}

	// Verify module exists - fail early if not
	if eacCfg.Repository.GetByMoniker(moniker) == nil {
		goRunnerLog.Debugf("FindTestRoot: unknown module %s for %s", moniker, featurePath)
		return ""
	}

	workspaceRoot := eacCfg.RepoRoot
	godogTestFile := eacCfg.Repository.Conventions.GodogTest

	if qualifier != "" {
		// Qualified: find godog root for this component
		basePath := eacCfg.Repository.TestImplPathForQualifier(moniker, qualifier)
		if basePath == "" {
			goRunnerLog.Debugf("FindTestRoot: no godog root for %s qualifier %s", moniker, qualifier)
			return ""
		}
		// Direct check
		if fileExists(filepath.Join(workspaceRoot, basePath, godogTestFile)) {
			goRunnerLog.Debugf("FindTestRoot: found %s at qualified basePath %s", godogTestFile, basePath)
			return basePath
		}
		// Subdirectory navigation for nested runners (e.g., create/commit-message)
		for i := 1; i < len(parts)-1; i++ {
			subPath := filepath.Join(basePath, strings.Join(parts[1:i+1], "/"))
			subPath = filepath.ToSlash(subPath)
			if fileExists(filepath.Join(workspaceRoot, subPath, godogTestFile)) {
				goRunnerLog.Debugf("FindTestRoot: found %s at qualified subPath %s", godogTestFile, subPath)
				return subPath
			}
		}
		goRunnerLog.Debugf("FindTestRoot: no %s found for qualified %s/%s", godogTestFile, moniker, qualifier)
		return ""
	}

	// Unqualified: existing single-root logic (backward compat for core, repository)
	basePath := eacCfg.Repository.TestImplPath(moniker)
	if basePath == "" {
		goRunnerLog.Debugf("FindTestRoot: no test-impl path for module %s", moniker)
		return ""
	}

	// Check if godog test file exists at base path
	if fileExists(filepath.Join(workspaceRoot, basePath, godogTestFile)) {
		goRunnerLog.Debugf("FindTestRoot: found %s at basePath %s", godogTestFile, basePath)
		return basePath
	}

	// Try adding subdirectories (skip the filename at the end)
	for i := 1; i < len(parts)-1; i++ {
		subPath := filepath.Join(basePath, strings.Join(parts[1:i+1], "/"))
		subPath = filepath.ToSlash(subPath)
		if fileExists(filepath.Join(workspaceRoot, subPath, godogTestFile)) {
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

// parsedGoTestPkg holds the parsed components of a Go test package path.
type parsedGoTestPkg struct {
	displayName    string
	relPkgPath     string
	relFeatureFile string
}

// parseGoTestPackagePath parses a package path in "featureName:testRoot:featurePath" or "testRoot" format.
func parseGoTestPackagePath(pkgPath string) parsedGoTestPkg {
	parts := strings.Split(pkgPath, ":")
	switch len(parts) {
	case 3:
		// BDD format: featureName:testRoot:featurePath
		return parsedGoTestPkg{
			displayName:    parts[0] + ":" + parts[1],
			relPkgPath:     parts[1],
			relFeatureFile: parts[2],
		}
	case 1:
		// Unit test format: just the package path
		return parsedGoTestPkg{
			displayName: parts[0],
			relPkgPath:  parts[0],
		}
	default:
		// Fallback for unexpected format
		return parsedGoTestPkg{
			displayName: pkgPath,
			relPkgPath:  pkgPath,
		}
	}
}

// buildGoTestArgs constructs the go test command arguments including build tags,
// coverage flags, and the package path.
func buildGoTestArgs(suiteTagFilterStr string, cfg testrunners.RunConfig) []string {
	args := []string{"test", "-json", "-v"}

	// Extract Go build tags from suite filter
	if buildTags := extractGoBuildTags(suiteTagFilterStr); buildTags != "" {
		args = append(args, "-tags", buildTags)
	}

	// Add coverage if enabled
	if cfg.Coverage && cfg.OutputDir != "" {
		coverageFile := filepath.Join(cfg.OutputDir, "coverage.out")
		args = append(args, "-cover", "-coverprofile="+coverageFile)
	}

	// Add package path
	args = append(args, ".")
	return args
}

// buildGodogEnvOverrides constructs the godog-specific environment variable overrides
// for BDD test execution including format, tag filters, output dir, and feature paths.
func buildGodogEnvOverrides(actualPkgDir string, cfg testrunners.RunConfig, suiteTagFilterStr, relFeatureFile string) map[string]string {
	envOverrides := map[string]string{
		"GODOG_FORMAT": "progress",
	}
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
	return envOverrides
}

// saveGoTestCTRF converts streaming go test events to CTRF format and writes
// the JSON artifact to the output directory. Only applies to non-godog tests.
func saveGoTestCTRF(streamingRunner *testrunners.StreamingRunner, outputDir string) {
	if outputDir == "" {
		return
	}
	events := streamingRunner.GetEvents()
	if len(events) == 0 {
		return
	}
	report := convertGoTestEventsToCTRF(events)
	if ctrfData, err := report.ToJSON(); err == nil {
		jsonPath := filepath.Join(outputDir, "unit.json")
		_ = os.WriteFile(jsonPath, ctrfData, 0o644) //nolint:errcheck // best-effort artifact save
	}
}

// Execute runs Go tests for a package and returns results.
// logWriter is provided by the orchestrator (UoW manages log files).
func (r *GoTestRunner) Execute(pkgPath string, tests []testing.TestReference, logWriter io.Writer, cfg testrunners.RunConfig) testrunners.RunResult {
	start := time.Now()

	pkg := parseGoTestPackagePath(pkgPath)
	result := testrunners.RunResult{
		PackageName:   pkg.displayName,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	actualPkgDir := filepath.Join(cfg.WorkspaceRoot, pkg.relPkgPath)

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

	goTestArgs := buildGoTestArgs(suiteTagFilterStr, cfg)

	// Build environment overrides
	isGodogTest := fileExists(filepath.Join(actualPkgDir, "godog_test.go"))
	testRunID := filepath.Base(cfg.TestRunDir)

	envOverrides := map[string]string{
		"CLIE_TEST_RUN_ID":         testRunID,
		"CLIE_TEST_LOGGING_ACTIVE": "true",
	}

	if isGodogTest {
		for k, v := range buildGodogEnvOverrides(actualPkgDir, cfg, suiteTagFilterStr, pkg.relFeatureFile) {
			envOverrides[k] = v
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
	if !isGodogTest {
		saveGoTestCTRF(streamingRunner, cfg.OutputDir)
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
