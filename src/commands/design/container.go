package design

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	// StructurizrImage is the Docker image for Structurizr Lite
	StructurizrImage = "structurizr/lite:latest"

	// ContainerPrefix is the prefix for Structurizr container names
	ContainerPrefix = "structurizr-"

	// HealthCheckTimeout is the maximum time to wait for container health check
	HealthCheckTimeout = 30 * time.Second

	// HealthCheckInterval is how often to check container health
	HealthCheckInterval = 2 * time.Second
)

// ContainerInfo contains information about a running Structurizr container
type ContainerInfo struct {
	ID     string
	Name   string
	Port   int
	Module string
	Status string
	URL    string
}

// ContainerManager manages Structurizr Lite Docker containers
type ContainerManager struct {
	client *client.Client
	ctx    context.Context
}

// NewContainerManager creates a new ContainerManager
func NewContainerManager(ctx context.Context) (*ContainerManager, error) {
	// Configure Docker client options
	clientOpts := []client.Opt{
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	}

	// Override Docker host if R2R_DOCKER_HOST is set
	if dockerHost := os.Getenv("R2R_DOCKER_HOST"); dockerHost != "" {
		clientOpts = append(clientOpts, client.WithHost(dockerHost))
	}

	cli, err := client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Verify Docker daemon is accessible
	_, err = cli.Ping(ctx)
	if err != nil {
		cli.Close()
		if strings.Contains(err.Error(), "docker_engine") ||
			strings.Contains(err.Error(), "cannot connect to the Docker daemon") ||
			strings.Contains(err.Error(), "Is the docker daemon running") {
			return nil, fmt.Errorf("Docker is not running. Please start Docker Desktop and try again")
		}
		return nil, fmt.Errorf("cannot connect to Docker daemon: %w", err)
	}

	return &ContainerManager{
		client: cli,
		ctx:    ctx,
	}, nil
}

// Close closes the Docker client connection
func (cm *ContainerManager) Close() error {
	if cm.client != nil {
		return cm.client.Close()
	}
	return nil
}

// GetContainerName returns the container name for a module
func GetContainerName(module string) string {
	return ContainerPrefix + module
}

// IsRunning checks if a container is currently running
func (cm *ContainerManager) IsRunning(module string) (bool, *ContainerInfo, error) {
	containerName := GetContainerName(module)

	containers, err := cm.client.ContainerList(cm.ctx, container.ListOptions{All: true})
	if err != nil {
		return false, nil, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				info := &ContainerInfo{
					ID:     c.ID,
					Name:   containerName,
					Module: module,
					Status: c.State,
				}

				// Extract port if available
				if len(c.Ports) > 0 {
					info.Port = int(c.Ports[0].PublicPort)
					info.URL = fmt.Sprintf("http://localhost:%d", info.Port)
				}

				isRunning := c.State == "running"
				return isRunning, info, nil
			}
		}
	}

	return false, nil, nil
}

