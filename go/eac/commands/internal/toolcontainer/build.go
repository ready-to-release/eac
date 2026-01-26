package toolcontainer

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/go/eac/commands/internal/dockerutil"
	"github.com/ready-to-release/eac/go/eac/commands/internal/serve"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// containerCounter ensures unique container names across parallel executions.
var containerCounter uint64

// log is the package-level logger.
var log = logging.C()

// BuildRunner builds containers from Dockerfile and executes commands.
type BuildRunner struct {
	toolName      string
	toolCfg       *config.ContainerToolConfig
	baseImages    map[string]string
	workspaceRoot string
	imageName     string
	client        serve.DockerClient
}

// NewBuildRunner creates a new BuildRunner for the specified tool.
func NewBuildRunner(ctx context.Context, toolName string, toolCfg *config.ContainerToolConfig, baseImages map[string]string, workspaceRoot string) (*BuildRunner, error) {
	client, err := serve.NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	runner := &BuildRunner{
		toolName:      toolName,
		toolCfg:       toolCfg,
		baseImages:    baseImages,
		workspaceRoot: workspaceRoot,
		imageName:     toolName + ":local",
		client:        client,
	}

	return runner, nil
}

// Mode returns ModeLocal.
func (r *BuildRunner) Mode() Mode {
	return ModeLocal
}

// Run builds the image if needed and executes the command.
func (r *BuildRunner) Run(ctx context.Context, cfg *RunConfig) *RunResult {
	// Ensure image is built
	if err := r.ensureImage(ctx); err != nil {
		return &RunResult{ExitCode: 1, Error: fmt.Errorf("failed to build image: %w", err)}
	}

	// Run the container
	return r.runContainer(ctx, cfg)
}

// Close releases the Docker client.
func (r *BuildRunner) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// ensureImage builds the Docker image if it doesn't exist or is outdated.
func (r *BuildRunner) ensureImage(ctx context.Context) error {
	// Check if image already exists
	images, err := r.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == r.imageName {
				log.Debugf("Image %s already exists locally", r.imageName)
				return nil
			}
		}
	}

	// Image doesn't exist, build it
	log.Infof("Building image %s from Dockerfile...", r.imageName)
	return r.buildImage(ctx)
}

// buildImage builds the Docker image from Dockerfile.
func (r *BuildRunner) buildImage(ctx context.Context) error {
	dockerfilePath := filepath.Join(r.workspaceRoot, r.toolCfg.Dockerfile)
	contextDir := filepath.Dir(dockerfilePath)

	// Create build context tar
	buildContext, err := r.createBuildContext(contextDir)
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}

	// Prepare build args from base_images config
	buildArgs := make(map[string]*string)
	for name, version := range r.baseImages {
		v := version
		// Convert base image name to ARG name (e.g., "python" -> "PYTHON_VERSION")
		argName := strings.ToUpper(strings.ReplaceAll(name, "-", "_")) + "_VERSION"
		buildArgs[argName] = &v
	}

	log.Debugf("Building with args: %v", buildArgs)

	// Build image
	resp, err := r.client.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Tags:       []string{r.imageName},
		Dockerfile: filepath.Base(dockerfilePath),
		BuildArgs:  buildArgs,
		Remove:     true,
	})
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	// Read build output to completion
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("failed to read build output: %w", err)
	}

	log.Infof("Image %s built successfully", r.imageName)
	return nil
}

// createBuildContext creates a tar archive of the Dockerfile directory.
func (r *BuildRunner) createBuildContext(contextDir string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := filepath.Walk(contextDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(contextDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content if it's a regular file
		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}

// runContainer creates and runs a container with the given configuration.
func (r *BuildRunner) runContainer(ctx context.Context, cfg *RunConfig) *RunResult {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

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
		Image:        r.imageName,
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

	log.Debugf("Creating container: name=%s image=%s cmd=%v", containerName, r.imageName, cmd)

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

// Ensure BuildRunner implements Runner.
var _ Runner = (*BuildRunner)(nil)
