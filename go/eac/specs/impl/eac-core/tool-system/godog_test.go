// Package toolsystem contains godog step implementations for specs/eac-core/tool-system.
//
// This package tests the tool resolution system - how component types are mapped
// to specific tools for build, test, lint, and scan operations.
// Tests run in isolated directories to avoid affecting the real repository.
package toolsystem

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

func TestToolSystemFeatures(t *testing.T) {
	cfg := internal.RunnerConfig{
		SpecsPath:         "../../../../../../specs/eac-core/tool-system",
		DefaultReportName: "cucumber-tool-system",
		RegisterSteps:     RegisterSteps,
	}

	opts := internal.BuildOptions(cfg.SpecsPath, cfg.DefaultReportName, t)
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			ctx := internal.NewTestContext()
			internal.RegisterCommonSteps(sc, ctx)
			cfg.RegisterSteps(sc, ctx)

			sc.After(func(c context.Context, sc *godog.Scenario, err error) (context.Context, error) {
				cleanupTestState()
				return c, nil
			})
		},
		Options: opts,
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
