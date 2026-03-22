package serve

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/adapters/docker"
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
func (c *DockerClient) StartDevServer(workspaceRoot string, port int) (*docker.ServeResult, error) {
	// Get the mkdocs-dev-oci tool definition from tool-config.yml
	toolDef := tool.GetToolDefinition("mkdocs-dev-oci")
	if toolDef == nil {
		return nil, fmt.Errorf("mkdocs-dev-oci tool not found in tool-config.yml")
	}

	// Check if repository has its own mkdocs.yml
	// Check both root and docs/ directory (common MkDocs layouts)
	var repoConfigPath string
	rootConfig := filepath.Join(workspaceRoot, "mkdocs.yml")
	docsConfig := filepath.Join(workspaceRoot, "docs", "mkdocs.yml")

	if _, err := os.Stat(rootConfig); err == nil {
		repoConfigPath = "/workspace/mkdocs.yml"
	} else if _, err := os.Stat(docsConfig); err == nil {
		repoConfigPath = "/workspace/docs/mkdocs.yml"
	}

	// Build environment variables for the container
	// These are used by entrypoint.sh to configure MkDocs
	envVars := []string{
		"DOCS_DIR=/workspace/docs",
	}
	if repoConfigPath != "" {
		// Use repository's mkdocs.yml (entrypoint will wrap with INHERIT)
		envVars = append(envVars, "CONFIG_FILE="+repoConfigPath)
	} else {
		// Use container's bundled fallback config
		envVars = append(envVars, "CONFIG_FILE=/docs/mkdocs.yml")
	}

	// Build serve config for dev mode
	// The container's entrypoint handles polling-based file watching
	// Mount workspace root to /workspace so docs are at /workspace/docs
	serveConfig := &docker.ServeConfig{
		Name:          c.containerName,
		ContentPath:   workspaceRoot,   // Mount workspace root
		ContainerPath: "/workspace",    // Mount target
		PreferredPort: port,
		ContainerPort: 8000,
		RestartPolicy: "unless-stopped",
		EnvVars:       envVars,
	}

	// Use LocalImageTag for local containers
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

// IsDevImageStale checks if the dev server image is stale.
// Uses configuration from the tool system (mkdocs-dev-oci tool definition).
func (c *DockerClient) IsDevImageStale(workspaceRoot string) (bool, string, error) {
	// Get the mkdocs-dev-oci tool definition from tool-config.yml
	toolDef := tool.GetToolDefinition("mkdocs-dev-oci")
	if toolDef == nil {
		return false, "", fmt.Errorf("mkdocs-dev-oci tool not found in tool-config.yml")
	}

	// Build serve config from tool definition
	serveConfig := &docker.ServeConfig{}

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

	return docker.CheckImageStale(c.ctx, serveConfig)
}

// StopContainer stops the container.
func (c *DockerClient) StopContainer() error {
	return docker.StopServe(c.ctx, c.containerName)
}

// IsImageStale checks if the container image is stale.
// Uses configuration from the tool system (static-site tool definition).
func (c *DockerClient) IsImageStale(workspaceRoot string) (bool, string, error) {
	// Get the static-site tool definition from tool-config.yml
	toolDef := tool.GetToolDefinition("static-site")
	if toolDef == nil {
		return false, "", fmt.Errorf("static-site tool not found in tool-config.yml")
	}

	// Build serve config from tool definition
	serveConfig := &docker.ServeConfig{}

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

	return docker.CheckImageStale(c.ctx, serveConfig)
}

// OpenBrowserWithFallback opens the browser.
func (c *DockerClient) OpenBrowserWithFallback(url string) (bool, error) {
	return docker.OpenBrowserWithFallback(url)
}

// GetRecentLogs retrieves the last N lines of container logs.
func (c *DockerClient) GetRecentLogs(tailLines string) (string, error) {
	return docker.GetContainerLogs(c.ctx, c.containerName, tailLines)
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
