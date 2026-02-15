package templates

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	templatesaicmds "github.com/ready-to-release/eac/go/commands/repository/templates/install/ai"
	templatesclaudecmds "github.com/ready-to-release/eac/go/commands/repository/templates/install/claude"
	templatesdocscmds "github.com/ready-to-release/eac/go/commands/repository/templates/install/docs"
	templatesreportscmds "github.com/ready-to-release/eac/go/commands/repository/templates/install/reports"
	templatesspeccmds "github.com/ready-to-release/eac/go/commands/repository/templates/install/specs"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

func TestFeatures(t *testing.T) {
	cfg := eacgodog.RunnerConfig{
		SpecsPath:         "../../../../specs/eac_templates",
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
	reg := registry.NewCommandRegistry()
	if err := reg.RegisterAll(Commands()...); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register templates commands: %v\n", err)
		os.Exit(1)
	}
	// Register install subcommands
	for _, err := range []error{
		reg.RegisterAll(templatesaicmds.Commands()...),
		reg.RegisterAll(templatesclaudecmds.Commands()...),
		reg.RegisterAll(templatesdocscmds.Commands()...),
		reg.RegisterAll(templatesreportscmds.Commands()...),
		reg.RegisterAll(templatesspeccmds.Commands()...),
	} {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to register template install commands: %v\n", err)
			os.Exit(1)
		}
	}
	registry.SetGlobal(reg)
	flags.SetRegistry(reg)
	os.Exit(m.Run())
}
