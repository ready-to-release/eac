// mkdocs.go - Build functions for MkDocs module types
package builders

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

func init() {
	// Register handler for "mkdocs" build system
	RegisterSystem("mkdocs", BuildMkDocsModule)
}

// BuildMkDocsModule builds MkDocs documentation sites using Docker.
// All MkDocs modules use this handler - behavior is the same for all.
func BuildMkDocsModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== Building %s: %s ===", module.Type, module.Moniker)

	// Check for mkdocs.yml at repository root
	mkdocsConfig := filepath.Join(workspaceRoot, "mkdocs.yml")
	if _, err := os.Stat(mkdocsConfig); os.IsNotExist(err) {
		Logln(logWriter, "⚠️  No mkdocs.yml found at: %s", mkdocsConfig)
		Logln(logWriter, "ℹ️  Skipping MkDocs build")
		return 0
	}

	Logln(logWriter, "📚 Building MkDocs site using Docker")
	Logln(logWriter, "   Config: %s", mkdocsConfig)

	// Ensure the Docker image exists
	imageName := "cli-mkdocs:latest"
	dockerfilePath := filepath.Join(workspaceRoot, "containers", "mkdocs", ".Dockerfile")
	contextPath := filepath.Join(workspaceRoot, "containers", "mkdocs")

	if err := ensureMkDocsImage(imageName, dockerfilePath, contextPath, logWriter); err != nil {
		Logln(logWriter, "❌ Failed to ensure Docker image: %v", err)
		return 1
	}

	// Calculate the site output directory
	siteDir := filepath.Join(outputDir, "site")

	if err := os.MkdirAll(siteDir, 0755); err != nil {
		Logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	relSiteDir, err := filepath.Rel(workspaceRoot, siteDir)
	if err != nil {
		Logln(logWriter, "❌ Failed to calculate relative path: %v", err)
		return 1
	}

	dockerVolume := FormatDockerVolumePath(workspaceRoot)
	dockerSiteDir := strings.ReplaceAll(relSiteDir, "\\", "/")

	// Check for --accept-warnings flag
	acceptWarnings := false
	for _, arg := range os.Args {
		if arg == "--accept-warnings" {
			acceptWarnings = true
			break
		}
	}

	buildArgs := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
		imageName,
		"mkdocs", "build",
		"--site-dir", dockerSiteDir,
		"--clean",
		"--strict",
	}

	Logln(logWriter, "   Image: %s", imageName)
	Logln(logWriter, "   Output: %s", siteDir)

	if acceptWarnings {
		Logln(logWriter, "   Mode: accepting warnings (--accept-warnings)")
	} else {
		Logln(logWriter, "   Mode: strict (warnings will fail build)")
	}

	exitCode := RunCommandWithLog(workspaceRoot, logWriter, "docker", buildArgs...)

	if acceptWarnings && exitCode != 0 {
		Logln(logWriter, "⚠️  Build completed with warnings (accepted)")
		exitCode = 0
	}

	if exitCode != 0 {
		Logln(logWriter, "❌ MkDocs build failed")
		return exitCode
	}

	Logln(logWriter, "✅ MkDocs site built successfully")
	Logln(logWriter, "   Output: %s", siteDir)

	return 0
}

// ensureMkDocsImage ensures the cli-mkdocs Docker image exists, building it if necessary
func ensureMkDocsImage(imageName, dockerfilePath, contextPath string, logWriter io.Writer) error {
	cmd := exec.Command("docker", "images", "-q", imageName)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check for Docker image: %w", err)
	}

	if len(strings.TrimSpace(string(output))) > 0 {
		Logln(logWriter, "   Using existing image: %s", imageName)
		return nil
	}

	Logln(logWriter, "   Building Docker image: %s", imageName)

	exitCode := RunCommandWithLog(contextPath, logWriter,
		"docker", "build",
		"-t", imageName,
		"-f", dockerfilePath,
		".")

	if exitCode != 0 {
		return fmt.Errorf("docker build failed with exit code %d", exitCode)
	}

	Logln(logWriter, "   Image built successfully: %s", imageName)
	return nil
}
