package serve

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ready-to-release/eac/go/adapters/docker"
	dockerutil "github.com/ready-to-release/eac/go/adapters/docker/util"
	"github.com/ready-to-release/eac/go/core/tool"
)

// DockerClient wraps the internal serve package for module serving.
type DockerClient struct {
	containerName string
	ctx           context.Context
}

// NewDockerClient creates a new Docker client.
func NewDockerClient(containerName string) (*DockerClient, error) {
	return &DockerClient{
		containerName: containerName,
		ctx:           context.Background(),
	}, nil
}

// Close closes the client (no-op for this wrapper).
func (c *DockerClient) Close() {}

// IsRunning checks if the container is running.
func (c *DockerClient) IsRunning() (bool, *docker.ServeResult, error) {
	result, running, err := docker.IsServing(c.ctx, c.containerName)
	return running, result, err
}

// StartContainer starts the serve container.
// Uses configuration from the tool system (static-site tool definition).
func (c *DockerClient) StartContainer(workspaceRoot, contentPath string, port int) (*docker.ServeResult, error) {
	// Get the static-site tool definition from tool-config.yml
	toolDef := tool.GetToolDefinition("static-site")
	if toolDef == nil {
		return nil, fmt.Errorf("static-site tool not found in tool-config.yml")
	}

	// Build serve config from tool definition
	serveConfig := &docker.ServeConfig{
		Name:          c.containerName,
		ContentPath:   contentPath,
		PreferredPort: port,
	}

	// Use LocalImageTag for local containers, or FullImage for external
	if toolDef.IsLocalContainer() {
		serveConfig.Image = toolDef.LocalImageTag()
		contextPath := toolDef.LocalContextPath(workspaceRoot)
		serveConfig.BuildInfo = &docker.BuildInfo{
			Dockerfile:  filepath.Join(contextPath, "Dockerfile"),
			ContextPath: contextPath,
		}
	} else {
		serveConfig.Image = toolDef.FullImage()
	}

	// Get container path from tool mounts (first mount with {content} source)
	containerPath := "/usr/share/nginx/html" // fallback
	for _, mount := range toolDef.Mounts {
		if mount.Source == "{content}" {
			containerPath = mount.Target
			break
		}
	}
	serveConfig.ContainerPath = containerPath

	// Get serve configuration from tool
	if toolDef.Serve != nil {
		serveConfig.ContainerPort = toolDef.Serve.ContainerPort
		serveConfig.RestartPolicy = toolDef.Serve.RestartPolicy
	}
	if serveConfig.ContainerPort == 0 {
		serveConfig.ContainerPort = 8000 // fallback
	}
	if serveConfig.RestartPolicy == "" {
		serveConfig.RestartPolicy = "unless-stopped"
	}

	return docker.StartServe(c.ctx, serveConfig)
}

// StartDevServer starts MkDocs in live-reload development mode.
// Uses mkdocs-dev-oci container with polling-based file watching.
// stagingDir contains pre-processed docs with expanded command markers.
func (c *DockerClient) StartDevServer(workspaceRoot, stagingDir string, port int) (*docker.ServeResult, error) {
	toolDef := tool.GetToolDefinition("mkdocs-dev-oci")
	if toolDef == nil {
		return nil, fmt.Errorf("mkdocs-dev-oci tool not found in tool-config.yml")
	}

	// Translate workspace root for secondary mount (read-only, for INHERIT config)
	workspaceMountSource, err := dockerutil.TranslatePathForMount(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to translate workspace path: %w", err)
	}

	// Primary mount: staging dir -> /workspace (has expanded docs + mkdocs.yml wrapper)
	// Secondary mount: workspace root -> /source (read-only, for INHERIT from source mkdocs.yml)
	serveConfig := &docker.ServeConfig{
		Name:          c.containerName,
		ContentPath:   stagingDir,    // Primary mount: staging
		ContainerPath: "/workspace",  // Mount target
		PreferredPort: port,
		ContainerPort: 8000,
		RestartPolicy: "unless-stopped",
		EnvVars: []string{
			"DOCS_DIR=/workspace/docs",
			"CONFIG_FILE=/workspace/mkdocs.yml",
		},
		AdditionalMounts: []string{
			fmt.Sprintf("%s:/source:ro", dockerutil.FormatDockerVolume(workspaceMountSource)),
		},
	}

	if toolDef.IsLocalContainer() {
		serveConfig.Image = toolDef.LocalImageTag()
		contextPath := toolDef.LocalContextPath(workspaceRoot)
		serveConfig.BuildInfo = &docker.BuildInfo{
			Dockerfile:  filepath.Join(contextPath, "Dockerfile"),
			ContextPath: contextPath,
		}
	} else {
		serveConfig.Image = toolDef.FullImage()
	}

	return docker.StartServe(c.ctx, serveConfig)
}

// StopContainer stops the container.
func (c *DockerClient) StopContainer() error {
	return docker.StopServe(c.ctx, c.containerName)
}

// OpenBrowserWithFallback opens the browser.
func (c *DockerClient) OpenBrowserWithFallback(url string) (bool, error) {
	return docker.OpenBrowserWithFallback(url)
}

// StreamLogsCtx streams container logs to stdout/stderr using the given context.
// Blocks until the context is cancelled or the container stops.
func (c *DockerClient) StreamLogsCtx(ctx context.Context) error {
	return docker.StreamContainerLogs(ctx, c.containerName)
}

// StreamLogs streams container logs to stdout/stderr.
// Blocks until the context is cancelled or the container stops.
func (c *DockerClient) StreamLogs() error {
	return docker.StreamContainerLogs(c.ctx, c.containerName)
}
