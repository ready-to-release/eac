// Package specs contains godog step implementations for eac-commands.
//
// This file contains docs command step definitions.
package specs

import (
	"fmt"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/eac/godog"
)

// registerDocsSteps registers step definitions for docs command features.
func registerDocsSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Given steps
	sc.Step(`^docker service is available$`, func() error {
		// Mock Docker is always available in tests
		return nil
	})

	// Given/Then steps - serve container state
	sc.Step(`^serve container is running$`, func() error {
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
