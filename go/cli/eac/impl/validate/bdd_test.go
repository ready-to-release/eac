package validate

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/create/spec" // register create spec
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

func TestFeatures(t *testing.T) {
	specsBase := "../../../../../specs/eac-cli"

	cfg := eacgodog.RunnerConfig{
		SpecsPath: specsBase + "/validate-specs," +
			specsBase + "/create-spec",
		DefaultReportName: "cucumber-specs",
		AssetsPath:        "go/cli/eac/impl/validate/assets",
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
