// mkdocs.go - Build functions for MkDocs module types
package builders

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
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
	Logln(logWriter, "   WorkspaceRoot: %s", workspaceRoot)

	// For Docker-in-Docker: use host path for volume mount
	// When running inside a container, the Docker daemon runs on the host,
	// so volume mounts need host paths instead of container paths
	hostRepoRoot := workspaceRoot
	isDinD := isDockerInDocker()
	if isDinD {
		if hostRoot := os.Getenv("R2R_HOST_REPOROOT"); hostRoot != "" {
			hostRepoRoot = hostRoot
			Logln(logWriter, "   Docker-in-Docker: using host path %s", hostRoot)
		}
	}

	// Ensure the Docker image exists
	// In DinD mode, docker build runs on host so paths must be host paths
	// We construct these paths manually for cross-platform compatibility
	imageName := "cli-mkdocs:latest"
	var dockerfilePath, contextPath string
	if isDinD {
		// Host is Windows, construct Windows paths manually
		dockerfilePath = hostRepoRoot + "\\containers\\mkdocs\\.Dockerfile"
		contextPath = hostRepoRoot + "\\containers\\mkdocs"
	} else {
		dockerfilePath = filepath.Join(hostRepoRoot, "containers", "mkdocs", ".Dockerfile")
		contextPath = filepath.Join(hostRepoRoot, "containers", "mkdocs")
	}

	if err := ensureMkDocsImage(imageName, dockerfilePath, contextPath, logWriter); err != nil {
		Logln(logWriter, "❌ Failed to ensure Docker image: %v", err)
		return 1
	}

	// Calculate the site output directory (local path within container)
	siteDir := filepath.Join(outputDir, "site")

	if err := os.MkdirAll(siteDir, 0755); err != nil {
		Logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	// Calculate relative site dir for Docker volume mount
	// In DinD mode, we need the relative path from container's workspaceRoot
	// since that maps to the host's repo root
	relSiteDir, err := filepath.Rel(workspaceRoot, siteDir)
	if err != nil {
		Logln(logWriter, "❌ Failed to calculate relative path: %v", err)
		return 1
	}

	volumeMountPath := hostRepoRoot

	dockerVolume := FormatDockerVolumePath(volumeMountPath)
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
	Logln(logWriter, "   Volume: %s:/docs", dockerVolume)
	Logln(logWriter, "   SiteDir: %s", dockerSiteDir)

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
// In DinD mode, dockerfilePath and contextPath are host (Windows) paths
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

	// Use current directory as working dir (doesn't matter for docker build)
	// The context path is passed explicitly to docker build
	exitCode := RunCommandWithLog("", logWriter,
		"docker", "build",
		"-t", imageName,
		"-f", dockerfilePath,
		contextPath)

	if exitCode != 0 {
		return fmt.Errorf("docker build failed with exit code %d", exitCode)
	}

	Logln(logWriter, "   Image built successfully: %s", imageName)
	return nil
}
