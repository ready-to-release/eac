package books

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/show"     // register show books
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/validate" // register validate books
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

func TestFeatures(t *testing.T) {
	cfg := eacgodog.RunnerConfig{
		SpecsPath:         "../../../../../../specs/eac-cli/books",
		DefaultReportName: "cucumber-books",
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
