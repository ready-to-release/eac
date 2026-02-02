// Package design provides constants for the design command
package design

import (
	"time"

	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

const (
	// File and directory names - use repository package for canonical paths.
	WorkspaceFileName = paths.WorkspaceDSL
	WorkspaceJSONFile = "workspace.json"
	SpecsDirectory    = paths.SpecsDir
	DesignDirectory   = paths.DesignDir
	SourceDirectory   = paths.SrcDir
	OutputDirectory   = paths.OutDir

	// Docker configuration.
	DockerWorkspaceMount = "/workspace"
	// Note: StructurizrLitePort removed - ports are now dynamically allocated in 9000-9999 range.

	// Default Docker images (used as fallbacks when tool-config.yml not loaded).
	DefaultStructurizrCLIImage  = "structurizr/cli:latest"
	DefaultStructurizrLiteImage = "structurizr/lite:latest"

	// Timeouts.
	DockerValidationTimeout = 30 * time.Second
	DockerStartTimeout      = 60 * time.Second

	// Buffer limits.
	MaxDockerOutputSize = 10 * 1024 * 1024 // 10MB

	// Validation.
	ValidationResultsFile = "design-validation-results.json"
)

// GetStructurizrCLIImage returns the Structurizr CLI Docker image.
// It first checks tool-config.yml, then falls back to the default.
func GetStructurizrCLIImage() string {
	return tool.GetToolImageWithDefault("structurizr-cli", DefaultStructurizrCLIImage)
}

// GetStructurizrLiteImage returns the Structurizr Lite Docker image.
// It first checks tool-config.yml, then falls back to the default.
func GetStructurizrLiteImage() string {
	return tool.GetToolImageWithDefault("structurizr-lite", DefaultStructurizrLiteImage)
}
