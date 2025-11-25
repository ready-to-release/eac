package serve

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// ServeConfig holds configuration for serving a website in a container.
type ServeConfig struct {
	// Name is the base container name (e.g., "cli-mkdocs", "structurizr-lite-mymodule")
	// The actual container name will include the port for multi-instance support.
	Name string

	// Image is the Docker image to use
	Image string

	// BuildInfo contains information for building the image locally.
	// If nil, the image is expected to exist or be pulled from a registry.
	BuildInfo *BuildInfo

	// ContentPath is the local path to the content to serve.
	// In DinD mode, this will be translated to the host path automatically.
	ContentPath string

	// ContainerPath is where to mount the content inside the container
	ContainerPath string

	// ContainerPort is the port the service listens on inside the container
	ContainerPort int

	// Command is the command to run (optional, uses image default if nil)
	Command []string

	// EnvVars are additional environment variables to set
	EnvVars []string

	// RestartPolicy sets the container restart policy (default: "unless-stopped")
	RestartPolicy string

	// PreferredPort is the preferred host port (0 = auto-allocate)
	PreferredPort int
}

// BuildInfo holds information for building a local Docker image.
type BuildInfo struct {
	// Dockerfile is the path to the Dockerfile
	Dockerfile string
	// ContextPath is the build context path
	ContextPath string
}

// ServeResult holds the result of starting a serve container.
type ServeResult struct {
	// ContainerID is the Docker container ID
	ContainerID string
	// ContainerName is the full container name (includes port)
	ContainerName string
	// HostPort is the port on the host where the service is accessible
	HostPort int
	// URL is the full URL to access the service
	URL string
}

// StartServe starts a container serving web content.
// It handles port allocation, DinD path translation, image building, and container creation.
func StartServe(ctx context.Context, config *ServeConfig) (*ServeResult, error) {
	// Allocate port
	hostPort, err := FindAvailablePortOrUse(config.PreferredPort)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate port: %w", err)
	}

	// Generate container name with port for multi-instance support
	containerName := fmt.Sprintf("%s-%d", config.Name, hostPort)

	// Create Docker client
	cli, err := createDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	// Check if container is already running
	running, existingResult, err := isContainerRunning(ctx, cli, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to check container status: %w", err)
	}
	if running {
		return existingResult, nil
	}

	// Ensure image exists
	if err := ensureImage(ctx, cli, config); err != nil {
		return nil, fmt.Errorf("failed to ensure image: %w", err)
	}

	// Translate content path for DinD
	mountSource, err := TranslatePathForMount(config.ContentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to translate path: %w", err)
	}

	// Format for Docker volume mount
	mountSource = FormatDockerVolume(mountSource)

	// Remove existing container with same name if stopped
	removeExistingContainer(ctx, cli, containerName)

	// Create container
	containerID, err := createContainer(ctx, cli, config, containerName, hostPort, mountSource)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for server to start
	time.Sleep(2 * time.Second)

	return &ServeResult{
		ContainerID:   containerID,
		ContainerName: containerName,
		HostPort:      hostPort,
		URL:           fmt.Sprintf("http://localhost:%d", hostPort),
	}, nil
}

// StopServe stops a serving container by name or name prefix.
// It stops and removes all containers matching the name pattern.
func StopServe(ctx context.Context, namePattern string) error {
	cli, err := createDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	found := false
	for _, c := range containers {
		for _, name := range c.Names {
			cleanName := strings.TrimPrefix(name, "/")
			// Match exact name or name with port suffix
			if cleanName == namePattern || strings.HasPrefix(cleanName, namePattern+"-") {
				found = true
				timeout := 10
				if err := cli.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
					return fmt.Errorf("failed to stop container %s: %w", cleanName, err)
				}
				if err := cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{}); err != nil {
					return fmt.Errorf("failed to remove container %s: %w", cleanName, err)
				}
			}
		}
	}

	if !found {
		return fmt.Errorf("no container found matching: %s", namePattern)
	}

	return nil
}

