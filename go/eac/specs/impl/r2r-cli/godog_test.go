// Package srccli contains godog step implementations for specs/r2r-cli.
//
// This package tests r2r-cli features (CLI invocation, verify-configuration, module-isolation)
// via subprocess execution, ensuring complete isolation from the implementation.
//
// Prerequisites:
// - Requires pre-built executable from "build module r2r-cli"
// - Executable location: out/build/r2r-cli/windows-r2r-cli.exe (or linux-r2r-cli, darwin-r2r-cli)
package srccli

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

func TestSrcCliFeatures(t *testing.T) {
	cfg := internal.RunnerConfig{
		SpecsPath:         "../../../../../specs/r2r-cli",
		DefaultReportName: "cucumber-r2r-cli",
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
