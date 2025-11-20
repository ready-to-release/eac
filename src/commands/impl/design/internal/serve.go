package design

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ready-to-release/eac/src/core/repository"
)

// StartStructurizrLite starts the Structurizr Lite Docker container and opens the browser
func StartStructurizrLite(moduleName string, autoStop bool) error {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	// Get workspace directory (contains workspace.dsl)
	workspaceDir := filepath.Join(repoRoot, SpecsDirectory, moduleName, DesignDirectory)

	// Verify workspace exists
	workspacePath := filepath.Join(workspaceDir, WorkspaceFileName)
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace file not found: %s", workspacePath)
	}

	// Convert Windows path to Docker volume format
	dockerVolume := formatDockerVolume(workspaceDir)

	// Container name (unique per module)
	containerName := fmt.Sprintf("structurizr-lite-%s", moduleName)

	// Check for port conflicts before starting
	hasConflict, conflictingContainer, err := detectPortConflict(StructurizrLitePort)
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not detect port conflicts: %v\n", err)
		fmt.Println("   Proceeding anyway...")
	} else if hasConflict {
		// Handle the port conflict
		shouldContinue, err := handlePortConflict(conflictingContainer, moduleName, autoStop)
		if err != nil {
			return fmt.Errorf("port conflict handling failed: %w", err)
		}
		if !shouldContinue {
			return fmt.Errorf("operation aborted due to port conflict")
		}
	}

	// Stop and remove existing container with same name if running
	exec.Command("docker", "stop", containerName).Run()
	exec.Command("docker", "rm", containerName).Run()

	// Start Structurizr Lite container
	// Using explicit version tag for consistency with validate command
	fmt.Println("🐳 Starting Structurizr Lite Docker container...")

	// Create context with timeout for Docker start
	ctx, cancel := context.WithTimeout(context.Background(), DockerStartTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "run",
		"--name", containerName,
		"-d",
		"-p", StructurizrLitePort+":"+StructurizrLitePort,
		"-v", dockerVolume+":/usr/local/structurizr",
		StructurizrLiteImage,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("Docker container start timed out after %v", DockerStartTimeout)
		}
		return fmt.Errorf("failed to start Docker container: %w\nOutput: %s", err, string(output))
	}

	fmt.Println("✅ Structurizr Lite started successfully")
	fmt.Println()
	fmt.Printf("📖 Opening browser at http://localhost:%s\n", StructurizrLitePort)
	fmt.Println()
	fmt.Println("💡 To stop the server:")
	fmt.Printf("   docker stop %s\n", containerName)
	fmt.Println()

	// Open browser
	url := "http://localhost:" + StructurizrLitePort
	if err := OpenBrowser(url); err != nil {
		fmt.Printf("⚠️  Could not open browser automatically: %v\n", err)
		fmt.Printf("   Please open manually: %s\n", url)
	}

	return nil
}
