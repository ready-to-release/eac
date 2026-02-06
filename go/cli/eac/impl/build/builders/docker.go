// docker.go - Build handler for Docker build system
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	h := &DockerHandler{}
	RegisterHandler(h)
	tool.GlobalBuildBridge().RegisterNativeHandler(h)
}

// resolveDockerfilePath resolves a dockerfile path template.
func resolveDockerfilePath(pathTemplate, moniker, root string) string {
	result := pathTemplate
	result = strings.ReplaceAll(result, "{moniker}", moniker)
	result = strings.ReplaceAll(result, "{root}", root)
	return result
}

// DockerHandler builds Docker container images.
type DockerHandler struct{}

func (h *DockerHandler) Name() string { return "docker" }

func (h *DockerHandler) Capabilities() []string { return []string{"container"} }

func (h *DockerHandler) Requirements() []string { return []string{"docker"} }

func (h *DockerHandler) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted (-v /var/run/docker.sock:/var/run/docker.sock)")
		}
		return fmt.Errorf("Docker is not available")
	}
	return nil
}

// IsContainer returns false as Docker builds run docker commands on the host.
func (h *DockerHandler) IsContainer() bool { return false }

// IsHostInstalled returns true as Docker builds use the local docker CLI.
func (h *DockerHandler) IsHostInstalled() bool { return true }

func (h *DockerHandler) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	return []string{fmt.Sprintf("docker-image:%s", module.GetMoniker())}
}

