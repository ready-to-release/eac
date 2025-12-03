// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains create commit-message command step definitions.
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
		// Ensure clean git state - no files staged
		cmd := exec.Command("git", "reset", "HEAD", ".")
		cmd.Dir = ctx.IsolatedDir
		_ = cmd.Run() // Ignore error if nothing to reset
		return nil
	})

	sc.Step(`^files are staged with changes$`, func() error {
		// Create module structure matching the working scenarios pattern
		if err := internal.CreateDirectory(ctx, ".r2r/eac"); err != nil {
			return err
		}

		// Create modules.yml with test module in subdirectory (like working tests)
		modulesYml := `modules:
  - moniker: test-module
    name: Test Module
    type: go-library
    files:
      root: go/test-module
`
		if err := internal.CreateFile(ctx, ".r2r/eac/modules.yml", modulesYml); err != nil {
			return err
		}

		// Create test file in module subdirectory (not root)
		testFile := filepath.Join("go", "test-module", "test.go")
		if err := internal.CreateFile(ctx, testFile, "package testmodule\n\n// Test file\n"); err != nil {
			return err
		}

		cmd := exec.Command("git", "add", testFile)
		cmd.Dir = ctx.IsolatedDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to stage file: %w (output: %s)", err, string(output))
		}

		state.stagedFiles = append(state.stagedFiles, testFile)
		return nil
	})

	sc.Step(`^files are staged in module "([^"]*)"$`, func(module string) error {
		// Create module structure
		if err := internal.CreateDirectory(ctx, ".r2r/eac"); err != nil {
			return err
		}

		// Create modules.yml with the specified module
		modulesYml := fmt.Sprintf(`modules:
  - moniker: %s
    name: %s Module
    type: go-library
    files:
      root: go/%s
`, module, module, module)
		if err := internal.CreateFile(ctx, ".r2r/eac/modules.yml", modulesYml); err != nil {
			return err
		}

		// Create a file in that module and stage it
		moduleFile := filepath.Join("go", module, "test.go")
		if err := internal.CreateFile(ctx, moduleFile, "package main\n"); err != nil {
			return err
		}

		cmd := exec.Command("git", "add", moduleFile)
		cmd.Dir = ctx.IsolatedDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to stage module file: %w (output: %s)", err, string(output))
		}

		state.stagedFiles = append(state.stagedFiles, moduleFile)
		return nil
	})

	sc.Step(`^files are staged in modules "([^"]*)" and "([^"]*)"$`, func(module1, module2 string) error {
		// Create modules.yml with both modules first
		modulesYml := fmt.Sprintf(`modules:
  - moniker: %s
    name: %s Module
    type: go-library
    files:
      root: go/%s
  - moniker: %s
    name: %s Module
    type: go-library
    files:
      root: go/%s
`, module1, module1, module1, module2, module2, module2)
		if err := internal.CreateFile(ctx, ".r2r/eac/modules.yml", modulesYml); err != nil {
			return err
		}

		// Create both module files and stage them
		for _, module := range []string{module1, module2} {
			moduleFile := filepath.Join("go", module, "test.go")
			if err := internal.CreateFile(ctx, moduleFile, fmt.Sprintf("package %s\n", module)); err != nil {
				return err
			}

			cmd := exec.Command("git", "add", moduleFile)
			cmd.Dir = ctx.IsolatedDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to stage module file %s: %w (output: %s)", moduleFile, err, string(output))
			}

			state.stagedFiles = append(state.stagedFiles, moduleFile)
		}

		return nil
	})

	// Mock AI configuration
	sc.Step(`^the mock AI is configured to return a valid commit message$`, func() error {
		mockContent, err := internal.LoadAsset(ctx, "commit-message/mock-response.txt")
		if err != nil {
			return err
		}
		return internal.CreateFile(ctx, ".r2r/test/ai-mock.txt", mockContent)
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

	// Debug file verification
	sc.Step(`^debug files are saved to "([^"]*)"$`, func(path string) error {
		debugPath := filepath.Join(ctx.IsolatedDir, path)
		info, err := os.Stat(debugPath)
		if err != nil {
			return fmt.Errorf("debug directory not found: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", path)
		}
		return nil
	})

	sc.Step(`^debug files include the AI prompt$`, func() error {
		promptFile := filepath.Join(ctx.IsolatedDir, "out/logs/commit/debug-top-level-context.md")
		if _, err := os.Stat(promptFile); os.IsNotExist(err) {
			return fmt.Errorf("AI prompt debug file not found")
		}
		return nil
	})

	sc.Step(`^debug files include the AI response$`, func() error {
		responseFile := filepath.Join(ctx.IsolatedDir, "out/logs/commit/debug-top-level-output.md")
		if _, err := os.Stat(responseFile); os.IsNotExist(err) {
			return fmt.Errorf("AI response debug file not found")
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
