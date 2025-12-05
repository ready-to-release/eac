// Package handlersconfig contains godog step implementations for specs/eac-core/handlers-config.
//
// This package tests the handlers configuration system including:
// - Handler loading and validation
// - Dispatch rule evaluation
// - Handler-specific configuration (Go, Docker, MkDocs, NPM)
// - Nil safety and edge cases
package handlersconfig

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

func TestHandlersConfigFeatures(t *testing.T) {
	cfg := internal.RunnerConfig{
		SpecsPath:         "../../../../../../specs/eac-core/handlers-config",
		DefaultReportName: "cucumber-handlers-config",
		RegisterSteps:     RegisterSteps,
	}

	opts := internal.BuildOptions(cfg.SpecsPath, cfg.DefaultReportName, t)
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			ctx := internal.NewTestContext()
			internal.RegisterCommonSteps(sc, ctx)
			cfg.RegisterSteps(sc, ctx)

			sc.After(func(c context.Context, sc *godog.Scenario, err error) (context.Context, error) {
				cleanupTestContext()
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
