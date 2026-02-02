// Package docker provides container running functionality using Docker.
package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"

	containerport "github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/adapters/docker/util"
)

// Compile-time interface check.
var _ containerport.ContainerPort = (*ContainerAdapter)(nil)

// ContainerAdapter implements ContainerPort using Docker.
type ContainerAdapter struct {
	client DockerClient
}

// NewContainerAdapter creates a Docker-based container adapter.
func NewContainerAdapter() (*ContainerAdapter, error) {
	cli, err := NewDockerClient()
	if err != nil {
		return nil, err
	}
	return &ContainerAdapter{client: cli}, nil
}

// NewContainerAdapterWithClient creates a container adapter with an existing Docker client.
// Useful for testing or when the client is already available.
func NewContainerAdapterWithClient(client DockerClient) *ContainerAdapter {
	return &ContainerAdapter{client: client}
}

// Execute runs a container and waits for completion.
func (a *ContainerAdapter) Execute(ctx context.Context, config *containerport.ContainerConfig) (*containerport.ContainerResult, error) {
	if config == nil {
		return nil, fmt.Errorf("container config is nil")
	}

	start := time.Now()

	// Apply timeout if specified
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Build mounts
	var mounts []mount.Mount
	for _, m := range config.Mounts {
		// Translate path for DinD
		source, err := util.TranslatePathForMount(m.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to translate mount path %s: %w", m.Source, err)
		}
		source = util.FormatDockerVolume(source)

		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		})
	}

	// Build environment variables
	var envVars []string
	for k, v := range config.Env {
		envVars = append(envVars, k+"="+v)
	}

	// Container configuration
	containerConfig := &container.Config{
		Image:        config.Image,
		Cmd:          config.Command,
		Env:          envVars,
		WorkingDir:   config.WorkingDir,
		AttachStdout: true,
		AttachStderr: true,
	}

	if len(config.Entrypoint) > 0 {
		containerConfig.Entrypoint = config.Entrypoint
	}

	if config.User != "" {
		containerConfig.User = config.User
	}

	// Host configuration
	hostConfig := &container.HostConfig{
		Mounts:     mounts,
		AutoRemove: true,
	}

	if config.Network != "" {
		hostConfig.NetworkMode = container.NetworkMode(config.Network)
	}

	if config.Privileged {
		hostConfig.Privileged = true
	}

	// Apply resource limits
	if config.Resources != nil {
		if config.Resources.Memory != "" {
			memLimit := parseMemoryLimit(config.Resources.Memory)
			if memLimit > 0 {
				hostConfig.Resources.Memory = memLimit
			}
		}
		if config.Resources.CPUs > 0 {
			hostConfig.Resources.NanoCPUs = int64(config.Resources.CPUs * 1e9)
		}
		if config.Resources.ShmSize != "" {
			shmSize := parseMemoryLimit(config.Resources.ShmSize)
			if shmSize > 0 {
				hostConfig.ShmSize = shmSize
			}
		}
	}

	// Generate container name if not specified
	containerName := config.ContainerName
	if containerName == "" {
		containerName = fmt.Sprintf("eac-run-%d", time.Now().UnixNano())
	}

	// Remove any existing container with the same name
	_ = a.client.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true}) //nolint:errcheck

	// Create container
	resp, err := a.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	// Ensure cleanup on any exit
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		_ = a.client.ContainerStop(cleanupCtx, containerID, container.StopOptions{})                //nolint:errcheck
		_ = a.client.ContainerRemove(cleanupCtx, containerID, container.RemoveOptions{Force: true}) //nolint:errcheck
	}()

	// Attach to container before starting to capture all output
	attachResp, err := a.client.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	// Start wait before starting container
	waitChan, errChan := a.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	// Start container
	if err := a.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Read output
	var stdout, stderr bytes.Buffer
	var stdoutWriter, stderrWriter io.Writer = &stdout, &stderr
	if config.LogWriter != nil {
		stdoutWriter = io.MultiWriter(&stdout, config.LogWriter)
		stderrWriter = io.MultiWriter(&stderr, config.LogWriter)
	}
	_, _ = stdcopy.StdCopy(stdoutWriter, stderrWriter, attachResp.Reader) //nolint:errcheck

	// Wait for container exit
	var exitCode int
	for {
		select {
		case err := <-errChan:
			if err != nil {
				return nil, fmt.Errorf("container wait error: %w", err)
			}
			continue
		case waitResp := <-waitChan:
			exitCode = int(waitResp.StatusCode)
		case <-ctx.Done():
			return nil, fmt.Errorf("container execution cancelled or timed out")
		}
		break
	}

	duration := time.Since(start)

	return &containerport.ContainerResult{
		ExitCode: exitCode,
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: duration,
	}, nil
}

