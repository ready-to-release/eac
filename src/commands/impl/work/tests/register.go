package tests

import (
	"context"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/commands/impl/work/tests/common"
)

// RegisterWorkSteps registers all work command step definitions
func RegisterWorkSteps(sc *godog.ScenarioContext, sharedCtx *SharedContext) {
	// Convert to common.SharedContext
	commonSharedCtx := &common.SharedContext{
		RepoRoot:      sharedCtx.RepoRoot,
		CurrentBranch: sharedCtx.CurrentBranch,
		ExitCode:      sharedCtx.ExitCode,
		Output:        sharedCtx.Output,
	}

	testCtx := &common.TestContext{
		SharedCtx: commonSharedCtx,
	}

	// Reset context before each scenario
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		testCtx.Reset()
		return ctx, nil
	})

	// Common steps
	sc.Step(`^I am in a git repository$`, func() error { return common.IAmInAGitRepository(testCtx) })
	sc.Step(`^I am not in a git repository$`, func() error { return common.IAmNotInAGitRepository(testCtx) })
	sc.Step(`^I run the command "([^"]*)"$`, func(cmd string) error { return common.IRunTheCommand(testCtx, cmd) })
	sc.Step(`^I run "([^"]*)"$`, func(cmd string) error { return common.IRunCommand(testCtx, cmd) })
	sc.Step(`^the exit code is (\d+)$`, func(code int) error { return common.TheExitCodeIs(testCtx, code) })
	sc.Step(`^I should see "([^"]*)"$`, func(text string) error { return common.IShouldSee(testCtx, text) })
	sc.Step(`^I see "([^"]*)"$`, func(text string) error { return common.ISee(testCtx, text) })
	sc.Step(`^I see error "([^"]*)"$`, func(text string) error { return common.ISeeError(testCtx, text) })
	sc.Step(`^I see warning "([^"]*)"$`, func(text string) error { return common.ISeeWarning(testCtx, text) })
	sc.Step(`^I see suggestion "([^"]*)"$`, func(text string) error { return common.ISeeSuggestion(testCtx, text) })
	sc.Step(`^I should see "([^"]*)" or "([^"]*)" or "([^"]*)"$`, func(o1, o2, o3 string) error { return common.IShouldSeeOneOf(testCtx, o1, o2, o3) })

	// Workspace state steps
	sc.Step(`^I am in the main workspace$`, func() error { return common.IAmInTheMainWorkspace(testCtx) })
	sc.Step(`^I am in the main workspace on branch "([^"]*)"$`, func(branch string) error { return common.IAmInTheMainWorkspaceOnBranch(testCtx, branch) })
	sc.Step(`^I am in a workspace$`, func() error { return common.IAmInAWorkspace(testCtx) })
	sc.Step(`^I am in a workspace for "([^"]*)"$`, func(branch string) error { return common.IAmInAWorkspaceFor(testCtx, branch) })
	sc.Step(`^a workspace exists for "([^"]*)"$`, func(branch string) error { return common.AWorkspaceExistsFor(testCtx, branch) })
	sc.Step(`^I have (\d+) active workspaces?$`, func(count int) error { return common.IHaveActiveWorkspaces(testCtx, count) })
	sc.Step(`^the workspace is removed$`, func() error { return common.TheWorkspaceIsRemoved(testCtx) })
	sc.Step(`^the workspace for "([^"]*)" is removed$`, func(branch string) error { return common.TheWorkspaceForIsRemoved(testCtx, branch) })
	sc.Step(`^the workspace is NOT removed$`, func() error { return common.TheWorkspaceIsNotRemoved(testCtx) })

	// Branch state steps
	sc.Step(`^the branch "([^"]*)" exists$`, func(branch string) error { return common.TheBranchExists(testCtx, branch) })
	sc.Step(`^the branch exists on remote$`, func() error { return common.TheBranchExistsOnRemote(testCtx) })
	sc.Step(`^the branch "([^"]*)" does not exist$`, func(branch string) error { return common.TheBranchDoesNotExist(testCtx, branch) })
	sc.Step(`^the local branch "([^"]*)" is deleted$`, func(branch string) error { return common.TheLocalBranchIsDeleted(testCtx, branch) })
	sc.Step(`^the local branch is deleted$`, func() error { return common.TheLocalBranchIsDeleted(testCtx) })
	sc.Step(`^the branch "([^"]*)" is deleted$`, func(branch string) error { return common.TheBranchIsDeleted(testCtx, branch) })
	sc.Step(`^the branch "([^"]*)" is NOT deleted$`, func(branch string) error { return common.TheBranchIsNotDeleted(testCtx, branch) })
	sc.Step(`^the remote branch is deleted$`, func() error { return common.TheRemoteBranchIsDeleted(testCtx) })
	sc.Step(`^the remote branch is NOT deleted$`, func() error { return common.TheRemoteBranchIsNotDeleted(testCtx) })
	sc.Step(`^I am switched to main branch$`, func() error { return common.IAmSwitchedToMainBranch(testCtx) })

	// Changes and state steps
	sc.Step(`^I have uncommitted changes$`, func() error { return common.IHaveUncommittedChanges(testCtx) })
	sc.Step(`^I have no uncommitted changes$`, func() error { return common.IHaveNoUncommittedChanges(testCtx) })
	sc.Step(`^uncommitted changes are discarded$`, func() error { return common.UncommittedChangesAreDiscarded(testCtx) })
	sc.Step(`^there are (\d+) commits? ahead of main$`, func(count int) error { return common.ThereAreCommitsAheadOfMain(testCtx, count) })
	sc.Step(`^there are conflicts in the files?$`, func() error { return common.ThereAreConflictsInTheFiles(testCtx) })

	// Debug mode steps
	sc.Step(`^debug logs are written to "([^"]*)"$`, func(path string) error { return common.DebugLogsAreWrittenTo(testCtx, path) })
	sc.Step(`^debug logs contain workspace details$`, func() error { return common.DebugLogsContainWorkspaceDetails(testCtx) })
	sc.Step(`^debug logs contain workspace removal details$`, func() error { return common.DebugLogsContainWorkspaceRemovalDetails(testCtx) })
	sc.Step(`^debug logs contain branch information$`, func() error { return common.DebugLogsContainBranchInformation(testCtx) })
	sc.Step(`^debug logs contain rebase details$`, func() error { return common.DebugLogsContainRebaseDetails(testCtx) })
	sc.Step(`^debug logs contain merge details$`, func() error { return common.DebugLogsContainMergeDetails(testCtx) })
}
