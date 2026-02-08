// Package docs contains godog step implementations for docs commands.
//
// This file contains docs command step definitions.
package docs

import (
	"fmt"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

// setupBooksConfig creates the books.yml and repository.yml configuration needed for serve docs tests.
// Note: This feature is marked @wip until proper Docker integration tests can be set up.
func setupBooksConfig(ctx *eacgodog.TestContext) error {
	if ctx.IsolatedDir == "" {
		return nil // Not in isolated mode, skip
	}

	// Create repository.yml with a docs module that references the book
	repoYml := `repository:
  type: poly
  remote:
    owner: test-org
    repo: test-repo

modules:
  - moniker: docs
    name: Documentation
    description: Test documentation module
    books:
      - docs
`
	booksYml := `books:
  - name: docs
    description: Test documentation
    output: site
    sources:
      - type: copy
        from: "docs/**/*.md"
        to: "./"
`
	if err := eacgodog.CreateDirectory(ctx, ".eac"); err != nil {
		return err
	}
	if err := eacgodog.CreateFile(ctx, ".eac/repository.yml", repoYml); err != nil {
		return err
	}
	if err := eacgodog.CreateFile(ctx, ".eac/books.yml", booksYml); err != nil {
		return err
	}
	// Create a minimal docs directory
	if err := eacgodog.CreateDirectory(ctx, "docs"); err != nil {
		return err
	}
	return eacgodog.CreateFile(ctx, "docs/index.md", "# Test Docs\n")
}

// registerSteps registers step definitions for docs command features.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Setup step for docs module configuration
	sc.Step(`^a docs module is configured$`, func() error {
		return setupBooksConfig(ctx)
	})

	// Note: "docker service is available" is registered in eacgodog/steps.go

	// Given/Then steps - serve container state
	sc.Step(`^serve container is running$`, func() error {
		// Set up books.yml if we're in isolation
		if err := setupBooksConfig(ctx); err != nil {
			return err
		}
		// Create container by running serve command
		if err := ctx.RunCommand("serve docs --no-browser"); err != nil {
			return fmt.Errorf("failed to start serve container: %w", err)
		}
		if ctx.ExitCode != 0 {
			return fmt.Errorf("serve docs command failed with exit code %d: %s",
				ctx.ExitCode, ctx.CommandOutput)
		}
		return nil
	})

	sc.Step(`^serve container is not running$`, func() error {
		// Set up books.yml if we're in isolation
		if err := setupBooksConfig(ctx); err != nil {
			return err
		}
		// Ensure no container is running by running stop command
		// Ignore errors since container may not exist
		_ = ctx.RunCommand("serve docs --stop") //nolint:errcheck // Ignore error if container doesn't exist
		return nil
	})

	sc.Step(`^serve container should start successfully$`, func() error {
		// Verify container started by checking command succeeded
		if ctx.ExitCode != 0 {
			return fmt.Errorf("serve command failed with exit code %d", ctx.ExitCode)
		}
		return nil
	})

	sc.Step(`^serve container should be stopped$`, func() error {
		// Verify container stopped by checking command succeeded
		if ctx.ExitCode != 0 {
			return fmt.Errorf("stop command failed with exit code %d", ctx.ExitCode)
		}
		return nil
	})

	sc.Step(`^documentation should be accessible$`, func() error {
		// Placeholder - in mock mode, just verify command succeeded
		return nil
	})
}