// StartContainer starts a Structurizr container for the given module
func (cm *ContainerManager) StartContainer(module string, port int) (*ContainerInfo, error) {
	containerName := GetContainerName(module)

	// Check if container already exists
	running, info, err := cm.IsRunning(module)
	if err != nil {
		return nil, err
	}

	if running {
		// Container is already running
		return info, nil
	}

	// If container exists but not running, remove it
	if info != nil {
		err = cm.client.ContainerRemove(cm.ctx, info.ID, container.RemoveOptions{Force: true})
		if err != nil {
			return nil, fmt.Errorf("failed to remove stopped container: %w", err)
		}
	}

	// Ensure image is available (pull if needed)
	err = cm.ensureImage()
	if err != nil {
		return nil, err
	}

	// Create container configuration
	hostPort := fmt.Sprintf("%d", port)
	containerPort := "8080/tcp"

	config := &container.Config{
		Image: StructurizrImage,
		ExposedPorts: nat.PortSet{
			nat.Port(containerPort): struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(containerPort): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: hostPort,
				},
			},
		},
		AutoRemove: false, // Keep container for reuse
	}

	// Create container
	resp, err := cm.client.ContainerCreate(
		cm.ctx,
		config,
		hostConfig,
		nil,
		nil,
		containerName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	err = cm.client.ContainerStart(cm.ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to be healthy before copying files
	// This ensures the /usr/local/structurizr directory is created
	err = cm.WaitForHealthy(resp.ID, HealthCheckTimeout)
	if err != nil {
		// Container not healthy, try copying anyway
		// Don't fail yet - just log
	}

	// Copy workspace files to container
	// Wait a bit more to ensure filesystem is ready
	time.Sleep(3 * time.Second)

	// Ensure the target directory exists in the container
	containerName = GetContainerName(module)
	mkdirCmd := exec.Command("docker", "exec", containerName, "mkdir", "-p", "/usr/local/structurizr")
	mkdirCmd.Run() // Ignore error - directory might already exist

	err = cm.CopyWorkspace(module, resp.ID)
	if err != nil {
		// Try to stop and remove container on failure
		cm.client.ContainerStop(cm.ctx, resp.ID, container.StopOptions{})
		cm.client.ContainerRemove(cm.ctx, resp.ID, container.RemoveOptions{Force: true})
		return nil, fmt.Errorf("failed to copy workspace files: %w", err)
	}

	return &ContainerInfo{
		ID:     resp.ID,
		Name:   containerName,
		Port:   port,
		Module: module,
		Status: "running",
		URL:    fmt.Sprintf("http://localhost:%d", port),
	}, nil
}

// StopContainer stops and removes a Structurizr container
func (cm *ContainerManager) StopContainer(module string) error {
	running, info, err := cm.IsRunning(module)
	if err != nil {
		return err
	}

	if !running || info == nil {
		return fmt.Errorf("no running container found for module '%s'", module)
	}

	// Stop container
	timeout := 10 // seconds
	err = cm.client.ContainerStop(cm.ctx, info.ID, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	// Remove container
	err = cm.client.ContainerRemove(cm.ctx, info.ID, container.RemoveOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

// CopyWorkspace copies workspace.dsl and optional docs/decisions to container
// Uses docker cp command as workaround for SDK issues
func (cm *ContainerManager) CopyWorkspace(module, containerID string) error {
	modulePath := GetModulePath(module)
	containerName := GetContainerName(module)

	// Copy workspace.dsl (required)
	workspacePath := filepath.Join(modulePath, "workspace.dsl")
	err := dockerCpToContainer(workspacePath, containerName+":/usr/local/structurizr/workspace.dsl")
	if err != nil {
		return fmt.Errorf("failed to copy workspace.dsl: %w", err)
	}

	// Copy docs folder (optional)
	docsPath := filepath.Join(modulePath, "docs")
	if stat, err := os.Stat(docsPath); err == nil && stat.IsDir() {
		err = dockerCpToContainer(docsPath, containerName+":/usr/local/structurizr/")
		if err != nil {
			return fmt.Errorf("failed to copy docs folder: %w", err)
		}
	}

	// Copy decisions folder (optional)
	decisionsPath := filepath.Join(modulePath, "decisions")
	if stat, err := os.Stat(decisionsPath); err == nil && stat.IsDir() {
		err = dockerCpToContainer(decisionsPath, containerName+":/usr/local/structurizr/")
		if err != nil {
			return fmt.Errorf("failed to copy decisions folder: %w", err)
		}
	}

	return nil
}

// dockerCpToContainer uses docker cp command to copy files
// This is a workaround for Docker SDK CopyToContainer issues on Windows
func dockerCpToContainer(src, dest string) error {
	cmd := exec.Command("docker", "cp", src, dest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

// WaitForHealthy waits for container to become healthy
func (cm *ContainerManager) WaitForHealthy(containerID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		inspect, err := cm.client.ContainerInspect(cm.ctx, containerID)
		if err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}

		if inspect.State.Running {
			// Container is running - consider it healthy
			// Structurizr Lite doesn't have built-in health checks
			return nil
		}

		time.Sleep(HealthCheckInterval)
	}

	return fmt.Errorf("timeout waiting for container to start")
}

// ensureImage ensures the Structurizr image is available
func (cm *ContainerManager) ensureImage() error {
	// Check if image exists locally
	images, err := cm.client.ImageList(cm.ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == StructurizrImage {
				return nil // Image already exists
			}
		}
	}

	// Image doesn't exist, pull it
	reader, err := cm.client.ImagePull(cm.ctx, StructurizrImage, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull Structurizr image: %w", err)
	}
	defer reader.Close()

	// Read pull output (discard for now)
	io.Copy(io.Discard, reader)

	return nil
}

// addFileToTar adds a single file to a tar archive
func addFileToTar(tw *tar.Writer, filePath, nameInTar string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    nameInTar,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}

// addDirToTar adds a directory recursively to a tar archive
func addDirToTar(tw *tar.Writer, dirPath, nameInTar string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories (we only add files)
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		// Name in tar archive
		tarPath := filepath.Join(nameInTar, relPath)
		// Convert Windows paths to Unix paths for tar
		tarPath = filepath.ToSlash(tarPath)

		return addFileToTar(tw, path, tarPath)
	})
}
