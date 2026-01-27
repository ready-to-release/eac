package tool

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Global executor instance for use throughout the codebase.
var (
	globalExecutor   *DefaultExecutor
	globalExecutorMu sync.RWMutex
)

// GlobalExecutor returns the global executor instance.
// Creates a new executor if one hasn't been set.
func GlobalExecutor() *DefaultExecutor {
	globalExecutorMu.RLock()
	e := globalExecutor
	globalExecutorMu.RUnlock()

	if e != nil {
		return e
	}

	// Double-check with write lock
	globalExecutorMu.Lock()
	defer globalExecutorMu.Unlock()

	if globalExecutor == nil {
		globalExecutor = NewExecutor()
	}
	return globalExecutor
}

// SetGlobalExecutor sets the global executor instance.
func SetGlobalExecutor(e *DefaultExecutor) {
	globalExecutorMu.Lock()
	defer globalExecutorMu.Unlock()
	globalExecutor = e
}

// DockerClient defines the interface for Docker operations.
// This mirrors the interface from commands/internal/serve but is defined here
// to avoid internal package import restrictions.
type DockerClient interface {
	ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error)
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	ContainerAttach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error)
	Close() error
}

// realDockerClient wraps the official Docker client.
type realDockerClient struct {
	*client.Client
}

// newDockerClient creates a Docker client.
func newDockerClient() (DockerClient, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	return &realDockerClient{Client: cli}, nil
}

// Executor executes tools against source code.
type Executor interface {
	// Execute runs a tool with the given context.
	Execute(ctx context.Context, tool *ToolDefinition, execCtx *ExecutionContext) (*ExecutionResult, error)

	// Validate checks if a tool can be executed (binary exists, Docker available, etc.).
	Validate(tool *ToolDefinition) error

	// Close releases any resources held by the executor.
	Close() error
}

// DefaultExecutor handles both system and container tool execution.
type DefaultExecutor struct {
	registry     *DefaultRegistry // For requirement validation via tool registry
	dockerClient DockerClient
	dockerErr    error // Lazily initialized; captures Docker client init error
}

// NewExecutor creates a new tool executor.
// Docker client is lazily initialized on first container tool execution.
func NewExecutor() *DefaultExecutor {
	return &DefaultExecutor{}
}

// NewExecutorWithRegistry creates an executor with a registry for requirement validation.
func NewExecutorWithRegistry(registry *DefaultRegistry) *DefaultExecutor {
	return &DefaultExecutor{
		registry: registry,
	}
}

// NewExecutorWithDocker creates an executor with a pre-initialized Docker client.
// Useful for testing or when Docker availability is already known.
func NewExecutorWithDocker(client DockerClient) *DefaultExecutor {
	return &DefaultExecutor{
		dockerClient: client,
	}
}

// Execute runs a tool with the given context.
func (e *DefaultExecutor) Execute(ctx context.Context, tool *ToolDefinition, execCtx *ExecutionContext) (*ExecutionResult, error) {
	if tool == nil {
		return nil, fmt.Errorf("tool is nil")
	}
	if execCtx == nil {
		return nil, fmt.Errorf("execution context is nil")
	}

	// Validate requirements before execution
	if err := e.validateRequirements(tool); err != nil {
		return nil, err
	}

	switch tool.Type {
	case ToolTypeSystem:
		return e.executeSystem(ctx, tool, execCtx)
	case ToolTypeContainer:
		return e.executeContainer(ctx, tool, execCtx)
	default:
		return nil, fmt.Errorf("unknown tool type: %s", tool.Type)
	}
}