// Build builds a container image from a Containerfile/Dockerfile.
func (a *ContainerAdapter) Build(ctx context.Context, config *containerport.BuildConfig) error {
	if config == nil {
		return fmt.Errorf("build config is nil")
	}

	// Build using docker CLI for buildx support
	args := []string{"buildx", "build"}

	// Platforms
	if len(config.Platforms) > 0 {
		args = append(args, "--platform", strings.Join(config.Platforms, ","))
	}

	// Tag
	if config.ImageTag != "" {
		args = append(args, "-t", config.ImageTag)
	}

	// Dockerfile
	if config.Dockerfile != "" {
		args = append(args, "-f", config.Dockerfile)
	}

	// Build args
	for k, v := range config.BuildArgs {
		args = append(args, "--build-arg", k+"="+v)
	}

	// Cache
	if config.CacheFrom != "" {
		args = append(args, "--cache-from", config.CacheFrom)
	}
	if config.CacheTo != "" {
		args = append(args, "--cache-to", config.CacheTo)
	}

	// Load or push
	if config.Load {
		args = append(args, "--load")
	}
	if config.Push {
		args = append(args, "--push")
	}

	// Context path
	contextPath := config.ContextPath
	if contextPath == "" {
		contextPath = "."
	}
	args = append(args, contextPath)

	cmd := exec.CommandContext(ctx, "docker", args...)
	// Use LogWriter if provided (TUI mode), otherwise fall back to stdout (non-TUI mode)
	// Callers MUST provide LogWriter when TUI is active to avoid display corruption
	if config.LogWriter != nil {
		cmd.Stdout = config.LogWriter
		cmd.Stderr = config.LogWriter
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	return nil
}

// Pull ensures an image is available locally.
func (a *ContainerAdapter) Pull(ctx context.Context, imageRef string) error {
	reader, err := a.client.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageRef, err)
	}
	defer reader.Close()

	// Drain the reader to complete the pull
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("failed to read pull output: %w", err)
	}

	return nil
}

// ImageExists checks if an image exists locally.
func (a *ContainerAdapter) ImageExists(ctx context.Context, imageRef string) bool {
	images, err := a.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false
	}
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageRef {
				return true
			}
		}
	}
	return false
}

// IsAvailable checks if the container runtime is available.
func (a *ContainerAdapter) IsAvailable() bool {
	return IsDockerAvailable()
}

// Close releases resources.
func (a *ContainerAdapter) Close() error {
	if a.client != nil {
		return a.client.Close()
	}
	return nil
}

// parseMemoryLimit parses memory limit strings like "8g", "512m", "1024k".
func parseMemoryLimit(s string) int64 {
	if s == "" {
		return 0
	}

	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) < 2 {
		return 0
	}

	var multiplier int64 = 1
	numStr := s

	switch {
	case strings.HasSuffix(s, "g"):
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(s, "g")
	case strings.HasSuffix(s, "gb"):
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(s, "gb")
	case strings.HasSuffix(s, "m"):
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(s, "m")
	case strings.HasSuffix(s, "mb"):
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(s, "mb")
	case strings.HasSuffix(s, "k"):
		multiplier = 1024
		numStr = strings.TrimSuffix(s, "k")
	case strings.HasSuffix(s, "kb"):
		multiplier = 1024
		numStr = strings.TrimSuffix(s, "kb")
	}

	var num int64
	fmt.Sscanf(numStr, "%d", &num)
	return num * multiplier
}

// unavailableAdapter is a ContainerPort that always reports unavailable.
type unavailableAdapter struct {
	err error
}

func (a *unavailableAdapter) Execute(_ context.Context, _ *containerport.ContainerConfig) (*containerport.ContainerResult, error) {
	return nil, fmt.Errorf("container runtime not available: %w", a.err)
}

func (a *unavailableAdapter) Build(_ context.Context, _ *containerport.BuildConfig) error {
	return fmt.Errorf("container runtime not available: %w", a.err)
}

func (a *unavailableAdapter) Pull(_ context.Context, _ string) error {
	return fmt.Errorf("container runtime not available: %w", a.err)
}

func (a *unavailableAdapter) ImageExists(_ context.Context, _ string) bool {
	return false
}

func (a *unavailableAdapter) IsAvailable() bool {
	return false
}

func (a *unavailableAdapter) Close() error {
	return nil
}
