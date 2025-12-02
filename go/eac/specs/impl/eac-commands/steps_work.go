// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains work command step definitions.
// Note: Most work scenarios are @skip:wip, so this is minimal for now.
package srccommands

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// registerWorkSteps registers step definitions for work command features.
func registerWorkSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Work-specific output verification (different patterns from common steps)
	sc.Step(`^I see "([^"]*)"$`, func(text string) error {
		return workOutputContains(ctx, text)
	})
	sc.Step(`^I see error "([^"]*)"$`, func(text string) error {
		return workOutputContains(ctx, text)
	})
	sc.Step(`^I see warning "([^"]*)"$`, func(text string) error {
		return workOutputContains(ctx, text)
	})
	sc.Step(`^I see suggestion "([^"]*)"$`, func(text string) error {
		return workOutputContains(ctx, text)
	})

	// Workspace state steps (for future when @skip:wip is removed)
	sc.Step(`^I am in a workspace$`, func() error {
		// Placeholder - needs git worktree setup
		return nil
	})
	sc.Step(`^I am in a workspace for "([^"]*)"$`, func(branch string) error {
		// Placeholder - needs git worktree setup
		return nil
	})
	sc.Step(`^I am in the main workspace$`, func() error {
		// Placeholder - verify we're in main worktree
		return nil
	})
	sc.Step(`^I am in the main workspace on branch "([^"]*)"$`, func(branch string) error {
		// Placeholder - verify branch
		return nil
	})
}

// workOutputContains checks if command output contains text.
func workOutputContains(ctx *internal.TestContext, text string) error {
	if !strings.Contains(ctx.CommandOutput, text) {
		return fmt.Errorf("expected output to contain %q, got:\n%s", text, ctx.CommandOutput)
	}
	return nil
}
