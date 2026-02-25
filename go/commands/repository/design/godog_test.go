package design

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	servedesign "github.com/ready-to-release/eac/go/commands/repository/serve/design"
	"github.com/ready-to-release/eac/go/commands/repository/validate"
	"github.com/ready-to-release/eac/go/core/tool"
)

func TestFeatures(t *testing.T) {
	cfg := eacgodog.RunnerConfig{
		SpecsPath:         "../../../../specs/eac_design",
		DefaultReportName: "cucumber-design",
		AssetsPath:        "go/commands/repository/design/assets",
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
	// Register commands that can run in-process.
	// Note: create/design and update/design are excluded because they import
	// this package (design), which would create a circular dependency.
	// Unregistered commands fall back to subprocess via ExitCodeDispatchDeclined.
	reg := registry.NewCommandRegistry()
	var cmds []core.CommandPort
	cmds = append(cmds, validate.Commands()...)
	cmds = append(cmds, servedesign.Commands()...)
	if err := reg.RegisterAll(cmds...); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register design commands: %v\n", err)
		os.Exit(1)
	}
	registry.SetGlobal(reg)
	flags.SetRegistry(reg)

	tool.SetGlobalToolSystemForTesting(tool.NewToolSystemForTesting())
	os.Exit(m.Run())
}

// registryLookup adapts the CommandRegistryPort to the CommandLookupFunc signature.
func registryLookup(cmdName string) (func() int, bool) {
	reg := registry.Global()
	if reg == nil {
		return nil, false
	}
	cmd, ok := reg.Get(cmdName)
	if !ok {
		return nil, false
	}
	simple, ok := cmd.(core.SimpleCommandPort)
	if !ok {
		return nil, false
	}
	return func() int {
		return simple.Execute(context.Background(), &core.CommandRequest{
			Args:   os.Args[1:],
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		})
	}, true
}
