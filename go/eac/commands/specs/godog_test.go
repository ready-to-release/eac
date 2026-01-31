// Package specs contains godog BDD test implementations for eac-commands.
//
// This package tests all eac-commands features via subprocess execution,
// ensuring complete isolation from the implementation.
package specs

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/eac/godog"
)

// TestEacCommandsFeatures runs all eac-commands Gherkin specifications.
// This is the single entry point for running all eac-commands BDD tests.
func TestEacCommandsFeatures(t *testing.T) {
	cfg := eacgodog.RunnerConfig{
		SpecsPath:         "../../../../specs/eac-commands",
		DefaultReportName: "cucumber-eac-commands",
		AssetsPath:        "go/eac/commands/specs/assets",
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

// registerAllSteps registers all step implementations for eac-commands features.
// Each feature's steps are organized in separate files.
func registerAllSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	registerTemplatesSteps(sc, ctx)
	registerInitSteps(sc, ctx)
	registerWorkSteps(sc, ctx)
	registerWorkspacesSteps(sc, ctx)
	registerDocsSteps(sc, ctx)
	registerCreateCommitMessageSteps(sc, ctx)
	registerCreateSquashMessageSteps(sc, ctx)
	registerDesignSteps(sc, ctx)
	registerSpecsSteps(sc, ctx)
	registerRiskSteps(sc, ctx)
	registerHelpSteps(sc, ctx)
	registerGitSetupSteps(sc, ctx)
	registerBooksSteps(sc, ctx)
	registerTestResultsSteps(sc, ctx)
	registerDrawioSteps(sc, ctx)
	registerPrunePackagesSteps(sc, ctx)
	registerPipelineSteps(sc, ctx)
	registerPipelineStatusSteps(sc, ctx)
	registerReleaseSteps(sc, ctx)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
