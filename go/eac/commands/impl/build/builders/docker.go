// docker.go - Build handler for Docker build system
package builders

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

func init() {
	// Register handler for "docker" build system (legacy)
	RegisterSystem("docker", BuildDockerModule)
	RegisterSystemArtifacts("docker", ListDockerArtifacts)

	// Register handler for "docker-build" system (new: uses docker_build config from module type)
	RegisterSystem("docker-build", BuildDockerBuildModule)
	RegisterSystemArtifacts("docker-build", ListDockerBuildArtifacts)
}

// ListDockerArtifacts returns the artifacts that would be produced by building this Docker module
func ListDockerArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	// Docker builds produce container images, not files
	// Return image reference pattern for CI to use
	return []string{fmt.Sprintf("docker-image:%s", module.Moniker)}
}

// BuildDockerModule builds a Docker container image.
// Behavior is determined by capabilities:
// - go_module → run go mod tidy first
// - container → build Docker image
func BuildDockerModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== Building %s: %s ===", module.Type, module.Moniker)

	// Check if Docker is available before attempting to build
	if !IsDockerAvailable() {
		Logln(logWriter, "❌ Docker is not available")
		if IsDockerInDocker() {
			Logln(logWriter, "   Container detected but Docker socket not mounted")
			Logln(logWriter, "   Ensure the Docker socket is mounted (-v /var/run/docker.sock:/var/run/docker.sock)")
		} else {
			Logln(logWriter, "   Ensure Docker is installed and the daemon is running")
		}
		return 1
	}

	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	// Get capabilities from contract
	cfg := config.Global()
	hasGoModule := cfg != nil && cfg.ModuleTypes != nil && cfg.ModuleTypes.HasCapability(module.Type, "go_module")

	// Step 1: go mod tidy (if Go module and enabled)
	if hasGoModule && opts.TidyFirst {
		goModPath := filepath.Join(moduleRoot, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			Logln(logWriter, "Running: go mod tidy")
			if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
				Logln(logWriter, "❌ go mod tidy failed")
				return exitCode
			}
		}
	}

	// Find Dockerfile using configured search paths
	var dockerfilePath string
	searchPaths := []string{"containers/{moniker}/Dockerfile", "{root}/Dockerfile"}
	if cfg != nil && cfg.Handlers != nil {
		searchPaths = cfg.Handlers.GetDockerfilePaths()
	}

	for _, pathTemplate := range searchPaths {
		resolvedPath := config.ResolveDockerfilePath(pathTemplate, module.Moniker, module.Files.Root)
		fullPath := filepath.Join(workspaceRoot, resolvedPath)
		if _, err := os.Stat(fullPath); err == nil {
			dockerfilePath = fullPath
			break
		}
	}

	if dockerfilePath == "" {
		Logln(logWriter, "❌ No Dockerfile found")
		return 1
	}

	// Build list of tags for the image
	tags := buildDockerTags(module.Moniker, workspaceRoot)

	Logln(logWriter, "📦 Building Docker image: %s", module.Moniker)
	Logln(logWriter, "   Tags: %v", tags)
	Logln(logWriter, "   Dockerfile: %s", dockerfilePath)
	Logln(logWriter, "   Build context: %s", workspaceRoot)

	isCI := os.Getenv("CI") == "true"

	if isCI {
		return buildDockerCI(module, workspaceRoot, outputDir, dockerfilePath, tags, logWriter)
	}
	return buildDockerLocal(workspaceRoot, outputDir, dockerfilePath, tags, logWriter)
}

// buildDockerTags generates the list of tags for a Docker image.
// For local builds:
//   - {moniker}:latest - for compatibility
//   - {moniker}:local - to distinguish from registry images
//   - {moniker}:local-{sha} - if in a git repo, includes short commit SHA
//
// For CI builds, tags are handled differently in the CI workflow.
func buildDockerTags(moniker string, workspaceRoot string) []string {
	tags := []string{
		fmt.Sprintf("%s:latest", moniker),
		fmt.Sprintf("%s:local", moniker),
	}

	// Try to get git short SHA for more specific tagging
	if sha := getGitShortSHA(workspaceRoot); sha != "" {
		tags = append(tags, fmt.Sprintf("%s:local-%s", moniker, sha))
	}

	return tags
}

