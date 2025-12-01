// Package srccli contains godog step implementations for specs/src-cli.
//
// This package tests src-cli features (CLI invocation, verify-configuration, module-isolation)
// via subprocess execution, ensuring complete isolation from the implementation.
//
// Prerequisites:
// - Requires pre-built executable from "build module src-cli"
// - Executable location: out/build/src-cli/windows-r2r-cli.exe (or linux-r2r-cli, darwin-r2r-cli)
package srccli

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/specs/internal"
)

func TestSrcCliFeatures(t *testing.T) {
	cfg := internal.RunnerConfig{
		SpecsPath:         "../../../../specs/src-cli",
		DefaultReportName: "cucumber-src-cli",
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
