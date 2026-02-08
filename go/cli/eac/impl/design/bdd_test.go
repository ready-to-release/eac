package design

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

func TestFeatures(t *testing.T) {
	specsBase := "../../../../../specs/eac-cli"

	cfg := eacgodog.RunnerConfig{
		SpecsPath: specsBase + "/create-design," +
			specsBase + "/serve-design," +
			specsBase + "/validate-design",
		DefaultReportName: "cucumber-design",
		AssetsPath:        "go/cli/eac/impl/design/assets",
		RegisterSteps:     registerSteps,
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

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