// getGitShortSHA returns the short git commit SHA, or empty string if not in a git repo
func getGitShortSHA(workspaceRoot string) string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = workspaceRoot
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// buildDockerLocal builds a Docker image locally using BuildKit for cache support
func buildDockerLocal(workspaceRoot string, outputDir string, dockerfilePath string, tags []string, logWriter io.Writer) int {
	// Check for Docker-in-Docker mode
	isDinD := IsDockerInDocker()
	var contextPath, dockerfileArg string

	if isDinD {
		// In DinD mode, the Docker daemon runs on the host but the client runs in the container.
		// The client needs to tar up the build context locally before sending to the daemon.
		// We use the container's mounted path (workspaceRoot = /var/task) as context because
		// the client can access it. The daemon receives the tarred context and builds locally.
		Logln(logWriter, "   Docker-in-Docker: using container path %s", workspaceRoot)
		contextPath = workspaceRoot
		dockerfileArg = dockerfilePath
	} else {
		contextPath = workspaceRoot
		dockerfileArg = dockerfilePath
	}

	// Build docker command with all tags
	// In DinD mode, use regular 'docker build' instead of 'docker buildx build' because
	// buildx is not always available in container images. Regular build with BuildKit
	// still supports --mount=type=cache directives in Dockerfiles.
	var args []string
	if isDinD {
		args = []string{"build"}
	} else {
		args = []string{"buildx", "build"}
	}
	for _, tag := range tags {
		args = append(args, "-t", tag)
	}
	args = append(args, "-f", dockerfileArg)
	if !isDinD {
		// --load is buildx-specific flag to load image into local docker daemon
		args = append(args, "--load")
	}
	args = append(args, contextPath)

	Logln(logWriter, "   Context: %s", contextPath)
	Logln(logWriter, "   Dockerfile: %s", dockerfileArg)

	// Use buildx (when available) to enable BuildKit cache mounts (RUN --mount=type=cache)
	// This significantly speeds up subsequent builds by caching Go modules and build artifacts
	exitCode := RunCommandWithLog(workspaceRoot, logWriter, "docker", args...)

	if exitCode != 0 {
		Logln(logWriter, "\n❌ Docker build failed")
		return exitCode
	}

	Logln(logWriter, "✅ Docker image built successfully with tags: %v", tags)

	// Save image info
	imageInfoPath := filepath.Join(outputDir, "image-info.txt")
	imageInfo := fmt.Sprintf("Tags: %v\nDockerfile: %s\nBuild Date: %s\n",
		tags, dockerfilePath, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
		Logln(logWriter, "⚠️  Warning: could not save image info: %v", err)
	}

	return 0
}

// buildDockerCI builds a Docker image in CI with multi-platform support
func buildDockerCI(module *modules.ModuleContract, workspaceRoot string, outputDir string, dockerfilePath string, tags []string, logWriter io.Writer) int {
	// Get CI platforms from handlers config
	cfg := config.Global()
	ciPlatforms := "linux/amd64,linux/arm64"
	if cfg != nil && cfg.Handlers != nil {
		ciPlatforms = cfg.Handlers.GetCIPlatformsString()
	}

	// Build docker command with all tags for single-platform
	Logln(logWriter, "\n--- CI Mode: Building single-platform for testing ---")
	args := []string{"buildx", "build", "--platform", "linux/amd64"}
	for _, tag := range tags {
		args = append(args, "-t", tag)
	}
	args = append(args, "-f", dockerfilePath, "--cache-from", "type=gha", "--cache-to", "type=gha,mode=max", "--load", ".")

	exitCode := RunCommandWithLog(workspaceRoot, logWriter, "docker", args...)

	if exitCode != 0 {
		Logln(logWriter, "\n❌ Docker build failed")
		return exitCode
	}
	Logln(logWriter, "✅ Single-platform image built successfully with tags: %v", tags)

	// Export multi-platform for release
	Logln(logWriter, "\n--- CI Mode: Building multi-platform for release ---")
	ociArchivePath := filepath.Join(outputDir, fmt.Sprintf("%s-ci-test.tar", module.Moniker))

	// Build multi-platform with all tags
	args = []string{"buildx", "build", "--platform", ciPlatforms}
	for _, tag := range tags {
		args = append(args, "-t", tag)
	}
	args = append(args, "-f", dockerfilePath, "--cache-from", "type=gha", "-o", fmt.Sprintf("type=oci,dest=%s", ociArchivePath), ".")

	exitCode = RunCommandWithLog(workspaceRoot, logWriter, "docker", args...)

	if exitCode != 0 {
		Logln(logWriter, "\n❌ Multi-platform build failed")
		return exitCode
	}

	Logln(logWriter, "✅ Multi-platform image exported: %s", ociArchivePath)

	// Compress the OCI archive
	Logln(logWriter, "Compressing OCI archive...")
	exitCode = RunCommandWithLog(outputDir, logWriter, "gzip", filepath.Base(ociArchivePath))
	if exitCode != 0 {
		Logln(logWriter, "⚠️  Warning: failed to compress archive")
	}

	// Save image info
	imageInfoPath := filepath.Join(outputDir, "image-info.txt")
	imageInfo := fmt.Sprintf("Tags: %v\nDockerfile: %s\nBuild Date: %s\nPlatforms: %s\nOCI Archive: %s.gz\n",
		tags, dockerfilePath, time.Now().Format(time.RFC3339), ciPlatforms, ociArchivePath)

	if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
		Logln(logWriter, "⚠️  Warning: could not save image info: %v", err)
	}

	return 0
}
