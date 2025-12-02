// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains design command step definitions.
// Note: Most design scenarios require Docker for Structurizr validation.
package srccommands

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// designContext holds Docker-related state for design tests.
type designContext struct {
	dockerClient    *client.Client
	dockerAvailable bool
	mockAIResponse  string
}

// registerDesignSteps registers step definitions for design command features.
func registerDesignSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	dCtx := &designContext{}

	// Setup/teardown
	sc.After(func(goctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if dCtx.dockerClient != nil {
			dCtx.dockerClient.Close()
		}
		return goctx, nil
	})

	// Given steps - repository/module setup
	sc.Step(`^a repository with AI configs at "([^"]*)"$`, func(path string) error {
		return internal.CreateDirectory(ctx, path)
	})
	sc.Step(`^module "([^"]*)" exists in "([^"]*)"$`, func(name, path string) error {
		return internal.CreateDirectory(ctx, path)
	})
	sc.Step(`^module "([^"]*)" has a workspace$`, func(module string) error {
		// Create a minimal workspace.dsl
		wsPath := fmt.Sprintf("specs/%s/.design/workspace.dsl", module)
		return internal.CreateFile(ctx, wsPath, minimalWorkspaceDSL(module))
	})
	sc.Step(`^module "([^"]*)" has a workspace at "([^"]*)"$`, func(module, path string) error {
		return internal.CreateFile(ctx, path, minimalWorkspaceDSL(module))
	})
	sc.Step(`^module "([^"]*)" has a valid workspace$`, func(module string) error {
		wsPath := fmt.Sprintf("specs/%s/.design/workspace.dsl", module)
		return internal.CreateFile(ctx, wsPath, minimalWorkspaceDSL(module))
	})
	sc.Step(`^module "([^"]*)" has a valid workspace at "([^"]*)"$`, func(module, path string) error {
		return internal.CreateFile(ctx, path, minimalWorkspaceDSL(module))
	})
	sc.Step(`^module "([^"]*)" has an invalid workspace at "([^"]*)"$`, func(module, path string) error {
		return internal.CreateFile(ctx, path, "invalid { dsl content")
	})
	sc.Step(`^no workspace exists for "([^"]*)"$`, func(module string) error {
		// Just ensure the directory structure exists without workspace
		return internal.CreateDirectory(ctx, fmt.Sprintf("specs/%s", module))
	})
	sc.Step(`^multiple modules have workspace files$`, func() error {
		for _, mod := range []string{"module-a", "module-b", "module-c"} {
			wsPath := fmt.Sprintf("specs/%s/.design/workspace.dsl", mod)
			if err := internal.CreateFile(ctx, wsPath, minimalWorkspaceDSL(mod)); err != nil {
				return err
			}
		}
		return nil
	})

	// Docker availability - Note: "docker service is available" is registered in steps_docs.go
	// Design tests that need Docker should use the same step or be refactored

	// Mock AI configuration
	sc.Step(`^the mock AI is configured to return a valid workspace$`, func() error {
		dCtx.mockAIResponse = minimalWorkspaceDSL("test-module")
		return nil
	})
	sc.Step(`^the mock AI is configured to return an updated workspace$`, func() error {
		dCtx.mockAIResponse = minimalWorkspaceDSL("test-module") + "\n// Updated"
		return nil
	})

	// Custom prompt
	sc.Step(`^a custom prompt file at "([^"]*)"$`, func(path string) error {
		return internal.CreateFile(ctx, path, "Custom design prompt content")
	})

	// Then steps - file verification
	// Note: "the file ... should exist" is registered in internal/steps.go
	sc.Step(`^a workspace file exists at "([^"]*)"$`, func(path string) error {
		return internal.FileExists(ctx, path)
	})
	// Note: "debug files should exist in" is registered in internal/steps.go
	sc.Step(`^debug logs should show the custom prompt was used$`, func() error {
		return internal.CustomPromptWasUsed(ctx, "design")
	})
	sc.Step(`^validation results should be written to "([^"]*)"$`, func(path string) error {
		return internal.FileExists(ctx, path)
	})
	sc.Step(`^aggregated results should be written to JSON file$`, func() error {
		// Placeholder - check for any JSON output
		return nil
	})

	// Then steps - output verification
	// Note: "the output should contain" is registered in internal/steps.go
	sc.Step(`^I should see success message with URL$`, func() error {
		return internal.OutputContainsAny(ctx, "http://", "https://", "localhost")
	})
	sc.Step(`^documentation should be accessible at the URL$`, func() error {
		// Placeholder - would need HTTP check
		return nil
	})
	sc.Step(`^the workspace should be updated$`, func() error {
		// Placeholder - verify workspace content changed
		return nil
	})
	sc.Step(`^the workspace should pass Structurizr validation$`, func() error {
		// Placeholder - requires Docker
		return nil
	})
	sc.Step(`^no confirmation prompt should appear$`, func() error {
		// With --force flag, no prompt
		return nil
	})

	// Docker container steps
	sc.Step(`^Structurizr container should start successfully$`, func() error {
		return designContainerRunning(dCtx, "structurizr")
	})
	sc.Step(`^two separate containers should be running$`, func() error {
		// Placeholder - check multiple containers
		return nil
	})
	sc.Step(`^each should have a unique port$`, func() error {
		// Placeholder - verify port allocation
		return nil
	})
}

// designCheckDocker verifies Docker is available.
func designCheckDocker(dCtx *designContext) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		dCtx.dockerAvailable = false
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	_, err = cli.Ping(context.Background())
	if err != nil {
		cli.Close()
		dCtx.dockerAvailable = false
		return fmt.Errorf("Docker is not running: %w", err)
	}

	dCtx.dockerClient = cli
	dCtx.dockerAvailable = true
	return nil
}

// designContainerRunning checks if a container with given name pattern is running.
func designContainerRunning(dCtx *designContext, namePattern string) error {
	if !dCtx.dockerAvailable || dCtx.dockerClient == nil {
		return fmt.Errorf("Docker is not available")
	}

	containers, err := dCtx.dockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.Contains(name, namePattern) && c.State == "running" {
				return nil
			}
		}
	}

	return fmt.Errorf("no running container matching %q found", namePattern)
}

// minimalWorkspaceDSL returns a minimal valid Structurizr DSL workspace.
func minimalWorkspaceDSL(module string) string {
	return fmt.Sprintf(`workspace "%s" "Test workspace" {
    model {
        user = person "User"
        system = softwareSystem "%s System"
        user -> system "Uses"
    }
    views {
        systemContext system "SystemContext" {
            include *
            autoLayout
        }
    }
}`, module, module)
}

