package godog

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
	coretesting "github.com/ready-to-release/eac/go/core/testing"
)

// log is shared across the package (declared in cache.go)

// Global process-level cache shared across ALL test packages
// This prevents 14+ concurrent git ls-files calls when running tests in parallel.
var (
	globalRepoCache     *TestCache
	globalRepoCacheOnce sync.Once
)

// RunnerConfig holds configuration for a spec runner.
type RunnerConfig struct {
	// SpecsPath is the relative path to feature files from the test file location
	SpecsPath string

	// DefaultReportName is the default name for the report file
	DefaultReportName string

	// AssetsPath is the path to the assets directory relative to the repository root.
	// Example: "go/cli/eac/specs/assets" for eac-cli.
	// If empty, LoadAsset will fail with an error.
	AssetsPath string

	// RegisterSteps is a function that registers spec-specific steps
	RegisterSteps func(sc *godog.ScenarioContext, ctx *TestContext)
}

// BuildTagFilter constructs the tag filter from environment and config.
func BuildTagFilter() string {
	// Load config for skip reasons
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Collect all filter parts, then join with " && "
	// This avoids invalid expressions like " && ~@pending" when some parts are empty
	var parts []string

	// Skip reasons filter: get raw tags and translate to godog syntax
	skipTags := cfg.Testing.GetSkipTags()
	if len(skipTags) > 0 {
		translator := &GodogTagTranslator{}
		skipFilter := translator.TranslateTagFilter(core.TagFilter{
			Selectors: []core.TagFilterSelector{{ExcludeTags: skipTags}},
		})
		if skipFilter != "" {
			parts = append(parts, skipFilter)
		}
	}

	// Always exclude @pending
	parts = append(parts, "~@pending")

	// Platform-specific exclusions
	if platformFilter := buildPlatformTagFilter(); platformFilter != "" {
		parts = append(parts, platformFilter)
	}

	// Suite tag filter from environment
	if suiteTagFilter := os.Getenv(environments.EnvGodogSuiteTags); suiteTagFilter != "" {
		parts = append(parts, suiteTagFilter)
	}

	return strings.Join(parts, " && ")
}

// buildPlatformTagFilter returns a tag filter that excludes platform-incompatible scenarios.
// On Windows, excludes @deps:linux and @deps:darwin
// On Linux, excludes @deps:windows and @deps:darwin
// On Darwin, excludes @deps:windows and @deps:linux.
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
// Generates cucumber.json report when output directory is set.
func BuildOptions(specsPath, defaultReportName string, t *testing.T) *godog.Options {
	outputDir := os.Getenv(environments.EnvGodogOutputDir)

	consoleFormat := os.Getenv(environments.EnvGodogFormat)
	if consoleFormat == "" {
		consoleFormat = "pretty"
	}

	pathsStr := os.Getenv(environments.EnvGodogPaths)
	if pathsStr == "" {
		pathsStr = specsPath
	}

	// Split comma-separated paths into array
	paths := strings.Split(pathsStr, ",")
	for i := range paths {
		paths[i] = strings.TrimSpace(paths[i])
	}

	opts := &godog.Options{
		Format:   consoleFormat,
		Paths:    paths,
		TestingT: t,
		Tags:     BuildTagFilter(),
		Strict:   true,
	}

	// Add cucumber report formatter if output directory is set
	if outputDir != "" {
		reportName := os.Getenv(environments.EnvGodogReportName)
		if reportName == "" {
			reportName = defaultReportName
		}
		// Strip any extension from reportName
		reportName = strings.TrimSuffix(reportName, ".json")

		cucumberPath := fmt.Sprintf("%s/%s.cucumber.json", outputDir, reportName)
		opts.Format = fmt.Sprintf("%s,cucumber:%s", consoleFormat, cucumberPath)
	}

	return opts
}

// GetRepoRoot returns the repository root, caching for later use.
func GetRepoRoot() (string, error) {
	return repository.GetRepositoryRoot("")
}

// suiteInitOnce ensures diagnostics are logged only once per test run.
var suiteInitOnce sync.Once

// logSuiteInitDiagnostics logs diagnostic information at suite initialization.
// This runs once per test suite and helps debug CI environment issues.
func logSuiteInitDiagnostics(repoRoot, specsPath string) {
	// Only log diagnostics if GODOG_DEBUG_INIT is set or if running in CI
	if os.Getenv(environments.EnvGodogDebugInit) == "" && os.Getenv(environments.EnvCI) == "" && os.Getenv(environments.EnvGitHubActions) == "" {
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
			fmt.Fprintf(os.Stderr, "  Ensure eac-cli is built before running tests.\n")
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
		panic(fmt.Sprintf("Failed to get repository root: %v", err))
	}

	// Log suite initialization diagnostics (only once per suite)
	logSuiteInitDiagnostics(repoRoot, cfg.SpecsPath)

	// Create fixture pool for this test suite (mandatory for test isolation)
	fixturePool := coretesting.NewFixturePool()
	fixtureTemplate := (*coretesting.FixtureTemplate)(nil)
	var templateMu sync.Mutex

	// Use GLOBAL process-level cache shared across all test packages
	// This prevents 14+ concurrent git ls-files calls when running tests in parallel
	// The cache is created once per process, not once per godog suite
	globalRepoCacheOnce.Do(func() {
		globalRepoCache = NewTestCache()
	})

	return func(sc *godog.ScenarioContext) {
		// Create context for this scenario
		ctx := NewTestContext()
		ctx.OriginalRepoRoot = repoRoot
		ctx.AssetsPath = cfg.AssetsPath
		ctx.FixturePool = fixturePool
		ctx.OriginalRepoCache = globalRepoCache // Share global cache across ALL test packages!

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
			hasIsolationTag := false
			for _, tag := range scenario.Tags {
				if tag.Name == "@env:isolated-test-project" {
					hasIsolationTag = true
					break
				}
			}

			if hasIsolationTag {
				// Create fixture template once (first scenario with isolation tag)
				templateMu.Lock()
				if fixtureTemplate == nil {
					start := time.Now()
					template, err := fixturePool.CreateTemplate(repoRoot)
					duration := time.Since(start)
					if err != nil {
						templateMu.Unlock()
						return gctx, fmt.Errorf("failed to create fixture template: %w", err)
					}
					fixtureTemplate = template
					log.Debugf("Fixture template created: %v", duration)
				}
				templateMu.Unlock()

				// Setup isolation (fast-copies from template, fails if template unavailable)
				start := time.Now()
				if err := ctx.SetupIsolation(); err != nil {
					return gctx, fmt.Errorf("failed to setup isolation: %w", err)
				}
				duration := time.Since(start)
				if duration > 100*time.Millisecond {
					log.Debugf("Isolation setup: %v", duration)
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
