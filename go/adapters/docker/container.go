// Package docker provides container running functionality using Docker.
package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"

	containerport "github.com/ready-to-release/eac/contracts/container-runtime/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/docker/util"
)

// Retry configuration constants.
const (
	maxRetries   = 3
	initialDelay = 100 * time.Millisecond
	maxDelay     = 2 * time.Second
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
// It includes retry logic with exponential backoff for transient errors
// such as Windows file locks and Docker cleanup race conditions.
func (a *ContainerAdapter) Execute(ctx context.Context, config *containerport.ContainerConfig) (*containerport.ContainerResult, error) {
	if config == nil {
		return nil, fmt.Errorf("container config is nil")
	}

	var lastResult *containerport.ContainerResult
	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := a.executeOnce(ctx, config)

		if err == nil && result != nil && result.ExitCode == 0 {
			return result, nil
		}

		lastResult = result
		lastErr = err

		// Check if this is a retryable error
		if !isRetryableError(err, result) {
			return result, err
		}

		// Don't retry on last attempt
		if attempt == maxRetries {
			break
		}

		// Use longer delay for image pull failures (rate limits, registry issues)
		retryDelay := delay
		if err != nil && strings.Contains(err.Error(), "failed to pull image") {
			retryDelay = max(delay, 2*time.Second)
		}

		// Log retry attempt
		if config.LogWriter != nil {
			fmt.Fprintf(config.LogWriter, "[container] Attempt %d/%d failed, retrying in %v...\n",
				attempt, maxRetries, retryDelay)
		}

		// Wait with exponential backoff
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryDelay):
		}

		delay = min(delay*2, maxDelay)
	}

	return lastResult, lastErr
}

// executeOnce runs a container a single time without retry.
func (a *ContainerAdapter) executeOnce(ctx context.Context, config *containerport.ContainerConfig) (*containerport.ContainerResult, error) {
	start := time.Now()

	// Apply timeout if specified
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Build Docker-level configurations from the port config
	mounts, err := buildMounts(config.Mounts)
	if err != nil {
		return nil, err
	}

	containerConfig := buildContainerConfig(config)
	hostConfig := buildHostConfig(config, mounts)

	// Resolve container name
	containerName := config.ContainerName
	if containerName == "" {
		containerName = GenerateContainerName()
	}

	// Remove any existing container with the same name
	_ = a.client.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true}) //nolint:errcheck

	// Auto-pull image if not available locally
	if !a.ImageExists(ctx, config.Image) {
		if config.LogWriter != nil {
			fmt.Fprintf(config.LogWriter, "[container] Pulling image %s...\n", config.Image)
		}
		if err := a.Pull(ctx, config.Image); err != nil {
			return nil, fmt.Errorf("failed to pull image %s: %w", config.Image, err)
		}
	}

	// Create container
	resp, err := a.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	// Ensure cleanup on any exit
	defer a.cleanupContainer(containerID)

	// Attach to container before starting to capture all output
	attachResp, err := a.client.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
		Stdin:  config.StdinReader != nil,
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

	// Pipe stdin if provided
	if config.StdinReader != nil {
		go func() {
			defer attachResp.CloseWrite() //nolint:errcheck
			_, _ = io.Copy(attachResp.Conn, config.StdinReader)
		}()
	}

	// Capture output and wait for container exit
	exitCode, stdout, stderr, err := a.captureOutputAndWait(ctx, config, attachResp, waitChan, errChan)
	if err != nil {
		return nil, err
	}

	return &containerport.ContainerResult{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
		Duration: time.Since(start),
	}, nil
}

