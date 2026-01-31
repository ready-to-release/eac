// Package godog provides shared infrastructure for godog BDD tests.
//
// This package is the godog-specific layer on top of eac-core/testing/,
// providing:
//
//   - TestContext: Wraps SharedTestContext with godog-specific features
//   - TestCache: Repository data caching for test performance
//   - RegisterCommonSteps: Generic godog step definitions
//   - Helpers: Step implementation building blocks
//   - Fixtures: Complete test environment setup
//   - Runner: Godog runner configuration
//   - Templates: EAC configuration templates for tests
//
// # Architecture
//
// This package sits between eac-core/testing/ (shared test infrastructure)
// and module-specific step implementations:
//
//	eac-core/testing/         ← Shared test infrastructure (unit tests + godog)
//	  - SharedTestContext
//	  - TestIsolation
//	  - FixturePool
//	  - mocks/
//
//	       ↑ imports
//
//	eac-godog/                ← This package (godog-specific layer)
//	  - TestContext
//	  - TestCache
//	  - RegisterCommonSteps()
//	  - helpers, fixtures, templates
//
//	       ↑ imports
//
//	<module>/specs/           ← Module step implementations
//	  - godog_test.go
//	  - steps_*_test.go
//
// # Usage
//
// Each module's specs/ directory creates a godog_test.go that:
//  1. Imports this package
//  2. Creates a RunnerConfig
//  3. Registers module-specific steps
//
// Example godog_test.go:
//
//	func TestFeatures(t *testing.T) {
//	    cfg := godog.RunnerConfig{
//	        SpecsPath:         "../../../../specs/eac-core/cache-invalidation",
//	        DefaultReportName: "eac-core-cache",
//	        RegisterSteps:     registerSteps,
//	    }
//	    suite := godog.TestSuite{
//	        ScenarioInitializer: godog.CreateScenarioInitializer(cfg),
//	        Options:             godog.BuildOptions(cfg.SpecsPath, cfg.DefaultReportName, t),
//	    }
//	    if suite.Run() != 0 {
//	        t.Fatal("non-zero status returned, failed to run feature tests")
//	    }
//	}
//
//	func registerSteps(sc *cucumber.ScenarioContext, ctx *godog.TestContext) {
//	    // Register module-specific steps here
//	}
package godog
