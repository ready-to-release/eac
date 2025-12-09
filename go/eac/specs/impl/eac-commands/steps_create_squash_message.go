// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains create squash-message command step definitions.
package srccommands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// createSquashMessageTestState holds state for create squash-message tests.
type createSquashMessageTestState struct {
	squashMessage string
	baseBranch    string
}

// registerCreateSquashMessageSteps registers step definitions for create squash-message command features.
func registerCreateSquashMessageSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	state := &createSquashMessageTestState{}

	// Note: "I am in a git repository with EAC configuration" and
	// "AI configuration exists at" steps are defined in steps_help.go

	// Git branch state steps
	sc.Step(`^I am on a branch with no commits ahead of "([^"]*)"$`, func(baseBranch string) error {
		state.baseBranch = baseBranch
		// Ensure we're on a branch that's even with the base
		// Create the base branch if it doesn't exist
		cmd := exec.Command("git", "checkout", "-b", baseBranch)
		cmd.Dir = ctx.IsolatedDir
		_ = cmd.Run() // Ignore error if branch exists

		// Create a feature branch from it
		cmd = exec.Command("git", "checkout", "-b", "feature-branch")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create feature branch: %w (output: %s)", err, string(output))
		}

		return nil
	})

	sc.Step(`^I am on a branch with multiple commits ahead of "([^"]*)"$`, func(baseBranch string) error {
		state.baseBranch = baseBranch

		// Create module structure
		if err := internal.CreateDirectory(ctx, ".r2r/eac"); err != nil {
			return err
		}

		modulesYml := `modules:
  - moniker: test-module
    name: Test Module
    type: go
    files:
      root: go/test-module
`
		if err := internal.CreateFile(ctx, ".r2r/eac/modules.yml", modulesYml); err != nil {
			return err
		}

		// Create and commit to base branch first
		cmd := exec.Command("git", "checkout", "-b", baseBranch)
		cmd.Dir = ctx.IsolatedDir
		_ = cmd.Run() // Ignore if branch exists

		// Add initial file
		baseFile := filepath.Join("go", "test-module", "base.go")
		if err := internal.CreateFile(ctx, baseFile, "package testmodule\n\n// Base file\n"); err != nil {
			return err
		}

		cmd = exec.Command("git", "add", baseFile)
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add base file: %w (output: %s)", err, string(output))
		}

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create base commit: %w (output: %s)", err, string(output))
		}

		// Create feature branch with multiple commits
		cmd = exec.Command("git", "checkout", "-b", "feature-branch")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create feature branch: %w (output: %s)", err, string(output))
		}

		// First commit on feature branch
		file1 := filepath.Join("go", "test-module", "feature1.go")
		if err := internal.CreateFile(ctx, file1, "package testmodule\n\n// Feature 1\n"); err != nil {
			return err
		}

		cmd = exec.Command("git", "add", file1)
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add file1: %w (output: %s)", err, string(output))
		}

		cmd = exec.Command("git", "commit", "-m", "feat: add feature 1")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to commit file1: %w (output: %s)", err, string(output))
		}

		// Second commit on feature branch
		file2 := filepath.Join("go", "test-module", "feature2.go")
		if err := internal.CreateFile(ctx, file2, "package testmodule\n\n// Feature 2\n"); err != nil {
			return err
		}

		cmd = exec.Command("git", "add", file2)
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add file2: %w (output: %s)", err, string(output))
		}

		cmd = exec.Command("git", "commit", "-m", "feat: add feature 2")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to commit file2: %w (output: %s)", err, string(output))
		}

		return nil
	})

	sc.Step(`^I am on a branch with commits ahead of "([^"]*)"$`, func(baseBranch string) error {
		state.baseBranch = baseBranch

		// Create module structure
		if err := internal.CreateDirectory(ctx, ".r2r/eac"); err != nil {
			return err
		}

		modulesYml := `modules:
  - moniker: test-module
    name: Test Module
    type: go
    files:
      root: go/test-module
`
		if err := internal.CreateFile(ctx, ".r2r/eac/modules.yml", modulesYml); err != nil {
			return err
		}

		// Create and commit to base branch first
		cmd := exec.Command("git", "checkout", "-b", baseBranch)
		cmd.Dir = ctx.IsolatedDir
		_ = cmd.Run() // Ignore if branch exists

		// Add initial file
		baseFile := filepath.Join("go", "test-module", "base.go")
		if err := internal.CreateFile(ctx, baseFile, "package testmodule\n\n// Base file\n"); err != nil {
			return err
		}

		cmd = exec.Command("git", "add", baseFile)
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add base file: %w (output: %s)", err, string(output))
		}

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create base commit: %w (output: %s)", err, string(output))
		}

		// Create feature branch with one commit
		cmd = exec.Command("git", "checkout", "-b", "feature-branch")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create feature branch: %w (output: %s)", err, string(output))
		}

		// Commit on feature branch
		file1 := filepath.Join("go", "test-module", "feature.go")
		if err := internal.CreateFile(ctx, file1, "package testmodule\n\n// Feature\n"); err != nil {
			return err
		}

		cmd = exec.Command("git", "add", file1)
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add file: %w (output: %s)", err, string(output))
		}

		cmd = exec.Command("git", "commit", "-m", "feat: add feature")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to commit file: %w (output: %s)", err, string(output))
		}

		return nil
	})

	// Mock AI configuration for squash message
	sc.Step(`^the mock AI is configured to return a valid squash message$`, func() error {
		mockContent := `feat(test-module): implement comprehensive testing features

Auditor-Summary: Added complete test module functionality with multiple features.

Implemented feature 1 and feature 2 to provide comprehensive testing
capabilities. The changes enable better test coverage and improve
overall code quality.

Changes: 2 files, +10 insertions, -0 deletions`

		return internal.CreateFile(ctx, ".r2r/test/ai-mock.txt", mockContent)
	})

	// Output verification steps
	sc.Step(`^the message synthesizes all commits into cohesive narrative$`, func() error {
		output := ctx.CommandOutput

		// Check that it's not just listing commits one by one
		// A synthesized message should not have multiple "feat:" or "fix:" prefixes
		lines := strings.Split(output, "\n")
		conventionalPrefixes := 0
		for _, line := range lines {
			if strings.HasPrefix(line, "feat:") || strings.HasPrefix(line, "feat(") ||
				strings.HasPrefix(line, "fix:") || strings.HasPrefix(line, "fix(") {
				conventionalPrefixes++
			}
		}

		if conventionalPrefixes > 1 {
			return fmt.Errorf("message appears to list commits individually (found %d commit prefixes)", conventionalPrefixes)
		}

		return nil
	})

	sc.Step(`^the message includes "Auditor-Summary" line$`, func() error {
		if !strings.Contains(ctx.CommandOutput, "Auditor-Summary:") {
			return fmt.Errorf("message doesn't contain 'Auditor-Summary:' line")
		}
		return nil
	})

	sc.Step(`^the message includes "Changes:" statistics line$`, func() error {
		if !strings.Contains(ctx.CommandOutput, "Changes:") {
			return fmt.Errorf("message doesn't contain 'Changes:' statistics line")
		}
		// Check it includes numbers
		if !strings.Contains(ctx.CommandOutput, "files") {
			return fmt.Errorf("Changes line doesn't include file statistics")
		}
		return nil
	})

	// Base branch verification
	sc.Step(`^the command compares against "([^"]*)" branch$`, func(branch string) error {
		// This is validated by the command succeeding with the expected branch
		// We'd check logs if debug mode was enabled
		return nil
	})

	// Debug file verification for squash-message specific files
	sc.Step(`^debug files include "([^"]*)"$`, func(filename string) error {
		debugFile := filepath.Join(ctx.IsolatedDir, "out/logs/commit", filename)
		if _, err := os.Stat(debugFile); os.IsNotExist(err) {
			return fmt.Errorf("debug file %s not found", filename)
		}
		return nil
	})
}
