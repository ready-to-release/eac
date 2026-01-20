package docker

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/cache"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/logging"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/terminal"
)

// CreateContainerConfig creates a container configuration based on mode and extension.
func (ch *ContainerHost) CreateContainerConfig(ext *ExtensionConfig, mode ContainerMode, args []string, imageInspect *image.InspectResponse) *container.Config {
	envVars := ch.BuildEnvironmentVars(ext)

	config := &container.Config{
		Image: ext.Image,
		Env:   envVars,
	}

	switch mode {
	case ModeInteractive:
		config.Tty = true
		config.OpenStdin = true
		// If no entrypoint, use shell; if entrypoint exists, let it run
		if len(imageInspect.Config.Entrypoint) == 0 {
			config.Cmd = []string{"/bin/sh"}
		}
	case ModeRun:
		// Only enable TTY for truly interactive sessions
		// When running commands, we don't want TTY to avoid slow terminal handshake (~5s on Windows/WSL2)
		// Terminal dimensions are passed via COLUMNS/LINES environment variables instead
		if len(args) == 0 {
			// No args means interactive mode
			logging.Debug("ModeRun: No args detected, enabling TTY for interactive session")
			config.Tty = true
			config.OpenStdin = true
		} else {
			// Args present means command mode - disable TTY for performance
			// TTY causes ~5s delay due to terminal handshake/cursor position queries
			logging.Debugf("ModeRun: Command mode, TTY disabled for performance: args_count=%d", len(args))
			config.Tty = false
			config.OpenStdin = false
		}
		config.Cmd = args
	}

	// Only set WorkingDir if container does NOT have an entrypoint defined
	if len(imageInspect.Config.Entrypoint) == 0 {
		workdir := "/var/task"
		logging.Debugf("No entrypoint found in extension container, setting workingdir: workdir=%s", workdir)
		config.WorkingDir = workdir
	} else {
		logging.Debug("Found entrypoint in extension container, not setting workingdir")
	}

	return config
}

// CreateHostConfig creates the host configuration with volume mounts
// volumeRequests are optional cache volumes requested by the extension via metadata.
func (ch *ContainerHost) CreateHostConfig(ext *ExtensionConfig, volumeRequests []cache.VolumeRequest) *container.HostConfig {
	// In Docker-in-Docker mode, we need to use the HOST path for mounts, not the container path.
	// R2R_HOST_REPOROOT is set by the parent r2r CLI and contains the original host path.
	mountSource := ch.rootDir
	if hostRoot := os.Getenv("R2R_HOST_REPOROOT"); hostRoot != "" {
		logging.Debugf("Docker-in-Docker detected, using host path for mount: host_root=%s container_root=%s", hostRoot, ch.rootDir)
		mountSource = hostRoot
	}

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: mountSource,
			Target: "/var/task",
		},
	}

	// Add Docker service mount based on platform
	dockerMount := ch.getDockerServiceMount()
	if dockerMount != nil {
		mounts = append(mounts, *dockerMount)
	}

	// Add cache volumes requested by the extension
	for _, vol := range volumeRequests {
		if vol.Type == "cache" {
			// Named volume: {extension-name}-{volume-name}
			volumeName := fmt.Sprintf("%s-%s", ext.Name, vol.Name)
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeVolume,
				Source: volumeName,
				Target: vol.Target,
			})
			logging.Debugf("Adding cache volume mount: extension=%s volume=%s target=%s", ext.Name, volumeName, vol.Target)
		}
		// Future: handle "bind" type for bind mounts if needed
	}

	return &container.HostConfig{
		AutoRemove: true,
		Mounts:     mounts,
	}
}

// getDockerServiceMount returns the appropriate Docker service mount for the current platform.
func (ch *ContainerHost) getDockerServiceMount() *mount.Mount {
	// For all platforms (including WSL2/Windows), use the Unix socket path
	// Docker Desktop on Windows exposes the socket at this path in WSL2
	return &mount.Mount{
		Type:   mount.TypeBind,
		Source: "/var/run/docker.sock",
		Target: "/var/run/docker.sock",
	}
}

// CreateContainer creates a new Docker container with the specified configuration.
func (ch *ContainerHost) CreateContainer(containerConfig *container.Config, hostConfig *container.HostConfig) (string, error) {
	// Ensure Docker connectivity before creating container (lazy Ping)
	if err := ch.EnsureConnected(); err != nil {
		return "", err
	}

	resp, err := ch.client.ContainerCreate(ch.ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("error creating container: %w", err)
	}

	// TTY resize will be done after container starts (in StartContainer)

	return resp.ID, nil
}

