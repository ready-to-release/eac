package docker

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/ready-to-release/eac/go/adapters/docker/util"
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

	// Memory is the memory limit in bytes (0 = no limit)
	Memory int64

	// CPUs is the number of CPUs to allocate (0 = no limit)
	CPUs float64

	// Output is the writer for status messages (nil defaults to os.Stdout)
	Output io.Writer
}

// output returns the configured output writer, defaulting to os.Stdout.
func (c *ServeConfig) output() io.Writer {
	if c.Output != nil {
		return c.Output
	}
	return os.Stdout
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
// Port reservation prevents race conditions when starting multiple containers simultaneously.
func StartServe(ctx context.Context, config *ServeConfig) (*ServeResult, error) {
	// Allocate and reserve port atomically
	hostPort, err := FindAndReservePortOrUse(config.PreferredPort)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate port: %w", err)
	}
	// Ensure port is released on error
	var portReleased bool
	defer func() {
		if !portReleased {
			ReleasePort(hostPort)
		}
	}()

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
		// Container already running, keep the port reservation
		portReleased = true
		return existingResult, nil
	}

	// Ensure image exists
	if err := ensureImage(ctx, cli, config); err != nil {
		return nil, fmt.Errorf("failed to ensure image: %w", err)
	}

	// Translate content path for DinD
	mountSource, err := util.TranslatePathForMount(config.ContentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to translate path: %w", err)
	}

	// Format for Docker volume mount
	mountSource = util.FormatDockerVolume(mountSource)

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

	// Container started successfully, keep the port reservation
	portReleased = true

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
	for i := range containers {
		for _, name := range containers[i].Names {
			cleanName := strings.TrimPrefix(name, "/")
			// Match exact name or name with port suffix
			if cleanName == namePattern || strings.HasPrefix(cleanName, namePattern+"-") {
				found = true
				timeout := 10
				if err := cli.ContainerStop(ctx, containers[i].ID, container.StopOptions{Timeout: &timeout}); err != nil {
					return fmt.Errorf("failed to stop container %s: %w", cleanName, err)
				}
				if err := cli.ContainerRemove(ctx, containers[i].ID, container.RemoveOptions{}); err != nil {
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

	for i := range containers {
		for _, name := range containers[i].Names {
			cleanName := strings.TrimPrefix(name, "/")
			if cleanName == namePattern || strings.HasPrefix(cleanName, namePattern+"-") {
				// Extract port from container ports
				var hostPort int
				for _, p := range containers[i].Ports {
					if p.PublicPort != 0 {
						hostPort = int(p.PublicPort)
						break
					}
				}

				return &ServeResult{
					ContainerID:   containers[i].ID,
					ContainerName: cleanName,
					HostPort:      hostPort,
					URL:           fmt.Sprintf("http://localhost:%d", hostPort),
				}, true, nil
			}
		}
	}

	return nil, false, nil
}

// CheckImageStale checks if the image for a config needs to be rebuilt.
// Returns (stale, reason) - if stale, the image should be rebuilt before serving.
func CheckImageStale(ctx context.Context, config *ServeConfig) (stale bool, reason string, err error) {
	if config.BuildInfo == nil {
		return false, "", nil // No local build info, can't check staleness
	}

	cli, err := createDockerClient()
	if err != nil {
		return false, "", fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	// Check if image exists and get its creation time
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, "", fmt.Errorf("failed to list images: %w", err)
	}

	var imageCreated time.Time
	imageExists := false
	for i := range images {
		for _, tag := range images[i].RepoTags {
			if tag == config.Image {
				imageExists = true
				imageCreated = time.Unix(images[i].Created, 0)
				break
			}
		}
		if imageExists {
			break
		}
	}

	if !imageExists {
		return true, "image does not exist", nil
	}

	// Check if source files are newer than image
	stale, reason = isImageStale(config.BuildInfo.ContextPath, imageCreated)
	return stale, reason, nil
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
	for i := range containers {
		for _, name := range containers[i].Names {
			cleanName := strings.TrimPrefix(name, "/")
			if cleanName == namePattern || strings.HasPrefix(cleanName, namePattern+"-") {
				var hostPort int
				for _, p := range containers[i].Ports {
					if p.PublicPort != 0 {
						hostPort = int(p.PublicPort)
						break
					}
				}

				results = append(results, &ServeResult{
					ContainerID:   containers[i].ID,
					ContainerName: cleanName,
					HostPort:      hostPort,
					URL:           fmt.Sprintf("http://localhost:%d", hostPort),
				})
			}
		}
	}

	return results, nil
}

// createDockerClient is a variable holding the Docker client factory function.
// This can be overridden in tests to inject mock clients.
var createDockerClient = func() (DockerClient, error) {
	return NewDockerClient()
}

// isContainerRunning checks if a specific container is running.
func isContainerRunning(ctx context.Context, cli DockerClient, containerName string) (bool, *ServeResult, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, nil, err
	}

	for i := range containers {
		for _, name := range containers[i].Names {
			if strings.TrimPrefix(name, "/") == containerName {
				if containers[i].State == "running" {
					var hostPort int
					for _, p := range containers[i].Ports {
						if p.PublicPort != 0 {
							hostPort = int(p.PublicPort)
							break
						}
					}

					return true, &ServeResult{
						ContainerID:   containers[i].ID,
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

// createBuildContext creates a tar archive of the build context.
func createBuildContext(contextPath, dockerfilePath string) (io.ReadCloser, error) {
	// Create a pipe for the tar stream
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		tw := tar.NewWriter(pw)
		defer tw.Close()

		// Walk the context directory
		err := filepath.Walk(contextPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get relative path
			relPath, err := filepath.Rel(contextPath, path)
			if err != nil {
				return err
			}

			// Skip the context root itself
			if relPath == "." {
				return nil
			}

			// Create tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(relPath)

			// Write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			// Write file content if it's a regular file
			if info.Mode().IsRegular() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				if _, err := io.Copy(tw, file); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	return pr, nil
}

// ensureImage ensures the Docker image exists and is up-to-date, building if necessary.
// For local builds (BuildInfo != nil), it checks if source files are newer than the image.
func ensureImage(ctx context.Context, cli DockerClient, config *ServeConfig) error {
	// Check if image exists and get its creation time
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	var imageCreated time.Time
	imageExists := false
	for i := range images {
		for _, tag := range images[i].RepoTags {
			if tag == config.Image {
				imageExists = true
				imageCreated = time.Unix(images[i].Created, 0)
				break
			}
		}
		if imageExists {
			break
		}
	}

	// For local builds, check if source files are newer than image
	needsBuild := !imageExists
	if imageExists && config.BuildInfo != nil {
		stale, reason := isImageStale(config.BuildInfo.ContextPath, imageCreated)
		if stale {
			fmt.Fprintf(config.output(), "Image %s is stale: %s\n", config.Image, reason)
			needsBuild = true
		}
	}

	if !needsBuild {
		return nil
	}

	// Build or pull the image
	if config.BuildInfo != nil {
		// Build locally using Docker SDK
		fmt.Fprintf(config.output(), "Building image %s...\n", config.Image)

		// Create build context tar
		buildContext, err := createBuildContext(config.BuildInfo.ContextPath, config.BuildInfo.Dockerfile)
		if err != nil {
			return fmt.Errorf("failed to create build context: %w", err)
		}
		defer buildContext.Close()

		// Get relative dockerfile path
		dockerfileRel, err := filepath.Rel(config.BuildInfo.ContextPath, config.BuildInfo.Dockerfile)
		if err != nil {
			return fmt.Errorf("failed to get relative dockerfile path: %w", err)
		}

		// Build image with cache
		buildOptions := types.ImageBuildOptions{
			Tags:        []string{config.Image},
			Dockerfile:  dockerfileRel,
			Remove:      true,
			ForceRemove: true,
			NoCache:     false, // Use Docker build cache
		}

		resp, err := cli.ImageBuild(ctx, buildContext, buildOptions)
		if err != nil {
			return fmt.Errorf("failed to build image: %w", err)
		}
		defer resp.Body.Close()

		// Stream build progress
		err = jsonmessage.DisplayJSONMessagesStream(resp.Body, config.output(), os.Stdout.Fd(), true, nil)
		if err != nil {
			return fmt.Errorf("error during image build: %w", err)
		}

		fmt.Fprintf(config.output(), "\nImage %s built successfully\n", config.Image)
	} else {
		// Pull from registry using Docker SDK
		fmt.Fprintf(config.output(), "Pulling image %s...\n", config.Image)

		pullOptions := image.PullOptions{}
		reader, err := cli.ImagePull(ctx, config.Image, pullOptions)
		if err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
		defer reader.Close()

		// Stream pull progress
		err = jsonmessage.DisplayJSONMessagesStream(reader, config.output(), os.Stdout.Fd(), true, nil)
		if err != nil {
			return fmt.Errorf("error during image pull: %w", err)
		}

		fmt.Fprintf(config.output(), "\nImage %s pulled successfully\n", config.Image)
	}

	return nil
}

// isImageStale checks if any file in the build context is newer than the image.
// Returns (true, reason) if stale, (false, "") if up-to-date.
func isImageStale(contextPath string, imageCreated time.Time) (stale bool, reason string) {
	var newestFile string
	var newestTime time.Time

	err := filepath.Walk(contextPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}
		if info.IsDir() {
			return nil
		}

		modTime := info.ModTime()
		if modTime.After(newestTime) {
			newestTime = modTime
			newestFile = path
		}
		return nil
	})
	if err != nil {
		return false, ""
	}

	if newestTime.After(imageCreated) {
		relPath, relErr := filepath.Rel(contextPath, newestFile)
		if relErr != nil || relPath == "" {
			relPath = filepath.Base(newestFile)
		}
		return true, fmt.Sprintf("%s modified", relPath)
	}

	return false, ""
}

// removeExistingContainer removes an existing container if it exists.
func removeExistingContainer(ctx context.Context, cli DockerClient, containerName string) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return
	}

	for i := range containers {
		for _, name := range containers[i].Names {
			if strings.TrimPrefix(name, "/") == containerName {
				_ = cli.ContainerRemove(ctx, containers[i].ID, container.RemoveOptions{Force: true}) //nolint:errcheck // best-effort cleanup
				return
			}
		}
	}
}

// createContainer creates a new Docker container with the specified configuration.
func createContainer(ctx context.Context, cli DockerClient, config *ServeConfig, containerName string, hostPort int, mountSource string) (string, error) {
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
		Resources: container.Resources{
			Memory:   config.Memory,
			NanoCPUs: int64(config.CPUs * 1e9),
		},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// StreamContainerLogs streams container logs to stdout/stderr.
// containerNamePrefix is used to find the container by name prefix.
func StreamContainerLogs(ctx context.Context, containerNamePrefix string) error {
	cli, err := NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	// Find the container
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	var containerID string
	for i := range containers {
		c := &containers[i]
		for _, name := range c.Names {
			cleanName := strings.TrimPrefix(name, "/")
			if cleanName == containerNamePrefix || strings.HasPrefix(cleanName, containerNamePrefix+"-") {
				containerID = c.ID
				break
			}
		}
		if containerID != "" {
			break
		}
	}

	if containerID == "" {
		return fmt.Errorf("container not found: %s", containerNamePrefix)
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

	// Copy logs to stdout and stderr (Docker multiplexes stdout and stderr)
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, logs)
	if err != nil {
		return fmt.Errorf("error reading logs: %w", err)
	}

	return nil
}
