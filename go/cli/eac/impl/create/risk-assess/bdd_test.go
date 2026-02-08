package riskassess

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

func TestFeatures(t *testing.T) {
	specsBase := "../../../../../../specs/eac-cli"

	cfg := eacgodog.RunnerConfig{
		SpecsPath: specsBase + "/create-risk-assess," +
			specsBase + "/create-risk-profile," +
			specsBase + "/validate-risk-catalog," +
			specsBase + "/validate-risk-profile",
		DefaultReportName: "cucumber-risk",
		AssetsPath:        "go/cli/eac/impl/create/risk-assess/assets",
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
