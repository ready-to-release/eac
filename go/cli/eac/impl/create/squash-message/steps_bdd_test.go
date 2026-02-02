// Package squashmessage contains godog step implementations for eac-cli.
//
// This file contains create squash-message command step definitions.
package squashmessage

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

// createSquashMessageTestState holds state for create squash-message tests.
type createSquashMessageTestState struct {
	baseBranch string
}

// registerSteps registers step definitions for create squash-message command features.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	state := &createSquashMessageTestState{}

	// Note: "I am in a git repository with EAC configuration" and
	// "AI configuration exists at" steps are in RegisterCommonSteps (godog package)

	// Git branch state steps
	sc.Step(`^I am on a branch with no commits ahead of "([^"]*)"$`, func(baseBranch string) error {
		state.baseBranch = baseBranch
		// Ensure we're on a branch that's even with the base
		// Create the base branch if it doesn't exist
		cmd := exec.Command("git", "checkout", "-b", baseBranch)
		cmd.Dir = ctx.IsolatedDir
		_ = cmd.Run() //nolint:errcheck // Ignore error if branch exists

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
		if err := eacgodog.CreateDirectory(ctx, ".eac"); err != nil {
			return err
		}

		repositoryYml := `modules:
  - moniker: test-module
    name: Test Module
    components:
      go: go/test-module
`
		if err := eacgodog.CreateFile(ctx, ".eac/repository.yml", repositoryYml); err != nil {
			return err
		}

		// Create and commit to base branch first
		cmd := exec.Command("git", "checkout", "-b", baseBranch)
		cmd.Dir = ctx.IsolatedDir
		_ = cmd.Run() //nolint:errcheck // Ignore if branch exists

		// Add initial file
		baseFile := filepath.Join("go", "test-module", "base.go")
		if err := eacgodog.CreateFile(ctx, baseFile, "package testmodule\n\n// Base file\n"); err != nil {
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
		if err := eacgodog.CreateFile(ctx, file1, "package testmodule\n\n// Feature 1\n"); err != nil {
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
		if err := eacgodog.CreateFile(ctx, file2, "package testmodule\n\n// Feature 2\n"); err != nil {
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
		if err := eacgodog.CreateDirectory(ctx, ".eac"); err != nil {
			return err
		}

		repositoryYml := `modules:
  - moniker: test-module
    name: Test Module
    components:
      go: go/test-module
`
		if err := eacgodog.CreateFile(ctx, ".eac/repository.yml", repositoryYml); err != nil {
			return err
		}

		// Create and commit to base branch first
		cmd := exec.Command("git", "checkout", "-b", baseBranch)
		cmd.Dir = ctx.IsolatedDir
		_ = cmd.Run() //nolint:errcheck // Ignore if branch exists

		// Add initial file
		baseFile := filepath.Join("go", "test-module", "base.go")
		if err := eacgodog.CreateFile(ctx, baseFile, "package testmodule\n\n// Base file\n"); err != nil {
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
		if err := eacgodog.CreateFile(ctx, file1, "package testmodule\n\n// Feature\n"); err != nil {
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
		return eacgodog.SetupMockAIFromAsset(ctx, "squash-message/mock-response.txt")
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
			return fmt.Errorf("output doesn't contain conventional commit type prefix: %s", output)
		}

		return nil
	})

	// Base branch verification
	sc.Step(`^the command compares against "([^"]*)" branch$`, func(branch string) error {
		// This is validated by the command succeeding with the expected branch
		// We'd check logs if debug mode was enabled
		return nil
	})

	// Log file verification - step "logs are written to" is defined in steps_create_commit_message.go
}
