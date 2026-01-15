// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains create commit-message command step definitions.
package srccommands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// createCommitMessageTestState holds state for create commit-message tests.
type createCommitMessageTestState struct {
	commitMessage string
	stagedFiles   []string
}

// registerCreateCommitMessageSteps registers step definitions for create commit-message command features.
func registerCreateCommitMessageSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	state := &createCommitMessageTestState{}

	// Note: "I am in a git repository with EAC configuration" and
	// "AI configuration exists at" steps are defined in steps_help.go

	// Git state steps
	sc.Step(`^no files are staged$`, func() error {
		return internal.UnstageAll(ctx)
	})

	sc.Step(`^files are staged with changes$`, func() error {
		// Use pre-built test repository layout for realistic test fixture
		if err := internal.CopyTestLayout(ctx, "single-go-module", true); err != nil {
			return err
		}
		state.stagedFiles = append(state.stagedFiles, filepath.Join("go", "test-module", "lib.go"))
		return nil
	})

	sc.Step(`^files are staged in module "([^"]*)"$`, func(module string) error {
		// Use template system for consistent EAC config
		if err := internal.SetupGoModuleWithEAC(ctx, module, true); err != nil {
			return err
		}
		state.stagedFiles = append(state.stagedFiles, filepath.Join("go", module, "lib.go"))
		return nil
	})


	// Mock AI configuration
	sc.Step(`^the mock AI is configured to return a valid commit message$`, func() error {
		return internal.SetupMockAIFromAsset(ctx, "commit-message/mock-response.txt")
	})

	// Output verification steps
	sc.Step(`^stdout contains a conventional commit message$`, func() error {
		output := ctx.CommandOutput

		// Check for conventional commit format (type: description or type(scope): description)
		hasConventionalFormat := strings.Contains(output, "feat:") ||
			strings.Contains(output, "feat(") ||
			strings.Contains(output, "fix:") ||
			strings.Contains(output, "fix(") ||
			strings.Contains(output, "docs:") ||
			strings.Contains(output, "docs(") ||
			strings.Contains(output, "refactor:") ||
			strings.Contains(output, "refactor(")

		if !hasConventionalFormat {
			return fmt.Errorf("output doesn't contain conventional commit type prefix")
		}

		state.commitMessage = output
		return nil
	})


	// Module reference verification
	sc.Step(`^the message references "([^"]*)"$`, func(module string) error {
		if !strings.Contains(ctx.CommandOutput, module) {
			return fmt.Errorf("message doesn't reference module %s", module)
		}
		return nil
	})
}
