package design

import (
	"context"
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/internal/serve"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

var log = logging.C()

// containerNamePrefix is the base name for structurizr-lite containers
// Actual containers will have module name and port suffix, e.g., structurizr-lite-mymodule-9001.
const containerNamePrefix = "structurizr-lite"

// StartStructurizrLite starts the Structurizr Lite Docker container and opens the browser
// The autoStop parameter is kept for backward compatibility but is less relevant with dynamic ports.
func StartStructurizrLite(moduleName string, autoStop bool) error {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	// Get workspace directory (contains workspace.dsl)
	workspaceDir := paths.DesignPath(repoRoot, moduleName)

	// Verify workspace exists
	workspacePath := paths.WorkspaceDSLPath(repoRoot, moduleName)
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace file not found: %s", workspacePath)
	}

	// Container name (unique per module, port will be appended by serve helper)
	containerName := fmt.Sprintf("%s-%s", containerNamePrefix, moduleName)

	// Build configuration for the serve helper
	config := &serve.ServeConfig{
		Name:          containerName,
		Image:         StructurizrLiteImage,
		BuildInfo:     nil, // Use official image from registry
		ContentPath:   workspaceDir,
		ContainerPath: "/usr/local/structurizr",
		ContainerPort: 8080, // Structurizr Lite internal port
		Command:       nil,  // Use image default
		PreferredPort: 0,    // Auto-allocate port in 9000-9999 range
		RestartPolicy: "",   // Use default (unless-stopped)
	}

	// Start Structurizr Lite container
	log.Info("Starting Structurizr Lite Docker container...")

	ctx := context.Background()
	result, err := serve.StartServe(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to start Docker container: %w", err)
	}

	log.Info("Structurizr Lite started successfully")
	log.Info("")
	log.Infof("Opening browser at %s", result.URL)
	log.Info("")
	log.Info("To stop the server:")
	log.Infof("   docker stop %s", result.ContainerName)
	log.Info("")

	// Open browser (skipped in DinD mode)
	opened, err := serve.OpenBrowserWithFallback(result.URL)
	if err != nil {
		log.Infof("Could not open browser automatically: %v", err)
		log.Infof("   Please open manually: %s", result.URL)
	} else if !opened {
		// DinD mode - browser not available
		log.Infof("Open in your browser: %s", result.URL)
	}

	return nil
}

// StopStructurizrLite stops any running Structurizr Lite container for the given module.
func StopStructurizrLite(moduleName string) error {
	containerName := fmt.Sprintf("%s-%s", containerNamePrefix, moduleName)
	ctx := context.Background()
	return serve.StopServe(ctx, containerName)
}

// IsStructurizrLiteRunning checks if Structurizr Lite is running for a module.
func IsStructurizrLiteRunning(moduleName string) (*serve.ServeResult, bool, error) {
	containerName := fmt.Sprintf("%s-%s", containerNamePrefix, moduleName)
	ctx := context.Background()
	return serve.IsServing(ctx, containerName)
}
