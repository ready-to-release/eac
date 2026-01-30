// Package srccommands contains godog step implementations for specs/eac-commands.
package srccommands

import (
	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// RegisterSteps registers all eac-commands step definitions.
// This is the main entry point called by the test runner.
// Note: Common steps are already registered by the runner in internal.CreateScenarioInitializer.
func RegisterSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Register feature-specific steps for each command domain
	registerTemplatesSteps(sc, ctx)
	registerInitSteps(sc, ctx)
	registerWorkSteps(sc, ctx)
	registerWorkspacesSteps(sc, ctx)
	registerDocsSteps(sc, ctx)
	registerCreateCommitMessageSteps(sc, ctx)
	registerCreateSquashMessageSteps(sc, ctx)
	registerDesignSteps(sc, ctx)
	registerSpecsSteps(sc, ctx)
	registerRiskSteps(sc, ctx)
	registerHelpSteps(sc, ctx)
	registerGitSetupSteps(sc, ctx)
	registerBooksSteps(sc, ctx)
	registerTestResultsSteps(sc, ctx)
	registerDrawioSteps(sc, ctx)
	registerPrunePackagesSteps(sc, ctx)
	registerPipelineSteps(sc, ctx)
	registerPipelineStatusSteps(sc, ctx)
	registerReleaseSteps(sc, ctx)
}
