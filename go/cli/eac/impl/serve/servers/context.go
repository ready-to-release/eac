// Package servers provides serve handlers using the pluggable tool system.
//
// This package delegates all handler registration and lookup to tool.GlobalServeBridge().
// Existing servers register via RegisterServer() in their init() functions,
// and the bridge integrates them with YAML-defined tools from tool-config.yml.
package servers

import "github.com/ready-to-release/eac/go/core/tool"

// ServeContext provides execution-time configuration for servers.
// It's populated by the framework before calling server functions.
// This allows native servers to access configuration that's only
// available at serve execution time (e.g., Docker images from tool-config.yml).
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

// GlobalServeContext is populated by the framework before serve execution.
// This allows native servers to access execution-time configuration.
//
// Usage pattern:
//  1. Framework populates GlobalServeContext before calling server
//  2. Server adapter reads from GlobalServeContext
//  3. Framework clears GlobalServeContext after server completes
//
// Thread safety: The framework serializes server execution per module,
// so concurrent access to GlobalServeContext is not a concern for the
// current single-threaded-per-module design.
var GlobalServeContext *ServeContext

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
