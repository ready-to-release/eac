package servers

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/core/tool"
)

// structurizrServeAdapter provides Structurizr Lite server for architecture diagrams.
// It reads configuration from GlobalServeContext and delegates to the internal serve package.
func structurizrServeAdapter(workspaceRoot, moduleRoot, contentPath string, port int, logWriter io.Writer, opts tool.ServeOptions) (*tool.ServeResult, error) {
	ctx := GlobalServeContext
	if ctx == nil {
		ctx = DefaultServeContext()
	}

	// Determine image
	image := ctx.StructurizrImage
	if image == "" {
		image = tool.GetToolImageWithDefault("structurizr-lite", "structurizr/lite:latest")
	}

	// Build container name from module
	moduleName := filepath.Base(moduleRoot)
	if moduleName == "." || moduleName == "" {
		moduleName = "design"
	}
	containerName := fmt.Sprintf("cli-structurizr-%s", moduleName)

	// For Structurizr, contentPath should point to the workspace directory
	// containing workspace.dsl
	workspacePath := contentPath
	if workspacePath == "" {
		// Default to module root if no content path specified
		workspacePath = filepath.Join(workspaceRoot, moduleRoot)
	}

	// Build serve config for Structurizr Lite
	serveConfig := &docker.ServeConfig{
		Name:          containerName,
		Image:         image,
		ContentPath:   workspacePath,
		ContainerPath: "/usr/local/structurizr",
		ContainerPort: 8080, // Structurizr default
		PreferredPort: port,
		RestartPolicy: "unless-stopped",
	}

	// Start the serve container
	result, err := docker.StartServe(context.Background(), serveConfig)
	if err != nil {
		if logWriter != nil {
			io.WriteString(logWriter, fmt.Sprintf("Error starting Structurizr server: %v\n", err))
		}
		return nil, fmt.Errorf("failed to start Structurizr server: %w", err)
	}

	return &tool.ServeResult{
		Port:    result.HostPort,
		URL:     result.URL,
		Running: true,
	}, nil
}