// buildMounts translates port-level mount configs into Docker mount specs.
func buildMounts(portMounts []containerport.MountConfig) ([]mount.Mount, error) {
	var mounts []mount.Mount
	for _, m := range portMounts {
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
	return mounts, nil
}

// buildContainerConfig creates a Docker container.Config from a port ContainerConfig.
func buildContainerConfig(config *containerport.ContainerConfig) *container.Config {
	var envVars []string
	for k, v := range config.Env {
		envVars = append(envVars, k+"="+v)
	}

	cc := &container.Config{
		Image:        config.Image,
		Cmd:          config.Command,
		Env:          envVars,
		WorkingDir:   config.WorkingDir,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  config.StdinReader != nil,
		OpenStdin:    config.StdinReader != nil,
		StdinOnce:    config.StdinReader != nil,
	}

	if len(config.Entrypoint) > 0 {
		cc.Entrypoint = config.Entrypoint
	}

	if config.User != "" {
		cc.User = config.User
	}

	return cc
}

// buildHostConfig creates a Docker container.HostConfig from a port ContainerConfig and resolved mounts.
func buildHostConfig(config *containerport.ContainerConfig, mounts []mount.Mount) *container.HostConfig {
	hc := &container.HostConfig{
		Mounts:     mounts,
		AutoRemove: true,
	}

	if config.Network != "" {
		hc.NetworkMode = container.NetworkMode(config.Network)
	}

	if config.Privileged {
		hc.Privileged = true
	}

	applyResourceLimits(hc, config.Resources)

	return hc
}

// applyResourceLimits sets memory, CPU, and SHM limits on a host config.
func applyResourceLimits(hc *container.HostConfig, resources *containerport.ResourceConfig) {
	if resources == nil {
		return
	}

	if resources.Memory != "" {
		if memLimit := parseMemoryLimit(resources.Memory); memLimit > 0 {
			hc.Resources.Memory = memLimit
		}
	}

	if resources.CPUs > 0 {
		hc.Resources.NanoCPUs = int64(resources.CPUs * 1e9)
	}

	if resources.ShmSize != "" {
		if shmSize := parseMemoryLimit(resources.ShmSize); shmSize > 0 {
			hc.ShmSize = shmSize
		}
	}
}

// cleanupContainer stops and removes a container, handling Windows file-handle delays.
func (a *ContainerAdapter) cleanupContainer(containerID string) {
	// Add small delay before cleanup on Windows to allow file handles to release
	if runtime.GOOS == "windows" {
		time.Sleep(50 * time.Millisecond)
	}

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cleanupCancel()
	_ = a.client.ContainerStop(cleanupCtx, containerID, container.StopOptions{})                //nolint:errcheck
	_ = a.client.ContainerRemove(cleanupCtx, containerID, container.RemoveOptions{Force: true}) //nolint:errcheck
}

// captureOutputAndWait copies container stdout/stderr and waits for the container to exit.
// It returns the exit code, captured stdout bytes, captured stderr bytes, and any error.
func (a *ContainerAdapter) captureOutputAndWait(
	ctx context.Context,
	config *containerport.ContainerConfig,
	attachResp types.HijackedResponse,
	waitChan <-chan container.WaitResponse,
	errChan <-chan error,
) (int, []byte, []byte, error) {
	// Determine output destinations (StdoutWriter > LogWriter > buffer capture)
	var stdout, stderr bytes.Buffer
	stdoutWriter := chooseWriter(config.StdoutWriter, config.LogWriter, &stdout)
	stderrWriter := chooseWriter(config.StderrWriter, config.LogWriter, &stderr)

	_, _ = stdcopy.StdCopy(stdoutWriter, stderrWriter, attachResp.Reader) //nolint:errcheck

	// Wait for container exit
	var exitCode int
	for {
		select {
		case err := <-errChan:
			if err != nil {
				return 0, nil, nil, fmt.Errorf("container wait error: %w", err)
			}
			continue
		case waitResp := <-waitChan:
			exitCode = int(waitResp.StatusCode)
		case <-ctx.Done():
			return 0, nil, nil, fmt.Errorf("container execution cancelled or timed out")
		}
		break
	}

	return exitCode, stdout.Bytes(), stderr.Bytes(), nil
}

// isRetryableError determines if an error should trigger a retry.
// Windows file locks and transient container errors are retryable.
func isRetryableError(err error, result *containerport.ContainerResult) bool {
	// Success is not retryable
	if err == nil && result != nil && result.ExitCode == 0 {
		return false
	}

	// Check for known retryable patterns in stderr
	if result != nil {
		stderr := string(result.Stderr)
		retryablePatterns := []string{
			"The process cannot access the file", // Windows file lock
			"text file busy",                     // Unix file lock
			"resource temporarily unavailable",   // Transient resource issue
			"Cannot delete",                      // Windows cleanup issue
			"EBUSY",                              // Node.js file busy error
		}
		for _, pattern := range retryablePatterns {
			if strings.Contains(stderr, pattern) {
				return true
			}
		}
	}

	// Check error message for transient Docker errors
	if err != nil {
		errStr := err.Error()
		// Check for image pull failures (transient: rate limits, network issues)
		if strings.Contains(errStr, "failed to pull image") {
			return true
		}
		// Check for container conflict errors
		if strings.Contains(errStr, "container already exists") {
			return true
		}
		// Check for network errors (e.g., "network not found", "network bridge not found")
		if strings.Contains(errStr, "network") && strings.Contains(errStr, "not found") {
			return true
		}
		// Check for removal issues
		if strings.Contains(errStr, "unable to remove") {
			return true
		}
	}

	return false
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

// ManifestExists checks if an image manifest exists in a remote registry.
// Uses the Docker CLI since the Docker SDK does not provide a direct manifest inspect API.
func (a *ContainerAdapter) ManifestExists(ctx context.Context, imageRef string) (bool, error) {
	cmd := exec.CommandContext(ctx, "docker", "manifest", "inspect", imageRef) //nolint:gosec // G204: image ref from trusted config
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

// ImageCreatedTime returns the creation timestamp of a local image.
func (a *ContainerAdapter) ImageCreatedTime(ctx context.Context, imageRef string) (time.Time, error) {
	resp, _, err := a.client.ImageInspectWithRaw(ctx, imageRef)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to inspect image %s: %w", imageRef, err)
	}
	t, err := time.Parse(time.RFC3339Nano, resp.Created)
	if err != nil {
		t, err = time.Parse(time.RFC3339, resp.Created)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse image created time %q: %w", resp.Created, err)
	}
	return t, nil
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

// chooseWriter selects the appropriate output writer based on priority:
// streaming writer > log writer > capture buffer.
func chooseWriter(streaming, log io.Writer, capture *bytes.Buffer) io.Writer {
	if streaming != nil {
		if log != nil {
			return io.MultiWriter(streaming, log)
		}
		return streaming
	}
	if log != nil {
		return io.MultiWriter(capture, log)
	}
	return capture
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

func (a *unavailableAdapter) ManifestExists(_ context.Context, _ string) (bool, error) {
	return false, fmt.Errorf("container runtime not available: %w", a.err)
}

func (a *unavailableAdapter) ImageCreatedTime(_ context.Context, _ string) (time.Time, error) {
	return time.Time{}, fmt.Errorf("container runtime not available: %w", a.err)
}

func (a *unavailableAdapter) IsAvailable() bool {
	return false
}

func (a *unavailableAdapter) Close() error {
	return nil
}
