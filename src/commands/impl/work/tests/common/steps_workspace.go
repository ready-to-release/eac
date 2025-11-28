package common

import (
	"fmt"
	"strings"
)

// Workspace state steps

func IAmInTheMainWorkspace(ctx *TestContext) error {
	ctx.CurrentBranch = "main"
	ctx.MockGitOps.SetCurrentBranch("main")
	ctx.SharedCtx.CurrentBranch = "main"
	return nil
}

func IAmInTheMainWorkspaceOnBranch(ctx *TestContext, branch string) error {
	ctx.CurrentBranch = branch
	ctx.MockGitOps.SetCurrentBranch(branch)
	ctx.SharedCtx.CurrentBranch = branch
	return nil
}

func IAmInAWorkspace(ctx *TestContext) error {
	ctx.CurrentBranch = "feature/test"
	ctx.MockGitOps.SetCurrentBranch("feature/test")
	ctx.MockGitOps.AddWorktree("../work-feature-test", "feature/test")
	ctx.SharedCtx.CurrentBranch = "feature/test"
	return nil
}

func IAmInAWorkspaceFor(ctx *TestContext, branch string) error {
	ctx.CurrentBranch = branch
	ctx.MockGitOps.SetCurrentBranch(branch)
	workDir := strings.ReplaceAll(branch, "/", "-")
	ctx.MockGitOps.AddWorktree("../work-"+workDir, branch)
	ctx.SharedCtx.CurrentBranch = branch
	return nil
}

func AWorkspaceExistsFor(ctx *TestContext, branch string) error {
	workDir := strings.ReplaceAll(branch, "/", "-")
	ctx.MockGitOps.AddWorktree("../work-"+workDir, branch)
	return nil
}

func IHaveActiveWorkspaces(ctx *TestContext, count int) error {
	ctx.MockGitOps.AddWorktree(".", "main")
	for i := 1; i < count; i++ {
		branch := fmt.Sprintf("feature/branch-%d", i)
		workDir := fmt.Sprintf("../work-feature-branch-%d", i)
		ctx.MockGitOps.AddWorktree(workDir, branch)
	}
	return nil
}

func TheWorkspaceIsRemoved(ctx *TestContext) error {
	if ctx.MockGitOps.RemoveWorktreeCalled {
		return nil
	}
	return fmt.Errorf("expected RemoveWorktree to be called")
}

func TheWorkspaceForIsRemoved(ctx *TestContext, branch string) error {
	exists, _ := ctx.MockGitOps.WorktreeExists(branch)
	if exists {
		return fmt.Errorf("expected worktree for %s to be removed", branch)
	}
	return nil
}

func TheWorkspaceIsNotRemoved(ctx *TestContext) error {
	if ctx.MockGitOps.RemoveWorktreeCalled {
		return fmt.Errorf("expected RemoveWorktree NOT to be called")
	}
	return nil
}
