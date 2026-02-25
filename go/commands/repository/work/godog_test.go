package work

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/core/tool"
)

func TestFeatures(t *testing.T) {
	cfg := eacgodog.RunnerConfig{
		SpecsPath: "../../../../specs/eac_work",
		DefaultReportName: "cucumber-work",
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
	tool.SetGlobalToolSystemForTesting(tool.NewToolSystemForTesting())
	os.Exit(m.Run())
}
