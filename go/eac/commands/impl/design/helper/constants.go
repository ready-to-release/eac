// Package design provides constants for the design command
package design

import (
	"time"

	"github.com/ready-to-release/eac/go/eac/core/paths"
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
	StructurizrCLIImage  = "structurizr/cli:latest"
	StructurizrLiteImage = "structurizr/lite:latest"
	// Note: StructurizrLitePort removed - ports are now dynamically allocated in 9000-9999 range.

	// Timeouts.
	DockerValidationTimeout = 30 * time.Second
	DockerStartTimeout      = 60 * time.Second

	// Buffer limits.
	MaxDockerOutputSize = 10 * 1024 * 1024 // 10MB

	// Validation.
	ValidationResultsFile = "design-validation-results.json"
)
