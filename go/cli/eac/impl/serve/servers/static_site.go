package servers

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/core/tool"
)

// staticSiteServeAdapter wraps the internal serve package for static site serving.
// It reads configuration from GlobalServeContext and delegates to the internal serve package.
func staticSiteServeAdapter(workspaceRoot, moduleRoot, contentPath string, port int, logWriter io.Writer, opts tool.ServeOptions) (*tool.ServeResult, error) {
	ctx := GlobalServeContext
	if ctx == nil {
		// Use defaults if no context set
		ctx = DefaultServeContext()
	}

	// Determine image and Dockerfile
	image := ctx.StaticSiteImage
	if image == "" {
		image = "cli-nginx-tool:latest"
	}

	// Build container name from module
	moduleName := filepath.Base(moduleRoot)
	if moduleName == "." || moduleName == "" {
		moduleName = "site"
	}
	containerName := fmt.Sprintf("cli-serve-%s", moduleName)

	// Build serve config
	serveConfig := &docker.ServeConfig{
		Name:          containerName,
		Image:         image,
		ContentPath:   contentPath,
		ContainerPath: "/usr/share/nginx/html",
		ContainerPort: 8000,
		PreferredPort: port,
		RestartPolicy: "unless-stopped",
	}

	// Add build info if Dockerfile path is available
	if ctx.StaticSiteDockerfile != "" {
		serveConfig.BuildInfo = &docker.BuildInfo{
			Dockerfile:  ctx.StaticSiteDockerfile,
			ContextPath: ctx.StaticSiteContext,
		}
	} else if workspaceRoot != "" {
		// Default Dockerfile location
		dockerfile := filepath.Join(workspaceRoot, "containers/nginx-tool/Dockerfile")
		serveConfig.BuildInfo = &docker.BuildInfo{
			Dockerfile:  dockerfile,
			ContextPath: filepath.Dir(dockerfile),
		}
	}

	// Override container port from context
	if ctx.ContainerPort > 0 {
		serveConfig.ContainerPort = ctx.ContainerPort
	}

	// Start the serve container
	result, err := docker.StartServe(context.Background(), serveConfig)
	if err != nil {
		if logWriter != nil {
			io.WriteString(logWriter, fmt.Sprintf("Error starting static site server: %v\n", err))
		}
		return nil, fmt.Errorf("failed to start static site server: %w", err)
	}

	return &tool.ServeResult{
		Port:    result.HostPort,
		URL:     result.URL,
		Running: true,
	}, nil
}
