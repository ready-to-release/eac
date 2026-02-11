package validate

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	speccmds "github.com/ready-to-release/eac/go/cli/eac/impl/create/spec"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

func TestFeatures(t *testing.T) {
	specsBase := "../../../../../specs/eac"

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
	reg := registry.NewCommandRegistry()
	if err := reg.RegisterAll(Commands()...); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register validate commands: %v\n", err)
		os.Exit(1)
	}
	if err := reg.RegisterAll(speccmds.Commands()...); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register spec commands: %v\n", err)
		os.Exit(1)
	}
	registry.SetGlobal(reg)
	flags.SetRegistry(reg)
	os.Exit(m.Run())
}