// executeSystem runs a system binary tool.
func (e *DefaultExecutor) executeSystem(ctx context.Context, tool *ToolDefinition, execCtx *ExecutionContext) (*ExecutionResult, error) {
	// Build argument list
	args := make([]string, 0, len(tool.Args)+len(execCtx.ArgsOverrides))
	args = append(args, tool.Args...)
	args = append(args, execCtx.ArgsOverrides...)

	// Resolve placeholders in arguments
	for i, arg := range args {
		args[i] = resolvePlaceholders(arg, execCtx.Placeholders)
	}

	// Create command
	cmd := exec.CommandContext(ctx, tool.Binary, args...)

	// Set working directory
	workDir := execCtx.ModuleRoot
	if tool.WorkDir != "" {
		workDir = resolvePlaceholders(tool.WorkDir, execCtx.Placeholders)
	}
	if workDir != "" {
		if filepath.IsAbs(workDir) {
			cmd.Dir = workDir
		} else {
			cmd.Dir = filepath.Join(execCtx.WorkspaceRoot, workDir)
		}
	} else {
		cmd.Dir = execCtx.WorkspaceRoot
	}

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range tool.Env {
		cmd.Env = append(cmd.Env, k+"="+resolvePlaceholders(v, execCtx.Placeholders))
	}
	for k, v := range execCtx.EnvOverrides {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	if execCtx.LogWriter != nil {
		cmd.Stdout = io.MultiWriter(&stdout, execCtx.LogWriter)
		cmd.Stderr = io.MultiWriter(&stderr, execCtx.LogWriter)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	// Execute
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("execution failed: %w", err)
		}
	}

	return &ExecutionResult{
		ExitCode: exitCode,
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: duration,
	}, nil
}

// executeContainer runs a container tool.
func (e *DefaultExecutor) executeContainer(ctx context.Context, tool *ToolDefinition, execCtx *ExecutionContext) (*ExecutionResult, error) {
	// Ensure Docker client is initialized
	if err := e.ensureDockerClient(); err != nil {
		return nil, err
	}

	// Resolve image reference
	imageRef := tool.FullImage()
	if imageRef == "" {
		return nil, fmt.Errorf("container tool %q has no image specified", tool.ID)
	}

	// Pull image if needed
	if err := e.ensureImage(ctx, imageRef); err != nil {
		return nil, err
	}

	// Resolve placeholders in mounts and workdir
	resolvedMounts := make([]string, 0, len(tool.Mounts))
	for _, mount := range tool.Mounts {
		resolved := mount.ResolvePlaceholders(execCtx.Placeholders)

		// Convert to Docker bind mount format: source:target[:ro]
		bind := resolved.Source + ":" + resolved.Target
		if resolved.ReadOnly {
			bind += ":ro"
		}
		resolvedMounts = append(resolvedMounts, bind)
	}

	// Resolve workdir
	workDir := tool.WorkDir
	if workDir != "" {
		workDir = resolvePlaceholders(workDir, execCtx.Placeholders)
	}

	// Build environment variables
	var envVars []string
	for k, v := range tool.Env {
		envVars = append(envVars, k+"="+resolvePlaceholders(v, execCtx.Placeholders))
	}
	for k, v := range execCtx.EnvOverrides {
		envVars = append(envVars, k+"="+v)
	}

	// Build command with overrides
	cmd := tool.Command
	if len(execCtx.ArgsOverrides) > 0 {
		cmd = append(cmd, execCtx.ArgsOverrides...)
	}

	// Build container config
	containerConfig := &container.Config{
		Image:        imageRef,
		Cmd:          cmd,
		Env:          envVars,
		WorkingDir:   workDir,
		AttachStdout: true,
		AttachStderr: true,
	}

	if len(tool.Entrypoint) > 0 {
		containerConfig.Entrypoint = tool.Entrypoint
	}

	if tool.User != "" {
		containerConfig.User = tool.User
	}

	// Build host config
	hostConfig := &container.HostConfig{
		Binds:      resolvedMounts,
		AutoRemove: true,
	}

	if tool.Network != "" {
		hostConfig.NetworkMode = container.NetworkMode(tool.Network)
	}

	if tool.Privileged {
		hostConfig.Privileged = true
	}

	if tool.MemoryLimit != "" {
		memLimit := parseMemoryLimit(tool.MemoryLimit)
		if memLimit > 0 {
			hostConfig.Resources.Memory = memLimit
		}
	}

	// Create unique container name
	containerName := fmt.Sprintf("tool-%s-%d", tool.ID, time.Now().UnixNano())

	// Execute container
	return e.runContainer(ctx, containerConfig, hostConfig, containerName, execCtx.LogWriter)
}

