package test

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	_ "github.com/ready-to-release/eac/go/cli/eac/impl/get" // register get commands
	"github.com/ready-to-release/eac/go/clibase/registry"
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

func TestFeatures(t *testing.T) {
	specsBase := "../../../../../specs/eac-cli"

	cfg := eacgodog.RunnerConfig{
		SpecsPath: specsBase + "/get-test-results," +
			specsBase + "/show-test-results",
		DefaultReportName: "cucumber-test-results",
		AssetsPath:        "go/cli/eac/impl/test/assets",
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

// registryLookup adapts registry.GetCommand to the CommandLookupFunc signature.
func registryLookup(cmdName string) (func() int, bool) {
	reg := registry.GetCommand(cmdName)
	if reg == nil {
		return nil, false
	}
	return reg.Func, true
}
