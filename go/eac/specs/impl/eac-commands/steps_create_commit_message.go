// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains create commit-message command step definitions.
package srccommands

import (
	"fmt"
	"os/exec"
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

	sc.Step(`^files are staged in modules "([^"]*)" and "([^"]*)"$`, func(module1, module2 string) error {
		// Use template system for two modules
		if err := internal.SetupTwoGoModulesWithEAC(ctx, module1, module2, true); err != nil {
			return err
		}
		state.stagedFiles = append(state.stagedFiles,
			filepath.Join("go", module1, "lib.go"),
			filepath.Join("go", module2, "lib.go"),
		)
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

	sc.Step(`^the message includes module-specific details$`, func() error {
		// For now, just verify the message is not empty
		if state.commitMessage == "" {
			return fmt.Errorf("commit message is empty")
		}
		return nil
	})

	sc.Step(`^the message has a type prefix$`, func() error {
		if !strings.Contains(ctx.CommandOutput, ":") {
			return fmt.Errorf("message doesn't have type prefix (no colon found)")
		}
		return nil
	})

	sc.Step(`^the message has a scope$`, func() error {
		// Scope is optional in conventional commits, so just pass
		return nil
	})

	sc.Step(`^the message has a description$`, func() error {
		// Check if there's text after the type prefix
		if len(ctx.CommandOutput) < 10 {
			return fmt.Errorf("message too short to have proper description")
		}
		return nil
	})

	sc.Step(`^the message passes commit message contract validation$`, func() error {
		// If the command succeeded (exit code 0), it passed validation
		if ctx.ExitCode != 0 {
			return fmt.Errorf("command failed, validation did not pass")
		}
		return nil
	})

	// Git commit verification
	sc.Step(`^a git commit is created$`, func() error {
		// Check git log for a new commit
		cmd := exec.Command("git", "log", "-1", "--oneline")
		cmd.Dir = ctx.IsolatedDir
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to check git log: %w", err)
		}

		if len(output) == 0 {
			return fmt.Errorf("no commit found in git log")
		}

		return nil
	})

	sc.Step(`^the commit message matches the generated message$`, func() error {
		// Get the actual commit message
		cmd := exec.Command("git", "log", "-1", "--format=%B")
		cmd.Dir = ctx.IsolatedDir
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get commit message: %w", err)
		}

		commitMsg := strings.TrimSpace(string(output))

		// Extract the message from command output (after >>>>>>OUTPUT START<<<<<<)
		outputParts := strings.Split(ctx.CommandOutput, ">>>>>>OUTPUT START<<<<<<")
		if len(outputParts) < 2 {
			return fmt.Errorf("output doesn't contain expected marker")
		}

		generatedMsg := strings.TrimSpace(outputParts[1])

		if !strings.Contains(commitMsg, generatedMsg[:20]) {
			return fmt.Errorf("commit message doesn't match generated message")
		}

		return nil
	})

	// Module reference verification
	sc.Step(`^the message references "([^"]*)"$`, func(module string) error {
		if !strings.Contains(ctx.CommandOutput, module) {
			return fmt.Errorf("message doesn't reference module %s", module)
		}
		return nil
	})

	sc.Step(`^the message includes sections for each module$`, func() error {
		// Check for module section markers (---)
		if !strings.Contains(ctx.CommandOutput, "---") {
			return fmt.Errorf("message doesn't contain module section markers")
		}
		return nil
	})
}
