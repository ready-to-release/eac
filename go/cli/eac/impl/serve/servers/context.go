// Package servers provides serve handlers using the pluggable tool system.
//
// This package delegates all handler registration and lookup to tool.GlobalServeBridge().
// Existing servers register via RegisterServer() in their init() functions,
// and the bridge integrates them with YAML-defined tools from tool-config.yml.
package servers

import "github.com/ready-to-release/eac/go/core/tool"

// ServeContext provides execution-time configuration for servers.
// It is passed as a parameter to serve adapter functions, allowing
// native servers to access configuration that's only available at
// serve execution time (e.g., Docker images from tool-config.yml).
// Use DefaultServeContext() to obtain a context with sensible defaults.
type ServeContext struct {
	// Docker images from tool-config.yml or repository.yml
	StaticSiteImage  string // e.g., "cli-nginx-oci:latest"
	MkDocsLiveImage  string // e.g., "squidfunk/mkdocs-material:latest"
	StructurizrImage string // e.g., "structurizr/lite:latest"

	// Container configuration
	StaticSiteDockerfile string // Path to static site Dockerfile
	StaticSiteContext    string // Docker build context path

	// Port configuration
	DefaultPort   int    // Default port if not specified (e.g., 9000)
	PortRangeMin  int    // Minimum auto-allocated port (e.g., 9000)
	PortRangeMax  int    // Maximum auto-allocated port (e.g., 9999)
	ContainerPort int    // Port inside container (e.g., 8000 for nginx)

	// Browser settings
	OpenBrowser bool   // Automatically open browser after start
	BrowserPath string // URL path to open (e.g., "/" or "/docs/")

	// Watch/live reload settings
	WatchEnabled bool     // Enable file watching for live reload
	WatchPaths   []string // Paths to watch for changes
	WatchExclude []string // Patterns to exclude from watching

	// Execution context
	WorkspaceRoot string // Repository root path
	ModuleMoniker string // Module being served
	ContentPath   string // Path to content being served
}

// DefaultServeContext returns a ServeContext with default values.
// Docker images are looked up from tool-config.yml with hardcoded fallbacks.
func DefaultServeContext() *ServeContext {
	return &ServeContext{
		StaticSiteImage:  tool.GetToolImageWithDefault("nginx-oci", "cli-nginx-oci:latest"),
		MkDocsLiveImage:  tool.GetToolImageWithDefault("mkdocs-live", "squidfunk/mkdocs-material:latest"),
		StructurizrImage: tool.GetToolImageWithDefault("structurizr-lite", "structurizr/lite:latest"),
		DefaultPort:      9000,
		PortRangeMin:     9000,
		PortRangeMax:     9999,
		ContainerPort:    8000,
		OpenBrowser:      true,
		BrowserPath:      "/",
		WatchEnabled:     false,
	}
}
