// Package specs contains godog BDD test implementations for core.
//
// This package runs all Gherkin specifications under specs/core/ and provides
// step implementations for testing the core domain libraries.
//
// Features tested:
//   - cache-invalidation: Build cache invalidation and module rebuild detection
//   - config-defaults: Configuration loading and defaults system
//   - logging: Structured logging behavior
//   - tool-system: Tool registry and configuration
package specs

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

// TestEacCoreFeatures runs all core Gherkin specifications.
// This is the single entry point for running all core BDD tests.
func TestEacCoreFeatures(t *testing.T) {
	cfg := eacgodog.RunnerConfig{
		SpecsPath:         "../../../specs/core",
		DefaultReportName: "cucumber-core",
		RegisterSteps:     registerAllSteps,
	}

	opts := eacgodog.BuildOptions(cfg.SpecsPath, cfg.DefaultReportName, t)
	suite := godog.TestSuite{
		ScenarioInitializer: eacgodog.CreateScenarioInitializer(cfg),
		Options:             opts,
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

// registerAllSteps registers all step implementations for core features.
// Each feature's steps are organized in separate files:
//   - steps_cache_test.go: cache-invalidation feature
//   - steps_config_test.go: config-defaults feature
//   - steps_logging_test.go: logging feature
//   - steps_tool_test.go: tool-system feature
func registerAllSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	registerCacheSteps(sc, ctx)
	registerConfigSteps(sc, ctx)
	registerLoggingSteps(sc, ctx)
	registerToolSteps(sc, ctx)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
