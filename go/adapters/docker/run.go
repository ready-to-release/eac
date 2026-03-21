// Package docker provides container running functionality.
package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/go/adapters/docker/util"
)

// MountConfig defines a volume mount for a container.
type MountConfig struct {
	// Source is the host path to mount
	Source string
	// Target is the path inside the container
	Target string
	// ReadOnly makes the mount read-only
	ReadOnly bool
}

// RunConfig defines configuration for running a one-shot container.
type RunConfig struct {
	// Image is the Docker image to use
	Image string
	// Command is the command to run in the container
	Command []string
	// Mounts defines volume mounts
	Mounts []MountConfig
	// EnvVars are environment variables to set
	EnvVars []string
	// WorkingDir is the working directory inside the container
	WorkingDir string
	// Memory is the memory limit in bytes (0 = no limit)
	Memory int64
	// CPUs is the number of CPUs to allocate (0 = no limit)
	CPUs float64
	// Timeout is the maximum time to wait for the container
	Timeout time.Duration
	// ContainerName is an optional name for the container (auto-generated if empty)
	ContainerName string
	// StreamLogs enables real-time log streaming during container execution.
	// When true, logs are streamed to LogWriter as they're produced.
	// When false, logs are captured after container exits into RunResult.
	StreamLogs bool
	// LogWriter receives container output when StreamLogs is true.
	// REQUIRED when StreamLogs is true - the adapter will NOT write to os.Stdout.
	// Core is responsible for providing a tee'd writer (file + observer).
	LogWriter io.Writer
}

// RunResult holds the result of a container run.
type RunResult struct {
	// ExitCode is the container exit code
	ExitCode int64
	// Stdout contains captured stdout output
	Stdout string
	// Stderr contains captured stderr output
	Stderr string
}

// RunContainer runs a container with the given configuration and waits for it to complete.
// This is for one-shot tasks, not long-running services.
func RunContainer(ctx context.Context, config *RunConfig) (*RunResult, error) {
	// Always use a real Docker client. RunContainer is a low-level utility that
	// executes actual containers — mock clients (CLIE_MOCK_DOCKER) are for the
	// higher-level ContainerAdapter/tool executor path only.
	cli, err := newRealDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

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

	// Container configuration
	containerConfig := &container.Config{
		Image:        config.Image,
		Cmd:          config.Command,
		Env:          config.EnvVars,
		WorkingDir:   config.WorkingDir,
		AttachStdout: true,
		AttachStderr: true,
	}

	// Host configuration
	hostConfig := &container.HostConfig{
		Mounts: mounts,
		Resources: container.Resources{
			Memory:   config.Memory,
			NanoCPUs: int64(config.CPUs * 1e9),
		},
	}

	// Generate container name if not specified
	containerName := config.ContainerName
	if containerName == "" {
		containerName = GenerateContainerName()
	}

	// Remove any existing container with the same name
	_ = cli.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true}) //nolint:errcheck

	// Create container
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	// Ensure cleanup on any exit
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		_ = cli.ContainerStop(cleanupCtx, containerID, container.StopOptions{})                //nolint:errcheck
		_ = cli.ContainerRemove(cleanupCtx, containerID, container.RemoveOptions{Force: true}) //nolint:errcheck
	}()

	// Attach to container before starting to reliably capture all output.
	// ContainerLogs after exit can miss output on short-lived containers.
	attachResp, err := cli.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	// Set up wait before starting container
	waitChan, errChan := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	// Start container
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Read output from the attach stream.
	// For StreamLogs mode, tee to the LogWriter while also capturing.
	var stdoutBuf, stderrBuf bytes.Buffer
	var stdoutWriter, stderrWriter io.Writer = &stdoutBuf, &stderrBuf

	if config.StreamLogs && config.LogWriter != nil {
		stdoutWriter = io.MultiWriter(&stdoutBuf, config.LogWriter)
		stderrWriter = io.MultiWriter(&stderrBuf, config.LogWriter)
	}

	// stdcopy.StdCopy demuxes the Docker multiplexed stream into stdout/stderr.
	// This blocks until the stream closes (container exits).
	_, _ = stdcopy.StdCopy(stdoutWriter, stderrWriter, attachResp.Reader) //nolint:errcheck

	// Wait for container to complete
	var exitCode int64
	select {
	case err := <-errChan:
		if err != nil {
			return nil, fmt.Errorf("error waiting for container: %w", err)
		}
	case waitResp := <-waitChan:
		exitCode = waitResp.StatusCode
	case <-ctx.Done():
		return nil, fmt.Errorf("container execution timed out")
	}

	return &RunResult{
		ExitCode: exitCode,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
	}, nil
}

