package squashmessage

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	_ "github.com/ready-to-release/eac/go/adapters/ai-test"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

func TestFeatures(t *testing.T) {
	cfg := eacgodog.RunnerConfig{
		SpecsPath:         "../../../../../specs/eac_get/squash-message",
		DefaultReportName: "cucumber-squash-message",
		AssetsPath:        "go/commands/repository/get/squash-message/assets",
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
	// Register this command for in-process dispatch.
	// This avoids subprocess overhead and removes dependency on pre-built binary.
	reg := registry.NewCommandRegistry()
	if err := reg.RegisterAll(Commands()...); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register commands: %v\n", err)
		os.Exit(1)
	}
	registry.SetGlobal(reg)
	flags.SetRegistry(reg)

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
