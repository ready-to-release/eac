// Package srccommands contains godog step implementations for specs/src-commands.
package srccommands

import (
	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/specs/internal"
)

// RegisterSteps registers all src-commands step definitions.
// This is the main entry point called by the test runner.
// Note: Common steps are already registered by the runner in internal.CreateScenarioInitializer.
func RegisterSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Register feature-specific steps for each command domain
	registerTemplatesSteps(sc, ctx)
	registerInitSteps(sc, ctx)
	registerWorkSteps(sc, ctx)
	registerDocsSteps(sc, ctx)
	registerCommitSteps(sc, ctx)
	registerDesignSteps(sc, ctx)
	registerSpecsSteps(sc, ctx)
	registerRisksSteps(sc, ctx)
	registerHelpSteps(sc, ctx)
	registerGitSetupSteps(sc, ctx)
}
