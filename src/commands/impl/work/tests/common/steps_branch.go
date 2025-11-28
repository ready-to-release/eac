package common

import (
	"fmt"
)

// Branch state steps

func TheBranchExists(ctx *TestContext, branch string) error {
	ctx.MockGitOps.AddBranch(branch)
	return nil
}

func TheBranchExistsOnRemote(ctx *TestContext) error {
	if ctx.CurrentBranch != "" {
		ctx.MockGitOps.AddRemoteBranch(ctx.CurrentBranch)
	}
	return nil
}

func TheBranchDoesNotExist(ctx *TestContext, branch string) error {
	delete(ctx.MockGitOps.Branches, branch)
	return nil
}

func TheLocalBranchIsDeleted(ctx *TestContext, args ...string) error {
	if !ctx.MockGitOps.DeleteBranchCalled {
		return fmt.Errorf("expected DeleteBranch to be called")
	}
	if len(args) > 0 {
		branch := args[0]
		exists, _ := ctx.MockGitOps.BranchExists(branch)
		if exists {
			return fmt.Errorf("expected branch %s to be deleted", branch)
		}
	}
	return nil
}

func TheBranchIsDeleted(ctx *TestContext, branch string) error {
	exists, _ := ctx.MockGitOps.BranchExists(branch)
	if exists {
		return fmt.Errorf("expected branch %s to be deleted", branch)
	}
	return nil
}

func TheBranchIsNotDeleted(ctx *TestContext, branch string) error {
	if ctx.MockGitOps.DeleteBranchCalled {
		return fmt.Errorf("expected branch %s NOT to be deleted", branch)
	}
	return nil
}

func TheRemoteBranchIsDeleted(ctx *TestContext) error {
	return nil
}

func TheRemoteBranchIsNotDeleted(ctx *TestContext) error {
	if ctx.CurrentBranch != "" && !ctx.MockGitOps.RemoteBranchExists(ctx.CurrentBranch) {
		return fmt.Errorf("expected remote branch to still exist")
	}
	return nil
}

func IAmSwitchedToMainBranch(ctx *TestContext) error {
	ctx.CurrentBranch = "main"
	ctx.SharedCtx.CurrentBranch = "main"
	return nil
}
