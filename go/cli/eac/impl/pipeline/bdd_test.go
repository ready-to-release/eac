package pipeline

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

func TestFeatures(t *testing.T) {
	specsBase := "../../../../../specs/eac-cli"

	cfg := eacgodog.RunnerConfig{
		SpecsPath: specsBase + "/pipeline-run," +
			specsBase + "/pipeline-status",
		DefaultReportName: "cucumber-pipeline",
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

// registerAllSteps registers all step definitions for pipeline tests.
func registerAllSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	registerSteps(sc, ctx)
	registerStatusSteps(sc, ctx)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
