// Package installer contains godog step implementations for specs/clie-cli/installer.
//
// This package tests clie-installer features (CLI installation scripts)
// via subprocess execution of PowerShell/Bash installer scripts.
//
// These tests invoke the installer scripts and verify they work correctly.
// Platform-specific scenarios use runtime detection to skip on non-matching platforms.
package installer

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

func TestCLIEInstallerFeatures(t *testing.T) {
	// Enable mock mode to skip GitHub API calls during tests
	// This avoids 3+ second network latency per test
	os.Setenv("__CLIE_TEST_MOCK", "1")

	cfg := eacgodog.RunnerConfig{
		SpecsPath:         "../../../../../specs/clie-cli/installer",
		DefaultReportName: "cucumber-clie-installer",
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
