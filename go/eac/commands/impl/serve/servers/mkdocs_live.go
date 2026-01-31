package servers

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/adapters/docker"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// mkdocsLiveServeAdapter provides live-reloading MkDocs server.
// It reads configuration from GlobalServeContext and delegates to the internal serve package.
func mkdocsLiveServeAdapter(workspaceRoot, moduleRoot, contentPath string, port int, logWriter io.Writer, opts tool.ServeOptions) (*tool.ServeResult, error) {
	ctx := GlobalServeContext
	if ctx == nil {
		ctx = DefaultServeContext()
	}

	// Determine image
	image := ctx.MkDocsLiveImage
	if image == "" {
		image = tool.GetToolImageWithDefault("mkdocs-live", "squidfunk/mkdocs-material:latest")
	}

	// Build container name from module
	moduleName := filepath.Base(moduleRoot)
	if moduleName == "." || moduleName == "" {
		moduleName = "docs"
	}
	containerName := fmt.Sprintf("cli-mkdocs-live-%s", moduleName)

	// For MkDocs live, we mount the docs source directory, not the built output
	// contentPath should point to the docs source (where mkdocs.yml lives)
	docsPath := contentPath
	if docsPath == "" {
		// Default to module root if no content path specified
		docsPath = filepath.Join(workspaceRoot, moduleRoot)
	}

	// Build serve config for MkDocs live server
	serveConfig := &docker.ServeConfig{
		Name:          containerName,
		Image:         image,
		ContentPath:   docsPath,
		ContainerPath: "/docs",
		ContainerPort: 8000, // MkDocs default
		PreferredPort: port,
		RestartPolicy: "unless-stopped",
		Command:       []string{"serve", "--dev-addr=0.0.0.0:8000"},
	}

	// Override container port from context
	if ctx.ContainerPort > 0 {
		serveConfig.ContainerPort = ctx.ContainerPort
	}

	// Start the serve container
	result, err := docker.StartServe(context.Background(), serveConfig)
	if err != nil {
		if logWriter != nil {
			io.WriteString(logWriter, fmt.Sprintf("Error starting MkDocs live server: %v\n", err))
		}
		return nil, fmt.Errorf("failed to start MkDocs live server: %w", err)
	}

	return &tool.ServeResult{
		Port:    result.HostPort,
		URL:     result.URL,
		Running: true,
	}, nil
}
