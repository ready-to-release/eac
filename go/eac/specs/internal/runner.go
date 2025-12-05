package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// RunnerConfig holds configuration for a spec runner.
type RunnerConfig struct {
	// SpecsPath is the relative path to feature files from the test file location
	SpecsPath string

	// DefaultReportName is the default name for the report file
	DefaultReportName string

	// RegisterSteps is a function that registers spec-specific steps
	RegisterSteps func(sc *godog.ScenarioContext, ctx *TestContext)
}

// BuildTagFilter constructs the tag filter from environment and config.
func BuildTagFilter() string {
	// Load config for skip reasons
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Build base tag filter
	skipFilter := cfg.TestingTags.BuildGodogSkipTagFilter()
	tagFilter := skipFilter + " && ~@pending"

	// Add platform-specific exclusions
	tagFilter = tagFilter + " && " + buildPlatformTagFilter()

	// Add suite tag filter if provided
	suiteTagFilter := os.Getenv("GODOG_SUITE_TAGS")
	if suiteTagFilter != "" {
		tagFilter = tagFilter + " && " + suiteTagFilter
	}

	return tagFilter
}

// buildPlatformTagFilter returns a tag filter that excludes platform-incompatible scenarios.
// On Windows, excludes @deps:linux and @deps:darwin
// On Linux, excludes @deps:windows and @deps:darwin
// On Darwin, excludes @deps:windows and @deps:linux
func buildPlatformTagFilter() string {
	switch runtime.GOOS {
	case "windows":
		return "~@deps:linux && ~@deps:darwin"
	case "linux":
		return "~@deps:windows && ~@deps:darwin"
	case "darwin":
		return "~@deps:windows && ~@deps:linux"
	default:
		return "" // No exclusions for unknown platforms
	}
}

// BuildOptions constructs godog options from environment and config.
func BuildOptions(specsPath, defaultReportName string, t *testing.T) *godog.Options {
	outputDir := os.Getenv("GODOG_OUTPUT_DIR")
	reportFormat := os.Getenv("GODOG_REPORT_FORMAT")
	if reportFormat == "" {
		reportFormat = "cucumber"
	}

	consoleFormat := os.Getenv("GODOG_FORMAT")
	if consoleFormat == "" {
		consoleFormat = "pretty"
	}

	paths := os.Getenv("GODOG_PATHS")
	if paths == "" {
		paths = specsPath
	}

	opts := &godog.Options{
		Format:   consoleFormat,
		Paths:    []string{paths},
		TestingT: t,
		Tags:     BuildTagFilter(),
		Strict:   true,
	}

	// Add report formatter if output directory is set
	if outputDir != "" {
		reportName := os.Getenv("GODOG_REPORT_NAME")
		if reportName == "" {
			if reportFormat == "junit" {
				reportName = defaultReportName + ".xml"
			} else {
				reportName = defaultReportName + ".json"
			}
		}
		reportPath := fmt.Sprintf("%s/%s", outputDir, reportName)
		opts.Format = fmt.Sprintf("%s,%s:%s", consoleFormat, reportFormat, reportPath)
	}

	return opts
}

// GetRepoRoot returns the repository root, caching for later use.
func GetRepoRoot() (string, error) {
	return repository.GetRepositoryRoot("")
}

// suiteInitOnce ensures diagnostics are logged only once per test run
var suiteInitOnce sync.Once

// logSuiteInitDiagnostics logs diagnostic information at suite initialization.
// This runs once per test suite and helps debug CI environment issues.
func logSuiteInitDiagnostics(repoRoot, specsPath string) {
	// Only log diagnostics if GODOG_DEBUG_INIT is set or if running in CI
	if os.Getenv("GODOG_DEBUG_INIT") == "" && os.Getenv("CI") == "" && os.Getenv("GITHUB_ACTIONS") == "" {
		return
	}

	suiteInitOnce.Do(func() {
		binaryPath := paths.CommandsBinaryPath(repoRoot)
		binaryExists := false
		binaryInfo := ""
		if info, err := os.Stat(binaryPath); err == nil {
			binaryExists = true
			binaryInfo = fmt.Sprintf("%d bytes, mode %s", info.Size(), info.Mode())
		} else if os.IsNotExist(err) {
			binaryInfo = "not found"
		} else {
			binaryInfo = fmt.Sprintf("error: %v", err)
		}

		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "══════════════════════════════════════════════════════════════════\n")
		fmt.Fprintf(os.Stderr, "  Godog Suite Initialization Diagnostics\n")
		fmt.Fprintf(os.Stderr, "══════════════════════════════════════════════════════════════════\n")
		fmt.Fprintf(os.Stderr, "  Platform:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Fprintf(os.Stderr, "  Repo root:     %s\n", repoRoot)
		fmt.Fprintf(os.Stderr, "  Specs path:    %s\n", specsPath)
		fmt.Fprintf(os.Stderr, "  Binary path:   %s\n", binaryPath)
		fmt.Fprintf(os.Stderr, "  Binary status: %s\n", binaryInfo)
		if !binaryExists {
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "  ⚠️  WARNING: Commands binary not found!\n")
			fmt.Fprintf(os.Stderr, "  Tests requiring command execution will fail.\n")
			fmt.Fprintf(os.Stderr, "  Ensure eac-commands is built before running tests.\n")
		}
		fmt.Fprintf(os.Stderr, "══════════════════════════════════════════════════════════════════\n")
		fmt.Fprintf(os.Stderr, "\n")
	})
}

// CreateScenarioInitializer creates a scenario initializer function.
func CreateScenarioInitializer(cfg RunnerConfig) func(sc *godog.ScenarioContext) {
	// Get repo root once
	repoRoot, err := GetRepoRoot()
	if err != nil {
		log.Fatalf("Failed to get repository root: %v", err)
	}

	// Log suite initialization diagnostics (only once per suite)
	logSuiteInitDiagnostics(repoRoot, cfg.SpecsPath)

	return func(sc *godog.ScenarioContext) {
		// Create context for this scenario
		ctx := NewTestContext()
		ctx.OriginalRepoRoot = repoRoot

		// Register common steps
		RegisterCommonSteps(sc, ctx)

		// Register spec-specific steps
		if cfg.RegisterSteps != nil {
			cfg.RegisterSteps(sc, ctx)
		}

		// Before hook: setup
		sc.Before(func(gctx context.Context, scenario *godog.Scenario) (context.Context, error) {
			ctx.Reset()

			// Check for @env:isolated-test-project tag
			for _, tag := range scenario.Tags {
				if tag.Name == "@env:isolated-test-project" {
					if err := ctx.SetupIsolation(); err != nil {
						return gctx, fmt.Errorf("failed to setup isolation: %w", err)
					}
					break
				}
			}

			return gctx, nil
		})

		// After hook: cleanup
		sc.After(func(gctx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
			ctx.CleanupIsolation()
			return gctx, nil
		})
	}
}
