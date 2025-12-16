// Package implicitr2rcli contains godog step implementations for specs/implicit-r2r-cli.
//
// This package tests implicit-r2r-cli features (importer scripts)
// via subprocess execution of PowerShell/Bash scripts.
//
// These tests invoke the importer scripts and verify they work correctly.
// Platform-specific scenarios use runtime detection to skip on non-matching platforms.
package implicitr2rcli

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

func TestScriptsImplicitCLIFeatures(t *testing.T) {
	cfg := internal.RunnerConfig{
		SpecsPath:         "../../../../../specs/implicit-r2r-cli",
		DefaultReportName: "cucumber-implicit-r2r-cli",
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
