package templates

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/templates/install/ai"      // register templates install ai
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/templates/install/claude"   // register templates install claude
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/templates/install/docs"     // register templates install docs
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/templates/install/reports"  // register templates install reports
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/templates/install/specs" // register templates install specs
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

func TestFeatures(t *testing.T) {
	cfg := eacgodog.RunnerConfig{
		SpecsPath:         "../../../../../specs/eac-cli/templates",
		DefaultReportName: "cucumber-templates",
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
