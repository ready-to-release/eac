// Package repository contains godog step implementations for specs/repository.
package repository

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
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
	os.Exit(m.Run())
}
