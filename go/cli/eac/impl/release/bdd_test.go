package release

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

func TestFeatures(t *testing.T) {
	specsBase := "../../../../../specs/eac-cli"

	cfg := eacgodog.RunnerConfig{
		SpecsPath: specsBase + "/release-this," +
			specsBase + "/release-prune-packages",
		DefaultReportName: "cucumber-release",
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

// registerAllSteps registers all step definitions for release tests.
func registerAllSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	registerSteps(sc, ctx)
	registerPruneSteps(sc, ctx)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