// runContainer creates, starts, and waits for a container to complete.
func (e *DefaultExecutor) runContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, name string, logWriter io.Writer) (*ExecutionResult, error) {
	start := time.Now()

	// Create container
	resp, err := e.dockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	// Attach to container before starting to capture all output
	attachResp, err := e.dockerClient.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	// Start wait before starting container
	waitChan, errChan := e.dockerClient.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	// Start container
	if err := e.dockerClient.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Read output
	var stdout, stderr bytes.Buffer
	var stdoutWriter, stderrWriter io.Writer = &stdout, &stderr
	if logWriter != nil {
		stdoutWriter = io.MultiWriter(&stdout, logWriter)
		stderrWriter = io.MultiWriter(&stderr, logWriter)
	}
	_, _ = stdcopy.StdCopy(stdoutWriter, stderrWriter, attachResp.Reader)

	// Wait for container exit
	var exitCode int
	select {
	case err := <-errChan:
		if err != nil {
			return nil, fmt.Errorf("container wait error: %w", err)
		}
	case waitResp := <-waitChan:
		exitCode = int(waitResp.StatusCode)
	case <-ctx.Done():
		return nil, fmt.Errorf("container execution cancelled")
	}

	duration := time.Since(start)

	return &ExecutionResult{
		ExitCode: exitCode,
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: duration,
	}, nil
}

// ensureDockerClient lazily initializes the Docker client.
func (e *DefaultExecutor) ensureDockerClient() error {
	if e.dockerClient != nil {
		return nil
	}
	if e.dockerErr != nil {
		return e.dockerErr
	}

	cli, err := newDockerClient()
	if err != nil {
		e.dockerErr = fmt.Errorf("Docker is not available: %w", err)
		return e.dockerErr
	}
	e.dockerClient = cli
	return nil
}

// ensureImage checks if an image exists locally and pulls it if not.
func (e *DefaultExecutor) ensureImage(ctx context.Context, imageRef string) error {
	// Check if image exists locally
	images, err := e.dockerClient.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageRef {
				return nil // Image exists
			}
		}
	}

	// Pull image
	reader, err := e.dockerClient.ImagePull(ctx, imageRef, image.PullOptions{})
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

// Validate checks if a tool can be executed.
func (e *DefaultExecutor) Validate(tool *ToolDefinition) error {
	if tool == nil {
		return fmt.Errorf("tool is nil")
	}

	if err := tool.Validate(); err != nil {
		return err
	}

	switch tool.Type {
	case ToolTypeSystem:
		// Check if binary exists in PATH
		_, err := exec.LookPath(tool.Binary)
		if err != nil {
			return fmt.Errorf("binary %q not found in PATH", tool.Binary)
		}
	case ToolTypeContainer:
		// Check Docker availability
		if err := e.ensureDockerClient(); err != nil {
			return err
		}
	}

	return nil
}

// Close releases resources held by the executor.
func (e *DefaultExecutor) Close() error {
	if e.dockerClient != nil {
		return e.dockerClient.Close()
	}
	return nil
}

// resolvePlaceholders replaces placeholders in a string with actual values.
func resolvePlaceholders(s string, placeholders map[string]string) string {
	if placeholders == nil {
		return s
	}
	result := s
	for placeholder, value := range placeholders {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
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

// validateRequirements checks if all tool requirements are met.
// Requirements are verified using the tool registry.
func (e *DefaultExecutor) validateRequirements(tool *ToolDefinition) error {
	if len(tool.Requirements) == 0 {
		return nil
	}

	// Use executor's registry, or fall back to global registry
	registry := e.registry
	if registry == nil {
		registry = GlobalRegistry()
	}

	missing := registry.GetMissingTools(tool.Requirements)
	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("tool %q requirements not met: missing %v", tool.ID, missing)
}
