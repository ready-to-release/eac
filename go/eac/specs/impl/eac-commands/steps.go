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
	registerDocsSteps(sc, ctx)
	registerCreateCommitMessageSteps(sc, ctx)
	registerDesignSteps(sc, ctx)
	registerSpecsSteps(sc, ctx)
	registerRiskSteps(sc, ctx)  // New risk command (OSCAL-based)
	registerHelpSteps(sc, ctx)
	registerGitSetupSteps(sc, ctx)
	registerSecuritySteps(sc, ctx)
}