// StartContainer starts a Docker container by ID.
func (ch *ContainerHost) StartContainer(containerID string) error {
	if err := ch.client.ContainerStart(ch.ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("error starting container: %w", err)
	}

	// After starting, resize the TTY if needed
	// Check if container has TTY enabled
	inspect, err := ch.client.ContainerInspect(ch.ctx, containerID)
	if err == nil && inspect.Config.Tty {
		if width, height, err := terminal.GetSize(); err == nil && width > 0 && height > 0 {
			logging.Debugf("Resizing container TTY after start: terminal_width=%d terminal_height=%d", width, height)
			resizeOptions := container.ResizeOptions{
				Height: uint(height), //nolint:gosec // G115: height > 0 checked above, no overflow
				Width:  uint(width),  //nolint:gosec // G115: width > 0 checked above, no overflow
			}
			if err := ch.client.ContainerResize(ch.ctx, containerID, resizeOptions); err != nil {
				logging.Debugf("Failed to resize container TTY after start: %v", err)
			} else {
				logging.Debug("Successfully resized container TTY after start")
			}
		}
	}

	return nil
}

// AttachToContainer attaches to a container for I/O operations.
func (ch *ContainerHost) AttachToContainer(containerID string) (types.HijackedResponse, error) {
	// Inspect container to determine if stdin should be attached
	inspect, err := ch.client.ContainerInspect(ch.ctx, containerID)
	if err != nil {
		return types.HijackedResponse{}, fmt.Errorf("error inspecting container: %w", err)
	}

	// Only attach stdin if the container has OpenStdin enabled
	attachStdin := inspect.Config.OpenStdin

	logging.Debugf("Attaching to container with appropriate stdin setting: attach_stdin=%v container_open_stdin=%v container_id=%s", attachStdin, inspect.Config.OpenStdin, containerID)

	attachResp, err := ch.client.ContainerAttach(ch.ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdin:  attachStdin,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return types.HijackedResponse{}, fmt.Errorf("error attaching to container: %w", err)
	}
	return attachResp, nil
}

// WaitForContainer waits for a container to finish execution.
func (ch *ContainerHost) WaitForContainer(containerID string) (<-chan container.WaitResponse, <-chan error) {
	return ch.client.ContainerWait(ch.ctx, containerID, container.WaitConditionNotRunning)
}

// StopContainer stops a running container.
func (ch *ContainerHost) StopContainer(containerID string) error {
	return ch.client.ContainerStop(ch.ctx, containerID, container.StopOptions{})
}

// StopContainerWithContext stops a running container with a specific context for timeout control.
func (ch *ContainerHost) StopContainerWithContext(ctx context.Context, containerID string) error {
	// Docker will send SIGTERM first, then SIGKILL after the timeout
	// The default timeout is 10 seconds, but we're using the context to control it
	return ch.client.ContainerStop(ctx, containerID, container.StopOptions{})
}

// GetRootDir returns the root directory path.
func (ch *ContainerHost) GetRootDir() string {
	return ch.rootDir
}

func (ch *ContainerHost) GetContainerSnapshot() (map[string]string, error) {
	containers, err := ch.client.ContainerList(ch.ctx, container.ListOptions{
		All: false, // Only running containers
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	snapshot := make(map[string]string)
	for _, cont := range containers {
		// Use container ID as key, image as value for identification
		snapshot[cont.ID] = cont.Image
	}

	return snapshot, nil
}

// WarnAboutNewContainers compares before/after snapshots and warns about new containers
// If autoRemove is true, it will stop and remove the containers instead of just warning
// expectedHostImages is a list of images that are expected to be created by the extension (e.g., serve commands).
func (ch *ContainerHost) WarnAboutNewContainers(beforeSnapshot, afterSnapshot map[string]string, extensionImage string, autoRemove bool, expectedHostImages []string) {
	for containerID, image := range afterSnapshot {
		if _, existed := beforeSnapshot[containerID]; !existed {
			// Skip our own main container
			if image == extensionImage {
				continue
			}

			// Skip expected host containers (e.g., from serve commands)
			if isExpectedHostImage(image, expectedHostImages) {
				logging.Debugf("Skipping expected host container: container_id=%s image=%s", containerID[:12], image)
				continue
			}

			if autoRemove {
				logging.Infof("Auto-removing detected child container: %s (container_id=%s extension=%s)", image, containerID[:12], extensionImage)

				// Stop and remove the container
				if err := ch.client.ContainerStop(ch.ctx, containerID, container.StopOptions{}); err != nil {
					logging.Warnf("Failed to stop child container: container_id=%s error=%v", containerID[:12], err)
				}

				if err := ch.client.ContainerRemove(ch.ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
					logging.Warnf("Failed to remove child container: container_id=%s error=%v", containerID[:12], err)
				} else {
					logging.Infof("Successfully removed child container: container_id=%s", containerID[:12])
				}
			} else {
				logging.Warnf("New container appeared during run: %s (container_id=%s extension=%s). This could be an indication of missing internal cleanup of docker-in-docker", image, containerID[:12], extensionImage)
			}
		}
	}
}

func (ch *ContainerHost) Close() error {
	return ch.client.Close()
}

// isExpectedHostImage checks if an image is in the list of expected host images.
func isExpectedHostImage(image string, expectedImages []string) bool {
	for _, expected := range expectedImages {
		if image == expected {
			return true
		}
	}
	return false
}
