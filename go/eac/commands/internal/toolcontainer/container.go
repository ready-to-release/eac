package toolcontainer

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/go/eac/commands/internal/dockerutil"
	"github.com/ready-to-release/eac/go/eac/commands/internal/serve"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

// ContainerRunner pulls containers from registry and executes commands.
type ContainerRunner struct {
	toolName      string
	toolCfg       *config.ContainerToolConfig
	workspaceRoot string
	client        serve.DockerClient
}

// NewContainerRunner creates a new ContainerRunner for the specified tool.
func NewContainerRunner(ctx context.Context, toolName string, toolCfg *config.ContainerToolConfig, workspaceRoot string) (*ContainerRunner, error) {
	client, err := serve.NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	runner := &ContainerRunner{
		toolName:      toolName,
		toolCfg:       toolCfg,
		workspaceRoot: workspaceRoot,
		client:        client,
	}

	return runner, nil
}

// Mode returns ModeContainer.
func (r *ContainerRunner) Mode() Mode {
	return ModeContainer
}

// Run pulls the image if needed and executes the command.
func (r *ContainerRunner) Run(ctx context.Context, cfg *RunConfig) *RunResult {
	// Ensure image is pulled
	if err := r.ensureImage(ctx); err != nil {
		return &RunResult{ExitCode: 1, Error: fmt.Errorf("failed to pull image: %w", err)}
	}

	// Run the container
	return r.runContainer(ctx, cfg)
}

// Close releases the Docker client.
func (r *ContainerRunner) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// ensureImage pulls the Docker image if it doesn't exist locally.
func (r *ContainerRunner) ensureImage(ctx context.Context) error {
	imageName := r.toolCfg.FullImage()
	if imageName == "" {
		return fmt.Errorf("no image configured for container %s", r.toolName)
	}

	// Check if image already exists
	images, err := r.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				log.Debugf("Image %s already exists locally", imageName)
				return nil
			}
		}
	}

	// Image doesn't exist, pull it
	log.Infof("Pulling image %s...", imageName)

	reader, err := r.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read pull output to completion
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("failed to read pull output: %w", err)
	}

	log.Infof("Image %s pulled successfully", imageName)
	return nil
}

// runContainer creates and runs a container with the given configuration.
func (r *ContainerRunner) runContainer(ctx context.Context, cfg *RunConfig) *RunResult {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	imageName := r.toolCfg.FullImage()

	// Prepare command
	cmd := []string{}
	if cfg.Command != "" {
		cmd = append(cmd, cfg.Command)
	}
	cmd = append(cmd, cfg.Args...)

	// Translate host path for DinD if needed
	hostWorkDir := cfg.WorkDir
	if dockerutil.IsDinD() {
		translated, err := dockerutil.TranslatePathForMount(cfg.WorkDir)
		if err != nil {
			return &RunResult{ExitCode: 1, Error: fmt.Errorf("failed to translate path: %w", err)}
		}
		hostWorkDir = translated
	}

	// Container configuration
	containerCfg := &container.Config{
		Image:        imageName,
		Cmd:          cmd,
		WorkingDir:   r.toolCfg.GetWorkdir(),
		Env:          cfg.Env,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  cfg.Stdin != nil,
		OpenStdin:    cfg.Stdin != nil,
		StdinOnce:    cfg.Stdin != nil,
	}

	// Host configuration with volume mount
	hostCfg := &container.HostConfig{
		AutoRemove: true,
		Binds: []string{
			fmt.Sprintf("%s:%s", dockerutil.FormatDockerVolume(hostWorkDir), r.toolCfg.GetWorkdir()),
		},
	}

	// Create container with unique name
	seq := atomic.AddUint64(&containerCounter, 1)
	containerName := fmt.Sprintf("toolcontainer-%s-%d-%d", r.toolName, time.Now().UnixNano(), seq)

	log.Debugf("Creating container: name=%s image=%s cmd=%v", containerName, imageName, cmd)

	resp, err := r.client.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, containerName)
	if err != nil {
		return &RunResult{ExitCode: 1, Error: fmt.Errorf("failed to create container: %w", err)}
	}
	containerID := resp.ID

	// Attach before starting to capture all output
	attachResp, err := r.client.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
		Stdin:  cfg.Stdin != nil,
	})
	if err != nil {
		return &RunResult{ExitCode: 1, Error: fmt.Errorf("failed to attach to container: %w", err)}
	}
	defer attachResp.Close()

	// Start waiting before starting container
	waitChan, errChan := r.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	// Start container
	if err := r.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return &RunResult{ExitCode: 1, Error: fmt.Errorf("failed to start container: %w", err)}
	}

	// Handle stdin if provided
	if cfg.Stdin != nil {
		go func() {
			io.Copy(attachResp.Conn, cfg.Stdin)
			attachResp.CloseWrite()
		}()
	}

	// Read output - use writers or discard
	stdout := cfg.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := cfg.Stderr
	if stderr == nil {
		stderr = io.Discard
	}

	_, copyErr := stdcopy.StdCopy(stdout, stderr, attachResp.Reader)

	// Wait for container to finish
	select {
	case err := <-errChan:
		if err != nil {
			return &RunResult{ExitCode: 1, Error: fmt.Errorf("container wait error: %w", err)}
		}
	case waitResp := <-waitChan:
		log.Debugf("Container completed: id=%s exitCode=%d", containerID, waitResp.StatusCode)
		if copyErr != nil {
			return &RunResult{ExitCode: int(waitResp.StatusCode), Error: fmt.Errorf("failed to read output: %w", copyErr)}
		}
		if waitResp.StatusCode != 0 {
			return &RunResult{ExitCode: int(waitResp.StatusCode), Error: fmt.Errorf("container exited with code %d", waitResp.StatusCode)}
		}
		return &RunResult{ExitCode: 0}
	case <-ctx.Done():
		return &RunResult{ExitCode: 1, Error: fmt.Errorf("container execution timed out")}
	}

	return &RunResult{ExitCode: 0}
}

// Ensure ContainerRunner implements Runner.
var _ Runner = (*ContainerRunner)(nil)
