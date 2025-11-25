package docs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/src/commands/internal/serve"
	"github.com/ready-to-release/eac/src/core/repository"
)

const (
	// containerNameBase is the base name for mkdocs containers
	// Actual containers will have port suffix, e.g., cli-mkdocs-9001
	containerNameBase = "cli-mkdocs"
	imageName         = "cli-mkdocs:latest"
	dockerfile        = "containers/mkdocs/.Dockerfile"

	// containerInternalPort is the port MkDocs listens on inside the container
	containerInternalPort = 8000
)

// getRepoRoot returns the repository root directory
func getRepoRoot() (string, error) {
	return repository.GetRepositoryRoot("")
}

// isContainerRunning checks if any MkDocs container is running
func isContainerRunning(cli *client.Client, ctx context.Context) (bool, *ContainerInfo, error) {
	result, running, err := serve.IsServing(ctx, containerNameBase)
	if err != nil {
		return false, nil, err
	}
	if !running || result == nil {
		return false, nil, nil
	}

	return true, &ContainerInfo{
		Name: result.ContainerName,
		URL:  result.URL,
		Port: result.HostPort,
	}, nil
}

// startMkDocsContainer starts the MkDocs container
func startMkDocsContainer(cli *client.Client, ctx context.Context, port int) (*ContainerInfo, error) {
	// Check if container already running
	running, info, err := isContainerRunning(cli, ctx)
	if err != nil {
		return nil, err
	}
	if running {
		return info, fmt.Errorf("container is already running")
	}

	// Get repo root
	repoRoot, err := getRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to determine repository root: %w", err)
	}

	// Build configuration for the serve helper
	config := &serve.ServeConfig{
		Name:  containerNameBase,
		Image: imageName,
		BuildInfo: &serve.BuildInfo{
			Dockerfile:  filepath.Join(repoRoot, "containers", "mkdocs", ".Dockerfile"),
			ContextPath: filepath.Join(repoRoot, "containers", "mkdocs"),
		},
		ContentPath:   repoRoot,
		ContainerPath: "/docs",
		ContainerPort: containerInternalPort,
		Command:       []string{"mkdocs", "serve", "--dev-addr=0.0.0.0:8000"},
		PreferredPort: port,
		RestartPolicy: "unless-stopped",
	}

	// Start the serve container
	result, err := serve.StartServe(ctx, config)
	if err != nil {
		return nil, err
	}

	return &ContainerInfo{
		Name: result.ContainerName,
		URL:  result.URL,
		Port: result.HostPort,
	}, nil
}

// stopMkDocsContainer stops the MkDocs container
func stopMkDocsContainer(cli *client.Client, ctx context.Context) error {
	return serve.StopServe(ctx, containerNameBase)
}

// streamContainerLogs streams container logs to stdout
func streamContainerLogs(cli *client.Client, ctx context.Context) error {
	// Find the container
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	var containerID string
	for _, c := range containers {
		for _, name := range c.Names {
			cleanName := strings.TrimPrefix(name, "/")
			if cleanName == containerNameBase || strings.HasPrefix(cleanName, containerNameBase+"-") {
				containerID = c.ID
				break
			}
		}
		if containerID != "" {
			break
		}
	}

	if containerID == "" {
		return fmt.Errorf("container not found")
	}

	// Stream logs
	logOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	}

	logs, err := cli.ContainerLogs(ctx, containerID, logOptions)
	if err != nil {
		return fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	// Copy logs to stdout and stderr
	// Docker multiplexes stdout and stderr, so we need to demultiplex it
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, logs)
	if err != nil {
		return fmt.Errorf("error reading logs: %w", err)
	}

	return nil
}
