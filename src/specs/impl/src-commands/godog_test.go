// Package srccommands contains godog step implementations for specs/src-commands.
//
// This package tests all src-commands features via subprocess execution,
// ensuring complete isolation from the implementation.
package srccommands

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/specs/internal"
)

func TestSrcCommandsFeatures(t *testing.T) {
	cfg := internal.RunnerConfig{
		SpecsPath:         "../../../../specs/src-commands",
		DefaultReportName: "cucumber-src-commands",
		RegisterSteps:     RegisterSteps,
	}

	opts := internal.BuildOptions(cfg.SpecsPath, cfg.DefaultReportName, t)
	suite := godog.TestSuite{
		ScenarioInitializer: internal.CreateScenarioInitializer(cfg),
		Options:             opts,
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
