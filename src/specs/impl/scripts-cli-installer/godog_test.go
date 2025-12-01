// Package scriptsinstaller contains godog step implementations for specs/scripts-cli-installer.
//
// This package tests scripts-cli-installer features (CLI installation scripts)
// via subprocess execution of PowerShell/Bash installer scripts.
//
// These tests invoke the installer scripts and verify they work correctly.
// Platform-specific scenarios use runtime detection to skip on non-matching platforms.
package scriptscliinstaller

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/specs/internal"
)

func TestScriptsInstallerFeatures(t *testing.T) {
	cfg := internal.RunnerConfig{
		SpecsPath:         "../../../../specs/scripts-cli-installer",
		DefaultReportName: "cucumber-scripts-cli-installer",
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
