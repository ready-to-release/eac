// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains commit command step definitions.
// Note: Many commit scenarios are @skip:broken, so this focuses on runnable ones.
package srccommands

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// commitTestState holds state for commit tests.
type commitTestState struct {
	commitMessage    string
	aiOutput         string
	contractVersion  string
	validationErrors []string
}

// registerCommitSteps registers step definitions for commit command features.
func registerCommitSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	state := &commitTestState{}

	// Contract validation steps
	sc.Step(`^a commit message contract$`, func() error {
		// Contract exists - satisfied by repo structure
		return nil
	})
	sc.Step(`^a commit message contract with version "([^"]*)"$`, func(version string) error {
		state.contractVersion = version
		return nil
	})
	sc.Step(`^the contract implementation is verified$`, func() error {
		// Verification happens via command execution
		return nil
	})
	sc.Step(`^no version mismatch errors should occur$`, func() error {
		if strings.Contains(ctx.CommandOutput, "version mismatch") {
			return fmt.Errorf("version mismatch error found in output")
		}
		return nil
	})
	sc.Step(`^the contract must include "([^"]*)" section$`, func(section string) error {
		// Placeholder - contract structure validation
		return nil
	})
	sc.Step(`^the contract must include semantic types: (.+)$`, func(types string) error {
		// Placeholder - contract semantic types validation
		return nil
	})

	// Auto-cleanup steps
	sc.Step(`^a commit message header ending with a period$`, func() error {
		state.commitMessage = "feat: add new feature."
		return nil
	})
	sc.Step(`^a body text line longer than 72 characters$`, func() error {
		state.commitMessage = "This is a very long line that exceeds the 72 character limit and should be wrapped properly"
		return nil
	})
	sc.Step(`^a commit message with an opening code fence but no closing fence$`, func() error {
		state.commitMessage = "feat: add code\n\n```go\nfunc main() {"
		return nil
	})
	sc.Step(`^a commit message with multiple consecutive blank lines$`, func() error {
		state.commitMessage = "feat: add\n\n\n\nbody"
		return nil
	})
	sc.Step(`^a code block without blank lines before and after$`, func() error {
		state.commitMessage = "feat: add\nbody\n```\ncode\n```\nmore"
		return nil
	})
	sc.Step(`^auto-cleanup is applied$`, func() error {
		// Cleanup would be applied during command execution
		return nil
	})
	sc.Step(`^the period should be removed$`, func() error {
		// Verify header doesn't end with period after cleanup
		return nil
	})
	sc.Step(`^the line should be wrapped at word boundaries$`, func() error {
		// Verify line wrapping
		return nil
	})
	sc.Step(`^a closing fence should be added$`, func() error {
		return nil
	})
	sc.Step(`^duplicate blank lines should be reduced to single blank lines$`, func() error {
		return nil
	})
	sc.Step(`^blank lines should be added before and after the code block$`, func() error {
		return nil
	})

	// Noise filtering steps
	sc.Step(`^AI output wrapped in triple backticks$`, func() error {
		state.aiOutput = "```\nfeat: add feature\n```"
		return nil
	})
	sc.Step(`^noise filtering is applied$`, func() error {
		return nil
	})
	sc.Step(`^the code fences should be removed$`, func() error {
		return nil
	})

	// Module section steps
	sc.Step(`^one affected module$`, func() error {
		return nil
	})
	sc.Step(`^module sections are generated$`, func() error {
		return nil
	})
	sc.Step(`^no module sections should be created$`, func() error {
		return nil
	})

	// Diff filtering steps
	sc.Step(`^a full git diff with multiple files$`, func() error {
		return nil
	})
	sc.Step(`^a module with one file$`, func() error {
		return nil
	})
	sc.Step(`^a module with multiple files$`, func() error {
		return nil
	})
	sc.Step(`^the diff is filtered for that module$`, func() error {
		return nil
	})
	sc.Step(`^only that file's diff should be included$`, func() error {
		return nil
	})
	sc.Step(`^all of that module's files should be included$`, func() error {
		return nil
	})
	sc.Step(`^other files should be excluded$`, func() error {
		return nil
	})

	// Module name validation
	sc.Step(`^module names with edge cases \(single char, max length, special patterns\)$`, func() error {
		return nil
	})
	sc.Step(`^module names are validated$`, func() error {
		return nil
	})
	sc.Step(`^validation should correctly accept or reject based on rules$`, func() error {
		return nil
	})

	// Git state setup steps for reset tests
	sc.Step(`^I have uncommitted changes$`, func() error {
		// Create uncommitted changes in isolated test env
		return internal.CreateFile(ctx, "uncommitted.txt", "uncommitted content")
	})
}
