package help

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

func TestFeatures(t *testing.T) {
	specsBase := "../../../../../specs/eac-cli"

	cfg := eacgodog.RunnerConfig{
		SpecsPath: specsBase + "/show-help," +
			specsBase + "/get-commands",
		DefaultReportName: "cucumber-help",
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
