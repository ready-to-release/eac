package common

// Changes and state steps

func IHaveUncommittedChanges(ctx *TestContext) error {
	ctx.MockGitOps.SetClean(false)
	return nil
}

func IHaveNoUncommittedChanges(ctx *TestContext) error {
	ctx.MockGitOps.SetClean(true)
	return nil
}

func UncommittedChangesAreDiscarded(ctx *TestContext) error {
	return TheWorkspaceIsRemoved(ctx)
}

func ThereAreCommitsAheadOfMain(ctx *TestContext, count int) error {
	ctx.MockGitOps.SetCommitCount(count)
	return nil
}

func ThereAreConflictsInTheFiles(ctx *TestContext) error {
	ctx.MockGitOps.SetConflicts([]string{"file1.go", "file2.go"})
	return nil
}
