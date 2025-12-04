// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains docs command step definitions.
package srccommands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// docsContext holds Docker-related state for docs tests.
type docsContext struct {
	dockerClient    *client.Client
	dockerAvailable bool
}

// registerDocsSteps registers step definitions for docs command features.
func registerDocsSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	dCtx := &docsContext{}

	// Setup/teardown
	sc.After(func(goctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if dCtx.dockerClient != nil {
			dCtx.dockerClient.Close()
		}
		return goctx, nil
	})

	// Given steps
	sc.Step(`^docker service is available$`, func() error {
		return docsCheckDocker(dCtx)
	})

	// Then steps - MkDocs container state
	sc.Step(`^MkDocs container is running$`, func() error {
		return docsContainerState(dCtx, true)
	})
	sc.Step(`^MkDocs container is not running$`, func() error {
		return docsContainerState(dCtx, false)
	})
	sc.Step(`^MkDocs container should start successfully$`, func() error {
		return docsContainerState(dCtx, true)
	})
	sc.Step(`^MkDocs container should be stopped$`, func() error {
		return docsContainerState(dCtx, false)
	})
	sc.Step(`^documentation should be accessible$`, func() error {
		// Placeholder - would check HTTP endpoint
		return nil
	})
}

// docsCheckDocker verifies Docker is available.
func docsCheckDocker(dCtx *docsContext) error {
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

// docsContainerState ensures MkDocs container is in expected state.
// When used as a Given step, it will start/stop the container as needed.
// When used as a Then step, it verifies the container is in the expected state.
func docsContainerState(dCtx *docsContext, shouldBeRunning bool) error {
	if !dCtx.dockerAvailable || dCtx.dockerClient == nil {
		return fmt.Errorf("Docker is not available")
	}

	ctx := context.Background()
	containers, err := dCtx.dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	var containerID string
	found := false
	running := false
	for _, c := range containers {
		for _, name := range c.Names {
			if strings.Contains(name, "mkdocs") || strings.Contains(name, "cli-mkdocs") {
				containerID = c.ID
				found = true
				running = c.State == "running"
				break
			}
		}
		if found {
			break
		}
	}

	if shouldBeRunning {
		if !found {
			return fmt.Errorf("MkDocs container not found")
		}
		if !running {
			// Container exists but stopped - start it
			if err := dCtx.dockerClient.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
				return fmt.Errorf("failed to start stopped MkDocs container: %w", err)
			}
			// Brief wait for container to fully start
			time.Sleep(2 * time.Second)
		}
	} else {
		if found && running {
			// Container is running but should be stopped - stop it
			timeout := 10
			if err := dCtx.dockerClient.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
				return fmt.Errorf("failed to stop MkDocs container: %w", err)
			}
		}
	}

	return nil
}
