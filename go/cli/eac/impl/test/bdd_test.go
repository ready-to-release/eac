package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	getcmds "github.com/ready-to-release/eac/go/cli/eac/impl/get"
	showcmds "github.com/ready-to-release/eac/go/cli/eac/impl/show"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

func TestFeatures(t *testing.T) {
	specsBase := "../../../../../specs/eac"

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
	reg := registry.NewCommandRegistry()
	for _, cmds := range [][]core.CommandPort{
		Commands(),
		getcmds.Commands(),
		showcmds.Commands(),
	} {
		if err := reg.RegisterAll(cmds...); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to register commands: %v\n", err)
			os.Exit(1)
		}
	}
	registry.SetGlobal(reg)
	flags.SetRegistry(reg)
	os.Exit(m.Run())
}

// registryLookup adapts the new CommandRegistryPort to the CommandLookupFunc signature.
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
