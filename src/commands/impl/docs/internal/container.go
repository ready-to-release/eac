package docs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/ready-to-release/eac/src/core/repository"
)

const (
	containerName = "cli-mkdocs"
	imageName     = "cli-mkdocs:latest"
	dockerfile    = "containers/mkdocs/.Dockerfile"
)

// getRepoRoot returns the repository root directory
func getRepoRoot() (string, error) {
	return repository.GetRepositoryRoot("")
}

// isContainerRunning checks if the MkDocs container is running
func isContainerRunning(cli *client.Client, ctx context.Context) (bool, *ContainerInfo, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, nil, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				if c.State == "running" {
					// Extract port
					port := 8000
					for _, p := range c.Ports {
						if p.PrivatePort == 8000 {
							port = int(p.PublicPort)
							break
						}
					}

					return true, &ContainerInfo{
						Name: containerName,
						URL:  fmt.Sprintf("http://localhost:%d", port),
						Port: port,
					}, nil
				}
				return false, nil, nil
			}
		}
	}

	return false, nil, nil
}

// startMkDocsContainer starts the MkDocs container
func startMkDocsContainer(cli *client.Client, ctx context.Context, port int) (*ContainerInfo, error) {
	// Check if container already exists
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

	// Ensure image exists by building it
	err = ensureImageExists(cli, ctx, repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure image exists: %w", err)
	}

	// Remove existing container if it exists but not running
	containers, _ := cli.ContainerList(ctx, container.ListOptions{All: true})
	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				err = cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true})
				if err != nil {
					return nil, fmt.Errorf("failed to remove existing container: %w", err)
				}
				break
			}
		}
	}

	// Create container configuration
	hostPort := fmt.Sprintf("%d", port)
	containerPort := "8000"

	config := &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			nat.Port(containerPort + "/tcp"): struct{}{},
		},
		WorkingDir: "/docs",
		Cmd:        []string{"mkdocs", "serve", "--dev-addr=0.0.0.0:8000"},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(containerPort + "/tcp"): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: hostPort,
				},
			},
		},
		Binds: []string{
			fmt.Sprintf("%s:/docs", repoRoot),
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// Create container
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Wait a moment for the server to start
	time.Sleep(2 * time.Second)

	return &ContainerInfo{
		Name: containerName,
		URL:  fmt.Sprintf("http://localhost:%d", port),
		Port: port,
	}, nil
}

// stopMkDocsContainer stops the MkDocs container
func stopMkDocsContainer(cli *client.Client, ctx context.Context) error {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	found := false
	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				found = true
				timeout := 10
				err = cli.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeout})
				if err != nil {
					return fmt.Errorf("failed to stop container: %w", err)
				}

				err = cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{})
				if err != nil {
					return fmt.Errorf("failed to remove container: %w", err)
				}
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("container not found")
	}

	return nil
}

// ensureImageExists checks if the MkDocs image exists, builds it if not
func ensureImageExists(cli *client.Client, ctx context.Context, repoRoot string) error {
	// Check if image exists
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				// Image exists
				return nil
			}
		}
	}

	// Image doesn't exist, build it
	fmt.Println("📦 MkDocs image not found, building...")

	buildContext := filepath.Join(repoRoot, "containers", "mkdocs")
	dockerfilePath := filepath.Join(buildContext, ".Dockerfile")

	// Build image using docker build command
	// This is more reliable than using the Docker SDK for building
	fmt.Printf("   Building from: %s\n", buildContext)
	fmt.Printf("   Using Dockerfile: %s\n", dockerfilePath)

	// Execute build command
	cmd := exec.Command("docker", "build", "-t", imageName, "-f", dockerfilePath, buildContext)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("   Build output:\n%s\n", string(output))
		return fmt.Errorf("failed to build image: %w", err)
	}

	fmt.Println("   ✅ Image built successfully")
	return nil
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
			if strings.TrimPrefix(name, "/") == containerName {
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
