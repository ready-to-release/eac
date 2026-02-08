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

	container "github.com/ready-to-release/eac/contracts/container-runtime/0.1.0"
	"github.com/ready-to-release/eac/go/core/paths"
)

// DinD environment variable constants.
const (
	// EnvHostRepoRoot is the environment variable that contains the host repository root
	// when running inside a Docker container (DinD mode).
	EnvHostRepoRoot = "R2R_HOST_REPOROOT"

	// EnvContainerRepoRoot is the environment variable that contains the container's
	// view of the repository root (typically /var/task).
	EnvContainerRepoRoot = "R2R_CONTAINER_REPOROOT"
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

// ContainerProvider is a function that returns a ContainerPort.
// This allows lazy initialization and dependency injection.
type ContainerProvider func() container.ContainerPort

// defaultContainerProvider is the default provider that uses GlobalContainer from the docker adapter.
// It's set by the docker adapter package via SetDefaultContainerProvider.
var defaultContainerProvider ContainerProvider

// SetDefaultContainerProvider sets the default container provider.
// Called by the docker adapter package during initialization.
func SetDefaultContainerProvider(provider ContainerProvider) {
	defaultContainerProvider = provider
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
	registry          *DefaultRegistry         // For requirement validation via tool registry
	container         container.ContainerPort  // Abstract container port
	containerErr      error                    // Lazily initialized; captures container init error
	containerProvider ContainerProvider        // Provider for container port
	imageManager      *ImageManager
	credentials       *CredentialsConfig       // Host env vars to forward to container tools
}

// NewExecutor creates a new tool executor.
// Container runtime is lazily initialized on first container tool execution.
func NewExecutor() *DefaultExecutor {
	return &DefaultExecutor{}
}

// NewExecutorWithRegistry creates an executor with a registry for requirement validation.
func NewExecutorWithRegistry(registry *DefaultRegistry) *DefaultExecutor {
	return &DefaultExecutor{
		registry: registry,
	}
}

// NewExecutorWithContainer creates an executor with a pre-initialized container port.
// Useful for testing or when container availability is already known.
func NewExecutorWithContainer(c container.ContainerPort) *DefaultExecutor {
	return &DefaultExecutor{
		container: c,
	}
}

// SetContainerProvider sets a custom container provider for this executor.
func (e *DefaultExecutor) SetContainerProvider(provider ContainerProvider) {
	e.containerProvider = provider
}

// SetCredentials sets the global credentials config for host env forwarding to containers.
func (e *DefaultExecutor) SetCredentials(creds *CredentialsConfig) {
	e.credentials = creds
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
	cmd := BuildCommand(ctx, tool, execCtx)
	if cmd == nil {
		return nil, fmt.Errorf("failed to build command for tool %q", tool.ID)
	}

	// Set up stdout: streaming writer or capture buffer
	var stdout bytes.Buffer
	if execCtx.StdoutWriter != nil {
		cmd.Stdout = execCtx.StdoutWriter
	} else if execCtx.LogWriter != nil {
		cmd.Stdout = io.MultiWriter(&stdout, execCtx.LogWriter)
	} else {
		cmd.Stdout = &stdout
	}

	// Set up stderr: streaming writer or capture buffer
	var stderr bytes.Buffer
	if execCtx.StderrWriter != nil {
		cmd.Stderr = execCtx.StderrWriter
	} else if execCtx.LogWriter != nil {
		cmd.Stderr = io.MultiWriter(&stderr, execCtx.LogWriter)
	} else {
		cmd.Stderr = &stderr
	}

	// Set up stdin if provided
	if execCtx.StdinReader != nil {
		cmd.Stdin = execCtx.StdinReader
	}

	// Execute
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	// Kill process group on context cancellation
	if ctx.Err() != nil {
		KillProcessGroup(cmd)
	}

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if ctx.Err() != nil {
			// Context cancelled/timed out — not an unexpected error
			exitCode = 1
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

// executeContainer runs a container tool using the ContainerPort interface.
func (e *DefaultExecutor) executeContainer(ctx context.Context, tool *ToolDefinition, execCtx *ExecutionContext) (*ExecutionResult, error) {
	// Ensure container runtime is initialized
	if err := e.ensureContainer(); err != nil {
		return nil, err
	}

	// Resolve and ensure image is available
	imageRef, err := e.resolveContainerImage(ctx, tool, execCtx)
	if err != nil {
		return nil, err
	}

	// Build mounts using ContainerPort's MountConfig
	var mounts []container.MountConfig
	for _, mount := range tool.Mounts {
		resolved := mount.ResolvePlaceholders(execCtx.Placeholders)

		// Ensure mount source is an absolute path (required for bind mounts)
		source := resolved.Source
		if !filepath.IsAbs(source) {
			source = filepath.Join(execCtx.WorkspaceRoot, source)
		}

		// DinD: Translate container path to host path for Docker daemon
		source = execCtx.TranslatePathForMount(source)

		mounts = append(mounts, container.MountConfig{
			Source:   source,
			Target:   resolved.Target,
			ReadOnly: resolved.ReadOnly,
		})
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
		}
	}

	// Build environment variables map with precedence:
	// 1. Global credentials (host-env, ci-env) - lowest priority
	// 2. Per-tool host-env
	// 3. Static tool.Env from YAML
	// 4. Per-call execCtx.EnvOverrides - highest priority
	env := make(map[string]string)

	// 1. Forward global credential env vars from host
	if e.credentials != nil {
		for _, name := range e.credentials.HostEnv {
			if val := os.Getenv(name); val != "" {
				env[name] = val
			}
		}
		for _, name := range e.credentials.CIEnv {
			if val := os.Getenv(name); val != "" {
				env[name] = val
			}
		}
	}

	// 2. Forward per-tool host env vars (additive to global)
	for _, name := range tool.HostEnv {
		if val := os.Getenv(name); val != "" {
			env[name] = val
		}
	}

	// 3. Static YAML env (overrides forwarded values)
	for k, v := range tool.Env {
		env[k] = resolvePlaceholders(v, execCtx.Placeholders)
	}

	// 4. Per-call overrides (highest priority)
	for k, v := range execCtx.EnvOverrides {
		env[k] = v
	}

	// Debug: log forwarded host env var names (never values)
	if execCtx.LogWriter != nil && (e.credentials != nil || len(tool.HostEnv) > 0) {
		var forwarded []string
		if e.credentials != nil {
			for _, name := range e.credentials.HostEnv {
				if _, ok := env[name]; ok {
					forwarded = append(forwarded, name)
				}
			}
			for _, name := range e.credentials.CIEnv {
				if _, ok := env[name]; ok {
					forwarded = append(forwarded, name)
				}
			}
		}
		for _, name := range tool.HostEnv {
			if _, ok := env[name]; ok {
				forwarded = append(forwarded, name)
			}
		}
		if len(forwarded) > 0 {
			fmt.Fprintf(execCtx.LogWriter, "[debug] forwarded host env: %v\n", forwarded)
		}
	}

	// Build command with overrides
	cmd := tool.Command
	if len(execCtx.ArgsOverrides) > 0 {
		cmd = append(cmd, execCtx.ArgsOverrides...)
	}

	// Build container config
	config := &container.ContainerConfig{
		Image:         imageRef,
		Command:       cmd,
		Entrypoint:    tool.Entrypoint,
		WorkingDir:    workDir,
		Env:           env,
		Mounts:        mounts,
		User:          tool.User,
		Privileged:    tool.Privileged,
		Network:       tool.Network,
		LogWriter:     execCtx.LogWriter,
		StdoutWriter:  execCtx.StdoutWriter,
		StderrWriter:  execCtx.StderrWriter,
		StdinReader:   execCtx.StdinReader,
		ContainerName: fmt.Sprintf("eac-%s-%d", tool.ContainerSafeName(), time.Now().UnixNano()),
	}

	// Apply resource limits from tool configuration, with optional amp multiplier.
	// Amp values: 1.0 = no change, 2.0 = double resources, 0.5 = half resources.
	amp := execCtx.Amp
	if amp <= 0 {
		amp = 1.0 // Default: no amplification
	}

	if tool.Resources != nil {
		resources := &container.ResourceConfig{}
		hasResources := false

		if tool.Resources.Memory != "" {
			memLimit := parseMemoryLimit(tool.Resources.Memory)
			if memLimit > 0 {
				// Apply amp and convert back to string
				ampedMem := int64(float64(memLimit) * amp)
				resources.Memory = formatMemoryBytes(ampedMem)
				hasResources = true
			}
		}
		if tool.Resources.CPUs > 0 {
			resources.CPUs = float64(tool.Resources.CPUs) * amp
			hasResources = true
		}
		if tool.Resources.ShmSize != "" {
			shmSize := parseMemoryLimit(tool.Resources.ShmSize)
			if shmSize > 0 {
				ampedShm := int64(float64(shmSize) * amp)
				resources.ShmSize = formatMemoryBytes(ampedShm)
				hasResources = true
			}
		}

		if hasResources {
			config.Resources = resources
		}
	}

	// Execute container
	result, err := e.container.Execute(ctx, config)
	if err != nil {
		return nil, err
	}

	// Workaround for tools that return exit code 0 despite errors (e.g., mkdocs 1.6.1 bug)
	// Check both stdout and stderr for known error patterns and override exit code if found
	exitCode := result.ExitCode
	if exitCode == 0 {
		combinedOutput := string(result.Stdout) + string(result.Stderr)
		if containsErrorPattern(combinedOutput) {
			if execCtx.LogWriter != nil {
				fmt.Fprintf(execCtx.LogWriter, "[debug] detected error in output despite exit code 0, overriding to 1\n")
			}
			exitCode = 1
		}
	}

	return &ExecutionResult{
		ExitCode: exitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		Duration: result.Duration,
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
			if !e.container.ImageExists(ctx, imageRef) {
				return "", fmt.Errorf("local container image %q not found - build it first or configure ImageManager", imageRef)
			}
			return imageRef, nil
		}
		return "", fmt.Errorf("container tool %q has no image specified", tool.ID)
	}

	// Pull image if needed
	if !e.container.ImageExists(ctx, imageRef) {
		if err := e.container.Pull(ctx, imageRef); err != nil {
			return "", err
		}
	}

	return imageRef, nil
}

// ensureContainer lazily initializes the container runtime.
func (e *DefaultExecutor) ensureContainer() error {
	if e.containerErr != nil {
		return e.containerErr
	}

	// If container not yet set, try to get one from providers
	if e.container == nil {
		// Try custom provider first
		if e.containerProvider != nil {
			e.container = e.containerProvider()
		} else if defaultContainerProvider != nil {
			// Fall back to default provider
			e.container = defaultContainerProvider()
		}
	}

	if e.container == nil {
		e.containerErr = fmt.Errorf("container runtime not configured - no container provider available")
		return e.containerErr
	}

	// Always check availability
	if !e.container.IsAvailable() {
		e.containerErr = fmt.Errorf("container runtime not available")
		return e.containerErr
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
		// Check container runtime availability
		if err := e.ensureContainer(); err != nil {
			return err
		}
	}

	return nil
}

// Close releases resources held by the executor.
func (e *DefaultExecutor) Close() error {
	if e.container != nil {
		return e.container.Close()
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

// formatMemoryBytes formats bytes as a human-readable memory string.
func formatMemoryBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%dg", bytes/GB)
	case bytes >= MB:
		return fmt.Sprintf("%dm", bytes/MB)
	case bytes >= KB:
		return fmt.Sprintf("%dk", bytes/KB)
	default:
		return fmt.Sprintf("%d", bytes)
	}
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
		execCtx.ContainerRepoRoot = paths.ContainerRepoRoot
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