func (h *DockerHandler) Build(module interfaces.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moniker := module.GetMoniker()
	Logln(logWriter, "\n=== Building dockerfile: %s ===", moniker)

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

	// Get relevant package roots (these methods are available via the port interface)
	dockerRoot := module.GetComponentRoot("dockerfile")
	goRoot := module.GetComponentRoot("go")

	// Check if module has go package type (available via port interface)
	hasGoModule := module.HasComponent("go")

	// Step 1: go mod tidy (if Go module and enabled)
	if hasGoModule && opts.TidyFirst && goRoot != "" {
		goModulePath := filepath.Join(workspaceRoot, goRoot)
		goModPath := filepath.Join(goModulePath, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			Logln(logWriter, "Running: go mod tidy")
			if exitCode := RunCommandWithLog(context.Background(), goModulePath, logWriter, "go", "mod", "tidy"); exitCode != 0 {
				Logln(logWriter, "❌ go mod tidy failed")
				return exitCode
			}
		}
	}

	// Find Dockerfile using default search paths
	var dockerfilePath string
	searchPaths := []string{"containers/{moniker}/Dockerfile", "{root}/Dockerfile"}

	for _, pathTemplate := range searchPaths {
		resolvedPath := resolveDockerfilePath(pathTemplate, moniker, dockerRoot)
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

	// Build list of tags for the image (using content hash for cache invalidation)
	tags := buildDockerTags(moniker, module)

	Logln(logWriter, "📦 Building Docker image: %s", moniker)
	Logln(logWriter, "   Tags: %v", tags)
	Logln(logWriter, "   Dockerfile: %s", dockerfilePath)
	Logln(logWriter, "   Build context: %s", workspaceRoot)

	isCI := os.Getenv("CI") == "true"

	if isCI {
		return buildDockerCI(moniker, workspaceRoot, outputDir, dockerfilePath, tags, logWriter, opts)
	}
	return buildDockerLocal(workspaceRoot, outputDir, dockerfilePath, tags, logWriter, opts)
}

// buildDockerTags generates the list of tags for a Docker image.
// For local builds:
//   - {moniker}:local - primary tag, matches tool system expectations (LocalImageTag())
//   - {moniker}:{hash} - content hash of module's owned files for cache invalidation
//
// We explicitly do NOT use :latest as it conflates with registry defaults.
// The content hash tag changes only when actual source files change.
// For CI builds, tags are handled differently in the CI workflow.
func buildDockerTags(moniker string, module interfaces.ModuleContractPort) []string {
	tags := []string{
		fmt.Sprintf("%s:local", moniker),
	}

	// Add content-hash tag for cache validation (changes only when source changes)
	if hash, err := module.GetContentHash(); err == nil && hash != "" {
		tags = append(tags, fmt.Sprintf("%s:%s", moniker, hash))
	}

	return tags
}

// buildDockerLocal builds a Docker image locally using buildx with the docker driver.
// Always uses buildx for consistent behavior and BuildKit features.
func buildDockerLocal(workspaceRoot, outputDir, dockerfilePath string, tags []string, logWriter io.Writer, opts BuildOptions) int {
	// Always use buildx with the default builder (docker driver).
	// The docker driver builds directly in Docker daemon - no slow tarball export.
	// This avoids the docker-container driver which requires exporting images.
	args := []string{"buildx", "build", "--builder", "default"}

	// Apply cache flags from CacheConfig
	if opts.CacheConfig != nil {
		// --skip-cache=local:layer -> --no-cache (bypass BuildKit layer cache)
		if opts.CacheConfig.ShouldForceNoCacheDocker() {
			args = append(args, "--no-cache")
			Logln(logWriter, "   Cache: --no-cache (local:layer skipped)")
		}
		// --skip-cache=local:registry -> --pull (force fresh base images)
		if opts.CacheConfig.ShouldForcePull() {
			args = append(args, "--pull")
			Logln(logWriter, "   Cache: --pull (local:registry skipped)")
		}
	}

	for _, tag := range tags {
		args = append(args, "-t", tag)
	}
	args = append(args, "-f", dockerfilePath)
	// --load loads the image into local docker daemon (required for buildx)
	args = append(args, "--load")
	args = append(args, workspaceRoot)

	// Use buildx to enable BuildKit cache mounts (RUN --mount=type=cache)
	// This significantly speeds up subsequent builds by caching Go modules and build artifacts
	exitCode := RunCommandWithLog(context.Background(), workspaceRoot, logWriter, "docker", args...)

	if exitCode != 0 {
		Logln(logWriter, "\n❌ Docker build failed")
		return exitCode
	}

	Logln(logWriter, "✅ Docker image built successfully with tags: %v", tags)

	// Save image info
	imageInfoPath := filepath.Join(outputDir, "image-info.txt")
	imageInfo := fmt.Sprintf("Tags: %v\nDockerfile: %s\nBuild Date: %s\n",
		tags, dockerfilePath, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0o644); err != nil {
		Logln(logWriter, "⚠️  Warning: could not save image info: %v", err)
	}

	return 0
}

// buildDockerCI builds a Docker image in CI with multi-platform support.
func buildDockerCI(moniker, workspaceRoot, outputDir, dockerfilePath string, tags []string, logWriter io.Writer, opts BuildOptions) int {
	// Default CI platforms
	ciPlatforms := "linux/amd64,linux/arm64"

	// Check if we're in GitHub Actions (has GHA cache available)
	isGitHubActions := os.Getenv("GITHUB_ACTIONS") == "true"

	// Build docker command with all tags for single-platform
	// Use default builder (docker driver) for single-platform - faster than docker-container driver
	Logln(logWriter, "\n--- CI Mode: Building single-platform for testing ---")
	args := []string{"buildx", "build", "--builder", "default", "--platform", "linux/amd64"}

	// Apply cache flags from CacheConfig
	if opts.CacheConfig != nil {
		// --skip-cache=local:layer -> --no-cache (bypass BuildKit layer cache)
		if opts.CacheConfig.ShouldForceNoCacheDocker() {
			args = append(args, "--no-cache")
			Logln(logWriter, "   Cache: --no-cache (local:layer skipped)")
		}
		// --skip-cache=local:registry -> --pull (force fresh base images)
		if opts.CacheConfig.ShouldForcePull() {
			args = append(args, "--pull")
			Logln(logWriter, "   Cache: --pull (local:registry skipped)")
		}
	}

	for _, tag := range tags {
		args = append(args, "-t", tag)
	}
	args = append(args, "-f", dockerfilePath)

	// Only use GitHub Actions cache when actually running in GitHub Actions
	// --skip-cache=remote:layer disables GHA cache
	useGHACache := isGitHubActions
	if opts.CacheConfig != nil && opts.CacheConfig.ShouldSkipRemoteLayer() {
		useGHACache = false
		Logln(logWriter, "   Cache: GHA cache disabled (remote:layer skipped)")
	}
	if useGHACache {
		args = append(args, "--cache-from", "type=gha", "--cache-to", "type=gha,mode=max")
	}

	args = append(args, "--load", ".")

	exitCode := RunCommandWithLog(context.Background(), workspaceRoot, logWriter, "docker", args...)

	if exitCode != 0 {
		Logln(logWriter, "\n❌ Docker build failed")
		return exitCode
	}
	Logln(logWriter, "✅ Single-platform image built successfully with tags: %v", tags)

	// Export multi-platform for release
	Logln(logWriter, "\n--- CI Mode: Building multi-platform for release ---")
	ociArchivePath := filepath.Join(outputDir, fmt.Sprintf("%s-ci-test.tar", moniker))

	// Build multi-platform with all tags
	args = []string{"buildx", "build", "--platform", ciPlatforms}

	// Apply cache flags from CacheConfig
	if opts.CacheConfig != nil {
		if opts.CacheConfig.ShouldForceNoCacheDocker() {
			args = append(args, "--no-cache")
		}
		if opts.CacheConfig.ShouldForcePull() {
			args = append(args, "--pull")
		}
	}

	for _, tag := range tags {
		args = append(args, "-t", tag)
	}
	args = append(args, "-f", dockerfilePath)

	// Only use GitHub Actions cache when actually running in GitHub Actions
	// unless --skip-cache=remote:layer is set
	if useGHACache {
		args = append(args, "--cache-from", "type=gha")
	}

	args = append(args, "-o", fmt.Sprintf("type=oci,dest=%s", ociArchivePath), ".")

	exitCode = RunCommandWithLog(context.Background(), workspaceRoot, logWriter, "docker", args...)

	if exitCode != 0 {
		Logln(logWriter, "\n❌ Multi-platform build failed")
		return exitCode
	}

	Logln(logWriter, "✅ Multi-platform image exported: %s", ociArchivePath)

	// Compress the OCI archive
	Logln(logWriter, "Compressing OCI archive...")
	exitCode = RunCommandWithLog(context.Background(), outputDir, logWriter, "gzip", filepath.Base(ociArchivePath))
	if exitCode != 0 {
		Logln(logWriter, "⚠️  Warning: failed to compress archive")
	}

	// Save image info
	imageInfoPath := filepath.Join(outputDir, "image-info.txt")
	imageInfo := fmt.Sprintf("Tags: %v\nDockerfile: %s\nBuild Date: %s\nPlatforms: %s\nOCI Archive: %s.gz\n",
		tags, dockerfilePath, time.Now().Format(time.RFC3339), ciPlatforms, ociArchivePath)

	if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0o644); err != nil {
		Logln(logWriter, "⚠️  Warning: could not save image info: %v", err)
	}

	return 0
}
