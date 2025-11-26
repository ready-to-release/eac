// Package reset provides BDD step definitions for the commit reset subcommand.
//
// This file contains Given step implementations for setting up test scenarios.
package reset

import (
	"github.com/ready-to-release/eac/src/commands/impl/commit"
)

// ============================================================================
// Git Repository State Setup Steps
// ============================================================================

// iAmInAGitRepositoryWithAtLeastTwoCommits sets up a repo with multiple commits.
func iAmInAGitRepositoryWithAtLeastTwoCommits() error {
	if MockOps == nil {
		MockOps = &MockResetGitOps{}
	}
	MockOps.HasParentCommitResult = true
	MockOps.SoftResetHeadCalled = false
	// Inject the mock into the commit package
	commit.SetResetGitOps(MockOps)
	return nil
}

// iAmInAGitRepositoryWithOnlyAnInitialCommit sets up a repo with only one commit.
func iAmInAGitRepositoryWithOnlyAnInitialCommit() error {
	if MockOps == nil {
		MockOps = &MockResetGitOps{}
	}
	MockOps.HasParentCommitResult = false
	MockOps.SoftResetHeadCalled = false
	// Inject the mock into the commit package
	commit.SetResetGitOps(MockOps)
	return nil
}

// theLatestCommitContainsFileChanges indicates the latest commit has file changes.
func theLatestCommitContainsFileChanges() error {
	// Mock is already set up - no additional setup needed
	// The soft reset will preserve these changes as staged
	return nil
}

// theLatestCommitAddedANewFile tracks which file was added in the latest commit.
func theLatestCommitAddedANewFile(filename string) error {
	// Track which file was "added" for verification
	if Ctx != nil {
		Ctx.ResetTestFile = filename
	}
	return nil
}

// iHaveUncommittedChangesInTheWorkingDirectory sets up uncommitted changes.
func iHaveUncommittedChangesInTheWorkingDirectory() error {
	// Mock handles this - working directory state is preserved by soft reset
	// The soft reset only affects the commit history, not the working directory
	return nil
}

// iAmInAGitRepositoryInDetachedHEADState sets up a detached HEAD state.
func iAmInAGitRepositoryInDetachedHEADState() error {
	if MockOps == nil {
		MockOps = &MockResetGitOps{}
	}
	// Detached HEAD can still have parent commits to reset to
	MockOps.HasParentCommitResult = true
	MockOps.SoftResetHeadCalled = false
	// Inject the mock into the commit package
	commit.SetResetGitOps(MockOps)
	return nil
}

// thereAreCommitsToReset ensures there are commits that can be reset.
func thereAreCommitsToReset() error {
	if MockOps == nil {
		MockOps = &MockResetGitOps{}
		commit.SetResetGitOps(MockOps)
	}
	MockOps.HasParentCommitResult = true
	return nil
}

// iRunCommitResetDirectly runs the commit reset command directly (not as subprocess).
// This allows the mock to be used and verified.
func iRunCommitResetDirectly() error {
	// Ensure mock is set up (but don't override existing mock configuration)
	if MockOps == nil {
		MockOps = &MockResetGitOps{HasParentCommitResult: true}
		commit.SetResetGitOps(MockOps)
	}

	// Run the command directly
	exitCode := commit.CommitReset()

	// Store result in context - initialize if nil
	if Ctx == nil {
		Ctx = &Context{}
	}
	Ctx.ExitCode = exitCode
	if exitCode == 0 {
		Ctx.CommandOutput = "Successfully reset latest commit. Changes are now staged."
	} else {
		Ctx.CommandOutput = "Error: no commits to reset (may be initial commit)"
	}
	return nil
}
