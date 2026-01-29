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

// DinD environment variable constants.
const (
	// EnvHostRepoRoot is the environment variable that contains the host repository root
	// when running inside a Docker container (DinD mode).
	EnvHostRepoRoot = "R2R_HOST_REPOROOT"

	// EnvContainerRepoRoot is the environment variable that contains the container's
	// view of the repository root (typically /var/task).
	EnvContainerRepoRoot = "R2R_CONTAINER_REPOROOT"

	// DefaultContainerRoot is the default path where the repository is mounted
	// inside the r2r CLI container.
	DefaultContainerRoot = "/var/task"
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
	imageManager *ImageManager
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

// SetImageManager sets the image manager for handling local containers and GHCR caching.
func (e *DefaultExecutor) SetImageManager(mgr *ImageManager) {
	e.imageManager = mgr
}

// GetImageManager returns the current image manager, or nil if not set.
func (e *DefaultExecutor) GetImageManager() *ImageManager {
	return e.imageManager
}

// Execute runs a tool with the given context.
func (e *DefaultExecutor) Execute(ctx context.Context, tool *ToolDefinition, execCtx *ExecutionContext) (*ExecutionResult, error) {
	if tool == nil {
		return nil, fmt.Errorf("tool is nil")
	}
	if execCtx == nil {
		return nil, fmt.Errorf("execution context is nil")
	}

	// Populate DinD context from environment (for container path translation)
	e.populateDinDContext(execCtx)

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

	// Resolve and ensure image is available
	imageRef, err := e.resolveContainerImage(ctx, tool, execCtx)
	if err != nil {
		return nil, err
	}

	// Resolve placeholders in mounts and workdir
	resolvedMounts := make([]string, 0, len(tool.Mounts))
	for _, mount := range tool.Mounts {
		resolved := mount.ResolvePlaceholders(execCtx.Placeholders)

		// Ensure mount source is an absolute path (required for Docker bind mounts)
		source := resolved.Source
		if !filepath.IsAbs(source) {
			source = filepath.Join(execCtx.WorkspaceRoot, source)
		}

		// DinD: Translate container path to host path for Docker daemon
		source = execCtx.TranslatePathForMount(source)

		// Format path for Docker volume mount (handles Windows C:\ -> /c/ conversion)
		source = formatDockerVolume(source)

		// Convert to Docker bind mount format: source:target[:ro]
		bind := source + ":" + resolved.Target
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

	// Debug: log container configuration
	if execCtx.LogWriter != nil {
		fmt.Fprintf(execCtx.LogWriter, "[debug] container workdir: %s\n", workDir)
		fmt.Fprintf(execCtx.LogWriter, "[debug] container command: %v\n", tool.Command)
		fmt.Fprintf(execCtx.LogWriter, "[debug] placeholders: %v\n", execCtx.Placeholders)
		if execCtx.IsDinD() {
			fmt.Fprintf(execCtx.LogWriter, "[dind] host workspace: %s\n", execCtx.HostWorkspaceRoot)
			fmt.Fprintf(execCtx.LogWriter, "[dind] container root: %s\n", execCtx.ContainerRepoRoot)
			fmt.Fprintf(execCtx.LogWriter, "[dind] translated mounts: %v\n", resolvedMounts)
		}
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

	// Apply resource limits from tool configuration, with optional amp multiplier.
	// Amp values: 1.0 = no change, 2.0 = double resources, 0.5 = half resources.
	amp := execCtx.Amp
	if amp <= 0 {
		amp = 1.0 // Default: no amplification
	}

	if tool.Resources != nil {
		if tool.Resources.Memory != "" {
			memLimit := parseMemoryLimit(tool.Resources.Memory)
			if memLimit > 0 {
				hostConfig.Resources.Memory = int64(float64(memLimit) * amp)
			}
		}
		if tool.Resources.CPUs > 0 {
			// Convert cpus to NanoCPUs (1 CPU = 1e9 NanoCPUs) with amp
			ampedCPUs := float64(tool.Resources.CPUs) * amp
			hostConfig.Resources.NanoCPUs = int64(ampedCPUs * 1e9)
		}
		if tool.Resources.ShmSize != "" {
			shmSize := parseMemoryLimit(tool.Resources.ShmSize)
			if shmSize > 0 {
				hostConfig.ShmSize = int64(float64(shmSize) * amp)
			}
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
	// We need to wait for waitChan to get the actual exit code.
	// errChan only receives errors from the wait operation itself.
	var exitCode int
	for {
		select {
		case err := <-errChan:
			if err != nil {
				return nil, fmt.Errorf("container wait error: %w", err)
			}
			// errChan received nil - continue waiting for waitChan
			if logWriter != nil {
				fmt.Fprintf(logWriter, "[debug] errChan received nil, waiting for waitChan\n")
			}
			continue
		case waitResp := <-waitChan:
			exitCode = int(waitResp.StatusCode)
			if logWriter != nil {
				fmt.Fprintf(logWriter, "[debug] container exited with code %d (StatusCode: %d, Error: %v)\n",
					exitCode, waitResp.StatusCode, waitResp.Error)
			}
			// Check if there's an error message in the wait response
			if waitResp.Error != nil && waitResp.Error.Message != "" {
				if logWriter != nil {
					fmt.Fprintf(logWriter, "[debug] container error message: %s\n", waitResp.Error.Message)
				}
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("container execution cancelled")
		}
		break
	}

	duration := time.Since(start)

	// Workaround for tools that return exit code 0 despite errors (e.g., mkdocs 1.6.1 bug)
	// Check both stdout and stderr for known error patterns and override exit code if found
	if exitCode == 0 {
		combinedOutput := stdout.String() + stderr.String()
		if containsErrorPattern(combinedOutput) {
			if logWriter != nil {
				fmt.Fprintf(logWriter, "[debug] detected error in output despite exit code 0, overriding to 1\n")
			}
			exitCode = 1
		}
	}

	return &ExecutionResult{
		ExitCode: exitCode,
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: duration,
	}, nil
}

// containsErrorPattern checks if output contains known error patterns from buggy tools.
// Some tools (e.g., mkdocs 1.6.1) return exit code 0 despite printing errors.
func containsErrorPattern(output string) bool {
	errorPatterns := []string{
		"Error:",          // Generic error prefix (mkdocs, etc.)
		"error:",          // Lowercase variant
		"FATAL:",          // Fatal errors
		"fatal:",          // Lowercase variant
		"Exception:",      // Python exceptions
		"Traceback (most", // Python tracebacks
		"panic:",          // Go panics
		"FAILED",          // Test failures
		"Aborted",         // mkdocs strict mode abort
	}

	for _, pattern := range errorPatterns {
		if strings.Contains(output, pattern) {
			return true
		}
	}
	return false
}

// resolveContainerImage resolves and ensures the container image is available.
// Uses ImageManager for local containers and GHCR caching, falls back to direct pull.
func (e *DefaultExecutor) resolveContainerImage(ctx context.Context, tool *ToolDefinition, execCtx *ExecutionContext) (string, error) {
	// Use ImageManager if available (handles local containers and GHCR caching)
	if e.imageManager != nil {
		imageRef, err := e.imageManager.EnsureImage(ctx, tool)
		if err != nil {
			return "", fmt.Errorf("image manager failed: %w", err)
		}
		if imageRef != "" {
			return imageRef, nil
		}
	}

	// Fallback: direct image resolution for external containers without ImageManager
	imageRef := tool.FullImage()
	if imageRef == "" {
		// Local container without ImageManager - try to use local tag
		if tool.IsLocalContainer() {
			imageRef = tool.LocalImageTag()
			if imageRef == "" {
				return "", fmt.Errorf("container tool %q has no image and no localPath", tool.ID)
			}
			// Check if local image exists
			if !e.imageExistsLocally(ctx, imageRef) {
				return "", fmt.Errorf("local container image %q not found - build it first or configure ImageManager", imageRef)
			}
			return imageRef, nil
		}
		return "", fmt.Errorf("container tool %q has no image specified", tool.ID)
	}

	// Pull image if needed
	if err := e.ensureImage(ctx, imageRef); err != nil {
		return "", err
	}

	return imageRef, nil
}

// imageExistsLocally checks if an image exists in the local Docker daemon.
func (e *DefaultExecutor) imageExistsLocally(ctx context.Context, imageRef string) bool {
	images, err := e.dockerClient.ImageList(ctx, image.ListOptions{})
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
	if e.imageExistsLocally(ctx, imageRef) {
		return nil
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

// populateDinDContext sets up DinD-related fields from environment variables.
// When R2R_HOST_REPOROOT is set, it indicates we're running in DinD mode
// and mount paths need to be translated from container paths to host paths.
func (e *DefaultExecutor) populateDinDContext(execCtx *ExecutionContext) {
	hostRoot := os.Getenv(EnvHostRepoRoot)
	if hostRoot == "" {
		return // Not in DinD mode
	}

	execCtx.HostWorkspaceRoot = hostRoot
	execCtx.ContainerRepoRoot = os.Getenv(EnvContainerRepoRoot)
	if execCtx.ContainerRepoRoot == "" {
		execCtx.ContainerRepoRoot = DefaultContainerRoot
	}
}

// formatDockerVolume formats a path for Docker volume mount.
// On Windows, converts paths like C:\path to /c/path for Docker compatibility.
func formatDockerVolume(path string) string {
	// Check if this is a Windows absolute path (e.g., C:\...)
	if len(path) >= 2 && path[1] == ':' {
		// Convert C:\path to /c/path
		driveLetter := strings.ToLower(string(path[0]))
		rest := strings.ReplaceAll(path[2:], "\\", "/")
		return "/" + driveLetter + rest
	}

	// Already Unix-style or relative path - just normalize slashes
	return strings.ReplaceAll(path, "\\", "/")
}
