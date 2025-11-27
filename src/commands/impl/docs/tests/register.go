// Package tests provides shared test context and step registration for docs command.
//
// This file registers all step definitions for the docs command test scenarios.
package tests

import (
	"context"
	"strings"

	"github.com/cucumber/godog"
	"github.com/docker/docker/api/types/container"
)

// InitializeDocsScenario registers all docs-related step definitions.
func InitializeDocsScenario(sc *godog.ScenarioContext) {
	// Initialize context before scenario runs
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		SyncContextFromParent()
		return ctx, nil
	})

	// Register setup steps (Given/When)
	registerSetupSteps(sc)

	// Register verification steps (Then)
	registerVerifySteps(sc)

	// Cleanup after scenario
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// Clean up Docker containers to prevent state leakage
		if Ctx != nil && Ctx.DockerClient != nil {
			containerName := "cli-mkdocs"
			containers, _ := Ctx.DockerClient.ContainerList(ctx, container.ListOptions{All: true})
			for _, c := range containers {
				for _, name := range c.Names {
					if strings.TrimPrefix(name, "/") == containerName || strings.Contains(name, containerName) {
						timeout := 5
						_ = Ctx.DockerClient.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeout})
						_ = Ctx.DockerClient.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true})
						break
					}
				}
			}
		}

		SyncContextToParent()
		return ctx, nil
	})
}

// SyncContextFromParent synchronizes context from the parent test runner.
// This must be called before each scenario to ensure we have the latest context.
func SyncContextFromParent() {
	// Context is synced by the bridge file in src/commands/tests/
}

// SyncContextToParent synchronizes changes back to the parent test runner.
// This should be called after each scenario to capture any state changes.
func SyncContextToParent() {
	// Context is synced by the bridge file in src/commands/tests/
}
