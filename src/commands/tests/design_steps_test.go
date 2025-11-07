// Feature: src-commands_design-command
// Godog step implementations for design command BDD scenarios
//
// This file implements steps for the specification at:
// specs/src-commands/design-command/specification.feature
package tests

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// designTestContext holds state specific to design command tests
type designTestContext struct {
	dockerClient     *client.Client
	dockerAvailable  bool
	containerStarted bool
	containerID      string
	containerURL     string
}

var designCtx *designTestContext

// ============================================================================
// Given Steps
// ============================================================================

func dockerIsRunning() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		designCtx.dockerAvailable = false
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	_, err = cli.Ping(context.Background())
	if err != nil {
		designCtx.dockerAvailable = false
		cli.Close()
		return fmt.Errorf("docker is not running: %w", err)
	}

	designCtx.dockerClient = cli
	designCtx.dockerAvailable = true
	return nil
}

func moduleHasWorkspaceDslFile(module string) error {
	// Tests run from src/commands/tests, so we need to go up three directories
	workspacePath := filepath.Join("..", "..", "..", "docs", "reference", "design", module, "workspace.dsl")
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace.dsl not found at %s", workspacePath)
	}
	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

func structurizrContainerShouldStartSuccessfully() error {
	if !designCtx.dockerAvailable {
		return fmt.Errorf("docker is not available")
	}

	// Check if container was created
	containerName := "structurizr-cli"
	containers, err := designCtx.dockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	found := false
	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				designCtx.containerID = c.ID
				designCtx.containerStarted = true
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("structurizr container was not created")
	}

	return nil
}

func iShouldSeeSuccessMessageWithURL() error {
	// Check if output contains a URL
	if !strings.Contains(ctx.commandOutput, "http://localhost:") {
		return fmt.Errorf("output does not contain URL, got:\n%s", ctx.commandOutput)
	}

	// Extract URL for later verification
	lines := strings.Split(ctx.commandOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "http://localhost:") {
			start := strings.Index(line, "http://")
			if start >= 0 {
				url := strings.TrimSpace(line[start:])
				// Remove any trailing characters
				if idx := strings.IndexAny(url, " \t\n"); idx > 0 {
					url = url[:idx]
				}
				designCtx.containerURL = url
				break
			}
		}
	}

	if designCtx.containerURL == "" {
		return fmt.Errorf("could not extract URL from output")
	}

	return nil
}

func documentationShouldBeAccessibleAtTheURL() error {
	if designCtx.containerURL == "" {
		return fmt.Errorf("no URL found to check")
	}

	// Wait a bit for Structurizr to fully start
	time.Sleep(3 * time.Second)

	// Try to access the URL
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(designCtx.containerURL)
	if err != nil {
		return fmt.Errorf("failed to access %s: %w", designCtx.containerURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return fmt.Errorf("expected status 200 or 302, got %d", resp.StatusCode)
	}

	return nil
}

func iShouldSeeAListOfAvailableModules() error {
	// Check if output contains table headers
	if !strings.Contains(ctx.commandOutput, "MODULE") {
		return fmt.Errorf("output does not contain module table, got:\n%s", ctx.commandOutput)
	}
	return nil
}

func moduleShouldBeInTheList(module string) error {
	if !strings.Contains(ctx.commandOutput, module) {
		return fmt.Errorf("module '%s' not found in output:\n%s", module, ctx.commandOutput)
	}
	return nil
}

// ============================================================================
// Scenario Initialization
// ============================================================================

func InitializeDesignScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		designCtx = &designTestContext{
			dockerAvailable:  false,
			containerStarted: false,
		}
		return ctx, nil
	})

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// Cleanup: stop and remove test containers
		if designCtx.dockerClient != nil {
			// Try to stop and remove structurizr-cli container
			containerName := "structurizr-cli"
			containers, listErr := designCtx.dockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
			if listErr == nil {
				for _, c := range containers {
					for _, name := range c.Names {
						if strings.TrimPrefix(name, "/") == containerName {
							timeout := 5
							designCtx.dockerClient.ContainerStop(context.Background(), c.ID, container.StopOptions{Timeout: &timeout})
							designCtx.dockerClient.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{Force: true})
							break
						}
					}
				}
			}
			designCtx.dockerClient.Close()
		}
		return ctx, nil
	})

	// Given steps
	sc.Step(`^Docker is running$`, dockerIsRunning)
	sc.Step(`^module "([^"]*)" has workspace\.dsl file$`, moduleHasWorkspaceDslFile)

	// Then steps
	sc.Step(`^Structurizr container should start successfully$`, structurizrContainerShouldStartSuccessfully)
	sc.Step(`^I should see success message with URL$`, iShouldSeeSuccessMessageWithURL)
	sc.Step(`^documentation should be accessible at the URL$`, documentationShouldBeAccessibleAtTheURL)
	sc.Step(`^I should see a list of available modules$`, iShouldSeeAListOfAvailableModules)
	sc.Step(`^"([^"]*)" module should be in the list$`, moduleShouldBeInTheList)
}
