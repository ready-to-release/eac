package common

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
)

// Common steps shared across all work subcommands

func IAmInAGitRepository(ctx *TestContext) error {
	ctx.SharedCtx.RepoRoot = "."
	return nil
}

func IAmNotInAGitRepository(ctx *TestContext) error {
	ctx.SharedCtx.RepoRoot = ""
	return nil
}

func IRunTheCommand(ctx *TestContext, command string) error {
	parts := strings.Fields(command)
	ctx.Args = parts
	return nil
}

func IRunCommand(ctx *TestContext, command string) error {
	ctx.Args = strings.Fields(command)
	internal.SetGitOps(ctx.MockGitOps)
	defer internal.SetGitOps(nil)
	ctx.SharedCtx.ExitCode = ctx.ExitCode
	ctx.SharedCtx.Output = ctx.Output
	return nil
}

func TheExitCodeIs(ctx *TestContext, expected int) error {
	actual := ctx.ExitCode
	if ctx.SharedCtx != nil && ctx.SharedCtx.ExitCode != 0 {
		actual = ctx.SharedCtx.ExitCode
	}
	if actual != expected {
		return fmt.Errorf("expected exit code %d, got %d", expected, actual)
	}
	return nil
}

func IShouldSee(ctx *TestContext, text string) error {
	output := ctx.Output
	if ctx.SharedCtx != nil && ctx.SharedCtx.Output != "" {
		output = ctx.SharedCtx.Output
	}
	if !strings.Contains(strings.ToLower(output), strings.ToLower(text)) {
		return fmt.Errorf("expected output to contain '%s', got: %s", text, output)
	}
	return nil
}

func ISee(ctx *TestContext, text string) error {
	return IShouldSee(ctx, text)
}

func ISeeError(ctx *TestContext, text string) error {
	if ctx.Error == nil {
		return fmt.Errorf("expected error containing '%s', got no error", text)
	}
	if !strings.Contains(strings.ToLower(ctx.Error.Error()), strings.ToLower(text)) {
		return fmt.Errorf("expected error to contain '%s', got: %s", text, ctx.Error.Error())
	}
	return nil
}

func ISeeWarning(ctx *TestContext, text string) error {
	return IShouldSee(ctx, text)
}

func ISeeSuggestion(ctx *TestContext, text string) error {
	return IShouldSee(ctx, text)
}

func IShouldSeeOneOf(ctx *TestContext, option1, option2, option3 string) error {
	output := strings.ToLower(ctx.Output)
	if strings.Contains(output, strings.ToLower(option1)) ||
		strings.Contains(output, strings.ToLower(option2)) ||
		strings.Contains(output, strings.ToLower(option3)) {
		return nil
	}
	return fmt.Errorf("expected output to contain one of '%s', '%s', or '%s', got: %s",
		option1, option2, option3, ctx.Output)
}
