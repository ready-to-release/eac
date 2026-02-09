// Package implicitcli contains godog step implementations for specs/eac/implicit-cli.
//
// This package tests implicit-cli features (importer scripts)
// via subprocess execution of PowerShell/Bash scripts.
//
// These tests invoke the importer scripts and verify they work correctly.
// Platform-specific scenarios use runtime detection to skip on non-matching platforms.
package implicitcli

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

func TestScriptsImplicitCLIFeatures(t *testing.T) {
	cfg := eacgodog.RunnerConfig{
		SpecsPath:         "../../../../../specs/eac/implicit-cli",
		DefaultReportName: "cucumber-implicit-cli",
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
	os.Exit(m.Run())
}
