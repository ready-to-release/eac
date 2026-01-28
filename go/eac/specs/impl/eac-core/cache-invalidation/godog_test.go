// Package cacheinvalidation contains godog step implementations for specs/eac-core/cache-invalidation.
//
// This package tests the cache invalidation system for detecting module rebuild requirements.
package cacheinvalidation

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

func TestCacheInvalidationFeatures(t *testing.T) {
	cfg := internal.RunnerConfig{
		SpecsPath:         "../../../../../../specs/eac-core/cache-invalidation",
		DefaultReportName: "cucumber-cache-invalidation",
		RegisterSteps:     RegisterSteps,
	}

	opts := internal.BuildOptions(cfg.SpecsPath, cfg.DefaultReportName, t)
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			ctx := internal.NewTestContext()
			internal.RegisterCommonSteps(sc, ctx)
			cfg.RegisterSteps(sc, ctx)

			sc.After(func(c context.Context, sc *godog.Scenario, err error) (context.Context, error) {
				cleanupCacheContext(ctx)
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
