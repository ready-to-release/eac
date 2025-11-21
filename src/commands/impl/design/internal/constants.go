// Package design provides constants for the design command
package design

import "time"

const (
	// File and directory names
	WorkspaceFileName = "workspace.dsl"
	WorkspaceJSONFile = "workspace.json"
	SpecsDirectory    = "specs"
	DesignDirectory   = ".design"
	SourceDirectory   = "src"
	OutputDirectory   = "out"

	// Docker configuration
	DockerWorkspaceMount = "/workspace"
	StructurizrCLIImage  = "structurizr/cli:latest"
	StructurizrLiteImage = "structurizr/lite:latest"
	StructurizrLitePort  = "8080"

	// Timeouts
	DockerValidationTimeout = 30 * time.Second
	DockerStartTimeout      = 60 * time.Second

	// Buffer limits
	MaxDockerOutputSize = 10 * 1024 * 1024 // 10MB

	// Validation
	ValidationResultsFile = "design-validation-results.json"
)
