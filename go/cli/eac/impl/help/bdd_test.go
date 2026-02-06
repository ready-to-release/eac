package help

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/build"         // register build
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/create/design" // register create design
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/create/spec"   // register create spec
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/describe"      // register get commands
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/list"          // register show help
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/show"          // register show modules, show books, etc.
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/work"          // register create pr
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
