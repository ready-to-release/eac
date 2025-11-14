// Feature: src-commands_docs-command
// Godog step implementations for docs command BDD scenarios
//
// This file implements steps for the specification at:
// specs/src-commands/docs-command/specification.feature
package tests

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// docsTestContext holds state specific to docs command tests
type docsTestContext struct {
	dockerClient     *client.Client
	dockerAvailable  bool
	containerStarted bool
	containerID      string
	containerURL     string
}

var docsCtx *docsTestContext

// ============================================================================
// Given Steps
// ============================================================================

// dockerIsRunning is already defined in design_steps_test.go
// We'll reuse it via the shared step registration

// mkdocsContainerIsRunning checks if MkDocs container is running
func mkdocsContainerIsRunning() error {
	if !docsCtx.dockerAvailable {
		return fmt.Errorf("docker is not available")
	}

	// Check if container exists and is running
	containerName := "cli-mkdocs"
	containers, err := docsCtx.dockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	found := false
	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				if c.State != "running" {
					return fmt.Errorf("container %s exists but is not running (state: %s)", containerName, c.State)
				}
				found = true
				docsCtx.containerID = c.ID
				docsCtx.containerStarted = true
				// Extract URL from port mapping
				for _, p := range c.Ports {
					if p.PrivatePort == 8000 {
						docsCtx.containerURL = fmt.Sprintf("http://localhost:%d", p.PublicPort)
						break
					}
				}
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("MkDocs container 'cli-mkdocs' is not running")
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

// mkdocsContainerShouldStartSuccessfully verifies container started
func mkdocsContainerShouldStartSuccessfully() error {
	if !docsCtx.dockerAvailable {
		return fmt.Errorf("docker is not available")
	}

	// Check if container was created
	containerName := "cli-mkdocs"
	containers, err := docsCtx.dockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	found := false
	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				if c.State != "running" {
					return fmt.Errorf("container %s exists but is not running (state: %s)", containerName, c.State)
				}
				found = true
				docsCtx.containerID = c.ID
				docsCtx.containerStarted = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("MkDocs container was not created")
	}

	return nil
}

// mkdocsContainerShouldBeStopped verifies container is stopped
func mkdocsContainerShouldBeStopped() error {
	if !docsCtx.dockerAvailable {
		return fmt.Errorf("docker is not available")
	}

	// Check that container no longer exists or is not running
	containerName := "cli-mkdocs"
	containers, err := docsCtx.dockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				return fmt.Errorf("container %s still exists (state: %s)", containerName, c.State)
			}
		}
	}

	return nil
}

// documentationShouldBeAccessibleAt verifies HTTP accessibility
func documentationShouldBeAccessibleAt(url string) error {
	// Wait up to 10 seconds for documentation to become accessible
	maxRetries := 10
	retryDelay := 1 * time.Second

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err == nil {
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				// Success - documentation is accessible
				return nil
			}

			lastErr = fmt.Errorf("HTTP %d: expected 200 OK", resp.StatusCode)
		} else {
			lastErr = err
		}

		// Wait before retry (except on last iteration)
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("documentation not accessible at %s after %d seconds: %v",
		url, maxRetries, lastErr)
}

// iShouldSeeStoppedMessage verifies "stopped" message
func iShouldSeeStoppedMessage() error {
	if !strings.Contains(ctx.commandOutput, "stopped") &&
		!strings.Contains(ctx.commandOutput, "Stopped") {
		return fmt.Errorf("expected 'stopped' in output, got:\n%s", ctx.commandOutput)
	}
	return nil
}

// ============================================================================
// Scenario Initialization
// ============================================================================

func InitializeDocsScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Only initialize for docs features
		if strings.Contains(sc.Uri, "docs-command") {
			docsCtx = &docsTestContext{}

			// Initialize Docker client
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				docsCtx.dockerAvailable = false
				return ctx, nil // Don't fail - test will skip if docker not available
			}

			_, err = cli.Ping(context.Background())
			if err != nil {
				docsCtx.dockerAvailable = false
				cli.Close()
				return ctx, nil // Don't fail - test will skip if docker not available
			}

			docsCtx.dockerClient = cli
			docsCtx.dockerAvailable = true
		}
		return ctx, nil
	})

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// Only cleanup for docs features
		if strings.Contains(sc.Uri, "docs-command") && docsCtx != nil {
			if docsCtx.dockerClient != nil {
				// Clean up test container if it exists
				containerName := "cli-mkdocs"
				containers, _ := docsCtx.dockerClient.ContainerList(ctx, container.ListOptions{All: true})
				for _, c := range containers {
					for _, name := range c.Names {
						if strings.TrimPrefix(name, "/") == containerName {
							timeout := 5
							docsCtx.dockerClient.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeout})
							docsCtx.dockerClient.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true})
							break
						}
					}
				}
				docsCtx.dockerClient.Close()
			}
		}
		return ctx, nil
	})

	// Docs-specific steps
	sc.Step(`^MkDocs container is running$`, mkdocsContainerIsRunning)
	sc.Step(`^MkDocs container should start successfully$`, mkdocsContainerShouldStartSuccessfully)
	sc.Step(`^MkDocs container should be stopped$`, mkdocsContainerShouldBeStopped)
	sc.Step(`^documentation should be accessible at "([^"]*)"$`, documentationShouldBeAccessibleAt)
	sc.Step(`^I should see "([^"]*)" message$`, func(msg string) error {
		if msg == "stopped" {
			return iShouldSeeStoppedMessage()
		}
		return iShouldSee(msg)
	})

	// Note: "I should see success message with URL" step is already registered
	// in design_steps_test.go and can be reused
}
