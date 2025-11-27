// Package tests provides BDD step definitions for the docs command.
//
// This file contains setup steps (Given/When) for docs command scenarios.
package tests

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// registerSetupSteps registers all Given/When step definitions
func registerSetupSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^docker service is available$`, dockerServiceIsAvailable)
	sc.Step(`^MkDocs container is running$`, mkdocsContainerIsRunning)
	sc.Step(`^MkDocs container is not running$`, mkdocsContainerIsNotRunning)

	// When steps
	// Note: "I run the command" steps are registered in the main test runner
}

// ============================================================================
// Given Steps
// ============================================================================

func dockerServiceIsAvailable() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	// Check Docker availability
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("docker client creation failed: %w", err)
	}

	_, err = cli.Ping(context.Background())
	if err != nil {
		cli.Close()
		return fmt.Errorf("docker daemon not responding: %w", err)
	}

	Ctx.DockerAvailable = true
	Ctx.DockerClient = cli
	return nil
}

func mkdocsContainerIsRunning() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	if !Ctx.DockerAvailable {
		return fmt.Errorf("docker is not available")
	}

	// Check if container exists and is running
	containerName := "cli-mkdocs"
	containers, err := Ctx.DockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	found := false
	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName || strings.Contains(name, containerName) {
				if c.State != "running" {
					return fmt.Errorf("container %s exists but is not running (state: %s)", containerName, c.State)
				}
				found = true
				Ctx.ContainerStarted = true
				// Extract URL from port mapping
				for _, p := range c.Ports {
					if p.PrivatePort == 8000 {
						Ctx.ContainerURL = fmt.Sprintf("http://localhost:%d", p.PublicPort)
						Ctx.ServerPort = int(p.PublicPort)
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
		// If not running, start it
		if RunCommand != nil {
			err := RunCommand("docs serve --no-browser")
			if err != nil {
				return fmt.Errorf("failed to start MkDocs container: %w", err)
			}
			Ctx.ContainerStarted = true
		}
	}

	return nil
}

func mkdocsContainerIsNotRunning() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	// Container should not be running
	Ctx.ContainerStarted = false
	return nil
}
