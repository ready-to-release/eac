// Package reset provides BDD step definitions for the commit reset subcommand.
//
// This file contains the step registration function for reset-specific steps.
package reset

import "github.com/cucumber/godog"

// InitializeScenario registers all reset-specific step definitions.
func InitializeScenario(sc *godog.ScenarioContext) {
	registerSetupSteps(sc)
	registerVerificationSteps(sc)
}

func registerSetupSteps(sc *godog.ScenarioContext) {
	// Git repository state setup
	sc.Step(`^I am in a git repository with at least two commits$`, iAmInAGitRepositoryWithAtLeastTwoCommits)
	sc.Step(`^I am in a git repository with only an initial commit$`, iAmInAGitRepositoryWithOnlyAnInitialCommit)
	sc.Step(`^the latest commit contains file changes$`, theLatestCommitContainsFileChanges)
	sc.Step(`^the latest commit added a new file "([^"]*)"$`, theLatestCommitAddedANewFile)
	sc.Step(`^I have uncommitted changes in the working directory$`, iHaveUncommittedChangesInTheWorkingDirectory)
	sc.Step(`^I am in a git repository in detached HEAD state$`, iAmInAGitRepositoryInDetachedHEADState)
	sc.Step(`^there are commits to reset$`, thereAreCommitsToReset)
	// Note: "I am not in a git repository" step is already registered in init_steps_test.go

	// Direct execution step for mock testing
	sc.Step(`^I run commit reset directly$`, iRunCommitResetDirectly)
}

func registerVerificationSteps(sc *godog.ScenarioContext) {
	// Reset operation verification
	sc.Step(`^the latest commit should be undone$`, theLatestCommitShouldBeUndone)
	sc.Step(`^the changes should remain staged$`, theChangesShouldRemainStaged)
	sc.Step(`^the working directory should be unchanged$`, theWorkingDirectoryShouldBeUnchanged)
	sc.Step(`^"([^"]*)" should be in the staging area$`, fileShouldBeInTheStagingArea)
	sc.Step(`^"([^"]*)" should exist in the working directory$`, fileShouldExistInTheWorkingDirectory)
	sc.Step(`^the uncommitted changes should be preserved$`, theUncommittedChangesShouldBePreserved)
	sc.Step(`^the reset should succeed$`, theResetShouldSucceed)
	sc.Step(`^I should see an error indicating not in a git repository$`, iShouldSeeAnErrorIndicatingNotInAGitRepository)
}
