package tests

import (
	"github.com/cucumber/godog"
	worktests "github.com/ready-to-release/eac/src/commands/impl/work/tests"
)

// InitializeWorkScenario registers work-specific step definitions
func InitializeWorkScenario(sc *godog.ScenarioContext) {
	// Create shared context for work tests
	workSharedCtx := &worktests.SharedContext{
		RepoRoot:      originalRepoRoot,
		CurrentBranch: "",
		ExitCode:      0,
		Output:        "",
	}

	// Register the work module's steps
	worktests.RegisterWorkSteps(sc, workSharedCtx)
}