// IsServing checks if a container with the given name pattern is currently serving.
// Returns the result if running, or (nil, false, nil) if not running.
func IsServing(ctx context.Context, namePattern string) (*ServeResult, bool, error) {
	cli, err := createDockerClient()
	if err != nil {
		return nil, false, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil {
		return nil, false, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			cleanName := strings.TrimPrefix(name, "/")
			if cleanName == namePattern || strings.HasPrefix(cleanName, namePattern+"-") {
				// Extract port from container ports
				var hostPort int
				for _, p := range c.Ports {
					if p.PublicPort != 0 {
						hostPort = int(p.PublicPort)
						break
					}
				}

				return &ServeResult{
					ContainerID:   c.ID,
					ContainerName: cleanName,
					HostPort:      hostPort,
					URL:           fmt.Sprintf("http://localhost:%d", hostPort),
				}, true, nil
			}
		}
	}

	return nil, false, nil
}

// ListServing returns all running serve containers matching the name pattern.
func ListServing(ctx context.Context, namePattern string) ([]*ServeResult, error) {
	cli, err := createDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var results []*ServeResult
	for _, c := range containers {
		for _, name := range c.Names {
			cleanName := strings.TrimPrefix(name, "/")
			if cleanName == namePattern || strings.HasPrefix(cleanName, namePattern+"-") {
				var hostPort int
				for _, p := range c.Ports {
					if p.PublicPort != 0 {
						hostPort = int(p.PublicPort)
						break
					}
				}

				results = append(results, &ServeResult{
					ContainerID:   c.ID,
					ContainerName: cleanName,
					HostPort:      hostPort,
					URL:           fmt.Sprintf("http://localhost:%d", hostPort),
				})
			}
		}
	}

	return results, nil
}

// createDockerClient creates a new Docker client with appropriate options.
func createDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
}

// isContainerRunning checks if a specific container is running.
func isContainerRunning(ctx context.Context, cli *client.Client, containerName string) (bool, *ServeResult, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, nil, err
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				if c.State == "running" {
					var hostPort int
					for _, p := range c.Ports {
						if p.PublicPort != 0 {
							hostPort = int(p.PublicPort)
							break
						}
					}

					return true, &ServeResult{
						ContainerID:   c.ID,
						ContainerName: containerName,
						HostPort:      hostPort,
						URL:           fmt.Sprintf("http://localhost:%d", hostPort),
					}, nil
				}
				return false, nil, nil
			}
		}
	}

	return false, nil, nil
}

// ensureImage ensures the Docker image exists, building it if necessary.
func ensureImage(ctx context.Context, cli *client.Client, config *ServeConfig) error {
	// Check if image exists
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == config.Image {
				return nil
			}
		}
	}

	// Image doesn't exist
	if config.BuildInfo != nil {
		// Build locally
		fmt.Printf("Building image %s...\n", config.Image)
		cmd := exec.CommandContext(ctx, "docker", "build",
			"-t", config.Image,
			"-f", config.BuildInfo.Dockerfile,
			config.BuildInfo.ContextPath,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to build image: %w\nOutput: %s", err, string(output))
		}
		fmt.Printf("Image %s built successfully\n", config.Image)
	} else {
		// Pull from registry
		fmt.Printf("Pulling image %s...\n", config.Image)
		cmd := exec.CommandContext(ctx, "docker", "pull", config.Image)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to pull image: %w\nOutput: %s", err, string(output))
		}
		fmt.Printf("Image %s pulled successfully\n", config.Image)
	}

	return nil
}

// removeExistingContainer removes an existing container if it exists.
func removeExistingContainer(ctx context.Context, cli *client.Client, containerName string) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true})
				return
			}
		}
	}
}

// createContainer creates a new Docker container with the specified configuration.
func createContainer(ctx context.Context, cli *client.Client, config *ServeConfig, containerName string, hostPort int, mountSource string) (string, error) {
	containerPortStr := fmt.Sprintf("%d", config.ContainerPort)
	hostPortStr := fmt.Sprintf("%d", hostPort)

	containerConfig := &container.Config{
		Image: config.Image,
		ExposedPorts: nat.PortSet{
			nat.Port(containerPortStr + "/tcp"): struct{}{},
		},
		WorkingDir: config.ContainerPath,
		Env:        config.EnvVars,
	}

	if len(config.Command) > 0 {
		containerConfig.Cmd = config.Command
	}

	restartPolicy := config.RestartPolicy
	if restartPolicy == "" {
		restartPolicy = "unless-stopped"
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(containerPortStr + "/tcp"): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: hostPortStr,
				},
			},
		},
		Binds: []string{
			fmt.Sprintf("%s:%s", mountSource, config.ContainerPath),
		},
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyMode(restartPolicy),
		},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}
