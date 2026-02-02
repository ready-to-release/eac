// structurizr.go - Build handler for Structurizr architecture diagrams
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	RegisterHandler(&StructurizrHandler{})
}

// StructurizrHandler builds Structurizr diagrams using the structurizr-cli container.
// It exports architecture diagrams to PlantUML format.
type StructurizrHandler struct{}

func (h *StructurizrHandler) Name() string { return "structurizr" }

func (h *StructurizrHandler) Capabilities() []string { return []string{"container", "diagrams"} }

func (h *StructurizrHandler) Requirements() []string { return []string{"docker"} }

func (h *StructurizrHandler) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted (-v /var/run/docker.sock:/var/run/docker.sock)")
		}
		return fmt.Errorf("Docker is not available")
	}

	// Check if workspace.dsl exists
	compRoot := module.GetComponentRoot(component)
	if compRoot == "" {
		return nil // Not an error - component may not exist
	}

	workspaceFile := filepath.Join(workspaceRoot, compRoot, "workspace.dsl")
	if _, err := os.Stat(workspaceFile); os.IsNotExist(err) {
		return fmt.Errorf("workspace.dsl not found at %s", workspaceFile)
	}

	return nil
}

// IsContainer returns true as this handler runs in a Docker container.
func (h *StructurizrHandler) IsContainer() bool { return true }

// IsHostInstalled returns false as this handler requires Docker.
func (h *StructurizrHandler) IsHostInstalled() bool { return false }

func (h *StructurizrHandler) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	return []string{
		"diagrams/",     // PlantUML exports
		"export/",       // Structurizr exports
		"views.json",    // View definitions (if exported)
	}
}

func (h *StructurizrHandler) Build(module interfaces.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moniker := module.GetMoniker()
	Logln(logWriter, "\n=== Building structurizr: %s ===", moniker)

	// Check if Docker is available
	if !IsDockerAvailable() {
		Logln(logWriter, "❌ Docker is not available")
		if IsDockerInDocker() {
			Logln(logWriter, "   Container detected but Docker socket not mounted")
		}
		return 1
	}

	// Find structurizr component root
	compRoot := module.GetComponentRoot("structurizr")
	if compRoot == "" {
		Logln(logWriter, "⚠️  No structurizr component found in module")
		return 0 // Not an error - module may not have structurizr
	}

	absRoot := filepath.Join(workspaceRoot, compRoot)
	workspaceFile := filepath.Join(absRoot, "workspace.dsl")

	if _, err := os.Stat(workspaceFile); os.IsNotExist(err) {
		Logln(logWriter, "⚠️  No workspace.dsl found at %s", workspaceFile)
		return 0 // Not an error if no DSL file exists
	}

	// Get container tool from registry
	containerTool, exists := tool.GlobalRegistry().Get("structurizr-cli")
	if !exists || containerTool == nil {
		Logln(logWriter, "❌ structurizr-cli container tool not found in tool-config.yml")
		return 1
	}

	image := containerTool.FullImage()
	Logln(logWriter, "📐 Using image: %s", image)
	Logln(logWriter, "   Workspace: %s", workspaceFile)
	Logln(logWriter, "   Output: %s", outputDir)

	// Create output directory
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		Logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	// Run structurizr-cli export
	// Command: export -workspace /workspace/workspace.dsl -format plantuml -output /output
	runCfg := &docker.RunConfig{
		Image: image,
		Command: []string{
			"export",
			"-workspace", "/workspace/workspace.dsl",
			"-format", "plantuml",
			"-output", "/output",
		},
		Mounts: []docker.MountConfig{
			{Source: absRoot, Target: "/workspace", ReadOnly: true},
			{Source: outputDir, Target: "/output", ReadOnly: false},
		},
		WorkingDir:    "/workspace",
		Timeout:       5 * time.Minute,
		ContainerName: fmt.Sprintf("eac-structurizr-%s", moniker),
		StreamLogs:    true,
		LogWriter:     logWriter, // Stream output to observer chain
	}

	Logln(logWriter, "🔄 Running structurizr-cli export...")

	ctx := context.Background()
	result, err := docker.RunContainer(ctx, runCfg)
	if err != nil {
		Logln(logWriter, "❌ structurizr-cli failed: %v", err)
		return 1
	}

	if result.ExitCode != 0 {
		Logln(logWriter, "❌ structurizr-cli exited with code %d", result.ExitCode)
		if result.Stderr != "" {
			Logln(logWriter, "stderr: %s", result.Stderr)
		}
		return int(result.ExitCode)
	}

	// Log output info
	if result.Stdout != "" {
		Logln(logWriter, "%s", result.Stdout)
	}

	// Count exported files
	exportCount := 0
	if entries, err := os.ReadDir(outputDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				exportCount++
			}
		}
	}

	Logln(logWriter, "✅ Exported %d diagram files to %s", exportCount, outputDir)

	return 0
}
