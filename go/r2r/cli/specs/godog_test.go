// Package specs contains godog step implementations for specs/r2r-cli.
//
// This package tests r2r-cli features (CLI invocation, verify-configuration, module-isolation)
// via subprocess execution, ensuring complete isolation from the implementation.
//
// Prerequisites:
// - Requires pre-built executable from "build module r2r-cli"
// - Executable location: out/build/r2r-cli/windows-r2r-cli.exe (or linux-r2r-cli, darwin-r2r-cli)
package specs

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/eac/godog"
)

func TestR2RCliFeatures(t *testing.T) {
	cfg := eacgodog.RunnerConfig{
		SpecsPath:         "../../../../specs/r2r-cli",
		DefaultReportName: "cucumber-r2r-cli",
		RegisterSteps:     registerAllSteps,
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

// registerAllSteps registers all r2r-cli specific step definitions.
// Currently all CLI steps are handled by the common eacgodog steps.
func registerAllSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// r2r-cli tests use common step definitions from eacgodog
	// No module-specific steps are currently needed.
}
