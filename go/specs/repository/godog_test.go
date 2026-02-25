// Package repository contains godog step implementations for specs/repository.
package repository

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	buildcmd "github.com/ready-to-release/eac/go/commands/build"
	"github.com/ready-to-release/eac/go/commands/repository/validate"
	specsunused "github.com/ready-to-release/eac/go/commands/repository/specs/unused"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

func TestRepositoryFeatures(t *testing.T) {
	cfg := eacgodog.RunnerConfig{
		SpecsPath:         "../../../specs/repository",
		DefaultReportName: "cucumber-repository",
		RegisterSteps:     RegisterSteps,
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

	var allCmds []core.CommandPort
	allCmds = append(allCmds, validate.Commands()...)
	allCmds = append(allCmds, specsunused.Commands()...)
	allCmds = append(allCmds, buildcmd.Commands()...)

	if err := reg.RegisterAll(allCmds...); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register commands: %v\n", err)
		os.Exit(1)
	}
	registry.SetGlobal(reg)
	flags.SetRegistry(reg)
	os.Exit(m.Run())
}
