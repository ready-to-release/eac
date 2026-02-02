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
	cli, err := NewDockerClient()
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
		Image:      config.Image,
		Cmd:        config.Command,
		Env:        config.EnvVars,
		WorkingDir: config.WorkingDir,
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
		containerName = fmt.Sprintf("eac-run-%d", time.Now().UnixNano())
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
		_ = cli.ContainerStop(cleanupCtx, containerID, container.StopOptions{})   //nolint:errcheck
		_ = cli.ContainerRemove(cleanupCtx, containerID, container.RemoveOptions{Force: true}) //nolint:errcheck
	}()

	// Start container
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Set up log streaming if requested (runs concurrently with container)
	var logsDone chan struct{}
	if config.StreamLogs && config.LogWriter != nil {
		logOptions := container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		}

		logs, err := cli.ContainerLogs(ctx, containerID, logOptions)
		if err != nil {
			// Non-fatal: continue without streaming, will capture at end
			// Note: Do NOT write to os.Stderr - TUI may be active
			_ = err // Silently ignore, logs will be captured at container exit
		} else {
			logsDone = make(chan struct{})
			go func() {
				defer close(logsDone)
				defer logs.Close()
				// Stream to provided LogWriter only - never os.Stdout
				_, _ = stdcopy.StdCopy(config.LogWriter, config.LogWriter, logs) //nolint:errcheck
			}()
		}
	}

	// Wait for container to complete
	waitChan, errChan := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

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

	// Wait for log streaming to complete (if active)
	if logsDone != nil {
		<-logsDone
	}

	// Get logs if not streaming (capture for RunResult)
	var stdout, stderr string
	if !config.StreamLogs || config.LogWriter == nil {
		stdout, stderr, _ = getContainerLogsInternal(ctx, cli, containerID)
	}

	return &RunResult{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}, nil
}

// getContainerLogsInternal retrieves container logs.
func getContainerLogsInternal(ctx context.Context, cli DockerClient, containerID string) (stdout, stderr string, err error) {
	logsReader, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", "", err
	}
	defer logsReader.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	_, _ = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, logsReader) //nolint:errcheck

	return stdoutBuf.String(), stderrBuf.String(), nil
}
